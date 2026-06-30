
transport inheritance structure
_______________________________

``ResourceServiceTransport`` is the ABC for all transports.

- public child ``ResourceServiceGrpcTransport`` for sync gRPC transport (defined in ``grpc.py``).
- public child ``ResourceServiceGrpcAsyncIOTransport`` for async gRPC transport (defined in ``grpc_asyncio.py``).
- private child ``_BaseResourceServiceRestTransport`` for base REST transport with inner classes ``_BaseMETHOD`` (defined in ``rest_base.py``).
- public child ``ResourceServiceRestTransport`` for sync REST transport with inner classes ``METHOD`` derived from the parent's corresponding ``_BaseMETHOD`` classes (defined in ``rest.py``).
