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

package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/hypochronos/timetable"
)

func labelState(srvName string) string {
	return fmt.Sprintf(hypochronosStateLabelPattern, srvName)
}

// NodeGetServiceStateLabel extracts the value of the status label on given node
// for given service.
func NodeGetServiceStateLabel(node *swarm.Node, srvName string) string {
	return node.Spec.Labels[labelState(srvName)]
}

// NodeSetServiceStateLabel writes a value to the status label on given node for
// given service.
func NodeSetServiceStateLabel(ctx context.Context, node *swarm.Node, srvName string, state timetable.State) error {
	if node.Spec.Labels == nil {
		node.Spec.Labels = make(map[string]string)
	}
	node.Spec.Labels[labelState(srvName)] = string(state)

	err := StdClient.NodeUpdate(ctx, node.ID, node.Version, node.Spec)
	if err != nil {
		return err
	}

	n, _, err := StdClient.NodeInspectWithRaw(ctx, node.ID)
	if err != nil {
		return err
	}

	*node = n
	return nil
}

// NodeDeleteServiceStateLabel removes the status label on given node for given
// service.
func NodeDeleteServiceStateLabel(ctx context.Context, node *swarm.Node, srvName string) error {
	if node.Spec.Labels == nil {
		return nil
	}

	delete(node.Spec.Labels, labelState(srvName))
	err := StdClient.NodeUpdate(ctx, node.ID, node.Version, node.Spec)
	if err != nil {
		return err
	}

	n, _, err := StdClient.NodeInspectWithRaw(ctx, node.ID)
	if err != nil {
		return err
	}

	*node = n
	return nil
}
