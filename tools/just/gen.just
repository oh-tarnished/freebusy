# Code generation: protobuf -> Go, OpenAPI, ORM, TypeScript, docs.
# Invoked as `just gen <recipe>` (e.g. `just gen orm`). `just gen` runs `all`.
# Recipes run from the repo root so buf/template paths resolve unchanged.
set working-directory := '../..'

# Regenerate everything: Go, OpenAPI (+merge+envoy), ORM, TypeScript, docs.
[group('codegen')]
all: deps go openapi envoy orm docs

# Update buf module dependencies.
[group('codegen')]
deps:
    buf dep update

# Generate the Go stubs.
[group('codegen')]
go:
    buf generate --template tools/protobuf/buf/go.buf.gen.yaml

# Generate the OpenAPI v3 specs and merge them into openapiv3-spec.yaml.
[group('codegen')]
openapi:
    buf generate --template tools/protobuf/buf/openapiv3.buf.gen.yaml
    go run ./tools/protobuf/openapi

# Generate the Envoy grpc-web proxy config (envoy/launch.yaml) from the OpenAPI
# specs: one route per gRPC service -> the freebusy backend.
[group('codegen')]
envoy:
    go run ./tools/protobuf/envoy

# Generate the ORM schemas (prisma + gorm). Requires protoc-gen-protorm on PATH.
[group('codegen')]
orm:
    buf generate --template tools/protobuf/buf/orm.buf.gen.yaml

# Generate protobuf docs (per-module + root README.md) via protodoc.
[group('codegen')]
docs:
    go run ./tools/protobuf/docs protobuf
