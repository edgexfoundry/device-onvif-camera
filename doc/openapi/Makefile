
.PHONY: gen install schema-gen

POSTMAN_COLLECTION = ../postman/device-onvif-camera.postman_collection.json
OPENAPI_FILE = v2/device-onvif-camera.yaml
OPENAPI_TEMP_FILE = ./tmp/device-onvif-camera.yaml
POSTMAN_2_OPENAPI_OPTIONS = p2o-options.json
SIDECAR_FILE = sidecar.yaml
PROFILE_FILE = ../../cmd/res/profiles/camera.yaml

SCHEMA_POSTMAN = ./device-onvif-camera.postman_collection.json
SCHEMA_TEMP = ./openapi.json
SCHEMA_OUTPUT = ../openapi/tmp/device-onvif-camera.yaml

# todo: replace the API version number based on the actual source code version

$(OPENAPI_FILE) gen: $(SIDECAR_FILE) $(OPENAPI_TEMP_FILE) python/postprocess.py
#$(MAKE) schema-gen
	python3 python/postprocess.py $(OPENAPI_TEMP_FILE) $(SIDECAR_FILE) $(PROFILE_FILE) $(OPENAPI_FILE)


$(OPENAPI_TEMP_FILE): $(POSTMAN_COLLECTION) $(POSTMAN_2_OPENAPI_OPTIONS) | ./tmp
	p2o $(POSTMAN_COLLECTION) -f $(OPENAPI_TEMP_FILE) -o $(POSTMAN_2_OPENAPI_OPTIONS)

./tmp:
	mkdir -p $@

install:
	npm install -g postman-to-openapi
	npm install -g postman-to-openapi-json-extended
	pip3 install -r python/requirements.txt
	snap install yq

schema-gen:
	cd ../postman; p2ojx extract $(SCHEMA_POSTMAN)
	cd ../postman; p2ojx build
	cd ../postman; yq $(SCHEMA_TEMP)
