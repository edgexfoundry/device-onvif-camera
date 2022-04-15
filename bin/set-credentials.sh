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

ALL="All Cameras"
CORE_METADATA_URL="${CORE_METADATA_URL:-http://localhost:59881}"
CONSUL_URL="${CONSUL_URL:-http://localhost:8500}"
DEVICE_SERVICE="${DEVICE_SERVICE:-device-onvif-camera}"
DEVICE_SERVICE_URL="${DEVICE_SERVICE_URL:-http://localhost:59984}"

BASE_URL="${CONSUL_URL}/v1/kv/edgex/devices/2.0/${DEVICE_SERVICE}/Writable/InsecureSecrets"

DEVICE_LIST=

DEVICE_NAME="${DEVICE_NAME:-}"
DEVICE_USERNAME="${DEVICE_USERNAME:-}"
DEVICE_PASSWORD="${DEVICE_PASSWORD:-}"
CURRENT_ARG=
SECURE_MODE=${SECURE_MODE:-0}

SELF_CMD="${0##*/}"

cleanup() {
    if [ -f "${curl_output}" ]; then
        rm -f "${curl_output}"
    fi
}

trap cleanup EXIT
curl_output="$(mktemp)"

print_output() {
    if [ -x "$(type -P jq)" ]; then
        jq . < "${curl_output}"
    else
        cat "${curl_output}"
        echo
    fi
}

# usage: do_curl "<payload>" curl_args...
do_curl() {
    echo '' > "${curl_output}"
    local payload="$1"
    shift
    echo "curl --data \"<redacted>\" $*" >&2
    # securely transfer the value through an auto-closing named pipe over stdin (prevent passwords on command line)
    local code
    code="$(curl -sS --location --data "@-" -w "%{http_code}" -o "${curl_output}" "$@" < <( set +x; echo -n "${payload}" ) || echo $?)"
    print_output
    if [ $((code)) -lt 200 ] || [ $((code)) -gt 299 ]; then
        echo -e "\033[31;1mFailed! curl returned a status code of '$((code))'\033[0m" >&2
        return $((code))
    fi
}

get_devices() {
    DEVICE_LIST="$(do_curl "" -X GET "${CORE_METADATA_URL}/api/v2/device/service/name/${DEVICE_SERVICE}" \
        | tr '{' '\n' \
        | sed -En 's/.*"name": *"([^"]+)".*/\1/p' \
        | xargs)"
}

pick_device() {
    get_devices

    PS3="Select a camera: "
    select DEVICE_NAME in "${ALL}" ${DEVICE_LIST}; do break; done

    if [ -z "${DEVICE_NAME}" ]; then
        echo "No device selected, exiting..."
        return 1
    fi
}

# usage: put_consul_field <sub-path> <value>
put_consul_field() {
    echo "Setting InsecureSecret: $1"

    do_curl "$2" -X PUT "${BASE_URL}/$1"
}

query_username_password() {
    if [ -z "${DEVICE_USERNAME}" ]; then
        printf "\033[1mUsername: \033[0m"
        read DEVICE_USERNAME

        if [ -z "${DEVICE_USERNAME}" ]; then
            echo "No username entered, exiting..."
            return 1
        fi
    fi

    if [ -z "${DEVICE_PASSWORD}" ]; then
        printf "\033[1mPassword: \033[0m"
        read -s DEVICE_PASSWORD
        echo

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

        -a | --all)
            DEVICE_NAME="${ALL}"
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
    do_curl "${payload}" -X POST "${DEVICE_SERVICE_URL}/api/v2/secret"
}

# usage: set_secrets "<device name>"
set_secrets() {
#     DEVICE_NAME="$1"
    if [ "${SECURE_MODE}" -eq 1 ]; then
        set_secure_secrets
    else
        set_insecure_secrets
    fi
}

main() {
    parse_args "$@"

    if [ -z "${DEVICE_NAME}" ]; then
        pick_device
    fi

    echo "Selected Device: ${DEVICE_NAME}"

    query_username_password

    if [ "${DEVICE_NAME}" == "${ALL}" ]; then
        get_devices
        for DEVICE_NAME in ${DEVICE_LIST}; do
            set_secrets
        done
    else
        set_secrets
    fi
}

main "$@"
