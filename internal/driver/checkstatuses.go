// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"net"
	"sync"
	"time"

	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// checkStatuses loops through all registered devices and tries to determine the most accurate connection state
func (d *Driver) checkStatuses() {
	d.lc.Debug("checkStatuses has been called")
	wg := sync.WaitGroup{}
	for _, device := range service.RunningService().Devices() {
		device := device                  // save the device value within the closure
		if device.Name == d.serviceName { // skip control plane device
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			status := d.testConnectionMethods(device)
			if statusChanged, updateDeviceStatusErr := d.updateDeviceStatus(device.Name, status); updateDeviceStatusErr != nil {
				d.lc.Warnf("Could not update device status for device %s: %s", device.Name, updateDeviceStatusErr.Error())

			} else if statusChanged && status == UpWithAuth {
				d.lc.Infof("Device %s is now %s, refreshing the device information.", device.Name, UpWithAuth)
				go func() { // refresh the device information in the background
					refreshErr := d.refreshDeviceInformation(device)
					if refreshErr != nil {
						d.lc.Warnf("An error occurred while refreshing the device information for %s: %s",
							device.Name, refreshErr.Error())
					}

					refreshErr = d.refreshNetworkInterfaces(device)
					if refreshErr != nil {
						d.lc.Warnf("An error occurred while refreshing the network information for %s: %s",
							device.Name, refreshErr.Error())
					}
				}()
			}
		}()
	}
	wg.Wait()
}

// testConnectionMethods will try to determine the state using different device calls
// and return the most accurate status
// Higher degrees of connection are tested first, becuase if they
// succeed, the lower levels of connection will too
func (d *Driver) testConnectionMethods(device models.Device) (status string) {

	// sends get capabilities command to device (does not require credentials)
	devClient, edgexErr := d.newTemporaryOnvifClient(device)
	if edgexErr != nil {
		d.lc.Debugf("Connection to %s failed when creating client: %s", device.Name, edgexErr.Message())
		// onvif connection failed, so lets probe it
		if d.tcpProbe(device) {
			return Reachable
		}
		return Unreachable

	}

	// sends get device information command to device (requires credentials)
	_, edgexErr = devClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetDeviceInformation, []byte{})
	if edgexErr != nil {
		d.lc.Debugf("%s command failed for device %s when using authentication: %s", onvif.GetDeviceInformation, device.Name, edgexErr.Message())
		return UpWithoutAuth
	}

	return UpWithAuth
}

// tcpProbe attempts to make a connection to a specific ip and port list to determine
// if there is a service listening at that ip+port.
func (d *Driver) tcpProbe(device models.Device) bool {
	proto, ok := device.Protocols[OnvifProtocol]
	if !ok {
		d.lc.Warnf("Device %s is missing required %s protocol info, cannot send probe.", device.Name, OnvifProtocol)
		return false
	}
	addr := proto[Address]
	port := proto[Port]

	if addr == "" || port == "" {
		d.lc.Warnf("Device %s has no network address, cannot send probe.", device.Name)
		return false
	}
	host := addr + ":" + port

	conn, err := net.DialTimeout("tcp", host, time.Duration(d.config.AppCustom.ProbeTimeoutMillis)*time.Millisecond)
	if err != nil {
		d.lc.Debugf("Connection to %s failed when using simple tcp dial, Error: %s ", device.Name, err.Error())
		return false
	}
	defer conn.Close()
	return true
}

// updateDeviceStatus updates the status of a device in the cache. Returns true if the status changed. Returns any errors that occur if failure.
func (d *Driver) updateDeviceStatus(deviceName string, status string) (bool, error) {
	// todo: maybe have connection levels known as ints, so that way we can log at different levels based on
	// if the connection level went up or down
	shouldUpdate := false

	// lookup device from cache to ensure we are updating the latest version
	device, err := service.RunningService().GetDeviceByName(deviceName)
	if err != nil {
		d.lc.Errorf("Unable to get device %s from cache while trying to update its status to %s. Error: %s",
			device.Name, status, err.Error())
		return false, err
	}

	statusChanged := false
	oldStatus := device.Protocols[OnvifProtocol][DeviceStatus]
	if oldStatus != status {
		d.lc.Infof("Device status for %s is now %s (used to be %s)", device.Name, status, oldStatus)
		device.Protocols[OnvifProtocol][DeviceStatus] = status
		shouldUpdate = true
		statusChanged = true
	}

	if status != Unreachable {
		device.Protocols[OnvifProtocol][LastSeen] = time.Now().Format(time.UnixDate)
		shouldUpdate = true
	}

	if shouldUpdate {
		return statusChanged, service.RunningService().UpdateDevice(device)
	}

	return statusChanged, nil
}

// taskLoop manages all of our custom background tasks such as checking camera statuses at regular intervals
func (d *Driver) taskLoop() {
	d.configMu.RLock()
	interval := d.config.AppCustom.CheckStatusInterval
	d.configMu.RUnlock()
	if interval > maxStatusInterval { // check the interval
		d.lc.Warnf("Status interval of %d seconds is larger than the maximum value of %d seconds. Status interval has been set to the max value.", interval, maxStatusInterval)
		interval = maxStatusInterval
	}

	statusTicker := time.NewTicker(time.Duration(interval) * time.Second) // TODO: Support dynamic updates for ticker interval

	defer statusTicker.Stop()

	d.lc.Info("Starting task loop.")

	for {
		select {
		case <-d.taskCh:
			return
		case <-statusTicker.C:
			start := time.Now()
			d.checkStatuses() // checks the status of every device
			d.lc.Debugf("checkStatuses completed in: %v", time.Since(start))
		}
	}
}
