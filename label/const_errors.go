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

import "errors"

var (
	// ErrNoType indicates that the parsing failed because no type was specified.
	ErrNoType = errors.New("label: no type specified")

	// ErrUnknownType indicates that the parsing failed because the type is
	// unknown.
	ErrUnknownType = errors.New("label: unknown type")

	// ErrNoJSONURL indicates that the parsing failed because no JSON URL was
	// specified.
	ErrNoJSONURL = errors.New("label: no JSON URL specified")

	// ErrInvalidHTTPUrl indicates that the parsing failed because the specified
	// URL is an invalid HTTP URL.
	ErrInvalidHTTPUrl = errors.New("label: invalid HTTP URL")

	// ErrInvalidDuration indicates that the parsing failed because the duration
	// is invalid.
	ErrInvalidDuration = errors.New("label: invalid duration")

	// ErrUnknownPolicy indicates that the parsing failed because the policy is
	// unknown.
	ErrUnknownPolicy = errors.New("label: unknown policy")
)
