# Control Plane Events

## Introduction
Control plane events have been added to enable the device service to emit events onto the message bus when a device has been added, updated, or deleted. This is in lieu of official Control Plane Events (CPE) in an upcoming release of EdgeX. Once EdgeX officially supports CPE, these events will be dropped in favor of those.

See edgex-docs [issue #581](https://github.com/edgexfoundry/edgex-docs/issues/581)

## Sample Events
The following are sample events emitted when a camera is added, updated or deleted from the device service.

### CameraAdded
The following is a sample event emitted when a camera is added to the device service. The top level object is an EdgeX `Event`, which contains 1 or more `Readings`.
Where each `Reading` contains the following information:
- **`origin`**: The timestamp at which the event occurred (in nanoseconds since Epoch)
- **`resourceName`**: The type of control plane event (`CameraAdded`)
- **`objectValue`**: The entire EdgeX Device object
  - **`Name`**: The newly added device's name
  - **`ProfileName`**: Which EdgeX Profile the device was assigned
  - **`Protocols.Onvif`**: Onvif specific connection information

```json
{
   "apiVersion":"v2",
   "id":"2b20ddfb-d6b2-4e63-b7dd-0abef38524dd",
   "deviceName":"device-onvif-camera",
   "profileName":"device-onvif-camera",
   "sourceName":"CameraAdded",
   "origin":1650895116540236500,
   "readings":[
      {
         "id":"5bf9958b-c504-4803-824d-5796971a3128",
         "origin":1650895116540236500,
         "deviceName":"device-onvif-camera",
         "resourceName":"CameraAdded",
         "profileName":"device-onvif-camera",
         "valueType":"Object",
         "objectValue":{
            "AdminState":"UNLOCKED",
            "AutoEvents":[
               
            ],
            "Created":0,
            "Description":"Intel SimCamera Camera",
            "Id":"",
            "Labels":[
               "auto-discovery",
               "Intel",
               "SimCamera"
            ],
            "LastConnected":0,
            "LastReported":0,
            "Location":null,
            "Modified":0,
            "Name":"Intel-SimCamera-c45a57b3-6fcb-4c51-83de-54495e5454a6",
            "Notify":false,
            "OperatingState":"UP",
            "ProfileName":"onvif-camera",
            "Protocols":{
               "Onvif":{
                  "Address":"172.20.25.54",
                  "AuthMode":"usernametoken",
                  "FirmwareVersion":"2.4a",
                  "HardwareId":"1.0",
                  "Manufacturer":"Intel",
                  "Model":"SimCamera",
                  "Port":"10000",
                  "SecretPath":"credentials002",
                  "SerialNumber":"c45a57b3"
               }
            },
            "ServiceName":"device-onvif-camera"
         }
      }
   ]
}
```
### CameraUpdated
The following is a sample event emitted when a camera is updated to the device service. The top level object is an EdgeX `Event`, which contains 1 or more `Readings`.
Where each `Reading` contains the following information:
- **`origin`**: The timestamp at which the event occurred (in nanoseconds since Epoch)
- **`resourceName`**: The type of control plane event (`CameraUpdated`)
- **`objectValue`**: The entire EdgeX Device object
  - **`Name`**: The newly updated device's name
  - **`ProfileName`**: Which EdgeX Profile the device was assigned
  - **`Protocols.Onvif`**: Onvif specific connection information

```json
{
   "apiVersion":"v2",
   "id":"c267a358-8c61-45e7-807f-18f582104f70",
   "deviceName":"device-onvif-camera",
   "profileName":"device-onvif-camera",
   "sourceName":"CameraUpdated",
   "origin":1650895193968984300,
   "readings":[
      {
         "id":"6cfe28c9-f65e-409a-9b3c-d6d8ba9368db",
         "origin":1650895193968984300,
         "deviceName":"device-onvif-camera",
         "resourceName":"CameraUpdated",
         "profileName":"device-onvif-camera",
         "valueType":"Object",
         "objectValue":{
            "AdminState":"UNLOCKED",
            "AutoEvents":[
               
            ],
            "Created":0,
            "Description":"Intel SimCamera Camera",
            "Id":"",
            "Labels":[
               "auto-discovery",
               "Intel",
               "SimCamera2"
            ],
            "LastConnected":0,
            "LastReported":0,
            "Location":null,
            "Modified":0,
            "Name":"Intel-SimCamera-4e5a4e47-1d31-430c-97b4-0e144a705f95",
            "Notify":false,
            "OperatingState":"UP",
            "ProfileName":"onvif-camera",
            "Protocols":{
               "Onvif":{
                  "Address":"172.20.25.54",
                  "AuthMode":"usernametoken",
                  "FirmwareVersion":"2.4a",
                  "HardwareId":"1.0",
                  "Manufacturer":"Intel",
                  "Model":"SimCamera",
                  "Port":"10001",
                  "SecretPath":"credentials002",
                  "SerialNumber":"4e5a4e47"
               }
            },
            "ServiceName":"device-onvif-camera"
         }
      }
   ]
}
```
### CameraDeleted
The following is a sample event emitted when a camera is deleted from the device service. The top level object is an EdgeX `Event`, which contains 1 or more `Readings`.
Where each `Reading` contains the following information:
- **`origin`**: The timestamp at which the event occurred (in nanoseconds since Epoch)
- **`resourceName`**: The type of control plane event (`CameraDeleted`)
- **`ProfileName`**: Which EdgeX Profile the device was assigned
- **`value`**: The deleted device's name

```json
{
   "apiVersion":"v2",
   "id":"5c69bee2-3471-47b8-a5e7-ec22e11d81d7",
   "deviceName":"device-onvif-camera",
   "profileName":"device-onvif-camera",
   "sourceName":"CameraDeleted",
   "origin":1650895307691330000,
   "readings":[
      {
         "id":"b2a01f7c-02ea-48d2-99b4-9d15132c1080",
         "origin":1650895307691330000,
         "deviceName":"device-onvif-camera",
         "resourceName":"CameraDeleted",
         "profileName":"device-onvif-camera",
         "valueType":"String",
         "value":"Intel-SimCamera-4e5a4e47-1d31-430c-97b4-0e144a705f95"
      }
   ]
}
```

## Control Plane Profile

An EdgeX Device Profile has been added to define the control plane event schemas. This profile is located at [../cmd/res/profiles/control-plane.profile.yaml](../cmd/res/profiles/control-plane.profile.yaml).
