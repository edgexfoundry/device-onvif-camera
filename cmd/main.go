// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/startup"

	"github.com/edgexfoundry/device-onvif-camera"
	"github.com/edgexfoundry/device-onvif-camera/internal/driver"
)

const (
	serviceName string = "device-onvif-camera"
)

func main() {
	startup.Bootstrap(serviceName, device_camera.Version, driver.NewDriver())
}
