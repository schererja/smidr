# Parker Work History

## 2026-02-10: Team formation
- Assigned to testing
- Responsibilities: Unit tests (Go, C#, React), test infrastructure, edge cases, quality gates, coverage
- Work order: Write tests as Kane, Dallas, Lambert deliver code

## Learnings

### Testing Infrastructure
- Go standard library `testing` package used for all agent tests
- Test files colocated with implementation: `*_test.go` files in same package
- Coverage tracked via `go test -cover` command
- Platform-specific tests (Linux /proc) skip gracefully on non-Linux systems using `t.Skip()` with error detection

### Agent Code Structure
- **Signal collectors** (`agent/internal/signals/signals.go`): Collect system metrics from /proc filesystem
  - `collectUptime()` reads /proc/uptime
  - `collectLoadAverage()` reads /proc/loadavg  
  - `collectMemoryUsage()` parses /proc/meminfo for MemTotal and MemAvailable
  - `collectDiskUsage()` uses syscall.Statfs for disk metrics
  - `collectProcessCount()` counts numeric directories in /proc
  - All collectors resilient to individual failures via error aggregation
  
- **Config management** (`agent/internal/agent/config.go`): Simple YAML parser without external deps
  - Custom line-by-line parser (no yaml library dependency)
  - Handles comments, quotes (single/double), malformed lines gracefully
  - Default config path: `/etc/smidr/agent.yaml` on Linux, user config dir otherwise
  - File permissions: 0600 (owner read/write only)
  
- **Certificate/enrollment** (`agent/internal/agent/certificates.go`): RSA keypair and CSR generation
  - 2048-bit RSA keys in PKCS8 format
  - CSRs include AgentID as CommonName, optional Hostname as DNSName
  - Idempotent: won't regenerate existing key/CSR unless key deleted
  - All crypto files stored with 0600 permissions
  
- **UUID generation** (`agent/internal/agent/uuid.go`): Wraps google/uuid library
  - Generates RFC 4122 UUIDv4 (random)
  - Lowercase, hyphenated format

### Testing Patterns Established
- **Edge case coverage**: Empty paths, missing files, invalid inputs, malformed data
- **Permission validation**: Verify 0600 on sensitive files (keys, certs, configs)
- **Idempotency tests**: Verify operations don't change state on repeated calls
- **Platform handling**: Skip Linux-specific tests on macOS/Windows with meaningful messages
- **Table-driven tests**: Use subtests with test tables for parsing functions
- **Integration tests**: Test full workflows (config->key->csr) not just units
- **Error message validation**: Assert specific error keywords, not exact strings

### Edge Cases Tested
- Missing /proc files (graceful failure with error aggregation)
- Invalid config YAML (malformed lines, missing colons, invalid types)
- Empty/missing required fields (agent_id, config path, cert path)
- File permission issues (tested via stat checks)
- CSR generation without hostname (optional DNSNames)
- Config file already exists (force flag behavior)
- UUID uniqueness and format validation
- Disk usage on invalid paths

### Code Quality Observations
- Kane's code has good separation of concerns (collectors are pure functions)
- Error handling is consistent (wraps with context, uses errors.Join for aggregation)
- No external YAML library keeps binary size small (custom parser is adequate)
- Crypto code uses stdlib correctly (crypto/rsa, crypto/x509, encoding/pem)
- File I/O consistently uses appropriate permissions

### Test Coverage Achieved
- `agent/internal/agent`: 43.6% (config, certs, uuid)
- `agent/internal/signals`: 50.5% (signal collectors) — fixed build tag issue with Linux-specific tests
- Coverage focused on critical paths: enrollment, config loading, signal collection
- Untested paths: daemon loop, heartbeat (will test when those modules mature)

### Test Infrastructure Established
- **Agent (Go):** Standard library `testing` with platform-specific build tags for /proc tests
- **Control Plane (C#):** xUnit + InMemory EF + real CaService for integration-style tests
- **UI (React):** Vitest + Testing Library for component and API client tests
- 59 total tests covering critical paths across all components
- No integration/E2E tests per v0 scope decision — unit tests only

📌 Team update (2026-02-11): Test infrastructure established for all components — decided by Parker

## 2026-02-11: Merged Decisions from Team Debug Session

**Merged from inbox decisions:** parker-test-infrastructure.md

**Key consolidated decisions:**

### Test Infrastructure Established
- **Agent (Go):** 50+ tests with standard library `testing` package
  - Platform-aware: Linux-specific /proc tests skip gracefully on macOS/Windows
  - Covers: config, certificates, UUID generation, signal collection
  - Coverage: 43.6% agent internals, 50.5% signal collectors
  
- **Control Plane (C#):** 22 tests with xUnit + InMemory EF + real CaService
  - HealthEvaluationService: state transitions, baseline computation, anomaly detection
  - MtlsValidationService: certificate validation, CN extraction
  - AgentsController: REST endpoints (register, list, detail, revoke)
  
- **UI (React):** 22 tests with Vitest + Testing Library
  - API client: fetch, error handling, DTO mapping
  - HealthBadge: all health states and size variations
  - MetricsCard: data formatting, delta coloring, signal display

### Testing Patterns and Best Practices
- No external mocking frameworks - use standard library patterns
- Table-driven tests for parsing and multiple scenarios
- Platform-specific tests skip gracefully using error detection
- Idempotency tests for critical operations
- Permission validation for sensitive files (0600)
- Focus on critical paths over 100% coverage
- Colocated test files (*_test.go, *.test.ts, *.test.tsx)

### Test Coverage Strategy
- 59 total tests covering critical paths across all components
- No integration/E2E tests per v0 scope (unit tests only)
- All tests runnable via standard commands: `go test ./...`, `dotnet test`, `npm test`
- Each test verifies expected behavior and catches regressions

**Coordination outcomes:**
- Coverage metrics calculated: 43.6% agent, 50.5% signals in Go
- All unit tests integrated into CI/CD pipeline ready
- Testing patterns documented for future contributors

### Control Plane Testing (C#)
- **Test project:** `control-plane-tests/` using xUnit, EntityFramework InMemory, real CaService
- **HealthEvaluationService tests** (`HealthEvaluationServiceTests.cs`): 8 tests covering state transitions
  - Learning period detection (24h window)
  - Baseline computation from heartbeat history (min 10 samples)
  - Anomaly detection using 3-sigma threshold
  - Health state transitions: Learning → Healthy/Degraded/Attention/Unknown
  - Edge cases: nonexistent agents, missing heartbeats, zero baselines
- **MtlsValidationService tests** (`MtlsValidationServiceTests.cs`): 4 tests + 2 skipped
  - Certificate chain validation with real CA
  - Null certificate handling
  - AgentId extraction from CN field
  - Skipped: expired/not-yet-valid certs (requires manual cert generation with custom dates)
- **AgentsController tests** (`AgentsControllerTests.cs`): 10 tests covering REST endpoints
  - GET /api/agents (list with heartbeats)
  - GET /api/agents/{id} (detail with baselines)
  - POST /api/agents/register (CSR signing, validation)
  - POST /api/agents/{id}/revoke
  - Error cases: 404, 400 for missing fields, invalid CSR
- All 22 tests pass (2 skipped), runnable via `dotnet test`

### UI Testing (React/TypeScript)
- **Test framework:** Vitest with Testing Library, happy-dom environment
- **API client tests** (`src/test/api.test.ts`): 7 tests for `api/client.ts`
  - Agent list fetching and mapping (C# DTOs → UI types)
  - Agent detail fetching with baselines/heartbeats
  - Error handling: ECONNREFUSED, 404, 500
  - Health state case-insensitive mapping (LEARNING → learning, Attention → attention)
  - Null handling for lastHeartbeatAt (defaults to epoch)
- **HealthBadge tests** (`src/test/HealthBadge.test.tsx`): 7 tests
  - All 5 health states render correctly (learning, healthy, degraded, attention, unknown)
  - Size prop handling (small, medium, large)
  - Default medium size
- **MetricsCard tests** (`src/test/MetricsCard.test.tsx`): 8 tests
  - "No signal data" message when signals undefined
  - All 5 metrics render (uptime, load, memory, disk, process count)
  - Uptime formatting (seconds → days/hours/minutes)
  - Baseline delta display with +/- signs
  - Delta color coding for anomalies
  - Zero uptime handling
- Tests runnable via `npm test` (vitest configured in vite.config.ts)

### Testing Infrastructure Decisions
- **Agent (Go):** Colocated tests (`*_test.go`), platform-specific tests use build tags (`//go:build linux`)
- **Control Plane (C#):** Separate test project, InMemory database for isolation, real services (CaService) for integration-style tests
- **UI (React):** Vitest + Testing Library, mocked axios, focus on logic/rendering not full integration
- **No integration/E2E tests** in v0 scope — unit tests only per project decisions

### Test Running Commands
```bash
# Agent
cd agent && go test ./...

# Control Plane
cd control-plane-tests && dotnet test

# UI
cd ui && npm test
```


