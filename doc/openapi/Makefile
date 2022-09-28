
.PHONY: gen install convert

POSTMAN_COLLECTION = ../postman/device-onvif-camera.postman_collection.json
OPENAPI_FILE = v2/device-onvif-camera.yaml
OPENAPI_TEMP_FILE = ./tmp/device-onvif-camera.yaml
POSTMAN_2_OPENAPI_OPTIONS = p2o-options.json
SIDECAR_FILE = sidecar.yaml

# todo: replace the API version number based on the actual source code version

$(OPENAPI_FILE) gen: $(POSTMAN_COLLECTION) | ./tmp
	p2o $(POSTMAN_COLLECTION) -f $(OPENAPI_TEMP_FILE) -o $(POSTMAN_2_OPENAPI_OPTIONS)
	$(MAKE) convert

./tmp:
	mkdir -p $@

install:
	npm install -g postman-to-openapi
	pip3 install -r python/requirements.txt

convert: $(SIDECAR_FILE) $(OPENAPI_TEMP_FILE)
	python3 python/postprocess.py $(OPENAPI_TEMP_FILE) $(SIDECAR_FILE) $(OPENAPI_FILE)
