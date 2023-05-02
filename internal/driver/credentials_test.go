// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
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
		secretName    string
		expected      Credentials
		errorExpected bool
		mockUsername  string
		mockPassword  string
		mockAuthMode  string
	}{
		{
			secretName:   noAuthSecretName,
			mockUsername: "",
			mockPassword: "",
			mockAuthMode: onvif.NoAuth,
			expected: Credentials{
				AuthMode: AuthModeNone,
			},
		},
		{
			secretName:   "validSecretName",
			mockUsername: "username",
			mockPassword: "password",
			mockAuthMode: onvif.DigestAuth,
			expected: Credentials{
				AuthMode: AuthModeDigest,
				Username: "username",
				Password: "password",
			},
		},
		{
			secretName:    "invalidSecretName",
			mockUsername:  "username",
			mockPassword:  "password",
			mockAuthMode:  onvif.DigestAuth,
			errorExpected: true,
		},
		{
			secretName:   "validSecretNameInvalidAuthMode",
			mockUsername: "username",
			mockPassword: "password",
			mockAuthMode: "invalidAuthMode",
			expected: Credentials{
				AuthMode: AuthModeUsernameToken,
				Username: "username",
				Password: "password",
			},
		},
	}

	driver, mockService := createDriverWithMockService()
	mockSecretProvider := &mocks.SecretProvider{}
	mockService.On("SecretProvider").Return(mockSecretProvider)
	lookups := 3

	for _, test := range tests {
		test := test
		t.Run(test.secretName, func(t *testing.T) {
			// reset the secret provider
			*mockSecretProvider = mocks.SecretProvider{}

			if test.errorExpected {
				mockSecretProvider.On("GetSecret", test.secretName, UsernameKey, PasswordKey, AuthModeKey).
					Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unit test error", nil)).
					Times(lookups) // expect to be called for every lookup
			} else if test.secretName != noAuthSecretName { // expect not to be called for noAuth
				mockSecretProvider.On("GetSecret", test.secretName, UsernameKey, PasswordKey, AuthModeKey).
					Return(map[string]string{"username": test.mockUsername, "password": test.mockPassword, "mode": test.mockAuthMode}, nil).
					Once() // expect to only be called once (cached lookups afterward)
			}

			// perform the lookup multiple times to check caching
			for i := 0; i < lookups; i++ {
				actual, err := driver.getCredentials(test.secretName)
				if test.errorExpected {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			}
			mockSecretProvider.AssertExpectations(t)
		})
	}
}

// TestTryGetCredentialsForDevice verifies correct credentials are returned for a device based on the MAC address of the device.
func TestTryGetCredentialsForDevice(t *testing.T) {

	tests := []struct {
		existingProtocols map[string]models.ProtocolProperties
		device            models.Device
		expected          Credentials
		secretName        string

		errorExpected bool
		username      string
		password      string
		authMode      string
	}{
		{
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: "",
				},
			},

			secretName:    "default_secret_name",
			username:      "username",
			password:      "password",
			authMode:      onvif.DigestAuth,
			errorExpected: true,
		},
		{
			existingProtocols: map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: "aa:bb:cc:dd:ee:ff",
				},
			},

			secretName: "secret_name",
			username:   "username",
			password:   "password",
			authMode:   onvif.UsernameTokenAuth,
			expected: Credentials{
				AuthMode: AuthModeUsernameToken,
				Username: "username",
				Password: "password",
			},
		},
	}

	driver, mockService := createDriverWithMockService()

	driver.macAddressMapper = NewMACAddressMapper(mockService)
	driver.macAddressMapper.credsMap = convertMACMappings(t, map[string]string{
		"secret_name": "aa:bb:cc:dd:ee:ff",
	})
	driver.config = &ServiceConfig{
		AppCustom: CustomConfig{
			DefaultSecretName: "default_secret_name",
		},
	}

	mockSecretProvider := &mocks.SecretProvider{}
	mockService.On("SecretProvider").Return(mockSecretProvider)

	for _, test := range tests {
		test := test
		t.Run(test.secretName, func(t *testing.T) {
			// reset the secret provider
			*mockSecretProvider = mocks.SecretProvider{}

			if test.errorExpected {
				mockSecretProvider.On("GetSecret", test.secretName, UsernameKey, PasswordKey, AuthModeKey).
					Return(nil, errors.NewCommonEdgeX(errors.KindServerError, "unit test error", nil))
			} else {
				mockSecretProvider.On("GetSecret", test.secretName, UsernameKey, PasswordKey, AuthModeKey).
					Return(map[string]string{"username": test.username, "password": test.password, "mode": test.authMode}, nil)
			}

			actual, err := driver.tryGetCredentialsForDevice(createTestDeviceWithProtocols(test.existingProtocols))
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected.Username, actual.Username)
			assert.Equal(t, test.expected.Password, actual.Password)
			assert.Equal(t, test.expected.AuthMode, actual.AuthMode)
		})
	}
}
