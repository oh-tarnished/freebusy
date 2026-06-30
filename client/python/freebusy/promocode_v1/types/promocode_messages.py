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
from freebusy.promocode_v1.types import promocode
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore


__protobuf__ = proto.module(
    package='freebusy.promocode.v1',
    manifest={
        'ListPromoCodesRequest',
        'ListPromoCodesResponse',
        'GetPromoCodeRequest',
        'CreatePromoCodeRequest',
        'UpdatePromoCodeRequest',
        'DeletePromoCodeRequest',
        'ValidatePromoCodeRequest',
        'ValidatePromoCodeResponse',
        'GetRedemptionRequest',
        'ListRedemptionsRequest',
        'ListRedemptionsResponse',
    },
)


class ListPromoCodesRequest(proto.Message):
    r"""Request message for ListPromoCodes.

    Attributes:
        page_size (int):
            Maximum number of promo codes to return. The
            server may cap this.
        page_token (str):
            Page token from a previous ListPromoCodes call's
            next_page_token.
        filter (str):
            Filter expression over promo code fields, e.g.
            ``state = ACTIVE`` or a substring match on code/display_name
            (AIP-160).
        order_by (str):
            Sort order, e.g. "code" or "create_time desc".
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


class ListPromoCodesResponse(proto.Message):
    r"""Response message for ListPromoCodes.

    Attributes:
        promo_codes (MutableSequence[freebusy.promocode_v1.types.PromoCode]):
            The page of promo codes.
        next_page_token (str):
            Token to pass as page_token to retrieve the next page; empty
            when no more.
        total_size (int):
            Total number of promo codes matching the
            filter across all pages, when the server
            computes it; 0 if unknown. Useful for rendering
            page counts (AIP-132).
    """

    @property
    def raw_page(self):
        return self

    promo_codes: MutableSequence[promocode.PromoCode] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=promocode.PromoCode,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )
    total_size: int = proto.Field(
        proto.INT32,
        number=3,
    )


class GetPromoCodeRequest(proto.Message):
    r"""Request message for GetPromoCode.

    Attributes:
        name (str):
            The promo code to retrieve. Format: promoCodes/{promo_code}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class CreatePromoCodeRequest(proto.Message):
    r"""Request message for CreatePromoCode.

    Attributes:
        promo_code (freebusy.promocode_v1.types.PromoCode):
            The promo code to create. Server-assigned fields are ignored
            on input: name, state, redemption_count, redemptions,
            create_time, update_time, etag. promo_code.code is honored
            only when code_generation is MANUAL (or unset).
        code_generation (freebusy.promocode_v1.types.CodeGeneration):
            Whether the human-facing code is supplied by the caller
            (MANUAL) or minted by the server (AUTO). Unset defaults to
            MANUAL. When AUTO, promo_code.code is ignored and the server
            returns the generated code on the created resource. (--
            api-linter: core::0133::request-unknown-fields=disabled
            aip.dev/not-precedent: Code generation mode is intrinsic to
            creating a promo code and has no standard AIP-133
            equivalent. --)
        promo_code_id (str):
            Optional caller-chosen ID for the promo code;
            the server generates one if unset.
        request_id (str):
            Caller-supplied idempotency key; identical
            retries return the first result.
        validate_only (bool):
            If true, validate the request and return what
            would happen, but don't commit.
    """

    promo_code: promocode.PromoCode = proto.Field(
        proto.MESSAGE,
        number=1,
        message=promocode.PromoCode,
    )
    code_generation: enums.CodeGeneration = proto.Field(
        proto.ENUM,
        number=5,
        enum=enums.CodeGeneration,
    )
    promo_code_id: str = proto.Field(
        proto.STRING,
        number=2,
    )
    request_id: str = proto.Field(
        proto.STRING,
        number=3,
    )
    validate_only: bool = proto.Field(
        proto.BOOL,
        number=4,
    )


class UpdatePromoCodeRequest(proto.Message):
    r"""Request message for UpdatePromoCode.

    Attributes:
        promo_code (freebusy.promocode_v1.types.PromoCode):
            The promo code to update; its name identifies
            the target. For optimistic concurrency, echo the
            etag you last read (AIP-154); the update fails
            if it no longer matches.
        update_mask (google.protobuf.field_mask_pb2.FieldMask):
            Fields to overwrite. Omit to replace all mutable fields.
            Nested fields use dotted paths, e.g. "discount.amount_off",
            "window.end_time", "scope.min_subtotal".
        validate_only (bool):
            If true, validate the request and return what
            would happen, but don't commit.
    """

    promo_code: promocode.PromoCode = proto.Field(
        proto.MESSAGE,
        number=1,
        message=promocode.PromoCode,
    )
    update_mask: field_mask_pb2.FieldMask = proto.Field(
        proto.MESSAGE,
        number=2,
        message=field_mask_pb2.FieldMask,
    )
    validate_only: bool = proto.Field(
        proto.BOOL,
        number=3,
    )


class DeletePromoCodeRequest(proto.Message):
    r"""Request message for DeletePromoCode.

    Attributes:
        name (str):
            The promo code to delete. Format: promoCodes/{promo_code}
        etag (str):
            Optional optimistic-concurrency guard: if
            set, the delete fails unless it matches the
            current PromoCode.etag (AIP-154). Echo the etag
            you last read.
        force (bool):
            If true, delete the promo code along with its
            redemptions; otherwise the delete fails when
            redemptions exist.
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )
    etag: str = proto.Field(
        proto.STRING,
        number=2,
    )
    force: bool = proto.Field(
        proto.BOOL,
        number=3,
    )


class ValidatePromoCodeRequest(proto.Message):
    r"""Request message for ValidatePromoCode. Computes the discount
    a code would apply to a prospective booking without redeeming
    it.

    Attributes:
        code (str):
            The human-entered code to validate (e.g.
            "SUMMER25").
        subtotal (google.type.money_pb2.Money):
            Subtotal the discount would apply to.
        resource (str):
            Resource being booked, for scope checks.
            Format: resources/{resource}
        offering (str):
            Offering being booked, for scope checks.
            Format:
            resources/{resource}/offerings/{offering}
        customer (str):
            Customer redeeming the code, for per-customer
            limit checks. Format: users/{user}
    """

    code: str = proto.Field(
        proto.STRING,
        number=1,
    )
    subtotal: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=2,
        message=money_pb2.Money,
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


class ValidatePromoCodeResponse(proto.Message):
    r"""Response message for ValidatePromoCode.

    Attributes:
        valid (bool):
            Whether the code is valid and applicable to
            the given context.
        invalid_reason (freebusy.promocode_v1.types.PromoCodeInvalidReason):
            Structured reason the code is not valid, when valid is
            false; branch on this rather than parsing ``reason``.
            UNSPECIFIED when valid is true.
        reason (str):
            Human-readable detail for invalid_reason, suitable for
            display (e.g. "Minimum subtotal of $50 not met"). Empty when
            valid is true.
        promo_code (str):
            The resolved promo code, when valid. Format:
            promoCodes/{promo_code}
        discount_amount (google.type.money_pb2.Money):
            Discount the code applies to the subtotal.
        final_total (google.type.money_pb2.Money):
            Subtotal minus the discount.
    """

    valid: bool = proto.Field(
        proto.BOOL,
        number=1,
    )
    invalid_reason: enums.PromoCodeInvalidReason = proto.Field(
        proto.ENUM,
        number=6,
        enum=enums.PromoCodeInvalidReason,
    )
    reason: str = proto.Field(
        proto.STRING,
        number=2,
    )
    promo_code: str = proto.Field(
        proto.STRING,
        number=3,
    )
    discount_amount: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=4,
        message=money_pb2.Money,
    )
    final_total: money_pb2.Money = proto.Field(
        proto.MESSAGE,
        number=5,
        message=money_pb2.Money,
    )


class GetRedemptionRequest(proto.Message):
    r"""Request message for GetRedemption.

    Attributes:
        name (str):
            The redemption to retrieve. Format:
            promoCodes/{promo_code}/redemptions/{redemption}
    """

    name: str = proto.Field(
        proto.STRING,
        number=1,
    )


class ListRedemptionsRequest(proto.Message):
    r"""Request message for ListRedemptions.

    Attributes:
        parent (str):
            The promo code whose redemptions to list. Format:
            promoCodes/{promo_code}
        page_size (int):
            Maximum number of redemptions to return. The
            server may cap this.
        page_token (str):
            Page token from a previous ListRedemptions call's
            next_page_token.
        filter (str):
            Filter expression over redemption fields, e.g.
            ``customer = "users/123"`` (AIP-160).
        order_by (str):
            Sort order, e.g. "redeemed_time desc".
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


class ListRedemptionsResponse(proto.Message):
    r"""Response message for ListRedemptions.

    Attributes:
        redemptions (MutableSequence[freebusy.promocode_v1.types.Redemption]):
            The page of redemptions, newest first by
            default.
        next_page_token (str):
            Token to pass as page_token to retrieve the next page; empty
            when no more.
        total_size (int):
            Total number of redemptions for the promo
            code, when the server computes it; 0 if unknown.
    """

    @property
    def raw_page(self):
        return self

    redemptions: MutableSequence[promocode.Redemption] = proto.RepeatedField(
        proto.MESSAGE,
        number=1,
        message=promocode.Redemption,
    )
    next_page_token: str = proto.Field(
        proto.STRING,
        number=2,
    )
    total_size: int = proto.Field(
        proto.INT32,
        number=3,
    )


__all__ = tuple(sorted(__protobuf__.manifest))
