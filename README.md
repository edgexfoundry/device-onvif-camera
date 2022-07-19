# Onvif Camera Device Service
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-onvif-camera/job/main/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-onvif-camera/job/main/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/device-onvif-camera/branch/main/graph/badge.svg?token=9AIEBTKLCC)](https://codecov.io/gh/edgexfoundry/device-onvif-camera) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/device-onvif-camera)](https://goreportcard.com/report/github.com/edgexfoundry/device-onvif-camera) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-onvif-camera?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/device-onvif-camera/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-onvif-camera?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/device-onvif-camera)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/device-onvif-camera) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/device-onvif-camera)](https://github.com/edgexfoundry/device-onvif-camera/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/device-onvif-camera)](https://github.com/edgexfoundry/device-onvif-camera/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/device-onvif-camera-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/device-onvif-camera)](https://github.com/edgexfoundry/device-onvif-camera/commits)

## Overview
The Open Network Video Interface Forum (ONVIF) Device Service is a microservice created to address the lack of standardization and automation of camera discovery and onboarding. EdgeX Foundry is a flexible microservice-based architecture created to promote the interoperability of multiple device interface combinations at the edge. In an EdgeX deployment, the ONVIF Device Service controls and communicates with ONVIF-compliant cameras, while EdgeX Foundry presents a standard interface to application developers. With normalized connectivity protocols and a vendor-neutral architecture, EdgeX paired with ONVIF Camera Device Service, simplifies deployment of edge camera devices. 


Use the ONVIF Device Service to streamline and scale your edge camera device deployment. 

## How It Works
The figure below illustrates the software flow through the architecture components.

![high-level-arch](./doc/images/ONVIFDeviceServiceArch.png)
<p align="left">
      <i>Figure 1: Software Flow</i>
</p>

1. **EdgeX Device Discovery:** Camera device microservices probe network and platform for video devices at a configurable interval. Devices that do not currently exist and that satisfy Provision Watcher filter criteria are added to Core Metadata.
2. **Application Device Discovery:** Query Core Metadata for devices and associated configuration.
3. **Application Device Configuration:** Set configuration and initiate device actions through a REST API representing the resources of the video device (e.g. stream URI, Pan-Tilt-Zoom position, Firmware Update).
4. **Pipeline Control:** The application initiates Video Analytics Pipeline through HTTP Post Request.
5. **Publish Inference Events/Data:** Analytics inferences are formatted and passed to the destination message bus specified in the request.
6.  **Export Data:** Publish prepared (transformed, enriched, filtered, etc.) and groomed (formatted, compressed, encrypted, etc.) data to external systems (be it analytics package, enterprise or on-premises application, cloud systems like Azure IoT, AWS IoT, or Google IoT Core, etc.


# Getting Started

Learn how to configure and run the service by following these [instructions](./doc/setup.md). 

For a full walkthrough of using the default images, use this [guide.](./doc/guides/SimpleStartupGuide.md)  

For a full walktrhough of building custom images, use this [guide.](./doc/guides/CustomStartupGuide.md)  


# Learn More 

### General
[Supported ONVIF features](./doc/ONVIF-protocol.md)  
[Credentials]()  
[Auto discovery](./doc/auto-discovery.md)  
[Control-plane events](./doc/control-plane-events.md)  


### Custom Features
[Custom Metadata](./doc/custom-metadata-feature.md)  
[Reboot Needed](./doc/custom-feature-rebootneeded.md)  

### API Support
[API Analytic Handling](./doc/api-analytic-support.md)  
[API Event Handling](./doc/api-event-handling.md)  
[API User Handling](./doc/api-usage-user-handling.md)  

### Miscellaneous
[Postman](./doc/test-with-postman.md)  
[User Authentication](./doc/onvif-user-authentication.md)  

## Resources
[Learn more about EdgeX Core Metadata](https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0)  
[Learn more about EdgeX Core Command](https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-command/2.1.0)


## References

- ONVIF Website: http://www.onvif.org
- EdgeX Foundry Project Wiki: https://wiki.edgexfoundry.org/
- EdgeX Source Code: https://github.com/edgexfoundry
- Edgex Developer Guide: https://docs.edgexfoundry.org/2.1/
- Docker Repos
   - Docker https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository
   - Docker Compose https://docs.docker.com/compose/install/#install-compose


## License

[Apache-2.0](https://github.com/edgexfoundry-holding/device-onvif-camera/blob/main/LICENSE)
