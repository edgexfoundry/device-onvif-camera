// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			input:    onvif.UsernameTokenAuth,
			expected: true,
		},
		{
			input:    onvif.Both,
			expected: true,
		},
		{
			input:    onvif.NoAuth,
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

// TestTryGetCredentials verifies correct credentials are returned.
func TestTryGetCredentials(t *testing.T) {

	tests := []struct {
		path          string
		expected      Credentials
		errorExpected bool
		username      string
		password      string
		authMode      string
	}{
		{
			path:     noAuthSecretPath,
			username: "username",
			password: "password",
			authMode: onvif.DigestAuth,
			expected: Credentials{
				AuthMode: AuthModeNone,
			},
			errorExpected: false,
		},
		{
			path:     "validPath",
			username: "username",
			password: "password",
			authMode: onvif.DigestAuth,
			expected: Credentials{
				AuthMode: AuthModeDigest,
				Username: "username",
				Password: "password",
			},
			errorExpected: false,
		},
		{
			path:          "invalidPath",
			username:      "username",
			password:      "password",
			authMode:      onvif.DigestAuth,
			expected:      Credentials{},
			errorExpected: true,
		},
	}

	driver, mockService := createDriverWithMockService()

	mockSecretProvider := &mocks.SecretProvider{}

	for i, _ := range tests {
		if tests[i].errorExpected {
			mockSecretProvider.On("GetSecret", tests[i].path, UsernameKey, PasswordKey, AuthModeKey).Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unit test error", nil)).Once()

		} else {
			mockSecretProvider.On("GetSecret", tests[i].path, UsernameKey, PasswordKey, AuthModeKey).Return(map[string]string{"username": tests[i].username, "password": tests[i].password, "mode": tests[i].authMode}, nil).Once()
		}
	}

	mockService.On("GetSecretProvider").Return(mockSecretProvider)

	for _, test := range tests {
		test := test
		t.Run(test.path, func(t *testing.T) {
			actual, err := driver.tryGetCredentials(test.path)

			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
