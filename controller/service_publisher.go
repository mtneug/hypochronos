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
	"time"

	log "github.com/Sirupsen/logrus"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/label"
	"github.com/mtneug/hypochronos/pkg/event"
	"github.com/mtneug/hypochronos/servicehandler"
	"github.com/mtneug/pkg/startstopper"
)

var serviceListOptions dockerTypes.ServiceListOptions

func init() {
	f := filters.NewArgs()
	f.Add("label", label.Type)
	serviceListOptions = dockerTypes.ServiceListOptions{Filter: f}
}

type serviceEventsPublisher struct {
	startstopper.StartStopper

	period            time.Duration
	eventQueue        chan<- event.Event
	serviceHandlerMap startstopper.Map

	// stored so that it doesn't need to be reallocated
	seen map[string]bool
}

func newServiceEventsPublisher(p time.Duration, eq chan<- event.Event, shm startstopper.Map) *serviceEventsPublisher {
	sp := &serviceEventsPublisher{
		period:            p,
		eventQueue:        eq,
		serviceHandlerMap: shm,
		seen:              make(map[string]bool),
	}
	sp.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(sp.run))
	return sp
}

func (sp *serviceEventsPublisher) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event publisher started")
	defer log.Debug("Service event publisher stopped")

	for {
		sp.tick(ctx)

		select {
		case <-time.After(sp.period):
			sp.tick(ctx)
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (sp *serviceEventsPublisher) tick(ctx context.Context) {
	services, err := docker.C.ServiceList(ctx, serviceListOptions)
	if err != nil {
		log.WithError(err).Error("Failed to get list of services")
		return
	}

	for _, srv := range services {
		sp.seen[srv.ID] = true

		ss, present := sp.serviceHandlerMap.Get(srv.ID)
		if !present {
			// Add
			sp.eventQueue <- event.New(types.EventTypeServiceCreated, srv)
		} else {
			sh, ok := ss.(*servicehandler.ServiceHandler)
			if !ok {
				log.
					WithError(errors.New("controller: type assertion failed")).
					Error("Failed to get service handler")
				return
			}
			if sh.Service.Version.Index < srv.Version.Index {
				// Update
				sp.eventQueue <- event.New(types.EventTypeServiceUpdated, srv)
			}
		}
	}

	sp.serviceHandlerMap.ForEach(func(id string, ss startstopper.StartStopper) {
		if !sp.seen[id] {
			// Delete
			sh, ok := ss.(*servicehandler.ServiceHandler)
			if !ok {
				log.
					WithError(errors.New("controller: type assertion failed")).
					Error("Failed to get service handler")
				return
			}
			sp.eventQueue <- event.New(types.EventTypeServiceDeleted, sh.Service)
		}
		delete(sp.seen, id)
	})
}
