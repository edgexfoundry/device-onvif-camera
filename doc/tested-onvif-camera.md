# Tested Onvif Cameras
The following table shows the Onvif functions tested for various Onvif cameras:

* '✔' means the function works for the specified camera.
* '❌' means the function does not work or is not implemented by the specified camera.

| Feature                                | Onvif Web Service | Onvif Function                      | Hikvision DFI6256TE | Tapo C200 | BOSCH DINION IP starlight 6000 HD | GeoVision GV-BX8700 |
|----------------------------------------|-------------------|-------------------------------------|---------------------|-----------|-----------------------------------|---------------------|
| **User Authentication**                | **Core**          | WS-UsernameToken                    | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | HTTP Digest                         | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   |                                     |                     |
| **Auto Discovery**                     | **Core**          | WS-Discovery                        | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        | **Device**        | GetDiscoveryMode                    | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | SetDiscoveryMode                    | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | GetScopes                           | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | SetScopes                           | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | AddScopes                           | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | RemoveScopes                        | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   |                                     |                     |
| **Network Configuration**              | **Device**        | GetHostname                         | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | SetHostname                         | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetDNS                              | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | SetDNS                              | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetNetworkInterfaces                | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | SetNetworkInterfaces                | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetNetworkProtocols                 | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | SetNetworkProtocols                 | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetNetworkDefaultGateway            | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | SetNetworkDefaultGateway            | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   |                                     |                     |
| **System Function**                    | **Device**        | GetDeviceInformation                | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | GetSystemDateAndTime                | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | SetSystemDateAndTime                | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | SetSystemFactoryDefault             | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | Reboot                              | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   |                                     |                     |
| **User Handling**                      | **Device**        | GetUsers                            | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | CreateUsers                         | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | DeleteUsers                         | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | SetUser                             | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   |                                     |                     |
| **Metadata Configuration**             | **Media**         | GetMetadataConfigurations           | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetMetadataConfiguration            | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetCompatibleMetadataConfigurations | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetMetadataConfigurationOptions     | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | AddMetadataConfiguration            | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | RemoveMetadataConfiguration         | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | SetMetadataConfiguration            | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   |                                     |                     |
| **Video Streaming**                    | **Media**         | GetProfiles                         | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | GetStreamUri                        | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        | **EdgeX**         | GetSnapshot                         | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   |                                     |                     |
| **VideoEncoder  Config**               | **Media**         | GetVideoEncoderConfiguration        | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   | SetVideoEncoderConfiguration        | ✔                   | ❌         | ✔                                 | ✔                   |
|                                        |                   | GetVideoEncoderConfigurationOptions | ✔                   | ✔         | ✔                                 | ✔                   |
|                                        |                   |                                     |                     |
| **PTZ Node**                           | **PTZ**           | GetNodes                            | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | GetNode                             | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   |                                     |                     |
| **PTZ Configuration**                  | **PTZ**           | GetConfigurations                   | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | GetConfiguration                    | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | GetConfigurationOptions             | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | SetConfiguration                    | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        | **Media**         | AddPTZConfiguration                 | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        | **Media**         | RemovePTZConfiguration              | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        |                   |                                     |                     |
| **PTZ Actuation**                      | **PTZ**           | AbsoluteMove                        | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | RelativeMove                        | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | ContinuousMove                      | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | Stop                                | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | GetStatus                           | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   |                                     |                     |
| **PTZ Preset**                         | **PTZ**           | SetPreset                           | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | GetPresets                          | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | GotoPreset                          | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   | RemovePreset                        | ❌                   | ✔         | ❌                                 | ❌                   |
|                                        |                   |                                     |                     |
| **PTZ Home Position**                  | **PTZ**           | GotoHomePosition                    | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        |                   | SetHomePosition                     | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        |                   |                                     |                     |
| **PTZ AuxiliaryOperations**            | **PTZ**           | SendAuxiliaryCommand                | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        |                   |                                     |                     |
| **Event Handling**                     | **Event**         | Notify                              | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | Subscribe                           | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | Renew                               | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | Unsubscribe                         | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | CreatePullPointSubscription         | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | PullMessages                        | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | TopicFilter                         | ✔                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | MessageContentFilter                | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        |                   |                                     |                     |
| **Configuration of Analytics profile** | **Media2**        | GetProfiles                         | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | GetAnalyticsConfigurations          | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | AddConfiguration                    | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | RemoveConfiguration                 | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   |                                     |                     | 
| **Analytics Module configuration**     | **Analytics**     | GetSupportedAnalyticsModules        | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | GetAnalyticsModules                 | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | CreateAnalyticsModules              | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        |                   | DeleteAnalyticsModules              | ❌                   | ❌         | ❌                                 | ❌                   |
|                                        |                   | GetAnalyticsModuleOptions           | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | ModifyAnalyticsModules              | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   |                                     |                     |
| **Rule configuration**                 | **Analytics**     | GetSupportedRules                   | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | GetRules                            | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | CreateRules                         | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | DeleteRules                         | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | GetRuleOptions                      | ❌                   | ❌         | ✔                                 | ❌                   |
|                                        |                   | ModifyRules                         | ❌                   | ❌         | ✔                                 | ❌                   |
