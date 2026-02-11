### 2026-02-11: Agent sends OS information to control plane

**By:** Kane

**What:** Added OS field to agent registration and heartbeat payloads. Agent now reports operating system using `runtime.GOOS` ("linux", "darwin", "windows").

**Why:** UI needs OS information to display appropriate icons for each agent. Registration payload captures OS during enrollment, and heartbeat payload includes OS for consistency. This enables the frontend to show platform-specific icons without requiring a separate API call or database migration.

**Implementation:**
- Added `OS string json:"os"` field to `registerAgentRequest` struct in `agent/internal/agent/daemon.go`
- Added `OS string json:"os"` field to `heartbeatRequest` struct in `agent/internal/heartbeat/client.go`
- Both payloads populate OS field with `runtime.GOOS` value
- Imported `runtime` package in both files
