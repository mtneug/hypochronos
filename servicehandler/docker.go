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
	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/timetable"
)

const (
	labelStatePattern = "de.mtneug.hypochronos.state.%s"
)

func (sh *ServiceHandler) applyTimetable(ctx context.Context, nodeID string) {
	log.Debugf("Applying timetable")
	sh.NodesMap.Write(func(nodes map[string]swarm.Node) {
		sh.timetableMutex.RLock()
		defer sh.timetableMutex.RUnlock()

		node, present := nodes[nodeID]
		if present {
			newState := sh.timetable.State(nodeID, time.Now().UTC())
			curState := timetable.State(node.Spec.Labels[labelState(sh.ServiceID)])
			if curState == "" {
				curState = timetable.StateUndefined
			}
			log.Debugf("Current state: %s; New state: %s", curState, newState)

			if newState != curState {
				newNode, err := setStateLabelAndInspect(ctx, sh.ServiceID, &node, newState)
				if err != nil {
					log.WithError(err).Error("Failed to apply timetable")
					return
				}
				nodes[nodeID] = *newNode
			}
		}
	})
	log.Debugf("Applied timetable")
}

func setStateLabelAndInspect(ctx context.Context, serviceID string, node *swarm.Node, state timetable.State) (*swarm.Node, error) {
	if node.Spec.Labels == nil {
		node.Spec.Labels = make(map[string]string)
	}
	node.Spec.Labels[labelState(serviceID)] = string(state)

	err := docker.C.NodeUpdate(ctx, node.ID, node.Version, node.Spec)
	if err != nil {
		return nil, err
	}

	n, _, err := docker.C.NodeInspectWithRaw(ctx, node.ID)
	return &n, err
}

func labelState(serviceID string) string {
	return fmt.Sprintf(labelStatePattern, serviceID)
}

func deleteStateLabel(ctx context.Context, serviceID string, node *swarm.Node) error {
	if node.Spec.Labels == nil {
		return nil
	}
	delete(node.Spec.Labels, labelState(serviceID))
	return docker.C.NodeUpdate(ctx, node.ID, node.Version, node.Spec)
}
