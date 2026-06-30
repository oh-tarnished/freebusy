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


__protobuf__ = proto.module(
    package='freebusy.availability.v1',
    manifest={
        'Code',
    },
)


class Code(proto.Enum):
    r"""Machine-readable reasons a span is not bookable.

    Values:
        CODE_UNSPECIFIED (0):
            Unset.
        CODE_NO_CAPACITY (1):
            Not enough free units for the requested
            count.
        CODE_OUTSIDE_HOURS (2):
            The span falls outside the resource's
            recurring hours.
        CODE_CLOSED (3):
            A closure exception (blackout/holiday) covers
            part of the span.
        CODE_MIN_NIGHTS (4):
            Shorter than the minimum stay (min_nights).
        CODE_MAX_NIGHTS (5):
            Longer than the maximum stay (max_nights).
        CODE_CHECKIN_DAY (6):
            Check-in falls on a disallowed weekday.
        CODE_CHECKOUT_DAY (7):
            Check-out falls on a disallowed weekday.
        CODE_MIN_NOTICE (8):
            The span starts sooner than the minimum
            notice allows.
        CODE_MAX_ADVANCE (9):
            The span starts further out than the advance
            window allows.
        CODE_BUFFER_CONFLICT (10):
            A buffer or gap rule around an adjacent
            booking conflicts.
        CODE_RESOURCE_ARCHIVED (11):
            The resource is archived.
    """
    CODE_UNSPECIFIED = 0
    CODE_NO_CAPACITY = 1
    CODE_OUTSIDE_HOURS = 2
    CODE_CLOSED = 3
    CODE_MIN_NIGHTS = 4
    CODE_MAX_NIGHTS = 5
    CODE_CHECKIN_DAY = 6
    CODE_CHECKOUT_DAY = 7
    CODE_MIN_NOTICE = 8
    CODE_MAX_ADVANCE = 9
    CODE_BUFFER_CONFLICT = 10
    CODE_RESOURCE_ARCHIVED = 11


__all__ = tuple(sorted(__protobuf__.manifest))
