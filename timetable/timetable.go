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
	"time"

	"github.com/mtneug/pkg/ulid"
)

// State of a resource.
type State string

const (
	// StateUndefined indicates that the state of the resource is undefined.
	StateUndefined State = "undefined"
)

// Type represents some category of timetables.
type Type string

const (
	// TypeJSON is a hypochronos JSON timetable.
	TypeJSON Type = "json"
)

// Spec specifies a timetable.
type Spec struct {
	// Type of the timetable.
	Type Type
	// JSONSpec for a hypochronos JSON timetable.
	JSONSpec JSONSpec
	// DefaultState if non is given.
	DefaultState State
}

// JSONSpec specifies a hypochronos JSON timetable.
type JSONSpec struct {
	// URL of the hypochronos JSON timetable.
	URL string
}

// Timetable for resources.
type Timetable struct {
	ID       string
	Spec     Spec
	FilledAt time.Time
}

// New creates a new timetable
func New(spec Spec) Timetable {
	tt := Timetable{
		ID:   ulid.New().String(),
		Spec: spec,
	}
	return tt
}

// State of the resource at given time.
func (tt *Timetable) State(id string, t time.Time) (state State, until time.Time) {
	// TODO: implement
	return tt.Spec.DefaultState, t.Add(24 * time.Hour)
}
