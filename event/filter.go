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

import "github.com/mtneug/hypochronos/api"

// Filter given event.
func Filter(filters api.Filters, event api.Event) (out bool) {
	args := filters.Args

	// by action
	a, ok := args[api.FilterKey_Action.String()]
	if ok && a != event.Action.String() {
		return true
	}

	// by actor type
	t, ok := args[api.FilterKey_ActorType.String()]
	if ok && t != event.ActorType.String() {
		return true
	}

	// by actor ID
	id, ok := args[api.FilterKey_ActorID.String()]
	if ok && id != event.ActorID {
		return true
	}

	// by node ID
	nodeID, ok := args[api.FilterKey_NodeID.String()]
	if node := event.GetNode(); ok && node != nil {
		if nodeID != node.ID {
			return true
		}
	} else if state := event.GetState(); ok && state != nil {
		if nodeID != state.Node.ID {
			return true
		}
	} else if ok {
		return true
	}

	// by service ID
	serviceID, ok := args[api.FilterKey_ServiceID.String()]
	if service := event.GetService(); ok && service != nil {
		if serviceID != service.ID {
			return true
		}
	} else if state := event.GetState(); ok && state != nil {
		if serviceID != state.Service.ID {
			return true
		}
	} else if ok {
		return true
	}

	return false
}
