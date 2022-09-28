#!/usr/bin/env python3

import dataclasses
import sys

import yaml

EDGEX_DEVICE_NAME = 'Camera001'


@dataclasses.dataclass
class YamlProcessor:
    input_file: str
    sidecar_file: str
    output_file: str
    yml = None
    extra = None

    """Read input yaml file and sidecar yaml files"""
    def _load(self):
        # read input yaml file
        with open(self.input_file) as f:
            self.yml = yaml.safe_load(f.read())

        # read sidecar file
        with open(self.sidecar_file) as f:
            self.extra = yaml.safe_load(f.read())

    """Output modified yaml file"""
    def _write(self):
        with open(self.output_file, 'w') as w:
            w.write(yaml.dump(self.yml))

    """Side load data from sidecar yaml file into yaml object"""
    def _sideload_data(self):
        pass

    """
    Goes through the paths and adds example values to all missing fields
    
    paths/[path]/[method]/parameters/[name=EDGEX_DEVICE_NAME]
    """
    def _add_example_vars(self):
        for p, path in self.yml['paths'].items():
            for m, method in path.items():
                for param in method['parameters']:
                    if param['name'] == 'EDGEX_DEVICE_NAME':
                        param['example'] = EDGEX_DEVICE_NAME

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
    proc.process()


if __name__ == '__main__':
    main()
