#!/usr/bin/env sh
set -e

REPO="allexandrecardos/dck"
NAME="dck"
INSTALL_DIR="/usr/local/bin"
ARCH="amd64"
OS="linux"

if [ -n "${DCK_INSTALL_DIR:-}" ]; then
  INSTALL_DIR="$DCK_INSTALL_DIR"
fi

if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
  ARCH="arm64"
fi

ASSET="${NAME}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

printf "[INFO] Downloading %s...\n" "$URL"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TMP_DIR/$NAME"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP_DIR/$NAME" "$URL"
else
  echo "[ERROR] curl or wget is required"
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
