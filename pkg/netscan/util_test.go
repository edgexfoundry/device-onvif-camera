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
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

type inetTest struct {
	inet  string
	first string
	last  string
	size  uint32
	err   bool
}

func mockIpWorker(ipCh <-chan uint32, result *inetTest) {
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

func ipGeneratorTest(input inetTest) (result inetTest) {
	var wg sync.WaitGroup
	ipCh := make(chan uint32, input.size)

	wg.Add(1)
	go func() {
		defer wg.Done()
		mockIpWorker(ipCh, &result)
	}()

	_, inet, err := net.ParseCIDR(input.inet)
	if err != nil {
		result.err = true
		return result
	}

	ipGenerator(context.Background(), inet, ipCh)
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
	}
	for _, input := range tests {
		input := input
		t.Run(input.inet, func(t *testing.T) {
			t.Parallel()
			result := ipGeneratorTest(input)
			if result.err && !input.err {
				t.Error("got unexpected error")
			} else if !result.err && input.err {
				t.Error("expected an error, but no error was returned")
			} else {
				if result.size != input.size {
					t.Errorf("expected %d ips, but got %d", input.size, result.size)
				}
				if result.first != input.first {
					t.Errorf("expected first ip in range to be %s, but got %s", input.first, result.first)
				}
				if result.last != input.last {
					t.Errorf("expected last ip in range to be %s, but got %s", input.last, result.last)
				}
			}
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
			if result.size != computeNetSz(i) {
				t.Errorf("expected %d ips, but got %d", computeNetSz(i), result.size)
			}
		})
	}
}
