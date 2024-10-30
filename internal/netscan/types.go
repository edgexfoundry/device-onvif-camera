// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

//go:generate mockery --name=ProtocolSpecificDiscovery --inpackage

package netscan

import (
	"context"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
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
	ConvertProbeResult(probeResult ProbeResult, params Params) (sdkModel.DiscoveredDevice, error)
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
	// ScanPorts is a slice of ports to scan for on each host. The first port is done synchronously
	// to test if the host is reachable, and any ports after that are done async.
	ScanPorts []string
	// AsyncLimit is the maximum amount of hosts to probe simultaneously. This does not include
	// any async scanning for multiple ports on the same host.
	AsyncLimit int
	// NetworkProtocol is the type of probe to make: tcp, udp, etc.
	NetworkProtocol string
	// Timeout is the maximum amount of time to wait when connecting to a host before giving up.
	Timeout time.Duration
	// Logger is a generic logging client for this code to log messages to.
	Logger logger.LoggingClient
}
