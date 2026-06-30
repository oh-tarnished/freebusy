
transport inheritance structure
_______________________________

``OrganisationServiceTransport`` is the ABC for all transports.

- public child ``OrganisationServiceGrpcTransport`` for sync gRPC transport (defined in ``grpc.py``).
- public child ``OrganisationServiceGrpcAsyncIOTransport`` for async gRPC transport (defined in ``grpc_asyncio.py``).
- private child ``_BaseOrganisationServiceRestTransport`` for base REST transport with inner classes ``_BaseMETHOD`` (defined in ``rest_base.py``).
- public child ``OrganisationServiceRestTransport`` for sync REST transport with inner classes ``METHOD`` derived from the parent's corresponding ``_BaseMETHOD`` classes (defined in ``rest.py``).
