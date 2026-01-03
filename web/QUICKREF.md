# Quick Reference

## Start Services

```bash
# Terminal 1: Backend
cd /Users/schererja/src/github.com/schererja/smidr
make build
./bin/smidr-core -config config.yaml

# Terminal 2: Frontend
cd /Users/schererja/src/github.com/schererja/smidr/web
npm run dev
```

## URLs

| Service | URL | Purpose |
|---------|-----|---------|
| Frontend | <http://localhost:3000> | Web UI |
| Backend API | <http://localhost:8080> | Web API |
| Agent API | <http://localhost:9001> | Agent Communication |

## Frontend Routes

| Path | Component | Purpose |
|------|-----------|---------|
| `/` | `app/page.tsx` | Landing page |
| `/dashboard` | `app/dashboard/page.tsx` | Dashboard overview |
| `/agents` | `app/agents/page.tsx` | Agent list |
| `/agents/[id]` | `app/agents/[id]/page.tsx` | Agent details |

## API Calls

### Register Agent

```bash
curl -X POST http://localhost:9001/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-1",
    "name": "My Agent",
    "capabilities": ["docker"],
    "metadata": {}
  }'
```

### List Agents (Frontend uses this)

```bash
curl http://localhost:8080/api/v1/agents
```

### Get Agent Details

```bash
curl http://localhost:8080/api/v1/agents/agent-1
```

## Key Files

| File | Purpose |
|------|---------|
| `lib/api.ts` | API client with all endpoints |
| `lib/hooks.ts` | React hooks for data fetching |
| `app/dashboard/page.tsx` | Main dashboard component |
| `app/agents/page.tsx` | Agents list component |
| `app/agents/[id]/page.tsx` | Agent detail component |

## Environment

```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## npm Commands

```bash
npm run dev      # Start dev server
npm run build    # Build for production
npm start        # Start production server
npm run lint     # Run linter
```

## Docker

```bash
# Build image
docker build -t smidr-web web/

# Run container
docker run -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  smidr-web

# Full stack
docker-compose -f docker-compose.web.yml up
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Can't connect to API | Verify backend running at port 8080 |
| No agents showing | Check `GET /api/v1/agents` returns data |
| CORS errors | Ensure `NEXT_PUBLIC_API_URL` is correct |
| TypeScript errors | Run `npm install` |
| Port already in use | Change port with `PORT=3001 npm run dev` |

## Data Flow

```
User opens /dashboard
    ↓
useAgents() hook fires
    ↓
apiClient.listAgents() called
    ↓
GET /api/v1/agents request
    ↓
Backend returns agents
    ↓
Dashboard renders agents
    ↓
Re-fetches every 10s
```

## Architecture

```
┌─────────────────┐
│   Frontend      │
│ (Next.js)       │
│ :3000           │
├─────────────────┤
│ - Home page     │
│ - Dashboard     │
│ - Agents list   │
│ - Agent detail  │
└────────┬────────┘
         │
         │ HTTP
         │ /api/v1/*
         ↓
┌─────────────────┐
│   Backend       │
│ (smidr-core)    │
│ :8080           │
├─────────────────┤
│ - Web API       │
│ - Store (SQLite)│
└─────────────────┘
         ↑
         │ HTTP
         │ :9001
         │
┌─────────────────┐
│   Agents        │
│ (Executors)     │
└─────────────────┘
```

## File Locations

```
/Users/schererja/src/github.com/schererja/smidr/
├── web/                      ← Frontend (Next.js)
│   ├── app/
│   ├── lib/api.ts           ← API client
│   ├── lib/hooks.ts         ← React hooks
│   └── package.json
├── cmd/smidr-core/
│   └── main.go              ← Backend entry
├── internal/core/server/
│   ├── web_server.go        ← Web API (port 8080)
│   └── agent_server.go      ← Agent API (port 9001)
└── config.yaml              ← Backend config
```

## Testing

### Check Backend Health

```bash
curl http://localhost:8080/health
```

### Check Frontend

```bash
curl http://localhost:3000
```

### Frontend Console

Open DevTools (F12) and check Console and Network tabs for API calls.
