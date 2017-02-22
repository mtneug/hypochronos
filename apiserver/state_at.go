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
	"time"

	"golang.org/x/net/context"

	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/servicehandler"
	"github.com/mtneug/pkg/startstopper"
)

// StatesAt implements the HypochronosServer interface.
func (s *Server) StatesAt(ctx context.Context, req *api.StatesAtRequest) (*api.StatesAtResponse, error) {
	s.ServiceHandlerMap.Lock()
	defer s.ServiceHandlerMap.Unlock()

	t := time.Unix(req.Time, 0)

	resp := api.StatesAtResponse{
		States: make([]*api.State, 0, s.ServiceHandlerMap.Len()),
	}

	s.ServiceHandlerMap.ForEach(func(key string, ss startstopper.StartStopper) {
		sh := ss.(*servicehandler.ServiceHandler)
		sh.TimetableMutex.RLock()
		defer sh.TimetableMutex.RUnlock()

		state, until := sh.Timetable.StateAt(req.NodeID, t)

		s := api.State{
			Value: state,
			Until: until.Unix(),
			Node: &api.Node{
				ID: req.NodeID,
			},
			Service: &api.Service{
				ID:   sh.ServiceID,
				Name: sh.ServiceName,
			},
		}

		resp.States = append(resp.States, &s)
	})

	return &resp, nil
}
