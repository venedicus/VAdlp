#!/usr/bin/env bash
# Prints VERSION=, COMMIT=, and DATE= for local builds and CI.
set -euo pipefail

if [[ "${GITHUB_REF_TYPE:-}" == "tag" ]]; then
	VERSION="${GITHUB_REF_NAME#v}"
elif [[ "${GITHUB_REF_NAME:-}" == "main" || "${GITHUB_REF_NAME:-}" == "master" ]]; then
	VERSION=main
elif v="$(git describe --tags --always --dirty 2>/dev/null)"; then
	VERSION="$v"
else
	VERSION=dev
fi

COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo none)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "")"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
	{
		echo "version=$VERSION"
		echo "commit=$COMMIT"
		echo "date=$DATE"
	} >>"$GITHUB_OUTPUT"
else
	echo "VERSION=$VERSION"
	echo "COMMIT=$COMMIT"
	echo "DATE=$DATE"
fi
