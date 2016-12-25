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
)

func (c *Controller) runEventLoop(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Main event loop started")
	defer log.Debug("Main event loop stopped")

	eventQueue, unsub := c.EventManager.Sub()
	defer unsub()

	for {
		select {
		case e := <-eventQueue:
			log.Debugf("Received %s event", e.Type)

			if types.IsServiceEvent(e) {
				srv, ok := e.Object.(swarm.Service)
				if ok {
					c.handleServiceEvent(ctx, e, srv)
				} else {
					log.
						WithError(errors.New("controller: type assertion failed")).
						Error("Failed to get service")
				}
			} else if types.IsNodeEvent(e) {
				node, ok := e.Object.(swarm.Node)
				if ok {
					c.handleNodeEvent(ctx, e, node)
				} else {
					log.
						WithError(errors.New("controller: type assertion failed")).
						Error("Failed to get node")
				}
			}
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Controller) handleNodeEvent(ctx context.Context, e event.Event, node swarm.Node) {
	switch e.Type {
	case types.EventTypeNodeCreated:
		c.NodesMap.Write(func(nodes map[string]swarm.Node) {
			// another Goroutine might have already added the node
			n, present := nodes[node.ID]
			if present && node.Version.Index < n.Version.Index {
				return
			}
			nodes[node.ID] = node
		})
		log.Info("Node added")

	case types.EventTypeNodeUpdated:
		c.NodesMap.Write(func(nodes map[string]swarm.Node) {
			// another Goroutine might have already updated the node
			n, present := nodes[node.ID]
			if present && node.Version.Index < n.Version.Index {
				return
			}
			nodes[node.ID] = node
		})
		log.Info("Node updated")

	case types.EventTypeNodeDeleted:
		c.NodesMap.Write(func(nodes map[string]swarm.Node) {
			delete(nodes, node.ID)
		})
		log.Info("Node deleted")
	}
}

func (c *Controller) handleServiceEvent(ctx context.Context, e event.Event, srv swarm.Service) {
	switch e.Type {
	case types.EventTypeServiceCreated:
		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			changed, err := c.addServiceHandler(ctx, srv)
			if err != nil {
				log.WithError(err).Error("Could not add service handler")
				return
			} else if changed {
				log.Info("Service handler added")
			}

			services[srv.ID] = srv
		})

	case types.EventTypeServiceUpdated:
		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			oldSrv, ok := services[srv.ID]
			if !ok {
				log.
					WithError(errors.New("controller: could not find old service")).
					Error("Could not update service handler")
				return
			}

			changed, err := c.updateServiceHandler(ctx, oldSrv, srv)
			if err != nil {
				log.WithError(err).Error("Could not update service handler")
				return
			} else if changed {
				log.Info("Service handler updated")
			}

			services[srv.ID] = srv
		})

	case types.EventTypeServiceDeleted:
		c.ServicesMap.Write(func(services map[string]swarm.Service) {
			changed, err := c.deleteServiceHandler(ctx, srv.ID)
			if err != nil {
				log.WithError(err).Error("Could not delete service handler")
				return
			} else if changed {
				log.Info("Service handler deleted")
			}

			delete(services, srv.ID)
		})
	}
}

func (c *Controller) addServiceHandler(ctx context.Context, srv swarm.Service) (bool, error) {
	sh, err := c.constructServiceHandler(srv)
	if err != nil {
		return false, err
	}

	return c.ServiceHandlerMap.AddAndStart(ctx, srv.ID, sh)
}

func (c *Controller) updateServiceHandler(ctx context.Context, srvOld, srvNew swarm.Service) (bool, error) {
	// We want to skip updates if non of the hypochronos labels have changed (for
	// instance if the service scale is changed).
	if label.Equal(srvOld.Spec.Labels, srvNew.Spec.Labels) {
		return false, nil
	}

	sh, err := c.constructServiceHandler(srvNew)
	if err != nil {
		return false, err
	}

	return c.ServiceHandlerMap.UpdateAndRestart(ctx, srvNew.ID, sh)
}

func (c *Controller) deleteServiceHandler(ctx context.Context, serviceID string) (bool, error) {
	return c.ServiceHandlerMap.DeleteAndStop(ctx, serviceID)
}

func (c *Controller) constructServiceHandler(srv swarm.Service) (*servicehandler.ServiceHandler, error) {
	labels := srv.Spec.Labels

	tt := types.TimetableSpec{}
	err := label.ParseTimetableSpec(&tt, labels)
	if err != nil {
		return nil, err
	}

	sh := servicehandler.New(srv.ID, tt, c.EventManager, c.NodesMap)
	err = label.ParseServiceHandler(sh, labels)
	if err != nil {
		return nil, err
	}

	return sh, nil
}
