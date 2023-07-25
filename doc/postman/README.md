# ONVIF Camera Device Postman Collections
This directory contains two sets of Postman collections which can be used to exercise an ONVIF camera device. To use either of these collections, simply import both the collection and the environment files into Postman.

### EdgeX ONVIF Collection
This collection utilizes the ONVIF device service to issue REST API's to the ONVIF camera device as facilitated by EdgeX.
- onvif_camera_with_edgex_postman_collection.json contains the set of EdgeX REST API's.
- onvif_camera_with_edgex_postman_environment.json contains the execution environment.

### Non-Edgex ONVIF Collection
This collection issues SOAP commands directly to the ONVIF camera device. EdgeX is not utilized here.
- onvif_camera_without_edgex_postman_collection.json contains the set of SOAP commands.
- onvif_camera_without_edgex_postman_environment.json file contains the exectuion environment.
