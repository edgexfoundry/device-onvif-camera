// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDiscoveryModeIsValid verifies discovery mode setting.
func TestDiscoveryModeIsValid(t *testing.T) {

	tests := []struct {
		input    DiscoveryMode
		expected bool
	}{
		{
			input:    ModeNetScan,
			expected: true,
		},
		{
			input:    ModeMulticast,
			expected: true,
		},
		{
			input:    ModeBoth,
			expected: true,
		},
		{
			input:    "invalidValue",
			expected: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.input), func(t *testing.T) {
			result := test.input.IsValid()
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestIsMulticastEnabled verifies multicast setting.
func TestIsMulticastEnabled(t *testing.T) {

	tests := []struct {
		input    DiscoveryMode
		expected bool
	}{
		{
			input:    ModeNetScan,
			expected: false,
		},
		{
			input:    ModeMulticast,
			expected: true,
		},
		{
			input:    ModeBoth,
			expected: true,
		},
		{
			input:    "invalidValue",
			expected: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.input), func(t *testing.T) {
			result := test.input.IsMulticastEnabled()
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestIsNetScanEnabled verifies netscan setting.
func TestIsNetScanEnabled(t *testing.T) {

	tests := []struct {
		input    DiscoveryMode
		expected bool
	}{
		{
			input:    ModeNetScan,
			expected: true,
		},
		{
			input:    ModeMulticast,
			expected: false,
		},
		{
			input:    ModeBoth,
			expected: true,
		},
		{
			input:    "invalidValue",
			expected: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.input), func(t *testing.T) {
			result := test.input.IsNetScanEnabled()
			assert.Equal(t, test.expected, result)
		})
	}
}
