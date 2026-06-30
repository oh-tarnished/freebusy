
transport inheritance structure
_______________________________

``BookingServiceTransport`` is the ABC for all transports.

- public child ``BookingServiceGrpcTransport`` for sync gRPC transport (defined in ``grpc.py``).
- public child ``BookingServiceGrpcAsyncIOTransport`` for async gRPC transport (defined in ``grpc_asyncio.py``).
- private child ``_BaseBookingServiceRestTransport`` for base REST transport with inner classes ``_BaseMETHOD`` (defined in ``rest_base.py``).
- public child ``BookingServiceRestTransport`` for sync REST transport with inner classes ``METHOD`` derived from the parent's corresponding ``_BaseMETHOD`` classes (defined in ``rest.py``).
