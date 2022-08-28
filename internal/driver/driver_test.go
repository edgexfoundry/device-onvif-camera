// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/IOTechSystems/onvif/device"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/interfaces/mocks"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDeviceName = "test-device"
)

func createDriverWithMockService() (*Driver, *mocks.DeviceServiceSDK) {

	mockService := &mocks.DeviceServiceSDK{}
	driver := &Driver{sdkService: mockService, lc: logger.MockLogger{}}
	return driver, mockService
}

func createTestDevice() models.Device {
	return models.Device{Name: testDeviceName, Protocols: map[string]models.ProtocolProperties{
		OnvifProtocol: map[string]string{
			DeviceStatus: Unreachable,
		},
	}}
}

func createTestDeviceWithProtocols(protocols map[string]models.ProtocolProperties) models.Device {
	return models.Device{Name: testDeviceName, Protocols: protocols}
}

func TestParametersFromURLRawQuery(t *testing.T) {
	parameters := `{ "ProfileToken": "Profile_1" }`
	base64EncodedStr := base64.StdEncoding.EncodeToString([]byte(parameters))
	req := sdkModel.CommandRequest{
		Attributes: map[string]interface{}{
			URLRawQuery: fmt.Sprintf("%s=%s", jsonObject, base64EncodedStr),
		},
	}
	data, err := parametersFromURLRawQuery(req)
	require.NoError(t, err)
	assert.Equal(t, parameters, string(data))
}

func TestDriver_HandleReadCommands(t *testing.T) {
	tests := []struct {
		name          string
		deviceName    string
		protocols     map[string]models.ProtocolProperties
		reqs          []sdkModel.CommandRequest
		expected      []*sdkModel.CommandValue
		errorExpected bool
	}{
		{
			name:       "simple read",
			deviceName: "testDevice",
			reqs: []sdkModel.CommandRequest{
				{
					DeviceResourceName: "GetDeviceInformation",
					Attributes: map[string]interface{}{
						"getFunction": "GetDeviceInformation",
						"service":     "Device",
					},
					Type: "Object",
				}},
			expected: []*sdkModel.CommandValue{
				{
					DeviceResourceName: "DeviceInformation",
					Type:               "Object",
					Value: device.GetDeviceInformationResponse{
						Manufacturer:    "Intel",
						Model:           "SimCamera",
						FirmwareVersion: "2.5a",
						SerialNumber:    "9a32410c",
						HardwareId:      "1.0",
					},
					Origin: 0,
					Tags:   nil,
				}},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			d := &Driver{}
			actual, err := d.HandleReadCommands(test.deviceName, test.protocols, test.reqs)
			if test.errorExpected {
				require.Error(t, err)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}
