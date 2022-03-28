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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/IOTechSystems/onvif"
	"github.com/IOTechSystems/onvif/device"
)

var once sync.Once
var driver *Driver

const (
	URLRawQuery = "urlRawQuery"
	jsonObject  = "jsonObject"
)

// Driver implements the sdkModel.ProtocolDriver interface for
// the device service
type Driver struct {
	lc            logger.LoggingClient
	asynchCh      chan<- *sdkModel.AsyncValues
	deviceCh      chan<- []sdkModel.DiscoveredDevice
	config        *configuration
	lock          *sync.RWMutex
	deviceClients map[string]*DeviceClient
	svc           ServiceWrapper
}

// NewProtocolDriver initializes the singleton Driver and
// returns it to the caller
func NewProtocolDriver() *Driver {
	once.Do(func() {
		driver = new(Driver)
		driver.deviceClients = make(map[string]*DeviceClient)
	})

	return driver
}

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues,
	deviceCh chan<- []sdkModel.DiscoveredDevice) error {
	d.lc = lc
	d.asynchCh = asyncCh
	d.deviceCh = deviceCh
	d.lock = new(sync.RWMutex)

	camConfig, err := loadCameraConfig(sdk.DriverConfigs())
	if err != nil {
		panic(fmt.Errorf("load camera configuration failed: %w", err))
	}
	d.config = camConfig

	deviceService := sdk.RunningService()
	d.svc = &DeviceSDKService{
		DeviceService: deviceService,
		lc:            lc,
	}

	for _, device := range deviceService.Devices() {
		d.lc.Infof("Initializing device client for '%s' camera", device.Name)

		deviceClient, err := NewDeviceClient(device, d.config, d.lc)
		if err != nil {
			d.lc.Errorf("failed to initial device client for '%s' camera, skipping this device.", device.Name)
			continue
		}
		d.lock.Lock()
		d.deviceClients[device.Name] = deviceClient
		d.lock.Unlock()
	}

	handler := NewRestHandler(sdk.RunningService(), lc, asyncCh)
	edgexErr := handler.Start()
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	d.lc.Info("Driver initialized.")
	return nil
}

func (d *Driver) getDeviceClient(deviceName string) (*DeviceClient, errors.EdgeX) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	deviceClient, ok := d.deviceClients[deviceName]
	if !ok {
		device, err := sdk.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		deviceClient, err = NewDeviceClient(device, d.config, d.lc)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initial device client for '%s' camera", device.Name), err)
		}
		d.deviceClients[deviceName] = deviceClient
	}
	return deviceClient, nil
}

func (d *Driver) removeDeviceClient(deviceName string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	_, ok := d.deviceClients[deviceName]
	if ok {
		delete(d.deviceClients, deviceName)
	}
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	var edgexErr errors.EdgeX
	var responses = make([]*sdkModel.CommandValue, len(reqs))

	deviceClient, edgexErr := d.getDeviceClient(deviceName)
	if edgexErr != nil {
		return responses, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	for i, req := range reqs {
		data, edgexErr := parametersFromURLRawQuery(req)
		if edgexErr != nil {
			return responses, errors.NewCommonEdgeXWrapper(edgexErr)
		}

		cv, edgexErr := deviceClient.CallOnvifFunction(req, GetFunction, data)
		if edgexErr != nil {
			return responses, errors.NewCommonEdgeX(errors.KindServerError, "fail to execute read command", edgexErr)
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
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to parse get command parameter for resource '%s'", req.DeviceResourceName), err)
	}
	param, exists := values[jsonObject]
	if !exists || len(param) == 0 {
		return []byte{}, nil
	}
	data, err := base64.StdEncoding.DecodeString(param[0])
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to decode '%v' parameter for resource '%s', the value should be json object with base64 encoded", jsonObject, req.DeviceResourceName), err)
	}
	return data, nil
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource (aka DeviceObject).
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	var edgexErr errors.EdgeX

	deviceClient, edgexErr := d.getDeviceClient(deviceName)
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
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to marshal set command parameter for resource '%s'", req.DeviceResourceName), err)
		}

		result, err := deviceClient.CallOnvifFunction(req, SetFunction, data)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to execute write command, %s", result), err)
		}
	}

	return nil
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (d *Driver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Warn("Driver's DisconnectDevice function didn't implement")
	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (d *Driver) Stop(force bool) error {
	close(d.asynchCh)
	for _, client := range d.deviceClients {
		client.pullPointManager.UnsubscribeAll()
		client.baseNotificationManager.UnsubscribeAll()
	}

	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	device, err := sdk.RunningService().GetDeviceByName(deviceName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	deviceClient, err := NewDeviceClient(device, d.config, d.lc)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initial device client for '%s' camera", device.Name), err)
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	d.deviceClients[deviceName] = deviceClient
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.removeDeviceClient(deviceName)
	return nil
}

func GetCredentials(secretPath string) (config.Credentials, errors.EdgeX) {
	credentials := config.Credentials{}
	deviceService := sdk.RunningService()

	timer := startup.NewTimer(driver.config.CredentialsRetryTime, driver.config.CredentialsRetryWait)

	var secretData map[string]string
	var err error
	for timer.HasNotElapsed() {
		secretData, err = deviceService.SecretProvider.GetSecret(secretPath, secret.UsernameKey, secret.PasswordKey)
		if err == nil {
			break
		}

		driver.lc.Warnf(
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
	maxSeconds := driver.config.MaxDiscoverDurationSeconds
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

//// Discover triggers protocol specific device discovery, which is an asynchronous operation.
//// Devices found as part of this discovery operation are written to the channel devices.
//func (d *Driver) Discover() {
//
//
//
//	onvifDevices := wsdiscovery.GetAvailableDevicesAtSpecificEthernetInterface(d.config.DiscoveryEthernetInterface)
//	var discoveredDevices []sdkModel.DiscoveredDevice
//	for _, onvifDevice := range onvifDevices {
//		if onvifDevice.GetDeviceParams().EndpointRefAddress == "" {
//			d.lc.Warnf("The EndpointRefAddress is empty from the Onvif camera, unable to add the camera %s", onvifDevice.GetDeviceParams().Xaddr)
//			continue
//		}
//		address, port := addressAndPort(onvifDevice.GetDeviceParams().Xaddr)
//		dev := models.Device{
//			// Using Xaddr as the temporary name
//			Name: onvifDevice.GetDeviceParams().Xaddr,
//			Protocols: map[string]models.ProtocolProperties{
//				OnvifProtocol: {
//					Address:    address,
//					Port:       port,
//					AuthMode:   d.config.DefaultAuthMode,
//					SecretPath: d.config.DefaultSecretPath,
//				},
//			},
//		}
//
//		devInfo, edgexErr := d.getDeviceInformation(dev)
//		endpointRef := onvifDevice.GetDeviceParams().EndpointRefAddress
//		var discovered sdkModel.DiscoveredDevice
//		if edgexErr != nil {
//			d.lc.Warnf("failed to get the device information for the camera %s, %v", endpointRef, edgexErr)
//			dev.Protocols[OnvifProtocol][SecretPath] = endpointRef
//			discovered = sdkModel.DiscoveredDevice{
//				Name:        endpointRef,
//				Protocols:   dev.Protocols,
//				Description: "Auto discovered Onvif camera",
//				Labels:      []string{"auto-discovery"},
//			}
//			d.lc.Debugf("Discovered unknown camera from the address '%s'", onvifDevice.GetDeviceParams().Xaddr)
//		} else {
//			dev.Protocols[OnvifProtocol][Manufacturer] = devInfo.Manufacturer
//			dev.Protocols[OnvifProtocol][Model] = devInfo.Model
//			dev.Protocols[OnvifProtocol][FirmwareVersion] = devInfo.FirmwareVersion
//			dev.Protocols[OnvifProtocol][SerialNumber] = devInfo.SerialNumber
//			dev.Protocols[OnvifProtocol][HardwareId] = devInfo.HardwareId
//
//			// Spaces are not allowed in the device name
//			deviceName := fmt.Sprintf("%s-%s-%s",
//				strings.ReplaceAll(devInfo.Manufacturer, " ", "-"),
//				strings.ReplaceAll(devInfo.Model, " ", "-"),
//				onvifDevice.GetDeviceParams().EndpointRefAddress)
//
//			discovered = sdkModel.DiscoveredDevice{
//				Name:        deviceName,
//				Protocols:   dev.Protocols,
//				Description: fmt.Sprintf("%s %s Camera", devInfo.Manufacturer, devInfo.Model),
//				Labels:      []string{"auto-discovery", devInfo.Manufacturer, devInfo.Model},
//			}
//			d.lc.Debugf("Discovered camera from the address '%s'", onvifDevice.GetDeviceParams().Xaddr)
//		}
//		discoveredDevices = append(discoveredDevices, discovered)
//	}
//
//	d.deviceCh <- discoveredDevices
//}

func (d *Driver) discover(ctx context.Context) {
	params := discoverParams{
		// split the comma separated string here to avoid issues with EdgeX's Consul implementation
		subnets:           strings.Split(d.config.DiscoverySubnets, ","),
		asyncLimit:        d.config.ProbeAsyncLimit,
		timeout:           time.Duration(d.config.ProbeTimeoutSeconds) * time.Second,
		scanPorts:         strings.Split(d.config.ScanPorts, ","),
		defaultAuthMode:   d.config.DefaultAuthMode,
		defaultSecretPath: d.config.DefaultSecretPath,
		lc:                d.lc,
		driver:            d,
	}

	t1 := time.Now()
	result := autoDiscover(ctx, params)
	if ctx.Err() != nil {
		d.lc.Warn("Discover process has been cancelled!", "ctxErr", ctx.Err())
	}

	d.lc.Info(fmt.Sprintf("Discovered %d new devices in %v.", len(result), time.Since(t1)))
	// pass the discovered devices to the EdgeX SDK to be passed through to the provision watchers
	d.deviceCh <- result
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
	devClient, edgexErr := newDeviceClient(dev, d.config, d.lc)
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

// newDeviceClient creates a temporary client for auto-discovery
func newDeviceClient(device models.Device, driverConfig *configuration, lc logger.LoggingClient) (*DeviceClient, errors.EdgeX) {
	cameraInfo, edgexErr := CreateCameraInfo(device.Protocols)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create cameraInfo for camera %s", device.Name), edgexErr)
	}

	var credential config.Credentials
	if cameraInfo.AuthMode != onvif.NoAuth {
		credential, edgexErr = GetCredentials(cameraInfo.SecretPath)
		if edgexErr != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get credentials for camera %s", device.Name), edgexErr)
		}
	}

	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    deviceAddress(cameraInfo),
		Username: credential.Username,
		Password: credential.Password,
		AuthMode: cameraInfo.AuthMode,
		HttpClient: &http.Client{
			Timeout: time.Duration(driverConfig.RequestTimeout) * time.Second,
		},
	})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServiceUnavailable, "fail to initial Onvif device client", err)
	}

	client := &DeviceClient{
		lc:          lc,
		DeviceName:  device.Name,
		cameraInfo:  cameraInfo,
		onvifDevice: dev,
	}
	return client, nil
}
