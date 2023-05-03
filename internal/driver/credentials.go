// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strings"

	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// Credentials encapsulates username, password, and AuthMode attributes.
// Assign AuthMode to "digest" | "usernametoken" | "both" | "none"
type Credentials struct {
	Username string
	Password string
	AuthMode string
}

const (
	AuthModeDigest        string = onvif.DigestAuth
	AuthModeUsernameToken string = onvif.UsernameTokenAuth
	AuthModeBoth          string = onvif.Both
	AuthModeNone          string = onvif.NoAuth
)

const (
	// noAuthSecretName is the magic string used to define a group which does not use credentials
	// this is defined in lowercase and compared in lowercase
	noAuthSecretName = "noauth"
)

const (
	UsernameKey = "username"
	PasswordKey = "password"
	AuthModeKey = "mode"
)

var (
	noAuthCredentials = Credentials{
		AuthMode: AuthModeNone,
	}
)

func IsAuthModeValid(mode string) bool {
	return mode == AuthModeDigest ||
		mode == AuthModeUsernameToken ||
		mode == AuthModeBoth ||
		mode == AuthModeNone
}

// tryGetCredentialsInternal will attempt one time to get the credentials located at secretName from
// secret provider and return them, otherwise return an error.
func (d *Driver) tryGetCredentialsInternal(secretName string) (Credentials, errors.EdgeX) {
	// if the secret name is the special NoAuth magic key, do not look it up, instead return the noAuthCredentials
	if strings.ToLower(secretName) == noAuthSecretName {
		return noAuthCredentials, nil
	}

	secretData, err := d.sdkService.SecretProvider().GetSecret(secretName, UsernameKey, PasswordKey, AuthModeKey)
	if err != nil {
		return Credentials{}, errors.NewCommonEdgeXWrapper(err)
	}

	credentials := Credentials{
		Username: secretData[UsernameKey],
		Password: secretData[PasswordKey],
		AuthMode: secretData[AuthModeKey],
	}

	if !IsAuthModeValid(secretData[AuthModeKey]) {
		d.lc.Warnf("AuthMode is set to an invalid value: %s. setting value to '%s'.", credentials.AuthMode, AuthModeUsernameToken)
		credentials.AuthMode = AuthModeUsernameToken
	}

	d.lc.Tracef("Found credentials from secret name %s", secretName)
	return credentials, nil
}

// tryGetCredentialsForDevice will attempt to use the device's MAC address to look up the credentials
// from the Secret Store. If a mapping does not exist, or the device's MAC address is missing or invalid,
// the default secret name will be used to look up the credentials. An error is returned if the secret name
// does not exist in the Secret Store.
func (d *Driver) tryGetCredentialsForDevice(device models.Device) (Credentials, errors.EdgeX) {
	d.configMu.RLock()
	defaultSecretName := d.config.AppCustom.DefaultSecretName
	d.configMu.RUnlock()

	secretName := defaultSecretName
	macAddress := ""
	if v, ok := device.Protocols[OnvifProtocol][MACAddress]; ok {
		macAddress = fmt.Sprintf("%v", v)
	}
	if macAddress != "" {
		secretName = d.macAddressMapper.TryGetSecretNameForMACAddress(macAddress, defaultSecretName)
	} else {
		d.lc.Warnf("Device %s is missing MAC Address, using default secret name", device.Name)
	}

	credentials, edgexErr := d.tryGetCredentialsInternal(secretName)
	if edgexErr != nil {
		d.lc.Errorf("Failed to retrieve credentials for the secret name %s: %s", secretName, edgexErr.Error())
		return Credentials{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to get credentials", edgexErr)
	}

	return credentials, nil
}

func (d *Driver) secretUpdated(secretName string) {
	d.lc.Infof("Secret updated callback called for secretName %s", secretName)

	d.credsCacheMu.Lock()
	// remove the cache entry for this secret
	delete(d.credsCache, secretName)
	d.credsCacheMu.Unlock()
	_, _ = d.getCredentials(secretName) // update cached data

	for _, device := range d.sdkService.Devices() {
		d.lc.Debugf("Updating onvif client for device %s", device.Name)
		err := d.updateOnvifClient(device)
		if err != nil {
			d.lc.Errorf("Unable to update onvif client for device: %s, %v", device.Name, err)
		}
	}

	d.lc.Debug("Done updating onvif clients")
}

func (d *Driver) getCredentials(secretName string) (Credentials, errors.EdgeX) {
	d.credsCacheMu.RLock()
	creds, found := d.credsCache[secretName]
	d.credsCacheMu.RUnlock()
	if found {
		return creds, nil
	}

	creds, err := d.tryGetCredentialsInternal(secretName)
	if err != nil {
		return Credentials{}, err
	}

	d.credsCacheMu.Lock()
	// cache the credentials and return them
	d.credsCache[secretName] = creds
	d.credsCacheMu.Unlock()
	return creds, nil
}
