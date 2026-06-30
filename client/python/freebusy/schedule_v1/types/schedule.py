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

from freebusy.schedule_v1.types import enums as fs_enums
import freebusy.shared.v1.enums_pb2 as enums_pb2  # type: ignore
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
import google.protobuf.duration_pb2 as duration_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.schedule.v1',
    manifest={
        'RecurringRule',
        'BufferSettings',
        'StayConstraints',
        'AvailabilityException',
        'Schedule',
        'CancellationPolicy',
        'RefundTier',
    },
)


class RecurringRule(proto.Message):
    r"""A recurring availability window expressed as an RRULE plus a
    daily open span. The freebusy engine expands these against the
    resource's timezone.

    Attributes:
        rrule (str):
            RFC 5545 RRULE, e.g.
            "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR".
        opens (str):
            Local time-of-day the window opens, 24-hour
            "HH:MM" (e.g. "09:00"). Unset (with closes also
            unset) means open the whole day.
        closes (str):
            Local time-of-day the window closes, 24-hour
            "HH:MM" (e.g. "17:00"). A closes at or before
            opens means the span crosses midnight into the
            next day (e.g. opens "22:00", closes "02:00").
    """

    rrule: str = proto.Field(
        proto.STRING,
        number=1,
    )
    opens: str = proto.Field(
        proto.STRING,
        number=2,
    )
    closes: str = proto.Field(
        proto.STRING,
        number=3,
    )


class BufferSettings(proto.Message):
    r"""Buffer and notice settings applied around bookings.

    Attributes:
        start_delta (google.protobuf.duration_pb2.Duration):
            Prep time reserved before each booking.
        end_delta (google.protobuf.duration_pb2.Duration):
            Turnover / cleaning time reserved after each
            booking.
        min_notice (google.protobuf.duration_pb2.Duration):
            Minimum lead time between now and a booking's
            start.
        max_advance (google.protobuf.duration_pb2.Duration):
            How far into the future bookings may be made.
        gap (google.protobuf.duration_pb2.Duration):
            Minimum gap enforced between two adjacent
            bookings.
    """

    start_delta: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=1,
        message=duration_pb2.Duration,
    )
    end_delta: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=2,
        message=duration_pb2.Duration,
    )
    min_notice: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=3,
        message=duration_pb2.Duration,
    )
    max_advance: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=4,
        message=duration_pb2.Duration,
    )
    gap: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=5,
        message=duration_pb2.Duration,
    )


class StayConstraints(proto.Message):
    r"""Stay rules that affect bookability for NIGHTLY resources.

    Attributes:
        min_nights (int):
            Minimum number of nights per booking.
        max_nights (int):
            Maximum number of nights per booking. Zero
            means no maximum.
        checkin_weekdays (MutableSequence[freebusy.shared.v1.enums_pb2.Weekday]):
            Allowed check-in weekdays. Empty means any
            day.
        checkout_weekdays (MutableSequence[freebusy.shared.v1.enums_pb2.Weekday]):
            Allowed check-out weekdays. Empty means any
            day.
        advance_min_days (int):
            Earliest a stay may begin, in days from now.
        advance_max_days (int):
            Latest a stay may begin, in days from now.
            Zero means no limit.
    """

    min_nights: int = proto.Field(
        proto.INT32,
        number=1,
    )
    max_nights: int = proto.Field(
        proto.INT32,
        number=2,
    )
    checkin_weekdays: MutableSequence[enums_pb2.Weekday] = proto.RepeatedField(
        proto.ENUM,
        number=3,
        enum=enums_pb2.Weekday,
    )
    checkout_weekdays: MutableSequence[enums_pb2.Weekday] = proto.RepeatedField(
        proto.ENUM,
        number=4,
        enum=enums_pb2.Weekday,
    )
    advance_min_days: int = proto.Field(
        proto.INT32,
        number=5,
    )
    advance_max_days: int = proto.Field(
        proto.INT32,
        number=6,
    )


class AvailabilityException(proto.Message):
    r"""An override of a resource's normal hours on a specific span:
    a blackout / holiday closure, or extra hours beyond the
    recurring rules.

    This message has `oneof`_ fields (mutually exclusive fields).
    For each oneof, at most one member field can be set at the same time.
    Setting any member of the oneof automatically clears all other
    members.

    .. _oneof: https://proto-plus-python.readthedocs.io/en/stable/fields.html#oneofs-mutually-exclusive-fields

    Attributes:
        name (str):
            The exception name. Format:
            resources/{resource}/availabilityExceptions/{availability_exception}
        kind (freebusy.schedule_v1.types.ExceptionKind):
            Whether this span closes the resource or adds
            extra availability.
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            An exact time window, the natural form for TIME_SLOT
            resources.

            This field is a member of `oneof`_ ``span``.
        date_range (freebusy.shared.v1.types_pb2.DateRange):
            A range of whole calendar dates in the
            resource's timezone, the natural form for
            NIGHTLY blackouts (e.g. "closed Dec 24 through
            Dec 26").

            This field is a member of `oneof`_ ``span``.
        reason (str):
            Human-readable reason (e.g. "Public
            holiday").
        create_time (google.protobuf.timestamp_pb2.Timestamp):
            Creation timestamp.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    kind: fs_enums.ExceptionKind = proto.Field(
        proto.ENUM,
        number=3,
        enum=fs_enums.ExceptionKind,
    )
    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=4,
        oneof='span',
        message=types_pb2.TimeWindow,
    )
    date_range: types_pb2.DateRange = proto.Field(
        proto.MESSAGE,
        number=7,
        oneof='span',
        message=types_pb2.DateRange,
    )
    reason: str = proto.Field(
        proto.STRING,
        number=5,
    )
    create_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=6,
        message=timestamp_pb2.Timestamp,
    )


class Schedule(proto.Message):
    r"""Aggregate read view of a resource's availability
    configuration: the inputs the freebusy engine consumes. Modeled
    as a singleton resource, one per resource.

    Attributes:
        name (str):
            The schedule name.
            Format: resources/{resource}/schedule
        recurring_rules (MutableSequence[freebusy.schedule_v1.types.RecurringRule]):
            Recurring working hours.
        buffers (freebusy.schedule_v1.types.BufferSettings):
            Buffer and notice settings.
        stay_constraints (freebusy.schedule_v1.types.StayConstraints):
            Stay rules (NIGHTLY resources).
        exceptions (MutableSequence[str]):
            Resource names of the active exceptions; manage them with
            the AvailabilityException standard methods. Format:
            resources/{resource}/availabilityExceptions/{availability_exception}
        cancellation_policy (freebusy.schedule_v1.types.CancellationPolicy):
            Refund rules applied when a booking on this
            resource is cancelled. Unset means cancellations
            are non-refundable by default.
        etag (str):
            Opaque version for optimistic concurrency
            (AIP-154); echo on update.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    recurring_rules: MutableSequence['RecurringRule'] = proto.RepeatedField(
        proto.MESSAGE,
        number=2,
        message='RecurringRule',
    )
    buffers: 'BufferSettings' = proto.Field(
        proto.MESSAGE,
        number=3,
        message='BufferSettings',
    )
    stay_constraints: 'StayConstraints' = proto.Field(
        proto.MESSAGE,
        number=4,
        message='StayConstraints',
    )
    exceptions: MutableSequence[str] = proto.RepeatedField(
        proto.STRING,
        number=5,
    )
    cancellation_policy: 'CancellationPolicy' = proto.Field(
        proto.MESSAGE,
        number=7,
        message='CancellationPolicy',
    )
    etag: str = proto.Field(
        proto.STRING,
        number=6,
    )


class CancellationPolicy(proto.Message):
    r"""Refund rules graded by how far ahead of a booking's start it
    is cancelled.

    Attributes:
        tiers (MutableSequence[freebusy.schedule_v1.types.RefundTier]):
            Ordered refund tiers. For a given cancellation the tier with
            the largest ``cutoff`` that is still satisfied (cancelled at
            least ``cutoff`` before the booking start) determines the
            refund. If no tier is satisfied, the booking is
            non-refundable.
    """

    tiers: MutableSequence['RefundTier'] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message='RefundTier',
    )


class RefundTier(proto.Message):
    r"""One tier of a CancellationPolicy: cancel at least ``cutoff`` before
    the booking start to receive ``refund_percent`` of the total back.

    Attributes:
        cutoff (google.protobuf.duration_pb2.Duration):
            Minimum lead time before the booking start
            for this tier to apply (e.g. 48h).
        refund_percent (int):
            Percentage of the booking total refunded at
            this tier (0-100).
    """

    cutoff: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=1,
        message=duration_pb2.Duration,
    )
    refund_percent: int = proto.Field(
        proto.INT32,
        number=2,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
