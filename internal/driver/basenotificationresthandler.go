// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/IOTechSystems/onvif/event"
	"github.com/IOTechSystems/onvif/gosoap"

	"github.com/gorilla/mux"
)

const (
	OnvifEventRestPath = "onvifevent"
	apiResourceRoute   = common.ApiBase + "/" + OnvifEventRestPath + "/{" + common.DeviceName + "}/{" + common.ResourceName + "}"
)

// RestNotificationHandler handle the notification from the camera and send to async value channel
type RestNotificationHandler struct {
	sdkService  SDKService
	lc          logger.LoggingClient
	asyncValues chan<- *models.AsyncValues
}

// NewRestNotificationHandler create a new RestNotificationHandler entity
func NewRestNotificationHandler(service SDKService, logger logger.LoggingClient, asyncValues chan<- *models.AsyncValues) *RestNotificationHandler {
	handler := RestNotificationHandler{
		sdkService:  service,
		lc:          logger,
		asyncValues: asyncValues,
	}

	return &handler
}

// AddRoute adds route for receiving the notification from the camera
func (handler RestNotificationHandler) AddRoute() errors.EdgeX {
	if err := handler.sdkService.AddRoute(apiResourceRoute, handler.processAsyncRequest, http.MethodPost); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("unable to add required route: %s: %s", apiResourceRoute, err.Error()), err)
	}

	handler.lc.Infof("Route %s added.", apiResourceRoute)
	return nil
}

// processAsyncRequest receives notification from Onvif camera and sends to the async reading channel
func (handler RestNotificationHandler) processAsyncRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	deviceName := vars[common.DeviceName]
	resourceName := vars[common.ResourceName]

	handler.lc.Debugf("Received POST for Device=%s Resource=%s", deviceName, resourceName)

	_, err := handler.sdkService.GetDeviceByName(deviceName)
	if err != nil {
		handler.lc.Errorf("Incoming reading ignored. Device '%s' not found", deviceName)
		http.Error(writer, fmt.Sprintf("Device '%s' not found", deviceName), http.StatusBadRequest)
		return
	}

	deviceResource, ok := handler.sdkService.DeviceResource(deviceName, resourceName)
	if !ok {
		handler.lc.Errorf("Incoming reading ignored. Resource '%s' not found", resourceName)
		http.Error(writer, fmt.Sprintf("Resource '%s' not found", resourceName), http.StatusBadRequest)
		return
	}

	data, err := handler.readBody(request)
	if err != nil {
		handler.lc.Errorf("Incoming reading ignored. Unable to read request body: %s", err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	notify := &event.Notify{}
	responseEnvelope := gosoap.NewSOAPEnvelope(notify)
	err = xml.Unmarshal(data, responseEnvelope)
	if err != nil {
		handler.lc.Errorf("Failed to create to the subscribe response for Device=%s Resource=%s, %s", deviceName, resourceName, err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	cv, err := sdkModel.NewCommandValue(deviceResource.Name, common.ValueTypeObject, notify)
	if err != nil {
		handler.lc.Errorf("Failed to create to the commandValue for Device=%s Resource=%s, %s", deviceName, resourceName, err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	asyncValues := &models.AsyncValues{
		DeviceName:    deviceName,
		CommandValues: []*models.CommandValue{cv},
	}

	handler.lc.Debugf("Incoming reading received: Device=%s Resource=%s", deviceName, resourceName)

	handler.asyncValues <- asyncValues
}

func (handler RestNotificationHandler) readBody(request *http.Request) ([]byte, error) {
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("no request body provided")
	}

	return body, nil
}
