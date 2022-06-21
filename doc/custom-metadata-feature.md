# Custom Metadata

We have added a custom device resource for setting and querying custom metadata for each camera added to the service.

## Usage

- The *CustomMetadata* map is an element in the *ProtocolProperties* device field. It is initialized to be empty, so the user can add their desired fields.

### Preset Custom Metadata

If you add pre-defined devices, set up the CustomMetadata object as shown in the camera.toml.example file.

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
    # Assign AuthMode to "usernametoken" | "digest" | "both" | "none"
    AuthMode = "digest"
    SecretPath = "credentials001"
    [DeviceList.Protocols.CustomMetadata]
    CommonName = "Door camera"
    Location = "Front door"
```


### Set Custom Metadata

Use the CustomMetadata resource to set the fields of *CustomMetadata*. Choose the key/value pairs to represent your custom fields.

1. Use this command to put the data in the CustomMetadata field.
```shell
curl --request PUT 'http://0.0.0.0:59882/api/v2/device/name/<device name>/CustomMetadata' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "CustomMetadata": {
            "CommonName":"Front Door Camera",
            "Location":"Front Door",
            "Color":"Black and white",
            "Condition": "Good working condition"
        }
    }' | json_pp
```
2. The response from the curl command.
```
{
    "apiVersion": "v2",
    "statusCode": 200
}
```


### Get Custom Metadata

Use the CustomMetadata resource to get and display the fields of *CustomMetadata*.

1. Use this command to return all of the data in the CustomMetadata field.

```shell
curl http://localhost:59882/api/v2/device/name/<device name>/CustomMetadata | json_pp
```
2. The repsonse from the curl command.
```shell
{
   "apiVersion" : "v2",
   "event" : {
      "apiVersion" : "v2",
      "deviceName" : "3fa1fe68-b915-4053-a3e1-cc32e5000688",
      "id" : "ba3987f9-b45b-480a-b582-f5501d673c4d",
      "origin" : 1655409814077374935,
      "profileName" : "onvif-camera",
      "readings" : [
         {
            "deviceName" : "3fa1fe68-b915-4053-a3e1-cc32e5000688",
            "id" : "cf96e5c0-bde1-4c0b-9fa4-8f765c8be456",
            "objectValue" : {
               "Color" : "Black and white",
               "CommonName" : "Front Door Camera",
               "Condition" : "Good working condition",
               "Location" : "Front Door"
            },
            "origin" : 1655409814077374935,
            "profileName" : "onvif-camera",
            "resourceName" : "CustomMetadata",
            "value" : "",
            "valueType" : "Object"
         }
      ],
      "sourceName" : "CustomMetadata"
   },
   "statusCode" : 200
}
```



### Get Specific Custom Metadata

Pass the *CustomMetadata* resource a query to get specific field(s) in CustomMetadata. The query must be a base64 encoded json object with an array of fields you want to access.

1. Json object holding an array of fields you want to query.
```json
'{
    "CustomMetadata": 
        [
            "CommonName",
            "Location"
        ]
}'
```

2. Use this command to convert the json object to base64.
```shell
echo '{
    "CustomMetadata": 
        [
            "CommonName",
            "Location"
        ]
}' | base64
```

3. The response converted to base64.
```shell
ewogICAgIkN1c3RvbU1ldGFkYXRhIjogCiAgICAgICAgWwogICAgICAgICAgICAiQ29tbW9uTmFtZSIsCiAgICAgICAgICAgICJMb2NhdGlvbiIKICAgICAgICBdCn0K
```

4. Use this command to query the fields you provided in the json object.
```shell
curl http://localhost:59882/api/v2/device/name/3fa1fe68-b915-4053-a3e1-cc32e5000688/CustomMetadata?jsonObject=ewogICAgIkN1c3RvbU1ldGFkYXRhIjogCiAgICAgICAgWwogICAgICAgICAgICAiQ29tbW9uTmFtZSIsCiAgICAgICAgICAgICJMb2NhdGlvbiIKICAgICAgICBdCn0K | json_pp

```

5. Curl response. 
```shell
{
   "apiVersion" : "v2",
   "event" : {
      "apiVersion" : "v2",
      "deviceName" : "3fa1fe68-b915-4053-a3e1-cc32e5000688",
      "id" : "24c3eb0a-48b1-4afe-b874-965aeb2e42a2",
      "origin" : 1655410556448058195,
      "profileName" : "onvif-camera",
      "readings" : [
         {
            "deviceName" : "3fa1fe68-b915-4053-a3e1-cc32e5000688",
            "id" : "d0c26303-20b5-4ccd-9e63-fb02b87b8ebc",
            "objectValue" : {
               "CommonName" : "Front Door Camera",
               "Location" : "Front Door"
            },
            "origin" : 1655410556448058195,
            "profileName" : "onvif-camera",
            "resourceName" : "CustomMetadata",
            "value" : "",
            "valueType" : "Object"
         }
      ],
      "sourceName" : "CustomMetadata"
   },
   "statusCode" : 200
}
```

### Additional Usage

Setting a field to "deletes" will delete that field.

1. Use this command to delete fields.
```shell
curl --request PUT 'http://0.0.0.0:59882/api/v2/device/name/<device name>/CustomMetadata' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "CustomMetadata": {
            "Color":"delete",
            "Condition": "delete"
        }
    }' | json_pp
```
2. The response from the curl command.
```
{
    "apiVersion": "v2",
    "statusCode": 200
}
```
