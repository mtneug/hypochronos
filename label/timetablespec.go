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

import (
	"net/url"

	"github.com/mtneug/hypochronos/model"
	"github.com/mtneug/hypochronos/timetable"
)

var (
	// DefaultState for timetable spec.
	DefaultState = model.StateActivated
)

// ParseTimetableSpec parses the labels and sets the corresponding values for
// given timetable specification.
func ParseTimetableSpec(tts *timetable.Spec, labels map[string]string) error {
	typeStr, ok := labels[TimetableType]
	if !ok {
		return ErrNoType
	}

	err := parseTimetableSpecCommon(tts, labels)
	if err != nil {
		return err
	}

	switch timetable.Type(typeStr) {
	case timetable.TypeJSON:
		return parseTimetableSpecJSON(tts, labels)
	}

	return ErrUnknownType
}

func parseTimetableSpecCommon(tts *timetable.Spec, labels map[string]string) error {
	// default state
	defaultStateStr, ok := labels[TimetableDefaultState]
	if !ok {
		tts.DefaultState = DefaultState
	} else {
		defaultState := timetable.State(defaultStateStr)
		if defaultState != model.StateActivated && defaultState != model.StateDeactivated {
			return ErrUnknownState
		}
		tts.DefaultState = defaultState
	}

	return nil
}

func parseTimetableSpecJSON(tts *timetable.Spec, labels map[string]string) error {
	tts.Type = timetable.TypeJSON

	// URL
	urlStr, ok := labels[TimetableJSONUrl]
	if !ok {
		return ErrNoJSONURL
	}
	url, err := url.Parse(urlStr)
	if err != nil || url.Scheme != "http" {
		return ErrInvalidHTTPUrl
	}
	tts.JSONSpec.URL = url.String()

	return nil
}
