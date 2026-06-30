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
import logging as std_logging
from collections import OrderedDict
import re
from typing import Dict, Callable, Mapping, MutableMapping, MutableSequence, Optional, Sequence, Tuple, Type, Union

from freebusy.organisation_v1 import gapic_version as package_version

from google.api_core.client_options import ClientOptions
from google.api_core import exceptions as core_exceptions
from google.api_core import gapic_v1
from google.api_core import retry_async as retries
from google.auth import credentials as ga_credentials   # type: ignore
from google.oauth2 import service_account              # type: ignore
import google.protobuf


try:
    OptionalRetry = Union[retries.AsyncRetry, gapic_v1.method._MethodDefault, None]
except AttributeError:  # pragma: NO COVER
    OptionalRetry = Union[retries.AsyncRetry, object, None]  # type: ignore

from freebusy.organisation_v1.services.organisation_service import pagers
from freebusy.organisation_v1.types import actions
from freebusy.organisation_v1.types import enums
from freebusy.organisation_v1.types import organisation
from freebusy.organisation_v1.types import organisation as fo_organisation
from freebusy.organisation_v1.types import organisation_message
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore
import google.protobuf.struct_pb2 as struct_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
from .transports.base import OrganisationServiceTransport, DEFAULT_CLIENT_INFO
from .transports.grpc_asyncio import OrganisationServiceGrpcAsyncIOTransport
from .client import OrganisationServiceClient

try:
    from google.api_core import client_logging  # type: ignore
    CLIENT_LOGGING_SUPPORTED = True  # pragma: NO COVER
except ImportError:  # pragma: NO COVER
    CLIENT_LOGGING_SUPPORTED = False

_LOGGER = std_logging.getLogger(__name__)

class OrganisationServiceAsyncClient:
    """OrganisationService manages tenants and their members.
    Day-to-day tenancy is enforced by the shell via row-level
    security from the caller's organisation; this service is where
    organisations and members are created and administered.
    """

    _client: OrganisationServiceClient

    # Copy defaults from the synchronous client for use here.
    # Note: DEFAULT_ENDPOINT is deprecated. Use _DEFAULT_ENDPOINT_TEMPLATE instead.
    DEFAULT_ENDPOINT = OrganisationServiceClient.DEFAULT_ENDPOINT
    DEFAULT_MTLS_ENDPOINT = OrganisationServiceClient.DEFAULT_MTLS_ENDPOINT
    _DEFAULT_ENDPOINT_TEMPLATE = OrganisationServiceClient._DEFAULT_ENDPOINT_TEMPLATE
    _DEFAULT_UNIVERSE = OrganisationServiceClient._DEFAULT_UNIVERSE

    member_path = staticmethod(OrganisationServiceClient.member_path)
    parse_member_path = staticmethod(OrganisationServiceClient.parse_member_path)
    organisation_path = staticmethod(OrganisationServiceClient.organisation_path)
    parse_organisation_path = staticmethod(OrganisationServiceClient.parse_organisation_path)
    common_billing_account_path = staticmethod(OrganisationServiceClient.common_billing_account_path)
    parse_common_billing_account_path = staticmethod(OrganisationServiceClient.parse_common_billing_account_path)
    common_folder_path = staticmethod(OrganisationServiceClient.common_folder_path)
    parse_common_folder_path = staticmethod(OrganisationServiceClient.parse_common_folder_path)
    common_organization_path = staticmethod(OrganisationServiceClient.common_organization_path)
    parse_common_organization_path = staticmethod(OrganisationServiceClient.parse_common_organization_path)
    common_project_path = staticmethod(OrganisationServiceClient.common_project_path)
    parse_common_project_path = staticmethod(OrganisationServiceClient.parse_common_project_path)
    common_location_path = staticmethod(OrganisationServiceClient.common_location_path)
    parse_common_location_path = staticmethod(OrganisationServiceClient.parse_common_location_path)

    @classmethod
    def from_service_account_info(cls, info: dict, *args, **kwargs):
        """Creates an instance of this client using the provided credentials
            info.

        Args:
            info (dict): The service account private key info.
            args: Additional arguments to pass to the constructor.
            kwargs: Additional arguments to pass to the constructor.

        Returns:
            OrganisationServiceAsyncClient: The constructed client.
        """
        sa_info_func = (
            OrganisationServiceClient.from_service_account_info.__func__  # type: ignore
        )
        return sa_info_func(OrganisationServiceAsyncClient, info, *args, **kwargs)

    @classmethod
    def from_service_account_file(cls, filename: str, *args, **kwargs):
        """Creates an instance of this client using the provided credentials
            file.

        Args:
            filename (str): The path to the service account private key json
                file.
            args: Additional arguments to pass to the constructor.
            kwargs: Additional arguments to pass to the constructor.

        Returns:
            OrganisationServiceAsyncClient: The constructed client.
        """
        sa_file_func = (
            OrganisationServiceClient.from_service_account_file.__func__  # type: ignore
        )
        return sa_file_func(OrganisationServiceAsyncClient, filename, *args, **kwargs)

    from_service_account_json = from_service_account_file

    @classmethod
    def get_mtls_endpoint_and_cert_source(cls, client_options: Optional[ClientOptions] = None):
        """Return the API endpoint and client cert source for mutual TLS.

        The client cert source is determined in the following order:
        (1) if `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is not "true", the
        client cert source is None.
        (2) if `client_options.client_cert_source` is provided, use the provided one; if the
        default client cert source exists, use the default one; otherwise the client cert
        source is None.

        The API endpoint is determined in the following order:
        (1) if `client_options.api_endpoint` if provided, use the provided one.
        (2) if `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is "always", use the
        default mTLS endpoint; if the environment variable is "never", use the default API
        endpoint; otherwise if client cert source exists, use the default mTLS endpoint, otherwise
        use the default API endpoint.

        More details can be found at https://google.aip.dev/auth/4114.

        Args:
            client_options (google.api_core.client_options.ClientOptions): Custom options for the
                client. Only the `api_endpoint` and `client_cert_source` properties may be used
                in this method.

        Returns:
            Tuple[str, Callable[[], Tuple[bytes, bytes]]]: returns the API endpoint and the
                client cert source to use.

        Raises:
            google.auth.exceptions.MutualTLSChannelError: If any errors happen.
        """
        return OrganisationServiceClient.get_mtls_endpoint_and_cert_source(client_options)  # type: ignore

    @property
    def transport(self) -> OrganisationServiceTransport:
        """Returns the transport used by the client instance.

        Returns:
            OrganisationServiceTransport: The transport used by the client instance.
        """
        return self._client.transport

    @property
    def api_endpoint(self) -> str:
        """Return the API endpoint used by the client instance.

        Returns:
            str: The API endpoint used by the client instance.
        """
        return self._client._api_endpoint

    @property
    def universe_domain(self) -> str:
        """Return the universe domain used by the client instance.

        Returns:
            str: The universe domain used
                by the client instance.
        """
        return self._client._universe_domain

    get_transport_class = OrganisationServiceClient.get_transport_class

    def __init__(self, *,
            credentials: Optional[ga_credentials.Credentials] = None,
            transport: Optional[Union[str, OrganisationServiceTransport, Callable[..., OrganisationServiceTransport]]] = "grpc_asyncio",
            client_options: Optional[ClientOptions] = None,
            client_info: gapic_v1.client_info.ClientInfo = DEFAULT_CLIENT_INFO,
            ) -> None:
        """Instantiates the organisation service async client.

        Args:
            credentials (Optional[google.auth.credentials.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify the application to the service; if none
                are specified, the client will attempt to ascertain the
                credentials from the environment.
            transport (Optional[Union[str,OrganisationServiceTransport,Callable[..., OrganisationServiceTransport]]]):
                The transport to use, or a Callable that constructs and returns a new transport to use.
                If a Callable is given, it will be called with the same set of initialization
                arguments as used in the OrganisationServiceTransport constructor.
                If set to None, a transport is chosen automatically.
            client_options (Optional[Union[google.api_core.client_options.ClientOptions, dict]]):
                Custom options for the client.

                1. The ``api_endpoint`` property can be used to override the
                default endpoint provided by the client when ``transport`` is
                not explicitly provided. Only if this property is not set and
                ``transport`` was not explicitly provided, the endpoint is
                determined by the GOOGLE_API_USE_MTLS_ENDPOINT environment
                variable, which have one of the following values:
                "always" (always use the default mTLS endpoint), "never" (always
                use the default regular endpoint) and "auto" (auto-switch to the
                default mTLS endpoint if client certificate is present; this is
                the default value).

                2. If the GOOGLE_API_USE_CLIENT_CERTIFICATE environment variable
                is "true", then the ``client_cert_source`` property can be used
                to provide a client certificate for mTLS transport. If
                not provided, the default SSL client certificate will be used if
                present. If GOOGLE_API_USE_CLIENT_CERTIFICATE is "false" or not
                set, no client certificate will be used.

                3. The ``universe_domain`` property can be used to override the
                default "googleapis.com" universe. Note that ``api_endpoint``
                property still takes precedence; and ``universe_domain`` is
                currently not supported for mTLS.

            client_info (google.api_core.gapic_v1.client_info.ClientInfo):
                The client info used to send a user-agent string along with
                API requests. If ``None``, then default info will be used.
                Generally, you only need to set this if you're developing
                your own client library.

        Raises:
            google.auth.exceptions.MutualTlsChannelError: If mutual TLS transport
                creation failed for any reason.
        """
        self._client = OrganisationServiceClient(
            credentials=credentials,
            transport=transport,
            client_options=client_options,
            client_info=client_info,

        )

        if CLIENT_LOGGING_SUPPORTED and _LOGGER.isEnabledFor(std_logging.DEBUG):  # pragma: NO COVER
            _LOGGER.debug(
                "Created client `freebusy.organisation_v1.OrganisationServiceAsyncClient`.",
                extra = {
                    "serviceName": "freebusy.organisation.v1.OrganisationService",
                    "universeDomain": getattr(self._client._transport._credentials, "universe_domain", ""),
                    "credentialsType": f"{type(self._client._transport._credentials).__module__}.{type(self._client._transport._credentials).__qualname__}",
                    "credentialsInfo": getattr(self.transport._credentials, "get_cred_info", lambda: None)(),
                } if hasattr(self._client._transport, "_credentials") else {
                    "serviceName": "freebusy.organisation.v1.OrganisationService",
                    "credentialsType": None,
                }
            )

    async def list_organisations(self,
            request: Optional[Union[organisation_message.ListOrganisationsRequest, dict]] = None,
            *,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListOrganisationsAsyncPager:
        r"""Lists the organisations the caller belongs to.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_list_organisations():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                request = organisation_v1.ListOrganisationsRequest(
                )

                # Make the request
                page_result = client.list_organisations(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.ListOrganisationsRequest, dict]]):
                The request object. Request message for
                ListOrganisations.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.services.organisation_service.pagers.ListOrganisationsAsyncPager:
                Response message for
                ListOrganisations.
                Iterating over this object will yield
                results and resolve additional pages
                automatically.

        """
        # Create or coerce a protobuf request object.
        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, organisation_message.ListOrganisationsRequest):
            request = organisation_message.ListOrganisationsRequest(request)

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_organisations]

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # This method is paged; wrap the response in a pager, which provides
        # an `__aiter__` convenience method.
        response = pagers.ListOrganisationsAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def get_organisation(self,
            request: Optional[Union[organisation_message.GetOrganisationRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> organisation.Organisation:
        r"""Gets a single organisation.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_get_organisation():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                request = organisation_v1.GetOrganisationRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_organisation(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.GetOrganisationRequest, dict]]):
                The request object. Request message for GetOrganisation.
            name (:class:`str`):
                The organisation to retrieve.
                Format: organisations/{organisation}

                This corresponds to the ``name`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.types.Organisation:
                A tenant. Organisation is the unit of
                multi-tenancy; the shell enforces
                isolation with row-level security keyed
                off the caller's organisation, so most
                resource names stay flat and the
                organisation appears explicitly only
                here.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [name]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, organisation_message.GetOrganisationRequest):
            request = organisation_message.GetOrganisationRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_organisation]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("name", request.name),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def create_organisation(self,
            request: Optional[Union[organisation_message.CreateOrganisationRequest, dict]] = None,
            *,
            organisation: Optional[fo_organisation.Organisation] = None,
            organisation_id: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> fo_organisation.Organisation:
        r"""Creates an organisation. The caller becomes its first
        owner.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_create_organisation():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                organisation = organisation_v1.Organisation()
                organisation.display_name = "display_name_value"

                request = organisation_v1.CreateOrganisationRequest(
                    organisation=organisation,
                )

                # Make the request
                response = await client.create_organisation(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.CreateOrganisationRequest, dict]]):
                The request object. Request message for
                CreateOrganisation.
            organisation (:class:`freebusy.organisation_v1.types.Organisation`):
                The organisation to create. The name
                and state fields are ignored. The caller
                becomes the organisation's first OWNER.

                This corresponds to the ``organisation`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            organisation_id (:class:`str`):
                Optional caller-chosen ID for the
                organisation; the server generates one
                if unset.

                This corresponds to the ``organisation_id`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.types.Organisation:
                A tenant. Organisation is the unit of
                multi-tenancy; the shell enforces
                isolation with row-level security keyed
                off the caller's organisation, so most
                resource names stay flat and the
                organisation appears explicitly only
                here.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [organisation, organisation_id]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, organisation_message.CreateOrganisationRequest):
            request = organisation_message.CreateOrganisationRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if organisation is not None:
            request.organisation = organisation
        if organisation_id is not None:
            request.organisation_id = organisation_id

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.create_organisation]

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def update_organisation(self,
            request: Optional[Union[organisation_message.UpdateOrganisationRequest, dict]] = None,
            *,
            organisation: Optional[fo_organisation.Organisation] = None,
            update_mask: Optional[field_mask_pb2.FieldMask] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> fo_organisation.Organisation:
        r"""Updates an organisation.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_update_organisation():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                organisation = organisation_v1.Organisation()
                organisation.display_name = "display_name_value"

                request = organisation_v1.UpdateOrganisationRequest(
                    organisation=organisation,
                )

                # Make the request
                response = await client.update_organisation(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.UpdateOrganisationRequest, dict]]):
                The request object. Request message for
                UpdateOrganisation.
            organisation (:class:`freebusy.organisation_v1.types.Organisation`):
                The organisation to update; its name
                identifies the target.

                This corresponds to the ``organisation`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            update_mask (:class:`google.protobuf.field_mask_pb2.FieldMask`):
                Fields to overwrite. Omit to replace
                all mutable fields.

                This corresponds to the ``update_mask`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.types.Organisation:
                A tenant. Organisation is the unit of
                multi-tenancy; the shell enforces
                isolation with row-level security keyed
                off the caller's organisation, so most
                resource names stay flat and the
                organisation appears explicitly only
                here.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [organisation, update_mask]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, organisation_message.UpdateOrganisationRequest):
            request = organisation_message.UpdateOrganisationRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if organisation is not None:
            request.organisation = organisation
        if update_mask is not None:
            request.update_mask = update_mask

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.update_organisation]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("organisation.name", request.organisation.name),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def delete_organisation(self,
            request: Optional[Union[organisation_message.DeleteOrganisationRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> None:
        r"""Deletes an organisation.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_delete_organisation():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                request = organisation_v1.DeleteOrganisationRequest(
                    name="name_value",
                )

                # Make the request
                await client.delete_organisation(request=request)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.DeleteOrganisationRequest, dict]]):
                The request object. Request message for
                DeleteOrganisation.
            name (:class:`str`):
                The organisation to delete.
                Format: organisations/{organisation}

                This corresponds to the ``name`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.
        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [name]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, organisation_message.DeleteOrganisationRequest):
            request = organisation_message.DeleteOrganisationRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.delete_organisation]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("name", request.name),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

    async def invite_member(self,
            request: Optional[Union[actions.InviteMemberRequest, dict]] = None,
            *,
            parent: Optional[str] = None,
            email: Optional[str] = None,
            role: Optional[enums.OrganisationRole] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> actions.InviteMemberResponse:
        r"""Invites a member to an organisation.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_invite_member():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                request = organisation_v1.InviteMemberRequest(
                    parent="parent_value",
                    email="email_value",
                    role="ORGANISATION_ROLE_VIEWER",
                )

                # Make the request
                response = await client.invite_member(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.InviteMemberRequest, dict]]):
                The request object. Request message for InviteMember.
            parent (:class:`str`):
                The organisation to invite the member
                to. Format: organisations/{organisation}

                This corresponds to the ``parent`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            email (:class:`str`):
                Email address to invite.
                This corresponds to the ``email`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            role (:class:`freebusy.organisation_v1.types.OrganisationRole`):
                Role to grant on acceptance.
                This corresponds to the ``role`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.types.InviteMemberResponse:
                Response message for InviteMember.
        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [parent, email, role]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, actions.InviteMemberRequest):
            request = actions.InviteMemberRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if parent is not None:
            request.parent = parent
        if email is not None:
            request.email = email
        if role is not None:
            request.role = role

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.invite_member]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("parent", request.parent),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def list_members(self,
            request: Optional[Union[organisation_message.ListMembersRequest, dict]] = None,
            *,
            parent: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListMembersAsyncPager:
        r"""Lists the members of an organisation.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_list_members():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                request = organisation_v1.ListMembersRequest(
                    parent="parent_value",
                )

                # Make the request
                page_result = client.list_members(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.ListMembersRequest, dict]]):
                The request object. Request message for ListMembers.
            parent (:class:`str`):
                The organisation whose members to
                list. Format:
                organisations/{organisation}

                This corresponds to the ``parent`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.services.organisation_service.pagers.ListMembersAsyncPager:
                Response message for ListMembers.

                Iterating over this object will yield
                results and resolve additional pages
                automatically.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [parent]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, organisation_message.ListMembersRequest):
            request = organisation_message.ListMembersRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if parent is not None:
            request.parent = parent

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_members]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("parent", request.parent),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # This method is paged; wrap the response in a pager, which provides
        # an `__aiter__` convenience method.
        response = pagers.ListMembersAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def get_member(self,
            request: Optional[Union[organisation_message.GetMemberRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> organisation.Member:
        r"""Gets a single member.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_get_member():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                request = organisation_v1.GetMemberRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_member(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.GetMemberRequest, dict]]):
                The request object. Request message for GetMember.
            name (:class:`str`):
                The member to retrieve.
                Format:
                organisations/{organisation}/members/{member}

                This corresponds to the ``name`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.types.Member:
                The membership of a user in an
                organisation, with their role.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [name]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, organisation_message.GetMemberRequest):
            request = organisation_message.GetMemberRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_member]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("name", request.name),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def update_member(self,
            request: Optional[Union[actions.UpdateMemberRequest, dict]] = None,
            *,
            member: Optional[organisation.Member] = None,
            update_mask: Optional[field_mask_pb2.FieldMask] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> organisation.Member:
        r"""Updates a member; the role is the only mutable field.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_update_member():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                member = organisation_v1.Member()
                member.email = "email_value"
                member.role = "ORGANISATION_ROLE_VIEWER"

                request = organisation_v1.UpdateMemberRequest(
                    member=member,
                )

                # Make the request
                response = await client.update_member(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.UpdateMemberRequest, dict]]):
                The request object. Request message for UpdateMember. The role is the only
                mutable field; set update_mask to "role" to change it.
            member (:class:`freebusy.organisation_v1.types.Member`):
                The member to update; its name
                identifies the target.

                This corresponds to the ``member`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            update_mask (:class:`google.protobuf.field_mask_pb2.FieldMask`):
                Fields to overwrite. Omit to replace
                all mutable fields.

                This corresponds to the ``update_mask`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.organisation_v1.types.Member:
                The membership of a user in an
                organisation, with their role.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [member, update_mask]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, actions.UpdateMemberRequest):
            request = actions.UpdateMemberRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if member is not None:
            request.member = member
        if update_mask is not None:
            request.update_mask = update_mask

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.update_member]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("member.name", request.member.name),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        response = await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def delete_member(self,
            request: Optional[Union[actions.DeleteMemberRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> None:
        r"""Removes a member from an organisation.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import organisation_v1

            async def sample_delete_member():
                # Create a client
                client = organisation_v1.OrganisationServiceAsyncClient()

                # Initialize request argument(s)
                request = organisation_v1.DeleteMemberRequest(
                    name="name_value",
                )

                # Make the request
                await client.delete_member(request=request)

        Args:
            request (Optional[Union[freebusy.organisation_v1.types.DeleteMemberRequest, dict]]):
                The request object. Request message for DeleteMember.
            name (:class:`str`):
                The member to remove.
                Format:
                organisations/{organisation}/members/{member}

                This corresponds to the ``name`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.
        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [name]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, actions.DeleteMemberRequest):
            request = actions.DeleteMemberRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.delete_member]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("name", request.name),
            )),
        )

        # Validate the universe domain.
        self._client._validate_universe_domain()

        # Send the request.
        await rpc(
            request,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

    async def __aenter__(self) -> "OrganisationServiceAsyncClient":
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.transport.close()

DEFAULT_CLIENT_INFO = gapic_v1.client_info.ClientInfo(gapic_version=package_version.__version__)

if hasattr(DEFAULT_CLIENT_INFO, "protobuf_runtime_version"):   # pragma: NO COVER
    DEFAULT_CLIENT_INFO.protobuf_runtime_version = google.protobuf.__version__


__all__ = (
    "OrganisationServiceAsyncClient",
)
