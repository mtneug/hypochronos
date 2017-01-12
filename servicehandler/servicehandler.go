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
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/event"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/hypochronos/timetable"
	"github.com/mtneug/pkg/startstopper"
)

// ServiceHandler manages the schedulability of nodes concerning a specific
// service.
type ServiceHandler struct {
	startstopper.StartStopper

	ServiceID   string
	ServiceName string
	NodesMap    *store.NodesMap

	Period      time.Duration
	MinDuration time.Duration

	Timetable      timetable.Timetable
	TimetableMutex sync.RWMutex

	EventManager event.Manager
	eventLoop    startstopper.StartStopper
	periodLoop   startstopper.StartStopper
}

// New creates a new ServiceHandler.
func New(serviceID, serviceName string, tt timetable.Timetable, em event.Manager, nm *store.NodesMap) *ServiceHandler {
	sh := &ServiceHandler{
		ServiceID:    serviceID,
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

			s := api.State{
				Value:   api.StateValue_undefined.String(),
				Until:   timetable.MaxTime.Unix(),
				Node:    &api.Node{ID: node.ID},
				Service: &api.Service{ID: sh.ServiceID, Name: sh.ServiceName},
			}
			sh.EventManager.Pub() <- event.New(api.EventAction_deleted, s)

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
	sh.TimetableMutex.RLock()
	defer sh.TimetableMutex.RUnlock()

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

func (sh *ServiceHandler) applyTimetable(ctx context.Context, node *swarm.Node) error {
	sh.TimetableMutex.RLock()
	defer sh.TimetableMutex.RUnlock()

	state, until := sh.Timetable.State(node.ID, time.Now().UTC())
	return sh.setState(ctx, node, state, until)
}

func (sh *ServiceHandler) setState(ctx context.Context, node *swarm.Node, state string,
	until time.Time) error {
	now := time.Now().UTC()

	curState := docker.NodeGetServiceStateLabel(node, sh.ServiceName)
	if curState == "" {
		curState = api.StateValue_undefined.String()
	}
	log.Debugf("Current state: %s; New state: %s until %s", curState, state, until)

	// If equal do nothing
	if state == curState {
		return nil
	}

	// TODO: does this belongs here or to the helper?
	if state == api.StateValue_activated.String() && until.Sub(now) < sh.MinDuration {
		log.Debug("Skipping activation: phase is to short")
		return nil
	}

	log.Debug("Setting service state label")
	err := docker.NodeSetServiceStateLabel(ctx, node, sh.ServiceName, state)
	if err != nil {
		return err
	}

	// Sending out event
	action := api.EventAction_updated
	if curState == api.StateValue_undefined.String() {
		action = api.EventAction_created
	}

	s := api.State{
		Value:   state,
		Until:   until.Unix(),
		Node:    &api.Node{ID: node.ID},
		Service: &api.Service{ID: sh.ServiceID, Name: sh.ServiceName},
	}

	sh.EventManager.Pub() <- event.New(action, s)

	return nil
}
