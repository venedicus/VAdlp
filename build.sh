#!/usr/bin/env bash
set -euo pipefail

mkdir -p bin

VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo none)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "")"

export CGO_ENABLED=1

go build \
  -ldflags "-X vadlp/internal/version.Version=${VERSION} -X vadlp/internal/version.Commit=${COMMIT} -X vadlp/internal/version.BuildDate=${DATE}" \
  -o bin/vadlp \
  ./cmd/vadlp

echo "Built bin/vadlp"
