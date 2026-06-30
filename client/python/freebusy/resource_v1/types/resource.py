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

from freebusy.resource_v1.types import enums as fr_enums
import freebusy.shared.v1.enums_pb2 as enums_pb2  # type: ignore
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
import google.protobuf.duration_pb2 as duration_pb2  # type: ignore
import google.protobuf.struct_pb2 as struct_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.resource.v1',
    manifest={
        'Resource',
        'Offering',
        'RateOverride',
        'LosDiscount',
        'Fee',
        'Tax',
    },
)


class Resource(proto.Message):
    r"""A bookable thing: a provider, room, piece of equipment, or a unit
    type. A resource is a pool of ``capacity`` interchangeable units;
    the freebusy engine computes how many are free for a given window.
    Its booking_mode decides whether availability is produced as time
    slots or per-night counts.

    Attributes:
        name (str):
            The resource name.
            Format: resources/{resource}
        display_name (str):
            Human-friendly name (e.g. "Dr. Lee", "Deluxe
            King", "Kayak #3").
        description (str):
            Free-form description.
        type_ (freebusy.resource_v1.types.ResourceType):
            What kind of bookable thing this is.
        booking_mode (freebusy.shared.v1.enums_pb2.BookingMode):
            How this resource is booked, and therefore
            the availability shape it yields. Immutable:
            flipping it after bookings exist would
            invalidate every existing booking and
            availability computation.
        capacity (int):
            Number of interchangeable units in the pool.
            Defaults to 1 when unset.
        time_zone (str):
            IANA timezone (e.g. "America/New_York") the resource's hours
            and dates are evaluated in. Required so availability is
            timezone-correct.
        tags (MutableSequence[str]):
            Arbitrary tags for grouping and filtering.
        attributes (google.protobuf.struct_pb2.Struct):
            Arbitrary attributes used for templating,
            policy, and segmentation.
        offerings (MutableSequence[str]):
            Resource names of the offerings attached to
            this resource (e.g. "30-min consult"); manage
            them with the Offering standard methods. Format:
            resources/{resource}/offerings/{offering}
        state (freebusy.resource_v1.types.ResourceState):
            Lifecycle state.
        create_time (google.protobuf.timestamp_pb2.Timestamp):
            Creation timestamp.
        update_time (google.protobuf.timestamp_pb2.Timestamp):
            Last-modification timestamp.
        etag (str):
            Opaque version for optimistic concurrency
            (AIP-154); echo on update/delete.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=3,
    )
    description: str = proto.Field(
        proto.STRING,
        number=4,
    )
    type_: fr_enums.ResourceType = proto.Field(
        proto.ENUM,
        number=5,
        enum=fr_enums.ResourceType,
    )
    booking_mode: enums_pb2.BookingMode = proto.Field(
        proto.ENUM,
        number=6,
        enum=enums_pb2.BookingMode,
    )
    capacity: int = proto.Field(
        proto.INT32,
        number=7,
    )
    time_zone: str = proto.Field(
        proto.STRING,
        number=8,
    )
    tags: MutableSequence[str] = proto.RepeatedField(
        proto.STRING,
        number=9,
    )
    attributes: struct_pb2.Struct = proto.Field(
        proto.MESSAGE,
        number=10,
        message=struct_pb2.Struct,
    )
    offerings: MutableSequence[str] = proto.RepeatedField(
        proto.STRING,
        number=11,
    )
    state: fr_enums.ResourceState = proto.Field(
        proto.ENUM,
        number=12,
        enum=fr_enums.ResourceState,
    )
    create_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=13,
        message=timestamp_pb2.Timestamp,
    )
    update_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=14,
        message=timestamp_pb2.Timestamp,
    )
    etag: str = proto.Field(
        proto.STRING,
        number=15,
    )


class Offering(proto.Message):
    r"""A specific way a resource can be booked, carrying its
    duration and price. A "30-min consult" and a "60-min session"
    are two offerings on the same provider. For NIGHTLY resources
    the duration is unused and price is per-night.

    Attributes:
        name (str):
            The offering name.
            Format:
            resources/{resource}/offerings/{offering}
        display_name (str):
            Human-friendly name (e.g. "30-min consult").
        description (str):
            Free-form description.
        duration (google.protobuf.duration_pb2.Duration):
            Slot length. Required for TIME_SLOT resources; ignored for
            NIGHTLY.
        price (google.type.money_pb2.Money):
            Price charged for the offering, interpreted per
            pricing_unit.
        pricing_unit (freebusy.resource_v1.types.PricingUnit):
            What the price is charged per.
        rate_overrides (MutableSequence[freebusy.resource_v1.types.RateOverride]):
            Rate calendar: date- and weekday-scoped overrides of
            ``price``. For NIGHTLY resources this is the
            seasonal/weekend rate calendar; for TIME_SLOT it varies slot
            price by date or day. ``price`` is the default when no
            override matches. Later-listed overrides win where they
            overlap.
        los_discounts (MutableSequence[freebusy.resource_v1.types.LosDiscount]):
            Length-of-stay discounts applied to the NIGHTLY subtotal
            when a stay is at least ``min_nights`` long. The most
            generous matching discount applies.
        fees (MutableSequence[freebusy.resource_v1.types.Fee]):
            Fees added on top of the base subtotal (e.g. cleaning,
            service). Each surfaces as a TYPE_FEE line in a booking's
            price_components.
        taxes (MutableSequence[freebusy.resource_v1.types.Tax]):
            Taxes applied to the taxable base (subtotal plus taxable
            fees). Each surfaces as a TYPE_TAX line in a booking's
            price_components.
        state (freebusy.resource_v1.types.OfferingState):
            Lifecycle state.
        create_time (google.protobuf.timestamp_pb2.Timestamp):
            Creation timestamp.
        update_time (google.protobuf.timestamp_pb2.Timestamp):
            Last-modification timestamp.
        etag (str):
            Opaque version for optimistic concurrency
            (AIP-154); echo on update/delete.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=3,
    )
    description: str = proto.Field(
        proto.STRING,
        number=4,
    )
    duration: duration_pb2.Duration = proto.Field(
        proto.MESSAGE,
        number=5,
        message=duration_pb2.Duration,
    )
    price: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=6,
        message=money_pb2.Money,
    )
    pricing_unit: fr_enums.PricingUnit = proto.Field(
        proto.ENUM,
        number=7,
        enum=fr_enums.PricingUnit,
    )
    rate_overrides: MutableSequence['RateOverride'] = proto.RepeatedField(
        proto.MESSAGE,
        number=12,
        message='RateOverride',
    )
    los_discounts: MutableSequence['LosDiscount'] = proto.RepeatedField(
        proto.MESSAGE,
        number=13,
        message='LosDiscount',
    )
    fees: MutableSequence['Fee'] = proto.RepeatedField(
        proto.MESSAGE,
        number=14,
        message='Fee',
    )
    taxes: MutableSequence['Tax'] = proto.RepeatedField(
        proto.MESSAGE,
        number=15,
        message='Tax',
    )
    state: fr_enums.OfferingState = proto.Field(
        proto.ENUM,
        number=8,
        enum=fr_enums.OfferingState,
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


class RateOverride(proto.Message):
    r"""A price override for a span of dates and/or specific weekdays,
    layered over an offering's base ``price``. The price is still
    interpreted per the offering's pricing_unit (per night, per booking,
    per person).

    Attributes:
        date_range (freebusy.shared.v1.types_pb2.DateRange):
            Dates the override applies to, in the
            resource's timezone. Unset means it applies on
            every date (a pure weekday rule).
        weekdays (MutableSequence[freebusy.shared.v1.enums_pb2.Weekday]):
            Weekdays the override applies to. Empty means every day
            within date_range.
        price (google.type.money_pb2.Money):
            The price in effect while this override
            matches.
    """

    date_range: types_pb2.DateRange = proto.Field(
        proto.MESSAGE,
        number=1,
        message=types_pb2.DateRange,
    )
    weekdays: MutableSequence[enums_pb2.Weekday] = proto.RepeatedField(
        proto.ENUM,
        number=2,
        enum=enums_pb2.Weekday,
    )
    price: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=3,
        message=money_pb2.Money,
    )


class LosDiscount(proto.Message):
    r"""A discount applied to a NIGHTLY subtotal once the stay reaches a
    minimum length. Exactly one of percent_off or amount_off is set.

    Attributes:
        min_nights (int):
            Minimum nights for the discount to apply.
        percent_off (int):
            Percent off the subtotal (1-100), when
            discounting by percentage.
        amount_off (google.type.money_pb2.Money):
            Fixed amount off the subtotal, when
            discounting by a flat amount.
    """

    min_nights: int = proto.Field(
        proto.INT32,
        number=1,
    )
    percent_off: int = proto.Field(
        proto.INT32,
        number=2,
    )
    amount_off: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=3,
        message=money_pb2.Money,
    )


class Fee(proto.Message):
    r"""A fee added on top of an offering's base subtotal. Exactly one of
    ``amount`` or ``percent`` is set. Surfaces as a TYPE_FEE line in a
    booking's price_components.

    Attributes:
        code (str):
            Stable machine code, e.g. "cleaning_fee".
        display_name (str):
            Human-readable label for receipts.
        amount (google.type.money_pb2.Money):
            Fixed fee amount, when charging a flat fee.
        percent (int):
            Percent of the base subtotal (1-100), when
            charging a proportional fee.
        pricing_unit (freebusy.resource_v1.types.PricingUnit):
            What the fee is charged per (per booking, per
            night, per person). Defaults to per booking.
        taxable (bool):
            Whether this fee is included in the taxable
            base.
    """

    code: str = proto.Field(
        proto.STRING,
        number=1,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=2,
    )
    amount: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=3,
        message=money_pb2.Money,
    )
    percent: int = proto.Field(
        proto.INT32,
        number=4,
    )
    pricing_unit: fr_enums.PricingUnit = proto.Field(
        proto.ENUM,
        number=5,
        enum=fr_enums.PricingUnit,
    )
    taxable: bool = proto.Field(
        proto.BOOL,
        number=6,
    )


class Tax(proto.Message):
    r"""A tax applied to the taxable base (base subtotal plus taxable fees).
    Surfaces as a TYPE_TAX line in a booking's price_components.

    Attributes:
        code (str):
            Stable machine code, e.g. "occupancy_tax" or "vat".
        display_name (str):
            Human-readable label for receipts.
        percent (float):
            Tax rate as a percentage, e.g. 8.5 for 8.5%.
    """

    code: str = proto.Field(
        proto.STRING,
        number=1,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=2,
    )
    percent: float = proto.Field(
        proto.DOUBLE,
        number=3,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
