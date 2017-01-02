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
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/mtneug/hypochronos/model"
)

// ContainerWriteTTL writes a TTL file to given container.
func ContainerWriteTTL(ctx context.Context, containerID string, until time.Time) error {
	resp := model.TTLResponse{
		API:      model.API{APIVersion: "1"},
		Metadata: model.Metadata{CreatedAt: time.Now().UTC()},
		Data:     model.TTLData{TTL: until},
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name: model.TTLFilePath,
		Mode: 0444,
		Size: int64(len(jsonResp)),
	}
	err = tw.WriteHeader(hdr)
	if err != nil {
		return err
	}

	_, err = tw.Write(jsonResp)
	if err != nil {
		return err
	}

	err = tw.Close()
	if err != nil {
		return err
	}

	opts := types.CopyToContainerOptions{AllowOverwriteDirWithFile: true}
	err = StdClient.CopyToContainer(ctx, containerID, "/", buf, opts)
	if err != nil {
		return err
	}

	return nil
}
