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

	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/pkg/startstopper"
)

// Controller monitors Docker Swarm state.
type Controller struct {
	startstopper.StartStopper
}

// New creates a new controller.
func New(p time.Duration, m startstopper.Map) *Controller {
	ctrl := &Controller{}
	ctrl.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(ctrl.run))
	return ctrl
}

func (c *Controller) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Controller loop started")
	defer log.Debug("Controller loop stopped")

	<-stopChan

	return nil
}
