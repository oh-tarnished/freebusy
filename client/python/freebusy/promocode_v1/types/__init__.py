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
from .enums import (
    CodeGeneration,
    PromoCodeInvalidReason,
    PromoCodeState,
)
from .promocode import (
    Discount,
    PromoCode,
    Redemption,
    RedemptionWindow,
    Scope,
    UsageLimits,
)
from .promocode_messages import (
    CreatePromoCodeRequest,
    DeletePromoCodeRequest,
    GetPromoCodeRequest,
    GetRedemptionRequest,
    ListPromoCodesRequest,
    ListPromoCodesResponse,
    ListRedemptionsRequest,
    ListRedemptionsResponse,
    UpdatePromoCodeRequest,
    ValidatePromoCodeRequest,
    ValidatePromoCodeResponse,
)

__all__ = (
    'CodeGeneration',
    'PromoCodeInvalidReason',
    'PromoCodeState',
    'Discount',
    'PromoCode',
    'Redemption',
    'RedemptionWindow',
    'Scope',
    'UsageLimits',
    'CreatePromoCodeRequest',
    'DeletePromoCodeRequest',
    'GetPromoCodeRequest',
    'GetRedemptionRequest',
    'ListPromoCodesRequest',
    'ListPromoCodesResponse',
    'ListRedemptionsRequest',
    'ListRedemptionsResponse',
    'UpdatePromoCodeRequest',
    'ValidatePromoCodeRequest',
    'ValidatePromoCodeResponse',
)
