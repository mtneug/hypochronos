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
	"sync"

	"docker.io/go-docker/api/types"
)

// ParallelForEachContainer executes op parallel for each given container.
func ParallelForEachContainer(ctx context.Context, containers []types.Container,
	op func(ctx context.Context, container types.Container) error) (err <-chan error) {

	var wg sync.WaitGroup
	errChan := make(chan error)

	for _, c := range containers {
		container := c

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := op(ctx, container)
			if err != nil {
				errChan <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}
