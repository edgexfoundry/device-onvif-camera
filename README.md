# Onvif Camera Device Service
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-onvif-camera/job/main/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-onvif-camera/job/main/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/device-onvif-camera/branch/main/graph/badge.svg?token=9AIEBTKLCC)](https://codecov.io/gh/edgexfoundry/device-onvif-camera) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/device-onvif-camera)](https://goreportcard.com/report/github.com/edgexfoundry/device-onvif-camera) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-onvif-camera?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/device-onvif-camera/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-onvif-camera?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/device-onvif-camera)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/device-onvif-camera) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/device-onvif-camera)](https://github.com/edgexfoundry/device-onvif-camera/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/device-onvif-camera)](https://github.com/edgexfoundry/device-onvif-camera/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/device-onvif-camera-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/device-onvif-camera)](https://github.com/edgexfoundry/device-onvif-camera/commits)


This Onvif Camera Device Service is developed to control/communicate ONVIF-compliant cameras accessible via http in an EdgeX deployment

## Onvif Features
The device service supports the onvif features listed in the following table:

| Feature                            | Onvif Web Service | Onvif Function                                                                                                                   | EdgeX Value Type |
|------------------------------------|-------------------|----------------------------------------------------------------------------------------------------------------------------------|------------------|
| User Authentication                | Core              | **WS-Usernametoken Authentication**                                                                                              | Object           |
|                                    |                   | **HTTP Digest**                                                                                                                  | Object           |
| Auto Discovery                     | Core              | **WS-Discovery**                                                                                                                 | Object           |
|                                    | Device            | [GetDiscoveryMode](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetDiscoveryMode)                                  | Object           |
|                                    |                   | [SetDiscoveryMode](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetDiscoveryMode)                                  | Object           |
|                                    |                   | [GetScopes](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetScopes)                                                | Object           |
|                                    |                   | [SetScopes](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetScopes)                                                | Object           |
|                                    |                   | [AddScopes](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.AddScopes)                                                | Object           |
|                                    |                   | [RemoveScopes](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.RemoveScopes)                                          | Object           |
| Network Configuration              | Device            | [GetHostname](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetHostname)                                            | Object           |
|                                    |                   | [SetHostname](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetHostname)                                            | Object           |
|                                    |                   | [GetDNS](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetDNS)                                                      | Object           |
|                                    |                   | [SetDNS](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetDNS)                                                      | Object           |
|                                    |                   | [**GetNetworkInterfaces**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetNetworkInterfaces)                      | Object           |
|                                    |                   | [**SetNetworkInterfaces**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetNetworkInterfaces)                      | Object           |
|                                    |                   | [GetNetworkProtocols](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetNetworkProtocols)                            | Object           |
|                                    |                   | [SetNetworkProtocols](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetNetworkProtocols)                            | Object           |
|                                    |                   | [**GetNetworkDefaultGateway**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetNetworkDefaultGateway)              | Object           |
|                                    |                   | [**SetNetworkDefaultGateway**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetNetworkDefaultGateway)              | Object           |
| System Function                    | Device            | [**GetDeviceInformation**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetDeviceInformation)                      | Object           |
|                                    |                   | [GetSystemDateAndTime](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetSystemDateAndTime)                          | Object           |
|                                    |                   | [SetSystemDateAndTime](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetSystemDateAndTime)                          | Object           |
|                                    |                   | [SetSystemFactoryDefault](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetSystemFactoryDefault)                    | Object           |
|                                    |                   | [SystemReboot](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SystemReboot)                                          | Object           |
| User Handling                      | Device            | [**GetUsers**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.GetUsers)                                              | Object           |
|                                    |                   | [**CreateUsers**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.CreateUsers)                                        | Object           |
|                                    |                   | [**DeleteUsers**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.DeleteUsers)                                        | Object           |
|                                    |                   | [**SetUser**](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl#op.SetUser)                                                | Object           |
| Metadata Configuration             | Media             | [GetMetadataConfiguration](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetMetadataConfiguration)                        | Object           |
|                                    |                   | [GetMetadataConfigurations](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetMetadataConfigurations)                      | Object           |
|                                    |                   | [GetCompatibleMetadataConfigurations](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetCompatibleMetadataConfigurations)  | Object           |
|                                    |                   | [**GetMetadataConfigurationOptions**](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetMetadataConfigurationOptions)      | Object           |
|                                    |                   | [AddMetadataConfiguration](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.AddMetadataConfiguration)                        | Object           |
|                                    |                   | [RemoveMetadataConfiguration](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.RemoveMetadataConfiguration)                  | Object           |
|                                    |                   | [**SetMetadataConfiguration**](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.SetMetadataConfiguration)                    | Object           |
| Video Streaming                    | Media             | [**GetProfiles**](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetProfiles)                                              | Object           |
|                                    |                   | [**GetStreamUri**](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetStreamUri)                                            | Object           |
| VideoEncoder  Config               | Media             | [GetVideoEncoderConfiguration](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetVideoEncoderConfiguration)                | Object           |
|                                    |                   | [**SetVideoEncoderConfiguration**](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.SetVideoEncoderConfiguration)            | Object           |
|                                    |                   | [GetVideoEncoderConfigurationOptions](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.GetVideoEncoderConfigurationOptions)  | Object           |
| PTZ Configuration                  | PTZ               | [GetNode](http://www.onvif.org/onvif/ver20/ptz/wsdl/ptz.wsdl#op.GetNode)                                                         | Object           |
|                                    |                   | [GetConfigurations](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.GetConfigurations)                                          | Object           |
|                                    |                   | [GetConfiguration](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.GetConfiguration)                                            | Object           |
|                                    |                   | [GetConfigurationOptions](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.GetConfigurationOptions)                              | Object           |
|                                    |                   | [SetConfiguration](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.SetConfiguration)                                            | Object           |
|                                    | Media             | [AddPTZConfiguration](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.AddPTZConfiguration)                                  | Object           |
|                                    | Media             | [RemovePTZConfiguration](https://www.onvif.org/ver10/media/wsdl/media.wsdl#op.RemovePTZConfiguration)                            | Object           |
| PTZ Actuation                      | PTZ               | [AbsoluteMove](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.AbsoluteMove)                                                    | Object           |
|                                    |                   | [RelativeMove](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.RelativeMove)                                                    | Object           |
|                                    |                   | [ContinuousMove](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.ContinuousMove)                                                | Object           |
|                                    |                   | [Stop](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.Stop)                                                                    | Object           |
|                                    |                   | [GetStatus](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.GetStatus)                                                          | Object           |
|                                    |                   | [GetPresets](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.GetPresets)                                                        | Object           |
|                                    |                   | [GotoPreset](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.GotoPreset)                                                        | Object           |
|                                    |                   | [RemovePreset](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.RemovePreset)                                                    | Object           |
| PTZ Home Position                  | PTZ               | [GotoHomePosition](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.GotoHomePosition)                                            | Object           |
|                                    |                   | [SetHomePosition](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.SetHomePosition)                                              | Object           |
| PTZ AuxiliaryOperations            | PTZ               | [SendAuxiliaryCommand](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl#op.SendAuxiliaryCommand)                                    | Object           |
| Event Handling                     | Event             | [Notify](https://docs.oasis-open.org/wsn/wsn-ws_base_notification-1.3-spec-os.pdf)                                               | Object           |
|                                    |                   | [Subscribe](https://docs.oasis-open.org/wsn/wsn-ws_base_notification-1.3-spec-os.pdf)                                            | Object           |
|                                    |                   | [Renew](https://docs.oasis-open.org/wsn/wsn-ws_base_notification-1.3-spec-os.pdf)                                                | Object           |
|                                    |                   | [Unsubscribe](https://www.onvif.org/ver10/events/wsdl/event.wsdl#op.Unsubscribe)                                                 | Object           |
|                                    |                   | [CreatePullPointSubscription](https://www.onvif.org/ver10/events/wsdl/event.wsdl#op.CreatePullPointSubscription)                 | Object           |
|                                    |                   | [PullMessages](https://www.onvif.org/ver10/events/wsdl/event.wsdl#op.PullMessages)                                               | Object           |
|                                    |                   | [TopicFilter](https://docs.oasis-open.org/wsn/wsn-ws_base_notification-1.3-spec-os.pdf)                                          | Object           |
|                                    |                   | [MessageContentFilter](https://docs.oasis-open.org/wsn/wsn-ws_base_notification-1.3-spec-os.pdf)                                 | Object           |
| Configuration of Analytics profile | Media2            | [GetProfiles](https://www.onvif.org/ver20/media/wsdl/media.wsdl#op.GetProfiles)                                                  | Object           |
|                                    |                   | [GetAnalyticsConfigurations](https://www.onvif.org/ver20/media/wsdl/media.wsdl#op.GetAnalyticsConfigurations)                    | Object           |
|                                    |                   | [AddConfiguration](https://www.onvif.org/ver20/media/wsdl/media.wsdl#op.AddConfiguration)                                        | Object           |
|                                    |                   | [RemoveConfiguration](https://www.onvif.org/ver20/media/wsdl/media.wsdl#op.RemoveConfiguration)                                  | Object           |
| Analytics Module configuration     | Analytics         | [GetSupportedAnalyticsModules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.GetSupportedAnalyticsModules)        | Object           |
|                                    |                   | [GetAnalyticsModules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.GetAnalyticsModules)                          | Object           |
|                                    |                   | [CreateAnalyticsModules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.CreateAnalyticsModules)                    | Object           |
|                                    |                   | [DeleteAnalyticsModules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.DeleteAnalyticsModules)                    | Object           |
|                                    |                   | [GetAnalyticsModuleOptions](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.GetAnalyticsModuleOptions)              | Object           |
|                                    |                   | [ModifyAnalyticsModules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.ModifyAnalyticsModules)                    | Object           |
| Rule configuration                 | Analytics         | [GetSupportedRules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.GetSupportedRules)                              | Object           |
|                                    |                   | [GetRules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.GetRules)                                                | Object           |
|                                    |                   | [CreateRules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.CreateRules)                                          | Object           |
|                                    |                   | [DeleteRules](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.DeleteRules)                                          | Object           |
|                                    |                   | [GetRuleOptions](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.GetRuleOptions)                                    | Object           |
|                                    |                   | [ModifyRule](https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl#op.ModifyRules)                                           | Object           |

**Note**: The functions in the bold text are **mandatory** for Onvif protocol.

## Custom Features
The device service also include custom function to enhance the usage for the EdgeX user.

| Feature         | Service | Function               | EdgeX Value Type | Description                                                                            |
|-----------------|---------|------------------------|------------------|----------------------------------------------------------------------------------------|
| System Function | EdgeX   | RebootNeeded           | Bool             | Read only. Used to indicate the camera should reboot to apply the configuration change |
| System Function | EdgeX   | CameraEvent            | Bool             | A device resource which is used to send the async event to north bound                 |
| System Function | EdgeX   | SubscribeCameraEvent   | Bool             | Create a subscription to subscribe the event from the camera                           |
| System Function | EdgeX   | UnsubscribeCameraEvent | Bool             | Unsubscribe all subscription from the camera                                           |
| Media           | EdgeX   | GetSnapshot            | Binary           | Get Snapshot from the snapshot uri                                                     |
| Custom Metadata | EdgeX   | CustomMetadata         | Object           | Read and write custom metadata to the camera entry in EdgeX                            | 
| Custom Metadata | EdgeX   | DeleteCustomMetadata   | Object           | Delete custom metadata fields from the camera entry in EdgeX                           |

## How does the device service work?

The Onvif camera uses Web Services standards such as XML, SOAP 1.2 and WSDL1.1 over an IP network. 
- XML is used as the data description syntax
- SOAP is used for message transfer 
- and WSDL is used for describing the services.

The spec can refer to [ONVIF-Core-Specification](https://www.onvif.org/specs/core/ONVIF-Core-Specification-v221.pdf).

For example, we can send a SOAP request to the Onvif camera as below:
```shell
curl --request POST 'http://192.168.12.128:2020/onvif/service' \
--header 'Content-Type: application/soap+xml' \
--data-raw '<?xml version="1.0" encoding="UTF-8"?>
<soap-env:Envelope xmlns:soap-env="http://www.w3.org/2003/05/soap-envelope" xmlns:soap-enc="http://www.w3.org/2003/05/soap-encoding" xmlns:tan="http://www.onvif.org/ver20/analytics/wsdl" xmlns:onvif="http://www.onvif.org/ver10/schema" xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:timg="http://www.onvif.org/ver20/imaging/wsdl" xmlns:tds="http://www.onvif.org/ver10/device/wsdl" xmlns:tev="http://www.onvif.org/ver10/events/wsdl" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" >
    <soap-env:Header>
        <Security xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
            <UsernameToken>
                <Username>myUsername</Username>
                <Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">+HKcvc+LCGClVwuros1sJuXepQY=</Password>
                <Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">w490bn6rlib33d5rb8t6ulnqlmz9h43m</Nonce>
                <Created xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">2021-10-21T03:43:21.02075Z</Created>
            </UsernameToken>
        </Security>
    </soap-env:Header>
    <soap-env:Body>
        <trt:GetStreamUri>
            <trt:ProfileToken>profile_1</trt:ProfileToken>
        </trt:GetStreamUri>
    </soap-env:Body>
  </soap-env:Envelope>'
```
And the response should be like the following XML data:
```shell
<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
	xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope" xmlns:SOAP-ENC="http://www.w3.org/2003/05/soap-encoding"
	xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
	xmlns:wsdd="http://schemas.xmlsoap.org/ws/2005/04/discovery" xmlns:chan="http://schemas.microsoft.com/ws/2005/02/duplex"
	xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd" xmlns:wsa5="http://www.w3.org/2005/08/addressing"
	xmlns:xmime="http://tempuri.org/xmime.xsd" xmlns:xop="http://www.w3.org/2004/08/xop/include" xmlns:wsrfbf="http://docs.oasis-open.org/wsrf/bf-2"
	xmlns:wstop="http://docs.oasis-open.org/wsn/t-1" xmlns:wsrfr="http://docs.oasis-open.org/wsrf/r-2" xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"
	xmlns:tt="http://www.onvif.org/ver10/schema" xmlns:ter="http://www.onvif.org/ver10/error" xmlns:tns1="http://www.onvif.org/ver10/topics"
	xmlns:tds="http://www.onvif.org/ver10/device/wsdl" xmlns:trt="http://www.onvif.org/ver10/media/wsdl"
	xmlns:tev="http://www.onvif.org/ver10/events/wsdl" xmlns:tdn="http://www.onvif.org/ver10/network/wsdl" xmlns:timg="http://www.onvif.org/ver20/imaging/wsdl"
	xmlns:trp="http://www.onvif.org/ver10/replay/wsdl" xmlns:tan="http://www.onvif.org/ver20/analytics/wsdl" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl">
	<SOAP-ENV:Header></SOAP-ENV:Header>
	<SOAP-ENV:Body>
		<trt:GetStreamUriResponse>
			<trt:MediaUri>
				<tt:Uri>rtsp://192.168.12.128:554/stream1</tt:Uri>
				<tt:InvalidAfterConnect>false</tt:InvalidAfterConnect>
				<tt:InvalidAfterReboot>false</tt:InvalidAfterReboot>
				<tt:Timeout>PT0H0M2S</tt:Timeout>
			</trt:MediaUri>
		</trt:GetStreamUriResponse>
	</SOAP-ENV:Body>
</SOAP-ENV:Envelope>
```

Since the SOAP is a HTTP call, the device service can just do the transformation between REST(JSON) and SOAP(XML).

For the concept of implementation:
- The device service accepts the REST request from the client, then transforms the request to SOAP format and forward it to the Onvif camera.
- Once the device service receives the response from the Onvif camera, the device service will transform the SOAP response to REST format for the client.
```
                  - Onvif Web Service

                  - Onvif Function  ┌────────────────────┐
                                    │                    │
┌──────────────┐  - Input Parameter │   Device Service   │               ┌─────────────────┐
│              │                    │                    │               │                 │
│              │ REST request       │                    │ SOAP request  │                 │
│    Client  ──┼────────────────────┼──►  Transform  ────┼───────────────┼──► Onvif Camera │
│              │                    │   to SOAP request  │               │                 │
│              │                    │                    │               │                 │
└──────────────┘                    └────────────────────┘               └─────────────────┘


                                    ┌────────────────────┐
                                    │                    │
┌──────────────┐                    │   Device Service   │               ┌─────────────────┐
│              │                    │                    │               │                 │
│              │ REST response      │                    │ SOAP response │                 │
│    Client  ◄─┼────────────────────┼───  Transform   ◄──┼───────────────┼── Onvif Camera  │
│              │                    │   to REST response │               │                 │
│              │                    │                    │               │                 │
└──────────────┘                    └────────────────────┘               └─────────────────┘
```

## General Usage

### Run Unit Test
```shell
make test
```

### Build the executable file
```shell
make build
```

### Build docker image
Build docker image named edgex/device-onvif-camera:0.0.0-dev with the following command:
```shell
make docker
```

### Define the device profile

The device resource should provide two attributes:
* **service** indicates the web service for the Onvif
* **function** indicates the SOAP action for the specified web service

For example:
```yaml
deviceResources:
  - name: "Hostname"
    isHidden: false
    description: "Camera Hostname"
    attributes:
      service: "Device"
      getFunction: "GetHostname"
      setFunction: "SetHostname"
    properties:
      valueType: "Object"
      readWrite: "RW"
```

See the sample at [cmd/res/profiles/camera.yaml](cmd/res/profiles/camera.yaml)

### Define the device

The device's protocol properties should contain:
* **Address** is the IP address of the Onvif camera
* **Port** is the server port of the Onvif camera
* **AuthMode** indicates the auth mode of the Onvif camera
* **SecretPath** indicates the path to retrieve the username and password

For example:
```yaml
[[DeviceList]]
Name = "Camera001"
ProfileName = "camera"
Description = "My test camera"
  [DeviceList.Protocols]
    [DeviceList.Protocols.Onvif]
    Address = "192.168.12.123"
    Port = 80
    # Assign AuthMode to "digest" | "usernametoken" | "both" | "none"
    AuthMode = "usernametoken"
    SecretPath = "credentials001"
```
See the sample at [cmd/res/devices/camera.toml.example](cmd/res/devices/camera.toml.example)


## Getting Started Guide
- [Getting started guide for developer](./doc/getting-started-guide.md)
- [Getting started guide for running docker container with none security mode](./doc/getting-started-with-docker.md)
- [Getting started guide for running docker container with security mode](./doc/getting-started-with-docker-security.md)

## Tested Onvif Camera
[Tested Onvif cameras with Onvif functions.](./doc/tested-onvif-camera.md)

## Pass parameter via URL query parameter
[Get Command Parameter](./doc/get-cmd-parameter.md)

## Custom feature for device-onvif-go
[RebootNeeded](./doc/custom-feature-rebootneeded.md)

## API Usage

- [User Handling](./doc/api-usage-user-handling.md)
- [Analytics Support](./doc/api-analytic-support.md)
- [Test with Postman](./doc/test-with-postman.md) - Test with Postman for ONVIF and device-onvif-camera APIs

## Onvif User Authentication
[Setup Onvif User Authentication](./doc/onvif-user-authentication.md)

## Auto Discovery
- [Onvif camera auto discovery](./doc/auto-discovery.md)
- [Set up auto discovery with docker](./doc/auto-discovery-docker.md)

## Control Plane Events
[Control Plane Events](./doc/control-plane-events.md)

## Custom Metadata
[CustomMetadata](./doc/custom-metadata-feature.md)

## Get and set Friendly Name and MAC Address
[FriendlyName and MACAddress](./doc/get-set-friendlyname-mac.md)

## License
[Apache-2.0](LICENSE)
