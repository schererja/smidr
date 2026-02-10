# Testing Strategy

## Overview

Yggdrasil employs a comprehensive testing strategy across all components: control plane (Go), web (React/TypeScript), and agent (Go). The goal is **80%+ test coverage** for critical paths while maintaining fast, reliable tests.

---

## Testing Pyramid

```bash
        ╱╲
       ╱  ╲       E2E Tests (5-10%)
      ╱____╲      - Critical user flows
     ╱      ╲     - Slow, brittle
    ╱________╲
   ╱          ╲   Integration Tests (20-30%)
  ╱____________╲  - API endpoints, DB queries
 ╱              ╲ - Moderate speed
╱________________╲
Unit Tests (60-75%)
- Functions, classes, services
- Fast, isolated
```

**Philosophy**: Write mostly unit tests, some integration tests, few E2E tests.

---

## Control Plane Testing (Go/Chi)

### Tools

- **testing** - Go standard library test framework
- **testcontainers-go** - Docker containers for integration tests
- **httptest** - HTTP testing utilities
- **pgx/v5** - PostgreSQL driver
- **testify** - Assertions and mocks (optional)

### Test Organization

```bash
# Go tests live alongside code in internal/
internal/
├── ticket/
│   ├── ticket.go
│   ├── ticket_test.go        # Unit tests
│   ├── service.go
│   ├── service_test.go       # Service tests
│   ├── handler.go
│   └── handler_test.go       # Integration tests
├── client/
│   ├── client_test.go
│   └── service_test.go
└── agent/
    ├── agent_test.go
    └── service_test.go
```

### Test Fixtures (testcontainers)

```go
package database

import (
    "context"
    "fmt"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()

    // Start postgres container
    req := testcontainers.ContainerRequest{
        Image:        "postgres:16-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_DB":       "testdb",
            "POSTGRES_PASSWORD": "password",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp"),
    }

    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        t.Fatal(err)
    }

    // Cleanup on test completion
    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    // Get connection string
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    connStr := fmt.Sprintf("postgres://postgres:password@%s:%s/testdb",
        host, port.Port())

    // Create connection pool
    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        t.Fatal(err)
    }

    // Run migrations
    if err := runMigrations(connStr); err != nil {
        t.Fatal(err)
    }

    return pool
}
```

### API Integration Tests

```go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/intrik8-labs/yggdrasil/internal/handlers"
    "github.com/stretchr/testify/assert"
)

func TestCreateTicket(t *testing.T) {
    // Setup
    pool := SetupTestDB(t)
    defer pool.Close()

    queries := database.New(pool)
    handler := handlers.NewTicketHandler(queries)

    // Create request
    ticket := map[string]interface{}{
        "subject": "Test ticket",
        "status":  "open",
    }
    body, _ := json.Marshal(ticket)
    req := httptest.NewRequest("POST", "/api/tickets", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Execute
    handler.CreateTicket(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "Test ticket", response["subject"])
}
```

### Test Data Factories

```go
package testdata

import (
    "time"

    "github.com/google/uuid"
    "github.com/intrik8-labs/yggdrasil/internal/ticket"
    "github.com/intrik8-labs/yggdrasil/internal/client"
)

// NewTestTenant creates a tenant for testing
func NewTestTenant() *client.Tenant {
    return &client.Tenant{
        ID:        uuid.New(),
        Name:      "Test MSP",
        Slug:      "test-msp",
        IsActive:  true,
        CreatedAt: time.Now(),
    }
}

// NewTestClient creates a client for testing
func NewTestClient(tenantID uuid.UUID) *client.Client {
    return &client.Client{
        ID:       uuid.New(),
        TenantID: tenantID,
        Name:     "Acme Corp",
        Status:   "active",
    }
}

// NewTestTicket creates a ticket for testing
func NewTestTicket(tenantID, clientID uuid.UUID) *ticket.Ticket {
    return &ticket.Ticket{
        ID:        uuid.New(),
        TenantID:  tenantID,
        ClientID:  clientID,
        Subject:   "Test Issue",
        Status:    ticket.StatusOpen,
        Priority:  ticket.PriorityMedium,
        CreatedAt: time.Now(),
    }
}
```

### Table-Driven Tests

```go
func TestTicketValidation(t *testing.T) {
    tests := []struct {
        name    string
        ticket  ticket.CreateTicketRequest
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid ticket",
            ticket: ticket.CreateTicketRequest{
                Subject:  "Valid subject",
                Status:   "open",
                Priority: "medium",
            },
            wantErr: false,
        },
        {
            name: "subject too short",
            ticket: ticket.CreateTicketRequest{
                Subject:  "Hi",
                Status:   "open",
                Priority: "medium",
            },
            wantErr: true,
            errMsg:  "subject must be at least 3 characters",
        },
        {
            name: "invalid status",
            ticket: ticket.CreateTicketRequest{
                Subject:  "Valid subject",
                Status:   "invalid",
                Priority: "medium",
            },
            wantErr: true,
            errMsg:  "invalid status",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.ticket.Validate()

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

    msp_tenant_id = factory.SubFactory(MSPTenantFactory)

class UserFactory(factory.Factory):
class Meta:
model = User

    email = factory.Faker('email')
    first_name = factory.Faker('first_name')
    last_name = factory.Faker('last_name')
    role = FuzzyChoice(['msp_admin', 'msp_tech', 'msp_manager'])
    is_active = True

    msp_tenant_id = factory.SubFactory(MSPTenantFactory)

class TicketFactory(factory.Factory):
class Meta:
model = Ticket

    title = factory.Faker('sentence', nb_words=6)
    description = factory.Faker('paragraph')
    status = FuzzyChoice(['new', 'assigned', 'in_progress', 'resolved', 'closed'])
    priority = FuzzyChoice(['low', 'medium', 'high', 'urgent'])

    msp_tenant_id = factory.SubFactory(MSPTenantFactory)
    client_id = factory.SubFactory(ClientFactory)
    created_by = factory.SubFactory(UserFactory)

````

### Unit Test Example

```go
// internal/ticket/service_test.go
package ticket_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/intrik8-labs/yggdrasil/internal/ticket"
    "github.com/intrik8-labs/yggdrasil/internal/testdata"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of ticket.Repository
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, t *ticket.Ticket) error {
    args := m.Called(ctx, t)
    return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id uuid.UUID) (*ticket.Ticket, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*ticket.Ticket), args.Error(1)
}

func TestCreateTicket(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockRepo := new(MockRepository)
    service := ticket.NewService(mockRepo)

    tenantID := uuid.New()
    cmd := ticket.CreateTicketCommand{
        TenantID:    tenantID,
        Subject:     "Test ticket",
        Description: "Test description",
        Priority:    ticket.PriorityHigh,
    }

    // Mock expects Create to be called once
    mockRepo.On("Create", ctx, mock.AnythingOfType("*ticket.Ticket")).
        Return(nil)

    // Act
    created, err := service.CreateTicket(ctx, cmd)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, created)
    assert.Equal(t, "Test ticket", created.Subject)
    assert.Equal(t, ticket.StatusOpen, created.Status)
    assert.Equal(t, ticket.PriorityHigh, created.Priority)
    mockRepo.AssertExpectations(t)
}

func TestListTickets_FiltersByTenant(t *testing.T) {
    // Arrange
    ctx := context.Background()
    pool := testdata.SetupTestDB(t)
    defer pool.Close()

    queries := database.New(pool)
    repo := ticket.NewPostgresRepository(queries)
    service := ticket.NewService(repo)

    // Create test data
    tenant1 := uuid.New()
    tenant2 := uuid.New()

    ticket1 := testdata.NewTestTicket(tenant1, uuid.New())
    ticket2 := testdata.NewTestTicket(tenant2, uuid.New())

    repo.Create(ctx, ticket1)
    repo.Create(ctx, ticket2)

    // Act
    tickets, err := service.ListTickets(ctx, tenant1, ticket.ListFilters{})

    // Assert
    assert.NoError(t, err)
    assert.Len(t, tickets, 1)
    assert.Equal(t, tenant1, tickets[0].TenantID)
}

    db_session.add_all([tenant1, tenant2])
    await db_session.commit()

    # Create tickets for both tenants
    ticket1 = TicketFactory.build(msp_tenant_id=tenant1.id)
    ticket2 = TicketFactory.build(msp_tenant_id=tenant2.id)

    db_session.add_all([ticket1, ticket2])
    await db_session.commit()

    service = TicketService(db_session)

    # Act
    tenant1_tickets = await service.list_tickets(tenant1.id)

    # Assert
    assert len(tenant1_tickets) == 1
    assert tenant1_tickets[0].id == ticket1.id
````

### Integration Test Example (API)

```go
// internal/handlers/ticket_handler_test.go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/intrik8-labs/yggdrasil/internal/handlers"
    "github.com/stretchr/testify/assert"
)

func TestCreateTicketEndpoint(t *testing.T) {
    // Arrange
    pool := testdata.SetupTestDB(t)
    defer pool.Close()

    queries := database.New(pool)
    handler := handlers.NewTicketHandler(queries)

    payload := map[string]interface{}{
        "client_id":   testdata.NewTestClient().ID.String(),
        "subject":     "Server down",
        "description": "Production server not responding",
        "priority":    "urgent",
    }
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest("POST", "/api/v1/tickets", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+testdata.GetTestToken())
    w := httptest.NewRecorder()

    // Act
    handler.CreateTicket(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    data := response["data"].(map[string]interface{})
    assert.Equal(t, "Server down", data["subject"])
    assert.Equal(t, "open", data["status"])
    assert.Equal(t, "urgent", data["priority"])
}

func TestListTicketsPagination(t *testing.T) {
    // Arrange
    pool := testdata.SetupTestDB(t)
    defer pool.Close()

    queries := database.New(pool)
    handler := handlers.NewTicketHandler(queries)

    // Create 25 test tickets
    tenantID := uuid.New()
    for i := 0; i < 25; i++ {
        ticket := testdata.NewTestTicket(tenantID, uuid.New())
        queries.CreateTicket(context.Background(), ticket)
    }

    // Act - request page 2 with 10 per page
    req := httptest.NewRequest("GET", "/api/v1/tickets?page=2&per_page=10", nil)
    req.Header.Set("Authorization", "Bearer "+testdata.GetTestToken())
    w := httptest.NewRecorder()

    handler.ListTickets(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    data := response["data"].([]interface{})
    meta := response["meta"].(map[string]interface{})

    assert.Len(t, data, 10)
    assert.Equal(t, float64(2), meta["page"])
    assert.Equal(t, float64(25), meta["total"])
    assert.Equal(t, float64(3), meta["total_pages"])
}

func TestMultiTenantIsolation(t *testing.T) {
    // Arrange
    pool := testdata.SetupTestDB(t)
    defer pool.Close()

    queries := database.New(pool)
    handler := handlers.NewTicketHandler(queries)

    tenant1 := uuid.New()
    tenant2 := uuid.New()

    // Tenant 1 creates a ticket
    ticket := testdata.NewTestTicket(tenant1, uuid.New())
    queries.CreateTicket(context.Background(), ticket)

    // Act - Tenant 2 tries to access Tenant 1's ticket
    req := httptest.NewRequest("GET", "/api/v1/tickets/"+ticket.ID.String(), nil)
    req.Header.Set("Authorization", "Bearer "+testdata.GetTestTokenForTenant(tenant2))
    w := httptest.NewRecorder()

    handler.GetTicket(w, req)

    // Assert - Should return 404 (due to RLS filtering)
    assert.Equal(t, http.StatusNotFound, w.Code)
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/ticket

# Run tests matching pattern
go test ./... -run TestTicket

# Run with verbose output
go test ./... -v

# Run integration tests
go test -tags=integration ./...

# Run with race detector
go test -race ./...
```

---

## Web Testing (React/TypeScript)

### Tools - Web Testing

- **Vitest** - Test framework (Vite-native)
- **@testing-library/react** - React component testing
- **@testing-library/user-event** - User interaction simulation
- **MSW (Mock Service Worker)** - API mocking

### Test Organization - Web Testing

```bash
web/tests/
├── setup.ts                 # Test setup
├── mocks/
│   ├── handlers.ts          # MSW handlers
│   └── server.ts            # MSW server setup
└── unit/
    ├── components/
    │   ├── TicketCard.test.tsx
    │   └── TicketForm.test.tsx
    ├── hooks/
    │   └── useTickets.test.ts
    └── utils/
        └── formatters.test.ts
```

### Component Test Example

```typescript
// tests/unit/components/TicketCard.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { TicketCard } from '@/components/tickets/TicketCard';

describe('TicketCard', () => {
  it('renders ticket information correctly', () => {
    const ticket = {
      id: '123',
      ticket_number: 1042,
      title: 'Server down',
      status: 'open',
      priority: 'urgent',
      client: { name: 'Acme Corp' },
      created_at: '2026-01-22T10:00:00Z',
    };

    render(<TicketCard ticket={ticket} />);

    expect(screen.getByText('#1042')).toBeInTheDocument();
    expect(screen.getByText('Server down')).toBeInTheDocument();
    expect(screen.getByText('Acme Corp')).toBeInTheDocument();
    expect(screen.getByText('Urgent')).toBeInTheDocument();
  });

  it('applies correct status badge styling', () => {
    const ticket = { ...baseTicket, status: 'in_progress' };

    render(<TicketCard ticket={ticket} />);

    const badge = screen.getByText('In Progress');
    expect(badge).toHaveClass('bg-blue-100');
  });
});
```

### Hook Test Example (with MSW)

```typescript
// tests/unit/hooks/useTickets.test.ts
import { describe, it, expect } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useTickets } from '@/hooks/useTickets';

describe('useTickets', () => {
  it('fetches tickets successfully', async () => {
    const queryClient = new QueryClient();
    const wrapper = ({ children }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    const { result } = renderHook(() => useTickets(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(result.current.data[0].title).toBe('Test Ticket 1');
  });
});
```

### MSW Setup (`tests/mocks/handlers.ts`)

```typescript
import { http, HttpResponse } from "msw";

export const handlers = [
  // List tickets
  http.get("/api/v1/tickets", () => {
    return HttpResponse.json({
      data: [
        { id: "1", title: "Test Ticket 1", status: "open" },
        { id: "2", title: "Test Ticket 2", status: "closed" },
      ],
      meta: { page: 1, per_page: 20, total: 2 },
    });
  }),

  // Create ticket
  http.post("/api/v1/tickets", async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json(
      {
        data: { id: "3", ...body, status: "new" },
      },
      { status: 201 },
    );
  }),
];
```

### Running Tests - Web Testing

```bash
cd web

# Run all tests
npm test

# Run with coverage
npm test -- --coverage

# Run in watch mode
npm test -- --watch

# Run specific test file
npm test -- TicketCard.test.tsx
```

---

## Agent Testing (Go)

### Tools - Agent Testing

- **testing** - Standard library
- **github.com/stretchr/testify** - Assertions and mocks

### Test Organization - Agent Testing

```bash
internal/agent/
├── plugins/
│   └── system_metrics/
│       ├── plugin.go
│       └── plugin_test.go
├── registry/
│   ├── registry.go
│   └── registry_test.go
└── scheduler/
    ├── scheduler.go
    └── scheduler_test.go
```

### Plugin Test Example

```go
// internal/agent/plugins/system_metrics/plugin_test.go
package system_metrics

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSystemMetricsPlugin(t *testing.T) {
    t.Run("Initialize", func(t *testing.T) {
        plugin := New()
        config := map[string]interface{}{
            "interval_seconds": 60.0,
        }

        err := plugin.Initialize(context.Background(), config)
        require.NoError(t, err)
        assert.Equal(t, 60*time.Second, plugin.interval)
    })

    t.Run("CollectMetrics", func(t *testing.T) {
        plugin := New()
        _ = plugin.Initialize(context.Background(), map[string]interface{}{})

        metrics, err := plugin.CollectMetrics(context.Background())
        require.NoError(t, err)
        assert.NotEmpty(t, metrics)

        // Verify CPU metric exists
        var cpuMetric *sdk.Metric
        for i := range metrics {
            if metrics[i].MetricName == "cpu_usage" {
                cpuMetric = &metrics[i]
                break
            }
        }
        require.NotNil(t, cpuMetric, "CPU metric should be collected")
        assert.GreaterOrEqual(t, cpuMetric.Value, 0.0)
        assert.LessOrEqual(t, cpuMetric.Value, 100.0)
    })

    t.Run("Shutdown", func(t *testing.T) {
        plugin := New()
        _ = plugin.Initialize(context.Background(), map[string]interface{}{})

        err := plugin.Shutdown(context.Background())
        assert.NoError(t, err)
    })
}
```

### Running Tests - Agent Testing

```bash
# Run all tests (from root)
go test ./...

# Run with coverage
go test ./... -cover

# Run with verbose output
go test ./... -v

# Run specific package
go test ./internal/agent/plugins/system_metrics

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## E2E Testing (Manual for v0.5)

For v0.5, E2E tests are **manual** using scripted test cases. Automated E2E with Playwright can be added in v1.0.

### Manual E2E Test Cases

#### **Test Case 1: Complete Ticket Workflow**

1. Login as MSP Admin
2. Create new client "Test Corp"
3. Create new ticket for Test Corp
4. Assign ticket to MSP Tech
5. Add time entry to ticket
6. Add comment to ticket
7. Resolve ticket
8. Verify ticket visible in client portal (read-only)

#### **Test Case 2: Agent Registration and Metrics**

1. Install agent on test system
2. Agent registers with control plane
3. MSP Admin approves agent
4. Agent sends metrics
5. Verify metrics appear in system detail page
6. Verify metrics stored in TimescaleDB

#### **Test Case 3: Multi-Tenant Isolation**

1. Create Tenant A and Tenant B
2. Login as Tenant A user
3. Create ticket for Tenant A client
4. Logout, login as Tenant B user
5. Verify Tenant B cannot see Tenant A's ticket
6. Verify API returns 404 for cross-tenant access

---

## Coverage Targets

| Component      | Target | Measurement                         |
| -------------- | ------ | ----------------------------------- |
| Control Plane  | 80%+   | Statement coverage (pytest-cov)     |
| Web            | 70%+   | Statement coverage (Vitest)         |
| Agent          | 80%+   | Statement coverage (go test -cover) |
| Critical Paths | 100%   | Auth, multi-tenancy, billing (v1.0) |
| Web            | 70%+   | Statement coverage (Vitest)         |
| Agent          | 80%+   | Statement coverage (go test -cover) |
| Critical Paths | 100%   | Auth, multi-tenancy, billing (v1.0) |

---

## CI Integration

All tests run automatically on pull requests. See [CI/CD Pipeline](./09-ci-cd-pipeline.md).

---

## Next Steps

1. Review [CI/CD Pipeline](./09-ci-cd-pipeline.md) for test automation
2. Set up test databases for local development
3. Write tests alongside features (TDD encouraged)
4. Monitor coverage reports in CI
