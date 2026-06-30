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
from freebusy.promocode import gapic_version as package_version

__version__ = package_version.__version__


from freebusy.promocode_v1.services.promo_code_service.client import PromoCodeServiceClient
from freebusy.promocode_v1.services.promo_code_service.async_client import PromoCodeServiceAsyncClient

from freebusy.promocode_v1.types.enums import CodeGeneration
from freebusy.promocode_v1.types.enums import PromoCodeInvalidReason
from freebusy.promocode_v1.types.enums import PromoCodeState
from freebusy.promocode_v1.types.promocode import Discount
from freebusy.promocode_v1.types.promocode import PromoCode
from freebusy.promocode_v1.types.promocode import Redemption
from freebusy.promocode_v1.types.promocode import RedemptionWindow
from freebusy.promocode_v1.types.promocode import Scope
from freebusy.promocode_v1.types.promocode import UsageLimits
from freebusy.promocode_v1.types.promocode_messages import CreatePromoCodeRequest
from freebusy.promocode_v1.types.promocode_messages import DeletePromoCodeRequest
from freebusy.promocode_v1.types.promocode_messages import GetPromoCodeRequest
from freebusy.promocode_v1.types.promocode_messages import GetRedemptionRequest
from freebusy.promocode_v1.types.promocode_messages import ListPromoCodesRequest
from freebusy.promocode_v1.types.promocode_messages import ListPromoCodesResponse
from freebusy.promocode_v1.types.promocode_messages import ListRedemptionsRequest
from freebusy.promocode_v1.types.promocode_messages import ListRedemptionsResponse
from freebusy.promocode_v1.types.promocode_messages import UpdatePromoCodeRequest
from freebusy.promocode_v1.types.promocode_messages import ValidatePromoCodeRequest
from freebusy.promocode_v1.types.promocode_messages import ValidatePromoCodeResponse

__all__ = ('PromoCodeServiceClient',
    'PromoCodeServiceAsyncClient',
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
