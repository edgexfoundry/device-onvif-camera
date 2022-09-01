// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package netscan

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

type inetTest struct {
	name  string
	inet  *net.IPNet
	first string
	last  string
	size  uint32
}

type inetTestResult struct {
	first string
	last  string
	size  uint32
}

func mockIpWorker(ipCh <-chan uint32, result *inetTestResult) {
	ip := net.IP([]byte{0, 0, 0, 0})
	var last uint32

	for a := range ipCh {
		result.size++
		last = a

		if result.first == "" {
			binary.BigEndian.PutUint32(ip, a)
			result.first = ip.String()
		}
	}

	binary.BigEndian.PutUint32(ip, last)
	result.last = ip.String()
}

func ipGeneratorTest(input inetTest) (result inetTestResult) {
	var wg sync.WaitGroup
	ipCh := make(chan uint32, input.size)

	wg.Add(1)
	go func() {
		defer wg.Done()
		mockIpWorker(ipCh, &result)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ipGenerator(ctx, input.inet, ipCh)
	close(ipCh)
	wg.Wait()

	return result
}

func mustParseCIDR(t *testing.T, cidr string) *net.IPNet {
	_, inet, err := net.ParseCIDR(cidr)
	require.NoError(t, err)
	return inet
}

// TestIpGenerator calls the ip generator and validates that the first ip, last ip, and size
// match the expected values.
func TestIpGenerator(t *testing.T) {
	tests := []inetTest{
		{
			name:  "basic /32 subnet",
			inet:  mustParseCIDR(t, "192.168.1.110/32"),
			first: "192.168.1.110",
			last:  "192.168.1.110",
			size:  computeNetSz(32),
		},
		{
			name:  "basic /31 subnet",
			inet:  mustParseCIDR(t, "192.168.1.20/31"),
			first: "192.168.1.20",
			last:  "192.168.1.20",
			size:  computeNetSz(31),
		},
		{
			name:  "basic /30 subnet",
			inet:  mustParseCIDR(t, "192.168.1.1/30"),
			first: "192.168.1.1",
			last:  "192.168.1.2",
			size:  computeNetSz(30),
		},
		{
			name:  "basic /24 subnet",
			inet:  mustParseCIDR(t, "192.168.1.110/24"),
			first: "192.168.1.1",
			last:  "192.168.1.254",
			size:  computeNetSz(24),
		},
		{
			name:  "basic /16 subnet",
			inet:  mustParseCIDR(t, "192.168.99.20/16"),
			first: "192.168.0.1",
			last:  "192.168.255.254",
			size:  computeNetSz(16),
		},
		{
			name:  "basic /8 subnet",
			inet:  mustParseCIDR(t, "10.10.10.10/8"),
			first: "10.0.0.1",
			last:  "10.255.255.254",
			size:  computeNetSz(8),
		},
		{
			name: "nil inet",
			inet: nil,
			size: 0,
		},
		{
			name: "nil IP and Mask",
			inet: &net.IPNet{
				IP:   nil,
				Mask: nil,
			},
			size: 0,
		},
		{
			name: "nil Mask",
			inet: &net.IPNet{
				IP:   net.IP{192, 168, 1, 100},
				Mask: nil,
			},
			size: 0,
		},
		{
			name: "invalid subnet size (too small)",
			inet: &net.IPNet{
				IP:   net.IP{192, 168, 1, 100},
				Mask: net.IPMask{1, 2, 3},
			},
			size: 0,
		},
		{
			name: "invalid subnet size (too large)",
			inet: &net.IPNet{
				IP:   net.IP{192, 168, 1, 100},
				Mask: net.IPMask{255, 255, 255, 255, 255, 255},
			},
			size: 0,
		},
		{
			name: "skip subnet-zero",
			inet: mustParseCIDR(t, "192.168.1.100/0"),
			size: 0,
		},
		//{
		//	name: "invalid subnet /99",
		//	inet: "1.1.1.1/99",
		//},
		//{
		//	name: "invalid ip address",
		//	inet: "1.1/24",
		//},
		//{
		//	name: "missing ip address",
		//	inet: "/24",
		//},
		{
			name: "skip ipv6 subnet",
			inet: mustParseCIDR(t, "2001:4860:4860::8888/32"),
			size: 0, // expect size of 0 because ipv6 is skipped
		},
	}
	for _, input := range tests {
		input := input
		t.Run(input.name, func(t *testing.T) {
			t.Parallel()
			result := ipGeneratorTest(input)

			assert.Equal(t, input.size, result.size)
			if input.size > 0 {
				assert.Equal(t, input.first, result.first)
				assert.Equal(t, input.last, result.last)
			}
		})
	}
}

func TestIPGeneratorTimeoutCancel(t *testing.T) {
	var result inetTestResult
	var wg sync.WaitGroup
	ipCh := make(chan uint32, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		mockIpWorker(ipCh, &result)
	}()

	// create a very short timeout that we know will trip before completed
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	ipGenerator(ctx, mustParseCIDR(t, "10.0.0.0/8"), ipCh) // start generating a large /8 subnet
	close(ipCh)
	wg.Wait()

	// make sure we generated less than all of them because we were cancelled
	assert.Less(t, result.size, computeNetSz(8))
}

// TestIpGeneratorSubnetSizes calls the ip generator for various subnet sizes and validates that
// the correct amount of IP addresses are generated
func TestIpGeneratorSubnetSizes(t *testing.T) {
	// stop at 16 because the time taken gets exponentially longer
	for i := 32; i >= 16; i-- {
		i := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			result := ipGeneratorTest(inetTest{
				size: uint32(i),
				inet: mustParseCIDR(t, fmt.Sprintf("192.168.1.1/%d", i)),
			})
			assert.Equal(t, computeNetSz(i), result.size)
		})
	}
}
