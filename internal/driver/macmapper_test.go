// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsValidMacAddress(t *testing.T) {
	tests := []struct {
		mac   string
		valid bool
	}{
		{
			mac:   "",
			valid: false,
		},
		{
			mac:   "ab:cd:ef:gh:ij:kl",
			valid: false,
		},
		{
			mac:   "11:22:33:44:55:66",
			valid: true,
		},
		{
			mac:   "11-22-33-44-55-66",
			valid: true,
		},
		{
			mac:   "112233445566",
			valid: true,
		},
		{
			mac:   "1122334455667",
			valid: false,
		},
		{
			mac:   "aa:bb-cc-dd-ee:ff",
			valid: false,
		},
		{
			mac:   "aa:bb:cc",
			valid: false,
		},
		{
			mac:   "1:2:3:4:5:6",
			valid: false,
		},
		{
			mac:   "1234:5678:9abc",
			valid: false,
		},
		{
			mac:   "12:34:56:78:9a:bc",
			valid: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.mac, func(t *testing.T) {
			assert.Equal(t, test.valid, IsValidMacAddress(test.mac))
		})
	}
}

func TestSanitizeMacAddress(t *testing.T) {
	tests := []struct {
		mac       string
		sanitized string
		err       bool
	}{
		{
			mac:       "aa:bb:cc:dd:ee:ff",
			sanitized: "AA:BB:CC:DD:EE:FF",
		},
		{
			mac: "aa:bb:cc:dd:ee",
			err: true,
		},
		{
			mac:       "AA:BB:CC:DD:EE:FF",
			sanitized: "AA:BB:CC:DD:EE:FF",
		},
		{
			mac:       "11-22-33-44-55-66",
			sanitized: "11:22:33:44:55:66",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.mac, func(t *testing.T) {
			sanitized, err := SanitizeMacAddress(test.mac)
			assert.Equal(t, test.sanitized, sanitized)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
