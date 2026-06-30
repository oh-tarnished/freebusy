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
from freebusy.booking import gapic_version as package_version

__version__ = package_version.__version__


from freebusy.booking_v1.services.booking_service.client import BookingServiceClient
from freebusy.booking_v1.services.booking_service.async_client import BookingServiceAsyncClient

from freebusy.booking_v1.types.booking import Booking
from freebusy.booking_v1.types.booking_actions import CancelBookingRequest
from freebusy.booking_v1.types.booking_actions import ConfirmBookingRequest
from freebusy.booking_v1.types.booking_actions import PreviewCancellationRequest
from freebusy.booking_v1.types.booking_actions import PreviewCancellationResponse
from freebusy.booking_v1.types.booking_actions import RescheduleBookingRequest
from freebusy.booking_v1.types.booking_mcp import BookSlotArgs
from freebusy.booking_v1.types.booking_messages import CreateBookingRequest
from freebusy.booking_v1.types.booking_messages import GetBookingRequest
from freebusy.booking_v1.types.booking_messages import ListBookingsRequest
from freebusy.booking_v1.types.booking_messages import ListBookingsResponse
from freebusy.booking_v1.types.enums import BookingState
from freebusy.booking_v1.types.enums import CancelReason

__all__ = ('BookingServiceClient',
    'BookingServiceAsyncClient',
    'Booking',
    'CancelBookingRequest',
    'ConfirmBookingRequest',
    'PreviewCancellationRequest',
    'PreviewCancellationResponse',
    'RescheduleBookingRequest',
    'BookSlotArgs',
    'CreateBookingRequest',
    'GetBookingRequest',
    'ListBookingsRequest',
    'ListBookingsResponse',
    'BookingState',
    'CancelReason',
)
