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

from freebusy.organisation_v1.types import organisation as fo_organisation
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.organisation.v1',
    manifest={
        'ListOrganisationsRequest',
        'ListOrganisationsResponse',
        'GetOrganisationRequest',
        'CreateOrganisationRequest',
        'UpdateOrganisationRequest',
        'DeleteOrganisationRequest',
        'ListMembersRequest',
        'ListMembersResponse',
        'GetMemberRequest',
    },
)


class ListOrganisationsRequest(proto.Message):
    r"""Request message for ListOrganisations.

    Attributes:
        page_size (int):
            Maximum number of organisations to return.
            The server may cap this. Zero lets the server
            pick a default.
        page_token (str):
            Page token from a previous ListOrganisations call's
            next_page_token.
        filter (str):
            Filter expression over organisation fields (e.g.
            display_name).
        order_by (str):
            Sort order, e.g. "create_time desc".
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


class ListOrganisationsResponse(proto.Message):
    r"""Response message for ListOrganisations.

    Attributes:
        organisations (MutableSequence[freebusy.organisation_v1.types.Organisation]):
            The page of organisations (those the caller
            belongs to).
        next_page_token (str):
            Token to pass as page_token to retrieve the next page; empty
            when no more.
    """

    @property
    def raw_page(self):
        return self

    organisations: MutableSequence[fo_organisation.Organisation] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=fo_organisation.Organisation,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )


class GetOrganisationRequest(proto.Message):
    r"""Request message for GetOrganisation.

    Attributes:
        name (str):
            The organisation to retrieve.
            Format: organisations/{organisation}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class CreateOrganisationRequest(proto.Message):
    r"""Request message for CreateOrganisation.

    Attributes:
        organisation (freebusy.organisation_v1.types.Organisation):
            The organisation to create. The name and
            state fields are ignored. The caller becomes the
            organisation's first OWNER.
        organisation_id (str):
            Optional caller-chosen ID for the
            organisation; the server generates one if unset.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
    """

    organisation: fo_organisation.Organisation = proto.Field(
        proto.MESSAGE,
        number=1,
        message=fo_organisation.Organisation,
    )
    organisation_id: str = proto.Field(
        proto.STRING,
        number=2,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=3,
    )


class UpdateOrganisationRequest(proto.Message):
    r"""Request message for UpdateOrganisation.

    Attributes:
        organisation (freebusy.organisation_v1.types.Organisation):
            The organisation to update; its name
            identifies the target.
        update_mask (google.protobuf.field_mask_pb2.FieldMask):
            Fields to overwrite. Omit to replace all
            mutable fields.
    """

    organisation: fo_organisation.Organisation = proto.Field(
        proto.MESSAGE,
        number=1,
        message=fo_organisation.Organisation,
    )
    update_mask: field_mask_pb2.FieldMask = proto.Field(
        proto.MESSAGE,
        number=2,
        message=field_mask_pb2.FieldMask,
    )


class DeleteOrganisationRequest(proto.Message):
    r"""Request message for DeleteOrganisation.

    Attributes:
        name (str):
            The organisation to delete.
            Format: organisations/{organisation}
        force (bool):
            If true, delete the organisation and its
            members; otherwise the call fails when the
            organisation still has members.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    force: bool = proto.Field(
        proto.BOOL,
        number=2,
    )


class ListMembersRequest(proto.Message):
    r"""Request message for ListMembers.

    Attributes:
        parent (str):
            The organisation whose members to list.
            Format: organisations/{organisation}
        page_size (int):
            Maximum number of members to return. The
            server may cap this.
        page_token (str):
            Page token from a previous ListMembers call's
            next_page_token.
        filter (str):
            Filter expression (AIP-160), e.g. ``role = ADMIN`` or
            ``state = ACTIVE``.
        order_by (str):
            Sort order, e.g. "create_time desc".
    """

    parent: str = proto.Field(
        proto.STRING,
        number=1,
    )
    page_size: int = proto.Field(
        proto.INT32,
        number=2,
    )
    page_token: str = proto.Field(
        proto.STRING,
        number=3,
    )
    filter: str = proto.Field(
        proto.STRING,
        number=4,
    )
    order_by: str = proto.Field(
        proto.STRING,
        number=5,
    )


class ListMembersResponse(proto.Message):
    r"""Response message for ListMembers.

    Attributes:
        members (MutableSequence[freebusy.organisation_v1.types.Member]):
            The page of members.
        next_page_token (str):
            Token to pass as page_token to retrieve the next page; empty
            when no more.
    """

    @property
    def raw_page(self):
        return self

    members: MutableSequence[fo_organisation.Member] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=fo_organisation.Member,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )


class GetMemberRequest(proto.Message):
    r"""Request message for GetMember.

    Attributes:
        name (str):
            The member to retrieve.
            Format:
            organisations/{organisation}/members/{member}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
