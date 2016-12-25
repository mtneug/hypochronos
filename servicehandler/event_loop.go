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
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/api/types"
)

func (sh *ServiceHandler) runEventLoop(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service event loop started")
	defer log.Debug("Service event loop stopped")

	eventQueue, unsub := sh.EventManager.Sub()
	defer unsub()

	for {
		select {
		case e := <-eventQueue:
			if e.Type == types.EventTypeNodeCreated ||
				e.Type == types.EventTypeNodeUpdated {
				log.Debugf("Received %s event", e.Type)

				n, ok := e.Object.(swarm.Node)
				if ok {
					sh.applyTimetable(ctx, n.ID)
				} else {
					log.
						WithError(errors.New("servicehandler: type assertion failed")).
						Error("Failed to get node")
				}
			}
			// TODO: add case container_create
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
