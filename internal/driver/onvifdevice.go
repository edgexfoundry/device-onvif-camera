// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/IOTechSystems/onvif"
	"net/http"
)

//go:generate mockery --name=OnvifDevice

// OnvifDevice is an interface that abstracts out the onvif.Device struct
type OnvifDevice interface {
	// GetServices returns all available endpoints by service
	GetServices() map[string]string

	// GetDeviceInfo returns the onvif.DeviceInfo
	GetDeviceInfo() onvif.DeviceInfo

	// GetEndpoint returns specific ONVIF service endpoint address
	GetEndpoint(name string) string

	// CallMethod function calls a method, defined <method> struct.
	// You should use Authenticate method to call authorized requests.
	CallMethod(method interface{}) (*http.Response, error)

	GetDeviceParams() onvif.DeviceParams

	GetEndpointByRequestStruct(requestStruct interface{}) (string, error)

	SendSoap(endpoint string, xmlRequestBody string) (resp *http.Response, err error)

	CallOnvifFunction(serviceName, functionName string, data []byte) (interface{}, error)

	// SendGetSnapshotRequest sends the Get request to retrieve the snapshot from the Onvif camera
	// The parameter url is come from the "GetSnapshotURI" command.
	SendGetSnapshotRequest(url string) (resp *http.Response, err error)
}
