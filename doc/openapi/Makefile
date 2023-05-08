# Copyright (C) 2022 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

.PHONY: gen install clean

SHELL := /bin/bash

POSTMAN_COLLECTION ?= ../postman/device-onvif-camera.postman_collection.json
POSTMAN_ENV ?= ../postman/device-onvif-camera.postman_environment.json
OPENAPI_FILE ?= v2/device-onvif-camera.yaml
OPENAPI_TEMP_FILE ?= tmp/device-onvif-camera.yaml
POSTMAN_2_OPENAPI_OPTIONS ?= p2o-options.json
SIDECAR_FILE ?= sidecar.yaml
PROFILE_FILE ?= ../../cmd/res/profiles/camera.yaml
EDGEX_DOCS_BASE ?= ../../../edgex-docs/docs_src/microservices/device/supported/device-onvif-camera/supplementary-info
ONVIF_TESTED_FILE ?= $(EDGEX_DOCS_BASE)/ONVIF-protocol.md
ONVIF_FOOTNOTES_FILE ?= $(EDGEX_DOCS_BASE)/onvif-footnotes.md

# list of all services in the ./ref folder
SERVICES = analytics device event imaging media media2 ptz

# todo: replace the API version number based on the actual source code version

$(OPENAPI_FILE) gen: $(SIDECAR_FILE) $(OPENAPI_TEMP_FILE) $(PROFILE_FILE) $(SERVICES:%=ref/out/%.yaml) $(ONVIF_TESTED_FILE) $(ONVIF_FOOTNOTES_FILE) $(POSTMAN_ENV) python/postprocess.py python/matrix.py python/cleaner.py Makefile
	python3 python/postprocess.py $(OPENAPI_TEMP_FILE) $(SIDECAR_FILE) $(PROFILE_FILE) $(OPENAPI_FILE) $(ONVIF_TESTED_FILE) $(ONVIF_FOOTNOTES_FILE) $(POSTMAN_ENV)

$(OPENAPI_TEMP_FILE): $(POSTMAN_COLLECTION) $(POSTMAN_ENV) $(POSTMAN_2_OPENAPI_OPTIONS) | ./tmp
	p2o $(POSTMAN_COLLECTION) -f $(OPENAPI_TEMP_FILE) -o $(POSTMAN_2_OPENAPI_OPTIONS) > /dev/null

./tmp ./ref/out:
	mkdir -p $@

install:
	npm install -g postman-to-openapi
	pip3 install -r python/requirements.txt

# generate the ref/out/<service>.yaml files from the ref/<service>.yaml files
$(SERVICES:%=ref/out/%.yaml) : ref/out/%.yaml : ref/%.yaml python/xmlstrip.py | ./ref/out
	python/xmlstrip.py "$*" "ref/$*.yaml" "ref/out/$*.yaml"

clean:
	rm -rf ref/out tmp $(SCHEMA_TEMP_JSON) $(EXTRACTED_SCHEMA)
