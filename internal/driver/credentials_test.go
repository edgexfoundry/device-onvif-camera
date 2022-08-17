// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/IOTechSystems/onvif"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTryGetCredentials_noAuth verifies when a credential is passed with a noAuth path, noAuth credentials are returned.
func TestTryGetCredentials_noAuth(t *testing.T) {
	driver, _ := createDriverWithMockService()
	result, err := driver.tryGetCredentials(noAuthSecretPath)

	expected := Credentials{
		AuthMode: AuthModeNone,
	}
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestIsAuthModeValid verifies auth mode is set correctly.
func TestIsAuthModeValid(t *testing.T) {

	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    onvif.DigestAuth,
			expected: true,
		},
		{
			input:    onvif.DigestAuth,
			expected: true,
		},
		{
			input:    onvif.DigestAuth,
			expected: true,
		},
		{
			input:    onvif.DigestAuth,
			expected: true,
		},
		{
			input:    "invalidValue",
			expected: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			result := IsAuthModeValid(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}
