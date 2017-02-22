// Copyright (c) 2016 Matthias Neugebauer <mtneug@mailbox.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/version"
	"github.com/spf13/cobra"
)

var (
	// nodeID of this worker.
	nodeID string

	// host address of hypochronos.
	host string

	// waitTime after a subscription failure.
	waitTime = 5 * time.Second

	// subTries determines how often a subscription request can fail before the
	// connection is reseted.
	subTries = 10

	// initChan is closed once serviceState is initialized after a connect.
	initChan chan struct{}

	// serviceState maps service IDs to state objects.
	serviceState map[string]*api.State
)

var rootCmd = &cobra.Command{
	Use:           "hypochronos-node-helper",
	Short:         "Temorary prevent Docker Swarm nodes from running certain services",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		levelStr, err := flags.GetString("log-level")
		if err != nil {
			return err
		}
		level, err := log.ParseLevel(levelStr)
		if err != nil {
			return err
		}
		log.SetLevel(level)

		// info
		i, err := flags.GetBool("info")
		if err != nil {
			return err
		}
		if i {
			_ = version.Hypochronos.PrintFull(os.Stdout)
			if docker.Err == nil {
				_ = docker.PrintInfo(context.Background(), os.Stdout)
			} else {
				fmt.Println("docker: not connected")
			}
			os.Exit(0)
		}

		// version
		v, err := flags.GetBool("version")
		if err != nil {
			return err
		}
		if v {
			fmt.Println(version.Hypochronos)
			os.Exit(0)
		}

		// Check if connected to Docker
		if docker.Err != nil {
			return errors.New("cmd: not connected to Docker")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		flags := cmd.Flags()

		var err error
		host, err = flags.GetString("host")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Retrieve node ID
		info, err := docker.StdClient.Info(ctx)
		if err != nil {
			cancel()
			return err
		}
		nodeID = info.Swarm.NodeID
		log.WithField("node", nodeID).Info("Node Helper")

		// Handle termination
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sig
			log.Info("Shutting down")
			cancel()
		}()

		return eventLoop(ctx)
	},
}

func init() {
	rootCmd.Flags().StringP("host", "H", "hypochronos:8080", "hypochronos host to connect to")
	rootCmd.Flags().Bool("info", false, "Print hypochronos environment information and exit")
	rootCmd.Flags().String("log-level", "info", "Log level ('debug', 'info', 'warn', 'error', 'fatal', 'panic')")
	rootCmd.Flags().BoolP("version", "v", false, "Print the version and exit")
}

// Execute invoces the top-level command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("An error occurred")
	}
}

func done(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func eventLoop(ctx context.Context) error {
	// Create hypochronos event channel
	hypochronosEvent, hypochronosErr := hypochronosEvents(ctx)

	// Create Docker event channel
	dockerEvent, dockerErr := docker.EventsContainerCreate(ctx)

	for {
		select {
		case e := <-hypochronosEvent:
			<-initChan
			handleHypochronosEvent(ctx, e)
		case e := <-dockerEvent:
			<-initChan
			handleDockerEvent(ctx, e)

		case <-hypochronosErr:
			time.Sleep(waitTime)
			hypochronosEvent, hypochronosErr = hypochronosEvents(ctx)
		case <-dockerErr:
			dockerEvent, dockerErr = docker.EventsContainerCreate(ctx)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func hypochronosEvents(ctx context.Context) (<-chan api.Event, <-chan error) {
	eventChan := make(chan api.Event, 20)
	errChan := make(chan error, 2)
	initChan = make(chan struct{})

	// Create a new connection
	cc, err := grpc.DialContext(ctx, host, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Error("Connection failed")
		errChan <- err
		return eventChan, errChan
	}

	// Create client
	client := api.NewHypochronosClient(cc)

	// init
	go func() {
		log.Debug("Initialize state")
		defer log.Debug("Initialize state finished")

		defer close(initChan)

		serviceState = make(map[string]*api.State)

		req := &api.StatesAtRequest{
			Time:   time.Now().UTC().Unix(),
			NodeID: nodeID,
		}

		resp, err2 := client.StatesAt(ctx, req)
		if err2 != nil {
			log.WithError(err2).Error("Initialize state failed")
			errChan <- err2
			return
		}

		for _, state := range resp.States {
			err2 = validateState(state)
			if err2 != nil {
				log.WithError(err2).Warn("Invalid state; skipping")
				continue
			}

			serviceState[state.Service.ID] = state
			applyStateToRelevantContainer(ctx, *state)
		}
	}()

	// event publisher
	go func() {
		defer func() {
			if cc != nil {
				<-initChan
				_ = cc.Close()
			}
		}()

		// Formulate request
		req := &api.EventsRequest{
			Filters: &api.Filters{
				Args: map[string]string{
					api.FilterKey_NodeID.String():    nodeID,
					api.FilterKey_ActorType.String(): api.EventActorType_state.String(),
				},
			},
		}

		// Subscribe to events
		var stream api.Hypochronos_EventsClient
		for i := 0; i < subTries && !done(ctx); i++ {
			stream, err = client.Events(ctx, req)
			if err == nil {
				break
			}

			log.WithError(err).Error("Subscription failed")
			time.Sleep(waitTime)
		}
		if err != nil {
			log.WithError(err).Errorf("Giving up after %d tries", subTries)
			errChan <- err
			return
		}

		// Receive events
		for !done(ctx) {
			resp, err := stream.Recv()
			if err != nil {
				if !done(ctx) {
					log.WithError(err).Error("Event reception failed")
					errChan <- err
				}
				return
			}

			if e := resp.Event; e != nil {
				eventChan <- *e
			}
		}
	}()

	return eventChan, errChan
}

func handleHypochronosEvent(ctx context.Context, e api.Event) {
	log.Debugf("Received %s_%s event", e.ActorType.String(), e.Action.String())

	state := e.GetState()

	// Can handle event?
	if e.ActorType != api.EventActorType_state {
		log.WithError(errors.New("helper: not a state event")).Warn("Invalid event; skipping")
		return
	}
	err := validateState(state)
	if err != nil {
		log.WithError(err).Warn("Invalid state; skipping")
	}

	switch e.Action {
	case api.EventAction_created:
		fallthrough
	case api.EventAction_updated:
		serviceState[state.Service.ID] = state
		applyStateToRelevantContainer(ctx, *state)

	case api.EventAction_deleted:
		delete(serviceState, state.Service.ID)
	}
}

func handleDockerEvent(ctx context.Context, e events.Message) {
	log.Debugf("Received %s_%s event", e.Type, e.Action)

	var (
		serviceID string
		state     *api.State
	)

	if e.Type != events.ContainerEventType {
		log.Debug("Not a container event; skipping")
		return
	}
	if e.Action != "create" {
		log.Debug("Not a create event; skipping")
		return
	}
	if serviceID = e.Actor.Attributes[docker.DockerSwarmServiceIDLabel]; serviceID == "" {
		log.Debug("Not a service container; skipping")
		return
	}
	if state = serviceState[serviceID]; state == nil {
		log.Debug("No state information; skipping")
		return
	}

	log.Debug("Retrieving container")
	args := filters.NewArgs()
	args.Add("id", e.Actor.ID)
	opts := types.ContainerListOptions{
		All:     true,
		Filters: args,
	}
	c, err := docker.StdClient.ContainerList(ctx, opts)
	if err != nil {
		log.WithError(err).Error("Retrieving containers failed")
		return
	}

	applyState(ctx, *state, c)
}

func validateState(state *api.State) error {
	if state == nil {
		return errors.New("helper: no state given")
	}
	if state.Node == nil {
		return errors.New("helper: no node given")
	}
	if state.Node.ID != nodeID {
		return errors.New("helper: not current node")
	}
	if state.Service == nil {
		return errors.New("helper: no service given")
	}
	return nil
}

func applyStateToRelevantContainer(ctx context.Context, state api.State) {
	log.Debug("Retrieving containers")
	containers, err := docker.ContainerListService(ctx, state.Service.ID)
	if err != nil {
		log.WithError(err).Error("Retrieving containers failed")
		return
	}

	applyState(ctx, state, containers)
}

func applyState(ctx context.Context, state api.State, cs []types.Container) {
	if len(cs) == 0 {
		log.Debug("No running containers")
		return
	}

	switch state.Value {
	case api.StateValue_activated.String():
		log.Debug("Writing container TTL")

		errChan := docker.ParallelForEachContainer(ctx, cs, func(ctx context.Context, c types.Container) error {
			until := time.Unix(state.Until, 0).UTC()
			err2 := docker.ContainerWriteTTL(ctx, c.ID, until)
			if err2 != nil {
				return err2
			}
			return nil
		})

		for err := range errChan {
			log.WithError(err).Warn("Writing container TTL failed")
		}

	case api.StateValue_deactivated.String():
		// NOTE: this should normally be done by Docker Swarm
		log.Debug("Stopping and removing running containers")

		// Get stop grace period
		var timeout *time.Duration
		srv, _, err := docker.StdClient.ServiceInspectWithRaw(ctx, state.Service.ID)
		if err != nil {
			log.WithError(err).Warn("Failed to get stop grace period of service")
		} else {
			timeout = srv.Spec.TaskTemplate.ContainerSpec.StopGracePeriod
		}

		errChan := docker.ParallelForEachContainer(ctx, cs, func(ctx context.Context, c types.Container) error {
			err2 := docker.ContainerStopAndRemoveGracefully(ctx, c.ID, timeout)
			if err2 != nil {
				return err2
			}
			return nil
		})

		for err := range errChan {
			log.WithError(err).Warn("Stopping and removing running containers failed")
		}
	}
}
