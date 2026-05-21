#!/usr/bin/env bash
set -euo pipefail

: "${GOOS:?GOOS required}"
: "${GOARCH:?GOARCH required}"
: "${BINARY:?BINARY required}"
: "${VERSION:?VERSION required}"
: "${COMMIT:?COMMIT required}"
: "${DATE:?DATE required}"

export CGO_ENABLED=1
export GOOS GOARCH

LDFLAGS=(
  "-X" "vadlp/internal/version.Version=${VERSION}"
  "-X" "vadlp/internal/version.Commit=${COMMIT}"
  "-X" "vadlp/internal/version.BuildDate=${DATE}"
)

if [[ "${RUNNER_OS:-}" == "Windows" ]]; then
  export PATH="/mingw64/bin:/usr/bin:${PATH}"
fi

go build -ldflags "${LDFLAGS[*]}" -o "${BINARY}" ./cmd/vadlp
