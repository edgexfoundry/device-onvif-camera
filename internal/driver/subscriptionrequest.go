// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// MinimumInitialTerminationTime indicates the minimum InitialTerminationTime because device service sends Renew request every ten second before termination time
const MinimumInitialTerminationTime = 11

type SubscriptionRequest struct {
	// AutoRenew indicate the device service should renew the subscription
	AutoRenew *bool

	// TopicFilter  indicates the optional XPATH expression to filter the event by topic
	TopicFilter *string
	// TopicFilter  indicates the optional XPATH expression to filter the event by message content
	MessageContentFilter *string
	// SubscriptionPolicy is the camera's subscription policy, the user should check the capability before using it
	SubscriptionPolicy *string
	// InitialTerminationTime indicates the subscription lifetime with specified duration
	InitialTerminationTime *string

	// MessageTimeout indicates the timeout to pull event message
	MessageTimeout *string
	// MessageTimeout indicates the limit for the number of messages to return at once
	MessageLimit *int
}

func subscriptionRequest(attributes map[string]interface{}, requestData []byte) (*SubscriptionRequest, errors.EdgeX) {
	request := &SubscriptionRequest{}
	err := json.Unmarshal(requestData, request)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "fail to unmarshal the json request body", err)
	}

	topicFilter, ok := attributes[DefaultTopicFilter]
	if request.TopicFilter == nil && ok {
		val := fmt.Sprint(topicFilter)
		request.TopicFilter = &val
	}

	messageContentFilter, ok := attributes[DefaultMessageContentFilter]
	if request.MessageContentFilter == nil && ok {
		val := fmt.Sprint(messageContentFilter)
		request.MessageContentFilter = &val
	}

	initialTerminationTime, ok := attributes[DefaultInitialTerminationTime]
	if request.InitialTerminationTime == nil && ok {
		val := fmt.Sprint(initialTerminationTime)
		request.InitialTerminationTime = &val
	}
	duration, err := ParseISO8601(*request.InitialTerminationTime)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid initial terminationTime, %v", err), err)
	}
	if duration.Seconds() < MinimumInitialTerminationTime {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("the initial terminationTime should greater then ten second"), nil)
	}

	subscriptionPolicy, ok := attributes[DefaultSubscriptionPolicy]
	if request.SubscriptionPolicy == nil && ok {
		val := fmt.Sprint(subscriptionPolicy)
		request.SubscriptionPolicy = &val
	}

	autoRenew, ok := attributes[DefaultAutoRenew]
	if request.AutoRenew == nil && ok {
		val, err := strconv.ParseBool(fmt.Sprint(autoRenew))
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to parse the request attribute '%s'", DefaultAutoRenew), err)
		}
		request.AutoRenew = &val
	}

	messageTimeout, ok := attributes[DefaultMessageTimeout]
	if request.MessageTimeout == nil && ok {
		val := fmt.Sprint(messageTimeout)
		request.MessageTimeout = &val
	}

	messageLimit, ok := attributes[DefaultMessageLimit]
	if request.MessageLimit == nil && ok {
		val, err := strconv.Atoi(fmt.Sprint(messageLimit))
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("fail to parse the request attribute '%s'", DefaultMessageLimit), err)
		}
		request.MessageLimit = &val
	}
	return request, nil
}

var pattern = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)

// ParseISO8601 parses an ISO8601 duration string.
// https://github.com/senseyeio/duration/blob/master/duration.go
// https://en.wikipedia.org/wiki/ISO_8601#Durations
func ParseISO8601(from string) (time.Duration, error) {
	var match []string
	var d Duration

	if pattern.MatchString(from) {
		match = pattern.FindStringSubmatch(from)
	} else {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid time duration, the format shoulb be like 'PT180S' ", nil)
	}

	for i, name := range pattern.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return 0, err
		}
		switch name {
		case "year":
			d.Y = val
		case "month":
			d.M = val
		case "week":
			d.W = val
		case "day":
			d.D = val
		case "hour":
			d.TH = val
		case "minute":
			d.TM = val
		case "second":
			d.TS = val
		default:
			return 0, fmt.Errorf("unknown field %s", name)
		}
	}

	return d.timeDuration(), nil
}

// Duration represents an ISO8601 Duration
type Duration struct {
	Y int
	M int
	W int
	D int
	// Time Component
	TH int
	TM int
	TS int
}

func (d Duration) timeDuration() time.Duration {
	var dur time.Duration
	dur = dur + (time.Duration(d.TH) * time.Hour)
	dur = dur + (time.Duration(d.TM) * time.Minute)
	dur = dur + (time.Duration(d.TS) * time.Second)
	return dur
}
