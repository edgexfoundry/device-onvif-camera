// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseISO8601 verifies accurate parsing of time.
func TestParseISO8601(t *testing.T) {
	tests := []struct {
		input         string
		expected      time.Duration
		errorExpected bool
	}{
		{
			input:    "PT180S",
			expected: 180000000000,
		},
		{
			input:    "P1Y2M3W4DT5H6M7S",
			expected: 18367000000000,
		},
		{
			input:    "P1YT5H",
			expected: 18000000000000,
		},
		{
			input:    "P5DT2M",
			expected: 120000000000,
		},
		{
			input:         "3Y6M4DT12H30M5S",
			errorExpected: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseISO8601(test.input)

			if test.errorExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}
