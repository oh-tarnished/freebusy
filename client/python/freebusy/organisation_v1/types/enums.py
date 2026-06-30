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
        'MemberState',
        'OrganisationState',
        'OrganisationRole',
    },
)


class MemberState(proto.Enum):
    r"""Confirmation state of a membership.

    Values:
        MEMBER_STATE_UNSPECIFIED (0):
            Unset.
        MEMBER_STATE_INVITED (1):
            Invited, awaiting acceptance.
        MEMBER_STATE_ACTIVE (2):
            Active member.
        MEMBER_STATE_SUSPENDED (3):
            Suspended within the organisation.
    """
    MEMBER_STATE_UNSPECIFIED = 0
    MEMBER_STATE_INVITED = 1
    MEMBER_STATE_ACTIVE = 2
    MEMBER_STATE_SUSPENDED = 3


class OrganisationState(proto.Enum):
    r"""Lifecycle state of an organisation.

    Values:
        ORGANISATION_STATE_UNSPECIFIED (0):
            Unset.
        ORGANISATION_STATE_ACTIVE (1):
            Active.
        ORGANISATION_STATE_SUSPENDED (2):
            Suspended; access blocked.
    """
    ORGANISATION_STATE_UNSPECIFIED = 0
    ORGANISATION_STATE_ACTIVE = 1
    ORGANISATION_STATE_SUSPENDED = 2


class OrganisationRole(proto.Enum):
    r"""A member's role within an organisation.

    Values:
        ORGANISATION_ROLE_UNSPECIFIED (0):
            Unset.
        ORGANISATION_ROLE_OWNER (1):
            Full control, including billing and deletion.
        ORGANISATION_ROLE_ADMIN (2):
            Manage members and resources.
        ORGANISATION_ROLE_MEMBER (3):
            Create and manage bookings and resources.
        ORGANISATION_ROLE_VIEWER (4):
            Read-only access.
    """
    ORGANISATION_ROLE_UNSPECIFIED = 0
    ORGANISATION_ROLE_OWNER = 1
    ORGANISATION_ROLE_ADMIN = 2
    ORGANISATION_ROLE_MEMBER = 3
    ORGANISATION_ROLE_VIEWER = 4


__all__ = tuple(sorted(__protobuf__.manifest))
