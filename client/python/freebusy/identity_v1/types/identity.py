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

import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.identity.v1',
    manifest={
        'User',
        'MembershipSummary',
        'GetUserRequest',
        'UpdateUserRequest',
        'ListUsersRequest',
        'ListUsersResponse',
    },
)


class User(proto.Message):
    r"""A signed-in person. Identity is deliberately thin: actual
    login is an OIDC redirect flow handled over plain HTTP by the
    IdP, so most of "auth" never appears as an RPC. Email and
    identity come from the IdP and are read-only here; only profile
    preferences are editable.

    Attributes:
        name (str):
            The user name. The alias "users/me" resolves
            to the caller. Format: users/{user}
        email (str):
            Email address, sourced from the IdP.
        display_name (str):
            Display name.
        avatar_url (str):
            Avatar image URL.
        locale (str):
            BCP 47 locale (e.g. "en-US").
        time_zone (str):
            IANA time zone (e.g. "America/New_York").
        memberships (MutableSequence[freebusy.identity_v1.types.MembershipSummary]):
            The organisations this user belongs to, with
            role.
        create_time (google.protobuf.timestamp_pb2.Timestamp):
            Creation timestamp.
        update_time (google.protobuf.timestamp_pb2.Timestamp):
            Last-modification timestamp.
        etag (str):
            Opaque version for optimistic concurrency
            (AIP-154); echo on update.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    email: str = proto.Field(
        proto.STRING,
        number=3,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=4,
    )
    avatar_url: str = proto.Field(
        proto.STRING,
        number=5,
    )
    locale: str = proto.Field(
        proto.STRING,
        number=6,
    )
    time_zone: str = proto.Field(
        proto.STRING,
        number=7,
    )
    memberships: MutableSequence['MembershipSummary'] = proto.RepeatedField(
        proto.MESSAGE,
        number=8,
        message='MembershipSummary',
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


class MembershipSummary(proto.Message):
    r"""A compact view of an organisation the user belongs to.

    Attributes:
        organisation (str):
            The organisation.
            Format: organisations/{organisation}
        org_display_name (str):
            Cached display name of the organisation.
        role (str):
            The user's role in the organisation (an
            OrganisationRole value name).
    """

    organisation: str = proto.Field(
        proto.STRING,
        number=1,
    )
    org_display_name: str = proto.Field(
        proto.STRING,
        number=2,
    )
    role: str = proto.Field(
        proto.STRING,
        number=3,
    )


class GetUserRequest(proto.Message):
    r"""Request message for GetUser.

    Attributes:
        name (str):
            The user to retrieve. Use "users/me" for the
            signed-in caller. Format: users/{user}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class UpdateUserRequest(proto.Message):
    r"""Request message for UpdateUser.

    Attributes:
        user (freebusy.identity_v1.types.User):
            The user to update; its name identifies the
            target (typically "users/me").
        update_mask (google.protobuf.field_mask_pb2.FieldMask):
            Fields to overwrite. Omit to replace all
            mutable profile fields.
    """

    user: 'User' = proto.Field(
        proto.MESSAGE,
        number=1,
        message='User',
    )
    update_mask: field_mask_pb2.FieldMask = proto.Field(
        proto.MESSAGE,
        number=2,
        message=field_mask_pb2.FieldMask,
    )


class ListUsersRequest(proto.Message):
    r"""Request message for ListUsers. Users are global, so the visible set
    is every user sharing at least one organisation with the caller; use
    ``organisation = "organisations/{organisation}"`` in filter to
    narrow to one organisation, or OrganisationService.ListMembers for a
    single organisation's roster with roles.

    Attributes:
        page_size (int):
            Maximum number of users to return. The server
            may cap this.
        page_token (str):
            Page token from a previous ListUsers call's next_page_token.
        filter (str):
            Filter expression (AIP-160), e.g.
            ``organisation = "organisations/7"`` or a match on
            display_name.
        order_by (str):
            Sort order, e.g. "display_name" or "create_time desc".
    """

    page_size: int = proto.Field(
        proto.INT32,
        number=1,
    )
    page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )
    filter: str = proto.Field(
        proto.STRING,
        number=3,
    )
    order_by: str = proto.Field(
        proto.STRING,
        number=4,
    )


class ListUsersResponse(proto.Message):
    r"""Response message for ListUsers.

    Attributes:
        users (MutableSequence[freebusy.identity_v1.types.User]):
            The page of users.
        next_page_token (str):
            Token to pass as page_token to retrieve the next page; empty
            when no more.
    """

    @property
    def raw_page(self):
        return self

    users: MutableSequence['User'] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message='User',
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
