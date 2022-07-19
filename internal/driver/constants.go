// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

// Enumerations for DeviceStatus
const (
	UpWithAuth    = "UpWithAuth"
	UpWithoutAuth = "UpWithoutAuth"
	Reachable     = "Reachable"
	Unreachable   = "Unreachable"
)

const (
	CustomMetadata       = "CustomMetadata"
	GetCustomMetadata    = "GetCustomMetadata"
	SetCustomMetadata    = "SetCustomMetadata"
	DeleteCustomMetadata = "DeleteCustomMetadata"
)

const (
	MACAddress      = "MACAddress"
	FriendlyName    = "FriendlyName"
	GetFriendlyName = "GetFriendlyName"
	SetFriendlyName = "SetFriendlyName"
	GetMACAddress   = "GetMACAddress"
	SetMACAddress   = "SetMACAddress"
)

const (
	OnvifProtocol      = "Onvif"
	Address            = "Address"
	Port               = "Port"
	EndpointRefAddress = "EndpointRefAddress"
	LastSeen           = "LastSeen"
	DeviceStatus       = "DeviceStatus"

	// Maximum interval for checkStatus interval
	maxStatusInterval = 300

	// Service is resource attribute and indicates the web service for the Onvif
	Service = "service"
	// GetFunction is resource attribute and indicates the SOAP action for the specified web service, it is used for the read operation
	GetFunction = "getFunction"
	// SetFunction is resource attribute and indicates the SOAP action for the specified web service, it is used for the write operation
	SetFunction = "setFunction"

	// SubscribeType indicates the way to fetch the event message. The value should be PullPoint or BaseNotification.
	SubscribeType    = "subscribeType"
	PullPoint        = "PullPoint"
	BaseNotification = "BaseNotification"
	// DefaultSubscriptionPolicy is optional, we should check the camera capabilities before using
	DefaultSubscriptionPolicy = "defaultSubscriptionPolicy"
	// DefaultInitialTerminationTime indicates the subscription lifetime with specified duration.  For example, PT1H.
	DefaultInitialTerminationTime = "defaultInitialTerminationTime"
	// DefaultAutoRenew indicates the subscription will auto-renew before out of date. For example, true or false.
	DefaultAutoRenew = "defaultAutoRenew"
	// DefaultTopicFilter indicates the optional XPATH expression to filter the event. For example, tns1:RuleEngine/TamperDetector
	DefaultTopicFilter = "defaultTopicFilter"
	// DefaultMessageContentFilter indicates the optional XPATH expression to filter the event. For example, boolean(//tt:SimpleItem[@Name=”IsTamper”])
	DefaultMessageContentFilter = "defaultMessageContentFilter"
	// DefaultMessageTimeout specify the Timeout for PullMessage. Maximum time to block until this method returns. For example, PT5S
	DefaultMessageTimeout = "defaultMessageTimeout"
	// DefaultMessageLimit specify the MessageLimit for PullMessage. Upper limit for the number of messages to return at once, For example, 10
	DefaultMessageLimit = "defaultMessageLimit"
	// DefaultConsumerURL point to the consumer's network location
	DefaultConsumerURL = "defaultConsumerURL"

	Manufacturer    = "Manufacturer"
	Model           = "Model"
	FirmwareVersion = "FirmwareVersion"
	SerialNumber    = "SerialNumber"
	HardwareId      = "HardwareId"

	UnknownDevicePrefix = "unknown_unknown_"
)
