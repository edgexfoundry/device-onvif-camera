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

func TestMacAddressMapper_IsValidMacAddress(t *testing.T) {
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
