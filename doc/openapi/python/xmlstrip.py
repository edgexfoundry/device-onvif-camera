#!/usr/bin/env python3

# Copyright (C) 2022 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

import re
import sys
import yaml

# schemas to not include in the exported file
EXCLUDED_SCHEMAS = ['wstop_Topic', 'wstop_TopicNamespaceType', 'wstop_TopicType']

TITLE_REGEX = re.compile(r' +title: .*')
TNS_REGEX = re.compile(r'([ /])tns_')
TT_REGEX = re.compile(r'([ /])tt_')


def main():
    if len(sys.argv) != 4:
        print(f'Usage: {sys.argv[0]} <service> <input_file> <output_file>')

    with open(sys.argv[2]) as f:
        yml = yaml.safe_load(f.read())

    for name, obj in yml.items():
        strip_xml(name, obj)

    for name, obj in yml.items():
        fix_refs(name, obj)

    yml['components']['schemas'] = fix_schemas(yml['components']['schemas'])

    service = sys.argv[1]
    with open(sys.argv[3], 'w') as w:
        # todo: this can be optimized better using streams. Right now it dumps the yaml to a string
        #       and then processes each raw line before actually writing it to the output file.
        lines = yaml.dump(yml).split('\n')
        for line in lines:
            if TITLE_REGEX.match(line):
                continue  # skip the title elements as they are all just superfluous
            # namespace all tns schemas to this specific service
            line = TNS_REGEX.sub(fr'\1{service}_', line)
            # namespace all tt schemas as common onvif
            line = TT_REGEX.sub(r'\1onvif_', line)
            w.write(line)
            w.write('\n')


def fix_schemas(schemas):
    """
    Namespace all schemas by service name
    """
    out = {}
    service = sys.argv[1]
    for k, v in schemas.items():

        # note: this section is a bit redundant with the output writer that does a raw replace all, though
        # this has the benefit of printing out the transformations.
        if k.startswith('tt_') | k.startswith('tns_'):
            k2 = k.replace('tt_', f'onvif_').replace('tns_', f'{service}_')
            print(f'{k} -> {k2}')
            out[k2] = v

        # only add schemas which are namespaced (ie. contain an underscore)
        elif '_' in k and k not in EXCLUDED_SCHEMAS:
            out[k] = v
    return out


def fix_refs(name, obj):
    """
    Turns this:

        PullMessagesFaultResponse:
          allOf:
          - $ref: '#/components/schemas/event_PullMessagesFaultResponse'

    into this:

        PullMessagesFaultResponse:
          $ref: '#/components/schemas/event_PullMessagesFaultResponse'
    """
    if isinstance(obj, dict):
        if 'allOf' in obj and len(obj) == 1:
            if len(obj['allOf']) == 1 and '$ref' in obj['allOf'][0]:
                print(f'fixing schema ref for {name}')
                obj['$ref'] = obj['allOf'][0]['$ref']
                del obj['allOf']
        else:
            for n2, o2 in obj.items():
                fix_refs(f'{name}.{n2}', o2)

    elif isinstance(obj, list):
        for o2 in obj:
            fix_refs(f'{name}[]', o2)


def strip_xml(name, obj):
    """
    - Remove xml definitions from openapi schema objects
    - Remove empty description fields
    - Redefine application/xml mime types to application/json
    - Remove empty values from arrays
    """
    if isinstance(obj, dict):
        if 'xml' in obj:
            print(f'Stripping xml field from {name}')
            del obj['xml']

        if 'description' in obj and obj['description'].strip() == '':
            del obj['description']

        if 'application/xml' in obj:
            print('Redefining mime type to application/json')
            obj['application/json'] = obj['application/xml']
            del obj['application/xml']

        for n2, o2 in obj.items():
            strip_xml(f'{name}.{n2}', o2)

    elif isinstance(obj, list):
        for o2 in obj:
            strip_xml(f'{name}[]', o2)
        if len(obj) == 2 and obj[1] == {}:
            print('Removing empty value from array')
            del obj[1]


if __name__ == '__main__':
    main()
