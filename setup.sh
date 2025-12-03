#!/bin/sh
#
# maestro-ios-device setup script
# Downloads the correct binary and runs setup
#

set -e

REPO="devicelab-dev/maestro-ios-device"
BINARY_NAME="maestro-ios-device"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo ""
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  maestro-ios-device — Unofficial Community Tool                ║"
echo "║                                                                ║"
echo "║  This is NOT affiliated with mobile.dev or Maestro.           ║"
echo "║  Stop-gap until PR #2856 is merged.                           ║"
echo "║                                                                ║"
echo "║  Built by DeviceLab — https://devicelab.dev                   ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

printf "Continue with installation? [y/N] "
read -r REPLY </dev/tty
if [ "$REPLY" != "y" ] && [ "$REPLY" != "Y" ]; then
    echo "Installation cancelled."
    exit 1
fi

echo ""

# Check OS
OS=$(uname -s)
if [ "$OS" != "Darwin" ]; then
    printf "${RED}Error: Only macOS is supported${NC}\n"
    exit 1
fi

# Check architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64)
        ARCH="arm64"
        ;;
    *)
        printf "${RED}Error: Unsupported architecture: $ARCH${NC}\n"
        exit 1
        ;;
esac

BINARY="maestro-ios-device-darwin-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

echo "Detected: macOS ${ARCH}"
echo ""

# Find where maestro is installed
MAESTRO_PATH=$(which maestro 2>/dev/null || true)
if [ -z "$MAESTRO_PATH" ]; then
    printf "${RED}Error: Maestro not found in PATH. Install Maestro first: https://maestro.mobile.dev${NC}\n"
    exit 1
fi

INSTALL_DIR=$(dirname "$MAESTRO_PATH")
echo "Found Maestro at: $INSTALL_DIR"
echo ""

# Download binary
echo "📥 Downloading ${BINARY}..."
TMP_FILE=$(mktemp)
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
    printf "${RED}Error: Failed to download binary${NC}\n"
    echo "URL: $DOWNLOAD_URL"
    rm -f "$TMP_FILE"
    exit 1
fi

# Install binary
echo "📦 Installing to $INSTALL_DIR/$BINARY_NAME..."
mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

printf "${GREEN}✅ Binary installed${NC}\n"
echo ""

# Run setup command
echo "🔧 Running maestro-ios-device setup..."
echo ""
"$INSTALL_DIR/$BINARY_NAME" setup

echo ""
printf "${GREEN}✅ Installation complete!${NC}\n"
