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
    package='freebusy.resource.v1',
    manifest={
        'ResourceState',
        'OfferingState',
        'ResourceType',
        'PricingUnit',
    },
)


class ResourceState(proto.Enum):
    r"""Lifecycle status of a resource.

    Values:
        RESOURCE_STATE_UNSPECIFIED (0):
            Unset.
        RESOURCE_STATE_ACTIVE (1):
            Bookable.
        RESOURCE_STATE_ARCHIVED (2):
            Retired; hidden from availability and new
            bookings.
    """
    RESOURCE_STATE_UNSPECIFIED = 0
    RESOURCE_STATE_ACTIVE = 1
    RESOURCE_STATE_ARCHIVED = 2


class OfferingState(proto.Enum):
    r"""Lifecycle state of an offering.

    Values:
        OFFERING_STATE_UNSPECIFIED (0):
            Unset; treated as active.
        OFFERING_STATE_ACTIVE (1):
            Bookable.
        OFFERING_STATE_INACTIVE (2):
            Hidden from new bookings.
    """
    OFFERING_STATE_UNSPECIFIED = 0
    OFFERING_STATE_ACTIVE = 1
    OFFERING_STATE_INACTIVE = 2


class ResourceType(proto.Enum):
    r"""Kind of bookable resource.

    Values:
        RESOURCE_TYPE_UNSPECIFIED (0):
            Unset.
        RESOURCE_TYPE_PROVIDER (1):
            A person who delivers a service (e.g. a
            doctor, stylist).
        RESOURCE_TYPE_ROOM (2):
            A bookable room or space (e.g. a meeting
            room).
        RESOURCE_TYPE_EQUIPMENT (3):
            A bookable piece of equipment (e.g. a kayak).
        RESOURCE_TYPE_UNIT_TYPE (4):
            A lodging unit type backed by a pool of
            identical units.
        RESOURCE_TYPE_SPACE (5):
            A generic space or venue.
    """
    RESOURCE_TYPE_UNSPECIFIED = 0
    RESOURCE_TYPE_PROVIDER = 1
    RESOURCE_TYPE_ROOM = 2
    RESOURCE_TYPE_EQUIPMENT = 3
    RESOURCE_TYPE_UNIT_TYPE = 4
    RESOURCE_TYPE_SPACE = 5


class PricingUnit(proto.Enum):
    r"""What an offering's price is charged per.

    Values:
        PRICING_UNIT_UNSPECIFIED (0):
            Unset; treated as per-booking.
        PRICING_UNIT_PER_BOOKING (1):
            A flat price for the whole booking.
        PRICING_UNIT_PER_NIGHT (2):
            Price multiplied by the number of nights
            (NIGHTLY resources).
        PRICING_UNIT_PER_PERSON (3):
            Price multiplied by party size / units
            booked.
    """
    PRICING_UNIT_UNSPECIFIED = 0
    PRICING_UNIT_PER_BOOKING = 1
    PRICING_UNIT_PER_NIGHT = 2
    PRICING_UNIT_PER_PERSON = 3


__all__ = tuple(sorted(__protobuf__.manifest))
