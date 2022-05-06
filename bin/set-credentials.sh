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

INSECURE_SECRETS_URL="${CONSUL_URL}/v1/kv/edgex/devices/2.0/${DEVICE_SERVICE}/Writable/InsecureSecrets"

DEVICE_LIST=

DEVICE_NAME="${DEVICE_NAME:-}"
DEVICE_USERNAME="${DEVICE_USERNAME:-}"
DEVICE_PASSWORD="${DEVICE_PASSWORD:-}"

# todo: auto-determine if service is running in secure mode
SECURE_MODE=${SECURE_MODE:-0}

SELF_CMD="${0##*/}"

# ANSI colors
red="\033[31m"
clear="\033[0m"
bold="\033[1m"
dim="\033[2m"

# print a message in bold
log_info() {
    echo -e "${bold}$*${clear}"
}

# print a message dimmed
log_debug() {
    echo -e "${dim}$*${clear}"
}

# log an error message to stderr in bold and red
log_error() {
    echo -e "${red}${bold}$*${clear}" >&2
}

# attempt to pretty print the output with jq. if jq is not available or
# jq fails to parse data, print it normally
format_output() {
    if [ ! -x "$(type -P jq)" ] || ! jq . <<< "$1" 2>/dev/null; then
        echo "$1"
    fi
    echo
}

# call the curl command with the specified payload and arguments.
# this function will print out the curl response and will return an error code
# if the curl request failed.
# usage: do_curl "<payload>" curl_args...
do_curl() {
    local payload="$1"
    shift
    # log the curl command so the user has insight into what the script is doing
    log_debug "curl --data \"<redacted>\" $*" >&2

    local code tmp output
    # the payload is securely transferred through an auto-closing named pipe.
    # this prevents any passwords or sensitive data being on the command line.
    # the http response code is written to stdout and stored in the variable 'code', while the full http response
    # is written to the temp file, and then read into the 'output' variable.
    tmp="$(mktemp)"
    code="$(curl -sS --location -w "%{http_code}" -o "${tmp}" "$@" --data "@"<( set +x; echo -n "${payload}" ) || echo $?)"
    output="$(<"${tmp}")"

    printf "Response [%3d] " "$((code))" >&2
    if [ $((code)) -lt 200 ] || [ $((code)) -gt 299 ]; then
        format_output "$output" >&2
        log_error "Failed! curl returned a status code of '${code}'"
        return $((code))
    else
        format_output "$output"
    fi
}

# this will update the device's Onvif protocol 'SecretPath' to equal
# the device name, which is the value inserted by this script
update_secret_path() {
    log_info "Patching protocols[\"Onvif\"].SecretPath to ${DEVICE_NAME}"
    local payload
    # query core metadata to get all the device information, and then
    # use sed to look for just the SecretPath and replace it. note that
    # currently this does not add one if it does not exist. Also, this
    # code might be better if it used jq, but then this script would require
    # the end user to have jq installed.
    payload="$(do_curl "" -X GET "${CORE_METADATA_URL}/api/v2/device/name/${DEVICE_NAME}" \
        | sed -E 's/"SecretPath" *: *"[^"]+"/"SecretPath":"'"${DEVICE_NAME}"'"/g')"
    # the patch endpoint requires an array of devices, so wrap in square brackets
    payload="[${payload}]"
    do_curl "${payload}" -X PATCH "${CORE_METADATA_URL}/api/v2/device"
}

# query EdgeX Core Metadata for the list of all devices
get_devices() {
    # grab the names of all devices for the specific device service.
    # filter out the fake control plane device (grep -v "${DEVICE_SERVICE}")
    DEVICE_LIST="$(do_curl "" -X GET "${CORE_METADATA_URL}/api/v2/device/service/name/${DEVICE_SERVICE}" \
        | tr '{' '\n' \
        | sed -En 's/.*"name": *"([^"]+)".*/\1/p' \
        | grep -v "${DEVICE_SERVICE}" \
        | xargs)"
    printf "\n\n"
}

# prompt the user to pick a device
pick_device() {
    get_devices

    # insert the option "All Cameras" first in the list. the reason first was chosen as opposed to
    # last was to keep the index of it the same no matter how many devices there are.
    PS3="Select a camera: "
    select DEVICE_NAME in "${ALL}" ${DEVICE_LIST}; do break; done

    if [ -z "${DEVICE_NAME}" ]; then
        log_error "No device selected, exiting..."
        return 1
    fi
    echo
}

# set an individual InsecureSecrets consul key to a specific value
# usage: put_insecure_secrets_field <sub-path> <value>
put_insecure_secrets_field() {
    log_info "Setting InsecureSecret: $1"
    do_curl "$2" -X PUT "${INSECURE_SECRETS_URL}/$1"
}

# prompt the user for the device's username and password
# and exit if not provided
query_username_password() {
    if [ -z "${DEVICE_USERNAME}" ]; then
        # shellcheck disable=SC2059
        printf "${bold}Username: ${clear}"
        read -r DEVICE_USERNAME

        if [ -z "${DEVICE_USERNAME}" ]; then
            log_error "No username entered, exiting..."
            return 1
        fi
    fi

    if [ -z "${DEVICE_PASSWORD}" ]; then
        # shellcheck disable=SC2059
        printf "${bold}Password: ${clear}"
        read -rs DEVICE_PASSWORD
        printf "\n\n"

        if [ -z "${DEVICE_PASSWORD}" ]; then
            log_error "No password entered, exiting..."
            return 1
        fi
    fi
}

# usage: try_set_argument "arg_name" "$@"
# attempts to set the global variable "arg_name" to the next value from the command line.
# if one is not provided, print error and return and error code.
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
    log_info "Usage: ${SELF_CMD} [-s/--secure-mode] [-d <device_name>] [-u <username>] [-p <password>] [-a/--all]"
}

parse_args() {
    while [ "$#" -gt 0 ]; do
        case "$1" in

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
            log_error "argument \"$1\" not recognized."
            return 1
            ;;

        esac

        shift
    done
}

# create or update the insecure secrets by setting the 3 required fields in Consul
set_insecure_secrets() {
    put_insecure_secrets_field "${DEVICE_NAME}/Path" "${DEVICE_NAME}"
    put_insecure_secrets_field "${DEVICE_NAME}/Secrets/username" "${DEVICE_USERNAME}"
    put_insecure_secrets_field "${DEVICE_NAME}/Secrets/password" "${DEVICE_PASSWORD}"
}

# set the secure secrets by posting to the device service's secret endpoint
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

# helper function to set the secrets using either secure or insecure mode, and then
# update the device's SecretPath.
set_secrets() {
    if [ "${SECURE_MODE}" -eq 1 ]; then
        set_secure_secrets
    else
        set_insecure_secrets
    fi

    update_secret_path
}

main() {
    parse_args "$@"

    if [ -z "${DEVICE_NAME}" ]; then
        pick_device
    fi

    log_info "Selected Device: ${DEVICE_NAME}"

    query_username_password

    if [ "${DEVICE_NAME}" == "${ALL}" ]; then
        get_devices # update the device list in the case where the user passed the --all flag
        for DEVICE_NAME in ${DEVICE_LIST}; do
            set_secrets
        done
    else
        set_secrets
    fi
}

main "$@"
