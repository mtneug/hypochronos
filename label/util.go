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

// Equal returns true if the same values for the hypochronos labels are used.
func Equal(labels1, labels2 map[string]string) bool {
	return labels1[Period] == labels2[Period] &&
		labels1[Policy] == labels2[Policy] &&
		labels1[Type] == labels2[Type] &&
		labels1[TimetableJSONUrl] == labels2[TimetableJSONUrl] &&
		labels1[NodeMinDuration] == labels2[NodeMinDuration]
}
