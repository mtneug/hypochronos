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
	"time"

	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/servicehandler"
)

var (
	// DefaultPeriod for service handler.
	DefaultPeriod = 30 * time.Minute

	// DefaultMinDuration for service handler.
	DefaultMinDuration = 1 * time.Minute

	// DefaultPolicy for service handler.
	DefaultPolicy = types.PolicyActivated
)

// ParseServiceHandler parses the labels and sets the corresponding values for
// given service handler.
func ParseServiceHandler(sh *servicehandler.ServiceHandler, labels map[string]string) error {
	// period
	periodStr, ok := labels[Period]
	if !ok {
		sh.Period = DefaultPeriod
	} else {
		period, err := time.ParseDuration(periodStr)
		if err != nil {
			return ErrInvalidDuration
		}
		sh.Period = period
	}

	// min duration
	minDurationStr, ok := labels[MinDuration]
	if !ok {
		sh.MinDuration = DefaultMinDuration
	} else {
		minDuration, err := time.ParseDuration(minDurationStr)
		if err != nil {
			return ErrInvalidDuration
		}
		sh.MinDuration = minDuration
	}

	// policy
	policyStr, ok := labels[Policy]
	if !ok {
		sh.Policy = DefaultPolicy
	} else {
		policy := types.Policy(policyStr)
		if policy != types.PolicyActivated && policy != types.PolicyDeactivated {
			return ErrUnknownPolicy
		}
		sh.Policy = policy
	}

	return nil
}
