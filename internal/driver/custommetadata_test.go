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

func getTestDevice() contract.Device {
	return contract.Device{
		Protocols: map[string]contract.ProtocolProperties{
			CustomMetadata: {
				"Location":          "Front door",
				"CommonName":        "Front door camera",
				"Installation date": "01/01/2022",
				"Maintenance date":  "05/01/2022",
			},
		},
	}
}

func TestOnvifClient_getCustomMetadata(t *testing.T) {
	tests := []struct {
		name          string
		device        contract.Device
		data          string
		expected      contract.ProtocolProperties
		errorExpected bool
	}{
		{
			name:     "happy path with no data (get all)",
			device:   getTestDevice(),
			data:     "",
			expected: getTestDevice().Protocols[CustomMetadata],
		},
		{
			name:   "happy path with data (single field)",
			device: getTestDevice(),
			data:   `{"CustomMetadata":["CommonName"]}`,
			expected: contract.ProtocolProperties{
				"CommonName": getTestDevice().Protocols[CustomMetadata]["CommonName"],
			},
		},
		{
			name:   "happy path with data (multiple fields)",
			device: getTestDevice(),
			data:   `{"CustomMetadata":["Location","CommonName"]}`,
			expected: contract.ProtocolProperties{
				"Location":   getTestDevice().Protocols[CustomMetadata]["Location"],
				"CommonName": getTestDevice().Protocols[CustomMetadata]["CommonName"],
			},
		},
		{
			name:     "happy path with data (single non-existent field)",
			device:   getTestDevice(),
			data:     `{"CustomMetadata":["Movie"]}`,
			expected: contract.ProtocolProperties{},
		},
		{
			name:     "happy path with data (multiple non-existent fields)",
			device:   getTestDevice(),
			data:     `{"CustomMetadata":["Movie", "Height"]}`,
			expected: contract.ProtocolProperties{},
		},
		{
			name:          "empty data (error)",
			device:        getTestDevice(),
			data:          `{"CustomMetadata":[]}`,
			expected:      contract.ProtocolProperties{},
			errorExpected: true,
		},
		{
			name:          "bad json (error)",
			device:        getTestDevice(),
			data:          "bogus",
			expected:      getTestDevice().Protocols[CustomMetadata],
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
				DeviceName: "myDevice",
			}
			actual, err := onvifClient.getCustomMetadata(test.device, []byte(test.data))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestOnvifClient_setCustomMetadata(t *testing.T) {
	tests := []struct {
		name          string
		device        contract.Device
		data          string
		expected      contract.ProtocolProperties
		errorExpected bool
	}{
		{
			name: "no data (error)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
			data:          `{}`,
			expected:      contract.ProtocolProperties{},
			errorExpected: true,
		},
		{
			name: "happy path set (single data field)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
			data: `{"CommonName":"Front door camera"}`,
			expected: contract.ProtocolProperties{
				"CommonName": "Front door camera",
			},
		},
		{
			name: "happy path set (multiple data fields)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
			data: `{
				"CommonName":"Front door camera",
				"Location":"Front door"
			}`,
			expected: contract.ProtocolProperties{
				"CommonName": "Front door camera",
				"Location":   "Front door",
			},
		},
		{
			name: "happy path update (single data field)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
					},
				},
			},
			data: `{"CommonName": "Outdoor camera"}`,
			expected: contract.ProtocolProperties{
				"CommonName": "Outdoor camera",
			},
		},
		{
			name: "happy path update (multiple data fields)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
						"Location":   "Front door",
						"Color":      "Black and white",
						"Condition":  "Good working condition",
					},
				},
			},
			data: `{
				"CommonName": "Outdoor camera multiple",
				"Location":   "Outside multiple",
				"Color":      "Purple multiple",
				"Condition":  "Bad working condition multiple"
			}`,
			expected: contract.ProtocolProperties{
				"CommonName": "Outdoor camera multiple",
				"Location":   "Outside multiple",
				"Color":      "Purple multiple",
				"Condition":  "Bad working condition multiple",
			},
		},
		{
			name: "happy path delete (single data field)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
					},
				},
			},
			data:     `{"CommonName":"delete"}`,
			expected: contract.ProtocolProperties{},
		},
		{
			name: "happy path delete (multple data fields)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
						"Location":   "Front door",
						"Color":      "Black and white",
						"Condition":  "Good working condition",
					},
				},
			},
			data: `{
				"CommonName":"delete",
				"Location":"delete",
				"Color":"delete",
				"Condition":"delete"
			}`,
			expected: contract.ProtocolProperties{},
		},
		{
			name:          "bad json (error)",
			device:        contract.Device{},
			data:          "bogus",
			expected:      contract.ProtocolProperties{},
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
				DeviceName: "getTestDevice()",
			}
			updatedDevice, err := onvifClient.setCustomMetadata(test.device, []byte(test.data))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, updatedDevice.Protocols[CustomMetadata])
		})
	}
}
