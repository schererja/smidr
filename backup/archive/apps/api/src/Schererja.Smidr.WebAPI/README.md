# Smidr REST API

REST API gateway for the Smidr Yocto build system. This provides HTTP/JSON endpoints that proxy to the underlying gRPC daemon.

## Quick Start

```bash
# Restore dependencies
dotnet restore

# Run the API
dotnet run

# API will be available at:
# - https://localhost:5001 (HTTPS)
# - http://localhost:5000 (HTTP)
# - OpenAPI docs: https://localhost:5001/openapi/v1.json
```

## Configuration

Configure the gRPC daemon URL in `appsettings.json`:

```json
{
  "Smidr": {
    "DaemonUrl": "http://localhost:50051"
  }
}
```

Or via environment variable:
```bash
export Smidr__DaemonUrl=http://localhost:50051
```

## API Endpoints

### Start a Build
```http
POST /api/builds
Content-Type: application/json

{
  "configPath": "smidr.yaml",
  "target": "core-image-minimal",
  "customer": "acme",
  "forceClean": false,
  "forceImageRebuild": false
}
```

**Response:**
```json
{
  "buildId": "build-20231115-120000",
  "message": "Build started successfully"
}
```

### Get Build Status
```http
GET /api/builds/{buildId}
```

**Response:**
```json
{
  "buildId": "build-20231115-120000",
  "state": "BuildStateBuilding",
  "startTime": "2023-11-15T12:00:00Z",
  "endTime": null,
  "errorMessage": "",
  "exitCode": 0
}
```

### List All Builds
```http
GET /api/builds?pageSize=10&state=BuildStateCompleted
```

**Response:**
```json
{
  "builds": [
    {
      "buildId": "build-20231115-120000",
      "state": "BuildStateCompleted",
      "startTime": "2023-11-15T12:00:00Z",
      "target": "core-image-minimal"
    }
  ]
}
```

### Cancel a Build
```http
DELETE /api/builds/{buildId}
```

**Response:**
```json
{
  "success": true,
  "message": "Build cancelled"
}
```

### List Build Artifacts
```http
GET /api/builds/{buildId}/artifacts
```

**Response:**
```json
{
  "buildId": "build-20231115-120000",
  "artifacts": [
    {
      "name": "core-image-minimal.wic",
      "path": "/path/to/artifact",
      "sizeBytes": 1048576,
      "downloadUrl": "http://localhost:8080/artifacts/...",
      "checksum": "sha256:..."
    }
  ]
}
```

### Stream Build Logs (Server-Sent Events)
```http
GET /api/builds/{buildId}/logs?follow=true
```

**Response (SSE stream):**
```
data: {"stream":"stdout","message":"Starting build...","timestamp":1700049600}

data: {"stream":"stdout","message":"Building recipe foo...","timestamp":1700049601}
```

**JavaScript Example:**
```javascript
const eventSource = new EventSource('/api/builds/build-123/logs?follow=true');

eventSource.onmessage = (event) => {
  const log = JSON.parse(event.data);
  console.log(`[${log.stream}] ${log.message}`);
};

eventSource.onerror = () => {
  console.error('Connection lost');
  eventSource.close();
};
```

## Build States

- `BuildStateUnspecified` - Unknown state
- `BuildStateQueued` - Waiting to start
- `BuildStatePreparing` - Setting up environment
- `BuildStateBuilding` - Build in progress
- `BuildStateExtractingArtifacts` - Extracting build outputs
- `BuildStateCompleted` - Build succeeded
- `BuildStateFailed` - Build failed
- `BuildStateCancelled` - Build was cancelled

## Error Handling

All endpoints return standard HTTP status codes:

- `200 OK` - Success
- `404 Not Found` - Build/resource not found
- `500 Internal Server Error` - Server error

**Error Response:**
```json
{
  "error": "Build not found"
}
```

## Development

### Running with Docker
```bash
docker build -t smidr-api .
docker run -p 5000:8080 -e Smidr__DaemonUrl=http://host.docker.internal:50051 smidr-api
```

### Testing with curl
```bash
# Start a build
curl -X POST http://localhost:5000/api/builds \
  -H "Content-Type: application/json" \
  -d '{"configPath":"smidr.yaml","target":"core-image-minimal"}'

# Get status
curl http://localhost:5000/api/builds/build-123

# Stream logs
curl http://localhost:5000/api/builds/build-123/logs?follow=true
```

## Architecture

```
HTTP Client → REST API (this project) → gRPC Client (SmidrLib) → gRPC Daemon (Go)
             ↓                         ↓
         JSON/HTTP                 Protobuf
```

The REST API is a thin translation layer that:
1. Accepts HTTP/JSON requests
2. Calls the gRPC daemon via SmidrLib
3. Transforms responses to JSON
4. Handles SSE streaming for logs

## License

MIT License - see the root repository LICENSE file for details.
