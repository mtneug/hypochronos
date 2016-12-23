// Copyright (c) 2016 Matthias Neugebauer <mtneug@mailbox.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Uless required by applicable law or agreed to in writing, software
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
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/pkg/startstopper"
)

type nodeEventsPublisher struct {
	startstopper.StartStopper

	period     time.Duration
	eventQueue chan<- event.Event
	nodesMap   *store.NodesMap

	// stored so that it doesn't need to be reallocated
	seen map[string]bool
}

func newNodeEventsPublisher(p time.Duration, eq chan<- event.Event, m *store.NodesMap) *nodeEventsPublisher {
	np := &nodeEventsPublisher{
		period:     p,
		eventQueue: eq,
		nodesMap:   m,
		seen:       make(map[string]bool),
	}
	np.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(np.run))
	return np
}

func (np *nodeEventsPublisher) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Node event publisher started")
	defer log.Debug("Node event publisher stopped")

	for {
		np.tick(ctx)

		select {
		case <-time.After(np.period):
			np.tick(ctx)
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (np *nodeEventsPublisher) tick(ctx context.Context) {
	nodes, err := docker.C.NodeList(ctx, dockerTypes.NodeListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed to get list of nodes")
		return
	}

	for _, node := range nodes {
		np.seen[node.ID] = true
		np.nodesMap.Read(func(nodes map[string]swarm.Node) {
			n, ok := nodes[node.ID]
			if !ok {
				// Add
				np.eventQueue <- event.New(types.EventTypeNodeCreated, node)
			} else if n.Version.Index < node.Version.Index {
				// Update
				np.eventQueue <- event.New(types.EventTypeNodeUpdated, node)
			}
		})
	}

	np.nodesMap.Read(func(nodes map[string]swarm.Node) {
		for id, node := range nodes {
			if !np.seen[id] {
				// Delete
				np.eventQueue <- event.New(types.EventTypeNodeDeleted, node)
			}
			delete(np.seen, id)
		}
	})
}
