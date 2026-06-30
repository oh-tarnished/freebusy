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
import google.protobuf.struct_pb2 as struct_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.organisation.v1',
    manifest={
        'Organisation',
        'Member',
    },
)


class Organisation(proto.Message):
    r"""A tenant. Organisation is the unit of multi-tenancy; the
    shell enforces isolation with row-level security keyed off the
    caller's organisation, so most resource names stay flat and the
    organisation appears explicitly only here.

    Attributes:
        name (str):
            The organisation name.
            Format: organisations/{organisation}
        display_name (str):
            Human-friendly organisation name (e.g. "Acme
            Inc.").
        slug (str):
            URL-safe slug, unique across organisations.
        billing_email (str):
            Billing contact email.
        state (freebusy.organisation_v1.types.OrganisationState):
            Lifecycle state.
        settings (google.protobuf.struct_pb2.Struct):
            Arbitrary organisation-level settings.
        member_count (int):
            Number of members across all states.
        create_time (google.protobuf.timestamp_pb2.Timestamp):
            Creation timestamp.
        update_time (google.protobuf.timestamp_pb2.Timestamp):
            Last-modification timestamp.
        etag (str):
            Opaque version for optimistic concurrency
            (AIP-154); echo on update/delete.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=3,
    )
    slug: str = proto.Field(
        proto.STRING,
        number=4,
    )
    billing_email: str = proto.Field(
        proto.STRING,
        number=5,
    )
    state: enums.OrganisationState = proto.Field(
        proto.ENUM,
        number=6,
        enum=enums.OrganisationState,
    )
    settings: struct_pb2.Struct = proto.Field(
        proto.MESSAGE,
        number=7,
        message=struct_pb2.Struct,
    )
    member_count: int = proto.Field(
        proto.INT64,
        number=8,
    )
    create_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=9,
        message=timestamp_pb2.Timestamp,
    )
    update_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=10,
        message=timestamp_pb2.Timestamp,
    )
    etag: str = proto.Field(
        proto.STRING,
        number=11,
    )


class Member(proto.Message):
    r"""The membership of a user in an organisation, with their role.

    Attributes:
        name (str):
            The member name.
            Format:
            organisations/{organisation}/members/{member}
        user (str):
            The user, once the invite is accepted.
            Format: users/{user}
        email (str):
            The invited email address.
        display_name (str):
            Cached display name of the member.
        role (freebusy.organisation_v1.types.OrganisationRole):
            The member's role in the organisation.
        state (freebusy.organisation_v1.types.MemberState):
            Confirmation state of the membership.
        inviter (str):
            The user who issued the invite.
            Format: users/{user}
        create_time (google.protobuf.timestamp_pb2.Timestamp):
            Creation timestamp (when the invite was
            created).
        update_time (google.protobuf.timestamp_pb2.Timestamp):
            Last-modification timestamp.
        etag (str):
            Opaque version for optimistic concurrency
            (AIP-154); echo on update/delete.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    user: str = proto.Field(
        proto.STRING,
        number=3,
    )
    email: str = proto.Field(
        proto.STRING,
        number=4,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=5,
    )
    role: enums.OrganisationRole = proto.Field(
        proto.ENUM,
        number=6,
        enum=enums.OrganisationRole,
    )
    state: enums.MemberState = proto.Field(
        proto.ENUM,
        number=7,
        enum=enums.MemberState,
    )
    inviter: str = proto.Field(
        proto.STRING,
        number=8,
    )
    create_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=9,
        message=timestamp_pb2.Timestamp,
    )
    update_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=10,
        message=timestamp_pb2.Timestamp,
    )
    etag: str = proto.Field(
        proto.STRING,
        number=11,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
