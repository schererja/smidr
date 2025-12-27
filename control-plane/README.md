# Smidr Control Plane

Clean architecture ASP.NET Core application providing gRPC services for agents and GraphQL API for clients/UI.

## Architecture

- **Domain Layer**: Core business entities (Tenant, Agent, Job, Project, Artifact)
- **Application Layer**: Business logic, interfaces, and DTOs
- **Infrastructure Layer**: EF Core DbContext, repositories, external services
- **API Layer**: gRPC services (agents) + GraphQL API (clients/UI)

## Features

- ✅ **Multi-tenancy**: Tenant isolation via headers/metadata
- ✅ **gRPC**: High-performance agent communication
- ✅ **GraphQL**: Flexible querying for web/mobile clients
- ✅ **TLS/mTLS**: Secure communication
- ✅ **Clean Architecture**: Separation of concerns, testability
- ✅ **Entity Framework Core**: SQLite (dev), easily swap for Postgres/SQL Server

## Project Structure

```
control-plane/
├── Smidr.ControlPlane.sln
├── Generated/                      # Proto-generated C# files
├── src/
│   ├── Smidr.ControlPlane.Api/            # gRPC + GraphQL endpoints
│   │   ├── Services/
│   │   │   └── AgentGrpcService.cs        # Agent gRPC implementation
│   │   ├── GraphQL/
│   │   │   ├── Query.cs                   # GraphQL queries
│   │   │   └── Mutation.cs                # GraphQL mutations
│   │   ├── Middleware/
│   │   │   └── TenantResolutionMiddleware.cs
│   │   ├── Program.cs
│   │   └── appsettings.json
│   ├── Smidr.ControlPlane.Application/     # Business logic
│   │   └── Common/
│   │       └── TenantContext.cs
│   ├── Smidr.ControlPlane.Domain/          # Entities & interfaces
│   │   └── Entities/
│   │       ├── Tenant.cs
│   │       ├── Agent.cs
│   │       ├── Job.cs
│   │       ├── Project.cs
│   │       └── Artifact.cs
│   └── Smidr.ControlPlane.Infrastructure/  # Data access
│       └── Persistence/
│           └── SmidrDbContext.cs
```

## Getting Started

### Prerequisites

- .NET 8 SDK
- (Optional) Docker for containerization
- (Optional) OpenSSL for TLS certificates

### Generate TLS Certificates (Development)

```bash
# Create certs directory
mkdir -p src/Smidr.ControlPlane.Api/certs
cd src/Smidr.ControlPlane.Api/certs

# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
  -days 365 -nodes -subj "/CN=localhost"

# Convert to PFX (for .NET)
openssl pkcs12 -export -out server.pfx -inkey server.key -in server.crt \
  -passout pass:

cd ../../../..
```

### Build & Run

```bash
# Restore dependencies
dotnet restore

# Build solution
dotnet build

# Run control plane
cd src/Smidr.ControlPlane.Api
dotnet run

# Control plane will start on:
# - HTTP: http://localhost:5000 (redirect to HTTPS)
# - HTTPS: https://localhost:5001 (GraphQL)
# - gRPC: https://localhost:5002 (Agents)
```

### Database

The application uses SQLite by default (`smidr.db`). The database is created automatically on first run.

**To use PostgreSQL or SQL Server:**

1. Update `appsettings.json` connection string
2. Install appropriate EF Core provider package
3. Update `Program.cs` to use the provider

## API Usage

### gRPC (Agents)

Agents connect via gRPC on port 5002 with mTLS. Implement the services defined in the proto files:

```bash
# Example: Register agent (using grpcurl)
grpcurl -plaintext -d '{
  "agent_name": "builder-01",
  "hostname": "builder-01.local",
  "capabilities": {
    "platform": "PLATFORM_LINUX",
    "architecture": "ARCHITECTURE_AMD64",
    "features": ["docker", "yocto"]
  }
}' \
  -H 'x-tenant-id: <tenant-guid>' \
  localhost:5002 smidr.agent.AgentService/RegisterAgent
```

### GraphQL (Clients/UI)

Access GraphQL Playground at `https://localhost:5001/graphql`

**Create a tenant:**
```graphql
mutation {
  createTenant(name: "Acme Corp", slug: "acme") {
    id
    name
    slug
    status
  }
}
```

**Create a project:**
```graphql
mutation {
  createProject(name: "IoT Build", description: "Yocto builds for IoT devices") {
    id
    name
  }
}
```
*Note: Include `X-Tenant-Id` header with tenant GUID*

**Submit a job:**
```graphql
mutation {
  submitJob(
    projectId: "<project-guid>",
    name: "Build core-image-minimal",
    jobType: "yocto-build",
    description: "Kirkstone build for RPI4"
  ) {
    id
    name
    state
    submittedAt
  }
}
```

**Query jobs:**
```graphql
query {
  jobs {
    id
    name
    state
    submittedAt
    assignedAgent {
      name
      state
    }
  }
}
```

## Multi-Tenancy

Tenant context is resolved from:

1. **HTTP Header**: `X-Tenant-Id: <guid>` or `X-Tenant-Slug: <slug>`
2. **gRPC Metadata**: `x-tenant-id: <guid>`

All queries/mutations are scoped to the tenant context. gRPC services verify tenant ownership.

## TLS Configuration

### Development
- Self-signed certificates (see above)
- Set `ASPNETCORE_ENVIRONMENT=Development`

### Production
- Use certificates from a trusted CA (Let's Encrypt, DigiCert, etc.)
- Update `appsettings.json` with cert paths
- Enable mTLS for agents:

```json
{
  "Kestrel": {
    "Endpoints": {
      "Grpc": {
        "ClientCertificateMode": "RequireCertificate",
        "ClientCertificateValidation": {
          "AllowedCertificateTypes": "SelfSigned",
          "ValidateCertificateUse": true
        }
      }
    }
  }
}
```

## Testing

### Health Check
```bash
curl http://localhost:5000/health
```

### gRPC Reflection (Development)
```bash
grpcurl -plaintext localhost:5002 list
```

## Docker

```bash
# Build image
docker build -t smidr-control-plane -f src/Smidr.ControlPlane.Api/Dockerfile .

# Run container
docker run -p 5000:5000 -p 5001:5001 -p 5002:5002 \
  -v $(pwd)/certs:/app/certs \
  smidr-control-plane
```

## Next Steps

- [ ] Add authentication (JWT/OAuth)
- [ ] Implement artifact storage (S3/Azure Blob)
- [ ] Add job scheduling policies
- [ ] Implement log streaming
- [ ] Add metrics/tracing (OpenTelemetry)
- [ ] Create agent SDK (Go/C#/Python)
- [ ] Add WebSocket subscriptions for real-time updates

## Related

- **Executor**: [executor/README.md](../executor/README.md)
- **Proto Definitions**: [proto/smidr/v1/](../proto/smidr/v1/)
- **Architecture**: [docs/architecture/](../docs/architecture/)
