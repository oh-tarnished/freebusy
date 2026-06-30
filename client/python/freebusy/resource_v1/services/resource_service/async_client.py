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

from freebusy.resource_v1 import gapic_version as package_version

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

from freebusy.resource_v1.services.resource_service import pagers
from freebusy.resource_v1.types import enums as fr_enums
from freebusy.resource_v1.types import resource
from freebusy.resource_v1.types import resource as fr_resource
from freebusy.resource_v1.types import resource_messages
import freebusy.shared.v1.enums_pb2 as enums_pb2  # type: ignore
import google.protobuf.duration_pb2 as duration_pb2  # type: ignore
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore
import google.protobuf.struct_pb2 as struct_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore
from .transports.base import ResourceServiceTransport, DEFAULT_CLIENT_INFO
from .transports.grpc_asyncio import ResourceServiceGrpcAsyncIOTransport
from .client import ResourceServiceClient

try:
    from google.api_core import client_logging  # type: ignore
    CLIENT_LOGGING_SUPPORTED = True  # pragma: NO COVER
except ImportError:  # pragma: NO COVER
    CLIENT_LOGGING_SUPPORTED = False

_LOGGER = std_logging.getLogger(__name__)

class ResourceServiceAsyncClient:
    """ResourceService is the read-heavy catalog of bookable things
    (providers, rooms, equipment, unit types) and the offerings
    attached to them.
    """

    _client: ResourceServiceClient

    # Copy defaults from the synchronous client for use here.
    # Note: DEFAULT_ENDPOINT is deprecated. Use _DEFAULT_ENDPOINT_TEMPLATE instead.
    DEFAULT_ENDPOINT = ResourceServiceClient.DEFAULT_ENDPOINT
    DEFAULT_MTLS_ENDPOINT = ResourceServiceClient.DEFAULT_MTLS_ENDPOINT
    _DEFAULT_ENDPOINT_TEMPLATE = ResourceServiceClient._DEFAULT_ENDPOINT_TEMPLATE
    _DEFAULT_UNIVERSE = ResourceServiceClient._DEFAULT_UNIVERSE

    offering_path = staticmethod(ResourceServiceClient.offering_path)
    parse_offering_path = staticmethod(ResourceServiceClient.parse_offering_path)
    resource_path = staticmethod(ResourceServiceClient.resource_path)
    parse_resource_path = staticmethod(ResourceServiceClient.parse_resource_path)
    common_billing_account_path = staticmethod(ResourceServiceClient.common_billing_account_path)
    parse_common_billing_account_path = staticmethod(ResourceServiceClient.parse_common_billing_account_path)
    common_folder_path = staticmethod(ResourceServiceClient.common_folder_path)
    parse_common_folder_path = staticmethod(ResourceServiceClient.parse_common_folder_path)
    common_organization_path = staticmethod(ResourceServiceClient.common_organization_path)
    parse_common_organization_path = staticmethod(ResourceServiceClient.parse_common_organization_path)
    common_project_path = staticmethod(ResourceServiceClient.common_project_path)
    parse_common_project_path = staticmethod(ResourceServiceClient.parse_common_project_path)
    common_location_path = staticmethod(ResourceServiceClient.common_location_path)
    parse_common_location_path = staticmethod(ResourceServiceClient.parse_common_location_path)

    @classmethod
    def from_service_account_info(cls, info: dict, *args, **kwargs):
        """Creates an instance of this client using the provided credentials
            info.

        Args:
            info (dict): The service account private key info.
            args: Additional arguments to pass to the constructor.
            kwargs: Additional arguments to pass to the constructor.

        Returns:
            ResourceServiceAsyncClient: The constructed client.
        """
        sa_info_func = (
            ResourceServiceClient.from_service_account_info.__func__  # type: ignore
        )
        return sa_info_func(ResourceServiceAsyncClient, info, *args, **kwargs)

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
            ResourceServiceAsyncClient: The constructed client.
        """
        sa_file_func = (
            ResourceServiceClient.from_service_account_file.__func__  # type: ignore
        )
        return sa_file_func(ResourceServiceAsyncClient, filename, *args, **kwargs)

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
        return ResourceServiceClient.get_mtls_endpoint_and_cert_source(client_options)  # type: ignore

    @property
    def transport(self) -> ResourceServiceTransport:
        """Returns the transport used by the client instance.

        Returns:
            ResourceServiceTransport: The transport used by the client instance.
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

    get_transport_class = ResourceServiceClient.get_transport_class

    def __init__(self, *,
            credentials: Optional[ga_credentials.Credentials] = None,
            transport: Optional[Union[str, ResourceServiceTransport, Callable[..., ResourceServiceTransport]]] = "grpc_asyncio",
            client_options: Optional[ClientOptions] = None,
            client_info: gapic_v1.client_info.ClientInfo = DEFAULT_CLIENT_INFO,
            ) -> None:
        """Instantiates the resource service async client.

        Args:
            credentials (Optional[google.auth.credentials.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify the application to the service; if none
                are specified, the client will attempt to ascertain the
                credentials from the environment.
            transport (Optional[Union[str,ResourceServiceTransport,Callable[..., ResourceServiceTransport]]]):
                The transport to use, or a Callable that constructs and returns a new transport to use.
                If a Callable is given, it will be called with the same set of initialization
                arguments as used in the ResourceServiceTransport constructor.
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
        self._client = ResourceServiceClient(
            credentials=credentials,
            transport=transport,
            client_options=client_options,
            client_info=client_info,

        )

        if CLIENT_LOGGING_SUPPORTED and _LOGGER.isEnabledFor(std_logging.DEBUG):  # pragma: NO COVER
            _LOGGER.debug(
                "Created client `freebusy.resource_v1.ResourceServiceAsyncClient`.",
                extra = {
                    "serviceName": "freebusy.resource.v1.ResourceService",
                    "universeDomain": getattr(self._client._transport._credentials, "universe_domain", ""),
                    "credentialsType": f"{type(self._client._transport._credentials).__module__}.{type(self._client._transport._credentials).__qualname__}",
                    "credentialsInfo": getattr(self.transport._credentials, "get_cred_info", lambda: None)(),
                } if hasattr(self._client._transport, "_credentials") else {
                    "serviceName": "freebusy.resource.v1.ResourceService",
                    "credentialsType": None,
                }
            )

    async def list_resources(self,
            request: Optional[Union[resource_messages.ListResourcesRequest, dict]] = None,
            *,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListResourcesAsyncPager:
        r"""Lists resources.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_list_resources():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                request = resource_v1.ListResourcesRequest(
                )

                # Make the request
                page_result = client.list_resources(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.ListResourcesRequest, dict]]):
                The request object. Request message for ListResources.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.resource_v1.services.resource_service.pagers.ListResourcesAsyncPager:
                Response message for ListResources.

                Iterating over this object will yield
                results and resolve additional pages
                automatically.

        """
        # Create or coerce a protobuf request object.
        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, resource_messages.ListResourcesRequest):
            request = resource_messages.ListResourcesRequest(request)

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_resources]

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
        response = pagers.ListResourcesAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def get_resource(self,
            request: Optional[Union[resource_messages.GetResourceRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> resource.Resource:
        r"""Gets a single resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_get_resource():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                request = resource_v1.GetResourceRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_resource(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.GetResourceRequest, dict]]):
                The request object. Request message for GetResource.
            name (:class:`str`):
                The resource to retrieve.
                Format: resources/{resource}

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
            freebusy.resource_v1.types.Resource:
                A bookable thing: a provider, room, piece of equipment, or a unit type. A
                   resource is a pool of capacity interchangeable units;
                   the freebusy engine computes how many are free for a
                   given window. Its booking_mode decides whether
                   availability is produced as time slots or per-night
                   counts.

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
        if not isinstance(request, resource_messages.GetResourceRequest):
            request = resource_messages.GetResourceRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_resource]

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

    async def create_resource(self,
            request: Optional[Union[resource_messages.CreateResourceRequest, dict]] = None,
            *,
            resource: Optional[fr_resource.Resource] = None,
            resource_id: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> fr_resource.Resource:
        r"""Creates a resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_create_resource():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                resource = resource_v1.Resource()
                resource.display_name = "display_name_value"
                resource.type_ = "RESOURCE_TYPE_SPACE"
                resource.booking_mode = "BOOKING_MODE_NIGHTLY"
                resource.time_zone = "time_zone_value"

                request = resource_v1.CreateResourceRequest(
                    resource=resource,
                )

                # Make the request
                response = await client.create_resource(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.CreateResourceRequest, dict]]):
                The request object. Request message for CreateResource.
            resource (:class:`freebusy.resource_v1.types.Resource`):
                The resource to create. The name,
                state, and offerings fields are ignored.

                This corresponds to the ``resource`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            resource_id (:class:`str`):
                Optional caller-chosen ID for the
                resource; the server generates one if
                unset.

                This corresponds to the ``resource_id`` field
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
            freebusy.resource_v1.types.Resource:
                A bookable thing: a provider, room, piece of equipment, or a unit type. A
                   resource is a pool of capacity interchangeable units;
                   the freebusy engine computes how many are free for a
                   given window. Its booking_mode decides whether
                   availability is produced as time slots or per-night
                   counts.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [resource, resource_id]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, resource_messages.CreateResourceRequest):
            request = resource_messages.CreateResourceRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if resource is not None:
            request.resource = resource
        if resource_id is not None:
            request.resource_id = resource_id

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.create_resource]

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

    async def update_resource(self,
            request: Optional[Union[resource_messages.UpdateResourceRequest, dict]] = None,
            *,
            resource: Optional[fr_resource.Resource] = None,
            update_mask: Optional[field_mask_pb2.FieldMask] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> fr_resource.Resource:
        r"""Updates a resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_update_resource():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                resource = resource_v1.Resource()
                resource.display_name = "display_name_value"
                resource.type_ = "RESOURCE_TYPE_SPACE"
                resource.booking_mode = "BOOKING_MODE_NIGHTLY"
                resource.time_zone = "time_zone_value"

                request = resource_v1.UpdateResourceRequest(
                    resource=resource,
                )

                # Make the request
                response = await client.update_resource(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.UpdateResourceRequest, dict]]):
                The request object. Request message for UpdateResource.
            resource (:class:`freebusy.resource_v1.types.Resource`):
                The resource to update; its name
                identifies the target.

                This corresponds to the ``resource`` field
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
            freebusy.resource_v1.types.Resource:
                A bookable thing: a provider, room, piece of equipment, or a unit type. A
                   resource is a pool of capacity interchangeable units;
                   the freebusy engine computes how many are free for a
                   given window. Its booking_mode decides whether
                   availability is produced as time slots or per-night
                   counts.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [resource, update_mask]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, resource_messages.UpdateResourceRequest):
            request = resource_messages.UpdateResourceRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if resource is not None:
            request.resource = resource
        if update_mask is not None:
            request.update_mask = update_mask

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.update_resource]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("resource.name", request.resource.name),
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

    async def archive_resource(self,
            request: Optional[Union[resource_messages.ArchiveResourceRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> resource.Resource:
        r"""Archives a resource, hiding it from availability and
        new bookings.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_archive_resource():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                request = resource_v1.ArchiveResourceRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.archive_resource(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.ArchiveResourceRequest, dict]]):
                The request object. Request message for ArchiveResource.
            name (:class:`str`):
                The resource to archive.
                Format: resources/{resource}

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
            freebusy.resource_v1.types.Resource:
                A bookable thing: a provider, room, piece of equipment, or a unit type. A
                   resource is a pool of capacity interchangeable units;
                   the freebusy engine computes how many are free for a
                   given window. Its booking_mode decides whether
                   availability is produced as time slots or per-night
                   counts.

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
        if not isinstance(request, resource_messages.ArchiveResourceRequest):
            request = resource_messages.ArchiveResourceRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.archive_resource]

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

    async def unarchive_resource(self,
            request: Optional[Union[resource_messages.UnarchiveResourceRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> resource.Resource:
        r"""Unarchives a resource, restoring it to the active
        state.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_unarchive_resource():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                request = resource_v1.UnarchiveResourceRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.unarchive_resource(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.UnarchiveResourceRequest, dict]]):
                The request object. Request message for
                UnarchiveResource.
            name (:class:`str`):
                The resource to restore to the active
                state. Format: resources/{resource}

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
            freebusy.resource_v1.types.Resource:
                A bookable thing: a provider, room, piece of equipment, or a unit type. A
                   resource is a pool of capacity interchangeable units;
                   the freebusy engine computes how many are free for a
                   given window. Its booking_mode decides whether
                   availability is produced as time slots or per-night
                   counts.

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
        if not isinstance(request, resource_messages.UnarchiveResourceRequest):
            request = resource_messages.UnarchiveResourceRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.unarchive_resource]

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

    async def list_offerings(self,
            request: Optional[Union[resource_messages.ListOfferingsRequest, dict]] = None,
            *,
            parent: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListOfferingsAsyncPager:
        r"""Lists the offerings attached to a resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_list_offerings():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                request = resource_v1.ListOfferingsRequest(
                    parent="parent_value",
                )

                # Make the request
                page_result = client.list_offerings(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.ListOfferingsRequest, dict]]):
                The request object. Request message for ListOfferings.
            parent (:class:`str`):
                The parent resource whose offerings
                to list. Format: resources/{resource}

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
            freebusy.resource_v1.services.resource_service.pagers.ListOfferingsAsyncPager:
                Response message for ListOfferings.

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
        if not isinstance(request, resource_messages.ListOfferingsRequest):
            request = resource_messages.ListOfferingsRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if parent is not None:
            request.parent = parent

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_offerings]

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
        response = pagers.ListOfferingsAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def get_offering(self,
            request: Optional[Union[resource_messages.GetOfferingRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> resource.Offering:
        r"""Gets a single offering.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_get_offering():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                request = resource_v1.GetOfferingRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_offering(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.GetOfferingRequest, dict]]):
                The request object. Request message for GetOffering.
            name (:class:`str`):
                The offering to retrieve.
                Format:
                resources/{resource}/offerings/{offering}

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
            freebusy.resource_v1.types.Offering:
                A specific way a resource can be
                booked, carrying its duration and price.
                A "30-min consult" and a "60-min
                session" are two offerings on the same
                provider. For NIGHTLY resources the
                duration is unused and price is
                per-night.

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
        if not isinstance(request, resource_messages.GetOfferingRequest):
            request = resource_messages.GetOfferingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_offering]

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

    async def create_offering(self,
            request: Optional[Union[resource_messages.CreateOfferingRequest, dict]] = None,
            *,
            parent: Optional[str] = None,
            offering: Optional[resource.Offering] = None,
            offering_id: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> resource.Offering:
        r"""Creates an offering on a resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_create_offering():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                offering = resource_v1.Offering()
                offering.display_name = "display_name_value"

                request = resource_v1.CreateOfferingRequest(
                    parent="parent_value",
                    offering=offering,
                )

                # Make the request
                response = await client.create_offering(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.CreateOfferingRequest, dict]]):
                The request object. Request message for CreateOffering.
            parent (:class:`str`):
                The resource to attach the offering
                to. Format: resources/{resource}

                This corresponds to the ``parent`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            offering (:class:`freebusy.resource_v1.types.Offering`):
                The offering to create. Its name
                field is ignored.

                This corresponds to the ``offering`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            offering_id (:class:`str`):
                Optional caller-chosen ID for the
                offering; the server generates one if
                unset.

                This corresponds to the ``offering_id`` field
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
            freebusy.resource_v1.types.Offering:
                A specific way a resource can be
                booked, carrying its duration and price.
                A "30-min consult" and a "60-min
                session" are two offerings on the same
                provider. For NIGHTLY resources the
                duration is unused and price is
                per-night.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [parent, offering, offering_id]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, resource_messages.CreateOfferingRequest):
            request = resource_messages.CreateOfferingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if parent is not None:
            request.parent = parent
        if offering is not None:
            request.offering = offering
        if offering_id is not None:
            request.offering_id = offering_id

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.create_offering]

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

    async def update_offering(self,
            request: Optional[Union[resource_messages.UpdateOfferingRequest, dict]] = None,
            *,
            offering: Optional[resource.Offering] = None,
            update_mask: Optional[field_mask_pb2.FieldMask] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> resource.Offering:
        r"""Updates an offering.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_update_offering():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                offering = resource_v1.Offering()
                offering.display_name = "display_name_value"

                request = resource_v1.UpdateOfferingRequest(
                    offering=offering,
                )

                # Make the request
                response = await client.update_offering(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.UpdateOfferingRequest, dict]]):
                The request object. Request message for UpdateOffering.
            offering (:class:`freebusy.resource_v1.types.Offering`):
                The offering to update; its name
                identifies the target.

                This corresponds to the ``offering`` field
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
            freebusy.resource_v1.types.Offering:
                A specific way a resource can be
                booked, carrying its duration and price.
                A "30-min consult" and a "60-min
                session" are two offerings on the same
                provider. For NIGHTLY resources the
                duration is unused and price is
                per-night.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [offering, update_mask]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, resource_messages.UpdateOfferingRequest):
            request = resource_messages.UpdateOfferingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if offering is not None:
            request.offering = offering
        if update_mask is not None:
            request.update_mask = update_mask

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.update_offering]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("offering.name", request.offering.name),
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

    async def delete_offering(self,
            request: Optional[Union[resource_messages.DeleteOfferingRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> None:
        r"""Deletes an offering.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import resource_v1

            async def sample_delete_offering():
                # Create a client
                client = resource_v1.ResourceServiceAsyncClient()

                # Initialize request argument(s)
                request = resource_v1.DeleteOfferingRequest(
                    name="name_value",
                )

                # Make the request
                await client.delete_offering(request=request)

        Args:
            request (Optional[Union[freebusy.resource_v1.types.DeleteOfferingRequest, dict]]):
                The request object. Request message for DeleteOffering.
            name (:class:`str`):
                The offering to delete.
                Format:
                resources/{resource}/offerings/{offering}

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
        if not isinstance(request, resource_messages.DeleteOfferingRequest):
            request = resource_messages.DeleteOfferingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.delete_offering]

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

    async def __aenter__(self) -> "ResourceServiceAsyncClient":
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.transport.close()

DEFAULT_CLIENT_INFO = gapic_v1.client_info.ClientInfo(gapic_version=package_version.__version__)

if hasattr(DEFAULT_CLIENT_INFO, "protobuf_runtime_version"):   # pragma: NO COVER
    DEFAULT_CLIENT_INFO.protobuf_runtime_version = google.protobuf.__version__


__all__ = (
    "ResourceServiceAsyncClient",
)
