# Smidr Setup & Integration Guide

## Architecture

The Smidr platform consists of two main components:

### Backend (smidr-core)

- **Agent Server** - Handles agent registration, heartbeats, and job polling (`localhost:9001`)
- **Web Server** - Provides REST API for the frontend (`localhost:8080`)
- **Store** - SQLite database for persistence

### Frontend (web)

- **Next.js Application** - React-based UI (`localhost:3000`)
- **API Client** - TypeScript client for backend communication
- **Pages** - Dashboard, Agents, and Agent Details

## Quick Start

### 1. Start the Backend

```bash
# From project root
make build  # Build the backend

# Run with default config (uses SQLite)
./bin/smidr-core -config config.yaml

# Or use the server deployment
make deploy-server
```

The backend will start on:

- Agent API: `http://localhost:9001`
- Web API: `http://localhost:8080`

### 2. Start the Frontend

```bash
# Navigate to web directory
cd web

# Install dependencies
npm install

# Set up environment (if needed)
cp .env.example .env.local

# Start development server
npm run dev
```

The frontend will be available at: `http://localhost:3000`

### 3. Register an Agent

When the backend is running, agents can register with the Agent Server:

```bash
curl -X POST http://localhost:9001/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-1",
    "name": "Build Server 1",
    "capabilities": ["docker", "kubernetes"],
    "metadata": {"region": "us-east-1"}
  }'
```

### 4. View in Dashboard

Navigate to the frontend and you should see:

- **Dashboard** - Overview of connected agents and system status
- **Agents** - Full list of registered agents with details
- **Agent Details** - Detailed information about each agent including capabilities and metadata

## API Endpoints

### Web API (Frontend <-> Backend)

```
GET  /health                  # Health check
GET  /api/v1/agents          # List all agents
GET  /api/v1/agents/{id}     # Get agent details
GET  /api/v1/jobs            # List jobs (WIP)
POST /api/v1/jobs            # Create job (WIP)
GET  /api/v1/jobs/{id}       # Get job details (WIP)
```

### Agent API (Agent <-> Backend)

```
GET  /health                  # Health check
POST /api/v1/register        # Register agent
POST /api/v1/heartbeat       # Send heartbeat
GET  /api/v1/jobs/poll       # Poll for jobs
```

## Frontend Components

### Pages

- **`/`** - Home/Landing page
- **`/dashboard`** - Main dashboard with agent overview
- **`/agents`** - List all registered agents
- **`/agents/[id]`** - Detailed agent information

### API Client (`lib/api.ts`)

Type-safe API client for all backend operations:

```typescript
import { apiClient } from '@/lib/api'

// Fetch agents
const response = await apiClient.listAgents()

// Get single agent
const agent = await apiClient.getAgent('agent-1')
```

### Hooks (`lib/hooks.ts`)

React hooks for data fetching with automatic polling:

```typescript
import { useAgents } from '@/lib/hooks'

function MyComponent() {
  const { agents, loading, error } = useAgents()
  // ...
}
```

## Configuration

### Backend Configuration (`config.yaml`)

```yaml
agent_server:
  address: "0.0.0.0:9001"

web_server:
  address: "0.0.0.0:8080"

store:
  type: "sqlite"
  sqlite_path: "./smidr.db"
```

### Frontend Configuration (`.env.local`)

```bash
# API endpoint for the backend
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Troubleshooting

### Frontend can't connect to backend

1. Verify the backend is running: `curl http://localhost:8080/health`
2. Check `NEXT_PUBLIC_API_URL` in `.env.local`
3. Ensure no firewall blocking the connection

### Agents not appearing on dashboard

1. Verify agents are registered: `curl http://localhost:8080/api/v1/agents`
2. Check agent heartbeats: Register and send heartbeat within 60 seconds
3. Check backend logs for registration errors

### CORS issues

Add CORS middleware to backend if needed. The frontend will automatically handle API calls.

## Development

### Building for Production

```bash
# Frontend
cd web
npm run build
npm start

# Backend
make build
./bin/smidr-server -config config.yaml
```

### Docker

Both components can be containerized:

```bash
# Frontend
cd web
docker build -t smidr-web .
docker run -p 3000:3000 -e NEXT_PUBLIC_API_URL=http://localhost:8080 smidr-web

# Backend
docker build -f Dockerfile.core -t smidr-core .
docker run -p 8080:8080 -p 9001:9001 smidr-core
```
