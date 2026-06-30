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
from freebusy.availability import gapic_version as package_version

__version__ = package_version.__version__


from freebusy.availability_v1.services.availability_service.client import AvailabilityServiceClient
from freebusy.availability_v1.services.availability_service.async_client import AvailabilityServiceAsyncClient

from freebusy.availability_v1.types.availability import AvailabilityMatch
from freebusy.availability_v1.types.availability import BatchComputeAvailabilityRequest
from freebusy.availability_v1.types.availability import BatchComputeAvailabilityResponse
from freebusy.availability_v1.types.availability import BookableRange
from freebusy.availability_v1.types.availability import CheckAvailabilityRequest
from freebusy.availability_v1.types.availability import CheckAvailabilityResponse
from freebusy.availability_v1.types.availability import ComputeAvailabilityRequest
from freebusy.availability_v1.types.availability import ComputeAvailabilityResponse
from freebusy.availability_v1.types.availability import ComputeBookableRangesRequest
from freebusy.availability_v1.types.availability import ComputeBookableRangesResponse
from freebusy.availability_v1.types.availability import NightAvailability
from freebusy.availability_v1.types.availability import ResourceAvailability
from freebusy.availability_v1.types.availability import SearchAvailabilityRequest
from freebusy.availability_v1.types.availability import SearchAvailabilityResponse
from freebusy.availability_v1.types.availability import Slot
from freebusy.availability_v1.types.availability import UnbookableReason
from freebusy.availability_v1.types.enums import Code

__all__ = ('AvailabilityServiceClient',
    'AvailabilityServiceAsyncClient',
    'AvailabilityMatch',
    'BatchComputeAvailabilityRequest',
    'BatchComputeAvailabilityResponse',
    'BookableRange',
    'CheckAvailabilityRequest',
    'CheckAvailabilityResponse',
    'ComputeAvailabilityRequest',
    'ComputeAvailabilityResponse',
    'ComputeBookableRangesRequest',
    'ComputeBookableRangesResponse',
    'NightAvailability',
    'ResourceAvailability',
    'SearchAvailabilityRequest',
    'SearchAvailabilityResponse',
    'Slot',
    'UnbookableReason',
    'Code',
)
