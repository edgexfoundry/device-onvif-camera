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

// tryGetCredentialsFromMac will attempt to retrieve the credentials associated with the given mac address.
func (d *Driver) tryGetCredentialsFromMac(mac string) (Credentials, errors.EdgeX) {
	if mac == "" {
		credential, edgexErr := tryGetCredentials(d.config.AppCustom.DefaultSecretPath)
		if edgexErr != nil {
			d.lc.Error("failed to get credentials from default secret path", "err", edgexErr)
			return Credentials{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to get default credentials for empty mac address", edgexErr)
		}

		d.lc.Debug("Using default credentials from default secret path for empty mac address")
		return credential, nil
	}

	credentials, edgexErr := d.macAddressMapper.TryGetCredentialsForMACAddress(mac)
	if edgexErr != nil {
		d.lc.Errorf("failed to get credentials for mac %s in lookup table", mac)

		credential, edgexErr := tryGetCredentials(d.config.AppCustom.DefaultSecretPath)
		if edgexErr != nil {
			d.lc.Error("failed to get credentials from default secret path", "err", edgexErr)
			return Credentials{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to get default credentials", edgexErr)
		}

		d.lc.Debug("Using credentials from default secret path for mac address not in lookup table")
		return credential, nil
	}

	d.lc.Debugf("Using credentials for mac %s", mac)
	return credentials, nil
}
