// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDeviceMac() contract.Device {
	return contract.Device{
		Protocols: map[string]contract.ProtocolProperties{
			OnvifProtocol: {
				MACAddress: "ab-cd-ef-12-34-56",
			},
		},
	}
}

func TestOnvifClient_getMac(t *testing.T) {
	tests := []struct {
		name          string
		device        contract.Device
		expected      string
		errorExpected bool
	}{
		{
			name:     "get",
			device:   getTestDeviceMac(),
			expected: getTestDeviceMac().Protocols[OnvifProtocol][MACAddress],
		},
		{
			name: "missing macaddress",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {},
				},
			},
			errorExpected: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver:     &Driver{},
				DeviceName: "myDevice",
			}
			actual, err := onvifClient.getMACAddress(test.device)
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestOnvifClient_setMACAddress(t *testing.T) {
	tests := []struct {
		name          string
		device        contract.Device
		data          string
		expected      contract.Device
		errorExpected bool
	}{
		{
			name: "no data (error)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {},
				},
			},
			data:          `{}`,
			errorExpected: true,
		},
		{
			name: "happy path set",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {},
				},
			},
			data: `{"MACAddress":"ab-cd-ef-12-34-56"}`,
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {"MACAddress": "ab-cd-ef-12-34-56"},
				},
			},
		},
		{
			name: "happy path update",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {
						"MACAddress": "ab-cd-ef-12-34-56s",
					},
				},
			},
			data: `{"MACAddress": "12-23-34-45-56-67"}`,
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {
						"MACAddress": "12-23-34-45-56-67",
					},
				},
			},
		},
		{
			name: "bad mac address",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {
						"MACAddress": "ab-cd-ef-12-34-56s",
					},
				},
			},
			data:          `{"MACAddress": "-cd-ef-12-34-56"}`,
			errorExpected: true,
		},
		{
			name:          "single empty field key",
			device:        getTestDeviceMac(),
			data:          `{"": "Bad key"}`,
			errorExpected: true,
		},
		{
			name:          "bad json (error)",
			device:        contract.Device{Protocols: map[string]contract.ProtocolProperties{}},
			data:          "bogus",
			expected:      contract.Device{},
			errorExpected: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver: &Driver{
					lc: logger.NewMockClient(),
				},
			}
			updatedDevice, err := onvifClient.setMACAddress(test.device, []byte(test.data))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, updatedDevice)
		})
	}
}
