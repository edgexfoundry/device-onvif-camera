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
- **`profileName`**: Which EdgeX Profile the device was assigned
- **`value`**: The newly added device's name

```json
{
   "apiVersion":"v2",
   "id":"1a4302ee-1df5-4193-9e6c-1b0256874ecb",
   "deviceName":"device-onvif-camera",
   "profileName":"device-onvif-camera",
   "sourceName":"CameraAdded",
   "origin":1651092452115594800,
   "readings":[
      {
         "id":"4102f0d2-2b77-49ee-aa6e-ea127493695f",
         "origin":1651092452115594800,
         "deviceName":"device-onvif-camera",
         "resourceName":"CameraAdded",
         "profileName":"device-onvif-camera",
         "valueType":"String",
         "value":"Camera000"
      }
   ]
}
```
### CameraUpdated
The following is a sample event emitted when a camera is updated to the device service. The top level object is an EdgeX `Event`, which contains 1 or more `Readings`.
Where each `Reading` contains the following information:
- **`origin`**: The timestamp at which the event occurred (in nanoseconds since Epoch)
- **`resourceName`**: The type of control plane event (`CameraUpdated`)
- **`profileName`**: Which EdgeX Profile the device was assigned
- **`value`**: The updated device's name

```json
{
   "apiVersion":"v2",
   "id":"e71f71b5-3d1b-47c6-946c-7d012669637c",
   "deviceName":"device-onvif-camera",
   "profileName":"device-onvif-camera",
   "sourceName":"CameraUpdated",
   "origin":1651092492515690200,
   "readings":[
      {
         "id":"8b1fa0a6-e1e9-4165-988b-318eadd66831",
         "origin":1651092492515690200,
         "deviceName":"device-onvif-camera",
         "resourceName":"CameraUpdated",
         "profileName":"device-onvif-camera",
         "valueType":"String",
         "value":"Camera000"
      }
   ]
}
```
### CameraDeleted
The following is a sample event emitted when a camera is deleted from the device service. The top level object is an EdgeX `Event`, which contains 1 or more `Readings`.
Where each `Reading` contains the following information:
- **`origin`**: The timestamp at which the event occurred (in nanoseconds since Epoch)
- **`resourceName`**: The type of control plane event (`CameraDeleted`)
- **`profileName`**: Which EdgeX Profile the device was assigned
- **`value`**: The deleted device's name

```json
{
   "apiVersion":"v2",
   "id":"1876f1d8-c4fb-444d-a480-c410a6a38294",
   "deviceName":"device-onvif-camera",
   "profileName":"device-onvif-camera",
   "sourceName":"CameraDeleted",
   "origin":1651092502773005000,
   "readings":[
      {
         "id":"698db996-641b-4469-9b0b-f1f8c7700a42",
         "origin":1651092502773005000,
         "deviceName":"device-onvif-camera",
         "resourceName":"CameraDeleted",
         "profileName":"device-onvif-camera",
         "valueType":"String",
         "value":"Camera000"
      }
   ]
}
```

## Control Plane Profile

An EdgeX Device Profile has been added to define the control plane event schemas. This profile is located at [../cmd/res/profiles/control-plane.profile.yaml](../cmd/res/profiles/control-plane.profile.yaml).
