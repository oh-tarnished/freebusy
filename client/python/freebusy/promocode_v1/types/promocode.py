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

from freebusy.promocode_v1.types import enums
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.protobuf.wrappers_pb2 as wrappers_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.promocode.v1',
    manifest={
        'PromoCode',
        'Discount',
        'RedemptionWindow',
        'UsageLimits',
        'Scope',
        'Redemption',
    },
)


class PromoCode(proto.Message):
    r"""A redeemable discount applied to a booking's subtotal. Scoped
    by a redemption window, usage caps, a minimum subtotal, and an
    optional set of resources / offerings it applies to.

    Attributes:
        name (str):
            The promo code name. Format: promoCodes/{promo_code}
        code (str):
            The human-entered code, unique across all
            promo codes (e.g. "SUMMER25").
        display_name (str):
            Internal display name (not shown to
            customers).
        description (str):
            Free-form description.
        discount (freebusy.promocode_v1.types.Discount):
            How the discount is computed → belongs-to
            child table promocode.discounts.
        window (freebusy.promocode_v1.types.RedemptionWindow):
            When the code is redeemable → belongs-to
            promocode.redemption_windows.
        limits (freebusy.promocode_v1.types.UsageLimits):
            Redemption caps → belongs-to promocode.usage_limits.
        scope (freebusy.promocode_v1.types.Scope):
            What the code applies to (eligibility) →
            belongs-to promocode.scopes.
        redemption_count (int):
            How many times the code has been redeemed. Cheap summary on
            the resource; the individual redemptions are a paginated
            sub-collection (promoCodes/{promo_code}/redemptions) listed
            via ListRedemptions, not inlined here, so reads stay bounded
            as redemptions accumulate.
        state (freebusy.promocode_v1.types.PromoCodeState):
            Derived lifecycle state, recomputed on every read/validate
            (no background job): DISABLED when ``disabled`` is set;
            EXPIRED once now > window.end_time (the expiry) or the code
            is out of redemptions; otherwise ACTIVE. This is how a code
            auto-disables at its expiry without a stored flag to flip.
        disabled (bool):
            If true, the code is manually disabled
            regardless of its window and caps.
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
    code: str = proto.Field(
        proto.STRING,
        number=3,
    )
    display_name: str = proto.Field(
        proto.STRING,
        number=4,
    )
    description: str = proto.Field(
        proto.STRING,
        number=5,
    )
    discount: 'Discount' = proto.Field(
        proto.MESSAGE,
        number=23,
        message='Discount',
    )
    window: 'RedemptionWindow' = proto.Field(
        proto.MESSAGE,
        number=24,
        message='RedemptionWindow',
    )
    limits: 'UsageLimits' = proto.Field(
        proto.MESSAGE,
        number=25,
        message='UsageLimits',
    )
    scope: 'Scope' = proto.Field(
        proto.MESSAGE,
        number=26,
        message='Scope',
    )
    redemption_count: int = proto.Field(
        proto.INT64,
        number=16,
    )
    state: enums.PromoCodeState = proto.Field(
        proto.ENUM,
        number=17,
        enum=enums.PromoCodeState,
    )
    disabled: bool = proto.Field(
        proto.BOOL,
        number=20,
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
    etag: str = proto.Field(
        proto.STRING,
        number=21,
    )


class Discount(proto.Message):
    r"""Discount describes how a promo code reduces a subtotal. Nested value
    object → belongs-to child table promocode.discounts (FK discount_id
    on promo_codes). Exactly one of percent_off / amount_off is set; the
    oneof case is the discriminator, so no separate type enum is needed.

    This message has `oneof`_ fields (mutually exclusive fields).
    For each oneof, at most one member field can be set at the same time.
    Setting any member of the oneof automatically clears all other
    members.

    .. _oneof: https://proto-plus-python.readthedocs.io/en/stable/fields.html#oneofs-mutually-exclusive-fields

    Attributes:
        percent_off (int):
            Percentage off the subtotal (1-100).

            This field is a member of `oneof`_ ``amount``.
        amount_off (google.type.money_pb2.Money):
            Fixed amount off the subtotal. Normalized into the shared
            common.moneys table (belongs-to via amount_off_id).

            This field is a member of `oneof`_ ``amount``.
    """

    percent_off: int = proto.Field(
        proto.INT32,
        number=1,
        oneof='amount',
    )
    amount_off: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=2,
        oneof='amount',
        message=money_pb2.Money,
    )


class RedemptionWindow(proto.Message):
    r"""RedemptionWindow bounds when a code can be redeemed; an unset bound
    is open-ended. Nested value object → belongs-to
    promocode.redemption_windows.

    Attributes:
        start_time (google.protobuf.timestamp_pb2.Timestamp):
            Earliest the code can be redeemed. Unset
            means no lower bound.
        end_time (google.protobuf.timestamp_pb2.Timestamp):
            The code's expiry: the latest moment it can
            be redeemed. Once now passes this, the derived
            PromoCode.state becomes EXPIRED. Unset means
            never expires.
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


class UsageLimits(proto.Message):
    r"""UsageLimits caps how often a code can be redeemed. Nested value
    object → belongs-to promocode.usage_limits. The caps are wrapper
    types so "unset" (unlimited) is distinct from an explicit value,
    including 0.

    Attributes:
        max_redemptions (google.protobuf.wrappers_pb2.Int64Value):
            Maximum total redemptions across all
            customers. Unset means unlimited.
        per_customer_limit (google.protobuf.wrappers_pb2.Int32Value):
            Maximum redemptions per customer. Unset means
            unlimited.
    """

    max_redemptions: wrappers_pb2.Int64Value = proto.Field(
        proto.MESSAGE,
        number=1,
        message=wrappers_pb2.Int64Value,
    )
    per_customer_limit: wrappers_pb2.Int32Value = proto.Field(
        proto.MESSAGE,
        number=2,
        message=wrappers_pb2.Int32Value,
    )


class Scope(proto.Message):
    r"""Scope restricts which bookings a code applies to. Nested
    value object → belongs-to promocode.scopes. Its repeated
    resource references and Money normalize one level deeper (array
    columns / common.moneys).

    Attributes:
        min_subtotal (google.type.money_pb2.Money):
            Minimum subtotal required for the code to apply. Normalized
            into the shared common.moneys table (belongs-to via
            min_subtotal_id).
        applicable_resources (MutableSequence[str]):
            Resources the code applies to. Empty means
            all resources. Format: resources/{resource}
        applicable_offerings (MutableSequence[str]):
            Offerings the code applies to. Empty means
            all offerings. Format:
            resources/{resource}/offerings/{offering}
    """

    min_subtotal: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=1,
        message=money_pb2.Money,
    )
    applicable_resources: MutableSequence[str] = proto.RepeatedField(
        proto.STRING,
        number=2,
    )
    applicable_offerings: MutableSequence[str] = proto.RepeatedField(
        proto.STRING,
        number=3,
    )


class Redemption(proto.Message):
    r"""Redemption is a single use of a promo code, modeled as a
    sub-resource of PromoCode rather than an inline list — so it has its
    own name/lifecycle and is listed with paging (ListRedemptions). The
    {promo_code} parent segment generates the promo_code_id FK back to
    the owning code (1:n into promocode.redemptions); amount_applied is
    the shared google.type.Money in common.moneys. Redemptions are
    created during CreateBooking, never directly.

    Attributes:
        name (str):
            The redemption resource name. Format:
            promoCodes/{promo_code}/redemptions/{redemption}
        customer (str):
            The customer who redeemed the code.
            Format: users/{user}
        booking (str):
            The booking the code was applied to.
            Format: bookings/{booking}
        redeemed_time (google.protobuf.timestamp_pb2.Timestamp):
            When the code was redeemed.
        amount_applied (google.type.money_pb2.Money):
            The discount actually applied at redemption.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    customer: str = proto.Field(
        proto.STRING,
        number=2,
    )
    booking: str = proto.Field(
        proto.STRING,
        number=3,
    )
    redeemed_time: timestamp_pb2.Timestamp = proto.Field(
        proto.MESSAGE,
        number=4,
        message=timestamp_pb2.Timestamp,
    )
    amount_applied: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=5,
        message=money_pb2.Money,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
