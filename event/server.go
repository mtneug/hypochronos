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
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/pkg/startstopper"
)

// Server implements a gRPC events server.
type Server struct {
	startstopper.StartStopper

	Addr         string
	EventManager Manager
}

// NewServer creates a new gRPC events server.
func NewServer(addr string, eventManager Manager) *Server {
	s := &Server{
		Addr:         addr,
		EventManager: eventManager,
	}
	s.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(s.run))

	return s
}

func (s *Server) run(ctx context.Context, stopChan <-chan struct{}) (err error) {
	log.Debug("Event server started")
	defer log.Debug("Event server stopped")

	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer func() { _ = lis.Close() }()

	srv := grpc.NewServer()
	api.RegisterEventServiceServer(srv, s)
	reflection.Register(srv)

	errChan := make(chan error)
	go func() { errChan <- srv.Serve(lis) }()

	select {
	case err = <-errChan:
	case <-stopChan:
	case <-ctx.Done():
		err = ctx.Err()
	}

	srv.Stop()

	return
}

// Sub implements the EventServiceServer interface.
func (s *Server) Sub(req *api.SubRequest, stream api.EventService_SubServer) error {
	log.Info("Client subscribed")
	defer log.Info("Client unsubscribed")

	eventQueue, unsub := s.EventManager.Sub()
	defer unsub()

	filters := api.Filters{}
	if f := req.GetFilters(); f != nil {
		filters = *f
	}

	for {
		select {
		case e := <-eventQueue:
			if out := Filter(filters, e); !out {
				err := stream.Send(&api.SubResponse{Event: &e})
				if err != nil {
					return err
				}
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}
