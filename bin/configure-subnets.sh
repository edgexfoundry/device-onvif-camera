#!/usr/bin/env bash

#
# Copyright (C) 2020-2022 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to make it easier for an end user to configure Onvif device discovery
# without the need to have knowledge about subnets and/or CIDR format. The "DiscoverySubnets" config
# option defaults to blank in the configuration.toml file, and needs to be provided before a discovery can occur.
# This allows the device-onvif-camera device service to be run in a NAT-ed environment without host-mode networking,
# because the subnet information is user-provided and does not rely on device-onvif-camera to detect it.
#
# Essentially how this script works is it polls the machine it is running on and finds the active subnet for
# any and all network interfaces that are on the machine which are physical (non-virtual) and online.
# It uses this information to automatically fill out the "DiscoverySubnets" configuration option through Consul of a deployed
# device-onvif-camera instance.
#
# NOTE 1: This script requires EdgeX Consul and the device-onvif-camera service to have been run before this
# script will function.
#
# NOTE 2: If the "DiscoverySubnets" config is provided via "configuration.toml" this script does
# not need to be run.
#


set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]:-${0}}")")"

# shellcheck source=./utils.sh
source "${SCRIPT_DIR}/utils.sh"

main() {
    parse_args "$@"

    dependencies_check
    consul_check

    get_subnets

    log_info "\nSetting DiscoverySubnets to '${SUBNETS}'"
    put_consul_kv "${APPCUSTOM_BASE_KEY}/DiscoverySubnets" "${SUBNETS}"
}

main "$@"
