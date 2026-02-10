# System Architecture

## Overview

Yggdrasil is a comprehensive MSP/CRM/ERP platform designed for Managed Service Providers. Inspired by Norse mythology, the platform consists of three main components with clear domain boundaries:

1. **Control Plane** - Go-based backend managing all business logic
2. **Web Application** - React/TypeScript frontend for MSP staff and clients
3. **Agent** - Go-based agent deployed on managed systems

### Service Domain Naming

The platform uses Norse mythology for service naming, providing memorable and marketable domain boundaries:

- **Týr** - Authentication & Identity (users, roles, permissions)
- **Mímir** - Work & Knowledge (tickets, tasks, time tracking)
- **Valhalla** - Administration (tenant management, system configuration)
- **Bifröst** - Customer Portal (client access, self-service)
- **Gjallarhorn** - Ticketing & Issues (escalation, SLAs)
- **Verdandi** - Notifications (email, webhook, realtime)
- **Heimdallr** - Monitoring (health, metrics, SLOs)
- **Urd** - Audit & Logging (compliance, trails)
- **Eir** - Remote Diagnostics (troubleshooting, remediation)
- **Smidr** - Edge Computing (agent orchestration, plugins)

## High-Level Architecture

```bash
┌─────────────────────────────────────────────────────────────────┐
│                         Web Application                         │
│                    (React + TypeScript + Vite)                  │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS/REST
                             │
┌────────────────────────────┴────────────────────────────────────┐
│                        Control Plane                            │
│                     (Go + Chi + sqlc + pgx)                     │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ REST API     │  │  gRPC API    │  │  Background  │         │
│  │ (MSP/Client) │  │  (Agents)    │  │  Workers     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                               │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │           Service Layer (Business Logic)                 │ │
│  │  • Týr (AuthService)      • Mímir (TicketingService)     │ │
│  │  • ClientService           • AgentService                │ │
│  │  • TimeTrackingService     • MetricsService              │ │
│  │  • Verdandi (Notifications) • Heimdallr (Monitoring)     │ │
│  └──────────────────────────────────────────────────────────┘ │
└───────┬──────────────────────────┬────────────────────────────┘
        │                          │
        │                          │ NATS Pub/Sub
        │                          │
        ├──────────────────────────┼──────────────────────────────┐
        │                          │                              │
        │                          │                              │
┌───────▼──────┐          ┌────────▼────────┐        ┌──────────▼────────┐
│  PostgreSQL  │          │      NATS       │        │   Agent (Go)      │
│ +TimescaleDB │          │  Message Queue  │        │                   │
│              │          │                 │        │  ┌──────────────┐ │
│ • Multi-     │          │ • Task Queue    │        │  │  Metrics     │ │
│   Tenant     │          │ • Agent Comms   │        │  │  Plugin      │ │
│ • Row-Level  │          │ • Per-Tenant    │        │  └──────────────┘ │
│   Security   │          │   Isolation     │        │                   │
│ • Time-      │          │                 │        │  Runs on Customer │
│   Series     │          └─────────────────┘        │  Systems          │
└──────────────┘                                     └───────────────────┘
```

## Component Details

### Control Plane (Go)

**Purpose**: Core business logic, API gateway, data management

**Technology Stack**:

- Go 1.22+
- Chi (web router - stdlib-based)
- sqlc (type-safe SQL code generation)
- pgx/v5 (PostgreSQL driver)
- golang-migrate (database migrations)
- koanf (configuration management)
- cobra (CLI framework)
- slog (structured logging - stdlib)
- google.golang.org/grpc (agent communication)
- nhooyr.io/websocket (real-time updates)

**Responsibilities**:

- REST API for web application (tickets, clients, time tracking, etc.)
- gRPC API for agent communication
- Business logic orchestration via service layer
- Multi-tenant data isolation (row-level security)
- Authentication & authorization (JWT-based)
- Background task processing
- Metrics aggregation and storage

**Deployment**:

- Docker container (single binary Go application)
- Kubernetes-ready (StatefulSet or Deployment)
- Horizontal scaling supported
- Stateless (session stored in JWT/database)

### Web Application (React/TypeScript)

**Purpose**: User interface for MSP staff and client portal

**Technology Stack**:

- React 18 + TypeScript 5
- Vite (build tool)
- React Router v6 (routing)
- TanStack Query (server state)
- React Hook Form + Zod (forms/validation)
- TailwindCSS + shadcn/ui (styling)
- TanStack Table (data tables)

**User Personas**:

1. **MSP Admin** - Full system access, tenant configuration
2. **MSP Technician** - Ticket management, time tracking, agent management
3. **MSP Manager** - Reporting, oversight, approvals
4. **Client User** - View tickets, invoices, system status (read-only in v0.5)

**Key Features (v0.5)**:

- Ticket management dashboard
- Client and system management
- Time entry tracking
- Agent status monitoring
- Metrics visualization (basic)

### Agent (Go)

**Purpose**: Lightweight agent deployed on managed systems for monitoring and task execution

**Technology Stack**:

- Go 1.22+ (single binary, cross-platform)
- google.golang.org/grpc (communication with control plane)
- NATS Go client (task queue subscription)
- koanf (configuration management)
- hashicorp/go-plugin (multi-language plugin system)
- gopsutil (system metrics collection)
- slog (structured logging - stdlib)

**Responsibilities**:

- System metrics collection (CPU, RAM, disk, network)
- Task execution (scripts, Docker commands, etc. in v1.0+)
- Heartbeat/health reporting
- Plugin execution (multi-language via go-plugin)

**Communication Model**:

- **Pull-based**: Agent polls NATS for tasks
- **Push metrics**: Agent sends metrics via gRPC to control plane
- **Authentication**: Mutual TLS (mTLS) with control plane

**Deployment**:

- Single binary distribution
- Runs as system service (systemd, Windows Service, launchd)
- Minimal resource footprint (~20-50MB RAM)
- Self-update capability (future)

## Data Flow Examples

### Ticket Creation Flow

```bash
User → Web App → REST API → TicketingService → PostgreSQL
                                  ↓
                           Background Task → Notification (email/webhook)
```

### Agent Metrics Collection Flow

```bash
Agent → gRPC → MetricsService → TimescaleDB
  ↓
Metrics Plugin (collects CPU/RAM/Disk)
  ↓
Batched every 60s
```

### Agent Task Execution Flow (v1.0+)

```bash
User → Web App → REST API → AgentService → NATS Queue
                                              ↓
                                           Agent (polls)
                                              ↓
                                        Execute Task
                                              ↓
                                        Report Result → gRPC → Control Plane
```

## Network Topology

### Ports and Protocols

| Component     | Port  | Protocol | Purpose             |
| ------------- | ----- | -------- | ------------------- |
| Control Plane | 8000  | HTTPS    | REST API (web app)  |
| Control Plane | 50051 | gRPC/TLS | Agent communication |
| NATS          | 4222  | TCP      | Message queue       |
| PostgreSQL    | 5432  | TCP      | Database            |
| Web App       | 3000  | HTTPS    | Frontend (dev)      |

### Network Requirements

**Agent → Control Plane**:

- Outbound HTTPS (443) for gRPC
- Outbound TCP (4222) for NATS
- No inbound ports required (agent initiates all connections)
- Works behind NAT/firewalls

**Web App → Control Plane**:

- HTTPS (443) for REST API

## Multi-Tenancy Model

### Tenant Hierarchy

```bash
MSP Tenant (Root)
  ├── Sub-Org 1 (optional)
  │   ├── Client A
  │   │   ├── System 1 (Agent)
  │   │   └── System 2 (Agent)
  │   └── Client B
  └── Sub-Org 2 (optional)
      └── Client C
```

### Data Isolation Strategy

**Row-Level Security (RLS)**: Default for v0.5

- All tables include `msp_tenant_id` column
- PostgreSQL RLS policies enforce tenant isolation
- Application sets `SET LOCAL app.current_tenant_id = ?` per request
- Prevents cross-tenant data leakage

**Database-per-Tenant**: Future premium option (v2.0+)

- Complete data isolation
- Independent backups/restores
- Higher resource usage
- Connection routing layer needed

### Tenant Context Propagation

```go
// Every request carries tenant context
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract tenant from JWT token
        token := extractToken(r)
        claims, err := validateToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        tenantID := claims.TenantID

        // Add tenant to request context
        ctx := context.WithValue(r.Context(), "tenant_id", tenantID)

        // Set PostgreSQL session variable for RLS
        // This happens in the repository layer per transaction

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// In repository layer, set RLS context
func (r *Repository) withTenantContext(ctx context.Context, fn func(context.Context) error) error {
    tenantID := ctx.Value("tenant_id").(uuid.UUID)

    // Execute in transaction with RLS set
    return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
        _, err := tx.Exec(ctx,
            "SET LOCAL app.current_tenant_id = $1",
            tenantID)
        if err != nil {
            return err
        }

        return fn(ctx)
    })
}
```

## Security Architecture

### Authentication

**JWT-based Authentication**:

- Access tokens (15min expiry)
- Refresh tokens (7 day expiry)
- RS256 signing (public/private key pair)
- Stored in httpOnly cookies (web) or secure storage (agent)

**Agent Authentication**:

- Mutual TLS (mTLS) certificates
- Agent certificate signed by control plane CA
- Certificate includes tenant_id in subject
- Automatic rotation (v1.0+)

### Authorization

**Role-Based Access Control (RBAC)**:

| Role           | Permissions                                           |
| -------------- | ----------------------------------------------------- |
| MSP Admin      | Full access to tenant, manage users, configure system |
| MSP Technician | Manage tickets, time entries, view clients/systems    |
| MSP Manager    | Read-only + reports, approve time entries             |
| Client User    | View own tickets/invoices (v0.5: read-only)           |

**Permission Model**:

```go
// Go middleware-based permission checks
func requirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := getUserFromContext(r.Context())
            if !user.HasPermission(permission) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func createTicketHandler(w http.ResponseWriter, r *http.Request) {
    // Only users with "tickets:write" can access
    // Handler logic here
}
```

### Data Security

- **Encryption at Rest**: PostgreSQL encryption (deployment-specific)
- **Encryption in Transit**: TLS 1.3 for all connections
- **Secrets Management**: Environment variables (v0.5), HashiCorp Vault (v1.5+)
- **Audit Logging**: All mutations logged with user/timestamp

## Scalability Considerations

### Horizontal Scaling

**Control Plane**:

- Stateless design allows multiple instances
- Load balancer distributes traffic
- Shared PostgreSQL and NATS (single instance v0.5, clustered v1.5+)

**Database**:

- Connection pooling (pgxpool + pgbouncer)
- Read replicas for reporting (v1.5+)
- Partitioning for time-series metrics (TimescaleDB automatic)

**NATS**:

- Single instance sufficient for 100s of agents (v0.5)
- Clustering for high availability (v1.5+)

### Performance Targets (v0.5)

- API response time: < 200ms (p95)
- Concurrent agents: 500+ per control plane instance
- Metrics ingestion: 10,000 data points/second
- Database connections: 20-50 per instance

## Deployment Architecture

### Development (Docker Compose)

```yaml
services:
  postgres:
    image: timescale/timescaledb:latest-pg15

  nats:
    image: nats:latest

  control-plane:
    build:
      context: .
      dockerfile: Dockerfile
      target: api
    depends_on: [postgres, nats]

  web:
    build: ./web
    depends_on: [control-plane]

  agent:
    build:
      context: .
      dockerfile: Dockerfile
      target: agent
    depends_on: [control-plane, nats]
```

### Production (Future - Kubernetes)

```bash
Kubernetes Cluster
  ├── Namespace: yggdrasil-control
  │   ├── Deployment: control-plane (3 replicas)
  │   ├── Service: control-plane-rest (LoadBalancer)
  │   ├── Service: control-plane-grpc (LoadBalancer)
  │   ├── StatefulSet: nats (3 replicas)
  │   └── StatefulSet: postgres (1 primary + 2 replicas)
  │
  └── Namespace: yggdrasil-web
      ├── Deployment: web (2 replicas)
      └── Ingress: HTTPS (Let's Encrypt)

Agents (External)
  └── Customer Systems (connect via Internet → LoadBalancer)
```

## Monitoring & Observability

### Metrics (Prometheus)

- **Application Metrics**: Request rate, latency, errors
- **Business Metrics**: Tickets created, time tracked, agents online
- **Infrastructure Metrics**: CPU, memory, database connections

### Logging (Structured)

```json
{
  "timestamp": "2026-01-22T10:30:00Z",
  "level": "info",
  "service": "control-plane",
  "tenant_id": "msp-123",
  "user_id": "user-456",
  "action": "ticket_created",
  "ticket_id": "ticket-789",
  "duration_ms": 45
}
```

### Tracing (Future - OpenTelemetry)

- Distributed tracing across services
- Agent → NATS → Control Plane → Database spans

## Disaster Recovery

### Backup Strategy

- **Database**: Daily full backups, continuous WAL archiving
- **Metrics**: 90-day retention in TimescaleDB (configurable)
- **Configuration**: Version controlled in Git

### Recovery Objectives

- **RTO (Recovery Time Objective)**: < 4 hours
- **RPO (Recovery Point Objective)**: < 1 hour (database), < 5 minutes (metrics acceptable loss)

## Technology Decisions Summary

| Concern        | Choice                    | Rationale                                                  |
| -------------- | ------------------------- | ---------------------------------------------------------- |
| API Framework  | Chi + Go                  | Idiomatic, stdlib-based, performant                        |
| Query Builder  | sqlc                      | Type-safe SQL, zero reflection, compile-time verification  |
| Frontend       | React + TypeScript        | Industry standard, huge ecosystem, type safety             |
| Agent Language | Go                        | Single binary, low resources, cross-platform               |
| Message Queue  | NATS                      | Lightweight, cloud-native, perfect for agent communication |
| Database       | PostgreSQL+TimescaleDB    | Robust, RLS support, time-series extension                 |
| Monorepo Tool  | Go workspaces + Makefiles | Native Go monorepo support                                 |

---

## Phased Evolution Strategy

### Current Phase: Modular Monolith (v0.5-1.0)

**Architecture**: Single deployment with clean service boundaries

- Týr, Mímir, Valhalla, Bifröst as internal modules
- Clean Architecture enables future extraction
- Shared database with row-level security

**Benefits**: Simple deployment, clear boundaries, easy development

### Future Phase: Progressive Service Extraction (v1.5-2.0)

**Architecture**: Hybrid monolith + extracted services

- Smidr agents remain independent (v0.5)
- Heimdallr/Verdandi extraction for scale (v1.5+)
- Database-per-tenant option for premium tiers

**Allfadr Orchestration Layer** (v2.0+):

- Service discovery and routing
- Container orchestration (Docker/Kubernetes)
- Unified observability and tracing

---

## Control Plane Evolution (Ymir Foundation)

### v0.5: Basic Control

- Manual configuration via environment variables
- Basic feature flags in database
- Simple API key management

### v1.0+: Centralized Governance (Ymir)

- **Feature Flag Service**: Dynamic feature management
- **Quota Management**: Per-tenant resource limits
- **API Key Management**: Self-service key generation/rotation
- **Policy Coordination**: Cross-service configuration
- **Secrets Management**: Centralized secret distribution

### Benefits of Phased Approach

- **Lower Complexity**: Start simple, add sophistication when needed
- **Clear Migration Path**: Services designed for extraction from day one
- **Operational Clarity**: Each phase has clear operational model
- **Customer Choice**: Customers can choose complexity level

---

## Next Steps

1. Review [Data Model](./02-data-model.md) for database schema design
2. Review [API Contract](./03-api-contract.md) for endpoint specifications
3. Review [Plugin Architecture](./04-plugin-architecture.md) for extensibility design
4. Review [Implementation Roadmap](./implementation-roadmap.md) for v0.5 plan
