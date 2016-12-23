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

package label

const (
	//
	// Hypochronos
	//

	// Hypochronos label namespace.
	Hypochronos = "de.mtneug.hypochronos"

	//
	// Timetable
	//

	// Timetable label namespace.
	Timetable = Hypochronos + ".timetable"

	// Type label.
	Type = Timetable + ".type"

	// TimetablePeriod label.
	TimetablePeriod = Timetable + ".period"

	// TimetableJSONUrl label.
	TimetableJSONUrl = Timetable + ".json.url"

	//
	// Node
	//

	// Node label namespace.
	Node = Hypochronos + ".node"

	// NodePolicy label.
	NodePolicy = Node + ".policy"

	// NodeMinDuration label.
	NodeMinDuration = Node + ".min_duration"
)