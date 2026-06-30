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


__protobuf__ = proto.module(
    package='freebusy.booking.v1',
    manifest={
        'BookingState',
        'CancelReason',
    },
)


class BookingState(proto.Enum):
    r"""Lifecycle state of a booking.

    Values:
        BOOKING_STATE_UNSPECIFIED (0):
            Unset.
        BOOKING_STATE_PENDING_HOLD (1):
            Held but not yet confirmed; expires at hold_expire_time.
        BOOKING_STATE_CONFIRMED (2):
            Confirmed and active.
        BOOKING_STATE_CANCELLED (3):
            Cancelled by customer or operator.
        BOOKING_STATE_EXPIRED (4):
            Hold lapsed before confirmation (released by
            the sweeper).
        BOOKING_STATE_COMPLETED (5):
            The booked time has passed and the booking
            was honored.
        BOOKING_STATE_NO_SHOW (6):
            The customer did not show.
    """
    BOOKING_STATE_UNSPECIFIED = 0
    BOOKING_STATE_PENDING_HOLD = 1
    BOOKING_STATE_CONFIRMED = 2
    BOOKING_STATE_CANCELLED = 3
    BOOKING_STATE_EXPIRED = 4
    BOOKING_STATE_COMPLETED = 5
    BOOKING_STATE_NO_SHOW = 6


class CancelReason(proto.Enum):
    r"""Why a booking was cancelled.

    Values:
        CANCEL_REASON_UNSPECIFIED (0):
            Unset.
        CANCEL_REASON_REQUESTED_BY_CUSTOMER (1):
            The customer requested cancellation.
        CANCEL_REASON_REQUESTED_BY_OPERATOR (2):
            An operator cancelled it.
        CANCEL_REASON_PAYMENT_FAILED (3):
            Payment failed or was not completed in time.
        CANCEL_REASON_NO_SHOW (4):
            The customer did not show.
        CANCEL_REASON_OTHER (5):
            Any other reason.
    """
    CANCEL_REASON_UNSPECIFIED = 0
    CANCEL_REASON_REQUESTED_BY_CUSTOMER = 1
    CANCEL_REASON_REQUESTED_BY_OPERATOR = 2
    CANCEL_REASON_PAYMENT_FAILED = 3
    CANCEL_REASON_NO_SHOW = 4
    CANCEL_REASON_OTHER = 5


__all__ = tuple(sorted(__protobuf__.manifest))
