# Set Up Auto Discovery with Docker

Since auto-discovery mechanism uses multicast UDP to find the available camera, so the device-onvif-camera needs to send out the probe message for searching cameras.

But we can’t send the multicast message from the **edgex network** to the **host network** in the **docker container**. 

The workaround is running the **dockerized device-onvif-camera** on the **host network**.

<img alt="overview" src="images/auto-discovery-docker-overview.jpg" width="75%"/>

Note: For macOS, the network_mode: "host" probably not working as expected: https://github.com/docker/for-mac/issues/1031

## None Security Mode

### Prepare edgex-compose/compose-builder

#### 1. Download the [edgex-compose](https://github.com/edgexfoundry/edgex-compose) and setup it according to the [docker-compose setup guide](./docker-compose/README.md)

#### 2. Replace the `add-device-onvif-camera.yml` with the following content:
```
services:
  device-onvif-camera:
    image: edgex/device-onvif-camera${ARCH}:${DEVICE_ONVIFCAM_VERSION}
    user: "${EDGEX_USER}:${EDGEX_GROUP}"
    container_name: edgex-device-onvif-camera
    hostname: edgex-device-onvif-camera
    read_only: true
    restart: always
    network_mode: "host"
    environment:
      SERVICE_HOST: 192.168.56.101
      EDGEX_SECURITY_SECRET_STORE: "false"
      DEVICE_DISCOVERY_ENABLED: "true"
      DRIVER_DISCOVERYETHERNETINTERFACE: enp0s3
      DRIVER_DEFAULTSECRETPATH: credentials001
      
      WRITABLE_LOGLEVEL: DEBUG
    depends_on:
      - consul
      - data
      - metadata
    security_opt:
      - no-new-privileges:true
    command: --registry --confdir=/res
```

**Note**: The user should replace the host IP to match their own machine IP

- Remove the useless port mapping when using the host network
- Replace **networks** with **network_mode: "host"**
- Remove `env_file` because we don't use the env like `CLIENTS_CORE_DATA_HOST=edgex-core-data`
- Modify SERVICE_HOST env to match the machine IP
- Add EDGEX_SECURITY_SECRET_STORE env with "false" value
- Enable auto-discovery by `DEVICE_DISCOVERY_ENABLED` with "true" value
- Use `DRIVER_DISCOVERYETHERNETINTERFACE` to specify the ethernet interface for discovering
- Use `DRIVER_DEFAULTSECRETPATH` to specify the default secret path
- Use `WRITABLE_LOGLEVEL` to specify the log level for debugging
- Add command to override the CMD because we don't use the configuration provider from Consul


### Deploy EdgeX services and device-onvif-camera
Deploy services with the following command:
```shell
make run no-secty ds-onvif-camera
```

### Inspect the device-onvif-camera
The user can docker logs to trace the auto-discovery
```shell
$ docker logs edgex-device-onvif-camera -f --tail 10
...
level=DEBUG ts=2021-12-17T08:26:38.686619358Z app=device-onvif-camera source=discovery.go:35 msg="protocol discovery triggered"
Onvif WS-Discovery: Find 192.168.56.101:10000 
Onvif WS-Discovery: Find 192.168.56.101:10001 
level=DEBUG ts=2021-12-17T08:26:39.726266165Z app=device-onvif-camera source=onvifclient.go:225 msg="SOAP Request: <tds:GetDeviceInformation></tds:GetDeviceInformation>"
level=DEBUG ts=2021-12-17T08:26:39.748227111Z app=device-onvif-camera source=onvifclient.go:243 msg="SOAP Response: <GetDeviceInformationResponse><Manufacturer>Happytimesoft</Manufacturer><Model>IPCamera</Model><FirmwareVersion>2.4</FirmwareVersion><SerialNumber>123456</SerialNumber><HardwareId>1.0</HardwareId></GetDeviceInformationResponse>"
level=DEBUG ts=2021-12-17T08:26:39.748270564Z app=device-onvif-camera source=driver.go:333 msg="Discovered camera from the address '192.168.56.101:10000'"
level=DEBUG ts=2021-12-17T08:26:39.761718293Z app=device-onvif-camera source=onvifclient.go:225 msg="SOAP Request: <tds:GetDeviceInformation></tds:GetDeviceInformation>"
level=DEBUG ts=2021-12-17T08:26:39.782834447Z app=device-onvif-camera source=onvifclient.go:243 msg="SOAP Response: <GetDeviceInformationResponse><Manufacturer>Happytimesoft</Manufacturer><Model>IPCamera</Model><FirmwareVersion>2.4</FirmwareVersion><SerialNumber>123456</SerialNumber><HardwareId>1.0</HardwareId></GetDeviceInformationResponse>"
level=DEBUG ts=2021-12-17T08:26:39.782871465Z app=device-onvif-camera source=driver.go:333 msg="Discovered camera from the address '192.168.56.101:10001'"
level=DEBUG ts=2021-12-17T08:26:39.782886193Z app=device-onvif-camera source=async.go:127 msg="Filtered device addition finished"
```
Then user can follow [the doc to add Provision Watcher](./auto-discovery.md).


## Security Mode

### Prepare edgex-compose/compose-builder

#### 1. Download the [edgex-compose](https://github.com/edgexfoundry/edgex-compose) and setup it according to the [docker-compose setup guide](./docker-compose/README.md)

#### 2. Replace the `add-device-onvif-camera.yml` with the following content:
```yaml
services:
  device-onvif-camera:
    image: edgex/device-onvif-camera${ARCH}:${DEVICE_ONVIFCAM_VERSION}
    user: "${EDGEX_USER}:${EDGEX_GROUP}"
    container_name: edgex-device-onvif-camera
    hostname: edgex-device-onvif-camera
    read_only: true
    restart: always
    network_mode: host
    environment:
      SERVICE_HOST: 192.168.56.101
      EDGEX_SECURITY_SECRET_STORE: "true"
      DEVICE_DISCOVERY_ENABLED: "true"
      DRIVER_DISCOVERYETHERNETINTERFACE: enp0s3
      DRIVER_DEFAULTSECRETPATH: credentials001
      
      SECRETSTORE_HOST: localhost
      STAGEGATE_BOOTSTRAPPER_HOST: localhost
      STAGEGATE_READY_TORUNPORT: 54329
      STAGEGATE_WAITFOR_TIMEOUT: 60s

      WRITABLE_LOGLEVEL: DEBUG
    depends_on:
      - consul
      - data
      - metadata
    security_opt:
      - no-new-privileges:true
    command: /device-onvif-camera --registry --confdir=/res
```

**Note**: The user should replace the host IP to match their own machine IP

- Remove the useless port mapping when using the host network
- Replace **networks** with **network_mode: "host"**
- Remove `env_file` because we don't use the env like `CLIENTS_CORE_DATA_HOST=edgex-core-data`
- Modify `SERVICE_HOST` env to match the machine IP
- Add `EDGEX_SECURITY_SECRET_STORE` env with "true" value
- Enable auto-discovery by `DEVICE_DISCOVERY_ENABLED` with "true" value
- Use `DRIVER_DISCOVERYETHERNETINTERFACE` to specify the ethernet interface for discovering
- Use `DRIVER_DEFAULTSECRETPATH` to specify the default secret path
- Use `SECRETSTORE_HOST` to specify the Vault's host
- Use `STAGEGATE_BOOTSTRAPPER_HOST`, `STAGEGATE_READY_TORUNPORT`, `STAGEGATE_WAITFOR_TIMEOUT` to specify the bootstrapper settings for waiting the security set up
- Use `WRITABLE_LOGLEVEL` to specify the log level for debugging
- Add command to override the CMD because we don't use the configuration provider from Consul

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
curl --request POST 'http://192.168.56.101:59985/api/v2/secret' \
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
level=DEBUG ts=2021-12-17T08:26:38.686619358Z app=device-onvif-camera source=discovery.go:35 msg="protocol discovery triggered"
Onvif WS-Discovery: Find 192.168.56.101:10000 
Onvif WS-Discovery: Find 192.168.56.101:10001 
level=DEBUG ts=2021-12-17T08:26:39.726266165Z app=device-onvif-camera source=onvifclient.go:225 msg="SOAP Request: <tds:GetDeviceInformation></tds:GetDeviceInformation>"
level=DEBUG ts=2021-12-17T08:26:39.748227111Z app=device-onvif-camera source=onvifclient.go:243 msg="SOAP Response: <GetDeviceInformationResponse><Manufacturer>Happytimesoft</Manufacturer><Model>IPCamera</Model><FirmwareVersion>2.4</FirmwareVersion><SerialNumber>123456</SerialNumber><HardwareId>1.0</HardwareId></GetDeviceInformationResponse>"
level=DEBUG ts=2021-12-17T08:26:39.748270564Z app=device-onvif-camera source=driver.go:333 msg="Discovered camera from the address '192.168.56.101:10000'"
level=DEBUG ts=2021-12-17T08:26:39.761718293Z app=device-onvif-camera source=onvifclient.go:225 msg="SOAP Request: <tds:GetDeviceInformation></tds:GetDeviceInformation>"
level=DEBUG ts=2021-12-17T08:26:39.782834447Z app=device-onvif-camera source=onvifclient.go:243 msg="SOAP Response: <GetDeviceInformationResponse><Manufacturer>Happytimesoft</Manufacturer><Model>IPCamera</Model><FirmwareVersion>2.4</FirmwareVersion><SerialNumber>123456</SerialNumber><HardwareId>1.0</HardwareId></GetDeviceInformationResponse>"
level=DEBUG ts=2021-12-17T08:26:39.782871465Z app=device-onvif-camera source=driver.go:333 msg="Discovered camera from the address '192.168.56.101:10001'"
level=DEBUG ts=2021-12-17T08:26:39.782886193Z app=device-onvif-camera source=async.go:127 msg="Filtered device addition finished"
```
Then user can follow [the doc to add Provision Watcher](./auto-discovery.md).
