This code generates OpenAPI spec based on the Postman Collection 
[device-onvif-camera.postman_collection.json](../postman/device-onvif-camera.postman_collection.json). 
It also does some automated find and replace to insert some missing example values.

Usage:
- Install `postman-to-openapi` by running `make install` from this directory.
- Update the latest postman collection [device-onvif-camera.postman_collection.json](../postman/device-onvif-camera.postman_collection.json)
- Run `make gen` to re-generate the OpenAPI files.
