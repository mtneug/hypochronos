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

package controller

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/hypochronos/event"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/pkg/startstopper"
)

// Controller monitors Docker Swarm state.
type Controller struct {
	startstopper.StartStopper

	Addr string

	NodeUpdatePeriod    time.Duration
	ServiceUpdatePeriod time.Duration

	NodesMap          *store.NodesMap
	ServicesMap       *store.ServicesMap
	ServiceHandlerMap startstopper.Map

	EventManager           event.Manager
	eventLoop              startstopper.StartStopper
	nodeEventsPublisher    startstopper.StartStopper
	serviceEventsPublisher startstopper.StartStopper
}

// New creates a new controller.
func New(addr string, nodeUpdatePeriod, serviceUpdatePeriod time.Duration) *Controller {
	ctrl := &Controller{
		Addr:                addr,
		NodeUpdatePeriod:    nodeUpdatePeriod,
		ServiceUpdatePeriod: serviceUpdatePeriod,
		NodesMap:            store.NewNodesMap(),
		ServicesMap:         store.NewServicesMap(),
		ServiceHandlerMap:   startstopper.NewInMemoryMap(),
		EventManager:        event.NewConcurrentManager(20),
	}

	ctrl.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(ctrl.run))
	ctrl.eventLoop = startstopper.NewGo(startstopper.RunnerFunc(ctrl.runEventLoop))
	ctrl.nodeEventsPublisher = startstopper.NewGo(startstopper.RunnerFunc(ctrl.runNodeEventsPublisher))
	ctrl.serviceEventsPublisher = startstopper.NewGo(startstopper.RunnerFunc(ctrl.runServiceEventsPublisher))

	return ctrl
}

func (c *Controller) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Controller started")
	defer log.Debug("Controller stopped")

	group := startstopper.NewGroup([]startstopper.StartStopper{
		c.EventManager,
		c.nodeEventsPublisher,
		c.serviceEventsPublisher,
		c.eventLoop,
	})

	_ = group.Start(ctx)

	select {
	case <-stopChan:
	case <-ctx.Done():
	}

	_ = group.Stop(ctx)
	err := group.Err(ctx)

	c.ServiceHandlerMap.ForEach(func(key string, serviceHandler startstopper.StartStopper) {
		_ = serviceHandler.Stop(ctx)
	})

	return err
}
