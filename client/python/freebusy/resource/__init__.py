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
from freebusy.resource import gapic_version as package_version

__version__ = package_version.__version__


from freebusy.resource_v1.services.resource_service.client import ResourceServiceClient
from freebusy.resource_v1.services.resource_service.async_client import ResourceServiceAsyncClient

from freebusy.resource_v1.types.enums import OfferingState
from freebusy.resource_v1.types.enums import PricingUnit
from freebusy.resource_v1.types.enums import ResourceState
from freebusy.resource_v1.types.enums import ResourceType
from freebusy.resource_v1.types.resource import Fee
from freebusy.resource_v1.types.resource import LosDiscount
from freebusy.resource_v1.types.resource import Offering
from freebusy.resource_v1.types.resource import RateOverride
from freebusy.resource_v1.types.resource import Resource
from freebusy.resource_v1.types.resource import Tax
from freebusy.resource_v1.types.resource_mcp import AddResourceArgs
from freebusy.resource_v1.types.resource_messages import ArchiveResourceRequest
from freebusy.resource_v1.types.resource_messages import CreateOfferingRequest
from freebusy.resource_v1.types.resource_messages import CreateResourceRequest
from freebusy.resource_v1.types.resource_messages import DeleteOfferingRequest
from freebusy.resource_v1.types.resource_messages import GetOfferingRequest
from freebusy.resource_v1.types.resource_messages import GetResourceRequest
from freebusy.resource_v1.types.resource_messages import ListOfferingsRequest
from freebusy.resource_v1.types.resource_messages import ListOfferingsResponse
from freebusy.resource_v1.types.resource_messages import ListResourcesRequest
from freebusy.resource_v1.types.resource_messages import ListResourcesResponse
from freebusy.resource_v1.types.resource_messages import UnarchiveResourceRequest
from freebusy.resource_v1.types.resource_messages import UpdateOfferingRequest
from freebusy.resource_v1.types.resource_messages import UpdateResourceRequest

__all__ = ('ResourceServiceClient',
    'ResourceServiceAsyncClient',
    'OfferingState',
    'PricingUnit',
    'ResourceState',
    'ResourceType',
    'Fee',
    'LosDiscount',
    'Offering',
    'RateOverride',
    'Resource',
    'Tax',
    'AddResourceArgs',
    'ArchiveResourceRequest',
    'CreateOfferingRequest',
    'CreateResourceRequest',
    'DeleteOfferingRequest',
    'GetOfferingRequest',
    'GetResourceRequest',
    'ListOfferingsRequest',
    'ListOfferingsResponse',
    'ListResourcesRequest',
    'ListResourcesResponse',
    'UnarchiveResourceRequest',
    'UpdateOfferingRequest',
    'UpdateResourceRequest',
)
