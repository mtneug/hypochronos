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

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mtneug/hypochronos/docker"
	"github.com/mtneug/hypochronos/model"
	"github.com/mtneug/hypochronos/timetable"
)

var nodeID string

func init() {
	info, err := docker.StdClient.Info(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	nodeID = info.Swarm.NodeID
}

func main() {
	http.HandleFunc("/tt.json", ttHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ttHandler(w http.ResponseWriter, r *http.Request) {
	var (
		now  = time.Now().UTC()
		resp = timetable.JSONFillerResponse{
			API:      timetable.API{APIVersion: "1"},
			Metadata: timetable.Metadata{CreatedAt: now},
			Spec: timetable.JSONFillerSpec{
				Timetable: map[string]map[time.Time]timetable.State{
					nodeID: {
						now.Add(-15 * time.Second): model.StateActivated,
						now.Add(+15 * time.Second): model.StateDeactivated,
						now.Add(+45 * time.Second): model.StateActivated,
						now.Add(+75 * time.Second): model.StateDeactivated,
					},
				},
			},
		}
	)

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}

	_, _ = w.Write(jsonResp)
	log.Print("GET /tt.json")
}
