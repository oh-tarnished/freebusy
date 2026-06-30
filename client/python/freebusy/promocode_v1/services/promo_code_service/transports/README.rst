
transport inheritance structure
_______________________________

``PromoCodeServiceTransport`` is the ABC for all transports.

- public child ``PromoCodeServiceGrpcTransport`` for sync gRPC transport (defined in ``grpc.py``).
- public child ``PromoCodeServiceGrpcAsyncIOTransport`` for async gRPC transport (defined in ``grpc_asyncio.py``).
- private child ``_BasePromoCodeServiceRestTransport`` for base REST transport with inner classes ``_BaseMETHOD`` (defined in ``rest_base.py``).
- public child ``PromoCodeServiceRestTransport`` for sync REST transport with inner classes ``METHOD`` derived from the parent's corresponding ``_BaseMETHOD`` classes (defined in ``rest.py``).
