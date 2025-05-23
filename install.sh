#!/bin/bash

# MD-to-PDF Installation Script
# This script downloads and installs MD-to-PDF for your platform

set -e

# Configuration
REPO="fredcamaral/md-to-pdf"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="md-to-pdf"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Detect platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $os"
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    log "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    log "Fetching latest release version..."
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
    
    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version"
    fi
    
    log "Latest version: $VERSION"
}

# Download binary
download_binary() {
    local url="https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME-$PLATFORM"
    local temp_file="/tmp/$BINARY_NAME-$PLATFORM"
    
    log "Downloading $BINARY_NAME from $url..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -L "$url" -o "$temp_file"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_file" "$url"
    else
        error "Neither curl nor wget is available"
    fi
    
    if [ ! -f "$temp_file" ]; then
        error "Download failed"
    fi
    
    log "Download completed"
    echo "$temp_file"
}

# Install binary
install_binary() {
    local temp_file="$1"
    local install_path="$INSTALL_DIR/$BINARY_NAME"
    
    log "Installing $BINARY_NAME to $install_path..."
    
    # Check if we need sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        log "Installation directory requires elevated privileges"
        sudo cp "$temp_file" "$install_path"
        sudo chmod +x "$install_path"
    else
        cp "$temp_file" "$install_path"
        chmod +x "$install_path"
    fi
    
    # Clean up
    rm -f "$temp_file"
    
    success "$BINARY_NAME installed successfully!"
}

# Verify installation
verify_installation() {
    log "Verifying installation..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version_output=$($BINARY_NAME version 2>/dev/null || echo "unknown")
        success "Installation verified: $version_output"
    else
        warn "Binary installed but not in PATH. You may need to restart your shell or add $INSTALL_DIR to your PATH."
    fi
}

# Install mermaid CLI (optional)
install_mermaid() {
    log "Checking for mermaid CLI..."
    
    if command -v mmdc >/dev/null 2>&1; then
        success "Mermaid CLI already installed"
        return
    fi
    
    if command -v npm >/dev/null 2>&1; then
        log "Installing mermaid CLI for diagram support..."
        if npm install -g @mermaid-js/mermaid-cli 2>/dev/null; then
            success "Mermaid CLI installed successfully"
        else
            warn "Failed to install mermaid CLI. Diagram support will use fallbacks."
            warn "You can install it manually with: npm install -g @mermaid-js/mermaid-cli"
        fi
    else
        warn "npm not found. Mermaid diagrams will show as placeholders."
        warn "Install Node.js and run: npm install -g @mermaid-js/mermaid-cli"
    fi
}

# Main installation function
main() {
    echo "MD-to-PDF Installation Script"
    echo "=============================="
    echo
    
    # Check requirements
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        error "Either curl or wget is required for installation"
    fi
    
    # Detect platform
    detect_platform
    
    # Get latest version
    get_latest_version
    
    # Download binary
    temp_file=$(download_binary)
    
    # Install binary
    install_binary "$temp_file"
    
    # Verify installation
    verify_installation
    
    # Install mermaid CLI (optional)
    install_mermaid
    
    echo
    echo "Installation completed! 🚀"
    echo
    echo "Quick start:"
    echo "  $BINARY_NAME convert document.md"
    echo "  $BINARY_NAME config list"
    echo "  $BINARY_NAME --help"
    echo
    echo "For more information, visit: https://github.com/$REPO"
}

# Run main function
main "$@"