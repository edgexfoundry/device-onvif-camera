// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
)

func createTestDeviceList() []contract.Device {
	return []models.Device{
		{
			Name: "device-onvif-camera", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: "abc",
				},
			},
		},
		{
			Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: "123",
				},
			},
		},
		{
			Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: "456",
				},
			},
		},
		{
			Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: "789",
				},
			},
		},
	}
}

func createDiscoveredList() []sdkModel.DiscoveredDevice {
	return []sdkModel.DiscoveredDevice{
		{
			Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: "123",
				},
			},
		},
		{
			Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: "456",
				},
			},
		},
		{
			Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: "789",
				},
			},
		},
	}
}

func TestOnvifDiscovery_makeDeviceMap(t *testing.T) {
	tests := []struct {
		name      string
		devices   []contract.Device
		deviceMap map[string]contract.Device
		nameCalls int
	}{
		{
			name:    "3 devices",
			devices: createTestDeviceList(),
			deviceMap: map[string]contract.Device{
				"123": {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "123",
						},
					},
				},
				"456": {
					Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "456",
						},
					},
				},
				"789": {
					Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "789",
						},
					},
				},
			},
			nameCalls: 4,
		},
		{
			name: "NoProtocol",
			devices: []contract.Device{
				{
					Name: "testDevice1",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "123",
						},
					},
				},
				{
					Name:      "testDevice2",
					Protocols: map[string]models.ProtocolProperties{},
				},
			},
			deviceMap: map[string]contract.Device{
				"123": {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "123",
						},
					},
				},
			},
			nameCalls: 2,
		},
		{
			name: "NoEndpointReference",
			devices: []contract.Device{
				{
					Name: "testDevice1",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "123",
						},
					},
				},
				{
					Name: "testDevice2",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "",
						},
					},
				},
			},
			deviceMap: map[string]contract.Device{
				"123": {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "123",
						},
					},
				},
			},
			nameCalls: 2,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			driver, mockService := createDriverWithMockService()
			mockService.On("Devices").
				Return(test.devices).Once()
			mockService.On("Name").
				Return("device-onvif-camera").Times(test.nameCalls)
			devices := driver.makeDeviceRefMap()
			mockService.AssertExpectations(t)

			assert.Equal(t, devices, test.deviceMap)
		})
	}
}

func TestOnvifDiscovery_discoveryFilter(t *testing.T) {
	tests := []struct {
		name              string
		devices           []contract.Device
		discoveredDevices []sdkModel.DiscoveredDevice
		filtered          []sdkModel.DiscoveredDevice
		nameCalls         int
	}{
		{
			name:              "No new devices",
			devices:           createTestDeviceList(),
			discoveredDevices: createDiscoveredList(),
			filtered:          []sdkModel.DiscoveredDevice{},
			nameCalls:         4,
		},
		{
			name: "All new devices",
			devices: []contract.Device{
				{
					Name: "device-onvif-camera", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "abc",
						},
					},
				},
			},
			discoveredDevices: createDiscoveredList(),
			filtered:          createDiscoveredList(),
			nameCalls:         1,
		},
		{
			name:    "new and old devices",
			devices: createTestDeviceList(),
			discoveredDevices: []sdkModel.DiscoveredDevice{
				{
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "123",
						},
					},
				},
				{
					Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "456",
						},
					},
				},
				{
					Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "789",
						},
					},
				},
				{
					Name: "testDevice4", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "xyz",
						},
					},
				},
				{
					Name: "testDevice5", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "def",
						},
					},
				},
			},
			filtered: []sdkModel.DiscoveredDevice{
				{
					Name: "testDevice4", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "xyz",
						},
					},
				},
				{
					Name: "testDevice5", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: "def",
						},
					},
				},
			},
			nameCalls: 4,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			driver, mockService := createDriverWithMockService()
			mockService.On("Devices").
				Return(test.devices).Times(2)
			mockService.On("Name").
				Return("device-onvif-camera").Times(2 * test.nameCalls)
			filtered := driver.discoverFilter(test.discoveredDevices)
			mockService.AssertExpectations(t)

			assert.Equal(t, test.filtered, filtered)
		})
	}
}

// func createParams() (params onvif.DeviceParams) {
// 	return onvif.DeviceParams{
// 		Xaddr:              "1.1.1.1:1",
// 		EndpointRefAddress: "1234",
// 		Username:           "hello",
// 		Password:           "world",
// 	}
// }

// func createDevice() onvif.Device {
// 	device, _ := onvif.NewDevice(createParams())
// 	return *device
// }

// func TestmockService_createDiscoveredDevice(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		device           onvif.Device
// 		discoveredDevice sdkModel.DiscoveredDevice
// 		errorExpected    bool
// 	}{
// 		{
// 			name:   "happy path",
// 			device: createDevice(),
// 			discoveredDevice: sdkModel.DiscoveredDevice{
// 				Name: "1.1.1.1:1",
// 				Protocols: map[string]contract.ProtocolProperties{
// 					OnvifProtocol: {
// 						Address:            "1.1.1.1",
// 						Port:               "1",
// 						SecretPath:         "credentials001",
// 						EndpointRefAddress: "1234",
// 						DeviceStatus:       "Reachable",
// 						LastSeen:           time.Now().Format(time.UnixDate),
// 					},
// 					CustomMetadata: {},
// 				},
// 			},
// 		},
// 	}
// 	for _, test := range tests {
// 		test := test
// 		t.Run(test.name, func(t *testing.T) {
// 			mockService := mockService{
// 				lc:       logger.NewMockClient(),
// 				configMu: &sync.RWMutex{},
// 			}
// 			actualDevice, err := mockService.createDiscoveredDevice(test.device)
// 			if test.errorExpected {
// 				assert.Error(t, err)
// 				return
// 			}
// 			require.NoError(t, err)

// 			assert.Equal(t, test.discoveredDevice, actualDevice)
// 		})
// 	}
// }
