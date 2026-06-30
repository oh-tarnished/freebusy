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

from freebusy.resource_v1.types import resource as fr_resource
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.resource.v1',
    manifest={
        'ListResourcesRequest',
        'ListResourcesResponse',
        'GetResourceRequest',
        'CreateResourceRequest',
        'UpdateResourceRequest',
        'ArchiveResourceRequest',
        'UnarchiveResourceRequest',
        'ListOfferingsRequest',
        'ListOfferingsResponse',
        'GetOfferingRequest',
        'CreateOfferingRequest',
        'UpdateOfferingRequest',
        'DeleteOfferingRequest',
    },
)


class ListResourcesRequest(proto.Message):
    r"""Request message for ListResources.

    Attributes:
        page_size (int):
            Maximum number of resources to return. The
            server may cap this.
        page_token (str):
            Page token from a previous ListResources
            call, for pagination.
        filter (str):
            Filter expression (AIP-160), e.g.
            ``type = RESOURCE_TYPE_ROOM``, ``state = STATE_ACTIVE``,
            ``tags:"beachfront"``, or a match on display_name.
        order_by (str):
            Sort order, e.g. "display_name" or "create_time desc".
    """

    page_size: int = proto.Field(
        proto.INT32,
        number=1,
    )
    page_token: str = proto.Field(
        proto.STRING,
        number=9,
    )
    filter: str = proto.Field(
        proto.STRING,
        number=10,
    )
    order_by: str = proto.Field(
        proto.STRING,
        number=11,
    )


class ListResourcesResponse(proto.Message):
    r"""Response message for ListResources.

    Attributes:
        resources (MutableSequence[freebusy.resource_v1.types.Resource]):
            The page of resources.
        next_page_token (str):
            Next page token, if more results remain.
            Omitted if this is the last page.
    """

    @property
    def raw_page(self):
        return self

    resources: MutableSequence[fr_resource.Resource] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=fr_resource.Resource,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=5,
    )


class GetResourceRequest(proto.Message):
    r"""Request message for GetResource.

    Attributes:
        name (str):
            The resource to retrieve.
            Format: resources/{resource}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class CreateResourceRequest(proto.Message):
    r"""Request message for CreateResource.

    Attributes:
        resource (freebusy.resource_v1.types.Resource):
            The resource to create. The name, state, and
            offerings fields are ignored.
        resource_id (str):
            Optional caller-chosen ID for the resource;
            the server generates one if unset.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
    """

    resource: fr_resource.Resource = proto.Field(
        proto.MESSAGE,
        number=1,
        message=fr_resource.Resource,
    )
    resource_id: str = proto.Field(
        proto.STRING,
        number=2,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=3,
    )


class UpdateResourceRequest(proto.Message):
    r"""Request message for UpdateResource.

    Attributes:
        resource (freebusy.resource_v1.types.Resource):
            The resource to update; its name identifies
            the target.
        update_mask (google.protobuf.field_mask_pb2.FieldMask):
            Fields to overwrite. Omit to replace all
            mutable fields.
    """

    resource: fr_resource.Resource = proto.Field(
        proto.MESSAGE,
        number=1,
        message=fr_resource.Resource,
    )
    update_mask: field_mask_pb2.FieldMask = proto.Field(
        proto.MESSAGE,
        number=2,
        message=field_mask_pb2.FieldMask,
    )


class ArchiveResourceRequest(proto.Message):
    r"""Request message for ArchiveResource.

    Attributes:
        name (str):
            The resource to archive.
            Format: resources/{resource}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class UnarchiveResourceRequest(proto.Message):
    r"""Request message for UnarchiveResource.

    Attributes:
        name (str):
            The resource to restore to the active state.
            Format: resources/{resource}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class ListOfferingsRequest(proto.Message):
    r"""Request message for ListOfferings.

    Attributes:
        parent (str):
            The parent resource whose offerings to list.
            Format: resources/{resource}
        page_size (int):
            Maximum number of offerings to return.
        page_token (str):
            Page token from a previous ListOfferings call's
            next_page_token.
        order_by (str):
            Sort order, e.g. "display_name" or "create_time desc".
        filter (str):
            Filter expression (AIP-160), e.g. a match on display_name.
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
    order_by: str = proto.Field(
        proto.STRING,
        number=4,
    )
    filter: str = proto.Field(
        proto.STRING,
        number=5,
    )


class ListOfferingsResponse(proto.Message):
    r"""Response message for ListOfferings.

    Attributes:
        offerings (MutableSequence[freebusy.resource_v1.types.Offering]):
            The page of offerings.
        next_page_token (str):
            next page token. Omitted if this is the last
            page.
    """

    @property
    def raw_page(self):
        return self

    offerings: MutableSequence[fr_resource.Offering] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=fr_resource.Offering,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )


class GetOfferingRequest(proto.Message):
    r"""Request message for GetOffering.

    Attributes:
        name (str):
            The offering to retrieve.
            Format:
            resources/{resource}/offerings/{offering}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class CreateOfferingRequest(proto.Message):
    r"""Request message for CreateOffering.

    Attributes:
        parent (str):
            The resource to attach the offering to.
            Format: resources/{resource}
        offering (freebusy.resource_v1.types.Offering):
            The offering to create. Its name field is
            ignored.
        offering_id (str):
            Optional caller-chosen ID for the offering;
            the server generates one if unset.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
    """

    parent: str = proto.Field(
        proto.STRING,
        number=1,
    )
    offering: fr_resource.Offering = proto.Field(
        proto.MESSAGE,
        number=2,
        message=fr_resource.Offering,
    )
    offering_id: str = proto.Field(
        proto.STRING,
        number=3,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=4,
    )


class UpdateOfferingRequest(proto.Message):
    r"""Request message for UpdateOffering.

    Attributes:
        offering (freebusy.resource_v1.types.Offering):
            The offering to update; its name identifies
            the target.
        update_mask (google.protobuf.field_mask_pb2.FieldMask):
            Fields to overwrite. Omit to replace all
            mutable fields.
    """

    offering: fr_resource.Offering = proto.Field(
        proto.MESSAGE,
        number=1,
        message=fr_resource.Offering,
    )
    update_mask: field_mask_pb2.FieldMask = proto.Field(
        proto.MESSAGE,
        number=2,
        message=field_mask_pb2.FieldMask,
    )


class DeleteOfferingRequest(proto.Message):
    r"""Request message for DeleteOffering.

    Attributes:
        name (str):
            The offering to delete.
            Format:
            resources/{resource}/offerings/{offering}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
