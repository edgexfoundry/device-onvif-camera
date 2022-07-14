/*
 * Copyright (C) 2022 Canonical Ltd
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * SPDX-License-Identifier: Apache-2.0'
 */

package main

import (
	"strings"

	"github.com/canonical/edgex-snap-hooks/v2/log"
	"github.com/canonical/edgex-snap-hooks/v2/options"
	"github.com/canonical/edgex-snap-hooks/v2/snapctl"
)

// configure is called by the main function
func configure() {
	log.SetComponentName("configure")

	log.Info("Enabling config options")
	err := snapctl.Set("app-options", "true").Run()
	if err != nil {
		log.Fatalf("could not enable config options: %v", err)
	}

	log.Info("Processing options")
	err = options.ProcessAppConfig("device-onvif-camera")
	if err != nil {
		log.Fatalf("could not process options: %v", err)
	}

	// If autostart is not explicitly set, default to "no"
	// as only example service configuration and profiles
	// are provided by default.
	autostart, err := snapctl.Get("autostart").Run()
	if err != nil {
		log.Fatalf("Reading config 'autostart' failed: %v", err)
	}
	if autostart == "" {
		log.Debug("autostart is NOT set, initializing to 'no'")
		autostart = "no"
	}
	autostart = strings.ToLower(autostart)
	log.Debugf("autostart=%s", autostart)

	// services are stopped/disabled by default in the install hook
	switch autostart {
	case "true", "yes":
		err = snapctl.Start("device-usb-camera").Enable().Run()
		if err != nil {
			log.Fatalf("Can't start service: %s", err)
		}
	case "false", "no":
		// no action necessary
	default:
		log.Fatalf("Invalid value for 'autostart': %s", autostart)
	}
}
