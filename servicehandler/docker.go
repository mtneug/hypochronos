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

package servicehandler

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/timetable"
)

const (
	dockerSwarmServiceNameLabel = "com.docker.swarm.service.name"
	stateLabelPattern           = "de.mtneug.hypochronos.state.%s"
)

func (sh *ServiceHandler) newDockerEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	args := filters.NewArgs()
	args.Add("type", "container")
	args.Add("event", "create")
	args.Add("label", dockerSwarmServiceNameLabel+"="+sh.ServiceName)
	return docker.C.Events(ctx, dockerTypes.EventsOptions{Filters: args})
}

func (sh *ServiceHandler) applyTimetable(ctx context.Context, nodeID string) {
	sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
		node, present := nodes[nodeID]
		if present {
			sh.timetableMutex.RLock()
			defer sh.timetableMutex.RUnlock()

			// TODO: timetable might be nil
			newState, until := sh.timetable.State(nodeID, time.Now().UTC())
			curState := timetable.State(node.Spec.Labels[labelState(sh.ServiceName)])
			if curState == "" {
				curState = timetable.StateUndefined
			}
			log.Debugf("Current state: %s; New state: %s until %s", curState, newState, until)
			// TODO: use until
			// TODO: remove containers

			if newState != curState {
				newNode, err := setStateLabelAndInspect(ctx, sh.ServiceName, &node, newState)
				if err != nil {
					log.WithError(err).Error("Failed to apply timetable")
					return
				}
				nodes[nodeID] = *newNode
			}
		}
	})
}

func labelState(serviceName string) string {
	return fmt.Sprintf(stateLabelPattern, serviceName)
}

func setStateLabelAndInspect(ctx context.Context, serviceName string, node *swarm.Node,
	state timetable.State) (*swarm.Node, error) {

	if node.Spec.Labels == nil {
		node.Spec.Labels = make(map[string]string)
	}
	node.Spec.Labels[labelState(serviceName)] = string(state)

	err := docker.C.NodeUpdate(ctx, node.ID, node.Version, node.Spec)
	if err != nil {
		return nil, err
	}

	n, _, err := docker.C.NodeInspectWithRaw(ctx, node.ID)
	return &n, err
}

func deleteStateLabel(ctx context.Context, serviceName string, node *swarm.Node) error {
	if node.Spec.Labels == nil {
		return nil
	}
	delete(node.Spec.Labels, labelState(serviceName))
	return docker.C.NodeUpdate(ctx, node.ID, node.Version, node.Spec)
}
