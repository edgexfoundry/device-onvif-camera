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

func NewPullPointManager(lc logger.LoggingClient) *PullPointManager {
	return &PullPointManager{
		lc:          lc,
		subscribers: make(map[string]*Subscriber),
		lock:        new(sync.RWMutex),
	}
}

func (manager *PullPointManager) NewSubscriber(deviceClient *DeviceClient, resourceName string, attributes map[string]interface{}, data []byte) errors.EdgeX {
	_, ok := manager.subscribers[resourceName]
	if ok {
		manager.lc.Infof("'%s' resource's Pull point subscriber already exists, skip adding new subscriber.", resourceName)
		return nil
	}

	request, edgexErr := subscriptionRequest(attributes, data)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	onvifDevice, err := manager.newSubscriberOnvifDevice(deviceClient.onvifDevice, *request.MessageTimeout)
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to create onvif device for pulling event, %v", err), edgexErr)
	}
	sub := &Subscriber{
		Name:                resourceName,
		lc:                  deviceClient.lc,
		manager:             manager,
		deviceClient:        deviceClient,
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
		return errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("fail to create the PullPoint subscription for resource '%s'", sub.Name), edgexErr)
	}
	manager.addSubscriber(sub)
	go sub.StartPullMessageLoop()
	return nil
}

func (manager *PullPointManager) newSubscriberOnvifDevice(device *onvif.Device, messageTimeout string) (*onvif.Device, error) {
	timeout, err := ParseISO8601(messageTimeout)
	if err != nil {
		return nil, err
	}
	timeout = timeout + time.Duration(driver.config.RequestTimeout)*time.Second
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

func (manager *PullPointManager) UnsubscribeAll() {
	for _, sub := range manager.subscribers {
		// subscriber will stop to pull message and unsubscribe the subscription when receiving the Stopped signal
		sub.Stopped <- true
	}
	manager.lc.Debug("Unsubscribe all subscriptions")
}
