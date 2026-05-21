#!/usr/bin/env bash
set -euo pipefail

VERSION=dev
if [[ "${GITHUB_REF_TYPE:-}" == "tag" ]]; then
  VERSION="${GITHUB_REF_NAME#v}"
elif [[ "${GITHUB_REF_NAME:-}" == "main" || "${GITHUB_REF_NAME:-}" == "master" ]]; then
  VERSION=main
fi

{
  echo "version=$VERSION"
  echo "commit=$(git rev-parse --short HEAD)"
  echo "date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
} >> "${GITHUB_OUTPUT}"
