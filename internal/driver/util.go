// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/base64"
	"fmt"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"net/url"
	"regexp"
	"strings"
)

const (
	// Per https://tools.ietf.org/html/rfc3986#section-2.3, unreserved characters = ALPHA / DIGIT / "-" / "." / "_" / "~"
	// Also due to names used in topics for Redis Pub/Sub, "." are not allowed
	// see: https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/common/validator.go
	//
	// Note: this is an inverted match of the unreserved characters from above, as we want to remove the reserved ones
	rFC3986ReservedCharsRegexString = "[^a-zA-Z0-9-_~]+"
)

var (
	rFC3986ReservedCharsRegex = regexp.MustCompile(rFC3986ReservedCharsRegexString)
)

type MultiErr []error

//goland:noinspection GoReceiverNames
func (me MultiErr) Error() string {
	strs := make([]string, len(me))
	for i, s := range me {
		strs[i] = s.Error()
	}

	return strings.Join(strs, "; ")
}

func addressAndPort(xaddr string) (string, string) {
	substrings := strings.Split(xaddr, ":")
	if len(substrings) == 1 {
		// The port might be empty from the discovered result, for example <d:XAddrs>http://192.168.12.123/onvif/device_service</d:XAddrs>
		return substrings[0], "80"
	} else {
		return substrings[0], substrings[1]
	}
}

func attributeByKey(attributes map[string]interface{}, key string) (attr string, err errors.EdgeX) {
	val, ok := attributes[key]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("attribute %s not exists", key), nil)
	}
	attr = fmt.Sprint(val)
	return attr, nil
}

func parametersFromURLRawQuery(req sdkModel.CommandRequest) ([]byte, errors.EdgeX) {
	values, err := url.ParseQuery(fmt.Sprint(req.Attributes[URLRawQuery]))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to parse get command parameter for resource '%s'", req.DeviceResourceName), err)
	}
	param, exists := values[jsonObject]
	if !exists || len(param) == 0 {
		return []byte{}, nil
	}
	data, err := base64.StdEncoding.DecodeString(param[0])
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to decode '%v' parameter for resource '%s', the value should be json object with base64 encoded", jsonObject, req.DeviceResourceName), err)
	}
	return data, nil
}

func buildDeviceName(manufacturer, model, endpointRefAddr string) string {
	return fmt.Sprintf("%s-%s-%s",
		// replace all the reserved chars with a dash, and trim any leftovers
		strings.Trim(rFC3986ReservedCharsRegex.ReplaceAllString(manufacturer, "-"), "-"),
		strings.Trim(rFC3986ReservedCharsRegex.ReplaceAllString(model, "-"), "-"),
		strings.Trim(rFC3986ReservedCharsRegex.ReplaceAllString(endpointRefAddr, "-"), "-"))
}
