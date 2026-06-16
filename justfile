default:
    @just --list

# run prek hooks (--all-files to force proto lint/generate when no .proto files staged)
prek *args:
    # Auto-fix trailing whitespace and EOF before running hooks
    python3 tools/protobuf/fix-precommit.py .
    prek run --all-files {{ args }}

# Build the in-repo protorm plugin so ORM generation always uses this monorepo's
# source, not a stale brew/global protoc-gen-protorm that can drift from protorm/.
build-protorm:
    cd protorm && go build -buildvcs=false -o bin/protoc-gen-protorm ./plugin/cmd/protoc-gen-protorm

# Generate protobuf docs (per-module + root README.md) via protodoc.
docs:
    go run ./tools/protobuf/docs protobuf

# Generate the ORM schemas (prisma + gorm) with the in-repo protorm plugin.
orm: build-protorm
    PATH="{{justfile_directory()}}/protorm/bin:$PATH" buf generate --template tools/protobuf/buf/orm.buf.gen.yaml

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
        echo "==> Building in-repo protorm plugin..."
        (cd protorm && go build -buildvcs=false -o bin/protoc-gen-protorm ./plugin/cmd/protoc-gen-protorm)
        echo "==> ORM generate (prisma + gorm)..."
        PATH="$PWD/protorm/bin:$PATH" buf generate --template tools/protobuf/buf/orm.buf.gen.yaml
        echo "==> Docs generate..."
        go run ./tools/protobuf/docs protobuf
    elif [ "{{language}}" = "orm" ]; then
        echo "==> Building in-repo protorm plugin..."
        (cd protorm && go build -buildvcs=false -o bin/protoc-gen-protorm ./plugin/cmd/protoc-gen-protorm)
        PATH="$PWD/protorm/bin:$PATH" buf generate --template tools/protobuf/buf/orm.buf.gen.yaml
    elif [ "{{language}}" = "docs" ]; then
        go run ./tools/protobuf/docs protobuf
    elif [ -f "tools/protobuf/buf/{{language}}.buf.gen.yaml" ]; then
        buf generate --template tools/protobuf/buf/{{language}}.buf.gen.yaml
    else
        echo "Error: Template for {{language}} not found!"
        echo "Available languages: go, openapiv3, orm, docs"
        exit 1
    fi
