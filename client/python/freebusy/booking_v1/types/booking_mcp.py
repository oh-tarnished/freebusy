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

import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.booking.v1',
    manifest={
        'BookSlotArgs',
    },
)


class BookSlotArgs(proto.Message):
    r"""Arguments for the "book_slot" prompt.

    Attributes:
        resource (str):
            Resource to book, as a resource name
            ("resources/42") or a display name.
        offering (str):
            Offering to book, as a resource name or a
            display name (e.g. "30-min consult").
        start_time (google.protobuf.timestamp_pb2.Timestamp):
            Start of the booking (RFC 3339, e.g.
            "2026-07-01T14:00:00Z").
        units (int):
            Number of units / party size. Defaults to 1.
        promo_code (str):
            Promo code to apply, if any.
    """

    resource: str = proto.Field(
        proto.STRING,
        number=1,
    )
    offering: str = proto.Field(
        proto.STRING,
        number=2,
    )
    start_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=3,
        message=timestamp_pb2.Timestamp,
    )
    units: int = proto.Field(
        proto.INT32,
        number=4,
    )
    promo_code: str = proto.Field(
        proto.STRING,
        number=5,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
