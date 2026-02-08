#!/bin/sh
set -e

# aiops installer — downloads the latest release binary from GitHub.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/voltic-software/aiops/main/install.sh | sh

REPO="voltic-software/aiops"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="aiops"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

# Get latest release tag
echo "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Error: Could not determine latest release. Check https://github.com/$REPO/releases"
  exit 1
fi

VERSION="${LATEST#v}"
ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$ARCHIVE"

echo "Downloading $BINARY $LATEST ($OS/$ARCH)..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL -o "$TMPDIR/$ARCHIVE" "$URL"
tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"

# Install binary
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "✅ $BINARY $LATEST installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Get started:"
echo "  cd your-project"
echo "  aiops init"
