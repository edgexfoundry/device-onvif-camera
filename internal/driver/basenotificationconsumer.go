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

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/IOTechSystems/onvif"
	"github.com/IOTechSystems/onvif/event"
	"github.com/IOTechSystems/onvif/gosoap"
	"github.com/IOTechSystems/onvif/xsd"
)

// Consumer holds the subscription info and receive the camera event
type Consumer struct {
	Name        string
	lc          logger.LoggingClient
	onvifClient *OnvifClient
	manager     *BaseNotificationManager

	// subscriptionRequest is used to create the BaseNotification subscription
	subscriptionRequest *SubscriptionRequest
	// SubscriptionAddress is the reference for the event producer
	SubscriptionAddress string
	// Stopped indicates the Consumer should stop the subscription
	Stopped chan bool
}

// StartRenewLoop renews the subscription before termination time
func (consumer *Consumer) StartRenewLoop() {
	consumer.lc.Infof("Consumer starts the Renew loop for '%s'", consumer.Name)
	// Remove self when subscription finished or renew failed
	defer consumer.manager.removeConsumer(consumer)

	duration, err := ParseISO8601(*consumer.subscriptionRequest.InitialTerminationTime)
	if err != nil {
		consumer.lc.Infof("invalid Initial termination time, %v", err)
		return
	}

	// Send Renew request every ten second before termination time
	renewTime := duration - 10*time.Second
	renewTicker := time.NewTicker(renewTime)
	for {
		select {
		case <-consumer.Stopped:
			consumer.lc.Infof("Stopping the subscription '%s'", consumer.Name)
			return
		case <-renewTicker.C:
			consumer.lc.Debugf("Renewing the subscription from '%s' for resource '%s'", consumer.SubscriptionAddress, consumer.Name)
			renewRequest := consumer.createRawRequest()
			renewRequestData, err := xml.Marshal(renewRequest)
			if err != nil {
				consumer.lc.Errorf("Failed to marshal subscription request for resource '%s', %v", consumer.Name, err)
				return
			}

			servResp, err := consumer.onvifClient.onvifDevice.SendSoap(consumer.SubscriptionAddress, string(renewRequestData))
			if err != nil {
				consumer.lc.Warnf("Failed to send the renew request from '%s' for resource '%s', %v. The pull point expired or dropped, try to create a new one.", consumer.SubscriptionAddress, consumer.Name, err)
				err = consumer.subscribe()
				if err != nil {
					consumer.lc.Errorf("Failed to subscribe again for resource '%s', %v", consumer.Name, err)
					return
				}
			} else if servResp.StatusCode >= http.StatusBadRequest {
				response, err := renewResponse(servResp)
				if err != nil {
					consumer.lc.Errorf("Failed to parse response for '%s', %v", consumer.Name, err)
					return
				}
				consumer.lc.Warnf("Failed to renew the subscription from '%s' for resource '%s', status code: %s, err: %v. The pull point expired or dropped, try to create a new one.", consumer.SubscriptionAddress, consumer.Name, response.Body.Fault.String())
				err = consumer.subscribe()
				if err != nil {
					consumer.lc.Errorf("Failed to subscribe again for resource '%s', %v", consumer.Name, err)
					return
				}
			}
		}
	}
}

func (consumer *Consumer) subscribe() errors.EdgeX {
	subscribe := consumer.subscribeRequest()
	subscribeData, err := json.Marshal(subscribe)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to marshal subscription request for resource '%s'", consumer.Name), err)
	}
	serviceName := onvif.EventWebService
	functionName := onvif.Subscribe
	respContent, edgexErr := consumer.onvifClient.callOnvifFunction(serviceName, functionName, subscribeData)
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to subscribe again for resource '%s', %v", consumer.Name, err), edgexErr)
	}
	subscribeResponse, ok := respContent.(*event.SubscribeResponse)
	if !ok {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid SubscribeResponse of type %T for the camera %s", respContent, consumer.onvifClient.DeviceName), nil)
	}
	consumer.SubscriptionAddress = fmt.Sprint(subscribeResponse.SubscriptionReference.Address)
	return nil
}

func (consumer *Consumer) createRawRequest() *event.Renew {
	terminationTime := xsd.String(*consumer.subscriptionRequest.InitialTerminationTime)
	return &event.Renew{
		TerminationTime: terminationTime,
	}
}

func renewResponse(servResp *http.Response) (*gosoap.SOAPEnvelope, errors.EdgeX) {
	defer servResp.Body.Close()

	rsp, err := ioutil.ReadAll(servResp.Body)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	response := &event.RenewResponse{}
	responseEnvelope := gosoap.NewSOAPEnvelope(response)
	err = xml.Unmarshal(rsp, responseEnvelope)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return responseEnvelope, nil
}

func (consumer *Consumer) subscribeRequest() *event.Subscribe {
	filter := &event.FilterType{}
	if *consumer.subscriptionRequest.TopicFilter != "" {
		filter.TopicExpression = &event.TopicExpressionType{TopicKinds: xsd.String(*consumer.subscriptionRequest.TopicFilter)}
	}
	if *consumer.subscriptionRequest.MessageContentFilter != "" {
		filter.MessageContent = &event.QueryExpressionType{MessageKind: xsd.String(*consumer.subscriptionRequest.MessageContentFilter)}
	}
	InitialTerminationTime := xsd.String(*consumer.subscriptionRequest.InitialTerminationTime)
	subscriptionPolicy := xsd.String(*consumer.subscriptionRequest.SubscriptionPolicy)

	consumer.onvifClient.driver.configMu.RLock()
	baseNotificationURL := consumer.onvifClient.driver.config.AppCustom.BaseNotificationURL
	consumer.onvifClient.driver.configMu.RUnlock()

	address := fmt.Sprintf("%s%s/%s/%s/%s",
		baseNotificationURL, common.ApiBase, OnvifEventRestPath, consumer.onvifClient.DeviceName, consumer.onvifClient.CameraEventResource.Name)
	consumerReference := &event.EndpointReferenceType{
		Address: event.AttributedURIType(address),
	}
	return &event.Subscribe{
		ConsumerReference:  consumerReference,
		Filter:             filter,
		TerminationTime:    &InitialTerminationTime,
		SubscriptionPolicy: &subscriptionPolicy,
	}
}
