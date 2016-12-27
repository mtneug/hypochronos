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
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/label"
	"github.com/mtneug/hypochronos/model"
	"github.com/mtneug/hypochronos/pkg/event"
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
			if model.IsServiceEvent(e) {
				log.Debugf("Received %s event", e.Type)

				serviceID, ok := e.Object.(string)
				if ok {
					c.handleServiceEvent(ctx, e, serviceID)
				} else {
					log.
						WithError(errors.New("controller: type assertion failed")).
						Error("Failed to get service ID")
				}
			}
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Controller) handleServiceEvent(ctx context.Context, e event.Event, serviceID string) {
	switch e.Type {
	case model.EventTypeServiceCreated:
		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			changed, err := c.addServiceHandler(ctx, services[serviceID])
			if err != nil {
				log.WithError(err).Error("Could not add service handler")
				return
			} else if changed {
				log.Info("Service handler added")
			}
		})

	case model.EventTypeServiceUpdated:
		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			changed, err := c.updateServiceHandler(ctx, services[serviceID])
			if err != nil {
				log.WithError(err).Error("Could not update service handler")
				return
			} else if changed {
				log.Info("Service handler updated")
			}
		})

	case model.EventTypeServiceDeleted:
		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			changed, err := c.deleteServiceHandler(ctx, serviceID)
			if err != nil {
				log.WithError(err).Error("Could not delete service handler")
				return
			} else if changed {
				log.Info("Service handler deleted")
			}

			delete(services, serviceID)
		})
	}
}

func (c *Controller) addServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
	sh, err := c.constructServiceHandler(srv)
	if err != nil {
		return false, err
	}

	return c.ServiceHandlerMap.AddAndStart(ctx, srv.ID, sh)
}

func (c *Controller) updateServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
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
	return c.ServiceHandlerMap.DeleteAndStop(ctx, serviceID)
}

func (c *Controller) constructServiceHandler(srv swarm.Service) (*servicehandler.ServiceHandler, error) {
	labels := srv.Spec.Labels

	tts := timetable.Spec{}
	err := label.ParseTimetableSpec(&tts, labels)
	if err != nil {
		return nil, err
	}

	sh := servicehandler.New(srv.Spec.Name, tts, c.EventManager, c.NodesMap)
	err = label.ParseServiceHandler(sh, labels)
	if err != nil {
		return nil, err
	}

	return sh, nil
}
