
transport inheritance structure
_______________________________

``IdentityServiceTransport`` is the ABC for all transports.

- public child ``IdentityServiceGrpcTransport`` for sync gRPC transport (defined in ``grpc.py``).
- public child ``IdentityServiceGrpcAsyncIOTransport`` for async gRPC transport (defined in ``grpc_asyncio.py``).
- private child ``_BaseIdentityServiceRestTransport`` for base REST transport with inner classes ``_BaseMETHOD`` (defined in ``rest_base.py``).
- public child ``IdentityServiceRestTransport`` for sync REST transport with inner classes ``METHOD`` derived from the parent's corresponding ``_BaseMETHOD`` classes (defined in ``rest.py``).
