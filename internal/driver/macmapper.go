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
			if !IsValidMacAddress(mac) {
				m.driver.lc.Warnf("'%s' is not a valid MAC address! Skipping entry.", mac)
				continue
			}
			credsMap[mac] = secretPath
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

func (m *MacAddressMapper) GetSecretPathForMacAddress(mac string) (string, bool) {
	m.credsMu.RLock()
	defer m.credsMu.RUnlock()

	// note: we cannot directly return the lookup
	secretPath, ok := m.credsMap[mac]
	return secretPath, ok
}

func (m *MacAddressMapper) TryGetCredentialsForMacAddress(mac string) (config.Credentials, error) {
	secretPath, ok := m.GetSecretPathForMacAddress(mac)
	if !ok {
		return config.Credentials{}, fmt.Errorf("no mapping exists for mac address %s", mac)
	}
	return m.driver.tryGetCredentials(secretPath)
}
