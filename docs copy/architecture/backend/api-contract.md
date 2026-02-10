# API Contract Design

## Overview

Yggdrasil exposes two primary APIs:

1. **REST API** - For web application (MSP users and client portal)
2. **gRPC API** - For agent communication (metrics, heartbeats, task execution)

This document defines v0.5 API contracts for both.

---

## REST API (Go/Chi)

### Base URL

```bash
https://api.yggdrasil.io/api/v1
```

### Authentication Method

**Method**: JWT Bearer Token

**Headers**:

```bash
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Token Payload**:

```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "role": "msp_admin",
  "exp": 1706000000
}
```

### Standard Response Formats

**Success Response**:

```json
{
  "data": { ... },
  "meta": {
    "timestamp": "2026-01-22T10:30:00Z"
  }
}
```

**Error Response**:

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Ticket not found",
    "details": {}
  }
}
```

**Paginated Response**:

```json
{
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### Error Codes

| HTTP Status | Code                  | Description                         |
| ----------- | --------------------- | ----------------------------------- |
| 400         | VALIDATION_ERROR      | Request validation failed           |
| 401         | UNAUTHORIZED          | Authentication required             |
| 403         | FORBIDDEN             | Insufficient permissions            |
| 404         | RESOURCE_NOT_FOUND    | Resource does not exist             |
| 409         | CONFLICT              | Resource conflict (e.g., duplicate) |
| 422         | UNPROCESSABLE_ENTITY  | Semantic validation error           |
| 500         | INTERNAL_SERVER_ERROR | Server error                        |
| 503         | SERVICE_UNAVAILABLE   | Service temporarily unavailable     |

---

## REST API Endpoints (v0.5)

### Authentication

#### POST /auth/login

Authenticate user and receive JWT tokens.

**Request**:

```json
{
  "email": "admin@acme-msp.com",
  "password": "securepassword"
}
```

**Response (200)**:

```json
{
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
    "token_type": "bearer",
    "expires_in": 900
  }
}
```

#### POST /auth/refresh

Refresh access token using refresh token.

**Request**:

```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**Response (200)**: Same as login

#### POST /auth/logout

Invalidate refresh token.

**Response (204)**: No content

---

### Tickets

#### GET /tickets

List tickets with filtering and pagination.

**Query Parameters**:

- `page` (int, default: 1)
- `per_page` (int, default: 20, max: 100)
- `status` (string, optional): Filter by status
- `priority` (string, optional): Filter by priority
- `assigned_to` (uuid, optional): Filter by assignee
- `client_id` (uuid, optional): Filter by client
- `sort` (string, default: "-created_at"): Sort field (prefix `-` for desc)

**Response (200)**:

```json
{
  "data": [
    {
      "id": "ticket-uuid",
      "ticket_number": 1042,
      "title": "Server down - web01",
      "description": "Production web server is not responding",
      "status": "in_progress",
      "priority": "urgent",
      "client": {
        "id": "client-uuid",
        "name": "Acme Corp"
      },
      "system": {
        "id": "system-uuid",
        "hostname": "web01.acme.com"
      },
      "assigned_to": {
        "id": "user-uuid",
        "first_name": "John",
        "last_name": "Doe"
      },
      "created_by": {
        "id": "user-uuid",
        "first_name": "Jane",
        "last_name": "Smith"
      },
      "created_at": "2026-01-22T08:30:00Z",
      "updated_at": "2026-01-22T10:15:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### GET /tickets/{ticket_id}

Get single ticket with full details.

**Response (200)**:

```json
{
  "data": {
    "id": "ticket-uuid",
    "ticket_number": 1042,
    "title": "Server down - web01",
    "description": "Production web server is not responding...",
    "status": "in_progress",
    "priority": "urgent",
    "client": { ... },
    "system": { ... },
    "assigned_to": { ... },
    "created_by": { ... },
    "time_entries": [
      {
        "id": "time-entry-uuid",
        "duration_minutes": 45,
        "description": "Investigated issue, restarted services",
        "billable": true,
        "user": { ... },
        "created_at": "2026-01-22T09:00:00Z"
      }
    ],
    "comments": [
      {
        "id": "comment-uuid",
        "content": "Server is back online",
        "is_internal": false,
        "user": { ... },
        "created_at": "2026-01-22T10:00:00Z"
      }
    ],
    "created_at": "2026-01-22T08:30:00Z",
    "updated_at": "2026-01-22T10:15:00Z"
  }
}
```

#### POST /tickets

Create new ticket.

**Request**:

```json
{
  "client_id": "client-uuid",
  "system_id": "system-uuid", // optional
  "title": "Server down - web01",
  "description": "Production web server is not responding",
  "priority": "urgent"
}
```

**Response (201)**:

```json
{
  "data": {
    "id": "ticket-uuid",
    "ticket_number": 1043,
    "title": "Server down - web01",
    "status": "new",
    "priority": "urgent",
    ...
  }
}
```

#### PATCH /tickets/{ticket_id}

Update ticket (partial update).

**Request**:

```json
{
  "status": "in_progress",
  "assigned_to": "user-uuid"
}
```

**Response (200)**: Updated ticket object

#### POST /tickets/{ticket_id}/comments

Add comment to ticket.

**Request**:

```json
{
  "content": "Server is back online after restart",
  "is_internal": false
}
```

**Response (201)**:

```json
{
  "data": {
    "id": "comment-uuid",
    "content": "Server is back online after restart",
    "is_internal": false,
    "user": { ... },
    "created_at": "2026-01-22T10:30:00Z"
  }
}
```

---

### Time Entries

#### GET /time-entries

List time entries with filtering.

**Query Parameters**:

- `page`, `per_page` (pagination)
- `ticket_id` (uuid, optional): Filter by ticket
- `user_id` (uuid, optional): Filter by user
- `billable` (bool, optional): Filter by billable status
- `start_date`, `end_date` (date, optional): Date range filter

**Response (200)**:

```json
{
  "data": [
    {
      "id": "time-entry-uuid",
      "ticket": {
        "id": "ticket-uuid",
        "ticket_number": 1042,
        "title": "Server down - web01"
      },
      "user": {
        "id": "user-uuid",
        "first_name": "John",
        "last_name": "Doe"
      },
      "duration_minutes": 45,
      "description": "Investigated issue, restarted services",
      "billable": true,
      "hourly_rate": 150.00,
      "started_at": "2026-01-22T09:00:00Z",
      "ended_at": "2026-01-22T09:45:00Z",
      "created_at": "2026-01-22T09:50:00Z"
    }
  ],
  "meta": { ... }
}
```

#### POST /time-entries

Create time entry.

**Request**:

```json
{
  "ticket_id": "ticket-uuid",
  "duration_minutes": 45,
  "description": "Investigated server issue",
  "billable": true,
  "started_at": "2026-01-22T09:00:00Z",
  "ended_at": "2026-01-22T09:45:00Z"
}
```

**Response (201)**: Created time entry object

#### PATCH /time-entries/{entry_id}

Update time entry.

**Request**:

```json
{
  "duration_minutes": 60,
  "billable": false
}
```

**Response (200)**: Updated time entry object

#### DELETE /time-entries/{entry_id}

Delete time entry.

**Response (204)**: No content

---

### Clients

#### GET /clients

List clients.

**Query Parameters**:

- `page`, `per_page`
- `status` (string, optional): Filter by status
- `sub_org_id` (uuid, optional): Filter by sub-org
- `search` (string, optional): Search by name

**Response (200)**:

```json
{
  "data": [
    {
      "id": "client-uuid",
      "name": "Acme Corp",
      "status": "active",
      "sub_org": {
        "id": "sub-org-uuid",
        "name": "West Coast Region"
      },
      "contact_name": "Bob Johnson",
      "contact_email": "bob@acme.com",
      "systems_count": 15,
      "open_tickets_count": 3,
      "created_at": "2025-06-01T00:00:00Z"
    }
  ],
  "meta": { ... }
}
```

#### GET /clients/{client_id}

Get single client with details.

**Response (200)**:

```json
{
  "data": {
    "id": "client-uuid",
    "name": "Acme Corp",
    "status": "active",
    "sub_org": { ... },
    "contact_name": "Bob Johnson",
    "contact_email": "bob@acme.com",
    "contact_phone": "+1-555-0100",
    "address": "123 Main St, San Francisco, CA",
    "notes": "VIP client, 24/7 support",
    "systems": [
      {
        "id": "system-uuid",
        "hostname": "web01.acme.com",
        "os_type": "linux",
        "last_seen": "2026-01-22T10:30:00Z"
      }
    ],
    "created_at": "2025-06-01T00:00:00Z",
    "updated_at": "2026-01-20T15:00:00Z"
  }
}
```

#### POST /clients

Create new client.

**Request**:

```json
{
  "name": "Acme Corp",
  "sub_org_id": "sub-org-uuid", // optional
  "contact_name": "Bob Johnson",
  "contact_email": "bob@acme.com",
  "contact_phone": "+1-555-0100",
  "address": "123 Main St, San Francisco, CA"
}
```

**Response (201)**: Created client object

#### PATCH /clients/{client_id}

Update client.

**Request**:

```json
{
  "status": "inactive",
  "notes": "Contract ended"
}
```

**Response (200)**: Updated client object

---

### Systems

#### GET /systems

List systems.

**Query Parameters**:

- `page`, `per_page`
- `client_id` (uuid, optional): Filter by client
- `os_type` (string, optional): Filter by OS
- `status` (string, optional): online, offline, unknown
- `search` (string, optional): Search by hostname

**Response (200)**:

```json
{
  "data": [
    {
      "id": "system-uuid",
      "hostname": "web01.acme.com",
      "ip_address": "192.168.1.100",
      "os_type": "linux",
      "os_version": "Ubuntu 22.04",
      "agent_version": "0.5.0",
      "client": {
        "id": "client-uuid",
        "name": "Acme Corp"
      },
      "status": "online",
      "last_seen": "2026-01-22T10:30:00Z",
      "created_at": "2025-08-15T00:00:00Z"
    }
  ],
  "meta": { ... }
}
```

#### GET /systems/{system_id}

Get single system with details and recent metrics.

**Response (200)**:

```json
{
  "data": {
    "id": "system-uuid",
    "hostname": "web01.acme.com",
    "ip_address": "192.168.1.100",
    "os_type": "linux",
    "os_version": "Ubuntu 22.04",
    "agent_version": "0.5.0",
    "client": { ... },
    "agent": {
      "id": "agent-uuid",
      "status": "approved",
      "last_heartbeat": "2026-01-22T10:30:00Z"
    },
    "recent_metrics": {
      "cpu_usage": 45.2,
      "memory_usage": 68.5,
      "disk_usage": 72.3
    },
    "metadata": {
      "cpu_cores": 8,
      "total_memory_gb": 32
    },
    "created_at": "2025-08-15T00:00:00Z",
    "updated_at": "2026-01-22T10:30:00Z"
  }
}
```

#### GET /systems/{system_id}/metrics

Get historical metrics for a system.

**Query Parameters**:

- `metric_name` (string, required): cpu_usage, memory_usage, disk_usage, etc.
- `start_time` (ISO8601, required): Start of time range
- `end_time` (ISO8601, required): End of time range
- `interval` (string, optional): Aggregation interval (1m, 5m, 1h, default: raw)

**Response (200)**:

```json
{
  "data": {
    "metric_name": "cpu_usage",
    "interval": "5m",
    "data_points": [
      {
        "time": "2026-01-22T10:00:00Z",
        "value": 42.5,
        "labels": {}
      },
      {
        "time": "2026-01-22T10:05:00Z",
        "value": 45.8,
        "labels": {}
      }
    ]
  }
}
```

#### POST /systems

Create new system (manual registration).

**Request**:

```json
{
  "client_id": "client-uuid",
  "hostname": "web02.acme.com",
  "ip_address": "192.168.1.101",
  "os_type": "linux",
  "os_version": "Ubuntu 22.04"
}
```

**Response (201)**: Created system object

#### PATCH /systems/{system_id}

Update system.

**Request**:

```json
{
  "metadata": {
    "cpu_cores": 16,
    "total_memory_gb": 64
  }
}
```

**Response (200)**: Updated system object

---

### Agents

#### GET /agents

List agents.

**Query Parameters**:

- `page`, `per_page`
- `status` (string, optional): pending, approved, revoked
- `client_id` (uuid, optional): Filter by client

**Response (200)**:

```json
{
  "data": [
    {
      "id": "agent-uuid",
      "system": {
        "id": "system-uuid",
        "hostname": "web01.acme.com",
        "client": {
          "id": "client-uuid",
          "name": "Acme Corp"
        }
      },
      "status": "approved",
      "last_heartbeat": "2026-01-22T10:30:00Z",
      "approved_at": "2025-08-15T12:00:00Z",
      "approved_by": {
        "id": "user-uuid",
        "first_name": "Admin",
        "last_name": "User"
      },
      "created_at": "2025-08-15T11:45:00Z"
    }
  ],
  "meta": { ... }
}
```

#### POST /agents/{agent_id}/approve

Approve pending agent.

**Response (200)**:

```json
{
  "data": {
    "id": "agent-uuid",
    "status": "approved",
    "certificate": "-----BEGIN CERTIFICATE-----\n...",
    "approved_at": "2026-01-22T10:35:00Z"
  }
}
```

#### POST /agents/{agent_id}/revoke

Revoke agent access.

**Response (200)**:

```json
{
  "data": {
    "id": "agent-uuid",
    "status": "revoked"
  }
}
```

---

### Users

#### GET /users

List users (MSP admin only).

**Query Parameters**:

- `page`, `per_page`
- `role` (string, optional): Filter by role
- `is_active` (bool, optional): Filter by active status

**Response (200)**:

```json
{
  "data": [
    {
      "id": "user-uuid",
      "email": "john.doe@acme-msp.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "msp_tech",
      "is_active": true,
      "last_login": "2026-01-22T08:00:00Z",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "meta": { ... }
}
```

#### POST /users

Create new user (MSP admin only).

**Request**:

```json
{
  "email": "jane.smith@acme-msp.com",
  "password": "securepassword",
  "first_name": "Jane",
  "last_name": "Smith",
  "role": "msp_tech"
}
```

**Response (201)**: Created user object (without password)

#### PATCH /users/{user_id}

Update user.

**Request**:

```json
{
  "role": "msp_manager",
  "is_active": false
}
```

**Response (200)**: Updated user object

---

### Sub-Orgs

#### GET /sub-orgs

List sub-organizations.

**Response (200)**:

```json
{
  "data": [
    {
      "id": "sub-org-uuid",
      "name": "West Coast Region",
      "parent": null,
      "clients_count": 25,
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

#### POST /sub-orgs

Create sub-organization.

**Request**:

```json
{
  "name": "West Coast Region",
  "description": "California and Oregon offices",
  "parent_id": null // optional, for nested orgs
}
```

**Response (201)**: Created sub-org object

---

## gRPC API (Agent Communication)

### Protocol Buffers Definition

**File**: `packages/proto/agent/v1/agent.proto`

```protobuf
syntax = "proto3";

package agent.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";

// Agent service definition
service AgentService {
  // Agent registration (initial setup)
  rpc Register(RegisterRequest) returns (RegisterResponse);

  // Heartbeat (keep-alive)
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);

  // Send metrics (batch)
  rpc SendMetrics(MetricBatch) returns (MetricBatchResponse);

  // Get pending tasks (v1.0+)
  rpc GetTasks(GetTasksRequest) returns (TaskList);

  // Report task result (v1.0+)
  rpc ReportTaskResult(TaskResult) returns (TaskResultResponse);
}

// Registration
message RegisterRequest {
  string hostname = 1;
  string ip_address = 2;
  string os_type = 3;
  string os_version = 4;
  string agent_version = 5;
  string mac_address = 6;  // For identification
  google.protobuf.Struct metadata = 7;
}

message RegisterResponse {
  string agent_id = 1;
  string system_id = 2;
  string status = 3;  // "pending" or "approved"
  string message = 4;
}

// Heartbeat
message HeartbeatRequest {
  string agent_id = 1;
}

message HeartbeatResponse {
  google.protobuf.Timestamp server_time = 1;
  bool config_updated = 2;  // Signal agent to fetch new config
}

// Metrics
message Metric {
  google.protobuf.Timestamp time = 1;
  string metric_name = 2;
  double value = 3;
  map<string, string> labels = 4;
}

message MetricBatch {
  string agent_id = 1;
  repeated Metric metrics = 2;
}

message MetricBatchResponse {
  int32 accepted = 1;
  int32 rejected = 2;
  repeated string errors = 3;
}

// Tasks (v1.0+)
message GetTasksRequest {
  string agent_id = 1;
}

message Task {
  string task_id = 1;
  string task_type = 2;
  google.protobuf.Struct payload = 3;
  google.protobuf.Timestamp created_at = 4;
}

message TaskList {
  repeated Task tasks = 1;
}

message TaskResult {
  string task_id = 1;
  string status = 2;  // "success" or "failed"
  google.protobuf.Struct result = 3;
  string error = 4;
  google.protobuf.Timestamp completed_at = 5;
}

message TaskResultResponse {
  bool acknowledged = 1;
}
```

### gRPC Service Implementation (Go)

```go
package grpc

import (
    "context"
    "google.golang.org/grpc"
    "github.com/intrik8-labs/yggdrasil/pkg/grpc/agent/v1"
    "github.com/intrik8-labs/yggdrasil/internal/agent/service"
)

type AgentServer struct {
    pb.UnimplementedAgentServiceServer
    agentService service.Service
}

func NewAgentServer(agentService service.Service) *AgentServer {
    return &AgentServer{
        agentService: agentService,
    }
}

func (s *AgentServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
    // Extract tenant from mTLS certificate
    tenantID := s.getTenantFromCert(ctx)

    agent, err := s.agentService.RegisterAgent(ctx, &service.RegisterAgentRequest{
        TenantID:     tenantID,
        Hostname:     req.Hostname,
        IPAddress:    req.IpAddress,
        OSType:       req.OsType,
        OSVersion:    req.OsVersion,
        AgentVersion: req.AgentVersion,
        MacAddress:   req.MacAddress,
        Metadata:     req.Metadata.AsMap(),
    })
    if err != nil {
        return nil, err
    }

    return &pb.RegisterResponse{
        AgentId:  agent.ID.String(),
        SystemId: agent.SystemID.String(),
        Status:    agent.Status,
        Message:   "Agent registered successfully",
    }, nil
}

func (s *AgentServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
    tenantID := s.getTenantFromCert(ctx)

    err := s.agentService.RecordHeartbeat(ctx, &service.RecordHeartbeatRequest{
        AgentID:  req.AgentId,
        TenantID:  tenantID,
    })
    if err != nil {
        return nil, err
    }

    return &pb.HeartbeatResponse{
        ServerTime:   timestamppb.New(time.Now()),
        ConfigUpdated: false,
    }, nil
}

func (s *AgentServer) SendMetrics(ctx context.Context, req *pb.MetricBatch) (*pb.MetricBatchResponse, error) {
    tenantID := s.getTenantFromCert(ctx)

    metrics := make([]*pb.Metric, 0, len(req.Metrics))
    for i, m := range req.Metrics {
        metrics[i] = &pb.Metric{
            Time:       timestamppb.New(m.Time.AsTime()),
            MetricName: m.MetricName,
            Value:      m.Value,
            Labels:     m.Labels.AsMap(),
        }
    }

    result, err := s.agentService.IngestMetrics(ctx, &service.IngestMetricsRequest{
        AgentID:  req.AgentId,
        TenantID:  tenantID,
        Metrics:   metrics,
    })
    if err != nil {
        return nil, err
    }

    return &pb.MetricBatchResponse{
        Accepted: result.Accepted,
        Rejected: result.Rejected,
        Errors:   result.Errors,
    }, nil
}
```

### gRPC Client (Go Agent)

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    pb "github.com/intrik8-labs/yggdrasil/packages/proto/agent/v1"
)

type Agent struct {
    client pb.AgentServiceClient
    agentID string
}

func (a *Agent) Register(ctx context.Context) error {
    req := &pb.RegisterRequest{
        Hostname: getHostname(),
        IpAddress: getIPAddress(),
        OsType: "linux",
        OsVersion: "Ubuntu 22.04",
        AgentVersion: "0.5.0",
        MacAddress: getMacAddress(),
    }

    resp, err := a.client.Register(ctx, req)
    if err != nil {
        return err
    }

    a.agentID = resp.AgentId
    log.Printf("Registered with agent_id: %s, status: %s", resp.AgentId, resp.Status)
    return nil
}

func (a *Agent) SendHeartbeat(ctx context.Context) error {
    req := &pb.HeartbeatRequest{
        AgentId: a.agentID,
    }

    _, err := a.client.Heartbeat(ctx, req)
    return err
}

func (a *Agent) SendMetrics(ctx context.Context, metrics []*pb.Metric) error {
    req := &pb.MetricBatch{
        AgentId: a.agentID,
        Metrics: metrics,
    }

    resp, err := a.client.SendMetrics(ctx, req)
    if err != nil {
        return err
    }

    log.Printf("Metrics sent: %d accepted, %d rejected", resp.Accepted, resp.Rejected)
    return nil
}
```

---

## API Versioning Strategy

### URL-Based Versioning

All APIs include version in URL path:

```bash
/api/v1/...
/api/v2/...
```

### Version Lifecycle

1. **Current Version (v1)**: Fully supported, receives all new features
2. **Previous Version (v0)**: Maintained for 6 months after new version release
3. **Deprecated Version**: 3-month notice before removal

### Breaking Changes

Breaking changes require new major version:

- Removing fields
- Changing field types
- Removing endpoints
- Changing authentication

Non-breaking changes allowed in same version:

- Adding optional fields
- Adding new endpoints
- Adding query parameters

---

## Rate Limiting

### Limits (v0.5)

| Endpoint Type      | Limit              | Window |
| ------------------ | ------------------ | ------ |
| Authentication     | 10 requests        | 1 min  |
| REST API (general) | 1000 requests      | 1 min  |
| Metrics ingestion  | 10,000 data points | 1 min  |
| gRPC (general)     | 100 requests       | 1 min  |
| gRPC (general)     | 100 requests       | 1 min  |

### Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1706000000
```

### Rate Limit Response (429)

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again in 30 seconds.",
    "retry_after": 30
  }
}
```

---

## OpenAPI/Swagger Documentation

Go with Chi can use swaggo to generate OpenAPI documentation:

- **Interactive Docs**: `https://api.yggdrasil.io/docs` (Swagger UI)
- **ReDoc**: `https://api.yggdrasil.io/redoc`
- **OpenAPI JSON**: `https://api.yggdrasil.io/openapi.json`

```go
// @title Yggdrasil API
// @version 1.0
// @description Yggdrasil MSP/CRM/ERP Platform API
// @host api.yggdrasil.io
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
    // Chi router setup with swaggo middleware
}
```

---

## Next Steps

1. Review [Plugin Architecture](./04-plugin-architecture.md) for extensibility
2. Review [Implementation Roadmap](../platform/implementation-roadmap.md) for build plan
3. Review [Testing Strategy](./08-testing-strategy.md) for API testing approach
