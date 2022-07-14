#
# Copyright (c) 2022 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG BASE=golang:1.17-alpine3.15
FROM ${BASE} AS builder

ARG MAKE="make build"
ARG ALPINE_PKG_BASE="make git gcc libc-dev zeromq-dev libsodium-dev"
ARG ALPINE_PKG_EXTRA=""

LABEL Name=edgex-device-onvif-camera

RUN sed -e 's/dl-cdn[.]alpinelinux.org/dl-4.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}

WORKDIR /device-onvif-camera

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .

RUN ${MAKE}

FROM alpine:3.15

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2022: Intel Corporation'

# dumb-init needed for injected secure bootstrapping entrypoint script when run in secure mode.
RUN apk add --update --no-cache zeromq dumb-init

WORKDIR /
COPY --from=builder /device-onvif-camera/cmd /
COPY --from=builder /device-onvif-camera/LICENSE /
COPY --from=builder /device-onvif-camera/Attribution.txt /

EXPOSE 59984

ENTRYPOINT ["/device-onvif-camera"]
CMD ["--cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
