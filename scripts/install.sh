#!/bin/bash

set -e

# Auto PR Installation Script
# This script installs the Auto PR CLI tool

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REPO="charles-adedotun/auto-pr"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_info() {
    echo -e "${YELLOW}$1${NC}"
}

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    
    case "$OS" in
        linux*)     OS="linux" ;;
        darwin*)    OS="darwin" ;;
        msys*|mingw*|cygwin*)    OS="windows" ;;
        *)          print_error "Unsupported OS: $OS"; exit 1 ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)             print_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    print_info "Detected platform: $PLATFORM"
}

# Check dependencies
check_dependencies() {
    print_info "Checking dependencies..."
    
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        print_error "curl or wget is required but not installed."
        exit 1
    fi
    
    if ! command -v tar &> /dev/null; then
        print_error "tar is required but not installed."
        exit 1
    fi
    
    # Check for GitHub CLI (optional but recommended)
    if command -v gh &> /dev/null; then
        print_success "✓ GitHub CLI (gh) found"
    else
        print_info "⚠ GitHub CLI (gh) not found - required for GitHub integration"
    fi
    
    # Check for GitLab CLI (optional)
    if command -v glab &> /dev/null; then
        print_success "✓ GitLab CLI (glab) found"
    else
        print_info "⚠ GitLab CLI (glab) not found - required for GitLab integration"
    fi
    
    # Check for Claude CLI (required)
    if command -v claude &> /dev/null; then
        print_success "✓ Claude CLI found"
    else
        print_info "⚠ Claude CLI not found - Please install Claude Code"
        print_info "   - Visit: https://docs.anthropic.com/en/docs/claude-code"
    fi
}

# Get latest version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        print_info "Getting latest version..."
        if command -v curl &> /dev/null; then
            VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        else
            VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        fi
        
        if [ -z "$VERSION" ]; then
            print_error "No releases found yet. Please specify a version with VERSION=<version> or build from source."
            echo
            echo "To build from source:"
            echo "  git clone https://github.com/$REPO.git"
            echo "  cd auto-pr"
            echo "  go build -o auto-pr ."
            echo "  sudo mv auto-pr /usr/local/bin/"
            exit 1
        fi
        print_info "Latest version: $VERSION"
    fi
}

# Download and install
download_and_install() {
    BINARY_NAME="auto-pr"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="auto-pr.exe"
    fi
    
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/auto-pr-$PLATFORM.tar.gz"
    TMP_DIR=$(mktemp -d)
    
    print_info "Downloading Auto PR $VERSION for $PLATFORM..."
    
    cd "$TMP_DIR"
    if command -v curl &> /dev/null; then
        curl -sL "$DOWNLOAD_URL" -o auto-pr.tar.gz
    else
        wget -q "$DOWNLOAD_URL" -O auto-pr.tar.gz
    fi
    
    if [ ! -f auto-pr.tar.gz ]; then
        print_error "Download failed"
        rm -rf "$TMP_DIR"
        exit 1
    fi
    
    print_info "Extracting..."
    tar xzf auto-pr.tar.gz
    
    if [ ! -f "$BINARY_NAME" ]; then
        print_error "Binary not found in archive"
        rm -rf "$TMP_DIR"
        exit 1
    fi
    
    print_info "Installing to $INSTALL_DIR..."
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_NAME" "$INSTALL_DIR/auto-pr"
        chmod +x "$INSTALL_DIR/auto-pr"
    else
        print_info "Root access required to install to $INSTALL_DIR"
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/auto-pr"
        sudo chmod +x "$INSTALL_DIR/auto-pr"
    fi
    
    rm -rf "$TMP_DIR"
}

# Verify installation
verify_installation() {
    if command -v auto-pr &> /dev/null; then
        print_success "✓ Auto PR installed successfully!"
        print_info "Version: $(auto-pr --version)"
    else
        print_error "Installation failed - auto-pr not found in PATH"
        print_info "You may need to add $INSTALL_DIR to your PATH"
        exit 1
    fi
}

# Setup initial configuration
setup_config() {
    print_info "\nSetting up initial configuration..."
    
    if [ ! -f "$HOME/.auto-pr/config.yaml" ]; then
        auto-pr config init
        print_success "✓ Configuration initialized at ~/.auto-pr/config.yaml"
    else
        print_info "Configuration already exists at ~/.auto-pr/config.yaml"
    fi
}

# Main installation flow
main() {
    echo "Auto PR Installer"
    echo "================"
    echo
    
    detect_platform
    check_dependencies
    get_latest_version
    download_and_install
    verify_installation
    setup_config
    
    echo
    print_success "Installation complete! 🎉"
    echo
    echo "Next steps:"
    echo "1. Ensure Claude Code is set up:"
    echo "   - Install Claude Code: https://docs.anthropic.com/en/docs/claude-code"
    echo "   - Authenticate your Claude Code installation"
    echo
    echo "2. Test the installation:"
    echo "   auto-pr status"
    echo
    echo "3. Create your first PR:"
    echo "   auto-pr create --dry-run"
    echo
    echo "For more information, visit: https://github.com/$REPO"
}

# Run main function
main "$@"