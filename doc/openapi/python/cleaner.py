#!/usr/bin/env python3
#
# Copyright (C) 2022 Intel Corporation
# SPDX-License-Identifier: Apache-2.0
#

from collections import defaultdict
from dataclasses import dataclass, field
from ruamel.yaml import YAML

yaml = YAML()


@dataclass
class Schema:
    uses: set = field(default_factory=set)
    used_by: set = field(default_factory=set)


@dataclass
class SchemaCleaner:
    """
    Implementation of a schema cleaner to remove unused schema definitions.

    Note: This is a one time use object, do not re-use it!
    """
    yml: any
    schemas: 'dict[str, Schema]' = field(default_factory=lambda: defaultdict(Schema))
    unused: set = field(default_factory=set)

    def _inner_parse(self, current, obj):
        """Inner recursive portion of _parse_schemas"""
        if isinstance(obj, dict):
            if '$ref' in obj:
                name = obj['$ref'].split('/')[-1]
                self.schemas[current].uses.add(name)
                self.schemas[name].used_by.add(current)
            for _, x_obj in obj.items():
                self._inner_parse(current, x_obj)
        elif isinstance(obj, list):
            for x_obj in obj:
                self._inner_parse(current, x_obj)

    def _parse_schemas(self):
        """Parse the yml for all schemas and recursively detect what they are used by, and who uses them"""
        for path, path_obj in self.yml['paths'].items():
            for method, method_obj in path_obj.items():
                name = f'{path.split("/")[-1]}_{method}'
                # this is a top level path, so add a used_by to prevent it from being detected as unused
                self.schemas[name].used_by.add('TOP_LEVEL')
                self._inner_parse(name, method_obj)

        for name, schema in self.yml['components']['schemas'].items():
            self._inner_parse(name, schema)

    def _inner_find_unused(self, ref, name):
        """Inner recursive part of _find_unused"""
        if ref is not None and ref in self.schemas[name].used_by:
            self.schemas[name].used_by.remove(ref)
        if len(self.schemas[name].used_by) == 0:
            for use in self.schemas[name].uses:
                self._inner_find_unused(name, use)

    def _find_unused(self):
        """
        Recursively find the schemas that are unused by removing references to a schema which itself is unused

        For Example: If Schema A is unused, and it refers to Schema B, remove the usage from Schema B. Now if Schema B has no
        more usages, it is considered unused as well.
        """
        for name in self.schemas.keys():
            self._inner_find_unused(None, name)

        # note: we need to loop through the original yml schema list (and NOT self.schemas) to find schemas that are
        # unused which do not have any usages themselves
        for name in self.yml['components']['schemas'].keys():
            if len(self.schemas[name].used_by) == 0:
                self.unused.add(name)

    def _remove_unused(self):
        """Remove all unused items from the schema portion of the yaml"""
        for name in self.unused:
            self.yml["components"]["schemas"].pop(name)

    def remove_unused_schemas(self) -> any:
        """
        Recursively parses the schemas to determine which ones are used and which ones are not. The unused
        ones are subsequently deleted from the input yaml data.
        """
        print('Removing unused schemas')
        total = len(self.yml["components"]["schemas"])  # cache the total number as it will be modified by the end
        self._parse_schemas()
        self._find_unused()
        self._remove_unused()
        print(f'Removed {len(self.unused)} unused schemas of {total} total schemas!')
        print(f'{len(self.yml["components"]["schemas"])} total schemas remain')

