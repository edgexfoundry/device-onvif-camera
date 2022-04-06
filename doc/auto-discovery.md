# Auto Discovery

## How does the WS-Discovery work?

ONVIF devices support WS-Discovery, which is a mechanism that supports probing a network to find ONVIF capable devices.

Probe messages are sent over UDP to a standardized multicast address and UDP port number.

<img src="images/auto-discovery.jpg" width="75%"/>

WS-Discovery is normally limited by the network segmentation since the multicast packages typically do not traverse routers.

- Find the WS-Discovery programmer guide from https://www.onvif.org/profiles/whitepapers/
- Wiki page https://en.wikipedia.org/wiki/WS-Discovery

For example:
1. The client sends Probe message to find Onvif camera on the network.
    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <soap-env:Envelope
            xmlns:soap-env="http://www.w3.org/2003/05/soap-envelope"
            xmlns:soap-enc="http://www.w3.org/2003/05/soap-encoding"
            xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
        <soap-env:Header>
            <a:Action mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
            <a:MessageID>uuid:a86f9421-b764-4256-8762-5ed0d8602a9c</a:MessageID>
            <a:ReplyTo>
                <a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
            </a:ReplyTo>
            <a:To mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
        </soap-env:Header>
        <soap-env:Body>
            <Probe
                    xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery"/>
        </soap-env:Body>
    </soap-env:Envelope>
    ```

2. The Onvif camera responds the Hello message according to the Probe message
   - The Hello message from HIKVISION:
    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <env:Envelope
        xmlns:env="http://www.w3.org/2003/05/soap-envelope"
        ...>
        <env:Header>
            <wsadis:MessageID>urn:uuid:cea94000-fb96-11b3-8260-686dbc5cb15d</wsadis:MessageID>
            <wsadis:RelatesTo>uuid:a86f9421-b764-4256-8762-5ed0d8602a9c</wsadis:RelatesTo>
            <wsadis:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsadis:To>
            <wsadis:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/ProbeMatches</wsadis:Action>
            <d:AppSequence InstanceId="1637072188" MessageNumber="17"/>
        </env:Header>
        <env:Body>
            <d:ProbeMatches>
                <d:ProbeMatch>
                    <wsadis:EndpointReference>
                        <wsadis:Address>urn:uuid:cea94000-fb96-11b3-8260-686dbc5cb15d</wsadis:Address>
                    </wsadis:EndpointReference>
                    <d:Types>dn:NetworkVideoTransmitter tds:Device</d:Types>
                    <d:Scopes>onvif://www.onvif.org/type/video_encoder onvif://www.onvif.org/Profile/Streaming onvif://www.onvif.org/MAC/68:6d:bc:5c:b1:5d onvif://www.onvif.org/hardware/DFI6256TE http:123</d:Scopes>
                    <d:XAddrs>http://192.168.12.123/onvif/device_service</d:XAddrs>
                    <d:MetadataVersion>10</d:MetadataVersion>
                </d:ProbeMatch>
            </d:ProbeMatches>
        </env:Body>
    </env:Envelope>
    ```
   - The Hello message from Tapo C200:
    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <SOAP-ENV:Envelope
        xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
        ...>
        <SOAP-ENV:Header>
            <wsa:MessageID>uuid:a86f9421-b764-4256-8762-5ed0d8602a9c</wsa:MessageID>
            <wsa:RelatesTo>uuid:a86f9421-b764-4256-8762-5ed0d8602a9c</wsa:RelatesTo>
            <wsa:ReplyTo SOAP-ENV:mustUnderstand="true">
                <wsa:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:Address>
            </wsa:ReplyTo>
            <wsa:To SOAP-ENV:mustUnderstand="true">urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
            <wsa:Action SOAP-ENV:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2005/04/discovery/ProbeMatches</wsa:Action>
        </SOAP-ENV:Header>
        <SOAP-ENV:Body>
            <wsdd:ProbeMatches>
                <wsdd:ProbeMatch>
                    <wsa:EndpointReference>
                        <wsa:Address>uuid:3fa1fe68-b915-4053-a3e1-c006c3afec0e</wsa:Address>
                        <wsa:ReferenceProperties></wsa:ReferenceProperties>
                        <wsa:PortType>ttl</wsa:PortType>
                    </wsa:EndpointReference>
                    <wsdd:Types>tdn:NetworkVideoTransmitter</wsdd:Types>
                    <wsdd:Scopes>onvif://www.onvif.org/name/TP-IPC onvif://www.onvif.org/hardware/MODEL onvif://www.onvif.org/Profile/Streaming onvif://www.onvif.org/location/ShenZhen onvif://www.onvif.org/type/NetworkVideoTransmitter </wsdd:Scopes>
                    <wsdd:XAddrs>http://192.168.12.128:2020/onvif/device_service</wsdd:XAddrs>
                    <wsdd:MetadataVersion>1</wsdd:MetadataVersion>
                </wsdd:ProbeMatch>
            </wsdd:ProbeMatches>
        </SOAP-ENV:Body>
    </SOAP-ENV:Envelope>
    ```

## EdgeX integrates the WS-Discovery
```
┌─────────────────┐               ┌──────────────┐                  ┌─────────────────┐
│                 │               │              │                  │                 │
│                 │3.Create device│    Onvif     │1.Discover camera │                 │
│ metadata service◄───────────────┼─   Device  ──┼──────────────────┼──► Onvif Camera │
│                 │               │    Service   │2.Get device info │                 │
│                 │               │              │                  │                 │
└─────────────────┘               └──────────────┘                  └─────────────────┘
```
1. Discover camera via WS-Discovery
2. Get device information via SOAP action
   ```xml
    <tds:GetDeviceInformationResponse>
        <tds:Manufacturer>HIKVISION</tds:Manufacturer>
        <tds:Model>DFI6256TE</tds:Model>
        <tds:FirmwareVersion>V5.5.80 build 190528</tds:FirmwareVersion>
        <tds:SerialNumber>DFI6256TE20190608AAWRD26707311</tds:SerialNumber>
        <tds:HardwareId>88</tds:HardwareId>
    </tds:GetDeviceInformationResponse>
   ```
3. Create device to metadata service
   ```json
    {
      "apiVersion": "v2",
      "device": {
        "name":"HIKVISION-DFI6256TE-cea94000-fb96-11b3-8260-686dbc5cb15d",
        "serviceName": "device-onvif-camera",
        "profileName": "onvif-camera",
        "description": "HIKVISION camera",
        "protocols": {
            "Onvif": {
                "Address": "192.168.12.123",
                "Port": "80",
                "AuthMode": "usernametoken",
                "SecretPath": "credentials001",
                "Manufacturer": "HIKVISION",
                "Model": "DFI6256TE",
                "FirmwareVersion": "V5.5.80 build 190528",
                "SerialNumber": "DFI6256TE20190608AAWRD26707311",
                "HardwareId": "88",
            }
         }
       }
    }
    ```

- Device Name:  Manufacturer-Model-UUID (The UUID extracted from the probe response's EndpointReference address)
- The serviceName, profileName, adminState, autoEvets are defined by the provisionWatcher
- Predefine the driver config authMode and secretPath for discovered device, for exmaple:
  - DefaultAuthMode="usernametoken"
  - DefaultSecretPath="credentials001"
- GetDeviceInformation function provides Manufacturer, Model, FirmwareVersion, SerialNumber, HardwareId to protocol properties for provision watcher to filter


## Usage
## 1. Define driver config

The device service expects the Onvif camera should be installed and configured, and we can discover the camera from the **specified network** and **get the required device information**.

To achieve that, we need to define the following config for auto-discovery mechanism:
* **DiscoveryEthernetInterface** - Specify the target EthernetInterface for discovering. The default value is `en0`, the user can modify it to meet their requirement.
* **DefaultAuthMode** - Specify the default AuthMode for discovered devices.
* **DefaultSecretPath** - Specify the default SecretPath for discovered devices.

For example:
```yaml
[Driver]
DiscoveryEthernetInterface = "en0"
DefaultAuthMode = "usernametoken"
DefaultSecretPath = "credentials001"
```


### 2. Enable the Discovery Mechanism
The Discovery is triggered by device SDK. Once the device service startup, the device service will discover the Onvif camera with the specified interval.

[Option 1] Enable from the configuration.toml
```yaml
[Device] 
...
    [Device.Discovery]
    Enabled = true
    Interval = "30s"
```

[Option 2] Enable from the env
```shell
export DEVICE_DISCOVERY_ENABLED=true
export DEVICE_DISCOVERY_INTERVAL=30s
```

### 3. Add Provision Watcher
The Provision Watcher is used to filter the discovered devices and provide information to create the devices.

#### Example - HIKVISION Onvif Camera Provision

Any discovered devices that match the Manufacturer and Model should be added to core metadata by device service
```shell
curl --request POST 'http://0.0.0.0:59881/api/v2/provisionwatcher' \
    --header 'Content-Type: application/json' \
    -d '[
       {
          "provisionwatcher":{
             "apiVersion":"v2",
             "name":"Test-Provision-Watcher",
             "adminState":"UNLOCKED",
             "identifiers":{
                "Manufacturer": "HIKVISION",
                "Model": "DFI6256TE"
             },
             "serviceName": "device-onvif-camera",
             "profileName": "hikvision-profile",
             "autoEvents": [
                 { "interval": "15s", "sourceName": "Users" }
              ]
          },
          "apiVersion":"v2"
       }
    ]'
```

#### Example - Unknown Onvif Camera Provision
Add any unknown discovered devices to core metadata with a generic profile.

```shell
curl --request POST 'http://0.0.0.0:59881/api/v2/provisionwatcher' \
    --header 'Content-Type: application/json' \
    -d '[
       {
          "provisionwatcher":{
             "apiVersion":"v2",
             "name":"Test-Provision-Watcher-Unknown",
             "adminState":"UNLOCKED",
             "identifiers":{
                "Address": "."
             },
             "blockingIdentifiers":{
                "Manufacturer": [ "HIKVISION" ]
             },
             "serviceName": "device-onvif-camera",
             "profileName": "onvif-camera"
          },
          "apiVersion":"v2"
       }
    ]'
```

### 4. Add Credentials to Unknown Camera
If a camera is discovered in which the credentials are unknown, it will be
added as a generic onvif camera, and will require the user to set the credentials
in order to call most ONVIF commands.

#### Non-Secure Mode
##### Helper Script
Run the [bin/set-credentials.sh](../bin/set-credentials.sh) script
```shell
# Usage: bin/set-credentials.sh [-s/--secure-mode] [-d <device_name>] [-u <username>] [-p <password>]
bin/set-credentials.sh

# Select which camera by device-name (uuid)
# Enter username when prompted
# Enter password when prompted

```

***

##### Manual
> **Note:** Replace `<device-name>` with the device name of the
> camera you want to set credentials for, `<username>` with the username, and
> `<password>` with the password.

Set Path to `<device-name>`
```shell
curl -X PUT --data "<device-name>" \
    "http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/<device-name>/Path"
```

Set username to `<username>`
```shell
curl -X PUT --data "<username>" \
    "http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/<device-name>/Secrets/username"
```

Set password to `<password>`
```shell
curl -X PUT --data "<password>" \
    "http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/<device-name>/Secrets/password"
```

***

#### Secure Mode
##### Helper Script
Run the [bin/set-credentials.sh](../bin/set-credentials.sh) script with `--secure-mode` flag
```shell
# Usage: bin/set-credentials.sh [-s/--secure-mode] [-d <device_name>] [-u <username>] [-p <password>]
bin/set-credentials.sh --secure-mode

# Select which camera by device-name (uuid)
# Enter username when prompted
# Enter password when prompted

```

***

##### Manual
Credentials can be added via EdgeX Secrets:

> **Note:** Replace `<device-name>` with the device name of the
> camera you want to set credentials for, `<username>` with the username, and
> `<password>` with the password.

```shell
curl --location --request POST 'http://0.0.0.0:59984/api/v2/secret' \
    --header 'Content-Type: application/json' \
    --data-raw '
{
    "apiVersion":"v2",
    "path": "<device-name>",
    "secretData":[
        {
            "key":"username",
            "value":"<username>"
        },
        {
            "key":"password",
            "value":"<password>"
        }
    ]
}'
```
