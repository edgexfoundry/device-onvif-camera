// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/edgexfoundry/device-onvif-camera/internal/driver/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testMACAddress   = "ab-cd-ef-12-34-56"
	testFriendlyName = "Outdoor camera"
)

func createOnvifClientWithMockDevice(driver *Driver, deviceName string) (*OnvifClient, *mocks.OnvifDevice) {
	mockDevice := &mocks.OnvifDevice{}
	return &OnvifClient{
		driver:      driver,
		DeviceName:  deviceName,
		onvifDevice: mockDevice,
		lc:          logger.NewMockClient(),
	}, mockDevice
}

func TestOnvifClient_getFriendlyName(t *testing.T) {
	driver, mockService := createDriverWithMockService()

	tests := []struct {
		name      string
		protocols map[string]models.ProtocolProperties
		expected  string
	}{
		{
			name: "basic get",
			protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					FriendlyName: testFriendlyName,
				},
			},
			expected: testFriendlyName,
		},
		{
			name: "missing",
			protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {},
			},
			expected: "",
		},
		{
			name: "blank",
			protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					FriendlyName: "",
				},
			},
			expected: "",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver:     driver,
				DeviceName: testDeviceName,
			}

			expected, err := sdkModel.NewCommandValue(FriendlyName, common.ValueTypeString, test.expected)
			require.NoError(t, err)

			mockService.On("GetDeviceByName", testDeviceName).
				Return(createTestDeviceWithProtocols(test.protocols), nil).Once()

			actual, err := onvifClient.callCustomFunction(FriendlyName, GetFriendlyName, nil, nil)
			mockService.AssertExpectations(t)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestOnvifClient_setFriendlyName(t *testing.T) {
	driver, mockService := createDriverWithMockService()

	tests := []struct {
		name              string
		existingProtocols map[string]models.ProtocolProperties
		data              string
		errorExpected     bool
	}{
		{
			name: "no data (error)",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {},
			},
			data:          ``,
			errorExpected: true,
		},
		{
			name: "happy path set",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {},
			},
			data: "Front door camera",
		},
		{
			name: "happy path set from empty",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					FriendlyName: "",
				},
			},
			data: "Front door camera",
		},
		{
			name: "happy path update",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					FriendlyName: "Front door camera",
				},
			},
			data: "Outdoor camera",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver:     driver,
				DeviceName: testDeviceName,
			}

			testDevice := createTestDeviceWithProtocols(test.existingProtocols)
			mockService.On("GetDeviceByName", testDeviceName).
				Return(testDevice, nil).Once()
			if !test.errorExpected {
				mockService.On("UpdateDevice", mock.AnythingOfType("models.Device")).
					Return(nil).Once()
			}

			_, err := onvifClient.callCustomFunction(FriendlyName, SetFriendlyName, nil, []byte(test.data))
			mockService.AssertExpectations(t)

			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// make sure UpdateDevice was called with the new value
			mockService.AssertCalled(t, "UpdateDevice", mock.MatchedBy(func(d models.Device) bool {
				return d.Protocols[OnvifProtocol][FriendlyName] == test.data
			}))

		})
	}
}

func TestOnvifClient_getMac(t *testing.T) {
	driver, mockService := createDriverWithMockService()

	tests := []struct {
		name      string
		protocols map[string]models.ProtocolProperties
		expected  string
	}{
		{
			name: "basic get",
			protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: testMACAddress,
				},
			},
			expected: testMACAddress,
		},
		{
			name: "missing",
			protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {},
			},
		},
		{
			name: "blank",
			protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: "",
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver:     driver,
				DeviceName: testDeviceName,
			}

			expected, err := sdkModel.NewCommandValue(MACAddress, common.ValueTypeString, test.expected)
			require.NoError(t, err)

			mockService.On("GetDeviceByName", testDeviceName).
				Return(createTestDeviceWithProtocols(test.protocols), nil).Once()

			actual, err := onvifClient.callCustomFunction(MACAddress, GetMACAddress, nil, nil)
			mockService.AssertExpectations(t)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestOnvifClient_setMACAddress(t *testing.T) {
	driver, mockService := createDriverWithMockService()

	tests := []struct {
		name              string
		existingProtocols map[string]models.ProtocolProperties
		data              string
		expected          string
		errorExpected     bool
	}{
		{
			name: "no data (error)",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {},
			},
			data:          "",
			errorExpected: true,
		},
		{
			name: "happy path set",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {},
			},
			data:     "ab-cd-ef-12-34-56",
			expected: "ab:cd:ef:12:34:56",
		},
		{
			name: "happy path set from empty",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: "",
				},
			},
			data:     "AB-12-eF-12-34-56",
			expected: "ab:12:ef:12:34:56",
		},
		{
			name: "happy path update",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: "ab:cd:ef:12:34:56",
				},
			},
			data:     "12-2A-BC-45-56-67",
			expected: "12:2a:bc:45:56:67",
		},
		{
			name: "bad mac address (error)",
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: "ab:cd:ef:12:34:56",
				},
			},
			data:          "-cd-ef-12-34-56",
			errorExpected: true,
		},
		{
			name:              "bogus (error)",
			existingProtocols: map[string]models.ProtocolProperties{},
			data:              "bogus",
			errorExpected:     true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver:     driver,
				DeviceName: testDeviceName,
			}

			testDevice := createTestDeviceWithProtocols(test.existingProtocols)
			mockService.On("GetDeviceByName", testDeviceName).
				Return(testDevice, nil).Once()
			if !test.errorExpected {
				mockService.On("UpdateDevice", mock.AnythingOfType("models.Device")).
					Return(nil).Once()
			}

			_, err := onvifClient.callCustomFunction(MACAddress, SetMACAddress, nil, []byte(test.data))
			mockService.AssertExpectations(t)
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// make sure UpdateDevice was called with the new value
			mockService.AssertCalled(t, "UpdateDevice", mock.MatchedBy(func(d models.Device) bool {
				return d.Protocols[OnvifProtocol][MACAddress] == test.expected
			}))

		})
	}
}
