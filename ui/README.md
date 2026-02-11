# Smidr UI

React frontend for the Smidr system monitoring platform.

## Tech Stack

- **React 19** with TypeScript
- **Vite** for fast dev server and builds
- **React Router** for client-side routing
- **Axios** for API calls
- **Tailwind CSS 4.x** for utility-first styling
- **shadcn/ui** for component library (Badge, Card, Table, Input)
- **lucide-react** for icons
- **recharts** for data visualization (ready for future charts)

## Project Structure

```
ui/
├── src/
│   ├── api/           # API client for control plane integration
│   ├── components/    # Reusable components
│   │   ├── AgentCard.tsx      # Agent card for grid view
│   │   ├── StatCard.tsx       # Stat card for dashboard metrics
│   │   ├── Header.tsx         # Top navigation header
│   │   ├── Sidebar.tsx        # Sidebar navigation
│   │   ├── HealthBadge.tsx    # Health state badge
│   │   └── MetricsCard.tsx    # Metrics display card
│   ├── pages/         # Page components
│   │   ├── SystemList.tsx     # Dashboard/agent list
│   │   └── SystemDetail.tsx   # Agent detail view
│   ├── types/         # TypeScript type definitions
│   ├── ui/            # shadcn/ui components (badge, card, table, input)
│   ├── lib/           # Utility functions (cn for className merging)
│   ├── App.tsx        # Main app with sidebar layout
│   ├── App.css        # Tailwind directives and CSS variables
│   └── main.tsx       # Entry point
├── index.html
├── vite.config.ts     # Vite config with path aliases and proxy
├── tailwind.config.js # Tailwind CSS configuration with shadcn theme
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

### Mobile Responsive Design

**All pages and components are fully responsive:**
- Hamburger menu on mobile (<1024px) with slide-out sidebar
- Responsive grid layouts (stat cards: 2 cols mobile, 4 cols desktop)
- Adaptive table columns (hide less critical columns on mobile)
- Responsive padding, font sizes, and icon sizes
- Touch-friendly tap targets and spacing

See `MOBILE-TESTING.md` for comprehensive mobile testing guide.

### Dashboard (`/systems`)

- **Modern SaaS Dashboard Design**
  - Professional sidebar navigation with icons (responsive hamburger menu on mobile)
  - Clean header with search and notifications
  - Summary stat cards showing total agents, healthy count, issues, and average uptime
  - Grid and table view modes for agent list
  - Color-coded health state badges
  - OS icons for each agent (Linux/Windows/macOS)
  - Hover effects and smooth transitions
  
- **Agent Display**
  - Card-based grid view with agent icons and key metrics
  - Table view with sortable columns (responsive - hides columns on mobile)
  - Search by hostname or agent ID
  - Real-time auto-refresh every 30s
  - Empty states with helpful messages

### System Detail (`/systems/:agentId`)

- **Hero Section**
  - Agent icon and name
  - Large health state badge
  - Key stats: last heartbeat, registration date, baseline count
  
- **Metrics Dashboard**
  - Card-based metric display with icons
  - Color-coded severity indicators
  - Baseline comparison badges (green = normal, orange = warning, red = critical)
  - Visual trend indicators (up/down arrows)
  
- **Baselines Detail**
  - Statistical baselines with sample counts
  - Status icons showing current vs baseline
  - Mean, standard deviation, and range for each metric
  
- **Recent Heartbeats**
  - Timeline-style heartbeat history
  - Formatted metrics display
  - Hover effects for better UX

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

Modern SaaS dashboard design inspired by TailAdmin and other professional admin templates.

### Design System

- **Layout:** Sidebar navigation + header + main content area
- **Color Palette:**
  - Primary: Purple gradient (`from-purple-600 to-purple-800`)
  - Success/Healthy: Green shades (`green-50`, `green-600`)
  - Warning/Degraded: Orange shades (`orange-50`, `orange-600`)
  - Error/Attention: Red shades (`red-50`, `red-600`)
  - Info/Learning: Blue shades (`blue-50`, `blue-600`)
  
- **Components:**
  - Card-based layouts with subtle shadows
  - Rounded corners (`rounded-lg`, `rounded-xl`)
  - Icon-first design with lucide-react icons
  - Hover states and smooth transitions
  - Semantic color usage via CSS variables (shadcn/ui theme)
  
- **Typography:**
  - Clear hierarchy with font weights (400, 500, 600, 700)
  - Consistent spacing (space-y-*, gap-*)
  - Monospace font for IDs and technical data

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
- [x] Modern SaaS dashboard design
- [x] Sidebar navigation and header
- [x] Summary stat cards
- [x] Grid and table view modes
- [x] Enhanced metrics visualization
- [x] Mobile responsive design with hamburger menu
- [x] OS icons (UI ready, needs backend support)
- [ ] User authentication (registration + login)
- [ ] Charts for metric trends (using recharts)
- [ ] Error boundary for better error handling
- [ ] Loading states with skeleton screens
- [ ] Dark mode toggle
- [ ] Unit tests with Vitest
- [ ] Multiple drives support (needs agent + control plane changes)

## Known Issues & Future Enhancements

### OS Detection (Ready, Needs Backend)
The UI is prepared to display OS-specific icons (Linux/Windows/macOS):
- `OSIcon` component created
- `OS` type added to Agent interface
- Agent cards and detail pages display icons

**Requires backend changes:**
1. Agent: Send `runtime.GOOS` in registration request
2. Control plane: Add `OS` field to Agent model
3. Control plane: Return OS in API responses

Currently shows generic server icon for all agents.

### Multiple Drives Support (Future Enhancement)
Currently displays aggregate disk usage for root filesystem.

**To support multiple drives:**
1. Agent: Detect all mounted filesystems, send array of disk metrics
2. Control plane: Update Heartbeat model to store disk array
3. Control plane: Calculate baselines per drive
4. UI: Display each drive separately in MetricsCard

See `.ai-team/decisions/inbox/lambert-os-and-drives.md` for details.

### Bugs Fixed
1. **Healthy fleet percentage**: Fixed calculation that always showed 0%
2. **Metrics count**: Clarified that 4 baselines is correct (excludes uptimeSeconds)
