
.PHONY: gen install schema-gen

POSTMAN_COLLECTION = ../postman/device-onvif-camera.postman_collection.json
OPENAPI_FILE = v2/device-onvif-camera.yaml
OPENAPI_TEMP_FILE = ./tmp/device-onvif-camera.yaml
POSTMAN_2_OPENAPI_OPTIONS = p2o-options.json
SIDECAR_FILE = sidecar.yaml
PROFILE_FILE = ../../cmd/res/profiles/camera.yaml

SCHEMA_POSTMAN = ./device-onvif-camera.postman_collection.json
SCHEMA_TEMP = ../postman/openapi.json
OPENAPI_SCHEMA_TEMP_FILE = ./tmp/device-onvif-camera.schema.yaml
EXTRACTED_SCHEMA = ../postman/schemas/_extracted.json

# todo: replace the API version number based on the actual source code version

$(OPENAPI_FILE) gen: $(SIDECAR_FILE) $(OPENAPI_SCHEMA_TEMP_FILE) $(PROFILE_FILE) python/postprocess.py
	python3 python/postprocess.py $(OPENAPI_SCHEMA_TEMP_FILE) $(SIDECAR_FILE) $(PROFILE_FILE) $(OPENAPI_FILE)
	rm -rf ./tmp


$(OPENAPI_TEMP_FILE): $(POSTMAN_COLLECTION) $(POSTMAN_2_OPENAPI_OPTIONS) | ./tmp
	p2o $(POSTMAN_COLLECTION) -f $(OPENAPI_TEMP_FILE) -o $(POSTMAN_2_OPENAPI_OPTIONS)

./tmp:
	mkdir -p $@

install:
	npm install -g postman-to-openapi
	npm install -g postman-to-openapi-json-extended
	pip3 install -r python/requirements.txt
	snap install yq

$(SCHEMA_TEMP) schema-gen:
	cd ../postman && \
	p2ojx extract $(SCHEMA_POSTMAN) && \
	p2ojx build
	rm -f $(EXTRACTED_SCHEMA)

$(OPENAPI_SCHEMA_TEMP_FILE): $(SCHEMA_TEMP) | ./tmp
	yq . -o=yaml $(SCHEMA_TEMP) -P > $(OPENAPI_SCHEMA_TEMP_FILE)
	rm -f $(SCHEMA_TEMP)
