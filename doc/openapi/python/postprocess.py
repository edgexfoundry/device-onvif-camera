#!/usr/bin/env python3

import dataclasses
import sys

import yaml

EDGEX = 'EdgeX'
EDGEX_DEVICE_NAME = 'Camera001'
API_PREFIX = '/api/v2/device/name/{EDGEX_DEVICE_NAME}'


SERVICE_WSDL = {
    'Analytics': 'https://www.onvif.org/ver20/analytics/wsdl/analytics.wsdl',
    'Device': 'https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl',
    'Event': 'https://www.onvif.org/ver10/events/wsdl/event.wsdl',
    'Imaging': 'https://www.onvif.org/ver20/imaging/wsdl/imaging.wsdl',
    'Media': 'https://www.onvif.org/ver10/media/wsdl/media.wsdl',
    'Media2': 'https://www.onvif.org/ver20/media/wsdl/media.wsdl',
    'PTZ': 'https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl'
}


class ProcessingError(RuntimeError):
    pass


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

    """Read input yaml file and sidecar yaml files"""
    def _load(self):
        print(f'Reading input OpenAPI file: {self.input_file}')
        with open(self.input_file) as f:
            self.yml = yaml.safe_load(f.read())

        print(f'Reading sidecar file: {self.sidecar_file}')
        with open(self.sidecar_file) as f:
            self.sidecar = yaml.safe_load(f.read())

        print(f'Reading profile file: {self.profile_file}')
        with open(self.profile_file) as f:
            self.profile = yaml.safe_load(f.read())

    """Parse the device resources into a lookup table"""
    def _parse(self):
        for resource in self.profile['deviceResources']:
            self.resources[resource['name']] = resource

    """Output modified yaml file"""
    def _write(self):
        print(f'Writing output OpenAPI file: {self.output_file}')
        with open(self.output_file, 'w') as w:
            w.write(yaml.dump(self.yml))

    """Sideload externalDocs using EdgeX profile file"""
    def _sideload_external_docs(self):
        for path, path_obj in self.yml['paths'].items():
            for method, method_obj in path_obj.items():
                cmd = path.split('/')[-1]

                prefix = 'set' if method == 'put' else method
                if cmd in self.resources:
                    attrs = self.resources[cmd]['attributes']
                    fn = attrs[f'{prefix}Function']
                    service = attrs['service']

                    if service != EDGEX:
                        print(f'| {method.upper()} | {cmd} | {fn} | {fn.replace(cmd, "")} |')
                        # print(f'Patching external docs for {method.upper()} {cmd}')
                        method_obj = path_obj[method]
                        method_obj['externalDocs'] = {
                            'description': 'Onvif Specification',
                            'url': f'{SERVICE_WSDL[service]}#op.{fn}'
                        }

    def _verify_complete(self):
        for cmd, cmd_obj in self.resources.items():
            api = f'{API_PREFIX}/{cmd}'
            if api not in self.yml['paths']:
                print(f'\033[33m[WARNING] API "{api}" is missing from input collection!\033[0m')
                continue

            path_obj = self.yml['paths'][api]
            if 'getFunction' in cmd_obj['attributes'] and 'get' not in path_obj:
                print(f'\033[33m[WARNING] Expected call GET "{cmd}" was not found in input yaml!\033[0m')
            if 'setFunction' in cmd_obj['attributes'] and 'put' not in path_obj:
                print(f'\033[33m[WARNING] Expected call PUT "{cmd}" was not found in input yaml!\033[0m')

    """
    Goes through the paths and adds example values to all missing fields
    
    paths/[path]/[method]/parameters/[name=EDGEX_DEVICE_NAME]
    """
    def _add_example_vars(self):
        for _, path_obj in self.yml['paths'].items():
            for _, method_obj in path_obj.items():
                for param_obj in method_obj['parameters']:
                    if param_obj['name'] == 'EDGEX_DEVICE_NAME':
                        param_obj['example'] = EDGEX_DEVICE_NAME

    """Process the input yaml files, and create the final output yaml file"""
    def process(self):
        self._load()
        self._parse()
        self._sideload_external_docs()
        self._add_example_vars()
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
