# Yggdrasil Style Guide

This guide establishes coding standards and conventions for the Yggdrasil MSP/CRM/ERP platform to ensure consistency, maintainability, and quality across all codebases.

## Table of Contents

1. [General Principles](#general-principles)
2. [Go Standards (Control Plane & Agent)](#go-standards-control-plane--agent)
3. [TypeScript/JavaScript Standards (Web App)](#typescriptjavascript-standards-web-app)
4. [Database Standards](#database-standards)
5. [API Standards](#api-standards)
6. [Documentation Standards](#documentation-standards)
7. [Testing Standards](#testing-standards)

## General Principles

### Code Quality

- **Readability over cleverness**: Write code that others can understand easily
- **Consistency**: Follow established patterns within each codebase
- **Simplicity**: Prefer simple solutions that solve the problem effectively
- **Testability**: Design code to be easily testable
- **Performance**: Consider performance implications but optimize based on metrics

### Naming Conventions

- **Be descriptive**: Use names that clearly indicate purpose
- **Be consistent**: Use the same naming patterns across the codebase
- **Avoid abbreviations**: Unless widely understood (e.g., `id`, `url`)
- **Use domain language**: Match terminology with business requirements

## Go Standards (Control Plane & Agent)

### Project Organization

Use **package-oriented design** with domain-based vertical slices:

```
internal/
├── ticket/              # Each domain is self-contained
│   ├── ticket.go        # Entity + business rules
│   ├── repository.go    # Repository interface
│   ├── service.go       # Use cases/orchestration
│   ├── handler.go       # HTTP handlers
│   ├── postgres.go      # Database implementation
│   ├── queries.sql      # sqlc queries
│   └── *_test.go        # Tests alongside code
├── client/
├── agent/
├── auth/
├── platform/            # Shared infrastructure
│   ├── config/
│   ├── database/
│   ├── messaging/
│   └── http/
└── shared/              # Shared utilities
```

### Code Style

Use standard Go formatting tools:

```bash
# Format code
gofmt -w .
goimports -w .

# Lint
golangci-lint run

# Vet
go vet ./...
```

### Naming Conventions

```go
// Variables and functions: camelCase (unexported) or PascalCase (exported)
userName := "john_doe"
func GetUserByID(userID string) (*User, error) {}

// Constants: PascalCase or SCREAMING_SNAKE_CASE for groups
const MaxRetries = 3
const (
    StatusActive   = "active"
    StatusInactive = "inactive"
)

// Interfaces: -er suffix for single-method, descriptive for multi-method
type Reader interface {}
type TicketRepository interface {}

// Types: PascalCase
type User struct {}
type TicketStatus string

// Package names: lowercase, single word
package ticket
package auth
```

### Domain Structure

Each domain package should contain:

```go
// Entity (ticket.go) - Business rules and domain logic
package ticket

type Ticket struct {
    ID          string
    TenantID    string
    ClientID    string
    Title       string
    Description string
    Status      Status
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (t *Ticket) Validate() error {
    if t.Title == "" {
        return errors.New("title is required")
    }
    if t.TenantID == "" {
        return errors.New("tenant_id is required")
    }
    return nil
}

// Repository Interface (repository.go)
type Repository interface {
    Create(ctx context.Context, ticket *Ticket) error
    GetByID(ctx context.Context, tenantID, id string) (*Ticket, error)
    Update(ctx context.Context, ticket *Ticket) error
    Delete(ctx context.Context, tenantID, id string) error
    List(ctx context.Context, tenantID string, filters Filters) ([]*Ticket, error)
}

// Service (service.go) - Use cases and orchestration
type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateTicket(ctx context.Context, ticket *Ticket) error {
    if err := ticket.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return s.repo.Create(ctx, ticket)
}

// Handler (handler.go) - HTTP handlers
type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Post("/", h.CreateTicket)
    r.Get("/{id}", h.GetTicket)
    r.Put("/{id}", h.UpdateTicket)
    r.Delete("/{id}", h.DeleteTicket)
    return r
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
    var req CreateTicketRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    ticket := &Ticket{
        Title:       req.Title,
        Description: req.Description,
        TenantID:    r.Context().Value("tenant_id").(string),
    }

    if err := h.svc.CreateTicket(r.Context(), ticket); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ticket)
}
```

### Error Handling

```go
// Use sentinel errors for expected conditions
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrInvalidInput  = errors.New("invalid input")
)

// Wrap errors with context
func (s *Service) GetTicket(ctx context.Context, id string) (*Ticket, error) {
    ticket, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get ticket: %w", err)
    }
    return ticket, nil
}

// Check for specific errors
if errors.Is(err, ErrNotFound) {
    http.Error(w, "not found", http.StatusNotFound)
    return
}
```

### Configuration with koanf

```go
package config

import (
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
)

type Config struct {
    Server   ServerConfig   `koanf:"server"`
    Database DatabaseConfig `koanf:"database"`
}

type ServerConfig struct {
    Address string `koanf:"address"`
    Port    int    `koanf:"port"`
}

func Load() (*Config, error) {
    k := koanf.New(".")

    // Load from file
    if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
        return nil, err
    }

    // Override with env vars
    k.Load(env.Provider("YGGDRASIL_", ".", func(s string) string {
        return strings.Replace(strings.ToLower(s), "yggdrasil_", "", 1)
    }), nil)

    var cfg Config
    if err := k.Unmarshal("", &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### Logging with slog

```go
import "log/slog"

// Structured logging
slog.Info("ticket created",
    "ticket_id", ticket.ID,
    "tenant_id", ticket.TenantID,
    "user_id", userID)

slog.Error("failed to create ticket",
    "error", err,
    "tenant_id", tenantID)

// With context
logger := slog.With("tenant_id", tenantID)
logger.Info("processing request")
```

## TypeScript/JavaScript Standards (Web App)

### Code Style

Use ESLint and Prettier for consistent formatting:

```json
// .eslintrc.js
module.exports = {
  extends: [
    '@typescript-eslint/recommended',
    'prettier'
  ],
  rules: {
    '@typescript-eslint/no-unused-vars': 'error',
    '@typescript-eslint/explicit-function-return-type': 'warn'
  }
};
```

### Naming Conventions

```typescript
// Variables and functions: camelCase
const userName = "john_doe"

const getUserById = (id: number): User => {
    // implementation
}

// Interfaces/Types: PascalCase
interface User {
    id:   number
    name: string
    email: string
}

type UserRole = "admin" | "user" | "readonly"

// Components: PascalCase
const UserProfile: React.FC<{ userId: number }> = ({ userId }) => {
  // component implementation
};

// Constants: UPPER_SNAKE_CASE for exports
export const API_BASE_URL = "/api/v1";
export const MAX_RETRIES = 3;

// File naming: kebab-case
// user-profile.component.tsx
// user-service.ts
// types/user.types.ts
```

### React Components

```typescript
// Functional components with hooks
interface UserProfileProps {
  userId: number;
  onUpdate?: (user: User) => void;
}

const UserProfile: React.FC<UserProfileProps> = ({
  userId,
  onUpdate
}) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        setLoading(true);
        const userData = await userService.getUser(userId);
        setUser(userData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [userId]);

  if (loading) return <Spinner />;
  if (error) return <ErrorMessage message={error} />;
  if (!user) return <div>User not found</div>;

  return (
    <div className="user-profile">
      <h2>{user.name}</h2>
      <p>{user.email}</p>
    </div>
  );
};
```

### API Services

```typescript
// services/userService.ts
import { apiClient } from '../lib/api';
import { User, CreateUserRequest, UpdateUserRequest } from '../types/user.types';

class UserService {
  private baseUrl = '/api/v1/users';

  async getUsers(): Promise<User[]> {
    const response = await apiClient.get<User[]>(this.baseUrl);
    return response.data;
  }

  async getUserById(id: number): Promise<User> {
    const response = await apiClient.get<User>(`${this.baseUrl}/${id}`);
    return response.data;
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    const response = await apiClient.post<User>(this.baseUrl, data);
    return response.data;
  }

  async updateUser(id: number, data: UpdateUserRequest): Promise<User> {
    const response = await apiClient.put<User>(`${this.baseUrl}/${id}`, data);
    return response.data;
  }

  async deleteUser(id: number): Promise<void> {
    await apiClient.delete(`${this.baseUrl}/${id}`);
  }
}

export const userService = new UserService();
```

### State Management

```typescript
// hooks/useUsers.ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { userService } from "../services/userService";
import { User } from "../types/user.types";

export const useUsers = () => {
  return useQuery({
    queryKey: ["users"],
    queryFn: () => userService.getUsers(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useCreateUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: userService.createUser,
    onSuccess: (newUser) => {
      queryClient.setQueryData(["users"], (old: User[] | undefined) =>
        old ? [...old, newUser] : [newUser],
      );
    },
    onError: (error) => {
      console.error("Failed to create user:", error);
    },
  });
};
```

## Go Standards (Agent)

The Agent follows the same Go standards as the Control Plane. See the [Go Standards (Control Plane & Agent)](#go-standards-control-plane--agent) section above for detailed coding standards.

### Agent-Specific Patterns

```go
// Agent-specific configuration
type AgentConfig struct {
    ServerURL    string        `koanf:"server_url"`
    APIKey       string        `koanf:"api_key"`
    TenantID     string        `koanf:"tenant_id"`
    SystemID     string        `koanf:"system_id"`
    Interval     time.Duration `koanf:"interval"`
}

// Metrics collection interface
type MetricsCollector interface {
    Collect(ctx context.Context) (*SystemMetrics, error)
    Name() string
    Interval() time.Duration
}

// Agent lifecycle management
type Agent struct {
    config    *AgentConfig
    collectors []MetricsCollector
    client    *http.Client
    logger    *slog.Logger
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}
```

## Database Standards

### Schema Design

```sql
-- Table naming: plural, snake_case
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Foreign key naming: {table}_id
CREATE TABLE tickets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'open',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_tickets_user_id ON tickets(user_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_created_at ON tickets(created_at);
```

### Migration Standards

```sql
-- golang-migrate migration example
-- 000001_create_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL DEFAULT 'technician',
    is_active BOOLEAN DEFAULT true,
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tenant ON users(msp_tenant_id);
CREATE INDEX idx_users_role ON users(role);
```

## API Standards

### REST API Design

```go
// Resource naming: plural nouns
GET    /api/v1/users          // List users
POST   /api/v1/users          // Create user
GET    /api/v1/users/{id}     // Get specific user
PUT    /api/v1/users/{id}     // Update user
DELETE /api/v1/users/{id}     // Delete user

// Nested resources
GET    /api/v1/users/{id}/tickets     // Get user's tickets
POST   /api/v1/users/{id}/tickets     // Create ticket for user

// Query parameters for filtering and pagination
GET /api/v1/tickets?status=open&assigned_to=123&page=2&limit=50

// Example Go handler
func (h *UserHandler) GetUsers(c echo.Context) error {
    params := &ListUsersParams{
        Status:     c.QueryParam("status"),
        AssignedTo: c.QueryParam("assigned_to"),
        Page:       c.QueryParam("page"),
        Limit:      c.QueryParam("limit"),
    }

    users, err := h.userService.List(c.Request().Context(), params)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": err.Error(),
        })
    }

    return c.JSON(http.StatusOK, users)
}
```

### Response Format

```json
{
  "data": [
    {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "meta": {
    "total": 150,
    "page": 2,
    "limit": 50,
    "total_pages": 3
  },
  "links": {
    "self": "/api/v1/users?page=2&limit=50",
    "first": "/api/v1/users?page=1&limit=50",
    "last": "/api/v1/users?page=3&limit=50",
    "next": "/api/v1/users?page=3&limit=50",
    "prev": "/api/v1/users?page=1&limit=50"
  }
}
```

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "email": ["Invalid email format"],
      "name": ["Name is required"]
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "path": "/api/v1/users"
}
```

## Documentation Standards

### Code Documentation

```go
// CalculateTicketPriority calculates ticket priority based on urgency and impact
//
// Args:
//   urgency: Urgency level (1-5, where 5 is highest)
//   impact: Impact level (1-5, where 5 is highest)
//
// Returns:
//   Priority score (1-10, where 10 is highest priority)
//   Error if urgency or impact are outside valid range
//
// Example:
//   CalculateTicketPriority(4, 3) // returns 7
func CalculateTicketPriority(urgency, impact int) (int, error) {
    if urgency < 1 || urgency > 5 {
        return 0, errors.New("urgency must be between 1 and 5")
    }
    if impact < 1 || impact > 5 {
        return 0, errors.New("impact must be between 1 and 5")
    }

    return (urgency + impact) // Simple priority calculation
}
```

### API Documentation

```go
// user_handler.go - API documentation with struct tags
package user

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

// CreateUserRequest represents the request body for creating a new user
type CreateUserRequest struct {
    // Email is the user's email address (required)
    Email string `json:"email" validate:"required,email" example:"user@example.com"`

    // Name is the user's full name (required)
    Name string `json:"name" validate:"required,min=1,max=255" example:"John Doe"`

    // CompanyID is optional company ID for organization users
    CompanyID *int `json:"company_id,omitempty" example:"123"`
}

// @Summary Create a new user
// @Description Create a new user account with provided details. The user will receive a welcome email upon successful creation.
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation request"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Implementation
    user, err := h.service.Create(r.Context(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

## Testing Standards

### Go Unit Testing

```go
// ticket_test.go - Entity tests
package ticket

import (
    "testing"
)

func TestTicket_Validate(t *testing.T) {
    tests := []struct {
        name    string
        ticket  *Ticket
        wantErr bool
    }{
        {
            name: "valid ticket",
            ticket: &Ticket{
                Title:    "Test Ticket",
                TenantID: "tenant-123",
                ClientID: "client-456",
            },
            wantErr: false,
        },
        {
            name: "missing title",
            ticket: &Ticket{
                TenantID: "tenant-123",
                ClientID: "client-456",
            },
            wantErr: true,
        },
        {
            name: "missing tenant_id",
            ticket: &Ticket{
                Title:    "Test Ticket",
                ClientID: "client-456",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.ticket.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

// service_test.go - Service tests with mocks
package ticket

import (
    "context"
    "errors"
    "testing"
)

type mockRepository struct {
    createFunc func(ctx context.Context, ticket *Ticket) error
}

func (m *mockRepository) Create(ctx context.Context, ticket *Ticket) error {
    if m.createFunc != nil {
        return m.createFunc(ctx, ticket)
    }
    return nil
}

func (m *mockRepository) GetByID(ctx context.Context, tenantID, id string) (*Ticket, error) {
    return nil, nil
}

func TestService_CreateTicket(t *testing.T) {
    ctx := context.Background()

    t.Run("success", func(t *testing.T) {
        repo := &mockRepository{}
        svc := NewService(repo)

        ticket := &Ticket{
            Title:    "Test",
            TenantID: "tenant-1",
            ClientID: "client-1",
        }

        err := svc.CreateTicket(ctx, ticket)
        if err != nil {
            t.Errorf("expected no error, got %v", err)
        }
    })

    t.Run("validation error", func(t *testing.T) {
        repo := &mockRepository{}
        svc := NewService(repo)

        ticket := &Ticket{} // invalid

        err := svc.CreateTicket(ctx, ticket)
        if err == nil {
            t.Error("expected validation error, got nil")
        }
    })

    t.Run("repository error", func(t *testing.T) {
        repo := &mockRepository{
            createFunc: func(ctx context.Context, ticket *Ticket) error {
                return errors.New("database error")
            },
        }
        svc := NewService(repo)

        ticket := &Ticket{
            Title:    "Test",
            TenantID: "tenant-1",
            ClientID: "client-1",
        }

        err := svc.CreateTicket(ctx, ticket)
        if err == nil {
            t.Error("expected error, got nil")
        }
    })
}
```

### Go Integration Testing (with testcontainers)

```go
// handler_test.go - Integration tests
package ticket

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
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

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("failed to start container: %v", err)
    }

    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    // Get connection string
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    connStr := fmt.Sprintf("postgres://postgres:password@%s:%s/testdb", host, port.Port())

    // Connect to database
    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        t.Fatalf("failed to connect: %v", err)
    }

    // Run migrations
    // ... migration code ...

    return pool
}

func TestHandler_CreateTicket(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    repo := NewPostgresRepository(db)
    svc := NewService(repo)
    handler := NewHandler(svc)

    // Create request
    reqBody := CreateTicketRequest{
        Title:       "Test Ticket",
        Description: "Test Description",
        ClientID:    "client-123",
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
    req = req.WithContext(context.WithValue(req.Context(), "tenant_id", "tenant-123"))
    w := httptest.NewRecorder()

    // Execute
    handler.CreateTicket(w, req)

    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }

    var ticket Ticket
    json.NewDecoder(w.Body).Decode(&ticket)
    if ticket.Title != reqBody.Title {
        t.Errorf("expected title %s, got %s", reqBody.Title, ticket.Title)
    }
}
```

### TypeScript Integration Testing

```typescript
// TypeScript integration test example
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { UserList } from './UserList';
import { userService } from '../services/userService';

// Mock the service
jest.mock('../services/userService');
const mockUserService = userService as jest.Mocked<typeof userService>;

const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
});

describe('UserList', () => {
  it('displays users when data is loaded', async () => {
    // Arrange
    const mockUsers = [
      { id: 1, name: 'John Doe', email: 'john@example.com' },
      { id: 2, name: 'Jane Smith', email: 'jane@example.com' },
    ];
    mockUserService.getUsers.mockResolvedValue(mockUsers);

    const queryClient = createTestQueryClient();

    // Act
    render(
      <QueryClientProvider client={queryClient}>
        <UserList />
      </QueryClientProvider>
    );

    // Assert
    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('jane@example.com')).toBeInTheDocument();
    });
  });

  it('shows loading state initially', () => {
    // Arrange
    mockUserService.getUsers.mockReturnValue(new Promise(() => {}));
    const queryClient = createTestQueryClient();

    // Act
    render(
      <QueryClientProvider client={queryClient}>
        <UserList />
      </QueryClientProvider>
    );

    // Assert
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
  });
});
```

### Go Testing (Agent)

```go
// Agent testing example
package agent

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock collector for testing
type MockCollector struct {
    mock.Mock
}

func (m *MockCollector) Collect(ctx context.Context) (*SystemMetrics, error) {
    args := m.Called(ctx)
    return args.Get(0).(*SystemMetrics), args.Error(1)
}

func (m *MockCollector) Name() string {
    args := m.Called()
    return args.String(0)
}

func (m *MockCollector) Interval() time.Duration {
    args := m.Called()
    return args.Get(0).(time.Duration)
}

func TestAgent_Start(t *testing.T) {
    // Arrange
    mockCollector := &MockCollector{}
    mockCollector.On("Name").Return("test")
    mockCollector.On("Interval").Return(time.Millisecond)
    mockCollector.On("Collect", mock.Anything).Return(&SystemMetrics{}, nil)

    agent := NewAgent(&AgentConfig{})
    agent.AddCollector(mockCollector)

    // Act
    err := agent.Start()
    time.Sleep(10 * time.Millisecond) // Allow goroutines to start
    agent.Stop()

    // Assert
    assert.NoError(t, err)
    mockCollector.AssertExpectations(t)
}

func TestSystemCollector_Collect(t *testing.T) {
    // Arrange
    config := &AgentConfig{
        Interval: time.Second,
    }
    collector := NewSystemCollector(config, nil)

    // Act
    metrics, err := collector.Collect(context.Background())

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, metrics)
    assert.False(metrics.Timestamp.IsZero())
}
```

This style guide serves as the foundation for maintaining code quality across the Yggdrasil platform. Regular reviews and updates ensure it stays relevant as the project evolves.
