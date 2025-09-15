// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
// Copyright (c) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/IOTechSystems/onvif"
	onvifdevice "github.com/IOTechSystems/onvif/device"
	"github.com/IOTechSystems/onvif/gosoap"
	"github.com/IOTechSystems/onvif/media"

	"github.com/spf13/cast"
)

const (
	EdgeXWebService        = "EdgeX"
	RebootNeeded           = "RebootNeeded"
	CameraEvent            = "CameraEvent"
	SubscribeCameraEvent   = "SubscribeCameraEvent"
	UnsubscribeCameraEvent = "UnsubscribeCameraEvent"
	GetSnapshot            = "GetSnapshot"
)

// OnvifClient manages the state required to issue ONVIF requests to the specified camera
type OnvifClient struct {
	driver      *Driver
	lc          logger.LoggingClient
	DeviceName  string
	onvifDevice OnvifDevice
	// RebootNeeded indicates the camera should reboot to apply the configuration change
	RebootNeeded bool
	// CameraEventResource is used to send the async event to north bound
	CameraEventResource     models.DeviceResource
	pullPointManager        *PullPointManager
	baseNotificationManager *BaseNotificationManager
}

// newOnvifClient returns a new OnvifClient for communicating with a single camera with all of the additional
// managers and resources needed for normal operation.
func (d *Driver) newOnvifClient(device models.Device) (*OnvifClient, errors.EdgeX) {
	return d.newOnvifClientInternal(device, false)
}

// newTemporaryOnvifClient returns a new OnvifClient for communicating with a single camera, however
// it is created without the extra managers and resources of a normal client for use in auto-discovery.
func (d *Driver) newTemporaryOnvifClient(device models.Device) (*OnvifClient, errors.EdgeX) {
	return d.newOnvifClientInternal(device, true)
}

// newOnvifClientInternal creates either a normal or a temporary OnvifClient
func (d *Driver) newOnvifClientInternal(device models.Device, temporary bool) (*OnvifClient, errors.EdgeX) {
	xAddr, edgexErr := GetCameraXAddr(device.Protocols)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create cameraInfo for camera %s", device.Name), edgexErr)
	}

	d.configMu.Lock()
	requestTimeout := d.config.AppCustom.RequestTimeout
	d.configMu.Unlock()

	credentials := d.getCredentialsForDevice(device)
	onvifDevice, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    xAddr,
		Username: credentials.Username,
		Password: credentials.Password,
		AuthMode: credentials.AuthMode,
		HttpClient: &http.Client{
			Timeout: time.Duration(requestTimeout) * time.Second,
		},
	})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServiceUnavailable, "failed to initialize Onvif device client", err)
	}

	client := &OnvifClient{
		driver:      d,
		lc:          d.lc,
		DeviceName:  device.Name,
		onvifDevice: onvifDevice,
	}

	if temporary {
		return client, nil
	}

	client.CameraEventResource, err = d.getCameraEventResourceByDeviceName(device.Name)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	// Create PullPointManager to control multiple pull points
	pullPointManager := newPullPointManager(d.lc)
	client.pullPointManager = pullPointManager

	// Create BaseNotificationManager to control multiple notification consumer
	baseNotificationManager := NewBaseNotificationManager(d.lc)
	client.baseNotificationManager = baseNotificationManager
	return client, nil
}

// updateOnvifClient updates the internal onvifDevice of an onvif client
func (d *Driver) updateOnvifClient(device models.Device) errors.EdgeX {
	xAddr, edgexErr := GetCameraXAddr(device.Protocols)
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create cameraInfo for camera %s", device.Name), edgexErr)
	}

	onvifClient, edgexErr := d.getOrCreateOnvifClient(device)
	if edgexErr != nil {
		return edgexErr
	}

	credentials := d.getCredentialsForDevice(device)
	existingParams := onvifClient.onvifDevice.GetDeviceParams()
	// check the internal parameters used when creating the onvif device vs the current ones
	if xAddr == existingParams.Xaddr && credentials.Username == existingParams.Username &&
		credentials.Password == existingParams.Password && credentials.AuthMode == existingParams.AuthMode {
		// XAddr and credentials are the same, skip creating new connection
		d.lc.Tracef("Skip creating new connection for un-modified device %s", device.Name)
		return nil
	}

	d.lc.Debugf("Updating connection for modified device %s", device.Name)

	d.configMu.Lock()
	requestTimeout := d.config.AppCustom.RequestTimeout
	d.configMu.Unlock()

	onvifDevice, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    xAddr,
		Username: credentials.Username,
		Password: credentials.Password,
		AuthMode: credentials.AuthMode,
		HttpClient: &http.Client{
			Timeout: time.Duration(requestTimeout) * time.Second,
		},
	})
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServiceUnavailable, "failed to update Onvif device client", err)
	}

	// lock the clients to prevent access while the update occurs
	d.clientsMu.Lock()
	onvifClient.onvifDevice = onvifDevice
	d.clientsMu.Unlock()

	d.checkStatusOfDevice(device)
	return nil
}

func (d *Driver) getOrCreateOnvifClient(device models.Device) (*OnvifClient, errors.EdgeX) {
	d.clientsMu.RLock()
	onvifClient, ok := d.onvifClients[device.Name]
	d.clientsMu.RUnlock()
	if ok {
		return onvifClient, nil
	}

	onvifClient, err := d.newOnvifClient(device)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to initialize onvif client for '%s' camera", device.Name), err)
	}

	d.clientsMu.Lock()
	d.onvifClients[device.Name] = onvifClient
	d.clientsMu.Unlock()
	return onvifClient, nil
}

func (d *Driver) removeOnvifClient(deviceName string) {
	d.clientsMu.Lock()
	defer d.clientsMu.Unlock()
	// note: delete on non-existing keys is a no-op
	delete(d.onvifClients, deviceName)
}

func (d *Driver) getCameraEventResourceByDeviceName(deviceName string) (r models.DeviceResource, edgexErr errors.EdgeX) {
	device, err := d.sdkService.GetDeviceByName(deviceName)
	if err != nil {
		return r, errors.NewCommonEdgeXWrapper(err)
	}
	profile, err := d.sdkService.GetProfileByName(device.ProfileName)
	if err != nil {
		return r, errors.NewCommonEdgeXWrapper(err)
	}
	for _, r := range profile.DeviceResources {
		val, ok := r.Attributes[GetFunction]
		if ok && fmt.Sprint(val) == CameraEvent {
			return r, nil
		}
	}
	return r, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device resource with Getfunciton '%s' not found", CameraEvent), nil)
}

// CallOnvifFunction send the request to the camera via onvif client
func (onvifClient *OnvifClient) CallOnvifFunction(req sdkModel.CommandRequest, functionType string, data []byte) (cv *sdkModel.CommandValue, edgexErr errors.EdgeX) {
	serviceName, edgexErr := attributeByKey(req.Attributes, Service)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	functionName, edgexErr := attributeByKey(req.Attributes, functionType)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if serviceName == EdgeXWebService {
		cv, edgexErr := onvifClient.callCustomFunction(req.DeviceResourceName, functionName, req.Attributes, data)
		if edgexErr != nil {
			return nil, errors.NewCommonEdgeXWrapper(edgexErr)
		}
		return cv, nil
	}

	responseContent, edgexErr := onvifClient.callOnvifFunction(serviceName, functionName, data)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if functionName == onvif.SetNetworkInterfaces {
		onvifClient.checkRebootNeeded(responseContent)
	} else if functionName == onvif.SystemReboot {
		onvifClient.RebootNeeded = false
	}
	cv, err := sdkModel.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, responseContent)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the function '%s' of web service '%s' ", functionName, serviceName), err)
	}
	return cv, nil
}

func (onvifClient *OnvifClient) callCustomFunction(resourceName, functionName string, attributes map[string]interface{}, data []byte) (cv *sdkModel.CommandValue, edgexErr errors.EdgeX) {
	var err error
	switch functionName {
	case GetCustomMetadata:
		deviceName := onvifClient.DeviceName
		device, err := onvifClient.driver.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
		}

		metadataObj, edgexError := onvifClient.getCustomMetadata(device, data)
		if edgexError != nil {
			onvifClient.driver.lc.Errorf("Failed to get customMetadata from the device %s", onvifClient.DeviceName)
			return nil, edgexError
		}
		cv, err = sdkModel.NewCommandValue(resourceName, common.ValueTypeObject, metadataObj)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the web service '%s' function '%s'", EdgeXWebService, functionName), err)
		}

		attributes[URLRawQuery] = "" // flush out the query so it resets with new calls
	case SetCustomMetadata:
		deviceName := onvifClient.DeviceName
		device, err := onvifClient.driver.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
		}

		updatedDevice, setErr := onvifClient.setCustomMetadata(device, data)
		if setErr != nil {
			onvifClient.driver.lc.Errorf("Failed to set customMetadata for the device '%s'", deviceName)
			return nil, setErr
		}
		err = onvifClient.driver.patchDeviceProtocols(deviceName, updatedDevice.Protocols)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to update device '%s'", deviceName), err)
		}
	case DeleteCustomMetadata:
		deviceName := onvifClient.DeviceName
		device, err := onvifClient.driver.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
		}

		updatedDevice, delErr := onvifClient.deleteCustomMetadata(device, data)
		if delErr != nil {
			onvifClient.driver.lc.Errorf("Failed to delete customMetadata for the device '%s'", deviceName)
			return nil, delErr
		}
		err = onvifClient.driver.patchDeviceProtocols(deviceName, updatedDevice.Protocols)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to update device '%s'", deviceName), err)
		}
	case RebootNeeded:
		cv, err = sdkModel.NewCommandValue(resourceName, common.ValueTypeBool, onvifClient.RebootNeeded)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the web service '%s' function '%s'", EdgeXWebService, functionName), err)
		}
	case SubscribeCameraEvent:
		err = onvifClient.callSubscribeCameraEventFunction(resourceName, EdgeXWebService, functionName, attributes, data)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
	case UnsubscribeCameraEvent:
		go func() {
			onvifClient.lc.Debugf("Unsubscribe camera event for the device '%v'", onvifClient.DeviceName)
			onvifClient.pullPointManager.UnsubscribeAll()
			onvifClient.baseNotificationManager.UnsubscribeAll()
		}()
	case GetSnapshot:
		res, edgexErr := onvifClient.callGetSnapshotFunction(data)
		if edgexErr != nil {
			return nil, errors.NewCommonEdgeXWrapper(edgexErr)
		}
		cv, err = sdkModel.NewCommandValue(resourceName, common.ValueTypeBinary, res)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the web service '%s' function '%s'", EdgeXWebService, functionName), err)
		}
	case SetFriendlyName:
		deviceName := onvifClient.DeviceName
		device, err := onvifClient.driver.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
		}

		friendlyName := strings.TrimSpace(string(data))
		if friendlyName == "" {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "no data in request body", nil)
		}
		device.Protocols[OnvifProtocol][FriendlyName] = friendlyName // create or update friendly name field
		err = onvifClient.driver.sdkService.UpdateDevice(device)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to update device '%s'", deviceName), err)
		}
	case GetFriendlyName:
		deviceName := onvifClient.DeviceName
		device, err := onvifClient.driver.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
		}
		friendlyName := ""
		if v, ok := device.Protocols[OnvifProtocol][FriendlyName]; ok {
			friendlyName = cast.ToString(v)
		}
		cv, err = sdkModel.NewCommandValue(resourceName, common.ValueTypeString, friendlyName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the web service '%s' function '%s'", EdgeXWebService, functionName), err)
		}
	case SetMACAddress:
		deviceName := onvifClient.DeviceName
		device, err := onvifClient.driver.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
		}

		mac := strings.TrimSpace(string(data))
		if mac == "" {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "no data in request body", nil)
		}

		mac, err = SanitizeMACAddress(mac)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "error setting MACAddress", err)
		}

		device.Protocols[OnvifProtocol][MACAddress] = mac // create or update mac address field
		err = onvifClient.driver.sdkService.UpdateDevice(device)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to update device '%s'", deviceName), err)
		}
	case GetMACAddress:
		deviceName := onvifClient.DeviceName
		device, err := onvifClient.driver.sdkService.GetDeviceByName(deviceName)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
		}
		macAddress := ""
		if v, ok := device.Protocols[OnvifProtocol][MACAddress]; ok {
			macAddress = cast.ToString(v)
		}
		cv, err = sdkModel.NewCommandValue(resourceName, common.ValueTypeString, macAddress)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the web service '%s' function '%s'", EdgeXWebService, functionName), err)
		}
	default:
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("not support the custom function '%s'", functionName), nil)
	}
	return cv, nil
}

func (onvifClient *OnvifClient) callSubscribeCameraEventFunction(resourceName, serviceName, functionName string, attributes map[string]interface{}, data []byte) errors.EdgeX {
	subscribeType, edgexErr := attributeByKey(attributes, SubscribeType)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	switch subscribeType {
	case PullPoint:
		edgexErr = onvifClient.pullPointManager.NewSubscriber(onvifClient, resourceName, attributes, data)
		if edgexErr != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the web service '%s' function '%s'", serviceName, functionName), edgexErr)
		}
	case BaseNotification:
		edgexErr = onvifClient.baseNotificationManager.NewConsumer(onvifClient, resourceName, attributes, data)
		if edgexErr != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue for the web service '%s' function '%s'", serviceName, functionName), edgexErr)
		}
	default:
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported subscribeType '%s'", subscribeType), nil)
	}
	return nil
}

// callGetSnapshotFunction returns a snapshot from the camera as a slice of bytes
// The implementation can refer to https://github.com/edgexfoundry/device-camera-go/blob/5c4f34d1d59b8e25e1a6316661d463e2495d45fe/internal/driver/onvifclient.go#L119
func (onvifClient *OnvifClient) callGetSnapshotFunction(data []byte) ([]byte, errors.EdgeX) {
	// data contains profile token json object
	respContent, edgexErr := onvifClient.callOnvifFunction(onvif.MediaWebService, onvif.GetSnapshotUri, data)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	uriResponse, ok := respContent.(*media.GetSnapshotUriResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetSnapshotUriResponse of type %T for the camera %s", respContent, onvifClient.DeviceName), nil)
	}
	url := uriResponse.MediaUri.Uri

	// Get the snapshot binary data
	resp, err := onvifClient.onvifDevice.SendGetSnapshotRequest(string(url))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to retrieve the snapshot from the url %s", url), err)
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "error reading http request", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("http request for image failed with status %v, %s", resp.StatusCode, string(buf)), nil)
	}
	return buf, nil
}

func (onvifClient *OnvifClient) callOnvifFunction(serviceName, functionName string, data []byte) (interface{}, errors.EdgeX) {
	function, edgexErr := onvif.FunctionByServiceAndFunctionName(serviceName, functionName)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	request, edgexErr := createRequest(function, data)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create '%s' request for the web service '%s'", functionName, serviceName), edgexErr)
	}

	endpoint, err := onvifClient.onvifDevice.GetEndpointByRequestStruct(request)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	requestBody, err := xml.Marshal(request)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	xmlRequestBody := string(requestBody)
	onvifClient.lc.Debugf("SOAP Request: %v", xmlRequestBody)

	servResp, err := onvifClient.onvifDevice.SendSoap(endpoint, xmlRequestBody)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to send the '%s' request for the web service '%s'", functionName, serviceName), err)
	}
	defer servResp.Body.Close()

	rsp, err := io.ReadAll(servResp.Body)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	responseEnvelope, edgexErr := createResponse(function, rsp)
	if edgexErr != nil {
		// log the raw response from the camera since it will not be logged further down
		onvifClient.lc.Debugf("Raw SOAP Response: %v", string(rsp))
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create '%s' response for the web service '%s'", functionName, serviceName), edgexErr)
	}
	res, _ := xml.Marshal(responseEnvelope)
	onvifClient.lc.Debugf("SOAP Response: %v", string(res))

	if servResp.StatusCode == http.StatusUnauthorized {
		return nil, errors.NewCommonEdgeX(errors.KindInvalidId,
			fmt.Sprintf("failed to verify the authentication for the function '%s' of web service '%s'. Onvif error: %s",
				functionName, serviceName, responseEnvelope.Body.Fault.String()), nil)
	} else if servResp.StatusCode == http.StatusBadRequest {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("invalid request for the function '%s' of web service '%s'. Onvif error: %s",
				functionName, serviceName, responseEnvelope.Body.Fault.String()), nil)
	} else if servResp.StatusCode > http.StatusNoContent {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to execute the request for the function '%s' of web service '%s'. Onvif error: %s",
				functionName, serviceName, responseEnvelope.Body.Fault.String()), nil)
	}
	return responseEnvelope.Body.Content, nil
}

func createRequest(function onvif.Function, data []byte) (interface{}, errors.EdgeX) {
	request := function.Request()
	if len(data) > 0 {
		err := json.Unmarshal(data, request)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
	}
	return request, nil
}

func createResponse(function onvif.Function, data []byte) (*gosoap.SOAPEnvelope, errors.EdgeX) {
	response := function.Response()
	responseEnvelope := gosoap.NewSOAPEnvelope(response)
	err := xml.Unmarshal(data, responseEnvelope)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return responseEnvelope, nil
}

func (onvifClient *OnvifClient) checkRebootNeeded(responseContent interface{}) {
	setNetworkInterfacesResponse, ok := responseContent.(*onvifdevice.SetNetworkInterfacesResponse)
	if ok {
		onvifClient.RebootNeeded = bool(setNetworkInterfacesResponse.RebootNeeded)
		return
	}
}

func (onvifClient *OnvifClient) getNetworkInterfaces(device models.Device) (netInfo *onvifdevice.GetNetworkInterfacesResponse, edgexErr errors.EdgeX) {
	devInfoResponse, edgexErr := onvifClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetNetworkInterfaces, []byte{})
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfo, ok := devInfoResponse.(*onvifdevice.GetNetworkInterfacesResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetNetworkInterfacesResponse of type %T for the camera %s", devInfoResponse, device.Name), nil)
	}
	return devInfo, nil
}

func (onvifClient *OnvifClient) getDeviceInformation(device models.Device) (devInfo *onvifdevice.GetDeviceInformationResponse, edgexErr errors.EdgeX) {
	devInfoResponse, edgexErr := onvifClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetDeviceInformation, []byte{})
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devInfo, ok := devInfoResponse.(*onvifdevice.GetDeviceInformationResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetDeviceInformationResponse of type %T for the camera %s", devInfoResponse, device.Name), nil)
	}
	return devInfo, nil
}

func (onvifClient *OnvifClient) getEndpointReference(device models.Device) (devInfo *onvifdevice.GetEndpointReferenceResponse, edgexErr errors.EdgeX) {
	endpointRefResponse, edgexErr := onvifClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetEndpointReference, []byte{})
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	devEndpointRef, ok := endpointRefResponse.(*onvifdevice.GetEndpointReferenceResponse)
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid GetEndpointReferenceResponse of type %T for the camera %s", endpointRefResponse, device.Name), nil)
	}
	return devEndpointRef, nil
}
