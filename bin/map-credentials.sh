#!/usr/bin/env bash

#
# Copyright (C) 2022 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to allow end-users to add credentials either through
# EdgeX InsecureSecrets via Consul, or EdgeX Secrets via the device service. It then allows the
# end-user to add a list of MAC Addresses to map to those credentials via Consul.
#

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]:-${0}}")")"

# shellcheck source=./utils.sh
source "${SCRIPT_DIR}/utils.sh"

main() {
    parse_args "$@"

    dependencies_check
    consul_check

    if [ -z "${SECRET_PATH}" ]; then
        pick_secret_path
    fi
    log_info "Secret Path: ${SECRET_PATH}"

    query_mac_address

    put_credentials_map_field "${SECRET_PATH}" "${MAC_ADDRESSES}"
}

main "$@"
