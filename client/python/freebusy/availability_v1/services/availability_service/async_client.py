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

from freebusy.availability_v1 import gapic_version as package_version

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

from freebusy.availability_v1.services.availability_service import pagers
from freebusy.availability_v1.types import availability
import freebusy.shared.v1.enums_pb2 as enums_pb2  # type: ignore
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
from .transports.base import AvailabilityServiceTransport, DEFAULT_CLIENT_INFO
from .transports.grpc_asyncio import AvailabilityServiceGrpcAsyncIOTransport
from .client import AvailabilityServiceClient

try:
    from google.api_core import client_logging  # type: ignore
    CLIENT_LOGGING_SUPPORTED = True  # pragma: NO COVER
except ImportError:  # pragma: NO COVER
    CLIENT_LOGGING_SUPPORTED = False

_LOGGER = std_logging.getLogger(__name__)

class AvailabilityServiceAsyncClient:
    """AvailabilityService is the read-only, cacheable surface over the
    pure freebusy engine. It has no side effects: given a resource and a
    window it returns what is bookable, in the shape matching the
    resource's booking_mode.

    Its operations are custom methods (AIP-136): they compute results
    rather than fetch or list a resource, so they intentionally do not
    use the Get/List/Batch standard-method names or shapes.
    """

    _client: AvailabilityServiceClient

    # Copy defaults from the synchronous client for use here.
    # Note: DEFAULT_ENDPOINT is deprecated. Use _DEFAULT_ENDPOINT_TEMPLATE instead.
    DEFAULT_ENDPOINT = AvailabilityServiceClient.DEFAULT_ENDPOINT
    DEFAULT_MTLS_ENDPOINT = AvailabilityServiceClient.DEFAULT_MTLS_ENDPOINT
    _DEFAULT_ENDPOINT_TEMPLATE = AvailabilityServiceClient._DEFAULT_ENDPOINT_TEMPLATE
    _DEFAULT_UNIVERSE = AvailabilityServiceClient._DEFAULT_UNIVERSE

    common_billing_account_path = staticmethod(AvailabilityServiceClient.common_billing_account_path)
    parse_common_billing_account_path = staticmethod(AvailabilityServiceClient.parse_common_billing_account_path)
    common_folder_path = staticmethod(AvailabilityServiceClient.common_folder_path)
    parse_common_folder_path = staticmethod(AvailabilityServiceClient.parse_common_folder_path)
    common_organization_path = staticmethod(AvailabilityServiceClient.common_organization_path)
    parse_common_organization_path = staticmethod(AvailabilityServiceClient.parse_common_organization_path)
    common_project_path = staticmethod(AvailabilityServiceClient.common_project_path)
    parse_common_project_path = staticmethod(AvailabilityServiceClient.parse_common_project_path)
    common_location_path = staticmethod(AvailabilityServiceClient.common_location_path)
    parse_common_location_path = staticmethod(AvailabilityServiceClient.parse_common_location_path)

    @classmethod
    def from_service_account_info(cls, info: dict, *args, **kwargs):
        """Creates an instance of this client using the provided credentials
            info.

        Args:
            info (dict): The service account private key info.
            args: Additional arguments to pass to the constructor.
            kwargs: Additional arguments to pass to the constructor.

        Returns:
            AvailabilityServiceAsyncClient: The constructed client.
        """
        sa_info_func = (
            AvailabilityServiceClient.from_service_account_info.__func__  # type: ignore
        )
        return sa_info_func(AvailabilityServiceAsyncClient, info, *args, **kwargs)

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
            AvailabilityServiceAsyncClient: The constructed client.
        """
        sa_file_func = (
            AvailabilityServiceClient.from_service_account_file.__func__  # type: ignore
        )
        return sa_file_func(AvailabilityServiceAsyncClient, filename, *args, **kwargs)

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
        return AvailabilityServiceClient.get_mtls_endpoint_and_cert_source(client_options)  # type: ignore

    @property
    def transport(self) -> AvailabilityServiceTransport:
        """Returns the transport used by the client instance.

        Returns:
            AvailabilityServiceTransport: The transport used by the client instance.
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

    get_transport_class = AvailabilityServiceClient.get_transport_class

    def __init__(self, *,
            credentials: Optional[ga_credentials.Credentials] = None,
            transport: Optional[Union[str, AvailabilityServiceTransport, Callable[..., AvailabilityServiceTransport]]] = "grpc_asyncio",
            client_options: Optional[ClientOptions] = None,
            client_info: gapic_v1.client_info.ClientInfo = DEFAULT_CLIENT_INFO,
            ) -> None:
        """Instantiates the availability service async client.

        Args:
            credentials (Optional[google.auth.credentials.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify the application to the service; if none
                are specified, the client will attempt to ascertain the
                credentials from the environment.
            transport (Optional[Union[str,AvailabilityServiceTransport,Callable[..., AvailabilityServiceTransport]]]):
                The transport to use, or a Callable that constructs and returns a new transport to use.
                If a Callable is given, it will be called with the same set of initialization
                arguments as used in the AvailabilityServiceTransport constructor.
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
        self._client = AvailabilityServiceClient(
            credentials=credentials,
            transport=transport,
            client_options=client_options,
            client_info=client_info,

        )

        if CLIENT_LOGGING_SUPPORTED and _LOGGER.isEnabledFor(std_logging.DEBUG):  # pragma: NO COVER
            _LOGGER.debug(
                "Created client `freebusy.availability_v1.AvailabilityServiceAsyncClient`.",
                extra = {
                    "serviceName": "freebusy.availability.v1.AvailabilityService",
                    "universeDomain": getattr(self._client._transport._credentials, "universe_domain", ""),
                    "credentialsType": f"{type(self._client._transport._credentials).__module__}.{type(self._client._transport._credentials).__qualname__}",
                    "credentialsInfo": getattr(self.transport._credentials, "get_cred_info", lambda: None)(),
                } if hasattr(self._client._transport, "_credentials") else {
                    "serviceName": "freebusy.availability.v1.AvailabilityService",
                    "credentialsType": None,
                }
            )

    async def compute_availability(self,
            request: Optional[Union[availability.ComputeAvailabilityRequest, dict]] = None,
            *,
            resource: Optional[str] = None,
            window: Optional[types_pb2.TimeWindow] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> availability.ComputeAvailabilityResponse:
        r"""Computes availability for a resource over a window.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import availability_v1

            async def sample_compute_availability():
                # Create a client
                client = availability_v1.AvailabilityServiceAsyncClient()

                # Initialize request argument(s)
                request = availability_v1.ComputeAvailabilityRequest(
                    resource="resource_value",
                )

                # Make the request
                response = await client.compute_availability(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.availability_v1.types.ComputeAvailabilityRequest, dict]]):
                The request object. Request message for
                ComputeAvailability.
            resource (:class:`str`):
                The resource to compute availability
                for. Format: resources/{resource}

                This corresponds to the ``resource`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            window (:class:`freebusy.shared.v1.types_pb2.TimeWindow`):
                An exact time window, the natural form for TIME_SLOT
                resources.

                This corresponds to the ``window`` field
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
            freebusy.availability_v1.types.ComputeAvailabilityResponse:
                Response message for
                ComputeAvailability.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [resource, window]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, availability.ComputeAvailabilityRequest):
            request = availability.ComputeAvailabilityRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if resource is not None:
            request.resource = resource
        if window is not None:
            request.window = window

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.compute_availability]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("resource", request.resource),
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

    async def check_availability(self,
            request: Optional[Union[availability.CheckAvailabilityRequest, dict]] = None,
            *,
            resource: Optional[str] = None,
            window: Optional[types_pb2.TimeWindow] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> availability.CheckAvailabilityResponse:
        r"""Tests whether one exact span is bookable.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import availability_v1

            async def sample_check_availability():
                # Create a client
                client = availability_v1.AvailabilityServiceAsyncClient()

                # Initialize request argument(s)
                request = availability_v1.CheckAvailabilityRequest(
                    resource="resource_value",
                )

                # Make the request
                response = await client.check_availability(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.availability_v1.types.CheckAvailabilityRequest, dict]]):
                The request object. Request message for
                CheckAvailability.
            resource (:class:`str`):
                The resource to test.
                Format: resources/{resource}

                This corresponds to the ``resource`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            window (:class:`freebusy.shared.v1.types_pb2.TimeWindow`):
                An exact time window, the natural form for TIME_SLOT
                resources.

                This corresponds to the ``window`` field
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
            freebusy.availability_v1.types.CheckAvailabilityResponse:
                Response message for
                CheckAvailability.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [resource, window]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, availability.CheckAvailabilityRequest):
            request = availability.CheckAvailabilityRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if resource is not None:
            request.resource = resource
        if window is not None:
            request.window = window

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.check_availability]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("resource", request.resource),
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

    async def compute_bookable_ranges(self,
            request: Optional[Union[availability.ComputeBookableRangesRequest, dict]] = None,
            *,
            resource: Optional[str] = None,
            window: Optional[types_pb2.TimeWindow] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> availability.ComputeBookableRangesResponse:
        r"""Computes contiguous bookable ranges within a window.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import availability_v1

            async def sample_compute_bookable_ranges():
                # Create a client
                client = availability_v1.AvailabilityServiceAsyncClient()

                # Initialize request argument(s)
                request = availability_v1.ComputeBookableRangesRequest(
                    resource="resource_value",
                )

                # Make the request
                response = await client.compute_bookable_ranges(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.availability_v1.types.ComputeBookableRangesRequest, dict]]):
                The request object. Request message for
                ComputeBookableRanges.
            resource (:class:`str`):
                The resource to compute bookable
                ranges for. Format: resources/{resource}

                This corresponds to the ``resource`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            window (:class:`freebusy.shared.v1.types_pb2.TimeWindow`):
                An exact time window, the natural form for TIME_SLOT
                resources.

                This corresponds to the ``window`` field
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
            freebusy.availability_v1.types.ComputeBookableRangesResponse:
                Response message for
                ComputeBookableRanges.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [resource, window]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, availability.ComputeBookableRangesRequest):
            request = availability.ComputeBookableRangesRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if resource is not None:
            request.resource = resource
        if window is not None:
            request.window = window

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.compute_bookable_ranges]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("resource", request.resource),
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

    async def batch_compute_availability(self,
            request: Optional[Union[availability.BatchComputeAvailabilityRequest, dict]] = None,
            *,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> availability.BatchComputeAvailabilityResponse:
        r"""Computes availability for several resources at once.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import availability_v1

            async def sample_batch_compute_availability():
                # Create a client
                client = availability_v1.AvailabilityServiceAsyncClient()

                # Initialize request argument(s)
                requests = availability_v1.ComputeAvailabilityRequest()
                requests.resource = "resource_value"

                request = availability_v1.BatchComputeAvailabilityRequest(
                    requests=requests,
                )

                # Make the request
                response = await client.batch_compute_availability(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.availability_v1.types.BatchComputeAvailabilityRequest, dict]]):
                The request object. Request message for
                BatchComputeAvailability. Each entry is
                a full ComputeAvailabilityRequest
                (AIP-231), so per-resource duration,
                offering, and units all work in batch
                exactly as they do in the single call.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.availability_v1.types.BatchComputeAvailabilityResponse:
                Response message for
                BatchComputeAvailability.

        """
        # Create or coerce a protobuf request object.
        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, availability.BatchComputeAvailabilityRequest):
            request = availability.BatchComputeAvailabilityRequest(request)

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.batch_compute_availability]

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

    async def search_availability(self,
            request: Optional[Union[availability.SearchAvailabilityRequest, dict]] = None,
            *,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.SearchAvailabilityAsyncPager:
        r"""Searches the catalog for resources bookable over a
        period.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import availability_v1

            async def sample_search_availability():
                # Create a client
                client = availability_v1.AvailabilityServiceAsyncClient()

                # Initialize request argument(s)
                request = availability_v1.SearchAvailabilityRequest(
                )

                # Make the request
                page_result = client.search_availability(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.availability_v1.types.SearchAvailabilityRequest, dict]]):
                The request object. Request message for
                SearchAvailability. Sweeps the catalog
                for resources that are bookable over a
                period for a given party size, narrowed
                by a resource filter and sorted for
                presentation. This is the storefront
                query: one call returns the matching
                resources with a lead price, rather than
                the caller listing resources and
                computing availability for each.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.availability_v1.services.availability_service.pagers.SearchAvailabilityAsyncPager:
                Response message for
                SearchAvailability.
                Iterating over this object will yield
                results and resolve additional pages
                automatically.

        """
        # Create or coerce a protobuf request object.
        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, availability.SearchAvailabilityRequest):
            request = availability.SearchAvailabilityRequest(request)

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.search_availability]

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
        response = pagers.SearchAvailabilityAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def __aenter__(self) -> "AvailabilityServiceAsyncClient":
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.transport.close()

DEFAULT_CLIENT_INFO = gapic_v1.client_info.ClientInfo(gapic_version=package_version.__version__)

if hasattr(DEFAULT_CLIENT_INFO, "protobuf_runtime_version"):   # pragma: NO COVER
    DEFAULT_CLIENT_INFO.protobuf_runtime_version = google.protobuf.__version__


__all__ = (
    "AvailabilityServiceAsyncClient",
)
