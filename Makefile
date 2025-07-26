.PHONY: all test build clean lint security race bench check install install-user install-system uninstall uninstall-system help build-linux build-darwin build-windows build-all dev-build dev-setup quick-test

# Variables
BINARY_NAME=carrion-lsp
MAIN_PATH=./cmd/carrion-lsp
COVERAGE_FILE=coverage.out
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=build
USER_INSTALL_DIR=$(HOME)/.local/bin
SYSTEM_INSTALL_DIR=/usr/local/bin

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: check build

# Run all checks before building
check: lint test race security

# Build the binary
build:
	@echo "Building Carrion LSP v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests with coverage
test:
	go test -v -cover -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html

# Run tests with race detector
race:
	go test -race ./...

# Run linters
lint:
	@echo "Running linters..."
	gofmt -l .
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --enable-all; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Security scanning
security:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./...; \
	else \
		echo "gosec not installed, skipping security scan..."; \
	fi

# Benchmarks
bench:
	go test -bench=. -benchmem ./...

# Install the binary (legacy - use install-user or install-system)
install: install-user

# Install to user's local bin directory
install-user: build
	@echo "Installing Carrion LSP to $(USER_INSTALL_DIR)..."
	@mkdir -p $(USER_INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(USER_INSTALL_DIR)/
	@chmod +x $(USER_INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(USER_INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Make sure $(USER_INSTALL_DIR) is in your PATH:"
	@echo "  export PATH=\"$(USER_INSTALL_DIR):\$$PATH\""
	@echo ""
	@echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.)"

# Install system-wide (requires sudo)
install-system: build
	@echo "Installing Carrion LSP system-wide..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(SYSTEM_INSTALL_DIR)/
	@sudo chmod +x $(SYSTEM_INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(SYSTEM_INSTALL_DIR)/$(BINARY_NAME)"

# Uninstall from user directory
uninstall:
	@echo "Uninstalling Carrion LSP from $(USER_INSTALL_DIR)..."
	@rm -f $(USER_INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstalled from $(USER_INSTALL_DIR)"

# Uninstall from system directory
uninstall-system:
	@echo "Uninstalling Carrion LSP from system..."
	@sudo rm -f $(SYSTEM_INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstalled from $(SYSTEM_INSTALL_DIR)"

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE) coverage.html
	rm -rf dist/ $(BUILD_DIR)/

# Development setup
dev-setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go mod download

# Quick test during development
quick-test:
	go test -short ./...

# Cross-compilation targets
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-all: build-linux build-darwin build-windows
	@echo "Built all platform binaries"

# Development build with race detection
dev-build:
	@echo "Building development version with race detection..."
	@mkdir -p $(BUILD_DIR)
	@go build -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-dev $(MAIN_PATH)

# Help target
help:
	@echo "Carrion LSP Build System"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "  Building:"
	@echo "    build          - Build the binary for current platform"
	@echo "    build-all      - Cross-compile for all platforms"
	@echo "    build-linux    - Build for Linux x64"
	@echo "    build-darwin   - Build for macOS (Intel and Apple Silicon)"
	@echo "    build-windows  - Build for Windows x64"
	@echo "    dev-build      - Build development version with race detection"
	@echo ""
	@echo "  Installation:"
	@echo "    install-user   - Install to ~/.local/bin (recommended)"
	@echo "    install-system - Install to /usr/local/bin (requires sudo)"
	@echo "    install        - Alias for install-user"
	@echo "    uninstall      - Uninstall from ~/.local/bin"
	@echo "    uninstall-system - Uninstall from /usr/local/bin (requires sudo)"
	@echo ""
	@echo "  Testing & Quality:"
	@echo "    test           - Run all tests with coverage"
	@echo "    quick-test     - Run tests without coverage"
	@echo "    race           - Run tests with race detector"
	@echo "    bench          - Run benchmarks"
	@echo "    lint           - Run linters"
	@echo "    security       - Run security checks"
	@echo "    check          - Run all quality checks"
	@echo ""
	@echo "  Maintenance:"
	@echo "    clean          - Clean build artifacts"
	@echo "    dev-setup      - Install development tools"
	@echo "    help           - Show this help message"