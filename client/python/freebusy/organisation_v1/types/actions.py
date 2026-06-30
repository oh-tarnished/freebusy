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

from freebusy.organisation_v1.types import enums
from freebusy.organisation_v1.types import organisation
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.organisation.v1',
    manifest={
        'InviteMemberRequest',
        'InviteMemberResponse',
        'UpdateMemberRequest',
        'DeleteMemberRequest',
    },
)


class InviteMemberRequest(proto.Message):
    r"""Request message for InviteMember.

    Attributes:
        parent (str):
            The organisation to invite the member to.
            Format: organisations/{organisation}
        email (str):
            Email address to invite.
        role (freebusy.organisation_v1.types.OrganisationRole):
            Role to grant on acceptance.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
    """

    parent: str = proto.Field(
        proto.STRING,
        number=1,
    )
    email: str = proto.Field(
        proto.STRING,
        number=2,
    )
    role: enums.OrganisationRole = proto.Field(
        proto.ENUM,
        number=3,
        enum=enums.OrganisationRole,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=4,
    )


class InviteMemberResponse(proto.Message):
    r"""Response message for InviteMember.

    Attributes:
        member (freebusy.organisation_v1.types.Member):
            The created member (in INVITED state).
    """

    member: organisation.Member = proto.Field(
        proto.MESSAGE,
        number=1,
        message=organisation.Member,
    )


class UpdateMemberRequest(proto.Message):
    r"""Request message for UpdateMember. The role is the only mutable
    field; set update_mask to "role" to change it.

    Attributes:
        member (freebusy.organisation_v1.types.Member):
            The member to update; its name identifies the
            target.
        update_mask (google.protobuf.field_mask_pb2.FieldMask):
            Fields to overwrite. Omit to replace all
            mutable fields.
    """

    member: organisation.Member = proto.Field(
        proto.MESSAGE,
        number=1,
        message=organisation.Member,
    )
    update_mask: field_mask_pb2.FieldMask = proto.Field(
        proto.MESSAGE,
        number=2,
        message=field_mask_pb2.FieldMask,
    )


class DeleteMemberRequest(proto.Message):
    r"""Request message for DeleteMember.

    Attributes:
        name (str):
            The member to remove.
            Format:
            organisations/{organisation}/members/{member}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
