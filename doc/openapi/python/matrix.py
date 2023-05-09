#!/usr/bin/env python3
#
# Copyright (C) 2022-2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0
#

from enum import Enum, auto
from dataclasses import dataclass, field
import sys
import os

from ruamel.yaml import YAML

yaml = YAML()


class State(Enum):
    NotStarted = auto()
    WantSection = auto()
    NewSection = auto()
    Commands = auto()


@dataclass
class CameraSupport:
    camera: str
    support: str
    notes: str = ''


@dataclass
class Command:
    section: str
    service: str
    func: str
    cameras: dict = field(default_factory=dict)

    @property
    def qualified_name(self):
        return f'{self.service}_{self.func}'


@dataclass
class Section:
    name: str
    cameras: list = field(default_factory=list)


@dataclass
class MarkdownMatrix:
    tested_file: str
    footnotes_file: str
    section: Section = None
    service: str = ''
    state: State = State.NotStarted
    validated: dict = field(default_factory=dict)
    footnotes: dict = field(default_factory=dict)

    def _load_footnotes(self):
        basename = os.path.basename(self.footnotes_file)
        current_key = None
        current_data = ''
        with open(self.footnotes_file) as f:
            while line := f.readline():
                if not line.startswith('### '):
                    if current_key is None:
                        continue
                    current_data += line.strip() + ' '
                else:
                    if current_key is not None:
                        self.footnotes[current_key] = current_data.strip()
                    current_data = ''
                    key = line.lstrip('### ').lower().replace(' ', '-').strip()
                    current_key = f'{basename}#{key}'
            if current_key is not None:
                self.footnotes[current_key] = current_data.strip()

    def parse(self):
        self._load_footnotes()
        with open(self.tested_file) as f:
            i = 0
            for line in f:
                i += 1
                line = line.strip()
                if self.state == State.NotStarted:
                    if line.startswith('## Tested Onvif Cameras'):
                        self.state = State.WantSection
                    continue

                if line.startswith('### '):
                    self.state = State.NewSection
                    self.section = Section(line.lstrip('### ').strip())
                    continue

                if self.state == State.NewSection:
                    if not line.startswith('| Onvif Web Service | Onvif Function'):
                        continue  # keep going until we get the proper line
                    tokens = line.lstrip('| Onvif Web Service | Onvif Function').split('|')
                    cameras = [x.strip() for x in tokens if x.strip() != '']
                    self.section.cameras = cameras
                    self.state = State.Commands
                elif self.state == State.Commands:
                    if not line.startswith('|'):
                        continue
                    fields = [x.strip() for x in line.split('|')][1:]
                    if fields[0].startswith('----'):
                        continue  # skip post-header line

                    if fields[0].startswith('**'):
                        # the service name is only placed on lines where it changes. cache the
                        # service name and only change it when a new one is found. Format: **ServiceName**
                        self.service = fields[0].strip('*')

                    command = Command(self.section.name, self.service, fields[1])
                    for i in range(len(self.section.cameras)):
                        camera = self.section.cameras[i]
                        data = fields[i+2].strip()
                        notes = ''
                        if data == '':
                            continue  # skip empty results

                        # replace footnotes link with actual footnotes data
                        if '[ⓘ](' in data:
                            key = data[data.rfind('[ⓘ](')+4:-1]
                            if key in self.footnotes:
                                data = data[:data.rfind('[ⓘ]')]
                                notes = self.footnotes[key]

                        command.cameras[camera] = CameraSupport(camera, data.replace('✔', '✔️').strip(), notes)
                    self.validated[command.qualified_name] = command
