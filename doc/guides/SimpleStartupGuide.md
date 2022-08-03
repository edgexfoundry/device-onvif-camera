# ONVIF Device Service Simple Start Up Guide

## Contents

[System Requirements](#system-requirements)  
[Dependencies](#dependencies)  
[Deploy the Service](#deploy-edgex-and-onvif-device-camera-microservice)  
[Verify the Service](#verify-service-and-device-profiles)   
[Manage Devices](#manage-devices)  
[Execute Example Command](#execute-getstreamuri-command-through-edgex)  
[Shutting Down](#shutting-down)  

## System Requirements

- Intel&#8482; Core&#174; processor
- Ubuntu 20.04.4 LTS
- ONVIF-compliant Camera

>NOTE: The instructions in this guide were developed and tested using Ubuntu 20.04 LTS and the Tapo C200 Pan/Tilt Wi-Fi Camera, referred to throughout this document as the **Tapo C200 Camera**. However, the software may work with other Linux distributions and ONVIF-compliant cameras. Refer to our [list of tested cameras for more information](../ONVIF-protocol.md#tested-onvif-cameras)

**Time to Complete**

10-20 minutes

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

Clone the device-onvif-camera repository:

   ```bash
   git clone https://github.com/edgexfoundry/device-onvif-camera.git
   ```

## Deploy EdgeX and ONVIF Device Camera Microservice

### Run the Service

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

1. Check the status of the container:

   ```bash 
   docker ps
   ```

   The status column will indicate if the container is running and how long it has been up.

   Example Output:

   ```docker
   CONTAINER ID   IMAGE                                         COMMAND                  CREATED       STATUS          PORTS                                                                                         NAMES
   33f9c5ecb70e   edgexfoundry/device-onvif-camera:0.0.0-dev    "/device-onvif-camer…"   7 weeks ago   Up 48 minutes   127.0.0.1:59985->59985/tcp                                                                    edgex-device-onvif-camera
   ```

2. Check that the device service is added to EdgeX:

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


3. Check that the device profile is added:

   ```bash
   curl -s http://localhost:59881/api/v2/deviceprofile/name/onvif-camera | jq -r '"profileName: " + '.profile.name' + "\nstatusCode: " + (.statusCode|tostring)'

   ```
   Successful:
   ```bash
   profileName: onvif-camera
   statusCode: 200
   ```
   Unsuccessful:
   ```bash
   profileName: 
   statusCode: 404
   ```
   > NOTE: The `jq -r` option is used to reduce the size of the displayed response. The entire device profile with all resources can be seen by removing `-r '"profileName: " + '.profile.name' + "\nstatusCode: " + (.statusCode|tostring)', and replacing it with '.'`

## Manage Devices
Follow these instructions to update devices.

### Curl Commands

#### Add Device

1. Edit the information to appropriately match the camera. The fields `Address`, and `Port` should match that of the camera:

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
                        "Address": "x.x.x.x",
                        "Port": "10000",
                        "AuthMode": "usernametoken",
                        "SecretPath": "credentials001"
                     },
                     "CustomMetadata": {
                        "CommonName":"Default Camera",
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

1. Map credentials using the `map-credentials.sh` script.  
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

2. Verify device(s) have been successfully added to core-metadata:

   ```bash
   curl -s http://localhost:59881/api/v2/device/all | jq -r '"deviceName: " + '.devices[].name''
   ```

   Example Output: 
   ```bash
   deviceName: Camera001
   deviceName: device-onvif-camera
   ```
   >NOTE: The device with name `device-onvif-camera` is a stand-in device and can be ignored.  
   >NOTE: The `jq -r` option is used in the curl command to reduce the size of the displayed response. The entire device with all information can be seen by removing `-r '"deviceName: " + '.devices[].name'', and replacing it with '.'`

#### Update Device

There are multiple commands that can update aspects of the camera entry in meta-data. Refer to the [Swagger documentation]() for more information (not implemented).

#### Delete Device

   ```bash
   curl -X 'DELETE' \
   'http://localhost:59881/api/v2/device/name/<device name>' \
   -H 'accept: application/json' 
   ```

### Use EdgeX Console to Verify Device Services, Devices, and Device Profiles
1. Visit http://localhost:4000 to go to the dashboard for EdgeX Console GUI:

   ![EdgeXConsoleDashboard](../images/EdgeXDashboard.png)
   <p align="left">
      <i>Figure 5: EdgeX Console Dashboard</i>
   </p>

2. To get device status information, click on the tabs **Device Services**, **Devices**, or **Device Profiles**:

   ![EdgeXConsoleDeviceServices](../images/EdgeXDeviceServices.png)
   <p align="left">
      <i>Figure 6: EdgeX Console Device Service List</i>
   </p>

   ![EdgeXConsoleDeviceList](../images/EdgeXDeviceList.png)
   <p align="left">
      <i>Figure 7: EdgeX Console Device List</i>
   </p>

   ![EdgeXConsoleDeviceProfileList](../images/EdgeXDeviceProfiles.png)
   <p align="left">
      <i>Figure 8: EdgeX Console Device Profile List</i>
   </p>

## Execute GetStreamURI Command through EdgeX

1. Get the profile token by executing the `GetProfiles` command:

   ```bash
   curl -s http://0.0.0.0:59882/api/v2/device/name/Camera001/Profiles | jq -r '"profileToken: " + '.event.readings[].objectValue.Profiles[].Token''
   ```
   Example Output: 

   ```bash
   profileToken: profile_1
   profileToken: profile_2
   ```

2. Convert the JSON input to Base64:

   >NOTE: Make sure to change the profile token to the one found in step 1. In this example, it is the string `profile_1`.

   ```json
   {
      "ProfileToken": "profile_1"
   }
   ```
   Example Output:

   ```bash
   echo -n '{
      "ProfileToken": "profile_1"
   }' | base64
   ewogICAgICAiUHJvZmlsZVRva2VuIjogInByb2ZpbGVfMSIKfQ==
   ```

3. Execute the `GetStreamURI` command to get RTSP URI from the ONVIF device. Make sure to put the Base64 JSON data after *?jsonObject=* in the command.

   ```bash
   curl -s http://0.0.0.0:59882/api/v2/device/name/Camera001/StreamUri?jsonObject=ewogICAgICAiUHJvZmlsZVRva2VuIjogInByb2ZpbGVfMSIKfQ== | jq -r '"streamURI: " + '.event.readings[].objectValue.MediaUri.Uri''
   ```
   
   Example Output:

   ```bash
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
   100   553  100   553    0     0  21269      0 --:--:-- --:--:-- --:--:-- 21269
   streamURI: rtsp://192.168.86.34:554/stream1
   ``` 

4. Stream the RTSP stream. 

   ffplay can be used to stream. The command follows this format: 
   
   `ffplay -rtsp_transport tcp rtsp://'<user>':'<password>'@<IP address>:<port>/<streamname>`.

   Using the `streamURI` returned from the previous step, run ffplay:
   
   ```bash
   ffplay -rtsp_transport tcp rtsp://'admin':'Password123'@192.168.86.34:554/stream1
   ```
   >NOTE: While the `streamURI` returned did not contain the username and password, those credentials are required in order to correctly authenticate the request and play the stream. Therefore, it is included in both the VLC and ffplay streaming examples.  
   >NOTE: If the password uses special characters, you must use percent-encoding. Use the resources found here to find the special character conversions. We do not recommend entering your full password to the site.

5. To shut down ffplay, use the ctrl-c command.

## Shutting Down
To stop all EdgeX services (containers), execute the `make down` command:

1. Navigate to the `edgex-compose/compose-builder` directory.
1. Run this command
   ```bash
   make down
   ```
1. To shut down and delete all volumes, run this command
   ```bash
   make clean
   ```

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
