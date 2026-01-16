# Smidr C# SDK

C# client library for interacting with the Smidr gRPC daemon.

## Installation

Add the SmidrLib project reference to your application:

```bash
dotnet add reference path/to/SmidrLib/SmidrLib.csproj
```

Or reference the NuGet package once published:

```bash
dotnet add package SmidrLib
```

## Quick Start

```csharp
using SmidrLib;
using Smidr.V1;

// Create a client connected to the daemon
using var client = new SmidrClient("http://localhost:50051");

// Start a new build
var status = await client.StartBuildAsync(
    configPath: "smidr.yaml",
    target: "core-image-minimal",
    customer: "acme"
);

Console.WriteLine($"Build started: {status.BuildIdentifier.BuildId}");

// Monitor build status
var currentStatus = await client.GetBuildStatusAsync(status.BuildIdentifier.BuildId);
Console.WriteLine($"State: {currentStatus.State}");

// Stream logs in real-time
await foreach (var logEntry in client.StreamLogsAsync(
    status.BuildIdentifier.BuildId,
    follow: true))
{
    Console.WriteLine($"[{logEntry.Stream}] {logEntry.Message}");
}

// List artifacts when complete
var artifacts = await client.ListArtifactsAsync(status.BuildIdentifier.BuildId);
foreach (var artifact in artifacts.Artifacts)
{
    Console.WriteLine($"Artifact: {artifact.Name} ({artifact.SizeBytes} bytes)");
    Console.WriteLine($"  Download: {artifact.DownloadUrl}");
}
```

## API Overview

### SmidrClient

The main client class providing high-level methods for interacting with the daemon.

#### Constructor

```csharp
var client = new SmidrClient("http://localhost:50051");
```

#### Methods

##### StartBuildAsync

Start a new build with the specified configuration.

```csharp
var response = await client.StartBuildAsync(
    configPath: "smidr.yaml",
    target: "core-image-minimal",
    customer: "acme",           // optional
    forceClean: false,          // optional
    forceImageRebuild: false    // optional
);
```

##### GetBuildStatusAsync

Get the current status of a build.

```csharp
var status = await client.GetBuildStatusAsync("build-123");
Console.WriteLine($"State: {status.State}");
Console.WriteLine($"Started: {DateTimeOffset.FromUnixTimeSeconds(status.Timestamps.StartTimeUnixSeconds)}");
```

##### ListBuildsAsync

List all builds, optionally filtered by state.

```csharp
// List all builds
var allBuilds = await client.ListBuildsAsync();

// List only running builds
var runningBuilds = await client.ListBuildsAsync(
    states: new[] { BuildState.Building }
);

// List with page size limit
var recentBuilds = await client.ListBuildsAsync(pageSize: 10);
```

##### CancelBuildAsync

Cancel a running build.

```csharp
var response = await client.CancelBuildAsync("build-123");
Console.WriteLine($"Cancelled: {response.Success}");
```

##### ListArtifactsAsync

List artifacts from a completed build.

```csharp
var artifacts = await client.ListArtifactsAsync("build-123");
foreach (var artifact in artifacts.Artifacts)
{
    Console.WriteLine($"{artifact.Name}: {artifact.DownloadUrl}");
}
```

##### StreamLogsAsync

Stream logs from a build in real-time.

```csharp
await foreach (var logEntry in client.StreamLogsAsync(
    buildId: "build-123",
    follow: true))
{
    if (logEntry.Stream == "stderr")
    {
        Console.Error.WriteLine(logEntry.Message);
    }
    else
    {
        Console.WriteLine(logEntry.Message);
    }
}
```

### Direct Service Access

For advanced scenarios, you can access the underlying gRPC service clients directly:

```csharp
// Direct access to BuildService
var buildDetails = await client.Builds.GetBuildAsync(
    new GetBuildRequest { BuildIdentifier = new BuildIdentifier { BuildId = "build-123" } }
);

// Direct access to LogService
var logsCall = client.Logs.StreamBuildLogs(
    new StreamBuildLogsRequest {
        BuildIdentifier = new BuildIdentifier { BuildId = "build-123" },
        Follow = true
    }
);

// Direct access to ArtifactService
var artifactsResponse = await client.Artifacts.ListArtifactsAsync(
    new ListArtifactsRequest {
        BuildIdentifier = new BuildIdentifier { BuildId = "build-123" }
    }
);
```

## Build States

The `BuildState` enum represents the possible states of a build:

- `BuildStateUnspecified` - Default/unknown state
- `BuildStateQueued` - Build is queued and waiting to start
- `BuildStatePreparing` - Build environment is being prepared
- `BuildStateBuilding` - Build is actively running
- `BuildStateExtractingArtifacts` - Build completed, extracting artifacts
- `BuildStateCompleted` - Build completed successfully
- `BuildStateFailed` - Build failed
- `BuildStateCancelled` - Build was cancelled

## Timestamps

Build timestamps are provided as Unix seconds (int64) in the `TimeStampRange` message:

```csharp
var startTime = DateTimeOffset.FromUnixTimeSeconds(status.Timestamps.StartTimeUnixSeconds);
var endTime = DateTimeOffset.FromUnixTimeSeconds(status.Timestamps.EndTimeUnixSeconds);
var duration = endTime - startTime;
Console.WriteLine($"Build took {duration.TotalMinutes:F2} minutes");
```

## Error Handling

All async methods can throw `RpcException` for gRPC errors:

```csharp
using Grpc.Core;

try
{
    var status = await client.GetBuildStatusAsync("invalid-build-id");
}
catch (RpcException ex) when (ex.StatusCode == StatusCode.NotFound)
{
    Console.WriteLine("Build not found");
}
catch (RpcException ex)
{
    Console.WriteLine($"RPC error: {ex.Status.Detail}");
}
```

## Development

### Building the SDK

```bash
cd /path/to/smidr/sdks/csharp/SmidrLib
dotnet build
```

### Regenerating from Proto Files

When the proto files change, regenerate the C# code:

```bash
cd /path/to/smidr/protos
buf generate
```

The generated files are placed in `sdks/csharp/Generated/` and automatically included in the SmidrLib project.

### Running Tests

```bash
cd /path/to/smidr/sdks/csharp/SmidrLib
dotnet test
```

## Example: Full Build Workflow

```csharp
using SmidrLib;
using Smidr.V1;

using var client = new SmidrClient("http://localhost:50051");

// Start the build
var startResponse = await client.StartBuildAsync(
    configPath: "smidr.yaml",
    target: "core-image-minimal",
    customer: "acme"
);

var buildId = startResponse.BuildIdentifier.BuildId;
Console.WriteLine($"✅ Build started: {buildId}");

// Stream logs with a cancellation token for Ctrl+C handling
using var cts = new CancellationTokenSource();
Console.CancelKeyPress += (sender, e) => {
    e.Cancel = true;
    cts.Cancel();
};

try
{
    await foreach (var log in client.StreamLogsAsync(buildId, follow: true, cts.Token))
    {
        Console.WriteLine(log.Message);
    }
}
catch (OperationCanceledException)
{
    Console.WriteLine("\n⚠️  Log streaming cancelled");
}

// Check final status
var finalStatus = await client.GetBuildStatusAsync(buildId);
if (finalStatus.State == BuildState.BuildStateCompleted)
{
    Console.WriteLine("✅ Build completed successfully!");

    // List and download artifacts
    var artifacts = await client.ListArtifactsAsync(buildId);
    foreach (var artifact in artifacts.Artifacts)
    {
        Console.WriteLine($"📦 {artifact.Name}");
        Console.WriteLine($"   Size: {artifact.SizeBytes / 1024 / 1024:F2} MB");
        Console.WriteLine($"   URL: {artifact.DownloadUrl}");
        Console.WriteLine($"   Checksum: {artifact.Checksum}");
    }
}
else if (finalStatus.State == BuildState.BuildStateFailed)
{
    Console.WriteLine($"❌ Build failed: {finalStatus.ErrorMessage}");
    Console.WriteLine($"   Exit code: {finalStatus.ExitCode}");
}
```

## License

MIT License - see the root repository LICENSE file for details.
