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

package model

import "time"

// TTLFilePath to write the TTL file to.
const TTLFilePath = "/etc/container-ttl"

// TTLData API object used in TTLResponse.
type TTLData struct {
	TTL time.Time `json:"ttl"`
}

// TTLResponse API object.
type TTLResponse struct {
	API
	Metadata Metadata `json:"metadata"`
	Data     TTLData  `json:"data"`
}
