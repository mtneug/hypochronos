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

package controller

import (
	"context"
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/label"
	"github.com/mtneug/hypochronos/pkg/event"
	"github.com/mtneug/hypochronos/servicehandler"
	"github.com/mtneug/hypochronos/store"
	"github.com/mtneug/pkg/startstopper"
)

type eventLoop struct {
	startstopper.StartStopper

	eventManager      event.Manager
	nodesMap          *store.NodesMap
	serviceHandlerMap startstopper.Map
}

func newEventLoop(em event.Manager, nm *store.NodesMap, shm startstopper.Map) *eventLoop {
	el := &eventLoop{
		eventManager:      em,
		nodesMap:          nm,
		serviceHandlerMap: shm,
	}
	el.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(el.run))
	return el
}

func (el *eventLoop) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Main event loop started")
	defer log.Debug("Main event loop stopped")

	eventQueue, unsub := el.eventManager.Sub()
	defer unsub()

	for {
		select {
		case e := <-eventQueue:
			el.handleEvent(ctx, e)
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (el *eventLoop) handleEvent(ctx context.Context, e event.Event) {
	log.Debugf("Received %s event", e.Type)

	var (
		ok   bool
		srv  swarm.Service
		node swarm.Node
	)

	if types.IsServiceEvent(e) {
		srv, ok = e.Object.(swarm.Service)
		if !ok {
			log.
				WithError(errors.New("controller: type assertion failed")).
				Error("Failed to get service")
			return
		}
	} else if types.IsNodeEvent(e) {
		node, ok = e.Object.(swarm.Node)
		if !ok {
			log.
				WithError(errors.New("controller: type assertion failed")).
				Error("Failed to get node")
			return
		}
	}

	switch e.Type {
	case types.EventTypeNodeCreated:
		el.nodesMap.Write(func(nodes map[string]swarm.Node) {
			// another Goroutine might have already added the node
			n, present := nodes[node.ID]
			if present && node.Version.Index < n.Version.Index {
				return
			}
			nodes[node.ID] = node
		})
		log.Info("Node added")
	case types.EventTypeNodeUpdated:
		el.nodesMap.Write(func(nodes map[string]swarm.Node) {
			// another Goroutine might have already updated the node
			n, present := nodes[node.ID]
			if present && node.Version.Index < n.Version.Index {
				return
			}
			nodes[node.ID] = node
		})
		log.Info("Node updated")
	case types.EventTypeNodeDeleted:
		el.nodesMap.Write(func(nodes map[string]swarm.Node) {
			delete(nodes, node.ID)
		})
		log.Info("Node deleted")
	case types.EventTypeServiceCreated:
		changed, err := el.addServiceHandler(ctx, srv)
		if err != nil {
			log.WithError(err).Error("Could not add service handler")
		} else if changed {
			log.Info("Service handler added")
		}
	case types.EventTypeServiceUpdated:
		changed, err := el.updateServiceHandler(ctx, srv)
		if err != nil {
			log.WithError(err).Error("Could not update service handler")
		} else if changed {
			log.Info("Service handler updated")
		}
	case types.EventTypeServiceDeleted:
		changed, err := el.deleteServiceHandler(ctx, srv)
		if err != nil {
			log.WithError(err).Error("Could not delete service handler")
		} else if changed {
			log.Info("Service handler deleted")
		}
	}
}

func (el *eventLoop) addServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
	sh, err := el.constructServiceHandler(srv)
	if err != nil {
		return false, err
	}

	return el.serviceHandlerMap.AddAndStart(ctx, srv.ID, sh)
}

func (el *eventLoop) updateServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
	sh, err := el.constructServiceHandler(srv)
	if err != nil {
		return false, err
	}

	return el.serviceHandlerMap.UpdateAndRestart(ctx, srv.ID, sh)
}

func (el *eventLoop) deleteServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
	return el.serviceHandlerMap.DeleteAndStop(ctx, srv.ID)
}

func (el *eventLoop) constructServiceHandler(srv swarm.Service) (*servicehandler.ServiceHandler, error) {
	labels := srv.Spec.Labels

	tt := types.TimetableSpec{}
	err := label.ParseTimetableSpec(&tt, labels)
	if err != nil {
		return nil, err
	}

	sh := servicehandler.New(srv, tt, el.eventManager, el.nodesMap)
	err = label.ParseServiceHandler(sh, labels)
	if err != nil {
		return nil, err
	}

	return sh, nil
}
