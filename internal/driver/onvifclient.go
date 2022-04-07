// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/IOTechSystems/onvif"
	"github.com/IOTechSystems/onvif/device"
	"github.com/IOTechSystems/onvif/gosoap"
)

const (
	EdgeXWebService        = "EdgeX"
	RebootNeeded           = "RebootNeeded"
	CameraEvent            = "CameraEvent"
	SubscribeCameraEvent   = "SubscribeCameraEvent"
	UnsubscribeCameraEvent = "UnsubscribeCameraEvent"
)

// OnvifClient manages the state required to issue ONVIF requests to the specified camera
type OnvifClient struct {
	lc          logger.LoggingClient
	DeviceName  string
	cameraInfo  *CameraInfo
	onvifDevice *onvif.Device
	// RebootNeeded indicates the camera should reboot to apply the configuration change
	RebootNeeded bool
	// CameraEventResource is used to send the async event to north bound
	CameraEventResource     models.DeviceResource
	pullPointManager        *PullPointManager
	baseNotificationManager *BaseNotificationManager
}

// NewOnvifClient returns an OnvifClient for a single camera
func NewOnvifClient(device models.Device, driverConfig *configuration, lc logger.LoggingClient) (*OnvifClient, errors.EdgeX) {
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

	resource, err := getCameraEventResourceByDeviceName(device.Name)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	client := &OnvifClient{
		lc:                  lc,
		DeviceName:          device.Name,
		cameraInfo:          cameraInfo,
		onvifDevice:         dev,
		CameraEventResource: resource,
	}
	// Create PullPointManager to control multiple pull points
	pullPointManager := NewPullPointManager(lc)
	client.pullPointManager = pullPointManager

	// Create BaseNotificationManager to control multiple notification consumer
	baseNotificationManager := NewBaseNotificationManager(lc)
	client.baseNotificationManager = baseNotificationManager
	return client, nil
}

func getCameraEventResourceByDeviceName(deviceName string) (r models.DeviceResource, edgexErr errors.EdgeX) {
	deviceService := sdk.RunningService()
	device, err := deviceService.GetDeviceByName(deviceName)
	if err != nil {
		return r, errors.NewCommonEdgeXWrapper(err)
	}
	profile, err := deviceService.GetProfileByName(device.ProfileName)
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

func deviceAddress(cameraInfo *CameraInfo) string {
	return fmt.Sprintf("%s:%d", cameraInfo.Address, cameraInfo.Port)
}

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
		cv, edgexErr := onvifClient.callCustomFunction(req.DeviceResourceName, serviceName, functionName, req.Attributes, data)
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
	}
	if functionName == onvif.SystemReboot {
		onvifClient.RebootNeeded = false
	}
	cv, err := sdkModel.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, responseContent)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to create commandValue for the function '%s' of web service '%s' ", functionName, serviceName), err)
	}
	return cv, nil
}

func (onvifClient *OnvifClient) callCustomFunction(resourceName, serviceName, functionName string, attributes map[string]interface{}, data []byte) (cv *sdkModel.CommandValue, edgexErr errors.EdgeX) {
	var err error
	switch functionName {
	case RebootNeeded:
		cv, err = sdkModel.NewCommandValue(resourceName, common.ValueTypeBool, onvifClient.RebootNeeded)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to create commandValue for the web service '%s' function '%s'", serviceName, functionName), err)
		}
	case SubscribeCameraEvent:
		err = onvifClient.callSubscribeCameraEventFunction(resourceName, serviceName, functionName, attributes, data)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
	case UnsubscribeCameraEvent:
		go func() {
			onvifClient.lc.Debugf("Unsubscribe camera event for the device '%v'", onvifClient.DeviceName)
			onvifClient.pullPointManager.UnsubscribeAll()
			onvifClient.baseNotificationManager.UnsubscribeAll()
		}()
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
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to create commandValue for the web service '%s' function '%s'", serviceName, functionName), edgexErr)
		}
	case BaseNotification:
		edgexErr = onvifClient.baseNotificationManager.NewConsumer(onvifClient, resourceName, attributes, data)
		if edgexErr != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to create commandValue for the web service '%s' function '%s'", serviceName, functionName), edgexErr)
		}
	default:
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported subscribeType '%s'", subscribeType), nil)
	}
	return nil
}

func (onvifClient *OnvifClient) callOnvifFunction(serviceName, functionName string, data []byte) (interface{}, errors.EdgeX) {
	function, edgexErr := onvif.FunctionByServiceAndFunctionName(serviceName, functionName)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	request, edgexErr := createRequest(function, data)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to create '%s' request for the web service '%s'", functionName, serviceName), edgexErr)
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
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to send the '%s' request for the web service '%s'", functionName, serviceName), err)
	}
	defer servResp.Body.Close()

	rsp, err := ioutil.ReadAll(servResp.Body)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	responseEnvelope, edgexErr := createResponse(function, rsp)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to create '%s' response for the web service '%s'", functionName, serviceName), edgexErr)
	}
	res, _ := xml.Marshal(responseEnvelope.Body.Content)
	onvifClient.lc.Debugf("SOAP Response: %v", string(res))

	if servResp.StatusCode == http.StatusUnauthorized {
		return nil, errors.NewCommonEdgeX(errors.KindInvalidId,
			fmt.Sprintf("fail to verify the authentication for the function '%s' of web service '%s'. Onvif error: %s",
				functionName, serviceName, responseEnvelope.Body.Fault.String()), nil)
	} else if servResp.StatusCode == http.StatusBadRequest {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("invalid request for the function '%s' of web service '%s'. Onvif error: %s",
				functionName, serviceName, responseEnvelope.Body.Fault.String()), nil)
	} else if servResp.StatusCode > http.StatusNoContent {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("fail to execute the request for the function '%s' of web service '%s'. Onvif error: %s",
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
	setNetworkInterfacesResponse, ok := responseContent.(*device.SetNetworkInterfacesResponse)
	if ok {
		onvifClient.RebootNeeded = bool(setNetworkInterfacesResponse.RebootNeeded)
		return
	}
}
