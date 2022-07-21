// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	error "errors"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// setFriendlyName will create or update friendly name of a camera
func (onvifClient *OnvifClient) setFriendlyName(device contract.Device, data []byte) (contract.Device, errors.EdgeX) {
	var dataObj contract.ProtocolProperties
	err := json.Unmarshal(data, &dataObj)
	if err != nil {
		return device, errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}
	if len(dataObj) == 0 {
		return device, errors.NewCommonEdgeX(errors.KindContractInvalid, "no data in request body", err)
	}

	for key, value := range dataObj {
		value = strings.TrimSpace(value)
		key = strings.TrimSpace(key)
		if key != FriendlyName {
			err := error.New("invalid key")
			return device, errors.NewCommonEdgeX(errors.KindContractInvalid, "error setting FriendlyName", err)
		}
		device.Protocols[OnvifProtocol][FriendlyName] = value // create or update friendly name field
	}

	return device, nil
}

// getFriendlyName returns the friendly name of a camera
func (onvifClient *OnvifClient) getFriendlyName(device contract.Device) (string, errors.EdgeX) {
	name, ok := device.Protocols[OnvifProtocol][FriendlyName]
	if !ok {
		err := error.New("device is missing friendly name")
		return "", errors.NewCommonEdgeX(errors.KindServerError, "error getting FriendlyName", err)
	}
	return name, nil
}
