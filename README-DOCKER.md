# Docker Deployment Guide

## Quick Start

```bash
# Generate development HTTPS certificate (first time only)
./scripts/generate-dev-cert.sh

# Build and start all services
docker-compose up --build

# Or run in detached mode
docker-compose up --build -d
```

Access services:
- **UI**: http://localhost:3000
- **Control Plane API**: https://localhost:5001
- **Swagger UI**: https://localhost:5001/swagger
- **PostgreSQL**: localhost:5432

## Certificate Setup

The control plane requires an HTTPS certificate. For development:

1. Create the `certs/` directory:
   ```bash
   mkdir -p certs
   ```

2. Generate a development certificate:
   ```bash
   dotnet dev-certs https -ep certs/aspnetapp.pfx -p development
   ```

   Or use the provided script:
   ```bash
   ./scripts/generate-dev-cert.sh
   ```

3. (Optional) Set a custom certificate password via environment variable:
   ```bash
   export CERT_PASSWORD=your-secure-password
   ```

For production, replace with a proper certificate from your CA.

## Services

### PostgreSQL Database
- **Image**: postgres:16-alpine
- **Port**: 5432
- **Credentials**: smidr/smidr
- **Volume**: postgres-data (persistent)

### Control Plane API
- **Built from**: ./control-plane/Dockerfile
- **Port**: 5001 (HTTPS)
- **Database**: PostgreSQL (migrations applied on startup)
- **CA Data**: Persisted to control-plane-data volume
- **Health Check**: Checks Swagger endpoint

### UI
- **Built from**: ./ui/Dockerfile (Lambert will create)
- **Port**: 3000
- **API Connection**: https://control-plane:5001 (internal network)

## Management Commands

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (full reset)
docker-compose down -v

# View logs
docker-compose logs -f

# View logs for specific service
docker-compose logs -f control-plane

# Rebuild specific service
docker-compose up --build control-plane

# Execute command in running container
docker-compose exec control-plane /bin/bash
docker-compose exec postgres psql -U smidr -d smidr
```

## Environment Variables

Override defaults in `.env` file or via command line:

```bash
# Example .env file
CERT_PASSWORD=my-secure-password
POSTGRES_PASSWORD=smidr
POSTGRES_USER=smidr
POSTGRES_DB=smidr
```

## Troubleshooting

### Control plane won't start
- Ensure certificate exists: `ls -la certs/aspnetapp.pfx`
- Check logs: `docker-compose logs control-plane`
- Verify PostgreSQL is healthy: `docker-compose ps`

### UI can't connect to API
- Check control plane health: `curl -k https://localhost:5001/swagger`
- Verify CORS settings in control-plane/Program.cs
- Check docker network: `docker network inspect smidr_smidr-network`

### Database connection issues
- Ensure PostgreSQL is ready: `docker-compose logs postgres`
- Wait for health check to pass before starting control plane
- Check connection string in docker-compose.yml

### Certificate errors
- Regenerate certificate: `./scripts/generate-dev-cert.sh`
- Trust certificate on host: `dotnet dev-certs https --trust`
- For browsers: Accept self-signed certificate warning

## Production Considerations

1. **Certificates**: Replace dev cert with production certificate from trusted CA
2. **Secrets**: Use Docker secrets or environment variable injection (not hardcoded)
3. **PostgreSQL**: Use managed database service or configure proper backups
4. **Volumes**: Ensure regular backups of postgres-data and control-plane-data
5. **Networking**: Use reverse proxy (nginx/traefik) for TLS termination
6. **Health Checks**: Configure monitoring and alerting on health check failures
