# Data Model

## Overview

This document defines the complete database schema for Yggdrasil v0.5. The data model is designed with multi-tenancy, scalability, and future extensibility in mind.

## Database Technology

- **Primary Database**: PostgreSQL 15+
- **Time-Series Extension**: TimescaleDB (for metrics)
- **Multi-Tenancy**: Row-Level Security (RLS)
- **Query Builder**: sqlc (type-safe SQL)
- **Migrations**: golang-migrate

## Entity Relationship Diagram

```bash
┌─────────────────┐
│   MSPTenant     │
│─────────────────│
│ id (PK)         │
│ name            │
│ slug            │
│ settings (JSON) │
│ created_at      │
└────────┬────────┘
         │ 1:N
         │
┌────────▼────────┐
│    SubOrg       │
│─────────────────│
│ id (PK)         │
│ msp_tenant_id FK│
│ name            │
│ parent_id FK    │◄──── Self-referential (nested orgs)
└────────┬────────┘
         │ 1:N
         │
┌────────▼────────┐         ┌─────────────────┐
│     Client      │         │      User       │
│─────────────────│         │─────────────────│
│ id (PK)         │         │ id (PK)         │
│ msp_tenant_id FK│         │ msp_tenant_id FK│
│ sub_org_id FK   │         │ email           │
│ name            │         │ password_hash   │
│ status          │         │ role            │
│ contact_info    │         │ is_client_user  │
│ created_at      │         │ client_id FK    │◄─── For client portal users
└────────┬────────┘         └────────┬────────┘
         │ 1:N                       │
         │                           │ 1:N
┌────────▼────────┐         ┌────────▼────────┐
│     System      │         │     Ticket      │
│─────────────────│         │─────────────────│
│ id (PK)         │         │ id (PK)         │
│ msp_tenant_id FK│         │ msp_tenant_id FK│
│ client_id FK    │         │ client_id FK    │
│ hostname        │         │ system_id FK    │
│ ip_address      │         │ title           │
│ os_type         │         │ description     │
│ agent_version   │         │ status          │
│ last_seen       │         │ priority        │
└────────┬────────┘         │ assigned_to FK  │
         │                  │ created_by FK   │
         │ 1:1              │ created_at      │
         │                  │ updated_at      │
┌────────▼────────┐         └────────┬────────┘
│     Agent       │                  │ 1:N
│─────────────────│                  │
│ id (PK)         │         ┌────────▼────────┐
│ msp_tenant_id FK│         │   TimeEntry     │
│ system_id FK    │         │─────────────────│
│ certificate     │         │ id (PK)         │
│ status          │         │ msp_tenant_id FK│
│ approved_at     │         │ ticket_id FK    │
│ approved_by FK  │         │ user_id FK      │
└─────────────────┘         │ duration_mins   │
                            │ description     │
         ┌──────────────────│ billable        │
         │                  │ hourly_rate     │
         │ 1:N              │ created_at      │
┌────────▼────────┐         └─────────────────┘
│  MetricData     │
│─────────────────│         ┌─────────────────┐
│ time (PK)       │         │  TicketComment  │
│ msp_tenant_id FK│         │─────────────────│
│ system_id FK    │         │ id (PK)         │
│ metric_name     │         │ msp_tenant_id FK│
│ value           │         │ ticket_id FK    │
│ labels (JSON)   │         │ user_id FK      │
└─────────────────┘         │ content         │
  (TimescaleDB hypertable)  │ is_internal     │
                            │ created_at      │
                            └─────────────────┘
```

## Core Entities

### MSPTenant

The root tenant entity representing an MSP business.

```sql
CREATE TABLE msp_tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,  -- URL-friendly identifier
    settings JSONB DEFAULT '{}',        -- Tenant-specific configuration
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_msp_tenants_slug ON msp_tenants(slug);
```

**Key Fields**:

- `settings`: Flexible JSON for tenant preferences (timezone, branding, feature flags, etc.)
- `slug`: Used in URLs and API paths (e.g., `/api/v1/tenants/acme-msp/...`)

**Go Model with sqlc**:

```go
// db/models.go
package models

import (
    "database/sql/driver"
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

type MSPTenant struct {
    ID        uuid.UUID      `json:"id" db:"id"`
    Name      string         `json:"name" db:"name"`
    Slug      string         `json:"slug" db:"slug"`
    Settings  TenantSettings `json:"settings" db:"settings"`
    IsActive  bool           `json:"is_active" db:"is_active"`
    CreatedAt time.Time      `json:"created_at" db:"created_at"`
    UpdatedAt *time.Time     `json:"updated_at" db:"updated_at"`
}

// TenantSettings handles JSONB column
type TenantSettings map[string]interface{}

func (ts TenantSettings) Value() (driver.Value, error) {
    return json.Marshal(ts)
}

func (ts *TenantSettings) Scan(value interface{}) error {
    if value == nil {
        *ts = make(TenantSettings)
        return nil
    }

    switch v := value.(type) {
    case []byte:
        return json.Unmarshal(v, ts)
    case string:
        return json.Unmarshal([]byte(v), ts)
    }
    return nil
}

// Go model relationships are handled at the service layer
// using foreign key fields and joins
```

---

### SubOrg

Organizational units within an MSP (regional offices, departments, etc.).

```sql
CREATE TABLE sub_orgs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES sub_orgs(id) ON DELETE SET NULL,  -- For nested orgs
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sub_orgs_tenant ON sub_orgs(msp_tenant_id);
CREATE INDEX idx_sub_orgs_parent ON sub_orgs(parent_id);
```

**Key Features**:

- Optional (MSP can operate without sub-orgs)
- Self-referential for nested hierarchy
- Future: separate billing per sub-org (v1.0+)

**Go Model with sqlc**:

```go
// db/models.go
type SubOrg struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    MSPTenantID uuid.UUID  `json:"msp_tenant_id" db:"msp_tenant_id"`
    ParentID    *uuid.UUID `json:"parent_id" db:"parent_id"`
    Name        string     `json:"name" db:"name"`
    Description *string    `json:"description" db:"description"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

// sqlc query (queries.sql)
-- name: GetSubOrg :one
SELECT * FROM sub_orgs WHERE id = $1;

-- name: ListSubOrgsByTenant :many
SELECT * FROM sub_orgs WHERE msp_tenant_id = $1 ORDER BY name;

// Go model relationships are handled at the service layer
// using foreign key fields and joins
```

---

### Client

Customers of the MSP (businesses being managed).

```sql
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    sub_org_id UUID REFERENCES sub_orgs(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',  -- active, inactive, suspended
    contact_name VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    address TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_clients_tenant ON clients(msp_tenant_id);
CREATE INDEX idx_clients_sub_org ON clients(sub_org_id);
CREATE INDEX idx_clients_status ON clients(status);
```

**Go Model with sqlc**:

```go
// db/models.go
package models

import (
    "database/sql/driver"
    "time"

    "github.com/google/uuid"
)

type SubOrg struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    MSPTenantID uuid.UUID  `json:"msp_tenant_id" db:"msp_tenant_id"`
    ParentID    *uuid.UUID `json:"parent_id" db:"parent_id"`
    Name        string     `json:"name" db:"name"`
    Description *string    `json:"description" db:"description"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

// sqlc query (queries.sql)
-- name: GetSubOrg :one
SELECT * FROM sub_orgs WHERE id = $1 AND msp_tenant_id = $2;

-- name: ListSubOrgsByTenant :many
SELECT * FROM sub_orgs WHERE msp_tenant_id = $1 ORDER BY name;
```

// sqlc query examples
-- name: GetClient :one
SELECT \* FROM clients WHERE id = $1 AND msp_tenant_id = $2;

-- name: ListClientsByTenant :many
SELECT \* FROM clients WHERE msp_tenant_id = $1 ORDER BY name;

````

---

### System

Managed systems (servers, workstations, devices) where agents run.

```sql
CREATE TABLE systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    hostname VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),  -- IPv4 or IPv6
    os_type VARCHAR(50),     -- linux, windows, macos
    os_version VARCHAR(100),
    agent_version VARCHAR(50),
    last_seen TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',  -- Flexible for custom fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_systems_tenant ON systems(msp_tenant_id);
CREATE INDEX idx_systems_client ON systems(client_id);
CREATE INDEX idx_systems_hostname ON systems(hostname);
CREATE INDEX idx_systems_last_seen ON systems(last_seen);

**Go Model with sqlc**:

```go
type System struct {
    ID           uuid.UUID  `json:"id" db:"id"`
    MSPTenantID  uuid.UUID  `json:"msp_tenant_id" db:"msp_tenant_id"`
    ClientID     uuid.UUID  `json:"client_id" db:"client_id"`
    Hostname     string     `json:"hostname" db:"hostname"`
    IPAddress    *string    `json:"ip_address" db:"ip_address"`
    OSType       *string    `json:"os_type" db:"os_type"`
    OSVersion    *string    `json:"os_version" db:"os_version"`
    AgentVersion *string    `json:"agent_version" db:"agent_version"`
    LastSeen     *time.Time `json:"last_seen" db:"last_seen"`
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
}

// sqlc query examples
-- name: GetSystemByID :one
SELECT * FROM systems WHERE id = $1 AND msp_tenant_id = $2;

-- name: ListSystemsByClient :many
SELECT * FROM systems WHERE client_id = $1 AND msp_tenant_id = $2 ORDER BY hostname;
// Go model relationships are handled at the service layer
// using foreign key fields and joins
````

---

### Agent

Agent registration and authentication data.

```sql
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    system_id UUID UNIQUE NOT NULL REFERENCES systems(id) ON DELETE CASCADE,
    certificate TEXT,           -- PEM-encoded client certificate
    status VARCHAR(50) DEFAULT 'pending',  -- pending, approved, revoked
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agents_tenant ON agents(msp_tenant_id);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_heartbeat ON agents(last_heartbeat);
```

**Agent Lifecycle**:

1. Agent starts, generates certificate signing request (CSR)
2. Agent registers with control plane → `status = 'pending'`
3. MSP admin approves agent → `status = 'approved'`, certificate issued
4. Agent uses certificate for mTLS authentication
5. Agent sends heartbeat every 60s → `last_heartbeat` updated

**Go Model with sqlc**:

```go
// db/models.go
package models

import (
    "database/sql/driver"
    "time"

    "github.com/google/uuid"
)

type Agent struct {
    ID          uuid.UUID `json:"id" db:"id"`
    MSPTenantID uuid.UUID `json:"msp_tenant_id" db:"msp_tenant_id"`
    SystemID    uuid.UUID `json:"system_id" db:"system_id"`
    Certificate string    `json:"certificate" db:"certificate"`
    Status      string    `json:"status" db:"status"`
    ApprovedAt  *time.Time `json:"approved_at" db:"approved_at"`
    ApprovedBy  *uuid.UUID `json:"approved_by" db:"approved_by"`
    LastHeartbeat *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

// sqlc query (queries.sql)
-- name: GetAgent :one
SELECT * FROM agents WHERE id = $1 AND msp_tenant_id = $2;

-- name: ListAgentsByTenant :many
SELECT * FROM agents WHERE msp_tenant_id = $1 ORDER BY created_at DESC;
```

---

### User

MSP staff and client portal users.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL,  -- msp_admin, msp_tech, msp_manager, client_user
    is_client_user BOOLEAN DEFAULT FALSE,
    client_id UUID REFERENCES clients(id) ON DELETE CASCADE,  -- For client portal users
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_tenant ON users(msp_tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_client ON users(client_id);
```

**User Roles**:

- `msp_admin`: Full tenant access, user management, configuration
- `msp_tech`: Ticket/time entry management, client/system viewing
- `msp_manager`: Read-only + reports, approvals
- `client_user`: View own client's tickets/invoices (limited in v0.5)

**Go Model with sqlc**:

```go
// db/models.go
package models

import (
    "database/sql/driver"
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID          uuid.UUID `json:"id" db:"id"`
    MSPTenantID uuid.UUID `json:"msp_tenant_id" db:"msp_tenant_id"`
    Email       string     `json:"email" db:"email"`
    PasswordHash string     `json:"-"` db:"password_hash"`
    FirstName   *string    `json:"first_name" db:"first_name"`
    LastName    *string    `json:"last_name" db:"last_name"`
    Role        string     `json:"role" db:"role"`
    IsClientUser bool       `json:"is_client_user" db:"is_client_user"`
    ClientID    *uuid.UUID `json:"client_id" db:"client_id"`
    IsActive    bool       `json:"is_active" db:"is_active"`
    LastLogin   *time.Time `json:"last_login" db:"last_login"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

// sqlc query (queries.sql)
-- name: GetUser :one
SELECT * FROM users WHERE id = $1 AND msp_tenant_id = $2;

-- name: ListUsersByTenant :many
SELECT * FROM users WHERE msp_tenant_id = $1 ORDER BY email;
```

---

### Ticket

Support tickets for tracking work.

```sql
CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    system_id UUID REFERENCES systems(id) ON DELETE SET NULL,  -- Optional
    ticket_number SERIAL,  -- Auto-incrementing per-tenant ticket number
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'new',  -- new, assigned, in_progress, waiting, resolved, closed
    priority VARCHAR(50) DEFAULT 'medium',  -- low, medium, high, urgent
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tickets_tenant ON tickets(msp_tenant_id);
CREATE INDEX idx_tickets_client ON tickets(client_id);
CREATE INDEX idx_tickets_system ON tickets(system_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_assigned ON tickets(assigned_to);
CREATE INDEX idx_tickets_created_at ON tickets(created_at DESC);

-- Unique ticket number per tenant
CREATE UNIQUE INDEX idx_tickets_number_tenant ON tickets(msp_tenant_id, ticket_number);
```

**Ticket Workflow States**:

```bash
  ┌──────────────────────────────────────────────────┐
  │              Ticket Status Workflow              │
  │──────────────────────────────────────────────────│
  │                                                  │
  │
new → assigned → in_progress → waiting → resolved → closed
  └──────────────────────────────────────────────────┘
                  (can move between states)
```

**Go Model with sqlc**:

```go
type Ticket struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    MSPTenantID uuid.UUID  `json:"msp_tenant_id" db:"msp_tenant_id"`
    ClientID    uuid.UUID  `json:"client_id" db:"client_id"`
    SystemID    *uuid.UUID `json:"system_id" db:"system_id"`
    TicketNumber int        `json:"ticket_number" db:"ticket_number"`
    Title       string     `json:"title" db:"title"`
    Description *string    `json:"description" db:"description"`
    Status      string     `json:"status" db:"status"`
    Priority    string     `json:"priority" db:"priority"`
    AssignedTo  *uuid.UUID `json:"assigned_to" db:"assigned_to"`
    CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
    ResolvedAt  *time.Time `json:"resolved_at" db:"resolved_at"`
    ClosedAt    *time.Time `json:"closed_at" db:"closed_at"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}
```

---

### TicketComment

Comments/notes on tickets.

```sql
CREATE TABLE ticket_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT FALSE,  -- Internal notes vs client-visible
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ticket_comments_tenant ON ticket_comments(msp_tenant_id);
CREATE INDEX idx_ticket_comments_ticket ON ticket_comments(ticket_id);
CREATE INDEX idx_ticket_comments_created_at ON ticket_comments(created_at);
```

**Go Model with sqlc**:

```go
type TicketComment struct {
    ID          uuid.UUID `json:"id" db:"id"`
    MSPTenantID uuid.UUID `json:"msp_tenant_id" db:"msp_tenant_id"`
    TicketID    uuid.UUID `json:"ticket_id" db:"ticket_id"`
    UserID      uuid.UUID `json:"user_id" db:"user_id"`
    Content     string    `json:"content" db:"content"`
    IsInternal  bool      `json:"is_internal" db:"is_internal"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
```

---

### TimeEntry

Time tracking for tickets.

```sql
CREATE TABLE time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    duration_minutes INTEGER NOT NULL,
    description TEXT,
    billable BOOLEAN DEFAULT TRUE,
    hourly_rate DECIMAL(10, 2),  -- Store rate at time of entry (for historical accuracy)
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_time_entries_tenant ON time_entries(msp_tenant_id);
CREATE INDEX idx_time_entries_ticket ON time_entries(ticket_id);
CREATE INDEX idx_time_entries_user ON time_entries(user_id);
CREATE INDEX idx_time_entries_billable ON time_entries(billable);
CREATE INDEX idx_time_entries_created_at ON time_entries(created_at);
```

**Go Model with sqlc**:

```go
type TimeEntry struct {
    ID             uuid.UUID `json:"id" db:"id"`
    MSPTenantID    uuid.UUID `json:"msp_tenant_id" db:"msp_tenant_id"`
    TicketID       uuid.UUID `json:"ticket_id" db:"ticket_id"`
    UserID         uuid.UUID `json:"user_id" db:"user_id"`
    DurationMinutes int       `json:"duration_minutes" db:"duration_minutes"`
    Description    *string   `json:"description" db:"description"`
    Billable      bool       `json:"billable" db:"billable"`
    HourlyRate    float64   `json:"hourly_rate" db:"hourly_rate"`
    StartedAt     time.Time  `json:"started_at" db:"started_at"`
    EndedAt       *time.Time `json:"ended_at" db:"ended_at"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// sqlc query examples
-- name: GetTimeEntry :one
SELECT * FROM time_entries WHERE id = $1;

-- name: ListTimeEntriesByTicket :many
SELECT * FROM time_entries WHERE ticket_id = $1 ORDER BY started_at DESC;

-- name: ListTimeEntriesByUser :many
SELECT * FROM time_entries WHERE user_id = $1 ORDER BY started_at DESC;
```

**Note**: `hourly_rate` stored per entry for historical accuracy (rates may change over time).

---

### MetricData

Time-series metrics from agents (TimescaleDB hypertable).

```sql
CREATE TABLE metric_data (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    system_id UUID NOT NULL REFERENCES systems(id) ON DELETE CASCADE,
    metric_name VARCHAR(100) NOT NULL,  -- cpu_usage, memory_usage, disk_usage, etc.
    value DOUBLE PRECISION NOT NULL,
    labels JSONB DEFAULT '{}',  -- Flexible labels (e.g., {"disk": "/dev/sda1"})
    PRIMARY KEY (time, system_id, metric_name)
);

-- Convert to TimescaleDB hypertable
SELECT create_hypertable('metric_data', 'time');

-- Create indexes
CREATE INDEX idx_metric_data_tenant ON metric_data(msp_tenant_id, time DESC);
CREATE INDEX idx_metric_data_system ON metric_data(system_id, time DESC);
CREATE INDEX idx_metric_data_name ON metric_data(metric_name, time DESC);
```

**TimescaleDB Features**:

- Automatic partitioning by time
- Compression for old data
- Continuous aggregates for rollups
- Retention policies

**Retention Policy** (example):

```sql
-- Keep raw data for 90 days
SELECT add_retention_policy('metric_data', INTERVAL '90 days');

-- Create continuous aggregate for hourly rollups (keep forever)
CREATE MATERIALIZED VIEW metric_data_hourly
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 hour', time) AS bucket,
       system_id,
       metric_name,
       AVG(value) AS avg_value,
       MAX(value) AS max_value,
       MIN(value) AS min_value
FROM metric_data
GROUP BY bucket, system_id, metric_name;
```

**Go Model with sqlc**:

```go
type MetricData struct {
    Time         time.Time `json:"time" db:"time"`
    MSPTenantID uuid.UUID `json:"msp_tenant_id" db:"msp_tenant_id"`
    SystemID    uuid.UUID `json:"system_id" db:"system_id"`
    MetricName  string    `json:"metric_name" db:"metric_name"`
    Value       float64   `json:"value" db:"value"`
    Labels      string    `json:"labels" db:"labels"` // JSON field for flexible labels
}

// sqlc query (queries.sql)
-- name: GetMetricData :many
SELECT * FROM metric_data
WHERE msp_tenant_id = $1
  AND time >= $2
  AND time <= $3
ORDER BY time DESC;

-- name: GetSystemMetrics :many
SELECT * FROM metric_data
WHERE system_id = $1 AND msp_tenant_id = $2
ORDER BY metric_name, time DESC;
```

---

## Multi-Tenancy Implementation

### Row-Level Security (RLS)

PostgreSQL RLS policies enforce tenant isolation at the database level.

```sql
-- Enable RLS on all tenant tables
ALTER TABLE msp_tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE sub_orgs ENABLE ROW LEVEL SECURITY;
ALTER TABLE clients ENABLE ROW LEVEL SECURITY;
ALTER TABLE systems ENABLE ROW LEVEL SECURITY;
ALTER TABLE agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE tickets ENABLE ROW LEVEL SECURITY;
ALTER TABLE ticket_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE metric_data ENABLE ROW LEVEL SECURITY;

-- Example policy for tickets table
CREATE POLICY tenant_isolation_policy ON tickets
    USING (msp_tenant_id::text = current_setting('app.current_tenant_id', TRUE));

-- Repeat for all tables with msp_tenant_id
```

### Setting Tenant Context (Go/pgx)

```go
func SetTenantContext(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) error {
    _, err := tx.Exec(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID.String())
    return err
}

// Middleware/dependency injection
func GetCurrentTenant(ctx context.Context, db *pgxpool.Pool) (*MSPTenant, error) {
    // Decode JWT from context, extract tenant_id
    claims := GetClaimsFromContext(ctx)
    tenantID, err := uuid.Parse(claims.TenantID)
    if err != nil {
        return nil, err
    }

    // Fetch tenant
    var tenant MSPTenant
    err = db.QueryRow(ctx,
        "SELECT id, name, slug, settings, is_active, created_at, updated_at FROM msp_tenants WHERE id = $1",
        tenantID).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Settings, &tenant.IsActive, &tenant.CreatedAt, &tenant.UpdatedAt)

    return &tenant, err
}
```

---

## Future Schema Extensions (v1.0+)

### Contracts (SLA/Billing)

```sql
CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    contract_type VARCHAR(50),  -- fixed, hourly, hybrid, prepaid_hours
    start_date DATE NOT NULL,
    end_date DATE,
    monthly_amount DECIMAL(10, 2),
    prepaid_hours INTEGER,
    hourly_rate DECIMAL(10, 2),
    auto_renew BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Invoices

```sql
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',  -- draft, sent, paid, overdue
    subtotal DECIMAL(10, 2) NOT NULL,
    tax DECIMAL(10, 2) DEFAULT 0,
    total DECIMAL(10, 2) NOT NULL,
    issued_date DATE,
    due_date DATE,
    paid_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE invoice_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity DECIMAL(10, 2) NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    time_entry_id UUID REFERENCES time_entries(id)  -- Link to time entries
);
```

### Agent Tasks (v1.0)

```sql
CREATE TABLE agent_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    task_type VARCHAR(100) NOT NULL,  -- script, docker, install, etc.
    payload JSONB NOT NULL,           -- Task-specific data
    status VARCHAR(50) DEFAULT 'queued',  -- queued, running, completed, failed
    result JSONB,                     -- Task output/result
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_agent_tasks_agent ON agent_tasks(agent_id, status);
```

---

## Migration Strategy

See [Database Migrations](./database-migrations.md) for detailed migration strategy.

**Key Principles**:

1. All schema changes via Alembic migrations
2. No manual SQL in production
3. Backwards-compatible migrations (add columns as nullable, then backfill, then set NOT NULL)
4. Test migrations against production-like data volume

---

## Performance Considerations

### Indexes

All critical query paths have indexes:

- Foreign keys (automatic in most cases)
- Status fields (for filtering)
- Timestamp fields (for sorting/ranges)
- Tenant IDs (for multi-tenant queries)

### Connection Pooling

```go
// pgxpool configuration
config, err := pgxpool.ParseConfig(DATABASE_URL)
if err != nil {
    log.Fatal(err)
}

config.MaxConns = 20           // Max connections per instance
config.MinConns = 5            // Minimum connections
config.MaxConnLifetime = time.Hour  // Recycle connections after 1 hour
config.HealthCheckPeriod = 30 * time.Second  // Check connection health
```

### Query Optimization

- Use proper JOINs with explicit foreign key relationships
- Paginate large result sets (LIMIT/OFFSET or cursor-based)
- Use `FOR UPDATE` for row locking when needed
- Select only needed columns to reduce data transfer
- Use prepared statements via sqlc for performance

---

## Next Steps

1. Review [API Contract](./03-api-contract.md) to see how this data model is exposed
2. Review [Database Migrations](./database-migrations.md) for migration setup
3. Review [Testing Strategy](./08-testing-strategy.md) for data seeding and fixtures
