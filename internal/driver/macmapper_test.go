// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeMACAddress(t *testing.T) {
	tests := []struct {
		mac       string // input mac address
		sanitized string // expected sanitized output mac address
		err       bool   // expect an error
	}{
		{
			mac:       "aa:bb:cc:dd:ee:ff",
			sanitized: "aa:bb:cc:dd:ee:ff",
		},
		{
			mac: "aa:bb:cc:dd:ee",
			err: true,
		},
		{
			mac: "aabbccddee",
			err: true,
		},
		{
			mac: "aa:bb:cc-dd-ee",
			err: true,
		},
		{
			mac:       "AA:BB:CC:DD:EE:FF",
			sanitized: "aa:bb:cc:dd:ee:ff",
		},
		{
			mac:       "11-22-33-44-55-66",
			sanitized: "11:22:33:44:55:66",
		},
		{
			mac:       " 11-22-33-44-55-66",
			sanitized: "11:22:33:44:55:66",
		},
		{
			mac:       "11-22-33-44-55-66 ",
			sanitized: "11:22:33:44:55:66",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.mac, func(t *testing.T) {
			sanitized, err := SanitizeMACAddress(test.mac)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.sanitized, sanitized)
		})
	}
}

func TestMACAddressBytewiseReverse(t *testing.T) {
	tests := []struct {
		mac      string // input mac address
		reversed string // expected output byte-wise reversed mac address
		err      bool   // expect an error or not
	}{
		{
			mac:      "aa:bb:cc:dd:ee:ff",
			reversed: "ff:ee:dd:cc:bb:aa",
		},
		{
			mac:      "12:34:56:78:9a:bc",
			reversed: "bc:9a:78:56:34:12",
		},
		{
			mac:      "00-00-00-00-fe-80-00-00-00-00-00-00-02-00-5e-10-00-00-00-01",
			reversed: "01:00:00:00:10:5e:00:02:00:00:00:00:00:00:80:fe:00:00:00:00",
		},
		{
			mac: "ab-cd-ef-ab",
			err: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.mac, func(t *testing.T) {
			reversed, err := macAddressBytewiseReverse(test.mac)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.reversed, reversed)
		})
	}
}

// convertMACMappings takes the place of macMapper.UpdateMappings, which is unable to be mocked currently.
func convertMACMappings(t *testing.T, raw map[string]string) map[string]string {
	credsMap := make(map[string]string)
	for secretPath, macs := range raw {
		for _, mac := range strings.Split(macs, ",") {
			sanitized, err := SanitizeMACAddress(mac)
			if err != nil {
				t.Logf("Skipping invalid mac address %s: %s", mac, err.Error())
				continue
			}
			// note: if the mac address already has a mapping, we do not overwrite it
			if existing, found := credsMap[sanitized]; found {
				t.Logf("Unable to set credential group to %s. MAC address '%s' already belongs to credential group %s.", secretPath, mac, existing)
			} else {
				credsMap[sanitized] = secretPath
			}
		}
	}
	return credsMap
}

func TestMatchEndpointRefAddressToMAC(t *testing.T) {
	_, mockService := createDriverWithMockService()
	macMapper := NewMACAddressMapper(mockService)
	macMapper.credsMap = convertMACMappings(t, map[string]string{
		"bosch":     "00:07:5f:c4:23:b6,00:07:5f:d8:85:f9",
		"geovision": "00:13:e2:25:95:6f",
		"tapo":      "10:27:f5:ea:88:f4",
		"honeywell": "00:40:84:f8:c1:05",
	})

	tests := []struct {
		endpointRef string
		mac         string
	}{
		{
			// Bosch DINION IP starlight 6000 HD (F0009143)
			// Present in byte-wise reverse towards the middle
			endpointRef: "00075fc4-23b6-b623-c45f-0700075fc45f",
			mac:         "00:07:5f:c4:23:b6",
		},
		{
			// Bosch DINION IP starlight 6000 HD (F0009143)
			// Present in byte-wise reverse towards the middle
			endpointRef: "00075fd8-85f9-f985-d85f-0700075fd85f",
			mac:         "00:07:5f:d8:85:f9",
		},
		{
			// Geovision GV-BX8700-FD
			// Present at the end
			endpointRef: "d4a02dea-afca-11ec-45e7-0013e225956f",
			mac:         "00:13:e2:25:95:6f",
		},
		{
			// Tapo C200
			// Present at the end
			endpointRef: "3fa1fe68-b915-4053-a3e1-1027f5ea88f4",
			mac:         "10:27:f5:ea:88:f4",
		},
		{
			// HONEYWELL HC30WB5R1 (ROSSINI)
			// Not Present!
			endpointRef: "54072677-7e74-dabf-24eb-a12a321db374",
			// real mac address is "00:40:84:f8:c1:05", but Honeywell does not use in endpoint ref
			mac: "",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.endpointRef, func(t *testing.T) {
			assert.Equal(t, test.mac, macMapper.MatchEndpointRefAddressToMAC(test.endpointRef))
		})
	}
}

func TestMACAddressMapper_UpdateMappings(t *testing.T) {
	tests := []struct {
		name              string
		currentMap        map[string]string
		expected          map[string]string
		alternateExpected map[string]string
	}{
		{
			name: "no update",
			currentMap: map[string]string{
				"creds1": "AA:BB:CC:DD:EE:FF",
				"creds2": "11:22:33:44:55:66",
			},
			expected: map[string]string{
				"aa:bb:cc:dd:ee:ff": "creds1",
				"11:22:33:44:55:66": "creds2",
			},
		},
		{
			name: "single update",
			currentMap: map[string]string{
				"creds1": "aa:bb:cc:dd:ee:ff",
				"creds2": "11:22:33:44:55:66",
				"creds3": "ff:ee:dd:cc:bb:aa",
			},
			expected: map[string]string{
				"aa:bb:cc:dd:ee:ff": "creds1",
				"11:22:33:44:55:66": "creds2",
				"ff:ee:dd:cc:bb:aa": "creds3",
			},
		},
		{
			name: "multiple valid creds",
			currentMap: map[string]string{
				"creds1": "aa:bb:cc:dd:ee:ff,12:23:34:45:56:11,a1:b2:c3:d4:e5:f6",
				"creds2": "11:22:33:44:55:66",
				"creds3": "ff:ee:dd:cc:bb:aa",
			},
			expected: map[string]string{
				"aa:bb:cc:dd:ee:ff": "creds1",
				"12:23:34:45:56:11": "creds1",
				"a1:b2:c3:d4:e5:f6": "creds1",
				"11:22:33:44:55:66": "creds2",
				"ff:ee:dd:cc:bb:aa": "creds3",
			},
		},
		{
			name: "Add invalid macs",
			currentMap: map[string]string{
				"creds3": "FF:EE:DD:CC:BB:AA,asbc,asdf",
				"creds1": "AA:BB:CC:DD:EE:FF",
				"creds2": "11:22:33:44:55:66",
			},
			expected: map[string]string{
				"aa:bb:cc:dd:ee:ff": "creds1",
				"11:22:33:44:55:66": "creds2",
				"ff:ee:dd:cc:bb:aa": "creds3",
			},
		},
		{
			name: "duplicate macs",
			currentMap: map[string]string{
				"creds1": "FF:EE:DD:CC:BB:AA",
				"creds2": "11:22:33:44:55:66",
				"creds3": "FF:EE:DD:CC:BB:AA",
			},
			expected: map[string]string{
				"ff:ee:dd:cc:bb:aa": "creds1",
				"11:22:33:44:55:66": "creds2",
			},
			alternateExpected: map[string]string{
				"ff:ee:dd:cc:bb:aa": "creds3",
				"11:22:33:44:55:66": "creds2",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {

			driver, mockService := createDriverWithMockService()
			driver.macAddressMapper = NewMACAddressMapper(mockService)
			driver.macAddressMapper.credsMap = test.currentMap
			mockSecretProvider := &mocks.SecretProvider{}
			mockLoggingClient := logger.NewMockClient()

			require.NotEmpty(t, test.currentMap)

			for secretPath := range test.currentMap {
				if strings.ToLower(secretPath) != noAuthSecretPath {
					mockSecretProvider.On("GetSecret", secretPath, UsernameKey, PasswordKey, AuthModeKey).
						Return(nil, nil)
				}
			}

			mockService.On("GetSecretProvider").
				Return(mockSecretProvider)
			mockService.On("GetLoggingClient").Return(mockLoggingClient)
			driver.macAddressMapper.UpdateMappings(test.currentMap)

			if test.alternateExpected != nil {
				ex1 := reflect.DeepEqual(test.expected, driver.macAddressMapper.credsMap)
				ex2 := reflect.DeepEqual(test.alternateExpected, driver.macAddressMapper.credsMap)
				assert.True(t, ex1 || ex2)
				return
			}
			assert.Equal(t, test.expected, driver.macAddressMapper.credsMap)
		})
	}
}

// TestTryGetSecretPathForMACAddress verifies the correct secret path is returned for a given mac address.
func TestTryGetSecretPathForMACAddress(t *testing.T) {

	mappedMac := "aa:bb:cc:dd:ee:ff"
	defaultSecretPath := "default_secret_path"

	tests := []struct {
		name     string
		mac      string // input mac address
		expected string
	}{
		{
			name:     "mac address for valid secret path",
			mac:      mappedMac,
			expected: "valid_secret_path",
		},
		{
			name:     "mac address for default secret path",
			mac:      "bb:bb:cc:dd:ee:ff",
			expected: defaultSecretPath,
		},
		{
			name:     "invalid mac address",
			mac:      "invalid_mac",
			expected: noAuthSecretPath,
		},
	}

	driver, mockService := createDriverWithMockService()
	mockLogger := logger.NewMockClient()
	mockService.On("GetLoggingClient").Return(mockLogger)

	driver.macAddressMapper = NewMACAddressMapper(mockService)
	driver.macAddressMapper.credsMap = convertMACMappings(t, map[string]string{
		"valid_secret_path": mappedMac,
	})
	driver.configMu = new(sync.RWMutex)
	driver.config = &ServiceConfig{
		AppCustom: CustomConfig{
			DefaultSecretPath: defaultSecretPath,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			actual := driver.macAddressMapper.TryGetSecretPathForMACAddress(test.mac, driver.config.AppCustom.DefaultSecretPath)
			assert.Equal(t, test.expected, actual)
		})
	}
}
