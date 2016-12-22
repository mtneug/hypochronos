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

package labels

const (
	//
	// Hypochronos
	//

	// HypochronosNs label namespace.
	HypochronosNs = "de.mtneug.hypochronos"

	//
	// Timetable
	//

	// TimetableNs label namespace.
	TimetableNs = HypochronosNs + ".timetable"

	// Type label.
	Type = TimetableNs + ".type"

	// TimetablePeriod label.
	TimetablePeriod = TimetableNs + ".period"

	// TimetableJSONUrl label.
	TimetableJSONUrl = TimetableNs + ".json.url"

	//
	// Node
	//

	// NodeNs label namespace.
	NodeNs = HypochronosNs + ".node"

	// NodePolicy label.
	NodePolicy = NodeNs + ".policy"

	// NodeMinDuration label.
	NodeMinDuration = NodeNs + ".min_duration"
)
