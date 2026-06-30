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
from freebusy.organisation import gapic_version as package_version

__version__ = package_version.__version__


from freebusy.organisation_v1.services.organisation_service.client import OrganisationServiceClient
from freebusy.organisation_v1.services.organisation_service.async_client import OrganisationServiceAsyncClient

from freebusy.organisation_v1.types.actions import DeleteMemberRequest
from freebusy.organisation_v1.types.actions import InviteMemberRequest
from freebusy.organisation_v1.types.actions import InviteMemberResponse
from freebusy.organisation_v1.types.actions import UpdateMemberRequest
from freebusy.organisation_v1.types.enums import MemberState
from freebusy.organisation_v1.types.enums import OrganisationRole
from freebusy.organisation_v1.types.enums import OrganisationState
from freebusy.organisation_v1.types.mcp import InviteMemberArgs
from freebusy.organisation_v1.types.organisation import Member
from freebusy.organisation_v1.types.organisation import Organisation
from freebusy.organisation_v1.types.organisation_message import CreateOrganisationRequest
from freebusy.organisation_v1.types.organisation_message import DeleteOrganisationRequest
from freebusy.organisation_v1.types.organisation_message import GetMemberRequest
from freebusy.organisation_v1.types.organisation_message import GetOrganisationRequest
from freebusy.organisation_v1.types.organisation_message import ListMembersRequest
from freebusy.organisation_v1.types.organisation_message import ListMembersResponse
from freebusy.organisation_v1.types.organisation_message import ListOrganisationsRequest
from freebusy.organisation_v1.types.organisation_message import ListOrganisationsResponse
from freebusy.organisation_v1.types.organisation_message import UpdateOrganisationRequest

__all__ = ('OrganisationServiceClient',
    'OrganisationServiceAsyncClient',
    'DeleteMemberRequest',
    'InviteMemberRequest',
    'InviteMemberResponse',
    'UpdateMemberRequest',
    'MemberState',
    'OrganisationRole',
    'OrganisationState',
    'InviteMemberArgs',
    'Member',
    'Organisation',
    'CreateOrganisationRequest',
    'DeleteOrganisationRequest',
    'GetMemberRequest',
    'GetOrganisationRequest',
    'ListMembersRequest',
    'ListMembersResponse',
    'ListOrganisationsRequest',
    'ListOrganisationsResponse',
    'UpdateOrganisationRequest',
)
