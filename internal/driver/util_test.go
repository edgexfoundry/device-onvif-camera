// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/base64"
	"fmt"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParametersFromURLRawQuery(t *testing.T) {
	parameters := `{ "ProfileToken": "Profile_1" }`
	base64EncodedStr := base64.StdEncoding.EncodeToString([]byte(parameters))
	req := sdkModel.CommandRequest{
		Attributes: map[string]interface{}{
			URLRawQuery: fmt.Sprintf("%s=%s", jsonObject, base64EncodedStr),
		},
	}
	data, err := parametersFromURLRawQuery(req)
	require.NoError(t, err)
	assert.Equal(t, parameters, string(data))
}

// TestAddressAndPort verifies splitting of address and port from a given string.
func TestAddressAndPort(t *testing.T) {

	tests := []struct {
		input           string
		expectedAddress string
		expectedPort    string
	}{
		{
			input:           "localhost:80",
			expectedAddress: "localhost",
			expectedPort:    "80",
		},
		{
			input:           "localhost",
			expectedAddress: "localhost",
			expectedPort:    "80",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			resultAddress, resultPort := addressAndPort(test.input)
			assert.Equal(t, test.expectedAddress, resultAddress)
			assert.Equal(t, test.expectedPort, resultPort)
		})
	}
}
