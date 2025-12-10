# Cross-Compilation Guide

## Overview

Smidr uses SQLite for build persistence, which requires CGO. Cross-compiling with CGO is more complex than pure Go builds.

## Native Builds (Recommended)

The easiest approach is to build natively on the target platform:

```bash
# On Linux
CGO_ENABLED=1 go build -o smidr ./cmd/smidr/main.go

# On macOS
CGO_ENABLED=1 go build -o smidr ./cmd/smidr/main.go
```

## Cross-Compiling from macOS to Linux

To cross-compile from macOS to Linux with CGO support, you need a cross-compiler toolchain.

### Prerequisites

Install the cross-compilation toolchain:

```bash
# Using Homebrew
brew install FiloSottile/musl-cross/musl-cross

# Or for glibc-based systems
brew install messense/macos-cross-toolchains/x86_64-unknown-linux-gnu
```

### Building for Linux amd64

```bash
# Set the cross-compiler
export CC=x86_64-linux-musl-gcc
export CXX=x86_64-linux-musl-g++

# Build with CGO enabled
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -o build/smidr-linux-amd64 ./cmd/smidr/main.go
```

### Alternative: Docker-based Build

Use Docker to build in a Linux environment:

```bash
# Create a build container
docker run --rm -v "$PWD":/app -w /app golang:1.23 \
  go build -o build/smidr-linux-amd64 ./cmd/smidr/main.go
```

### Alternative: Build on Target Server

The simplest approach for single-platform deployment:

```bash
# Copy source to target server
rsync -av . user@server:/tmp/smidr-src/

# SSH and build natively
ssh user@server "cd /tmp/smidr-src && go build -o smidr ./cmd/smidr/main.go"
```

## CI/CD Recommendations

For CI/CD pipelines:

1. **GitHub Actions**: Use matrix builds with native runners

   ```yaml
   strategy:
     matrix:
       os: [ubuntu-latest, macos-latest]
   ```

2. **Docker-based CI**: Build in Linux containers for consistent results

3. **Native Builds**: Each platform builds its own binary

## Troubleshooting

### Error: `Binary was compiled with 'CGO_ENABLED=0'`

This means the binary was built without CGO support. Rebuild with:

```bash
CGO_ENABLED=1 go build -o smidr ./cmd/smidr/main.go
```

### Cross-compilation Failures

If cross-compilation fails, consider:

- Using Docker for Linux builds
- Building natively on the target platform
- Using GitHub Actions with platform-specific runners
