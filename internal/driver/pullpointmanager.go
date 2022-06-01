// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/IOTechSystems/onvif"
	"github.com/IOTechSystems/onvif/event"
	"github.com/IOTechSystems/onvif/xsd"
)

// PullPointManager manages the subscribers to pull event from specified PullPoints
type PullPointManager struct {
	lc          logger.LoggingClient
	lock        *sync.RWMutex
	subscribers map[string]*Subscriber
}

// newPullPointManager create a new PullPointManager entity
func newPullPointManager(lc logger.LoggingClient) *PullPointManager {
	return &PullPointManager{
		lc:          lc,
		subscribers: make(map[string]*Subscriber),
		lock:        new(sync.RWMutex),
	}
}

// NewSubscriber creates a new subscriber entity and start pulling the event from the camera
func (manager *PullPointManager) NewSubscriber(onvifClient *OnvifClient, resourceName string, attributes map[string]interface{}, data []byte) errors.EdgeX {
	_, ok := manager.subscribers[resourceName]
	if ok {
		manager.lc.Warnf("'%s' resource's Pull point subscriber already exists, skip adding new subscriber.", resourceName)
		return nil
	}

	request, edgexErr := newSubscriptionRequest(attributes, data)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	onvifClient.driver.configMu.RLock()
	requestTimeout := onvifClient.driver.config.AppCustom.RequestTimeout
	onvifClient.driver.configMu.RUnlock()

	onvifDevice, err := manager.newSubscriberOnvifDevice(onvifClient.onvifDevice, *request.MessageTimeout, requestTimeout)
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to create onvif device for pulling event, %v", err), edgexErr)
	}
	sub := &Subscriber{
		Name:                resourceName,
		manager:             manager,
		onvifClient:         onvifClient,
		onvifDevice:         onvifDevice,
		subscriptionRequest: request,
		pullMessageRequestBody: event.PullMessages{
			Timeout:      xsd.Duration(*request.MessageTimeout),
			MessageLimit: xsd.Int(*request.MessageLimit),
		},
		Stopped: make(chan bool),
	}
	edgexErr = sub.createPullPoint()
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to create the PullPoint subscription for resource '%s'", sub.Name), edgexErr)
	}
	manager.addSubscriber(sub)
	go sub.StartPullMessageLoop()
	return nil
}

func (manager *PullPointManager) newSubscriberOnvifDevice(device *onvif.Device, messageTimeout string, httpRequestTimeout int) (*onvif.Device, error) {
	timeout, err := ParseISO8601(messageTimeout)
	if err != nil {
		return nil, err
	}
	timeout = timeout + time.Duration(httpRequestTimeout)*time.Second
	params := device.GetDeviceParams()
	params.HttpClient = &http.Client{
		Timeout: timeout,
	}
	return onvif.NewDevice(params)
}

func (manager *PullPointManager) addSubscriber(sub *Subscriber) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.subscribers[sub.Name] = sub
}

func (manager *PullPointManager) removeSubscriber(sub *Subscriber) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	delete(manager.subscribers, sub.Name)
}

// UnsubscribeAll stops all subscriptions
func (manager *PullPointManager) UnsubscribeAll() {
	for _, sub := range manager.subscribers {
		// subscriber will stop to pull message and unsubscribe the subscription when receiving the Stopped signal
		sub.Stopped <- true
	}
	manager.lc.Debug("Unsubscribe all subscriptions")
}
