// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// setCustomMetadata will return a map containing the fields provided in the call to the function
func (onvifClient *OnvifClient) setCustomMetadata(data []byte) errors.EdgeX {
	// onvifClient.CameraInfo does not contain protocol properties, this is the alternative
	deviceName := onvifClient.DeviceName
	device, err := sdk.RunningService().GetDeviceByName(deviceName)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
	}

	var dataObj contract.ProtocolProperties

	err = json.Unmarshal(data, &dataObj)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}

	for key, value := range dataObj {
		device.Protocols[CustomMetadata][key] = value // create or update a field in CustomMetadata
	}

	cleanUpMetadata(device)

	return nil
}

func (onvifClient *OnvifClient) getCustomMetadata(data []byte) (contract.ProtocolProperties, errors.EdgeX) {
	// onvifClient.CameraInfo does not contain protocol properties, this is the alternative
	deviceName := onvifClient.DeviceName
	device, err := sdk.RunningService().GetDeviceByName(deviceName)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get device '%s'", deviceName), err)
	}

	var metadataObj contract.ProtocolProperties

	if len(data) == 0 { // if no list is provided, return all
		metadataObj = device.Protocols[CustomMetadata]
	} else { // if a list of fields is provided, return those specific fields
		metadataObj, err = onvifClient.getSpecificCustomMetadata(device, data)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get specific metadata for device %s", deviceName), err)
		}
	}

	return metadataObj, nil
}

// getSpecificCustomMetadata will return a map of the key/value pairs corresponding to the array of keys provided in the resource call
func (onvifClient *OnvifClient) getSpecificCustomMetadata(device contract.Device, data []byte) (obj contract.ProtocolProperties, error errors.EdgeX) {
	dataArray := make(map[string][]string)
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

// cleanUpMetadata will look for empty fields in CustomMetadata and delete themm
func cleanUpMetadata(device contract.Device) {
	for key, value := range device.Protocols[CustomMetadata] {
		if value == "" {
			delete(device.Protocols[CustomMetadata], key)
		}
	}
}
