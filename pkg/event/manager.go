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

package event

import (
	"context"
	"sync"

	"github.com/mtneug/pkg/startstopper"
	"github.com/mtneug/pkg/ulid"
)

// Manager of event handlers.
type Manager interface {
	startstopper.StartStopper

	// Register an event handler. Calling the returned function will unregister
	// the event handler.
	Register(Handler) (unregister func())

	// Publish returns a write-only channel for publishing events.
	Publish() chan<- Event
}

// Handler for events.
type Handler interface {
	Handle(Event)
}

// ConcurrentManager manages events concurrently.
type ConcurrentManager struct {
	startstopper.StartStopper

	mutex   sync.Mutex
	handler map[string]Handler
	queue   chan Event
}

// NewConcurrentManager creates a new ConcurrentManager with given queue size.
func NewConcurrentManager(queueSize int) *ConcurrentManager {
	em := &ConcurrentManager{
		handler: make(map[string]Handler),
		queue:   make(chan Event, queueSize),
	}
	em.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(em.run))
	return em
}

// Register an event handler.
func (em *ConcurrentManager) Register(handler Handler) (unregister func()) {
	id := ulid.New().String()

	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.handler[id] = handler

	return func() {
		em.mutex.Lock()
		defer em.mutex.Unlock()
		delete(em.handler, id)
	}
}

// Publish to the event channel
func (em *ConcurrentManager) Publish() chan<- Event {
	return em.queue
}

func (em *ConcurrentManager) run(ctx context.Context, stopChan <-chan struct{}) error {
	for {
		select {
		case e := <-em.queue:
			for _, h := range em.handler {
				go h.Handle(e)
			}
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
