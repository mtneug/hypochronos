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

package apiserver

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/event"
)

// Events implements the HypochronosServer interface.
func (s *Server) Events(req *api.EventsRequest, stream api.Hypochronos_EventsServer) error {
	log.Info("Client subscribed to events")
	defer log.Info("Client unsubscribed from events")

	eventQueue, unsub := s.EventManager.Sub()
	defer unsub()

	filters := api.Filters{}
	if f := req.GetFilters(); f != nil {
		filters = *f
	}

	for {
		select {
		case e := <-eventQueue:
			if out := event.Filter(filters, e); !out {
				err := stream.Send(&api.EventsResponse{Event: &e})
				if err != nil {
					return err
				}
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}
