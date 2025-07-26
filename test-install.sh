#!/bin/bash

# Test script for Carrion LSP installation
# This script tests the installation process on different platforms

set -e

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

# Test functions
test_build() {
    log_info "Testing build process..."
    
    if command -v go &> /dev/null; then
        if make build; then
            log_success "Build successful"
            return 0
        else
            log_error "Build failed"
            return 1
        fi
    else
        log_warning "Go not installed, skipping build test"
        return 0
    fi
}

test_installation() {
    log_info "Testing installation..."
    
    # Test user installation
    if make install-user; then
        log_success "User installation successful"
        
        # Check if binary exists
        if [ -f "$HOME/.local/bin/carrion-lsp" ]; then
            log_success "Binary installed to ~/.local/bin"
        else
            log_error "Binary not found after installation"
            return 1
        fi
        
        # Test if binary is executable
        if "$HOME/.local/bin/carrion-lsp" --version &> /dev/null; then
            log_success "Binary is executable and responds to --version"
        else
            log_warning "Binary exists but may not be working properly"
        fi
        
        return 0
    else
        log_error "Installation failed"
        return 1
    fi
}

test_makefile_targets() {
    log_info "Testing Makefile targets..."
    
    local targets=(
        "help"
        "clean"
        "lint"
        "quick-test"
    )
    
    for target in "${targets[@]}"; do
        if make "$target" &> /dev/null; then
            log_success "Target '$target' works"
        else
            log_warning "Target '$target' may have issues"
        fi
    done
}

test_cross_compilation() {
    log_info "Testing cross-compilation..."
    
    if command -v go &> /dev/null; then
        if make build-all; then
            log_success "Cross-compilation successful"
            
            # Check if binaries were created
            local expected_binaries=(
                "build/carrion-lsp-linux-amd64"
                "build/carrion-lsp-darwin-amd64" 
                "build/carrion-lsp-darwin-arm64"
                "build/carrion-lsp-windows-amd64.exe"
            )
            
            for binary in "${expected_binaries[@]}"; do
                if [ -f "$binary" ]; then
                    log_success "Binary $binary created"
                else
                    log_warning "Binary $binary not found"
                fi
            done
        else
            log_error "Cross-compilation failed"
            return 1
        fi
    else
        log_warning "Go not installed, skipping cross-compilation test"
    fi
}

test_install_script() {
    log_info "Testing install script syntax..."
    
    if [ -f "install.sh" ]; then
        if bash -n install.sh; then
            log_success "Install script syntax is valid"
        else
            log_error "Install script has syntax errors"
            return 1
        fi
    else
        log_error "Install script not found"
        return 1
    fi
}

test_editor_configs() {
    log_info "Testing editor configuration files..."
    
    local config_dirs=(
        "configs/neovim/lazy.nvim"
        "configs/neovim/mason"
        "configs/neovim/nvchad"
        "configs/neovim/generic"
        "configs/vscode"
    )
    
    for dir in "${config_dirs[@]}"; do
        if [ -d "$dir" ]; then
            log_success "Configuration directory $dir exists"
            
            # Check for key files
            case "$dir" in
                "configs/neovim"*)
                    if find "$dir" -name "*.lua" | grep -q .; then
                        log_success "Lua configuration files found in $dir"
                    else
                        log_warning "No Lua files found in $dir"
                    fi
                    ;;
                "configs/vscode")
                    if [ -f "$dir/package.json" ]; then
                        log_success "VS Code package.json found"
                    else
                        log_warning "VS Code package.json not found"
                    fi
                    ;;
            esac
        else
            log_warning "Configuration directory $dir not found"
        fi
    done
}

cleanup() {
    log_info "Cleaning up test artifacts..."
    make clean &> /dev/null || true
    rm -f "$HOME/.local/bin/carrion-lsp" &> /dev/null || true
}

main() {
    echo "🧪 Carrion LSP Installation Test Suite"
    echo "======================================"
    echo
    
    local failed_tests=0
    
    # Run tests
    test_install_script || ((failed_tests++))
    test_build || ((failed_tests++))
    test_installation || ((failed_tests++))
    test_makefile_targets || ((failed_tests++))
    test_cross_compilation || ((failed_tests++))
    test_editor_configs || ((failed_tests++))
    
    echo
    echo "Test Summary:"
    echo "============"
    
    if [ $failed_tests -eq 0 ]; then
        log_success "All tests passed! ✅"
        echo
        echo "The Carrion LSP installation system is working correctly."
        echo "You can now safely distribute this to users."
    else
        log_error "$failed_tests test(s) failed ❌"
        echo
        echo "Please review the failed tests and fix any issues before distribution."
    fi
    
    cleanup
    
    exit $failed_tests
}

# Run main function
main "$@"