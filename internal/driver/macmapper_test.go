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

func TestSanitizeMACAddress(t *testing.T) {
	tests := []struct {
		mac       string // input mac address
		sanitized string // expected sanitized output mac address
		err       bool   // expect an error
	}{
		{
			mac:       "aa:bb:cc:dd:ee:ff",
			sanitized: "aa:bb:cc:dd:ee:ff",
		},
		{
			mac: "aa:bb:cc:dd:ee",
			err: true,
		},
		{
			mac: "aabbccddee",
			err: true,
		},
		{
			mac: "aa:bb:cc-dd-ee",
			err: true,
		},
		{
			mac:       "AA:BB:CC:DD:EE:FF",
			sanitized: "aa:bb:cc:dd:ee:ff",
		},
		{
			mac:       "11-22-33-44-55-66",
			sanitized: "11:22:33:44:55:66",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.mac, func(t *testing.T) {
			sanitized, err := SanitizeMACAddress(test.mac)
			assert.Equal(t, test.sanitized, sanitized)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
