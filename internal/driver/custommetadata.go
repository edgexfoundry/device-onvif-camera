// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

func setCustomMetadata(device contract.Device, data []byte) (cv *sdkModel.CommandValue, edgexErr errors.EdgeX) {
	var obj contract.ProtocolProperties
	err := json.Unmarshal(data, obj)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}
	return cv, nil
}

func getDevice(deviceName string) (device contract.Device, err error) {
	device, err = sdk.RunningService().GetDeviceByName(deviceName) // CameraInfo does not contain protocol proerties, this is the alternative
	return device, err
}
