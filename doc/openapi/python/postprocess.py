#!/usr/bin/env python3
#
# Copyright (C) 2022-2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0
#

import base64
import json
from dataclasses import dataclass, field
import sys
import copy
import textwrap
from typing import Dict
import logging
import os

from ruamel.yaml import YAML
from ruamel.yaml.scalarstring import LiteralScalarString

from cleaner import SchemaCleaner
from matrix import MarkdownMatrix

yaml = YAML()
log = logging.getLogger('postprocess')

EDGEX = 'EdgeX'
EDGEX_DEVICE_NAME = 'Camera001'
API_PREFIX = '/api/v3/device/name/{EDGEX_DEVICE_NAME}'
SNAPSHOT = 'Snapshot'
SNAPSHOT_URI_FN = 'GetSnapshotUri'
MEDIA = 'Media'

# mapping of service name to wsdl file for externalDocs
SERVICE_WSDL = {
    'Analytics': 'https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl',
    'Device': 'https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl',
    'Event': 'https://www.onvif.org/ver10/events/wsdl/event.wsdl',
    'Imaging': 'https://www.onvif.org/ver20/imaging/wsdl/imaging.wsdl',
    'Media': 'https://www.onvif.org/ver10/media/wsdl/media.wsdl',
    'Media2': 'https://www.onvif.org/ver20/media/wsdl/media.wsdl',
    'PTZ': 'https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl'
}

# list of superfluous headers to remove from response objects
HEADERS_TO_REMOVE = [
    'Content-Length',
    'Date',
    'Transfer-Encoding'
]


class ProcessingError(RuntimeError):
    pass


def multiline_string(val: str):
    """Takes a string value and wraps it so that ruamel.yaml will format it as a raw multi-line string"""
    return LiteralScalarString(textwrap.dedent(val))


def single_quote(s: str):
    """Returns the input value wrapped in single quote marks"""
    return f"'{s}'"


@dataclass
class YamlProcessor:
    input_file: str
    sidecar_file: str
    profile_file: str
    output_file: str
    matrix: MarkdownMatrix
    postman_env_file: str

    yml: any = None
    sidecar: any = None
    profile: any = None
    postman_env: Dict[str, str] = field(default_factory=dict)
    resources: Dict[str, any] = field(default_factory=dict)
    wsdl_files: Dict[str, any] = field(default_factory=dict)

    def _load(self):
        """Read input yaml file and sidecar yaml files"""
        log.info(f'Reading input OpenAPI file: {self.input_file}')
        with open(self.input_file) as f:
            self.yml = yaml.load(f)

        log.info(f'Reading sidecar file: {self.sidecar_file}')
        with open(self.sidecar_file) as f:
            self.sidecar = yaml.load(f)

        log.info(f'Reading profile file: {self.profile_file}')
        with open(self.profile_file) as f:
            self.profile = yaml.load(f)

        log.info(f'Loading validation matrix file: {self.matrix.tested_file}')
        log.info(f'Loading footnotes file: {self.matrix.footnotes_file}')
        self.matrix.parse()

        log.info(f'Loading postman env file: {self.postman_env_file}')
        with open(self.postman_env_file) as f:
            env = json.load(f)
            for item in env['values']:
                self.postman_env[item['key']] = item['value']

    def _parse(self):
        """Parse the device resources into a lookup table"""
        for resource in self.profile['deviceResources']:
            self.resources[resource['name']] = resource

    def _write(self):
        """Output modified yaml file"""
        log.info(f'Writing output OpenAPI file: {self.output_file}')
        with open(self.output_file, 'w') as w:
            yaml.dump(self.yml, w)

    def _process_apis(self):
        """
        - Side-load externalDocs using EdgeX profile file
        - update descriptions
        - set schemas
        - add response codes
        etc...
        """

        for path, path_obj in self.yml['paths'].items():
            for method, method_obj in path_obj.items():
                cmd = path.split('/')[-1]

                prefix = 'set' if method == 'put' else method
                if cmd in self.resources:
                    attrs = self.resources[cmd]['attributes']
                    fn = attrs[f'{prefix}Function']
                    service = attrs['service']
                    service_fn = f'{service}_{fn}'

                    # set the unique operationId. this also will match the schema name of the command
                    method_obj['operationId'] = f'{service.lower()}_{fn}'

                    # add all responses
                    for code, resp_obj in self.sidecar['responses']['canned'].items():
                        if code not in method_obj['responses']:
                            method_obj['responses'][code] = resp_obj
                        elif code == '200' and method == 'put':
                            content = method_obj['responses']['200']['content']
                            if 'application/json' not in content or \
                                    len(content['application/json']) == 0 or \
                                    ('example' in content['application/json'] and len(
                                        content['application/json']['example']) == 2):
                                log.debug(f'Overriding empty 200 response for {service_fn}')
                                method_obj['responses'][code] = resp_obj

                    if service == EDGEX:
                        # --- Custom EdgeX function patching ---
                        if method == 'get':
                            if cmd in self.sidecar['responses']['edgex']:
                                # clone the 200 response to avoid mangling pointer references
                                resp_200 = copy.deepcopy(method_obj['responses']['200'])
                                # apply the defined schema
                                resp_200['content']['application/json']['schema'] = self.sidecar['responses']['edgex'][cmd]
                                # override with cloned one
                                method_obj['responses']['200'] = resp_200
                            else:
                                log.warning(f'\033[33m*** Missing schema response definition for EdgeX command {method.upper()} {cmd} ***\033[0m')
                        elif method == 'put':
                            if cmd in self.sidecar['requests']['edgex']:
                                # look for the json response object, so we can modify it
                                jscontent = method_obj['requestBody']['content']['application/json']

                                # move the example outside the schema to preserve it (and it belongs better up there)
                                if 'example' in jscontent['schema']:
                                    jscontent['example'] = jscontent['schema']['example']

                                # patch PUT call schema by using service name and onvif function name
                                jscontent['schema'] = {
                                    'properties': {
                                        cmd: self.sidecar['requests']['edgex'][cmd]
                                    },
                                    'required': [cmd],
                                    'type': 'object'
                                }
                            else:
                                log.warning(f'\033[33m*** Missing schema request definition for EdgeX command {method.upper()} {cmd} ***\033[0m')

                            # override the response schema with default 200 response
                            method_obj['responses']['200'] = self.sidecar['responses']['canned']['200']

                    else:
                        # --- ONVIF function patching ---
                        method_obj = path_obj[method]
                        method_obj['externalDocs'] = {
                            'description': 'Onvif Specification',
                            'url': f'{SERVICE_WSDL[service]}#op.{fn}'
                        }

                        # patch description for endpoints missing it
                        paths = self.wsdl_files[service]['paths']
                        if f'/{fn}' in paths:
                            # note: all SOAP calls are POST
                            api = paths[f'/{fn}']['post']
                            if 'description' in api and \
                                    ('description' not in method_obj or method_obj['description'].strip() == ''):
                                log.debug(f'Copying description for {service_fn}')
                                method_obj['description'] = api['description']

                    if service_fn in self.matrix.validated:
                        log.debug(f'Adding validated camera list in description for {service_fn}')
                        val_desc = f'''

<details>
<summary><strong>Tested Camera Models</strong></summary>

Below is a list of camera models that this command has been tested against, and whether or not the command is supported.

| Camera | Supported? &nbsp;&nbsp; | Notes |
|--------|:------------|-------|
'''
                        for camera, support in self.matrix.validated[service_fn].cameras.items():
                            val_desc += f'| **{camera}** | {support.support} | {support.notes} |\n'
                        method_obj['description'] = multiline_string(method_obj['description'] + val_desc + '</details>')
                    elif service != EDGEX:
                        # only print warning for non-EdgeX functions
                        log.warning(f'\033[33m*** Missing camera validation entry for command {service_fn} ***\033[0m')

                    if service == EDGEX:
                        if cmd == SNAPSHOT:
                            # Special handling for the EdgeX.Snapshot command, as it uses the same input as
                            # Media.GetSnapshotUri, so patch the jsonObject documentation.
                            log.info(f'Handling {SNAPSHOT} command...')
                            self._lookup_json_object(cmd, method_obj, MEDIA, SNAPSHOT_URI_FN)

                        # nothing left to patch for custom edgex functions, as they do not exist in onvif spec
                        continue

                    # --- More ONVIF function patching ---

                    method_obj['summary'] = f'{service}: {fn}'

                    # Special handling for PUT calls:
                    # - Move example out of schema into json object itself
                    # - Patch the input body schema based on the EdgeX command name and the Onvif function name
                    if method == 'put':
                        # look for the json response object, so we can modify it
                        jscontent = method_obj['requestBody']['content']['application/json']

                        # move the example outside the schema to preserve it (and it belongs better up there)
                        if 'example' in jscontent['schema']:
                            jscontent['example'] = jscontent['schema']['example']
                            self._insert_postman_env(jscontent['example'])

                        # patch PUT call schema by using service name and onvif function name
                        jscontent['schema'] = {
                            'properties': {
                                # EdgeX commands always require the command name as the object key.
                                # Note that this will actually insert the name of the command
                                cmd: {
                                    # this generated name assumes the onvif schemas are named after the commands,
                                    # prefixed by the service name by the xmlstrip.py script.
                                    '$ref': f'#/components/schemas/{service.lower()}_{fn}'
                                }
                            },
                            'required': [cmd],
                            'type': 'object'
                        }

                    # Special handling for GET calls:
                    # - Ensure a 200 OK json response exists
                    # - Generate and override response schema for 200 OK data types based on Onvif spec
                    elif method == 'get':
                        # clone the 200 response to avoid mangling pointer references
                        resp_200 = copy.deepcopy(method_obj['responses']['200'])
                        # get a pointer to the json response body content
                        resp_content = resp_200['content']['application/json']
                        # clone our example get response schema
                        schema = copy.deepcopy(self.sidecar['responses']['onvif']['get'])

                        # patch the response schema to set the objectValue portion to be a reference to the
                        # onvif spec's function response. This assumes the onvif schemas are in the proper format of
                        # function name followed by Response, with the service as the prefix, as set by xmlstrip.py.
                        schema['allOf'][1]['properties']['event']['properties']['readings']['items']['properties']['objectValue'] = {
                            '$ref': f'#/components/schemas/{service.lower()}_{fn}Response'
                        }
                        # override the original schema with this modified one
                        resp_content['schema'] = schema
                        # override with cloned one
                        method_obj['responses']['200'] = resp_200

                        req_schema = f'{service.lower()}_{fn}'
                        if req_schema in self.yml['components']['schemas']:
                            schema = self.yml['components']['schemas'][req_schema]
                            if 'type' in schema and schema['type'] == 'object' and len(schema) == 1:
                                log.debug(f'Skipping empty request schema for {service_fn}')
                            else:
                                self._lookup_json_object(cmd, method_obj, service, fn)

    def _lookup_json_object(self, cmd, method_obj, service, fn):
        found = False
        for param in method_obj['parameters']:
            if param['name'] == 'jsonObject':
                found = True
                self._set_json_object(param, service, fn)
                break
        if not found:
            log.warning(f'\033[33m*** Expected jsonObject parameter for command {cmd}! Creating one. ***\033[0m')
            param = {
                'name': 'jsonObject',
                'in': 'query',
                'schema': {
                    'type': 'string'
                },
                'example': ''
            }
            self._set_json_object(param, service, fn)
            method_obj['parameters'].insert(0, param)

    def _set_json_object(self, param, service, fn):
        """
        This sets the param (which is a jsonObject param) description field to include auto-generated docs
        """
        desc = f'''**Format:**<br/>
This field is a Base64 encoded json string.

**JSON Schema:**
```yaml
{self._gen_pretty_schema_for(f'{service.lower()}_{fn}')}
```

**Field Descriptions:**
{self._gen_field_desc_for(f'{service.lower()}_{fn}')}

**Schema Reference:** [{service.lower()}_{fn}](#{service.lower()}_{fn})
'''
        if param['example'].startswith('{{'):
            key = param['example'].lstrip('{{').rstrip('}}')
            if key in self.postman_env:
                param['example'] = self.postman_env[key]
                if self.postman_env[key].startswith('eyJ'):
                    # if the value is base64, lets insert it
                    js = json.loads(base64.b64decode(self.postman_env[key]))
                    desc += f'''
**Example JSON:**<br/>
> _Note: This value must be encoded to base64!_

```json
{json.dumps(js, indent=2)}
```
'''
        param['description'] = multiline_string(desc)
        # todo: setting the format to Base64 messes with the Swagger UI, and makes it
        #       choose file box, which does not end up working
        # param['schema']['format'] = 'base64'

    def _combine_schemas(self):
        """
        Load all schema files for the onvif spec, and combine them into the loaded yaml file.
        This will also append the pre-defined schemas from the sidecar.yaml file.
        """
        self.wsdl_files = {}
        for service in SERVICE_WSDL.keys():
            fname = f'ref/out/{service.lower()}.yaml'
            log.info(f'Loading schema file: {fname}')
            with open(fname) as f:
                self.wsdl_files[service] = yaml.load(f)

        log.info('Combining schema files')
        schemas = {}
        for schema_file in self.wsdl_files.values():
            for k, v in schema_file['components']['schemas'].items():
                schemas[k] = v

        self.yml['components'] = {
            'schemas': schemas,
            'headers': {},
            'examples': {},
            'parameters': {}
        }

        # note: sidecar should always be added last to override the onvif schemas
        if 'components' in self.sidecar:
            for component in ['schemas', 'headers', 'examples', 'parameters']:
                # for each of the component types, copy the items to the output file
                if component in self.sidecar['components']:
                    for k, v in self.sidecar['components'][component].items():
                        self.yml['components'][component][k] = v

    def _gen_field_desc_for(self, typ):
        """
        Generate a pretty field description in markdown for a specific onvif type
        :param typ: the onvif type
        :return: string
        """
        return self._gen_field_desc(None, '', self.yml['components']['schemas'][typ], indent=0, all_types=set(typ))

    def _gen_field_desc(self, name, desc, val, indent, all_types):
        """Internal recursive field description generator"""
        if 'allOf' in val:
            desc2 = ''
            if len(val['allOf']) > 1 and 'description' in val['allOf'][1]:
                desc2 = val['allOf'][1]['description']
            return self._gen_field_desc(name, desc2, val['allOf'][0], indent, all_types)
        if '$ref' in val:
            typ = val['$ref'].split('/')[-1]
            if typ in all_types:
                output = " " * indent + f'- **{name}** _[Recursive object of type [{typ}](#{typ})]_\n'
                if desc.strip() != '':
                    output += f'<br/>{" " * indent}  {desc}\n'
                return output
            all_types.add(typ)
            return self._gen_field_desc(name, desc, self.yml['components']['schemas'][typ], indent, all_types)

        desc2 = desc if 'description' not in val else val['description']

        if val['type'] == 'object':
            output = ''
            if name is not None:
                output = " " * indent + f'- **{name}** _[object]_\n'
                if desc2.strip() != '':
                    output += f'<br/>{" " * indent}  {desc2}\n'
            for prop, prop_val in val['properties'].items():
                output += self._gen_field_desc(prop, '', prop_val, indent + 2, all_types)
            return output
        else:
            output = f'{" " * indent}- **{name}** _[{val["type"]}]_\n'
            if desc2.strip() != '':  # add description if present
                output += f'<br/>{" " * indent}  {desc2}\n'
            if 'enum' in val:  # add enum values
                output += f'<br/>{" " * indent}  _Enum: [{", ".join([single_quote(e) for e in val["enum"]])}]_\n'
            return output

    def _gen_pretty_schema_for(self, typ):
        """
        Return the pretty json formatted schema for a specific onvif type
        :param typ: the onvif type
        :return: string
        """
        return self._gen_pretty_schema(None, self.yml['components']['schemas'][typ], indent=0, all_types=set(typ))

    def _gen_pretty_schema(self, name, val, indent, all_types):
        """Internal recursive pretty json schema generator"""
        if 'allOf' in val:
            return self._gen_pretty_schema(name, val['allOf'][0], indent, all_types)
        if '$ref' in val:
            typ = val['$ref'].split('/')[-1]
            if typ in all_types:
                return " " * indent + '"%s": { $ref: %s }' % (name, typ)
            all_types.add(typ)
            return self._gen_pretty_schema(name, self.yml['components']['schemas'][typ], indent, all_types)

        if val['type'] == 'object':
            if name is None:
                output = " " * indent + '{\n'
            else:
                output = " " * indent + '"%s": {\n' % name
            inners = []
            for prop, prop_val in val['properties'].items():
                inners.append(self._gen_pretty_schema(prop, prop_val, indent + 2, all_types))
            output += ',\n'.join(inners) + '\n' + " " * indent + '}'
            return output

        elif val['type'] == 'array':
            return f'{" " * indent}"{name}": []'

        elif val['type'] == 'string':
            example = f'<{name}>'
            if 'enum' in val:
                # example = f"Enum: <{', '.join([single_quote(e) for e in val['enum']])}>"
                example = '|'.join(val['enum'])
            return f'{" " * indent}"{name}": "{example}"'

        elif val['type'] == 'integer' or val['type'] == 'number':
            return f'{" " * indent}"{name}": <{name}>'

        elif val['type'] == 'boolean':
            return f'{" " * indent}"{name}": true|false'

        raise ProcessingError('unsupported data type')

    def _verify_complete(self):
        """Checks that all functions from the device profile exist in the openapi file"""
        for cmd, cmd_obj in self.resources.items():
            if cmd_obj['isHidden'] is True:
                continue  # skip hidden commands (not callable by the core-command service)

            api = f'{API_PREFIX}/{cmd}'
            path_obj = None
            if api not in self.yml['paths']:
                log.warning(f'\033[33m*** API "{api}" is missing from input collection! ***\033[0m')
            else:
                path_obj = self.yml['paths'][api]

            if 'getFunction' in cmd_obj['attributes'] and (path_obj is None or 'get' not in path_obj):
                log.warning(f'\033[33m*** Expected call GET "{cmd}" was not found in input yaml! ***\033[0m')
            if 'setFunction' in cmd_obj['attributes'] and (path_obj is None or 'put' not in path_obj):
                log.warning(f'\033[33m*** Expected call PUT "{cmd}" was not found in input yaml! ***\033[0m')

    def _patch_parameters(self):
        """
        Goes through the paths and links parameters to pre-defined ones

        paths/[path]/[method]/parameters/[name=EDGEX_DEVICE_NAME]
        """
        for _, path_obj in self.yml['paths'].items():
            for _, method_obj in path_obj.items():
                for i in range(len(method_obj['parameters'])):
                    name = method_obj['parameters'][i]['name']
                    if name in self.yml['components']['parameters']:
                        # patch the parameter to just be a reference
                        method_obj['parameters'][i] = {'$ref': f'#/components/parameters/{name}'}

    def _clean_response_headers(self):
        """
        - Remove superfluous headers from response objects
        - Patches X-Correlation-Id header to be a $ref
        """
        for _, path_obj in self.yml['paths'].items():
            for _, method_obj in path_obj.items():
                for _, response_obj in method_obj['responses'].items():
                    if 'headers' in response_obj:
                        for header in HEADERS_TO_REMOVE:
                            if header in response_obj['headers']:
                                del response_obj['headers'][header]
                        # patch the correlation id to reference the stock one
                        if 'X-Correlation-Id' in response_obj['headers']:
                            response_obj['headers']['X-Correlation-Id'] = {
                                '$ref': '#/components/headers/correlatedResponseHeader'
                            }

    def _clean_schemas(self):
        cleaner = SchemaCleaner(self.yml)
        cleaner.remove_unused_schemas()

    def process(self):
        """Process the input yaml files, and create the final output yaml file"""
        self._load()
        self._parse()
        self._combine_schemas()
        self._process_apis()
        self._patch_parameters()
        self._clean_response_headers()
        self._verify_complete()
        self._clean_schemas()
        self._write()

    def _insert_postman_env(self, obj):
        if isinstance(obj, dict):
            for k, v in obj.items():
                if isinstance(v, str):
                    if v.startswith('{{'):
                        key = v.lstrip('{{').rstrip('}}')
                        if key in self.postman_env:
                            log.debug(f'Patching postman env: {key}')
                            obj[k] = self.postman_env[key]
                        else:
                            log.warning(f'\033[33m*** Reference to postman env {key} was not found in environment ***\033[0m')
                else:
                    self._insert_postman_env(v)
        elif isinstance(obj, list):
            for item in obj:
                self._insert_postman_env(item)


def main():
    if len(sys.argv) != 8:
        print(f'Usage: {sys.argv[0]} <input_file> <sidecar_file> <profile_file> <output_file> <onvif_tested_file> <onvif_footnotes_file> <postman_env_file>')
        sys.exit(1)

    logging.basicConfig(level=(logging.DEBUG if os.getenv('DEBUG_LOGGING') == '1' else logging.INFO),
                        format='%(asctime)-15s %(levelname)-8s %(name)-12s %(message)s')

    proc = YamlProcessor(sys.argv[1],  # input_file
                         sys.argv[2],  # sidecar_file
                         sys.argv[3],  # profile_file
                         sys.argv[4],  # output_file
                         MarkdownMatrix(sys.argv[5], sys.argv[6]),  # onvif_tested_file, onvif_footnotes_file
                         sys.argv[7])  # postman_env_file
    proc.process()


if __name__ == '__main__':
    main()
