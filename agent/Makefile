.PHONY: help build build-linux build-darwin test test-verbose test-coverage clean install fmt vet lint check deps all

# Binary name
BINARY_NAME=smidr-agent
BINARY_PATH=bin/$(BINARY_NAME)

# Build variables
GO=go
GOFMT=gofmt
GOVET=$(GO) vet
GOTEST=$(GO) test
GOBUILD=$(GO) build

# Version info (can be overridden)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Test flags
TEST_FLAGS=-v -race -timeout 30s
COVERAGE_FILE=coverage.out

# Default target
all: clean deps fmt vet test build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the agent binary for current OS
	@echo "Building $(BINARY_NAME) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/agent
	@echo "Binary created at $(BINARY_PATH)"

build-linux: ## Build the agent binary for Linux x86_64
	@echo "Building $(BINARY_NAME) for linux/amd64..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/agent
	@echo "Binary created at bin/$(BINARY_NAME)-linux-amd64"

build-darwin: ## Build the agent binary for macOS (arm64 and amd64)
	@echo "Building $(BINARY_NAME) for darwin/amd64..."
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/agent
	@echo "Binary created at bin/$(BINARY_NAME)-darwin-amd64"
	@echo "Building $(BINARY_NAME) for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/agent
	@echo "Binary created at bin/$(BINARY_NAME)-darwin-arm64"

build-all: build-linux build-darwin ## Build binaries for all supported platforms

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) $(TEST_FLAGS) ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report generated at $(COVERAGE_FILE)"
	@$(GO) tool cover -func=$(COVERAGE_FILE)

test-coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "HTML coverage report generated at coverage.html"

fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) -w -s .

fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Code is not formatted. Run 'make fmt'" && $(GOFMT) -l . && exit 1)

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run ./...

check: fmt-check vet ## Run all checks (format, vet)

clean: ## Remove build artifacts and test cache
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f $(COVERAGE_FILE) coverage.html
	@$(GO) clean -testcache

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

install: build ## Install the agent binary to /usr/local/bin (requires sudo)
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo install -m 755 $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed successfully"

run-init: build ## Run agent init command
	@echo "Running $(BINARY_NAME) init..."
	@$(BINARY_PATH) init

run-daemon: build ## Run agent daemon (requires config)
	@echo "Running $(BINARY_NAME) daemon..."
	@$(BINARY_PATH) daemon

dev: clean deps fmt vet test build ## Development workflow: clean, deps, format, vet, test, build
	@echo "Development build complete"

ci: deps fmt-check vet test-coverage ## CI workflow: deps, format check, vet, test with coverage
	@echo "CI checks passed"
