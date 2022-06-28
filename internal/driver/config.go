// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// CustomConfig holds the values for the driver configuration
type CustomConfig struct {
	CredentialsRetryTime int
	CredentialsRetryWait int
	RequestTimeout       int
	// DefaultSecretPath indicates the secret path to retrieve username and password from secret store.
	DefaultSecretPath string
	// DiscoveryEthernetInterface indicates the target EthernetInterface for discovering. The default value is `en0`, the user can modify it to meet their requirement.
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

	// ProvisionWatcherDir is the location of Provision Watchers
	ProvisionWatcherDir string

	// CredentialsMap is a map of SecretPath -> Comma separated list of mac addresses
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

// CameraInfo holds the camera connection info
type CameraInfo struct {
	Address    string
	Port       int
	SecretPath string
}

// CreateCameraInfo creates new CameraInfo entity from the protocol properties
func CreateCameraInfo(protocols map[string]models.ProtocolProperties) (*CameraInfo, errors.EdgeX) {
	info := new(CameraInfo)
	protocol, ok := protocols[OnvifProtocol]
	if !ok {
		return info, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unable to load config, Protocol '%s' not exist", OnvifProtocol), nil)
	}

	if err := load(protocol, info); err != nil {
		return info, errors.NewCommonEdgeXWrapper(err)
	}

	return info, nil
}

// load by reflect to check map key and then fetch the value
func load(config map[string]string, des interface{}) error {
	errorMessage := "unable to load config, '%s' not exist"
	val := reflect.ValueOf(des).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		valueField := val.Field(i)

		val, ok := config[typeField.Name]
		if !ok {
			return fmt.Errorf(errorMessage, typeField.Name)
		}

		switch valueField.Kind() {
		case reflect.Int:
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			valueField.SetInt(int64(intVal))
		case reflect.String:
			valueField.SetString(val)
		default:
			return fmt.Errorf("none supported value type %v ,%v", valueField.Kind(), typeField.Name)
		}
	}
	return nil
}
