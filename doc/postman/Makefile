
.PHONY: gen install 

POSTMAN_COLLECTION = ../postman/device-onvif-camera.postman_collection.json
OPENAPI_FILE_JSON = openapi.json
OPENAPI_FILE_YAML = device-onvif-camera.yaml
POSTMAN_2_OPENAPI_OPTIONS = p2o-options.json

install:
	npm install -g postman-to-openapi
gen:
	@p2ojx extract $(POSTMAN_COLLECTION)
	@p2ojx build -p $(POSTMAN_COLLECTION)
	@yq -o=yaml $(OPENAPI_FILE_JSON) -P | tee $(OPENAPI_FILE_YAML)
	@rm $(OPENAPI_FILE_JSON)