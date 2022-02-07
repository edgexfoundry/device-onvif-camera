# Run device-onvif-camera with edgex-compose

The files in this directory are temporary and will be used to add device-onvif-camera into [edgex-compose](https://github.com/edgexfoundry/edgex-compose) in the near future.

> **Note to Developers**:
> *If you want to test the dockerized device-onvif-camera in this stage, please download [edgex-compose](https://github.com/edgexfoundry/edgex-compose) and follow the instructions below to generate a compose file including EdgeX Core services and device-onvif-camera service.*

## Prepare Compose file and Make file
1. Copy `add-device-onvif-camera.yml` from this directory to edgex-compose/compose-builder.

2. Go to edgex-compose/compose-builder.

3. Update `.env` file to add the registry and image version variable for device-onvif-camera.
```
DEVICE_ONVIFCAM_VERSION=latest
```
4. Update Makefile to add new option `ds-onvif-camera` to the [OPTIONS list](https://github.com/edgexfoundry/edgex-compose/blob/main/compose-builder/Makefile#L38).
```
define OPTIONS
 - arm64 no-secty dev app-dev -
 ...
 - ds-onvif-camera ds-bacnet ds-camera ds-grove ds-modbus ds-mqtt ds-rest ds-snmp ds-virtual ds-llrp ds-coap ds-gpio -
```

5. Update Makefile to add new device device-onvif-camera. Search for keyword `# Add Device Services` and then add the following code snippet below that line.
```
# Add Device Services
ifeq (ds-onvif-camera, $(filter ds-onvif-camera,$(ARGS)))
	ifeq (mqtt-bus, $(filter mqtt-bus,$(ARGS)))
	  extension_file:= $(shell GEN_EXT_DIR="$(GEN_EXT_DIR)" ./gen_mqtt_messagebus_compose_ext.sh device-onvif-camera -d)
	  COMPOSE_FILES:=$(COMPOSE_FILES) -f $(extension_file)
	endif
	ifneq (no-secty, $(filter no-secty,$(ARGS)))
		ifeq ($(TOKEN_LIST),"")
			TOKEN_LIST:=device-onvif-camera
		else
			TOKEN_LIST:=$(TOKEN_LIST),device-onvif-camera
		endif
		ifeq ($(KNOWN_SECRETS_LIST),"")
			KNOWN_SECRETS_LIST:=redisdb[device-onvif-camera]
		else
			KNOWN_SECRETS_LIST:=$(KNOWN_SECRETS_LIST),redisdb[device-onvif-camera]
		endif
		extension_file:= $(shell GEN_EXT_DIR="$(GEN_EXT_DIR)" ./gen_secure_compose_ext.sh device-onvif-camera)
		COMPOSE_FILES:=$(COMPOSE_FILES) -f $(extension_file)
	endif
	
	COMPOSE_FILES:=$(COMPOSE_FILES) -f add-device-onvif-camera.yml
endif
```

6. Update Makefile to add `add-device-onvif-camera.yml` to `COMPOSE_DOWN`
```
define COMPOSE_DOWN
	DEV= \
	APP_SVC_DEV= \
	ARCH= \
	TOKEN_LIST= \
	KNOWN_SECRETS_LIST= \
	EXTRA_PROXY_ROUTE_LIST= \
	docker-compose -p edgex \
		-f docker-compose-base.yml \
		-f add-device-bacnet.yml \
		...
		-f add-device-onvif-camera.yml \
```

7. Add the following snippets to `add-security.yml` for security mode
```yaml
  secretstore-setup:
    ...
    environment:
      ...
      ADD_SECRETSTORE_TOKENS: device-onvif-camera
      
  ...
  device-onvif-camera:
    env_file:
      - common-security.env
      - common-sec-stage-gate.env
    environment:
      SECRETSTORE_TOKENFILE: /tmp/edgex/secrets/device-onvif-camera/secrets-token.json
    entrypoint: ["/edgex-init/ready_to_run_wait_install.sh"]
    command: "/device-onvif-camera ${DEFAULT_EDGEX_RUN_CMD_PARMS}"
    volumes:
      - edgex-init:/edgex-init:ro,z
      - /tmp/edgex/secrets/device-onvif-camera:/tmp/edgex/secrets/device-onvif-camera:ro,z
    depends_on:
      - security-bootstrapper
      - secretstore-setup
      - database
```

## Run None Security Mode
1. Change the directory to `edgex-compose/compose-builder`
2. Enter the command `make gen no-secty ds-onvif-camera` to generate non-secure compose file
3. Or run the EdgeX with the command `make run no-secty ds-onvif-camera`

## Run Security Mode
1. Change the directory to `edgex-compose/compose-builder`
2. Enter the command `make gen ds-onvif-camera` to generate non-secure compose file
3. Or run the EdgeX with the command `make run ds-onvif-camera`
