# Custom Metadata

We have added a custom device resource for setting and querying custom metadata for each camera added to the service.

## Usage

- The *CustomMetadata* map is an element in the *ProtocolProperties* device field. It is initialized to be empty, so the user can add their desired fields.

### Set Custom Metadata

Use the CustomMetadata resource to set the fields of *CustomMetadata*. Choose the key/value pairs to represent your custom fields.

```shell
curl -X POST http://localhost:59882/api/v2/device/name/<device name>/CustomMetadata /
-H 'Content-Type: application/json'/
-d '{
        "CustomMetadata: {
            "CommonName":"Front Door Camera",
            "Location":"Front Door"
        }
    }
```
>Example of SetCustomMetadata command.

### Get Custom Metadata

Use the CustomMetadata resource to get and display the fields of *CustomMetadata*.

```shell
curl http://localhost://59882/api/v2/device/name/<device name>/CustomMetadata
```
