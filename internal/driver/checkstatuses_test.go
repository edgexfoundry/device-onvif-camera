// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpdateDeviceStatus_update(t *testing.T) {
	mockService, driver := createDriverWithMockService()
	mockService.On("GetDeviceByName", testDeviceName).
		Return(createTestDevice(), nil).Once()
	mockService.On("UpdateDevice", mock.AnythingOfType("models.Device")).
		Return(nil).Once()

	err := driver.updateDeviceStatus(testDeviceName, UpWithAuth)
	mockService.AssertExpectations(t)
	require.NoError(t, err)
}

func TestUpdateDeviceStatus_noUpdate(t *testing.T) {
	mockService, driver := createDriverWithMockService()
	mockService.On("GetDeviceByName", testDeviceName).
		Return(createTestDevice(), nil).Once()

	err := driver.updateDeviceStatus(testDeviceName, Unreachable)
	mockService.AssertExpectations(t)
	require.NoError(t, err)
}
