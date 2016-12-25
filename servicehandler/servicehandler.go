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
	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/pkg/event"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/pkg/startstopper"
)

// ServiceHandler manages the schedulability of nodes concerning a specific
// service.
type ServiceHandler struct {
	startstopper.StartStopper

	Service       swarm.Service
	TimetableSpec types.TimetableSpec
	NodesMap      *store.NodesMap

	Period      time.Duration
	Policy      types.Policy
	MinDuration time.Duration

	EventManager event.Manager
}

// New creates a new ServiceHandler.
func New(srv swarm.Service, tt types.TimetableSpec, em event.Manager, nm *store.NodesMap) *ServiceHandler {
	sh := &ServiceHandler{
		Service:       srv,
		TimetableSpec: tt,
		NodesMap:      nm,
		EventManager:  em,
	}
	sh.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(sh.run))

	return sh
}

func (sh *ServiceHandler) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Service handler started")
	defer log.Debug("Service handler stopped")

	select {
	case <-stopChan:
	case <-ctx.Done():
	}

	return nil
}
