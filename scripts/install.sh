#!/bin/bash
# Based on Deno and nvm installer: Copyright 2023 the Deno authors. All rights reserved. MIT license.
# TODO(everyone): Keep this script simple and easily auditable.
set -e

BINARY_NAME="wso2-integration-platform"
REPO="wso2/integration-platform-tools"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.wso2-integration-platform/bin}"

getArchitecture() {
    local ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
    if [[ "$ARCH" == "x86_64" ]]; then
        echo "amd64"
    elif [[ "$ARCH" == "i386" ]]; then
        echo "386"
    elif [[ "$ARCH" == "arm64" || "$ARCH" == "aarch64" ]]; then
        echo "arm64"
    elif [[ "$ARCH" == "arm" ]]; then
        echo "arm"
    else
        echo "Unsupported architecture: $ARCH" >&2
        exit 1
    fi
}

getVersion() {
    if [ -n "$1" ]; then
        echo "$1"
        return
    fi
    local LAST_RELEASE=${LAST_RELEASE:-"false"}

    if [ "$LAST_RELEASE" == "true" ]; then
        local VERSION=$(curl --silent "https://api.github.com/repos/$REPO/releases" | grep -E 'tag_name|prerelease": true' | grep -E 'prerelease": true' -B1 | head -n 1 | awk -F '"' '{print $4}')
    else
        local VERSION=$(curl -Ls -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest" | cut -d/ -f8)
    fi

    echo "$VERSION"
}

main() {
    local OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    local ARCH=$(getArchitecture)
    local SHELL_TYPE=$(basename "$SHELL")
    local TMP_DIR=$(mktemp -d -t wso2-ip-XXXXXXXXXX)
    local LATEST_VERSION=$(getVersion "$1")
    mkdir -p "$INSTALL_DIR"

    local FILE_TYPE=""
    if [[ "$OS" == "linux" ]]; then
        FILE_TYPE=".tar.gz"
    elif [[ "$OS" == "darwin" ]]; then
        FILE_TYPE=".zip"
    else
        echo "Unsupported OS: $OS" >&2
        echo "On Windows, use: iwr https://raw.githubusercontent.com/$REPO/main/scripts/install.ps1 | iex" >&2
        exit 1
    fi

    local FILE_NAME="${BINARY_NAME}-${LATEST_VERSION}-${OS}-${ARCH}"
    local DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${FILE_NAME}${FILE_TYPE}"

    echo "Installing $BINARY_NAME $LATEST_VERSION..."
    echo "Downloading from: $DOWNLOAD_URL"
    curl -q --fail --location --progress-bar --output "$TMP_DIR/${FILE_NAME}${FILE_TYPE}" "$DOWNLOAD_URL"

    echo "Extracting archive..."
    if [[ "$FILE_TYPE" == ".tar.gz" ]]; then
        tar -xzf "$TMP_DIR/${FILE_NAME}${FILE_TYPE}" -C "$TMP_DIR"
    elif [[ "$FILE_TYPE" == ".zip" ]]; then
        unzip -q "$TMP_DIR/${FILE_NAME}${FILE_TYPE}" -d "$TMP_DIR"
    fi

    echo "Installing to $INSTALL_DIR/$BINARY_NAME"
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    echo "Cleaning up..."
    rm -rf "$TMP_DIR"

    local PROFILE=$(detect_profile)
    local PATH_LINE="export PATH=\"$INSTALL_DIR:\$PATH\""

    if [ -z "$PROFILE" ]; then
        echo ""
        echo "Could not detect shell profile. Add this to your profile manually:"
        echo "  $PATH_LINE"
    else
        if ! grep -qc "$INSTALL_DIR" "$PROFILE" 2>/dev/null; then
            echo "" >> "$PROFILE"
            echo "# wso2-integration-platform" >> "$PROFILE"
            echo "$PATH_LINE" >> "$PROFILE"
            echo "# wso2-integration-platform end" >> "$PROFILE"
            echo ""
            echo "Added $BINARY_NAME to PATH in $PROFILE"
            echo "Run 'source $PROFILE' or open a new terminal to use it."
        else
            echo "$INSTALL_DIR is already in $PROFILE"
        fi
    fi

    echo ""
    echo "$BINARY_NAME $LATEST_VERSION installed successfully!"
    echo ""
    echo "Get started:"
    echo "  $BINARY_NAME login"
    echo ""
    echo "Use as MCP server with Claude Code:"
    echo "  claude mcp add wso2-integration-platform -- $BINARY_NAME start-mcp-server"
    echo ""
    echo "Or install the npm package for \`claude mcp add --package\`:"
    echo "  npm install -g @wso2/integration-platform-mcp"
}

detect_profile() {
    if [ "${PROFILE-}" = '/dev/null' ]; then
        return
    fi

    if [ -n "${PROFILE}" ] && [ -f "${PROFILE}" ]; then
        echo "${PROFILE}"
        return
    fi

    local DETECTED_PROFILE=''

    if [ "${SHELL#*bash}" != "$SHELL" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            DETECTED_PROFILE="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            DETECTED_PROFILE="$HOME/.bash_profile"
        fi
    elif [ "${SHELL#*zsh}" != "$SHELL" ]; then
        if [ -f "$HOME/.zshrc" ]; then
            DETECTED_PROFILE="$HOME/.zshrc"
        elif [ -f "$HOME/.zprofile" ]; then
            DETECTED_PROFILE="$HOME/.zprofile"
        fi
    fi

    if [ -z "$DETECTED_PROFILE" ]; then
        if [ -f "$HOME/.profile" ]; then
            DETECTED_PROFILE="$HOME/.profile"
        elif [ -f "$HOME/.bashrc" ]; then
            DETECTED_PROFILE="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            DETECTED_PROFILE="$HOME/.bash_profile"
        elif [ -f "$HOME/.zshrc" ]; then
            DETECTED_PROFILE="$HOME/.zshrc"
        elif [ -f "$HOME/.zprofile" ]; then
            DETECTED_PROFILE="$HOME/.zprofile"
        fi
    fi

    if [ -n "$DETECTED_PROFILE" ]; then
        echo "$DETECTED_PROFILE"
    fi
}


main "$@"
unset -f main detect_profile getArchitecture getVersion
