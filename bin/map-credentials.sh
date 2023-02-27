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

    if [ -z "${SECRET_NAME}" ]; then
        pick_secret_name 1 1
    fi
    log_info "Secret Name: ${SECRET_NAME}"

    # if the user manually passed credentials in via the command line, create or update the secret
    if [ "$USER_SET_CREDENTIALS" -eq 1 ]; then
        create_or_update_credentials
    fi

    query_mac_address

    put_credentials_map_field "${SECRET_NAME}" "${MAC_ADDRESSES}"

    echo -e "${green}${bold}Success${clear}"
}

main "$@"
