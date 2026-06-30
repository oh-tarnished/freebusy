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

from freebusy.schedule_v1 import gapic_version as package_version

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

from freebusy.schedule_v1.services.schedule_service import pagers
from freebusy.schedule_v1.types import enums as fs_enums
from freebusy.schedule_v1.types import schedule
from freebusy.schedule_v1.types import schedule as fs_schedule
from freebusy.schedule_v1.types import schedule_messages
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
from .transports.base import ScheduleServiceTransport, DEFAULT_CLIENT_INFO
from .transports.grpc_asyncio import ScheduleServiceGrpcAsyncIOTransport
from .client import ScheduleServiceClient

try:
    from google.api_core import client_logging  # type: ignore
    CLIENT_LOGGING_SUPPORTED = True  # pragma: NO COVER
except ImportError:  # pragma: NO COVER
    CLIENT_LOGGING_SUPPORTED = False

_LOGGER = std_logging.getLogger(__name__)

class ScheduleServiceAsyncClient:
    """ScheduleService is the write side of availability
    configuration: a resource's recurring working hours,
    blackout/holiday exceptions, buffers, and stay rules. These are
    the inputs the freebusy engine consumes to compute availability.
    """

    _client: ScheduleServiceClient

    # Copy defaults from the synchronous client for use here.
    # Note: DEFAULT_ENDPOINT is deprecated. Use _DEFAULT_ENDPOINT_TEMPLATE instead.
    DEFAULT_ENDPOINT = ScheduleServiceClient.DEFAULT_ENDPOINT
    DEFAULT_MTLS_ENDPOINT = ScheduleServiceClient.DEFAULT_MTLS_ENDPOINT
    _DEFAULT_ENDPOINT_TEMPLATE = ScheduleServiceClient._DEFAULT_ENDPOINT_TEMPLATE
    _DEFAULT_UNIVERSE = ScheduleServiceClient._DEFAULT_UNIVERSE

    availability_exception_path = staticmethod(ScheduleServiceClient.availability_exception_path)
    parse_availability_exception_path = staticmethod(ScheduleServiceClient.parse_availability_exception_path)
    schedule_path = staticmethod(ScheduleServiceClient.schedule_path)
    parse_schedule_path = staticmethod(ScheduleServiceClient.parse_schedule_path)
    common_billing_account_path = staticmethod(ScheduleServiceClient.common_billing_account_path)
    parse_common_billing_account_path = staticmethod(ScheduleServiceClient.parse_common_billing_account_path)
    common_folder_path = staticmethod(ScheduleServiceClient.common_folder_path)
    parse_common_folder_path = staticmethod(ScheduleServiceClient.parse_common_folder_path)
    common_organization_path = staticmethod(ScheduleServiceClient.common_organization_path)
    parse_common_organization_path = staticmethod(ScheduleServiceClient.parse_common_organization_path)
    common_project_path = staticmethod(ScheduleServiceClient.common_project_path)
    parse_common_project_path = staticmethod(ScheduleServiceClient.parse_common_project_path)
    common_location_path = staticmethod(ScheduleServiceClient.common_location_path)
    parse_common_location_path = staticmethod(ScheduleServiceClient.parse_common_location_path)

    @classmethod
    def from_service_account_info(cls, info: dict, *args, **kwargs):
        """Creates an instance of this client using the provided credentials
            info.

        Args:
            info (dict): The service account private key info.
            args: Additional arguments to pass to the constructor.
            kwargs: Additional arguments to pass to the constructor.

        Returns:
            ScheduleServiceAsyncClient: The constructed client.
        """
        sa_info_func = (
            ScheduleServiceClient.from_service_account_info.__func__  # type: ignore
        )
        return sa_info_func(ScheduleServiceAsyncClient, info, *args, **kwargs)

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
            ScheduleServiceAsyncClient: The constructed client.
        """
        sa_file_func = (
            ScheduleServiceClient.from_service_account_file.__func__  # type: ignore
        )
        return sa_file_func(ScheduleServiceAsyncClient, filename, *args, **kwargs)

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
        return ScheduleServiceClient.get_mtls_endpoint_and_cert_source(client_options)  # type: ignore

    @property
    def transport(self) -> ScheduleServiceTransport:
        """Returns the transport used by the client instance.

        Returns:
            ScheduleServiceTransport: The transport used by the client instance.
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

    get_transport_class = ScheduleServiceClient.get_transport_class

    def __init__(self, *,
            credentials: Optional[ga_credentials.Credentials] = None,
            transport: Optional[Union[str, ScheduleServiceTransport, Callable[..., ScheduleServiceTransport]]] = "grpc_asyncio",
            client_options: Optional[ClientOptions] = None,
            client_info: gapic_v1.client_info.ClientInfo = DEFAULT_CLIENT_INFO,
            ) -> None:
        """Instantiates the schedule service async client.

        Args:
            credentials (Optional[google.auth.credentials.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify the application to the service; if none
                are specified, the client will attempt to ascertain the
                credentials from the environment.
            transport (Optional[Union[str,ScheduleServiceTransport,Callable[..., ScheduleServiceTransport]]]):
                The transport to use, or a Callable that constructs and returns a new transport to use.
                If a Callable is given, it will be called with the same set of initialization
                arguments as used in the ScheduleServiceTransport constructor.
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
        self._client = ScheduleServiceClient(
            credentials=credentials,
            transport=transport,
            client_options=client_options,
            client_info=client_info,

        )

        if CLIENT_LOGGING_SUPPORTED and _LOGGER.isEnabledFor(std_logging.DEBUG):  # pragma: NO COVER
            _LOGGER.debug(
                "Created client `freebusy.schedule_v1.ScheduleServiceAsyncClient`.",
                extra = {
                    "serviceName": "freebusy.schedule.v1.ScheduleService",
                    "universeDomain": getattr(self._client._transport._credentials, "universe_domain", ""),
                    "credentialsType": f"{type(self._client._transport._credentials).__module__}.{type(self._client._transport._credentials).__qualname__}",
                    "credentialsInfo": getattr(self.transport._credentials, "get_cred_info", lambda: None)(),
                } if hasattr(self._client._transport, "_credentials") else {
                    "serviceName": "freebusy.schedule.v1.ScheduleService",
                    "credentialsType": None,
                }
            )

    async def get_schedule(self,
            request: Optional[Union[schedule_messages.GetScheduleRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> schedule.Schedule:
        r"""Reads the full availability configuration for a
        resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import schedule_v1

            async def sample_get_schedule():
                # Create a client
                client = schedule_v1.ScheduleServiceAsyncClient()

                # Initialize request argument(s)
                request = schedule_v1.GetScheduleRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_schedule(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.schedule_v1.types.GetScheduleRequest, dict]]):
                The request object. Request message for GetSchedule.
            name (:class:`str`):
                The schedule to read.
                Format: resources/{resource}/schedule

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
            freebusy.schedule_v1.types.Schedule:
                Aggregate read view of a resource's
                availability configuration: the inputs
                the freebusy engine consumes. Modeled as
                a singleton resource, one per resource.

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
        if not isinstance(request, schedule_messages.GetScheduleRequest):
            request = schedule_messages.GetScheduleRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_schedule]

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

    async def update_schedule(self,
            request: Optional[Union[schedule_messages.UpdateScheduleRequest, dict]] = None,
            *,
            schedule: Optional[fs_schedule.Schedule] = None,
            update_mask: Optional[field_mask_pb2.FieldMask] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> fs_schedule.Schedule:
        r"""Updates a resource's availability configuration. Set update_mask
        to the section(s) to replace: recurring_rules, buffers,
        stay_constraints, and/or cancellation_policy.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import schedule_v1

            async def sample_update_schedule():
                # Create a client
                client = schedule_v1.ScheduleServiceAsyncClient()

                # Initialize request argument(s)
                request = schedule_v1.UpdateScheduleRequest(
                )

                # Make the request
                response = await client.update_schedule(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.schedule_v1.types.UpdateScheduleRequest, dict]]):
                The request object. Request message for UpdateSchedule. Set update_mask to
                the section(s) to replace: "recurring_rules", "buffers",
                "stay_constraints", and/or "cancellation_policy".
            schedule (:class:`freebusy.schedule_v1.types.Schedule`):
                The schedule to update; its name
                identifies the target.

                This corresponds to the ``schedule`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            update_mask (:class:`google.protobuf.field_mask_pb2.FieldMask`):
                Fields to overwrite. Omit to replace
                all mutable sections.

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
            freebusy.schedule_v1.types.Schedule:
                Aggregate read view of a resource's
                availability configuration: the inputs
                the freebusy engine consumes. Modeled as
                a singleton resource, one per resource.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [schedule, update_mask]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, schedule_messages.UpdateScheduleRequest):
            request = schedule_messages.UpdateScheduleRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if schedule is not None:
            request.schedule = schedule
        if update_mask is not None:
            request.update_mask = update_mask

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.update_schedule]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("schedule.name", request.schedule.name),
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

    async def list_availability_exceptions(self,
            request: Optional[Union[schedule_messages.ListAvailabilityExceptionsRequest, dict]] = None,
            *,
            parent: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListAvailabilityExceptionsAsyncPager:
        r"""Lists the exceptions configured for a resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import schedule_v1

            async def sample_list_availability_exceptions():
                # Create a client
                client = schedule_v1.ScheduleServiceAsyncClient()

                # Initialize request argument(s)
                request = schedule_v1.ListAvailabilityExceptionsRequest(
                    parent="parent_value",
                )

                # Make the request
                page_result = client.list_availability_exceptions(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.schedule_v1.types.ListAvailabilityExceptionsRequest, dict]]):
                The request object. Request message for
                ListAvailabilityExceptions.
            parent (:class:`str`):
                The resource whose exceptions to
                list. Format: resources/{resource}

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
            freebusy.schedule_v1.services.schedule_service.pagers.ListAvailabilityExceptionsAsyncPager:
                Response message for
                ListAvailabilityExceptions.
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
        if not isinstance(request, schedule_messages.ListAvailabilityExceptionsRequest):
            request = schedule_messages.ListAvailabilityExceptionsRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if parent is not None:
            request.parent = parent

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_availability_exceptions]

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
        response = pagers.ListAvailabilityExceptionsAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def get_availability_exception(self,
            request: Optional[Union[schedule_messages.GetAvailabilityExceptionRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> schedule.AvailabilityException:
        r"""Gets a single availability exception.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import schedule_v1

            async def sample_get_availability_exception():
                # Create a client
                client = schedule_v1.ScheduleServiceAsyncClient()

                # Initialize request argument(s)
                request = schedule_v1.GetAvailabilityExceptionRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_availability_exception(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.schedule_v1.types.GetAvailabilityExceptionRequest, dict]]):
                The request object. Request message for
                GetAvailabilityException.
            name (:class:`str`):
                The exception to retrieve. Format:
                resources/{resource}/availabilityExceptions/{availability_exception}

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
            freebusy.schedule_v1.types.AvailabilityException:
                An override of a resource's normal
                hours on a specific span: a blackout /
                holiday closure, or extra hours beyond
                the recurring rules.

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
        if not isinstance(request, schedule_messages.GetAvailabilityExceptionRequest):
            request = schedule_messages.GetAvailabilityExceptionRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_availability_exception]

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

    async def create_availability_exception(self,
            request: Optional[Union[schedule_messages.CreateAvailabilityExceptionRequest, dict]] = None,
            *,
            parent: Optional[str] = None,
            availability_exception: Optional[schedule.AvailabilityException] = None,
            availability_exception_id: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> schedule.AvailabilityException:
        r"""Adds an availability exception to a resource.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import schedule_v1

            async def sample_create_availability_exception():
                # Create a client
                client = schedule_v1.ScheduleServiceAsyncClient()

                # Initialize request argument(s)
                availability_exception = schedule_v1.AvailabilityException()
                availability_exception.kind = "EXCEPTION_KIND_EXTRA_HOURS"

                request = schedule_v1.CreateAvailabilityExceptionRequest(
                    parent="parent_value",
                    availability_exception=availability_exception,
                )

                # Make the request
                response = await client.create_availability_exception(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.schedule_v1.types.CreateAvailabilityExceptionRequest, dict]]):
                The request object. Request message for
                CreateAvailabilityException.
            parent (:class:`str`):
                The resource to add the exception to.
                Format: resources/{resource}

                This corresponds to the ``parent`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            availability_exception (:class:`freebusy.schedule_v1.types.AvailabilityException`):
                The exception to add. Its name field
                is ignored.

                This corresponds to the ``availability_exception`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            availability_exception_id (:class:`str`):
                Optional caller-chosen ID for the
                exception; the server generates one if
                unset.

                This corresponds to the ``availability_exception_id`` field
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
            freebusy.schedule_v1.types.AvailabilityException:
                An override of a resource's normal
                hours on a specific span: a blackout /
                holiday closure, or extra hours beyond
                the recurring rules.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [parent, availability_exception, availability_exception_id]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, schedule_messages.CreateAvailabilityExceptionRequest):
            request = schedule_messages.CreateAvailabilityExceptionRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if parent is not None:
            request.parent = parent
        if availability_exception is not None:
            request.availability_exception = availability_exception
        if availability_exception_id is not None:
            request.availability_exception_id = availability_exception_id

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.create_availability_exception]

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

    async def delete_availability_exception(self,
            request: Optional[Union[schedule_messages.DeleteAvailabilityExceptionRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> None:
        r"""Removes an availability exception.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import schedule_v1

            async def sample_delete_availability_exception():
                # Create a client
                client = schedule_v1.ScheduleServiceAsyncClient()

                # Initialize request argument(s)
                request = schedule_v1.DeleteAvailabilityExceptionRequest(
                    name="name_value",
                )

                # Make the request
                await client.delete_availability_exception(request=request)

        Args:
            request (Optional[Union[freebusy.schedule_v1.types.DeleteAvailabilityExceptionRequest, dict]]):
                The request object. Request message for
                DeleteAvailabilityException.
            name (:class:`str`):
                The exception to remove. Format:
                resources/{resource}/availabilityExceptions/{availability_exception}

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
        if not isinstance(request, schedule_messages.DeleteAvailabilityExceptionRequest):
            request = schedule_messages.DeleteAvailabilityExceptionRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.delete_availability_exception]

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

    async def __aenter__(self) -> "ScheduleServiceAsyncClient":
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.transport.close()

DEFAULT_CLIENT_INFO = gapic_v1.client_info.ClientInfo(gapic_version=package_version.__version__)

if hasattr(DEFAULT_CLIENT_INFO, "protobuf_runtime_version"):   # pragma: NO COVER
    DEFAULT_CLIENT_INFO.protobuf_runtime_version = google.protobuf.__version__


__all__ = (
    "ScheduleServiceAsyncClient",
)
