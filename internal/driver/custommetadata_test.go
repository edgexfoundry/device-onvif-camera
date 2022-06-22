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

func TestOnvifClient_getCustomMetadata(t *testing.T) {

	myDevice := contract.Device{
		Protocols: map[string]contract.ProtocolProperties{
			CustomMetadata: {
				"Location":          "Front door",
				"CommonName":        "Front door camera",
				"Installation date": "01/01/2022",
				"Maintenance date":  "05/01/2022",
			},
		},
	}

	tests := []struct {
		name          string
		device        contract.Device
		data          string
		expected      contract.ProtocolProperties
		errorExpected bool
	}{
		{
			"happy path without data - getAll",
			myDevice,
			"",
			myDevice.Protocols[CustomMetadata],
			false},
		{
			"happy path with data for single field",
			myDevice,
			`{"CustomMetadata":["CommonName"]}`,
			contract.ProtocolProperties{
				"CommonName": myDevice.Protocols[CustomMetadata]["CommonName"],
			},
			false,
		},
		{
			"happy path with data for multiple field",
			myDevice,
			`{"CustomMetadata":["Location","CommonName"]}`,
			contract.ProtocolProperties{
				"Location":   myDevice.Protocols[CustomMetadata]["Location"],
				"CommonName": myDevice.Protocols[CustomMetadata]["CommonName"],
			},
			false,
		},
		{
			"happy path with data for non-existent field",
			myDevice,
			`{"CustomMetadata":["Movie"]}`,
			contract.ProtocolProperties{},
			false,
		},
		{
			"happy path with data for multiple non-existent fields",
			myDevice,
			`{"CustomMetadata":["Movie", "Height"]}`,
			contract.ProtocolProperties{},
			false,
		},
		{
			"with empty data",
			myDevice,
			`{"CustomMetadata":[]}`,
			contract.ProtocolProperties{},
			true,
		},
		{
			"badJson",
			myDevice,
			"bogus",
			myDevice.Protocols[CustomMetadata],
			true,
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
			"happy-path-withoutData",
			contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
			`{}`,
			contract.ProtocolProperties{},
			true,
		},
		{
			"happy-path-withSingleData",
			contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
			`{"CommonName":"Front door camera"}`,
			contract.ProtocolProperties{
				"CommonName": "Front door camera",
			},
			false,
		},
		{
			"happy-path-withMultipleData",
			contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {},
				},
			},
			`{
				"CommonName":"Front door camera",
				"Location":"Front door"
			}`,
			contract.ProtocolProperties{
				"CommonName": "Front door camera",
				"Location":   "Front door",
			},
			false,
		},
		{
			"happy-path-updateSingleData",
			contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
					},
				},
			},
			`{"CommonName": "Outdoor camera"}`,
			contract.ProtocolProperties{
				"CommonName": "Outdoor camera",
			},
			false,
		},
		{
			"happy-path-updateMultipleData",
			contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
						"Location":   "Front door",
						"Color":      "Black and white",
						"Condition":  "Good working condition",
					},
				},
			},
			`{
			"CommonName": "Outdoor camera multiple",
			"Location":   "Outside multiple",
			"Color":      "Purple multiple",
			"Condition":  "Bad working condition multiple"
		}`,
			contract.ProtocolProperties{
				"CommonName": "Outdoor camera multiple",
				"Location":   "Outside multiple",
				"Color":      "Purple multiple",
				"Condition":  "Bad working condition multiple",
			},
			false,
		},
		{
			"happy-path-deleteSingleData",
			contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
					},
				},
			},
			`{"CommonName":"delete"}`,
			contract.ProtocolProperties{},
			false,
		},
		{
			"happy-path-deleteMultipleData",
			contract.Device{
				Protocols: map[string]contract.ProtocolProperties{
					CustomMetadata: {
						"CommonName": "Front door camera",
						"Location":   "Front door",
						"Color":      "Black and white",
						"Condition":  "Good working condition",
					},
				},
			},
			`{
			"CommonName":"delete",
			"Location":"delete",
			"Color":"delete",
			"Condition":"delete"
			}`,
			contract.ProtocolProperties{},
			false,
		},
		{
			"badJson",
			contract.Device{},
			"bogus",
			contract.ProtocolProperties{},
			true,
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
			updatedDevice, err := onvifClient.setCustomMetadata(test.device, []byte(test.data))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, updatedDevice.Protocols[CustomMetadata])
		})
	}
}
