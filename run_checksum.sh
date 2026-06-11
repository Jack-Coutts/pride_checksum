#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin)
    case "$arch" in
      arm64) platform="darwin-arm64" ;;
      x86_64) platform="darwin-amd64" ;;
      *)
        echo "Unsupported macOS architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  Linux)
    case "$arch" in
      x86_64|amd64) platform="linux-amd64" ;;
      *)
        echo "Unsupported Linux architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported operating system: $os"
    exit 1
    ;;
esac

HELPER="$SCRIPT_DIR/bin/$platform/pride-checksum-helper"

if [ ! -x "$HELPER" ]; then
  echo "Could not find helper binary at:"
  echo "$HELPER"
  echo
  echo "Download the latest version of this folder, or ask the maintainer to run build_binaries.sh."
  exit 1
fi

if [ "$#" -ne 1 ]; then
  echo "Usage: ./run_checksum.sh /path/to/data/folder"
  exit 1
fi

exec "$HELPER" "$1"
