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
	"sync/atomic"
	"testing"
	"time"
)

type inetTest struct {
	inet      string
	first     string
	last      string
	size      uint32
	expectErr bool
}

type inetTestResult struct {
	first string
	last  string
	size  uint32
	err   error
}

func mockIpWorker(ipCh <-chan uint32, result *inetTestResult) {
	ip := net.IP([]byte{0, 0, 0, 0})
	var last uint32

	for a := range ipCh {
		atomic.AddUint32(&result.size, 1)
		atomic.StoreUint32(&last, a)

		if result.first == "" {
			binary.BigEndian.PutUint32(ip, a)
			result.first = ip.String()
		}
		binary.BigEndian.PutUint32(ip, last)
		result.last = ip.String()
	}

}

func ipGeneratorTest(input inetTest) (result inetTestResult) {
	var wg sync.WaitGroup
	ipCh := make(chan uint32, input.size)

	wg.Add(1)
	go func() {
		defer wg.Done()
		mockIpWorker(ipCh, &result)
	}()

	_, inet, err := net.ParseCIDR(input.inet)
	if err != nil {
		result.err = err
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ipGenerator(ctx, inet, ipCh)
	close(ipCh)
	wg.Wait()

	return result
}

// TestIpGenerator calls the ip generator and validates that the first ip, last ip, and size
// match the expected values.
func TestIpGenerator(t *testing.T) {
	tests := []inetTest{
		{
			inet:  "192.168.1.110/24",
			first: "192.168.1.1",
			last:  "192.168.1.254",
			size:  computeNetSz(24),
		},
		{
			inet:  "192.168.1.110/32",
			first: "192.168.1.110",
			last:  "192.168.1.110",
			size:  computeNetSz(32),
		},
		{
			inet:  "192.168.1.20/31",
			first: "192.168.1.20",
			last:  "192.168.1.20",
			size:  computeNetSz(31),
		},
		{
			inet:  "192.168.99.20/16",
			first: "192.168.0.1",
			last:  "192.168.255.254",
			size:  computeNetSz(16),
		},
		{
			inet:  "10.10.10.10/8",
			first: "10.0.0.1",
			last:  "10.255.255.254",
			size:  computeNetSz(8),
		},
		{
			inet:      "invalid inet",
			expectErr: true,
		},
		{
			inet:      "1.1.1.1/33",
			expectErr: true,
		},
		{
			inet:      "1.1.1.1",
			expectErr: true,
		},
		{
			inet:      "1.1.1.1/99",
			expectErr: true,
		},
		{
			inet:      "1.1/24",
			expectErr: true,
		},
		{
			inet:      "/24",
			expectErr: true,
		},
	}
	for _, input := range tests {
		input := input
		t.Run(input.inet, func(t *testing.T) {
			t.Parallel()
			result := ipGeneratorTest(input)
			if input.expectErr {
				require.Error(t, result.err)
			} else {
				require.NoError(t, result.err)
			}

			assert.Equal(t, input.size, result.size)
			assert.Equal(t, input.first, result.first)
			assert.Equal(t, input.last, result.last)
		})
	}
}

// TestIpGeneratorSubnetSizes calls the ip generator for various subnet sizes and validates that
// the correct amount of IP addresses are generated
func TestIpGeneratorSubnetSizes(t *testing.T) {
	// stop at 10 because the time taken gets exponentially longer
	for i := 32; i >= 10; i-- {
		i := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			result := ipGeneratorTest(inetTest{size: uint32(i), inet: fmt.Sprintf("192.168.1.1/%d", i)})
			assert.Equal(t, computeNetSz(i), result.size)
		})
	}
}
