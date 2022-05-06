# Tested Onvif Camera
The following table shows the tested Onvif cameras with Onvif functions:

* 'O' means the function works for the camera.
* 'X' means the function not work for the camera. The camera might not perform the request or return empty response.

| Feature                            | Onvif Web Service | Onvif Function                      | Hikvision DFI6256TE | Tapo C200 | BOSCH DINION IP starlight 6000 HD | GeoVision GV-BX8700 |
|------------------------------------|-------------------|-------------------------------------|---------------------|-----------|-----------------------------------|---------------------|
| User Authentication                | Core              | **WS-Usernametoken Authentication** | O                   | O         | O                                 | O                   |
|                                    |                   | **HTTP Digest**                     | O                   | X         | O                                 | X                   |
| Auto Discovery                     | Core              | **WS-Discovery**                    | O                   | O         | O                                 | O                   |
|                                    | Device            | GetDiscoveryMode                    | O                   | O         | O                                 | O                   |
|                                    |                   | SetDiscoveryMode                    | O                   | O         | O                                 | O                   |
|                                    |                   | GetScopes                           | O                   | O         | O                                 | O                   |
|                                    |                   | SetScopes                           | O                   | O         | O                                 | O                   |
|                                    |                   | AddScopes                           | O                   | X         | O                                 | O                   |
|                                    |                   | RemoveScopes                        | O                   | X         | O                                 | O                   |
| Network Configuration              | Device            | GetHostname                         | O                   | O         | O                                 | O                   |
|                                    |                   | SetHostname                         | O                   | X         | O                                 | O                   |
|                                    |                   | GetDNS                              | O                   | X         | O                                 | O                   |
|                                    |                   | SetDNS                              | O                   | X         | O                                 | O                   |
|                                    |                   | **GetNetworkInterfaces**            | O                   | O         | O                                 | O                   |
|                                    |                   | **SetNetworkInterfaces**            | O                   | X         | O                                 | O                   |
|                                    |                   | GetNetworkProtocols                 | O                   | O         | O                                 | O                   |
|                                    |                   | SetNetworkProtocols                 | O                   | X         | O                                 | O                   |
|                                    |                   | **GetNetworkDefaultGateway**        | O                   | X         | O                                 | O                   |
|                                    |                   | **SetNetworkDefaultGateway**        | O                   | X         | O                                 | O                   |
| System Function                    | Device            | **GetDeviceInformation**            | O                   | O         | O                                 | O                   |
|                                    |                   | GetSystemDateAndTime                | O                   | O         | O                                 | O                   |
|                                    |                   | SetSystemDateAndTime                | O                   | X         | O                                 | O                   |
|                                    |                   | SetSystemFactoryDefault             | O                   | O         | O                                 | O                   |
|                                    |                   | Reboot                              | O                   | O         | O                                 | O                   |
| User Handling                      | Device            | **GetUsers**                        | O                   | X         | O                                 | O                   |
|                                    |                   | **CreateUsers**                     | O                   | X         | O                                 | O                   |
|                                    |                   | **DeleteUsers**                     | O                   | X         | O                                 | O                   |
|                                    |                   | **SetUser**                         | O                   | X         | O                                 | O                   |
| Metadata Configuration             | Media             | GetMetadataConfigurations           | O                   | X         | O                                 | O                   |
|                                    |                   | GetMetadataConfiguration            | O                   | X         | O                                 | O                   |
|                                    |                   | GetCompatibleMetadataConfigurations | O                   | X         | O                                 | O                   |
|                                    |                   | **GetMetadataConfigurationOptions** | O                   | X         | O                                 | O                   |
|                                    |                   | AddMetadataConfiguration            | O                   | X         | O                                 | O                   |
|                                    |                   | RemoveMetadataConfiguration         | O                   | X         | O                                 | O                   |
|                                    |                   | **SetMetadataConfiguration**        | O                   | X         | O                                 | O                   |
| Video Streaming                    | Media             | **GetProfiles**                     | O                   | O         | O                                 | O                   |
|                                    |                   | **GetStreamUri**                    | O                   | O         | O                                 | O                   |
|                                    | EdgeX             | **GetSnapshot**                     | O                   | X         | O                                 | X                   |
| VideoEncoder  Config               | Media             | GetVideoEncoderConfiguration        | O                   | O         | O                                 | O                   |
|                                    |                   | **SetVideoEncoderConfiguration**    | O                   | X         | O                                 | O                   |
|                                    |                   | GetVideoEncoderConfigurationOptions | O                   | O         | O                                 | O                   |
| PTZ Node                           | PTZ               | GetNodes                            | X                   | O         | X                                 | X                   |
|                                    |                   | GetNode                             | X                   | O         | X                                 | X                   |
| PTZ Configuration                  | PTZ               | GetConfigurations                   | X                   | O         | X                                 | X                   |
|                                    |                   | GetConfiguration                    | X                   | O         | X                                 | X                   |
|                                    |                   | GetConfigurationOptions             | X                   | O         | X                                 | X                   |
|                                    |                   | SetConfiguration                    | X                   | X         | X                                 | X                   |
|                                    | Media             | AddPTZConfiguration                 | X                   | X         | X                                 | X                   |
|                                    | Media             | RemovePTZConfiguration              | X                   | X         | X                                 | X                   |
| PTZ Actuation                      | PTZ               | AbsoluteMove                        | X                   | O         | X                                 | X                   |
|                                    |                   | RelativeMove                        | X                   | O         | X                                 | X                   |
|                                    |                   | ContinuousMove                      | X                   | O         | X                                 | X                   |
|                                    |                   | Stop                                | X                   | O         | X                                 | X                   |
|                                    |                   | GetStatus                           | X                   | O         | X                                 | X                   |
| PTZ Preset                         | PTZ               | SetPreset                           | X                   | O         | X                                 | X                   |
|                                    |                   | GetPresets                          | X                   | O         | X                                 | X                   |
|                                    |                   | GotoPreset                          | X                   | O         | X                                 | X                   |
|                                    |                   | RemovePreset                        | X                   | O         | X                                 | X                   |
| PTZ Home Position                  | PTZ               | GotoHomePosition                    | X                   | X         | X                                 | X                   |
|                                    |                   | SetHomePosition                     | X                   | X         | X                                 | X                   |
| PTZ AuxiliaryOperations            | PTZ               | SendAuxiliaryCommand                | X                   | X         | X                                 | X                   |
| Event Handling                     | Event             | Notify                              | O                   | X         | O                                 | X                   |
|                                    |                   | Subscribe                           | O                   | X         | O                                 | X                   |
|                                    |                   | Renew                               | X                   | X         | O                                 | X                   |
|                                    |                   | Unsubscribe                         | O                   | X         | O                                 | X                   |
|                                    |                   | CreatePullPointSubscription         | O                   | X         | O                                 | X                   |
|                                    |                   | PullMessages                        | O                   | X         | O                                 | X                   |
|                                    |                   | TopicFilter                         | O                   | X         | O                                 | X                   |
|                                    |                   | MessageContentFilter                | X                   | X         | X                                 | X                   |
| Configuration of Analytics profile | Media2            | GetProfiles                         | X                   | X         | O                                 | X                   |
|                                    |                   | GetAnalyticsConfigurations          | X                   | X         | O                                 | X                   |
|                                    |                   | AddConfiguration                    | X                   | X         | O                                 | X                   |
|                                    |                   | RemoveConfiguration                 | X                   | X         | O                                 | X                   |
| Analytics Module configuration     | Analytics         | GetSupportedAnalyticsModules        | X                   | X         | O                                 | X                   |
|                                    |                   | GetAnalyticsModules                 | X                   | X         | O                                 | X                   |
|                                    |                   | CreateAnalyticsModules              | X                   | X         | X                                 | X                   |
|                                    |                   | DeleteAnalyticsModules              | X                   | X         | X                                 | X                   |
|                                    |                   | GetAnalyticsModuleOptions           | X                   | X         | O                                 | X                   |
|                                    |                   | ModifyAnalyticsModules              | X                   | X         | O                                 | X                   |
| Rule configuration                 | Analytics         | GetSupportedRules                   | X                   | X         | O                                 | X                   |
|                                    |                   | GetRules                            | X                   | X         | O                                 | X                   |
|                                    |                   | CreateRules                         | X                   | X         | O                                 | X                   |
|                                    |                   | DeleteRules                         | X                   | X         | O                                 | X                   |
|                                    |                   | GetRuleOptions                      | X                   | X         | O                                 | X                   |
|                                    |                   | ModifyRules                         | X                   | X         | O                                 | X                   |
