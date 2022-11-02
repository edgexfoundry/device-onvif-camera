
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

## [v2.3.0] Levski - 2022-11-09 (Not Compatible with 1.x releases)

### Features ✨

- Add Service Metrics configuration ([#187](https://github.com/edgexfoundry/device-onvif-camera/issues/187)) ([#c606ad2](https://github.com/edgexfoundry/device-onvif-camera/commits/c606ad2))
- Add NATS configuration and build option ([#164](https://github.com/edgexfoundry/device-onvif-camera/issues/164)) ([#0803938](https://github.com/edgexfoundry/device-onvif-camera/commits/0803938))
- Add commanding via message configuration ([#4fecb75](https://github.com/edgexfoundry/device-onvif-camera/commits/4fecb75))
- Add GetCapabilities and GetEndpointReference ([#189](https://github.com/edgexfoundry/device-onvif-camera/issues/189)) ([#8e26798](https://github.com/edgexfoundry/device-onvif-camera/commits/8e26798))
- **snap:** add config interface with unique identifier ([#179](https://github.com/edgexfoundry/device-onvif-camera/issues/179)) ([#b5d5716](https://github.com/edgexfoundry/device-onvif-camera/commits/b5d5716))

### Code Refactoring ♻

- Remove control-plane-device ([#153](https://github.com/edgexfoundry/device-onvif-camera/issues/153)) ([#93a2116](https://github.com/edgexfoundry/device-onvif-camera/commits/93a2116))
- Remove custom service interface and use sdk interface ([#150](https://github.com/edgexfoundry/device-onvif-camera/issues/150)) ([#87f15e2](https://github.com/edgexfoundry/device-onvif-camera/commits/87f15e2))

### Test

- Add nestscan discovery tests ([#152](https://github.com/edgexfoundry/device-onvif-camera/issues/152)) ([#d315776](https://github.com/edgexfoundry/device-onvif-camera/commits/d315776))
- Unit test additions ([#149](https://github.com/edgexfoundry/device-onvif-camera/issues/149)) ([#06d0f2a](https://github.com/edgexfoundry/device-onvif-camera/commits/06d0f2a))

### Bug Fixes 🐛

- Status delay [#121](https://github.com/edgexfoundry/device-onvif-camera/issues/121) ([#182](https://github.com/edgexfoundry/device-onvif-camera/issues/182)) ([#55762a2](https://github.com/edgexfoundry/device-onvif-camera/commits/55762a2))
- Discovery mode netscan when using docker-compose ([#176](https://github.com/edgexfoundry/device-onvif-camera/issues/176)) ([#6644798](https://github.com/edgexfoundry/device-onvif-camera/commits/6644798))
- Check whether the onvif response is valid ([#146](https://github.com/edgexfoundry/device-onvif-camera/issues/146)) ([#dc3ccaf](https://github.com/edgexfoundry/device-onvif-camera/commits/dc3ccaf))
- Netscan search only for NetworkVideoTransmitter ([#136](https://github.com/edgexfoundry/device-onvif-camera/issues/136)) ([#a7f4bbd](https://github.com/edgexfoundry/device-onvif-camera/commits/a7f4bbd))

### Documentation 📖

- Integrate onvif schemas and validation matrix to swagger ([#180](https://github.com/edgexfoundry/device-onvif-camera/issues/180)) ([#68a981f](https://github.com/edgexfoundry/device-onvif-camera/commits/68a981f))
- Updated date in onvif function info ([#177](https://github.com/edgexfoundry/device-onvif-camera/issues/177)) ([#c32d12c](https://github.com/edgexfoundry/device-onvif-camera/commits/c32d12c))
- Changes based on validation feeback ([#171](https://github.com/edgexfoundry/device-onvif-camera/issues/171)) ([#bca8455](https://github.com/edgexfoundry/device-onvif-camera/commits/bca8455))
- Add video reference and text ([#166](https://github.com/edgexfoundry/device-onvif-camera/issues/166)) ([#eff61ec](https://github.com/edgexfoundry/device-onvif-camera/commits/eff61ec))
- Add Go 1.18+ required for native install ([#60be3eb](https://github.com/edgexfoundry/device-onvif-camera/commits/60be3eb))
- Fix yaml tag for multicast ([#143](https://github.com/edgexfoundry/device-onvif-camera/issues/143)) ([#51d17d6](https://github.com/edgexfoundry/device-onvif-camera/commits/51d17d6))
- Add missing param to postman call ([#2f0c3ce](https://github.com/edgexfoundry/device-onvif-camera/commits/2f0c3ce))
- Add openapi spec from updated postman collection ([#123](https://github.com/edgexfoundry/device-onvif-camera/issues/123)) ([#9690a38](https://github.com/edgexfoundry/device-onvif-camera/commits/9690a38))
- Publish swagger files, update swagger readme ([#130](https://github.com/edgexfoundry/device-onvif-camera/issues/130)) ([#59a7980](https://github.com/edgexfoundry/device-onvif-camera/commits/59a7980))
- In-depth discovery and credential guides ([#126](https://github.com/edgexfoundry/device-onvif-camera/issues/126)) ([#803b45c](https://github.com/edgexfoundry/device-onvif-camera/commits/803b45c))

### Build 👷

- Upgrade to Go 1.18, Alpine 3.16, linter version and latest SDK/go-mod versions ([#127](https://github.com/edgexfoundry/device-onvif-camera/issues/127)) ([#616f9b7](https://github.com/edgexfoundry/device-onvif-camera/commits/616f9b7))

## [v2.2.0] Kamakura - 2022-07-26 (Not Compatible with 1.x releases)

This is the initial release of this ONVIF camera device service. Refer to the [README](https://github.com/edgexfoundry/device-onvif-camera/blob/v2.2.0/README.md) for details about this service.

### Known Issues 

The following issues are known at the time of this initial release and will be addressed in a future release:

-  [#121](https://github.com/edgexfoundry/device-onvif-camera/issues/121) Delay in obtaining DeviceStatus for pre-defined and manually added devices
