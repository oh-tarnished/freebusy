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

from freebusy.promocode_v1 import gapic_version as package_version

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

from freebusy.promocode_v1.services.promo_code_service import pagers
from freebusy.promocode_v1.types import enums
from freebusy.promocode_v1.types import promocode
from freebusy.promocode_v1.types import promocode_messages
import google.protobuf.field_mask_pb2 as field_mask_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore
from .transports.base import PromoCodeServiceTransport, DEFAULT_CLIENT_INFO
from .transports.grpc_asyncio import PromoCodeServiceGrpcAsyncIOTransport
from .client import PromoCodeServiceClient

try:
    from google.api_core import client_logging  # type: ignore
    CLIENT_LOGGING_SUPPORTED = True  # pragma: NO COVER
except ImportError:  # pragma: NO COVER
    CLIENT_LOGGING_SUPPORTED = False

_LOGGER = std_logging.getLogger(__name__)

class PromoCodeServiceAsyncClient:
    """PromoCodeService manages redeemable discount codes and
    validates them against a prospective booking. Redemption itself
    happens inside CreateBooking.
    """

    _client: PromoCodeServiceClient

    # Copy defaults from the synchronous client for use here.
    # Note: DEFAULT_ENDPOINT is deprecated. Use _DEFAULT_ENDPOINT_TEMPLATE instead.
    DEFAULT_ENDPOINT = PromoCodeServiceClient.DEFAULT_ENDPOINT
    DEFAULT_MTLS_ENDPOINT = PromoCodeServiceClient.DEFAULT_MTLS_ENDPOINT
    _DEFAULT_ENDPOINT_TEMPLATE = PromoCodeServiceClient._DEFAULT_ENDPOINT_TEMPLATE
    _DEFAULT_UNIVERSE = PromoCodeServiceClient._DEFAULT_UNIVERSE

    promo_code_path = staticmethod(PromoCodeServiceClient.promo_code_path)
    parse_promo_code_path = staticmethod(PromoCodeServiceClient.parse_promo_code_path)
    redemption_path = staticmethod(PromoCodeServiceClient.redemption_path)
    parse_redemption_path = staticmethod(PromoCodeServiceClient.parse_redemption_path)
    common_billing_account_path = staticmethod(PromoCodeServiceClient.common_billing_account_path)
    parse_common_billing_account_path = staticmethod(PromoCodeServiceClient.parse_common_billing_account_path)
    common_folder_path = staticmethod(PromoCodeServiceClient.common_folder_path)
    parse_common_folder_path = staticmethod(PromoCodeServiceClient.parse_common_folder_path)
    common_organization_path = staticmethod(PromoCodeServiceClient.common_organization_path)
    parse_common_organization_path = staticmethod(PromoCodeServiceClient.parse_common_organization_path)
    common_project_path = staticmethod(PromoCodeServiceClient.common_project_path)
    parse_common_project_path = staticmethod(PromoCodeServiceClient.parse_common_project_path)
    common_location_path = staticmethod(PromoCodeServiceClient.common_location_path)
    parse_common_location_path = staticmethod(PromoCodeServiceClient.parse_common_location_path)

    @classmethod
    def from_service_account_info(cls, info: dict, *args, **kwargs):
        """Creates an instance of this client using the provided credentials
            info.

        Args:
            info (dict): The service account private key info.
            args: Additional arguments to pass to the constructor.
            kwargs: Additional arguments to pass to the constructor.

        Returns:
            PromoCodeServiceAsyncClient: The constructed client.
        """
        sa_info_func = (
            PromoCodeServiceClient.from_service_account_info.__func__  # type: ignore
        )
        return sa_info_func(PromoCodeServiceAsyncClient, info, *args, **kwargs)

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
            PromoCodeServiceAsyncClient: The constructed client.
        """
        sa_file_func = (
            PromoCodeServiceClient.from_service_account_file.__func__  # type: ignore
        )
        return sa_file_func(PromoCodeServiceAsyncClient, filename, *args, **kwargs)

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
        return PromoCodeServiceClient.get_mtls_endpoint_and_cert_source(client_options)  # type: ignore

    @property
    def transport(self) -> PromoCodeServiceTransport:
        """Returns the transport used by the client instance.

        Returns:
            PromoCodeServiceTransport: The transport used by the client instance.
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

    get_transport_class = PromoCodeServiceClient.get_transport_class

    def __init__(self, *,
            credentials: Optional[ga_credentials.Credentials] = None,
            transport: Optional[Union[str, PromoCodeServiceTransport, Callable[..., PromoCodeServiceTransport]]] = "grpc_asyncio",
            client_options: Optional[ClientOptions] = None,
            client_info: gapic_v1.client_info.ClientInfo = DEFAULT_CLIENT_INFO,
            ) -> None:
        """Instantiates the promo code service async client.

        Args:
            credentials (Optional[google.auth.credentials.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify the application to the service; if none
                are specified, the client will attempt to ascertain the
                credentials from the environment.
            transport (Optional[Union[str,PromoCodeServiceTransport,Callable[..., PromoCodeServiceTransport]]]):
                The transport to use, or a Callable that constructs and returns a new transport to use.
                If a Callable is given, it will be called with the same set of initialization
                arguments as used in the PromoCodeServiceTransport constructor.
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
        self._client = PromoCodeServiceClient(
            credentials=credentials,
            transport=transport,
            client_options=client_options,
            client_info=client_info,

        )

        if CLIENT_LOGGING_SUPPORTED and _LOGGER.isEnabledFor(std_logging.DEBUG):  # pragma: NO COVER
            _LOGGER.debug(
                "Created client `freebusy.promocode_v1.PromoCodeServiceAsyncClient`.",
                extra = {
                    "serviceName": "freebusy.promocode.v1.PromoCodeService",
                    "universeDomain": getattr(self._client._transport._credentials, "universe_domain", ""),
                    "credentialsType": f"{type(self._client._transport._credentials).__module__}.{type(self._client._transport._credentials).__qualname__}",
                    "credentialsInfo": getattr(self.transport._credentials, "get_cred_info", lambda: None)(),
                } if hasattr(self._client._transport, "_credentials") else {
                    "serviceName": "freebusy.promocode.v1.PromoCodeService",
                    "credentialsType": None,
                }
            )

    async def list_promo_codes(self,
            request: Optional[Union[promocode_messages.ListPromoCodesRequest, dict]] = None,
            *,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListPromoCodesAsyncPager:
        r"""Lists promo codes.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_list_promo_codes():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                request = promocode_v1.ListPromoCodesRequest(
                )

                # Make the request
                page_result = client.list_promo_codes(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.ListPromoCodesRequest, dict]]):
                The request object. Request message for ListPromoCodes.
            retry (google.api_core.retry_async.AsyncRetry): Designation of what errors, if any,
                should be retried.
            timeout (float): The timeout for this request.
            metadata (Sequence[Tuple[str, Union[str, bytes]]]): Key/value pairs which should be
                sent along with the request as metadata. Normally, each value must be of type `str`,
                but for metadata keys ending with the suffix `-bin`, the corresponding values must
                be of type `bytes`.

        Returns:
            freebusy.promocode_v1.services.promo_code_service.pagers.ListPromoCodesAsyncPager:
                Response message for ListPromoCodes.

                Iterating over this object will yield
                results and resolve additional pages
                automatically.

        """
        # Create or coerce a protobuf request object.
        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, promocode_messages.ListPromoCodesRequest):
            request = promocode_messages.ListPromoCodesRequest(request)

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_promo_codes]

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
        response = pagers.ListPromoCodesAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def get_promo_code(self,
            request: Optional[Union[promocode_messages.GetPromoCodeRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> promocode.PromoCode:
        r"""Gets a single promo code.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_get_promo_code():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                request = promocode_v1.GetPromoCodeRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_promo_code(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.GetPromoCodeRequest, dict]]):
                The request object. Request message for GetPromoCode.
            name (:class:`str`):
                The promo code to retrieve. Format:
                promoCodes/{promo_code}

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
            freebusy.promocode_v1.types.PromoCode:
                A redeemable discount applied to a
                booking's subtotal. Scoped by a
                redemption window, usage caps, a minimum
                subtotal, and an optional set of
                resources / offerings it applies to.

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
        if not isinstance(request, promocode_messages.GetPromoCodeRequest):
            request = promocode_messages.GetPromoCodeRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_promo_code]

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

    async def create_promo_code(self,
            request: Optional[Union[promocode_messages.CreatePromoCodeRequest, dict]] = None,
            *,
            promo_code: Optional[promocode.PromoCode] = None,
            promo_code_id: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> promocode.PromoCode:
        r"""Creates a promo code.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_create_promo_code():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                promo_code = promocode_v1.PromoCode()
                promo_code.code = "code_value"
                promo_code.discount.percent_off = 1163

                request = promocode_v1.CreatePromoCodeRequest(
                    promo_code=promo_code,
                )

                # Make the request
                response = await client.create_promo_code(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.CreatePromoCodeRequest, dict]]):
                The request object. Request message for CreatePromoCode.
            promo_code (:class:`freebusy.promocode_v1.types.PromoCode`):
                The promo code to create. Server-assigned fields are
                ignored on input: name, state, redemption_count,
                redemptions, create_time, update_time, etag.
                promo_code.code is honored only when code_generation is
                MANUAL (or unset).

                This corresponds to the ``promo_code`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            promo_code_id (:class:`str`):
                Optional caller-chosen ID for the
                promo code; the server generates one if
                unset.

                This corresponds to the ``promo_code_id`` field
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
            freebusy.promocode_v1.types.PromoCode:
                A redeemable discount applied to a
                booking's subtotal. Scoped by a
                redemption window, usage caps, a minimum
                subtotal, and an optional set of
                resources / offerings it applies to.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [promo_code, promo_code_id]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, promocode_messages.CreatePromoCodeRequest):
            request = promocode_messages.CreatePromoCodeRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if promo_code is not None:
            request.promo_code = promo_code
        if promo_code_id is not None:
            request.promo_code_id = promo_code_id

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.create_promo_code]

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

    async def update_promo_code(self,
            request: Optional[Union[promocode_messages.UpdatePromoCodeRequest, dict]] = None,
            *,
            promo_code: Optional[promocode.PromoCode] = None,
            update_mask: Optional[field_mask_pb2.FieldMask] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> promocode.PromoCode:
        r"""Updates a promo code.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_update_promo_code():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                promo_code = promocode_v1.PromoCode()
                promo_code.code = "code_value"
                promo_code.discount.percent_off = 1163

                request = promocode_v1.UpdatePromoCodeRequest(
                    promo_code=promo_code,
                )

                # Make the request
                response = await client.update_promo_code(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.UpdatePromoCodeRequest, dict]]):
                The request object. Request message for UpdatePromoCode.
            promo_code (:class:`freebusy.promocode_v1.types.PromoCode`):
                The promo code to update; its name
                identifies the target. For optimistic
                concurrency, echo the etag you last read
                (AIP-154); the update fails if it no
                longer matches.

                This corresponds to the ``promo_code`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            update_mask (:class:`google.protobuf.field_mask_pb2.FieldMask`):
                Fields to overwrite. Omit to replace all mutable fields.
                Nested fields use dotted paths, e.g.
                "discount.amount_off", "window.end_time",
                "scope.min_subtotal".

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
            freebusy.promocode_v1.types.PromoCode:
                A redeemable discount applied to a
                booking's subtotal. Scoped by a
                redemption window, usage caps, a minimum
                subtotal, and an optional set of
                resources / offerings it applies to.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [promo_code, update_mask]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, promocode_messages.UpdatePromoCodeRequest):
            request = promocode_messages.UpdatePromoCodeRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if promo_code is not None:
            request.promo_code = promo_code
        if update_mask is not None:
            request.update_mask = update_mask

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.update_promo_code]

        # Certain fields should be provided within the metadata header;
        # add these here.
        metadata = tuple(metadata) + (
            gapic_v1.routing_header.to_grpc_metadata((
                ("promo_code.name", request.promo_code.name),
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

    async def delete_promo_code(self,
            request: Optional[Union[promocode_messages.DeletePromoCodeRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> None:
        r"""Deletes a promo code.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_delete_promo_code():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                request = promocode_v1.DeletePromoCodeRequest(
                    name="name_value",
                )

                # Make the request
                await client.delete_promo_code(request=request)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.DeletePromoCodeRequest, dict]]):
                The request object. Request message for DeletePromoCode.
            name (:class:`str`):
                The promo code to delete. Format:
                promoCodes/{promo_code}

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
        if not isinstance(request, promocode_messages.DeletePromoCodeRequest):
            request = promocode_messages.DeletePromoCodeRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.delete_promo_code]

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

    async def validate_promo_code(self,
            request: Optional[Union[promocode_messages.ValidatePromoCodeRequest, dict]] = None,
            *,
            code: Optional[str] = None,
            subtotal: Optional[money_pb2.Money] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> promocode_messages.ValidatePromoCodeResponse:
        r"""Validates a code against a prospective booking and
        returns the discount.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_validate_promo_code():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                request = promocode_v1.ValidatePromoCodeRequest(
                    code="code_value",
                )

                # Make the request
                response = await client.validate_promo_code(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.ValidatePromoCodeRequest, dict]]):
                The request object. Request message for
                ValidatePromoCode. Computes the discount
                a code would apply to a prospective
                booking without redeeming it.
            code (:class:`str`):
                The human-entered code to validate
                (e.g. "SUMMER25").

                This corresponds to the ``code`` field
                on the ``request`` instance; if ``request`` is provided, this
                should not be set.
            subtotal (:class:`google.type.money_pb2.Money`):
                Subtotal the discount would apply to.
                This corresponds to the ``subtotal`` field
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
            freebusy.promocode_v1.types.ValidatePromoCodeResponse:
                Response message for
                ValidatePromoCode.

        """
        # Create or coerce a protobuf request object.
        # - Quick check: If we got a request object, we should *not* have
        #   gotten any keyword arguments that map to the request.
        flattened_params = [code, subtotal]
        has_flattened_params = len([param for param in flattened_params if param is not None]) > 0
        if request is not None and has_flattened_params:
            raise ValueError("If the `request` argument is set, then none of "
                             "the individual field arguments should be set.")

        # - Use the request object if provided (there's no risk of modifying the input as
        #   there are no flattened fields), or create one.
        if not isinstance(request, promocode_messages.ValidatePromoCodeRequest):
            request = promocode_messages.ValidatePromoCodeRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if code is not None:
            request.code = code
        if subtotal is not None:
            request.subtotal = subtotal

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.validate_promo_code]

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

    async def list_redemptions(self,
            request: Optional[Union[promocode_messages.ListRedemptionsRequest, dict]] = None,
            *,
            parent: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> pagers.ListRedemptionsAsyncPager:
        r"""Lists redemptions of a promo code (paged).
        Redemptions are created during CreateBooking, so this
        service exposes read-only access to them.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_list_redemptions():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                request = promocode_v1.ListRedemptionsRequest(
                    parent="parent_value",
                )

                # Make the request
                page_result = client.list_redemptions(request=request)

                # Handle the response
                async for response in page_result:
                    print(response)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.ListRedemptionsRequest, dict]]):
                The request object. Request message for ListRedemptions.
            parent (:class:`str`):
                The promo code whose redemptions to list. Format:
                promoCodes/{promo_code}

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
            freebusy.promocode_v1.services.promo_code_service.pagers.ListRedemptionsAsyncPager:
                Response message for ListRedemptions.

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
        if not isinstance(request, promocode_messages.ListRedemptionsRequest):
            request = promocode_messages.ListRedemptionsRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if parent is not None:
            request.parent = parent

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.list_redemptions]

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
        response = pagers.ListRedemptionsAsyncPager(
            method=rpc,
            request=request,
            response=response,
            retry=retry,
            timeout=timeout,
            metadata=metadata,
        )

        # Done; return the response.
        return response

    async def get_redemption(self,
            request: Optional[Union[promocode_messages.GetRedemptionRequest, dict]] = None,
            *,
            name: Optional[str] = None,
            retry: OptionalRetry = gapic_v1.method.DEFAULT,
            timeout: Union[float, object] = gapic_v1.method.DEFAULT,
            metadata: Sequence[Tuple[str, Union[str, bytes]]] = (),
            ) -> promocode.Redemption:
        r"""Gets a single redemption by resource name.

        .. code-block:: python

            # This snippet has been automatically generated and should be regarded as a
            # code template only.
            # It will require modifications to work:
            # - It may require correct/in-range values for request initialization.
            # - It may require specifying regional endpoints when creating the service
            #   client as shown in:
            #   https://googleapis.dev/python/google-api-core/latest/client_options.html
            from freebusy import promocode_v1

            async def sample_get_redemption():
                # Create a client
                client = promocode_v1.PromoCodeServiceAsyncClient()

                # Initialize request argument(s)
                request = promocode_v1.GetRedemptionRequest(
                    name="name_value",
                )

                # Make the request
                response = await client.get_redemption(request=request)

                # Handle the response
                print(response)

        Args:
            request (Optional[Union[freebusy.promocode_v1.types.GetRedemptionRequest, dict]]):
                The request object. Request message for GetRedemption.
            name (:class:`str`):
                The redemption to retrieve. Format:
                promoCodes/{promo_code}/redemptions/{redemption}

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
            freebusy.promocode_v1.types.Redemption:
                Redemption is a single use of a promo code, modeled as a sub-resource of
                   PromoCode rather than an inline list — so it has its
                   own name/lifecycle and is listed with paging
                   (ListRedemptions). The {promo_code} parent segment
                   generates the promo_code_id FK back to the owning
                   code (1:n into promocode.redemptions); amount_applied
                   is the shared google.type.Money in common.moneys.
                   Redemptions are created during CreateBooking, never
                   directly.

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
        if not isinstance(request, promocode_messages.GetRedemptionRequest):
            request = promocode_messages.GetRedemptionRequest(request)

        # If we have keyword arguments corresponding to fields on the
        # request, apply these.
        if name is not None:
            request.name = name

        # Wrap the RPC method; this adds retry and timeout information,
        # and friendly error handling.
        rpc = self._client._transport._wrapped_methods[self._client._transport.get_redemption]

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

    async def __aenter__(self) -> "PromoCodeServiceAsyncClient":
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.transport.close()

DEFAULT_CLIENT_INFO = gapic_v1.client_info.ClientInfo(gapic_version=package_version.__version__)

if hasattr(DEFAULT_CLIENT_INFO, "protobuf_runtime_version"):   # pragma: NO COVER
    DEFAULT_CLIENT_INFO.protobuf_runtime_version = google.protobuf.__version__


__all__ = (
    "PromoCodeServiceAsyncClient",
)
