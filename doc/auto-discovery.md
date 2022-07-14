# Auto Discovery
There are two methods that the device service can use to discover and add ONVIF compliant cameras using WS-Discovery, multicast and netscan.


## How does WS-Discovery work?

ONVIF devices support WS-Discovery, which is a mechanism that supports probing a network to find ONVIF capable devices.

Probe messages are sent over UDP to a standardized multicast address and UDP port number.

<img src="images/auto-discovery.jpg" width="75%"/>

WS-Discovery is generally faster than netscan becuase it only sends out one broadcast signal. However, it is normally limited by the network segmentation since the multicast packages typically do not traverse routers.

- Find the WS-Discovery programmer guide from https://www.onvif.org/profiles/whitepapers/
- Wiki page https://en.wikipedia.org/wiki/WS-Discovery

Example:
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
    > The Hello message from HIKVISION
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
    
    >The Hello message from Tapo C200
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
    


## How does netscan work?
An alternative method of discovery is a netscan, where the device service is provided a set of IP addresses on a network to scan for ONVIF protocol devices using unicast.

For example, if the provided CIDR is 10.0.0.0/24, it will probe the IP's subnet mask for ONVIF compliant devices using soap commands, directly connecting to each address. It then returns any devices it finds and adds them to the device service using the protocol information from the probes.

This method is going to be slower and more network-intensive than multicast WS-Discovery, becuase it has to make individual connections. However, it can reach a much wider set of networks and works better behind NATs (such as docker networks).


## Adding the Devices to EdgeX
```
┌─────────────────┐               ┌──────────────┐                  ┌─────────────────┐
│                 │               │              │                  │                 │
│                 │3.Create device│    Onvif     │1.Discover camera │                 │
│ metadata service◄───────────────┼─   Device  ──┼──────────────────┼──► Onvif Camera │
│                 │               │    Service   │2.Get device info │                 │
│                 │               │           ◄──┼──────────────────┼──               │
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
- The serviceName, profileName, adminState, autoEvents are defined by the provisionWatcher
- Predefine the secretPath for discovered device, for exmaple:
  - DefaultSecretPath="credentials001"
- GetDeviceInformation function provides Manufacturer, Model, FirmwareVersion, SerialNumber, HardwareId to protocol properties for provision watcher to filter


## Usage
## 1. Define driver config

1. Ensure that the cameras are all installed and configured before attempting discovery. 

2. Define the following configurations in `cmd/res/configuration.toml` for auto-discovery mechanism:


```toml
# Custom configs
[AppCustom]
CredentialsRetryTime = "120" # Seconds
CredentialsRetryWait = "1" # Seconds
RequestTimeout = "5" # Seconds
DiscoveryEthernetInterface = ""
DefaultSecretPath = "credentials001"
# BaseNotificationURL indicates the device service network location
BaseNotificationURL = "http://192.168.12.112:59984"

# Select which discovery mechanism(s) to use
DiscoveryMode = "both" # netscan, multicast, or both

# List of IPv4 subnets to perform netscan discovery on, in CIDR format (X.X.X.X/Y)
# separated by commas ex: "192.168.1.0/24,10.0.0.0/24"
DiscoverySubnets = ""

# Maximum simultaneous network probes
ProbeAsyncLimit = "4000"

# Maximum amount of milliseconds to wait for each IP probe before timing out.
# This will also be the minimum time the discovery process can take.
ProbeTimeoutMillis = "2000"

# Maximum amount of seconds the discovery process is allowed to run before it will be cancelled.
# It is especially important to have this configured in the case of larger subnets such as /16 and /8
MaxDiscoverDurationSeconds = "300"

EnableStatusCheck = true

# The interval in seconds at which the service will check the connection of all known cameras and update the device status 
# A longer interval will mean the service will detect changes in status less quickly
# Maximum 300s (1 hour)
CheckStatusInterval = 30

```
>Example of configuration.toml contents

### 2. Enable the Discovery Mechanism
Device discovery is triggered by the device SDK. Once the device service starts, it will discover the Onvif camera(s) at the specified interval.
>NOTE: you can also manually call discovery using this command: `curl -X POST http://<service-host>:59984/api/v2/discovery`

[Option 1] Enable from `configuration.toml`
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
The provision watcher is used to filter the discovered devices and provide information to create the devices.

#### Example - HIKVISION Onvif Camera Provision

Any discovered devices that match the `Manufacturer` and `Model` should be added to core metadata by the device service
```shell
curl --request POST 'http://localhost:59881/api/v2/provisionwatcher' \
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
             "profileName": "onvif-camera",
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
curl --request POST 'http://localhost:59881/api/v2/provisionwatcher' \
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
> `<password>` with the password, and <mode> with the auth mode.

```shell
curl --location --request POST 'http://localhost:59984/api/v2/secret' \
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
        },
        {
            "key":"mode",
            "value":"<mode>"
        }
    ]
}'
```
## Rediscovery

The device service is able to rediscover and update devices that have been discovered previously. Nothing additional is needed to enable this. It will run whenever the discover call is sent, regardless of whether it is a manual or automated call to discover.


## Device Status
The device status goes hand in hand with the rediscovery of the cameras, but goes beyond the scope of just discovery. It is a separate background task running at a specified interval (default 30s) to determine the most accurate operating status of the existing cameras.

### States and Descriptions
Currently there are 4 different statuses that a camera can have  

**UpWithAuth**: Can execute commands requiring credentials  
**UpWithoutAuth**: Can only execute commands that do not require credentials. Usually this means the camera's credentials have not been registered with the service yet, or have been changed.  
**Reachable**: Can be discovered but no commands can be recieved.  
**Unreachable**: Cannot be seen by service at all. Usually this means that there is a connection issue either physically or with the network.   

### Configuration
- Use `EnableStatusCheck` to enable the device status background service.
- `CheckStatusInterval` is the interval at which the service will determine the status of each camera.

```toml
EnableStatusCheck = true

# The interval in seconds at which the service will check the connection of all known cameras and update the device status 
# A longer interval will mean the service will detect changes in status less quickly
# Maximum 300s (1 hour)
CheckStatusInterval = 30
```
