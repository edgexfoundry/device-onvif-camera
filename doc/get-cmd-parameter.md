# Get Command Parameter

For Get command, the client side should pass input parameter vai URL query parameter instead of request payload([see the discussion](https://github.com/edgexfoundry/edgex-go/issues/3754)).

Since device-onvif-camera treat the Onvif input parameter as JSON object, so we can encode the JSON data to **base64** string and pass to the URL query parameter.  

The device service will retrieve the base64 string from **jsonObject** parameter and decodes the base64 string to the JSON data as the Onvif input parameter.

## Pass One Parameter
For Example, the GetStreamUri require ProfileToken as input parameter,

<img src="images/get-streamuri-input-parameter.jpg"  width="80%"/>

then we can do the following steps:
1. Encode the `{ "ProfileToken": "Profile_1" }` json string to `eyAiUHJvZmlsZVRva2VuIjogIlByb2ZpbGVfMSIgfQ==` base64 string
2. Send the command with the base64 string:
   `curl http://0.0.0.0:59882/api/v2/device/name/Camera001/StreamUri?jsonObject=eyAiUHJvZmlsZVRva2VuIjogIlByb2ZpbGVfMSIgfQ==`


## Pass Multiple Parameters
If we need to pass multiple parameters like:
```json
{ 
   "ProfileToken": "Profile_1",
   "StreamSetup": {
      "Stream":"RTP-Multicast"
   }
}
```
then we can do the following steps:
1. Encode the json string to `eyAKICAgIlByb2ZpbGVUb2tlbiI6ICJQcm9maWxlXzEiLAogICAiU3RyZWFtU2V0dXAiOiB7CiAgICAgICJTdHJlYW0iOiJSVFAtTXVsdGljYXN0IgogICB9Cn0=`
2. Send the command with the base64 string:
   `curl http://0.0.0.0:59882/api/v2/device/name/Camera001/StreamUri?jsonObject=eyAKICAgIlByb2ZpbGVUb2tlbiI6ICJQcm9maWxlXzEiLAogICAiU3RyZWFtU2V0dXAiOiB7CiAgICAgICJTdHJlYW0iOiJSVFAtTXVsdGljYXN0IgogICB9Cn0=`


## Combine the Parameters for Different Resources
If we want to read Multiple resources which require different parameter, for example:
- The `GetMetadataConfiguration` function need `ConfigurationToken` parameter
   ```json
   { 
      "ConfigurationToken": "MetaDataToken" 
   }
   ```
- The `GetMetadataConfigurationOptions` function need `ProfileToken` parameter
   ```json
   { 
      "ProfileToken": "Profile_1" 
   }
   ```

Then we need to do the following steps:  
1. combine the parameter:
   ```json
   { 
      "ConfigurationToken": "MetaDataToken",
      "ProfileToken": "Profile_1"
   }
   ```
2. Encode the json string to `eyAKICAgICAgIkNvbmZpZ3VyYXRpb25Ub2tlbiI6ICJNZXRhRGF0YVRva2VuIiwKICAgICAgIlByb2ZpbGVUb2tlbiI6ICJQcm9maWxlXzEiCiAgIH0=`
3. Send the command with the base64 string:
   `curl http://0.0.0.0:59882/api/v2/device/name/Camera001/MetadataConfigurationAndOptions?jsonObject=eyAKICAgICAgIkNvbmZpZ3VyYXRpb25Ub2tlbiI6ICJNZXRhRGF0YVRva2VuIiwKICAgICAgIlByb2ZpbGVUb2tlbiI6ICJQcm9maWxlXzEiCiAgIH0=`
