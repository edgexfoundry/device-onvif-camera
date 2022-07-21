# Get and Set Friendly Name and MAC Address

Friendly name and MAC address can be set and retrieved for each camera added to the service.


## Preset FriendlyName
`FriendlyName` is an element in the `Onvif ProtocolProperties` device field. It is initialized to be empty or `<Manufacturer+Model>`
if credentials are provided on discovery. The user can also pre-define this field in a camera.toml file.

If you add pre-defined devices, set up the `FriendlyName` field as shown in the
[camera.toml.example file](cmd/res/devices/camera.toml.example).

```toml
# Pre-defined Devices
[[DeviceList]]
Name = "Camera001"
ProfileName = "onvif-camera"
Description = "onvif conformant camera"
  [DeviceList.Protocols]
    [DeviceList.Protocols.Onvif]
    Address = "192.168.12.123"
    Port = "80"
    FriendlyName = "Home camera"
    [DeviceList.Protocols.CustomMetadata]
    Location = "Front door"
```

## Set Friendly Name

Use the FriendlyName device resource to set `FriendlyName` of a camera.

1. Use this command to set FriendlyName field.
```shell
curl --request PUT 'http://0.0.0.0:59882/api/v2/device/name/<device name>/FriendlyName' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "FriendlyName": {
            "FriendlyName":"Home camera"
        }
    }' | jq .
```
2. The response from the curl command.
```
{
    "apiVersion": "v2",
    "statusCode": 200
}
```
>Note: ensure all data is properly formatted json, and that all special characters are escaped if necessary


## Get Friendly Name

Use the FriendlyName device resource to retrieve `FriendlyName` of a camera.

1. Use this command to return FriendlyName field.

```shell
curl http://localhost:59882/api/v2/device/name/<device name>/FriendlyName | jq .
```
2. Response from the curl command. FriendlyName value can be found under `objectValue` in the response.
```shell
{
   "apiVersion" : "v2",
   "event" : {
      "apiVersion" : "v2",
      "deviceName" : "TP-Link-C200-3fa1fe68-b915-4053-a3e1-cc32e5000688",
      "id" : "a96e6fb1-d61c-49bf-b243-6e7b7db579d0",
      "origin" : 1657921895566531752,
      "profileName" : "onvif-camera",
      "readings" : [
         {
            "deviceName" : "TP-Link-C200-3fa1fe68-b915-4053-a3e1-cc32e5000688",
            "id" : "0cc6ae1f-1755-45ef-8a62-29bb69b3b821",
            "objectValue" : "Home camera",
            "origin" : 1657921895566531752,
            "profileName" : "onvif-camera",
            "resourceName" : "FriendlyName",
            "value" : "",
            "valueType" : "Object"
         }
      ],
      "sourceName" : "FriendlyName"
   },
   "statusCode" : 200
}
```
## Set MAC Address

Use the device resource MACAddress to set `MACAddress` of a camera.

1. Use this command to set MACAddress field.
```shell
curl --request PUT 'http://0.0.0.0:59882/api/v2/device/name/<device name>/MACAddress' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "MACAddress": {
            "MACAddress":"11:22:33:44:55:66"
        }
    }' | jq .
```
2. The response from the curl command.
```
{
    "apiVersion": "v2",
    "statusCode": 200
}
```
>Note: ensure all data is properly formatted json, and that all special characters are escaped if necessary.


## Get MAC Address

Use the MACAddress device resource to retrieve `MACAddress` of a camera.

1. Use this command to return MACAddress field.

```shell
curl http://localhost:59882/api/v2/device/name/<device name>/MACAddress | jq .
```
2. Response from the curl command. MACAddress value can be found under `objectValue` in the response.
```shell
{
   "apiVersion" : "v2",
   "event" : {
      "apiVersion" : "v2",
      "deviceName" : "TP-Link-C200-3fa1fe68-b915-4053-a3e1-1027f5ea8888",
      "id" : "433ba9bd-fe5a-4120-91b1-024a16005ab4",
      "origin" : 1658188028882145325,
      "profileName" : "onvif-camera",
      "readings" : [
         {
            "deviceName" : "TP-Link-C200-3fa1fe68-b915-4053-a3e1-1027f5ea8888",
            "id" : "ff3293ce-98b8-4d3f-9c12-dec7608ef2a8",
            "objectValue" : "11:22:33:44:55:66",
            "origin" : 1658188028882145325,
            "profileName" : "onvif-camera",
            "resourceName" : "MACAddress",
            "value" : "",
            "valueType" : "Object"
         }
      ],
      "sourceName" : "MACAddress"
   },
   "statusCode" : 200
}
```
