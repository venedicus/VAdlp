#!/usr/bin/env bash
set -euo pipefail

mkdir -p bin
eval "$(bash scripts/build-metadata.sh)"
export CGO_ENABLED=1

go build \
	-ldflags "-X vadlp/internal/version.Version=${VERSION} -X vadlp/internal/version.Commit=${COMMIT} -X vadlp/internal/version.BuildDate=${DATE}" \
	-o bin/vadlp \
	./cmd/vadlp

echo "Built bin/vadlp (${VERSION})"
