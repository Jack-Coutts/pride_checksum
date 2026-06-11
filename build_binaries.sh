#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

build() {
  local goos="$1"
  local goarch="$2"
  local outdir="$ROOT/bin/${goos}-${goarch}"
  local outname="pride-checksum-helper"

  mkdir -p "$outdir"

  if [ "$goos" = "windows" ]; then
    outname="${outname}.exe"
  fi

  echo "Building ${goos}/${goarch}..."
  GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "$outdir/$outname" .
}

build windows amd64
build darwin amd64
build darwin arm64
build linux amd64

echo
echo "Built binaries in bin/"
