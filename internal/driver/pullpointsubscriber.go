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
	"io"
	"net/http"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/IOTechSystems/onvif"
	"github.com/IOTechSystems/onvif/event"
	"github.com/IOTechSystems/onvif/xsd"
)

type Subscriber struct {
	Name        string
	manager     *PullPointManager
	onvifClient *OnvifClient

	// onvifDevice is used to send the pullMessage onvif function with specified request timeout
	onvifDevice OnvifDevice
	// SubscriptionAddress is used to pull the event from the camera
	SubscriptionAddress string
	// subscriptionRequest is used to create the PullPoint subscription
	subscriptionRequest *SubscriptionRequest
	// pullMessageRequestBody is the pullMessage onvif function's request body
	pullMessageRequestBody event.PullMessages
	// Stopped indicates the Subscriber should stop the PullMessageLoop
	Stopped chan bool
}

// StartPullMessageLoop implements the long-polling strategy to pull the camera event
func (sub *Subscriber) StartPullMessageLoop() {
	sub.onvifClient.lc.Infof("Subscriber starts the PullMessage loop for '%s'", sub.Name)
	// Remove self when subscription finished or pull message failed
	defer sub.manager.removeSubscriber(sub)
	for {
		select {
		case <-sub.Stopped:
			sub.onvifClient.lc.Infof("Removing the subscription '%s'", sub.Name)
			edgexErr := sub.unsubscribe()
			if edgexErr != nil {
				sub.onvifClient.lc.Warnf(edgexErr.Message())
				return
			}
			return
		default:
			sub.onvifClient.lc.Debugf("Pull the event from '%s' for resource '%s'", sub.SubscriptionAddress, sub.Name)
			// The camera will block the request according to the SubscribeCameraEvent's MessageTimeout
			// and the device service will renew the expired subscription if AutoRenew is enabled.
			edgexErr := sub.pullMessage()
			if edgexErr != nil {
				sub.onvifClient.lc.Warnf(edgexErr.Message())
				return
			}
		}
	}
}

func (sub *Subscriber) pullMessage() errors.EdgeX {
	requestBody, err := xml.Marshal(sub.pullMessageRequestBody)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to marshal the PullMessage request for '%s', %v", sub.Name, err), err)
	}
	servResp, err := sub.onvifDevice.SendSoap(sub.SubscriptionAddress, string(requestBody))
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to send the '%s' pull event message request, %v", onvif.PullMessages, err), err)
	}
	defer servResp.Body.Close()
	if *sub.subscriptionRequest.AutoRenew && (servResp.StatusCode == http.StatusNotFound || servResp.StatusCode == http.StatusBadRequest) {
		sub.onvifClient.lc.Warnf("The pull point expired, try to create a new one")

		edgexErr := sub.createPullPoint()
		if edgexErr != nil {
			return errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to create the PullPoint subscription for resource '%s'", sub.Name), edgexErr)
		}
		return nil
	}

	rsp, err := io.ReadAll(servResp.Body)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to read the PullMessage response for '%s', %v", sub.Name, err), err)
	}

	function := &event.PullMessagesFunction{}
	response, edgexErr := createResponse(function, rsp)
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get the PullMessage response for '%s', %v", sub.Name, edgexErr), edgexErr)
	}

	res, ok := response.Body.Content.(*event.PullMessagesResponse)
	if !ok {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid PullMessagesResponse of type %T for the camera %s", response.Body.Content, sub.onvifClient.DeviceName), nil)
	}
	if len(res.NotificationMessage) == 0 {
		return nil
	}
	cv, err := sdkModel.NewCommandValue(sub.onvifClient.CameraEventResource.Name, common.ValueTypeObject, response.Body.Content)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create commandValue  for '%s', %v", sub.Name, err), err)
	}
	asyncValues := &sdkModel.AsyncValues{
		DeviceName:    sub.onvifClient.DeviceName,
		CommandValues: []*sdkModel.CommandValue{cv},
	}

	sub.onvifClient.driver.asynchCh <- asyncValues
	return nil
}

func (sub *Subscriber) createPullPoint() errors.EdgeX {
	serviceName := onvif.EventWebService
	functionName := onvif.CreatePullPointSubscription
	subscription := sub.createPullPointSubscription()
	subscriptionData, err := json.Marshal(subscription)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal subscription request for resource", err)
	}
	respContent, edgexErr := sub.onvifClient.callOnvifFunction(serviceName, functionName, subscriptionData)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	subscriptionResponse, ok := respContent.(*event.CreatePullPointSubscriptionResponse)
	if !ok {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("invalid CreatePullPointSubscriptionResponse of type %T for the camera %s", respContent, sub.onvifClient.DeviceName), nil)
	}
	sub.SubscriptionAddress = fmt.Sprint(subscriptionResponse.SubscriptionReference.Address)
	return nil
}

func (sub *Subscriber) createPullPointSubscription() *event.CreatePullPointSubscription {
	filter := &event.FilterType{}
	if sub.subscriptionRequest.TopicFilter != nil {
		filter.TopicExpression = &event.TopicExpressionType{TopicKinds: xsd.String(*sub.subscriptionRequest.TopicFilter)}
	}
	if sub.subscriptionRequest.MessageContentFilter != nil {
		filter.MessageContent = &event.QueryExpressionType{MessageKind: xsd.String(*sub.subscriptionRequest.MessageContentFilter)}
	}
	InitialTerminationTime := xsd.String(*sub.subscriptionRequest.InitialTerminationTime)
	subscriptionPolicy := xsd.String(*sub.subscriptionRequest.SubscriptionPolicy)
	return &event.CreatePullPointSubscription{
		Filter:                 filter,
		InitialTerminationTime: &InitialTerminationTime,
		SubscriptionPolicy:     &subscriptionPolicy,
	}
}

func (sub *Subscriber) unsubscribe() errors.EdgeX {
	request := event.Unsubscribe{}
	requestBody, err := xml.Marshal(request)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to marshal the unsubscribe request for '%s', %v", sub.Name, err), err)
	}
	_, edgexErr := sub.onvifClient.onvifDevice.SendSoap(sub.SubscriptionAddress, string(requestBody))
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	sub.onvifClient.lc.Debugf("Unsubscribe the subscription '%s' from %s", sub.Name, sub.SubscriptionAddress)
	return nil
}
