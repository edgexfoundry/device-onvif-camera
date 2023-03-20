#!/usr/bin/env bash

#
# Copyright (C) 2022-2023 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to allow end-users to modify credentials either through
# EdgeX InsecureSecrets via Consul, or EdgeX Secrets via the device service.
#

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]:-${0}}")")"

# shellcheck source=./utils.sh
source "${SCRIPT_DIR}/utils.sh"

main() {
    parse_args "$@"

    dependencies_check
    consul_check

    if [ "${SECURE_MODE}" -eq 1 ] && [ -z "${REST_API_JWT}" ]; then
        query_rest_api_jwt
    fi

    pick_secret_name 0 0
    if [ "${SECRET_NAME}" == "NoAuth" ]; then
        log_error "NoAuth is a built-in value and cannot be modified by the user. It contains no actual credentials!"
        return 1
    fi
    create_or_update_credentials

    echo -e "${green}${bold}Success${clear}"
}

main "$@"
