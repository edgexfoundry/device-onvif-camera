# Set Up Auto Discovery with Docker

This is a guide on how to configure the ONVIF device service for automatic device discovery using Docke.

<img alt="overview" src="images/auto-discovery-docker-overview.jpg" width="75%"/>

Note: For macOS, the network_mode: "host" probably does not work as expected: https://github.com/docker/for-mac/issues/1031

## Non-Security Mode

### Prepare edgex-compose/compose-builder

#### 1. Download the [edgex-compose](https://github.com/edgexfoundry/edgex-compose) and setup it according to the [docker-compose setup guide](./docker-compose/README.md)

#### 2. Update  the `add-device-onvif-camera.yml` file with the following content:

```yaml
version: '3.7'

services:
  device-onvif-camera:
    image: edgexfoundry/device-onvif-camera${ARCH}:${DEVICE_ONVIFCAM_VERSION}
    ports:
      - "127.0.0.1:59984:59984"
    container_name: edgex-device-onvif-camera
    hostname: edgex-device-onvif-camera
    read_only: true
    restart: always
    networks:
      - edgex-network
    env_file:
      - common.env
      - device-common.env
    environment:
      SERVICE_HOST: edgex-device-onvif-camera
      MESSAGEQUEUE_HOST: edgex-redis
    depends_on:
      - consul
      - data
      - metadata
    security_opt:
      - no-new-privileges:true
    user: "${EDGEX_USER}:${EDGEX_GROUP}"

```
> Example add-device-onvif-camera.yml contents

### Deploy EdgeX services and device-onvif-camera
Deploy services with the following command:
```shell
make run no-secty ds-onvif-camera
```

### Inspect the device-onvif-camera
The user can use docker logs to trace the auto-discovery
```shell
$ docker logs edgex-device-onvif-camera -f --tail 10
...
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10011          EndpointRefAddress: 76a3186a-bcc2-43a2-9ef5-458adbf2262e
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10012          EndpointRefAddress: 83f9a999-149b-44ca-88a2-d58a766c738e
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10013          EndpointRefAddress: 2a3d6329-91b4-41de-8d4d-85bb2c737c02
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10014          EndpointRefAddress: dc9e23a7-50db-44f7-bb34-a9b1d3c87fa8
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10015          EndpointRefAddress: cf0280d6-4909-4671-8972-90adb3e60181
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10016          EndpointRefAddress: da1cd97d-d18a-40fd-a250-1f0c4f13f293
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10017          EndpointRefAddress: 8af3658d-adf9-4898-8146-86a0cb50c9fb
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10018          EndpointRefAddress: 436eb565-ae32-4d79-a9a3-6d520af2c647
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10019          EndpointRefAddress: 6e17721c-d861-4609-880b-efd72e00b8bc
level=INFO ts=2022-05-17T17:17:03.069067014Z app=device-onvif-camera source=driver.go:422 msg="Discovered 20 device(s) in 1.133374208s via netscan."
```
Then user can follow [this doc to add a provision watcher](./auto-discovery.md) to add the discovered devices to EdgeX.


## Security Mode

### Prepare edgex-compose/compose-builder

#### 1. Download the [edgex-compose](https://github.com/edgexfoundry/edgex-compose) and setup it according to the [docker-compose setup guide](./docker-compose/README.md)

#### 2. Replace the `add-device-onvif-camera.yml` with the following content:
```yaml
version: '3.7'

services:
  device-onvif-camera:
    image: edgexfoundry/device-onvif-camera${ARCH}:${DEVICE_ONVIFCAM_VERSION}
    ports:
      - "127.0.0.1:59984:59984"
    container_name: edgex-device-onvif-camera
    hostname: edgex-device-onvif-camera
    read_only: true
    restart: always
    networks:
      - edgex-network
    env_file:
      - common.env
      - device-common.env
    environment:
      SERVICE_HOST: edgex-device-onvif-camera
      MESSAGEQUEUE_HOST: edgex-redis
      SECRETSTORE_HOST: localhost
      STAGEGATE_BOOTSTRAPPER_HOST: localhost
      STAGEGATE_READY_TORUNPORT: 54329
      STAGEGATE_WAITFOR_TIMEOUT: 60s
    depends_on:
      - consul
      - data
      - metadata
    security_opt:
      - no-new-privileges:true
    user: "${EDGEX_USER}:${EDGEX_GROUP}"
```

#### 3. Export the Security Bootstrapper
Open the `add-security.yml` file and modify the `security-bootstrapper` section to export the port. This port is used for the device-onvif-camera to wait for the security setup.
```yaml
services:
  security-bootstrapper:
    ...
    ports:
      - "54329:54329"
```

### Deploy EdgeX services and device-onvif-camera
Deploy services with the following command:
```shell
make run device-onvif-camera
```

### Add Secrets to Secret Store (Vault)

```shell
curl --request POST 'http://192.168.56.101:59984/api/v2/secret' \
--header 'Content-Type: application/json' \
--data-raw '{
    "apiVersion":"v2",
    "path": "credentials001",
    "secretData":[
        {
            "key":"username",
            "value":"administrator"
        },
        {
            "key":"password",
            "value":"Password1!"
        }
    ]
}'
```

**Note**: The user should replace the host IP to match their own machine IP

### Inspect the device-onvif-camera
The user can docker logs to trace the auto-discovery
```shell
$ docker logs edgex-device-onvif-camera -f --tail 10
...
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10011          EndpointRefAddress: 76a3186a-bcc2-43a2-9ef5-458adbf2262e
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10012          EndpointRefAddress: 83f9a999-149b-44ca-88a2-d58a766c738e
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10013          EndpointRefAddress: 2a3d6329-91b4-41de-8d4d-85bb2c737c02
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10014          EndpointRefAddress: dc9e23a7-50db-44f7-bb34-a9b1d3c87fa8
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10015          EndpointRefAddress: cf0280d6-4909-4671-8972-90adb3e60181
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10016          EndpointRefAddress: da1cd97d-d18a-40fd-a250-1f0c4f13f293
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10017          EndpointRefAddress: 8af3658d-adf9-4898-8146-86a0cb50c9fb
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10018          EndpointRefAddress: 436eb565-ae32-4d79-a9a3-6d520af2c647
Onvif WS-Discovery: Find Xaddr: 10.0.0.147:10019          EndpointRefAddress: 6e17721c-d861-4609-880b-efd72e00b8bc
level=INFO ts=2022-05-17T17:17:03.069067014Z app=device-onvif-camera source=driver.go:422 msg="Discovered 20 device(s) in 1.133374208s via netscan."
```
Then user can follow [this doc to add a provision watcher](./auto-discovery.md) to add the discovered devices to EdgeX.
