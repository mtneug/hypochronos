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

	"docker.io/go-docker/api/types/swarm"
	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/hypochronos/api"
)

func (sh *ServiceHandler) runEventLoop(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event loop started")
	defer log.Debug("Service event loop stopped")

	eventQueue, unsub := sh.EventManager.Sub()
	defer unsub()

	for {
		select {
		case e := <-eventQueue:
			if e.ActorType == api.EventActorType_node &&
				(e.Action == api.EventAction_created || e.Action == api.EventAction_updated) {
				log.Debugf("Received %s_%s event", e.ActorType.String(), e.Action.String())

				ctx2, cancel2 := sh.WithPeriod(ctx)

				sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
					node := nodes[e.ActorID]

					err := sh.applyTimetable(ctx2, &node)
					if err != nil {
						log.WithError(err).Error("Failed to apply timetable to node")
						return
					}

					nodes[e.ActorID] = node
				})

				cancel2()
			}
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
