// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/startup"

	"github.com/edgexfoundry/device-onvif-camera"
	"github.com/edgexfoundry/device-onvif-camera/internal/driver"
)

const (
	serviceName string = "device-onvif-camera"
)

func main() {
	sd := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, device_camera.Version, sd)
}
