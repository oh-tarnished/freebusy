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
import json
import logging as std_logging
import pickle
import warnings
from typing import Callable, Dict, Optional, Sequence, Tuple, Union

from google.api_core import grpc_helpers
from google.api_core import gapic_v1
import google.auth                         # type: ignore
from google.auth import credentials as ga_credentials  # type: ignore
from google.auth.transport.grpc import SslCredentials  # type: ignore
from google.protobuf.json_format import MessageToJson
import google.protobuf.message

import grpc  # type: ignore
import proto  # type: ignore

from freebusy.promocode_v1.types import promocode
from freebusy.promocode_v1.types import promocode_messages
import google.protobuf.empty_pb2 as empty_pb2  # type: ignore
from .base import PromoCodeServiceTransport, DEFAULT_CLIENT_INFO

try:
    from google.api_core import client_logging  # type: ignore
    CLIENT_LOGGING_SUPPORTED = True  # pragma: NO COVER
except ImportError:  # pragma: NO COVER
    CLIENT_LOGGING_SUPPORTED = False

_LOGGER = std_logging.getLogger(__name__)


class _LoggingClientInterceptor(grpc.UnaryUnaryClientInterceptor):  # pragma: NO COVER
    def intercept_unary_unary(self, continuation, client_call_details, request):
        logging_enabled = CLIENT_LOGGING_SUPPORTED and _LOGGER.isEnabledFor(std_logging.DEBUG)
        if logging_enabled:  # pragma: NO COVER
            request_metadata = client_call_details.metadata
            if isinstance(request, proto.Message):
                request_payload = type(request).to_json(request)
            elif isinstance(request, google.protobuf.message.Message):
                request_payload = MessageToJson(request)
            else:
                request_payload = f"{type(request).__name__}: {pickle.dumps(request)!r}"

            request_metadata = {
                key: value.decode("utf-8") if isinstance(value, bytes) else value
                for key, value in request_metadata
            }
            grpc_request = {
                "payload": request_payload,
                "requestMethod": "grpc",
                "metadata": dict(request_metadata),
            }
            _LOGGER.debug(
                f"Sending request for {client_call_details.method}",
                extra = {
                    "serviceName": "freebusy.promocode.v1.PromoCodeService",
                    "rpcName": str(client_call_details.method),
                    "request": grpc_request,
                    "metadata": grpc_request["metadata"],
                },
            )
        response = continuation(client_call_details, request)
        if logging_enabled:  # pragma: NO COVER
            response_metadata = response.trailing_metadata()
            # Convert gRPC metadata `<class 'grpc.aio._metadata.Metadata'>` to list of tuples
            metadata = dict([(k, str(v)) for k, v in response_metadata]) if response_metadata else None
            result = response.result()
            if isinstance(result, proto.Message):
                response_payload = type(result).to_json(result)
            elif isinstance(result, google.protobuf.message.Message):
                response_payload = MessageToJson(result)
            else:
                response_payload = f"{type(result).__name__}: {pickle.dumps(result)!r}"
            grpc_response = {
                "payload": response_payload,
                "metadata": metadata,
                "status": "OK",
            }
            _LOGGER.debug(
                f"Received response for {client_call_details.method}.",
                extra = {
                    "serviceName": "freebusy.promocode.v1.PromoCodeService",
                    "rpcName": client_call_details.method,
                    "response": grpc_response,
                    "metadata": grpc_response["metadata"],
                },
            )
        return response


class PromoCodeServiceGrpcTransport(PromoCodeServiceTransport):
    """gRPC backend transport for PromoCodeService.

    PromoCodeService manages redeemable discount codes and
    validates them against a prospective booking. Redemption itself
    happens inside CreateBooking.

    This class defines the same methods as the primary client, so the
    primary client can load the underlying transport implementation
    and call it.

    It sends protocol buffers over the wire using gRPC (which is built on
    top of HTTP/2); the ``grpcio`` package must be installed.
    """
    _stubs: Dict[str, Callable]

    def __init__(self, *,
            host: str = 'freebusy.ohtarnished.dev',
            credentials: Optional[ga_credentials.Credentials] = None,
            credentials_file: Optional[str] = None,
            scopes: Optional[Sequence[str]] = None,
            channel: Optional[Union[grpc.Channel, Callable[..., grpc.Channel]]] = None,
            api_mtls_endpoint: Optional[str] = None,
            client_cert_source: Optional[Callable[[], Tuple[bytes, bytes]]] = None,
            ssl_channel_credentials: Optional[grpc.ChannelCredentials] = None,
            client_cert_source_for_mtls: Optional[Callable[[], Tuple[bytes, bytes]]] = None,
            quota_project_id: Optional[str] = None,
            client_info: gapic_v1.client_info.ClientInfo = DEFAULT_CLIENT_INFO,
            always_use_jwt_access: Optional[bool] = False,
            api_audience: Optional[str] = None,
            ) -> None:
        """Instantiate the transport.

        Args:
            host (Optional[str]):
                 The hostname to connect to (default: 'freebusy.ohtarnished.dev').
            credentials (Optional[google.auth.credentials.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify the application to the service; if none
                are specified, the client will attempt to ascertain the
                credentials from the environment.
                This argument is ignored if a ``channel`` instance is provided.
            credentials_file (Optional[str]): Deprecated. A file with credentials that can
                be loaded with :func:`google.auth.load_credentials_from_file`.
                This argument is ignored if a ``channel`` instance is provided.
                This argument will be removed in the next major version of this library.
            scopes (Optional(Sequence[str])): A list of scopes. This argument is
                ignored if a ``channel`` instance is provided.
            channel (Optional[Union[grpc.Channel, Callable[..., grpc.Channel]]]):
                A ``Channel`` instance through which to make calls, or a Callable
                that constructs and returns one. If set to None, ``self.create_channel``
                is used to create the channel. If a Callable is given, it will be called
                with the same arguments as used in ``self.create_channel``.
            api_mtls_endpoint (Optional[str]): Deprecated. The mutual TLS endpoint.
                If provided, it overrides the ``host`` argument and tries to create
                a mutual TLS channel with client SSL credentials from
                ``client_cert_source`` or application default SSL credentials.
            client_cert_source (Optional[Callable[[], Tuple[bytes, bytes]]]):
                Deprecated. A callback to provide client SSL certificate bytes and
                private key bytes, both in PEM format. It is ignored if
                ``api_mtls_endpoint`` is None.
            ssl_channel_credentials (grpc.ChannelCredentials): SSL credentials
                for the grpc channel. It is ignored if a ``channel`` instance is provided.
            client_cert_source_for_mtls (Optional[Callable[[], Tuple[bytes, bytes]]]):
                A callback to provide client certificate bytes and private key bytes,
                both in PEM format. It is used to configure a mutual TLS channel. It is
                ignored if a ``channel`` instance or ``ssl_channel_credentials`` is provided.
            quota_project_id (Optional[str]): An optional project to use for billing
                and quota.
            client_info (google.api_core.gapic_v1.client_info.ClientInfo):
                The client info used to send a user-agent string along with
                API requests. If ``None``, then default info will be used.
                Generally, you only need to set this if you're developing
                your own client library.
            always_use_jwt_access (Optional[bool]): Whether self signed JWT should
                be used for service account credentials.
            api_audience (Optional[str]): The intended audience for the API calls
                to the service that will be set when using certain 3rd party
                authentication flows. Audience is typically a resource identifier.
                If not set, the host value will be used as a default.

        Raises:
          google.auth.exceptions.MutualTLSChannelError: If mutual TLS transport
              creation failed for any reason.
          google.api_core.exceptions.DuplicateCredentialArgs: If both ``credentials``
              and ``credentials_file`` are passed.
        """
        self._grpc_channel = None
        self._ssl_channel_credentials = ssl_channel_credentials
        self._stubs: Dict[str, Callable] = {}

        if api_mtls_endpoint:
            warnings.warn("api_mtls_endpoint is deprecated", DeprecationWarning)
        if client_cert_source:
            warnings.warn("client_cert_source is deprecated", DeprecationWarning)

        if isinstance(channel, grpc.Channel):
            # Ignore credentials if a channel was passed.
            credentials = None
            self._ignore_credentials = True
            # If a channel was explicitly provided, set it.
            self._grpc_channel = channel
            self._ssl_channel_credentials = None

        else:
            if api_mtls_endpoint:
                host = api_mtls_endpoint

                # Create SSL credentials with client_cert_source or application
                # default SSL credentials.
                if client_cert_source:
                    cert, key = client_cert_source()
                    self._ssl_channel_credentials = grpc.ssl_channel_credentials(
                        certificate_chain=cert, private_key=key
                    )
                else:
                    self._ssl_channel_credentials = SslCredentials().ssl_credentials

            else:
                if client_cert_source_for_mtls and not ssl_channel_credentials:
                    cert, key = client_cert_source_for_mtls()
                    self._ssl_channel_credentials = grpc.ssl_channel_credentials(
                        certificate_chain=cert, private_key=key
                    )

        # The base transport sets the host, credentials and scopes
        super().__init__(
            host=host,
            credentials=credentials,
            credentials_file=credentials_file,
            scopes=scopes,
            quota_project_id=quota_project_id,
            client_info=client_info,
            always_use_jwt_access=always_use_jwt_access,
            api_audience=api_audience,
        )

        if not self._grpc_channel:
            # initialize with the provided callable or the default channel
            channel_init = channel or type(self).create_channel
            self._grpc_channel = channel_init(
                self._host,
                # use the credentials which are saved
                credentials=self._credentials,
                # Set ``credentials_file`` to ``None`` here as
                # the credentials that we saved earlier should be used.
                credentials_file=None,
                scopes=self._scopes,
                ssl_credentials=self._ssl_channel_credentials,
                quota_project_id=quota_project_id,
                options=[
                    ("grpc.max_send_message_length", -1),
                    ("grpc.max_receive_message_length", -1),
                ],
            )

        self._interceptor = _LoggingClientInterceptor()
        self._logged_channel =  grpc.intercept_channel(self._grpc_channel, self._interceptor)

        # Wrap messages. This must be done after self._logged_channel exists
        self._prep_wrapped_messages(client_info)

    @classmethod
    def create_channel(cls,
                       host: str = 'freebusy.ohtarnished.dev',
                       credentials: Optional[ga_credentials.Credentials] = None,
                       credentials_file: Optional[str] = None,
                       scopes: Optional[Sequence[str]] = None,
                       quota_project_id: Optional[str] = None,
                       **kwargs) -> grpc.Channel:
        """Create and return a gRPC channel object.
        Args:
            host (Optional[str]): The host for the channel to use.
            credentials (Optional[~.Credentials]): The
                authorization credentials to attach to requests. These
                credentials identify this application to the service. If
                none are specified, the client will attempt to ascertain
                the credentials from the environment.
            credentials_file (Optional[str]): Deprecated. A file with credentials that can
                be loaded with :func:`google.auth.load_credentials_from_file`.
                This argument is mutually exclusive with credentials.  This argument will be
                removed in the next major version of this library.
            scopes (Optional[Sequence[str]]): A optional list of scopes needed for this
                service. These are only used when credentials are not specified and
                are passed to :func:`google.auth.default`.
            quota_project_id (Optional[str]): An optional project to use for billing
                and quota.
            kwargs (Optional[dict]): Keyword arguments, which are passed to the
                channel creation.
        Returns:
            grpc.Channel: A gRPC channel object.

        Raises:
            google.api_core.exceptions.DuplicateCredentialArgs: If both ``credentials``
              and ``credentials_file`` are passed.
        """

        return grpc_helpers.create_channel(
            host,
            credentials=credentials,
            credentials_file=credentials_file,
            quota_project_id=quota_project_id,
            default_scopes=cls.AUTH_SCOPES,
            scopes=scopes,
            default_host=cls.DEFAULT_HOST,
            **kwargs
        )

    @property
    def grpc_channel(self) -> grpc.Channel:
        """Return the channel designed to connect to this service.
        """
        return self._grpc_channel

    @property
    def list_promo_codes(self) -> Callable[
            [promocode_messages.ListPromoCodesRequest],
            promocode_messages.ListPromoCodesResponse]:
        r"""Return a callable for the list promo codes method over gRPC.

        Lists promo codes.

        Returns:
            Callable[[~.ListPromoCodesRequest],
                    ~.ListPromoCodesResponse]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'list_promo_codes' not in self._stubs:
            self._stubs['list_promo_codes'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/ListPromoCodes',
                request_serializer=promocode_messages.ListPromoCodesRequest.serialize,
                response_deserializer=promocode_messages.ListPromoCodesResponse.deserialize,
            )
        return self._stubs['list_promo_codes']

    @property
    def get_promo_code(self) -> Callable[
            [promocode_messages.GetPromoCodeRequest],
            promocode.PromoCode]:
        r"""Return a callable for the get promo code method over gRPC.

        Gets a single promo code.

        Returns:
            Callable[[~.GetPromoCodeRequest],
                    ~.PromoCode]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'get_promo_code' not in self._stubs:
            self._stubs['get_promo_code'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/GetPromoCode',
                request_serializer=promocode_messages.GetPromoCodeRequest.serialize,
                response_deserializer=promocode.PromoCode.deserialize,
            )
        return self._stubs['get_promo_code']

    @property
    def create_promo_code(self) -> Callable[
            [promocode_messages.CreatePromoCodeRequest],
            promocode.PromoCode]:
        r"""Return a callable for the create promo code method over gRPC.

        Creates a promo code.

        Returns:
            Callable[[~.CreatePromoCodeRequest],
                    ~.PromoCode]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'create_promo_code' not in self._stubs:
            self._stubs['create_promo_code'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/CreatePromoCode',
                request_serializer=promocode_messages.CreatePromoCodeRequest.serialize,
                response_deserializer=promocode.PromoCode.deserialize,
            )
        return self._stubs['create_promo_code']

    @property
    def update_promo_code(self) -> Callable[
            [promocode_messages.UpdatePromoCodeRequest],
            promocode.PromoCode]:
        r"""Return a callable for the update promo code method over gRPC.

        Updates a promo code.

        Returns:
            Callable[[~.UpdatePromoCodeRequest],
                    ~.PromoCode]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'update_promo_code' not in self._stubs:
            self._stubs['update_promo_code'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/UpdatePromoCode',
                request_serializer=promocode_messages.UpdatePromoCodeRequest.serialize,
                response_deserializer=promocode.PromoCode.deserialize,
            )
        return self._stubs['update_promo_code']

    @property
    def delete_promo_code(self) -> Callable[
            [promocode_messages.DeletePromoCodeRequest],
            empty_pb2.Empty]:
        r"""Return a callable for the delete promo code method over gRPC.

        Deletes a promo code.

        Returns:
            Callable[[~.DeletePromoCodeRequest],
                    ~.Empty]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'delete_promo_code' not in self._stubs:
            self._stubs['delete_promo_code'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/DeletePromoCode',
                request_serializer=promocode_messages.DeletePromoCodeRequest.serialize,
                response_deserializer=empty_pb2.Empty.FromString,
            )
        return self._stubs['delete_promo_code']

    @property
    def validate_promo_code(self) -> Callable[
            [promocode_messages.ValidatePromoCodeRequest],
            promocode_messages.ValidatePromoCodeResponse]:
        r"""Return a callable for the validate promo code method over gRPC.

        Validates a code against a prospective booking and
        returns the discount.

        Returns:
            Callable[[~.ValidatePromoCodeRequest],
                    ~.ValidatePromoCodeResponse]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'validate_promo_code' not in self._stubs:
            self._stubs['validate_promo_code'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/ValidatePromoCode',
                request_serializer=promocode_messages.ValidatePromoCodeRequest.serialize,
                response_deserializer=promocode_messages.ValidatePromoCodeResponse.deserialize,
            )
        return self._stubs['validate_promo_code']

    @property
    def list_redemptions(self) -> Callable[
            [promocode_messages.ListRedemptionsRequest],
            promocode_messages.ListRedemptionsResponse]:
        r"""Return a callable for the list redemptions method over gRPC.

        Lists redemptions of a promo code (paged).
        Redemptions are created during CreateBooking, so this
        service exposes read-only access to them.

        Returns:
            Callable[[~.ListRedemptionsRequest],
                    ~.ListRedemptionsResponse]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'list_redemptions' not in self._stubs:
            self._stubs['list_redemptions'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/ListRedemptions',
                request_serializer=promocode_messages.ListRedemptionsRequest.serialize,
                response_deserializer=promocode_messages.ListRedemptionsResponse.deserialize,
            )
        return self._stubs['list_redemptions']

    @property
    def get_redemption(self) -> Callable[
            [promocode_messages.GetRedemptionRequest],
            promocode.Redemption]:
        r"""Return a callable for the get redemption method over gRPC.

        Gets a single redemption by resource name.

        Returns:
            Callable[[~.GetRedemptionRequest],
                    ~.Redemption]:
                A function that, when called, will call the underlying RPC
                on the server.
        """
        # Generate a "stub function" on-the-fly which will actually make
        # the request.
        # gRPC handles serialization and deserialization, so we just need
        # to pass in the functions for each.
        if 'get_redemption' not in self._stubs:
            self._stubs['get_redemption'] = self._logged_channel.unary_unary(
                '/freebusy.promocode.v1.PromoCodeService/GetRedemption',
                request_serializer=promocode_messages.GetRedemptionRequest.serialize,
                response_deserializer=promocode.Redemption.deserialize,
            )
        return self._stubs['get_redemption']

    def close(self):
        self._logged_channel.close()

    @property
    def kind(self) -> str:
        return "grpc"


__all__ = (
    'PromoCodeServiceGrpcTransport',
)
