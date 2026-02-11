# Smidr Agent

The Smidr agent is a lightweight Go service that runs on Linux x86_64 systems, collecting system signals and communicating with the control plane over mTLS.

## Architecture

### Components

```
agent/
├── cmd/agent/          # Main entry point
├── internal/
│   ├── agent/          # Core daemon logic, config, enrollment
│   ├── signals/        # System signal collectors
│   ├── heartbeat/      # mTLS heartbeat client
│   └── logging/        # Structured logging
├── systemd/            # systemd service unit
└── install.sh          # Installation script
```

### Key Features

- **Signal Collection**: Gathers uptime, load average, memory usage, disk usage, and process count
- **mTLS Authentication**: All communication uses mutual TLS with client certificates
- **Enrollment Flow**: Auto-generates UUID and keypair, submits CSR, installs signed certificate
- **Heartbeat Loop**: Sends periodic telemetry to control plane (default: 60s interval)
- **Systemd Integration**: Runs as a hardened systemd service with automatic restart
- **Configuration**: YAML-based config at `/etc/smidr/agent.yaml`

## Signal Collectors

All signals are collected from `/proc` and system calls, no external dependencies required.

### Collected Signals

| Signal | Source | Type | Description |
|--------|--------|------|-------------|
| `uptimeSeconds` | `/proc/uptime` | float64 | System uptime in seconds |
| `loadAverage1m` | `/proc/loadavg` | float64 | 1-minute load average |
| `memoryUsedPct` | `/proc/meminfo` | float64 | Memory usage percentage |
| `diskUsedPct` | `statfs(/)` | float64 | Root disk usage percentage |
| `processCount` | `/proc` | int | Total process count |

Implementation: `internal/signals/signals.go`

## API Contracts

The agent communicates with the control plane using JSON over HTTPS.

### Enrollment: POST /agents/register

**Request:**
```json
{
  "agentId": "550e8400-e29b-41d4-a716-446655440000",
  "hostname": "web-server-01",
  "token": "optional-enrollment-token",
  "csrPem": "-----BEGIN CERTIFICATE REQUEST-----\n..."
}
```

**Response:**
```json
{
  "certPem": "-----BEGIN CERTIFICATE-----\n..."
}
```

**Flow:**
1. Agent generates UUID and RSA keypair on first start
2. Agent creates CSR with agent ID as Common Name
3. Agent sends CSR to control plane for signing
4. Control plane validates request and signs certificate
5. Agent receives and persists signed certificate
6. Agent can now use mTLS for all future communication

### Heartbeat: POST /v0/agents/heartbeat

**Authentication:** mTLS (client certificate required)

**Request:**
```json
{
  "agentId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-02-10T15:04:05Z",
  "uptimeSeconds": 86400.5,
  "loadAverage1m": 1.23,
  "memoryUsedPct": 45.6,
  "diskUsedPct": 67.8,
  "processCount": 156
}
```

**Response:**
- `200 OK`: Heartbeat accepted
- `4xx/5xx`: Error (agent will retry on next interval)

**Flow:**
1. Agent collects all system signals
2. Agent sends heartbeat payload with mTLS
3. Control plane validates certificate and processes signals
4. Agent logs success/failure and waits for next interval

## Configuration

Config file: `/etc/smidr/agent.yaml`

```yaml
control_plane_url: "https://control-plane.example.com"
agent_id: ""                  # auto-generated on first start
hostname: ""                  # auto-detected from system
key_path: ""                  # auto-set to /etc/smidr/agent.key.pem
csr_path: ""                  # auto-set to /etc/smidr/agent.csr.pem
cert_path: ""                 # auto-set to /etc/smidr/agent.crt.pem
token: ""                     # optional enrollment token
heartbeat_interval: 60        # heartbeat interval in seconds
```

All empty fields are populated automatically on first run and persisted to the config file.

## Installation

### Prerequisites

- Linux x86_64 system
- systemd
- Network access to control plane (outbound HTTPS only)

### Quick Install

```bash
# Build the agent
go build -o smidr-agent ./cmd/agent

# Install and start (requires root)
sudo ./install.sh --control-plane-url https://your-control-plane.example.com
```

### Manual Installation

```bash
# 1. Create smidr user
sudo useradd --system --no-create-home --shell /usr/sbin/nologin smidr

# 2. Install binary
sudo cp smidr-agent /usr/local/bin/
sudo chmod +x /usr/local/bin/smidr-agent

# 3. Initialize config
sudo mkdir -p /etc/smidr
sudo /usr/local/bin/smidr-agent init --config /etc/smidr/agent.yaml

# 4. Edit config
sudo vim /etc/smidr/agent.yaml
# Set control_plane_url to your control plane

# 5. Set permissions
sudo chown -R smidr:smidr /etc/smidr
sudo chmod 700 /etc/smidr
sudo chmod 600 /etc/smidr/agent.yaml

# 6. Install systemd service
sudo cp systemd/smidr-agent.service /etc/systemd/system/
sudo systemctl daemon-reload

# 7. Start service
sudo systemctl enable smidr-agent
sudo systemctl start smidr-agent
```

## Usage

### Commands

```bash
# Initialize config file
smidr-agent init [--config PATH] [--force]

# Run daemon (normally run via systemd)
smidr-agent daemon [--config PATH]

# Reset enrollment state (deletes cert/key/csr and clears agent ID)
smidr-agent reset-enrollment [--config PATH]

# Help
smidr-agent help
```

### Service Management

```bash
# Start service
sudo systemctl start smidr-agent

# Stop service
sudo systemctl stop smidr-agent

# Restart service
sudo systemctl restart smidr-agent

# Check status
sudo systemctl status smidr-agent

# View logs
sudo journalctl -u smidr-agent -f
```

## Development

### Building

The agent includes a comprehensive Makefile for building, testing, and development workflows:

```bash
# Show all available targets
make help

# Build for current platform
make build

# Build for Linux (production target)
make build-linux

# Build for all platforms
make build-all

# Run tests with race detector
make test

# Run tests with coverage report
make test-coverage

# Format code and run checks
make fmt vet

# Full development workflow (clean, deps, fmt, vet, test, build)
make dev
```

Or use Go commands directly:

```bash
go build -o smidr-agent ./cmd/agent
```

### Testing Signal Collection (requires Linux)

```bash
# Test on a Linux system
go run ./cmd/agent daemon --config /tmp/test-config.yaml

# Or run a specific test
go test ./internal/signals/...

# Or use Makefile
make test
```

## Security

- Agent runs as unprivileged `smidr` user
- Read-only access to `/proc` for signal collection
- Write access restricted to `/etc/smidr` only
- Private keys never leave the agent system (2048-bit RSA)
- All communication over mTLS (TLS 1.2+)
- systemd security hardening enabled (NoNewPrivileges, PrivateTmp, ProtectSystem)

## Troubleshooting

### Agent won't start

```bash
# Check service status
sudo systemctl status smidr-agent

# Check logs
sudo journalctl -u smidr-agent -n 50

# Common issues:
# - control_plane_url not set in config
# - Network connectivity to control plane
# - Permission issues on /etc/smidr
```

### Enrollment fails

```bash
# Check if cert already exists
ls -la /etc/smidr/*.pem

# Re-enroll using reset command (recommended)
sudo systemctl stop smidr-agent
sudo smidr-agent reset-enrollment
sudo systemctl start smidr-agent
```

### Heartbeats failing

```bash
# Check if enrolled (cert exists)
sudo ls -la /etc/smidr/agent.crt.pem

# Check mTLS cert validity
openssl x509 -in /etc/smidr/agent.crt.pem -noout -text

# If cert is for wrong agent ID (after control plane database reset):
# This happens when the control plane database was recreated but agent still has old cert
sudo systemctl stop smidr-agent
sudo smidr-agent reset-enrollment
sudo systemctl start smidr-agent

# Test connectivity
curl -v https://your-control-plane.example.com/v0/agents/heartbeat
```

## Integration Points for Dallas (Control Plane)

### What Dallas Needs to Implement

1. **POST /agents/register**
   - Parse CSR from agent
   - Validate enrollment token (if required)
   - Sign CSR with internal CA (coordinate with Ash)
   - Return signed certificate
   - Store agent registration in database

2. **POST /v0/agents/heartbeat**
   - Require mTLS authentication
   - Extract agent ID from client certificate
   - Parse heartbeat payload
   - Store signals in time-series database
   - Return 200 OK on success

### Expected Error Handling

- `400 Bad Request`: Invalid JSON or missing fields
- `401 Unauthorized`: Invalid or missing client certificate
- `403 Forbidden`: Invalid enrollment token
- `500 Internal Server Error`: Server-side errors

Agent will retry on next interval for all errors.

## Next Steps

- [ ] Add unit tests for signal collectors
- [ ] Add integration tests with mock control plane
- [ ] Build release pipeline with GoReleaser
- [ ] Add support for CA certificate verification (CACertPath)
- [ ] Add metrics for agent health (heartbeat success rate, etc.)
- [ ] Support graceful certificate rotation
