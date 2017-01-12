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

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/label"
	"github.com/mtneug/hypochronos/servicehandler"
	"github.com/mtneug/hypochronos/timetable"
)

func (c *Controller) runEventLoop(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Main event loop started")
	defer log.Debug("Main event loop stopped")

	eventQueue, unsub := c.EventManager.Sub()
	defer unsub()

	for {
		select {
		case e := <-eventQueue:
			if e.ActorType == api.EventActorType_service {
				log.Debugf("Received %s_%s event", e.ActorType.String(), e.Action.String())
				c.handleServiceEvent(ctx, e)
			}
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Controller) handleServiceEvent(ctx context.Context, e api.Event) {
	var f func(swarm.Service) (bool, error)
	switch e.Action {
	case api.EventAction_created:
		f = func(s swarm.Service) (bool, error) { return c.addServiceHandler(ctx, s) }
	case api.EventAction_updated:
		f = func(s swarm.Service) (bool, error) { return c.updateServiceHandler(ctx, s) }
	case api.EventAction_deleted:
		f = func(swarm.Service) (bool, error) { return c.deleteServiceHandler(ctx, e.ActorID) }
	}

	c.ServicesMap.Write(func(services map[string]swarm.Service) {
		changed, err := f(services[e.ActorID])
		if err != nil {
			log.WithError(err).Error("Handling service event failed")
		} else if changed {
			log.Infof("Service handler %s", e.Action.String())
		}
	})
}

func (c *Controller) addServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
	c.ServiceHandlerMap.Lock()
	defer c.ServiceHandlerMap.Unlock()

	sh, err := c.constructServiceHandler(srv)
	if err != nil {
		return false, err
	}

	return c.ServiceHandlerMap.AddAndStart(ctx, srv.ID, sh)
}

func (c *Controller) updateServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
	c.ServiceHandlerMap.Lock()
	defer c.ServiceHandlerMap.Unlock()

	// We want to skip updates if non of the hypochronos labels have changed (for
	// instance if the service scale is changed).
	if label.Equal(srv.Spec.Labels, srv.PreviousSpec.Labels) {
		return false, nil
	}

	sh, err := c.constructServiceHandler(srv)
	if err != nil {
		return false, err
	}

	return c.ServiceHandlerMap.UpdateAndRestart(ctx, srv.ID, sh)
}

func (c *Controller) deleteServiceHandler(ctx context.Context, serviceID string) (bool, error) {
	c.ServiceHandlerMap.Lock()
	defer c.ServiceHandlerMap.Unlock()

	return c.ServiceHandlerMap.DeleteAndStop(ctx, serviceID)
}

func (c *Controller) constructServiceHandler(srv swarm.Service) (*servicehandler.ServiceHandler, error) {
	labels := srv.Spec.Labels

	// Timetable spec
	tts := timetable.Spec{}
	err := label.ParseTimetableSpec(&tts, labels)
	if err != nil {
		return nil, err
	}

	// Timetable
	tt := timetable.New(tts)

	// Service handler
	sh := servicehandler.New(srv.ID, srv.Spec.Name, tt, c.EventManager, c.NodesMap)
	err = label.ParseServiceHandler(sh, labels)
	if err != nil {
		return nil, err
	}

	return sh, nil
}
