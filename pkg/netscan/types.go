// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package netscan

import (
	"context"
	"net"
	"time"
)

type ProtocolSpecificDiscovery interface {
	// ProbeFilter takes in a host and a slice of ports to be scanned. It should return a slice
	// of ports to actually scan, or a nil/empty slice if the host is to not be scanned at all.
	// Can be used to filter out known devices from being probed again if required.
	ProbeFilter(host string, ports []string) []string

	// OnConnectionDialed handles the protocol specific verification if there is actually
	// a valid device or devices at the other end of the connection.
	OnConnectionDialed(host string, port string, conn net.Conn, params Params) ([]ProbeResult, error)

	// ConvertProbeResult takes a raw ProbeResult and transforms it into a
	// processed DiscoveredDevice struct.
	ConvertProbeResult(probeResult ProbeResult, params Params) (DiscoveredDevice, error)
}

// ProbeResult holds the pre-processed information about a discovered device
type ProbeResult struct {
	// Host is the IP address that was probed
	Host string
	// Port is the port that was probed
	Port string
	// Data is the generic response details captured by the ProtocolSpecificDiscovery code
	// to be used to further process the result.
	Data interface{}
}

// workerParams is a helper struct to store shared parameters to ipWorkers
type workerParams struct {
	Params

	proto    ProtocolSpecificDiscovery
	ipCh     <-chan uint32
	resultCh chan<- []ProbeResult
	ctx      context.Context
}

// Params is the input configuration for a Discovery Net Scan
type Params struct {
	// Subnets is a slice of CIDR formatted subnets to scan
	Subnets []string
	// ScanPorts is a slice of ports to scan for on each host
	ScanPorts []string
	// AsyncLimit is the maximum amount of probes to run simultaneously
	AsyncLimit int
	// NetworkProtocol is the type of probe to make: tcp, udp, etc.
	NetworkProtocol string
	// MaxTimeoutsPerHost is the amount of ports that timeout before we assume the host is offline and
	// skip the rest of the ports. Set the value to 0 to disable this and always scan each port.
	MaxTimeoutsPerHost int
	// Timeout is the maximum amount of time to wait when connecting to a host before giving up.
	Timeout time.Duration
	// Logger is a generic logging client for this code to log messages to
	Logger Logger
}

// DiscoveredDevice defines the required information for a found device.
type DiscoveredDevice struct {
	Name string
	Info interface{}
}

// Logger is a generic logging interface in order to not lock this code
// into a specific logging framework. It is directly compatible with, but not limited to:
// - github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger LoggingClient
// - github.com/sirupsen/logrus Logger
type Logger interface {
	// Debugf logs a formatted message at the DEBUG severity level
	Debugf(format string, args ...interface{})
	// Errorf logs a formatted message at the ERROR severity level
	Errorf(format string, args ...interface{})
	// Infof logs a formatted message at the INFO severity level
	Infof(format string, args ...interface{})
	// Tracef logs a formatted message at the TRACE severity level
	Tracef(format string, args ...interface{})
	// Warnf logs a formatted message at the WARN severity level
	Warnf(format string, args ...interface{})
}
