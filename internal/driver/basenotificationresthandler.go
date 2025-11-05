// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/IOTechSystems/onvif/event"
	"github.com/IOTechSystems/onvif/gosoap"

	"github.com/labstack/echo/v4"
)

const (
	OnvifEventRestPath = "onvifevent"
	apiResourceRoute   = common.ApiBase + "/" + OnvifEventRestPath + "/:deviceName/:resourceName"
)

// RestNotificationHandler handle the notification from the camera and send to async value channel
type RestNotificationHandler struct {
	sdkService interfaces.DeviceServiceSDK
	lc         logger.LoggingClient
}

// NewRestNotificationHandler create a new RestNotificationHandler entity
func NewRestNotificationHandler(service interfaces.DeviceServiceSDK) *RestNotificationHandler {
	handler := RestNotificationHandler{
		sdkService: service,
		lc:         service.LoggingClient(),
	}
	return &handler
}

// AddRoute adds route for receiving the notification from the camera
func (handler RestNotificationHandler) AddRoute() errors.EdgeX {
	if err := handler.sdkService.AddCustomRoute(apiResourceRoute, interfaces.Authenticated, handler.processAsyncRequest, http.MethodPost); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("unable to add required route: %s: %s", apiResourceRoute, err.Error()), err)
	}

	handler.lc.Infof("Route %s added.", apiResourceRoute)
	return nil
}

// processAsyncRequest receives notification from Onvif camera and sends to the async reading channel
func (handler RestNotificationHandler) processAsyncRequest(c echo.Context) error {
	deviceName := c.Param(common.DeviceName)
	resourceName := c.Param(common.ResourceName)

	handler.lc.Debugf("Received POST for Device=%s Resource=%s", deviceName, resourceName)

	_, err := handler.sdkService.GetDeviceByName(deviceName)
	if err != nil {
		handler.lc.Errorf("Incoming reading ignored. Device '%s' not found", deviceName)
		return c.String(http.StatusBadRequest, fmt.Sprintf("Device '%s' not found", deviceName))
	}

	deviceResource, ok := handler.sdkService.DeviceResource(deviceName, resourceName)
	if !ok {
		handler.lc.Errorf("Incoming reading ignored. Resource '%s' not found", resourceName)
		return c.String(http.StatusBadRequest, fmt.Sprintf("Resource '%s' not found", resourceName))
	}

	data, err := handler.readBody(c.Request())
	if err != nil {
		handler.lc.Errorf("Incoming reading ignored. Unable to read request body: %s", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}

	notify := &event.Notify{}
	responseEnvelope := gosoap.NewSOAPEnvelope(notify)
	err = xml.Unmarshal(data, responseEnvelope)
	if err != nil {
		handler.lc.Errorf("Failed to create to the subscribe response for Device=%s Resource=%s, %s", deviceName, resourceName, err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}

	cv, err := models.NewCommandValue(deviceResource.Name, common.ValueTypeObject, notify)
	if err != nil {
		handler.lc.Errorf("Failed to create to the commandValue for Device=%s Resource=%s, %s", deviceName, resourceName, err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	asyncValues := &models.AsyncValues{
		DeviceName:    deviceName,
		CommandValues: []*models.CommandValue{cv},
	}

	handler.lc.Debugf("Incoming reading received: Device=%s Resource=%s", deviceName, resourceName)

	handler.sdkService.AsyncValuesChannel() <- asyncValues

	return nil
}

func (handler RestNotificationHandler) readBody(request *http.Request) ([]byte, error) {
	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("no request body provided")
	}

	return body, nil
}
