// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package discover

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
	Host string
	Port string
	Info interface{}
}

// workerParams is a helper struct to store shared parameters to ipWorkers
type workerParams struct {
	Params

	proto    ProtocolSpecificDiscovery
	ipCh     <-chan uint32
	resultCh chan<- []ProbeResult
	ctx      context.Context
}

type Params struct {
	Subnets            []string
	ScanPorts          []string
	AsyncLimit         int
	NetworkProtocol    string
	MaxTimeoutsPerHost int
	Timeout            time.Duration
	Logger             LoggingClient
}

// DiscoveredDevice defines the required information for a found device.
type DiscoveredDevice struct {
	Name string
	Info interface{}
}

// LoggingClient is a generic logging interface in order to not lock this code
// into a specific logging framework
type LoggingClient interface {
	// Debugf logs a formatted message at the DEBUG severity level
	Debugf(msg string, args ...interface{})
	// Errorf logs a formatted message at the ERROR severity level
	Errorf(msg string, args ...interface{})
	// Infof logs a formatted message at the INFO severity level
	Infof(msg string, args ...interface{})
	// Tracef logs a formatted message at the TRACE severity level
	Tracef(msg string, args ...interface{})
	// Warnf logs a formatted message at the WARN severity level
	Warnf(msg string, args ...interface{})
}
