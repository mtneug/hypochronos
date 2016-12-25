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
	"github.com/mtneug/hypochronos/pkg/event"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/pkg/startstopper"
)

// Controller monitors Docker Swarm state.
type Controller struct {
	startstopper.StartStopper

	nodesMap          *store.NodesMap
	serviceHandlerMap startstopper.Map

	eventManager           event.Manager
	nodeEventsPublisher    *nodeEventsPublisher
	serviceEventsPublisher *serviceEventsPublisher
	eventLoop              *eventLoop
}

// New creates a new controller.
func New(nodeUpdatePeriod, serviceUpdatePeriod time.Duration, nodesMap *store.NodesMap,
	serviceHandlerMap startstopper.Map) *Controller {

	em := event.NewConcurrentManager(20)
	np := newNodeEventsPublisher(nodeUpdatePeriod, em.Pub(), nodesMap)
	sp := newServiceEventsPublisher(serviceUpdatePeriod, em.Pub(), serviceHandlerMap)
	el := newEventLoop(em, nodesMap, serviceHandlerMap)

	ctrl := &Controller{
		nodesMap:               nodesMap,
		serviceHandlerMap:      serviceHandlerMap,
		eventManager:           em,
		nodeEventsPublisher:    np,
		serviceEventsPublisher: sp,
		eventLoop:              el,
	}
	ctrl.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(ctrl.run))
	return ctrl
}

func (c *Controller) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Controller started")
	defer log.Debug("Controller stopped")

	group := startstopper.NewGroup([]startstopper.StartStopper{
		c.eventManager,
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

	c.serviceHandlerMap.ForEach(func(key string, serviceHandler startstopper.StartStopper) {
		_ = serviceHandler.Stop(ctx)
	})

	return err
}
