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

package servicehandler

import (
	"context"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/event"
	"github.com/mtneug/hypochronos/model"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/hypochronos/timetable"
	"github.com/mtneug/pkg/startstopper"
)

// ServiceHandler manages the schedulability of nodes concerning a specific
// service.
type ServiceHandler struct {
	startstopper.StartStopper

	ServiceName string
	NodesMap    *store.NodesMap

	Period      time.Duration
	MinDuration time.Duration

	Timetable      timetable.Timetable
	timetableMutex sync.RWMutex

	EventManager event.Manager
	eventLoop    startstopper.StartStopper
	periodLoop   startstopper.StartStopper
}

// New creates a new ServiceHandler.
func New(serviceName string, tt timetable.Timetable, em event.Manager, nm *store.NodesMap) *ServiceHandler {
	sh := &ServiceHandler{
		ServiceName:  serviceName,
		Timetable:    tt,
		NodesMap:     nm,
		EventManager: em,
	}
	sh.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(sh.run))
	sh.eventLoop = startstopper.NewGo(startstopper.RunnerFunc(sh.runEventLoop))
	sh.periodLoop = startstopper.NewGo(startstopper.RunnerFunc(sh.runPeriodLoop))

	return sh
}

func (sh *ServiceHandler) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service handler started")
	defer log.Debug("Service handler stopped")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	group := startstopper.NewGroup([]startstopper.StartStopper{
		sh.eventLoop,
		sh.periodLoop,
	})

	_ = group.Start(ctx)

	select {
	case <-stopChan:
	case <-ctx.Done():
	}

	_ = group.Stop(ctx)
	err := group.Err(ctx)

	log.Debug("Delete service state labels")
	sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
		errChan := forEachKeyNodePair(ctx, nodes, func(ctx context.Context, key string, node swarm.Node) error {
			err = docker.NodeDeleteServiceStateLabel(ctx, &node, sh.ServiceName)
			if err != nil {
				return err
			}
			nodes[key] = node
			return nil
		})

		for err := range errChan {
			log.WithError(err).Warn("Deletion of service state label failed")
		}
	})

	return err
}

// WithPeriod creates a new context with the deadline set to the end of the
// current period.
func (sh *ServiceHandler) WithPeriod(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, sh.PeriodEnd())
}

// PeriodStart time of current period.
func (sh *ServiceHandler) PeriodStart() time.Time {
	sh.timetableMutex.RLock()
	defer sh.timetableMutex.RUnlock()

	return sh.Timetable.FilledAt
}

// PeriodEnd time of current period.
func (sh *ServiceHandler) PeriodEnd() time.Time {
	return sh.PeriodStart().Add(sh.Period)
}

func forEachKeyNodePair(ctx context.Context, nm map[string]swarm.Node,
	op func(ctx context.Context, key string, node swarm.Node) error) (err <-chan error) {

	var wg sync.WaitGroup
	errChan := make(chan error)

	for k, n := range nm {
		key, node := k, n

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := op(ctx, key, node)
			if err != nil {
				errChan <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}

func forEachContainer(ctx context.Context, containers []types.Container,
	op func(ctx context.Context, container types.Container) error) (err <-chan error) {

	var wg sync.WaitGroup
	errChan := make(chan error)

	for _, c := range containers {
		container := c

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := op(ctx, container)
			if err != nil {
				errChan <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}

func (sh *ServiceHandler) applyTimetable(ctx context.Context, node *swarm.Node) error {
	sh.timetableMutex.RLock()
	defer sh.timetableMutex.RUnlock()

	state, until := sh.Timetable.State(node.ID, time.Now().UTC())
	return sh.setState(ctx, node, state, until)
}

func (sh *ServiceHandler) setState(ctx context.Context, node *swarm.Node, state timetable.State,
	until time.Time) error {
	now := time.Now().UTC()

	curState := timetable.State(docker.NodeGetServiceStateLabel(node, sh.ServiceName))
	if curState == "" {
		curState = timetable.StateUndefined
	}
	log.Debugf("Current state: %s; New state: %s until %s", curState, state, until)

	// If equal do nothing
	if state == curState {
		return nil
	}

	if state == model.StateActivated && until.Sub(now) < sh.MinDuration {
		log.Debug("Skipping activation: phase is to short")
		return nil
	}

	log.Debug("Setting service state label")
	err := docker.NodeSetServiceStateLabel(ctx, node, sh.ServiceName, state)
	if err != nil {
		return err
	}

	// Apply appropriate actions to running containersx
	log.Debug("Retrieving containers from node")
	// TODO: This doesn't work with external Docker Engines
	containers, err := docker.ContainerListNode(ctx, node.ID)
	if err != nil {
		return err
	}

	if state == model.StateActivated {
		log.Debug("Writing container TTL")

		errChan := forEachContainer(ctx, containers, func(ctx context.Context, container types.Container) error {
			err2 := docker.ContainerWriteTTL(ctx, container.ID, until)
			if err2 != nil {
				return err2
			}
			return nil
		})

		for err := range errChan {
			log.WithError(err).Warn("Writing container TTL failed")
		}
	} else {
		// FIX: this should normally be done by Docker Swarm
		log.Debug("Stopping and removing running containers")

		// Get stop grace period
		var timeout *time.Duration
		srv, _, err := docker.StdClient.ServiceInspectWithRaw(ctx, sh.ServiceName)
		if err != nil {
			log.WithError(err).Warn("Failed to get stop grace period of service")
		} else {
			timeout = srv.Spec.TaskTemplate.ContainerSpec.StopGracePeriod
		}

		errChan := forEachContainer(ctx, containers, func(ctx context.Context, container types.Container) error {
			err2 := docker.ContainerStopAndRemoveGracefully(ctx, container.ID, timeout)
			if err2 != nil {
				return err2
			}
			return nil
		})

		for err := range errChan {
			log.WithError(err).Warn("Stopping and removing running containers failed")
		}
	}

	return nil
}
