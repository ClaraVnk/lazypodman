#!/bin/bash
set -euo pipefail

# Installs or updates the latest lazypodman release binary.
# Override the destination with DIR=... (default: ~/.local/bin).
DIR="${DIR:-"$HOME/.local/bin"}"
REPO="ClaraVnk/lazypodman"

# Map the host architecture to the GoReleaser arch label.
ARCH=$(uname -m)
case $ARCH in
    x86_64 | amd64) ARCH=amd64 ;;
    aarch64 | arm64) ARCH=arm64 ;;
    *)
        echo "unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

# GoReleaser names archives with a lowercase OS (linux/darwin/windows).
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Resolve the latest release tag.
GITHUB_LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' \
    "https://github.com/${REPO}/releases/latest" |
    sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
if [ -z "$GITHUB_LATEST_VERSION" ]; then
    echo "could not determine the latest release of ${REPO} (none published yet?)" >&2
    exit 1
fi

GITHUB_FILE="lazypodman_${GITHUB_LATEST_VERSION//v/}_${OS}_${ARCH}.tar.gz"
GITHUB_URL="https://github.com/${REPO}/releases/download/${GITHUB_LATEST_VERSION}/${GITHUB_FILE}"

# Download and install into a temporary working directory.
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
curl -L -o "$TMP/lazypodman.tar.gz" "$GITHUB_URL"
tar xzf "$TMP/lazypodman.tar.gz" -C "$TMP" lazypodman
install -Dm 755 "$TMP/lazypodman" -t "$DIR"

echo "lazypodman ${GITHUB_LATEST_VERSION} installed to ${DIR}/lazypodman"
