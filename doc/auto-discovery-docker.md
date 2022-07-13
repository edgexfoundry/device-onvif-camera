# Set Up Auto Discovery with Docker

This is a guide on how to configure the ONVIF device service for automatic device discovery using Docker.

<img alt="overview" src="images/auto-discovery-docker-overview.jpg" width="75%"/>

## Prepare edgex-compose/compose-builder
1. Download  or clone [edgex-compose](https://github.com/edgexfoundry/edgex-compose)
2. Change the directory to ` edgex-compose/compose-builder`

## Update the `./add-device-onvif-camera.yml`
Add subnets for discovering cameras:
```yaml
services:
  device-onvif-camera:
    ...
    environment:
      ...
      APPCUSTOM_DISCOVERYSUBNETS: "192.168.1.0/24,10.0.0.0/24"
```
See more settings at the [AppCustom section of the configuration.toml](https://github.com/edgexfoundry/device-onvif-camera/blob/main/cmd/res/configuration.toml)

## Non-Security Mode

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


## Secure Mode

### Deploy EdgeX services and device-onvif-camera
Deploy services with the following command:
```shell
make run ds-onvif-camera
```

### Add Secrets to Secret Store (Vault)

```shell
curl --request POST 'http://localhost:59984/api/v2/secret' \
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
        },
        {
            "key":"mode",
            "value":"usernametoken"
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
