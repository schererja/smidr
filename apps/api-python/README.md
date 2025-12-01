# Smidr REST API (Python)

REST API gateway for the Smidr Yocto build system. This provides HTTP/JSON endpoints that proxy to the underlying gRPC daemon, implemented using FastAPI.

## Quick Start

### 1. Set up Virtual Environment

```bash
cd apps/api-python

# Create virtual environment
python3 -m venv venv

# Activate virtual environment
# On macOS/Linux:
source venv/bin/activate
# On Windows:
# venv\Scripts\activate
```

### 2. Install Dependencies

**Important:** The SDK must be installed as a package for the protobuf imports to work correctly.

```bash
# Install WebAPI dependencies
pip install -r requirements.txt

# Install the SDK in development mode (REQUIRED)
# This ensures the 'generated' package is properly importable
cd ../../sdks/python
pip install -e .
cd ../../apps/api-python
```

**Why?** The generated protobuf files use absolute imports (e.g., `import common_pb2`), which require the SDK to be installed as a package so Python can resolve the imports correctly.

### 3. Run the API

**Option A: Using the startup script**
```bash
./run.sh
```

**Option B: Using uvicorn directly (Recommended)**
```bash
uvicorn smidr_webapi.main:app --host 0.0.0.0 --port 8000 --reload
```

**Option C: Install in development mode**
```bash
pip install -e .
uvicorn smidr_webapi.main:app --host 0.0.0.0 --port 8000 --reload
```

**Option D: Direct Python execution**
```bash
python3 src/smidr_webapi/main.py
```

### 4. Deactivate Virtual Environment

When done:
```bash
deactivate
```

**API will be available at:**
- http://localhost:8000
- OpenAPI docs: http://localhost:8000/swagger (if DEBUG=true)
- ReDoc: http://localhost:8000/redoc (if DEBUG=true)

## Configuration

Configure the gRPC daemon URL and server settings via environment variables or a `.env` file:

```bash
# .env file
SMIDR_DAEMON_URL=http://localhost:50051
HOST=0.0.0.0
PORT=8000
DEBUG=true
CORS_ORIGINS=*
```

Or via environment variables:

```bash
export SMIDR_DAEMON_URL=http://localhost:50051
export PORT=8000
export DEBUG=true
```

## API Endpoints

### Start a Build

```http
POST /api/builds
Content-Type: application/json

{
  "config_path": "smidr.yaml",
  "target": "core-image-minimal",
  "customer": "acme",
  "force_clean": false,
  "force_image_rebuild": false,
  "environment_variables": {
    "MY_VAR": "value"
  }
}
```

**Response:**
```json
{
  "buildId": "build-20231115-120000"
}
```

### Get Build Status

```http
GET /api/builds/{buildId}
```

**Response:**
```json
{
  "build_id": "build-20231115-120000",
  "state": "BUILD_STATE_BUILDING",
  "start_time": "2023-11-15T12:00:00",
  "end_time": null,
  "error_message": null,
  "exit_code": null
}
```

### List All Builds

```http
GET /api/builds?page_size=10&state=BUILD_STATE_COMPLETED&customer=acme
```

**Response:**
```json
{
  "builds": [
    {
      "build_id": "build-20231115-120000",
      "state": "BUILD_STATE_COMPLETED",
      "start_time": "2023-11-15T12:00:00",
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
  "build_id": "build-20231115-120000",
  "artifacts": [
    {
      "name": "core-image-minimal.wic",
      "size_bytes": 1048576,
      "download_url": "http://localhost:8080/artifacts/...",
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
event: log
data: {"stream":"stdout","message":"Starting build...","timestamp":1700049600}

event: log
data: {"stream":"stdout","message":"Building recipe foo...","timestamp":1700049601}
```

**JavaScript Example:**
```javascript
const eventSource = new EventSource('/api/builds/build-123/logs?follow=true');

eventSource.addEventListener('log', (event) => {
  const log = JSON.parse(event.data);
  console.log(`[${log.stream}] ${log.message}`);
});

eventSource.addEventListener('error', (event) => {
  const error = JSON.parse(event.data);
  console.error('Error:', error.error);
  eventSource.close();
});
```

**Python Example:**
```python
import requests
import json

url = "http://localhost:8000/api/builds/build-123/logs?follow=true"
response = requests.get(url, stream=True)

for line in response.iter_lines():
    if line:
        # SSE format: "event: log\ndata: {...}"
        if line.startswith(b"data: "):
            data = json.loads(line[6:])  # Skip "data: "
            print(f"[{data['stream']}] {data['message']}")
```

## Build States

- `BUILD_STATE_UNSPECIFIED` - Unknown state
- `BUILD_STATE_QUEUED` - Waiting to start
- `BUILD_STATE_PREPARING` - Setting up environment
- `BUILD_STATE_BUILDING` - Build in progress
- `BUILD_STATE_EXTRACTING_ARTIFACTS` - Extracting build outputs
- `BUILD_STATE_COMPLETED` - Build succeeded
- `BUILD_STATE_FAILED` - Build failed
- `BUILD_STATE_CANCELLED` - Build was cancelled

## Error Handling

All endpoints return standard HTTP status codes:

- `200 OK` - Success
- `404 Not Found` - Build/resource not found
- `500 Internal Server Error` - Server error

**Error Response:**
```json
{
  "detail": "Build not found"
}
```

## Development

### Setup

```bash
cd apps/api-python

# Create and activate virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Install SDK in development mode
cd ../../sdks/python
pip install -e .
cd ../../apps/api-python

# Install WebAPI in development mode (optional, for easier imports)
pip install -e .
```

### Running in Development Mode

```bash
# With auto-reload
uvicorn smidr_webapi.main:app --reload --host 0.0.0.0 --port 8000

# Or using the main function
python -m smidr_webapi.main
```

### Testing with curl

```bash
# Start a build
curl -X POST http://localhost:8000/api/builds \
  -H "Content-Type: application/json" \
  -d '{"config_path":"smidr.yaml","target":"core-image-minimal"}'

# Get status
curl http://localhost:8000/api/builds/build-123

# List builds
curl "http://localhost:8000/api/builds?page_size=10&state=BUILD_STATE_COMPLETED"

# Stream logs
curl http://localhost:8000/api/builds/build-123/logs?follow=true
```

### Testing with Python SDK

```python
import asyncio
from smidr_sdk import SmidrClient

async def test_api():
    async with SmidrClient("http://localhost:50051") as client:
        # Start build via gRPC
        status = await client.start_build(
            config_path="smidr.yaml",
            target="core-image-minimal"
        )
        print(f"Build started: {status.build_identifier.build_id}")

if __name__ == "__main__":
    asyncio.run(test_api())
```

## Architecture

```
HTTP Client → REST API (this project) → Python SDK → gRPC Daemon (Go)
             ↓                         ↓
         JSON/HTTP                 Protobuf
```

The REST API is a thin translation layer that:
1. Accepts HTTP/JSON requests via FastAPI
2. Calls the gRPC daemon via the Python SDK
3. Transforms responses to JSON using Pydantic models
4. Handles SSE streaming for logs using sse-starlette

## Python vs C# WebAPI Comparison

### Similarities

- Both provide REST endpoints that proxy to gRPC daemon
- Both support Server-Sent Events for log streaming
- Both use dependency injection for client management
- Both provide OpenAPI/Swagger documentation
- Both support CORS configuration

### Differences

- **Framework**: FastAPI (Python) vs ASP.NET Core (C#)
- **Async Model**: Python asyncio vs C# Task-based async
- **Validation**: Pydantic models vs C# data annotations
- **Dependency Injection**: FastAPI's dependency system vs ASP.NET DI container
- **Configuration**: python-dotenv vs ASP.NET Configuration
- **SSE**: sse-starlette library vs native C# SSE support
- **Documentation**: FastAPI auto-generates OpenAPI, C# uses Swashbuckle

### Performance Considerations

- FastAPI is built on Starlette and is one of the fastest Python web frameworks
- C# ASP.NET Core generally has better raw performance
- Both are suitable for production use
- Python may be preferred for rapid development and ecosystem integration
- C# may be preferred for .NET ecosystem integration and performance-critical scenarios

## Dependencies

- **fastapi** - Modern, fast web framework
- **uvicorn** - ASGI server
- **python-dotenv** - Environment variable management
- **sse-starlette** - Server-Sent Events support
- **pydantic** - Data validation using Python type annotations
- **smidr-sdk** - Python SDK for Smidr (local dependency)

## License

MIT License - see the root repository LICENSE file for details.

