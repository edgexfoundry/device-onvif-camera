// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultSecretName = "default_secret_name"
	secret1Name       = "secret1"
	testMAC1          = ""
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
					Times(lookups) // expect to be called for every lookup
			}

			// perform the lookup multiple times
			for i := 0; i < lookups; i++ {
				actual, err := driver.tryGetCredentialsInternal(test.secretName)
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

// TestGetCredentialsForDevice verifies correct credentials are returned for a device based on the MAC address of the device.
func TestGetCredentialsForDevice(t *testing.T) {
	testMAC2 := "aa:bb:cc:11:11:11"
	noAuthMAC := "cc:cc:dd:dd:11:22"
	bogusMAC := "14:36:90:65:ff:ab"

	existingSecrets := map[string]Credentials{
		defaultSecretName: {
			Username: "default",
			Password: "password",
			AuthMode: AuthModeBoth,
		},
		secret1Name: {
			Username: "user1",
			Password: "pass1",
			AuthMode: AuthModeUsernameToken,
		},
	}

	tests := []struct {
		name        string
		macAddress  string
		secretStore map[string]Credentials
		expected    Credentials
	}{
		{
			name:        "missing MAC, fallback to default credentials",
			macAddress:  "",
			secretStore: existingSecrets,
			expected:    existingSecrets[defaultSecretName],
		},
		{
			name:        "no mapping for MAC, fallback to default credentials",
			macAddress:  bogusMAC,
			secretStore: existingSecrets,
			expected:    existingSecrets[defaultSecretName],
		},
		{
			name:        "success secret1",
			macAddress:  testMACAddress,
			secretStore: existingSecrets,
			expected:    existingSecrets[secret1Name],
		},
		{
			name:        "mapping points to missing secret, fallback to no auth",
			macAddress:  testMAC2,
			secretStore: existingSecrets,
			expected:    noAuthCredentials,
		},
		{
			name:        "no MAC specified, but missing default credentials secret, fallback to no auth",
			macAddress:  "",
			secretStore: map[string]Credentials{},
			expected:    noAuthCredentials,
		},
		{
			name:        "success - explicitly map to no auth",
			macAddress:  noAuthMAC,
			secretStore: map[string]Credentials{},
			expected:    noAuthCredentials,
		},
	}

	driver, mockService := createDriverWithMockService()

	driver.macAddressMapper = NewMACAddressMapper(mockService)
	driver.macAddressMapper.credsMap = convertMACMappings(t, map[string]string{
		secret1Name:      testMACAddress,
		"bogus":          testMAC2,
		noAuthSecretName: noAuthMAC,
	})
	driver.config = &ServiceConfig{
		AppCustom: CustomConfig{
			DefaultSecretName: defaultSecretName,
		},
	}

	mockSecretProvider := &mocks.SecretProvider{}
	getSecret := mockSecretProvider.On("GetSecret", mock.AnythingOfType("string"), UsernameKey, PasswordKey, AuthModeKey)
	mockService.On("SecretProvider").Return(mockSecretProvider)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			// each run override the response of getSecret based on the provided secret store input
			getSecret.Run(func(args mock.Arguments) {
				if secret, ok := test.secretStore[args.String(0)]; ok {
					getSecret.Return(map[string]string{"username": secret.Username, "password": secret.Password, "mode": secret.AuthMode}, nil)
				} else {
					getSecret.Return(nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("secret %s does not exist", args.String(0)), nil))
				}
			})

			device := createTestDeviceWithProtocols(map[string]models.ProtocolProperties{
				OnvifProtocol: {
					MACAddress: test.macAddress,
				},
			})

			actual := driver.getCredentialsForDevice(device)
			assert.Equal(t, test.expected, actual)
		})
	}
}
