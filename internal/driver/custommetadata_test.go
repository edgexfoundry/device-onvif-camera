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
			data:   `["CommonName"]`,
			expected: contract.ProtocolProperties{
				"CommonName": getTestDevice().Protocols[CustomMetadata]["CommonName"],
			},
		},
		{
			name:   "happy path with data (multiple fields)",
			device: getTestDevice(),
			data:   `["Location","CommonName"]`,
			expected: contract.ProtocolProperties{
				"Location":   getTestDevice().Protocols[CustomMetadata]["Location"],
				"CommonName": getTestDevice().Protocols[CustomMetadata]["CommonName"],
			},
		},
		{
			name:     "happy path with data (single non-existent field)",
			device:   getTestDevice(),
			data:     `["Movie"]`,
			expected: contract.ProtocolProperties{},
		},
		{
			name:     "happy path with data (multiple non-existent fields)",
			device:   getTestDevice(),
			data:     `["Movie", "Height"]`,
			expected: contract.ProtocolProperties{},
		},
		{
			name:     "happy path create CustomMetadata",
			device:   contract.Device{},
			expected: contract.ProtocolProperties{},
		},
		{
			name:          "empty data (error)",
			device:        getTestDevice(),
			data:          `[]`,
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
		expected      contract.Device
		errorExpected bool
	}{
		{
			name: "no data (error)",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
			data: `{}`,
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
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
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {"CommonName": "Front door camera"},
				},
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
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
						"Location":   "Front door",
					},
				},
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
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Outdoor camera",
					},
				},
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
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Outdoor camera multiple",
						"Location":   "Outside multiple",
						"Color":      "Purple multiple",
						"Condition":  "Bad working condition multiple",
					},
				},
			},
		},
		{
			name:     "happy path with data (single empty field key)",
			device:   getTestDevice(),
			data:     `{"": "Bad key"}`,
			expected: getTestDevice(),
		},
		{
			name:     "happy path with data (multiple empty field keys)",
			device:   getTestDevice(),
			data:     `{"": "Bad key","":"Another bad key"}`,
			expected: getTestDevice(),
		},
		{
			name: "Custom Metadata doesn't exist",
			device: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {
						"Hello": "World",
					},
				},
			},
			data: `{
				"CommonName": "Outdoor camera multiple",
				"Location":   "Outside multiple",
				"Color":      "Purple multiple",
				"Condition":  "Bad working condition multiple"
			}`,
			expected: contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					OnvifProtocol: {
						"Hello": "World",
					},
					CustomMetadata: {
						"CommonName": "Outdoor camera multiple",
						"Location":   "Outside multiple",
						"Color":      "Purple multiple",
						"Condition":  "Bad working condition multiple",
					},
				},
			},
		},
		{
			name:          "bad json (error)",
			device:        contract.Device{},
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
				DeviceName: "getTestDevice()",
			}
			updatedDevice, err := onvifClient.setCustomMetadata(test.device, []byte(test.data))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, updatedDevice)
		})
	}
}

func TestOnvifClient_deleteCustomMetadata(t *testing.T) {
	tests := []struct {
		name          string
		device        contract.Device
		data          string
		expected      contract.ProtocolProperties
		errorExpected bool
	}{
		{
			name:   "happy path with data (single field)",
			device: getTestDevice(),
			data:   `["CommonName"]`,
			expected: contract.ProtocolProperties{
				"Location":          "Front door",
				"Installation date": "01/01/2022",
				"Maintenance date":  "05/01/2022",
			},
		},
		{
			name:   "happy path with data (multiple fields)",
			device: getTestDevice(),
			data:   `["Location","CommonName"]`,
			expected: contract.ProtocolProperties{
				"Installation date": "01/01/2022",
				"Maintenance date":  "05/01/2022",
			},
		},
		{
			name:     "happy path with data (single non-existent field)",
			device:   getTestDevice(),
			data:     `["Movie"]`,
			expected: getTestDevice().Protocols[CustomMetadata],
		},
		{
			name:     "happy path with data (multiple non-existent fields)",
			device:   getTestDevice(),
			data:     `["Movie", "Car"]`,
			expected: getTestDevice().Protocols[CustomMetadata],
		},
		{
			name:          "empty data (error)",
			device:        getTestDevice(),
			data:          `[]`,
			expected:      getTestDevice().Protocols[CustomMetadata],
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
			actual, err := onvifClient.deleteCustomMetadata(test.device, []byte(test.data))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual.Protocols[CustomMetadata])
		})
	}
}
