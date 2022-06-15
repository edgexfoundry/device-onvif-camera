// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

func (onvifClient *OnvifClient) setCustomMetadata(device contract.Device, data []byte) errors.EdgeX {
	var dataObj contract.ProtocolProperties
	err := json.Unmarshal(data, &dataObj)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}

	for key, value := range dataObj {
		device.Protocols[CustomMetadata][key] = value
	}

	return nil
}

// getSpecificCustomMetadata will return a map of the key/value pairs corresponding to the array of keys provided in the resource call
func (onvifClient *OnvifClient) getSpecificCustomMetadata(device contract.Device, data []byte) (obj contract.ProtocolProperties, error errors.EdgeX) {
	var dataArray map[string][]string
	dataMap := make(map[string]string)
	err := json.Unmarshal(data, &dataArray)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}

	for _, key := range dataArray[CustomMetadata] {
		value := device.Protocols[CustomMetadata][key]
		if value == "" {
			onvifClient.driver.lc.Warnf("Failed to find custom metadata field %s", key) // TODO: should this also be displayed in command response?
			dataMap[key] = "This field does not exist in custom metadata"
			continue
		}
		dataMap[key] = value
	}
	return dataMap, nil
}

func cleanUpMetadata(device contract.Device) models.ProtocolProperties {
	for key, value := range device.Protocols[CustomMetadata] {
		if value == "" {
			delete(device.Protocols[CustomMetadata], key)
		}
	}
	return device.Protocols[CustomMetadata]
}
