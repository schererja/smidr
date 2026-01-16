# Smidr TypeScript/JavaScript SDK

TypeScript/JavaScript client library for interacting with the Smidr gRPC daemon using Connect-RPC.

## Installation

```bash
npm install @smidr/sdk @bufbuild/protobuf @connectrpc/connect @connectrpc/connect-node
```

Or with yarn:

```bash
yarn add @smidr/sdk @bufbuild/protobuf @connectrpc/connect @connectrpc/connect-node
```

Or with pnpm:

```bash
pnpm add @smidr/sdk @bufbuild/protobuf @connectrpc/connect @connectrpc/connect-node
```

## Quick Start

```typescript
import { SmidrClient, BuildState } from "@smidr/sdk";

// Create a client connected to the daemon
const client = new SmidrClient({
  address: "http://localhost:50051",
});

// Start a new build
const status = await client.startBuild({
  config: "smidr.yaml",
  target: "core-image-minimal",
  customer: "acme",
});

console.log(`Build started: ${status.buildIdentifier?.buildId}`);

// Monitor build status
const currentStatus = await client.getBuildStatus(status.buildIdentifier!.buildId);
console.log(`State: ${BuildState[currentStatus.state]}`);

// Stream logs in real-time
for await (const logEntry of client.streamLogs(status.buildIdentifier!.buildId, true)) {
  console.log(`[${logEntry.stream}] ${logEntry.message}`);
}

// List artifacts when complete
const artifacts = await client.listArtifacts(status.buildIdentifier!.buildId);
for (const artifact of artifacts.artifacts) {
  console.log(`Artifact: ${artifact.name} (${artifact.sizeBytes} bytes)`);
  console.log(`  Download: ${artifact.downloadUrl}`);
}
```

## API Overview

### SmidrClient

The main client class providing high-level methods for interacting with the daemon.

#### Constructor

```typescript
const client = new SmidrClient({
  address: "http://localhost:50051",
  httpVersion: "2", // optional, defaults to "2"
});
```

#### Methods

##### startBuild

Start a new build with the specified configuration.

```typescript
const response = await client.startBuild({
  config: "smidr.yaml",
  target: "core-image-minimal",
  customer: "acme", // optional
  forceClean: false, // optional
  forceImageRebuild: false, // optional
  environmentVariables: {
    // optional
    MY_VAR: "value",
  },
});
```

##### getBuildStatus

Get the current status of a build.

```typescript
const status = await client.getBuildStatus("build-123");
console.log(`State: ${BuildState[status.state]}`);
console.log(
  `Started: ${new Date(Number(status.timestamps?.startTimeUnixSeconds) * 1000).toISOString()}`
);
```

##### listBuilds

List all builds, optionally filtered by state.

```typescript
// List all builds
const allBuilds = await client.listBuilds();

// List only running builds
const runningBuilds = await client.listBuilds({
  stateFilter: [BuildState.BUILD_STATE_BUILDING],
});

// List with page size limit
const recentBuilds = await client.listBuilds({
  pageSize: 10,
});
```

##### cancelBuild

Cancel a running build.

```typescript
const response = await client.cancelBuild("build-123");
console.log(`Cancelled: ${response.success}`);
```

##### listArtifacts

List artifacts from a completed build.

```typescript
const artifacts = await client.listArtifacts("build-123");
for (const artifact of artifacts.artifacts) {
  console.log(`${artifact.name}: ${artifact.downloadUrl}`);
}
```

##### streamLogs

Stream logs from a build in real-time using async iteration.

```typescript
for await (const logEntry of client.streamLogs("build-123", true)) {
  if (logEntry.stream === "stderr") {
    console.error(logEntry.message);
  } else {
    console.log(logEntry.message);
  }
}
```

### Direct Service Access

For advanced scenarios, you can access the underlying Connect service clients directly:

```typescript
// Direct access to BuildService
const buildDetails = await client.builds.getBuild({
  buildIdentifier: create(BuildIdentifierSchema, { buildId: "build-123" }),
});

// Direct access to LogService
const logsStream = client.logs.streamBuildLogs({
  buildIdentifier: create(BuildIdentifierSchema, { buildId: "build-123" }),
  follow: true,
});

// Direct access to ArtifactService
const artifactsResponse = await client.artifacts.listArtifacts({
  buildIdentifier: create(BuildIdentifierSchema, { buildId: "build-123" }),
});
```

## Build States

The `BuildState` enum represents the possible states of a build:

```typescript
enum BuildState {
  BUILD_STATE_UNSPECIFIED = 0,
  BUILD_STATE_QUEUED = 1,
  BUILD_STATE_PREPARING = 2,
  BUILD_STATE_BUILDING = 3,
  BUILD_STATE_EXTRACTING_ARTIFACTS = 4,
  BUILD_STATE_COMPLETED = 5,
  BUILD_STATE_FAILED = 6,
  BUILD_STATE_CANCELLED = 7,
}
```

## Timestamps

Build timestamps are provided as Unix seconds (bigint) in the `TimeStampRange` message:

```typescript
const startTime = new Date(Number(status.timestamps?.startTimeUnixSeconds) * 1000);
const endTime = new Date(Number(status.timestamps?.endTimeUnixSeconds) * 1000);
const duration = endTime.getTime() - startTime.getTime();
console.log(`Build took ${duration / 1000 / 60} minutes`);
```

## Error Handling

Connect-RPC errors are thrown as `ConnectError`:

```typescript
import { ConnectError, Code } from "@connectrpc/connect";

try {
  const status = await client.getBuildStatus("invalid-build-id");
} catch (err) {
  if (err instanceof ConnectError) {
    if (err.code === Code.NotFound) {
      console.log("Build not found");
    } else {
      console.log(`RPC error: ${err.message}`);
    }
  }
}
```

## Usage in Next.js / React

For web applications, you can use the browser-compatible transport:

```bash
npm install @connectrpc/connect-web
```

```typescript
import { SmidrClient } from "@smidr/sdk";
import { createConnectTransport } from "@connectrpc/connect-web";

// Use the Connect transport for browsers
const transport = createConnectTransport({
  baseUrl: "http://localhost:50051",
});

// Pass the transport to the client constructor
// (Note: you'll need to modify SmidrClient to accept a custom transport)
```

For Next.js API routes or server components, continue using the Node.js transport as shown above.

## Example: Full Build Workflow

```typescript
import { SmidrClient, BuildState } from "@smidr/sdk";

const client = new SmidrClient({
  address: "http://localhost:50051",
});

async function runBuild() {
  // Start the build
  const startResponse = await client.startBuild({
    config: "smidr.yaml",
    target: "core-image-minimal",
    customer: "acme",
  });

  const buildId = startResponse.buildIdentifier!.buildId;
  console.log(`✅ Build started: ${buildId}`);

  // Stream logs with cancellation support
  const abortController = new AbortController();
  process.on("SIGINT", () => {
    abortController.abort();
  });

  try {
    for await (const log of client.streamLogs(buildId, true)) {
      console.log(log.message);
    }
  } catch (err) {
    if (err.name === "AbortError") {
      console.log("\n⚠️  Log streaming cancelled");
    } else {
      throw err;
    }
  }

  // Check final status
  const finalStatus = await client.getBuildStatus(buildId);
  if (finalStatus.state === BuildState.BUILD_STATE_COMPLETED) {
    console.log("✅ Build completed successfully!");

    // List and download artifacts
    const artifacts = await client.listArtifacts(buildId);
    for (const artifact of artifacts.artifacts) {
      console.log(`📦 ${artifact.name}`);
      console.log(`   Size: ${(Number(artifact.sizeBytes) / 1024 / 1024).toFixed(2)} MB`);
      console.log(`   URL: ${artifact.downloadUrl}`);
      console.log(`   Checksum: ${artifact.checksum}`);
    }
  } else if (finalStatus.state === BuildState.BUILD_STATE_FAILED) {
    console.log(`❌ Build failed: ${finalStatus.errorMessage}`);
    console.log(`   Exit code: ${finalStatus.exitCode}`);
  }
}

runBuild().catch(console.error);
```

## Development

### Building the SDK

```bash
npm install
npm run build
```

### Regenerating from Proto Files

When the proto files change, regenerate the TypeScript code:

```bash
cd /path/to/smidr/protos
buf generate
```

The generated files are placed in `sdks/ts/Generated/` and automatically included in the SDK.

## Generated Files

The SDK includes these generated files in `Generated/`:

- **Message types** (`*_pb.ts`): Protocol buffer message definitions
  - `common_pb.ts` - Common types (BuildIdentifier, BuildState, etc.)
  - `builds_pb.ts` - Build-related messages
  - `logs_pb.ts` - Log-related messages
  - `artifacts_pb.ts` - Artifact-related messages
  - `smidr_service_pb.ts` - Service definitions

- **Service clients** (`*_connect.ts`): Connect-RPC service definitions
  - `builds_connect.ts` - BuildService client
  - `logs_connect.ts` - LogService client
  - `artifacts_connect.ts` - ArtifactService client

## License

MIT License - see the root repository LICENSE file for details.
