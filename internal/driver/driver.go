// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"

	"github.com/edgexfoundry/device-onvif-camera/internal/netscan"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/IOTechSystems/onvif"
	onvifdevice "github.com/IOTechSystems/onvif/device"
	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
)

const (
	URLRawQuery = "urlRawQuery"
	jsonObject  = "jsonObject"

	cameraAdded   = "CameraAdded"
	cameraUpdated = "CameraUpdated"
	cameraDeleted = "CameraDeleted"

	wsDiscoveryPort = "3702"

	// enable this by default, otherwise discovery will not work.
	registerProvisionWatchers = true

	// discoverDebounceDuration is the amount of time to wait for additional changes to discover
	// configuration before auto-triggering a discovery
	discoverDebounceDuration = 10 * time.Second
)

// Driver implements the sdkModel.ProtocolDriver interface for
// the device service
type Driver struct {
	lc       logger.LoggingClient
	asynchCh chan<- *sdkModel.AsyncValues
	deviceCh chan<- []sdkModel.DiscoveredDevice

	sdkService SDKService

	onvifClients map[string]*OnvifClient
	clientsMu    *sync.RWMutex

	config   *ServiceConfig
	configMu *sync.RWMutex

	addedWatchers bool
	watchersMu    sync.Mutex

	macAddressMapper *MACAddressMapper

	// debounceTimer and debounceMu keep track of when to fire a debounced discovery call
	debounceTimer *time.Timer
	debounceMu    sync.Mutex

	// taskCh is used to send signals to the taskLoop
	taskCh chan struct{}
	wg     sync.WaitGroup
}

type MultiErr []error

//goland:noinspection GoReceiverNames
func (me MultiErr) Error() string {
	strs := make([]string, len(me))
	for i, s := range me {
		strs[i] = s.Error()
	}

	return strings.Join(strs, "; ")
}

// EdgeX's Device SDK takes an interface{}
// and uses a runtime-check to determine that it implements ProtocolDriver,
// at which point it will abruptly exit without a panic.
// This restores type-safety by making it so that we can't compile
// unless we meet the runtime-required interface.
var _ sdkModel.ProtocolDriver = (*Driver)(nil)

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues,
	deviceCh chan<- []sdkModel.DiscoveredDevice) error {
	d.lc = lc
	d.asynchCh = asyncCh
	d.deviceCh = deviceCh
	d.taskCh = make(chan struct{})
	d.clientsMu = new(sync.RWMutex)
	d.configMu = new(sync.RWMutex)
	d.onvifClients = make(map[string]*OnvifClient)
	d.sdkService = &DeviceSDKService{
		DeviceService: service.RunningService(),
	}
	d.macAddressMapper = NewMACAddressMapper(d.sdkService)
	d.config = &ServiceConfig{}

	err := d.sdkService.LoadCustomConfig(d.config, "AppCustom")
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "custom driver configuration failed to load", err)
	}

	lc.Debugf("Custom config is : %+v", d.config)

	if !d.config.AppCustom.DiscoveryMode.IsValid() {
		d.lc.Errorf("DiscoveryMode is set to an invalid value: %q. Discovery will be unable to be performed.",
			d.config.AppCustom.DiscoveryMode)
	}

	d.macAddressMapper.UpdateMappings(d.config.AppCustom.CredentialsMap)

	err = d.sdkService.ListenForCustomConfigChanges(&d.config.AppCustom, "AppCustom", d.updateWritableConfig)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to listen to custom config changes", err)
	}

	for _, device := range d.sdkService.Devices() {
		// onvif client should not be created for the control-plane device
		if device.Name == d.sdkService.Name() {
			continue
		}

		d.lc.Infof("Initializing onvif client for '%s' camera", device.Name)

		onvifClient, err := d.newOnvifClient(device)
		if err != nil {
			d.lc.Errorf("failed to initialize onvif client for '%s' camera, skipping this device.", device.Name)
			continue
		}
		d.clientsMu.Lock()
		d.onvifClients[device.Name] = onvifClient
		d.clientsMu.Unlock()
	}

	handler := NewRestNotificationHandler(d.sdkService, lc, asyncCh)
	edgexErr := handler.AddRoute()
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

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

	d.lc.Info("Driver initialized.")
	return nil
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

			d.Discover()
		}()
	}
}

// todo: remove this method once the Device SDK has been updated as per https://github.com/edgexfoundry/device-sdk-go/issues/1100
func (d *Driver) addProvisionWatchers() error {

	// this setting is a workaround for the fact that there is no standard way to define this directory using the SDK
	// the snap needs to be able to change the location of the provision watchers
	d.configMu.RLock()
	provisionWatcherFolder := d.config.AppCustom.ProvisionWatcherDir
	d.configMu.RUnlock()
	if provisionWatcherFolder == "" {
		provisionWatcherFolder = "res/provision_watchers"
	}
	d.lc.Infof("Adding provision watchers from %s", provisionWatcherFolder)

	files, err := ioutil.ReadDir(provisionWatcherFolder)
	if err != nil {
		return err
	}

	d.lc.Debugf("%d provision watcher files found", len(files))

	var errs []error
	for _, file := range files {
		filename := filepath.Join(provisionWatcherFolder, file.Name())
		d.lc.Debugf("processing %s", filename)
		var watcher dtos.ProvisionWatcher
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			errs = append(errs, errors.NewCommonEdgeX(errors.KindServerError, "error reading file "+filename, err))
			continue
		}

		if err := json.Unmarshal(data, &watcher); err != nil {
			errs = append(errs, errors.NewCommonEdgeX(errors.KindServerError, "error unmarshalling provision watcher "+filename, err))
			continue
		}

		err = common.Validate(watcher)
		if err != nil {
			errs = append(errs, errors.NewCommonEdgeX(errors.KindServerError, "provision watcher validation failed "+filename, err))
			continue
		}

		if _, err := d.sdkService.GetProvisionWatcherByName(watcher.Name); err == nil {
			d.lc.Debugf("skip existing provision watcher %s", watcher.Name)
			continue // provision watcher already exists
		}

		watcherModel := dtos.ToProvisionWatcherModel(watcher)

		d.lc.Infof("Adding provision watcher:%s", watcherModel.Name)
		id, err := d.sdkService.AddProvisionWatcher(watcherModel)
		if err != nil {
			errs = append(errs, errors.NewCommonEdgeX(errors.KindServerError, "error adding provision watcher "+watcherModel.Name, err))
			continue
		}
		d.lc.Infof("Successfully added provision watcher: %s,  ID: %s", watcherModel.Name, id)
	}

	if errs != nil {
		return MultiErr(errs)
	}
	return nil
}

func (d *Driver) getOnvifClient(deviceName string) (*OnvifClient, errors.EdgeX) {
	d.clientsMu.RLock()
	defer d.clientsMu.RUnlock()
	onvifClient, ok := d.onvifClients[deviceName]
	if !ok {
		device, err := d.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		onvifClient, err = d.newOnvifClient(device)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initialize onvif client for '%s' camera", device.Name), err)
		}
		d.onvifClients[deviceName] = onvifClient
	}
	return onvifClient, nil
}

func (d *Driver) removeOnvifClient(deviceName string) {
	d.clientsMu.Lock()
	defer d.clientsMu.Unlock()
	// note: delete on non-existing keys is a no-op
	delete(d.onvifClients, deviceName)
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	var edgexErr errors.EdgeX
	var responses = make([]*sdkModel.CommandValue, len(reqs))

	onvifClient, edgexErr := d.getOnvifClient(deviceName)
	if edgexErr != nil {
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

func attributeByKey(attributes map[string]interface{}, key string) (attr string, err errors.EdgeX) {
	val, ok := attributes[key]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("attribute %s not exists", key), nil)
	}
	attr = fmt.Sprint(val)
	return attr, nil
}

func parametersFromURLRawQuery(req sdkModel.CommandRequest) ([]byte, errors.EdgeX) {
	values, err := url.ParseQuery(fmt.Sprint(req.Attributes[URLRawQuery]))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to parse get command parameter for resource '%s'", req.DeviceResourceName), err)
	}
	param, exists := values[jsonObject]
	if !exists || len(param) == 0 {
		return []byte{}, nil
	}
	data, err := base64.StdEncoding.DecodeString(param[0])
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to decode '%v' parameter for resource '%s', the value should be json object with base64 encoded", jsonObject, req.DeviceResourceName), err)
	}
	return data, nil
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource (aka DeviceObject).
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	var edgexErr errors.EdgeX

	onvifClient, edgexErr := d.getOnvifClient(deviceName)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	for i, req := range reqs {
		parameters, err := params[i].ObjectValue()
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		data, err := json.Marshal(parameters)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to marshal set command parameter for resource '%s'", req.DeviceResourceName), err)
		}

		result, err := onvifClient.CallOnvifFunction(req, SetFunction, data)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to execute write command, %s", result), err)
		}
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (d *Driver) Stop(force bool) error {
	close(d.asynchCh)
	for _, client := range d.onvifClients {
		client.pullPointManager.UnsubscribeAll()
		client.baseNotificationManager.UnsubscribeAll()
	}

	close(d.taskCh) // send signal for taskLoop to finish
	d.wg.Wait()     // wait for taskLoop goroutine to return

	return nil
}

func (d *Driver) publishControlPlaneEvent(deviceName, eventType string) {
	var cv *sdkModel.CommandValue
	var err error

	cv, err = sdkModel.NewCommandValue(eventType, common.ValueTypeString, deviceName)
	if err != nil {
		d.lc.Errorf("issue creating control plane-event %s for device %s: %v", eventType, deviceName, err)
		return
	}

	asyncValues := &sdkModel.AsyncValues{
		DeviceName:    d.sdkService.Name(),
		CommandValues: []*sdkModel.CommandValue{cv},
	}
	d.asynchCh <- asyncValues
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// only execute if this was not called for the control-plane device
	if deviceName != d.sdkService.Name() {
		d.publishControlPlaneEvent(deviceName, cameraAdded)
		err := d.createOnvifClient(deviceName)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// only execute if this was not called for the control-plane device
	if deviceName != d.sdkService.Name() {
		d.publishControlPlaneEvent(deviceName, cameraUpdated)
		// Invoke the createOnvifClient func to create new onvif client and replace the old one
		err := d.createOnvifClient(deviceName)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	// only execute if this was not called for the control-plane device
	if deviceName != d.sdkService.Name() {
		d.publishControlPlaneEvent(deviceName, cameraDeleted)
		d.removeOnvifClient(deviceName)
	}
	return nil
}

// createOnvifClient creates the Onvif client used to communicate with the specified the device
func (d *Driver) createOnvifClient(deviceName string) error {
	device, err := d.sdkService.GetDeviceByName(deviceName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	onvifClient, err := d.newOnvifClient(device)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initialize onvif client for '%s' camera", device.Name), err)
	}

	d.clientsMu.Lock()
	defer d.clientsMu.Unlock()
	d.onvifClients[deviceName] = onvifClient
	return nil
}

// Discover performs a discovery on the network and passes them to EdgeX to get provisioned
func (d *Driver) Discover() {
	d.lc.Info("Discover was called.")

	d.configMu.RLock()
	maxSeconds := d.config.AppCustom.MaxDiscoverDurationSeconds
	discoveryMode := d.config.AppCustom.DiscoveryMode
	d.configMu.RUnlock()

	if !discoveryMode.IsValid() {
		d.lc.Errorf("DiscoveryMode is set to an invalid value: %s. Refusing to do discovery.", discoveryMode)
		return
	}

	if registerProvisionWatchers {
		d.watchersMu.Lock()
		if !d.addedWatchers {
			if err := d.addProvisionWatchers(); err != nil {
				d.lc.Errorf("Error adding provision watchers. Newly discovered devices may fail to register with EdgeX: %s",
					err.Error())
				// Do not return on failure, as it is possible there are alternative watchers registered.
				// And if not, the discovered devices will just not be registered with EdgeX, but will
				// still be available for discovery again.
			} else {
				d.addedWatchers = true
			}
		}
		d.watchersMu.Unlock()
	}

	var discoveredDevices []sdkModel.DiscoveredDevice

	if discoveryMode.IsMulticastEnabled() {
		discoveredDevices = append(discoveredDevices, d.discoverMulticast(discoveredDevices)...)
	}

	if discoveryMode.IsNetScanEnabled() {
		ctx := context.Background()
		if maxSeconds > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(),
				time.Duration(maxSeconds)*time.Second)
			defer cancel()
		}
		discoveredDevices = append(discoveredDevices, d.discoverNetscan(ctx, discoveredDevices)...)
	}

	// pass the discovered devices to the EdgeX SDK to be passed through to the provision watchers
	filtered := d.discoverFilter(discoveredDevices)
	d.deviceCh <- filtered
}

// multicast enable/disable via config option
func (d *Driver) discoverMulticast(discovered []sdkModel.DiscoveredDevice) []sdkModel.DiscoveredDevice {
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
func (d *Driver) discoverNetscan(ctx context.Context, discovered []sdkModel.DiscoveredDevice) []sdkModel.DiscoveredDevice {

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

func addressAndPort(xaddr string) (string, string) {
	substrings := strings.Split(xaddr, ":")
	if len(substrings) == 1 {
		// The port the might be empty from the discovered result, for example <d:XAddrs>http://192.168.12.123/onvif/device_service</d:XAddrs>
		return substrings[0], "80"
	} else {
		return substrings[0], substrings[1]
	}
}

// todo: this should be integrated better with getDeviceInformation to avoid creating another temporary client
func (d *Driver) getNetworkInterfaces(device models.Device) (netInfo *onvifdevice.GetNetworkInterfacesResponse, edgexErr errors.EdgeX) {
	devClient, edgexErr := d.newTemporaryOnvifClient(device)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfoResponse, edgexErr := devClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetNetworkInterfaces, []byte{})
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfo, ok := devInfoResponse.(*onvifdevice.GetNetworkInterfacesResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetNetworkInterfacesResponse for the camera %s", device.Name), nil)
	}
	return devInfo, nil
}

func (d *Driver) getDeviceInformation(device models.Device) (devInfo *onvifdevice.GetDeviceInformationResponse, edgexErr errors.EdgeX) {
	devClient, edgexErr := d.newTemporaryOnvifClient(device)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfoResponse, edgexErr := devClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetDeviceInformation, []byte{})
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfo, ok := devInfoResponse.(*onvifdevice.GetDeviceInformationResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetDeviceInformationResponse for the camera %s", device.Name), nil)
	}
	return devInfo, nil
}

func (d *Driver) getEndpointReference(device models.Device) (devInfo *onvifdevice.GetEndpointReferenceResponse, edgexErr errors.EdgeX) {
	devClient, edgexErr := d.newTemporaryOnvifClient(device)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	endpointRefResponse, edgexErr := devClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetEndpointReference, []byte{})
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devEndpointRef, ok := endpointRefResponse.(*onvifdevice.GetEndpointReferenceResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetEndpointReferenceResponse for the camera %s", device.Name), nil)
	}
	return devEndpointRef, nil
}

// newOnvifClient creates a temporary client for auto-discovery
func (d *Driver) newTemporaryOnvifClient(device models.Device) (*OnvifClient, errors.EdgeX) {
	xAddr, edgexErr := GetCameraXAddr(device.Protocols)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create cameraInfo for camera %s", device.Name), edgexErr)
	}

	credential, edgexErr := d.tryGetCredentialsForDevice(device)
	if edgexErr != nil {
		// if credentials are not found, instead of returning an error, set the AuthMode to NoAuth
		// and allow the user to call unauthenticated endpoints
		d.lc.Warnf("failed to get credentials for camera %s, setting AuthMode to none for temporary client", device.Name)
		credential = noAuthCredentials
	}

	d.configMu.Lock()
	requestTimeout := d.config.AppCustom.RequestTimeout
	d.configMu.Unlock()

	onvifDevice, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    xAddr,
		Username: credential.Username,
		Password: credential.Password,
		AuthMode: credential.AuthMode,
		HttpClient: &http.Client{
			Timeout: time.Duration(requestTimeout) * time.Second,
		},
	})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServiceUnavailable, "failed to initialize Onvif device client", err)
	}

	client := &OnvifClient{
		lc:          d.lc,
		DeviceName:  device.Name,
		onvifDevice: onvifDevice,
	}
	return client, nil
}

// refreshDevice will attempt to retrieve the MAC address and the device info for the specified camera
// and update the values in the protocol properties
// Also the device name is updated if the name starts with the UnknownDevicePrefix and the status is UpWithAuth
func (d *Driver) refreshDevice(device models.Device) error {
	// save the MAC Address in case it was changed by the calling code
	hwAddress := device.Protocols[OnvifProtocol][MACAddress]

	devInfo, err := d.getDeviceInformation(device)
	if err != nil {
		return err
	}

	netInfo, netErr := d.getNetworkInterfaces(device)
	if netErr != nil {
		d.lc.Warnf("Error trying to get network interfaces for device %s: %s", device.Name, netErr.Error())
	}

	endpointRef, endpointErr := d.getEndpointReference(device)
	if endpointErr != nil {
		d.lc.Warnf("Error trying to get get endpoint reference for device %s: %s", device.Name, endpointErr.Error())
	}

	// update device to latest version in cache to prevent race conditions
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
				strings.ReplaceAll(devInfo.Manufacturer, " ", "-"),
				strings.ReplaceAll(devInfo.Model, " ", "-"),
				device.Protocols[OnvifProtocol][EndpointRefAddress])
			d.lc.Infof("Adding device back with the updated name '%s'", device.Name)
			_, err = d.sdkService.AddDevice(device)
			return err
		}

		return d.sdkService.UpdateDevice(device)
	}
	return nil
}
