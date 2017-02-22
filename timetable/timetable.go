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
	DefaultState string
}

// JSONSpec specifies a hypochronos JSON timetable.
type JSONSpec struct {
	// URL of the hypochronos JSON timetable.
	URL string
}

// Entry of a timetable.
type Entry struct {
	StartsAt time.Time
	State    string
}

// SortedEntries of a timetable.
type SortedEntries []Entry

// Since filters for entries starting after given time.
func (e SortedEntries) Since(t time.Time) SortedEntries {
	i := binarySearch(e, t, 0, len(e)-1)

	if i == -1 {
		return e
	}

	if e[i].StartsAt.Equal(t) {
		return e[i:]
	}

	if i+1 < len(e) {
		return e[i+1:]
	}

	return make([]Entry, 0, 0)
}

// Until filters for entries starting before given time.
func (e SortedEntries) Until(t time.Time) SortedEntries {
	i := binarySearch(e, t, 0, len(e)-1)

	if i == -1 {
		return make([]Entry, 0, 0)
	}

	return e[:i+1]
}

type byTime []Entry

func (e byTime) Len() int           { return len(e) }
func (e byTime) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e byTime) Less(i, j int) bool { return e[i].StartsAt.Before(e[j].StartsAt) }

// Timetable for resources.
type Timetable struct {
	ID       string
	Spec     Spec
	FilledAt time.Time

	idSortedEntriesMap map[string][]Entry
}

var (
	// MaxTime that can be un/marshaled.
	MaxTime = time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC)
)

// New creates a new timetable
func New(spec Spec) Timetable {
	tt := Timetable{
		ID:   ulid.New().String(),
		Spec: spec,
	}
	return tt
}

// Entries returns a copy of the internal entries for the resource with given
// id. The entries are sorted by time.
func (tt *Timetable) Entries(id string) SortedEntries {
	eOrig, ok := tt.idSortedEntriesMap[id]
	if !ok {
		return make([]Entry, 0, 0)
	}

	e := make([]Entry, len(eOrig))
	copy(e, eOrig)
	return e
}

// StateAt of the resource at given time.
func (tt *Timetable) StateAt(id string, t time.Time) (state string, until time.Time) {
	entries, ok := tt.idSortedEntriesMap[id]
	if !ok {
		return tt.Spec.DefaultState, MaxTime
	}

	l := len(entries)
	i := binarySearch(entries, t, 0, l-1)

	if i == -1 {
		state = tt.Spec.DefaultState
	} else {
		state = entries[i].State
	}

	if i+1 < l {
		until = entries[i+1].StartsAt
	} else {
		until = MaxTime
	}

	return
}

func binarySearch(entries []Entry, t time.Time, sIdx, eIdx int) int {
	if eIdx < sIdx {
		// before first entry
		return -1
	}

	mIdx := (sIdx + eIdx) / 2

	if entries[mIdx].StartsAt.After(t) {
		// left side
		return binarySearch(entries, t, sIdx, mIdx-1)
	}

	if mIdx == eIdx || entries[mIdx+1].StartsAt.After(t) {
		// found
		return mIdx
	}

	// right side
	return binarySearch(entries, t, mIdx+1, eIdx)
}
