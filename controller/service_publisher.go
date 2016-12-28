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
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/model"
	"github.com/mtneug/hypochronos/pkg/event"
)

func (c *Controller) runServiceEventsPublisher(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event publisher started")
	defer log.Debug("Service event publisher stopped")

	eventQueue := c.EventManager.Pub()
	seen := make(map[string]bool)

	tick := func() {
		services, err := docker.StdClient.ServiceList(ctx, dockerTypes.ServiceListOptions{})
		if err != nil {
			log.WithError(err).Error("Failed to get list of services")
			return
		}

		for _, srv := range services {
			seen[srv.ID] = true
			c.ServicesMap.Write(func(services map[string]swarm.Service) {
				n, ok := services[srv.ID]
				if !ok {
					// Add
					services[srv.ID] = srv
					log.Info("Service added")
					eventQueue <- event.New(model.EventTypeServiceCreated, srv.ID)
				} else if n.Version.Index < srv.Version.Index {
					// Update
					services[srv.ID] = srv
					log.Info("Service updated")
					eventQueue <- event.New(model.EventTypeServiceUpdated, srv.ID)
				}
			})
		}

		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			for id, srv := range services {
				if !seen[id] {
					// Delete
					delete(services, srv.ID)
					log.Info("Service updated")
					eventQueue <- event.New(model.EventTypeServiceDeleted, srv.ID)
				}
				delete(seen, id)
			}
		})
	}

	tick()
	for {
		select {
		case <-time.After(c.ServiceUpdatePeriod):
			tick()
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
