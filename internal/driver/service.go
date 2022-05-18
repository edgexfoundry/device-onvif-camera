//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// ServiceWrapper wraps an EdgeX SDK service so it can be easily mocked in tests.
type ServiceWrapper interface {
	Devices() []contract.Device
	GetDeviceByName(name string) (contract.Device, error)
	UpdateDevice(device contract.Device) error
	UpdateDeviceOperatingState(deviceName string, state string) error
	GetProvisionWatcherByName(name string) (contract.ProvisionWatcher, error)
	AddProvisionWatcher(watcher contract.ProvisionWatcher) (id string, err error)
	AddDevice(device contract.Device) (id string, err error)
	LoadCustomConfig(customConfig service.UpdatableConfig, sectionName string) error
	ListenForCustomConfigChanges(configToWatch interface{}, sectionName string, changedCallback func(interface{})) error

	DriverConfigs() map[string]string
}

type DeviceSDKService struct {
	*service.DeviceService
	lc logger.LoggingClient
}

// DriverConfigs retrieves the driver specific configuration
func (s *DeviceSDKService) DriverConfigs() map[string]string {
	return service.DriverConfigs()
}
