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
from freebusy.schedule import gapic_version as package_version

__version__ = package_version.__version__


from freebusy.schedule_v1.services.schedule_service.client import ScheduleServiceClient
from freebusy.schedule_v1.services.schedule_service.async_client import ScheduleServiceAsyncClient

from freebusy.schedule_v1.types.enums import ExceptionKind
from freebusy.schedule_v1.types.schedule import AvailabilityException
from freebusy.schedule_v1.types.schedule import BufferSettings
from freebusy.schedule_v1.types.schedule import CancellationPolicy
from freebusy.schedule_v1.types.schedule import RecurringRule
from freebusy.schedule_v1.types.schedule import RefundTier
from freebusy.schedule_v1.types.schedule import Schedule
from freebusy.schedule_v1.types.schedule import StayConstraints
from freebusy.schedule_v1.types.schedule_messages import CreateAvailabilityExceptionRequest
from freebusy.schedule_v1.types.schedule_messages import DeleteAvailabilityExceptionRequest
from freebusy.schedule_v1.types.schedule_messages import GetAvailabilityExceptionRequest
from freebusy.schedule_v1.types.schedule_messages import GetScheduleRequest
from freebusy.schedule_v1.types.schedule_messages import ListAvailabilityExceptionsRequest
from freebusy.schedule_v1.types.schedule_messages import ListAvailabilityExceptionsResponse
from freebusy.schedule_v1.types.schedule_messages import UpdateScheduleRequest

__all__ = ('ScheduleServiceClient',
    'ScheduleServiceAsyncClient',
    'ExceptionKind',
    'AvailabilityException',
    'BufferSettings',
    'CancellationPolicy',
    'RecurringRule',
    'RefundTier',
    'Schedule',
    'StayConstraints',
    'CreateAvailabilityExceptionRequest',
    'DeleteAvailabilityExceptionRequest',
    'GetAvailabilityExceptionRequest',
    'GetScheduleRequest',
    'ListAvailabilityExceptionsRequest',
    'ListAvailabilityExceptionsResponse',
    'UpdateScheduleRequest',
)
