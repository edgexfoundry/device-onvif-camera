#!/usr/bin/env python3

import dataclasses
import sys

import yaml

EDGEX_DEVICE_NAME = 'Camera001'
API_PREFIX = '/api/v2/device/name/{EDGEX_DEVICE_NAME}'


class ProcessingError(RuntimeError):
    pass


@dataclasses.dataclass
class YamlProcessor:
    input_file: str
    sidecar_file: str
    output_file: str
    yml = None
    sidecar = None

    """Read input yaml file and sidecar yaml files"""
    def _load(self):
        # read input yaml file
        with open(self.input_file) as f:
            self.yml = yaml.safe_load(f.read())

        # read sidecar file
        with open(self.sidecar_file) as f:
            self.sidecar = yaml.safe_load(f.read())

    """Output modified yaml file"""
    def _write(self):
        with open(self.output_file, 'w') as w:
            w.write(yaml.dump(self.yml))

    """Side load data from sidecar yaml file into yaml object"""
    def _sideload_data(self):
        self._sideload_external_docs()

    """Side load externalDocs"""
    def _sideload_external_docs(self):
        for cmd, methods in self.sidecar['externalDocs'].items():
            api = f'{API_PREFIX}/{cmd}'
            if api not in self.yml['paths']:
                raise ProcessingError(f'Expected api "{api}" was not found in input yaml!')
            path_obj = self.yml['paths'][api]

            for method, url in methods.items():
                if method not in path_obj:
                    raise ProcessingError(f'Method {method} missing from api "{api}" in input yaml!')

                print(f'Patching external docs for [{method}] {api}')
                method_obj = path_obj[method]
                method_obj['externalDocs'] = {
                    'description': 'Onvif Specification',
                    'url': url
                }

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
        self._sideload_data()
        self._add_example_vars()
        self._write()


def main():
    if len(sys.argv) != 4:
        print(f'Usage: {sys.argv[0]} <input_file> <sidecar_file> <output_file>')
        sys.exit(1)

    proc = YamlProcessor(sys.argv[1], sys.argv[2], sys.argv[3])
    # todo: try-catch ProcessingError and dump trace
    proc.process()


if __name__ == '__main__':
    main()
