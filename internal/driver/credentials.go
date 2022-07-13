// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"strings"
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
	// noAuthSecretPath is the magic string used to define a group which does not use credentials
	// this is defined in lowercase and compared in lowercase
	noAuthSecretPath = "noauth"
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

// tryGetCredentials will attempt one time to get the credentials located at secretPath from
// secret provider and return them, otherwise return an error.
func (d *Driver) tryGetCredentials(secretPath string) (Credentials, errors.EdgeX) {
	// if the secret path is the special NoAuth magic key, do not look it up, instead return the noAuthCredentials
	// todo: add unit tests for noAuth magic key
	if strings.ToLower(secretPath) == noAuthSecretPath {
		return noAuthCredentials, nil
	}

	secretData, err := d.sdkService.GetSecretProvider().GetSecret(secretPath, UsernameKey, PasswordKey, AuthModeKey)
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
	defaultSecretPath := d.config.AppCustom.DefaultSecretPath
	d.configMu.RUnlock()

	secretPath := defaultSecretPath
	if mac, hasMAC := device.Protocols[OnvifProtocol][MACAddress]; hasMAC {
		secretPath = d.macAddressMapper.TryGetSecretPathForMACAddress(mac, defaultSecretPath)
	} else {
		d.lc.Warnf("Device %s is missing MAC Address, using default secret path", device.Name)
	}

	credentials, edgexErr := d.tryGetCredentials(secretPath)
	if edgexErr != nil {
		d.lc.Errorf("Failed to retrieve credentials for the secret path %s: %s", secretPath, edgexErr.Error())
		return Credentials{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to get credentials", edgexErr)
	}

	d.lc.Infof("Found credentials for device %s", device.Name)

	return credentials, nil
}
