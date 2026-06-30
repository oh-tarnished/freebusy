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
import os
import asyncio
from unittest import mock
from unittest.mock import AsyncMock

import grpc
from grpc.experimental import aio
import json
import math
import pytest
from collections.abc import Sequence, Mapping
from google.api_core import api_core_version
from proto.marshal.rules.dates import DurationRule, TimestampRule
from proto.marshal.rules import wrappers

try:
    from google.auth.aio import credentials as ga_credentials_async
    HAS_GOOGLE_AUTH_AIO = True
except ImportError: # pragma: NO COVER
    HAS_GOOGLE_AUTH_AIO = False

from freebusy.booking_v1.services.booking_service import BookingServiceAsyncClient
from freebusy.booking_v1.services.booking_service import BookingServiceClient
from freebusy.booking_v1.services.booking_service import pagers
from freebusy.booking_v1.services.booking_service import transports
from freebusy.booking_v1.types import booking
from freebusy.booking_v1.types import booking as fb_booking
from freebusy.booking_v1.types import booking_actions
from freebusy.booking_v1.types import booking_messages
from freebusy.booking_v1.types import enums
from google.api_core import client_options
from google.api_core import exceptions as core_exceptions
from google.api_core import gapic_v1
from google.api_core import grpc_helpers
from google.api_core import grpc_helpers_async
from google.api_core import path_template
from google.api_core import retry as retries
from google.auth import credentials as ga_credentials
from google.auth.exceptions import MutualTLSChannelError
from google.oauth2 import service_account
import freebusy.shared.v1.types_pb2 as types_pb2  # type: ignore
import google.auth
import google.protobuf.duration_pb2 as duration_pb2  # type: ignore
import google.protobuf.struct_pb2 as struct_pb2  # type: ignore
import google.protobuf.timestamp_pb2 as timestamp_pb2  # type: ignore
import google.type.money_pb2 as money_pb2  # type: ignore



CRED_INFO_JSON = {
    "credential_source": "/path/to/file",
    "credential_type": "service account credentials",
    "principal": "service-account@example.com",
}
CRED_INFO_STRING = json.dumps(CRED_INFO_JSON)


async def mock_async_gen(data, chunk_size=1):
    for i in range(0, len(data)):  # pragma: NO COVER
        chunk = data[i : i + chunk_size]
        yield chunk.encode("utf-8")

def client_cert_source_callback():
    return b"cert bytes", b"key bytes"

# TODO: use async auth anon credentials by default once the minimum version of google-auth is upgraded.
# See related issue: https://github.com/googleapis/gapic-generator-python/issues/2107.
def async_anonymous_credentials():
    if HAS_GOOGLE_AUTH_AIO:
        return ga_credentials_async.AnonymousCredentials()
    return ga_credentials.AnonymousCredentials()

# If default endpoint is localhost, then default mtls endpoint will be the same.
# This method modifies the default endpoint so the client can produce a different
# mtls endpoint for endpoint testing purposes.
def modify_default_endpoint(client):
    return "foo.googleapis.com" if ("localhost" in client.DEFAULT_ENDPOINT) else client.DEFAULT_ENDPOINT

# If default endpoint template is localhost, then default mtls endpoint will be the same.
# This method modifies the default endpoint template so the client can produce a different
# mtls endpoint for endpoint testing purposes.
def modify_default_endpoint_template(client):
    return "test.{UNIVERSE_DOMAIN}" if ("localhost" in client._DEFAULT_ENDPOINT_TEMPLATE) else client._DEFAULT_ENDPOINT_TEMPLATE


@pytest.fixture(autouse=True)
def set_event_loop():
    try:
        asyncio.get_running_loop()
        yield
    except RuntimeError:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            yield
        finally:
            loop.close()
            asyncio.set_event_loop(None)


def test__get_default_mtls_endpoint():
    api_endpoint = "example.googleapis.com"
    api_mtls_endpoint = "example.mtls.googleapis.com"
    sandbox_endpoint = "example.sandbox.googleapis.com"
    sandbox_mtls_endpoint = "example.mtls.sandbox.googleapis.com"
    non_googleapi = "api.example.com"
    custom_endpoint = ".custom"

    assert BookingServiceClient._get_default_mtls_endpoint(None) is None
    assert BookingServiceClient._get_default_mtls_endpoint(api_endpoint) == api_mtls_endpoint
    assert BookingServiceClient._get_default_mtls_endpoint(api_mtls_endpoint) == api_mtls_endpoint
    assert BookingServiceClient._get_default_mtls_endpoint(sandbox_endpoint) == sandbox_mtls_endpoint
    assert BookingServiceClient._get_default_mtls_endpoint(sandbox_mtls_endpoint) == sandbox_mtls_endpoint
    assert BookingServiceClient._get_default_mtls_endpoint(non_googleapi) == non_googleapi
    assert BookingServiceClient._get_default_mtls_endpoint(custom_endpoint) == custom_endpoint

def test__read_environment_variables():
    assert BookingServiceClient._read_environment_variables() == (False, "auto", None)

    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "true"}):
        assert BookingServiceClient._read_environment_variables() == (True, "auto", None)

    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "false"}):
        assert BookingServiceClient._read_environment_variables() == (False, "auto", None)

    with mock.patch.dict(
        os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "Unsupported"}
    ):
        if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
            with pytest.raises(ValueError) as excinfo:
                BookingServiceClient._read_environment_variables()
            assert (
                str(excinfo.value)
                == "Environment variable `GOOGLE_API_USE_CLIENT_CERTIFICATE` must be either `true` or `false`"
            )
        else:
            assert BookingServiceClient._read_environment_variables() == (
            False,
            "auto",
            None,
        )

    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "never"}):
        assert BookingServiceClient._read_environment_variables() == (False, "never", None)

    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "always"}):
        assert BookingServiceClient._read_environment_variables() == (False, "always", None)

    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "auto"}):
        assert BookingServiceClient._read_environment_variables() == (False, "auto", None)

    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "Unsupported"}):
        with pytest.raises(MutualTLSChannelError) as excinfo:
            BookingServiceClient._read_environment_variables()
    assert str(excinfo.value) == "Environment variable `GOOGLE_API_USE_MTLS_ENDPOINT` must be `never`, `auto` or `always`"

    with mock.patch.dict(os.environ, {"GOOGLE_CLOUD_UNIVERSE_DOMAIN": "foo.com"}):
        assert BookingServiceClient._read_environment_variables() == (False, "auto", "foo.com")


def test_use_client_cert_effective():
    # Test case 1: Test when `should_use_client_cert` returns True.
    # We mock the `should_use_client_cert` function to simulate a scenario where
    # the google-auth library supports automatic mTLS and determines that a
    # client certificate should be used.
    if hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch("google.auth.transport.mtls.should_use_client_cert", return_value=True):
            assert BookingServiceClient._use_client_cert_effective() is True

    # Test case 2: Test when `should_use_client_cert` returns False.
    # We mock the `should_use_client_cert` function to simulate a scenario where
    # the google-auth library supports automatic mTLS and determines that a
    # client certificate should NOT be used.
    if hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch("google.auth.transport.mtls.should_use_client_cert", return_value=False):
            assert BookingServiceClient._use_client_cert_effective() is False

    # Test case 3: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to "true".
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "true"}):
            assert BookingServiceClient._use_client_cert_effective() is True

    # Test case 4: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to "false".
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "false"}):
            assert BookingServiceClient._use_client_cert_effective() is False

    # Test case 5: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to "True".
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "True"}):
            assert BookingServiceClient._use_client_cert_effective() is True

    # Test case 6: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to "False".
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "False"}):
            assert BookingServiceClient._use_client_cert_effective() is False

    # Test case 7: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to "TRUE".
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "TRUE"}):
            assert BookingServiceClient._use_client_cert_effective() is True

    # Test case 8: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to "FALSE".
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "FALSE"}):
            assert BookingServiceClient._use_client_cert_effective() is False

    # Test case 9: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is not set.
    # In this case, the method should return False, which is the default value.
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, clear=True):
            assert BookingServiceClient._use_client_cert_effective() is False

    # Test case 10: Test when `should_use_client_cert` is unavailable and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to an invalid value.
    # The method should raise a ValueError as the environment variable must be either
    # "true" or "false".
    if not hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "unsupported"}):
            with pytest.raises(ValueError):
                BookingServiceClient._use_client_cert_effective()

    # Test case 11: Test when `should_use_client_cert` is available and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is set to an invalid value.
    # The method should return False as the environment variable is set to an invalid value.
    if  hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "unsupported"}):
            assert BookingServiceClient._use_client_cert_effective() is False

    # Test case 12: Test when `should_use_client_cert` is available and the
    # `GOOGLE_API_USE_CLIENT_CERTIFICATE` environment variable is unset. Also,
    # the GOOGLE_API_CONFIG environment variable is unset.
    if  hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": ""}):
            with mock.patch.dict(os.environ, {"GOOGLE_API_CERTIFICATE_CONFIG": ""}):
                assert BookingServiceClient._use_client_cert_effective() is False

def test__get_client_cert_source():
    mock_provided_cert_source = mock.Mock()
    mock_default_cert_source = mock.Mock()

    assert BookingServiceClient._get_client_cert_source(None, False) is None
    assert BookingServiceClient._get_client_cert_source(mock_provided_cert_source, False) is None
    assert BookingServiceClient._get_client_cert_source(mock_provided_cert_source, True) == mock_provided_cert_source

    with mock.patch('google.auth.transport.mtls.has_default_client_cert_source', return_value=True):
        with mock.patch('google.auth.transport.mtls.default_client_cert_source', return_value=mock_default_cert_source):
            assert BookingServiceClient._get_client_cert_source(None, True) is mock_default_cert_source
            assert BookingServiceClient._get_client_cert_source(mock_provided_cert_source, "true") is mock_provided_cert_source

@mock.patch.object(BookingServiceClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceClient))
@mock.patch.object(BookingServiceAsyncClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceAsyncClient))
def test__get_api_endpoint():
    api_override = "foo.com"
    mock_client_cert_source = mock.Mock()
    default_universe = BookingServiceClient._DEFAULT_UNIVERSE
    default_endpoint = BookingServiceClient._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=default_universe)
    mock_universe = "bar.com"
    mock_endpoint = BookingServiceClient._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=mock_universe)

    assert BookingServiceClient._get_api_endpoint(api_override, mock_client_cert_source, default_universe, "always") == api_override
    assert BookingServiceClient._get_api_endpoint(None, mock_client_cert_source, default_universe, "auto") == BookingServiceClient.DEFAULT_MTLS_ENDPOINT
    assert BookingServiceClient._get_api_endpoint(None, None, default_universe, "auto") == default_endpoint
    assert BookingServiceClient._get_api_endpoint(None, None, default_universe, "always") == BookingServiceClient.DEFAULT_MTLS_ENDPOINT
    assert BookingServiceClient._get_api_endpoint(None, mock_client_cert_source, default_universe, "always") == BookingServiceClient.DEFAULT_MTLS_ENDPOINT
    assert BookingServiceClient._get_api_endpoint(None, None, mock_universe, "never") == mock_endpoint
    assert BookingServiceClient._get_api_endpoint(None, None, default_universe, "never") == default_endpoint

    with pytest.raises(MutualTLSChannelError) as excinfo:
        BookingServiceClient._get_api_endpoint(None, mock_client_cert_source, mock_universe, "auto")
    assert str(excinfo.value) == "mTLS is not supported in any universe other than googleapis.com."


def test__get_universe_domain():
    client_universe_domain = "foo.com"
    universe_domain_env = "bar.com"

    assert BookingServiceClient._get_universe_domain(client_universe_domain, universe_domain_env) == client_universe_domain
    assert BookingServiceClient._get_universe_domain(None, universe_domain_env) == universe_domain_env
    assert BookingServiceClient._get_universe_domain(None, None) == BookingServiceClient._DEFAULT_UNIVERSE

    with pytest.raises(ValueError) as excinfo:
        BookingServiceClient._get_universe_domain("", None)
    assert str(excinfo.value) == "Universe Domain cannot be an empty string."

@pytest.mark.parametrize("error_code,cred_info_json,show_cred_info", [
    (401, CRED_INFO_JSON, True),
    (403, CRED_INFO_JSON, True),
    (404, CRED_INFO_JSON, True),
    (500, CRED_INFO_JSON, False),
    (401, None, False),
    (403, None, False),
    (404, None, False),
    (500, None, False)
])
def test__add_cred_info_for_auth_errors(error_code, cred_info_json, show_cred_info):
    cred = mock.Mock(["get_cred_info"])
    cred.get_cred_info = mock.Mock(return_value=cred_info_json)
    client = BookingServiceClient(credentials=cred)
    client._transport._credentials = cred

    error = core_exceptions.GoogleAPICallError("message", details=["foo"])
    error.code = error_code

    client._add_cred_info_for_auth_errors(error)
    if show_cred_info:
        assert error.details == ["foo", CRED_INFO_STRING]
    else:
        assert error.details == ["foo"]

@pytest.mark.parametrize("error_code", [401,403,404,500])
def test__add_cred_info_for_auth_errors_no_get_cred_info(error_code):
    cred = mock.Mock([])
    assert not hasattr(cred, "get_cred_info")
    client = BookingServiceClient(credentials=cred)
    client._transport._credentials = cred

    error = core_exceptions.GoogleAPICallError("message", details=[])
    error.code = error_code

    client._add_cred_info_for_auth_errors(error)
    assert error.details == []

@pytest.mark.parametrize("client_class,transport_name", [
    (BookingServiceClient, "grpc"),
    (BookingServiceAsyncClient, "grpc_asyncio"),
])
def test_booking_service_client_from_service_account_info(client_class, transport_name):
    creds = ga_credentials.AnonymousCredentials()
    with mock.patch.object(service_account.Credentials, 'from_service_account_info') as factory:
        factory.return_value = creds
        info = {"valid": True}
        client = client_class.from_service_account_info(info, transport=transport_name)
        assert client.transport._credentials == creds
        assert isinstance(client, client_class)

        assert client.transport._host == (
            'freebusy.ohtarnished.dev:443'
        )


@pytest.mark.parametrize("transport_class,transport_name", [
    (transports.BookingServiceGrpcTransport, "grpc"),
    (transports.BookingServiceGrpcAsyncIOTransport, "grpc_asyncio"),
])
def test_booking_service_client_service_account_always_use_jwt(transport_class, transport_name):
    with mock.patch.object(service_account.Credentials, 'with_always_use_jwt_access', create=True) as use_jwt:
        creds = service_account.Credentials(None, None, None)
        transport = transport_class(credentials=creds, always_use_jwt_access=True)
        use_jwt.assert_called_once_with(True)

    with mock.patch.object(service_account.Credentials, 'with_always_use_jwt_access', create=True) as use_jwt:
        creds = service_account.Credentials(None, None, None)
        transport = transport_class(credentials=creds, always_use_jwt_access=False)
        use_jwt.assert_not_called()


@pytest.mark.parametrize("client_class,transport_name", [
    (BookingServiceClient, "grpc"),
    (BookingServiceAsyncClient, "grpc_asyncio"),
])
def test_booking_service_client_from_service_account_file(client_class, transport_name):
    creds = ga_credentials.AnonymousCredentials()
    with mock.patch.object(service_account.Credentials, 'from_service_account_file') as factory:
        factory.return_value = creds
        client = client_class.from_service_account_file("dummy/file/path.json", transport=transport_name)
        assert client.transport._credentials == creds
        assert isinstance(client, client_class)

        client = client_class.from_service_account_json("dummy/file/path.json", transport=transport_name)
        assert client.transport._credentials == creds
        assert isinstance(client, client_class)

        assert client.transport._host == (
            'freebusy.ohtarnished.dev:443'
        )


def test_booking_service_client_get_transport_class():
    transport = BookingServiceClient.get_transport_class()
    available_transports = [
        transports.BookingServiceGrpcTransport,
    ]
    assert transport in available_transports

    transport = BookingServiceClient.get_transport_class("grpc")
    assert transport == transports.BookingServiceGrpcTransport


@pytest.mark.parametrize("client_class,transport_class,transport_name", [
    (BookingServiceClient, transports.BookingServiceGrpcTransport, "grpc"),
    (BookingServiceAsyncClient, transports.BookingServiceGrpcAsyncIOTransport, "grpc_asyncio"),
])
@mock.patch.object(BookingServiceClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceClient))
@mock.patch.object(BookingServiceAsyncClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceAsyncClient))
def test_booking_service_client_client_options(client_class, transport_class, transport_name):
    # Check that if channel is provided we won't create a new one.
    with mock.patch.object(BookingServiceClient, 'get_transport_class') as gtc:
        transport = transport_class(
            credentials=ga_credentials.AnonymousCredentials()
        )
        client = client_class(transport=transport)
        gtc.assert_not_called()

    # Check that if channel is provided via str we will create a new one.
    with mock.patch.object(BookingServiceClient, 'get_transport_class') as gtc:
        client = client_class(transport=transport_name)
        gtc.assert_called()

    # Check the case api_endpoint is provided.
    options = client_options.ClientOptions(api_endpoint="squid.clam.whelk")
    with mock.patch.object(transport_class, '__init__') as patched:
        patched.return_value = None
        client = client_class(transport=transport_name, client_options=options)
        patched.assert_called_once_with(
            credentials=None,
            credentials_file=None,
            host="squid.clam.whelk",
            scopes=None,
            client_cert_source_for_mtls=None,
            quota_project_id=None,
            client_info=transports.base.DEFAULT_CLIENT_INFO,
            always_use_jwt_access=True,
            api_audience=None,
        )

    # Check the case api_endpoint is not provided and GOOGLE_API_USE_MTLS_ENDPOINT is
    # "never".
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "never"}):
        with mock.patch.object(transport_class, '__init__') as patched:
            patched.return_value = None
            client = client_class(transport=transport_name)
            patched.assert_called_once_with(
                credentials=None,
                credentials_file=None,
                host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
                scopes=None,
                client_cert_source_for_mtls=None,
                quota_project_id=None,
                client_info=transports.base.DEFAULT_CLIENT_INFO,
                always_use_jwt_access=True,
                api_audience=None,
            )

    # Check the case api_endpoint is not provided and GOOGLE_API_USE_MTLS_ENDPOINT is
    # "always".
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "always"}):
        with mock.patch.object(transport_class, '__init__') as patched:
            patched.return_value = None
            client = client_class(transport=transport_name)
            patched.assert_called_once_with(
                credentials=None,
                credentials_file=None,
                host=client.DEFAULT_MTLS_ENDPOINT,
                scopes=None,
                client_cert_source_for_mtls=None,
                quota_project_id=None,
                client_info=transports.base.DEFAULT_CLIENT_INFO,
                always_use_jwt_access=True,
                api_audience=None,
            )

    # Check the case api_endpoint is not provided and GOOGLE_API_USE_MTLS_ENDPOINT has
    # unsupported value.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "Unsupported"}):
        with pytest.raises(MutualTLSChannelError) as excinfo:
            client = client_class(transport=transport_name)
    assert str(excinfo.value) == "Environment variable `GOOGLE_API_USE_MTLS_ENDPOINT` must be `never`, `auto` or `always`"

    # Check the case quota_project_id is provided
    options = client_options.ClientOptions(quota_project_id="octopus")
    with mock.patch.object(transport_class, '__init__') as patched:
        patched.return_value = None
        client = client_class(client_options=options, transport=transport_name)
        patched.assert_called_once_with(
            credentials=None,
            credentials_file=None,
            host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
            scopes=None,
            client_cert_source_for_mtls=None,
            quota_project_id="octopus",
            client_info=transports.base.DEFAULT_CLIENT_INFO,
            always_use_jwt_access=True,
            api_audience=None,
        )
    # Check the case api_endpoint is provided
    options = client_options.ClientOptions(api_audience="https://language.googleapis.com")
    with mock.patch.object(transport_class, '__init__') as patched:
        patched.return_value = None
        client = client_class(client_options=options, transport=transport_name)
        patched.assert_called_once_with(
            credentials=None,
            credentials_file=None,
            host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
            scopes=None,
            client_cert_source_for_mtls=None,
            quota_project_id=None,
            client_info=transports.base.DEFAULT_CLIENT_INFO,
            always_use_jwt_access=True,
            api_audience="https://language.googleapis.com"
        )

@pytest.mark.parametrize("client_class,transport_class,transport_name,use_client_cert_env", [
    (BookingServiceClient, transports.BookingServiceGrpcTransport, "grpc", "true"),
    (BookingServiceAsyncClient, transports.BookingServiceGrpcAsyncIOTransport, "grpc_asyncio", "true"),
    (BookingServiceClient, transports.BookingServiceGrpcTransport, "grpc", "false"),
    (BookingServiceAsyncClient, transports.BookingServiceGrpcAsyncIOTransport, "grpc_asyncio", "false"),
])
@mock.patch.object(BookingServiceClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceClient))
@mock.patch.object(BookingServiceAsyncClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceAsyncClient))
@mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "auto"})
def test_booking_service_client_mtls_env_auto(client_class, transport_class, transport_name, use_client_cert_env):
    # This tests the endpoint autoswitch behavior. Endpoint is autoswitched to the default
    # mtls endpoint, if GOOGLE_API_USE_CLIENT_CERTIFICATE is "true" and client cert exists.

    # Check the case client_cert_source is provided. Whether client cert is used depends on
    # GOOGLE_API_USE_CLIENT_CERTIFICATE value.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": use_client_cert_env}):
        options = client_options.ClientOptions(client_cert_source=client_cert_source_callback)
        with mock.patch.object(transport_class, '__init__') as patched:
            patched.return_value = None
            client = client_class(client_options=options, transport=transport_name)

            if use_client_cert_env == "false":
                expected_client_cert_source = None
                expected_host = client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE)
            else:
                expected_client_cert_source = client_cert_source_callback
                expected_host = client.DEFAULT_MTLS_ENDPOINT

            patched.assert_called_once_with(
                credentials=None,
                credentials_file=None,
                host=expected_host,
                scopes=None,
                client_cert_source_for_mtls=expected_client_cert_source,
                quota_project_id=None,
                client_info=transports.base.DEFAULT_CLIENT_INFO,
                always_use_jwt_access=True,
                api_audience=None,
            )

    # Check the case ADC client cert is provided. Whether client cert is used depends on
    # GOOGLE_API_USE_CLIENT_CERTIFICATE value.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": use_client_cert_env}):
        with mock.patch.object(transport_class, '__init__') as patched:
            with mock.patch('google.auth.transport.mtls.has_default_client_cert_source', return_value=True):
                with mock.patch('google.auth.transport.mtls.default_client_cert_source', return_value=client_cert_source_callback):
                    if use_client_cert_env == "false":
                        expected_host = client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE)
                        expected_client_cert_source = None
                    else:
                        expected_host = client.DEFAULT_MTLS_ENDPOINT
                        expected_client_cert_source = client_cert_source_callback

                    patched.return_value = None
                    client = client_class(transport=transport_name)
                    patched.assert_called_once_with(
                        credentials=None,
                        credentials_file=None,
                        host=expected_host,
                        scopes=None,
                        client_cert_source_for_mtls=expected_client_cert_source,
                        quota_project_id=None,
                        client_info=transports.base.DEFAULT_CLIENT_INFO,
                        always_use_jwt_access=True,
                        api_audience=None,
                    )

    # Check the case client_cert_source and ADC client cert are not provided.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": use_client_cert_env}):
        with mock.patch.object(transport_class, '__init__') as patched:
            with mock.patch("google.auth.transport.mtls.has_default_client_cert_source", return_value=False):
                patched.return_value = None
                client = client_class(transport=transport_name)
                patched.assert_called_once_with(
                    credentials=None,
                    credentials_file=None,
                    host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
                    scopes=None,
                    client_cert_source_for_mtls=None,
                    quota_project_id=None,
                    client_info=transports.base.DEFAULT_CLIENT_INFO,
                    always_use_jwt_access=True,
                    api_audience=None,
                )


@pytest.mark.parametrize("client_class", [
    BookingServiceClient, BookingServiceAsyncClient
])
@mock.patch.object(BookingServiceClient, "DEFAULT_ENDPOINT", modify_default_endpoint(BookingServiceClient))
@mock.patch.object(BookingServiceAsyncClient, "DEFAULT_ENDPOINT", modify_default_endpoint(BookingServiceAsyncClient))
def test_booking_service_client_get_mtls_endpoint_and_cert_source(client_class):
    mock_client_cert_source = mock.Mock()

    # Test the case GOOGLE_API_USE_CLIENT_CERTIFICATE is "true".
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "true"}):
        mock_api_endpoint = "foo"
        options = client_options.ClientOptions(client_cert_source=mock_client_cert_source, api_endpoint=mock_api_endpoint)
        api_endpoint, cert_source = client_class.get_mtls_endpoint_and_cert_source(options)
        assert api_endpoint == mock_api_endpoint
        assert cert_source == mock_client_cert_source

    # Test the case GOOGLE_API_USE_CLIENT_CERTIFICATE is "false".
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "false"}):
        mock_client_cert_source = mock.Mock()
        mock_api_endpoint = "foo"
        options = client_options.ClientOptions(client_cert_source=mock_client_cert_source, api_endpoint=mock_api_endpoint)
        api_endpoint, cert_source = client_class.get_mtls_endpoint_and_cert_source(options)
        assert api_endpoint == mock_api_endpoint
        assert cert_source is None

    # Test the case GOOGLE_API_USE_CLIENT_CERTIFICATE is "Unsupported".
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "Unsupported"}):
        if hasattr(google.auth.transport.mtls, "should_use_client_cert"):
            mock_client_cert_source = mock.Mock()
            mock_api_endpoint = "foo"
            options = client_options.ClientOptions(
                client_cert_source=mock_client_cert_source, api_endpoint=mock_api_endpoint
            )
            api_endpoint, cert_source = client_class.get_mtls_endpoint_and_cert_source(
                options
            )
            assert api_endpoint == mock_api_endpoint
            assert cert_source is None

    # Test cases for mTLS enablement when GOOGLE_API_USE_CLIENT_CERTIFICATE is unset.
    test_cases = [
        (
            # With workloads present in config, mTLS is enabled.
            {
                "version": 1,
                "cert_configs": {
                    "workload": {
                        "cert_path": "path/to/cert/file",
                        "key_path": "path/to/key/file",
                    }
                },
            },
            mock_client_cert_source,
        ),
        (
            # With workloads not present in config, mTLS is disabled.
            {
                "version": 1,
                "cert_configs": {},
            },
            None,
        ),
    ]
    if hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        for config_data, expected_cert_source in test_cases:
            env = os.environ.copy()
            env.pop("GOOGLE_API_USE_CLIENT_CERTIFICATE", None)
            with mock.patch.dict(os.environ, env, clear=True):
                    config_filename = "mock_certificate_config.json"
                    config_file_content = json.dumps(config_data)
                    m = mock.mock_open(read_data=config_file_content)
                    with mock.patch("builtins.open", m):
                        with mock.patch.dict(
                            os.environ, {"GOOGLE_API_CERTIFICATE_CONFIG": config_filename}
                        ):
                            mock_api_endpoint = "foo"
                            options = client_options.ClientOptions(
                                client_cert_source=mock_client_cert_source,
                                api_endpoint=mock_api_endpoint,
                            )
                            api_endpoint, cert_source = (
                                client_class.get_mtls_endpoint_and_cert_source(options)
                            )
                            assert api_endpoint == mock_api_endpoint
                            assert cert_source is expected_cert_source

    # Test cases for mTLS enablement when GOOGLE_API_USE_CLIENT_CERTIFICATE is unset(empty).
    test_cases = [
        (
            # With workloads present in config, mTLS is enabled.
            {
                "version": 1,
                "cert_configs": {
                    "workload": {
                        "cert_path": "path/to/cert/file",
                        "key_path": "path/to/key/file",
                    }
                },
            },
            mock_client_cert_source,
        ),
        (
            # With workloads not present in config, mTLS is disabled.
            {
                "version": 1,
                "cert_configs": {},
            },
            None,
        ),
    ]
    if hasattr(google.auth.transport.mtls, "should_use_client_cert"):
        for config_data, expected_cert_source in test_cases:
            env = os.environ.copy()
            env.pop("GOOGLE_API_USE_CLIENT_CERTIFICATE", "")
            with mock.patch.dict(os.environ, env, clear=True):
                    config_filename = "mock_certificate_config.json"
                    config_file_content = json.dumps(config_data)
                    m = mock.mock_open(read_data=config_file_content)
                    with mock.patch("builtins.open", m):
                        with mock.patch.dict(
                            os.environ, {"GOOGLE_API_CERTIFICATE_CONFIG": config_filename}
                        ):
                            mock_api_endpoint = "foo"
                            options = client_options.ClientOptions(
                                client_cert_source=mock_client_cert_source,
                                api_endpoint=mock_api_endpoint,
                            )
                            api_endpoint, cert_source = (
                                client_class.get_mtls_endpoint_and_cert_source(options)
                            )
                            assert api_endpoint == mock_api_endpoint
                            assert cert_source is expected_cert_source

    # Test the case GOOGLE_API_USE_MTLS_ENDPOINT is "never".
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "never"}):
        api_endpoint, cert_source = client_class.get_mtls_endpoint_and_cert_source()
        assert api_endpoint == client_class.DEFAULT_ENDPOINT
        assert cert_source is None

    # Test the case GOOGLE_API_USE_MTLS_ENDPOINT is "always".
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "always"}):
        api_endpoint, cert_source = client_class.get_mtls_endpoint_and_cert_source()
        assert api_endpoint == client_class.DEFAULT_MTLS_ENDPOINT
        assert cert_source is None

    # Test the case GOOGLE_API_USE_MTLS_ENDPOINT is "auto" and default cert doesn't exist.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "true"}):
        with mock.patch('google.auth.transport.mtls.has_default_client_cert_source', return_value=False):
            api_endpoint, cert_source = client_class.get_mtls_endpoint_and_cert_source()
            assert api_endpoint == client_class.DEFAULT_ENDPOINT
            assert cert_source is None

    # Test the case GOOGLE_API_USE_MTLS_ENDPOINT is "auto" and default cert exists.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "true"}):
        with mock.patch('google.auth.transport.mtls.has_default_client_cert_source', return_value=True):
            with mock.patch('google.auth.transport.mtls.default_client_cert_source', return_value=mock_client_cert_source):
                api_endpoint, cert_source = client_class.get_mtls_endpoint_and_cert_source()
                assert api_endpoint == client_class.DEFAULT_MTLS_ENDPOINT
                assert cert_source == mock_client_cert_source

    # Check the case api_endpoint is not provided and GOOGLE_API_USE_MTLS_ENDPOINT has
    # unsupported value.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "Unsupported"}):
        with pytest.raises(MutualTLSChannelError) as excinfo:
            client_class.get_mtls_endpoint_and_cert_source()

        assert str(excinfo.value) == "Environment variable `GOOGLE_API_USE_MTLS_ENDPOINT` must be `never`, `auto` or `always`"

@pytest.mark.parametrize("client_class", [
    BookingServiceClient, BookingServiceAsyncClient
])
@mock.patch.object(BookingServiceClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceClient))
@mock.patch.object(BookingServiceAsyncClient, "_DEFAULT_ENDPOINT_TEMPLATE", modify_default_endpoint_template(BookingServiceAsyncClient))
def test_booking_service_client_client_api_endpoint(client_class):
    mock_client_cert_source = client_cert_source_callback
    api_override = "foo.com"
    default_universe = BookingServiceClient._DEFAULT_UNIVERSE
    default_endpoint = BookingServiceClient._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=default_universe)
    mock_universe = "bar.com"
    mock_endpoint = BookingServiceClient._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=mock_universe)

    # If ClientOptions.api_endpoint is set and GOOGLE_API_USE_CLIENT_CERTIFICATE="true",
    # use ClientOptions.api_endpoint as the api endpoint regardless.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_CLIENT_CERTIFICATE": "true"}):
        with mock.patch("google.auth.transport.requests.AuthorizedSession.configure_mtls_channel"):
            options = client_options.ClientOptions(client_cert_source=mock_client_cert_source, api_endpoint=api_override)
            client = client_class(client_options=options, credentials=ga_credentials.AnonymousCredentials())
            assert client.api_endpoint == api_override

    # If ClientOptions.api_endpoint is not set and GOOGLE_API_USE_MTLS_ENDPOINT="never",
    # use the _DEFAULT_ENDPOINT_TEMPLATE populated with GDU as the api endpoint.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "never"}):
        client = client_class(credentials=ga_credentials.AnonymousCredentials())
        assert client.api_endpoint == default_endpoint

    # If ClientOptions.api_endpoint is not set and GOOGLE_API_USE_MTLS_ENDPOINT="always",
    # use the DEFAULT_MTLS_ENDPOINT as the api endpoint.
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "always"}):
        client = client_class(credentials=ga_credentials.AnonymousCredentials())
        assert client.api_endpoint == client_class.DEFAULT_MTLS_ENDPOINT

    # If ClientOptions.api_endpoint is not set, GOOGLE_API_USE_MTLS_ENDPOINT="auto" (default),
    # GOOGLE_API_USE_CLIENT_CERTIFICATE="false" (default), default cert source doesn't exist,
    # and ClientOptions.universe_domain="bar.com",
    # use the _DEFAULT_ENDPOINT_TEMPLATE populated with universe domain as the api endpoint.
    options = client_options.ClientOptions()
    universe_exists = hasattr(options, "universe_domain")
    if universe_exists:
        options = client_options.ClientOptions(universe_domain=mock_universe)
        client = client_class(client_options=options, credentials=ga_credentials.AnonymousCredentials())
    else:
        client = client_class(client_options=options, credentials=ga_credentials.AnonymousCredentials())
    assert client.api_endpoint == (mock_endpoint if universe_exists else default_endpoint)
    assert client.universe_domain == (mock_universe if universe_exists else default_universe)

    # If ClientOptions does not have a universe domain attribute and GOOGLE_API_USE_MTLS_ENDPOINT="never",
    # use the _DEFAULT_ENDPOINT_TEMPLATE populated with GDU as the api endpoint.
    options = client_options.ClientOptions()
    if hasattr(options, "universe_domain"):
        delattr(options, "universe_domain")
    with mock.patch.dict(os.environ, {"GOOGLE_API_USE_MTLS_ENDPOINT": "never"}):
        client = client_class(client_options=options, credentials=ga_credentials.AnonymousCredentials())
        assert client.api_endpoint == default_endpoint


@pytest.mark.parametrize("client_class,transport_class,transport_name", [
    (BookingServiceClient, transports.BookingServiceGrpcTransport, "grpc"),
    (BookingServiceAsyncClient, transports.BookingServiceGrpcAsyncIOTransport, "grpc_asyncio"),
])
def test_booking_service_client_client_options_scopes(client_class, transport_class, transport_name):
    # Check the case scopes are provided.
    options = client_options.ClientOptions(
        scopes=["1", "2"],
    )
    with mock.patch.object(transport_class, '__init__') as patched:
        patched.return_value = None
        client = client_class(client_options=options, transport=transport_name)
        patched.assert_called_once_with(
            credentials=None,
            credentials_file=None,
            host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
            scopes=["1", "2"],
            client_cert_source_for_mtls=None,
            quota_project_id=None,
            client_info=transports.base.DEFAULT_CLIENT_INFO,
            always_use_jwt_access=True,
            api_audience=None,
        )

@pytest.mark.parametrize("client_class,transport_class,transport_name,grpc_helpers", [
    (BookingServiceClient, transports.BookingServiceGrpcTransport, "grpc", grpc_helpers),
    (BookingServiceAsyncClient, transports.BookingServiceGrpcAsyncIOTransport, "grpc_asyncio", grpc_helpers_async),
])
def test_booking_service_client_client_options_credentials_file(client_class, transport_class, transport_name, grpc_helpers):
    # Check the case credentials file is provided.
    options = client_options.ClientOptions(
        credentials_file="credentials.json"
    )

    with mock.patch.object(transport_class, '__init__') as patched:
        patched.return_value = None
        client = client_class(client_options=options, transport=transport_name)
        patched.assert_called_once_with(
            credentials=None,
            credentials_file="credentials.json",
            host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
            scopes=None,
            client_cert_source_for_mtls=None,
            quota_project_id=None,
            client_info=transports.base.DEFAULT_CLIENT_INFO,
            always_use_jwt_access=True,
            api_audience=None,
        )

def test_booking_service_client_client_options_from_dict():
    with mock.patch('freebusy.booking_v1.services.booking_service.transports.BookingServiceGrpcTransport.__init__') as grpc_transport:
        grpc_transport.return_value = None
        client = BookingServiceClient(
            client_options={'api_endpoint': 'squid.clam.whelk'}
        )
        grpc_transport.assert_called_once_with(
            credentials=None,
            credentials_file=None,
            host="squid.clam.whelk",
            scopes=None,
            client_cert_source_for_mtls=None,
            quota_project_id=None,
            client_info=transports.base.DEFAULT_CLIENT_INFO,
            always_use_jwt_access=True,
            api_audience=None,
        )


@pytest.mark.parametrize("client_class,transport_class,transport_name,grpc_helpers", [
    (BookingServiceClient, transports.BookingServiceGrpcTransport, "grpc", grpc_helpers),
    (BookingServiceAsyncClient, transports.BookingServiceGrpcAsyncIOTransport, "grpc_asyncio", grpc_helpers_async),
])
def test_booking_service_client_create_channel_credentials_file(client_class, transport_class, transport_name, grpc_helpers):
    # Check the case credentials file is provided.
    options = client_options.ClientOptions(
        credentials_file="credentials.json"
    )

    with mock.patch.object(transport_class, '__init__') as patched:
        patched.return_value = None
        client = client_class(client_options=options, transport=transport_name)
        patched.assert_called_once_with(
            credentials=None,
            credentials_file="credentials.json",
            host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
            scopes=None,
            client_cert_source_for_mtls=None,
            quota_project_id=None,
            client_info=transports.base.DEFAULT_CLIENT_INFO,
            always_use_jwt_access=True,
            api_audience=None,
        )

    # test that the credentials from file are saved and used as the credentials.
    with mock.patch.object(
        google.auth, "load_credentials_from_file", autospec=True
    ) as load_creds, mock.patch.object(
        google.auth, "default", autospec=True
    ) as adc, mock.patch.object(
        grpc_helpers, "create_channel"
    ) as create_channel:
        creds = ga_credentials.AnonymousCredentials()
        file_creds = ga_credentials.AnonymousCredentials()
        load_creds.return_value = (file_creds, None)
        adc.return_value = (creds, None)
        client = client_class(client_options=options, transport=transport_name)
        create_channel.assert_called_with(
            "freebusy.ohtarnished.dev:443",
            credentials=file_creds,
            credentials_file=None,
            quota_project_id=None,
            default_scopes=(
),
            scopes=None,
            default_host="freebusy.ohtarnished.dev",
            ssl_credentials=None,
            options=[
                ("grpc.max_send_message_length", -1),
                ("grpc.max_receive_message_length", -1),
            ],
        )


@pytest.mark.parametrize("request_type", [
  booking_messages.CreateBookingRequest(),
  {},
])
def test_create_booking(request_type, transport: str = 'grpc'):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.create_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = fb_booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        )
        response = client.create_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        request = booking_messages.CreateBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, fb_booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'


def test_create_booking_non_empty_request_with_auto_populated_field():
    # This test is a coverage failsafe to make sure that UUID4 fields are
    # automatically populated, according to AIP-4235, with non-empty requests.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport='grpc',
    )

    # Populate all string fields in the request which are not UUID4
    # since we want to check that UUID4 are populated automatically
    # if they meet the requirements of AIP 4235.
    request = booking_messages.CreateBookingRequest(
        booking_id='booking_id_value',
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.create_booking),
            '__call__') as call:
        call.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client.create_booking(request=request)
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.CreateBookingRequest(
            booking_id='booking_id_value',
        )
        assert args[0] == request_msg

def test_create_booking_use_cached_wrapped_rpc():
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method.wrap_method") as wrapper_fn:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport="grpc",
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._transport.create_booking in client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.Mock()
        mock_rpc.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client._transport._wrapped_methods[client._transport.create_booking] = mock_rpc
        request = {}
        client.create_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        client.create_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
async def test_create_booking_async_use_cached_wrapped_rpc(transport: str = "grpc_asyncio"):
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method_async.wrap_method") as wrapper_fn:
        client = BookingServiceAsyncClient(
            credentials=async_anonymous_credentials(),
            transport=transport,
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._client._transport.create_booking in client._client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.AsyncMock()
        mock_rpc.return_value = mock.Mock()
        client._client._transport._wrapped_methods[client._client._transport.create_booking] = mock_rpc

        request = {}
        await client.create_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        await client.create_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
@pytest.mark.parametrize("request_type", [
  booking_messages.CreateBookingRequest(),
  {},
])
async def test_create_booking_async(request_type, transport: str = 'grpc_asyncio'):
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.create_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value =grpc_helpers_async.FakeUnaryUnaryCall(fb_booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        response = await client.create_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        request = booking_messages.CreateBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, fb_booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'


def test_create_booking_flattened():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.create_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = fb_booking.Booking()
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        client.create_booking(
            booking=fb_booking.Booking(name='name_value'),
            booking_id='booking_id_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        arg = args[0].booking
        mock_val = fb_booking.Booking(name='name_value')
        assert arg == mock_val
        arg = args[0].booking_id
        mock_val = 'booking_id_value'
        assert arg == mock_val


def test_create_booking_flattened_error():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        client.create_booking(
            booking_messages.CreateBookingRequest(),
            booking=fb_booking.Booking(name='name_value'),
            booking_id='booking_id_value',
        )

@pytest.mark.asyncio
async def test_create_booking_flattened_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.create_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = fb_booking.Booking()

        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(fb_booking.Booking())
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        response = await client.create_booking(
            booking=fb_booking.Booking(name='name_value'),
            booking_id='booking_id_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        arg = args[0].booking
        mock_val = fb_booking.Booking(name='name_value')
        assert arg == mock_val
        arg = args[0].booking_id
        mock_val = 'booking_id_value'
        assert arg == mock_val

@pytest.mark.asyncio
async def test_create_booking_flattened_error_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        await client.create_booking(
            booking_messages.CreateBookingRequest(),
            booking=fb_booking.Booking(name='name_value'),
            booking_id='booking_id_value',
        )


@pytest.mark.parametrize("request_type", [
  booking_messages.GetBookingRequest(),
  {},
])
def test_get_booking(request_type, transport: str = 'grpc'):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        )
        response = client.get_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        request = booking_messages.GetBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'


def test_get_booking_non_empty_request_with_auto_populated_field():
    # This test is a coverage failsafe to make sure that UUID4 fields are
    # automatically populated, according to AIP-4235, with non-empty requests.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport='grpc',
    )

    # Populate all string fields in the request which are not UUID4
    # since we want to check that UUID4 are populated automatically
    # if they meet the requirements of AIP 4235.
    request = booking_messages.GetBookingRequest(
        name='name_value',
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        call.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client.get_booking(request=request)
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.GetBookingRequest(
            name='name_value',
        )
        assert args[0] == request_msg

def test_get_booking_use_cached_wrapped_rpc():
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method.wrap_method") as wrapper_fn:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport="grpc",
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._transport.get_booking in client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.Mock()
        mock_rpc.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client._transport._wrapped_methods[client._transport.get_booking] = mock_rpc
        request = {}
        client.get_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        client.get_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
async def test_get_booking_async_use_cached_wrapped_rpc(transport: str = "grpc_asyncio"):
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method_async.wrap_method") as wrapper_fn:
        client = BookingServiceAsyncClient(
            credentials=async_anonymous_credentials(),
            transport=transport,
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._client._transport.get_booking in client._client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.AsyncMock()
        mock_rpc.return_value = mock.Mock()
        client._client._transport._wrapped_methods[client._client._transport.get_booking] = mock_rpc

        request = {}
        await client.get_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        await client.get_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
@pytest.mark.parametrize("request_type", [
  booking_messages.GetBookingRequest(),
  {},
])
async def test_get_booking_async(request_type, transport: str = 'grpc_asyncio'):
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value =grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        response = await client.get_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        request = booking_messages.GetBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'

def test_get_booking_field_headers():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_messages.GetBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.get_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


@pytest.mark.asyncio
async def test_get_booking_field_headers_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_messages.GetBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        await client.get_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


def test_get_booking_flattened():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        client.get_booking(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val


def test_get_booking_flattened_error():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        client.get_booking(
            booking_messages.GetBookingRequest(),
            name='name_value',
        )

@pytest.mark.asyncio
async def test_get_booking_flattened_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()

        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        response = await client.get_booking(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val

@pytest.mark.asyncio
async def test_get_booking_flattened_error_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        await client.get_booking(
            booking_messages.GetBookingRequest(),
            name='name_value',
        )


@pytest.mark.parametrize("request_type", [
  booking_messages.ListBookingsRequest(),
  {},
])
def test_list_bookings(request_type, transport: str = 'grpc'):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking_messages.ListBookingsResponse(
            next_page_token='next_page_token_value',
        )
        response = client.list_bookings(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        request = booking_messages.ListBookingsRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, pagers.ListBookingsPager)
    assert response.next_page_token == 'next_page_token_value'


def test_list_bookings_non_empty_request_with_auto_populated_field():
    # This test is a coverage failsafe to make sure that UUID4 fields are
    # automatically populated, according to AIP-4235, with non-empty requests.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport='grpc',
    )

    # Populate all string fields in the request which are not UUID4
    # since we want to check that UUID4 are populated automatically
    # if they meet the requirements of AIP 4235.
    request = booking_messages.ListBookingsRequest(
        page_token='page_token_value',
        filter='filter_value',
        order_by='order_by_value',
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__') as call:
        call.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client.list_bookings(request=request)
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.ListBookingsRequest(
            page_token='page_token_value',
            filter='filter_value',
            order_by='order_by_value',
        )
        assert args[0] == request_msg

def test_list_bookings_use_cached_wrapped_rpc():
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method.wrap_method") as wrapper_fn:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport="grpc",
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._transport.list_bookings in client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.Mock()
        mock_rpc.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client._transport._wrapped_methods[client._transport.list_bookings] = mock_rpc
        request = {}
        client.list_bookings(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        client.list_bookings(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
async def test_list_bookings_async_use_cached_wrapped_rpc(transport: str = "grpc_asyncio"):
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method_async.wrap_method") as wrapper_fn:
        client = BookingServiceAsyncClient(
            credentials=async_anonymous_credentials(),
            transport=transport,
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._client._transport.list_bookings in client._client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.AsyncMock()
        mock_rpc.return_value = mock.Mock()
        client._client._transport._wrapped_methods[client._client._transport.list_bookings] = mock_rpc

        request = {}
        await client.list_bookings(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        await client.list_bookings(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
@pytest.mark.parametrize("request_type", [
  booking_messages.ListBookingsRequest(),
  {},
])
async def test_list_bookings_async(request_type, transport: str = 'grpc_asyncio'):
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value =grpc_helpers_async.FakeUnaryUnaryCall(booking_messages.ListBookingsResponse(
            next_page_token='next_page_token_value',
        ))
        response = await client.list_bookings(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        request = booking_messages.ListBookingsRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, pagers.ListBookingsAsyncPager)
    assert response.next_page_token == 'next_page_token_value'


def test_list_bookings_pager(transport_name: str = "grpc"):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport_name,
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__') as call:
        # Set the response to a series of pages.
        call.side_effect = (
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                    booking.Booking(),
                ],
                next_page_token='abc',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[],
                next_page_token='def',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                ],
                next_page_token='ghi',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                ],
            ),
            RuntimeError,
        )

        expected_metadata = ()
        retry = retries.Retry()
        timeout = 5
        pager = client.list_bookings(request={}, retry=retry, timeout=timeout)

        assert pager._metadata == expected_metadata
        assert pager._retry == retry
        assert pager._timeout == timeout

        assert pager.next_page_token == 'abc'
        assert str(pager).startswith(f'{pager.__class__.__name__}<')

        results = list(pager)
        assert len(results) == 6
        assert all(isinstance(i, booking.Booking)
                   for i in results)
def test_list_bookings_pages(transport_name: str = "grpc"):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport_name,
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__') as call:
        # Set the response to a series of pages.
        call.side_effect = (
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                    booking.Booking(),
                ],
                next_page_token='abc',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[],
                next_page_token='def',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                ],
                next_page_token='ghi',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                ],
            ),
            RuntimeError,
        )
        pages = list(client.list_bookings(request={}).pages)
        for page_, token in zip(pages, ['abc','def','ghi', '']):
            assert page_.raw_page.next_page_token == token

@pytest.mark.asyncio
async def test_list_bookings_async_pager():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__', new_callable=mock.AsyncMock) as call:
        # Set the response to a series of pages.
        call.side_effect = (
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                    booking.Booking(),
                ],
                next_page_token='abc',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[],
                next_page_token='def',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                ],
                next_page_token='ghi',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                ],
            ),
            RuntimeError,
        )
        async_pager = await client.list_bookings(request={},)
        assert async_pager.next_page_token == 'abc'
        assert str(async_pager).startswith(f'{async_pager.__class__.__name__}<')

        responses = []
        async for response in async_pager: # pragma: no branch
            responses.append(response)

        assert len(responses) == 6
        assert all(isinstance(i, booking.Booking)
                for i in responses)


@pytest.mark.asyncio
async def test_list_bookings_async_pages():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__', new_callable=mock.AsyncMock) as call:
        # Set the response to a series of pages.
        call.side_effect = (
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                    booking.Booking(),
                ],
                next_page_token='abc',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[],
                next_page_token='def',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                ],
                next_page_token='ghi',
            ),
            booking_messages.ListBookingsResponse(
                bookings=[
                    booking.Booking(),
                    booking.Booking(),
                ],
            ),
            RuntimeError,
        )
        pages = []
        async for page_ in (
            await client.list_bookings(request={})
        ).pages:
            pages.append(page_)
        for page_, token in zip(pages, ['abc','def','ghi', '']):
            assert page_.raw_page.next_page_token == token

@pytest.mark.parametrize("request_type", [
  booking_actions.ConfirmBookingRequest(),
  {},
])
def test_confirm_booking(request_type, transport: str = 'grpc'):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        )
        response = client.confirm_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        request = booking_actions.ConfirmBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'


def test_confirm_booking_non_empty_request_with_auto_populated_field():
    # This test is a coverage failsafe to make sure that UUID4 fields are
    # automatically populated, according to AIP-4235, with non-empty requests.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport='grpc',
    )

    # Populate all string fields in the request which are not UUID4
    # since we want to check that UUID4 are populated automatically
    # if they meet the requirements of AIP 4235.
    request = booking_actions.ConfirmBookingRequest(
        name='name_value',
        payment_ref='payment_ref_value',
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        call.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client.confirm_booking(request=request)
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.ConfirmBookingRequest(
            name='name_value',
            payment_ref='payment_ref_value',
        )
        assert args[0] == request_msg

def test_confirm_booking_use_cached_wrapped_rpc():
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method.wrap_method") as wrapper_fn:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport="grpc",
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._transport.confirm_booking in client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.Mock()
        mock_rpc.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client._transport._wrapped_methods[client._transport.confirm_booking] = mock_rpc
        request = {}
        client.confirm_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        client.confirm_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
async def test_confirm_booking_async_use_cached_wrapped_rpc(transport: str = "grpc_asyncio"):
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method_async.wrap_method") as wrapper_fn:
        client = BookingServiceAsyncClient(
            credentials=async_anonymous_credentials(),
            transport=transport,
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._client._transport.confirm_booking in client._client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.AsyncMock()
        mock_rpc.return_value = mock.Mock()
        client._client._transport._wrapped_methods[client._client._transport.confirm_booking] = mock_rpc

        request = {}
        await client.confirm_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        await client.confirm_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
@pytest.mark.parametrize("request_type", [
  booking_actions.ConfirmBookingRequest(),
  {},
])
async def test_confirm_booking_async(request_type, transport: str = 'grpc_asyncio'):
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value =grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        response = await client.confirm_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        request = booking_actions.ConfirmBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'

def test_confirm_booking_field_headers():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.ConfirmBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.confirm_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


@pytest.mark.asyncio
async def test_confirm_booking_field_headers_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.ConfirmBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        await client.confirm_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


def test_confirm_booking_flattened():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        client.confirm_booking(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val


def test_confirm_booking_flattened_error():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        client.confirm_booking(
            booking_actions.ConfirmBookingRequest(),
            name='name_value',
        )

@pytest.mark.asyncio
async def test_confirm_booking_flattened_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()

        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        response = await client.confirm_booking(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val

@pytest.mark.asyncio
async def test_confirm_booking_flattened_error_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        await client.confirm_booking(
            booking_actions.ConfirmBookingRequest(),
            name='name_value',
        )


@pytest.mark.parametrize("request_type", [
  booking_actions.CancelBookingRequest(),
  {},
])
def test_cancel_booking(request_type, transport: str = 'grpc'):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        )
        response = client.cancel_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        request = booking_actions.CancelBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'


def test_cancel_booking_non_empty_request_with_auto_populated_field():
    # This test is a coverage failsafe to make sure that UUID4 fields are
    # automatically populated, according to AIP-4235, with non-empty requests.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport='grpc',
    )

    # Populate all string fields in the request which are not UUID4
    # since we want to check that UUID4 are populated automatically
    # if they meet the requirements of AIP 4235.
    request = booking_actions.CancelBookingRequest(
        name='name_value',
        note='note_value',
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        call.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client.cancel_booking(request=request)
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.CancelBookingRequest(
            name='name_value',
            note='note_value',
        )
        assert args[0] == request_msg

def test_cancel_booking_use_cached_wrapped_rpc():
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method.wrap_method") as wrapper_fn:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport="grpc",
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._transport.cancel_booking in client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.Mock()
        mock_rpc.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client._transport._wrapped_methods[client._transport.cancel_booking] = mock_rpc
        request = {}
        client.cancel_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        client.cancel_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
async def test_cancel_booking_async_use_cached_wrapped_rpc(transport: str = "grpc_asyncio"):
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method_async.wrap_method") as wrapper_fn:
        client = BookingServiceAsyncClient(
            credentials=async_anonymous_credentials(),
            transport=transport,
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._client._transport.cancel_booking in client._client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.AsyncMock()
        mock_rpc.return_value = mock.Mock()
        client._client._transport._wrapped_methods[client._client._transport.cancel_booking] = mock_rpc

        request = {}
        await client.cancel_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        await client.cancel_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
@pytest.mark.parametrize("request_type", [
  booking_actions.CancelBookingRequest(),
  {},
])
async def test_cancel_booking_async(request_type, transport: str = 'grpc_asyncio'):
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value =grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        response = await client.cancel_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        request = booking_actions.CancelBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'

def test_cancel_booking_field_headers():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.CancelBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.cancel_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


@pytest.mark.asyncio
async def test_cancel_booking_field_headers_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.CancelBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        await client.cancel_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


def test_cancel_booking_flattened():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        client.cancel_booking(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val


def test_cancel_booking_flattened_error():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        client.cancel_booking(
            booking_actions.CancelBookingRequest(),
            name='name_value',
        )

@pytest.mark.asyncio
async def test_cancel_booking_flattened_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()

        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        response = await client.cancel_booking(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val

@pytest.mark.asyncio
async def test_cancel_booking_flattened_error_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        await client.cancel_booking(
            booking_actions.CancelBookingRequest(),
            name='name_value',
        )


@pytest.mark.parametrize("request_type", [
  booking_actions.PreviewCancellationRequest(),
  {},
])
def test_preview_cancellation(request_type, transport: str = 'grpc'):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking_actions.PreviewCancellationResponse(
            refundable=True,
            refund_percent=1492,
            policy_summary='policy_summary_value',
        )
        response = client.preview_cancellation(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        request = booking_actions.PreviewCancellationRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking_actions.PreviewCancellationResponse)
    assert response.refundable is True
    assert response.refund_percent == 1492
    assert response.policy_summary == 'policy_summary_value'


def test_preview_cancellation_non_empty_request_with_auto_populated_field():
    # This test is a coverage failsafe to make sure that UUID4 fields are
    # automatically populated, according to AIP-4235, with non-empty requests.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport='grpc',
    )

    # Populate all string fields in the request which are not UUID4
    # since we want to check that UUID4 are populated automatically
    # if they meet the requirements of AIP 4235.
    request = booking_actions.PreviewCancellationRequest(
        name='name_value',
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        call.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client.preview_cancellation(request=request)
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.PreviewCancellationRequest(
            name='name_value',
        )
        assert args[0] == request_msg

def test_preview_cancellation_use_cached_wrapped_rpc():
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method.wrap_method") as wrapper_fn:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport="grpc",
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._transport.preview_cancellation in client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.Mock()
        mock_rpc.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client._transport._wrapped_methods[client._transport.preview_cancellation] = mock_rpc
        request = {}
        client.preview_cancellation(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        client.preview_cancellation(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
async def test_preview_cancellation_async_use_cached_wrapped_rpc(transport: str = "grpc_asyncio"):
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method_async.wrap_method") as wrapper_fn:
        client = BookingServiceAsyncClient(
            credentials=async_anonymous_credentials(),
            transport=transport,
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._client._transport.preview_cancellation in client._client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.AsyncMock()
        mock_rpc.return_value = mock.Mock()
        client._client._transport._wrapped_methods[client._client._transport.preview_cancellation] = mock_rpc

        request = {}
        await client.preview_cancellation(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        await client.preview_cancellation(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
@pytest.mark.parametrize("request_type", [
  booking_actions.PreviewCancellationRequest(),
  {},
])
async def test_preview_cancellation_async(request_type, transport: str = 'grpc_asyncio'):
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value =grpc_helpers_async.FakeUnaryUnaryCall(booking_actions.PreviewCancellationResponse(
            refundable=True,
            refund_percent=1492,
            policy_summary='policy_summary_value',
        ))
        response = await client.preview_cancellation(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        request = booking_actions.PreviewCancellationRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking_actions.PreviewCancellationResponse)
    assert response.refundable is True
    assert response.refund_percent == 1492
    assert response.policy_summary == 'policy_summary_value'

def test_preview_cancellation_field_headers():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.PreviewCancellationRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        call.return_value = booking_actions.PreviewCancellationResponse()
        client.preview_cancellation(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


@pytest.mark.asyncio
async def test_preview_cancellation_field_headers_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.PreviewCancellationRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking_actions.PreviewCancellationResponse())
        await client.preview_cancellation(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


def test_preview_cancellation_flattened():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking_actions.PreviewCancellationResponse()
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        client.preview_cancellation(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val


def test_preview_cancellation_flattened_error():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        client.preview_cancellation(
            booking_actions.PreviewCancellationRequest(),
            name='name_value',
        )

@pytest.mark.asyncio
async def test_preview_cancellation_flattened_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking_actions.PreviewCancellationResponse()

        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking_actions.PreviewCancellationResponse())
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        response = await client.preview_cancellation(
            name='name_value',
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val

@pytest.mark.asyncio
async def test_preview_cancellation_flattened_error_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        await client.preview_cancellation(
            booking_actions.PreviewCancellationRequest(),
            name='name_value',
        )


@pytest.mark.parametrize("request_type", [
  booking_actions.RescheduleBookingRequest(),
  {},
])
def test_reschedule_booking(request_type, transport: str = 'grpc'):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        )
        response = client.reschedule_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        request = booking_actions.RescheduleBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'


def test_reschedule_booking_non_empty_request_with_auto_populated_field():
    # This test is a coverage failsafe to make sure that UUID4 fields are
    # automatically populated, according to AIP-4235, with non-empty requests.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport='grpc',
    )

    # Populate all string fields in the request which are not UUID4
    # since we want to check that UUID4 are populated automatically
    # if they meet the requirements of AIP 4235.
    request = booking_actions.RescheduleBookingRequest(
        name='name_value',
        offering='offering_value',
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        call.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client.reschedule_booking(request=request)
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.RescheduleBookingRequest(
            name='name_value',
            offering='offering_value',
        )
        assert args[0] == request_msg

def test_reschedule_booking_use_cached_wrapped_rpc():
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method.wrap_method") as wrapper_fn:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport="grpc",
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._transport.reschedule_booking in client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.Mock()
        mock_rpc.return_value.name = "foo" # operation_request.operation in compute client(s) expect a string.
        client._transport._wrapped_methods[client._transport.reschedule_booking] = mock_rpc
        request = {}
        client.reschedule_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        client.reschedule_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
async def test_reschedule_booking_async_use_cached_wrapped_rpc(transport: str = "grpc_asyncio"):
    # Clients should use _prep_wrapped_messages to create cached wrapped rpcs,
    # instead of constructing them on each call
    with mock.patch("google.api_core.gapic_v1.method_async.wrap_method") as wrapper_fn:
        client = BookingServiceAsyncClient(
            credentials=async_anonymous_credentials(),
            transport=transport,
        )

        # Should wrap all calls on client creation
        assert wrapper_fn.call_count > 0
        wrapper_fn.reset_mock()

        # Ensure method has been cached
        assert client._client._transport.reschedule_booking in client._client._transport._wrapped_methods

        # Replace cached wrapped function with mock
        mock_rpc = mock.AsyncMock()
        mock_rpc.return_value = mock.Mock()
        client._client._transport._wrapped_methods[client._client._transport.reschedule_booking] = mock_rpc

        request = {}
        await client.reschedule_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert mock_rpc.call_count == 1

        await client.reschedule_booking(request)

        # Establish that a new wrapper was not created for this call
        assert wrapper_fn.call_count == 0
        assert mock_rpc.call_count == 2

@pytest.mark.asyncio
@pytest.mark.parametrize("request_type", [
  booking_actions.RescheduleBookingRequest(),
  {},
])
async def test_reschedule_booking_async(request_type, transport: str = 'grpc_asyncio'):
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport=transport,
    )

    # Everything is optional in proto3 as far as the runtime is concerned,
    # and we are mocking out the actual API, so just send an empty request.
    request = request_type

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value =grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        response = await client.reschedule_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        request = booking_actions.RescheduleBookingRequest()
        assert args[0] == request

    # Establish that the response is the type that we expect.
    assert isinstance(response, booking.Booking)
    assert response.name == 'name_value'
    assert response.resource == 'resource_value'
    assert response.offering == 'offering_value'
    assert response.customer == 'customer_value'
    assert response.units == 563
    assert response.assigned_unit == 'assigned_unit_value'
    assert response.state == enums.BookingState.BOOKING_STATE_PENDING_HOLD
    assert response.promo_code == 'promo_code_value'
    assert response.notes == 'notes_value'
    assert response.cancel_reason == enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER
    assert response.refund_percent == 1492
    assert response.etag == 'etag_value'

def test_reschedule_booking_field_headers():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.RescheduleBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.reschedule_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


@pytest.mark.asyncio
async def test_reschedule_booking_field_headers_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Any value that is part of the HTTP/1.1 URI should be sent as
    # a field header. Set these to a non-empty value.
    request = booking_actions.RescheduleBookingRequest()

    request.name = 'name_value'

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        await client.reschedule_booking(request)

        # Establish that the underlying gRPC stub method was called.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        assert args[0] == request

    # Establish that the field header was sent.
    _, _, kw = call.mock_calls[0]
    assert (
        'x-goog-request-params',
        'name=name_value',
    ) in kw['metadata']


def test_reschedule_booking_flattened():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        client.reschedule_booking(
            name='name_value',
            window=types_pb2.TimeWindow(start_time=timestamp_pb2.Timestamp(seconds=751)),
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls) == 1
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val
        arg = args[0].window
        mock_val = types_pb2.TimeWindow(start_time=timestamp_pb2.Timestamp(seconds=751))
        assert arg == mock_val


def test_reschedule_booking_flattened_error():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        client.reschedule_booking(
            booking_actions.RescheduleBookingRequest(),
            name='name_value',
            window=types_pb2.TimeWindow(start_time=timestamp_pb2.Timestamp(seconds=751)),
        )

@pytest.mark.asyncio
async def test_reschedule_booking_flattened_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Mock the actual call within the gRPC stub, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = booking.Booking()

        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking())
        # Call the method with a truthy value for each flattened field,
        # using the keyword arguments to the method.
        response = await client.reschedule_booking(
            name='name_value',
            window=types_pb2.TimeWindow(start_time=timestamp_pb2.Timestamp(seconds=751)),
        )

        # Establish that the underlying call was made with the expected
        # request object values.
        assert len(call.mock_calls)
        _, args, _ = call.mock_calls[0]
        arg = args[0].name
        mock_val = 'name_value'
        assert arg == mock_val
        arg = args[0].window
        mock_val = types_pb2.TimeWindow(start_time=timestamp_pb2.Timestamp(seconds=751))
        assert arg == mock_val

@pytest.mark.asyncio
async def test_reschedule_booking_flattened_error_async():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
    )

    # Attempting to call a method with both a request object and flattened
    # fields is an error.
    with pytest.raises(ValueError):
        await client.reschedule_booking(
            booking_actions.RescheduleBookingRequest(),
            name='name_value',
            window=types_pb2.TimeWindow(start_time=timestamp_pb2.Timestamp(seconds=751)),
        )


def test_credentials_transport_error():
    # It is an error to provide credentials and a transport instance.
    transport = transports.BookingServiceGrpcTransport(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    with pytest.raises(ValueError):
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport=transport,
        )

    # It is an error to provide a credentials file and a transport instance.
    transport = transports.BookingServiceGrpcTransport(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    with pytest.raises(ValueError):
        client = BookingServiceClient(
            client_options={"credentials_file": "credentials.json"},
            transport=transport,
        )

    # It is an error to provide an api_key and a transport instance.
    transport = transports.BookingServiceGrpcTransport(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    options = client_options.ClientOptions()
    options.api_key = "api_key"
    with pytest.raises(ValueError):
        client = BookingServiceClient(
            client_options=options,
            transport=transport,
        )

    # It is an error to provide an api_key and a credential.
    options = client_options.ClientOptions()
    options.api_key = "api_key"
    with pytest.raises(ValueError):
        client = BookingServiceClient(
            client_options=options,
            credentials=ga_credentials.AnonymousCredentials()
        )

    # It is an error to provide scopes and a transport instance.
    transport = transports.BookingServiceGrpcTransport(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    with pytest.raises(ValueError):
        client = BookingServiceClient(
            client_options={"scopes": ["1", "2"]},
            transport=transport,
        )


def test_transport_instance():
    # A client may be instantiated with a custom transport instance.
    transport = transports.BookingServiceGrpcTransport(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    client = BookingServiceClient(transport=transport)
    assert client.transport is transport

def test_transport_get_channel():
    # A client may be instantiated with a custom transport instance.
    transport = transports.BookingServiceGrpcTransport(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    channel = transport.grpc_channel
    assert channel

    transport = transports.BookingServiceGrpcAsyncIOTransport(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    channel = transport.grpc_channel
    assert channel

@pytest.mark.parametrize("transport_class", [
    transports.BookingServiceGrpcTransport,
    transports.BookingServiceGrpcAsyncIOTransport,
])
def test_transport_adc(transport_class):
    # Test default credentials are used if not provided.
    with mock.patch.object(google.auth, 'default') as adc:
        adc.return_value = (ga_credentials.AnonymousCredentials(), None)
        transport_class()
        adc.assert_called_once()

def test_transport_kind_grpc():
    transport = BookingServiceClient.get_transport_class("grpc")(
        credentials=ga_credentials.AnonymousCredentials()
    )
    assert transport.kind == "grpc"


def test_initialize_client_w_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc"
    )
    assert client is not None


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
def test_create_booking_empty_call_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.create_booking),
            '__call__') as call:
        call.return_value = fb_booking.Booking()
        client.create_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.CreateBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
def test_get_booking_empty_call_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.get_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.GetBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
def test_list_bookings_empty_call_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__') as call:
        call.return_value = booking_messages.ListBookingsResponse()
        client.list_bookings(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.ListBookingsRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
def test_confirm_booking_empty_call_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.confirm_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.ConfirmBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
def test_cancel_booking_empty_call_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.cancel_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.CancelBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
def test_preview_cancellation_empty_call_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        call.return_value = booking_actions.PreviewCancellationResponse()
        client.preview_cancellation(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.PreviewCancellationRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
def test_reschedule_booking_empty_call_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        call.return_value = booking.Booking()
        client.reschedule_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.RescheduleBookingRequest()
        assert args[0] == request_msg


def test_transport_kind_grpc_asyncio():
    transport = BookingServiceAsyncClient.get_transport_class("grpc_asyncio")(
        credentials=async_anonymous_credentials()
    )
    assert transport.kind == "grpc_asyncio"


def test_initialize_client_w_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio"
    )
    assert client is not None


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
@pytest.mark.asyncio
async def test_create_booking_empty_call_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.create_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(fb_booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        await client.create_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.CreateBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
@pytest.mark.asyncio
async def test_get_booking_empty_call_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.get_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        await client.get_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.GetBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
@pytest.mark.asyncio
async def test_list_bookings_empty_call_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.list_bookings),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking_messages.ListBookingsResponse(
            next_page_token='next_page_token_value',
        ))
        await client.list_bookings(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_messages.ListBookingsRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
@pytest.mark.asyncio
async def test_confirm_booking_empty_call_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.confirm_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        await client.confirm_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.ConfirmBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
@pytest.mark.asyncio
async def test_cancel_booking_empty_call_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.cancel_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        await client.cancel_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.CancelBookingRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
@pytest.mark.asyncio
async def test_preview_cancellation_empty_call_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.preview_cancellation),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking_actions.PreviewCancellationResponse(
            refundable=True,
            refund_percent=1492,
            policy_summary='policy_summary_value',
        ))
        await client.preview_cancellation(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.PreviewCancellationRequest()
        assert args[0] == request_msg


# This test is a coverage failsafe to make sure that totally empty calls,
# i.e. request == None and no flattened fields passed, work.
@pytest.mark.asyncio
async def test_reschedule_booking_empty_call_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio",
    )

    # Mock the actual call, and fake the request.
    with mock.patch.object(
            type(client.transport.reschedule_booking),
            '__call__') as call:
        # Designate an appropriate return value for the call.
        call.return_value = grpc_helpers_async.FakeUnaryUnaryCall(booking.Booking(
            name='name_value',
            resource='resource_value',
            offering='offering_value',
            customer='customer_value',
            units=563,
            assigned_unit='assigned_unit_value',
            state=enums.BookingState.BOOKING_STATE_PENDING_HOLD,
            promo_code='promo_code_value',
            notes='notes_value',
            cancel_reason=enums.CancelReason.CANCEL_REASON_REQUESTED_BY_CUSTOMER,
            refund_percent=1492,
            etag='etag_value',
        ))
        await client.reschedule_booking(request=None)

        # Establish that the underlying stub method was called.
        call.assert_called()
        _, args, _ = call.mock_calls[0]
        request_msg = booking_actions.RescheduleBookingRequest()
        assert args[0] == request_msg


def test_transport_grpc_default():
    # A client should use the gRPC transport by default.
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
    )
    assert isinstance(
        client.transport,
        transports.BookingServiceGrpcTransport,
    )

def test_booking_service_base_transport_error():
    # Passing both a credentials object and credentials_file should raise an error
    with pytest.raises(core_exceptions.DuplicateCredentialArgs):
        transport = transports.BookingServiceTransport(
            credentials=ga_credentials.AnonymousCredentials(),
            credentials_file="credentials.json"
        )


def test_booking_service_base_transport():
    # Instantiate the base transport.
    with mock.patch('freebusy.booking_v1.services.booking_service.transports.BookingServiceTransport.__init__') as Transport:
        Transport.return_value = None
        transport = transports.BookingServiceTransport(
            credentials=ga_credentials.AnonymousCredentials(),
        )

    # Every method on the transport should just blindly
    # raise NotImplementedError.
    methods = (
        'create_booking',
        'get_booking',
        'list_bookings',
        'confirm_booking',
        'cancel_booking',
        'preview_cancellation',
        'reschedule_booking',
    )
    for method in methods:
        with pytest.raises(NotImplementedError):
            getattr(transport, method)(request=object())

    with pytest.raises(NotImplementedError):
        transport.close()

    # Catch all for all remaining methods and properties
    remainder = [
        'kind',
    ]
    for r in remainder:
        with pytest.raises(NotImplementedError):
            getattr(transport, r)()


def test_booking_service_base_transport_with_credentials_file():
    # Instantiate the base transport with a credentials file
    with mock.patch.object(google.auth, 'load_credentials_from_file', autospec=True) as load_creds, mock.patch('freebusy.booking_v1.services.booking_service.transports.BookingServiceTransport._prep_wrapped_messages') as Transport:
        Transport.return_value = None
        load_creds.return_value = (ga_credentials.AnonymousCredentials(), None)
        transport = transports.BookingServiceTransport(
            credentials_file="credentials.json",
            quota_project_id="octopus",
        )
        load_creds.assert_called_once_with("credentials.json",
            scopes=None,
            default_scopes=(
),
            quota_project_id="octopus",
        )


def test_booking_service_base_transport_with_adc():
    # Test the default credentials are used if credentials and credentials_file are None.
    with mock.patch.object(google.auth, 'default', autospec=True) as adc, mock.patch('freebusy.booking_v1.services.booking_service.transports.BookingServiceTransport._prep_wrapped_messages') as Transport:
        Transport.return_value = None
        adc.return_value = (ga_credentials.AnonymousCredentials(), None)
        transport = transports.BookingServiceTransport()
        adc.assert_called_once()


def test_booking_service_auth_adc():
    # If no credentials are provided, we should use ADC credentials.
    with mock.patch.object(google.auth, 'default', autospec=True) as adc:
        adc.return_value = (ga_credentials.AnonymousCredentials(), None)
        BookingServiceClient()
        adc.assert_called_once_with(
            scopes=None,
            default_scopes=(
),
            quota_project_id=None,
        )


@pytest.mark.parametrize(
    "transport_class",
    [
        transports.BookingServiceGrpcTransport,
        transports.BookingServiceGrpcAsyncIOTransport,
    ],
)
def test_booking_service_transport_auth_adc(transport_class):
    # If credentials and host are not provided, the transport class should use
    # ADC credentials.
    with mock.patch.object(google.auth, 'default', autospec=True) as adc:
        adc.return_value = (ga_credentials.AnonymousCredentials(), None)
        transport_class(quota_project_id="octopus", scopes=["1", "2"])
        adc.assert_called_once_with(
            scopes=["1", "2"],
            default_scopes=(),
            quota_project_id="octopus",
        )


@pytest.mark.parametrize(
    "transport_class",
    [
        transports.BookingServiceGrpcTransport,
        transports.BookingServiceGrpcAsyncIOTransport,
    ],
)
def test_booking_service_transport_auth_gdch_credentials(transport_class):
    host = 'https://language.com'
    api_audience_tests = [None, 'https://language2.com']
    api_audience_expect = [host, 'https://language2.com']
    for t, e in zip(api_audience_tests, api_audience_expect):
        with mock.patch.object(google.auth, 'default', autospec=True) as adc:
            gdch_mock = mock.MagicMock()
            type(gdch_mock).with_gdch_audience = mock.PropertyMock(return_value=gdch_mock)
            adc.return_value = (gdch_mock, None)
            transport_class(host=host, api_audience=t)
            gdch_mock.with_gdch_audience.assert_called_once_with(
                e
            )


@pytest.mark.parametrize(
    "transport_class,grpc_helpers",
    [
        (transports.BookingServiceGrpcTransport, grpc_helpers),
        (transports.BookingServiceGrpcAsyncIOTransport, grpc_helpers_async)
    ],
)
def test_booking_service_transport_create_channel(transport_class, grpc_helpers):
    # If credentials and host are not provided, the transport class should use
    # ADC credentials.
    with mock.patch.object(google.auth, "default", autospec=True) as adc, mock.patch.object(
        grpc_helpers, "create_channel", autospec=True
    ) as create_channel:
        creds = ga_credentials.AnonymousCredentials()
        adc.return_value = (creds, None)
        transport_class(
            quota_project_id="octopus",
            scopes=["1", "2"]
        )

        create_channel.assert_called_with(
            "freebusy.ohtarnished.dev:443",
            credentials=creds,
            credentials_file=None,
            quota_project_id="octopus",
            default_scopes=(
),
            scopes=["1", "2"],
            default_host="freebusy.ohtarnished.dev",
            ssl_credentials=None,
            options=[
                ("grpc.max_send_message_length", -1),
                ("grpc.max_receive_message_length", -1),
            ],
        )


@pytest.mark.parametrize("transport_class", [transports.BookingServiceGrpcTransport, transports.BookingServiceGrpcAsyncIOTransport])
def test_booking_service_grpc_transport_client_cert_source_for_mtls(
    transport_class
):
    cred = ga_credentials.AnonymousCredentials()

    # Check ssl_channel_credentials is used if provided.
    with mock.patch.object(transport_class, "create_channel") as mock_create_channel:
        mock_ssl_channel_creds = mock.Mock()
        transport_class(
            host="squid.clam.whelk",
            credentials=cred,
            ssl_channel_credentials=mock_ssl_channel_creds
        )
        mock_create_channel.assert_called_once_with(
            "squid.clam.whelk:443",
            credentials=cred,
            credentials_file=None,
            scopes=None,
            ssl_credentials=mock_ssl_channel_creds,
            quota_project_id=None,
            options=[
                ("grpc.max_send_message_length", -1),
                ("grpc.max_receive_message_length", -1),
            ],
        )

    # Check if ssl_channel_credentials is not provided, then client_cert_source_for_mtls
    # is used.
    with mock.patch.object(transport_class, "create_channel", return_value=mock.Mock()):
        with mock.patch("grpc.ssl_channel_credentials") as mock_ssl_cred:
            transport_class(
                credentials=cred,
                client_cert_source_for_mtls=client_cert_source_callback
            )
            expected_cert, expected_key = client_cert_source_callback()
            mock_ssl_cred.assert_called_once_with(
                certificate_chain=expected_cert,
                private_key=expected_key
            )


@pytest.mark.parametrize("transport_name", [
    "grpc",
    "grpc_asyncio",
])
def test_booking_service_host_no_port(transport_name):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        client_options=client_options.ClientOptions(api_endpoint='freebusy.ohtarnished.dev'),
         transport=transport_name,
    )
    assert client.transport._host == (
        'freebusy.ohtarnished.dev:443'
    )

@pytest.mark.parametrize("transport_name", [
    "grpc",
    "grpc_asyncio",
])
def test_booking_service_host_with_port(transport_name):
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        client_options=client_options.ClientOptions(api_endpoint='freebusy.ohtarnished.dev:8000'),
        transport=transport_name,
    )
    assert client.transport._host == (
        'freebusy.ohtarnished.dev:8000'
    )

def test_booking_service_grpc_transport_channel():
    channel = grpc.secure_channel('http://localhost/', grpc.local_channel_credentials())

    # Check that channel is used if provided.
    transport = transports.BookingServiceGrpcTransport(
        host="squid.clam.whelk",
        channel=channel,
    )
    assert transport.grpc_channel == channel
    assert transport._host == "squid.clam.whelk:443"
    assert transport._ssl_channel_credentials == None


def test_booking_service_grpc_asyncio_transport_channel():
    channel = aio.secure_channel('http://localhost/', grpc.local_channel_credentials())

    # Check that channel is used if provided.
    transport = transports.BookingServiceGrpcAsyncIOTransport(
        host="squid.clam.whelk",
        channel=channel,
    )
    assert transport.grpc_channel == channel
    assert transport._host == "squid.clam.whelk:443"
    assert transport._ssl_channel_credentials == None


# Remove this test when deprecated arguments (api_mtls_endpoint, client_cert_source) are
# removed from grpc/grpc_asyncio transport constructor.
@pytest.mark.filterwarnings("ignore::FutureWarning")
@pytest.mark.parametrize("transport_class", [transports.BookingServiceGrpcTransport, transports.BookingServiceGrpcAsyncIOTransport])
def test_booking_service_transport_channel_mtls_with_client_cert_source(
    transport_class
):
    with mock.patch("grpc.ssl_channel_credentials", autospec=True) as grpc_ssl_channel_cred:
        with mock.patch.object(transport_class, "create_channel") as grpc_create_channel:
            mock_ssl_cred = mock.Mock()
            grpc_ssl_channel_cred.return_value = mock_ssl_cred

            mock_grpc_channel = mock.Mock()
            grpc_create_channel.return_value = mock_grpc_channel

            cred = ga_credentials.AnonymousCredentials()
            with pytest.warns(DeprecationWarning):
                with mock.patch.object(google.auth, 'default') as adc:
                    adc.return_value = (cred, None)
                    transport = transport_class(
                        host="squid.clam.whelk",
                        api_mtls_endpoint="mtls.squid.clam.whelk",
                        client_cert_source=client_cert_source_callback,
                    )
                    adc.assert_called_once()

            grpc_ssl_channel_cred.assert_called_once_with(
                certificate_chain=b"cert bytes", private_key=b"key bytes"
            )
            grpc_create_channel.assert_called_once_with(
                "mtls.squid.clam.whelk:443",
                credentials=cred,
                credentials_file=None,
                scopes=None,
                ssl_credentials=mock_ssl_cred,
                quota_project_id=None,
                options=[
                    ("grpc.max_send_message_length", -1),
                    ("grpc.max_receive_message_length", -1),
                ],
            )
            assert transport.grpc_channel == mock_grpc_channel
            assert transport._ssl_channel_credentials == mock_ssl_cred


# Remove this test when deprecated arguments (api_mtls_endpoint, client_cert_source) are
# removed from grpc/grpc_asyncio transport constructor.
@pytest.mark.parametrize("transport_class", [transports.BookingServiceGrpcTransport, transports.BookingServiceGrpcAsyncIOTransport])
def test_booking_service_transport_channel_mtls_with_adc(
    transport_class
):
    mock_ssl_cred = mock.Mock()
    with mock.patch.multiple(
        "google.auth.transport.grpc.SslCredentials",
        __init__=mock.Mock(return_value=None),
        ssl_credentials=mock.PropertyMock(return_value=mock_ssl_cred),
    ):
        with mock.patch.object(transport_class, "create_channel") as grpc_create_channel:
            mock_grpc_channel = mock.Mock()
            grpc_create_channel.return_value = mock_grpc_channel
            mock_cred = mock.Mock()

            with pytest.warns(DeprecationWarning):
                transport = transport_class(
                    host="squid.clam.whelk",
                    credentials=mock_cred,
                    api_mtls_endpoint="mtls.squid.clam.whelk",
                    client_cert_source=None,
                )

            grpc_create_channel.assert_called_once_with(
                "mtls.squid.clam.whelk:443",
                credentials=mock_cred,
                credentials_file=None,
                scopes=None,
                ssl_credentials=mock_ssl_cred,
                quota_project_id=None,
                options=[
                    ("grpc.max_send_message_length", -1),
                    ("grpc.max_receive_message_length", -1),
                ],
            )
            assert transport.grpc_channel == mock_grpc_channel


def test_booking_path():
    booking = "squid"
    expected = "bookings/{booking}".format(booking=booking, )
    actual = BookingServiceClient.booking_path(booking)
    assert expected == actual


def test_parse_booking_path():
    expected = {
        "booking": "clam",
    }
    path = BookingServiceClient.booking_path(**expected)

    # Check that the path construction is reversible.
    actual = BookingServiceClient.parse_booking_path(path)
    assert expected == actual

def test_common_billing_account_path():
    billing_account = "whelk"
    expected = "billingAccounts/{billing_account}".format(billing_account=billing_account, )
    actual = BookingServiceClient.common_billing_account_path(billing_account)
    assert expected == actual


def test_parse_common_billing_account_path():
    expected = {
        "billing_account": "octopus",
    }
    path = BookingServiceClient.common_billing_account_path(**expected)

    # Check that the path construction is reversible.
    actual = BookingServiceClient.parse_common_billing_account_path(path)
    assert expected == actual

def test_common_folder_path():
    folder = "oyster"
    expected = "folders/{folder}".format(folder=folder, )
    actual = BookingServiceClient.common_folder_path(folder)
    assert expected == actual


def test_parse_common_folder_path():
    expected = {
        "folder": "nudibranch",
    }
    path = BookingServiceClient.common_folder_path(**expected)

    # Check that the path construction is reversible.
    actual = BookingServiceClient.parse_common_folder_path(path)
    assert expected == actual

def test_common_organization_path():
    organization = "cuttlefish"
    expected = "organizations/{organization}".format(organization=organization, )
    actual = BookingServiceClient.common_organization_path(organization)
    assert expected == actual


def test_parse_common_organization_path():
    expected = {
        "organization": "mussel",
    }
    path = BookingServiceClient.common_organization_path(**expected)

    # Check that the path construction is reversible.
    actual = BookingServiceClient.parse_common_organization_path(path)
    assert expected == actual

def test_common_project_path():
    project = "winkle"
    expected = "projects/{project}".format(project=project, )
    actual = BookingServiceClient.common_project_path(project)
    assert expected == actual


def test_parse_common_project_path():
    expected = {
        "project": "nautilus",
    }
    path = BookingServiceClient.common_project_path(**expected)

    # Check that the path construction is reversible.
    actual = BookingServiceClient.parse_common_project_path(path)
    assert expected == actual

def test_common_location_path():
    project = "scallop"
    location = "abalone"
    expected = "projects/{project}/locations/{location}".format(project=project, location=location, )
    actual = BookingServiceClient.common_location_path(project, location)
    assert expected == actual


def test_parse_common_location_path():
    expected = {
        "project": "squid",
        "location": "clam",
    }
    path = BookingServiceClient.common_location_path(**expected)

    # Check that the path construction is reversible.
    actual = BookingServiceClient.parse_common_location_path(path)
    assert expected == actual


def test_client_with_default_client_info():
    client_info = gapic_v1.client_info.ClientInfo()

    with mock.patch.object(transports.BookingServiceTransport, '_prep_wrapped_messages') as prep:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            client_info=client_info,
        )
        prep.assert_called_once_with(client_info)

    with mock.patch.object(transports.BookingServiceTransport, '_prep_wrapped_messages') as prep:
        transport_class = BookingServiceClient.get_transport_class()
        transport = transport_class(
            credentials=ga_credentials.AnonymousCredentials(),
            client_info=client_info,
        )
        prep.assert_called_once_with(client_info)


def test_transport_close_grpc():
    client = BookingServiceClient(
        credentials=ga_credentials.AnonymousCredentials(),
        transport="grpc"
    )
    with mock.patch.object(type(getattr(client.transport, "_grpc_channel")), "close") as close:
        with client:
            close.assert_not_called()
        close.assert_called_once()


@pytest.mark.asyncio
async def test_transport_close_grpc_asyncio():
    client = BookingServiceAsyncClient(
        credentials=async_anonymous_credentials(),
        transport="grpc_asyncio"
    )
    with mock.patch.object(type(getattr(client.transport, "_grpc_channel")), "close") as close:
        async with client:
            close.assert_not_called()
        close.assert_called_once()


def test_client_ctx():
    transports = [
        'grpc',
    ]
    for transport in transports:
        client = BookingServiceClient(
            credentials=ga_credentials.AnonymousCredentials(),
            transport=transport
        )
        # Test client calls underlying transport.
        with mock.patch.object(type(client.transport), "close") as close:
            close.assert_not_called()
            with client:
                pass
            close.assert_called()

@pytest.mark.parametrize("client_class,transport_class", [
    (BookingServiceClient, transports.BookingServiceGrpcTransport),
    (BookingServiceAsyncClient, transports.BookingServiceGrpcAsyncIOTransport),
])
def test_api_key_credentials(client_class, transport_class):
    with mock.patch.object(
        google.auth._default, "get_api_key_credentials", create=True
    ) as get_api_key_credentials:
        mock_cred = mock.Mock()
        get_api_key_credentials.return_value = mock_cred
        options = client_options.ClientOptions()
        options.api_key = "api_key"
        with mock.patch.object(transport_class, "__init__") as patched:
            patched.return_value = None
            client = client_class(client_options=options)
            patched.assert_called_once_with(
                credentials=mock_cred,
                credentials_file=None,
                host=client._DEFAULT_ENDPOINT_TEMPLATE.format(UNIVERSE_DOMAIN=client._DEFAULT_UNIVERSE),
                scopes=None,
                client_cert_source_for_mtls=None,
                quota_project_id=None,
                client_info=transports.base.DEFAULT_CLIENT_INFO,
                always_use_jwt_access=True,
                api_audience=None,
            )
