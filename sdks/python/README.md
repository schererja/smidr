# Smidr Python SDK

Python client library for interacting with the Smidr gRPC daemon.

## Installation

### Using Virtual Environment (Recommended)

```bash
cd sdks/python

# Create virtual environment
python3 -m venv venv

# Activate virtual environment
# On macOS/Linux:
source venv/bin/activate
# On Windows:
# venv\Scripts\activate

# Install in development mode
pip install -e .

# Or install dependencies directly
pip install -r requirements.txt
```

### Without Virtual Environment

```bash
cd sdks/python
pip install -e .
```

### From PyPI (when published)

```bash
pip install smidr-sdk
```

## Quick Start

```python
import asyncio
from smidr_sdk import SmidrClient
from generated.smidr.v1 import common_pb2

async def main():
    # Create a client connected to the daemon
    async with SmidrClient("http://localhost:50051") as client:
        # Start a new build
        status = await client.start_build(
            config_path="smidr.yaml",
            target="core-image-minimal",
            customer="acme"
        )

        print(f"Build started: {status.build_identifier.build_id}")

        # Monitor build status
        current_status = await client.get_build_status(status.build_identifier.build_id)
        print(f"State: {current_status.state}")

        # Stream logs in real-time
        async for log_entry in client.stream_logs(
            status.build_identifier.build_id,
            follow=True
        ):
            print(f"[{log_entry.stream}] {log_entry.message}")

        # List artifacts when complete
        artifacts = await client.list_artifacts(status.build_identifier.build_id)
        for artifact in artifacts.artifacts:
            print(f"Artifact: {artifact.name} ({artifact.size_bytes} bytes)")
            print(f"  Download: {artifact.download_url}")

if __name__ == "__main__":
    asyncio.run(main())
```

## API Overview

### SmidrClient

The main client class providing high-level methods for interacting with the daemon.

#### Constructor

```python
client = SmidrClient("http://localhost:50051")
```

The client supports async context manager for automatic cleanup:

```python
async with SmidrClient("http://localhost:50051") as client:
    # Use client
    pass
# Channel is automatically closed
```

#### Methods

##### start_build

Start a new build with the specified configuration.

```python
response = await client.start_build(
    config_path="smidr.yaml",
    target="core-image-minimal",
    customer="acme",              # optional
    force_clean=False,            # optional
    force_image_rebuild=False,    # optional
    environment_variables={        # optional
        "MY_VAR": "value"
    }
)
```

##### get_build_status

Get the current status of a build.

```python
status = await client.get_build_status("build-123")
print(f"State: {status.state}")
from datetime import datetime
start_time = datetime.fromtimestamp(status.timestamps.start_time_unix_seconds)
print(f"Started: {start_time}")
```

##### list_builds

List all builds, optionally filtered by state.

```python
# List all builds
all_builds = await client.list_builds()

# List only running builds
from generated.smidr.v1 import common_pb2
running_builds = await client.list_builds(
    states=[common_pb2.BUILD_STATE_BUILDING]
)

# List with page size limit
recent_builds = await client.list_builds(page_size=10)

# Filter by customer
customer_builds = await client.list_builds(customer="acme")
```

##### cancel_build

Cancel a running build.

```python
response = await client.cancel_build("build-123")
print(f"Cancelled: {response.success}")
```

##### list_artifacts

List artifacts from a completed build.

```python
artifacts = await client.list_artifacts("build-123")
for artifact in artifacts.artifacts:
    print(f"{artifact.name}: {artifact.download_url}")
```

##### stream_logs

Stream logs from a build in real-time.

```python
async for log_entry in client.stream_logs("build-123", follow=True):
    if log_entry.stream == "stderr":
        print(f"ERROR: {log_entry.message}", file=sys.stderr)
    else:
        print(log_entry.message)
```

### Direct Service Access

For advanced scenarios, you can access the underlying gRPC service clients directly:

```python
# Direct access to BuildService
from generated.smidr.v1 import builds_pb2, common_pb2

build_details = await client.builds.GetBuild(
    builds_pb2.GetBuildRequest(
        build_identifier=common_pb2.BuildIdentifier(build_id="build-123")
    )
)

# Direct access to LogService
logs_stream = client.logs.StreamBuildLogs(
    logs_pb2.StreamBuildLogsRequest(
        build_identifier=common_pb2.BuildIdentifier(build_id="build-123"),
        follow=True
    )
)

# Direct access to ArtifactService
artifacts_response = await client.artifacts.ListArtifacts(
    artifacts_pb2.ListArtifactsRequest(
        build_identifier=common_pb2.BuildIdentifier(build_id="build-123")
    )
)
```

## Build States

The `BuildState` enum represents the possible states of a build:

- `BUILD_STATE_UNSPECIFIED` - Default/unknown state
- `BUILD_STATE_QUEUED` - Build is queued and waiting to start
- `BUILD_STATE_PREPARING` - Build environment is being prepared
- `BUILD_STATE_BUILDING` - Build is actively running
- `BUILD_STATE_EXTRACTING_ARTIFACTS` - Build completed, extracting artifacts
- `BUILD_STATE_COMPLETED` - Build completed successfully
- `BUILD_STATE_FAILED` - Build failed
- `BUILD_STATE_CANCELLED` - Build was cancelled

## Timestamps

Build timestamps are provided as Unix seconds (int64) in the `TimeStampRange` message:

```python
from datetime import datetime

start_time = datetime.fromtimestamp(status.timestamps.start_time_unix_seconds)
end_time = datetime.fromtimestamp(status.timestamps.end_time_unix_seconds)
duration = end_time - start_time
print(f"Build took {duration.total_seconds() / 60:.2f} minutes")
```

## Error Handling

All async methods can raise `grpc.RpcError` for gRPC errors:

```python
import grpc

try:
    status = await client.get_build_status("invalid-build-id")
except grpc.RpcError as e:
    if e.code() == grpc.StatusCode.NOT_FOUND:
        print("Build not found")
    else:
        print(f"RPC error: {e.details()}")
```

## Development

### Building the SDK

```bash
cd sdks/python
pip install -e .
```

### Regenerating from Proto Files

When the proto files change, regenerate the Python code:

```bash
cd protos
buf generate
```

The generated files are placed in `sdks/python/generated/` and should be automatically included in the SDK.

### Running Tests

```bash
cd sdks/python
pytest
```

## Example: Full Build Workflow

```python
import asyncio
import sys
from smidr_sdk import SmidrClient
from generated.smidr.v1 import common_pb2

async def run_build():
    async with SmidrClient("http://localhost:50051") as client:
        # Start the build
        start_response = await client.start_build(
            config_path="smidr.yaml",
            target="core-image-minimal",
            customer="acme"
        )

        build_id = start_response.build_identifier.build_id
        print(f"✅ Build started: {build_id}")

        # Stream logs with cancellation support
        try:
            async for log in client.stream_logs(build_id, follow=True):
                print(log.message)
        except KeyboardInterrupt:
            print("\n⚠️  Log streaming cancelled")

        # Check final status
        final_status = await client.get_build_status(build_id)
        if final_status.state == common_pb2.BUILD_STATE_COMPLETED:
            print("✅ Build completed successfully!")

            # List and download artifacts
            artifacts = await client.list_artifacts(build_id)
            for artifact in artifacts.artifacts:
                print(f"📦 {artifact.name}")
                print(f"   Size: {artifact.size_bytes / 1024 / 1024:.2f} MB")
                print(f"   URL: {artifact.download_url}")
                print(f"   Checksum: {artifact.checksum}")
        elif final_status.state == common_pb2.BUILD_STATE_FAILED:
            print(f"❌ Build failed: {final_status.error_message}")
            print(f"   Exit code: {final_status.exit_code}")

if __name__ == "__main__":
    asyncio.run(run_build())
```

## Python vs C# Comparison

### Similarities

- Both use async/await for asynchronous operations
- Both provide high-level client methods wrapping gRPC calls
- Both support direct access to underlying service clients
- Both support streaming logs via async iteration

### Differences

- **Context Management**: Python uses `async with` for resource cleanup, C# uses `using` statements
- **Type System**: Python uses type hints (optional), C# has static typing
- **Error Handling**: Python raises `grpc.RpcError`, C# throws `RpcException`
- **Streaming**: Python uses async generators (`async for`), C# uses `IAsyncEnumerable`
- **Naming**: Python uses `snake_case` for methods, C# uses `PascalCase` with `Async` suffix
- **Imports**: Python imports are more explicit about generated code location

### Performance Considerations

- Both implementations use the same underlying gRPC protocol
- Python's async/await is built on asyncio, C# uses Task-based async
- Performance characteristics are similar, with C# typically having slightly better raw performance
- Python's ecosystem and ease of use may be preferable for some use cases

## License

MIT License - see the root repository LICENSE file for details.

