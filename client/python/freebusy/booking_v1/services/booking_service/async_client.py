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

from freebusy.booking_v1 import gapic_version as package_version

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

from freebusy.booking_v1.services.booking_service import pagers
from freebusy.booking_v1.types import booking
from freebusy.booking_v1.types import booking as fb_booking
from freebusy.booking_v1.types import booking_actions
from freebusy.booking_v1.types import booking_messages
from freebusy.booking_v1.types import enums
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
import google.protobuf.duration_pb2 as duration_pb2  # type: ignore
import google.protobuf.struct_pb2 as struct_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore
from .transports.base import BookingServiceTransport, DEFAULT_CLIENT_INFO
from .transports.grpc_asyncio import BookingServiceGrpcAsyncIOTransport
from .client import BookingServiceClient

try:
    from google.api_core import client_logging  # type: ignore
    CLIENT_LOGGING_SUPPORTED = True  # pragma: NO COVER
except ImportError:  # pragma: NO COVER
    CLIENT_LOGGING_SUPPORTED = False

_LOGGER = std_logging.getLogger(__name__)

class BookingServiceAsyncClient:
    """BookingService is the transactional heart. CreateBooking
    carries the idempotency key, runs policy checks, and does the
    exclusion-constraint insert that places a hold. The hold
    lifecycle is modeled as booking states. Two things are
    deliberately NOT RPCs: the sweeper that releases expired holds
    is an internal goroutine/cron, and confirmation usually arrives
    as a payment webhook that calls ConfirmBooking server-side.
    """

    _client: BookingServiceClient

    # Copy defaults from the synchronous client for use here.
    # Note: DEFAULT_ENDPOINT is deprecated. Use _DEFAULT_ENDPOINT_TEMPLATE instead.
    DEFAULT_ENDPOINT = BookingServiceClient.DEFAULT_ENDPOINT
    DEFAULT_MTLS_ENDPOINT = BookingServiceClient.DEFAULT_MTLS_ENDPOINT
    _DEFAULT_ENDPOINT_TEMPLATE = BookingServiceClient._DEFAULT_ENDPOINT_TEMPLATE
    _DEFAULT_UNIVERSE = BookingServiceClient._DEFAULT_UNIVERSE

    booking_path = staticmethod(BookingServiceClient.booking_path)
    parse_booking_path = staticmethod(BookingServiceClient.parse_booking_path)
    common_billing_account_path = staticmethod(BookingServiceClient.common_billing_account_path)
    parse_common_billing_account_path = staticmethod(BookingServiceClient.parse_common_billing_account_path)
    common_folder_path = staticmethod(BookingServiceClient.common_folder_path)
    parse_common_folder_path = staticmethod(BookingServiceClient.parse_common_folder_path)
    common_organization_path = staticmethod(BookingServiceClient.common_organization_path)
    parse_common_organization_path = staticmethod(BookingServiceClient.parse_common_organization_path)
    common_project_path = staticmethod(BookingServiceClient.common_project_path)
    parse_common_project_path = staticmethod(BookingServiceClient.parse_common_project_path)
    common_location_path = staticmethod(BookingServiceClient.common_location_path)
    parse_common_location_path = staticmethod(BookingServiceClient.parse_common_location_path)

    @classmethod
    def from_service_account_info(cls, info: dict, *args, **kwargs):
        """Creates an instance of this client using the provided credentials
            info.

        Args:
            info (dict): The service account private key info.
            args: Additional arguments to pass to the constructor.
            kwargs: Additional arguments to pass to the constructor.

        Returns:
            BookingServiceAsyncClient: The constructed client.
        """
        sa_info_func = (
            BookingServiceClient.from_service_account_info.__func__  # type: ignore
        )
        return sa_info_func(BookingServiceAsyncClient, info, *args, **kwargs)

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
            BookingServiceAsyncClient: The constructed client.
        """
        sa_file_func = (
            BookingServiceClient.from_service_account_file.__func__  # type: ignore
        )
        return sa_file_func(BookingServiceAsyncClient, filename, *args, **kwargs)

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
        return BookingServiceClient.get_mtls_endpoint_and_cert_source(client_options)  # type: ignore

    @property
    def transport(self) -> BookingServiceTransport:
        """Returns the transport used by the client instance.

        Returns:
            BookingServiceTransport: The transport used by the client instance.
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

    get_transport_class = BookingServiceClient.get_transport_class

    def __init__(self, *,
            credentials: Optional[ga_credentials.Credentials] = None,
            transport: Optional[Union[str, BookingServiceTransport, Callable[..., BookingServiceTransport]]] = "grpc_asyncio",
            client_options: Optional[ClientOptions] = None,
            client_info: gapic_v1.client_info.ClientInfo = DEFAULT_CLIENT_INFO,
            ) -> None:
        """Instantiates the booking service async client.

        Args:
            credentials (Optional[google.auth.credentials.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify the application to the service; if none
                are specified, the client will attempt to ascertain the
                credentials from the environment.
            transport (Optional[Union[str,BookingServiceTransport,Callable[..., BookingServiceTransport]]]):
                The transport to use, or a Callable that constructs and returns a new transport to use.
                If a Callable is given, it will be called with the same set of initialization
                arguments as used in the BookingServiceTransport constructor.
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
        self._client = BookingServiceClient(
            credentials=credentials,
            transport=transport,
            client_options=client_options,
            client_info=client_info,

        )

        if CLIENT_LOGGING_SUPPORTED and _LOGGER.isEnabledFor(std_logging.DEBUG):  # pragma: NO COVER
            _LOGGER.debug(
                "Created client `freebusy.booking_v1.BookingServiceAsyncClient`.",
                extra = {
                    "serviceName": "freebusy.booking.v1.BookingService",
                    "universeDomain": getattr(self._client._transport._credentials, "universe_domain", ""),
                    "credentialsType": f"{type(self._client._transport._credentials).__module__}.{type(self._client._transport._credentials).__qualname__}",
                    "credentialsInfo": getattr(self.transport._credentials, "get_cred_info", lambda: None)(),
                } if hasattr(self._client._transport, "_credentials") else {
                    "serviceName": "freebusy.booking.v1.BookingService",
                    "credentialsType": None,
                }
            )

    async def create_booking(self,
            request: Optional[Union[booking_messages.CreateBookingRequest, dict]] = None,
            *,
            booking: Optional[fb_booking.Booking] = None,
            booking_id: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> fb_booking.Booking:
        r"""Creates a booking, placing a hold.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import booking_v1

            async def sample_create_booking():
                # Create a client
                client = booking_v1.BookingServiceAsyncClient()

                # Initialize request argument(s)
                booking = booking_v1.Booking()
                booking.resource = "resource_value"

                request = booking_v1.CreateBookingRequest(
                    booking=booking,
                )

                # Make the request
                response = await client.create_booking(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.booking_v1.types.CreateBookingRequest, dict]]):
                The request object. Request message for CreateBooking. This places a hold
                transactionally; the request_id makes retries safe (the
                same id always yields the same booking instead of a
                duplicate hold). The promo code and hold TTL are set on
                the Booking resource itself.
            booking (:class:`freebusy.booking_v1.types.Booking`):
                The booking to create. Supply resource, window, and
                optionally offering, units, customer, contact, notes,
                attributes, promo_code, and hold_ttl. Provide contact
                when there is no customer (a guest booking). Output-only
                fields are ignored.

                This corresponds to the ``booking`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            booking_id (:class:`str`):
                Optional caller-chosen ID for the
                booking; the server generates one if
                unset.

                This corresponds to the ``booking_id`` field
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
            freebusy.booking_v1.types.Booking:
                A reservation against a resource. The hold lifecycle lives here as states
                   rather than a separate service: CreateBooking places
                   a PENDING_HOLD, confirmation flips it to CONFIRMED,
                   and an internal sweeper expires holds that are never
                   confirmed.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [booking, booking_id]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, booking_messages.CreateBookingRequest):
            request = booking_messages.CreateBookingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if booking is not None:
            request.booking = booking
        if booking_id is not None:
            request.booking_id = booking_id

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.create_booking]

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

    async def get_booking(self,
            request: Optional[Union[booking_messages.GetBookingRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> booking.Booking:
        r"""Gets a single booking.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import booking_v1

            async def sample_get_booking():
                # Create a client
                client = booking_v1.BookingServiceAsyncClient()

                # Initialize request argument(s)
                request = booking_v1.GetBookingRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_booking(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.booking_v1.types.GetBookingRequest, dict]]):
                The request object. Request message for GetBooking.
            name (:class:`str`):
                The booking to retrieve.
                Format: bookings/{booking}

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
            freebusy.booking_v1.types.Booking:
                A reservation against a resource. The hold lifecycle lives here as states
                   rather than a separate service: CreateBooking places
                   a PENDING_HOLD, confirmation flips it to CONFIRMED,
                   and an internal sweeper expires holds that are never
                   confirmed.

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
        if not isinstance(request, booking_messages.GetBookingRequest):
            request = booking_messages.GetBookingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_booking]

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

    async def list_bookings(self,
            request: Optional[Union[booking_messages.ListBookingsRequest, dict]] = None,
            *,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListBookingsAsyncPager:
        r"""Lists bookings.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import booking_v1

            async def sample_list_bookings():
                # Create a client
                client = booking_v1.BookingServiceAsyncClient()

                # Initialize request argument(s)
                request = booking_v1.ListBookingsRequest(
                )

                # Make the request
                page_result = client.list_bookings(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.booking_v1.types.ListBookingsRequest, dict]]):
                The request object. Request message for ListBookings.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.booking_v1.services.booking_service.pagers.ListBookingsAsyncPager:
                Response message for ListBookings.

                Iterating over this object will yield
                results and resolve additional pages
                automatically.

        """
        # Create or coerce a protobuf request object.
        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, booking_messages.ListBookingsRequest):
            request = booking_messages.ListBookingsRequest(request)

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_bookings]

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
        response = pagers.ListBookingsAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def confirm_booking(self,
            request: Optional[Union[booking_actions.ConfirmBookingRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> booking.Booking:
        r"""Confirms a held booking.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import booking_v1

            async def sample_confirm_booking():
                # Create a client
                client = booking_v1.BookingServiceAsyncClient()

                # Initialize request argument(s)
                request = booking_v1.ConfirmBookingRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.confirm_booking(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.booking_v1.types.ConfirmBookingRequest, dict]]):
                The request object. Request message for ConfirmBooking. Confirmation
                normally arrives via the payment webhook rather than a
                direct client call; either way it flips a PENDING_HOLD
                to CONFIRMED.
            name (:class:`str`):
                The booking to confirm.
                Format: bookings/{booking}

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
            freebusy.booking_v1.types.Booking:
                A reservation against a resource. The hold lifecycle lives here as states
                   rather than a separate service: CreateBooking places
                   a PENDING_HOLD, confirmation flips it to CONFIRMED,
                   and an internal sweeper expires holds that are never
                   confirmed.

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
        if not isinstance(request, booking_actions.ConfirmBookingRequest):
            request = booking_actions.ConfirmBookingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.confirm_booking]

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

    async def cancel_booking(self,
            request: Optional[Union[booking_actions.CancelBookingRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> booking.Booking:
        r"""Cancels a booking.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import booking_v1

            async def sample_cancel_booking():
                # Create a client
                client = booking_v1.BookingServiceAsyncClient()

                # Initialize request argument(s)
                request = booking_v1.CancelBookingRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.cancel_booking(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.booking_v1.types.CancelBookingRequest, dict]]):
                The request object. Request message for CancelBooking.
            name (:class:`str`):
                The booking to cancel.
                Format: bookings/{booking}

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
            freebusy.booking_v1.types.Booking:
                A reservation against a resource. The hold lifecycle lives here as states
                   rather than a separate service: CreateBooking places
                   a PENDING_HOLD, confirmation flips it to CONFIRMED,
                   and an internal sweeper expires holds that are never
                   confirmed.

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
        if not isinstance(request, booking_actions.CancelBookingRequest):
            request = booking_actions.CancelBookingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.cancel_booking]

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

    async def preview_cancellation(self,
            request: Optional[Union[booking_actions.PreviewCancellationRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> booking_actions.PreviewCancellationResponse:
        r"""Previews the refund a cancellation would yield now,
        without cancelling.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import booking_v1

            async def sample_preview_cancellation():
                # Create a client
                client = booking_v1.BookingServiceAsyncClient()

                # Initialize request argument(s)
                request = booking_v1.PreviewCancellationRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.preview_cancellation(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.booking_v1.types.PreviewCancellationRequest, dict]]):
                The request object. Request message for
                PreviewCancellation. Computes the refund
                a cancellation would yield right now,
                under the resource's cancellation
                policy, without cancelling the booking.
            name (:class:`str`):
                The booking to preview a cancellation
                for. Format: bookings/{booking}

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
            freebusy.booking_v1.types.PreviewCancellationResponse:
                Response message for
                PreviewCancellation.

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
        if not isinstance(request, booking_actions.PreviewCancellationRequest):
            request = booking_actions.PreviewCancellationRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.preview_cancellation]

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

    async def reschedule_booking(self,
            request: Optional[Union[booking_actions.RescheduleBookingRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            window: Optional[types_pb2.TimeWindow] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> booking.Booking:
        r"""Reschedules a booking to a new span.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import booking_v1

            async def sample_reschedule_booking():
                # Create a client
                client = booking_v1.BookingServiceAsyncClient()

                # Initialize request argument(s)
                request = booking_v1.RescheduleBookingRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.reschedule_booking(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.booking_v1.types.RescheduleBookingRequest, dict]]):
                The request object. Request message for
                RescheduleBooking. Atomically moves a
                booking to a new span (and optionally
                offering), re-running availability and
                the exclusion check on the new window.
            name (:class:`str`):
                The booking to reschedule.
                Format: bookings/{booking}

                This corresponds to the ``name`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            window (:class:`freebusy.shared.v1.types_pb2.TimeWindow`):
                The new reserved span.
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
            freebusy.booking_v1.types.Booking:
                A reservation against a resource. The hold lifecycle lives here as states
                   rather than a separate service: CreateBooking places
                   a PENDING_HOLD, confirmation flips it to CONFIRMED,
                   and an internal sweeper expires holds that are never
                   confirmed.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [name, window]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, booking_actions.RescheduleBookingRequest):
            request = booking_actions.RescheduleBookingRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name
        if window is not None:
            request.window = window

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.reschedule_booking]

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

    async def __aenter__(self) -> "BookingServiceAsyncClient":
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.transport.close()

DEFAULT_CLIENT_INFO = gapic_v1.client_info.ClientInfo(gapic_version=package_version.__version__)

if hasattr(DEFAULT_CLIENT_INFO, "protobuf_runtime_version"):   # pragma: NO COVER
    DEFAULT_CLIENT_INFO.protobuf_runtime_version = google.protobuf.__version__


__all__ = (
    "BookingServiceAsyncClient",
)
