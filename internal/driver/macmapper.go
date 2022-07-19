// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type MACAddressMapper struct {
	// credsMu is for locking access to the credsMap
	credsMu sync.RWMutex
	// credsMap is a map between mac address to secretPath
	credsMap map[string]string

	sdkService SDKService
}

// NewMACAddressMapper creates a new MACAddressMapper object
func NewMACAddressMapper(sdkService SDKService) *MACAddressMapper {
	return &MACAddressMapper{
		credsMap:   make(map[string]string),
		sdkService: sdkService,
	}
}

// UpdateMappings takes the raw map of secret path to csv list of mac addresses and
// inverts it into a quick lookup map of mac address to secret path.
func (m *MACAddressMapper) UpdateMappings(raw map[string]string) {
	m.credsMu.Lock()
	defer m.credsMu.Unlock()

	credsMap := make(map[string]string)
	for secretPath, macs := range raw {
		if strings.ToLower(secretPath) != noAuthSecretPath { // do not check for noAuth
			if _, err := m.sdkService.GetSecretProvider().GetSecret(secretPath, UsernameKey, PasswordKey, AuthModeKey); err != nil {
				m.sdkService.GetLoggingClient().Warnf("One or more MAC address mappings exist for the secret path '%s' which does not exist in the Secret Store!", secretPath)
			}
		}

		for _, mac := range strings.Split(macs, ",") {
			sanitized, err := SanitizeMACAddress(mac)
			if err != nil {
				m.sdkService.GetLoggingClient().Warnf("Skipping invalid mac address %s: %s", mac, err.Error())
				continue
			}
			// note: if the mac address already has a mapping, we do not overwrite it
			if existing, found := credsMap[sanitized]; found {
				m.sdkService.GetLoggingClient().Warnf("Unable to set credential group to %s. MAC address '%s' already belongs to credential group %s.", secretPath, mac, existing)
			} else {
				credsMap[sanitized] = secretPath
			}
		}
	}

	m.credsMap = credsMap
}

// ListMACAddresses will return a slice of mac addresses that have been assigned credentials
func (m *MACAddressMapper) ListMACAddresses() []string {
	m.credsMu.RLock()
	defer m.credsMu.RUnlock()

	macs := make([]string, len(m.credsMap))

	i := 0
	for mac := range m.credsMap {
		macs[i] = mac
		i++
	}

	return macs
}

// TryGetSecretPathForMACAddress will return the secret path associated with the mac address passed if a mapping exists,
// the default secret path if the mapping is not found, or no auth if the mac address is invalid.
func (m *MACAddressMapper) TryGetSecretPathForMACAddress(mac string, defaultSecretPath string) string {
	// sanitize the mac address before looking up to ensure they all match the same format
	sanitized, err := SanitizeMACAddress(mac)
	if err != nil {
		m.sdkService.GetLoggingClient().Warnf("Unable to sanitize mac address: %s. Using no authentication.", err.Error())
		return noAuthSecretPath
	}

	m.credsMu.RLock()
	defer m.credsMu.RUnlock()

	secretPath, found := m.credsMap[sanitized]
	if !found {
		m.sdkService.GetLoggingClient().Debugf("No credential mapping exists for mac address '%s', will use default secret path.", mac)
		return defaultSecretPath
	}

	return secretPath
}

// SanitizeMACAddress takes in a MAC address in one of the IEEE 802 MAC-48, EUI-48, EUI-64 formats
// and will return it in the standard go format, using colons and lower case letters:
// Example:	aa:bb:cc:dd:ee:ff
func SanitizeMACAddress(mac string) (string, error) {
	hwAddr, err := net.ParseMAC(strings.TrimSpace(mac))
	if err != nil {
		return "", err
	}
	return hwAddr.String(), nil
}

// macAddressBytewiseReverse returns the byte-wise reverse of the input MAC Address.
// Examples:
// 		aa:bb:cc:dd:ee:ff -> ff:ee:dd:cc:bb:aa
//		12:34:56:78:9a:bc -> bc:9a:78:56:34:12
func macAddressBytewiseReverse(mac string) (string, error) {
	var err error
	if mac, err = SanitizeMACAddress(mac); err != nil {
		return "", err
	}
	mac = strings.ReplaceAll(mac, ":", "")
	if len(mac)%2 != 0 {
		return "", fmt.Errorf("mac address %s has invalid length of %d", mac, len(mac))
	}

	buf := strings.Builder{}
	// loop through the string backwards two characters at a time (1-byte)
	for i := len(mac); i > 0; i -= 2 {
		buf.WriteString(mac[i-2 : i])
		if i > 2 { // only write delimiter if more bytes exist
			buf.WriteByte(':')
		}
	}
	return buf.String(), nil
}

// MatchEndpointRefAddressToMAC will return a mac address if one is found in the Endpoint Reference Address,
// or empty string if not
func (m *MACAddressMapper) MatchEndpointRefAddressToMAC(endpointRef string) string {
	endpointRef = strings.ToLower(strings.ReplaceAll(endpointRef, "-", ""))

	m.credsMu.RLock()
	defer m.credsMu.RUnlock()

	for mac := range m.credsMap {
		if strings.Contains(endpointRef, strings.ReplaceAll(mac, ":", "")) {
			return mac
		}

		reversedMAC, err := macAddressBytewiseReverse(mac)
		if err != nil {
			m.sdkService.GetLoggingClient().Warnf("issue computing byte-wise reverse of MAC address %s: %s", mac, err.Error())
			continue
		}
		if strings.Contains(endpointRef, reversedMAC) {
			return mac
		}
	}

	return "" // not found
}
