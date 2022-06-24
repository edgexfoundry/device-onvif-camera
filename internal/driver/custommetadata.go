// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// setCustomMetadata will return a map containing the fields provided in the call to the function
func (onvifClient *OnvifClient) setCustomMetadata(device contract.Device, data []byte) (contract.Device, errors.EdgeX) {
	if _, found := device.Protocols[CustomMetadata]; !found {
		saveOnvifProtocol := device.Protocols[OnvifProtocol]
		device.Protocols = map[string]contract.ProtocolProperties{
			OnvifProtocol:  saveOnvifProtocol,
			CustomMetadata: {},
		}
	}

	var dataObj contract.ProtocolProperties
	err := json.Unmarshal(data, &dataObj)
	if err != nil {
		return device, errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}
	if len(dataObj) == 0 {
		return device, errors.NewCommonEdgeX(errors.KindContractInvalid, "no data in PUT command", err)
	}

	for okey, ovalue := range device.Protocols[OnvifProtocol] {
		fmt.Printf("%s, %s", okey, ovalue)
	}
	for key, value := range dataObj {
		value = strings.TrimSpace(value)
		key = strings.TrimSpace(key)
		if len(key) == 0 {
			onvifClient.driver.lc.Warn("tried to add an empty key: {\"\": \"%s\"}", value)
			continue
		}

		device.Protocols[CustomMetadata][key] = value // create or update a field in CustomMetadata
	}

	return device, nil
}

// getCustomMetadata will return all metdata or enter getSpecificCustomMetadata if a list is provided
func (onvifClient *OnvifClient) getCustomMetadata(device contract.Device, data []byte) (contract.ProtocolProperties, errors.EdgeX) {
	if _, found := device.Protocols[CustomMetadata]; !found {
		saveOnvifProtocol := device.Protocols[OnvifProtocol]
		device.Protocols = map[string]contract.ProtocolProperties{
			OnvifProtocol:  saveOnvifProtocol,
			CustomMetadata: {},
		}
	}

	if len(data) == 0 { // if no list is provided, return all
		return device.Protocols[CustomMetadata], nil
	}

	// if an array of fields is provided, return those specific fields
	metadataObj, err := onvifClient.getSpecificCustomMetadata(device, data)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to get specific metadata for device %s", onvifClient.DeviceName), err)
	}
	return metadataObj, nil
}

// getSpecificCustomMetadata will return a map of the key/value pairs corresponding to the array of keys provided in the resource call
func (onvifClient *OnvifClient) getSpecificCustomMetadata(device contract.Device, data []byte) (obj contract.ProtocolProperties, error errors.EdgeX) {
	input := make(map[string][]string)
	response := make(map[string]string)

	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}
	if len(input[CustomMetadata]) == 0 {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "no data in query body", err)
	}

	for _, key := range input[CustomMetadata] {
		value, found := device.Protocols[CustomMetadata][key]
		if !found {
			onvifClient.driver.lc.Warnf("Failed to find custom metadata field %s", key) // TODO: should this also be displayed in command response?
			continue
		}
		response[key] = value
	}

	return response, nil
}

// deleteCustomMetadata will delete specified entries in custom metadata
func (onvifClient *OnvifClient) deleteCustomMetadata(device contract.Device, data []byte) (contract.Device, errors.EdgeX) {
	var input []string
	err := json.Unmarshal(data, &input)
	if err != nil {
		return device, errors.NewCommonEdgeX(errors.KindServerError, "failed to unmarshal the json request body", err)
	}
	if len(input) == 0 {
		return device, errors.NewCommonEdgeX(errors.KindContractInvalid, "no data in query body", err)
	}

	for _, deleteKey := range input {
		deleteKey = strings.TrimSpace(deleteKey)
		_, found := device.Protocols[CustomMetadata][deleteKey]
		if !found {
			onvifClient.driver.lc.Warnf("can't delete field '%s', it does not exist in Custom Metadata", deleteKey)
			continue
		}
		delete(device.Protocols[CustomMetadata], deleteKey)
	}

	return device, nil
}
