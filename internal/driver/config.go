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

type configuration struct {
	CredentialsRetryTime int
	CredentialsRetryWait int
	RequestTimeout       int
	// DefaultAuthMode indicates the Onvif camera default auth mode. "digest" | "usernametoken" | "both" | "none"
	DefaultAuthMode string
	// DefaultSecretPath indicates the secret path to retrieve username and password from secret store.
	DefaultSecretPath string
	// DiscoveryEthernetInterface indicates the target EthernetInterface for discovering. The default value is `en0`, the user can modify it to meet their requirement.
	DiscoveryEthernetInterface string
	// BaseNotificationURL indicates the device service network location
	BaseNotificationURL string

	DiscoverySubnets           string
	ProbeAsyncLimit            int
	ProbeTimeoutMillis         int
	ScanPorts                  string
	MaxDiscoverDurationSeconds int
}

// CameraInfo holds the camera connection info
type CameraInfo struct {
	Address    string
	Port       int
	AuthMode   string
	SecretPath string
}

// loadCameraConfig loads the camera configuration
func loadCameraConfig(configMap map[string]string) (*configuration, error) {
	config := new(configuration)
	err := load(configMap, config)
	if err != nil {
		return config, err
	}
	return config, nil
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
