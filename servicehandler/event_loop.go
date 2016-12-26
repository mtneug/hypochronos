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
	"errors"

	log "github.com/Sirupsen/logrus"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/docker"
)

func (sh *ServiceHandler) runEventLoop(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event loop started")
	defer log.Debug("Service event loop stopped")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eventQueue, unsub := sh.EventManager.Sub()
	defer unsub()

	dockerEvent, dockerErr := sh.newDockerEvents(ctx)

	for {
		select {
		case e := <-eventQueue:
			if e.Type == types.EventTypeNodeCreated ||
				e.Type == types.EventTypeNodeUpdated {
				log.Debugf("Received %s event", e.Type)

				nodeID, ok := e.Object.(string)
				if ok {
					sh.applyTimetable(ctx, nodeID)
				} else {
					log.
						WithError(errors.New("servicehandler: type assertion failed")).
						Error("Failed to get node ID")
				}
			}
		case e := <-dockerEvent:
			log.Debugf("Received %s_%s event", e.Type, e.Action)

			containerID := e.Actor.ID
			log.Debugf("Container ID: %s", containerID)
			// TODO: implement
		case <-dockerErr:
			dockerEvent, dockerErr = sh.newDockerEvents(ctx)
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (sh *ServiceHandler) newDockerEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	args := filters.NewArgs()
	args.Add("type", "container")
	args.Add("event", "create")
	args.Add("label", "com.docker.swarm.service.id="+sh.ServiceID)
	return docker.C.Events(ctx, dockerTypes.EventsOptions{Filters: args})
}
