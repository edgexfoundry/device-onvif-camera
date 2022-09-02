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

const (
	uuid1 = "1793ddc8-28b0-11ed-a261-0242ac120002"
	uuid2 = "1793dfb2-28b0-11ed-a261-0242ac120002"
	uuid3 = "1793e0a2-28b0-11ed-a261-0242ac120002"
	uuid4 = "1793e19c-28b0-11ed-a261-0242ac120002"
	uuid5 = "8076305c-28b0-11ed-a261-0242ac120002"
	uuid6 = "80763188-28b0-11ed-a261-0242ac120002"
)

func createTestDeviceList() []contract.Device {
	return []models.Device{
		{
			Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: uuid2,
				},
			},
		},
		{
			Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: uuid3,
				},
			},
		},
		{
			Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: uuid4,
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
					EndpointRefAddress: uuid2,
				},
			},
		},
		{
			Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: uuid3,
				},
			},
		},
		{
			Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]string{
					EndpointRefAddress: uuid4,
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
	}{
		{
			name:    "3 devices",
			devices: createTestDeviceList(),
			deviceMap: map[string]contract.Device{
				uuid2: {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid2,
						},
					},
				},
				uuid3: {
					Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid3,
						},
					},
				},
				uuid4: {
					Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid4,
						},
					},
				},
			},
		},
		{
			name: "NoProtocol",
			devices: []contract.Device{
				{
					Name: "testDevice1",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid2,
						},
					},
				},
				{
					Name:      "testDevice2",
					Protocols: map[string]models.ProtocolProperties{},
				},
			},
			deviceMap: map[string]contract.Device{
				uuid2: {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid2,
						},
					},
				},
			},
		},
		{
			name: "NoEndpointReference",
			devices: []contract.Device{
				{
					Name: "testDevice1",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid2,
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
				uuid2: {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid2,
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			driver, mockService := createDriverWithMockService()
			mockService.On("Devices").
				Return(test.devices).Once()
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
	}{
		{
			name:              "No new devices",
			devices:           createTestDeviceList(),
			discoveredDevices: createDiscoveredList(),
			filtered:          []sdkModel.DiscoveredDevice{},
		},
		{
			name:              "All new devices",
			devices:           []contract.Device{},
			discoveredDevices: createDiscoveredList(),
			filtered:          createDiscoveredList(),
		},
		{
			name:    "new and old devices",
			devices: createTestDeviceList(),
			discoveredDevices: []sdkModel.DiscoveredDevice{
				{
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid2,
						},
					},
				},
				{
					Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid3,
						},
					},
				},
				{
					Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid4,
						},
					},
				},
				{
					Name: "testDevice4", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid6,
						},
					},
				},
				{
					Name: "testDevice5", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid5,
						},
					},
				},
			},
			filtered: []sdkModel.DiscoveredDevice{
				{
					Name: "testDevice4", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid6,
						},
					},
				},
				{
					Name: "testDevice5", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]string{
							EndpointRefAddress: uuid5,
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			driver, mockService := createDriverWithMockService()
			mockService.On("Devices").
				Return(test.devices)
			filtered := driver.discoverFilter(test.discoveredDevices)
			mockService.AssertExpectations(t)

			assert.Equal(t, test.filtered, filtered)
		})
	}
}
