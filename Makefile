# MD-to-PDF Makefile

.PHONY: all build test clean install plugins docs

# Build configuration
BINARY_NAME=md-to-pdf
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: build

# Build the main binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build with race detection for development
build-dev:
	go build -race $(LDFLAGS) -o $(BINARY_NAME) .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
test-race:
	go test -race -v ./...

# Build example plugins
plugins:
	$(MAKE) -C examples/plugins all

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(MAKE) -C examples/plugins clean

# Install binary to system
install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Install for development (with plugins)
install-dev: build plugins
	sudo cp $(BINARY_NAME) /usr/local/bin/
	mkdir -p ~/.local/share/md-to-pdf/plugins
	cp examples/plugins/*.so ~/.local/share/md-to-pdf/plugins/ 2>/dev/null || true

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Vet code
vet:
	go vet ./...

# Check dependencies
deps:
	go mod tidy
	go mod verify

# Security scan
security:
	gosec ./...

# Generate documentation
docs:
	@echo "Documentation generated in README.md and CONTRIBUTING.md"

# Release build (cross-platform)
release:
	mkdir -p dist
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

# Docker build
docker:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Development workflow
dev: clean fmt vet test build plugins

# CI workflow
ci: deps fmt vet lint test-race test-coverage

# Help
help:
	@echo "Available targets:"
	@echo "  all          - Build the binary (default)"
	@echo "  build        - Build the binary"
	@echo "  build-dev    - Build with race detection"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  test-race    - Run tests with race detection"
	@echo "  plugins      - Build example plugins"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install binary to system"
	@echo "  install-dev  - Install for development"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  vet          - Vet code"
	@echo "  deps         - Check and tidy dependencies"
	@echo "  security     - Run security scan"
	@echo "  docs         - Generate documentation"
	@echo "  release      - Build for all platforms"
	@echo "  docker       - Build Docker image"
	@echo "  dev          - Development workflow"
	@echo "  ci           - CI workflow"
	@echo "  help         - Show this help"