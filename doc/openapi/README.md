# OpenAPI / Swagger Spec Files
This code generates OpenAPI 3.0 spec based on the [Postman Collection](../postman/device-onvif-camera.postman_collection.json), 
Onvif WSDL Schema Files, and [sidecar.yaml](sidecar.yaml).

Usage:
- Install `postman-to-openapi` and python3 dependencies by running `make install` from this directory.
- Update the latest postman collection [device-onvif-camera.postman_collection.json](../postman/device-onvif-camera.postman_collection.json)
- Update the latest postman environment [device-onvif-camera.postman_environment.json](../postman/device-onvif-camera.postman_environment.json)
- Run `make gen` to re-generate the OpenAPI files.

## [python](python) folder
Contains scripts for processing the input and output files when
generating the OpenAPI spec file.

### [xmlstrip.py](python/xmlstrip.py)
This script cleans up the yaml files in the [ref](ref) folder by removing all the xml 
references in the schema, and tweaking the schema values to clean them up for use in 
the [postprocess.py](python/postprocess.py) script.

### [postprocess.py](python/postprocess.py)
This script takes in the preliminary OpenAPI file that was generated from the Postman collection
and adds additional metadata to it, as well as cleaning up the data and format.

### [xmlstrip.py](python/xmlstrip.py)
This script cleans up the yaml files in the [ref](ref) folder by removing all the xml
references in the schema, and tweaking the schema values to clean them up for use in
the [postprocess.py](python/postprocess.py) script.

### [matrix.py](python/matrix.py)
This script adds the compatibility matrix to each endpoint

### [cleaner.py](python/cleaner.py)
This script removes all unused schema definition files

## [ref](ref) folder
This folder contains files generated from the official Onvif wsdl
spec files. The wsdl files were imported into [apimatic.io](https://apimatic.io)
and exported as OpenAPI 3.0 spec yaml files.

## [v2](v2) folder
This folder contains the final exported OpenAPI 3.0 spec file, to be
exported to SwaggerHub.

