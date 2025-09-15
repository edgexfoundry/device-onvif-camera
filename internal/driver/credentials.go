// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"strings"

	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/spf13/cast"
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

// getCredentialsForDevice will attempt to use the device's MAC address to look up the credentials
// from the Secret Store. If a mapping does not exist, or the device's MAC address is missing or invalid,
// the default secret name will be used to look up the credentials. If the resolved secret name
// does not exist in the Secret Store, noAuthCredentials are returned, allowing the user
// to still call unauthenticated endpoints.
func (d *Driver) getCredentialsForDevice(device models.Device) Credentials {
	d.configMu.RLock()
	defaultSecretName := d.config.AppCustom.DefaultSecretName
	d.configMu.RUnlock()

	secretName := defaultSecretName
	macAddress := ""
	if v, ok := device.Protocols[OnvifProtocol][MACAddress]; ok {
		macAddress = cast.ToString(v)
	}
	if macAddress != "" {
		secretName = d.macAddressMapper.TryGetSecretNameForMACAddress(macAddress, defaultSecretName)
	} else {
		d.lc.Warnf("Device %s is missing MAC Address, using default secret name", device.Name)
	}

	credentials, edgexErr := d.tryGetCredentialsInternal(secretName)
	if edgexErr != nil {
		// if credentials are not found, instead of returning an error, set the AuthMode to NoAuth
		// and allow the user to call unauthenticated endpoints
		d.lc.Errorf("Failed to retrieve credentials for the secret name %s. Falling back to using NoAuth: %s", secretName, edgexErr.Error())
		return noAuthCredentials
	}

	return credentials
}

func (d *Driver) secretUpdated(secretName string) {
	d.lc.Infof("Secret updated callback called for secretName '%s'", secretName)

	for _, device := range d.sdkService.Devices() {
		d.lc.Tracef("Updating onvif client for device %s", device.Name)
		err := d.updateOnvifClient(device)
		if err != nil {
			d.lc.Errorf("Unable to update onvif client for device: %s, %v", device.Name, err)
		}
	}

	d.lc.Trace("Done updating onvif clients")
}
