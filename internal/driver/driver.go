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
	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
	"github.com/edgexfoundry/device-onvif-camera/pkg/netscan"
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
	"github.com/IOTechSystems/onvif/device"
)

const (
	URLRawQuery = "urlRawQuery"
	jsonObject  = "jsonObject"

	cameraAdded   = "CameraAdded"
	cameraUpdated = "CameraUpdated"
	cameraDeleted = "CameraDeleted"
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
	svc          ServiceWrapper
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
	d.svc = &DeviceSDKService{
		DeviceService: deviceService,
		lc:            lc,
	}

	for _, dev := range deviceService.Devices() {
		// onvif client should not be created for the control-plane device
		if dev.Name == d.serviceName {
			continue
		}

		d.lc.Infof("Initializing onvif client for '%s' camera", dev.Name)

		onvifClient, err := d.newOnvifClient(dev)
		if err != nil {
			d.lc.Errorf("failed to initial onvif client for '%s' camera, skipping this device.", dev.Name)
			continue
		}
		d.lock.Lock()
		d.onvifClients[dev.Name] = onvifClient
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
		dev, err := sdk.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		onvifClient, err = d.newOnvifClient(dev)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initial onvif client for '%s' camera", dev.Name), err)
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

// createOnvifClient create the Onvif client for specified the device
func (d *Driver) createOnvifClient(deviceName string) error {
	dev, err := sdk.RunningService().GetDeviceByName(deviceName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	onvifClient, err := d.newOnvifClient(dev)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initial onvif client for '%s' camera", dev.Name), err)
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	d.onvifClients[deviceName] = onvifClient
	return nil
}

func (d *Driver) getCredentials(secretPath string) (config.Credentials, errors.EdgeX) {
	credentials := config.Credentials{}
	deviceService := sdk.RunningService()

	timer := startup.NewTimer(d.config.CredentialsRetryTime, d.config.CredentialsRetryWait)

	var secretData map[string]string
	var err error
	for timer.HasNotElapsed() {
		secretData, err = deviceService.SecretProvider.GetSecret(secretPath, secret.UsernameKey, secret.PasswordKey)
		if err == nil {
			break
		}

		d.lc.Warnf(
			"Unable to retrieve camera credentials from SecretProvider at path '%s': %s. Retrying for %s",
			secretPath,
			err.Error(),
			timer.RemainingAsString())
		timer.SleepForInterval()
	}

	if err != nil {
		return credentials, errors.NewCommonEdgeXWrapper(err)
	}

	credentials.Username = secretData[secret.UsernameKey]
	credentials.Password = secretData[secret.PasswordKey]

	return credentials, nil
}

// Discover performs a discovery on the network and passes them to EdgeX to get provisioned
func (d *Driver) Discover() {
	d.lc.Info("Discover was called.")
	//
	//d.configMu.RLock()
	maxSeconds := d.config.MaxDiscoverDurationSeconds
	//d.configMu.RUnlock()
	//
	//if registerProvisionWatchers {
	//	d.watchersMu.Lock()
	//	if !d.addedWatchers {
	//		if err := d.addProvisionWatchers(); err != nil {
	//			d.lc.Error("Error adding provision watchers. Newly discovered devices may fail to register with EdgeX.",
	//				"error", err.Error())
	//			// Do not return on failure, as it is possible there are alternative watchers registered.
	//			// And if not, the discovered devices will just not be registered with EdgeX, but will
	//			// still be available for discovery again.
	//		} else {
	//			d.addedWatchers = true
	//		}
	//	}
	//	d.watchersMu.Unlock()
	//}

	ctx := context.Background()
	if maxSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(),
			time.Duration(maxSeconds)*time.Second)
		defer cancel()
	}

	d.discover(ctx)
}

func (d *Driver) discover(ctx context.Context) {
	var discovered []sdkModel.DiscoveredDevice

	// TODO: support multicast via config option
	t0 := time.Now()
	onvifDevices := wsdiscovery.GetAvailableDevicesAtSpecificEthernetInterface(d.config.DiscoveryEthernetInterface)
	d.lc.Info(fmt.Sprintf("Discovered %d device(s) in %v via multicast.", len(onvifDevices), time.Since(t0)))
	for _, onvifDevice := range onvifDevices {
		dev, err := d.createDiscoveredDevice(onvifDevice)
		if err != nil {
			d.lc.Warnf(err.Error())
			continue
		}
		discovered = append(discovered, dev)
	}

	params := netscan.Params{
		// split the comma separated string here to avoid issues with EdgeX's Consul implementation
		Subnets:            strings.Split(d.config.DiscoverySubnets, ","),
		AsyncLimit:         d.config.ProbeAsyncLimit,
		Timeout:            time.Duration(d.config.ProbeTimeoutMillis) * time.Millisecond,
		ScanPorts:          strings.Split(d.config.ScanPorts, ","),
		Logger:             d.lc,
		NetworkProtocol:    netscan.NetworkTCP, // todo: configurable?
		MaxTimeoutsPerHost: 2,                  // todo: configurable?
	}

	t1 := time.Now()
	result := netscan.AutoDiscover(ctx, NewOnvifProtocolDiscovery(d), params)
	if ctx.Err() != nil {
		d.lc.Warn("Discover process has been cancelled!", "ctxErr", ctx.Err())
	}

	d.lc.Debugf("NetScan result: %+v", result)
	d.lc.Info(fmt.Sprintf("Discovered %d device(s) in %v via netscan.", len(result), time.Since(t1)))

	for _, res := range result {
		dev, ok := res.Info.(sdkModel.DiscoveredDevice)
		if !ok {
			d.lc.Warnf("unable to cast res.Data into sdkModel.DiscoveredDevice. type=%T", res.Info)
			continue
		}
		discovered = append(discovered, dev)
	}
	// pass the discovered devices to the EdgeX SDK to be passed through to the provision watchers
	d.deviceCh <- discovered
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

func (d *Driver) getDeviceInformation(dev models.Device) (devInfo *device.GetDeviceInformationResponse, edgexErr errors.EdgeX) {
	devClient, edgexErr := d.newTemporaryOnvifClient(dev)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfoResponse, edgexErr := devClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetDeviceInformation, []byte{})
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfo, ok := devInfoResponse.(*device.GetDeviceInformationResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetDeviceInformationResponse for the camera %s", dev.Name), nil)
	}
	return devInfo, nil
}

// newOnvifClient creates a temporary client for auto-discovery
func (d *Driver) newTemporaryOnvifClient(dev models.Device) (*OnvifClient, errors.EdgeX) {
	cameraInfo, edgexErr := CreateCameraInfo(dev.Protocols)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create cameraInfo for camera %s", dev.Name), edgexErr)
	}

	var credential config.Credentials
	if cameraInfo.AuthMode != onvif.NoAuth {
		credential, edgexErr = d.getCredentials(cameraInfo.SecretPath)
		if edgexErr != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get credentials for camera %s", dev.Name), edgexErr)
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
		return nil, errors.NewCommonEdgeX(errors.KindServiceUnavailable, "failed to initial Onvif device client", err)
	}

	client := &OnvifClient{
		lc:          d.lc,
		DeviceName:  dev.Name,
		cameraInfo:  cameraInfo,
		onvifDevice: onvifDevice,
	}
	return client, nil
}
