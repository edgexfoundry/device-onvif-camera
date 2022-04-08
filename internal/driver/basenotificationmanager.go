// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// BaseNotificationManager manages the consumers to renew the subscription
type BaseNotificationManager struct {
	lc        logger.LoggingClient
	lock      *sync.RWMutex
	consumers map[string]*Consumer
}

// NewBaseNotificationManager create the new BaseNotificationManager entity
func NewBaseNotificationManager(lc logger.LoggingClient) *BaseNotificationManager {
	return &BaseNotificationManager{
		lc:        lc,
		consumers: make(map[string]*Consumer),
		lock:      new(sync.RWMutex),
	}
}

// NewConsumer create the new NewConsumer entity and send the subscription request to the camera
func (manager *BaseNotificationManager) NewConsumer(onvifClient *OnvifClient, resourceName string, attributes map[string]interface{}, data []byte) errors.EdgeX {
	_, ok := manager.consumers[resourceName]
	if ok {
		manager.lc.Warnf("'%s' resource's base notification consumer already exists, skip adding new subscriber.", resourceName)
		return nil
	}

	request, edgexErr := newSubscriptionRequest(attributes, data)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	consumer := &Consumer{
		Name:                resourceName,
		lc:                  onvifClient.lc,
		onvifClient:         onvifClient,
		manager:             manager,
		subscriptionRequest: request,
		Stopped:             make(chan bool),
	}
	edgexErr = consumer.subscribe()
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to create the BaseNotification for resource '%s'", consumer.Name), edgexErr)
	}
	manager.addConsumer(consumer)
	if *request.AutoRenew {
		go consumer.StartRenewLoop()
	}
	return nil
}

func (manager *BaseNotificationManager) addConsumer(consumer *Consumer) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.consumers[consumer.Name] = consumer
}

func (manager *BaseNotificationManager) removeConsumer(consumer *Consumer) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	delete(manager.consumers, consumer.Name)
}

func (manager *BaseNotificationManager) UnsubscribeAll() {
	for _, consumer := range manager.consumers {
		// consumer will stop to renew the subscription when receiving the Stopped signal
		consumer.Stopped <- true
	}
	manager.lc.Debug("Unsubscribe all subscriptions")
}
