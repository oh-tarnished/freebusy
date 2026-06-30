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

from freebusy.booking_v1.types import booking as fb_booking


__protobuf__ = proto.module(
    package='freebusy.booking.v1',
    manifest={
        'CreateBookingRequest',
        'GetBookingRequest',
        'ListBookingsRequest',
        'ListBookingsResponse',
    },
)


class CreateBookingRequest(proto.Message):
    r"""Request message for CreateBooking. This places a hold
    transactionally; the request_id makes retries safe (the same id
    always yields the same booking instead of a duplicate hold). The
    promo code and hold TTL are set on the Booking resource itself.

    Attributes:
        booking (freebusy.booking_v1.types.Booking):
            The booking to create. Supply resource, window, and
            optionally offering, units, customer, contact, notes,
            attributes, promo_code, and hold_ttl. Provide contact when
            there is no customer (a guest booking). Output-only fields
            are ignored.
        request_id (str):
            Caller-supplied idempotency key that dedupes
            retries of this create. Reusing an id returns
            the booking created by the first call.
        booking_id (str):
            Optional caller-chosen ID for the booking;
            the server generates one if unset.
        validate_only (bool):
            If true, validate the request (availability +
            policy) and report what would happen, but place
            no hold.
    """

    booking: fb_booking.Booking = proto.Field(
        proto.MESSAGE,
        number=1,
        message=fb_booking.Booking,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=2,
    )
    booking_id: str = proto.Field(
        proto.STRING,
        number=3,
    )
    validate_only: bool = proto.Field(
        proto.BOOL,
        number=4,
    )


class GetBookingRequest(proto.Message):
    r"""Request message for GetBooking.

    Attributes:
        name (str):
            The booking to retrieve.
            Format: bookings/{booking}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class ListBookingsRequest(proto.Message):
    r"""Request message for ListBookings.

    Attributes:
        page_size (int):
            Maximum number of bookings to return. The
            server may cap this.
        page_token (str):
            Page token from a previous ListBookings call's
            next_page_token.
        filter (str):
            Filter expression (AIP-160), e.g.
            ``resource = "resources/42"``, ``customer = "users/7"``,
            ``state = CONFIRMED``, or a window overlap predicate.
        order_by (str):
            Sort order, e.g. "create_time desc" or "window.start_time".
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


class ListBookingsResponse(proto.Message):
    r"""Response message for ListBookings.

    Attributes:
        bookings (MutableSequence[freebusy.booking_v1.types.Booking]):
            The page of bookings.
        next_page_token (str):
            Token to pass as page_token to retrieve the next page; empty
            when no more.
    """

    @property
    def raw_page(self):
        return self

    bookings: MutableSequence[fb_booking.Booking] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=fb_booking.Booking,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
