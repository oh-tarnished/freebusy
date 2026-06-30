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
    package='freebusy.schedule.v1',
    manifest={
        'ExceptionKind',
    },
)


class ExceptionKind(proto.Enum):
    r"""Whether an exception removes or adds availability.

    Values:
        EXCEPTION_KIND_UNSPECIFIED (0):
            Unset.
        EXCEPTION_KIND_CLOSURE (1):
            The resource is closed for the span (blackout
            / holiday).
        EXCEPTION_KIND_EXTRA_HOURS (2):
            The resource is open beyond its recurring
            hours for the span.
    """
    EXCEPTION_KIND_UNSPECIFIED = 0
    EXCEPTION_KIND_CLOSURE = 1
    EXCEPTION_KIND_EXTRA_HOURS = 2


__all__ = tuple(sorted(__protobuf__.manifest))
