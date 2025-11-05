// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
)

const (
	uuid1 = "1793dfb2-28b0-11ed-a261-0242ac120002"
	uuid2 = "1793e0a2-28b0-11ed-a261-0242ac120002"
	uuid3 = "1793e19c-28b0-11ed-a261-0242ac120002"
	uuid4 = "8076305c-28b0-11ed-a261-0242ac120002"
	uuid5 = "80763188-28b0-11ed-a261-0242ac120002"
)

func createTestDeviceList() []models.Device {
	return []models.Device{
		{
			Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]interface{}{
					EndpointRefAddress: uuid1,
				},
			},
		},
		{
			Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]interface{}{
					EndpointRefAddress: uuid2,
				},
			},
		},
		{
			Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]interface{}{
					EndpointRefAddress: uuid3,
				},
			},
		},
	}
}

func createDiscoveredList() []sdkModel.DiscoveredDevice {
	return []sdkModel.DiscoveredDevice{
		{
			Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]interface{}{
					EndpointRefAddress: uuid1,
				},
			},
		},
		{
			Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]interface{}{
					EndpointRefAddress: uuid2,
				},
			},
		},
		{
			Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
				OnvifProtocol: map[string]interface{}{
					EndpointRefAddress: uuid3,
				},
			},
		},
	}
}

func TestOnvifDiscovery_makeDeviceMap(t *testing.T) {
	tests := []struct {
		name      string
		devices   []models.Device
		deviceMap map[string]models.Device
	}{
		{
			name:    "3 devices",
			devices: createTestDeviceList(),
			deviceMap: map[string]models.Device{
				uuid1: {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid1,
						},
					},
				},
				uuid2: {
					Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid2,
						},
					},
				},
				uuid3: {
					Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid3,
						},
					},
				},
			},
		},
		{
			name: "NoProtocol",
			devices: []models.Device{
				{
					Name: "testDevice1",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid1,
						},
					},
				},
				{
					Name:      "testDevice2",
					Protocols: map[string]models.ProtocolProperties{},
				},
			},
			deviceMap: map[string]models.Device{
				uuid1: {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid1,
						},
					},
				},
			},
		},
		{
			name: "NoEndpointReference",
			devices: []models.Device{
				{
					Name: "testDevice1",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid1,
						},
					},
				},
				{
					Name: "testDevice2",
					Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: "",
						},
					},
				},
			},
			deviceMap: map[string]models.Device{
				uuid1: {
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid1,
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
		devices           []models.Device
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
			devices:           []models.Device{},
			discoveredDevices: createDiscoveredList(),
			filtered:          createDiscoveredList(),
		},
		{
			name:    "new and old devices",
			devices: createTestDeviceList(),
			discoveredDevices: []sdkModel.DiscoveredDevice{
				{
					Name: "testDevice1", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid1,
						},
					},
				},
				{
					Name: "testDevice2", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid2,
						},
					},
				},
				{
					Name: "testDevice3", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid3,
						},
					},
				},
				{
					Name: "testDevice4", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid4,
						},
					},
				},
				{
					Name: "testDevice5", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid5,
						},
					},
				},
			},
			filtered: []sdkModel.DiscoveredDevice{
				{
					Name: "testDevice4", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
							EndpointRefAddress: uuid4,
						},
					},
				},
				{
					Name: "testDevice5", Protocols: map[string]models.ProtocolProperties{
						OnvifProtocol: map[string]interface{}{
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
