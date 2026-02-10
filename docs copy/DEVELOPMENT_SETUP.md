# Development Setup Guide

This comprehensive guide helps you set up a productive development environment for Yggdrasil MSP/CRM/ERP platform, supporting both local and cloud-based workflows.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Development Environment Options](#development-environment-options)
4. [Local Development Setup](#local-development-setup)
5. [Cloud IDE Setup](#cloud-ide-setup)
6. [IDE Configuration](#ide-configuration)
7. [Project Structure](#project-structure)
8. [Database Setup](#database-setup)
9. [Development Workflows](#development-workflows)
10. [Testing Setup](#testing-setup)
11. [Debugging Guide](#debugging-guide)
12. [Common Issues](#common-issues)
13. [Performance Optimization](#performance-optimization)

## Prerequisites

### System Requirements

**Minimum Requirements:**

- CPU: 4 cores
- Memory: 8GB RAM
- Storage: 20GB free space
- Internet connection

**Recommended for Full Development:**

- CPU: 8+ cores
- Memory: 16GB+ RAM
- Storage: 50GB+ SSD
- Multiple monitors recommended

### Software Requirements

**Essential Tools:**

```bash
# Version requirements
Node.js >= 20.0.0
Go >= 1.22.0
Docker >= 20.10.0
Docker Compose >= 2.0.0
Git >= 2.30.0
```

**Package Managers:**

- `npm` or `yarn` (for Node.js dependencies)
- `go mod` (Go modules, built-in)
- Homebrew (macOS) or equivalent package manager
- `sqlc` (for SQL code generation)
- `buf` (for protocol buffer compilation)

### Account Setup

**Required Accounts:**

- [GitHub](https://github.com) account with SSH keys configured
- [Docker Hub](https://hub.docker.com) account (for container operations)
- Optional: [NPM](https://www.npmjs.com) account for package publishing

## Quick Start

### 5-Minute Setup (Experienced Developers)

```bash
# 1. Clone repository
git clone https://github.com/intrik8-labs/yggdrasil.git
cd yggdrasil

# 2. Copy environment
cp .env.example .env

# 3. Start infrastructure
docker compose up -d postgres nats

# 4. Run migrations
migrate -path migrations -database "postgres://localhost/yggdrasil?sslmode=disable" up

# 5. Setup Go dependencies and generate code
go mod download
sqlc generate

# 6. Start control plane
go run cmd/api/main.go &

# 7. Setup and start web frontend
cd web
npm ci
npm run dev &
cd ..

# 8. Start agent
go run cmd/agent/main.go &

# 9. Access applications
# Web: http://localhost:3000
# API: http://localhost:8080
# API Docs: http://localhost:8080/swagger (if configured)
```

## Development Environment Options

### 1. Local Development (Recommended for Regular Contributors)

**Advantages:**

- Full control over environment
- Fast local compilation
- Access to all system resources
- Better debugging capabilities

**Best For:**

- Regular contributors
- Complex feature development
- Performance testing
- Offline development

### 2. Cloud IDEs (Recommended for Quick Contributions)

**Supported Platforms:**

- [GitHub Codespaces](https://github.com/features/codespaces)
- [Gitpod](https://gitpod.io/)
- [AWS Cloud9](https://aws.amazon.com/cloud9/)

**Advantages:**

- Zero local setup required
- Consistent environments
- Pre-configured tools
- Easy collaboration

**Best For:**

- New contributors
- Quick fixes
- Code reviews
- Remote development

### 3. Docker Development

**Advantages:**

- Isolated dependencies
- Reproducible builds
- Easy cleanup
- Production-like environment

**Best For:**

- CI/CD pipeline testing
- Cross-platform development
- Dependency management

## Local Development Setup

### Step 1: Install Core Dependencies

**macOS:**

```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew install node@20 go docker docker-compose git sqlc buf
```

**Ubuntu/Debian:**

```bash
# Update package list
sudo apt update

# Install core dependencies
sudo apt install -y curl git build-essential

# Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# Install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install buf
go install github.com/bufbuild/buf/cmd/buf@latest
```

**Windows (WSL2):**

```powershell
# Install WSL2
wsl --install

# Install Ubuntu in WSL2
wsl --install -d Ubuntu

# Inside WSL2 Ubuntu, follow Ubuntu instructions above
```

### Step 2: Install Development Tools

**Go Tools:**

```bash
# Install Go development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add GOPATH/bin to PATH
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

**Node.js Tools:**

```bash
# Install global packages
npm install -g @typescript-eslint/cli prettier typescript

# Verify installations
node --version
npm --version
```

**Docker Setup:**

```bash
# Verify Docker installation
docker --version
docker-compose --version

# Create Docker group (Linux)
sudo usermod -aG docker $USER

# Restart shell or run: newgrp docker
```

### Step 3: Clone and Configure Repository

```bash
# Clone repository
git clone https://github.com/intrik8-labs/yggdrasil.git
cd yggdrasil

# Set up git hooks (if available)
cp scripts/git-hooks/* .git/hooks/
chmod +x .git/hooks/*

# Configure git user (if not configured)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Create development branch
git checkout -b develop
git pull origin develop
```

### Step 4: Environment Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit environment file
nano .env  # or your preferred editor
```

**Key Development Settings:**

```bash
# Development configuration
ENVIRONMENT=development
DEBUG=true
LOG_LEVEL=DEBUG

# Database (will be started with Docker)
DATABASE_URL=postgresql://postgres:password@localhost:5432/yggdrasil_dev

# Local service ports
CONTROL_PLANE_PORT=8080
WEB_PORT=3000

# Development tools
HOT_RELOAD=true
ENABLE_SWAGGER=false
ENABLE_CORS=true
```

### Step 5: Start Development Infrastructure

```bash
# Start only essential services for development
docker compose up -d postgres nats redis

# Verify services are running
docker compose ps

# Check logs if needed
docker compose logs postgres
```

## Development Workflows

### Feature Development Workflow

1. **Create Feature Branch:**

```bash
git checkout develop
git pull origin develop
git checkout -b feature/your-feature-name
```

1. **Make Changes:**

```bash
# Work on your feature
# Make small, focused commits
git add .
git commit -m "feat: implement user authentication"
```

1. **Run Tests:**

```bash
# Run all tests
make test

# Run specific component tests
make test-control-plane
make test-web
make test-agent
```

1. **Format and Lint:**

```bash
# Format code
make format

# Run linters
make lint
```

1. **Push and Create PR:**

```bash
git push origin feature/your-feature-name
# Create pull request on GitHub
```

### Development Commands (Makefile)

```makefile
# Root Makefile
.PHONY: setup-dev test lint format clean

setup-dev:
 @echo "Setting up development environment..."
 cd web && npm ci
 cd . && go mod download
 docker compose up -d postgres nats

test:
 @echo "Running all tests..."
 cd . && go test -v ./...
 cd web && npm test

lint:
 @echo "Running linters..."
 cd . && golangci-lint run
 cd . && gofmt -l .
 cd web && npm run lint

format:
 @echo "Formatting code..."
 cd . && gofmt -w . && goimports -w .
 cd web && npm run format

clean:
 @echo "Cleaning up..."
 docker compose down -v
 cd . && rm -rf tmp
 cd web && rm -rf node_modules
```

### Hot Reload Configuration

**Go with Air:**

```toml
# .air.toml (root level)
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
# Control Plane
args_bin = []
bin = "./tmp/api"
cmd = "go build -o ./tmp/api ./cmd/api/main.go"
delay = 1000
exclude_dir = ["assets", "tmp", "vendor", "testdata", "web"]
exclude_file = []
exclude_regex = ["_test.go"]
exclude_unchanged = false
follow_symlink = false
full_bin = ""
include_dir = ["cmd", "internal"]
include_ext = ["go", "tpl", "tmpl", "html"]
kill_delay = "0s"
log = "build-errors.log"
rerun = false
rerun_delay = 500
send_interrupt = false
stop_on_root = false

[color]
app = ""
build = "yellow"
main = "magenta"
runner = "green"
watcher = "cyan"

[log]
time = false
main_only = false
```

## Testing Setup

### Test Database Configuration

```bash
# Create test database
docker exec -it yggdrasil-dev-db psql -U postgres -c "CREATE DATABASE yggdrasil_test;"

# Run tests with test database
DATABASE_URL=postgresql://postgres:password@localhost:5432/yggdrasil_test go test ./...
```

### Component Testing

**Go (Control Plane & Agent):**

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./internal/ticket

# Watch mode for development (with entr or air)
find . -name '*.go' | entr -c go test ./...

# Integration tests
go test -tags=integration ./...
```

**JavaScript/TypeScript (Web):**

```bash
cd web

# Run tests
npm test

# Run with coverage
npm run test:coverage

# Watch mode
npm run test:watch

# E2E tests
npm run test:e2e
```

go test ./...

# Run with coverage

go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests

go test -v ./internal/ticket

# Run integration tests

go test -tags=integration ./...

# Race condition testing

go test -race ./...

````

### Test Data Management

```go
// internal/platform/database/testing.go
package database

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()

    // Start postgres container
    req := testcontainers.ContainerRequest{
        Image:        "postgres:16-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_DB":       "testdb",
            "POSTGRES_PASSWORD": "password",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp"),
    }

    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        t.Fatalf("failed to start container: %v", err)
    }

    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    // Get connection string
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    connStr := fmt.Sprintf("postgres://postgres:password@%s:%s/testdb",
        host, port.Port())

    // Connect
    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        t.Fatalf("failed to connect: %v", err)
    }

    return pool
}
````

## Debugging Guide

### Go Debugging (VS Code)

**Launch Configuration (.vscode/launch.json):**

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Control Plane",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/api/main.go",
      "cwd": "${workspaceFolder}",
      "env": {
        "ENVIRONMENT": "development",
        "DEBUG": "true"
      },
      "args": []
    },
    {
      "name": "Debug Agent",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/agent/main.go",
      "cwd": "${workspaceFolder}",
      "env": {
        "ENVIRONMENT": "development"
      }
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/ticket",
      "args": ["-test.v"]
    }
  ]
}
```

**Debugging Tips:**

```go
// Use delve for debugging
dlv debug ./cmd/api/main.go

// Run with race detector
go run -race ./cmd/api/main.go

// Print debugging (structured logging)
slog.Debug("debugging info", "ticket_id", ticketID, "value", value)
```

### Frontend Debugging

**React DevTools:**

1. Install React Developer Tools browser extension
2. Open browser developer tools
3. Use "Components" and "Profiler" tabs

**Network Debugging:**

```bash
# Monitor API calls
curl -v http://localhost:8080/api/v1/users

# Check CORS headers
curl -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: X-Requested-With" \
     -X OPTIONS http://localhost:8080/api/v1/users
```

### Database Debugging

**Connection Issues:**

```bash
# Test database connection
docker exec -it yggdrasil-dev-db psql -U postgres -d yggdrasil_dev -c "SELECT version();"

# Check active connections
docker exec -it yggdrasil-dev-db psql -U postgres -d yggdrasil_dev -c "SELECT * FROM pg_stat_activity;"

# Kill hanging connections
docker exec -it yggdrasil-dev-db psql -U postgres -d yggdrasil_dev -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'active' AND pid <> pg_backend_pid();"
```

**Query Performance:**

```sql
-- Explain slow queries
EXPLAIN ANALYZE SELECT * FROM tickets WHERE status = 'open';

-- Check indexes
SELECT * FROM pg_indexes WHERE tablename = 'tickets';

-- Analyze table statistics
ANALYZE tickets;
```

## Common Issues

### Port Conflicts

**Problem:** Services already running on ports 3000, 8000, 5432

**Solution:**

```bash
# Find process using port
lsof -i :3000
lsof -i :8080
lsof -i :5432

# Kill process
kill -9 <PID>

# Or change ports in .env
WEB_PORT=3001
CONTROL_PLANE_PORT=8001
```

### Go Module Issues

**Problem:** Go module dependencies not resolving

**Solution:**

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify Go version
go version

# List modules
go list -m all
```

### Node.js Dependency Issues

**Problem:** npm install failures

**Solution:**

```bash
# Clear npm cache
npm cache clean --force

# Remove node_modules and package-lock.json
rm -rf node_modules package-lock.json

# Reinstall
npm install

# Or try yarn
npm install -g yarn
yarn install
```

### Docker Issues

**Problem:** Container startup failures

**Solution:**

```bash
# Check container logs
docker compose logs postgres
docker compose logs nats

# Reset containers
docker compose down -v
docker system prune -f
docker compose up -d

# Check Docker resources
docker system df
docker system events
```

### Database Migration Issues

**Problem:** Alembic migration failures

**Solution:**

```bash
# Check current migration status
migrate -path migrations -database "${DATABASE_URL}" version

# Check migration history
ls -la migrations/

# Force migration version (if manually applied)
migrate -path migrations -database "${DATABASE_URL}" force <version>

# Create new migration
migrate create -ext sql -dir migrations -seq <name>
```

## Performance Optimization

### Local Development Performance

**Hot Reload Optimization:**

```bash
# Use air for hot reload with optimized settings
# .air.toml configuration excludes unnecessary directories
air
```

**Database Optimization:**

```sql
-- Development database optimizations
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET work_mem = '4MB';
SELECT pg_reload_conf();
```

### Build Performance

**Parallel Builds:**

```bash
# Go parallel builds
go build -p 4 ./cmd/agent

# TypeScript parallel compilation
cd web
npm run build -- --parallel

# Docker multi-stage builds
docker build --target development .
```

**Caching Strategies:**

```bash
# Docker layer caching
docker build --cache-from yggdrasil/control-plane:dev .

# npm cache
npm ci --prefer-offline --no-audit

# Go module cache
export GOCACHE=$HOME/.cache/go-build
export GOMODCACHE=$HOME/.cache/go-mod
```

---

This comprehensive setup guide ensures you can develop productively across all components of the Yggdrasil platform. For additional help or questions, open an issue on GitHub with the "development" label.
