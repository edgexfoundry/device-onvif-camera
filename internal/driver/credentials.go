// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
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

// tryGetCredentials will attempt one time to get the credentials located at secretName from
// secret provider and return them, otherwise return an error.
func (d *Driver) tryGetCredentials(secretName string) (Credentials, errors.EdgeX) {
	// if the secret path is the special NoAuth magic key, do not look it up, instead return the noAuthCredentials
	if strings.ToLower(secretName) == noAuthSecretName {
		return noAuthCredentials, nil
	}

	secretData, err := d.sdkService.GetSecretProvider().GetSecret(secretName, UsernameKey, PasswordKey, AuthModeKey)
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

	return credentials, nil
}

// tryGetCredentialsForDevice will attempt to use the device's MAC address to look up the credentials
// from the Secret Store. If a mapping does not exist, or the device's MAC address is missing or invalid,
// the default secret path will be used to look up the credentials. An error is returned if the secret path
// does not exist in the Secret Store.
func (d *Driver) tryGetCredentialsForDevice(device models.Device) (Credentials, errors.EdgeX) {
	d.configMu.RLock()
	defaultSecretName := d.config.AppCustom.DefaultSecretName
	d.configMu.RUnlock()

	secretName := defaultSecretName
	if mac := device.Protocols[OnvifProtocol][MACAddress]; mac != "" {
		secretName = d.macAddressMapper.TryGetSecretNameForMACAddress(mac, defaultSecretName)
	} else {
		d.lc.Warnf("Device %s is missing MAC Address, using default secret name", device.Name)
	}

	credentials, edgexErr := d.tryGetCredentials(secretName)
	if edgexErr != nil {
		d.lc.Errorf("Failed to retrieve credentials for the secret name %s: %s", secretName, edgexErr.Error())
		return Credentials{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to get credentials", edgexErr)
	}

	d.lc.Debugf("Found credentials from secret name %s for device %s", secretName, device.Name)

	return credentials, nil
}
