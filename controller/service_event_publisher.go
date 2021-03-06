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

	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	"docker.io/go-docker/api/types/swarm"
	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/event"
	"github.com/mtneug/hypochronos/label"
)

var serviceListOptions types.ServiceListOptions

func init() {
	f := filters.NewArgs()
	f.Add("label", label.TimetableType)
	serviceListOptions = types.ServiceListOptions{Filters: f}
}

func (c *Controller) runServiceEventPublisher(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event publisher started")
	defer log.Debug("Service event publisher stopped")

	seen := make(map[string]bool)

	tick := func() {
		services, err := docker.StdClient.ServiceList(ctx, serviceListOptions)
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

					c.EventManager.Pub() <- event.New(api.EventAction_created, api.Service{ID: srv.ID})
				} else if n.Version.Index < srv.Version.Index {
					// Update
					services[srv.ID] = srv
					log.Info("Service updated")

					c.EventManager.Pub() <- event.New(api.EventAction_updated, api.Service{ID: srv.ID})
				}
			})
		}

		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			for id, srv := range services {
				if !seen[id] {
					// Delete
					delete(services, srv.ID)
					log.Info("Service deleted")

					c.EventManager.Pub() <- event.New(api.EventAction_deleted, api.Service{ID: srv.ID})
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
