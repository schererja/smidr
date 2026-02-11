# Smidr Control Plane

ASP.NET Core REST API that manages agent enrollment, heartbeat ingestion, health evaluation, and certificate authority.

## Features

- **Agent Enrollment**: Validates CSRs and issues signed certificates via internal CA
- **Heartbeat Ingestion**: Receives mTLS-authenticated agent telemetry (uptime, load, memory, disk, process count)
- **Health Evaluation**: Learns baselines over 24h, detects anomalies, transitions health states (learning → healthy/degraded/attention)
- **Certificate Authority**: Self-signed CA for agent certificate issuance
- **Database**: Stores agents, heartbeats, baselines, health states (SQLite default, PostgreSQL supported)

## Run

```bash
cd control-plane
dotnet restore
dotnet run
```

Swagger UI: `https://localhost:5001/swagger`

## API Endpoints

### POST /agents/register

Enrolls a new agent by signing its CSR.

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

### POST /v0/agents/heartbeat

Receives agent heartbeat with system signals. In production, requires mTLS authentication.

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

**Response:** `200 OK`

Triggers async health evaluation after storing heartbeat.

### GET /agents

Lists all registered agents with current health status.

### POST /agents/{id}/revoke

Revokes an agent certificate.

## Health Evaluation

**Learning Phase (24h):**
- Collects baseline data
- Computes mean, std dev, min, max per metric
- Agent status: `Learning`

**Evaluation (post-learning):**
- Compares current values to baselines
- Detects anomalies (> 3 standard deviations)
- Updates health status:
  - `Healthy`: no anomalies
  - `Degraded`: 1 anomaly
  - `Attention`: 2+ anomalies
  - `Unknown`: no recent heartbeats

## Database

**Default:** SQLite at `control-plane/data/controlplane.db`

**PostgreSQL (production):**
```json
{
  "UsePostgres": true,
  "ConnectionStrings": {
    "PostgreSQL": "Host=localhost;Database=smidr;Username=smidr;Password=smidr"
  }
}
```

## Certificate Authority

CA keypair generated on first start:
- `data/ca/ca.key.pem` (private key)
- `data/ca/ca.crt.pem` (self-signed certificate)

CA signs agent CSRs with 365-day validity.

## Development

### Build
```bash
dotnet build
```

### Test
```bash
dotnet test
```

### Database Migrations (PostgreSQL)
```bash
dotnet ef migrations add InitialCreate
dotnet ef database update
```

## Integration with Agent

Agent sends:
1. **Enrollment:** POST `/agents/register` with CSR → receives signed cert
2. **Heartbeats:** POST `/v0/agents/heartbeat` every 60s with mTLS

Control plane:
1. Signs agent certificates via internal CA
2. Stores heartbeat signals in time-series
3. Evaluates health asynchronously after each heartbeat
4. Maintains baselines and health state transitions

## Security

- Internal CA for agent certificates (2048-bit RSA)
- mTLS authentication for heartbeats (to be enforced)
- Agent private keys never leave agent systems
- TLS 1.2+ required

## Next Steps

- [ ] Enforce mTLS authentication on heartbeat endpoint
- [ ] Add UI query endpoints (agent detail, health history)
- [ ] Implement certificate rotation
- [ ] Add gRPC/WebSocket support
- [ ] Multi-tenant support
