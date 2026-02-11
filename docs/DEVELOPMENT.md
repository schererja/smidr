# Smidr Development Guide

This guide covers local development setup, building, testing, and contributing to Smidr.

---

## Prerequisites

### Required Software

- **Go 1.21+** - Agent development
- **.NET 8.0 SDK** - Control plane development
- **Node.js 18+** and **npm** - UI development
- **PostgreSQL 14+** - Production database (SQLite works for development)
- **OpenSSL** - Certificate operations (usually pre-installed)
- **Git** - Version control

### Operating System

- **Development:** macOS, Linux, or WSL2 on Windows
- **Agent deployment:** Linux x86_64 only
- **Control plane deployment:** Any platform supporting .NET 8

### Recommended Tools

- **curl** - API testing
- **jq** - JSON parsing for scripts
- **systemd** - Agent service management (Linux only)
- **Visual Studio Code** - Editor with Go, C#, and TypeScript extensions

---

## Repository Structure

```
smidr/
├── agent/               # Go agent (Linux daemon)
│   ├── cmd/agent/       # Main entry point
│   ├── internal/        # Agent implementation
│   ├── systemd/         # systemd service unit
│   ├── install.sh       # Installation script
│   ├── go.mod           # Go dependencies
│   └── README.md        # Agent-specific docs
│
├── control-plane/       # C# control plane (ASP.NET Core API)
│   ├── Controllers/     # REST API endpoints
│   ├── Services/        # Business logic (CA, health evaluation)
│   ├── Data/            # Database context and repositories
│   ├── Models/          # Entity models
│   ├── Middleware/      # mTLS authentication
│   ├── Migrations/      # EF Core migrations
│   ├── appsettings.json # Configuration
│   ├── Program.cs       # Application entry point
│   └── README.md        # Control plane-specific docs
│
├── ui/                  # React UI (TypeScript)
│   ├── src/
│   │   ├── api/         # API client
│   │   ├── components/  # Reusable components
│   │   ├── pages/       # Page components
│   │   ├── types/       # TypeScript types
│   │   └── App.tsx      # Main app with routing
│   ├── package.json     # npm dependencies
│   ├── vite.config.ts   # Vite configuration
│   └── README.md        # UI-specific docs
│
├── scripts/             # Utility scripts
│   ├── ca-tool.sh       # CA certificate operations
│   └── test-mtls.sh     # mTLS testing
│
├── docs/                # Documentation
│   ├── ARCHITECTURE.md  # System architecture
│   ├── API.md           # API reference
│   └── DEVELOPMENT.md   # This file
│
└── README.md            # Project overview
```

---

## Getting Started

### 1. Clone Repository

```bash
git clone https://github.com/your-org/smidr.git
cd smidr
```

### 2. Build All Components

#### Agent (Go)

```bash
cd agent
go mod download
go build -o bin/smidr-agent ./cmd/agent
```

Binary: `agent/bin/smidr-agent`

#### Control Plane (C#)

```bash
cd control-plane
dotnet restore
dotnet build
```

Output: `control-plane/bin/Debug/net8.0/`

#### UI (React)

```bash
cd ui
npm install
npm run build
```

Output: `ui/dist/`

---

## Running Locally

### Control Plane

The control plane must start first to provide the API for agents and UI.

```bash
cd control-plane
dotnet run
```

**Default URLs:**
- HTTPS: `https://localhost:5001`
- HTTP: `http://localhost:5000`
- Swagger UI: `https://localhost:5001/swagger`

**First Run:**
- Auto-generates CA keypair at `control-plane/data/ca/`
- Creates SQLite database at `control-plane/data/controlplane.db`

**Configuration:**

Edit `appsettings.json` to change database, ports, or TLS settings:

```json
{
  "Logging": {
    "LogLevel": {
      "Default": "Information"
    }
  },
  "UsePostgres": false,
  "ConnectionStrings": {
    "PostgreSQL": "Host=localhost;Database=smidr;Username=smidr;Password=smidr"
  }
}
```

Set `UsePostgres: true` to use PostgreSQL instead of SQLite.

### UI (Development Mode)

```bash
cd ui
npm run dev
```

**Dev server:** `http://localhost:3000`

**Features:**
- Hot module replacement (HMR)
- Auto-reload on file changes
- Proxies API requests to `https://localhost:5001`

**Configuration:**

Create `.env` file (copy from `.env.example`):

```bash
VITE_API_BASE_URL=https://localhost:5001
```

### Agent (Development Mode)

**Note:** Agent requires Linux. Use a VM or WSL2 if developing on macOS/Windows.

#### Initialize Config

```bash
cd agent
sudo ./bin/smidr-agent init --config /tmp/smidr-test.yaml
```

Edit `/tmp/smidr-test.yaml`:

```yaml
control_plane_url: "https://localhost:5001"
agent_id: ""  # auto-generated on first run
hostname: ""  # auto-detected
key_path: "/tmp/smidr-test.key.pem"
csr_path: "/tmp/smidr-test.csr.pem"
cert_path: "/tmp/smidr-test.crt.pem"
token: ""
heartbeat_interval: 60
```

#### Run Agent Daemon

```bash
sudo ./bin/smidr-agent daemon --config /tmp/smidr-test.yaml
```

**First run:**
- Generates agent_id and keypair
- Enrolls with control plane
- Starts sending heartbeats every 60 seconds

**Logs:** Output to stdout/stderr (use `journalctl` when running via systemd)

---

## Development Workflow

### Typical Development Session

1. **Start control plane:**
   ```bash
   cd control-plane && dotnet run
   ```

2. **Start UI dev server:**
   ```bash
   cd ui && npm run dev
   ```

3. **Access UI:**
   Open `http://localhost:3000` in browser

4. **Test API:**
   Use Swagger UI at `https://localhost:5001/swagger`

5. **Run agent (optional):**
   ```bash
   cd agent && sudo ./bin/smidr-agent daemon --config /tmp/smidr-test.yaml
   ```

### Hot Reload Support

- **Control plane:** Restart required for code changes (`dotnet run` or `dotnet watch`)
- **UI:** Automatic via Vite HMR
- **Agent:** Restart required for code changes

### Using `dotnet watch` (Control Plane)

For automatic restart on file changes:

```bash
cd control-plane
dotnet watch run
```

---

## Testing

### Agent Tests

```bash
cd agent
go test ./... -v
```

**Test Coverage:**

```bash
go test ./... -cover
```

**Run Specific Test:**

```bash
go test ./internal/signals -run TestCollectSignals -v
```

**Note:** Signal collection tests require Linux. Mock tests work on all platforms.

### Control Plane Tests

```bash
cd control-plane
dotnet test
```

**Run with Coverage:**

```bash
dotnet test /p:CollectCoverage=true
```

**Run Specific Test:**

```bash
dotnet test --filter "FullyQualifiedName~CaServiceTests.IssueAgentCertificate_ValidCsr_ReturnsSignedCert"
```

### UI Tests

**Note:** v0 does not include UI tests yet. Planned for future:

```bash
cd ui
npm test              # Vitest unit tests
npm run test:e2e      # Playwright E2E tests (future)
```

### Integration Tests

**Script:** `scripts/test-mtls.sh`

End-to-end test of enrollment and heartbeat flow:

```bash
cd scripts
./test-mtls.sh
```

**What it does:**
1. Generates test agent keypair
2. Creates CSR
3. Enrolls agent with control plane
4. Sends mTLS-authenticated heartbeat
5. Verifies 200 OK response

---

## Database Management

### SQLite (Development)

**Location:** `control-plane/data/controlplane.db`

**Reset database:**

```bash
cd control-plane
rm data/controlplane.db
dotnet run  # Recreates database on startup
```

**Inspect with sqlite3:**

```bash
sqlite3 control-plane/data/controlplane.db
.tables
SELECT * FROM agents;
.quit
```

### PostgreSQL (Production)

#### Setup PostgreSQL

**macOS (Homebrew):**

```bash
brew install postgresql@14
brew services start postgresql@14
createdb smidr
psql smidr -c "CREATE USER smidr WITH PASSWORD 'smidr';"
psql smidr -c "GRANT ALL PRIVILEGES ON DATABASE smidr TO smidr;"
```

**Linux (Ubuntu):**

```bash
sudo apt install postgresql-14
sudo -u postgres createuser -P smidr  # Enter password when prompted
sudo -u postgres createdb -O smidr smidr
```

#### Configure Control Plane

Edit `appsettings.json`:

```json
{
  "UsePostgres": true,
  "ConnectionStrings": {
    "PostgreSQL": "Host=localhost;Database=smidr;Username=smidr;Password=smidr"
  }
}
```

#### Run Migrations

```bash
cd control-plane
dotnet ef database update
```

**Create new migration:**

```bash
dotnet ef migrations add MigrationName
```

---

## Building for Production

### Agent (Go)

**Single platform:**

```bash
cd agent
GOOS=linux GOARCH=amd64 go build -o bin/smidr-agent-linux-amd64 ./cmd/agent
```

**GoReleaser (multi-platform, future):**

```bash
goreleaser release --snapshot --clean
```

Output: `agent/dist/`

### Control Plane (C#)

**Publish for deployment:**

```bash
cd control-plane
dotnet publish -c Release -o ./publish
```

Output: `control-plane/publish/`

**Self-contained binary (includes .NET runtime):**

```bash
dotnet publish -c Release -r linux-x64 --self-contained -o ./publish
```

### UI (React)

**Production build:**

```bash
cd ui
npm run build
```

Output: `ui/dist/`

**Serve static files:**

Configure control plane to serve UI from `wwwroot/`:

1. Copy `ui/dist/*` to `control-plane/wwwroot/`
2. Control plane serves static files at root path

---

## Debugging

### Control Plane (Visual Studio Code)

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": ".NET Core Launch (control-plane)",
      "type": "coreclr",
      "request": "launch",
      "preLaunchTask": "build",
      "program": "${workspaceFolder}/control-plane/bin/Debug/net8.0/control-plane.dll",
      "args": [],
      "cwd": "${workspaceFolder}/control-plane",
      "stopAtEntry": false,
      "serverReadyAction": {
        "action": "openExternally",
        "pattern": "\\bNow listening on:\\s+(https?://\\S+)"
      },
      "env": {
        "ASPNETCORE_ENVIRONMENT": "Development"
      }
    }
  ]
}
```

**Set breakpoints** in Controllers, Services, or Middleware.

### Agent (VS Code with Delve)

Install Delve:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Run with debugger:

```bash
cd agent
sudo dlv debug ./cmd/agent -- daemon --config /tmp/smidr-test.yaml
```

Or use VS Code Go extension for integrated debugging.

### UI (Browser DevTools)

- Use browser DevTools (F12)
- Source maps enabled in Vite dev mode
- React DevTools extension recommended

---

## Code Style and Conventions

### Go (Agent)

- Follow Go standard style (`gofmt`, `go vet`)
- Use `golangci-lint` for linting (future)
- Package-level comments for exported types
- Error handling: wrap with `fmt.Errorf` and `%w`

**Example:**

```go
func CollectSignals() (Signals, error) {
    uptime, err := readUptime()
    if err != nil {
        return Signals{}, fmt.Errorf("failed to read uptime: %w", err)
    }
    // ...
}
```

### C# (Control Plane)

- Follow .NET conventions (PascalCase, async/await, `sealed` classes)
- Use `required` for mandatory properties
- Prefer records for DTOs
- XML comments for public APIs (future)

**Example:**

```csharp
public sealed record RegisterAgentRequest(
    string AgentId,
    string Hostname,
    string? Token,
    string CsrPem
);
```

### TypeScript (UI)

- Use TypeScript strict mode
- Prefer functional components with hooks
- Props interfaces with PascalCase
- camelCase for variables and functions

**Example:**

```typescript
interface SystemListProps {
  systems: System[];
  onSystemClick: (id: string) => void;
}

export function SystemList({ systems, onSystemClick }: SystemListProps) {
  // ...
}
```

---

## Contributing

### Branching Strategy

- `main` - production-ready code
- `develop` - integration branch for features
- `feature/xyz` - feature branches
- `fix/xyz` - bug fix branches

### Commit Messages

Follow conventional commits:

```
feat(agent): add disk usage signal collector
fix(control-plane): handle null baselines in health evaluation
docs(api): document heartbeat endpoint response codes
test(ui): add SystemList component tests
```

**Types:** `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

### Pull Request Process

1. Create feature branch from `develop`
2. Make changes and commit
3. Write tests for new functionality
4. Run all tests locally
5. Push branch and open PR to `develop`
6. Address review feedback
7. Squash and merge after approval

### Code Review Checklist

- [ ] Tests pass locally
- [ ] Code follows project style conventions
- [ ] New functionality has tests
- [ ] API changes documented in `docs/API.md`
- [ ] Architecture changes documented in `docs/ARCHITECTURE.md`
- [ ] Breaking changes called out in PR description

---

## Troubleshooting

### Control Plane Won't Start

**Error:** `Unable to bind to https://localhost:5001`

**Solution:** Port already in use. Change in `appsettings.json` or kill existing process.

**Error:** `CA certificate generation failed`

**Solution:** Ensure write permissions on `control-plane/data/ca/` directory.

**Error:** `Database connection failed`

**Solution:** Check PostgreSQL is running and credentials in `appsettings.json` are correct.

### Agent Won't Enroll

**Error:** `failed to register: connection refused`

**Solution:** Control plane not running. Start with `cd control-plane && dotnet run`.

**Error:** `certificate verify failed`

**Solution:** Agent can't verify control plane's TLS certificate. Use `--insecure` flag for self-signed certs (development only).

**Error:** `agentId is required`

**Solution:** Config file missing or corrupt. Re-run `init` command.

### UI Won't Connect to API

**Error:** `Cannot connect to control plane API`

**Solution:** 
- Check control plane is running (`https://localhost:5001`)
- Verify `VITE_API_BASE_URL` in `.env` is correct
- Check browser console for CORS errors

**Error:** `Mixed content blocked`

**Solution:** UI served over HTTPS but API is HTTP. Use HTTPS for control plane or serve UI over HTTP.

### Tests Failing

**Go tests fail on macOS:**

**Solution:** Signal collection tests require Linux. Use `GOOS=linux go test` or run in Docker/VM.

**dotnet test timeout:**

**Solution:** Database tests may be slow. Increase timeout or use in-memory database for tests.

---

## Development Tools

### Recommended VS Code Extensions

**Go:**
- `golang.go` - Go language support
- `ms-vscode.vscode-go-testing` - Go test UI

**C#:**
- `ms-dotnettools.csharp` - C# language support
- `ms-dotnettools.csdevkit` - .NET SDK integration

**TypeScript/React:**
- `dbaeumer.vscode-eslint` - ESLint integration
- `esbenp.prettier-vscode` - Prettier formatting
- `dsznajder.es7-react-js-snippets` - React snippets

**General:**
- `eamodio.gitlens` - Git history and blame
- `humao.rest-client` - Test HTTP endpoints from editor

### REST Client Examples

Create `.vscode/api-tests.http`:

```http
### Enroll Agent
POST https://localhost:5001/agents/register
Content-Type: application/json

{
  "agentId": "test-agent-123",
  "hostname": "test-host",
  "csrPem": "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----"
}

### List Agents
GET https://localhost:5001/api/agents

### Get Agent Detail
GET https://localhost:5001/api/agents/test-agent-123
```

Run with "Send Request" (Ctrl+Alt+R).

---

## CI/CD

### GitHub Actions (Future)

Planned workflows:

**`.github/workflows/test.yml`** - Run tests on PR
**`.github/workflows/build.yml`** - Build all components
**`.github/workflows/release.yml`** - Create release with GoReleaser

### Docker (Future)

Planned Dockerfiles:

- `agent/Dockerfile` - Agent container
- `control-plane/Dockerfile` - Control plane container
- `docker-compose.yml` - Full stack for testing

---

## Performance Profiling

### Control Plane (dotnet-trace)

```bash
dotnet tool install --global dotnet-trace
cd control-plane
dotnet run &
dotnet trace collect --process-id $(pgrep -f control-plane) --providers Microsoft-Windows-DotNETRuntime
```

### Agent (pprof)

Add to `main.go`:

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Profile:

```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

---

## Security Testing

### mTLS Validation

Test certificate rejection:

```bash
# Generate invalid cert (not signed by CA)
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout invalid.key.pem -out invalid.crt.pem \
  -days 365 -subj "/CN=invalid-agent"

# Attempt heartbeat (should fail)
curl -X POST https://localhost:5001/v0/agents/heartbeat \
  --cert invalid.crt.pem --key invalid.key.pem \
  -H 'Content-Type: application/json' -d '{...}'
# Expected: TLS handshake failure or 403 Forbidden
```

### SQL Injection Testing (Future)

Use sqlmap or similar tools to test query endpoints:

```bash
sqlmap -u "https://localhost:5001/api/agents?id=test" --batch
```

Expected: No vulnerabilities (EF Core uses parameterized queries).

---

## Getting Help

- **Documentation:** See `docs/` directory
- **Issues:** Open GitHub issue with reproduction steps
- **Slack/Discord:** (Add team communication channel if applicable)

---

**End of Development Guide**
