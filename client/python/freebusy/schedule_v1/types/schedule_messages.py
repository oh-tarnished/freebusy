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

from freebusy.schedule_v1.types import schedule as fs_schedule
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.schedule.v1',
    manifest={
        'GetScheduleRequest',
        'UpdateScheduleRequest',
        'ListAvailabilityExceptionsRequest',
        'ListAvailabilityExceptionsResponse',
        'GetAvailabilityExceptionRequest',
        'CreateAvailabilityExceptionRequest',
        'DeleteAvailabilityExceptionRequest',
    },
)


class GetScheduleRequest(proto.Message):
    r"""Request message for GetSchedule.

    Attributes:
        name (str):
            The schedule to read.
            Format: resources/{resource}/schedule
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class UpdateScheduleRequest(proto.Message):
    r"""Request message for UpdateSchedule. Set update_mask to the
    section(s) to replace: "recurring_rules", "buffers",
    "stay_constraints", and/or "cancellation_policy".

    Attributes:
        schedule (freebusy.schedule_v1.types.Schedule):
            The schedule to update; its name identifies
            the target.
        update_mask (google.protobuf.field_mask_pb2.FieldMask):
            Fields to overwrite. Omit to replace all
            mutable sections.
    """

    schedule: fs_schedule.Schedule = proto.Field(
        proto.MESSAGE,
        number=1,
        message=fs_schedule.Schedule,
    )
    update_mask: field_mask_pb2.FieldMask = proto.Field(
        proto.MESSAGE,
        number=2,
        message=field_mask_pb2.FieldMask,
    )


class ListAvailabilityExceptionsRequest(proto.Message):
    r"""Request message for ListAvailabilityExceptions.

    Attributes:
        parent (str):
            The resource whose exceptions to list.
            Format: resources/{resource}
        page_size (int):
            Maximum number of exceptions to return.
        page_token (str):
            Token for the page of results to return.
            Empty for the first page.
        filter (str):
            Filter expression (AIP-160), e.g.
            ``kind = EXCEPTION_KIND_CLOSURE``.
        order_by (str):
            Sort order, e.g. "window.start_time" or "create_time desc".
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


class ListAvailabilityExceptionsResponse(proto.Message):
    r"""Response message for ListAvailabilityExceptions.

    Attributes:
        availability_exceptions (MutableSequence[freebusy.schedule_v1.types.AvailabilityException]):
            The page of exceptions.
        next_page_token (str):
            Token for the next page of results. Empty if
            there are no more pages.
    """

    @property
    def raw_page(self):
        return self

    availability_exceptions: MutableSequence[fs_schedule.AvailabilityException] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=fs_schedule.AvailabilityException,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )


class GetAvailabilityExceptionRequest(proto.Message):
    r"""Request message for GetAvailabilityException.

    Attributes:
        name (str):
            The exception to retrieve. Format:
            resources/{resource}/availabilityExceptions/{availability_exception}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class CreateAvailabilityExceptionRequest(proto.Message):
    r"""Request message for CreateAvailabilityException.

    Attributes:
        parent (str):
            The resource to add the exception to.
            Format: resources/{resource}
        availability_exception (freebusy.schedule_v1.types.AvailabilityException):
            The exception to add. Its name field is
            ignored.
        availability_exception_id (str):
            Optional caller-chosen ID for the exception;
            the server generates one if unset.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
    """

    parent: str = proto.Field(
        proto.STRING,
        number=1,
    )
    availability_exception: fs_schedule.AvailabilityException = proto.Field(
        proto.MESSAGE,
        number=2,
        message=fs_schedule.AvailabilityException,
    )
    availability_exception_id: str = proto.Field(
        proto.STRING,
        number=3,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=4,
    )


class DeleteAvailabilityExceptionRequest(proto.Message):
    r"""Request message for DeleteAvailabilityException.

    Attributes:
        name (str):
            The exception to remove. Format:
            resources/{resource}/availabilityExceptions/{availability_exception}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
