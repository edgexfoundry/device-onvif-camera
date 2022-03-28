//
// Copyright (C) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"sync/atomic"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/pkg/errors"
)

type MockSDKService struct {
	devices map[string]models.Device
	added   uint32
	Config  map[string]string
}

func NewMockSdkService() *MockSDKService {
	return &MockSDKService{
		devices: make(map[string]models.Device),
	}
}

func (s *MockSDKService) clearDevices() {
	s.devices = make(map[string]models.Device)
	s.resetAddedCount()
}

func (s *MockSDKService) resetAddedCount() {
	atomic.StoreUint32(&s.added, 0)
}

func (s *MockSDKService) Devices() []models.Device {
	devices := make([]models.Device, len(s.devices))
	for _, v := range s.devices {
		devices = append(devices, v)
	}
	return devices
}

func (s *MockSDKService) AddDiscoveredDevices(discovered []dsModels.DiscoveredDevice) {
	for _, d := range discovered {
		_, _ = s.AddDevice(models.Device{
			Name:        d.Name,
			Protocols:   d.Protocols,
			ProfileName: "LLRP-Device-Profile",
		})
	}
}

func (s *MockSDKService) AddDevice(device models.Device) (id string, err error) {
	if device.Id == "" {
		device.Id = device.Name
	}
	s.devices[device.Name] = device
	atomic.AddUint32(&s.added, 1)
	return device.Id, nil
}

func (s *MockSDKService) DriverConfigs() map[string]string {
	return s.Config
}

func (s *MockSDKService) GetDeviceByName(name string) (models.Device, error) {
	device, ok := s.devices[name]
	if ok {
		return device, nil
	}
	return models.Device{}, fmt.Errorf("device %s was not found", name)
}

func (s *MockSDKService) UpdateDevice(device models.Device) error {
	s.devices[device.Name] = device
	return nil
}

func (s *MockSDKService) UpdateDeviceOperatingState(deviceName string, state string) error {
	if d, ok := s.devices[deviceName]; ok {
		d.OperatingState = models.OperatingState(state)
		return nil
	}
	return fmt.Errorf("device with name %s not found", deviceName)
}

func (s *MockSDKService) SetDeviceOpState(_ string, _ models.OperatingState) error {
	return errors.New("Method not implemented.")
}

func (s *MockSDKService) GetProvisionWatcherByName(_ string) (models.ProvisionWatcher, error) {
	return models.ProvisionWatcher{}, errors.New("Method not implemented.")
}

func (s *MockSDKService) AddProvisionWatcher(_ models.ProvisionWatcher) (id string, err error) {
	return "", errors.New("Method not implemented.")
}

func (s *MockSDKService) LoadCustomConfig(customConfig service.UpdatableConfig, sectionName string) error {
	return nil
}

func (s *MockSDKService) ListenForCustomConfigChanges(configToWatch interface{}, sectionName string, changedCallback func(interface{})) error {
	return nil
}
