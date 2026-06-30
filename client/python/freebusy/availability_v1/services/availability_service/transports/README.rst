
transport inheritance structure
_______________________________

``AvailabilityServiceTransport`` is the ABC for all transports.

- public child ``AvailabilityServiceGrpcTransport`` for sync gRPC transport (defined in ``grpc.py``).
- public child ``AvailabilityServiceGrpcAsyncIOTransport`` for async gRPC transport (defined in ``grpc_asyncio.py``).
- private child ``_BaseAvailabilityServiceRestTransport`` for base REST transport with inner classes ``_BaseMETHOD`` (defined in ``rest_base.py``).
- public child ``AvailabilityServiceRestTransport`` for sync REST transport with inner classes ``METHOD`` derived from the parent's corresponding ``_BaseMETHOD`` classes (defined in ``rest.py``).
