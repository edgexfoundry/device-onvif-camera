# EdgeX Open Network Video Interface Forum (ONVIF) Device Service User Guide

## Overview
The Open Network Video Interface Forum (ONVIF) Device Service is a microservice created to address the lack of standardization and automation of camera discovery and onboarding. EdgeX Foundry is a flexible microservice-based architecture created to promote the interoperability of multiple device interface combinations at the edge. In an EdgeX deployment, the ONVIF Device Service controls and communicates with ONVIF-compliant cameras, while EdgeX Foundry presents a standard interface to application developers. With normalized connectivity protocols and a vendor-neutral architecture, EdgeX paired with ONVIF Camera Device Service, simplifies deployment of edge camera devices. 

This user guide describes how to:

- build, configure, and install the ONVIF Device Service and EdgeX
- configure EdgeX with an ONVIF camera
- acquire and view the configured camera's Real Time Streaming Protocol (RTSP) video stream

Use the ONVIF Device Service to streamline and scale your edge camera device deployment. 

### Contents
[System Requirements](#system-requirements)  
[Dependencies](#dependencies)  
[Tool Descriptions](#tool-descriptions)  
[Camera Setup](#camera-setup)  
[Get the Source](#get-the-source)  
[Configure and Build the ONVIF Device Service Docker Image](#configure-and-build-the-onvif-device-service-docker-image)  
[Deploy EdgeX and ONVIF Device Camera Microservice](#deploy-edgex-and-onvif-device-camera-microservice)  
[Verify Services, Devices, and Device Profiles](#verify-services-devices-and-device-profiles)  
[Execute GetStreamURI Command through EdgeX](#execute-getstreamuri-command-through-edgex)  
[Summary and Next Steps](#summary-and-next-steps)  
[References](#references)

## System Requirements

- Intel&#8482; Core&#174; processor
- Ubuntu 20.04.4 LTS
- ONVIF-compliant Camera

>NOTE: The instructions in this User Guide were developed and tested using Ubuntu 20.04 LTS and the Tapo C200 Pan/Tilt Wi-Fi Camera, referred to throughout this document as the **Tapo C200 Camera**. However, the software may work with other Linux distributions and ONVIF-compliant cameras.

**Time to Complete**

30-40 minutes

**Other Requirements**

You must have administrator (sudo) privileges to execute the user guide commands.

## How It Works
The figure below illustrates the software flow through the architecture components.

![high-level-arch](../images/ONVIFDeviceServiceArch.png)
<p align="left">
      <i>Figure 1: Software Flow</i>
</p>

1. **EdgeX Device Discovery:** Camera device microservices probe network and platform for video devices at a configurable interval. Devices that do not currently exist and that satisfy Provision Watcher filter criteria are added to Core Metadata.
2. **Application Device Discovery:** Query Core Metadata for devices and associated configuration.
3. **Application Device Configuration:** Set configuration and initiate device actions through a REST API representing the resources of the video device (e.g. stream URI, Pan-Tilt-Zoom position, Firmware Update).
4. **Pipeline Control:** The application initiates Video Analytics Pipeline through HTTP Post Request.
5. **Publish Inference Events/Data:** Analytics inferences are formatted and passed to the destination message bus specified in the request.
6.  **Export Data:** Publish prepared (transformed, enriched, filtered, etc.) and groomed (formatted, compressed, encrypted, etc.) data to external systems (be it analytics package, enterprise or on-premises application, cloud systems like Azure IoT, AWS IoT, or Google IoT Core, etc.

## Dependencies
The software has dependencies, including Git, Docker, Docker Compose, and assorted tools (e.g., curl). Follow the instructions below to install any dependency that is not already installed. 

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

1. Update installation repositories:
   ```bash
   sudo apt update
   ```

2. Install dependencies:
   ```bash
   sudo apt-get install ca-certificates curl gnupg lsb-release
   ```

3. Add Docker's GPG Key:
   ```bash
   curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
   ```

4. Update to point to the Stable Release:
   ```bash
   echo \
   "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \ 
   $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
   ```

5. Update the installation repositories:
   ```bash
   sudo apt-get update
   ```

6. Install Docker:
   ```bash
   sudo apt-get install docker-ce docker-ce-cli containerd.io
   ```

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
   >NOTE: When this User Guide was created, verison 1.29.2 was current.

2. Set permissions:
   ```bash
   sudo chmod +x /usr/local/bin/docker-compose
   ```


### Install Tools
Install the build, media streaming, and parsing tools:

```bash
sudo apt install build-essential vlc ffmpeg jq
```

## Tool Descriptions
The table below lists command line tools this guide uses to help with EdgeX configuration and device setup.

| Tool        | Description | Note |
| ----------- | ----------- |----------- |
| **curl**     | Allows the user to connect to services such as EdgeX. |Use curl to get transfer information either to or from this service. In the tutorial, use `curl` to communicate with the EdgeX API. The call will return a JSON object.|
| **jq**   |Parses the JSON object returned from the `curl` requests. |The `jq` command includes parameters that are used to parse and format data. In this tutorial, the `jq` command has been configured to return and format appropriate data for each `curl` command that is piped into it. |
| **base64**   | Converts data into the Base64 format.| |

Table 1: Command Line Tools

## Camera Setup

Follow the instructions included with the Tapo C200 Camera to set up and configure the camera.

The ONVIF service available on the Tapo C200 Camera requires a camera account. This is different from the Tapo account used to set up the camera. 

>NOTE: The location of the <b>Camera Account</b> setting may differ by application version.

To create the account in the Tapo Camera application:

1. Choose the **Gear** icon.
2. Choose **Camera Settings** > **Advanced Settings** > **Camera Account**.
3. Provide a **Username** and **Password**. 

>NOTE: After the Camera Account is created, note the username, password, and IP address, which will be used in a later step to configure the camera's credentials.

   ![alt text](https://static.tp-link.com/image003_1579056157342a.jpg)
   ![alt text](https://static.tp-link.com/image004_1579056164876m.jpg)
   ![alt text](https://static.tp-link.com/image005_1579056172449b.jpg)
<p align="left">
   <i>Figure 1: Tapo app settings</i>
</p>


### Verify Camera Operation
After the Tapo C200 Camera is set up and the Computer Account is created, the camera stream will be available for viewing. VLC Player can be used to view the camera's RTSP (Real Time Streaming Protocol) stream. 

To see the RTSP stream:

1. Run VLC Player.
2. Choose **File** > **Open Network Stream**:

   <p align="center">
      <img src="../images/vlcstream1.png" alt="NetworkVLC"><br>
   <i>Figure 2: VLC Open Network Stream</i>
   </p>

2. Provide network URI source using this format:

   `rtsp://<Username>:<Password>@<IPAddress>:554/stream1`
 
   Example:
 
   ` rtsp://admin:Pasword123@192.168.86.34:554/stream1`

   Replace these elements of the URI:
   - `Username`: username specified when creating the Camera Account.
   - `Password`: password specified when creating the Camera Account.
   - `IPAddress`: IP address of the Tapo C200 Camera found in **Camera Settings** in the Tapo Camera app.

   <p align="center">
      <img src="../images/vlcstream2.png" alt="NetworkVLC"><br>
   <i>Figure 3: VLC Enter Network URL</i>
   </p>

3. Click on the **Play** button so that VLC Player plays the RTSP stream from the camera.

   <p align="center">
      <img src="../images/vlcstream3.png" alt="NetworkVLC"><br>
   <i>Figure 4: VLC RTSP Stream</i>
   </p>

   >NOTE: If the camera is unable to play the RTSP stream, check your camera configuration. See [C200 RTSP Stream Setup](#references) in the references for more help.
## Get the Source

1. Clone the ONVIF Device Service repository:

   ```bash
   git clone https://github.com/edgexfoundry/device-onvif-camera.git
   ```

2. Clone the EdgeX Compose repository:

   ```bash
   git clone -b jakarta https://github.com/edgexfoundry/edgex-compose.git
   ```

## Configure and Build the ONVIF Device Service Docker Image

### Configure the Camera

>NOTE: Before configuring, make sure the username, password, and IP address for the camera are available, as described in [TapoC200 Camera Setup](#tapoc200-camera-setup)*

1. Make a copy of the `camera.toml.example`:  
   ```bash
   cp device-onvif-camera/cmd/res/devices/camera.toml.example device-onvif-camera/cmd/res/devices/camera.toml
   ```
   >NOTE: The `camera.toml` file will contain the camera definition.
2. Open the `camera.toml` file and update the `Address` and `Port` fields to match the IP address of the Tapo C200 Camera and port used for ONVIF services (default is 2020):

   ```bash
   nano device-onvif-camera/cmd/res/devices/camera.toml
   ```
   > NOTE: The application `nano` is used here since this editor is typically already installed, but any editor can be used.
   &*()
   ```toml
   # If having more than one camera, uncomment the following config settings
   [[DeviceList]]
   Name = "Camera001"                         # Modify as desired
   ProfileName = "onvif-camera"
   Description = "onvif conformant camera"    # Modify as desired
   [DeviceList.Protocols]
   [DeviceList.Protocols.Onvif]
   Address = "192.168.86.34"              # Set to your Tapo C200 IP address
   Port = "2020"                          # Set to 2020; default for Tapo C200
   # Assign AuthMode to "usernametoken" | "digest" | "both" | "none"
   AuthMode = "usernametoken"             # Set to 'usernametoken'
   SecretPath = "credentials001"
   ```
   <p align="left">
      <i>Sample: Snippet from camera.toml</i>
   </p>

3. Ensure the `AuthMode` is set to `"usernametoken"`.

4. Optionally, modify the `Name` and `Description` fields to more easily identify the Tapo C200 Camera. The `Name` is the camera name used when using ONVIF Device Service Rest APIs. The `Description` is simply a more detailed explanation of the camera.

### Set the Credentials &*()

1. Open `configuration.toml` file:

   ```bash
   nano device-onvif-camera/cmd/res/configuration.toml
   ```
2. Make sure `path` is set to match `SecretPath` in `camera.toml`. In the sample below, it is `"credentials001"`. 
3. Under `path`, set `username` and `password` to your Tapo C200 Camera Account credentials. 

```toml
[Writable]
LogLevel = "INFO"
  # Example InsecureSecrets configuration that simulates SecretStore for when EDGEX_SECURITY_SECRET_STORE=false
  # InsecureSecrets are required for when Redis is used for message bus
  [Writable.InsecureSecrets]
    [Writable.InsecureSecrets.DB]
    path = "redisdb"
      [Writable.InsecureSecrets.DB.Secrets]
      username = ""
      password = ""
    [Writable.InsecureSecrets.Camera001]
    path = "credentials001"
      [Writable.InsecureSecrets.Camera001.Secrets]
      username = "admin"                  # Set to your Tapo C200 Camera Account Username
      password = "Password123"            # Set to your Tapo C200 Camera Account Password
```

<p align="left">
   <i>Sample: Snippet from configuration.toml</i>
</p>

### Build the Docker Image

1. Navigate to the `device-onvif-camera` directory and run make:

   ```bash
   cd device-onvif-camera
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

3. Update `.env` file to add the registry and image version variable for device-onvif-camera:

   a. Open the `.env` file to edit:
      ```bash
      nano edgex-compose/compose-builder/.env
      ```

   b. Add the following registry and version information:
      ```
      DEVICE_ONVIFCAM_VERSION=0.0.0-dev
      ```

# Deploy EdgeX and ONVIF Device Camera Microservice

## Run the Service

1. Go to the EdgeX compose-builder directory:

   ```bash
   cd edgex-compose/compose-builder/
   ```

2. Generate a non-secure compose file:

   ```bash
   make gen no-secty ds-onvif-camera
   ```

3. Run EdgeX with the microservice:

   ```bash
   make run no-secty ds-onvif-camera
   ```

## Verify Services, Devices, and Device Profiles

1. Check the status of the container:

   ```bash 
   docker ps
   ```
   Example Output:

   The status column will indicate if the container is running, and how long it has been up.

   Example Output:

   ```docker
   CONTAINER ID   IMAGE                                         COMMAND                  CREATED       STATUS          PORTS                                                                                         NAMES
   33f9c5ecb70e   edgexfoundry/device-onvif-camera:0.0.0-dev    "/device-onvif-camer…"   7 weeks ago   Up 48 minutes   127.0.0.1:59985->59985/tcp                                                                    edgex-device-onvif-camera
   ```

2. Check whether the device service is added to EdgeX:

   ```bash
   curl http://localhost:59881/api/v2/deviceservice/name/device-onvif-camera | jq
   ```
   Example Output:

   ```bash
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
      100   250  100   250    0     0   1222k      0 --:--:-- --:--:-- --:--:--  122k
      {
         "apiVersion" : "v2",
         "service" : {
            "adminState" : "UNLOCKED",
            "baseAddress" : "http://edgex-device-onvif-camera:59985",
            "created" : 1644512266873,
            "id" : "9b7c952a-5fd6-40a3-a08d-51f97161debe",
            "modified" : 1648744261712,
            "name" : "device-onvif-camera"
      },
         "statusCode" : 200
      }
   ```

3. Check whether the ONVIF camera device and device-profile is added:

   ```bash
   curl -s http://localhost:59882/api/v2/device/name/Camera001 | jq -r '"deviceName: " + '.deviceCoreCommand.deviceName' + "\nstatusCode: " + (.statusCode|tostring)'
   ```
   Example Output:

   ```bash
   deviceName: Camera001
   statusCode: 200
   ```

4. Check that all cameras are added to EdgeX:

   ```bash
   curl -s http://localhost:59881/api/v2/device/all | jq -r '"deviceName: " + '.deviceCoreCommands[].deviceName''
   ```
   Example Output: 

   ```bash
   deviceName: Camera002
   deviceName: Camera001
   ```

### Use EdgeX Console to Verify Device Services, Devices, and Device Profiles
1. Visit http://localhost:4000 to go to the dashboard for EdgeX Console GUI:

   ![EdgeXConsoleDashboard](../images/EdgeXDashboard.png)
   <p align="left">
      <i>Figure 5: EdgeX Console Dashboard</i>
   </p>

2. To see **Device Services**, **Devices**, or **Device Profiles**, click on their respective tab:

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
   curl http://0.0.0.0:59882/api/v2/device/name/Camera001/Profiles | jq -r '"profileToken: " + '.event.readings[].objectValue.Profiles[].Token''
   ```
   Example Output: 

   ```bash
     % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
     100  5151    0  5151    0     0  55989      0 --:--:-- --:--:-- --:--:-- 55989
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

3. Execute `GetStreamURI` command to get RTSP URI from the ONVIF device. Make sure to put the Base64 JSON data after *?jsonObject=* in the command.

   ```bash
   curl  http://0.0.0.0:59882/api/v2/device/name/Camera001/StreamUri?jsonObject=ewogICAgICAiUHJvZmlsZVRva2VuIjogInByb2ZpbGVfMSIKfQ== | jq -r '"streamURI: " + '.event.readings[].objectValue.MediaUri.Uri''
   ```
   
   Example Output:

   ```bash
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
   100   553  100   553    0     0  21269      0 --:--:-- --:--:-- --:--:-- 21269
   streamURI: rtsp://192.168.86.34:554/stream1
   ``` 

4. Stream the RTSP stream:

   VLC can be used as shown in the [Verify Camera Operation](#verify-camera-operation) section.

   Alternatively, ffplay can be used to stream. The command follows this format: 
   
   `ffplay -rtsp_transport tcp rtsp://<user>:<password>@<IP address>:<port>/<streamname>`.

   Using the `streamURI` retuned from the previous step, run ffplay:
   
   ```bash
   ffplay -rtsp_transport tcp rtsp://admin:Password123@192.168.86.34:554/stream1
   ```
   >NOTE: While the `streamURI` returned did not contain the username and password, those credentials are required in order to correctly authenticate the request and play the stream. Therefore, it is included in both the VLC and ffplay streaming examples.

### Shutting Down
To stop all EdgeX services (containers), execute the `make down` command. This will stop all services but not the images and volumes, which still exist.

```bash
make down
```

## Summary and Next Steps
This user guide demonstrated how to:

- configure the Tapo C200 Camera, the ONVIF Device Service, and EdgeX
- deploy EdgeX with the ONVIF Device Service 
- use the EdgeX REST APIs and the ONVIF Device Service to acquire the camera's RTSP stream

## References

- ONVIF Website: http://www.onvif.org
- EdgeX Foundry Project Wiki: https://wiki.edgexfoundry.org/
- EdgeX Source Code: https://github.com/edgexfoundry
- Edgex Developer Guide: https://docs.edgexfoundry.org/2.1/
- Tapo C200 RTSP Stream Setup: https://www.tapo.com/us/faq/34/
- Docker Repos
   - Docker https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository
   - Docker Compose https://docs.docker.com/compose/install/#install-compose

# License

[Apache-2.0](https://github.com/edgexfoundry-holding/device-onvif-camera/blob/main/LICENSE)
