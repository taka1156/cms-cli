#!/usr/bin/env bash
# Build a single cmsc binary artifact.
#
# Required environment variables:
#   GOOS    - target OS   (e.g. linux, darwin, windows)
#   GOARCH  - target arch (e.g. amd64, arm64)
#   ARCHIVE - output type (tar.gz | exe)
#
# Example (local):
#   GOOS=linux GOARCH=amd64 ARCHIVE=tar.gz bash scripts/build.sh

set -eu

: "${GOOS:?GOOS is required}"
: "${GOARCH:?GOARCH is required}"
: "${ARCHIVE:?ARCHIVE is required}"

export CGO_ENABLED=0

mkdir -p dist tmp

VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")

if [ "${ARCHIVE}" = "exe" ]; then
	out="dist/cms-cli_${GOOS}_${GOARCH}.exe"
	go build -trimpath -o "${out}" ./cmd/cms-cli
else
	bin="tmp/cms-cli"
	go build -trimpath -o "${bin}" ./cmd/cms-cli
	tar -C tmp -czf "dist/cms-cli_${GOOS}_${GOARCH}.tar.gz" cms-cli
	rm -f "${bin}"
fi
