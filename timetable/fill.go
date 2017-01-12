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

package timetable

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Fill error values
var (
	// ErrUnknownType indicates that filling failed because the timetable type is
	// unknown.
	ErrUnknownType = errors.New("timetable: unknown type")
)

// Fill a timetable.
func Fill(tt *Timetable) error {
	switch tt.Spec.Type {
	case TypeJSON:
		return jsonFiller(tt)
	}

	return ErrUnknownType
}

// API holds values of every API object.
type API struct {
	APIVersion string `json:"apiVersion"`
}

// Metadata of API objects.
type Metadata struct {
	CreatedAt time.Time `json:"createdAt"`
}

// JSONFillerSpec API object for JSONFillerResponse.
type JSONFillerSpec struct {
	Timetable map[string]map[time.Time]string `json:"createdAt"`
}

// JSONFillerResponse API object.
type JSONFillerResponse struct {
	API
	Metadata Metadata       `json:"metadata"`
	Spec     JSONFillerSpec `json:"spec"`
}

func jsonFiller(tt *Timetable) error {
	// Load
	resp, err := http.Get(tt.Spec.JSONSpec.URL)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Decode
	var rawTt JSONFillerResponse
	err = json.NewDecoder(resp.Body).Decode(&rawTt)
	if err != nil {
		return err
	}

	if rawTt.APIVersion != "1" {
		return errors.New("API version does not match")
	}

	// Remove old entries
	tt.idSortedEntriesMap = make(map[string][]Entry, len(rawTt.Spec.Timetable))

	// Fill new entries
	var wg sync.WaitGroup
	for _id, _timeStateMap := range rawTt.Spec.Timetable {
		id, timeStateMap := _id, _timeStateMap
		wg.Add(1)

		go func() {
			defer wg.Done()

			// Create entries
			entries := make([]Entry, 0, len(timeStateMap))
			for t, s := range timeStateMap {
				entries = append(entries, Entry{StartsAt: t, State: s})
			}

			// Sort entries
			sort.Sort(byTime(entries))

			// Remove unneeded entries
			for i := 1; i < len(entries); i++ {
				if entries[i-1].State == entries[i].State {
					entries = append(entries[:i], entries[i+1:]...)
					i--
				}
			}

			// Store entries
			tt.idSortedEntriesMap[id] = entries
		}()
	}
	wg.Wait()

	tt.FilledAt = time.Now().UTC()
	return nil
}
