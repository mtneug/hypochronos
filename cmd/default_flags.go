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

package cmd

import (
	"errors"
	"time"

	"github.com/mtneug/hypochronos/api/types"
	"github.com/mtneug/hypochronos/label"
	flag "github.com/spf13/pflag"
)

func readAndSetDefaults(flags *flag.FlagSet) error {
	// --default-timetable-period
	defaultTimetablePeriodStr, err := flags.GetString("default-timetable-period")
	if err != nil {
		return err
	}
	defaultTimetablePeriod, err := time.ParseDuration(defaultTimetablePeriodStr)
	if err != nil {
		return err
	}
	label.DefaultPeriod = defaultTimetablePeriod

	// --default-minimum-scheduling-duration
	defaultMinimumSchedulingDurationStr, err := flags.GetString("default-minimum-scheduling-duration")
	if err != nil {
		return err
	}
	defaultMinimumSchedulingDuration, err := time.ParseDuration(defaultMinimumSchedulingDurationStr)
	if err != nil {
		return err
	}
	label.DefaultMinDuration = defaultMinimumSchedulingDuration

	// --default-policy
	defaultPolicyStr, err := flags.GetString("default-policy")
	if err != nil {
		return err
	}
	defaultPolicy := types.Policy(defaultPolicyStr)
	if defaultPolicy != types.PolicyActivated && defaultPolicy != types.PolicyDeactivated {
		return errors.New("cmd: unknown policy")
	}
	label.DefaultPolicy = defaultPolicy

	return nil
}

func init() {
	rootCmd.Flags().String("default-timetable-period", "1m", "Default timetable update period")
	rootCmd.Flags().String("default-policy", "activated", "Default policy (activated or deactivated)")
	rootCmd.Flags().String("default-minimum-scheduling-duration", "1m", "Default minimum sheduling duration")
}
