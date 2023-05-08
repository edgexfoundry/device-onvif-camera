// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
// Copyright (c) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"

	"github.com/edgexfoundry/device-onvif-camera/internal/netscan"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	onvifdevice "github.com/IOTechSystems/onvif/device"
	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
)

const (
	URLRawQuery = "urlRawQuery"
	jsonObject  = "jsonObject"

	wsDiscoveryPort = "3702"
	// discoverDebounceDuration is the amount of time to wait for additional changes to discover
	// configuration before auto-triggering a discovery
	discoverDebounceDuration = 10 * time.Second
)

// Driver implements the sdkModel.ProtocolDriver interface for
// the device service
type Driver struct {
	lc         logger.LoggingClient
	sdkService interfaces.DeviceServiceSDK

	onvifClients map[string]*OnvifClient
	clientsMu    sync.RWMutex

	config   *ServiceConfig
	configMu sync.RWMutex

	macAddressMapper *MACAddressMapper

	// debounceTimer and debounceMu keep track of when to fire a debounced discovery call
	debounceTimer *time.Timer
	debounceMu    sync.Mutex

	// taskCh is used to send signals to the taskLoop
	taskCh chan struct{}
	wg     sync.WaitGroup
}

func NewDriver() *Driver {
	return &Driver{
		onvifClients: make(map[string]*OnvifClient),
		config:       &ServiceConfig{},
		taskCh:       make(chan struct{}),
	}
}

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	d.sdkService = sdk
	d.lc = sdk.LoggingClient()
	d.macAddressMapper = NewMACAddressMapper(sdk)

	err := d.sdkService.LoadCustomConfig(d.config, "AppCustom")
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "custom driver configuration failed to load", err)
	}

	d.lc.Debugf("Custom config is : %+v", d.config)

	if !d.config.AppCustom.DiscoveryMode.IsValid() {
		d.lc.Errorf("DiscoveryMode is set to an invalid value: %q. Discovery will be unable to be performed.",
			d.config.AppCustom.DiscoveryMode)
	}

	err = d.sdkService.SecretProvider().RegisterSecretUpdatedCallback(secret.WildcardName, d.secretUpdated)
	if err != nil {
		d.lc.Errorf("failed to register secret update callback: %v", err)
	}

	d.macAddressMapper.UpdateMappings(d.config.AppCustom.CredentialsMap)

	err = d.sdkService.ListenForCustomConfigChanges(&d.config.AppCustom, "AppCustom", d.updateWritableConfig)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to listen to custom config changes", err)
	}

	handler := NewRestNotificationHandler(d.sdkService)
	edgexErr := handler.AddRoute()
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	d.lc.Info("Driver initialized.")
	return nil
}

// Start is called after the device sdk is fully initialized. This function creates connections to all the cameras
// and checks their statuses. It then runs the task loop if enabled.
func (d *Driver) Start() error {
	wg := sync.WaitGroup{}
	for _, device := range d.sdkService.Devices() {
		device := device
		wg.Add(1)
		go func() {
			defer wg.Done()

			d.lc.Infof("Initializing onvif client for '%s' camera", device.Name)
			_, err := d.getOrCreateOnvifClient(device)
			if err != nil {
				d.lc.Errorf("failed to initialize onvif client for '%s' camera, skipping this device.", device.Name)
				return
			}
			d.checkStatusOfDevice(device)
		}()
	}
	wg.Wait()

	d.configMu.RLock()
	enableStatusCheck := d.config.AppCustom.EnableStatusCheck
	d.configMu.RUnlock()

	if enableStatusCheck {
		// starts loop to check connection and determine device status
		d.wg.Add(1)
		go func() {
			defer d.wg.Done() // wait for taskLoop to return
			d.taskLoop()
			d.lc.Info("taskLoop has stopped.")
		}()
	}

	d.lc.Info("Driver started.")
	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (d *Driver) Stop(force bool) error {
	if d.sdkService.AsyncValuesChannel() != nil {
		close(d.sdkService.AsyncValuesChannel())
	}

	d.clientsMu.Lock()
	for _, client := range d.onvifClients {
		client.pullPointManager.UnsubscribeAll()
		client.baseNotificationManager.UnsubscribeAll()
	}
	d.onvifClients = make(map[string]*OnvifClient)
	d.clientsMu.Unlock()

	close(d.taskCh) // send signal for taskLoop to finish
	d.wg.Wait()     // wait for taskLoop goroutine to return

	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	_, err := d.getOrCreateOnvifClient(models.Device{Name: deviceName, Protocols: protocols})
	if err != nil {
		d.lc.Errorf("Failed to initialize onvif client for camera '%s'", deviceName)
		return errors.NewCommonEdgeXWrapper(err)
	}
	// check the status of the newly added device
	d.checkStatusOfDevice(models.Device{Name: deviceName, Protocols: protocols})
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Invoke the updateOnvifClient func to update the old onvif client if needed
	err := d.updateOnvifClient(models.Device{Name: deviceName, Protocols: protocols})
	if err != nil {
		d.lc.Errorf("Unable to update onvif device client for device %s, %v", deviceName, err)
	}
	return err
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.removeOnvifClient(deviceName)
	return nil
}

// ValidateDevice is called by core-metadata service anytime a device is added or updated.
func (d *Driver) ValidateDevice(device models.Device) error {
	_, err := GetCameraXAddr(device.Protocols)
	if err != nil {
		return fmt.Errorf("invalid protocol properties, %v", err)
	}
	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	var edgexErr errors.EdgeX
	var responses = make([]*sdkModel.CommandValue, len(reqs))

	onvifClient, edgexErr := d.getOrCreateOnvifClient(models.Device{Name: deviceName, Protocols: protocols})
	if edgexErr != nil {
		d.lc.Errorf("Failed to retrieve onvif client for camera '%s'", deviceName)
		return responses, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	for i, req := range reqs {
		data, edgexErr := parametersFromURLRawQuery(req)
		if edgexErr != nil {
			return responses, errors.NewCommonEdgeXWrapper(edgexErr)
		}

		cv, edgexErr := onvifClient.CallOnvifFunction(req, GetFunction, data)
		if edgexErr != nil {
			return responses, errors.NewCommonEdgeX(errors.KindServerError, "failed to execute read command", edgexErr)
		}
		responses[i] = cv
	}

	return responses, nil
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource (aka DeviceObject).
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	var edgexErr errors.EdgeX

	onvifClient, edgexErr := d.getOrCreateOnvifClient(models.Device{Name: deviceName, Protocols: protocols})
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	for i, req := range reqs {
		var data []byte

		switch req.Type {
		case common.ValueTypeString:
			str, err := params[i].StringValue()
			if err != nil {
				return errors.NewCommonEdgeXWrapper(err)
			}
			data = []byte(str)
		case common.ValueTypeObject:
			parameters, err := params[i].ObjectValue()
			if err != nil {
				return errors.NewCommonEdgeXWrapper(err)
			}
			data, err = json.Marshal(parameters)
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to marshal set command parameter for resource '%s'", req.DeviceResourceName), err)
			}
		default:
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("value type %s is not supported for write commands", req.Type), nil)
		}

		result, err := onvifClient.CallOnvifFunction(req, SetFunction, data)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to execute write command, %s", result), err)
		}
	}

	return nil
}

// Discover performs a discovery on the network and passes them to EdgeX to get provisioned
func (d *Driver) Discover() error {
	d.lc.Info("Discover was called.")

	d.configMu.RLock()
	maxSeconds := d.config.AppCustom.MaxDiscoverDurationSeconds
	discoveryMode := d.config.AppCustom.DiscoveryMode
	d.configMu.RUnlock()

	if !discoveryMode.IsValid() {
		return fmt.Errorf("DiscoveryMode is set to an invalid value: %s. Refusing to do discovery", discoveryMode)
	}

	var discoveredDevices []sdkModel.DiscoveredDevice

	if discoveryMode.IsMulticastEnabled() {
		discoveredDevices = append(discoveredDevices, d.discoverMulticast()...)
	}

	if discoveryMode.IsNetScanEnabled() {
		ctx := context.Background()
		if maxSeconds > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(),
				time.Duration(maxSeconds)*time.Second)
			defer cancel()
		}
		discoveredDevices = append(discoveredDevices, d.discoverNetscan(ctx)...)
	}

	// pass the discovered devices to the EdgeX SDK to be passed through to the provision watchers
	filtered := d.discoverFilter(discoveredDevices)
	d.sdkService.DiscoveredDeviceChannel() <- filtered
	return nil
}

// multicast enable/disable via config option
func (d *Driver) discoverMulticast() []sdkModel.DiscoveredDevice {
	var discovered []sdkModel.DiscoveredDevice

	d.configMu.RLock()
	discoveryEthernetInterface := d.config.AppCustom.DiscoveryEthernetInterface
	d.configMu.RUnlock()

	t0 := time.Now()
	onvifDevices := wsdiscovery.GetAvailableDevicesAtSpecificEthernetInterface(discoveryEthernetInterface)
	d.lc.Infof("Discovered %d device(s) in %v via multicast.", len(onvifDevices), time.Since(t0))
	for _, onvifDevice := range onvifDevices {
		device, err := d.createDiscoveredDevice(onvifDevice)
		if err != nil {
			d.lc.Warnf(err.Error())
			continue
		}
		discovered = append(discovered, device)
	}

	return discovered
}

// netscan enable/disable via config option
func (d *Driver) discoverNetscan(ctx context.Context) []sdkModel.DiscoveredDevice {
	var discovered []sdkModel.DiscoveredDevice

	if len(strings.TrimSpace(d.config.AppCustom.DiscoverySubnets)) == 0 {
		d.lc.Warn("netscan discovery was called, but DiscoverySubnets are empty!")
		return nil
	}

	d.configMu.RLock()
	params := netscan.Params{
		// split the comma separated string here to avoid issues with EdgeX's Consul implementation
		Subnets:         strings.Split(d.config.AppCustom.DiscoverySubnets, ","),
		AsyncLimit:      d.config.AppCustom.ProbeAsyncLimit,
		Timeout:         time.Duration(d.config.AppCustom.ProbeTimeoutMillis) * time.Millisecond,
		ScanPorts:       []string{wsDiscoveryPort},
		Logger:          d.lc,
		NetworkProtocol: netscan.NetworkUDP,
	}
	d.configMu.RUnlock()

	t0 := time.Now()
	result := netscan.AutoDiscover(ctx, NewOnvifProtocolDiscovery(d), params)
	if ctx.Err() != nil {
		d.lc.Warnf("Discover process has been cancelled!", "ctxErr", ctx.Err())
	}

	d.lc.Debugf("NetScan result: %+v", result)
	d.lc.Infof("Discovered %d device(s) in %v via netscan.", len(result), time.Since(t0))

	discovered = append(discovered, result...)
	return discovered
}

// debouncedDiscover adds or updates a future call to Discover. This function is intended to be
// called by the config watcher in response to any configuration changes related to discovery.
// The reason Discover calls are being debounced is to allow the user to make multiple changes to
// their configuration, and only fire the discovery once.
//
// The way it works is that this code creates and starts a timer for discoverDebounceDuration.
// Every subsequent call to this function before that timer elapses resets the timer to
// discoverDebounceDuration. Once the timer finally elapses, the Discover function is called.
func (d *Driver) debouncedDiscover() {
	d.lc.Debugf("trigger debounced discovery in %v", discoverDebounceDuration)

	// everything in this function is mutex-locked and is safe to access asynchronously
	d.debounceMu.Lock()
	defer d.debounceMu.Unlock()

	if d.debounceTimer != nil {
		// timer is already active, so reset it (debounce)
		d.debounceTimer.Reset(discoverDebounceDuration)
	} else {
		// no timer is active, so create and start a new one
		d.debounceTimer = time.NewTimer(discoverDebounceDuration)

		// asynchronously listen for the timer to elapse. this go routine will only ever be run
		// once due to mutex locking and the above if statement.
		go func() {
			// wait for timer to tick
			<-d.debounceTimer.C

			// remove timer. we must lock the mutex as this go routine runs separately from the
			// outer function's locked scope
			d.debounceMu.Lock()
			d.debounceTimer = nil
			d.debounceMu.Unlock()

			err := d.Discover()
			if err != nil {
				d.lc.Errorf("failed to run device discovery, %v", err)
			}
		}()
	}
}

func (d *Driver) updateWritableConfig(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*CustomConfig)
	if !ok {
		d.lc.Errorf("Unable to update writable custom config: Cannot cast raw config of type %T into type 'CustomConfig'",
			rawWritableConfig)
		return
	}

	d.configMu.Lock()
	oldSubnets := d.config.AppCustom.DiscoverySubnets
	d.config.AppCustom = *updated
	d.configMu.Unlock()

	if updated.DiscoverySubnets != oldSubnets {
		d.lc.Info("Discover configuration has changed! Discovery will be triggered momentarily.")
		d.debouncedDiscover()
	}

	d.macAddressMapper.UpdateMappings(d.config.AppCustom.CredentialsMap)
	// check device statuses in case the credentials map was updated
	d.checkStatuses()
}

// refreshDevice will attempt to retrieve the MAC address and the device info for the specified camera
// and update the values in the protocol properties
// Also the device name is updated if the name starts with the UnknownDevicePrefix and the status is UpWithAuth
func (d *Driver) refreshDevice(device models.Device) error {
	onvifClient, edgexErr := d.getOrCreateOnvifClient(device)
	if edgexErr != nil {
		return edgexErr
	}

	// save the MAC Address in case it was changed by the calling code
	hwAddress := device.Protocols[OnvifProtocol][MACAddress]

	devInfo, err := onvifClient.getDeviceInformation(device)
	if err != nil {
		return err
	}

	netInfo, netErr := onvifClient.getNetworkInterfaces(device)
	if netErr != nil {
		d.lc.Warnf("Error trying to get network interfaces for device %s: %s", device.Name, netErr.Error())
	}

	endpointRef, endpointErr := onvifClient.getEndpointReference(device)
	if endpointErr != nil {
		d.lc.Warnf("Error trying to get get endpoint reference for device %s: %s", device.Name, endpointErr.Error())
	}

	// update device to latest version in cache to prevent race conditions and ensure we have all associated metadata
	device, edgeXErr := d.sdkService.GetDeviceByName(device.Name)
	if err != nil {
		return edgeXErr
	}

	isChanged := false

	if netErr == nil { // only update if there was no error querying the net info
		hwAddress = string(netInfo.NetworkInterfaces.Info.HwAddress)
	}
	if hwAddress != device.Protocols[OnvifProtocol][MACAddress] {
		device.Protocols[OnvifProtocol][MACAddress] = hwAddress
		isChanged = true
	}

	if endpointErr == nil { // only update if there was no error querying the endpoint ref address
		uuidElements := strings.Split(endpointRef.GUID, ":")
		device.Protocols[OnvifProtocol][EndpointRefAddress] = uuidElements[len(uuidElements)-1]
		isChanged = true
	}

	if devInfo.Manufacturer != device.Protocols[OnvifProtocol][Manufacturer] ||
		devInfo.Model != device.Protocols[OnvifProtocol][Model] ||
		devInfo.FirmwareVersion != device.Protocols[OnvifProtocol][FirmwareVersion] ||
		devInfo.SerialNumber != device.Protocols[OnvifProtocol][SerialNumber] ||
		devInfo.HardwareId != device.Protocols[OnvifProtocol][HardwareId] {

		device.Protocols[OnvifProtocol][Manufacturer] = devInfo.Manufacturer
		device.Protocols[OnvifProtocol][Model] = devInfo.Model
		device.Protocols[OnvifProtocol][FirmwareVersion] = devInfo.FirmwareVersion
		device.Protocols[OnvifProtocol][SerialNumber] = devInfo.SerialNumber
		device.Protocols[OnvifProtocol][HardwareId] = devInfo.HardwareId
		isChanged = true
	}

	if device.Protocols[OnvifProtocol][FriendlyName] == "" { // initialize the friendly name if it is blank
		device.Protocols[OnvifProtocol][FriendlyName] = devInfo.Manufacturer + " " + devInfo.Model
	}

	if isChanged {
		return d.updateDevice(device, devInfo)
	}

	return nil
}

func (d *Driver) updateDevice(device models.Device, deviceInfo *onvifdevice.GetDeviceInformationResponse) error {
	if strings.HasPrefix(device.Name, UnknownDevicePrefix) {
		d.lc.Infof("Removing device '%s' to update device with the updated name", device.Name)
		err := d.sdkService.RemoveDeviceByName(device.Name)
		if err != nil {
			d.lc.Warnf("An error occurred while removing the device %s: %s",
				device.Name, err)
		}

		device.Id = ""
		// Spaces are not allowed in the device name
		device.Name = fmt.Sprintf("%s-%s-%s",
			strings.ReplaceAll(deviceInfo.Manufacturer, " ", "-"),
			strings.ReplaceAll(deviceInfo.Model, " ", "-"),
			device.Protocols[OnvifProtocol][EndpointRefAddress])
		d.lc.Infof("Adding device back with the updated name '%s'", device.Name)
		_, err = d.sdkService.AddDevice(device)
		return err
	}

	return d.sdkService.UpdateDevice(device)
}
