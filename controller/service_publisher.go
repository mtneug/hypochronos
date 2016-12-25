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
	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/pkg/event"
)

func (c *Controller) runServiceEventsPublisher(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event publisher started")
	defer log.Debug("Service event publisher stopped")

	eventQueue := c.EventManager.Pub()
	seen := make(map[string]bool)

	tick := func() {
		services, err := docker.C.ServiceList(ctx, dockerTypes.ServiceListOptions{})
		if err != nil {
			log.WithError(err).Error("Failed to get list of services")
			return
		}

		for _, service := range services {
			seen[service.ID] = true
			c.ServicesMap.Read(func(services map[string]swarm.Service) {
				n, ok := services[service.ID]
				if !ok {
					// Add
					eventQueue <- event.New(types.EventTypeServiceCreated, service)
				} else if n.Version.Index < service.Version.Index {
					// Update
					eventQueue <- event.New(types.EventTypeServiceUpdated, service)
				}
			})
		}

		c.ServicesMap.Read(func(services map[string]swarm.Service) {
			for id, service := range services {
				if !seen[id] {
					// Delete
					eventQueue <- event.New(types.EventTypeServiceDeleted, service)
				}
				delete(seen, id)
			}
		})
	}

	for {
		tick()

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
