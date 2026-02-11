# Smidr UI

React frontend for the Smidr system monitoring platform.

## Tech Stack

- **React 19** with TypeScript
- **Vite** for fast dev server and builds
- **React Router** for client-side routing
- **Axios** for API calls
- **Tailwind CSS** for styling
- **shadcn/ui** for component library (Badge, Card, Table, Input)
- **lucide-react** for icons

## Project Structure

```
ui/
├── src/
│   ├── api/           # API client for control plane integration
│   ├── components/    # Reusable components (HealthBadge, MetricsCard)
│   ├── pages/         # Page components (SystemList, SystemDetail)
│   ├── types/         # TypeScript type definitions
│   ├── ui/            # shadcn/ui components (badge, card, table, input)
│   ├── lib/           # Utility functions (cn for className merging)
│   ├── App.tsx        # Main app component with routing
│   ├── App.css        # Tailwind directives
│   └── main.tsx       # Entry point
├── index.html
├── vite.config.ts     # Vite config with path aliases and proxy
├── tailwind.config.js # Tailwind CSS configuration
├── postcss.config.js  # PostCSS with @tailwindcss/postcss
└── package.json
```

## Development

### Prerequisites

- Node.js 18+ and npm

### Install Dependencies

```bash
npm install
```

### Run Dev Server

```bash
npm run dev
```

Runs on http://localhost:3000

### Build for Production

```bash
npm run build
```

Output in `dist/` directory.

## Features

### System List (`/systems`)

- View all registered agents
- Color-coded health states (learning, healthy, degraded, attention, unknown)
- Quick stats: load, memory, disk usage
- Last seen timestamp
- Auto-refresh every 30s

### System Detail (`/systems/:agentId`)

- Detailed view of single agent
- Current signals with baseline comparisons
- Learned baselines (mean, std dev, range)
- Recent heartbeats history
- Auto-refresh every 30s

## API Integration

The UI now connects to the **real control plane API** at `https://localhost:5001`.

### Configuration

Create a `.env` file (use `.env.example` as template):

```bash
VITE_API_BASE_URL=https://localhost:5001
```

The API base URL can be changed via environment variable for different deployment environments.

### API Endpoints

- `GET /api/agents` - List all agents with latest signals and health state
- `GET /api/agents/:id` - Get agent detail with baselines and recent heartbeats

### Response Mapping

The control plane returns C# PascalCase responses (serialized to camelCase). The client automatically maps to UI types:
- `id` → `agentId`
- `lastHeartbeatAt` → `lastHeartbeat`
- `currentHealth` → `healthState`
- `latestSignals` → `signals`
- `metricName` → `metric` (in baselines)

### Error Handling

The client provides user-friendly error messages:
- Connection refused: "Cannot connect to control plane API. Is the server running?"
- 404: "Agent not found"
- 500: "Server error loading agents/agent details"
- Network errors: Specific axios error messages

## Health States

| State      | Color  | Meaning                                    |
|------------|--------|-------------------------------------------|
| learning   | Blue   | Agent in baseline learning phase          |
| healthy    | Green  | All metrics within baseline thresholds    |
| degraded   | Orange | Some metrics exceeding warning thresholds |
| attention  | Red    | Critical thresholds exceeded              |
| unknown    | Gray   | Missing heartbeats or insufficient data   |

## Styling

Clean, minimal design with no UI framework dependencies. All CSS is custom and scoped to components.

Color palette:
- Primary: Purple gradient (`#667eea` to `#764ba2`)
- Success: Green (`#2e7d32`)
- Warning: Orange (`#e65100`)
- Error: Red (`#c62828`)
- Info: Blue (`#1565c0`)

## Integration Points

### For Dallas (Control Plane)

The UI expects these response formats:

**GET /api/v0/agents:**
```json
[
  {
    "agentId": "uuid",
    "hostname": "string",
    "lastHeartbeat": "ISO8601 timestamp",
    "healthState": "learning|healthy|degraded|attention|unknown",
    "signals": {
      "uptimeSeconds": 86400,
      "loadAverage1m": 1.23,
      "memoryUsedPct": 45.6,
      "diskUsedPct": 67.8,
      "processCount": 156
    }
  }
]
```

**GET /api/v0/agents/:id:**
```json
{
  "agentId": "uuid",
  "hostname": "string",
  "lastHeartbeat": "ISO8601 timestamp",
  "healthState": "learning|healthy|degraded|attention|unknown",
  "signals": { ... },
  "baselines": [
    {
      "metric": "loadAverage1m",
      "mean": 1.2,
      "stdDev": 0.3,
      "min": 0.5,
      "max": 2.1
    }
  ],
  "recentHeartbeats": [
    {
      "agentId": "uuid",
      "timestamp": "ISO8601 timestamp",
      "uptimeSeconds": 86400,
      "loadAverage1m": 1.23,
      "memoryUsedPct": 45.6,
      "diskUsedPct": 67.8,
      "processCount": 156
    }
  ]
}
```

## Development Workflow

### Starting the Full Stack

1. Start the control plane API:
   ```bash
   cd control-plane
   dotnet run
   ```

2. Start the UI dev server:
   ```bash
   cd ui
   npm run dev
   ```

3. Access UI at http://localhost:3000

The Vite dev server automatically proxies `/api` requests to the control plane.

## Next Steps

- [x] Integration with real control plane API
- [ ] User authentication (registration + login)
- [ ] Error boundary for better error handling
- [ ] Loading states with skeleton screens
- [ ] Responsive design improvements for mobile
- [ ] Dark mode toggle
- [ ] Unit tests with Vitest
