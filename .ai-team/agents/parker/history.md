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
- `agent/internal/signals`: 50.5% (signal collectors)
- Coverage focused on critical paths: enrollment, config loading, signal collection
- Untested paths: daemon loop, heartbeat (will test when those modules mature)

