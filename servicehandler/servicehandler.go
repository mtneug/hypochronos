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
	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/pkg/event"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/hypochronos/timetable"
	"github.com/mtneug/pkg/startstopper"
)

// ServiceHandler manages the schedulability of nodes concerning a specific
// service.
type ServiceHandler struct {
	startstopper.StartStopper

	ServiceID string
	NodesMap  *store.NodesMap

	timetable      *timetable.Timetable
	timetableMutex sync.RWMutex
	TimetableSpec  types.TimetableSpec

	Period      time.Duration
	Policy      types.Policy
	MinDuration time.Duration

	EventManager event.Manager
	eventLoop    startstopper.StartStopper
}

// New creates a new ServiceHandler.
func New(serviceID string, tt types.TimetableSpec, em event.Manager, nm *store.NodesMap) *ServiceHandler {
	sh := &ServiceHandler{
		ServiceID:     serviceID,
		TimetableSpec: tt,
		NodesMap:      nm,
		EventManager:  em,

		// TODO: implement
		timetable: &timetable.Timetable{
			DefaultState: timetable.StateUndefined,
		},
	}
	sh.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(sh.run))
	sh.eventLoop = startstopper.NewGo(startstopper.RunnerFunc(sh.runEventLoop))

	return sh
}

func (sh *ServiceHandler) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service handler started")
	defer log.Debug("Service handler stopped")

	group := startstopper.NewGroup([]startstopper.StartStopper{
		sh.eventLoop,
	})

	_ = group.Start(ctx)

	select {
	case <-stopChan:
	case <-ctx.Done():
	}

	_ = group.Stop(ctx)
	err := group.Err(ctx)

	log.Debug("Delete node state labels")
	sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
		for _, node := range nodes {
			err = deleteStateLabel(ctx, sh.ServiceID, &node)
			if err != nil {
				log.WithError(err).Warn("Deletion of node state label failed")
			}
		}
	})

	return err
}
