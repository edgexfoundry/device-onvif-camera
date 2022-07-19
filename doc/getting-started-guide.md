# Getting Started Guide

## 1. Deploy the EdgeX
Run core-data, core-metadata, core-command service according to the EdgeX doc
https://docs.edgexfoundry.org/2.2/getting-started/Ch-GettingStartedGoDevelopers/.

## 2. Deploy the device-onvif-camera
Build the executable file with the following command:
```shell
make build
```

Run the executable file:
```shell
export EDGEX_SECURITY_SECRET_STORE=false
./cmd/device-onvif-camera -c ./cmd/res/
```

Check the device service is added to EdgeX
```shell
curl http://localhost:59881/api/v2/deviceservice/name/device-onvif-camera | jq .
{
   "apiVersion" : "v2",
   "service" : {
      "adminState" : "UNLOCKED",
      "baseAddress" : "http://localhost:59984",
      "created" : 1635740039367,
      "id" : "dd79a417-6bbf-4a66-b9ab-d59cb9f3b324",
      "modified" : 1635740039367,
      "name" : "device-onvif-camera"
   },
   "statusCode" : 200
}
```

## 3. Add the device profile to EdgeX
Add the device profile to core-metadata service with the following command:
```shell
curl http://localhost:59881/api/v2/deviceprofile/uploadfile \
  -F "file=@./cmd/res/profiles/camera.yaml"
```

## 4. Add the device to EdgeX
Add the device data to core-metadata service with the following command:
```shell
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
                "operatingState": "UNKNOWN",
                "protocols": {
                    "Onvif": {
                        "Address": "192.168.12.123",
                        "Port": "80",
                        "AuthMode": "usernametoken",
                        "SecretPath": "credentials001"
                    }
                }
            }
          }
  ]'
```

Check the available commands from core-command service:
```shell
$ curl http://localhost:59882/api/v2/device/name/Camera001 | jq .
{
   "apiVersion" : "v2",
   "deviceCoreCommand" : {
      "coreCommands" : [
         {
            "get" : true,
            "set" : true,
            "name" : "DNS",
            "parameters" : [
               {
                  "resourceName" : "DNS",
                  "valueType" : "Object"
               }
            ],
            "path" : "/api/v2/device/name/Camera001/DNS",
            "url" : "http://0.0.0.0:59882"
         },
         ...
         {
            "get" : true,
            "name" : "StreamUri",
            "parameters" : [
               {
                  "resourceName" : "StreamUri",
                  "valueType" : "Object"
               }
            ],
            "path" : "/api/v2/device/name/Camera001/StreamUri",
            "url" : "http://0.0.0.0:59882"
         }
      ],
      "deviceName" : "Camera001",
      "profileName" : "onvif-camera"
   },
   "statusCode" : 200
}
```

## 5. Execute a Get Command - Read Single Resource
```shell
$ curl http://0.0.0.0:59882/api/v2/device/name/Camera001/Hostname | jq .
{
   "apiVersion" : "v2",
   "event" : {
      "apiVersion" : "v2",
      "deviceName" : "Camera001",
      "id" : "6b46d058-d8e0-4095-ba80-4a6de1787510",
      "origin" : 1635749209227019000,
      "profileName" : "onvif-camera",
      "readings" : [
         {
            "deviceName" : "Camera001",
            "id" : "a1b0d809-c88a-4889-920e-8ac64e6aa658",
            "objectValue" : {
               "HostnameInformation" : {
                  "FromDHCP" : false,
                  "Name" : "localhost"
               }
            },
            "origin" : 1635749209227019000,
            "profileName" : "onvif-camera",
            "resourceName" : "Hostname",
            "valueType" : "Object"
         }
      ],
      "sourceName" : "Hostname"
   },
   "statusCode" : 200
}
```
## 6. Execute a Get Command - Read Multiple Resources
```shell
$ curl http://0.0.0.0:59882/api/v2/device/name/Camera001/NetworkConfiguration | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  1942  100  1942    0     0  10167      0 --:--:-- --:--:-- --:--:-- 10167
{
   "apiVersion" : "v2",
   "event" : {
      "apiVersion" : "v2",
      "sourceName" : "NetworkConfiguration",
      "deviceName" : "Camera001",
      "id" : "24d5e391-0dcd-48f5-8706-6abb11797d29",
      "origin" : 1635868623002677000,
      "profileName" : "onvif-camera",
      "readings" : [
         {
            "deviceName" : "Camera001",
            "id" : "87d0bcfd-aecf-4ab7-a871-2b85a3c90f00",
            "objectValue" : {
               "HostnameInformation" : {
                  "FromDHCP" : false,
                  "Name" : "localhost"
               }
            },
            "origin" : 1635868623002677000,
            "profileName" : "onvif-camera",
            "resourceName" : "Hostname",
            "valueType" : "Object"
         },
         {
            "deviceName" : "Camera001",
            "id" : "edfa8d6f-a96e-49a8-96c9-595905cbe170",
            "objectValue" : {
               "DNSInformation" : {
                  "DNSManual" : {
                     "IPv4Address" : "192.168.12.1",
                     "Type" : "IPv4"
                  },
                  "FromDHCP" : false
               }
            },
            "origin" : 1635868623002677000,
            "profileName" : "onvif-camera",
            "resourceName" : "DNS",
            "valueType" : "Object"
         },
         ...
      ]
   },
   "statusCode" : 200
}
```

## 7. Execute a Set Command - Write Single Resource
```shell
curl -X PUT -H 'Content-Type: application/json' 'http://0.0.0.0:59882/api/v2/device/name/Camera001/Hostname' \
    -d '{
        "Hostname": {
            "Name": "localhost555"
        }
    }'
```

## 8. Execute a Set Command - Write Multiple Resource
```shell
curl -X PUT -H 'Content-Type: application/json' 'http://0.0.0.0:59882/api/v2/device/name/Camera001/NetworkConfiguration' \
    -d '{
        "Hostname": {
            "Name": "localhost"
        },
        "DNS": {
            "FromDHCP": false,
            "DNSManual": {
                "Type": "IPv4",
                "IPv4Address": "192.168.12.1"
            }
        },
        "NetworkInterfaces": {
            "InterfaceToken": "eth0",
            "NetworkInterface": {
                "Enabled": true,
                "IPv4": {
                    "DHCP": false
                }
            }
            
        },
        "NetworkProtocols": {
            "NetworkProtocols": [ 
                {
                    "Name": "HTTP",
                    "Enabled": true,
                    "Port": 80
                }
            ]
        },
        "NetworkDefaultGateway": {
            "IPv4Address": "192.168.12.1"
        }
    }'
```
