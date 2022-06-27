// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/IOTechSystems/onvif"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
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
func tryGetCredentials(secretPath string) (Credentials, errors.EdgeX) {
	secretData, err := sdk.RunningService().SecretProvider.GetSecret(secretPath, UsernameKey, PasswordKey, AuthModeKey)
	if err != nil {
		return Credentials{}, errors.NewCommonEdgeXWrapper(err)
	}

	credentials := Credentials{
		Username: secretData[UsernameKey],
		Password: secretData[PasswordKey],
		AuthMode: secretData[AuthModeKey],
	}

	if !IsAuthModeValid(secretData[AuthModeKey]) {
		sdk.RunningService().LoggingClient.Warnf("AuthMode is set to an invalid value: %s. setting value to 'usernametoken'.", credentials.AuthMode)
		credentials.AuthMode = AuthModeUsernameToken
	}

	return credentials, nil
}

// getCredentials will repeatedly try and get the credentials located at secretPath from
// secret provider every CredentialsRetryTime seconds for a maximum of CredentialsRetryWait seconds.
// Note that this function will block until either the credentials are found, or CredentialsRetryWait
// seconds have elapsed.
func (d *Driver) getCredentials(secretPath string) (credentials Credentials, err errors.EdgeX) {
	d.configMu.RLock()
	timer := startup.NewTimer(d.config.AppCustom.CredentialsRetryTime, d.config.AppCustom.CredentialsRetryWait)
	d.configMu.RUnlock()

	for timer.HasNotElapsed() {
		if credentials, err = tryGetCredentials(secretPath); err == nil {
			return credentials, nil
		}

		d.lc.Warnf(
			"Unable to retrieve camera credentials from SecretProvider at path '%s': %s. Retrying for %s",
			secretPath,
			err.Error(),
			timer.RemainingAsString())
		timer.SleepForInterval()
	}

	return credentials, err
}

// tryGetCredentialsForDevice will attempt to use the device's MAC address to look up the credentials
// from the Secret Store. If a mapping does not exist, or the device's MAC address is missing or invalid,
// the default secret path will be used to look up the credentials. An error is returned if the secret path
// does not exist in the Secret Store.
// todo: remove nolint once function is used.
//nolint:golint,unused
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

	credentials, edgexErr := tryGetCredentials(secretPath)
	if edgexErr != nil {
		d.lc.Errorf("Failed to retrieve credentials for the secret path %s: %s", secretPath, edgexErr.Error())
		return Credentials{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to get credentials", edgexErr)
	}

	return credentials, nil
}
