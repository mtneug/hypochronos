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

package store

import (
	"sync"

	"github.com/docker/docker/api/types/swarm"
)

// ServicesMap is a locked string to service map.
type ServicesMap struct {
	mutex    sync.RWMutex
	services map[string]swarm.Service
}

// NewServicesMap creates a new ServicesMap.
func NewServicesMap() *ServicesMap {
	m := &ServicesMap{
		services: make(map[string]swarm.Service),
	}
	return m
}

// Read transaction.
func (m *ServicesMap) Read(f func(map[string]swarm.Service)) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	f(m.services)
}

// Write transaction.
func (m *ServicesMap) Write(f func(map[string]swarm.Service)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	f(m.services)
}
