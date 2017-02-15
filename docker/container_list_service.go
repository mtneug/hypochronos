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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// ContainerListService lists all container of a specific service.
func ContainerListService(ctx context.Context, serviceID string) ([]types.Container, error) {
	args := filters.NewArgs()
	args.Add("label", DockerSwarmServiceIDLabel+"="+serviceID)

	return StdClient.ContainerList(ctx, types.ContainerListOptions{Filters: args})
}
