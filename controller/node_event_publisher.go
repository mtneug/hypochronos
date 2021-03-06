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
	"docker.io/go-docker/api/types/swarm"
	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/event"
)

func (c *Controller) runNodeEventPublisher(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Node event publisher started")
	defer log.Debug("Node event publisher stopped")

	seen := make(map[string]bool)

	tick := func() {
		nodes, err := docker.StdClient.NodeList(ctx, types.NodeListOptions{})
		if err != nil {
			log.WithError(err).Error("Failed to get list of nodes")
			return
		}

		for _, node := range nodes {
			seen[node.ID] = true
			c.NodesMap.Write(func(nodes map[string]swarm.Node) {
				n, ok := nodes[node.ID]
				if !ok {
					// Add
					nodes[node.ID] = node
					log.Info("Node added")

					c.EventManager.Pub() <- event.New(api.EventAction_created, api.Node{ID: node.ID})
				} else if n.Version.Index < node.Version.Index {
					// Update
					nodes[node.ID] = node
					log.Info("Node updated")

					c.EventManager.Pub() <- event.New(api.EventAction_updated, api.Node{ID: node.ID})
				}
			})
		}

		c.NodesMap.Write(func(nodes map[string]swarm.Node) {
			for id, node := range nodes {
				if !seen[id] {
					// Delete
					delete(nodes, node.ID)
					log.Info("Node deleted")

					c.EventManager.Pub() <- event.New(api.EventAction_deleted, api.Node{ID: node.ID})
				}
				delete(seen, id)
			}
		})
	}

	tick()
	for {
		select {
		case <-time.After(c.NodeUpdatePeriod):
			tick()
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
