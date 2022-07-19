#!/usr/bin/env bash

#
# Copyright (C) 2022 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to be sourced by other utility scripts from
# this service in order to reduce duplicated code.
#

CORE_METADATA_URL="${CORE_METADATA_URL:-http://localhost:59881}"
CONSUL_URL="${CONSUL_URL:-http://localhost:8500}"
DEVICE_SERVICE="${DEVICE_SERVICE:-device-onvif-camera}"
DEVICE_SERVICE_URL="${DEVICE_SERVICE_URL:-http://localhost:59984}"

SECRET_PATH="${SECRET_PATH:-}"
SECRET_USERNAME="${SECRET_USERNAME:-}"
SECRET_PASSWORD="${SECRET_PASSWORD:-}"
AUTH_MODE="${AUTH_MODE:-}"
MAC_ADDRESSES="${MAC_ADDRESSES:-}"

CONSUL_KV_BASE_URL="${CONSUL_URL}/v1/kv"
CONSUL_BASE_KEY="edgex/devices/2.0/${DEVICE_SERVICE}"
APPCUSTOM_BASE_KEY="${CONSUL_BASE_KEY}/AppCustom"
CREDENTIALS_MAP_KEY="${APPCUSTOM_BASE_KEY}/CredentialsMap"
WRITABLE_BASE_KEY="${CONSUL_BASE_KEY}/Writable"
INSECURE_SECRETS_KEY="${WRITABLE_BASE_KEY}/InsecureSecrets"

CONSUL_TOKEN="${CONSUL_TOKEN:-}"

CREDENTIALS_MAP_KEYS=
CREDENTIALS_COUNT=
declare -A CREDENTIALS_MAP

NET_IFACES=
SUBNETS=
CURL_CODE=
CURL_OUTPUT=

# note: we must use a separate array here to preserve order
AUTH_MODES=("usernametoken" "digest" "both")
declare -A AUTH_MODES_DESC=(
    ["usernametoken"]="Username/Token"
    ["digest"]="Digest Auth"
    ["both"]="Both"
)

# todo: auto-determine if service is running in secure mode
SECURE_MODE=${SECURE_MODE:-0}

SELF_CMD="${0##*/}"

# ANSI colors
red="\033[31m"
green="\033[32m"
clear="\033[0m"
bold="\033[1m"
dim="\033[2m"
normal="\e[22;24m"

# these are used for printing out messages
spacing=18
prev_line="\e[1A\e[$((spacing + 2))C"

ADD_NEW_SECRET_KEY="_ADD_NEW_"

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
    # redact the consul token and password in case of sensitive data
    local redacted_args="${*//${CONSUL_TOKEN}/<redacted>}"
    redacted_args="${redacted_args//-H X-Consul-Token: /}"
    local redacted_data=""
    if [ -n "${payload}" ]; then
        redacted_data="--data '${payload//${SECRET_PASSWORD}/<redacted>}' "
    fi
    log_debug "curl ${redacted_data}${redacted_args}" >&2

    local tmp code output
    # the payload is securely transferred through an auto-closing named pipe.
    # this prevents any passwords or sensitive data being on the command line.
    # the http response code is written to stdout and stored in the variable 'code', while the full http response
    # is written to the temp file, and then read into the 'output' variable.
    tmp="$(mktemp)"
    code="$(curl -sS --location -w "%{http_code}" -o "${tmp}" "$@" --data "@"<( set +x; echo -n "${payload}" ) || echo $?)"
    output="$(<"${tmp}")"

    declare -g CURL_CODE="$((code))"
    declare -g CURL_OUTPUT="${output}"
    printf "Response [%3d] " "$((code))" >&2
    if [ $((code)) -lt 200 ] || [ $((code)) -gt 299 ]; then
        format_output "$output" >&2
        log_error "Failed! curl returned a status code of '$((code))'"
        return $((code))
    else
        format_output "$output"
    fi

    echo >&2
}

# query EdgeX Core Metadata for the list of all devices
get_devices() {
    # grab the names of all devices for the specific device service.
    # filter out the fake control plane device (grep -v "${DEVICE_SERVICE}")
    DEVICE_LIST="$(do_curl "" -X GET "${CORE_METADATA_URL}/api/v2/device/service/name/${DEVICE_SERVICE}" \
        | tr '{' '\n' \
        | sed -En 's/.*"name": *"([^"]+)".*/\1/p' \
        | grep -v "${DEVICE_SERVICE}" \
        | sort -u \
        | xargs)"
    printf "\n\n"

    DEVICE_COUNT=$(wc -w <<< "${DEVICE_LIST}")
}

# prompt the user to pick a device
pick_device() {
    get_devices

    # insert the option "All Cameras" first in the list. the reason first was chosen as opposed to
    # last was to keep the index of it the same no matter how many devices there are.
    local options=("ALL" "All Cameras")
    for d in ${DEVICE_LIST}; do
        options+=("$d" "$d")
    done

    DEVICE_NAME=$(whiptail --menu "Please pick a device" --notags \
        0 0 "${DEVICE_COUNT}" \
        "${options[@]}" 3>&1 1>&2 2>&3)

    if [ -z "${DEVICE_NAME}" ]; then
        log_error "No device selected, exiting..."
        return 1
    fi
    echo
}

# prompt the user to pick an auth mode
query_auth_mode() {
    local options=()
    for mode in "${AUTH_MODES[@]}"; do
        options+=("$mode" "${AUTH_MODES_DESC[$mode]}")
    done

    AUTH_MODE=$(whiptail --menu "Select an authentication mode" --notags \
        0 0 3 \
        "${options[@]}" 3>&1 1>&2 2>&3)
}

query_consul_token() {
    CONSUL_TOKEN=$(whiptail --inputbox "Enter Consul ACL Token (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)" \
        10 0 3>&1 1>&2 2>&3)

    if [ -z "${CONSUL_TOKEN}" ]; then
        log_error "No Consul token entered, exiting..."
        return 1
    fi
}

# usage: get_consul_kv <path from base> <args>
get_consul_kv() {
    if [ "${SECURE_MODE}" -eq 1 ] && [ -z "${CONSUL_TOKEN}" ]; then
        query_consul_token
    fi

    do_curl "" -H "X-Consul-Token:${CONSUL_TOKEN}" -X GET "${CONSUL_KV_BASE_URL}/$1?$2"
}

# returns a list of keys for a given sub-path in sorted order
# usage: get_consul_kv_keys <path from base>
get_consul_kv_keys() {
    get_consul_kv "$1" "keys=true" \
        | tr ',' '\n' \
        | sed -En "s|.*\"$1\/([^\"]+)\".*|\1|p" \
        | sort -u \
        | xargs
}

# returns the raw value for a given key
# usage: get_consul_kv_raw <path from base>
get_consul_kv_raw() {
    get_consul_kv "$1" "raw=true"
}

# usage: put_consul_kv <key path from base> <value>
put_consul_kv() {
    if [ "${SECURE_MODE}" -eq 1 ] && [ -z "${CONSUL_TOKEN}" ]; then
        query_consul_token
    fi

    do_curl "$2" -H "X-Consul-Token:${CONSUL_TOKEN}" -X PUT "${CONSUL_KV_BASE_URL}/$1"
}

# set an individual InsecureSecrets consul key to a specific value
# usage: put_insecure_secrets_field <sub-path> <value>
put_insecure_secrets_field() {
    log_info "Setting InsecureSecret: $1"
    put_consul_kv "${INSECURE_SECRETS_KEY}/$1" "$2"
}

# usage: put_credentials_map_field <secret-path> <value>
put_credentials_map_field() {
    log_info "Setting Credentials Map: $1 = '$2'"
    put_consul_kv "${CREDENTIALS_MAP_KEY}/$1" "$2"
}

create_or_update_credentials() {
    if [ -z "${SECRET_PATH}" ]; then
        query_secret_path
    fi
    log_info "Secret Path: ${SECRET_PATH}"
    # we need to inject the secret path into the map to avoid key not found errors later on
    CREDENTIALS_MAP[$SECRET_PATH]=$(get_credentials_map_field "$SECRET_PATH" 2>/dev/null || printf '')

    query_username_password

    if [ -z "${AUTH_MODE}" ]; then
        query_auth_mode
    fi

    # store the credentials
    set_secret

    put_credentials_map_field "${SECRET_PATH}" "${CREDENTIALS_MAP[$SECRET_PATH]}"
}

# prompt the user to pick a secret path mapping
pick_secret_path() {
    get_credentials_map_keys

    local options=()
    for d in ${CREDENTIALS_MAP_KEYS}; do
        options+=("$d" "$d")
    done
    # insert the option "_ADD_NEW_" last in the list. if the user selects this, prompt them to
    # create a new secret
    options+=("${ADD_NEW_SECRET_KEY}" "(Create New)")

    SECRET_PATH=$(whiptail --menu "Please pick credentials" --notags \
        0 0 "${CREDENTIALS_COUNT}" \
        "${options[@]}" 3>&1 1>&2 2>&3)

    if [ "${SECRET_PATH}" == "${ADD_NEW_SECRET_KEY}" ]; then
        SECRET_PATH=""
        create_or_update_credentials
    fi

    if [ -z "${SECRET_PATH}" ]; then
        log_error "No secret path selected, exiting..."
        return 1
    fi
    echo
}

# sets CREDENTIALS_MAP_KEYS to an array of secret names, and CREDENTIALS_COUNT to the count of secrets
get_credentials_map_keys() {
    CREDENTIALS_MAP_KEYS="$(get_consul_kv_keys "${CREDENTIALS_MAP_KEY}")"
    CREDENTIALS_COUNT=$(wc -w <<< "${CREDENTIALS_MAP_KEYS}")
}

# retrieves the list of secret keys, and then queries the value of each one and puts them in CREDENTIALS_MAP
get_credentials_map() {
    get_credentials_map_keys

    for key in ${CREDENTIALS_MAP_KEYS}; do
        CREDENTIALS_MAP[$key]=$(get_credentials_map_field "$key")
    done
}

# usage: get_credentials_map_field <secret-path>
get_credentials_map_field() {
    get_consul_kv_raw "${CREDENTIALS_MAP_KEY}/$1"
}

# prompt the user for a name for the secret path
query_secret_path() {
    if [ -z "${SECRET_PATH}" ]; then
        SECRET_PATH=$(whiptail --inputbox "Enter a name for the credentials (aka Secret Path)" \
            10 0 3>&1 1>&2 2>&3)

        if [ -z "${SECRET_PATH}" ]; then
            log_error "No secret path entered, exiting..."
            return 1
        fi
    fi
}

# prompt the user for a mac address, and pre-fill it with the existing value.
# todo: mac address and csv field validation
query_mac_address() {
    if [ -z "${MAC_ADDRESSES}" ]; then
        CREDENTIALS_MAP[$SECRET_PATH]=$(get_credentials_map_field "$SECRET_PATH")
        MAC_ADDRESSES=$(whiptail --inputbox "Enter one or more mac addresses to associate with credentials: '${SECRET_PATH}'" \
            10 0 "${CREDENTIALS_MAP[$SECRET_PATH]}" 3>&1 1>&2 2>&3)

        if [ -z "${MAC_ADDRESSES}" ]; then
            log_error "No mac addresses entered, exiting..."

            return 1
        fi
    fi
}

# prompt the user for the credential's username and password
# and exit if not provided
query_username_password() {
    if [ -z "${SECRET_USERNAME}" ]; then
        SECRET_USERNAME=$(whiptail --inputbox "Enter username for ${SECRET_PATH}" \
            10 0 3>&1 1>&2 2>&3)

        if [ -z "${SECRET_USERNAME}" ]; then
            log_error "No username entered, exiting..."
            return 1
        fi
    fi

    if [ -z "${SECRET_PASSWORD}" ]; then
        SECRET_PASSWORD=$(whiptail --passwordbox "Enter password for ${SECRET_PATH}" \
            10 0 3>&1 1>&2 2>&3)

        if [ -z "${SECRET_PASSWORD}" ]; then
            log_error "No password entered, exiting..."
            return 1
        fi
    fi
}

# Detect online physical network interfaces
get_net_ifaces() {
    NET_IFACES=$(
        find /sys/class/net -mindepth 1 -maxdepth 2   `# list all network interfaces`  \
            -not -lname '*devices/virtual*'           `# filter out all virtual interfaces` \
            -execdir grep -q 'up' "{}/operstate" \;   `# ensure interface is online (operstate == up)` \
            -printf '%f\n'                            `# print them one per line` \
            | paste -sd\| -                           `# join them separated by | for regex matching`
    )

    if [ -z "${NET_IFACES}" ]; then
        log_error "No online physical network interfaces detected."
        return 1
    fi
}

# Detect active physical ipv4 subnets
#
# print all ipv4 subnets, filter for just the ones associated with our physical interfaces,
# grab the unique ones and join them by commas
#
# sed -n followed by "s///p" means find and print (with replacements) only the lines containing a match
# 'eno1|eno2' becomes "s/ dev (eno1|eno2).+//p"
# (eno1|eno2) is a matched group of possible values (| means OR)
# .+ is a catch-all to prevent printing the rest of the line
#
# Example Input:
#   10.0.0.0/24 dev eno1 proto kernel src 10.0.0.212 metric 600
#   192.168.1.0/24 dev eno2 proto kernel src 192.168.1.134 metric 900
#   172.17.0.0/16 dev docker0 proto kernel src 172.17.0.1 linkdown
#
# Example Output:
#   10.0.0.0/24
#   192.168.1.0/24
#
# Explanation:
# - The first line matched the 'eno1' interface, so everything starting from " dev eno1 ..."
#     is stripped out, and we are left with just the subnet (10.0.0.0/24).
# - The second line matched the 'eno2' interface, same process as before, and we are left with just the subnet.
# - The third line does not match either interface and is not printed.
get_subnets() {
    get_net_ifaces

    SUBNETS=$(
        # Print all IPv4 routes, one per line
        ip -4 -o route list scope link |
            # Regex match it against all of our online physical interfaces
            sed -En "s/ dev (${NET_IFACES}).+//p" |
            # Remove [link-local subnet](https://en.wikipedia.org/wiki/Link-local_address) using grep reverse match (-v)
            grep -v "169.254.0.0/16" |
            # Sort and remove potential duplicates
            sort -u |
            # Merge all lines into a single line separated by commas (no trailing ,)
            paste -sd, -
    )

    if [ -z "${SUBNETS}" ]; then
        log_error "No subnets detected."
        return 1
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
    log_info "Usage: ${SELF_CMD} [-s/--secure-mode] [-d <device_name>] [-u <username>] [-p <password>] [-a/--all] [--auth-mode <auth mode>] [-t <consul token>]"
}

parse_args() {
    while [ "$#" -gt 0 ]; do
        case "$1" in

        -s | --secure | --secure-mode)
            SECURE_MODE=1
            ;;

        -t | --token | --consul-token)
            try_set_argument "CONSUL_TOKEN" "$@"
            shift
            ;;

        -d | --device | --device-name)
            try_set_argument "DEVICE_NAME" "$@"
            shift
            ;;

        -A | --auth | --auth-mode)
            try_set_argument "AUTH_MODE" "$@"
            shift
            ;;

        -a | --all)
            DEVICE_NAME="${ALL}"
            ;;

        -u | --user | --username)
            try_set_argument "SECRET_USERNAME" "$@"
            shift
            ;;

        -p | --pass | --password)
            try_set_argument "SECRET_PASSWORD" "$@"
            shift
            ;;

        -P | --path | --secret-path)
            try_set_argument "SECRET_PATH" "$@"
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
set_insecure_secret() {
    put_insecure_secrets_field "${SECRET_PATH}/Path"                "${SECRET_PATH}"
    put_insecure_secrets_field "${SECRET_PATH}/Secrets/username"    "${SECRET_USERNAME}"
    put_insecure_secrets_field "${SECRET_PATH}/Secrets/password"    "${SECRET_PASSWORD}"
    put_insecure_secrets_field "${SECRET_PATH}/Secrets/mode"        "${AUTH_MODE}"
}

# set the secure secrets by posting to the device service's secret endpoint
set_secure_secret() {
    local payload="{
    \"apiVersion\":\"v2\",
    \"path\": \"${SECRET_PATH}\",
    \"secretData\":[
        {
            \"key\":\"username\",
            \"value\":\"${SECRET_USERNAME}\"
        },
        {
            \"key\":\"password\",
            \"value\":\"${SECRET_PASSWORD}\"
        },
        {
            \"key\":\"mode\",
            \"value\":\"${AUTH_MODE}\"
        }
    ]
}"
    do_curl "${payload}" -X POST "${DEVICE_SERVICE_URL}/api/v2/secret"
}

# helper function to set the secrets using either secure or insecure mode
set_secret() {
    if [ "${SECURE_MODE}" -eq 1 ]; then
        set_secure_secret
    else
        set_insecure_secret
    fi
}

# Dependencies Check
dependencies_check() {
    printf "${bold}%${spacing}s${clear}: ...\n" "Dependencies Check"
    if ! type -P curl >/dev/null; then
        log_error "${prev_line}${bold}${red}Failed!${normal}\nPlease install ${bold}curl${normal} in order to use this script!${clear}"
        return 1
    fi
    echo -e "${prev_line}${green}Success${clear}"
}

check_consul_return_code() {
    if [ $((CURL_CODE)) -ne 200 ]; then
        if [ $((CURL_CODE)) -eq 7 ]; then
            # Special message for error code 7
            echo -e "${red}* Error code '7' denotes 'Failed to connect to host or proxy'${clear}"
        elif [ $((CURL_CODE)) -eq 404 ]; then
            # Error 404 means it connected to consul but couldn't find the key
            echo -e "${red}* Have you deployed the ${bold}${DEVICE_SERVICE}${normal} service?${clear}"
        elif [ $((CURL_CODE)) -eq 401 ]; then
            if [ "${CURL_OUTPUT}" == "ACL support disabled" ]; then
                SECURE_MODE=0
                CONSUL_TOKEN=""
                return
            fi
            echo -e "${red}* Are you running in secure mode? Is your Consul token correct?${clear}"
        elif [ $((CURL_CODE)) -eq 403 ]; then
            # Error 401 and 403 are authentication errors
            if [ -z "${CONSUL_TOKEN}" ]; then
                SECURE_MODE=1
                query_consul_token
                consul_check
                return
            fi
            echo -e "${red}* Are you running in secure mode? Is your Consul token correct?${clear}"
        else
            echo -e "${red}* Is Consul deployed and accessible?${clear}"
        fi
        return $((CURL_CODE))
    fi
}

# Consul Check
consul_check() {
    printf "${bold}%${spacing}s${clear}: ...\n%${spacing}s  " "Consul Check" ""

    # use || true because we want to handle the result and not let the script auto exit
    do_curl '[{"Resource":"key","Access":"read"},{"Resource":"key","Access":"write"}]' \
        -H "X-Consul-Token:${CONSUL_TOKEN}" -X POST "${CONSUL_URL}/v1/internal/acl/authorize" 2>/dev/null || true
    check_consul_return_code


    if [ $((CURL_CODE)) -eq 200 ]; then
        local authorized
        # use || true because we want to handle the result and not let the script auto exit
        # this could be parsed better if using `jq`, but don't want to require the user to have it installed
        authorized=$(grep -c '"Allow":true'<<<"${CURL_OUTPUT}" || true)
        if [ $((authorized)) -ne 2 ]; then
            SECURE_MODE=1
            query_consul_token
        fi
    fi

    # use || true because we want to handle the result and not let the script auto exit
    get_consul_kv "${CONSUL_BASE_KEY}" "keys=true" > /dev/null || true
    check_consul_return_code

    echo -e "${prev_line}${green}Success${clear}"
}
