#!/usr/bin/env bash

#
# Copyright (C) 2022 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to allow end-users to assign credentials either through
# EdgeX InsecureSecrets via Consul, or EdgeX Secrets via the device service for discovered
# devices which lack them.
#

set -euo pipefail

CORE_METADATA_URL="${CORE_METADATA_HOST:-http://localhost:59881}"
CONSUL_URL="${CONSUL_HOST:-http://localhost:8500}"
DEVICE_SERVICE="${DEVICE_SERVICE:-device-onvif-camera}"
DEVICE_SERVICE_URL="${DEVICE_SERVICE_URL:-http://localhost:59984}"

BASE_URL="${CONSUL_URL}/v1/kv/edgex/devices/2.0/${DEVICE_SERVICE}/Writable/InsecureSecrets"

DEVICE_LIST=
DEVICE_COUNT=

DEVICE_NAME="${DEVICE_NAME:-}"
DEVICE_USERNAME="${DEVICE_USERNAME:-}"
DEVICE_PASSWORD="${DEVICE_PASSWORD:-}"
CURRENT_ARG=
SECURE_MODE=${SECURE_MODE:-0}

HEIGHT="$(tput lines)"
WIDTH="$(tput cols)"
SELF_CMD="${0##*/}"


get_devices() {
    DEVICE_LIST="$(curl --silent "${CORE_METADATA_URL}/api/v2/device/service/name/${DEVICE_SERVICE}" \
        | tr '{' '\n' \
        | sed -En 's/.*"name": *"([^"]+)".*/\1/p' \
        | grep -v "${DEVICE_SERVICE}" \
        | xargs)"

    DEVICE_COUNT=$(wc -w <<< "${DEVICE_LIST}")
}

pick_device() {
    get_devices

    local options=()
    for d in ${DEVICE_LIST}; do
        options+=("$d" "$d")
    done

    DEVICE_NAME=$(whiptail --menu "Please pick a device" --notags \
        "${HEIGHT}" "${WIDTH}" "${DEVICE_COUNT}" \
        "${options[@]}" 3>&1 1>&2 2>&3)

    if [ -z "${DEVICE_NAME}" ]; then
        echo "No device selected, exiting..."
        return 1
    fi
}

# usage: put_consul_field <sub-path> <value>
put_consul_field() {
    echo "Setting InsecureSecret: $1"
    local code
    # securely transfer the value through an auto-closing named pipe over stdin (prevent passwords on command line)
    code=$(curl -X PUT --data "@-" -w "%{http_code}" -o /dev/null -s "${BASE_URL}/$1" < <( set +x; echo -n "$2" ) || echo $?)
    if [ $((code)) -ne 200 ]; then
        echo -e "Failed! curl returned a status code of '${code}'"
        return $((code))
    fi
}

query_username_password() {
    if [ -z "${DEVICE_USERNAME}" ]; then
        DEVICE_USERNAME=$(whiptail --inputbox "Please specify device username" \
            "${HEIGHT}" "${WIDTH}" 3>&1 1>&2 2>&3)

        if [ -z "${DEVICE_USERNAME}" ]; then
            echo "No username entered, exiting..."
            return 1
        fi
    fi

    if [ -z "${DEVICE_PASSWORD}" ]; then
        DEVICE_PASSWORD=$(whiptail --passwordbox "Please specify device password" \
            "${HEIGHT}" "${WIDTH}" 3>&1 1>&2 2>&3)

        if [ -z "${DEVICE_PASSWORD}" ]; then
            echo "No password entered, exiting..."
            return 1
        fi
    fi
}

# usage: try_set_argument "arg_name" "$@"
# note: call shift AFTER this, as we want to see the flag_name as first argument after arg_name
try_set_argument() {
    local arg_name="$1"
    local flag_name="$2"
    shift 2
    if [ "$#" -lt 1 ]; then
        log_error "Missing required argument: ${flag_name} ${arg_name}"
        return 2
    fi
    declare -g "${arg_name}"="$1"
}

print_usage() {
    echo "Usage: ${SELF_CMD} [-s/--secure-mode] [-d <device_name>] [-u <username>] [-p <password>]"
}

parse_args() {
    while [ "$#" -gt 0 ]; do
        CURRENT_ARG="$1"
        case "${CURRENT_ARG}" in

        -s | --secure | --secure-mode)
            SECURE_MODE=1
            ;;

        -d | --device | --device-name)
            try_set_argument "DEVICE_NAME" "$@"
            shift
            ;;

        -u | --user | --username)
            try_set_argument "DEVICE_USERNAME" "$@"
            shift
            ;;

        -p | --pass | --password)
            try_set_argument "DEVICE_PASSWORD" "$@"
            shift
            ;;

        -c | --consul-url)
            try_set_argument "CONSUL_URL" "$@"
            shift
            ;;

        -m | --core-metadata-url)
            try_set_argument "CORE_METADATA_URL" "$@"
            shift
            ;;

        -U | --device-service-url)
            try_set_argument "DEVICE_SERVICE_URL" "$@"
            shift
            ;;

        --help)
            print_usage
            exit 0
            ;;

        *)
            echo "argument \"${CURRENT_ARG}\" not recognized."
            return 1
            ;;

        esac

        shift
    done
}

set_insecure_secrets() {
    put_consul_field "${DEVICE_NAME}/Path" "${DEVICE_NAME}"
    put_consul_field "${DEVICE_NAME}/Secrets/username" "${DEVICE_USERNAME}"
    put_consul_field "${DEVICE_NAME}/Secrets/password" "${DEVICE_PASSWORD}"
}

set_secure_secrets() {
    local payload="{
    \"apiVersion\":\"v2\",
    \"path\": \"${DEVICE_NAME}\",
    \"secretData\":[
        {
            \"key\":\"username\",
            \"value\":\"${DEVICE_USERNAME}\"
        },
        {
            \"key\":\"password\",
            \"value\":\"${DEVICE_PASSWORD}\"
        }
    ]
}"
    local code
    # securely transfer the value through an auto-closing named pipe over stdin (prevent passwords on command line)
    code=$(curl --location --request POST --data "@-" -w "%{http_code}" -o /dev/null -s "${DEVICE_SERVICE_URL}/api/v2/secret" < <( set +x; echo -n "${payload}" ) || echo $?)
    if [ $((code)) -ne 200 ]; then
        echo -e "Failed! curl returned a status code of '${code}'"
        return $((code))
    fi
}

main() {
    parse_args "$@"

    if [ -z "${DEVICE_NAME}" ]; then
        pick_device
    fi

    echo "Selected Device: ${DEVICE_NAME}"

    query_username_password

    if [ "${SECURE_MODE}" -eq 1 ]; then
        set_secure_secrets
    else
        set_insecure_secrets
    fi
}

main "$@"
