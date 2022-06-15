// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"regexp"
	"strings"
	"sync"
)

var (
	// macRegex is a regular expression that matches MAC addresses in the following 3 formats:
	//		11:22:33:44:55:66, 11-22-33-44-55-66, and 112233445566
	// It does not match the use of mixed separators (both - and : used in the same MAC address)
	macRegex = regexp.MustCompile("^[a-fA-F0-9]{2}((:[a-fA-F0-9]{2}){5}|(-[a-fA-F0-9]{2}){5}|([a-fA-F0-9]{2}){5})$")
)

type MacAddressMapper struct {
	// driver is a pointer to the current Driver instance
	driver *Driver
	// credsMu is for locking access to the credsMap
	credsMu sync.RWMutex
	// credsMap is a map between mac address to secretPath
	credsMap map[string]string
}

func NewMacAddressMapper(driver *Driver) *MacAddressMapper {
	return &MacAddressMapper{
		driver:   driver,
		credsMap: make(map[string]string),
	}
}

func IsValidMacAddress(mac string) bool {
	return macRegex.MatchString(mac)
}

func (m *MacAddressMapper) UpdateMappings(raw map[string]string) {
	m.credsMu.Lock()
	defer m.credsMu.Unlock()

	credsMap := make(map[string]string)
	for secretPath, macs := range raw {
		if _, err := sdk.RunningService().SecretProvider.GetSecret(secretPath, secret.UsernameKey); err != nil {
			m.driver.lc.Warnf("One or more MAC address mappings exist for the secret path '%s' which does not exist in the Secret Store!", secretPath)
		}

		for _, mac := range strings.Split(macs, ",") {
			sanitized, err := SanitizeMacAddress(mac)
			if err != nil {
				m.driver.lc.Warnf("Skipping entry: %s", err.Error())
				continue
			}
			credsMap[sanitized] = secretPath
		}
	}

	m.credsMap = credsMap
}

// ListMacAddresses will return a slice of mac addresses that have been assigned credentials
func (m *MacAddressMapper) ListMacAddresses() []string {
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

func (m *MacAddressMapper) GetSecretPathForMacAddress(mac string) (string, error) {
	m.credsMu.RLock()
	defer m.credsMu.RUnlock()

	// sanitize the mac address before looking up to ensure they all match the same format
	sanitized, err := SanitizeMacAddress(mac)
	if err != nil {
		return "", err
	}

	secretPath, ok := m.credsMap[sanitized]
	if !ok {
		return "", fmt.Errorf("no mapping exists for mac address '%s'", mac)
	}

	return secretPath, nil
}

func (m *MacAddressMapper) TryGetCredentialsForMacAddress(mac string) (config.Credentials, error) {
	secretPath, err := m.GetSecretPathForMacAddress(mac)
	if err != nil {
		return config.Credentials{}, err
	}
	return m.driver.tryGetCredentials(secretPath)
}

// SanitizeMacAddress takes in a MAC address in one of the 3 formats:
// 		aa:bb:cc:dd:ee:ff, 11-22-33-44-55-66, and 112233445566
// and will return it in the following format, using all capital letters:
//		AA:BB:CC:DD:EE:FF
func SanitizeMacAddress(mac string) (string, error) {
	if !IsValidMacAddress(mac) {
		return "", fmt.Errorf("'%s' is not a valid MAC Address", mac)
	}
	mac = strings.ToUpper(mac)
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	// note: we already verified that it is a valid MAC address, so no need to do length checking
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12]), nil
}
