# Development Workflow

## Overview

This document establishes development workflows, coding standards, and operational procedures for Yggdrasil. It combines proven practices from enterprise development with modern tooling to ensure consistency, quality, and maintainability.

---

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (control plane and agent)
- Node.js 20+ (web application)
- Git and GitHub account

### Initial Setup

```bash
# Clone repository
git clone <repository-url>
cd yggdrasil

# Install dependencies
cd . && go mod download
cd ../web && npm install
cd ../agent && go mod download

# Start development environment
docker-compose up -d
```

### Development Services

Running services after setup:

- **PostgreSQL 15**: `localhost:5432`
  - User: `postgres`, Password: `yggdrasil_dev`
  - Database: `yggdrasil`
- **NATS**: `localhost:4222`
- **Control Plane API**: `http://localhost:8080`
- **Web Application**: `http://localhost:3000`

---

## 📊 Monorepo Structure

```bash
yggdrasil/
├── cmd/                     # Application entry points
│   ├── control-plane/        # Go backend (Týr, Mímir, etc.)
│   ├── web/                # React frontend (Valhalla, Bifröst)
│   └── agent/              # Go agent (Smidr)
├── packages/               # Shared code
│   ├── proto/             # gRPC protocol definitions
│   ├── agent-sdk/         # Agent SDK libraries
│   └── shared/           # Common types and utilities
├── docs/                 # Architecture and documentation
└── scripts/              # Build and deployment scripts
```

### Development Commands

```bash
# Control Plane (Go/Chi)
cd .
go run cmd/api/main.go

# Web Application (React/Vite)
cd web
npm run dev

# Agent (Go)
cd internal/agent
go run cmd/agent/main.go

# All services (using Make)
make dev              # Start all development servers
make test              # Run all tests
make lint              # Run all linters
make build             # Build all applications
```

---

## 🛠️ Development Workflow

### Daily Development

1. **Start Services**

   ```bash
   docker-compose up -d    # Database and NATS
   make dev               # All applications
   ```

2. **Development Branches**

   ```bash
   git checkout main
   git pull
   git checkout -b feature/your-feature-name
   ```

3. **Code Quality Checks**

   ```bash
   make check              # Run linting and type checking
   make test               # Run all tests
   ```

4. **Commit Changes**

   ```bash
   git add .
   git commit -m "feat(tickets): add ticket filtering by status"
   git push origin feature/your-feature-name
   ```

### Code Quality Standards

#### Go (Control Plane)

```bash
# Formatting
gofmt -w ./
goimports -w ./

# Linting
golangci-lint run ./...

# Vetting
go vet ./...

# Testing
cd . && go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### TypeScript (Web Application)

```bash
# Linting
cd web
npm run lint
npm run lint:fix

# Type checking
npm run typecheck

# Formatting
npm run format

# Testing
npm run test
npm run test:coverage
```

#### Go (Agent)

```bash
# Formatting
go fmt ./internal/agent/...

# Linting
golangci-lint run ./internal/agent/...

# Testing
go test ./internal/agent/... -v -race -cover
```

---

## 🔧 Environment Configuration

### Required Environment Variables

Create `.env` file from `.env.example`:

```bash
# Database
DATABASE_URL=postgresql://postgres:yggdrasil_dev@localhost:5432/yggdrasil

# Authentication
JWT_SECRET_KEY=your-super-secret-jwt-key
REFRESH_SECRET_KEY=your-super-secret-refresh-key

# NATS
NATS_URL=nats://localhost:4222

# Agent (if running locally)
AGENT_SERVER_URL=localhost:8080
AGENT_GRPC_PORT=50051
```

### Environment-Specific Configurations

- **Development**: Use `.env.local` for local overrides
- **Testing**: Use test database and in-memory NATS
- **Production**: Use managed services and secret managers

---

## 📋 Service Development Patterns

### Control Plane Service Layer (Go)

Each service follows domain-driven design:

```go
// internal/ticket/service.go
package ticket

import (
    "context"
    "fmt"
    "github.com/google/uuid"
)

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateTicket(ctx context.Context, tenantID uuid.UUID, subject, description string, priority Priority) (*Ticket, error) {
    // Business logic validation
    ticket, err := New(tenantID, subject, description, priority)
    if err != nil {
        return nil, err
    }

    if err := s.repo.Create(ctx, ticket); err != nil {
        return nil, fmt.Errorf("create ticket: %w", err)
    }

    return ticket, nil
}

func (s *Service) GetTickets(ctx context.Context, filter Filter) ([]*Ticket, error) {
        """Get filtered tickets with proper tenant isolation."""
        return await self.db.execute(
            select(Ticket).where(Ticket.tenant_id == tenant_id, **filters)
        )
```

### Web Component Patterns (React/TypeScript)

```typescript
// web/src/components/tickets/TicketList.tsx
import { useQuery } from '@tanstack/react-query';
import { TicketService } from '@/services/tickets';

interface TicketListProps {
  clientId?: string;
  status?: string;
}

export function TicketList({ clientId, status }: TicketListProps) {
  const { data: tickets, isLoading, error } = useQuery({
    queryKey: ['tickets', { clientId, status }],
    queryFn: () => TicketService.getTickets({ clientId, status }),
  });

  if (isLoading) return <div>Loading tickets...</div>;
  if (error) return <div>Error loading tickets</div>;

  return (
    <div className="space-y-4">
      {tickets?.map(ticket => (
        <TicketCard key={ticket.id} ticket={ticket} />
      ))}
    </div>
  );
}
```

### Agent Plugin Patterns (Go)

```go
// internal/agent/internal/plugins/metrics/system.go
package metrics

import (
    "context"
    "github.com/intrik8-labs/yggdrasil/packages/agent-sdk"
)

type SystemMetricsPlugin struct {
    sdk agent_sdk.SDK
}

func (p *SystemMetricsPlugin) Initialize(sdk agent_sdk.SDK) error {
    p.sdk = sdk
    return nil
}

func (p *SystemMetricsPlugin) Collect(ctx context.Context) ([]agent_sdk.Metric, error) {
    // Collect system metrics
    cpuUsage := getCpuUsage()
    memUsage := getMemoryUsage()

    return []agent_sdk.Metric{
        {Name: "cpu_usage", Value: cpuUsage, Labels: nil},
        {Name: "memory_usage", Value: memUsage, Labels: nil},
    }, nil
}

func (p *SystemMetricsPlugin) Name() string {
    return "system_metrics"
}
```

---

## 🔄 Database Operations

### Migrations (Alembic)

```bash
# Create new migration
cd .
migrate create -ext sql -dir migrations -seq add_ticket_comments

# Apply migrations
migrate -path migrations -database "${DATABASE_URL}" up

# Rollback migration
migrate -path migrations -database "${DATABASE_URL}" down 1

# View migration version
migrate -path migrations -database "${DATABASE_URL}" version
```

### Development Data

```bash
# Seed development data
cd .
go run cmd/seed/main.go

# Reset database (development only)
docker-compose down -v postgres
docker-compose up -d postgres
migrate -path migrations -database "${DATABASE_URL}" up
go run cmd/seed/main.go
```

---

## 🧪 Testing Strategy

### Test Structure

```bash
cmd/
├── control-plane/
│   └── tests/
│       ├── unit/           # Service layer tests
│       ├── integration/    # API endpoint tests
│       └── conftest.py    # Test fixtures and configuration
├── web/
│   └── src/
│       └── __tests__/      # Component tests
└── agent/
    └── internal/          # Go unit tests
        └── plugins/
            └── system_test.go
```

### Running Tests

```bash
# All tests
make test

# Specific service
cd . && go test ./internal/ticket/... -v
cd web && npm test -- --testNamePattern="TicketList"
cd internal/agent && go test ./internal/plugins/... -v

# Coverage
make test-coverage
```

### Test Patterns

**Unit Tests (Go)**:

```go
package ticket_test

    import (
        "context"
        "testing"
        "yggdrasil/internal/ticket"
        "github.com/google/uuid"
    )

@pytest.mark.asyncio
func TestCreateTicket(t *testing.T) {
    # Arrange
    service := ticket.NewService(mock_db)
    ticketData := &ticket.CreateTicketCommand{
        Title:       "Test Ticket",
        Description: "Test",
    }
    # Act
    result, err := service.CreateTicket(ctx, ticketData, "tenant-123")
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Test Ticket", result.Title)
    assert.Equal(t, "tenant-123", result.TenantID)
```

---

## 🚨 Common Issues & Solutions

### Database Connection Issues

```bash
# Check PostgreSQL status
docker-compose ps postgres

# Restart PostgreSQL
docker-compose restart postgres

# View logs
docker-compose logs postgres
```

### Port Conflicts

If ports are in use, modify `docker-compose.yml`:

```yaml
services:
  postgres:
    ports:
      - "5433:5432" # Change host port
```

### Agent Registration Failures

1. Verify control plane is running: `curl http://localhost:8080/health`
2. Check NATS connection: `docker-compose logs nats`
3. Review agent logs: `cd internal/agent && go run cmd/agent/main.go -v`

---

## 📝 Git Workflow

### Commit Message Convention

```bash
type(scope): description

feat(tickets): add ticket filtering by status
fix(agent): handle heartbeat timeout gracefully
docs(api): update ticket endpoints documentation
refactor(auth): simplify JWT token validation
test(backend): add integration tests for ticket service
```

### Branch Strategy

- `main` - Production-ready code
- `develop` - Integration branch
- `feature/` - New features
- `fix/` - Bug fixes
- `hotfix/` - Critical production fixes

### Pull Request Process

1. Create PR from feature branch to `develop`
2. Ensure all CI checks pass
3. Request code review
4. Address feedback
5. Merge after approval

---

## 📊 Monitoring & Debugging

### Application Logs

```bash
# Control Plane logs
cd . && go run cmd/api/main.go 2>&1 | jq

# Web application logs
cd web && npm run dev 2>&1 | tee web.log

# Agent logs
cd internal/agent && go run cmd/agent/main.go -v
```

### Database Debugging

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U postgres yggdrasil

# View recent queries
SELECT query, calls, total_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;
```

### API Testing

```bash
# Health check
curl http://localhost:8080/health

# API documentation
open http://localhost:8080/docs

# Example API call
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@test.com", "password": "test123"}'
```

---

## 🚀 Deployment Preparation

### Build Process

```bash
# Build all applications
make build

# Build specific service
make build-control-plane
make build-web
make build-agent

# Production containers
docker-compose -f docker-compose.prod.yml build
```

### Pre-Deployment Checklist

- [ ] All tests passing (`make test`)
- [ ] Code quality checks passing (`make check`)
- [ ] Security scan completed
- [ ] Database migrations tested
- [ ] Environment variables documented
- [ ] API documentation updated
- [ ] Performance benchmarks run

---

## 🔍 Performance Tips

### Development Performance

- Use hot reload features in all applications
- Run database queries in development mode for debugging
- Enable fast builds (Turbo for TypeScript, Go's native compiler)

### Application Performance

- Optimize database queries with proper indexing
- Use connection pooling for database connections
- Implement caching for frequently accessed data
- Monitor memory usage in long-running agents

---

## 📚 Additional Resources

### Documentation

- [System Architecture](../platform/system-architecture.md)
- [Data Model](./02-data-model.md)
- [API Contract](./03-api-contract.md)
- [Implementation Roadmap](../platform/implementation-roadmap.md)

### Tools & Libraries

- **Go Chi**: [Chi Documentation](https://github.com/go-chi/chi)
- **sqlc**: [sqlc Documentation](https://sqlc.dev/)
- **React**: [React Documentation](https://react.dev/)
- **Go**: [Go Documentation](https://golang.org/doc/)
- **PostgreSQL**: [PostgreSQL Docs](https://www.postgresql.org/docs/)

---

This development workflow ensures consistency across the entire Yggdrasil platform while enabling efficient development and deployment practices.
