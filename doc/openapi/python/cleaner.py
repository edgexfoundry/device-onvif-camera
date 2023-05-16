#!/usr/bin/env python3
#
# Copyright (C) 2022-2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0
#

import textwrap
from collections import defaultdict
from dataclasses import dataclass, field
from ruamel.yaml import YAML
from typing import Dict
import logging

from ruamel.yaml.scalarstring import LiteralScalarString

yaml = YAML()
log = logging.getLogger('cleaner')


def multiline_string(val: str):
    """Takes a string value and wraps it so that ruamel.yaml will format it as a raw multi-line string"""
    return LiteralScalarString(textwrap.dedent(val))


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
    schemas: Dict[str, Schema] = field(default_factory=lambda: defaultdict(Schema))
    unused: set = field(default_factory=set)

    def _inner_parse(self, current, obj):
        """Inner recursive portion of _parse_schemas"""
        if isinstance(obj, dict):
            if '$ref' in obj:
                name = obj['$ref'].split('/')[-1]
                self.schemas[current].uses.add(name)
                self.schemas[name].used_by.add(current)
            elif 'operationId' in obj:
                name = obj['operationId']
                self.schemas[current].uses.add(name)
                self.schemas[name].used_by.add(current)

            # check if there is a description field that is more than one line, but is not already a
            # ruamel.Yaml multi-line string, and patch it to be one.
            if 'description' in obj and '\n' in obj['description'] and type(obj['description']) == str:
                # Typically, the onvif schema definitions that are multi-line do not indent the first line, but
                # indent the rest. The textwrap.dedent() method will not dedent it due to this. To fix this,
                # we need to take the first line as is, and dedent the rest of the lines.
                lines = str(obj['description']).split('\n', 1)
                obj['description'] = multiline_string(lines[0] + '\n' + textwrap.dedent(lines[1]))

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
        log.debug('Removing unused schemas')
        total = len(self.yml["components"]["schemas"])  # cache the total number as it will be modified by the end
        self._parse_schemas()
        self._find_unused()
        self._remove_unused()
        log.info(f'Removed {len(self.unused)} unused schemas of {total} total schemas!')
        log.debug(f'{len(self.yml["components"]["schemas"])} total schemas remain')

