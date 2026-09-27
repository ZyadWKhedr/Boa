#!/usr/bin/env bash
set -e

# Boa universal installer for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/ZyadWKhedr/Boa/main/install.sh | bash

REPO="ZyadWKhedr/Boa"
BINARY_NAME="boa"
ALIAS_NAME="bo"

# Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin|linux)
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

INSTALL_DIR="${HOME}/.local/bin"
if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
fi

mkdir -p "$INSTALL_DIR"

echo "🐍 Installing Boa for ${OS}-${ARCH} into ${INSTALL_DIR}..."

# Download latest release asset
TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "latest")

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/boa-${OS}-${ARCH}"

TMP_DIR="$(mktemp -d)"
TMP_FILE="${TMP_DIR}/boa"

if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" 2>/dev/null; then
    chmod +x "$TMP_FILE"
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
else
    # Fallback to Go build if binary not found or pre-release
    echo "Pre-built binary not yet published for ${TAG}; installing via Go..."
    if command -v go >/dev/null 2>&1; then
        go install "github.com/${REPO}@latest"
        GOBIN="$(go env GOPATH)/bin"
        cp "${GOBIN}/compressor" "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || cp "${GOBIN}/boa" "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || true
    else
        echo "Error: Could not download release binary and Go compiler is not installed."
        exit 1
    fi
fi

rm -rf "$TMP_DIR"

# Create symlinks
ln -sf "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${ALIAS_NAME}"
ln -sf "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/compressor"

echo "✔ Successfully installed Boa (${BINARY_NAME}, ${ALIAS_NAME}) into ${INSTALL_DIR}!"
echo ""
echo "To get started, run:"
echo "  bo"
echo ""
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Note: Make sure ${INSTALL_DIR} is in your PATH. Add this to ~/.zshrc or ~/.bashrc:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
