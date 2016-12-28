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
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/pkg/event"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/hypochronos/timetable"
	"github.com/mtneug/pkg/startstopper"
)

// ServiceHandler manages the schedulability of nodes concerning a specific
// service.
type ServiceHandler struct {
	startstopper.StartStopper

	ServiceName string
	NodesMap    *store.NodesMap

	Period      time.Duration
	MinDuration time.Duration

	Timetable      timetable.Timetable
	timetableMutex sync.RWMutex

	EventManager event.Manager
	eventLoop    startstopper.StartStopper
	periodLoop   startstopper.StartStopper
}

// New creates a new ServiceHandler.
func New(serviceName string, tt timetable.Timetable, em event.Manager, nm *store.NodesMap) *ServiceHandler {
	sh := &ServiceHandler{
		ServiceName:  serviceName,
		Timetable:    tt,
		NodesMap:     nm,
		EventManager: em,
	}
	sh.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(sh.run))
	sh.eventLoop = startstopper.NewGo(startstopper.RunnerFunc(sh.runEventLoop))
	sh.periodLoop = startstopper.NewGo(startstopper.RunnerFunc(sh.runPeriodLoop))

	return sh
}

func (sh *ServiceHandler) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service handler started")
	defer log.Debug("Service handler stopped")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	group := startstopper.NewGroup([]startstopper.StartStopper{
		sh.eventLoop,
		sh.periodLoop,
	})

	_ = group.Start(ctx)

	select {
	case <-stopChan:
	case <-ctx.Done():
	}

	_ = group.Stop(ctx)
	err := group.Err(ctx)

	log.Debug("Delete service state labels")
	sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
		errChan := forEachKeyNodePair(ctx, nodes, func(ctx context.Context, key string, node swarm.Node) error {
			err = docker.NodeDeleteServiceStateLabel(ctx, &node, sh.ServiceName)
			if err != nil {
				return err
			}
			nodes[key] = node
			return nil
		})

		for err := range errChan {
			log.WithError(err).Warn("Deletion of service state label failed")
		}
	})

	return err
}
