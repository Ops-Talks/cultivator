#!/bin/bash
# Development environment setup script for Cultivator
# This script sets up pre-commit hooks and development dependencies

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        print_success "$1 is installed"
        return 0
    else
        print_error "$1 is not installed"
        return 1
    fi
}

# Main setup
echo "================================================"
echo "  Cultivator Development Environment Setup"
echo "================================================"
echo ""

# Check required tools
print_info "Checking required tools..."
echo ""

MISSING_TOOLS=0

# Check Go
if check_command go; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "  Version: $GO_VERSION"
else
    ((MISSING_TOOLS++))
    echo "  Install: https://golang.org/dl/"
fi
echo ""

# Check Git
if check_command git; then
    GIT_VERSION=$(git --version | awk '{print $3}')
    echo "  Version: $GIT_VERSION"
else
    ((MISSING_TOOLS++))
fi
echo ""

# Check Make
if check_command make; then
    MAKE_VERSION=$(make --version | head -n1 | awk '{print $3}')
    echo "  Version: $MAKE_VERSION"
else
    ((MISSING_TOOLS++))
fi
echo ""

# Exit if required tools are missing
if [ $MISSING_TOOLS -gt 0 ]; then
    print_error "Missing $MISSING_TOOLS required tool(s). Please install them and try again."
    exit 1
fi

print_success "All required tools are installed!"
echo ""

# Install optional tools
print_info "Checking optional tools..."
echo ""

# Check Python (for pre-commit)
if ! check_command python3 && ! check_command python; then
    print_error "Python is not installed (required for pre-commit)"
    echo "  Install: https://www.python.org/downloads/"
    exit 1
fi
echo ""

# Check pip
if ! check_command pip3 && ! check_command pip; then
    print_error "pip is not installed (required for pre-commit)"
    echo "  Install: python -m ensurepip --upgrade"
    exit 1
fi
echo ""

# Install pre-commit
print_info "Installing pre-commit..."
if check_command pre-commit; then
    PRECOMMIT_VERSION=$(pre-commit --version | awk '{print $2}')
    echo "  Already installed: $PRECOMMIT_VERSION"
else
    if pip3 install --user pre-commit 2>/dev/null || pip install --user pre-commit 2>/dev/null; then
        print_success "pre-commit installed successfully"
        
        # Add to PATH if needed
        if ! command -v pre-commit >/dev/null 2>&1; then
            print_info "Adding pre-commit to PATH..."
            export PATH="$HOME/.local/bin:$PATH"
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc 2>/dev/null || true
        fi
    else
        print_error "Failed to install pre-commit"
        exit 1
    fi
fi
echo ""

# Check golangci-lint
print_info "Checking golangci-lint..."
if check_command golangci-lint; then
    GOLANGCI_VERSION=$(golangci-lint --version | head -n1 | awk '{print $4}')
    echo "  Version: $GOLANGCI_VERSION"
else
    print_info "golangci-lint not found. Installing..."
    if curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin; then
        print_success "golangci-lint installed successfully"
    else
        print_error "Failed to install golangci-lint"
        echo "  Manual install: https://golangci-lint.run/usage/install/"
    fi
fi
echo ""

# Download Go dependencies
print_info "Downloading Go dependencies..."
if make deps; then
    print_success "Go dependencies downloaded"
else
    print_error "Failed to download Go dependencies"
    exit 1
fi
echo ""

# Install pre-commit hooks
print_info "Installing pre-commit hooks..."
if make pre-commit-install; then
    print_success "Pre-commit hooks installed"
else
    print_error "Failed to install pre-commit hooks"
    exit 1
fi
echo ""

# Verify setup
print_info "Verifying setup..."
echo ""

if go build -o /tmp/cultivator-test ./cmd/cultivator >/dev/null 2>&1; then
    print_success "Build test passed"
    rm -f /tmp/cultivator-test
else
    print_error "Build test failed"
    exit 1
fi
echo ""

# Summary
echo "================================================"
echo "  Setup Complete!"
echo "================================================"
echo ""
echo "Next steps:"
echo "  1. Run 'make check' to verify everything works"
echo "  2. Read CONTRIBUTING.md for development guidelines"
echo "  3. Start coding! Pre-commit hooks will run automatically"
echo ""
echo "Useful commands:"
echo "  make build           - Build the binary"
echo "  make test            - Run tests"
echo "  make check           - Run all checks (fmt, vet, lint, test)"
echo "  make pre-commit-run  - Manually run pre-commit hooks"
echo "  make help            - Show all available commands"
echo ""
print_success "Happy coding!"
