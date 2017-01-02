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

	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/pkg/startstopper"
	"github.com/mtneug/pkg/ulid"
)

// Manager of event handlers.
type Manager interface {
	startstopper.StartStopper

	// Sub returns a read-only channel for subscribing to events. Calling the
	// returned function will unsubscribe this channel.
	Sub() (queue <-chan api.Event, unsub func())

	// Pub returns a write-only channel for publishing events.
	Pub() (queue chan<- api.Event)
}

// ConcurrentManager manages events concurrently.
type ConcurrentManager struct {
	startstopper.StartStopper

	mutex sync.Mutex
	subs  map[string]chan api.Event
	queue chan api.Event
}

// NewConcurrentManager creates a new ConcurrentManager with given queue size.
func NewConcurrentManager(queueSize int) *ConcurrentManager {
	em := &ConcurrentManager{
		subs:  make(map[string]chan api.Event),
		queue: make(chan api.Event, queueSize),
	}
	em.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(em.run))
	return em
}

// Sub to the event channel.
func (em *ConcurrentManager) Sub() (<-chan api.Event, func()) {
	id := ulid.New().String()
	sub := make(chan api.Event)
	unsub := func() {
		em.mutex.Lock()
		defer em.mutex.Unlock()
		delete(em.subs, id)
		close(sub)
	}

	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.subs[id] = sub

	return sub, unsub
}

// Pub to the event channel.
func (em *ConcurrentManager) Pub() chan<- api.Event {
	return em.queue
}

func (em *ConcurrentManager) run(ctx context.Context, stopChan <-chan struct{}) error {
	for {
		select {
		case e := <-em.queue:
			em.mutex.Lock()
			for _, s := range em.subs {
				sub := s
				go func() { sub <- e }()
			}
			em.mutex.Unlock()
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
