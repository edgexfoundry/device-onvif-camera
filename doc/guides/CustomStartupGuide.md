# ONVIF Device Service Custom Start Up Guide

## Contents

[System Requirements](#system-requirements)  
[Dependencies](#dependencies)  
[Get the Source Code](#get-the-source-code)  
[Configure the Pre-Defined Devices](#configure-the-pre-defined-devices)   
[Configure the Device Service](#configure-the-device-service)  
[Build the Docker Image](#build-the-docker-image)  
[Deploy the Service](#deploy-edgex-and-onvif-device-camera-microservice)  
[Verify the Service](#verify-service-and-device-profiles)  
[Manage Devices](#manage-devices)  
[Execute Example Command](#execute-getstreamuri-command-through-edgex)  
[Shutting Down](#shutting-down)  
[Additional Configuration](#additional-configuration)  
[Next Steps](#summary-and-next-steps)    


## System Requirements

- Intel&#8482; Core&#174; processor
- Ubuntu 20.04.4 LTS
- ONVIF-compliant Camera

>NOTE: The instructions in this guide were developed and tested using Ubuntu 20.04 LTS and the Tapo C200 Pan/Tilt Wi-Fi Camera. However, the software may work with other Linux distributions and ONVIF-compliant cameras. Refer to our [list of tested cameras for more information](./ONVIF-protocol.md#tested-onvif-cameras)

**Time to Complete**

20-30 minutes

**Other Requirements**

You must have administrator (sudo) privileges to execute the user guide commands.

## How It Works
For an explanation of the architecture, see the [User Guide](UserGuide.md#how-it-works).

## Dependencies
The software has dependencies, including Git, Docker, Docker Compose, and assorted tools. Follow the instructions below to install any dependency that is not already installed. 

### Install Git
Install Git from the official repository as documented on the [Git SCM](https://git-scm.com/download/linux) site.

1. Update installation repositories:
   ```bash
   sudo apt update
   ```

2. Add the Git repository:
   ```bash
   sudo add-apt-repository ppa:git-core/ppa -y
   ```

3. Install Git:
   ```bash
   sudo apt install git
   ```

### Install Docker
Install Docker from the official repository as documented on the [Docker](https://docs.docker.com/engine/install/ubuntu/) site.

### Verify Docker
To enable running Docker commands without the preface of sudo, add the user to the Docker group. Then run Docker with the `hello-world` test.

1. Create Docker group:
   ```bash
   sudo groupadd docker
   ```
   >NOTE: If the group already exists, `groupadd` outputs a message: **groupadd: group `docker` already exists**. This is OK.

2. Add User to group:
   ```bash
   sudo usermod -aG docker $USER
   ```

3. Refresh the group:
   ```bash
   newgrp docker 
   ```

4. To verify the Docker installation, run `hello-world`:

   ```bash
   docker run hello-world
   ```
   A **Hello from Docker!** greeting indicates successful installation.

   ```bash
   Unable to find image 'hello-world:latest' locally
   latest: Pulling from library/hello-world
   2db29710123e: Pull complete 
   Digest: sha256:10d7d58d5ebd2a652f4d93fdd86da8f265f5318c6a73cc5b6a9798ff6d2b2e67
   Status: Downloaded newer image for hello-world:latest

   Hello from Docker!
   This message shows that your installation appears to be working correctly.
   ...
   ```

### Install Docker Compose
Install Docker from the official repository as documented on the [Docker Compose](https://docs.docker.com/compose/install/#install-compose) site. See the Linux tab. 

1. Download current stable Docker Compose:
   ```bash
   sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   ```
   >NOTE: When this guide was created, version 1.29.2 was current.

2. Set permissions:
   ```bash
   sudo chmod +x /usr/local/bin/docker-compose
   ```

###  Download EdgeX Compose
Clone the EdgeX compose repository

   ```bash
   git clone https://github.com/edgexfoundry/edgex-compose.git
   ```

### Install Tools
Install the build, media streaming, and parsing tools:

```bash
sudo apt install build-essential ffmpeg jq curl
```

### Tool Descriptions
The table below lists command line tools this guide uses to help with EdgeX configuration and device setup.

| Tool        | Description | Note |
| ----------- | ----------- |----------- |
| **curl**     | Allows the user to connect to services such as EdgeX. |Use curl to get transfer information either to or from this service. In the tutorial, use `curl` to communicate with the EdgeX API. The call will return a JSON object.|
| **jq**   |Parses the JSON object returned from the `curl` requests. |The `jq` command includes parameters that are used to parse and format data. In this tutorial, the `jq` command has been configured to return and format appropriate data for each `curl` command that is piped into it. |
| **base64**   | Converts data into the Base64 format.| |

>Table 1: Command Line Tools

## Get the Source Code

1. Clone the device-onvif-camera repository:

   ```bash
   git clone https://github.com/edgexfoundry/device-onvif-camera.git
   ```

2. Navigate into the directory;

   ```bash
   cd device-onvif-camera
   ```

## Configuration

### Configure the Pre-Defined Devices

Configuring pre-defined devices will allow the service to automatically provision them into core-metadata. Create a list of devices with the appropriate information as outlined below.

1. Make a copy of the `camera.toml.example`:  

   ```bash
   cp ./cmd/res/devices/camera.toml.example ./cmd/res/devices/camera.toml
   ```

2. Open the `cmd/res/devices/camera.toml` file using your preferred text editor and update the `Address` and `Port` fields to match the IP address of the Camera and port used for ONVIF services:

   ```toml
   [[DeviceList]]
   Name = "Camera001"                         # Modify as desired
   ProfileName = "onvif-camera"               # Default profile
   Description = "onvif conformant camera"    # Modify as desired
      [DeviceList.Protocols]
         [DeviceList.Protocols.Onvif]
         Address = "191.168.86.34"              # Set to your camera IP address
         Port = "2020"                          # Set to the port your camera uses
         SecretPath = "credentials001"
         [DeviceList.Protocols.CustomMetadata]
         CommonName = "Outdoor camera"
   ```
   <p align="left">
      <i>Sample: Snippet from camera.toml</i>
   </p>

3. Optionally, modify the `Name` and `Description` fields to more easily identify the camera. The `Name` is the camera name used when using ONVIF Device Service Rest APIs. The `Description` is simply a more detailed explanation of the camera.

4. You can also optionally configure the `CustomMetadata` with custom fields and values to store any extra information you would like.

5. To add more pre-defined devices, copy the above configuration and edit to match your extra devices.


### Configure the Device Service
1. Open the [configuration.toml](../../cmd/res/configuration.toml) file using your preferred text editor.

2. Make sure `path` is set to match `SecretPath` in `camera.toml`. In the sample below, it is `"credentials001"`. If you have multiple cameras, make sure the secret paths match.

3. Under `path`, set `username` and `password` to your camera credentials. If you have multiple cameras copy the `Writable.InsecureSecrets` section and edit to include the new information.

```toml
[Writable]
    [Writable.InsecureSecrets.credentials001]
    path = "credentials001"
      [Writable.InsecureSecrets.credentials001.Secrets]
      username = "<Credentials 1 username>"
      password = "<Credentials 1 password>"
      mode = "usernametoken" # assign "digest" | "usernametoken" | "both" | "none"

    [Writable.InsecureSecrets.credentials002]
    path = "credentials002"
      [Writable.InsecureSecrets.credentials002.Secrets]
      username = "<Credentials 1 password>"
      password = "<Credentials 2 password>"
      mode = "usernametoken" # assign "digest" | "usernametoken" | "both" | "none"

```

<p align="left">
   <i>Sample: Snippet from configuration.toml</i>
</p>


### Additional Configuration Options
For optional configurations, see [here.](#additional-configuration)

## Build the Docker Image

1. In the `device-onvif-camera` directory, run make docker:

   ```bash
   make docker
   ```

2. Verify the ONVIF Device Service Docker image was successfully created:

   ```bash
   docker images
   ```
   ```docker
   REPOSITORY                                 TAG          IMAGE ID       CREATED        SIZE
   edgexfoundry-holding/device-onvif-camera   0.0.0-dev    75684e673feb   6 weeks ago    21.3MB
   ```

3. Navigate to `edgex-compose` and enter the `compose-builder` directory:

   ```bash
   cd edgex-compose/compose-builder
   ```

4. Update `.env` file to add the registry and image version variable for device-onvif-camera:

   Add the following registry and version information:
   ```env
   DEVICE_ONVIFCAM_VERSION=0.0.0-dev
   ```

5. Update the `add-device-onvif-camera.yml` to point to the local image:

   ```yml
   services:
      device-onvif-camera:
         image: edgexfoundry/device-onvif-camera:${DEVICE_ONVIFCAM_VERSION}
   ```

## Deploy EdgeX and ONVIF Device Camera Microservice

<details>
<summary><strong>Run the Service using Docker</strong></summary>

   1. Navigate to the EdgeX `compose-builder` directory:

      ```bash
      cd edgex-compose/compose-builder/
      ```

   1. Run EdgeX with the microservice in non-secure mode:

      ```bash
      make run no-secty ds-onvif-camera
      ```
   
   1. Run EdgeX with the microservice in secure mode:

      ```bash
      make run ds-onvif-camera
      ```
</details>

<details>
<summary><strong>Run the Service natively</summary><strong>

<br/>

>NOTE: Go version 1.18+ is required to run natively.

<br/>

   1. Navigate to the EdgeX `compose-builder` directory:

      ```bash
      cd edgex-compose/compose-builder/
      ```

   1. Run EdgeX:

      ```bash
      make run no-secty
      ```

   1. Navigate out of the `edgex-compose` directory to the `device-onvif-camera` directory:

      ```bash
      cd device-onvif-camera
      ```

   1. Run the service:

      ```bash
      make run 
      ```

</details>

## Verify Service and Device Profiles

### Using Command Line
1. Check the status of the container:

   ```bash 
   docker ps
   ```

   The status column will indicate if the container is running, and how long it has been up.

   Example Output:

    ```docker
   CONTAINER ID   IMAGE                                                       COMMAND                  CREATED       STATUS          PORTS                                                                                         NAMES
   33f9c5ecb70e  edgexfoundry/device-onvif-camera:0.0.0-dev    "/device-onvif-camer…"   7 weeks ago   Up 48 minutes   127.0.0.1:59985->59985/tcp                                                                    edgex-device-onvif-camera
   ```

2. Check whether the device service is added to EdgeX:

   ```bash
   curl -s http://localhost:59881/api/v2/deviceservice/name/device-onvif-camera | jq .
   ```
   Successful:
   ```json
      {
         "apiVersion": "v2",
         "statusCode": 200,
         "service": {
            "created": 1657227634593,
            "modified": 1657291447649,
            "id": "e1883aa7-f440-447f-ad4d-effa2aeb0ade",
            "name": "device-onvif-camera",
            "baseAddress": "http://edgex-device-onvif-camera:59984",
            "adminState": "UNLOCKED"
         }         
      }
   ```
   Unsuccessful:
   ```json
   {
      "apiVersion": "v2",
      "message": "fail to query device service by name device-onvif-camer",
      "statusCode": 404
   }
   ```


3. Check whether the device profile is added:

   ```bash
   curl -s http://localhost:59881/api/v2/deviceprofile/name/onvif-camera | jq -r '"profileName: " + '.profile.name' + "\nstatusCode: " + (.statusCode|tostring)'

   ```
   Good response:
   ```bash
   profileName: onvif-camera
   statusCode: 200
   ```
   Bad response:
   ```bash
   profileName: 
   statusCode: 404
   ```
   > NOTE: The `jq -r` option is used to reduce the size of the displayed response. The entire device profile with all resources can be seen by removing `-r '"profileName: " + '.profile.name' + "\nstatusCode: " + (.statusCode|tostring)', and replacing it with '.'`

### Using EdgeX UI
1. Visit http://localhost:4000 to go to the dashboard for EdgeX Console GUI:

   ![EdgeXConsoleDashboard](../images/EdgeXDashboard.png)
   <p align="left">
      <i>Figure 1: EdgeX Console Dashboard</i>
   </p>

2. To see **Device Services**, **Devices**, or **Device Profiles**, click on their respective tab:

   ![EdgeXConsoleDeviceServices](../images/EdgeXDeviceServices.png)
   <p align="left">
      <i>Figure 2: EdgeX Console Device Service List</i>
   </p>

   ![EdgeXConsoleDeviceList](../images/EdgeXDeviceList.png)
   <p align="left">
      <i>Figure 3: EdgeX Console Device List</i>
   </p>

   ![EdgeXConsoleDeviceProfileList](../images/EdgeXDeviceProfiles.png)
   <p align="left">
      <i>Figure 4: EdgeX Console Device Profile List</i>
   </p>
## Manage Devices
Follow these instructions to update devices.

### Curl Commands

#### Add Device

1. Edit the information to appropriately match the camera. The fields `Address`, `MACAddress` and `Port` should match that of the camera:

   ```bash
   curl -X POST -H 'Content-Type: application/json'  \
   http://localhost:59881/api/v2/device \
   -d '[
            {
               "apiVersion": "v2",
               "device": {
                  "name":"Camera001",
                  "serviceName": "device-onvif-camera",
                  "profileName": "onvif-camera",
                  "description": "My test camera",
                  "adminState": "UNLOCKED",
                  "operatingState": "UP",
                  "protocols": {
                     "Onvif": {
                        "Address": "10.0.0.0",
                        "Port": "10000",
                        "MACAddress": "aa:bb:cc:11:22:33",
                        "FriendlyName":"Default Camera"
                     },
                     "CustomMetadata": {
                        "Location":"Front door"
                     }
                  }
               }
            }
   ]'
   ```

   Example Output: 
   ```bash
   [{"apiVersion":"v2","statusCode":201,"id":"fb5fb7f2-768b-4298-a916-d4779523c6b5"}]
   ```

 Map credentials using the `map-credentials.sh` script.  
   a. Run `bin/map-credentials.sh`    
   b. Select `(Create New)`
      ![](../images/create_new.png)
   c. Enter the Secret Path to associate with these credentials  
      ![](../images/secret_path.png)
   d. Enter the username  
      ![](../images/set_username.png)
   e. Enter the password  
      ![](../images/set_password.png)
   f. Choose the Authentication Mode  
      ![](../images/auth_mode.png)
   g. Assign one or more MAC Addresses to the credential group  
      ![](../images/assign_mac.png)  

   >NOTE: The MAC address field can be left blank if the SecretPath from the "Enter Secret Path ..." step above, is set to the DefaultSecretPath (credentials001) from the [cmd/res/configuration.toml](../cmd/res/configuration.toml).  

   h. Learn more about updating credentials [here](../utility-scripts.md)  

   Successful:
   
   ```bash 
   Dependencies Check: Success
         Consul Check: ...
                     curl -X GET http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera?keys=true
   Response [200]      Success
   curl -X GET http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/AppCustom/CredentialsMap?keys=true
   Response [200] 
   Secret Path: a
   curl -X GET http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/AppCustom/CredentialsMap/a?raw=true
   Response [404] 
   Failed! curl returned a status code of '404'
   Setting InsecureSecret: a/Path
   curl --data "<redacted>" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/a/Path
   Response [200] true


   Setting InsecureSecret: a/Secrets/username
   curl --data "<redacted>" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/a/Secrets/username
   Response [200] true


   Setting InsecureSecret: a/Secrets/password
   curl --data "<redacted>" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/a/Secrets/password
   Response [200] true


   Setting InsecureSecret: a/Secrets/mode
   curl --data "usern<redacted>metoken" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/a/Secrets/mode
   Response [200] true


   Setting Credentials Map: a = ''
   curl -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/AppCustom/CredentialsMap/a
   Response [200] true



   Secret Path: a
   curl -X GET http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/AppCustom/CredentialsMap/a?raw=true
   Response [200] 
   Setting Credentials Map: a = '11:22:33:44:55:66'
   curl --data "11:22:33:44:55:66" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/AppCustom/CredentialsMap/a
   Response [200] true
   ``` 

   <a name="verify-device"></a>  

2. Verify device(s) have been successfully added to core-metadata.

   ```bash
   curl -s http://localhost:59881/api/v2/device/all | jq -r '"deviceName: " + '.devices[].name''
   ```

   Example Output: 
   ```bash
   deviceName: Camera001
   deviceName: device-onvif-camera
   ```
   >NOTE: The device with name `device-onvif-camera` is a stand-in device and can be ignored.  
   >NOTE: The `jq -r` option is used to reduce the size of the displayed response. The entire device with all information can be seen by removing `-r '"deviceName: " + '.devices[].name'', and replacing it with '.'`

#### Update Device

   There are multiple commands that can update aspects of the camera entry in meta-data. Refer to the [Swagger documentation]() for more information (not implemented).

#### Delete Device

   ```bash
   curl -X 'DELETE' \
   'http://localhost:59881/api/v2/device/name/<device name>' \
   -H 'accept: application/json' 
   ```
## Execute GetStreamURI Command through EdgeX

1. <a name="step1"></a>Get the profile token by executing the `GetProfiles` command:

   >NOTE: Make sure to replace `Camera001` in all the commands below, with the deviceName returned in the ["Verify device(s) have been successfully added to core-metadata"](#verify-device) step above.  

   ```bash
   curl -s http://0.0.0.0:59882/api/v2/device/name/Camera001/Profiles | jq -r '"profileToken: " + '.event.readings[].objectValue.Profiles[].Token''
   ```
   Example Output: 

   ```bash
   profileToken: profile_1
   profileToken: profile_2
   ```

2. Get the RTSP URI, from the ONVIF device, by executing the `GetStreamURI` command with the profileToken found in [step 1](#step1):  
   In this example, `profile_1` is the ProfileToken:  

   ```bash
      curl -s "http://0.0.0.0:59882/api/v2/device/name/Camera001/StreamUri?jsonObject=$(base64 -w 0 <<< '{
         "StreamSetup" : {
            "Stream" : "RTP-Unicast",
            "Transport" : {
               "Protocol" : "RTSP"
            }
         },
         "ProfileToken": "profile_1"
      }')" | jq -r '"streamURI: " + '.event.readings[].objectValue.MediaUri.Uri''
   ```
   
   Example Output:

   ```bash
   streamURI: rtsp://192.168.86.34:554/stream1
   ``` 

3. Stream the RTSP stream. 

   ffplay can be used to stream. The command follows this format: 
   
   `ffplay -rtsp_transport tcp "rtsp://<user>:<password>@<IP address>:<port>/<streamname>"`.

   Using the `streamURI` returned from the previous step, run ffplay:
   
   ```bash
   ffplay -rtsp_transport tcp "rtsp://admin:Password123@192.168.86.34:554/stream1"
   ```

   >NOTE: While the `streamURI` returned did not contain the username and password, those credentials are required in order to correctly authenticate the request and play the stream. Therefore, it is included in both the VLC and ffplay streaming examples.  
   >NOTE: If the password uses special characters, you must use percent-encoding.

5. To shut down ffplay, use the ctrl-c command.

## Shutting Down
To stop all EdgeX services (containers), execute the `make down` command. This will stop all services but not the images and volumes, which still exist.

1. Navigate to the `edgex-compose/compose-builder` directory.
1. Run this command
   ```bash
   make down
   ```
1. To shut down and delete all volumes, run this command
   ```bash
   make clean
   ```
   >NOTE: As this command deletes all volumes, you will need to rerun the [Add Device](#add-device) steps to re-enable your device(s). 

## Additional Configuration

Here is some information on how to specially configure parts of the service beyond the provided defaults.  

### Configure the Device Profiles

The device profile contains general information about the camera and includes all the device resources and commands that the device resources can use to manage the cameras. The default [profile](../../cmd/res/profiles/camera.yaml) contains all possible resources a camera could implement. Enable and disable supported resources in this file, or create an entirely new profile. It is important to set up the device profile to match the capabilities of the camera. Information on the resources supported by specific cameras can be found [here](../ONVIF-protocol.md#tested-onvif-cameras). Learn more about device profiles in EdgeX [here.](https://docs.edgexfoundry.org/1.2/microservices/device/profile/Ch-DeviceProfile/)

```yaml
name: "onvif-camera" # general information about the profile
manufacturer:  "Generic"
model: "Generic ONVIF"
labels:
  - "onvif"
description: "EdgeX device profile for ONVIF-compliant IP camera."

deviceResources:
  # Network Configuration
  - name: "Hostname" # an example of a resource with get/set values
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
<p align="left">
   <i>Sample: Snippet from camera.yaml</i>
</p>


### Configure the Provision Watchers

The provision watcher sets up parameters for EdgeX to automatically add devices to core-metadata. They can be configured to look for certain features, as well as block features. The default provision watcher is sufficient unless you plan on having multiple different cameras with different profiles and resources. Learn more about provision watchers [here](https://docs.edgexfoundry.org/2.2/microservices/core/metadata/Ch-Metadata/#provision-watcher).

```json
{
    "name":"Generic-Onvif-Provision-Watcher",
    "identifiers":{  // Use the identifiers to filter through specific features of the protocol
         "Address": ".",
         "Manufacturer": "Intel", // example of a feature to allow through 
         "Model": "DFI6256TE" 
    },
    "blockingIdentifiers":{
    },
    "serviceName": "device-onvif-camera",
    "profileName": "onvif-camera",
    "adminState":"UNLOCKED"
}
```
<p align="left">
   <i>Sample: Snippet from generic.provision.watcher.json</i>
</p>

## Summary and Next Steps
This guide demonstrated how to:

- deploy EdgeX with the ONVIF Device Service 
- use the EdgeX REST APIs and the ONVIF Device Service to acquire the camera's RTSP stream

### Next Steps

[Explore how to further use this device service](../general-usage.md)

Refer to the main [README](../../README.md) to find links to the rest of the documents.

## References

- ONVIF Website: http://www.onvif.org
- EdgeX Foundry Project Wiki: https://wiki.edgexfoundry.org/
- EdgeX Source Code: https://github.com/edgexfoundry
- Edgex Developer Guide: https://docs.edgexfoundry.org/2.1/
- Docker Repos
   - Docker https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository
   - Docker Compose https://docs.docker.com/compose/install/#install-compose

# License

[Apache-2.0](https://github.com/edgexfoundry-holding/device-onvif-camera/blob/main/LICENSE)
