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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSortedEntriesSince(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	examples := []struct {
		entries  SortedEntries
		filtered SortedEntries
		time     time.Time
	}{
		{
			time:     now,
			entries:  []Entry{},
			filtered: []Entry{},
		},
		{
			time: now,
			entries: []Entry{
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
				{StartsAt: now.Add(+70 * time.Second)},
			},
			filtered: []Entry{
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
				{StartsAt: now.Add(+70 * time.Second)},
			},
		},
		{
			time: now,
			entries: []Entry{
				{StartsAt: now},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
			},
			filtered: []Entry{
				{StartsAt: now},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
			},
		},
		{
			time: now,
			entries: []Entry{
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
			},
			filtered: []Entry{
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
			},
		},
		{
			time: now,
			entries: []Entry{
				{StartsAt: now.Add(-60 * time.Second)},
				{StartsAt: now.Add(-50 * time.Second)},
				{StartsAt: now.Add(-40 * time.Second)},
				{StartsAt: now.Add(-30 * time.Second)},
				{StartsAt: now.Add(-20 * time.Second)},
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now},
			},
			filtered: []Entry{
				{StartsAt: now},
			},
		},
		{
			time: now,
			entries: []Entry{
				{StartsAt: now.Add(-60 * time.Second)},
				{StartsAt: now.Add(-50 * time.Second)},
				{StartsAt: now.Add(-40 * time.Second)},
				{StartsAt: now.Add(-30 * time.Second)},
				{StartsAt: now.Add(-20 * time.Second)},
				{StartsAt: now.Add(-10 * time.Second)},
			},
			filtered: []Entry{},
		},
	}

	for _, e := range examples {
		filtered := e.entries.Since(e.time)
		require.Equal(t, filtered, e.filtered)
	}
}

func TestTimetableEntries(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	entries := SortedEntries([]Entry{
		{StartsAt: now.Add(+10 * time.Second)},
		{StartsAt: now.Add(+20 * time.Second)},
		{StartsAt: now.Add(+30 * time.Second)},
		{StartsAt: now.Add(+40 * time.Second)},
		{StartsAt: now.Add(+50 * time.Second)},
		{StartsAt: now.Add(+60 * time.Second)},
		{StartsAt: now.Add(+70 * time.Second)},
	})

	tt := Timetable{
		idSortedEntriesMap: map[string][]Entry{
			"test": entries,
		},
	}

	e := tt.Entries("test")
	require.Equal(t, entries, e)
}

func TestBinarySearch(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	examples := []struct {
		entries []Entry
		time    time.Time
		idx     int
	}{
		{
			entries: []Entry{},
			time:    now,
			idx:     -1,
		},
		{
			entries: []Entry{
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
				{StartsAt: now.Add(+70 * time.Second)},
			},
			time: now,
			idx:  -1,
		},
		{
			entries: []Entry{
				{StartsAt: now},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
			},
			time: now,
			idx:  0,
		},
		{
			entries: []Entry{
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
			},
			time: now,
			idx:  0,
		},
		{
			entries: []Entry{
				{StartsAt: now.Add(-30 * time.Second)},
				{StartsAt: now.Add(-20 * time.Second)},
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
			},
			time: now,
			idx:  3,
		},
		{
			entries: []Entry{
				{StartsAt: now.Add(-30 * time.Second)},
				{StartsAt: now.Add(-20 * time.Second)},
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now.Add(-1 * time.Second)},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
			},
			time: now,
			idx:  3,
		},
		{
			entries: []Entry{
				{StartsAt: now.Add(-30 * time.Second)},
				{StartsAt: now.Add(-20 * time.Second)},
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now.Add(-1 * time.Second)},
				{StartsAt: now.Add(+10 * time.Second)},
				{StartsAt: now.Add(+20 * time.Second)},
				{StartsAt: now.Add(+30 * time.Second)},
				{StartsAt: now.Add(+40 * time.Second)},
				{StartsAt: now.Add(+50 * time.Second)},
				{StartsAt: now.Add(+60 * time.Second)},
			},
			time: now,
			idx:  3,
		},
		{
			entries: []Entry{
				{StartsAt: now.Add(-60 * time.Second)},
				{StartsAt: now.Add(-50 * time.Second)},
				{StartsAt: now.Add(-40 * time.Second)},
				{StartsAt: now.Add(-30 * time.Second)},
				{StartsAt: now.Add(-20 * time.Second)},
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now},
			},
			time: now,
			idx:  6,
		},
		{
			entries: []Entry{
				{StartsAt: now.Add(-60 * time.Second)},
				{StartsAt: now.Add(-50 * time.Second)},
				{StartsAt: now.Add(-40 * time.Second)},
				{StartsAt: now.Add(-30 * time.Second)},
				{StartsAt: now.Add(-20 * time.Second)},
				{StartsAt: now.Add(-10 * time.Second)},
				{StartsAt: now.Add(-1 * time.Second)},
			},
			time: now,
			idx:  6,
		},
	}

	for _, e := range examples {
		idx := binarySearch(e.entries, e.time, 0, len(e.entries)-1)
		require.Equal(t, e.idx, idx)
	}
}
