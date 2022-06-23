// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	stdErrors "errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/IOTechSystems/onvif"
	wsdiscovery "github.com/IOTechSystems/onvif/ws-discovery"
	"github.com/edgexfoundry/device-onvif-camera/internal/netscan"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/gofrs/uuid"
)

const (
	bufSize = 8192
)

// OnvifProtocolDiscovery implements netscan.ProtocolSpecificDiscovery
type OnvifProtocolDiscovery struct {
	driver *Driver
}

func NewOnvifProtocolDiscovery(driver *Driver) *OnvifProtocolDiscovery {
	return &OnvifProtocolDiscovery{driver: driver}
}

// ProbeFilter takes in a host and a slice of ports to be scanned. It should return a slice
// of ports to actually scan, or a nil/empty slice if the host is to not be scanned at all.
// Can be used to filter out known devices from being probed again if required.
func (proto *OnvifProtocolDiscovery) ProbeFilter(_ string, ports []string) []string {
	// For onvif we do not want to do any filtering
	return ports
}

// OnConnectionDialed handles the protocol specific verification if there is actually
// a valid device or devices at the other end of the connection.
func (proto *OnvifProtocolDiscovery) OnConnectionDialed(host string, port string, conn net.Conn, params netscan.Params) ([]netscan.ProbeResult, error) {
	// attempt a basic direct probe approach using the open connection
	devices, err := executeRawProbe(conn, params)
	if err != nil {
		params.Logger.Debug(err.Error())
	} else if len(devices) > 0 {
		return mapProbeResults(host, port, devices), nil
	}
	return nil, err
}

// ConvertProbeResult takes a raw ProbeResult and transforms it into a
// processed DiscoveredDevice struct.
func (proto *OnvifProtocolDiscovery) ConvertProbeResult(probeResult netscan.ProbeResult, params netscan.Params) (sdkModel.DiscoveredDevice, error) {
	onvifDevice, ok := probeResult.Data.(onvif.Device)
	if !ok {
		return sdkModel.DiscoveredDevice{}, fmt.Errorf("unable to cast probe result into onvif.Device. type=%T", probeResult.Data)
	}

	discovered, err := proto.driver.createDiscoveredDevice(onvifDevice)
	if err != nil {
		return sdkModel.DiscoveredDevice{}, err
	}

	return discovered, nil
}

// createDiscoveredDevice will take an onvif.Device that was detected on the network and
// attempt to get more information about the device and create an EdgeX compatible DiscoveredDevice.
func (d *Driver) createDiscoveredDevice(onvifDevice onvif.Device) (sdkModel.DiscoveredDevice, error) {
	xaddr := onvifDevice.GetDeviceParams().Xaddr
	endpointRefAddr := onvifDevice.GetDeviceParams().EndpointRefAddress
	if endpointRefAddr == "" {
		d.lc.Warnf("The EndpointRefAddress is empty from the Onvif camera, unable to add the camera %s", xaddr)
		return sdkModel.DiscoveredDevice{}, fmt.Errorf("empty EndpointRefAddress for XAddr %s", xaddr)
	}
	address, port := addressAndPort(xaddr)
	timestamp := time.Now().Format(time.UnixDate)

	d.configMu.RLock()
	device := contract.Device{
		// Using Xaddr as the temporary name
		Name: xaddr,
		Protocols: map[string]contract.ProtocolProperties{
			OnvifProtocol: {
				Address:            address,
				Port:               port,
				SecretPath:         d.config.AppCustom.DefaultSecretPath,
				EndpointRefAddress: endpointRefAddr,
				DeviceStatus:       Reachable,
				LastSeen:           timestamp,
			},
		},
	}
	d.configMu.RUnlock()

	devInfo, edgexErr := d.getDeviceInformation(device)
	if edgexErr != nil {
		// try again using the endpointRefAddr as the SecretPath. the reason for this is that
		// the user may have pre-filled the secret store with per-device credentials based on the endpointRefAddr.
		device.Protocols[OnvifProtocol][SecretPath] = endpointRefAddr
		devInfo, edgexErr = d.getDeviceInformation(device)
	}

	var discovered sdkModel.DiscoveredDevice
	if edgexErr != nil {
		d.lc.Warnf("failed to get the device information for the camera %s, %v", endpointRefAddr, edgexErr)
		device.Protocols[OnvifProtocol][DeviceStatus] = Reachable // update device status in this error case
		discovered = sdkModel.DiscoveredDevice{
			Name:        endpointRefAddr,
			Protocols:   device.Protocols,
			Description: "Auto discovered Onvif camera",
			Labels:      []string{"auto-discovery"},
		}
		d.lc.Debugf("Discovered unknown camera '%s' from the address '%s'", discovered.Name, xaddr)
	} else {
		device.Protocols[OnvifProtocol][Manufacturer] = devInfo.Manufacturer
		device.Protocols[OnvifProtocol][Model] = devInfo.Model
		device.Protocols[OnvifProtocol][FirmwareVersion] = devInfo.FirmwareVersion
		device.Protocols[OnvifProtocol][SerialNumber] = devInfo.SerialNumber
		device.Protocols[OnvifProtocol][HardwareId] = devInfo.HardwareId
		device.Protocols[OnvifProtocol][DeviceStatus] = UpWithAuth
		device.Protocols[OnvifProtocol][LastSeen] = time.Now().Format(time.UnixDate)

		netInfo, err := d.getNetworkInterfaces(device)
		if err != nil {
			d.lc.Warnf("failed to get the network information for the camera %s, %v", endpointRefAddr, edgexErr)
		} else {
			device.Protocols[OnvifProtocol][MACAddress] = string(netInfo.NetworkInterfaces.Info.HwAddress)
		}

		// Spaces are not allowed in the device name
		deviceName := fmt.Sprintf("%s-%s-%s",
			strings.ReplaceAll(devInfo.Manufacturer, " ", "-"),
			strings.ReplaceAll(devInfo.Model, " ", "-"),
			endpointRefAddr)

		discovered = sdkModel.DiscoveredDevice{
			Name:        deviceName,
			Protocols:   device.Protocols,
			Description: fmt.Sprintf("%s %s Camera", devInfo.Manufacturer, devInfo.Model),
			Labels:      []string{"auto-discovery", devInfo.Manufacturer, devInfo.Model},
		}
		d.lc.Debugf("Discovered camera '%s' from the address '%s'", discovered.Name, xaddr)
	}
	return discovered, nil
}

// mapProbeResults converts a slice of discovered onvif.Device into the generic
// netscan.ProbeResult.
func mapProbeResults(host, port string, devices []onvif.Device) (res []netscan.ProbeResult) {
	for _, device := range devices {
		res = append(res, netscan.ProbeResult{
			Host: host,
			Port: port,
			Data: device,
		})
	}
	return res
}

// executeRawProbe essentially performs a UDP unicast ws-discovery probe by sending the
// probe message directly over the connection and listening for any responses. Those
// responses are then converted into a slice of onvif.Device.
func executeRawProbe(conn net.Conn, params netscan.Params) ([]onvif.Device, error) {
	probeSOAP := wsdiscovery.BuildProbeMessage(uuid.Must(uuid.NewV4()).String(), nil, nil,
		map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})

	addr := conn.RemoteAddr().String()

	if err := conn.SetDeadline(time.Now().Add(params.Timeout)); err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("%s: failed to set read/write deadline", addr), err)
	}

	if _, err := conn.Write([]byte(probeSOAP.String())); err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to write probe message", err)
	}

	var responses []string
	buf := make([]byte, bufSize)
	// keep reading from the PacketConn until the read deadline expires or an error occurs
	for {
		n, _, err := (conn.(net.PacketConn)).ReadFrom(buf)
		if err != nil {
			// ErrDeadlineExceeded is expected once the read timeout is expired
			if !stdErrors.Is(err, os.ErrDeadlineExceeded) {
				params.Logger.Debugf("Unexpected error occurred while reading ws-discovery responses: %s", err.Error())
			}
			break
		}
		responses = append(responses, string(buf[0:n]))
	}

	if len(responses) == 0 {
		// log as trace because when using UDP this will be logged for all devices that are probed
		// that do not respond or refuse the connection.
		params.Logger.Tracef("%s: No Response", addr)
		return nil, nil
	}
	for i, resp := range responses {
		params.Logger.Debugf("%s: Response %d of %d: %s", addr, i+1, len(responses), resp)
	}

	devices, err := wsdiscovery.DevicesFromProbeResponses(responses)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		params.Logger.Debugf("%s: no devices matched from probe response", addr)
		return nil, nil
	}

	return devices, nil
}

// makeDeviceMap creates a lookup table of existing devices by EndpointRefAddress
func (d *Driver) makeDeviceMap() map[string]contract.Device {
	devices := sdk.RunningService().Devices()
	deviceMap := make(map[string]contract.Device, len(devices))

	for _, dev := range devices {
		if dev.Name == d.serviceName {
			// skip control plane device
			continue
		}

		onvifInfo, ok := dev.Protocols[OnvifProtocol]
		if !ok {
			d.lc.Warnf("Found registered device %s without %s protocol information.", dev.Name, OnvifProtocol)
			continue
		}

		endpointRef := onvifInfo[EndpointRefAddress]
		if endpointRef == "" {
			d.lc.Warnf("Registered device %s is missing required %s protocol information: %s.",
				dev.Name, OnvifProtocol, EndpointRefAddress)
			continue
		}

		deviceMap[endpointRef] = dev
	}

	return deviceMap
}

// discoverFilter iterates through the discovered devices, and returns any that are not duplicates
// of devices in metadata or are from an alternate discovery method
// will return an empty slice if no new devices are discovered
func (d *Driver) discoverFilter(discoveredDevices []sdkModel.DiscoveredDevice) (filtered []sdkModel.DiscoveredDevice) {
	// filter out duplicate discovered devices by endpoint reference
	discoveredMap := make(map[string]sdkModel.DiscoveredDevice)
	for _, device := range discoveredDevices {
		endpointRef := device.Protocols[OnvifProtocol][EndpointRefAddress]
		if _, found := discoveredMap[endpointRef]; !found {
			discoveredMap[endpointRef] = device
		}
	}

	existingDevices := d.makeDeviceMap() // create comparison map

	// loop through discovered devices and see if they are already discovered
	for endpointRef, device := range discoveredMap {
		if existingDevice, found := existingDevices[endpointRef]; found {
			if err := d.updateExistingDevice(existingDevice, device); err != nil {
				d.lc.Errorf("error occurred while updating existing device %s: %s", existingDevice.Name, err.Error())
			}
			continue // skip registering existing device
		}
		// if device was not found, add it to the list of new devices to be registered with EdgeX
		filtered = append(filtered, device)
	}

	return filtered
}

// updateExistingDevice compares a discovered device and a matching existing device, and updates the existing
// device network address and port if necessary
func (d *Driver) updateExistingDevice(device contract.Device, discDev sdkModel.DiscoveredDevice) error {
	shouldUpdate := false
	if device.OperatingState == contract.Down {
		device.OperatingState = contract.Up
		shouldUpdate = true
	}

	device.Protocols[OnvifProtocol][LastSeen] = time.Now().Format(time.UnixDate)

	existAddr := device.Protocols[OnvifProtocol][Address]
	existPort := device.Protocols[OnvifProtocol][Port]
	discAddr := discDev.Protocols[OnvifProtocol][Address]
	discPort := discDev.Protocols[OnvifProtocol][Port]
	if existAddr != discAddr ||
		existPort != discPort {
		d.lc.Infof("Existing device %s has been discovered with a different network address. Old: %s, Discovered: %s",
			device.Name, existAddr+":"+existPort, discAddr+":"+discPort)
		device.Protocols[OnvifProtocol][Address] = discAddr
		device.Protocols[OnvifProtocol][Port] = discPort

		shouldUpdate = true
	}

	if !shouldUpdate {
		d.lc.Debug("Re-discovered existing device at the same network address, nothing to do")
		return nil
	}

	err := sdk.RunningService().UpdateDevice(device)
	if err != nil {
		d.lc.Errorf("There was an error updating the network address for device %s: %s", device.Name, err.Error())
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}
