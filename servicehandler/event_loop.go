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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/model"
)

func (sh *ServiceHandler) runEventLoop(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event loop started")
	defer log.Debug("Service event loop stopped")

	eventQueue, unsub := sh.EventManager.Sub()
	defer unsub()

	dockerEvent, dockerErr := docker.EventsServiceContainerCreate(ctx, sh.ServiceName)

	for {
		select {
		case e := <-eventQueue:
			if e.Type == model.EventTypeNodeCreated || e.Type == model.EventTypeNodeUpdated {
				log.Debugf("Received %s event", e.Type)
				ctx2, cancel := sh.WithPeriod(ctx)

				sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
					nodeID := e.Object.(string)
					node := nodes[nodeID]

					err := sh.applyTimetable(ctx2, &node)
					if err != nil {
						log.WithError(err).Error("Failed to apply timetable to node")
						return
					}

					nodes[nodeID] = node
				})

				cancel()
			}

		case e := <-dockerEvent:
			if e.Type == events.ContainerEventType && e.Action == "create" {
				log.Debugf("Received %s_%s event", e.Type, e.Action)
				ctx2, cancel := sh.WithPeriod(ctx)
				sh.timetableMutex.RLock()

				containerID := e.Actor.ID
				nodeID := e.Actor.Attributes["com.docker.swarm.node.id"]

				// at this point it is ensured that the state is "activated"
				_, until := sh.Timetable.State(nodeID, time.Now().UTC())
				docker.ContainerWriteTTL(ctx2, containerID, until)

				sh.timetableMutex.RUnlock()
				cancel()
			}

		case <-dockerErr:
			dockerEvent, dockerErr = docker.EventsServiceContainerCreate(ctx, sh.ServiceName)
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
