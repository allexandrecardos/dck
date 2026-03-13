#!/usr/bin/env sh
set -e

REPO="allexandrecardos/dck"
NAME="dck"
INSTALL_DIR="/usr/local/bin"
ARCH="amd64"
OS="linux"
VERSION=""

if [ -n "${DCK_INSTALL_DIR:-}" ]; then
  INSTALL_DIR="$DCK_INSTALL_DIR"
fi

if [ -n "${DCK_VERSION:-}" ]; then
  VERSION="$DCK_VERSION"
fi

if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
  ARCH="arm64"
fi

if [ -z "$VERSION" ]; then
  if command -v curl >/dev/null 2>&1; then
    VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name"\s*:\s*"([^"]+)".*/\1/')"
  elif command -v wget >/dev/null 2>&1; then
    VERSION="$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name"\s*:\s*"([^"]+)".*/\1/')"
  else
    echo "[ERROR] curl or wget is required" >&2
    exit 1
  fi
fi

if [ -z "$VERSION" ]; then
  echo "[ERROR] Failed to resolve latest version. Set DCK_VERSION=vX.Y.Z" >&2
  exit 1
fi

ASSET="${NAME}_${VERSION}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

printf "[INFO] Downloading %s...\n" "$URL"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TMP_DIR/$NAME"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP_DIR/$NAME" "$URL"
else
  echo "[ERROR] curl or wget is required" >&2
  exit 1
fi

chmod +x "$TMP_DIR/$NAME"

if [ ! -w "$INSTALL_DIR" ]; then
  printf "[INFO] Installing with sudo to %s\n" "$INSTALL_DIR"
  sudo mv "$TMP_DIR/$NAME" "$INSTALL_DIR/$NAME"
else
  mv "$TMP_DIR/$NAME" "$INSTALL_DIR/$NAME"
fi

printf "[INFO] Installed %s to %s\n" "$NAME" "$INSTALL_DIR"
printf "[INFO] Try: %s version\n" "$NAME"
