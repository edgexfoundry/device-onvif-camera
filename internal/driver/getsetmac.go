// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	error "errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"strings"
)

// setMACAddress will create or update mac address of a camera
func (onvifClient *OnvifClient) setMACAddress(device contract.Device, data []byte) (contract.Device, errors.EdgeX) {
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
		if key != MACAddress {
			err := error.New("invalid key")
			return device, errors.NewCommonEdgeX(errors.KindContractInvalid, "error setting MACAddress", err)
		}
		_, err := SanitizeMACAddress(value)
		if err != nil {
			return device, errors.NewCommonEdgeX(errors.KindContractInvalid, "error setting MACAddress:", err)
		}

		device.Protocols[OnvifProtocol][MACAddress] = value // create or update mac address field
	}

	return device, nil
}

// getMACAddress returns the MAC address of a camera
func (onvifClient *OnvifClient) getMACAddress(device contract.Device) (string, errors.EdgeX) {
	mac, ok := device.Protocols[OnvifProtocol][MACAddress]
	if !ok {
		err := error.New("device is missing mac address")
		return "", errors.NewCommonEdgeX(errors.KindServerError, "error getting MACAddress", err)
	}
	return mac, nil
}
