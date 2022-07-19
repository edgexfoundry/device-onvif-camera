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

func getTestDeviceFriendly() contract.Device {
	return contract.Device{
		Protocols: map[string]contract.ProtocolProperties{
			OnvifProtocol: {
				FriendlyName: "Outdoor camera",
			},
		},
	}
}

func TestOnvifClient_getFriendlyName(t *testing.T) {
	tests := []struct {
		name          string
		device        contract.Device
		expected      string
		errorExpected bool
	}{
		{
			name:     "get",
			device:   getTestDeviceFriendly(),
			expected: getTestDeviceFriendly().Protocols[OnvifProtocol][FriendlyName],
		},
		{
			name: "missing friendly name (error)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {},
				},
			},
			expected:      "",
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
			actual, err := onvifClient.getFriendlyName(test.device)
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestOnvifClient_setFriendlyName(t *testing.T) {
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
			data: `{"FriendlyName":"Front door camera"}`,
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {"FriendlyName": "Front door camera"},
				},
			},
		},
		{
			name: "happy path update",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {
						"FriendlyName": "Front door camera",
					},
				},
			},
			data: `{"FriendlyName": "Outdoor camera"}`,
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {
						"FriendlyName": "Outdoor camera",
					},
				},
			},
		},
		{
			name:          "empty field key",
			device:        getTestDeviceFriendly(),
			data:          `{"": "Bad key"}`,
			errorExpected: true,
		},
		{
			name:          "bad json (error)",
			device:        contract.Device{Protocols: map[string]contract.ProtocolProperties{}},
			data:          "bogus",
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
			updatedDevice, err := onvifClient.setFriendlyName(test.device, []byte(test.data))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, updatedDevice)
		})
	}
}
