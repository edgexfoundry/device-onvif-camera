# OpenAPI / Swagger Spec Files
This code generates OpenAPI 3.0 spec based on the [Postman Collection](../postman/device-onvif-camera.postman_collection.json), 
[Onvif WSDL Schema Files](./ref), [sidecar.yaml](sidecar.yaml), [default camera device profile](../../cmd/res/profiles/camera.yaml),
[ONVIF Tested Cameras Matrix](../ONVIF-protocol.md#tested-onvif-cameras), and [ONVIF footnotes](../onvif-footnotes.md).

## Usage
- Install `postman-to-openapi` and python3 dependencies by running `make install` from this directory.
- Import the latest [postman collection][collection] and latest [postman environment][env] into Postman App
- Modify the postman collection and/or postman environment files
- Export the modified postman collection and environment
- Overwrite the postman collection and environment in this repo with the exported files
- Run `make gen` to re-generate the OpenAPI files.
  - Use `make -B gen` to force it to no use any cached files
  - Use `DEBUG_LOGGING=1 make ...` to enable debug logging of the python scripts

[collection]: ../postman/device-onvif-camera.postman_collection.json
[env]: ../postman/device-onvif-camera.postman_environment.json

## Explanation
### [python](python) folder
Contains scripts for processing the input and output files when
generating the OpenAPI spec file.

#### [xmlstrip.py](python/xmlstrip.py)
This script cleans up the yaml files in the [ref](ref) folder by removing all the xml 
references in the schema, and tweaking the schema values to clean them up for use in 
the [postprocess.py](python/postprocess.py) script.

#### [postprocess.py](python/postprocess.py)
This script takes in the preliminary OpenAPI file that was generated from the Postman collection
and adds additional metadata to it, as well as cleaning up the data and format.

#### [xmlstrip.py](python/xmlstrip.py)
This script cleans up the yaml files in the [ref](ref) folder by removing all the xml
references in the schema, and tweaking the schema values to clean them up for use in
the [postprocess.py](python/postprocess.py) script.

#### [matrix.py](python/matrix.py)
This script adds the compatibility matrix to each endpoint

#### [cleaner.py](python/cleaner.py)
This script removes all unused schema definition files

### [sidecar.yaml](sidecar.yaml)
Contains additional canned responses, template data, and schemas to include in the final OpenAPI spec file.

### [ref](ref) folder
This folder contains files generated from the official Onvif wsdl
spec files. The wsdl files were imported into [apimatic.io](https://apimatic.io)
and exported as OpenAPI 3.0 spec yaml files.

### [v2](v2) folder
This folder contains the final exported OpenAPI 3.0 spec file, to be
exported to SwaggerHub.
