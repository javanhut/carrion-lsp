#!/bin/bash

# Carrion LSP Installation Script
# This script downloads and installs the Carrion Language Server

set -e

# Configuration
REPO="javanhut/carrion-lsp"
BINARY_NAME="carrion-lsp"
INSTALL_DIR="$HOME/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*|MSYS*) os="windows";;
        *)          os="unknown";;
    esac
    
    case "$(uname -m)" in
        x86_64)     arch="amd64";;
        arm64|aarch64) arch="arm64";;
        *)          arch="amd64";;  # Default to amd64
    esac
    
    # Special case for Apple Silicon Macs
    if [[ "$os" == "darwin" && "$(uname -m)" == "arm64" ]]; then
        arch="arm64"
    fi
    
    echo "${os}-${arch}"
}

# Check if binary is already installed
check_existing() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        local existing_version
        existing_version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        log_warning "Carrion LSP is already installed: $existing_version"
        echo -n "Do you want to reinstall? [y/N]: "
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            log_info "Installation cancelled"
            exit 0
        fi
    fi
}

# Get latest release version
get_latest_version() {
    log_info "Fetching latest release version..."
    local version
    version=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    if [[ -z "$version" ]]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
    echo "$version"
}

# Download and install binary
install_binary() {
    local platform="$1"
    local version="$2"
    local binary_url binary_file
    
    # Construct download URL
    if [[ "$platform" == *"windows"* ]]; then
        binary_file="${BINARY_NAME}-${platform}.exe"
    else
        binary_file="${BINARY_NAME}-${platform}"
    fi
    
    binary_url="https://github.com/$REPO/releases/download/$version/$binary_file"
    
    log_info "Downloading from: $binary_url"
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Download binary
    local temp_file
    temp_file=$(mktemp)
    if ! curl -L -o "$temp_file" "$binary_url"; then
        log_error "Failed to download binary"
        rm -f "$temp_file"
        exit 1
    fi
    
    # Install binary
    local install_path="$INSTALL_DIR/$BINARY_NAME"
    if [[ "$platform" == *"windows"* ]]; then
        install_path="${install_path}.exe"
    fi
    
    mv "$temp_file" "$install_path"
    chmod +x "$install_path"
    
    log_success "Installed to $install_path"
}

# Build from source if binary not available
build_from_source() {
    log_info "Building from source..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go first: https://golang.org/dl/"
        exit 1
    fi
    
    # Clone repository
    local temp_dir
    temp_dir=$(mktemp -d)
    log_info "Cloning repository to $temp_dir"
    
    if ! git clone "https://github.com/$REPO.git" "$temp_dir"; then
        log_error "Failed to clone repository"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Build binary
    cd "$temp_dir"
    log_info "Building binary..."
    
    if ! make build; then
        log_error "Failed to build binary"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Install binary
    mkdir -p "$INSTALL_DIR"
    cp "build/$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Cleanup
    rm -rf "$temp_dir"
    
    log_success "Built and installed to $INSTALL_DIR/$BINARY_NAME"
}

# Update PATH if needed
update_path() {
    local shell_profile
    
    # Detect shell profile
    case "$SHELL" in
        */bash) shell_profile="$HOME/.bashrc";;
        */zsh)  shell_profile="$HOME/.zshrc";;
        */fish) shell_profile="$HOME/.config/fish/config.fish";;
        *)      shell_profile="$HOME/.profile";;
    esac
    
    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        log_warning "$INSTALL_DIR is not in your PATH"
        echo -n "Add it to your PATH in $shell_profile? [Y/n]: "
        read -r response
        
        if [[ ! "$response" =~ ^[Nn]$ ]]; then
            echo "" >> "$shell_profile"
            echo "# Added by Carrion LSP installer" >> "$shell_profile"
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_profile"
            log_success "Added $INSTALL_DIR to PATH in $shell_profile"
            log_info "Restart your shell or run: source $shell_profile"
        fi
    fi
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."
    
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    if [[ ! -f "$binary_path" ]]; then
        log_error "Binary not found at $binary_path"
        exit 1
    fi
    
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]] || command -v "$BINARY_NAME" &> /dev/null; then
        local version
        version=$("$binary_path" --version 2>/dev/null || echo "unknown")
        log_success "Installation verified: $version"
        log_info "Run 'carrion-lsp --help' for usage information"
    else
        log_success "Binary installed to $binary_path"
        log_warning "Add $INSTALL_DIR to your PATH to use 'carrion-lsp' command"
    fi
}

# Main installation function
main() {
    echo "🦅 Carrion LSP Installer"
    echo "========================"
    echo
    
    # Check for existing installation
    check_existing
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get latest version
    local version
    version=$(get_latest_version)
    log_info "Latest version: $version"
    
    # Try to install binary from release
    if install_binary "$platform" "$version" 2>/dev/null; then
        log_success "Binary installation completed"
    else
        log_warning "Binary not available for $platform, building from source"
        build_from_source
    fi
    
    # Update PATH if needed
    update_path
    
    # Verify installation
    verify_installation
    
    echo
    log_success "Carrion LSP installation completed!"
    echo
    echo "Next steps:"
    echo "1. Configure your editor (see README.md for editor configurations)"
    echo "2. Open a .crl file to test the LSP"
    echo
}

# Run main function
main "$@"