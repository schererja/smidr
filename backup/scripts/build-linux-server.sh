#!/bin/bash
set -e

# Build server on Linux (with CGO support) using Docker
# This avoids cross-compilation issues with CGO

DOCKER_IMAGE="golang:1.25-alpine"
OUTPUT_BIN="bin/smidr-server-linux-amd64"

mkdir -p bin

# Build using Docker with Alpine Linux
docker run --rm -v "$(pwd)":/workspace -w /workspace "$DOCKER_IMAGE" sh -c '
  apk add --no-cache gcc musl-dev sqlite-dev
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o '"$OUTPUT_BIN"' -ldflags="" ./cmd/smidr-core
'

echo "Built: $OUTPUT_BIN"
file "$OUTPUT_BIN"
