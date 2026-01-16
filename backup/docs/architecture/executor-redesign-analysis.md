# Executor Redesign Analysis: gRPC vs HTTP & Model Organization

## Executive Summary

**Recommendation: Migrate to gRPC** for executor-to-control-plane communication with a phased approach. Your protos are well-structured but need better organization in the executor codebase.

## Current State Assessment

### What You Have Now

- **Protocol**: HTTP/REST via `pkg/protocols/httpExecutorClient`
- **Models**: Scattered across `pkg/jobs/` with Go-native types
- **Proto definitions**: Well-organized control plane and agent service definitions
- **Integration**: Loose coupling between proto types and executor implementation

### Pain Points

1. **Dual type systems**: Proto types AND Go types for jobs/steps
2. **Manual serialization**: JSON marshaling in HTTP client
3. **No streaming**: Logs are pushed via individual HTTP POSTs
4. **Type drift**: Easy for proto and Go models to diverge
5. **Limited error handling**: HTTP status codes less expressive than gRPC status codes

---

## gRPC vs HTTP Analysis

### Why gRPC is Better for Smidr

#### 1. **Streaming (Critical for Build Systems)**

```proto
// Native bi-directional streaming for logs
rpc StreamJobLogs(StreamJobLogsRequest) returns (stream JobLogEntry);

// Agent can push logs as they happen without HTTP overhead
```

**Impact for Yocto/large builds:**

- Current: Each log line = 1 HTTP POST (network overhead, connection setup)
- gRPC: Single connection, push logs as they come, lower latency
- Better for long-running builds (Yocto can take hours)

#### 2. **Type Safety & Code Generation**

```go
// Current (manual):
type Step struct {
    Name    string
    Command []string
    // ... manual JSON tags, validation
}

// gRPC (generated):
import pb "github.com/schererja/smidr/sdks/pkg/smidr-sdk/v1"
step := &pb.Step{
    Name: "build",
    // Type-safe, compiler-checked
}
```

**Benefits:**

- No type drift between control plane and executor
- Breaking changes caught at compile time
- Less boilerplate code

#### 3. **Performance**

- HTTP/2 multiplexing (built-in)
- Binary protocol vs JSON (smaller payloads for large job definitions)
- Connection pooling handled by gRPC library
- Keepalive/heartbeat built into gRPC

**For your use case (build automation):**

- Job definitions with large artifact lists compress better
- Yocto metadata (layers, recipes) = lots of strings
- Binary encoding = ~30-50% smaller than JSON

#### 4. **Error Handling**

```go
// Current HTTP:
if resp.StatusCode != 200 {
    return fmt.Errorf("status: %d", resp.StatusCode)
}

// gRPC:
_, err := client.FetchJob(ctx, req)
if st, ok := status.FromError(err); ok {
    switch st.Code() {
    case codes.ResourceExhausted:
        // Queue full, backoff longer
    case codes.DeadlineExceeded:
        // Network issue, retry
    case codes.FailedPrecondition:
        // Agent not registered
    }
}
```

Rich error codes align with job lifecycle states.

#### 5. **Observability**

gRPC middleware for:

- OpenTelemetry tracing (see where time is spent in job dispatch)
- Prometheus metrics (call duration, error rates per RPC)
- Structured logging per RPC

**Critical for CI debugging:** Track why a Yocto build stalled at the control plane vs executor level.

### When HTTP Makes Sense

- Public webhooks (GitHub/GitLab trigger builds) → Keep HTTP
- Health checks / simple status → HTTP is fine
- Browser-based UI → GraphQL over HTTP or REST

**Recommendation:** Keep HTTP for external integrations, use gRPC for executor ↔ control plane.

---

## Model Organization Redesign

### Problem: Current Structure

```
executor/
├── pkg/
│   ├── jobs/          # Go-native types (Job, Step)
│   ├── executor/      # Uses jobs.Job
│   └── protocols/
│       └── httpExecutorClient/  # Manually maps Go → JSON
```

**Issues:**

- `jobs.Job` != proto `Job` (easy to drift)
- Manual mapping code in HTTP client
- No single source of truth

### Proposed Structure (gRPC-based)

```
executor/
├── pkg/
│   ├── agent/         # Agent implementation (uses proto types)
│   │   ├── client.go          # gRPC client wrapper
│   │   ├── registration.go    # RegisterAgent, Heartbeat
│   │   └── job_polling.go     # FetchJob, ReportResult
│   │
│   ├── runner/        # Execution layer (adapts proto → execution)
│   │   ├── smart_runner.go    # Step executor
│   │   ├── local_runner.go
│   │   └── docker_runner.go
│   │
│   ├── materialize/   # Build system detection (unchanged)
│   │
│   └── types/         # Executor-specific extensions
│       └── execution_context.go  # Runtime state not in proto
│
├── internal/
│   └── config/        # Executor config (smidr.yaml)
│
└── api/               # Generated proto Go code (symlink)
    └── v1/
        ├── agent.pb.go
        └── agent_grpc.pb.go
```

### Key Changes

#### 1. Use Proto Types Directly

```go
// executor/pkg/agent/client.go
import pb "github.com/schererja/smidr/sdks/pkg/smidr-sdk/v1"

type Client struct {
    conn   *grpc.ClientConn
    agent  pb.AgentServiceClient
}

func (c *Client) FetchJob(ctx context.Context) (*pb.Job, error) {
    resp, err := c.agent.FetchJob(ctx, &pb.FetchJobRequest{
        AgentId: c.agentID,
        Capabilities: c.capabilities,
    })
    return resp.GetJob(), err
}
```

**No more manual mapping!** Proto defines the contract.

#### 2. Adapter Layer for Execution

```go
// executor/pkg/runner/adapter.go
package runner

import pb "github.com/schererja/smidr/sdks/pkg/smidr-sdk/v1"

// ExecutionContext augments proto Step with runtime state
type ExecutionContext struct {
    ProtoStep *pb.Step
    WorkDir   string
    LogWriter io.Writer
    // Executor-specific fields not in proto
}

func (sr *SmartRunner) Run(ctx context.Context, job *pb.Job) (*pb.JobResult, error) {
    for _, step := range job.GetSteps() {
        execCtx := &ExecutionContext{
            ProtoStep: step,
            WorkDir:   sr.workDir,
        }
        if err := sr.executeStep(ctx, execCtx); err != nil {
            return &pb.JobResult{Success: false, Error: err.Error()}, err
        }
    }
    return &pb.JobResult{Success: true}, nil
}
```

**Why?**

- Proto defines what to execute
- ExecutionContext adds _how_ to execute (local state)
- Clean separation of concerns

#### 3. Streaming Logs

```go
// executor/pkg/agent/log_stream.go
func (c *Client) StreamLogs(ctx context.Context, jobID *pb.JobID) error {
    stream, err := c.agent.StreamJobLogs(ctx)
    if err != nil {
        return err
    }

    for logEntry := range c.logChan {
        stream.Send(&pb.JobLogEntry{
            JobId:     jobID,
            StepName:  logEntry.StepName,
            Line:      logEntry.Line,
            Timestamp: timestamppb.Now(),
        })
    }
    return stream.CloseSend()
}
```

Replace current `LogCollector` HTTP POST loop with streaming.

---

## Proto Model Improvements

### Current Proto Structure (Good!)

```
proto/smidr/v1/
├── common/
│   ├── types.proto        # JobID, AgentID, enums
│   ├── policy.proto
│   └── execution.proto
├── agent/
│   └── agent.proto        # AgentService RPCs
└── control_plane/
    ├── job.proto          # JobService RPCs
    └── artifact.proto
```

**This is well-organized!** Keep this structure.

### Recommended Additions

#### 1. Add execution-specific types

```proto
// proto/smidr/v1/agent/execution.proto
syntax = "proto3";
package smidr.agent;

import "smidr/v1/common/types.proto";

// Step execution result (richer than JobResult)
message StepResult {
  string step_name = 1;
  bool success = 2;
  int32 exit_code = 3;
  string error_message = 4;
  google.protobuf.Duration duration = 5;
  map<string, string> metrics = 6;  // e.g., "cache_hits": "42"
}

// Build system detection result (from materialize package)
message BuildSystemInfo {
  enum BuildSystem {
    BUILD_SYSTEM_UNSPECIFIED = 0;
    BUILD_SYSTEM_YOCTO = 1;
    BUILD_SYSTEM_LIVE_BUILD = 2;
    BUILD_SYSTEM_TINYCORE = 3;
    BUILD_SYSTEM_MAKE = 4;
  }
  BuildSystem system = 1;
  repeated string detected_files = 2;  // e.g., "build.yaml"
  map<string, string> config = 3;
}
```

**Why?** Formalize build system detection in proto so control plane can query "what build systems does agent X support?"

#### 2. Richer Step types (align with current executor)

```proto
// Extend smidr/v1/common/execution.proto
enum StepType {
  STEP_TYPE_UNSPECIFIED = 0;
  STEP_TYPE_COMMAND = 1;
  STEP_TYPE_SCRIPT = 2;
  STEP_TYPE_BINARY = 3;
  STEP_TYPE_BUILD_CONFIG = 4;  // Yocto/live-build/tinycore
  STEP_TYPE_DOCKER_BUILD = 5;
  STEP_TYPE_CONTAINER_RUN = 6;
}

enum RuntimeType {
  RUNTIME_TYPE_UNSPECIFIED = 0;
  RUNTIME_TYPE_LOCAL = 1;
  RUNTIME_TYPE_DOCKER = 2;
  RUNTIME_TYPE_PODMAN = 3;
  RUNTIME_TYPE_FIRECRACKER = 4;  // Future: micro-VMs
}

message Step {
  string name = 1;
  StepType type = 2;
  RuntimeType runtime = 3;

  oneof execution {
    CommandExecution command = 4;
    ScriptExecution script = 5;
    BuildConfigExecution build_config = 6;
  }

  map<string, string> env = 7;
  string workdir = 8;
  int32 timeout_seconds = 9;
  bool continue_on_error = 10;
  repeated StepInput inputs = 11;
}

message CommandExecution {
  repeated string argv = 1;
}

message ScriptExecution {
  string shell = 1;  // "bash", "sh", "python"
  string script = 2;
}

message BuildConfigExecution {
  string config_file = 1;  // "build.yaml"
  BuildSystemInfo.BuildSystem system = 2;
  map<string, string> parameters = 3;
}
```

**Benefits:**

- Type-safe execution modes
- Control plane knows what each step does
- Easier to add new step types (Firecracker, Kubernetes)

---

## Migration Path

### Phase 1: Proto Consolidation (Week 1)

1. ✅ Fix proto compilation errors (DONE!)
2. Add execution.proto with richer Step types
3. Generate Go code: `buf generate`
4. Update executor to import generated types

### Phase 2: gRPC Client (Week 2)

1. Create `executor/pkg/agent/client.go` with gRPC
2. Keep HTTP client as fallback
3. Add feature flag: `SMIDR_PROTOCOL=grpc|http`
4. Test with local control plane

### Phase 3: Streaming (Week 3)

1. Replace `LogCollector` HTTP POST with gRPC streaming
2. Benchmark: HTTP vs gRPC for long-running Yocto build
3. Document latency improvements

### Phase 4: Remove HTTP (Week 4)

1. Delete `pkg/protocols/httpExecutorClient`
2. Remove `pkg/jobs` (use proto types)
3. Update docs

### Phase 5: Polish (Ongoing)

1. Add OpenTelemetry tracing
2. gRPC health checks
3. Connection pooling tuning

---

## Decision Matrix

| Factor                     | HTTP                      | gRPC                        | Winner   |
| -------------------------- | ------------------------- | --------------------------- | -------- |
| **Streaming logs**         | Requires polling/webhooks | Native bi-directional       | **gRPC** |
| **Type safety**            | Manual JSON mapping       | Code generation             | **gRPC** |
| **Performance**            | JSON overhead             | Binary protocol             | **gRPC** |
| **Browser compatibility**  | Native                    | Requires proxy (grpc-web)   | HTTP     |
| **Debugging**              | curl, Postman             | grpcurl, more setup         | HTTP     |
| **Long-lived connections** | Tricky (HTTP/1.1)         | Built-in (HTTP/2)           | **gRPC** |
| **Error handling**         | Status codes              | Rich status codes + details | **gRPC** |
| **Learning curve**         | Lower                     | Steeper                     | HTTP     |

**For CI/build automation:** gRPC wins on technical merits. HTTP better for webhooks/public APIs.

---

## Recommended Architecture

```
┌─────────────────┐         gRPC          ┌──────────────────┐
│  Control Plane  │◄─────────────────────►│  Executor Agent  │
│   (Go/C#)       │   (AgentService)       │     (smidrd)     │
└─────────────────┘                        └──────────────────┘
        │                                            │
        │ HTTP/GraphQL                               │ Execution
        ▼                                            ▼
┌─────────────────┐                        ┌──────────────────┐
│   Web UI        │                        │  Local/Docker    │
│   (Next.js)     │                        │     Runners      │
└─────────────────┘                        └──────────────────┘
```

**Protocol usage:**

- Executor ↔ Control Plane: **gRPC** (performance, streaming)
- Web UI ↔ Control Plane: **HTTP/GraphQL** (browser compatibility)
- Webhooks → Control Plane: **HTTP** (external integrations)

---

## Next Steps

### Immediate (Fix protos)

- ✅ Resolved `smidr.common.ID` errors
- ✅ Proto generation working

### Short-term (This week)

1. Design Step proto with execution modes
2. Add BuildSystemInfo to proto
3. Create `executor/pkg/agent` with gRPC client stub

### Medium-term (Next 2 weeks)

1. Implement gRPC client parallel to HTTP
2. Add streaming log support
3. Benchmark both approaches with Yocto test build

### Long-term (Next month)

1. Migrate control plane to gRPC
2. Remove HTTP client from executor
3. Add OpenTelemetry tracing

---

## Questions to Resolve

1. **Control plane language?** (Go, C#, or both?)
   - Affects proto package structure
   - Both: use `github.com/schererja/smidr/sdks/{go,csharp}`

2. **Backward compatibility?**
   - Keep HTTP for 1-2 releases as deprecated?
   - Recommend: Yes, make migration gradual

3. **Authentication?**
   - Current: None (executor_id in requests)
   - gRPC: Use mTLS or JWT in metadata
   - Add `AuthenticationConfig` to `smidr.yaml`

4. **Load balancing?**
   - gRPC: Use client-side LB or Envoy/Linkerd
   - Important if you have multiple control plane instances

---

## Summary

**Do it.** Migrate to gRPC, but do it incrementally:

1. Fix protos (done ✅)
2. Add gRPC client alongside HTTP
3. Test with real Yocto builds
4. Measure latency/throughput improvements
5. Deprecate HTTP once confident

Your proto structure is solid. Main work is:

- Use proto types directly in executor (kill `pkg/jobs` duplication)
- Add execution-specific protos (BuildSystemInfo, StepResult)
- Implement gRPC client with streaming

**Estimated effort:** 2-3 weeks for full migration with proper testing.

**Biggest win:** Streaming logs for long Yocto builds + no more type drift.
