# Project Structure

## Overview

Yggdrasil uses a **monorepo** structure managed by Turborepo for TypeScript/JavaScript packages and Makefiles for Go components. This provides a unified development experience while maintaining clear boundaries between components.

---

## Repository Layout

```bash
yggdrasil/
├── cmd/
│   ├── api/                    # Control plane entry point
│   │   └── main.go
│   └── agent/                  # Agent entry point
│       └── main.go
├── internal/
│   ├── ticket/                 # Ticket domain (vertical slice)
│   ├── client/                 # Client domain
│   ├── agent/                  # Agent domain
│   ├── auth/                   # Authentication domain
│   ├── platform/               # Platform infrastructure
│   │   ├── config/
│   │   ├── database/
│   │   ├── messaging/
│   │   └── http/
│   └── shared/                 # Shared utilities
├── pkg/                        # Public packages (if needed)
├── web/                        # React frontend
│   ├── src/
│   ├── public/
│   └── package.json
├── packages/
│   ├── proto/                  # Protocol Buffer definitions
│   └── agent-sdk/              # Go SDK for agent plugins
├── migrations/                 # Database migrations (golang-migrate)
├── docs/
│   ├── architecture/           # Architecture documentation (this folder)
│   ├── api/                    # API documentation
│   └── guides/                 # User guides
├── scripts/
│   ├── setup-dev.sh            # Development environment setup
│   ├── generate-certs.sh       # Generate mTLS certificates
│   └── seed-data.sh            # Seed database with test data
├── .github/
│   └── workflows/              # GitHub Actions CI/CD
├── docker-compose.yml          # Local development environment
├── Makefile                    # Build automation
├── go.mod                      # Go module (root level)
├── go.sum
├── sqlc.yaml                   # sqlc configuration
├── .air.toml                   # Live reload for development
├── .env.example                # Example environment variables
└── README.md
```

---

## Domain-Based Structure

Yggdrasil uses **package-oriented design** with domain-based vertical slices. Each domain is self-contained with entity, repository, service, handler, and database implementation.

### Control Plane & Agent (Go/Chi)

```bash
cmd/
├── api/                        # Control plane entry point
│   └── main.go
└── agent/                      # Agent entry point
    └── main.go

internal/
├── ticket/                     # Ticket domain (vertical slice)
│   ├── ticket.go               # Entity + business rules
│   ├── repository.go           # Repository interface
│   ├── service.go              # Use cases/orchestration
│   ├── handler.go              # HTTP handlers
│   ├── postgres.go             # Database implementation
│   ├── queries.sql             # sqlc queries
│   ├── models.go               # sqlc generated (gitignored)
│   ├── ticket_test.go          # Entity tests
│   ├── service_test.go         # Service tests
│   └── handler_test.go         # Handler tests
│
├── client/                     # Client domain
│   ├── client.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go
│   ├── postgres.go
│   └── queries.sql
│
├── agent/                      # Agent domain
│   ├── agent.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go              # HTTP handlers
│   ├── grpc_server.go          # gRPC service for agent<->control plane
│   ├── postgres.go
│   └── queries.sql
│
├── auth/                       # Authentication domain
│   ├── user.go
│   ├── session.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go              # Login, register, etc.
│   ├── middleware.go           # Auth middleware
│   ├── postgres.go
│   └── queries.sql
│
├── platform/                   # Platform infrastructure (shared)
│   ├── config/
│   │   └── config.go           # koanf configuration
│   ├── database/
│   │   └── postgres.go         # Database connection
│   ├── messaging/
│   │   └── nats.go             # NATS client
│   └── http/
│       ├── router.go           # Chi router setup
│       └── middleware/
│           ├── logging.go
│           ├── recovery.go
│           └── cors.go
│
└── shared/                     # Shared utilities
    ├── errors.go               # Error types
    ├── tenant.go               # Multi-tenancy utilities
    └── pagination.go           # Pagination helpers

pkg/                            # Public packages (if needed)

migrations/                     # golang-migrate SQL files
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000002_add_tickets.up.sql
└── 000002_add_tickets.down.sql
```

**Key Files**:

`cmd/api/main.go`:

```go
package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"

    "yggdrasil/internal/ticket"
    "yggdrasil/internal/client"
    "yggdrasil/internal/agent"
    "yggdrasil/internal/auth"
    "yggdrasil/internal/platform/config"
    "yggdrasil/internal/platform/database"
    "yggdrasil/internal/platform/http/router"
)
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        slog.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    // Setup structured logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    // Initialize database
    db, err := database.Connect(cfg.Database)
    if err != nil {
        slog.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    // Initialize repositories
    ticketRepo := ticket.NewPostgresRepository(db)
    clientRepo := client.NewPostgresRepository(db)
    agentRepo := agent.NewPostgresRepository(db)
    authRepo := auth.NewPostgresRepository(db)

    // Initialize services
    ticketSvc := ticket.NewService(ticketRepo)
    clientSvc := client.NewService(clientRepo)
    agentSvc := agent.NewService(agentRepo)
    authSvc := auth.NewService(authRepo)

    // Initialize handlers
    ticketHandler := ticket.NewHandler(ticketSvc)
    clientHandler := client.NewHandler(clientSvc)
    agentHandler := agent.NewHandler(agentSvc)
    authHandler := auth.NewHandler(authSvc)

    // Setup router
    r := router.New(router.Config{
        Auth:   authHandler,
        Ticket: ticketHandler,
        Client: clientHandler,
        Agent:  agentHandler,
    })

    // Start server
    srv := &http.Server{
        Addr:    cfg.Server.Address,
        Handler: r,
    }

    go func() {
        slog.Info("starting server", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server error", "error", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("server shutdown error", "error", err)
    }
}
```

Example domain implementation (`internal/ticket/ticket.go`):

```go
package ticket

import (
    "time"
    "errors"
)

// Entity
type Ticket struct {
    ID          string
    TenantID    string
    ClientID    string
    Title       string
    Description string
    Status      Status
    Priority    Priority
    AssignedTo  *string
    CreatedBy   string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Status string

const (
    StatusOpen       Status = "open"
    StatusInProgress Status = "in_progress"
    StatusResolved   Status = "resolved"
    StatusClosed     Status = "closed"
)

type Priority string

const (
    PriorityLow      Priority = "low"
    PriorityMedium   Priority = "medium"
    PriorityHigh     Priority = "high"
    PriorityCritical Priority = "critical"
)

// Business rules
func (t *Ticket) Validate() error {
    if t.Title == "" {
        return errors.New("title is required")
    }
    if t.TenantID == "" {
        return errors.New("tenant_id is required")
    }
    if t.ClientID == "" {
        return errors.New("client_id is required")
    }
    return nil
}

func (t *Ticket) CanTransitionTo(newStatus Status) bool {
    // Business logic for valid state transitions
    validTransitions := map[Status][]Status{
        StatusOpen:       {StatusInProgress, StatusClosed},
        StatusInProgress: {StatusResolved, StatusClosed},
        StatusResolved:   {StatusClosed, StatusInProgress},
        StatusClosed:     {StatusOpen},
    }

    allowed := validTransitions[t.Status]
    for _, s := range allowed {
        if s == newStatus {
            return true
        }
    }
    return false
}
```

Example repository interface (`internal/ticket/repository.go`):

```go
package ticket

import "context"

type Repository interface {
    Create(ctx context.Context, ticket *Ticket) error
    GetByID(ctx context.Context, tenantID, id string) (*Ticket, error)
    Update(ctx context.Context, ticket *Ticket) error
    Delete(ctx context.Context, tenantID, id string) error
    List(ctx context.Context, tenantID string, filters Filters) ([]*Ticket, error)
}

type Filters struct {
    ClientID   *string
    Status     *Status
    Priority   *Priority
    AssignedTo *string
    Limit      int
    Offset     int
}
```

Example sqlc queries (`internal/ticket/queries.sql`):

```sql
-- name: CreateTicket :exec
INSERT INTO tickets (
    id, tenant_id, client_id, title, description,
    status, priority, assigned_to, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: GetTicketByID :one
SELECT * FROM tickets
WHERE tenant_id = $1 AND id = $2;

-- name: UpdateTicket :exec
UPDATE tickets
SET title = $2,
    description = $3,
    status = $4,
    priority = $5,
    assigned_to = $6,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2;

-- name: ListTickets :many
SELECT * FROM tickets
WHERE tenant_id = $1
  AND ($2::text IS NULL OR client_id = $2)
  AND ($3::text IS NULL OR status = $3)
  AND ($4::text IS NULL OR priority = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;
```

Example configuration (`internal/platform/config/config.go`):

```go
package config

import (
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
    NATS     NATSConfig
}

type ServerConfig struct {
    Address string `koanf:"address"`
    Port    int    `koanf:"port"`
}

type DatabaseConfig struct {
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    Name     string `koanf:"name"`
    User     string `koanf:"user"`
    Password string `koanf:"password"`
    SSLMode  string `koanf:"ssl_mode"`
}

type AuthConfig struct {
    JWTSecret     string `koanf:"jwt_secret"`
    TokenDuration int    `koanf:"token_duration"`
}

type NATSConfig struct {
    URL string `koanf:"url"`
}

func Load() (*Config, error) {
    k := koanf.New(".")

    // Load from config file
    if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
        return nil, err
    }

    // Override with environment variables
    k.Load(env.Provider("YGGDRASIL_", ".", func(s string) string {
        return strings.Replace(strings.ToLower(s), "yggdrasil_", "", 1)
    }), nil)

    var cfg Config
    if err := k.Unmarshal("", &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

---

### Web (React/TypeScript)

```bash
web/
├── src/
│   ├── main.tsx                # Application entry point
│   ├── App.tsx                 # Root component
│   ├── routes/
│   │   ├── index.tsx           # Route definitions
│   │   ├── auth/
│   │   │   ├── LoginPage.tsx
│   │   │   └── LogoutPage.tsx
│   │   ├── dashboard/
│   │   │   └── DashboardPage.tsx
│   │   ├── tickets/
│   │   │   ├── TicketsListPage.tsx
│   │   │   ├── TicketDetailPage.tsx
│   │   │   ├── TicketCreatePage.tsx
│   │   │   └── TicketEditPage.tsx
│   │   ├── clients/
│   │   │   ├── ClientsListPage.tsx
│   │   │   └── ClientDetailPage.tsx
│   │   ├── systems/
│   │   │   ├── SystemsListPage.tsx
│   │   │   └── SystemDetailPage.tsx
│   │   ├── time-entries/
│   │   │   └── TimeEntriesListPage.tsx
│   │   ├── agents/
│   │   │   └── AgentsListPage.tsx
│   │   └── users/
│   │       └── UsersListPage.tsx
│   ├── components/
│   │   ├── layout/
│   │   │   ├── Navbar.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Layout.tsx
│   │   ├── tickets/
│   │   │   ├── TicketCard.tsx
│   │   │   ├── TicketTable.tsx
│   │   │   ├── TicketForm.tsx
│   │   │   ├── TicketStatusBadge.tsx
│   │   │   └── TicketCommentList.tsx
│   │   ├── ui/                 # shadcn/ui components (copied, not npm)
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── table.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── badge.tsx
│   │   │   └── ...
│   │   └── common/
│   │       ├── LoadingSpinner.tsx
│   │       ├── ErrorMessage.tsx
│   │       └── Pagination.tsx
│   ├── lib/
│   │   ├── api.ts              # API client (axios or fetch)
│   │   ├── auth.ts             # Auth utilities (token management)
│   │   └── utils.ts            # General utilities (cn, formatDate, etc.)
│   ├── hooks/
│   │   ├── useAuth.ts          # Authentication hook
│   │   ├── useTickets.ts       # TanStack Query hooks for tickets
│   │   ├── useClients.ts
│   │   ├── useSystems.ts
│   │   └── useAgents.ts
│   ├── types/
│   │   ├── api.ts              # API response types
│   │   ├── ticket.ts           # Ticket types
│   │   ├── client.ts
│   │   ├── system.ts
│   │   └── user.ts
│   └── styles/
│       └── globals.css         # Tailwind imports
├── public/
│   └── vite.svg
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── components.json             # shadcn/ui config
└── README.md
```

**Key Files**:

`src/lib/api.ts`:

```typescript
import axios from "axios";

const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor (add auth token)
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("access_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor (handle 401, refresh token)
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Handle token refresh or redirect to login
    }
    return Promise.reject(error);
  },
);
```

`src/hooks/useTickets.ts`:

```typescript
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { Ticket, TicketCreate } from "@/types/ticket";

export function useTickets(params?: { status?: string; page?: number }) {
  return useQuery({
    queryKey: ["tickets", params],
    queryFn: async () => {
      const { data } = await api.get("/tickets", { params });
      return data;
    },
  });
}

export function useCreateTicket() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (ticket: TicketCreate) => {
      const { data } = await api.post("/tickets", ticket);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tickets"] });
    },
  });
}
```

---

### Agent (Go)

```bash
├── cmd/
│   └── agent/
│       └── main.go             # Agent entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management (viper)
│   ├── registry/
│   │   └── registry.go         # Plugin registry
│   ├── scheduler/
│   │   └── scheduler.go        # Metrics collection scheduler
│   ├── grpc/
│   │   ├── client.go           # gRPC client for control plane
│   │   └── auth.go             # mTLS authentication
│   └── nats/
│       └── client.go           # NATS client (for task queue, v1.0+)
├── plugins/
│   ├── system_metrics/
│   │   ├── plugin.go           # System metrics plugin
│   │   └── plugin_test.go
│   └── README.md               # Plugin development guide
├── pkg/
│   └── version/
│       └── version.go          # Version information
├── .goreleaser.yml             # GoReleaser config for builds
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

**Key Files**:

`cmd/agent/main.go`:

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/intrik8-labs/yggdrasil/internal/agent/internal/config"
    "github.com/intrik8-labs/yggdrasil/internal/agent/internal/registry"
    "github.com/intrik8-labs/yggdrasil/internal/agent/internal/scheduler"
    "github.com/intrik8-labs/yggdrasil/internal/agent/plugins/system_metrics"
)

func main() {
    // Setup logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        slog.Error("Failed to load config", "error", err)
        os.Exit(1)
    }

    // Create context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create plugin registry
    pluginRegistry := registry.NewRegistry()

    // Register plugins
    if err := pluginRegistry.Register(system_metrics.New()); err != nil {
        slog.Error("Failed to register plugin", "error", err)
        os.Exit(1)
    }

    // Initialize plugins
    for _, plugin := range pluginRegistry.List() {
        if err := plugin.Initialize(ctx, cfg.PluginConfig(plugin.Name())); err != nil {
            slog.Error("Failed to initialize plugin", "plugin", plugin.Name(), "error", err)
        } else {
            slog.Info("Plugin initialized", "plugin", plugin.Name(), "version", plugin.Version())
        }
    }

    // Start metrics scheduler
    metricsScheduler := scheduler.NewMetricsScheduler(cfg, pluginRegistry)
    go metricsScheduler.Start(ctx)

    // Wait for interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

    slog.Info("Shutting down agent...")
    cancel()

    // Shutdown plugins
    for _, plugin := range pluginRegistry.List() {
        if err := plugin.Shutdown(ctx); err != nil {
            slog.Error("Failed to shutdown plugin", "plugin", plugin.Name(), "error", err)
        }
    }
}
```

---

## Packages Structure

### Proto (Protocol Buffers)

```bash
packages/proto/
├── agent/
│   └── v1/
│       ├── agent.proto         # Agent service definitions
│       └── buf.gen.yaml        # Buf code generation config
├── buf.yaml                    # Buf workspace config
├── buf.lock
└── README.md
```

**Code Generation**:

```bash
# Generate Go code
buf generate --template buf.gen.go.yaml
```

---

### Agent SDK (Go)

```bash
packages/agent-sdk/
├── plugin.go                   # Plugin interfaces
├── metric.go                   # Metric types
├── task.go                     # Task types (v1.0+)
├── go.mod
└── README.md
```

---

## Tooling Configuration

### Turborepo (`turbo.json`)

```json
{
  "$schema": "https://turbo.build/schema.json",
  "tasks": {
    "lint": { "dependsOn": ["^lint"], "outputs": [] },
    "build": {
      "dependsOn": ["lint", "^build"],
      "outputs": ["dist/**", "build/**"]
    },
    "test": { "dependsOn": ["build"], "outputs": [] },
    "dev": { "cache": false }
  }
}
```

### Top-Level Makefile

```makefile
.PHONY: help setup dev test lint clean

help: ## Show this help
 @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Setup development environment
 @./scripts/setup-dev.sh

dev: ## Start development environment
 docker-compose up -d
 air  # Hot reload for API
 cd web && npm run dev &
 go run cmd/agent/main.go &

test: ## Run all tests
 go test ./...
 cd web && npm test

lint: ## Lint all code
 golangci-lint run
 cd web && npm run lint

clean: ## Clean build artifacts
 docker-compose down -v
 rm -rf coverage.out
 rm -rf web/dist
 rm -rf dist
```

---

## Environment Variables

`.env.example`:

```bash
# Control Plane
DATABASE_URL=postgresql+asyncpg://yggdrasil:password@localhost:5432/yggdrasil
SECRET_KEY=your-secret-key-here
DEBUG=true
CORS_ORIGINS=["http://localhost:3000"]

# NATS
NATS_URL=nats://localhost:4222

# gRPC
GRPC_PORT=50051

# Agent (config.yaml)
AGENT_ID=
CONTROL_PLANE_URL=localhost:50051
CERT_PATH=./certs/agent.crt
KEY_PATH=./certs/agent.key
CA_CERT_PATH=./certs/ca.crt
```

---

## Docker Compose

`docker-compose.yml`:

```yaml
version: "3.9"

services:
  postgres:
    image: timescale/timescaledb:latest-pg15
    environment:
      POSTGRES_DB: yggdrasil
      POSTGRES_USER: yggdrasil
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "8222:8222"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "8222:8222"  # HTTP monitoring

  # control-plane and web can be added here for production-like environment

volumes:
  postgres_data:
```

---

## Development Workflow

### Initial Setup

```bash
# Clone repository
git clone https://github.com/intrik8-labs/yggdrasil.git
cd yggdrasil

# Run setup script
make setup

# Start infrastructure
docker-compose up -d

# Run migrations
migrate -path migrations -database "${DATABASE_URL}" up

# Seed test data
./scripts/seed-data.sh
```

### Daily Development

```bash
# Start all services
make dev

# Or individually:
# Terminal 1: Control Plane (with hot reload)
air

# Terminal 2: Web
cd web
npm run dev

# Terminal 3: Agent (optional)
# Run from root
make run
```

### Testing

```bash
# Run all tests
make test

# Or individually
go test ./...
cd web && npm test
go test ./...
```

---

## Next Steps

1. Review [Database Migrations](../backend/database-migrations.md) for migration setup
2. Review [Testing Strategy](./08-testing-strategy.md) for test organization
3. Review [CI/CD Pipeline](./09-ci-cd-pipeline.md) for automation
4. Start building! Begin with database setup in Week 1
