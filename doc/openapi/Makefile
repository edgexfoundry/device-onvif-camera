
.PHONY: gen install replace

POSTMAN_COLLECTION = ../postman/device-onvif-camera.postman_collection.json
OPENAPI_FILE = device-onvif-camera.yaml

$(OPENAPI_FILE) gen: $(POSTMAN_COLLECTION)
	p2o $(POSTMAN_COLLECTION) -f device-onvif-camera.yaml -o p2o-options.json
	$(MAKE) replace

install:
	npm install -g postman-to-openapi


define FINDSTR
      parameters:\n        - name: EDGEX_DEVICE_NAME\n          in: path\n          schema:\n            type: string\n          required: true
endef

define REPLACESTR
      parameters:\n        - name: EDGEX_DEVICE_NAME\n          in: path\n          schema:\n            type: string\n          required: true\n          example: Camera001
endef

# insert example camera name
replace:
	perl -0pe "s/$(FINDSTR)/$(REPLACESTR)/g" $(OPENAPI_FILE) > $(OPENAPI_FILE).bak
	mv -f $(OPENAPI_FILE).bak $(OPENAPI_FILE)
