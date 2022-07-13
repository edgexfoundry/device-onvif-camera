#!/usr/bin/env bash

#
# Copyright (C) 2022 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to allow end-users to see what MAC Addresses are
# mapped to what credentials.
#

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]:-${0}}")")"

# shellcheck source=./utils.sh
source "${SCRIPT_DIR}/utils.sh"

main() {
    parse_args "$@"

    dependencies_check
    consul_check

    get_credentials_map

    printf "\n\n%20s:\n" "Credentials Map"
    for key in ${CREDENTIALS_MAP_KEYS}; do
        printf "%20s = '%s'\n" "$key" "${CREDENTIALS_MAP[$key]}"
    done
}

main "$@"
