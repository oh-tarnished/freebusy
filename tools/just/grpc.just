# Exercise the gRPC API via reflection (server must be running; needs grpcurl).
# Invoked as `just grpc <recipe>`.
set working-directory := '../..'

grpc_port := env_var_or_default("FREEBUSY_GRPC_PORT", "50051")

# List the gRPC services exposed via reflection.
[group('api')]
list:
    grpcurl -plaintext localhost:{{ grpc_port }} list

# gRPC health check.
[group('api')]
health:
    grpcurl -plaintext localhost:{{ grpc_port }} grpc.health.v1.Health/Check
