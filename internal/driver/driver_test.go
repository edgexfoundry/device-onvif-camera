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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// TestAddressAndPort splits the address and port from a given string.
func TestAddressAndPort(t *testing.T) {

	tests := []struct {
		input           string
		expectedAddress string
		expectedPort    string
	}{
		{
			input:           "localhost:80",
			expectedAddress: "localhost",
			expectedPort:    "80",
		},
		{
			input:           "localhost",
			expectedAddress: "localhost",
			expectedPort:    "80",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			resultAddress, resultPort := addressAndPort(test.input)
			assert.Equal(t, test.expectedAddress, resultAddress)
			assert.Equal(t, test.expectedPort, resultPort)
		})
	}
}

type mockGetOnvifClient struct {
	mock.Mock
}

type mockCallOnvifFunction struct {
	mock.Mock
}

func (m *mockGetOnvifClient) GetOnvifClient(deviceName string) (*OnvifClient, errors.EdgeX) {
	return &OnvifClient{}, nil
}

func (m *mockCallOnvifFunction) CallOnvifFunction(req sdkModel.CommandRequest, functionType string, data []byte) (cv *sdkModel.CommandValue, edgexErr errors.EdgeX) {
	return &sdkModel.CommandValue{}, nil
}

// func TestDriver_HandleReadCommands(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		deviceName    string
// 		protocols     map[string]models.ProtocolProperties
// 		reqs          []sdkModel.CommandRequest
// 		getFunction   string
// 		data          string
// 		expected      []*sdkModel.CommandValue
// 		errorExpected bool
// 	}{
// 		{
// 			name:       "simple read",
// 			deviceName: "testDevice",
// 			reqs: []sdkModel.CommandRequest{
// 				{
// 					DeviceResourceName: "GetDeviceInformation",
// 					Attributes: map[string]interface{}{
// 						"getFunction": "GetDeviceInformation",
// 						"service":     "Device",
// 					},
// 					Type: "Object",
// 				}},
// 			getFunction: "GetDeviceInformation",
// 			data:        "",
// 			expected: []*sdkModel.CommandValue{
// 				{
// 					DeviceResourceName: "DeviceInformation",
// 					Type:               "Object",
// 					Value: device.GetDeviceInformationResponse{
// 						Manufacturer:    "Intel",
// 						Model:           "SimCamera",
// 						FirmwareVersion: "2.5a",
// 						SerialNumber:    "9a32410c",
// 						HardwareId:      "1.0",
// 					},
// 					Origin: 0,
// 					Tags:   nil,
// 				}},
// 		},
// 	}
// 	for _, test := range tests {
// 		test := test
// 		t.Run(test.name, func(t *testing.T) {
// 			d := &Driver{}
// 			mockCallOnvifFunction := &mockCallOnvifFunction{}
// 			mockGetOnvifClient := &mockGetOnvifClient{}
// 			for _, req := range test.reqs {
// 				mockCallOnvifFunction.On("CallOnvifFunction", req, test.getFunction, []byte(test.data)).
// 					Return()
// 			}
// 			mockGetOnvifClient.GetOnvifClient(test.deviceName)
// 			actual, err := d.HandleReadCommands(test.deviceName, test.protocols, test.reqs)
// 			if test.errorExpected {
// 				require.Error(t, err)
// 			}
// 			assert.Equal(t, test.expected, actual)
// 		})
// 	}
// }

// TestUpdateDevice: test for the proper updating of device information
func TestUpdateDevice(t *testing.T) {

	driver, mockService := createDriverWithMockService()

	tests := []struct {
		device  models.Device
		devInfo *device.GetDeviceInformationResponse

		expectedDevice models.Device
		errorExpected  bool
	}{
		{
			device: contract.Device{
				Name: "testName",
			},
			devInfo: &device.GetDeviceInformationResponse{
				Manufacturer:    "Intel",
				Model:           "SimCamera",
				FirmwareVersion: "2.5a",
				SerialNumber:    "9a32410c",
				HardwareId:      "1.0",
			},
			errorExpected: false,
		},
		{
			device: contract.Device{
				Name: "unknown_unknown_device",
			},
			devInfo: &device.GetDeviceInformationResponse{
				Manufacturer:    "Intel",
				Model:           "SimCamera",
				FirmwareVersion: "2.5a",
				SerialNumber:    "9a32410c",
				HardwareId:      "1.0",
			},
			expectedDevice: contract.Device{
				Name: "Intel-SimCamera-",
			},
			errorExpected: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.device.Name, func(t *testing.T) {
			mockService.On("RemoveDeviceByName", test.device.Name).Return(nil).Once()
			mockService.On("AddDevice", test.expectedDevice).Return(test.expectedDevice.Name, nil).Once()
			mockService.On("UpdateDevice", test.device).Return(nil).Once()

			err := driver.updateDevice(test.device, test.devInfo)

			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
