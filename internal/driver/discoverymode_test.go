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
		mode     DiscoveryMode
		expected bool
	}{
		{
			mode:     ModeNetScan,
			expected: true,
		},
		{
			mode:     ModeMulticast,
			expected: true,
		},
		{
			mode:     ModeBoth,
			expected: true,
		},
		{
			mode:     "invalidValue",
			expected: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.mode), func(t *testing.T) {
			result := test.mode.IsValid()
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestIsNetScanAndIsMulticastEnabled verifies netscan and multicast settings.
func TestIsNetScanAndIsMulticastEnabled(t *testing.T) {

	tests := []struct {
		mode              DiscoveryMode
		multicastExpected bool
		netscanExpected   bool
	}{
		{
			mode:              ModeNetScan,
			netscanExpected:   true,
			multicastExpected: false,
		},
		{
			mode:              ModeMulticast,
			netscanExpected:   false,
			multicastExpected: true,
		},
		{
			mode:              ModeBoth,
			netscanExpected:   true,
			multicastExpected: true,
		},
		{
			mode:              "invalidValue",
			netscanExpected:   false,
			multicastExpected: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.mode), func(t *testing.T) {
			multicastActual := test.mode.IsMulticastEnabled()
			netscanActual := test.mode.IsNetScanEnabled()
			assert.Equal(t, test.multicastExpected, multicastActual)
			assert.Equal(t, test.netscanExpected, netscanActual)
		})
	}
}
