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

from freebusy.availability_v1.types import enums as fa_enums
import freebusy.shared.v1.enums_pb2 as enums_pb2  # type: ignore
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
import google.protobuf.duration_pb2 as duration_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.type.date_pb2 as date_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.availability.v1',
    manifest={
        'Slot',
        'NightAvailability',
        'BookableRange',
        'ResourceAvailability',
        'ComputeAvailabilityRequest',
        'ComputeAvailabilityResponse',
        'CheckAvailabilityRequest',
        'UnbookableReason',
        'CheckAvailabilityResponse',
        'ComputeBookableRangesRequest',
        'ComputeBookableRangesResponse',
        'BatchComputeAvailabilityRequest',
        'BatchComputeAvailabilityResponse',
        'SearchAvailabilityRequest',
        'AvailabilityMatch',
        'SearchAvailabilityResponse',
    },
)


class Slot(proto.Message):
    r"""A discrete bookable time slot, produced for TIME_SLOT resources.

    Attributes:
        start_time (google.protobuf.timestamp_pb2.Timestamp):
            Inclusive start of the slot.
        end_time (google.protobuf.timestamp_pb2.Timestamp):
            Exclusive end of the slot (start +
            offering/requested duration).
        free_count (int):
            Number of units free in this slot (capacity
            minus overlapping bookings).
        bookable (bool):
            Whether the slot can actually be booked (free
            and passes policy).
        price (google.type.money_pb2.Money):
            Price for booking this slot, when an offering
            was supplied.
    """

    start_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=1,
        message=timestamp_pb2.Timestamp,
    )
    end_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=2,
        message=timestamp_pb2.Timestamp,
    )
    free_count: int = proto.Field(
        proto.INT32,
        number=3,
    )
    bookable: bool = proto.Field(
        proto.BOOL,
        number=4,
    )
    price: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=5,
        message=money_pb2.Money,
    )


class NightAvailability(proto.Message):
    r"""Per-night availability, produced for NIGHTLY resources.

    Attributes:
        night (google.type.date_pb2.Date):
            The night, in the resource's local timezone.
        free_units (int):
            Number of units of the pool free that night.
        closed (bool):
            Whether the resource is closed that night
            (exception/blackout).
        price (google.type.money_pb2.Money):
            Nightly price, when an offering was supplied.
    """

    night: date_pb2.Date = proto.Field(
        proto.MESSAGE,
        number=1,
        message=date_pb2.Date,
    )
    free_units: int = proto.Field(
        proto.INT32,
        number=2,
    )
    closed: bool = proto.Field(
        proto.BOOL,
        number=3,
    )
    price: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=4,
        message=money_pb2.Money,
    )


class BookableRange(proto.Message):
    r"""A contiguous span that is bookable end to end.

    Attributes:
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            The bookable span.
        bookable (bool):
            Whether the whole span is bookable.
    """

    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=1,
        message=types_pb2.TimeWindow,
    )
    bookable: bool = proto.Field(
        proto.BOOL,
        number=2,
    )


class ResourceAvailability(proto.Message):
    r"""Availability for one resource, used in batch responses.

    Attributes:
        resource (str):
            The resource these results are for.
            Format: resources/{resource}
        mode (freebusy.shared.v1.enums_pb2.BookingMode):
            Which shape is populated, matching the resource's
            booking_mode.
        slots (MutableSequence[freebusy.availability_v1.types.Slot]):
            Slots, when mode is TIME_SLOT.
        nights (MutableSequence[freebusy.availability_v1.types.NightAvailability]):
            Per-night availability, when mode is NIGHTLY.
    """

    resource: str = proto.Field(
        proto.STRING,
        number=1,
    )
    mode: enums_pb2.BookingMode = proto.Field(
        proto.ENUM,
        number=2,
        enum=enums_pb2.BookingMode,
    )
    slots: MutableSequence['Slot'] = proto.RepeatedField(
        proto.MESSAGE,
        number=3,
        message='Slot',
    )
    nights: MutableSequence['NightAvailability'] = proto.RepeatedField(
        proto.MESSAGE,
        number=4,
        message='NightAvailability',
    )


class ComputeAvailabilityRequest(proto.Message):
    r"""Request message for ComputeAvailability.

    This message has `oneof`_ fields (mutually exclusive fields).
    For each oneof, at most one member field can be set at the same time.
    Setting any member of the oneof automatically clears all other
    members.

    .. _oneof: https://proto-plus-python.readthedocs.io/en/stable/fields.html#oneofs-mutually-exclusive-fields

    Attributes:
        resource (str):
            The resource to compute availability for.
            Format: resources/{resource}
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            An exact time window, the natural form for TIME_SLOT
            resources.

            This field is a member of `oneof`_ ``period``.
        date_range (freebusy.shared.v1.types_pb2.DateRange):
            A calendar-date range in the resource's timezone, the
            natural form for NIGHTLY resources; end_date is the
            check-out date.

            This field is a member of `oneof`_ ``period``.
        duration (google.protobuf.duration_pb2.Duration):
            Slot length for TIME_SLOT resources. Ignored when offering
            is set or for NIGHTLY resources.
        offering (str):
            Offering to derive duration and price from.
            Takes precedence over duration. Format:
            resources/{resource}/offerings/{offering}
        units (int):
            Number of units required to be free. Defaults
            to 1.
    """

    resource: str = proto.Field(
        proto.STRING,
        number=1,
    )
    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=2,
        oneof='period',
        message=types_pb2.TimeWindow,
    )
    date_range: types_pb2.DateRange = proto.Field(
        proto.MESSAGE,
        number=6,
        oneof='period',
        message=types_pb2.DateRange,
    )
    duration: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=3,
        message=duration_pb2.Duration,
    )
    offering: str = proto.Field(
        proto.STRING,
        number=4,
    )
    units: int = proto.Field(
        proto.INT32,
        number=5,
    )


class ComputeAvailabilityResponse(proto.Message):
    r"""Response message for ComputeAvailability.

    Attributes:
        mode (freebusy.shared.v1.enums_pb2.BookingMode):
            Which shape is populated, matching the resource's
            booking_mode.
        slots (MutableSequence[freebusy.availability_v1.types.Slot]):
            Slots, when mode is TIME_SLOT.
        nights (MutableSequence[freebusy.availability_v1.types.NightAvailability]):
            Per-night availability, when mode is NIGHTLY.
    """

    mode: enums_pb2.BookingMode = proto.Field(
        proto.ENUM,
        number=1,
        enum=enums_pb2.BookingMode,
    )
    slots: MutableSequence['Slot'] = proto.RepeatedField(
        proto.MESSAGE,
        number=2,
        message='Slot',
    )
    nights: MutableSequence['NightAvailability'] = proto.RepeatedField(
        proto.MESSAGE,
        number=3,
        message='NightAvailability',
    )


class CheckAvailabilityRequest(proto.Message):
    r"""Request message for CheckAvailability.

    This message has `oneof`_ fields (mutually exclusive fields).
    For each oneof, at most one member field can be set at the same time.
    Setting any member of the oneof automatically clears all other
    members.

    .. _oneof: https://proto-plus-python.readthedocs.io/en/stable/fields.html#oneofs-mutually-exclusive-fields

    Attributes:
        resource (str):
            The resource to test.
            Format: resources/{resource}
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            An exact time window, the natural form for TIME_SLOT
            resources.

            This field is a member of `oneof`_ ``period``.
        date_range (freebusy.shared.v1.types_pb2.DateRange):
            A calendar-date range in the resource's timezone, the
            natural form for NIGHTLY stays; end_date is the check-out
            date.

            This field is a member of `oneof`_ ``period``.
        units (int):
            Number of units required to be free. Defaults
            to 1.
        offering (str):
            Offering whose duration/rules apply, when
            relevant. Format:
            resources/{resource}/offerings/{offering}
    """

    resource: str = proto.Field(
        proto.STRING,
        number=1,
    )
    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=2,
        oneof='period',
        message=types_pb2.TimeWindow,
    )
    date_range: types_pb2.DateRange = proto.Field(
        proto.MESSAGE,
        number=5,
        oneof='period',
        message=types_pb2.DateRange,
    )
    units: int = proto.Field(
        proto.INT32,
        number=3,
    )
    offering: str = proto.Field(
        proto.STRING,
        number=4,
    )


class UnbookableReason(proto.Message):
    r"""A reason a span is not bookable, as a machine-readable code
    plus a human-readable explanation. Clients should branch on
    code, never on message.

    Attributes:
        code (freebusy.availability_v1.types.Code):
            Why the span is not bookable.
        message (str):
            Human-readable explanation suitable for
            display, not for parsing.
    """

    code: fa_enums.Code = proto.Field(
        proto.ENUM,
        number=1,
        enum=fa_enums.Code,
    )
    message: str = proto.Field(
        proto.STRING,
        number=2,
    )


class CheckAvailabilityResponse(proto.Message):
    r"""Response message for CheckAvailability.

    Attributes:
        bookable (bool):
            Whether the span is bookable.
        free_count (int):
            Free units across the span (the minimum over
            the span).
        reasons (MutableSequence[freebusy.availability_v1.types.UnbookableReason]):
            Why the span is not bookable, when bookable
            is false.
    """

    bookable: bool = proto.Field(
        proto.BOOL,
        number=1,
    )
    free_count: int = proto.Field(
        proto.INT32,
        number=2,
    )
    reasons: MutableSequence['UnbookableReason'] = proto.RepeatedField(
        proto.MESSAGE,
        number=3,
        message='UnbookableReason',
    )


class ComputeBookableRangesRequest(proto.Message):
    r"""Request message for ComputeBookableRanges.

    This message has `oneof`_ fields (mutually exclusive fields).
    For each oneof, at most one member field can be set at the same time.
    Setting any member of the oneof automatically clears all other
    members.

    .. _oneof: https://proto-plus-python.readthedocs.io/en/stable/fields.html#oneofs-mutually-exclusive-fields

    Attributes:
        resource (str):
            The resource to compute bookable ranges for.
            Format: resources/{resource}
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            An exact time window, the natural form for TIME_SLOT
            resources.

            This field is a member of `oneof`_ ``period``.
        date_range (freebusy.shared.v1.types_pb2.DateRange):
            A calendar-date range in the resource's
            timezone, the natural form for NIGHTLY
            resources.

            This field is a member of `oneof`_ ``period``.
        duration (google.protobuf.duration_pb2.Duration):
            Minimum span length for TIME_SLOT resources.
        offering (str):
            Offering to derive duration/rules from.
            Format:
            resources/{resource}/offerings/{offering}
        units (int):
            Number of units required to be free. Defaults
            to 1.
    """

    resource: str = proto.Field(
        proto.STRING,
        number=1,
    )
    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=2,
        oneof='period',
        message=types_pb2.TimeWindow,
    )
    date_range: types_pb2.DateRange = proto.Field(
        proto.MESSAGE,
        number=6,
        oneof='period',
        message=types_pb2.DateRange,
    )
    duration: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=3,
        message=duration_pb2.Duration,
    )
    offering: str = proto.Field(
        proto.STRING,
        number=4,
    )
    units: int = proto.Field(
        proto.INT32,
        number=5,
    )


class ComputeBookableRangesResponse(proto.Message):
    r"""Response message for ComputeBookableRanges.

    Attributes:
        ranges (MutableSequence[freebusy.availability_v1.types.BookableRange]):
            The bookable ranges within the window.
    """

    ranges: MutableSequence['BookableRange'] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message='BookableRange',
    )


class BatchComputeAvailabilityRequest(proto.Message):
    r"""Request message for BatchComputeAvailability. Each entry is a
    full ComputeAvailabilityRequest (AIP-231), so per-resource
    duration, offering, and units all work in batch exactly as they
    do in the single call.

    Attributes:
        requests (MutableSequence[freebusy.availability_v1.types.ComputeAvailabilityRequest]):
            The individual compute requests. Results are
            returned in the same order.
    """

    requests: MutableSequence['ComputeAvailabilityRequest'] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message='ComputeAvailabilityRequest',
    )


class BatchComputeAvailabilityResponse(proto.Message):
    r"""Response message for BatchComputeAvailability.

    Attributes:
        resources (MutableSequence[freebusy.availability_v1.types.ResourceAvailability]):
            Availability per request, in request order.
    """

    resources: MutableSequence['ResourceAvailability'] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message='ResourceAvailability',
    )


class SearchAvailabilityRequest(proto.Message):
    r"""Request message for SearchAvailability. Sweeps the catalog
    for resources that are bookable over a period for a given party
    size, narrowed by a resource filter and sorted for presentation.
    This is the storefront query: one call returns the matching
    resources with a lead price, rather than the caller listing
    resources and computing availability for each.

    This message has `oneof`_ fields (mutually exclusive fields).
    For each oneof, at most one member field can be set at the same time.
    Setting any member of the oneof automatically clears all other
    members.

    .. _oneof: https://proto-plus-python.readthedocs.io/en/stable/fields.html#oneofs-mutually-exclusive-fields

    Attributes:
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            An exact time window, the natural form for TIME_SLOT
            resources.

            This field is a member of `oneof`_ ``period``.
        date_range (freebusy.shared.v1.types_pb2.DateRange):
            A calendar-date range in each resource's timezone, the
            natural form for NIGHTLY resources; end_date is the
            check-out date.

            This field is a member of `oneof`_ ``period``.
        units (int):
            Number of units / party size required free.
            Defaults to 1.
        filter (str):
            Filter (AIP-160) over resource fields to narrow the catalog,
            e.g. ``type = RESOURCE_TYPE_ROOM``, ``tags:"beachfront"``,
            or a display_name match.
        order_by (str):
            Sort order for matches, e.g. "price" or
            "price desc". Defaults to price ascending.
        page_size (int):
            Maximum number of matches to return. The
            server may cap this.
        page_token (str):
            Page token from a previous SearchAvailability call's
            next_page_token.
        include_unavailable (bool):
            If true, include resources that matched the
            filter but are not bookable for the period (with
            bookable=false), instead of dropping them.
    """

    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=1,
        oneof='period',
        message=types_pb2.TimeWindow,
    )
    date_range: types_pb2.DateRange = proto.Field(
        proto.MESSAGE,
        number=2,
        oneof='period',
        message=types_pb2.DateRange,
    )
    units: int = proto.Field(
        proto.INT32,
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
    page_size: int = proto.Field(
        proto.INT32,
        number=6,
    )
    page_token: str = proto.Field(
        proto.STRING,
        number=7,
    )
    include_unavailable: bool = proto.Field(
        proto.BOOL,
        number=8,
    )


class AvailabilityMatch(proto.Message):
    r"""One resource matched by SearchAvailability, with a lead price
    for the period. Detailed slots/nights are fetched per resource
    via ComputeAvailability.

    Attributes:
        resource (str):
            The matching resource.
            Format: resources/{resource}
        display_name (str):
            Cached display name of the resource, for
            convenience.
        mode (freebusy.shared.v1.enums_pb2.BookingMode):
            The resource's booking mode.
        bookable (bool):
            Whether the resource is bookable for the
            requested period and units.
        price (google.type.money_pb2.Money):
            Lead price for the requested period: the stay total for
            NIGHTLY, or the slot price for TIME_SLOT. Used for sorting
            and display.
    """

    resource: str = proto.Field(
        proto.STRING,
        number=1,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=2,
    )
    mode: enums_pb2.BookingMode = proto.Field(
        proto.ENUM,
        number=3,
        enum=enums_pb2.BookingMode,
    )
    bookable: bool = proto.Field(
        proto.BOOL,
        number=4,
    )
    price: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=5,
        message=money_pb2.Money,
    )


class SearchAvailabilityResponse(proto.Message):
    r"""Response message for SearchAvailability.

    Attributes:
        matches (MutableSequence[freebusy.availability_v1.types.AvailabilityMatch]):
            The matching resources, ordered per order_by.
        next_page_token (str):
            Token to pass as page_token to retrieve the next page; empty
            when no more.
    """

    @property
    def raw_page(self):
        return self

    matches: MutableSequence['AvailabilityMatch'] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message='AvailabilityMatch',
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
