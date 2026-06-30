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
import google.type.money_pb2 as money_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.booking.v1',
    manifest={
        'ConfirmBookingRequest',
        'CancelBookingRequest',
        'RescheduleBookingRequest',
        'PreviewCancellationRequest',
        'PreviewCancellationResponse',
    },
)


class ConfirmBookingRequest(proto.Message):
    r"""Request message for ConfirmBooking. Confirmation normally arrives
    via the payment webhook rather than a direct client call; either way
    it flips a PENDING_HOLD to CONFIRMED.

    Attributes:
        name (str):
            The booking to confirm.
            Format: bookings/{booking}
        payment_ref (str):
            Opaque reference to the payment/settlement
            that confirmed this booking.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    payment_ref: str = proto.Field(
        proto.STRING,
        number=2,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=3,
    )


class CancelBookingRequest(proto.Message):
    r"""Request message for CancelBooking.

    Attributes:
        name (str):
            The booking to cancel.
            Format: bookings/{booking}
        reason (freebusy.booking_v1.types.CancelReason):
            Why the booking is being cancelled.
        note (str):
            Free-form note explaining the cancellation.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    reason: enums.CancelReason = proto.Field(
        proto.ENUM,
        number=2,
        enum=enums.CancelReason,
    )
    note: str = proto.Field(
        proto.STRING,
        number=3,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=4,
    )


class RescheduleBookingRequest(proto.Message):
    r"""Request message for RescheduleBooking. Atomically moves a
    booking to a new span (and optionally offering), re-running
    availability and the exclusion check on the new window.

    Attributes:
        name (str):
            The booking to reschedule.
            Format: bookings/{booking}
        window (freebusy.shared.v1.types_pb2.TimeWindow):
            The new reserved span.
        offering (str):
            The new offering, when changing it as part of
            the reschedule. Format:
            resources/{resource}/offerings/{offering}
        request_id (str):
            Caller-supplied idempotency key that dedupes
            retries of this reschedule.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    window: types_pb2.TimeWindow = proto.Field(
        proto.MESSAGE,
        number=2,
        message=types_pb2.TimeWindow,
    )
    offering: str = proto.Field(
        proto.STRING,
        number=3,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=4,
    )


class PreviewCancellationRequest(proto.Message):
    r"""Request message for PreviewCancellation. Computes the refund
    a cancellation would yield right now, under the resource's
    cancellation policy, without cancelling the booking.

    Attributes:
        name (str):
            The booking to preview a cancellation for.
            Format: bookings/{booking}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class PreviewCancellationResponse(proto.Message):
    r"""Response message for PreviewCancellation.

    Attributes:
        refundable (bool):
            Whether cancelling now would refund anything.
        refund_percent (int):
            Percentage of the total that would be
            refunded (0-100).
        refund_amount (google.type.money_pb2.Money):
            Amount that would be refunded now.
        non_refundable_amount (google.type.money_pb2.Money):
            Amount that would be retained (total minus refund_amount).
        policy_summary (str):
            Human-readable summary of the policy tier
            that applied, for display.
    """

    refundable: bool = proto.Field(
        proto.BOOL,
        number=1,
    )
    refund_percent: int = proto.Field(
        proto.INT32,
        number=2,
    )
    refund_amount: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=3,
        message=money_pb2.Money,
    )
    non_refundable_amount: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=4,
        message=money_pb2.Money,
    )
    policy_summary: str = proto.Field(
        proto.STRING,
        number=5,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
