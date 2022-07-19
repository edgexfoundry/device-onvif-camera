// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTryGetCredentials_noAuth verifies when a credential is passed with a noAuth path, noAuth credentials are returned.
func TestTryGetCredentials_noAuth(t *testing.T) {
	driver, _ := createDriverWithMockService()
	result, err := driver.tryGetCredentials(noAuthSecretPath)

	expected := Credentials{
		AuthMode: AuthModeNone,
	}
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
