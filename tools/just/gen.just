# Code generation: protobuf -> Go, OpenAPI, ORM, TypeScript, docs.
# Invoked as `just gen <recipe>` (e.g. `just gen orm`). `just gen` runs `all`.
# Recipes run from the repo root so buf/template paths resolve unchanged.
set working-directory := '../..'

# Regenerate everything: Go, OpenAPI (+merge+envoy), ORM, TypeScript, docs.
[group('codegen')]
all: deps go openapi envoy orm typescript docs

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

# Generate the TypeScript client.
[group('codegen')]
typescript:
    buf generate --template tools/protobuf/buf/typescript.buf.gen.yaml

# Install the Python GAPIC toolchain into a repo-local venv (one-time).
# gapic-generator provides protoc-gen-python_gapic; pypandoc-binary bundles the
# pandoc that gapic-generator needs to render proto comments. `just gen gapic`
# auto-detects and uses this venv, so nothing has to be on your PATH.
[group('codegen')]
gapic-deps:
    #!/usr/bin/env bash
    set -euo pipefail
    venv=tools/protobuf/buf/gapic/.venv
    # gapic-generator's deps (google-api-core) want Python >= 3.10; pick the
    # newest available interpreter and fall back to python3.
    py=python3
    for c in python3.13 python3.12 python3.11 python3.10; do command -v "$c" >/dev/null 2>&1 && { py="$c"; break; }; done
    rm -rf "$venv"
    "$py" -m venv "$venv"
    "$venv/bin/pip" install --quiet --upgrade pip
    "$venv/bin/pip" install --quiet gapic-generator pypandoc-binary
    echo "Installed protoc-gen-python_gapic + bundled pandoc into $venv ($("$venv/bin/python" --version))"

# Generate the GAPIC clients (Go + Python).
#
# Go GAPIC emits one library per proto package, so we loop the services and
# generate each into its own apiv1 dir (client/go/freebusy/<svc>/apiv1).
# Python GAPIC emits one unified library per API, generated in a single pass
# into client/python.
#
# Python uses the repo-local venv from `just gen gapic-deps` if present (no
# PATH setup needed); otherwise it falls back to protoc-gen-python_gapic on
# PATH. It is skipped (with a warning) when neither the plugin nor pandoc is
# available.
[group('codegen')]
gapic:
    #!/usr/bin/env bash
    set -euo pipefail
    tmpl=tools/protobuf/buf/gapic.buf.gen.yaml
    venv=tools/protobuf/buf/gapic/.venv
    services=(availability booking identity organisation promocode resource schedule)
    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT
    # buf treats an extensionless --template path as inline config, so every
    # generated template below must keep a .yaml extension.
    header() { printf 'version: v2\nclean: false\ninputs:\n  - directory: protobuf\nplugins:\n'; }
    # Prefer the repo-local venv's plugin (and its bundled pandoc).
    [ -x "$venv/bin/protoc-gen-python_gapic" ] && export PATH="$(pwd)/$venv/bin:$PATH"
    # pandoc is satisfied either by PATH or by the venv's bundled copy (pypandoc).
    pandoc_ok() { command -v pandoc >/dev/null 2>&1 || { [ -x "$venv/bin/python" ] && "$venv/bin/python" -c 'import pypandoc; pypandoc.get_pandoc_path()' >/dev/null 2>&1; }; }

    # --- Go: one GAPIC package per service ---
    for svc in "${services[@]}"; do
        tmp="$tmpdir/go-$svc.yaml"
        sed "s/__SVC__/$svc/g" "$tmpl" | awk '/^ *# PYTHON_PLUGIN$/{exit} {print}' > "$tmp"
        echo "==> gapic go: $svc"
        buf generate --template "$tmp" --path "protobuf/freebusy/$svc/v1"
    done

    # --- Python: one unified library (proto-plus + clients) + raw _pb2 commons ---
    if ! command -v protoc-gen-python_gapic >/dev/null 2>&1; then
        echo "WARN: protoc-gen-python_gapic not found — skipped Python. Run: just gen gapic-deps" >&2
    elif ! pandoc_ok; then
        echo "WARN: pandoc not found (gapic-generator-python needs it) — skipped Python. Run: just gen gapic-deps (or brew install pandoc)" >&2
    else
        # Scope the GAPIC pass to the SERVICE packages only; freebusy.shared.v1
        # stays an import and is emitted as raw _pb2 below (see template notes).
        svc_paths=(); for svc in "${services[@]}"; do svc_paths+=(--path "protobuf/freebusy/$svc/v1"); done
        echo "==> gapic python: services"
        { header; awk '/^ *# PYTHON_PLUGIN$/{f=1;next} /^ *# PYTHON_PB2$/{exit} f' "$tmpl"; } > "$tmpdir/py-gapic.yaml"
        buf generate --template "$tmpdir/py-gapic.yaml" "${svc_paths[@]}"
        echo "==> gapic python: shared _pb2"
        { header; awk '/^ *# PYTHON_PB2$/{f=1;next} f' "$tmpl"; } > "$tmpdir/py-pb2.yaml"
        buf generate --template "$tmpdir/py-pb2.yaml" --path protobuf/freebusy/shared/v1
    fi

# Generate protobuf docs (per-module + root README.md) via protodoc.
[group('codegen')]
docs:
    go run ./tools/protobuf/docs protobuf
