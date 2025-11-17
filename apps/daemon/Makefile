# Makefile for smidr - Embedded Linux Build Tool
.PHONY: build test clean install fmt lint help run dev deps check coverage yocto-smoke yocto-fetch yocto-sstate

# Variables
BINARY_NAME=smidr
CMD_DIR=cmd/smidr
BUILD_DIR=build
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: build

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "✅ Built $(BINARY_NAME)"

# Build for multiple platforms
# Note: CGO is enabled for SQLite support (go-sqlite3 requires it)
# macOS builds natively, Linux build uses Docker for cross-compilation
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@echo "🍎 Building for macOS..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/main.go
	@echo "🐧 Building for Linux (using Docker)..."
	@$(MAKE) build-linux-docker
	@echo "✅ Built binaries in $(BUILD_DIR)/"

# Build Linux binary using Docker (works from any platform)
build-linux-docker:
	@echo "🐳 Building Linux binary in Docker..."
	@mkdir -p $(BUILD_DIR)
	docker run --rm \
		--platform linux/amd64 \
		-v "$(PWD)":/app \
		-w /app \
		-e CGO_ENABLED=1 \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		golang:1.25.1 \
		go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total: | awk '{print "📊 Total coverage: " $$3}'
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage HTML report: coverage.html"

# Run tests in specific package
test-pkg:
	@echo "🧪 Running tests for specific package..."
	@read -p "Enter package (e.g., ./internal/config): " pkg; \
	go test -v $$pkg

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean
	@echo "✅ Cleaned"

# Install binary to GOPATH/bin
install:
	@echo "📦 Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(CMD_DIR)/main.go
	@echo "✅ Installed $(BINARY_NAME) to $(shell go env GOPATH)/bin"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi
	@echo "✅ Linting complete"

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies updated"

# Run development build with example config
dev: build
	@echo "🚀 Running development build..."
	@if [ -f "smidr-example.yaml" ]; then \
		./$(BINARY_NAME) --config smidr-example.yaml build --help; \
	else \
		./$(BINARY_NAME) --help; \
	fi

# Quick development workflow
check: fmt lint test
	@echo "✅ All checks passed"

# Run smidr with arguments
run: build
	@echo "🚀 Running $(BINARY_NAME) with args: $(ARGS)"
	./$(BINARY_NAME) $(ARGS)

# Example: make run ARGS="build --target my-image"
run-build: build
	./$(BINARY_NAME) build $(ARGS)

run-init: build
	./$(BINARY_NAME) init $(ARGS)

run-status: build
	./$(BINARY_NAME) status $(ARGS)

# Docker-related targets (if needed for container testing)
docker-build:
	@echo "🐳 Building test container..."
	docker build -t smidr-test -f docker/Dockerfile.test .

# Release preparation (tags, builds all platforms)
release: clean check build-all
	@echo "🚀 Release preparation complete"
	@echo "📦 Binaries available in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Help target
help:
	@echo "Smidr Build System"
	@echo "=================="
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build the binary (default)"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  test         - Run all tests"
	@echo "  coverage     - Run tests with coverage report (prints percent and generates coverage.html)"
	@echo "  test-pkg     - Run tests for specific package (interactive)"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code (requires golangci-lint)"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  check        - Run fmt, lint, and test"
	@echo "  dev          - Build and run with example config"
	@echo "  run          - Build and run with ARGS"
	@echo "  run-build    - Build and run 'smidr build' with ARGS"
	@echo "  run-init     - Build and run 'smidr init' with ARGS"
	@echo "  run-status   - Build and run 'smidr status' with ARGS"
	@echo "  release      - Prepare release (clean, check, build-all)"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make run ARGS='--help'"
	@echo "  make run-build ARGS='--target my-image'"
	@echo "  make check    # fmt + lint + test"

# --- Fast Yocto CI tiers ---
# Tier 0: container+env smoke (no network, no build). Validates container starts and bitbake is callable.
yocto-smoke: build
	@echo "🚦 Yocto smoke: parse-only in container (no network/build)"
	SMIDR_TEST_ENTRYPOINT='sh,-c,sleep 3600' \
	SMIDR_TEST_WRITE_MARKERS=1 \
	./$(BINARY_NAME) $(ARGS) build --customer ci --target core-image-minimal
	@echo "✅ Smoke complete"

# Tier 1: fetch-only with preseeded downloads (kept fast via restored DL_DIR). No compile.
yocto-fetch: build
	@echo "📦 Yocto fetch-only: download sources using mirrors/cache"
	SMIDR_TEST_ENTRYPOINT='sh,-c,sleep 3600' \
	SMIDR_TEST_WRITE_MARKERS=1 \
	./$(BINARY_NAME) $(ARGS) build --customer ci --target core-image-minimal --fetch-only
	@echo "✅ Fetch-only complete"

# Tier 2: sstate-restore build (fast if SSTATE_MIRRORS hits). Intended for CI where caches are restored.
yocto-sstate: build
	@echo "🧩 Yocto sstate build: attempts bitbake with sstate restore"
	./$(BINARY_NAME) $(ARGS) build --customer ci --target core-image-minimal
	@echo "✅ Sstate tier invoked (ensure SSTATE_MIRRORS via smidr.yaml/local.conf)"