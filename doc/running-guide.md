# ONVIF Device Service Running Guide

## Table of Contents

[Deploy EdgeX and ONVIF Device Service](#deploy-edgex-and-onvif-device-service)  
[Verify Service and Device Profiles](#verify-service-and-device-profiles)  
[Add Device](#add-device)  
[Using EdgeX UI](#using-edgex-ui)  
[Manage Devices](#manage-devices)  
[Next Steps](#next-steps)  


## Deploy EdgeX and ONVIF Device Camera Microservice

### Run the Service

1. Go to the EdgeX compose-builder directory:

   ```bash
   cd edgex-compose/compose-builder/
   ```

1. Run EdgeX with the microservice:

   ```bash
   make run no-secty ds-onvif-camera
   ```

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
   33f9c5ecb70e   nexus3.edgexfoundry.org:10004/device-onvif-camera:latest    "/device-onvif-camer…"   7 weeks ago   Up 48 minutes   127.0.0.1:59985->59985/tcp                                                                    edgex-device-onvif-camera
   ```

2. Check whether the device service is added to EdgeX:

   ```bash
   curl -s http://localhost:59881/api/v2/deviceservice/name/device-onvif-camera | jq
   ```
   Good response:
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
   Bad response:
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
   > note: `jq -r` is used to reduce the size of the displayed response. The entire device profile with all resources can be seen by removing `-r '"profileName: " + '.profile.name' + "\nstatusCode: " + (.statusCode|tostring)'`

### Using EdgeX UI
1. Visit http://localhost:4000 to go to the dashboard for EdgeX Console GUI:

   ![EdgeXConsoleDashboard](./images/EdgeXDashboard.png)
   <p align="left">
      <i>Figure 1: EdgeX Console Dashboard</i>
   </p>

2. To see **Device Services**, **Devices**, or **Device Profiles**, click on their respective tab:

   ![EdgeXConsoleDeviceServices](./images/EdgeXDeviceServices.png)
   <p align="left">
      <i>Figure 2: EdgeX Console Device Service List</i>
   </p>

   ![EdgeXConsoleDeviceList](./images/EdgeXDeviceList.png)
   <p align="left">
      <i>Figure 3: EdgeX Console Device List</i>
   </p>

   ![EdgeXConsoleDeviceProfileList](./images/EdgeXDeviceProfiles.png)
   <p align="left">
      <i>Figure 4: EdgeX Console Device Profile List</i>
   </p>
## Manage Devices
Follow these instructions to update devices.

<!-- Do we want this?
### EdgeX Console

#### Add Device

#### Update Device

#### Delete Device -->

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

1. Add credentials using the `set-credentials.sh` script.  
   1. Navigate to the `device-onvif-camera/bin` directory.
   1. Run the `set-credentials.sh` script

      ```bash
      ./set-credentials.sh
      ```
   1. Select the device(s) you want to set the credentials for.

      ![creds-select-devices](./images/set-credentials-start.png)
      <p align="left">
         <i>Figure 5: Select device for set-credentials.sh</i>
      </p>

   1. Set the username for the device(s).

      ![creds-set-username](./images/set-credentials-username.png)
      <p align="left">
         <i>Figure 6: Select username for devices</i>
      </p>

   1. Set the password for the device(s).

      ![creds-set-password](./images/set-credentials-password.png)
      <p align="left">
         <i>Figure 7: Select password for devices</i>
      </p>

   1. Set the authmode for the device(s).

      ![creds-set-authmode](./images/set-credentials-authmode.png)
      <p align="left">
         <i>Figure 8: Select authmode for devices</i>
      </p>

   Good output:
   
   ```bash
   curl --data "<redacted>" -X GET http://localhost:59881/api/v2/device/service/name/device-onvif-camera
   Response [200] 


   Selected Device: Camera001
   Setting InsecureSecret: Camera001/Path
   curl --data "<redacted>" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/Camera001/Path
   Response [200] true

   Setting InsecureSecret: Camera001/Secrets/username
   curl --data "<redacted>" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/Camera001/Secrets/username
   Response [200] true

   Setting InsecureSecret: Camera001/Secrets/password
   curl --data "<redacted>" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/Camera001/Secrets/password
   Response [200] true

   Setting InsecureSecret: Camera001/Secrets/mode
   curl --data "<redacted>" -X PUT http://localhost:8500/v1/kv/edgex/devices/2.0/device-onvif-camera/Writable/InsecureSecrets/Camera001/Secrets/mode
   Response [200] true

   Patching protocols["Onvif"].SecretPath to Camera001
   curl --data "<redacted>" -X GET http://localhost:59881/api/v2/device/name/Camera001
   Response [200] curl --data "<redacted>" -X PATCH http://localhost:59881/api/v2/device
   Response [207] [
      {
         "apiVersion": "v2",
         "statusCode": 200
      }
   ]
   ```

1. Verify device(s) have been succesfully added to core-metadata.

   ```bash
   curl -s http://localhost:59881/api/v2/device/all | jq -r '"deviceName: " + '.devices[].name''
   ```

   Example Output: 
   ```bash
   deviceName: Camera001
   deviceName: device-onvif-camera
   ```
   >note: device with name `device-onvif-camera` is a stand-in device and can be ignored.  
   >note: `jq -r` is used to reduce the size of the displayed response. The entire device with all information can be seen by removing `-r '"deviceName: " + '.devices[].name''`

#### Update Device

   There are multiple commands that can update aspects of the camera entry in meta-data. Refer to the [Swagger documentation]() for more information (not implemented).

#### Delete Device

   ```bash
   curl -X 'DELETE' \
   'http://localhost:59881/api/v2/device/name/<device name>' \
   -H 'accept: application/json' 
   ```

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

## Next Steps

[Learn how to use the device service](./general-usage.md)

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
