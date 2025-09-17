// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/spf13/cast"
)

// CustomConfig holds the values for the driver configuration
type CustomConfig struct {
	// RequestTimeout is the number of seconds to wait when making an Onvif request before timing out.
	RequestTimeout int
	// DefaultSecretName indicates the secret name to retrieve username and password from secret store.
	DefaultSecretName string
	// DiscoveryEthernetInterface indicates the target EthernetInterface for multicast discovering.
	DiscoveryEthernetInterface string
	// BaseNotificationURL indicates the device service network location
	BaseNotificationURL string

	// DiscoveryMode indicates mode used to discovery devices on the network.
	DiscoveryMode DiscoveryMode
	// DiscoverySubnets indicates the network segments used when discovery is scanning for devices.
	DiscoverySubnets string
	// ProbeAsyncLimit indicates the maximum number of simultaneous network probes.
	ProbeAsyncLimit int
	// ProbeTimeoutMillis indicates the maximum amount of milliseconds to wait for each IP probe before timing out.
	ProbeTimeoutMillis int
	// MaxDiscoverDurationSeconds indicates the amount of seconds discovery will run before timing out.
	MaxDiscoverDurationSeconds int

	// EnableStatusCheck indicates if status checking should be enabled
	EnableStatusCheck bool
	// CheckStatusInterval indicates the interval in seconds at which the device service will check device statuses
	CheckStatusInterval int

	// CredentialsMap is a map of SecretName -> Comma separated list of mac addresses
	CredentialsMap map[string]string
}

// ServiceConfig a struct that wraps CustomConfig which holds the values for driver configuration
type ServiceConfig struct {
	AppCustom CustomConfig
}

// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false
	}

	*c = *configuration

	return true
}

// GetCameraXAddr returns the Address:Port of the camera from the Onvif protocol properties
func GetCameraXAddr(protocols map[string]models.ProtocolProperties) (string, errors.EdgeX) {
	protocol, ok := protocols[OnvifProtocol]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unable to load config, Protocol '%s' not exist", OnvifProtocol), nil)
	}

	address := ""
	if v, ok := protocol[Address]; ok {
		address = cast.ToString(v)
	}
	if address == "" {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unable to load XAddr, %s Address does not exist", OnvifProtocol), nil)
	}
	port := ""
	if v, ok := protocol[Port]; ok {
		port = cast.ToString(v)
	}

	xAddr := address
	if port != "" {
		xAddr += ":" + port
	}

	return xAddr, nil
}
