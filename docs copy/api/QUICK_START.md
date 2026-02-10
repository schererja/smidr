# API Quick Start Guide

This comprehensive guide helps developers integrate with the Yggdrasil MSP/CRM/ERP platform API, covering authentication, basic operations, and advanced workflows across multiple programming languages.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [API Overview](#api-overview)
4. [Base URLs](#base-urls)
5. [Rate Limiting](#rate-limiting)
6. [Error Handling](#error-handling)
7. [SDK Setup](#sdk-setup)
8. [Core API Operations](#core-api-operations)
9. [Advanced Workflows](#advanced-workflows)
10. [Webhooks](#webhooks)
11. [Examples by Language](#examples-by-language)
12. [Testing](#testing)
13. [Best Practices](#best-practices)

## Getting Started

### Prerequisites

- Yggdrasil account with API access
- API credentials (client ID and secret)
- HTTP client or programming language SDK
- Basic understanding of RESTful APIs

### Quick Start (5 Minutes)

**For a complete overview of authentication methods, see the [Authentication](#authentication) section.**

```bash
# 1. Get access token
ACCESS_TOKEN=$(curl -s -X POST https://api.yourdomain.com/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "your_client_id",
    "client_secret": "your_client_secret",
    "grant_type": "client_credentials"
  }' | jq -r '.access_token')

# 2. Make authenticated request
curl -X GET https://api.yourdomain.com/api/v1/tickets \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

## Authentication

Yggdrasil uses OAuth 2.0 for API authentication with multiple grant types.

### Authentication Methods

#### 1. Client Credentials (Service-to-Service)

```bash
curl -X POST https://api.yourdomain.com/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "your_client_id",
    "client_secret": "your_client_secret",
    "grant_type": "client_credentials"
  }'
```

#### 2. Resource Owner Password (User Login)

```bash
curl -X POST https://api.yourdomain.com/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "your_client_id",
    "client_secret": "your_client_secret",
    "grant_type": "password",
    "username": "user@company.com",
    "password": "user_password"
  }'
```

#### 3. Authorization Code (Web Apps)

```bash
# Redirect user to authorization
GET https://api.yourdomain.com/api/v1/auth/authorize?
  response_type=code&
  client_id=your_client_id&
  redirect_uri=https://yourapp.com/callback&
  scope=read write&
  state=random_string

# Exchange code for token
curl -X POST https://api.yourdomain.com/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "your_client_id",
    "client_secret": "your_client_secret",
    "grant_type": "authorization_code",
    "code": "auth_code_from_callback",
    "redirect_uri": "https://yourapp.com/callback"
  }'
```

### Token Response

```json
{
  "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIs...",
  "refresh_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read write"
}
```

### Using Tokens

```bash
# Include Authorization header
curl -X GET https://api.yourdomain.com/api/v1/users \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIs..."
```

### Refresh Token

```bash
curl -X POST https://api.yourdomain.com/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "your_client_id",
    "client_secret": "your_client_secret",
    "grant_type": "refresh_token",
    "refresh_token": "your_refresh_token"
  }'
```

## API Overview

### RESTful Design

- **Base URL**: `https://api.yourdomain.com/api/v1`
- **Content-Type**: `application/json`
- **Authentication**: Bearer token
- **HTTPS Required**: All API calls must use HTTPS

### HTTP Methods

- **GET**: Retrieve resources
- **POST**: Create new resources
- **PUT**: Update existing resources
- **PATCH**: Partial updates
- **DELETE**: Remove resources

### Response Format

All API responses follow consistent JSON format:

```json
{
  "data": [
    {
      "id": 1,
      "name": "Example Resource",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "meta": {
    "total": 1,
    "page": 1,
    "limit": 50,
    "total_pages": 1
  },
  "links": {
    "self": "/api/v1/resources?page=1&limit=50",
    "first": "/api/v1/resources?page=1&limit=50",
    "last": "/api/v1/resources?page=1&limit=50",
    "next": null,
    "prev": null
  }
}
```

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "email": ["Invalid email format"],
      "name": ["Name is required"]
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "path": "/api/v1/users"
}
```

## Base URLs

### Environments

| Environment | Base URL                                    | Purpose                 |
| ----------- | ------------------------------------------- | ----------------------- |
| Production  | `https://api.yourdomain.com/api/v1`         | Live production API     |
| Staging     | `https://staging-api.yourdomain.com/api/v1` | Pre-production testing  |
| Development | `https://dev-api.yourdomain.com/api/v1`     | Development environment |

### API Endpoints Overview

```bash
/api/v1/
├── auth/                    # Authentication endpoints
│   ├── token               # OAuth token endpoint
│   ├── authorize           # OAuth authorization endpoint
│   └── refresh            # Token refresh endpoint
├── users/                   # User management
├── tickets/                 # Ticket management
├── customers/               # Customer management
├── systems/                 # System/device management
├── agents/                  # Agent management
├── reports/                 # Reports and analytics
├── webhooks/                # Webhook management
└── health/                  # Health check endpoints
```

## Rate Limiting

### Limits

- **Standard Rate**: 1,000 requests per hour per API key
- **Burst Limit**: 100 requests per minute
- **Concurrent Connections**: 10 per client

### Rate Limit Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1704067200
```

### Rate Limiting Response

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again later."
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "retry_after": 3600
}
```

## Error Handling

### HTTP Status Codes

| Status | Meaning               | Description                       |
| ------ | --------------------- | --------------------------------- |
| 200    | OK                    | Request successful                |
| 201    | Created               | Resource created successfully     |
| 400    | Bad Request           | Invalid request data              |
| 401    | Unauthorized          | Invalid or missing authentication |
| 403    | Forbidden             | Insufficient permissions          |
| 404    | Not Found             | Resource not found                |
| 409    | Conflict              | Resource already exists           |
| 422    | Unprocessable Entity  | Validation errors                 |
| 429    | Too Many Requests     | Rate limit exceeded               |
| 500    | Internal Server Error | Server error                      |

### Common Error Codes

| Code                       | HTTP Status | Description                  |
| -------------------------- | ----------- | ---------------------------- |
| `VALIDATION_ERROR`         | 400         | Request validation failed    |
| `AUTHENTICATION_FAILED`    | 401         | Invalid credentials          |
| `INSUFFICIENT_PERMISSIONS` | 403         | Missing required permissions |
| `RESOURCE_NOT_FOUND`       | 404         | Requested resource not found |
| `RESOURCE_CONFLICT`        | 409         | Resource already exists      |
| `RATE_LIMIT_EXCEEDED`      | 429         | Too many requests            |
| `INTERNAL_ERROR`           | 500         | Internal server error        |

### Error Handling Best Practices

```go
// Go error handling example
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

func main() {
    client := &http.Client{Timeout: 30 * time.Second}
    token := "your-access-token"

    req, err := http.NewRequest("GET", "https://api.yourdomain.com/api/v1/tickets", nil)
    if err != nil {
        panic(err)
    }
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case 401:
        // Refresh token and retry
        token, err = refreshAccessToken()
        if err != nil {
            panic(err)
        }
        req.Header.Set("Authorization", "Bearer "+token)
        resp, err = client.Do(req)
    case 429:
        // Wait and retry
        retryAfter := resp.Header.Get("Retry-After")
        if retryAfter != "" {
            time.Sleep(parseRetryAfter(retryAfter))
        }
    }
        response = requests.get(
            'https://api.yourdomain.com/api/v1/tickets',
            headers={'Authorization': f'Bearer {token}'}
        )

    response.raise_for_status()
    return response.json()

except requests.exceptions.RequestException as e:
    print(f"API request failed: {e}")
    return None
```

## SDK Setup

### Go SDK

```bash
# Install
go get github.com/intrik8-labs/yggdrasil/pkg/sdk

# Import
import "github.com/intrik8-labs/yggdrasil/pkg/sdk"

client = YggdrasilClient(
    client_id='your_client_id',
    client_secret='your_client_secret',
    base_url='https://api.yourdomain.com/api/v1'
)

try:
    token = client.auth.get_token()
    print(f"Access token: {token.access_token}")
except Exception as e:
    print(f"Authentication failed: {e}")
```

#### JavaScript

```javascript
import { YggdrasilClient } from "@yggdrasil/sdk";

const client = new YggdrasilClient({
  clientId: "your_client_id",
  clientSecret: "your_client_secret",
  baseUrl: "https://api.yourdomain.com/api/v1",
});

try {
  const token = await client.auth.getToken();
  console.log("Access token:", token.accessToken);
} catch (error) {
  console.error("Authentication failed:", error.message);
}
```

#### JavaScript

```javascript
import { YggdrasilClient } from "@yggdrasil/sdk";

const client = new YggdrasilClient({
  clientId: "your_client_id",
  clientSecret: "your_client_secret",
  baseUrl: "https://api.yourdomain.com/api/v1",
});

try {
  const token = await client.auth.getToken();
  console.log("Access token:", token.accessToken);
} catch (error) {
  console.error("Authentication failed:", error.message);
}
```

#### Go

```go
package main

import (
    "fmt"
    "github.com/intrik8-labs/yggdrasil/pkg/sdk"
)

func main() {
    client := sdk.NewClient(&sdk.Config{
        ClientID:     "your_client_id",
        ClientSecret: "your_client_secret",
        BaseURL:      "https://api.yourdomain.com/api/v1",
    })

    token, err := client.Auth.GetToken(context.Background())
    if err != nil {
        panic(err)
    }

    fmt.Printf("Access token: %s\n", token.AccessToken)
}
```

### User Management

#### Create User

```go
// Go
ticketData := sdk.TicketCreate{
    Title:       "Server Down - Web Server",
    Description: "Web server is not responding to requests",
    Priority:    1,
    CustomerID:  456,
    AssignedTo:  intPtr(123),
    Tags:        []string{"server", "web", "urgent"},
}

ticket, err := client.Tickets.Create(ctx, ticketData)
if err != nil {
    log.Fatalf("Failed to create ticket: %v", err)
}
fmt.Printf("Created ticket: %s - %s\n", ticket.ID, ticket.Title)

// JavaScript
const ticketData = {
  title: 'Server Down - Web Server',
  description: 'Web server is not responding to requests',
  priority: 1,
  customerId: 456,
  assignedTo: 123,
  tags: ['server', 'web', 'urgent']
};

const ticket = await client.tickets.create(ticketData);
console.log(`Created ticket: ${ticket.id} - ${ticket.title}`);
```

#### Search Tickets

```go
// Go with filters
tickets, err := client.Tickets.Search(ctx, &sdk.SearchTicketsParams{
    Status:      stringPtr("open"),
    PriorityIn:  []int{1, 2},
    CreatedAfter: timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
    CustomerID:  intPtr(456),
    SortBy:      stringPtr("created_at"),
    Order:       stringPtr("desc"),
})

for ticket in tickets.data:
    print(f"Ticket {ticket.id}: {ticket.title}")

// JavaScript with query builder
const filters = {
  status: 'open',
  priority_in: [1, 2],
  created_after: '2024-01-01',
  customer_id: 456,
  sort_by: 'created_at',
  order: 'desc'
};

const tickets = await client.tickets.search(filters);
tickets.data.forEach(ticket => {
  console.log(`Ticket ${ticket.id}: ${ticket.title}`);
});
```

### Customer Management

#### Create Customer

```go
// Go
customerData := sdk.CustomerCreate{
    Name:  "Acme Corporation",
    Email: "contact@acme.com",
    Phone: "+1-555-0123",
    Address: &sdk.Address{
        Street: "123 Business St",
        City:   "New York",
        "state": "NY",
        "zip": "10001",
        "country": "US"
    },
    "contact_person": {
        "name": "Jane Smith",
        "email": "jane@acme.com",
        "phone": "+1-555-0124"
    }
}

customer = client.customers.create(customer_data)
print(f"Created customer: {customer.id}")

# JavaScript
const customerData = {
  name: 'Acme Corporation',
  email: 'contact@acme.com',
  phone: '+1-555-0123',
  address: {
    street: '123 Business St',
    city: 'New York',
    state: 'NY',
    zip: '10001',
    country: 'US'
  },
  contactPerson: {
    name: 'Jane Smith',
    email: 'jane@acme.com',
    phone: '+1-555-0124'
  }
};

const customer = await client.customers.create(customerData);
console.log(`Created customer: ${customer.id}`);
```

### System/Device Management

#### Register System

```go
// Go
systemData := sdk.SystemCreate{
    Hostname:    "web-server-01",
    IPAddress:   stringPtr("192.168.1.100"),
    OSType:      stringPtr("Ubuntu 22.04"),
    CPUCores:    intPtr(4),
    MemoryGB:    intPtr(8),
    DiskGB:      intPtr(500),
    CustomerID:  456,
    Tags:        []string{"web", "production", "ubuntu"},
}

system, err := client.Systems.Create(ctx, systemData)
print(f"Registered system: {system.id} - {system.hostname}")

# JavaScript
const systemData = {
  hostname: 'web-server-01',
  ipAddress: '192.168.1.100',
  osType: 'Ubuntu 22.04',
  cpuCores: 4,
  memoryGb: 8,
  diskGb: 500,
  customerId: 456,
  tags: ['web', 'production', 'ubuntu']
};

const system = await client.systems.create(systemData);
console.log(`Registered system: ${system.id} - ${system.hostname}`);
```

#### Get System Metrics

```go
// Go
metrics, err := client.Systems.GetMetrics(ctx, &sdk.GetMetricsParams{
    SystemID:  123,
    StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    EndTime:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
    Interval:  stringPtr("5m"),
})
if err != nil {
    log.Fatalf("Failed to get metrics: %v", err)
}

for _, metric := range metrics.Data {
    fmt.Printf("Time: %s, CPU: %.2f%%\n", metric.Timestamp, metric.CPUUsage)
}

# JavaScript
const metrics = await client.systems.getMetrics({
  systemId: 123,
  startTime: '2024-01-01T00:00:00Z',
  endTime: '2024-01-01T23:59:59Z',
  interval: '5m'
});

metrics.data.forEach(metric => {
  console.log(`Time: ${metric.timestamp}, CPU: ${metric.cpuUsage}%`);
});
```

## Advanced Workflows

### Multi-Step Ticket Workflow

```go
// Complete ticket resolution workflow
func resolveTicketWorkflow(ctx context.Context, client *sdk.Client, ticketID int, resolutionNotes string) error {
    // 1. Get ticket details
    ticket, err := client.Tickets.Get(ctx, ticketID)
    if err != nil {
        return fmt.Errorf("failed to get ticket: %w", err)
    }
    fmt.Printf("Processing ticket: %s\n", ticket.Title)

    // 2. Assign to technician
    if ticket.AssignedTo == nil {
        tech, err := findAvailableTechnician(ctx, client)
        if err != nil {
            return fmt.Errorf("failed to find technician: %w", err)
        }
        client.tickets.assign(ticket_id, tech.id)
        fmt.Printf("Assigned to technician: %s\n", tech.Name)

    // 3. Add work log entry
    workLog := sdk.WorkLogCreate{
        TicketID:         ticketID,
        Entry:            "Initial assessment performed",
        TimeSpentMinutes: 15,
        TechnicianID:     tech.ID,
    }
    _, err = client.Tickets.AddWorkLog(ctx, workLog)
    if err != nil {
        return fmt.Errorf("failed to add work log: %w", err)
    }

    // 4. Update ticket status
    _, err = client.Tickets.UpdateStatus(ctx, ticketID, "in_progress")
    if err != nil {
        return fmt.Errorf("failed to update ticket status: %w", err)
    }

    // 5. Perform resolution (simulated)
    # ... your business logic here ...

    // 6. Add final work log
    finalLog := sdk.WorkLogCreate{
        TicketID:         ticketID,
        Entry:            resolutionNotes,
        TimeSpentMinutes: 45,
        TechnicianID:     tech.ID,
    }
    _, err = client.Tickets.AddWorkLog(ctx, finalLog)
    if err != nil {
        return fmt.Errorf("failed to add final work log: %w", err)
    }

    // 7. Resolve ticket
    err = client.Tickets.Resolve(ctx, ticketID, resolutionNotes)
    if err != nil {
        return fmt.Errorf("failed to resolve ticket: %w", err)
    }
    fmt.Printf("Ticket %d resolved successfully\n", ticketID)

    // 8. Notify customer
    err = client.Notifications.SendTicketResolution(ctx, ticketID, resolutionNotes)
    if err != nil {
        return fmt.Errorf("failed to send notification: %w", err)
    }

    return nil
}

// Usage
func main() {
    ctx := context.Background()
    client := sdk.NewClient("your-api-key")

    err := resolveTicketWorkflow(ctx, client, 12345, "Fixed web server configuration issue")
    if err != nil {
        log.Fatalf("Workflow failed: %v", err)
    }
}
```

### Automated Customer Onboarding

```javascript
// JavaScript automated onboarding
async function onboardNewCompany(customerData) {
  try {
    // 1. Create customer
    const customer = await client.customers.create({
      name: customerData.name,
      email: customerData.email,
      phone: customerData.phone,
      plan: customerData.plan,
    });

    console.log(`Created customer: ${customer.id}`);

    // 2. Create admin user
    const adminUser = await client.users.create({
      email: customerData.adminEmail,
      name: customerData.adminName,
      role: "customer_admin",
      customer_id: customer.id,
    });

    console.log(`Created admin user: ${adminUser.id}`);

    // 3. Create initial systems
    for (const system of customerData.systems) {
      const createdSystem = await client.systems.create({
        ...system,
        customer_id: customer.id,
      });
      console.log(`Created system: ${createdSystem.id}`);
    }

    // 4. Generate welcome notification
    await client.notifications.sendWelcomeEmail(customer.id);

    // 5. Create onboarding ticket
    const onboardingTicket = await client.tickets.create({
      title: `Onboarding: ${customer.name}`,
      description: `Complete onboarding checklist for new customer`,
      customer_id: customer.id,
      priority: 2,
      assigned_to: 1, // Onboarding team
      tags: ["onboarding", "new-customer"],
    });

    return {
      customer,
      adminUser,
      onboardingTicket,
      success: true,
    };
  } catch (error) {
    console.error("Onboarding failed:", error);
    return { success: false, error: error.message };
  }
}

// Usage
const newCustomer = {
  name: "Tech Solutions Inc",
  email: "info@techsolutions.com",
  phone: "+1-555-0123",
  adminName: "John Manager",
  adminEmail: "john@techsolutions.com",
  plan: "enterprise",
  systems: [
    {
      hostname: "main-server",
      ipAddress: "10.0.1.10",
      osType: "Windows Server 2022",
    },
  ],
};

onboardNewCompany(newCustomer);
```

### Bulk Operations

```go
// Go bulk operations
type BulkOperations struct {
    client  *sdk.Client
    results []OperationResult
}

type OperationResult struct {
    Type  string      `json:""`
    Data  interface{} `json:""`
    Error string      `json:""`
}

func NewBulkOperations(client *sdk.Client) *BulkOperations {
    return &BulkOperations{
        client:  client,
        results: make([]OperationResult, 0),
    }
}

func (b *BulkOperations) BulkCreateTickets(ctx context.Context, ticketsData []sdk.TicketCreate) ([]sdk.Ticket, error) {
    var createdTickets []sdk.Ticket

    // Use batch endpoint if available
    if batchTickets, err := b.client.Tickets.CreateBatch(ctx, ticketsData); err == nil {
        return batchTickets, nil
    }

    // Fallback to individual calls
    for _, ticketData := range ticketsData {
        ticket, err := b.client.Tickets.Create(ctx, ticketData)
        if err != nil {
            b.results = append(b.results, OperationResult{
                Type:  "error",
                Data:  ticketData,
                Error: err.Error(),
            })
            continue
        }
        createdTickets = append(createdTickets, ticket)
    }

    return createdTickets, nil
}

func (b *BulkOperations) BulkUpdateSystems(ctx context.Context, updates map[int]sdk.SystemUpdate) ([]sdk.System, error) {
    var updatedSystems []sdk.System

    for systemID, updateData := range updates {
        system, err := b.client.Systems.Update(ctx, systemID, updateData)
        if err != nil {
                system = self.client.systems.update(system_id, update_data)
                updated_systems.append(system)
            except Exception as e:
                self.results.append({
                    'type': 'error',
                    'system_id': system_id,
                    'error': str(e)
                })

        return updated_systems

        func GetResultsSummary(results []BulkResult) BulkResultSummary {
            successCount := 0
            errorCount := 0
            for _, r := range results {
                if r.Type == "success" {
                    successCount++
                } else {
                    errorCount++
                }
            }

            return BulkResultSummary{
                Total:     len(results),
                Success:   successCount,
                Errors:    errorCount,
                ErrorList: getErrorList(results),
            }
        }

        func getErrorList(results []BulkResult) []BulkError {
            var errors []BulkError
            for _, r := range results {
                if r.Type == "error" {
                    errors = append(errors, r.Error)
                }
            }
            return errors
        }

# Usage
bulk_ops = BulkOperations(client)

# Bulk create tickets
tickets_to_create = [
    {
        "title": f"Server Maintenance {i}",
        "description": f"Regular maintenance for server {i}",
        "priority": 3,
        "customer_id": 456
    }
    for i in range(1, 6)
]

created = bulk_ops.bulk_create_tickets(tickets_to_create)
summary = bulk_ops.get_results_summary()
print(f"Bulk operation summary: {summary}")
```

## Webhooks

### Webhook Setup

#### Create Webhook

```go
// Go
webhookData := sdk.WebhookCreate{
    URL:    "https://yourapp.com/webhooks/yggdrasil",
    Events: []string{"ticket.created", "ticket.updated", "user.created"},
    Secret: stringPtr("your_webhook_secret"),
    Active: boolPtr(true),
}

webhook, err := client.Webhooks.Create(ctx, webhookData)
if err != nil {
    log.Fatalf("Failed to create webhook: %v", err)
}
fmt.Printf("Created webhook: %s\n", webhook.ID)

# JavaScript
const webhookData = {
  url: 'https://yourapp.com/webhooks/yggdrasil',
  events: ['ticket.created', 'ticket.updated', 'user.created'],
  secret: 'your_webhook_secret',
  active: true
};

const webhook = await client.webhooks.create(webhookData);
console.log(`Created webhook: ${webhook.id}`);
```

#### Webhook Event Structure

```json
{
  "event": "ticket.created",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "id": 12345,
    "title": "Server Down - Web Server",
    "description": "Web server is not responding",
    "status": "open",
    "priority": 1,
    "customer_id": 456,
    "created_at": "2024-01-01T12:00:00Z"
  },
  "signature": "sha256=a1b2c3d4..."
}
```

#### Webhook Verification

```go
// Go webhook verification
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func verifyWebhookSignature(payload []byte, signature string, secret string) bool {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    expectedSignature := hex.EncodeToString(h.Sum(nil))

    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

    return hmac.compare_digest(expected_signature, signature)

def handle_webhook(request):
    """Handle incoming webhook"""
    signature = request.headers.get('X-Yggdrasil-Signature')
    payload = request.body

    if not verify_webhook_signature(payload, signature, 'your_webhook_secret'):
        return {"error": "Invalid signature"}, 401

    event_data = request.json()
    event_type = event_data['event']

    if event_type == 'ticket.created':
        handle_new_ticket(event_data['data'])
    elif event_type == 'ticket.updated':
        handle_ticket_update(event_data['data'])

    return {"status": "ok"}

# JavaScript webhook verification
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');

  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

function handleWebhook(req, res) {
  const signature = req.headers['x-yggdrasil-signature'];
  const payload = JSON.stringify(req.body);

  if (!verifyWebhookSignature(payload, signature, 'your_webhook_secret')) {
    return res.status(401).json({ error: 'Invalid signature' });
  }

  const { event, data } = req.body;

  switch (event) {
    case 'ticket.created':
      handleNewTicket(data);
      break;
    case 'ticket.updated':
      handleTicketUpdate(data);
      break;
  }

  res.json({ status: 'ok' });
}
```

## Examples by Language

### cURL Examples

#### Authentication

```bash
# Get access token
ACCESS_TOKEN=$(curl -s -X POST https://api.yourdomain.com/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "your_client_id",
    "client_secret": "your_client_secret",
    "grant_type": "client_credentials"
  }' | jq -r '.access_token')

# Use token
curl -X GET https://api.yourdomain.com/api/v1/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

#### Create Ticket

```bash
curl -X POST https://api.yourdomain.com/api/v1/tickets \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Server Down - Web Server",
    "description": "Web server is not responding to requests",
    "priority": 1,
    "customer_id": 456
  }'
```

### PowerShell Examples

```powershell
# Authentication
$headers = @{
    "Content-Type" = "application/json"
}

$body = @{
    client_id = "your_client_id"
    client_secret = "your_client_secret"
    grant_type = "client_credentials"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "https://api.yourdomain.com/api/v1/auth/token" `
    -Method POST -Headers $headers -Body $body

$accessToken = $response.access_token

# API Request
$headers.Authorization = "Bearer $accessToken"
$users = Invoke-RestMethod -Uri "https://api.yourdomain.com/api/v1/users" `
    -Method GET -Headers $headers
```

### PHP Examples

```php
<?php
require 'vendor/autoload.php';

use Yggdrasil\Client as YggdrasilClient;

$client = new YggdrasilClient([
    'client_id' => 'your_client_id',
    'client_secret' => 'your_client_secret',
    'base_url' => 'https://api.yourdomain.com/api/v1'
]);

// Get token
$token = $client->auth->getToken();
echo "Access token: " . $token->access_token . "\n";

// Create user
$userData = [
    'email' => 'user@example.com',
    'name' => 'John Doe',
    'role' => 'technician'
];

$user = $client->users->create($userData);
echo "Created user: " . $user->id . "\n";
?>
```

### Ruby Examples

```ruby
require 'yggdrasil'

client = Yggdrasil::Client.new(
  client_id: 'your_client_id',
  client_secret: 'your_client_secret',
  base_url: 'https://api.yourdomain.com/api/v1'
)

# Get token
token = client.auth.get_token
puts "Access token: #{token.access_token}"

# Create ticket
ticket_data = {
  title: 'Server Down - Web Server',
  description: 'Web server is not responding',
  priority: 1,
  customer_id: 456
}

ticket = client.tickets.create(ticket_data)
puts "Created ticket: #{ticket.id} - #{ticket.title}"
```

## Testing

### API Testing with Postman

#### Import Collection

1. Open Postman
2. Click Import
3. Select "Link" and enter:
   `https://api.yourdomain.com/api/v1/postman-collection.json`
4. Save collection as "Yggdrasil API"

#### Environment Variables

```json
{
  "name": "Yggdrasil API",
  "values": [
    {
      "key": "base_url",
      "value": "https://api.yourdomain.com/api/v1",
      "enabled": true
    },
    {
      "key": "client_id",
      "value": "your_client_id",
      "enabled": true
    },
    {
      "key": "client_secret",
      "value": "your_client_secret",
      "enabled": true
    },
    {
      "key": "access_token",
      "value": "",
      "enabled": true
    }
  ]
}
```

#### Pre-request Script

```javascript
// Auto-refresh token if expired
if (pm.variables.get("access_token")) {
  // Check if token is valid (simplified check)
  // In production, parse JWT and check expiration
  const token = pm.variables.get("access_token");
  if (!token || token.includes("expired")) {
    // Refresh token
    pm.sendRequest(
      {
        url: pm.variables.get("base_url") + "/auth/token",
        method: "POST",
        header: {
          "Content-Type": "application/json",
        },
        body: {
          mode: "raw",
          raw: JSON.stringify({
            client_id: pm.variables.get("client_id"),
            client_secret: pm.variables.get("client_secret"),
            grant_type: "client_credentials",
          }),
        },
      },
      function (err, response) {
        if (response.json().access_token) {
          pm.variables.set("access_token", response.json().access_token);
        }
      },
    );
  }
} else {
  // Get initial token
  pm.sendRequest(
    {
      url: pm.variables.get("base_url") + "/auth/token",
      method: "POST",
      header: {
        "Content-Type": "application/json",
      },
      body: {
        mode: "raw",
        raw: JSON.stringify({
          client_id: pm.variables.get("client_id"),
          client_secret: pm.variables.get("client_secret"),
          grant_type: "client_credentials",
        }),
      },
    },
    function (err, response) {
      if (response.json().access_token) {
        pm.variables.set("access_token", response.json().access_token);
      }
    },
  );
}
```

### Go API Tests

```go
// api_test.go
package api_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
    ctx := context.Background()
    client := NewTestClient()

    user := &sdk.UserCreate{
        Email: "test@example.com",
        Name:  "Test User",
        Role:  "technician",
    }

    created, err := client.Users.Create(ctx, user)
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", created.Email)
    assert.Equal(t, "Test User", created.Name)
    assert.Equal(t, "technician", created.Role)
}

func TestListTickets(t *testing.T) {
    ctx := context.Background()
    client := NewTestClient()

    tickets, err := client.Tickets.List(ctx, &sdk.ListTicketsParams{
        Limit: 10,
    })
    assert.NoError(t, err)
    assert.LessOrEqual(t, len(tickets.Data), 10)
    assert.Equal(t, 10, tickets.Meta.Limit)
}

func TestCreateTicket(t *testing.T) {
    ctx := context.Background()
    client := NewTestClient()

    ticket := &sdk.TicketCreate{
        Title:       "Test Ticket",
        Description: "This is a test ticket",
        Priority:    sdk.PriorityMedium,
    }

    created, err := client.Tickets.Create(ctx, ticket)
    assert.NoError(t, err)
    assert.Equal(t, "Test Ticket", created.Title)
    assert.Equal(t, sdk.PriorityMedium, created.Priority)
}

        assert ticket.title == ticket_data["title"]
        assert ticket.description == ticket_data["description"]
        assert ticket.priority == ticket_data["priority"]
```

#### Jest API Tests (JavaScript)

```javascript
// api.test.js
const { YggdrasilClient } = require("@yggdrasil/sdk");

describe("Yggdrasil API", () => {
  let client;
  let token;

  beforeAll(async () => {
    client = new YggdrasilClient({
      clientId: "test_client_id",
      clientSecret: "test_client_secret",
      baseUrl: "https://api.yourdomain.com/api/v1",
    });

    token = await client.auth.getToken();
  });

  test("should create user", async () => {
    const userData = {
      email: "test@example.com",
      name: "Test User",
      role: "technician",
    };

    const user = await client.users.create(userData);

    expect(user.email).toBe(userData.email);
    expect(user.name).toBe(userData.name);
    expect(user.role).toBe(userData.role);
  });

  test("should list tickets", async () => {
    const response = await client.tickets.list({ limit: 10 });

    expect(response.data.length).toBeLessThanOrEqual(10);
    expect(response.meta.limit).toBe(10);
  });

  test("should create ticket", async () => {
    const ticketData = {
      title: "Test Ticket",
      description: "This is a test ticket",
      priority: 3,
    };

    const ticket = await client.tickets.create(ticketData);

    expect(ticket.title).toBe(ticketData.title);
    expect(ticket.description).toBe(ticketData.description);
    expect(ticket.priority).toBe(ticketData.priority);
  });
});
```

## Best Practices

### Security

1. **Use HTTPS always** - Never send credentials over HTTP
2. **Protect API keys** - Use environment variables, not hardcode
3. **Validate inputs** - Sanitize all user inputs
4. **Use least privilege** - Request only necessary permissions
5. **Implement retry logic** - Handle network failures gracefully

### Performance

1. **Batch operations** - Use bulk endpoints when available
2. **Pagination** - Use page limits for large datasets
3. **Caching** - Cache frequently accessed data
4. **Compression** - Accept gzip responses
5. **Connection pooling** - Reuse HTTP connections

### Error Handling

1. **Check status codes** - Handle different HTTP responses appropriately
2. **Parse error messages** - Use specific error details
3. **Implement retry logic** - Handle temporary failures
4. **Log errors** - Record debugging information
5. **User feedback** - Provide clear error messages

### Rate Limiting

1. **Monitor headers** - Check rate limit information
2. **Implement backoff** - Exponential backoff on rate limits
3. **Cache tokens** - Reduce authentication requests
4. **Batch requests** - Reduce API call count

### Webhooks

1. **Verify signatures** - Always validate webhook authenticity
2. **Handle duplicates** - Implement idempotent processing
3. **Quick responses** - Respond quickly to webhook requests
4. **Error handling** - Graceful failure handling
5. **Logging** - Log webhook processing for debugging

---

For additional API documentation, visit [API Reference](../architecture/03-api-contract.md) or open an issue on GitHub with "api" label.
