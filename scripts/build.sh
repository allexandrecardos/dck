#!/usr/bin/env sh
set -e

if [ -z "${1:-}" ]; then
  echo "Usage: $0 <version>" >&2
  exit 1
fi

VERSION="$1"
DIST="dist"

mkdir -p "$DIST"

LDFLAGS="-X github.com/allexandrecardos/dck/cmd.version=${VERSION}"

echo "[INFO] Building $VERSION"

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$DIST/dck_${VERSION}_windows_amd64.exe" .
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$DIST/dck_${VERSION}_linux_amd64" .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$DIST/dck_${VERSION}_linux_arm64" .

echo "[INFO] Done. Outputs in dist/"
