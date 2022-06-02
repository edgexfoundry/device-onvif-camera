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
	"github.com/edgexfoundry/device-onvif-camera/internal/netscan"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
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
)

type DiscoveryMode string

const (
	NetScan   DiscoveryMode = "netscan"
	Multicast DiscoveryMode = "multicast"
	Both      DiscoveryMode = "both"
)

// Driver implements the sdkModel.ProtocolDriver interface for
// the device service
type Driver struct {
	lc           logger.LoggingClient
	asynchCh     chan<- *sdkModel.AsyncValues
	deviceCh     chan<- []sdkModel.DiscoveredDevice
	config       *configuration
	lock         *sync.RWMutex
	onvifClients map[string]*OnvifClient
	serviceName  string
}

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues,
	deviceCh chan<- []sdkModel.DiscoveredDevice) error {
	d.lc = lc
	d.asynchCh = asyncCh
	d.deviceCh = deviceCh
	d.lock = new(sync.RWMutex)
	d.onvifClients = make(map[string]*OnvifClient)
	d.serviceName = sdk.RunningService().Name()

	camConfig, err := loadCameraConfig(sdk.DriverConfigs())
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to load camera configuration", err)
	}
	d.config = camConfig

	deviceService := sdk.RunningService()

	for _, device := range deviceService.Devices() {
		// onvif client should not be created for the control-plane device
		if device.Name == d.serviceName {
			continue
		}

		d.lc.Infof("Initializing onvif client for '%s' camera", device.Name)

		onvifClient, err := d.newOnvifClient(device)
		if err != nil {
			d.lc.Errorf("failed to initialize onvif client for '%s' camera, skipping this device.", device.Name)
			continue
		}
		d.lock.Lock()
		d.onvifClients[device.Name] = onvifClient
		d.lock.Unlock()
	}

	handler := NewRestNotificationHandler(deviceService, lc, asyncCh)
	edgexErr := handler.AddRoute()
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	d.lc.Info("Driver initialized.")
	return nil
}

func (d *Driver) getOnvifClient(deviceName string) (*OnvifClient, errors.EdgeX) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	onvifClient, ok := d.onvifClients[deviceName]
	if !ok {
		device, err := sdk.RunningService().GetDeviceByName(deviceName)
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
	d.lock.Lock()
	defer d.lock.Unlock()
	_, ok := d.onvifClients[deviceName]
	if ok {
		delete(d.onvifClients, deviceName)
	}
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

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (d *Driver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Warn("Driver's DisconnectDevice function not implemented")
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
		DeviceName:    d.serviceName,
		CommandValues: []*sdkModel.CommandValue{cv},
	}
	d.asynchCh <- asyncValues
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// only execute if this was not called for the control-plane device
	if deviceName != d.serviceName {
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
	if deviceName != d.serviceName {
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
	if deviceName != d.serviceName {
		d.publishControlPlaneEvent(deviceName, cameraDeleted)
		d.removeOnvifClient(deviceName)
	}
	return nil
}

// createOnvifClient creates the Onvif client used to communicate with the specified the device
func (d *Driver) createOnvifClient(deviceName string) error {
	device, err := sdk.RunningService().GetDeviceByName(deviceName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	onvifClient, err := d.newOnvifClient(device)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initialize onvif client for '%s' camera", device.Name), err)
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	d.onvifClients[deviceName] = onvifClient
	return nil
}

// tryGetCredentials will attempt one time to get the credentials located at secretPath from
// secret provider and return them, otherwise return an error.
func (d *Driver) tryGetCredentials(secretPath string) (config.Credentials, errors.EdgeX) {
	secretData, err := sdk.RunningService().SecretProvider.GetSecret(secretPath, secret.UsernameKey, secret.PasswordKey)
	if err != nil {
		return config.Credentials{}, errors.NewCommonEdgeXWrapper(err)
	}
	return config.Credentials{
		Username: secretData[secret.UsernameKey],
		Password: secretData[secret.PasswordKey],
	}, nil
}

// getCredentials will repeatedly try and get the credentials located at secretPath from
// secret provider every CredentialsRetryTime seconds for a maximum of CredentialsRetryWait seconds.
// Note that this function will block until either the credentials are found, or CredentialsRetryWait
// seconds have elapsed.
func (d *Driver) getCredentials(secretPath string) (credentials config.Credentials, err errors.EdgeX) {
	timer := startup.NewTimer(d.config.CredentialsRetryTime, d.config.CredentialsRetryWait)

	for timer.HasNotElapsed() {
		if credentials, err = d.tryGetCredentials(secretPath); err == nil {
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

// Discover performs a discovery on the network and passes them to EdgeX to get provisioned
func (d *Driver) Discover() {
	d.lc.Info("Discover was called.")

	maxSeconds := d.config.MaxDiscoverDurationSeconds
	ctx := context.Background()
	if maxSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(),
			time.Duration(maxSeconds)*time.Second)
		defer cancel()
	}

	var discoveredDevices []sdkModel.DiscoveredDevice
	if d.config.DiscoveryMode == Multicast || d.config.DiscoveryMode == Both {
		discoveredDevices = append(discoveredDevices, d.discoverMulticast(discoveredDevices)...)
	}
	if d.config.DiscoveryMode == NetScan || d.config.DiscoveryMode == Both {
		discoveredDevices = append(discoveredDevices, d.discoverNetscan(ctx, discoveredDevices)...)
	}
	// pass the discovered devices to the EdgeX SDK to be passed through to the provision watchers
	d.deviceCh <- discoveredDevices
}

// multicast enable/disable via config option
func (d *Driver) discoverMulticast(discovered []sdkModel.DiscoveredDevice) []sdkModel.DiscoveredDevice {
	t0 := time.Now()
	onvifDevices := wsdiscovery.GetAvailableDevicesAtSpecificEthernetInterface(d.config.DiscoveryEthernetInterface)
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

	if len(strings.TrimSpace(d.config.DiscoverySubnets)) == 0 {
		d.lc.Warn("netscan discovery was called, but DiscoverySubnets are empty!")
		return nil
	}

	params := netscan.Params{
		// split the comma separated string here to avoid issues with EdgeX's Consul implementation
		Subnets:         strings.Split(d.config.DiscoverySubnets, ","),
		AsyncLimit:      d.config.ProbeAsyncLimit,
		Timeout:         time.Duration(d.config.ProbeTimeoutMillis) * time.Millisecond,
		ScanPorts:       []string{wsDiscoveryPort},
		Logger:          d.lc,
		NetworkProtocol: netscan.NetworkUDP,
	}

	t0 := time.Now()
	result := netscan.AutoDiscover(ctx, NewOnvifProtocolDiscovery(d), params)
	if ctx.Err() != nil {
		d.lc.Warn("Discover process has been cancelled!", "ctxErr", ctx.Err())
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

// newOnvifClient creates a temporary client for auto-discovery
func (d *Driver) newTemporaryOnvifClient(device models.Device) (*OnvifClient, errors.EdgeX) {
	cameraInfo, edgexErr := CreateCameraInfo(device.Protocols)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create cameraInfo for camera %s", device.Name), edgexErr)
	}

	var credential config.Credentials
	if cameraInfo.AuthMode != onvif.NoAuth {
		// since this is just a temporary client, we do not want to wait for credentials to be available
		credential, edgexErr = d.tryGetCredentials(cameraInfo.SecretPath)
		if edgexErr != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get credentials for camera %s", device.Name), edgexErr)
		}
	}

	onvifDevice, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    deviceAddress(cameraInfo),
		Username: credential.Username,
		Password: credential.Password,
		AuthMode: cameraInfo.AuthMode,
		HttpClient: &http.Client{
			Timeout: time.Duration(d.config.RequestTimeout) * time.Second,
		},
	})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServiceUnavailable, "failed to initialize Onvif device client", err)
	}

	client := &OnvifClient{
		lc:          d.lc,
		DeviceName:  device.Name,
		cameraInfo:  cameraInfo,
		onvifDevice: onvifDevice,
	}
	return client, nil
}
