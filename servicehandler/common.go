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
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/model"
	"github.com/mtneug/hypochronos/timetable"
)

// WithPeriod creates a new context with the deadline set to the end of the
// current period.
func (sh *ServiceHandler) WithPeriod(ctx context.Context) (context.Context, context.CancelFunc) {
	sh.timetableMutex.RLock()
	defer sh.timetableMutex.RUnlock()
	return context.WithDeadline(ctx, sh.Timetable.FilledAt.Add(sh.Period))
}

func forEachKeyNodePair(ctx context.Context, nm map[string]swarm.Node,
	op func(ctx context.Context, key string, node swarm.Node) error) (err <-chan error) {

	var wg sync.WaitGroup
	errChan := make(chan error)

	for k, n := range nm {
		key, node := k, n

		wg.Add(1)
		go func() {
			err := op(ctx, key, node)
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}

func forEachContainer(ctx context.Context, containers []types.Container,
	op func(ctx context.Context, container types.Container) error) (err <-chan error) {

	var wg sync.WaitGroup
	errChan := make(chan error)

	for _, c := range containers {
		container := c

		wg.Add(1)
		go func() {
			err := op(ctx, container)
			if err != nil {
				errChan <- err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}

func (sh *ServiceHandler) applyTimetable(ctx context.Context, node *swarm.Node) error {
	sh.timetableMutex.RLock()
	defer sh.timetableMutex.RUnlock()

	// Get new state
	newState, until := sh.Timetable.State(node.ID, time.Now().UTC())
	curState := timetable.State(docker.NodeGetServiceStateLabel(node, sh.ServiceName))

	if curState == "" {
		curState = timetable.StateUndefined
	}
	log.Debugf("Current state: %s; New state: %s until %s", curState, newState, until)

	// If different: change it
	if newState != curState {
		log.Debug("Setting service state label")
		err := docker.NodeSetServiceStateLabel(ctx, node, sh.ServiceName, newState)
		if err != nil {
			return err
		}

		// Apply appropriate actions to running containers
		log.Debug("Retrieving containers from node")
		containers, err := docker.ContainerListNode(ctx, node.ID)
		if err != nil {
			return err
		}

		if newState == model.StateActivated {
			log.Debug("Writing container TTL")

			errChan := forEachContainer(ctx, containers, func(ctx context.Context, container types.Container) error {
				err2 := docker.ContainerWriteTTL(ctx, container.ID, until)
				if err2 != nil {
					return err2
				}
				return nil
			})

			for err := range errChan {
				log.WithError(err).Warn("Writing container TTL failed")
			}
		} else {
			log.Debug("Stopping and removing running containers")

			errChan := forEachContainer(ctx, containers, func(ctx context.Context, container types.Container) error {
				err2 := docker.ContainerRemoveGracefully(ctx, container.ID)
				if err2 != nil {
					return err2
				}
				return nil
			})

			for err := range errChan {
				log.WithError(err).Warn("Stopping and removing running containers failed")
			}
		}
	}

	return nil
}
