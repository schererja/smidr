# Smidr UI - Docker Setup

## Overview

This directory contains a production-ready Dockerfile for the Smidr UI.

## Architecture

**Multi-stage build:**
1. **Build stage:** Uses Node 22 Alpine to install dependencies and build the Vite production bundle
2. **Serve stage:** Uses nginx Alpine to serve the static files

## Configuration

### API Endpoint

The UI needs to know the control plane API URL. Vite bakes environment variables into the build at **build time**.

In `docker-compose.yml`, Dallas should pass the API URL as a build argument:

```yaml
services:
  ui:
    build:
      context: ./ui
      args:
        - VITE_API_BASE_URL=https://control-plane:5001
    ports:
      - "3000:80"
    depends_on:
      - control-plane
```

### Ports

- Container exposes port **80** (nginx)
- Map to host port 3000 (or any available port)

### Health Check

Nginx serves a `/health` endpoint that returns "healthy" for docker-compose health checks.

## Files

- **Dockerfile** - Multi-stage build (node build + nginx serve)
- **nginx.conf** - Custom nginx config with:
  - Client-side routing support (serves index.html for all routes)
  - Gzip compression
  - Security headers
  - Aggressive caching for static assets
  - Health check endpoint
- **.dockerignore** - Excludes node_modules, dist, etc. from build context

## Build & Run

Local build test:
```bash
cd ui
docker build --build-arg VITE_API_BASE_URL=https://localhost:5001 -t smidr-ui .
docker run -p 3000:80 smidr-ui
```

Open browser: http://localhost:3000

## Notes for Dallas

- The UI is a static site after build, so no environment variables at runtime
- Control plane URL must be set at **build time** via `VITE_API_BASE_URL`
- Use build args in docker-compose to inject the service URL
- UI expects HTTPS API (control plane uses TLS)
- Make sure control-plane service name matches the URL in build args
