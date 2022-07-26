
<a name="Onvif Camera Device Service (found in device-onvif-camera) Changelog"></a>
## EdgeX ONVIF Camera Device Service
[Github repository](https://github.com/edgexfoundry/device-onvif-camera)

### Change Logs for EdgeX Dependencies
- [device-sdk-go](https://github.com/edgexfoundry/device-sdk-go/blob/main/CHANGELOG.md)
- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/CHANGELOG.md)
- [go-mod-bootstrap](https://github.com/edgexfoundry/go-mod-bootstrap/blob/main/CHANGELOG.md)
- [go-mod-messaging](https://github.com/edgexfoundry/go-mod-messaging/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-registry](https://github.com/edgexfoundry/go-mod-registry/blob/main/CHANGELOG.md)  (indirect dependency)
- [go-mod-secrets](https://github.com/edgexfoundry/go-mod-secrets/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-configuration](https://github.com/edgexfoundry/go-mod-configuration/blob/main/CHANGELOG.md) (indirect dependency)

## [v2.2.0] Kamakura - 2022-07-26

This is the initial release of this ONVIF camera device service. Refer to the [README](https://github.com/edgexfoundry/device-onvif-camera/blob/v2.2.0/README.md) for details about this service.

### Known Issues 

The following issues are known at the time of this initial release and will be addressed in a future release:

-  [#121](https://github.com/edgexfoundry/device-onvif-camera/issues/121) Delay in obtaining DeviceStatus for pre-defined and manually added devices
