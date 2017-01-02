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
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/pkg/ulid"
)

// Type represents some category of events.
type Type string

// New creates a new event.
func New(a api.EventAction, actor interface{}) api.Event {
	e := api.Event{
		ID:     ulid.New().String(),
		Action: a,
	}

	switch a := actor.(type) {
	case api.Node:
		e.Actor = &api.Event_Node{Node: &a}
		e.ActorID = a.ID
		e.ActorType = api.EventActorType_node
	case api.Service:
		e.Actor = &api.Event_Service{Service: &a}
		e.ActorID = a.ID
		e.ActorType = api.EventActorType_service
	case api.State:
		e.Actor = &api.Event_State{State: a}
		e.ActorType = api.EventActorType_state
	}

	return e
}
