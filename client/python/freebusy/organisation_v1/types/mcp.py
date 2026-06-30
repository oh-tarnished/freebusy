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
    package='freebusy.organisation.v1',
    manifest={
        'InviteMemberArgs',
    },
)


class InviteMemberArgs(proto.Message):
    r"""Arguments for the "invite_member" prompt.

    Attributes:
        organisation (str):
            Organisation to invite into, as a resource
            name ("organisations/7") or a display name.
        email (str):
            Email address to invite.
        role (str):
            Role to grant, e.g. "owner", "admin",
            "member", or "viewer".
    """

    organisation: str = proto.Field(
        proto.STRING,
        number=1,
    )
    email: str = proto.Field(
        proto.STRING,
        number=2,
    )
    role: str = proto.Field(
        proto.STRING,
        number=3,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
