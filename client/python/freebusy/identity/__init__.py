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
from freebusy.identity import gapic_version as package_version

__version__ = package_version.__version__


from freebusy.identity_v1.services.identity_service.client import IdentityServiceClient
from freebusy.identity_v1.services.identity_service.async_client import IdentityServiceAsyncClient

from freebusy.identity_v1.types.identity import GetUserRequest
from freebusy.identity_v1.types.identity import ListUsersRequest
from freebusy.identity_v1.types.identity import ListUsersResponse
from freebusy.identity_v1.types.identity import MembershipSummary
from freebusy.identity_v1.types.identity import UpdateUserRequest
from freebusy.identity_v1.types.identity import User

__all__ = ('IdentityServiceClient',
    'IdentityServiceAsyncClient',
    'GetUserRequest',
    'ListUsersRequest',
    'ListUsersResponse',
    'MembershipSummary',
    'UpdateUserRequest',
    'User',
)
