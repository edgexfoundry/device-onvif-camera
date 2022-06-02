// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

type DiscoveryMode string

const (
	ModeNetScan   DiscoveryMode = "netscan"
	ModeMulticast DiscoveryMode = "multicast"
	ModeBoth      DiscoveryMode = "both"
)

func (mode DiscoveryMode) IsValid() bool {
	return mode == ModeNetScan || mode == ModeMulticast || mode == ModeBoth
}

func (mode DiscoveryMode) IsMulticastEnabled() bool {
	return mode == ModeMulticast || mode == ModeBoth
}

func (mode DiscoveryMode) IsNetScanEnabled() bool {
	return mode == ModeNetScan || mode == ModeBoth
}
