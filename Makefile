.PHONY: build test clean install lint fmt docker-build pre-commit-install pre-commit-uninstall pre-commit-run pre-commit-update

# Variables
BINARY_NAME=cultivator
GO=go
GOOS=linux
GOARCH=amd64
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Git info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build the binary
build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/cultivator

# Install the binary
install:
	$(GO) install $(GOFLAGS) $(LDFLAGS) ./cmd/cultivator

# Run tests
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
coverage: test
	$(GO) tool cover -html=coverage.out

# Format code
fmt:
	$(GO) fmt ./...
	gofmt -s -w .

# Lint code
lint:
	golangci-lint run ./...

# Vet code
vet:
	$(GO) vet ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	$(GO) clean

# Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Run all checks
check: fmt vet lint test coverage

# Build Docker image
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		-t cultivator:$(VERSION) \
		-t cultivator:latest \
		.

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/cultivator
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/cultivator
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/cultivator
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/cultivator

# Install pre-commit hooks
pre-commit-install:
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || { echo "Error: pre-commit not found. Install with: pip install pre-commit"; exit 1; }
	pre-commit install --install-hooks
	pre-commit install --hook-type commit-msg
	@echo "Pre-commit hooks installed successfully!"

# Uninstall pre-commit hooks
pre-commit-uninstall:
	@echo "Uninstalling pre-commit hooks..."
	pre-commit uninstall
	pre-commit uninstall --hook-type commit-msg
	@echo "Pre-commit hooks uninstalled."

# Run pre-commit on all files
pre-commit-run:
	@echo "Running pre-commit on all files..."
	pre-commit run --all-files

# Update pre-commit hooks to latest versions
pre-commit-update:
	@echo "Updating pre-commit hooks..."
	pre-commit autoupdate
	@echo "Pre-commit hooks updated."

# Setup development environment
setup-dev: deps pre-commit-install
	@echo "Development environment setup complete!"
	@echo "Run 'make check' to verify everything is working."

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  install    - Install the binary to GOPATH/bin"
	@echo "  test       - Run tests"
	@echo "  coverage   - Run tests and show coverage report"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  vet        - Vet code"
	@echo "  clean      - Remove build artifacts"
	@echo "  deps       - Download dependencies"
	@echo "  check        - Run all checks (fmt, vet, lint, test)"
	@echo "  docker-build - Build Docker image"
	@echo "  build-all    - Build for all platforms"
	@echo ""
	@echo "Pre-commit hooks:"
	@echo "  pre-commit-install   - Install pre-commit hooks"
	@echo "  pre-commit-uninstall - Uninstall pre-commit hooks"
	@echo "  pre-commit-run       - Run pre-commit on all files"
	@echo "  pre-commit-update    - Update pre-commit hooks"
	@echo "  setup-dev            - Setup development environment"
