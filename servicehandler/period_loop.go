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
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/timetable"
)

func (sh *ServiceHandler) runPeriodLoop(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Timetable filler started")
	defer log.Debug("Timetable filler stopped")

	var periodCtx context.Context
	var cancelPeriodCtx context.CancelFunc

	tick := func() {
		// Cancel last period
		if cancelPeriodCtx != nil {
			cancelPeriodCtx()
			log.Debug("Period ended")
		}

		// Filling Timetable
		sh.timetableMutex.Lock()
		log.Debug("Filling timetable")
		err := timetable.Fill(&sh.Timetable)
		if err != nil {
			log.WithError(err).Error("Failed to fill timetable")
			return
		}
		sh.timetableMutex.Unlock()

		// Start a new period
		log.Debug("Period started")
		periodCtx, cancelPeriodCtx = sh.WithPeriod(ctx)

		go func() {
			// Applying timetable
			log.Debug("Applying timetable")
			sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
				errChan := forEachKeyNodePair(periodCtx, nodes, func(ctx context.Context, key string, node swarm.Node) error {
					err := sh.applyTimetable(ctx, &node)
					if err != nil {
						return err
					}

					nodes[key] = node
					return nil
				})

				for err := range errChan {
					log.WithError(err).Error("Applying timetable failed")
				}
			})

			// Schedule state changes
			log.Debug("Schedule state changes")
			// TODO: implement
		}()
	}

	for {
		tick()

		select {
		case <-time.After(sh.Period):
			tick()
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
