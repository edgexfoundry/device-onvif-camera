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

type testCase struct {
	device           contract.Device
	data             string
	expectedResponse contract.ProtocolProperties
}

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

	singleDataTest := testCase{
		data: `{
		"CustomMetadata":[
			"CommonName"
		]
		}`,
		expectedResponse: contract.ProtocolProperties{
			"CommonName": myDevice.Protocols[CustomMetadata]["CommonName"],
		},
	}

	happyMultipleDataTest := testCase{
		data: `{
		"CustomMetadata":[
			"Location",
			"CommonName"
		]
		}`,
		expectedResponse: contract.ProtocolProperties{
			"Location":   myDevice.Protocols[CustomMetadata]["Location"],
			"CommonName": myDevice.Protocols[CustomMetadata]["CommonName"],
		},
	}

	noFieldDataTest := testCase{
		data: `{
		"CustomMetadata":["Movie"]
		}`,
		expectedResponse: contract.ProtocolProperties{
			"Movie": "This field does not exist in custom metadata",
		},
	}

	noFieldsDataTest := testCase{
		data: `{
		"CustomMetadata":["Movie", "Height"]
		}`,
		expectedResponse: contract.ProtocolProperties{
			"Movie":  "This field does not exist in custom metadata",
			"Height": "This field does not exist in custom metadata",
		},
	}

	tests := []struct {
		name          string
		device        contract.Device
		data          string
		expected      contract.ProtocolProperties
		errorExpected bool
	}{
		{"happy path without data", myDevice, "", myDevice.Protocols[CustomMetadata], false},
		{"happy path with data for single field", myDevice, singleDataTest.data, singleDataTest.expectedResponse, false},
		{"happy path with data for multiple field", myDevice, happyMultipleDataTest.data, happyMultipleDataTest.expectedResponse, false},
		{"happy path with data for non-existent field", myDevice, noFieldDataTest.data, noFieldDataTest.expectedResponse, false},
		{"happy path with data for multiple non-existent fields", myDevice, noFieldsDataTest.data, noFieldsDataTest.expectedResponse, false},
		{"badJson", myDevice, "bogus", myDevice.Protocols[CustomMetadata], true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver: &Driver{
					lc: logger.NewMockClient(),
				},
				DeviceName: "myDevice",
			}
			actual, err := onvifClient.getCustomMetadata(tt.device, []byte(tt.data))
			if tt.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestOnvifClient_setCustomMetadata(t *testing.T) {

	noDataTest := testCase{
		device: contract.Device{
			Protocols: map[string]contract.ProtocolProperties{
				CustomMetadata: {},
			},
		},
		data:             `{}`,
		expectedResponse: contract.ProtocolProperties{},
	}
	singleDataTest := testCase{
		device: contract.Device{
			Protocols: map[string]contract.ProtocolProperties{
				CustomMetadata: {},
			},
		},
		data: `{
			"CommonName":"Front door camera"
		}`,
		expectedResponse: contract.ProtocolProperties{
			"CommonName": "Front door camera",
		},
	}
	multipleDataTest := testCase{
		device: contract.Device{
			Protocols: map[string]contract.ProtocolProperties{
				CustomMetadata: {},
			},
		},
		data: `{
			"CommonName":"Front door camera",
			"Location":"Front door"
		}`,
		expectedResponse: contract.ProtocolProperties{
			"CommonName": "Front door camera",
			"Location":   "Front door",
		},
	}
	updateSingleDataTest := testCase{
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
			"CommonName": "Outdoor camera"
				}`,
		expectedResponse: contract.ProtocolProperties{
			"CommonName": "Outdoor camera",
			"Location":   "Front door",
			"Color":      "Black and white",
			"Condition":  "Good working condition",
		},
	}
	updateMultipleDataTest := testCase{
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
		expectedResponse: contract.ProtocolProperties{
			"CommonName": "Outdoor camera multiple",
			"Location":   "Outside multiple",
			"Color":      "Purple multiple",
			"Condition":  "Bad working condition multiple",
		},
	}

	deleteSingleDataTest := testCase{
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
		data: `{"CommonName":""}`,
		expectedResponse: contract.ProtocolProperties{
			"Location":  "Front door",
			"Color":     "Black and white",
			"Condition": "Good working condition",
		},
	}
	deleteMultipleDataTest := testCase{
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
			"CommonName":"",
			"Location":"",
			"Color":"",
			"Condition":""
		}`,
		expectedResponse: contract.ProtocolProperties{},
	}
	badJson := testCase{
		device:           contract.Device{},
		data:             "bogus",
		expectedResponse: contract.ProtocolProperties{},
	}

	tests := []struct {
		name          string
		test          testCase
		errorExpected bool
	}{
		{"happy-path-withoutData", noDataTest, false},
		{"happy-path-withSingleData", singleDataTest, false}, // TODO: add base 64
		{"happy-path-withMultipleData", multipleDataTest, false},
		{"happy-path-updateSingleData", updateSingleDataTest, false},
		{"happy-path-updateMultipleData", updateMultipleDataTest, false},
		{"happy-path-deleteSingleData", deleteSingleDataTest, false},
		{"happy-path-deleteMultipleData", deleteMultipleDataTest, false},
		{"badJson", badJson, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onvifClient := &OnvifClient{
				driver: &Driver{
					lc: logger.NewMockClient(),
				},
				DeviceName: "myDevice",
			}
			err := onvifClient.setCustomMetadata(tt.test.device, []byte(tt.test.data))
			if tt.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.test.expectedResponse, tt.test.device.Protocols[CustomMetadata])
		})
	}
}
