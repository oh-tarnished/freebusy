default:
    @just --list

# run prek hooks (--all-files to force proto lint/generate when no .proto files staged)
prek *args:
    # Auto-fix trailing whitespace and EOF before running hooks
    python3 tools/protobuf/fix-precommit.py .
    prek run --all-files {{ args }}

# generated protos for languages. if language is not specified, generates for all supported languages.
generate language="all" descriptors="true":
    #!/usr/bin/env sh
    if $descriptors ; then
        ./tools/protobuf/protobuf/buf/gen_descriptor.sh
    else
        echo "==> Error: Script for descriptors not found!"
    fi

    echo "==> Updating buf deps..."
    buf dep update

    echo "==> Language Specific Buf generate..."

    if [ "{{language}}" = "all" ]; then
        buf generate --template tools/protobuf/buf/go.buf.gen.yaml
        buf generate --template tools/protobuf/buf/openapiv3.buf.gen.yaml
    else
        if [ -f "tools/protobuf/buf/{{language}}.buf.gen.yaml" ]; then
            buf generate --template tools/protobuf/buf/{{language}}.buf.gen.yaml
            if [ "{{language}}" = "go" ]; then
                rm -rf generated/go/client
                ./tools/protobuf/buf/generate-gapic.sh
            fi
        else
            echo "Error: Template for {{language}} not found!"
            echo "Available languages: dart, go, py, rust, ts, openapiv3"
            exit 1
        fi
    fi
