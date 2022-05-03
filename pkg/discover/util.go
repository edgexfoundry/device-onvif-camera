// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package discover

// computeNetSz computes the total amount of valid IP addresses for a given subnet size
// Subnets of size 31 and 32 have only 1 valid IP address
// Ex. For a /24 subnet, computeNetSz(24) -> 254
func computeNetSz(subnetSz int) uint32 {
	if subnetSz >= 31 {
		return 1
	}
	return ^uint32(0)>>subnetSz - 1
}
