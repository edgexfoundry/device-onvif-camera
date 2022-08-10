
.PHONY: gen install replace

POSTMAN_COLLECTION = ../postman/device-onvif-camera.postman_collection.json
OPENAPI_FILE = v2/device-onvif-camera.yaml
POSTMAN_2_OPENAPI_OPTIONS = p2o-options.json

# todo: replace the API version number based on the actual source code version

$(OPENAPI_FILE) gen: $(POSTMAN_COLLECTION)
	p2o $(POSTMAN_COLLECTION) -f $(OPENAPI_FILE) -o $(POSTMAN_2_OPENAPI_OPTIONS)
	$(MAKE) replace

install:
	npm install -g postman-to-openapi


# Note: this is a basic find and replace to set the Example Camera name. In this future this can be expanded or replaced
# with a better solution.
define FINDSTR
      parameters:\n        - name: EDGEX_DEVICE_NAME\n          in: path\n          schema:\n            type: string\n          required: true
endef

define REPLACESTR
      parameters:\n        - name: EDGEX_DEVICE_NAME\n          in: path\n          schema:\n            type: string\n          required: true\n          example: Camera001
endef

# insert example camera name into the generated swagger file
replace:
	perl -0pe "s/$(FINDSTR)/$(REPLACESTR)/g" $(OPENAPI_FILE) > $(OPENAPI_FILE).bak
	mv -f $(OPENAPI_FILE).bak $(OPENAPI_FILE)
