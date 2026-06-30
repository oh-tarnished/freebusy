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

from freebusy.booking_v1.types import enums
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
import google.protobuf.duration_pb2 as duration_pb2  # type: ignore
import google.protobuf.struct_pb2 as struct_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.booking.v1',
    manifest={
        'Booking',
    },
)


class Booking(proto.Message):
    r"""A reservation against a resource. The hold lifecycle lives here as
    states rather than a separate service: CreateBooking places a
    PENDING_HOLD, confirmation flips it to CONFIRMED, and an internal
    sweeper expires holds that are never confirmed.

    Attributes:
        name (str):
            The booking name.
            Format: bookings/{booking}
        resource (str):
            The resource being booked.
            Format: resources/{resource}
        offering (str):
            The offering being booked, when applicable.
            Format:
            resources/{resource}/offerings/{offering}
        customer (str):
            The user the booking is for.
            Format: users/{user}
        contact (freebusy.shared.v1.types_pb2.Contact):
            Contact details for the booker. Required when ``customer``
            is unset (a guest / walk-in booking); when ``customer`` is
            set these supplement or override the user's profile contact
            for this booking.
        units (int):
            Number of units / party size reserved.
            Defaults to 1.
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            The reserved span. For NIGHTLY resources this
            spans check-in to check-out.
        assigned_unit (str):
            Which specific unit of the pool was assigned
            (the shell's atomic pick).
        state (freebusy.booking_v1.types.BookingState):
            Current lifecycle state.
        hold_expire_time (google.protobuf.timestamp_pb2.Timestamp):
            When the pending hold lapses, if not
            confirmed first.
        price (google.type.money_pb2.Money):
            Computed subtotal before discounts.
        promo_code (str):
            The promo code to apply to this booking, set at creation, if
            any. Format: promo-codes/{promo_code}
        discount (google.type.money_pb2.Money):
            Discount applied from the promo code.
        total (google.type.money_pb2.Money):
            Final total after discounts.
        price_components (MutableSequence[freebusy.shared.v1.types_pb2.PriceComponent]):
            Itemized breakdown behind the total: the base charge, each
            fee and tax, and each discount, as signed lines. ``price``
            is the TYPE_BASE subtotal and ``total`` is the sum of every
            component; these lines expose the fees and taxes in between.
            Empty for simple bookings with no fees or taxes configured.
        notes (str):
            Free-form notes on the booking.
        attributes (google.protobuf.struct_pb2.Struct):
            Arbitrary attributes.
        cancel_reason (freebusy.booking_v1.types.CancelReason):
            Why the booking was cancelled, when state is
            CANCELLED.
        create_time (google.protobuf.timestamp_pb2.Timestamp):
            Creation timestamp (when the hold was
            placed).
        update_time (google.protobuf.timestamp_pb2.Timestamp):
            Last-modification timestamp.
        confirm_time (google.protobuf.timestamp_pb2.Timestamp):
            When the booking was confirmed, if at all.
        cancel_time (google.protobuf.timestamp_pb2.Timestamp):
            When the booking was cancelled, if at all.
        refund_amount (google.type.money_pb2.Money):
            Amount refunded on cancellation, computed
            from the resource's cancellation policy and how
            far ahead of the booking start it was cancelled.
            Set only once the booking is CANCELLED. Use
            PreviewCancellation to see this before
            committing.
        refund_percent (int):
            Percentage of the total that ``refund_amount`` represents
            (0-100).
        hold_ttl (google.protobuf.duration_pb2.Duration):
            Requested time-to-live of the hold, set at creation. The
            server caps this and reflects the effective expiry in
            hold_expire_time.
        etag (str):
            Opaque version for optimistic concurrency
            (AIP-154); echo on update/delete.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    resource: str = proto.Field(
        proto.STRING,
        number=3,
    )
    offering: str = proto.Field(
        proto.STRING,
        number=4,
    )
    customer: str = proto.Field(
        proto.STRING,
        number=5,
    )
    contact: types_pb2.Contact = proto.Field(
        proto.MESSAGE,
        number=24,
        message=types_pb2.Contact,
    )
    units: int = proto.Field(
        proto.INT32,
        number=6,
    )
    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=7,
        message=types_pb2.TimeWindow,
    )
    assigned_unit: str = proto.Field(
        proto.STRING,
        number=8,
    )
    state: enums.BookingState = proto.Field(
        proto.ENUM,
        number=9,
        enum=enums.BookingState,
    )
    hold_expire_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=10,
        message=timestamp_pb2.Timestamp,
    )
    price: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=11,
        message=money_pb2.Money,
    )
    promo_code: str = proto.Field(
        proto.STRING,
        number=12,
    )
    discount: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=13,
        message=money_pb2.Money,
    )
    total: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=14,
        message=money_pb2.Money,
    )
    price_components: MutableSequence[types_pb2.PriceComponent] = proto.RepeatedField(
        proto.MESSAGE,
        number=25,
        message=types_pb2.PriceComponent,
    )
    notes: str = proto.Field(
        proto.STRING,
        number=15,
    )
    attributes: struct_pb2.Struct = proto.Field(
        proto.MESSAGE,
        number=16,
        message=struct_pb2.Struct,
    )
    cancel_reason: enums.CancelReason = proto.Field(
        proto.ENUM,
        number=17,
        enum=enums.CancelReason,
    )
    create_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=18,
        message=timestamp_pb2.Timestamp,
    )
    update_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=19,
        message=timestamp_pb2.Timestamp,
    )
    confirm_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=20,
        message=timestamp_pb2.Timestamp,
    )
    cancel_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=21,
        message=timestamp_pb2.Timestamp,
    )
    refund_amount: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=26,
        message=money_pb2.Money,
    )
    refund_percent: int = proto.Field(
        proto.INT32,
        number=27,
    )
    hold_ttl: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=22,
        message=duration_pb2.Duration,
    )
    etag: str = proto.Field(
        proto.STRING,
        number=23,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
