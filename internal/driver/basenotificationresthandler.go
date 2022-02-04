// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
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
	handlerContextKey  = "RestHandler"
)

type RestHandler struct {
	service     *sdk.DeviceService
	logger      logger.LoggingClient
	asyncValues chan<- *models.AsyncValues
}

func NewRestHandler(service *sdk.DeviceService, logger logger.LoggingClient, asyncValues chan<- *models.AsyncValues) *RestHandler {
	handler := RestHandler{
		service:     service,
		logger:      logger,
		asyncValues: asyncValues,
	}

	return &handler
}

func (handler RestHandler) Start() errors.EdgeX {
	if err := handler.service.AddRoute(apiResourceRoute, handler.addContext(deviceHandler), http.MethodPost); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("unable to add required route: %s: %s", apiResourceRoute, err.Error()), err)
	}

	handler.logger.Infof("Route %s added.", apiResourceRoute)
	return nil
}

func (handler RestHandler) addContext(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	// Add the context with the handler so the endpoint handling code can get back to this handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), handlerContextKey, handler)
		next(w, r.WithContext(ctx))
	})
}

// processAsyncRequest receives notification from Onvif camera and sends to the async reading channel
func (handler RestHandler) processAsyncRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	deviceName := vars[common.DeviceName]
	resourceName := vars[common.ResourceName]

	handler.logger.Debugf("Received POST for Device=%s Resource=%s", deviceName, resourceName)

	_, err := handler.service.GetDeviceByName(deviceName)
	if err != nil {
		handler.logger.Errorf("Incoming reading ignored. Device '%s' not found", deviceName)
		http.Error(writer, fmt.Sprintf("Device '%s' not found", deviceName), http.StatusNotFound)
		return
	}

	deviceResource, ok := handler.service.DeviceResource(deviceName, resourceName)
	if !ok {
		handler.logger.Errorf("Incoming reading ignored. Resource '%s' not found", resourceName)
		http.Error(writer, fmt.Sprintf("Resource '%s' not found", resourceName), http.StatusNotFound)
		return
	}

	data, err := handler.readBody(request)
	if err != nil {
		handler.logger.Errorf("Incoming reading ignored. Unable to read request body: %s", err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	notify := &event.Notify{}
	responseEnvelope := gosoap.NewSOAPEnvelope(notify)
	err = xml.Unmarshal(data, responseEnvelope)
	if err != nil {
		handler.logger.Errorf("Fail to create to the subscribe response for Device=%s Resource=%s, %s", deviceName, resourceName, err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	cv, err := sdkModel.NewCommandValue(deviceResource.Name, common.ValueTypeObject, notify)
	if err != nil {
		handler.logger.Errorf("Fail to create to the commandValue for Device=%s Resource=%s, %s", deviceName, resourceName, err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	asyncValues := &models.AsyncValues{
		DeviceName:    deviceName,
		CommandValues: []*models.CommandValue{cv},
	}

	handler.logger.Debugf("Incoming reading received: Device=%s Resource=%s", deviceName, resourceName)

	handler.asyncValues <- asyncValues
}

func (handler RestHandler) readBody(request *http.Request) ([]byte, error) {
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

func deviceHandler(writer http.ResponseWriter, request *http.Request) {
	handler, ok := request.Context().Value(handlerContextKey).(RestHandler)
	if !ok {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Bad context pass to handler"))
		return
	}

	handler.processAsyncRequest(writer, request)
}
