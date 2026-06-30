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
    package='freebusy.promocode.v1',
    manifest={
        'PromoCodeState',
        'PromoCodeInvalidReason',
        'CodeGeneration',
    },
)


class PromoCodeState(proto.Enum):
    r"""Lifecycle state of a promo code.

    Values:
        PROMO_CODE_STATE_UNSPECIFIED (0):
            Unset; treated as active.
        PROMO_CODE_STATE_ACTIVE (1):
            Redeemable (subject to window and caps).
        PROMO_CODE_STATE_DISABLED (2):
            Manually disabled.
        PROMO_CODE_STATE_EXPIRED (3):
            Past its redemption window or out of
            redemptions.
    """
    PROMO_CODE_STATE_UNSPECIFIED = 0
    PROMO_CODE_STATE_ACTIVE = 1
    PROMO_CODE_STATE_DISABLED = 2
    PROMO_CODE_STATE_EXPIRED = 3


class PromoCodeInvalidReason(proto.Enum):
    r"""Why a promo code failed validation, returned on
    ValidatePromoCodeResponse. Each value maps to one of the code's
    nested rules (window / limits / scope) or its lifecycle, so a
    client can branch without parsing the human-readable reason.

    Values:
        PROMO_CODE_INVALID_REASON_UNSPECIFIED (0):
            The code is valid, or no specific reason
            applies.
        PROMO_CODE_INVALID_REASON_NOT_FOUND (1):
            No promo code matches the supplied code
            string.
        PROMO_CODE_INVALID_REASON_DISABLED (2):
            The code is manually disabled (``disabled`` is set).
        PROMO_CODE_INVALID_REASON_NOT_STARTED (3):
            Now is before window.start_time; the code isn't redeemable
            yet.
        PROMO_CODE_INVALID_REASON_EXPIRED (4):
            Now is after window.end_time; the code has expired.
        PROMO_CODE_INVALID_REASON_OUT_OF_REDEMPTIONS (5):
            The code hit its total redemption cap
            (limits.max_redemptions).
        PROMO_CODE_INVALID_REASON_PER_CUSTOMER_LIMIT_REACHED (6):
            The customer hit their per-customer cap
            (limits.per_customer_limit).
        PROMO_CODE_INVALID_REASON_BELOW_MIN_SUBTOTAL (7):
            The subtotal is below scope.min_subtotal.
        PROMO_CODE_INVALID_REASON_OUT_OF_SCOPE (8):
            The booked resource/offering is outside
            scope.applicable_resources/offerings.
    """
    PROMO_CODE_INVALID_REASON_UNSPECIFIED = 0
    PROMO_CODE_INVALID_REASON_NOT_FOUND = 1
    PROMO_CODE_INVALID_REASON_DISABLED = 2
    PROMO_CODE_INVALID_REASON_NOT_STARTED = 3
    PROMO_CODE_INVALID_REASON_EXPIRED = 4
    PROMO_CODE_INVALID_REASON_OUT_OF_REDEMPTIONS = 5
    PROMO_CODE_INVALID_REASON_PER_CUSTOMER_LIMIT_REACHED = 6
    PROMO_CODE_INVALID_REASON_BELOW_MIN_SUBTOTAL = 7
    PROMO_CODE_INVALID_REASON_OUT_OF_SCOPE = 8


class CodeGeneration(proto.Enum):
    r"""How the human-facing code string is chosen when creating a
    promo code.

    Values:
        CODE_GENERATION_UNSPECIFIED (0):
            Unset; treated as MANUAL (use the
            caller-provided code).
        CODE_GENERATION_MANUAL (1):
            The caller provides the code in promo_code.code; the server
            uses it verbatim.
        CODE_GENERATION_AUTO (2):
            The server generates a unique code; any code in
            promo_code.code is ignored.
    """
    CODE_GENERATION_UNSPECIFIED = 0
    CODE_GENERATION_MANUAL = 1
    CODE_GENERATION_AUTO = 2


__all__ = tuple(sorted(__protobuf__.manifest))
