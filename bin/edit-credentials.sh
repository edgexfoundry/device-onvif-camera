#!/usr/bin/env bash

#
# Copyright (C) 2022 Intel Corporation
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

    pick_secret_path 0 0
    if [ "${SECRET_PATH}" == "NoAuth" ]; then
        log_error "NoAuth is a built-in value and cannot be modified by the user. It contains no actual credentials!"
        return 1
    fi
    create_or_update_credentials

    echo -e "${green}${bold}Success${clear}"
}

main "$@"
