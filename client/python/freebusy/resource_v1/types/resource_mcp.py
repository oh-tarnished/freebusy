# -*- coding: utf-8 -*-
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
from __future__ import annotations

from typing import MutableMapping, MutableSequence

import proto  # type: ignore

import freebusy.shared.v1.enums_pb2 as enums_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.resource.v1',
    manifest={
        'AddResourceArgs',
    },
)


class AddResourceArgs(proto.Message):
    r"""Arguments for the "add_resource" prompt.

    Attributes:
        display_name (str):
            Human-friendly name of the resource.
        type_ (str):
            What kind of bookable thing it is (e.g. "provider", "room",
            "unit_type").
        booking_mode (freebusy.shared.v1.enums_pb2.BookingMode):
            How it is booked: time-slot appointments or
            nightly stays.
        time_zone (str):
            IANA timezone the resource operates in (e.g.
            "America/New_York").
        capacity (int):
            Number of interchangeable units in the pool.
            Defaults to 1.
    """

    display_name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    type_: str = proto.Field(
        proto.STRING,
        number=2,
    )
    booking_mode: enums_pb2.BookingMode = proto.Field(
        proto.ENUM,
        number=3,
        enum=enums_pb2.BookingMode,
    )
    time_zone: str = proto.Field(
        proto.STRING,
        number=4,
    )
    capacity: int = proto.Field(
        proto.INT32,
        number=5,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
