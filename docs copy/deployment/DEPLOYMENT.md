# Deployment Guide

This guide covers deployment strategies for the Yggdrasil MSP/CRM/ERP platform across different environments and infrastructure setups.

## Table of Contents

- [Deployment Guide](#deployment-guide)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
    - [System Requirements](#system-requirements)
    - [Software Requirements](#software-requirements)
    - [Network Requirements](#network-requirements)
  - [Infrastructure Requirements](#infrastructure-requirements)
    - [Database Server](#database-server)
    - [Message Broker (NATS)](#message-broker-nats)
  - [Environment Configuration](#environment-configuration)
    - [Environment Variables](#environment-variables)
    - [Production Environment File](#production-environment-file)
  - [Deployment Methods](#deployment-methods)
    - [Docker Compose (Small Deployments)](#docker-compose-small-deployments)
    - [Kubernetes (Production)](#kubernetes-production)
    - [Bare Metal (Advanced)](#bare-metal-advanced)
  - [Database Setup](#database-setup)
    - [Initial Database Migration](#initial-database-migration)
    - [Database Backups](#database-backups)
  - [Authentication Configuration](#authentication-configuration)
    - [Basic Authentication Setup](#basic-authentication-setup)
    - [OAuth2/OIDC Setup (Future)](#oauth2oidc-setup-future)
  - [SSL/TLS Configuration](#ssltls-configuration)
    - [Nginx Configuration](#nginx-configuration)
    - [SSL Certificate Setup](#ssl-certificate-setup)
  - [Post-Deployment Verification](#post-deployment-verification)
    - [Health Check Endpoints](#health-check-endpoints)
    - [Smoke Tests](#smoke-tests)
  - [Monitoring and Health Checks](#monitoring-and-health-checks)
    - [Health Check Endpoints](#health-check-endpoints-1)
    - [Monitoring Stack (Optional)](#monitoring-stack-optional)
  - [Troubleshooting](#troubleshooting)
    - [Common Issues](#common-issues)
    - [Performance Optimization](#performance-optimization)
    - [Security Checklist](#security-checklist)

## Prerequisites

### System Requirements

**Minimum Requirements:**

- CPU: 4 cores
- Memory: 8GB RAM
- Storage: 50GB SSD
- Network: 100Mbps

**Recommended Production Requirements:**

- CPU: 8+ cores
- Memory: 16GB+ RAM
- Storage: 100GB+ SSD (separate for database)
- Network: 1Gbps

### Software Requirements

- Docker 20.10+
- Docker Compose 2.0+
- Kubernetes 1.25+ (for K8s deployment)
- PostgreSQL 15+ with TimescaleDB 2.8+
- NATS Server 2.9+
- Node.js 20+ (for frontend builds)
- Go 1.22+ (for control plane and agent builds)

### Network Requirements

- Ports 80, 443 (web traffic)
- Port 5432 (PostgreSQL)
- Port 4222 (NATS)
- Port 8080 (Go control plane API, internal)
- Port 3000 (React dev server, internal)

## Infrastructure Requirements

### Database Server

**PostgreSQL + TimescaleDB Setup:**

```bash
# Install PostgreSQL with TimescaleDB
# Ubuntu/Debian
curl https://packagecloud.io/timescale/timescaledb/gpgkey | sudo apt-key add -
echo "deb https://packagecloud.io/timescale/timescaledb/ubuntu/ $(lsb_release -c -s) main" | sudo tee /etc/apt/sources.list.d/timescaledb.list
sudo apt update
sudo apt install timescaledb-2-postgresql-14 postgresql-14

# Enable TimescaleDB
sudo timescaledb-tune --quiet --yes
sudo systemctl restart postgresql
```

**Database Creation:**

```sql
-- Create database and user
CREATE DATABASE yggdrasil;
CREATE USER yggdrasil_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE yggdrasil TO yggdrasil_user;

-- Enable TimescaleDB
\c yggdrasil
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
```

### Message Broker (NATS)

**Standalone NATS Server:**

```bash
# Using Docker (recommended)
docker run -d --name nats \
  -p 4222:4222 \
  -p 6222:6222 \
  -p 8222:8222 \
  nats:2.9-alpine \
  -js -m 8222

# Or using package manager
sudo apt install nats-server
sudo systemctl enable nats-server
sudo systemctl start nats-server
```

## Environment Configuration

### Environment Variables

Create `.env` file in project root:

```bash
# Database Configuration
DATABASE_URL=postgresql://yggdrasil_user:secure_password@localhost:5432/yggdrasil
DATABASE_POOL_SIZE=20
DATABASE_MAX_OVERFLOW=30

# NATS Configuration
NATS_URL=nats://localhost:4222
NATS_JETSTREAM=true

# Application Configuration
ENVIRONMENT=production
DEBUG=false
LOG_LEVEL=INFO

# Security Configuration
# Use RS256 for asymmetric signing (public/private key pair)
# This aligns with mTLS agent certificate authentication
SECRET_KEY=your-super-secret-key-change-in-production
JWT_SECRET=your-jwt-secret-change-in-production
JWT_ALGORITHM=RS256
JWT_EXPIRE_MINUTES=1440

# Control Plane Configuration
CONTROL_PLANE_HOST=0.0.0.0
CONTROL_PLANE_PORT=8000
CORS_ORIGINS=["https://yourdomain.com"]

# Web Application Configuration
WEB_HOST=0.0.0.0
WEB_PORT=3000
API_BASE_URL=https://api.yourdomain.com

# Agent Configuration
AGENT_REGISTRATION_TOKEN=your-agent-token-here
AGENT_HEARTBEAT_INTERVAL=60
# Agent heartbeat and metrics collection occur every 60 seconds (aligned with implementation roadmap)

# External Services (v1.0+ features, not required for v0.5)
# Redis caching (v1.0+)
REDIS_URL=redis://localhost:6379
# Email notifications (v1.0+)
SMTP_HOST=smtp.yourprovider.com
SMTP_PORT=587
SMTP_USER=your-smtp-user
SMTP_PASSWORD=your-smtp-password
```

### Production Environment File

For production, create `.env.production`:

```bash
# Production-specific settings
ENVIRONMENT=production
DEBUG=false
LOG_LEVEL=WARNING

# Performance settings
WORKERS=4
WORKER_CONNECTIONS=1000
KEEPALIVE_TIMEOUT=65

# Security settings
ALLOWED_HOSTS=["yourdomain.com", "www.yourdomain.com"]
SECURE_SSL_REDIRECT=true
SECURE_HSTS_SECONDS=31536000
SECURE_HSTS_INCLUDE_SUBDOMAINS=true
SECURE_HSTS_PRELOAD=true
```

## Deployment Methods

### Docker Compose (Small Deployments)

This method is ideal for development, testing, and small production deployments.

**Production Docker Compose File:**

```yaml
# docker-compose.prod.yml
version: "3.8"

services:
  postgres:
    image: timescale/timescaledb:latest-pg14
    container_name: yggdrasil-db
    environment:
      POSTGRES_DB: yggdrasil
      POSTGRES_USER: yggdrasil_user
      POSTGRES_PASSWORD: ${DATABASE_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    ports:
      - "5432:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U yggdrasil_user -d yggdrasil"]
      interval: 30s
      timeout: 10s
      retries: 3

  nats:
    image: nats:2.9-alpine
    container_name: yggdrasil-nats
    command: ["-js", "-m", "8222"]
    ports:
      - "4222:4222"
      - "6222:6222"
      - "8222:8222"
    volumes:
      - nats_data:/data
    restart: unless-stopped
    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--quiet",
          "--tries=1",
          "--spider",
          "http://localhost:8222/varz",
        ]
      interval: 30s
      timeout: 10s
      retries: 3

  control-plane:
    build:
      context: .
      dockerfile: Dockerfile
      target: api
    container_name: yggdrasil-control-plane
    environment:
      - DATABASE_URL=postgresql://yggdrasil_user:${DATABASE_PASSWORD}@postgres:5432/yggdrasil
      - NATS_URL=nats://nats:4222
    env_file:
      - .env.production
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      nats:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    container_name: yggdrasil-web
    environment:
      - API_BASE_URL=http://control-plane:8080
    ports:
      - "3000:3000"
    depends_on:
      - control-plane
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    container_name: yggdrasil-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - static_files:/var/www/static
    depends_on:
      - web
      - control-plane
    restart: unless-stopped

volumes:
  postgres_data:
  nats_data:
  static_files:
```

**Deployment Commands:**

```bash
# Build and start services
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d

# Run database migrations
docker-compose -f docker-compose.prod.yml exec control-plane migrate -path /app/migrations -database "${DATABASE_URL}" up

# Check service status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f control-plane
```

### Kubernetes (Production)

For scalable production deployments, use Kubernetes with the following manifests.

**Namespace:**

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: yggdrasil
```

**ConfigMap:**

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yggdrasil-config
  namespace: yggdrasil
data:
  ENVIRONMENT: "production"
  LOG_LEVEL: "INFO"
  DATABASE_HOST: "postgres-service"
  DATABASE_PORT: "5432"
  DATABASE_NAME: "yggdrasil"
  NATS_URL: "nats://nats-service:4222"
```

**Secret:**

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: yggdrasil-secrets
  namespace: yggdrasil
type: Opaque
data:
  DATABASE_PASSWORD: <base64-encoded-password>
  SECRET_KEY: <base64-encoded-secret>
  JWT_SECRET: <base64-encoded-jwt-secret>
```

**PostgreSQL Deployment:**

```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: yggdrasil
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: timescale/timescaledb:latest-pg14
          env:
            - name: POSTGRES_DB
              valueFrom:
                configMapKeyRef:
                  name: yggdrasil-config
                  key: DATABASE_NAME
            - name: POSTGRES_USER
              value: yggdrasil_user
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: yggdrasil-secrets
                  key: DATABASE_PASSWORD
          ports:
            - containerPort: 5432
          volumeMounts:
            - name: postgres-storage
              mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
    - metadata:
        name: postgres-storage
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 50Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: yggdrasil
spec:
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432
```

**Control Plane Deployment:**

```yaml
# k8s/control-plane.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: control-plane
  namespace: yggdrasil
spec:
  replicas: 3
  selector:
    matchLabels:
      app: control-plane
  template:
    metadata:
      labels:
        app: control-plane
    spec:
      containers:
        - name: control-plane
          image: yggdrasil/control-plane:latest
          env:
            - name: DATABASE_URL
              value: "postgresql://yggdrasil_user:$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)"
            - name: DATABASE_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: yggdrasil-secrets
                  key: DATABASE_PASSWORD
            - name: SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: yggdrasil-secrets
                  key: SECRET_KEY
          envFrom:
            - configMapRef:
                name: yggdrasil-config
          ports:
            - containerPort: 8000
          livenessProbe:
            httpGet:
              path: /health
              port: 8000
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8000
            initialDelaySeconds: 5
            periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: control-plane-service
  namespace: yggdrasil
spec:
  selector:
    app: control-plane
  ports:
    - port: 8000
      targetPort: 8000
```

**Deployment Commands:**

```bash
# Apply all manifests
kubectl apply -f k8s/

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=control-plane -n yggdrasil --timeout=300s

# Run migrations
kubectl exec -n yggdrasil deployment/control-plane -- migrate -path /app/migrations -database "${DATABASE_URL}" up

# Check logs
kubectl logs -n yggdrasil -l app=control-plane -f
```

### Bare Metal (Advanced)

For organizations that prefer bare metal deployments without containers.

**System Preparation:**

```bash
# Install dependencies
sudo apt update
sudo apt install -y nginx postgresql-14 nats-server curl wget

# Create application user
sudo useradd -m -s /bin/bash yggdrasil
sudo usermod -aG sudo yggdrasil
```

**Application Installation:**

```bash
# As yggdrasil user
sudo su - yggdrasil

# Clone and setup control plane
git clone https://github.com/intrik8-labs/yggdrasil.git
cd yggdrasil

# Install Go dependencies
go mod download

# Generate sqlc code
sqlc generate

# Copy environment file
cp .env.production .env
# Edit .env with your settings

# Run database migrations
migrate -path migrations -database "${DATABASE_URL}" up
```

**Systemd Service Files:**

```ini
# /etc/systemd/system/yggdrasil-control-plane.service
[Unit]
Description=Yggdrasil Control Plane
After=network.target postgresql.service nats-server.service

[Service]
Type=exec
User=yggdrasil
Group=yggdrasil
WorkingDirectory=/home/yggdrasil/yggdrasil
ExecStart=/home/yggdrasil/yggdrasil/bin/yggdrasil-api
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start services
sudo systemctl daemon-reload
sudo systemctl enable yggdrasil-control-plane
sudo systemctl start yggdrasil-control-plane
```

## Database Setup

### Initial Database Migration

```bash
# For Docker deployments
docker-compose exec control-plane migrate -path /app/migrations -database "${DATABASE_URL}" up

# For Kubernetes
kubectl exec deployment/control-plane -n yggdrasil -- migrate -path /app/migrations -database "${DATABASE_URL}" up

# For bare metal
cd /home/yggdrasil/yggdrasil
migrate -path migrations -database "${DATABASE_URL}" up
```

### Database Backups

**Automated Backup Script:**

```bash
#!/bin/bash
# scripts/backup-db.sh

BACKUP_DIR="/backups/yggdrasil"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/yggdrasil_backup_$DATE.sql"

mkdir -p $BACKUP_DIR

# Create backup
pg_dump -h localhost -U yggdrasil_user -d yggdrasil > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Remove old backups (keep last 7 days)
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

**Cron Job for Daily Backups:**

```bash
# Add to crontab: crontab -e
0 2 * * * /path/to/yggdrasil/scripts/backup-db.sh
```

## Authentication Configuration

### Basic Authentication Setup

For initial deployment, Yggdrasil uses basic authentication with JWT tokens.

**Create Default Admin User:**

```bash
# Using CLI tool
docker-compose exec control-plane go run cmd/admin/create.go \
  --email admin@yourdomain.com \
  --password secure_admin_password \
  --name "System Administrator"
```

**Authentication Configuration:**

```yaml
# Add to your environment configuration
AUTH_METHOD=basic
JWT_EXPIRE_MINUTES=1440
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBERS=true
PASSWORD_REQUIRE_SYMBOLS=true
```

### OAuth2/OIDC Setup (Future)

For enterprise SSO integration:

```yaml
# OAuth2 Configuration
AUTH_METHOD=oauth2
OAUTH2_CLIENT_ID=your-client-id
OAUTH2_CLIENT_SECRET=your-client-secret
OAUTH2_AUTHORIZATION_URL=https://your-oauth-provider.com/auth
OAUTH2_TOKEN_URL=https://your-oauth-provider.com/token
OAUTH2_USER_INFO_URL=https://your-oauth-provider.com/userinfo
OAUTH2_CALLBACK_URL=https://yourdomain.com/auth/callback
```

## SSL/TLS Configuration

### Nginx Configuration

```nginx
# nginx/nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream web {
        server web:3000;
    }

    upstream api {
        server control-plane:8080;
    }

    # HTTP to HTTPS redirect
    server {
        listen 80;
        server_name yourdomain.com www.yourdomain.com;
        return 301 https://$server_name$request_uri;
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name yourdomain.com www.yourdomain.com;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header Strict-Transport-Security "max-age=63072000" always;
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";

        # Web application
        location / {
            proxy_pass http://web;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # API endpoints
        location /api/ {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Websocket support for real-time features
        location /ws/ {
            proxy_pass http://api;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
        }
    }
}
```

### SSL Certificate Setup

**Using Let's Encrypt:**

```bash
# Install certbot (Python dependency for Let's Encrypt)
sudo apt install certbot python3-certbot-nginx

# Generate certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

**Manual SSL Setup:**

```bash
# Generate self-signed certificate for testing
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem
```

## Post-Deployment Verification

### Health Check Endpoints

```bash
# Check control plane health
curl https://api.yourdomain.com/health

# Check web application
curl https://yourdomain.com/health

# Check database connectivity
curl https://api.yourdomain.com/health/db

# Check NATS connectivity
curl https://api.yourdomain.com/health/nats
```

### Smoke Tests

```bash
# Test authentication
curl -X POST https://api.yourdomain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yourdomain.com","password":"your_password"}'

# Test API access
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  https://api.yourdomain.com/api/v1/users
```

## Monitoring and Health Checks

### Health Check Endpoints

The control plane provides comprehensive health check endpoints:

- `/health` - Overall system health
- `/health/db` - Database connectivity
- `/health/nats` - Message broker connectivity
- `/health/redis` - Cache connectivity (if configured)
- `/metrics` - Prometheus metrics (if configured)

### Monitoring Stack (Optional)

```yaml
# Add to docker-compose.prod.yml
prometheus:
  image: prom/prometheus:latest
  ports:
    - "9090:9090"
  volumes:
    - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
  command:
    - "--config.file=/etc/prometheus/prometheus.yml"
    - "--storage.tsdb.path=/prometheus"
    - "--web.console.libraries=/etc/prometheus/console_libraries"
    - "--web.console.templates=/etc/prometheus/consoles"

grafana:
  image: grafana/grafana:latest
  ports:
    - "3001:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=admin
  volumes:
    - grafana_data:/var/lib/grafana
    - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
    - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
```

## Troubleshooting

### Common Issues

**Database Connection Issues:**

```bash
# Check database status
docker-compose exec postgres pg_isready

# Check database logs
docker-compose logs postgres

# Test connection manually
docker-compose exec postgres psql -U yggdrasil_user -d yggdrasil
```

**NATS Connection Issues:**

```bash
# Check NATS status
curl http://localhost:8222/varz

# Test NATS connection
docker-compose exec nats nats -s nats://localhost:4222 pub test "hello"
```

**Control Plane Issues:**

```bash
# Check logs
docker-compose logs control-plane

# Debug mode
docker-compose exec control-plane /app/bin/api

# Database migration issues
docker-compose exec control-plane migrate -path /app/migrations -database "${DATABASE_URL}" version
docker-compose exec control-plane migrate -path /app/migrations -database "${DATABASE_URL}" up
```

**Web Application Issues:**

```bash
# Check logs
docker-compose logs web

# Build issues
docker-compose build --no-cache web

# API connectivity
docker-compose exec web curl http://control-plane:8080/health
```

### Performance Optimization

**Database Performance:**

```sql
-- Check slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Check database size
SELECT pg_size_pretty(pg_database_size('yggdrasil'));

-- Check table sizes
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Application Performance:**

```bash
# Monitor resource usage
docker stats

# Check response times
curl -w "@curl-format.txt" -o /dev/null -s https://api.yourdomain.com/health

# Load testing
ab -n 100 -c 10 https://api.yourdomain.com/health
```

### Security Checklist

- [ ] Change all default passwords
- [ ] Configure SSL/TLS certificates
- [ ] Set up firewall rules
- [ ] Enable database SSL
- [ ] Configure security headers
- [ ] Set up log monitoring
- [ ] Enable audit logging
- [ ] Configure backup encryption
- [ ] Set up intrusion detection
- [ ] Regular security updates

This deployment guide provides comprehensive coverage for deploying Yggdrasil in production environments. Adjust the configurations based on your specific infrastructure and security requirements.

```bash
# Monitor resource usage
docker stats

# Check response times
curl -w "@curl-format.txt" -o /dev/null -s https://api.yourdomain.com/health

# Load testing
ab -n 100 -c 10 https://api.yourdomain.com/health
```

### Security Checklist

- [ ] Change all default passwords
- [ ] Configure SSL/TLS certificates
- [ ] Set up firewall rules
- [ ] Enable database SSL
- [ ] Configure security headers
- [ ] Set up log monitoring
- [ ] Enable audit logging
- [ ] Configure backup encryption
- [ ] Set up intrusion detection
- [ ] Regular security updates

This deployment guide provides comprehensive coverage for deploying Yggdrasil in production environments. Adjust the configurations based on your specific infrastructure and security requirements.
