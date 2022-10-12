#!/usr/bin/env python3

import dataclasses
import io
import sys
import copy
import textwrap

from ruamel.yaml import YAML
from ruamel.yaml.scalarstring import LiteralScalarString
yaml = YAML()

EDGEX = 'EdgeX'
EDGEX_DEVICE_NAME = 'Camera001'
API_PREFIX = '/api/v2/device/name/{EDGEX_DEVICE_NAME}'

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
    'Content-Type',
    'Date',
    'Transfer-Encoding',
]


class ProcessingError(RuntimeError):
    pass


def make_scalar(val):
    return LiteralScalarString(textwrap.dedent(val))


@dataclasses.dataclass
class YamlProcessor:
    input_file: str
    sidecar_file: str
    profile_file: str
    output_file: str
    yml = None
    sidecar = None
    profile = None
    resources = {}
    wsdl_files = {}

    def _load(self):
        """Read input yaml file and sidecar yaml files"""
        print(f'Reading input OpenAPI file: {self.input_file}')
        with open(self.input_file) as f:
            self.yml = yaml.load(f)

        print(f'Reading sidecar file: {self.sidecar_file}')
        with open(self.sidecar_file) as f:
            self.sidecar = yaml.load(f)

        print(f'Reading profile file: {self.profile_file}')
        with open(self.profile_file) as f:
            self.profile = yaml.load(f)

    def _parse(self):
        """Parse the device resources into a lookup table"""
        for resource in self.profile['deviceResources']:
            self.resources[resource['name']] = resource

    def _write(self):
        """Output modified yaml file"""
        print(f'Writing output OpenAPI file: {self.output_file}')
        with open(self.output_file, 'w') as w:
            yaml.dump(self.yml, w)

    def _process_apis(self):
        """
        Sideload externalDocs using EdgeX profile file, update descriptions, and set schemas
        """

        for path, path_obj in self.yml['paths'].items():
            for method, method_obj in path_obj.items():
                cmd = path.split('/')[-1]

                prefix = 'set' if method == 'put' else method
                if cmd in self.resources:
                    attrs = self.resources[cmd]['attributes']
                    fn = attrs[f'{prefix}Function']
                    service = attrs['service']

                    # add all responses
                    for code, resp_obj in self.sidecar['responses']['canned'].items():
                        if code not in method_obj['responses']:
                            method_obj['responses'][code] = resp_obj
                        elif code == '200':
                            content = method_obj['responses']['200']['content']
                            if 'application/json' not in content or \
                                    len(content['application/json']) == 0 or \
                                    ('example' in content['application/json'] and len(content['application/json']['example']) == 2):
                                print(f'Overriding empty 200 response for {service}_{fn}')
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
                                print(f'\033[33m[WARNING] \t -- Missing schema response definition for EdgeX command {method.upper()} {cmd}\033[0m')
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
                                print(f'\033[33m[WARNING] \t -- Missing schema request definition for EdgeX command {method.upper()} {cmd}\033[0m')

                            # override the response schema with default 200 response
                            method_obj['responses']['200'] = self.sidecar['responses']['canned']['200']

                        # nothing left to patch for custom edgex functions, as they do not exist in onvif spec
                        continue

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
                            print(f'Copying description for {service}_{fn}')
                            method_obj['description'] = api['description']

                    # Special handling for PUT calls:
                    # - Move example out of schema into json object itself
                    # - Patch the input body schema based on the EdgeX command name and the Onvif function name
                    if method == 'put':
                        # look for the json response object, so we can modify it
                        jscontent = method_obj['requestBody']['content']['application/json']

                        # move the example outside the schema to preserve it (and it belongs better up there)
                        if 'example' in jscontent['schema']:
                            jscontent['example'] = jscontent['schema']['example']

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
                                print(f'Skipping empty request schema for {service}_{fn}')
                            else:
                                buf = io.BytesIO()
                                yaml.dump({
                                    f'{service.lower()}_{fn}':
                                        self.yml['components']['schemas'][f'{service.lower()}_{fn}']},
                                    buf)
                                method_obj['description'] = make_scalar(method_obj['description'] + f'''

<hr/>

**`jsonObject` Schema:** 

_See: [{service.lower()}_{fn}](#{service.lower()}_{fn})_

```yaml
{self._gen_pretty_schema(None, self.yml['components']['schemas'][f'{service.lower()}_{fn}'], 
                         indent=0, all_types=set(f'{service.lower()}_{fn}'))}
```''')

    def _combine_schemas(self):
        """
        Load all schema files for the onvif spec, and combine them into the loaded yaml file.
        This will also append the pre-defined schemas from the sidecar.yaml file.
        """
        self.wsdl_files = {}
        for service in SERVICE_WSDL.keys():
            fname = f'ref/out/{service.lower()}.yaml'
            print(f'Loading schema file: {fname}')
            with open(fname) as f:
                self.wsdl_files[service] = yaml.load(f)

        print('Combining schema files')
        schemas = {}
        for schema_file in self.wsdl_files.values():
            for k, v in schema_file['components']['schemas'].items():
                schemas[k] = v

        self.yml['components'] = {
            'schemas': schemas,
            'headers': {},
            'examples': {},
        }

        # note: sidecar should always be added last to override the onvif schemas
        if 'components' in self.sidecar:
            for component in ['schemas', 'headers', 'examples']:
                # for each of the component types, copy the items to the output file
                if component in self.sidecar['components']:
                    for k, v in self.sidecar['components'][component].items():
                        self.yml['components'][component][k] = v

    def _gen_pretty_schema(self, name, val, indent, all_types):
        if 'allOf' in val:
            return self._gen_pretty_schema(name, val['allOf'][0], indent, all_types)
        if '$ref' in val:
            typ = val['$ref'].split('/')[-1]
            if typ in all_types:
                return " "*indent + '"%s": { $ref: %s }' % (name, typ)
            all_types.add(typ)
            return self._gen_pretty_schema(name, self.yml['components']['schemas'][typ], indent, all_types)

        if val['type'] == 'object':
            if name is None:
                output = " "*indent + '{\n'
            else:
                output = " "*indent + '"%s": {\n' % name
            inners = []
            for prop, prop_val in val['properties'].items():
                inners.append(self._gen_pretty_schema(prop, prop_val, indent+2, all_types))
            output += ',\n'.join(inners) + '\n' + " "*indent + '}'
            return output

        elif val['type'] == 'array':
            return f'{" "*indent}"{name}": []'

        elif val['type'] == 'string':
            return f'{" "*indent}"{name}": "<{name}>"'

        elif val['type'] == 'integer' or val['type'] == 'number':
            return f'{" "*indent}"{name}": <{name}>'

        elif val['type'] == 'boolean':
            return f'{" "*indent}"{name}": true|false'
        
        raise ProcessingError('unsupported data type')

    def _verify_complete(self):
        """
        _verify_complete checks that all functions from the device profile exist in the openapi file
        """
        for cmd, cmd_obj in self.resources.items():
            if cmd_obj['isHidden'] is True:
                continue  # skip hidden commands (not callable by the core-command service)

            api = f'{API_PREFIX}/{cmd}'
            path_obj = None
            if api not in self.yml['paths']:
                print(f'\033[33m[WARNING] API "{api}" is missing from input collection!\033[0m')
            else:
                path_obj = self.yml['paths'][api]

            if 'getFunction' in cmd_obj['attributes'] and (path_obj is None or 'get' not in path_obj):
                print(f'\033[33m[WARNING] \t -- Expected call GET "{cmd}" was not found in input yaml!\033[0m')
            if 'setFunction' in cmd_obj['attributes'] and (path_obj is None or 'put' not in path_obj):
                print(f'\033[33m[WARNING] \t -- Expected call PUT "{cmd}" was not found in input yaml!\033[0m')

    def _add_example_vars(self):
        """
        Goes through the paths and adds example values to all missing fields

        paths/[path]/[method]/parameters/[name=EDGEX_DEVICE_NAME]
        """
        for _, path_obj in self.yml['paths'].items():
            for _, method_obj in path_obj.items():
                for param_obj in method_obj['parameters']:
                    if param_obj['name'] == 'EDGEX_DEVICE_NAME':
                        param_obj['example'] = EDGEX_DEVICE_NAME

    def _clean_response_headers(self):
        """
        Remove superfluous headers from response objects
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

    def process(self):
        """Process the input yaml files, and create the final output yaml file"""
        self._load()
        self._parse()
        self._combine_schemas()
        self._process_apis()
        self._add_example_vars()
        self._clean_response_headers()
        self._verify_complete()
        self._write()


def main():
    if len(sys.argv) != 5:
        print(f'Usage: {sys.argv[0]} <input_file> <sidecar_file> <profile_file> <output_file>')
        sys.exit(1)

    proc = YamlProcessor(sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4])
    proc.process()


if __name__ == '__main__':
    main()
