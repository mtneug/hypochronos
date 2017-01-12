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
	"context"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/hypochronos/api"
	"github.com/mtneug/hypochronos/event"
	"github.com/mtneug/pkg/startstopper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server implements the hypochronos API.
type Server struct {
	startstopper.StartStopper

	Addr string

	ServiceHandlerMap startstopper.Map
	EventManager      event.Manager
}

// New creates a new hypochronos API server.
func New(addr string, serviceHandlerMap startstopper.Map, eventManager event.Manager) *Server {
	s := &Server{
		Addr:              addr,
		ServiceHandlerMap: serviceHandlerMap,
		EventManager:      eventManager,
	}
	s.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(s.run))

	return s
}

func (s *Server) run(ctx context.Context, stopChan <-chan struct{}) (err error) {
	log.Debug("API server started")
	defer log.Debug("API server stopped")

	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer func() { _ = lis.Close() }()

	srv := grpc.NewServer()
	api.RegisterHypochronosServer(srv, s)
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
