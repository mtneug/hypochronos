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

package model

import "github.com/mtneug/hypochronos/pkg/event"

const (
	// EventTypeNodeCreated indicates that a node was created.
	EventTypeNodeCreated event.Type = "node_created"

	// EventTypeNodeUpdated indicates that a node was updated.
	EventTypeNodeUpdated event.Type = "node_updated"

	// EventTypeNodeDeleted indicates that a node was deleted.
	EventTypeNodeDeleted event.Type = "node_deleted"

	// EventTypeServiceCreated indicates that a service was created.
	EventTypeServiceCreated event.Type = "service_created"

	// EventTypeServiceUpdated indicates that a service was updated.
	EventTypeServiceUpdated event.Type = "service_updated"

	// EventTypeServiceDeleted indicates that a service was deleted.
	EventTypeServiceDeleted event.Type = "service_deleted"
)

// TODO: maybe split type to objectType and action

// IsServiceEvent checks if the event is a service event.
func IsServiceEvent(e event.Event) bool {
	return e.Type == EventTypeServiceCreated ||
		e.Type == EventTypeServiceUpdated ||
		e.Type == EventTypeServiceDeleted
}

// IsNodeEvent checks if the event is a node event.
func IsNodeEvent(e event.Event) bool {
	return e.Type == EventTypeNodeCreated ||
		e.Type == EventTypeNodeUpdated ||
		e.Type == EventTypeNodeDeleted
}
