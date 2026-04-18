# gsecret Makefile
# Simple installation and management for the GitHub secrets retrieval tool

.PHONY: all build install uninstall clean test help

# Variables
BINARY_NAME=gsecret
INSTALL_PATH=~/bin
GO=go
GOFLAGS=-ldflags="-s -w"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "✓ Build complete: ./$(BINARY_NAME)"

# Install to system path
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Installed successfully!"
	@echo ""
	@echo "You can now run '$(BINARY_NAME)' from anywhere."
	@echo "Try: $(BINARY_NAME) --help"

# Install without sudo (to user's home bin)
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/bin..."
	@mkdir -p ~/bin
	@cp $(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@chmod +x ~/bin/$(BINARY_NAME)
	@echo "✓ Installed successfully!"
	@echo ""
	@echo "Make sure ~/bin is in your PATH."
	@echo "Add this to your ~/.bashrc or ~/.zshrc if not already present:"
	@echo '  export PATH="$$HOME/bin:$$PATH"'
	@echo ""
	@echo "Then run: source ~/.bashrc  (or ~/.zshrc)"
	@echo "You can now run '$(BINARY_NAME)' from anywhere."

# Uninstall from system path
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PATH)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Uninstalled successfully!"

# Uninstall from user path
uninstall-user:
	@echo "Uninstalling $(BINARY_NAME) from ~/bin..."
	@rm -f ~/bin/$(BINARY_NAME)
	@echo "✓ Uninstalled successfully!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f debug.log
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@$(GO) test ./... -v

# Verify installation
verify:
	@echo "Verifying installation..."
	@which $(BINARY_NAME) > /dev/null && echo "✓ $(BINARY_NAME) is installed at: $$(which $(BINARY_NAME))" || echo "✗ $(BINARY_NAME) not found in PATH"
	@$(BINARY_NAME) --version 2>/dev/null || echo "✗ Cannot execute $(BINARY_NAME)"

# Check prerequisites
check:
	@echo "Checking prerequisites..."
	@which go > /dev/null && echo "✓ Go is installed: $$(go version)" || echo "✗ Go is not installed"
	@which gh > /dev/null && echo "✓ GitHub CLI is installed: $$(gh --version | head -1)" || echo "✗ GitHub CLI (gh) is not installed"
	@gh auth status > /dev/null 2>&1 && echo "✓ GitHub CLI is authenticated" || echo "⚠ GitHub CLI is not authenticated (run: gh auth login)"

# Development build (with debug symbols)
dev:
	@echo "Building $(BINARY_NAME) with debug symbols..."
	@$(GO) build -o $(BINARY_NAME) .
	@echo "✓ Development build complete"

# Show help
help:
	@echo "gsecret Makefile - Available targets:"
	@echo ""
	@echo "  make build          - Build the binary (default)"
	@echo "  make install        - Build and install to /usr/local/bin (requires sudo)"
	@echo "  make install-user   - Build and install to ~/bin (no sudo required)"
	@echo "  make uninstall      - Remove from /usr/local/bin (requires sudo)"
	@echo "  make uninstall-user - Remove from ~/bin"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make verify         - Check if installation is working"
	@echo "  make check          - Check prerequisites (Go, gh CLI)"
	@echo "  make dev            - Build with debug symbols"
	@echo "  make help           - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  make check          # Verify prerequisites"
	@echo "  make install-user   # Install to ~/bin (recommended for first-time)"
	@echo "  gsecret --help      # Test the installation"
	@echo ""
	@echo "For system-wide install:"
	@echo "  make install        # Install to /usr/local/bin (requires sudo)"
