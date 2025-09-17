// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
// Copyright (c) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/IOTechSystems/onvif"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/spf13/cast"
)

// checkStatuses loops through all registered devices and tries to determine the most accurate connection state
func (d *Driver) checkStatuses() {
	d.lc.Debug("checkStatuses has been called")
	start := time.Now()
	defer func() {
		d.lc.Debugf("checkStatuses completed in: %v", time.Since(start))
	}()

	wg := sync.WaitGroup{}
	for _, device := range d.sdkService.Devices() {
		device := device // save the device value within the closure

		wg.Add(1)
		go func() {
			defer wg.Done()

			d.checkStatusOfDevice(device)
		}()
	}
	wg.Wait()
}

// checkStatusOfDevice checks the status of an individual device
func (d *Driver) checkStatusOfDevice(device models.Device) {
	d.lc.Debugf("checking status of device %s", device.Name)

	// if device is unknown, and missing a MAC Address, try and determine the MAC address via the endpoint reference
	if strings.HasPrefix(device.Name, UnknownDevicePrefix) && device.Protocols[OnvifProtocol][MACAddress] == "" {
		if v, ok := device.Protocols[OnvifProtocol][EndpointRefAddress]; ok {
			endpointRefAddr := cast.ToString(v)
			if endpointRefAddr != "" {
				if mac := d.macAddressMapper.MatchEndpointRefAddressToMAC(endpointRefAddr); mac != "" {
					// the mac address for the device was found, so set it here which will allow the
					// code below to use the mac address for looking up the credentials. Because the mac mapper
					// already contains them, the credentials will be found (whether they are valid or invalid).
					device.Protocols[OnvifProtocol][MACAddress] = mac
				}
			}
		}
	}

	status := d.testConnectionMethods(device)
	if statusChanged, updateDeviceStatusErr := d.updateDeviceStatus(device.Name, status); updateDeviceStatusErr != nil {
		d.lc.Warnf("Could not update device status for device %s: %s", device.Name, updateDeviceStatusErr.Error())

	} else if statusChanged && status == UpWithAuth {
		d.lc.Infof("Device %s is now %s, refreshing the device information.", device.Name, UpWithAuth)
		go func() { // refresh the device information in the background
			if refreshErr := d.refreshDevice(device); refreshErr != nil {
				d.lc.Errorf("An error occurred while refreshing the device %s: %s",
					device.Name, refreshErr.Error())
			}
		}()
	}

	d.lc.Debugf("device %s status is %s", device.Name, status)
}

// testConnectionMethods will try to determine the state using different device calls
// and return the most accurate status
// Higher degrees of connection are tested first, because if they
// succeed, the lower levels of connection will too
func (d *Driver) testConnectionMethods(device models.Device) (status string) {
	devClient, err := d.getOrCreateOnvifClient(device)
	if err != nil {
		d.lc.Warnf("Error getting onvif client for device %s", device.Name)
		// if we do not have a valid onvif client, lets just tcp probe it
		if d.tcpProbe(device) {
			return Reachable
		}
		return Unreachable
	}

	// sends GetDeviceInformation command to device (requires authentication)
	_, edgexErr := devClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetDeviceInformation, []byte{})
	if edgexErr == nil {
		return UpWithAuth // we are authenticated
	}
	d.lc.Debugf("%s command failed for device %s when using authentication: %s", onvif.GetDeviceInformation, device.Name, edgexErr.Message())

	// sends GetSystemDateAndTime command to device (does not require authentication)
	_, edgexErr = devClient.callOnvifFunction(onvif.DeviceWebService, onvif.GetSystemDateAndTime, []byte{})
	if edgexErr == nil {
		return UpWithoutAuth // non-authenticated onvif command is working
	}
	d.lc.Debugf("%s command failed for device %s without using authentication: %s", onvif.GetSystemDateAndTime, device.Name, edgexErr.Message())

	// onvif commands are not working, so let us probe it
	if d.tcpProbe(device) {
		return Reachable
	}
	return Unreachable
}

// tcpProbe attempts to make a connection to a specific ip and port list to determine
// if there is a service listening at that ip+port.
func (d *Driver) tcpProbe(device models.Device) bool {
	xAddr, edgexErr := GetCameraXAddr(device.Protocols)
	if edgexErr != nil {
		d.lc.Warnf("Device %s is missing required %s protocol info, cannot send probe: %v", device.Name, OnvifProtocol, edgexErr)
		return false
	}

	conn, err := net.DialTimeout("tcp", xAddr, time.Duration(d.config.AppCustom.ProbeTimeoutMillis)*time.Millisecond)
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
	//       if the connection level went up or down
	shouldUpdate := false

	// lookup device from cache to ensure we are updating the latest version
	device, err := d.sdkService.GetDeviceByName(deviceName)
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
		return statusChanged, d.sdkService.PatchDevice(dtos.UpdateDevice{
			Name:      &deviceName,
			Protocols: dtos.FromProtocolModelsToDTOs(device.Protocols),
		})
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
			d.checkStatuses() // checks the status of every device
		}
	}
}
