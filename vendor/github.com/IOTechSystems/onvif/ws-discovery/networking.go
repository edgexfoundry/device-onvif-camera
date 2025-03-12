package wsdiscovery

/*******************************************************
 * Copyright (C) 2018 Palanjyan Zhorzhik
 *
 * This file is part of ws-discovery project.
 *
 * ws-discovery can be copied and/or distributed without the express
 * permission of Palanjyan Zhorzhik
 *******************************************************/

// Copyright (C) 2022 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/IOTechSystems/onvif"
	"github.com/beevik/etree"
	"github.com/google/uuid"
	"golang.org/x/net/ipv4"
)

const (
	bufSize = 8192
)

// GetAvailableDevicesAtSpecificEthernetInterface sends a ws-discovery Probe Message via
// UDP multicast to Discover NVT type Devices
func GetAvailableDevicesAtSpecificEthernetInterface(interfaceName string) ([]onvif.Device, error) {
	types := []string{"dn:NetworkVideoTransmitter"}
	namespaces := map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl", "ds": "http://www.onvif.org/ver10/device/wsdl"}

	probeResponses, err := SendProbe(interfaceName, nil, types, namespaces)
	if err != nil {
		return nil, fmt.Errorf("failed to probe: %w", err)
	}

	nvtDevices, err := DevicesFromProbeResponses(probeResponses)
	if err != nil {
		return nil, fmt.Errorf("failed to discover Onvif devices: %w", err)
	}

	return nvtDevices, nil
}

func DevicesFromProbeResponses(probeResponses []string) ([]onvif.Device, error) {
	nvtDevices := make([]onvif.Device, 0)
	xaddrSet := make(map[string]struct{})
	for _, j := range probeResponses {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(j); err != nil {
			return nil, err
		}

		probeMatches := doc.Root().FindElements("./Body/ProbeMatches/ProbeMatch")
		for _, probeMatch := range probeMatches {
			var xaddr string
			if address := probeMatch.FindElement("./XAddrs"); address != nil {
				u, err := url.Parse(address.Text())
				if err != nil {
					// TODO: Add logger for fmt.Printf("Invalid XAddrs: %s\n", address.Text())
					continue
				}
				xaddr = u.Host
			}
			if _, dupe := xaddrSet[xaddr]; dupe {
				// TODO: Add logger for fmt.Printf("Skipping duplicate XAddr: %s\n", xaddr)
				continue
			}

			var endpointRefAddress string
			if ref := probeMatch.FindElement("./EndpointReference/Address"); ref != nil {
				uuidElements := strings.Split(ref.Text(), ":")
				endpointRefAddress = uuidElements[len(uuidElements)-1]
			}

			dev, err := onvif.NewDevice(onvif.DeviceParams{
				Xaddr:              xaddr,
				EndpointRefAddress: endpointRefAddress,
				HttpClient: &http.Client{
					Timeout: 2 * time.Second,
				},
			})
			if err != nil {
				// TODO: Add logger for fmt.Printf("Failed to connect to camera at %s: %s\n", xaddr, err.Error())
				continue
			}

			var scopes []string
			ref := probeMatch.FindElement("./Scopes")
			if ref != nil {
				scopes = strings.Split(ref.Text(), " ")
			}
			dev.SetDeviceInfoFromScopes(scopes)

			xaddrSet[xaddr] = struct{}{}
			nvtDevices = append(nvtDevices, *dev)
			// TODO: Add logger for fmt.Printf("Onvif WS-Discovery: Find Xaddr: %-25s EndpointRefAddress: %s\n", xaddr, string(endpointRefAddress))
		}
	}

	return nvtDevices, nil
}

// SendProbe to device
func SendProbe(interfaceName string, scopes, types []string, namespaces map[string]string) ([]string, error) {
	probeSOAP := BuildProbeMessage(uuid.NewString(), scopes, types, namespaces)
	return SendUDPMulticast(probeSOAP.String(), interfaceName)
}

func SendUDPMulticast(msg string, interfaceName string) ([]string, error) {
	var responses []string
	data := []byte(msg)

	c, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)

	// 239.255.255.250 port 3702 is the multicast address and port used by ws-discovery
	group := net.IPv4(239, 255, 255, 250)
	dest := &net.UDPAddr{IP: group, Port: 3702}

	var iface *net.Interface
	if interfaceName == "" {
		iface = nil
	} else {
		iface, err = net.InterfaceByName(interfaceName)
		if err != nil {
			return nil, fmt.Errorf("failed to call InterfaceByName for interface %q: %w", interfaceName, err)
		}
	}

	if err = p.JoinGroup(iface, &net.UDPAddr{IP: group}); err != nil {
		return nil, fmt.Errorf("failed to JoinGroup for ws-discovery: %w", err)
	}
	if iface != nil {
		if err = p.SetMulticastInterface(iface); err != nil {
			return nil, fmt.Errorf("failed to SetMulticastInterface for interface %q: %w", interfaceName, err)
		}
		if err = p.SetMulticastTTL(2); err != nil {
			return nil, fmt.Errorf("failed to SetMulticastTTL: %w", err)
		}
	}
	if _, err = p.WriteTo(data, nil, dest); err != nil {
		return nil, fmt.Errorf("failed to write to ws-discovery multicast address %s: %w", dest.String(), err)
	}

	if err = p.SetReadDeadline(time.Now().Add(time.Second * 1)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	b := make([]byte, bufSize)

	// keep reading from the PacketConn until the read deadline expires or an error occurs
	for {
		n, _, _, err := p.ReadFrom(b)
		if err != nil {
			// ErrDeadlineExceeded is expected once the read timeout is expired
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				return nil, fmt.Errorf("unexpected error occurred while reading ws-discovery responses: %w", err)
			}
			break
		}
		responses = append(responses, string(b[0:n]))
	}
	return responses, nil
}
