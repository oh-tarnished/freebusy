default:
    @just --list

# run prek hooks (--all-files to force proto lint/generate when no .proto files staged)
prek *args:
    # Auto-fix trailing whitespace and EOF before running hooks
    python3 tools/protobuf/fix-precommit.py .
    prek run --all-files {{ args }}

# Generate protobuf docs (per-module + root README.md) via protodoc.
docs:
    go run ./tools/protobuf/docs protobuf

# Generate the ORM schemas (prisma + gorm). Requires protoc-gen-protorm on PATH (brew/go install).
orm:
    buf generate --template tools/protobuf/buf/orm.buf.gen.yaml

# Merge the per-service generated OpenAPI specs into a single openapiv3-spec.yaml.
openapi-merge:
    go run ./tools/protobuf/openapi

# Generate the Envoy grpc-web proxy config (envoy/launch.yaml) from the
# per-service OpenAPI specs: one route per gRPC service -> the freebusy backend.
envoy-gen:
    go run ./tools/protobuf/envoy

# Start the Envoy grpc-web proxy (envoy/docker-compose.yaml). Fronts the backend
# on :8080; admin UI on http://localhost:9901. Run the freebusy backend first.
envoy-up:
    docker compose -f envoy/docker-compose.yaml up -d

# Stop the Envoy grpc-web proxy.
envoy-down:
    docker compose -f envoy/docker-compose.yaml down

# Regenerate Hasura DDN metadata after a DB schema change (introspect connector,
# rebuild supergraph, restart engine, refresh generateql client). Pass a domain
# to drop+regenerate its changed models, e.g. `just hasura-regen promocode`.
hasura-regen domain="":
    tools/db/hasura/regen.sh {{ domain }}

# generated protos for languages. if language is not specified, generates for all supported languages.
generate language="all" descriptors="true":
    #!/usr/bin/env sh
    set -e
    echo "==> Updating buf deps..."
    buf dep update
    echo "==> Language Specific Buf generate..."
    if [ "{{language}}" = "all" ]; then
        buf generate --template tools/protobuf/buf/go.buf.gen.yaml
        buf generate --template tools/protobuf/buf/openapiv3.buf.gen.yaml
        echo "==> OpenAPI merge..."
        go run ./tools/protobuf/openapi
        echo "==> Envoy config generate..."
        go run ./tools/protobuf/envoy
        echo "==> ORM generate (prisma + gorm)..."
        buf generate --template tools/protobuf/buf/orm.buf.gen.yaml
        echo "==> TypeScript generate..."
        buf generate --template tools/protobuf/buf/typescript.buf.gen.yaml
        echo "==> Docs generate..."
        go run ./tools/protobuf/docs protobuf
    elif [ "{{language}}" = "orm" ]; then
        buf generate --template tools/protobuf/buf/orm.buf.gen.yaml
    elif [ "{{language}}" = "docs" ]; then
        go run ./tools/protobuf/docs protobuf
    elif [ "{{language}}" = "openapiv3" ]; then
        buf generate --template tools/protobuf/buf/openapiv3.buf.gen.yaml
        echo "==> OpenAPI merge..."
        go run ./tools/protobuf/openapi
        echo "==> Envoy config generate..."
        go run ./tools/protobuf/envoy
    elif [ -f "tools/protobuf/buf/{{language}}.buf.gen.yaml" ]; then
        buf generate --template tools/protobuf/buf/{{language}}.buf.gen.yaml
    else
        echo "Error: Template for {{language}} not found!"
        echo "Available languages: go, openapiv3, orm, docs"
        exit 1
    fi

# ============================================================================
# Application (server) — run, build, test, and exercise the gRPC API.
# Database settings (provider + connection) come from config: the embedded
# config/freebusy.release.toml defaults, overlaid by config/freebusy.dev.toml
# for local development. Switch backends by editing [database].provider there.
# ============================================================================

# Server ports (exported so the runtime picks them up). Override on the CLI or
# via the matching FREEBUSY_* env var.
grpc_port := env_var_or_default("FREEBUSY_GRPC_PORT", "50051")
http_port := env_var_or_default("FREEBUSY_HTTP_PORT", "8080")

export FREEBUSY_GRPC_PORT := grpc_port
export FREEBUSY_HTTP_PORT := http_port

# Run the server (DB provider + connection come from config; dev overlay applies).
run:
    go run .

# DEV ONLY: create the Postgres schemas + AutoMigrate every generated model.
# Run once before `just run`. Uses the same config as the server.
migrate:
    go run ./cmd/migrate

# Dev convenience: migrate the schema, then run the server.
dev: migrate run

# Compile everything.
build:
    go build ./...

# Vet everything.
vet:
    go vet ./...

# Run the unit tests (pure-logic packages; no database required).
test:
    go test ./...

# Verbose unit tests for the pulse-free pure-logic packages.
test-unit:
    go test -v ./internal/discount/ ./internal/database/repository/ ./internal/service/gorm/

# Format all Go files.
fmt:
    gofmt -w .

# Tidy module dependencies.
gotidy:
    go mod tidy

# CI-style gate: build, vet, and test.
check: build vet test

# Remove the Go build/test caches.
clean:
    go clean -cache -testcache

# --- exercise the API (server must be running; needs grpcurl) ---

# List the gRPC services exposed via reflection.
grpc-list:
    grpcurl -plaintext localhost:{{grpc_port}} list

# gRPC health check.
grpc-health:
    grpcurl -plaintext localhost:{{grpc_port}} grpc.health.v1.Health/Check

# List promo codes.
promo-list:
    grpcurl -plaintext -d '{}' localhost:{{grpc_port}} freebusy.promocode.v1.PromoCodeService/ListPromoCodes

# Create a sample 25%-off promo code.
promo-create:
    grpcurl -plaintext -d '{"promo_code":{"code":"SUMMER25","display_name":"Summer Sale","discount_type":"DISCOUNT_TYPE_PERCENTAGE","percent_off":25}}' \
        localhost:{{grpc_port}} freebusy.promocode.v1.PromoCodeService/CreatePromoCode

# Validate a code against a $100 subtotal.
promo-validate code="SUMMER25":
    grpcurl -plaintext -d '{"code":"{{code}}","subtotal":{"currency_code":"USD","units":100}}' \
        localhost:{{grpc_port}} freebusy.promocode.v1.PromoCodeService/ValidatePromoCode
