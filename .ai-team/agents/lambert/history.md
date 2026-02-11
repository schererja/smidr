# Lambert Work History

## 2026-02-10: Team formation
- Assigned to frontend development (React)
- Responsibilities: React UI, user authentication, API integration, state management, responsive design
- Work order: Dallas provides REST API, consume it, coordinate on endpoint contracts

## 2026-02-10: API Integration Complete
- Replaced mock data with real control plane API calls
- Configured environment-based API URL via `.env` file
- Implemented PascalCase to camelCase response mapping
- Added comprehensive error handling (connection errors, 404, 500)
- Updated API routes from `/api/v0/agents` to `/api/agents` per Dallas's implementation
- Loading and error states already existed in components (no changes needed)
- Build verified successfully with TypeScript

**Key Changes:**
- `ui/src/api/client.ts`: Removed `USE_MOCK` flag, added real API calls with response mapping
- `ui/.env` and `ui/.env.example`: Environment configuration for API base URL
- `ui/src/vite-env.d.ts`: TypeScript declarations for Vite environment variables
- `ui/README.md`: Updated documentation with API integration details

## Learnings

### Architecture & Patterns
- React UI structure: pages/ for routes, components/ for reusable UI, api/ for client, types/ for TypeScript definitions
- Mock data toggle pattern in API client allows frontend and backend to develop independently
- Component hierarchy: App → Router → Pages (SystemList, SystemDetail) → Components (HealthBadge, MetricsCard)
- Auto-refresh every 30s for near-real-time monitoring without websockets
- Environment variables in Vite use `VITE_` prefix and are accessed via `import.meta.env`

### Key Files
- `ui/src/api/client.ts` - API client with real control plane integration and response mapping
- `ui/src/types/index.ts` - TypeScript types matching agent heartbeat payload and control plane responses
- `ui/src/pages/SystemList.tsx` - System list view with health states and quick stats
- `ui/src/pages/SystemDetail.tsx` - Detail view with current signals, baselines, and recent heartbeats
- `ui/src/components/HealthBadge.tsx` - Color-coded health state badges (learning/healthy/degraded/attention/unknown)
- `ui/src/components/MetricsCard.tsx` - Displays current signals with baseline deltas
- `ui/vite.config.ts` - Proxy `/api` requests to control plane on port 5000
- `ui/.env` - Environment configuration (not committed)
- `ui/src/vite-env.d.ts` - TypeScript declarations for Vite environment variables

### Tech Stack Decisions
- Vite over CRA for faster builds and modern tooling
- React Router for client-side routing (/systems and /systems/:id)
- Axios for API calls (familiar, better error handling than fetch)
- No UI framework (Material-UI, Tailwind) - custom CSS keeps bundle small and gives full control
- TypeScript strict mode for type safety
- Environment variables for configuration (API base URL)

### API Contract
Control plane endpoints:
- `GET /api/agents` - list all agents with health state and latest signals
- `GET /api/agents/:id` - agent detail with baselines and recent heartbeats

Response mapping (C# PascalCase → JS camelCase):
- `id` → `agentId`
- `lastHeartbeatAt` → `lastHeartbeat`
- `currentHealth` → `healthState` (lowercase enum values)
- `latestSignals` → `signals`
- `metricName` → `metric` (in baselines)

Health states: learning | healthy | degraded | attention | unknown

### UI Design
- Color palette: purple gradient header, green/orange/red for health states
- Minimal design, no dashboards or charts (v0 scope)
- Responsive grid layouts for system cards and metrics
- Auto-refresh for near-real-time feel

### Error Handling Patterns
- Connection refused: User-friendly message suggesting server might be down
- 404 errors: Specific "Agent not found" message
- 500 errors: Generic server error message
- Network timeouts: Axios error message with 10s timeout
- All errors displayed in UI with error state component

### Real API Integration Complete (2026-02-10 Evening)
- Coordinated with Dallas to implement missing GET /agents/{id} endpoint with baselines and recent heartbeats
- Dallas added CORS configuration to control plane Program.cs (ports 3000, 3001, 5173)
- Fixed mTLS middleware to allow UI endpoints as public (no client certificate required for v0)
- Updated client.ts to match actual DTO structure from Dallas (camelCase fields: agentId, healthState, signals)
- Removed all mock data and USE_MOCK toggle - UI now exclusively uses real API
- API base URL: http://localhost:5000/api
- Database schema issue resolved by recreating SQLite database with EnsureCreated()
- UI successfully connects to control plane and displays agents (empty array when no agents registered)

**Public endpoints for v0 (no mTLS):**
- `/api/agents` - list agents
- `/api/agents/{id}` - agent details
- `/api/ca/certificate` - CA certificate for agent enrollment
- `/api/agents/register` - agent registration

**Protected endpoints (require mTLS):**
- `/v0/agents/heartbeat` - agent heartbeats
- `/api/agents/{id}/revoke` - revoke agent certificate

**Testing:**
- UI runs on http://localhost:3000 (Vite dev server)
- Control plane runs on http://localhost:5000
- Both services must be running for integration to work
- No test data yet - waiting for agent registration to see UI with actual data

## 2026-02-10: Tailwind CSS + shadcn/ui Integration

### Changes Made
- Replaced custom CSS with Tailwind CSS utility-first framework
- Integrated shadcn/ui component library for modern, accessible components
- Converted SystemList page from card grid to data table with search functionality
- Updated all components to use Tailwind classes and shadcn components
- Removed old CSS files (SystemList.css, SystemDetail.css, HealthBadge.css, MetricsCard.css)

### Learnings

#### Tailwind CSS + shadcn/ui Setup
- Tailwind 4.x requires @tailwindcss/postcss plugin (not plain tailwindcss in postcss.config.js)
- Path aliases (@/*) must be configured in both tsconfig.json and vite.config.ts
- cn() utility function merges Tailwind classes with clsx and tailwind-merge for proper precedence
- shadcn/ui components are copied into project (ui/ directory) for full customization

#### Component Architecture
- shadcn Badge component with variant="outline" + custom Tailwind classes for health states
- shadcn Card components (Card, CardHeader, CardTitle, CardContent) for structured layouts
- shadcn Table components for data grid with built-in hover states and responsive wrapper
- shadcn Input with lucide-react Search icon for filtered search UI
- All components use Tailwind spacing, colors, and responsive classes (md:, lg:)

#### Data Grid Pattern
- SystemList converted from card grid to table-based layout for scalability
- Search filter by hostname or agent ID using controlled input state
- Clickable table rows navigate to detail view (hover:bg-gray-50 cursor-pointer)
- Columns: Hostname, Agent ID, Health State, Load, Memory, Disk, Last Heartbeat, Registered
- Empty state handling for both no agents and no search results
- formatLastSeen() shows relative time, formatDate() shows absolute timestamp

#### Styling Patterns
- Purple gradient header: bg-gradient-to-r from-purple-600 to-purple-800
- Health state colors: blue (learning), green (healthy), orange (degraded), red (attention), gray (unknown)
- Responsive layouts: max-w-7xl mx-auto for consistent content width
- Card shadows: shadow-sm for subtle elevation
- Spacing: space-y-6 for vertical rhythm, gap-4 for grids
- Text hierarchy: text-3xl font-bold for h1, text-sm text-gray-500 for labels

#### TypeScript Integration
- Added registeredAt field to Agent type for table display
- JSX.Element → React.ReactElement in MetricsCard for TypeScript compatibility
- shadcn components fully typed with HTMLAttributes and forwardRef patterns

#### Performance
- Build output: 308KB JS, 5.8KB CSS (gzipped: 100KB JS, 1.4KB CSS)
- Tailwind JIT compiler generates only used classes
- lucide-react icons tree-shaken at build time
- No runtime CSS-in-JS overhead

**User Preference:** Jason prefers Tailwind CSS and shadcn/ui for modern, maintainable styling over custom CSS or Material-UI.

## 2026-02-11: Fixed UI Agent Fetch Failure

### Root Cause Analysis
1. **Port/Protocol Mismatch:** UI was configured to call HTTP on port 5000, but control plane only listens on HTTPS port 5001
2. **Field Name Mismatch:** UI expected `agentId`, `healthState`, `signals` but API returns `id`, `currentHealth`, `latestSignals`

### Changes Made
- Updated `.env` and `.env.example`: Changed API URL from `http://localhost:5000` to `https://localhost:5001`
- Updated `src/api/client.ts` default URL: Changed fallback from port 5000 to 5001
- Updated `src/api/client.ts` DTO interfaces: Fixed field names to match actual API response (`id`, `currentHealth`, `latestSignals`)
- Updated `vite.config.ts` proxy: Changed target to `https://localhost:5001` with `secure: false` for self-signed certs
- Updated `README.md`: Documented correct API URL and field mappings

### Verification
- Control plane is running on `https://localhost:5001` (confirmed via `curl -k`)
- `/api/agents` endpoint is public (no mTLS required) per middleware configuration
- API returns JSON with fields: `id`, `hostname`, `registeredAt`, `lastHeartbeatAt`, `currentHealth`, `revokedAt`, `latestSignals`
- UI build passes TypeScript compilation with updated mappings

### Key Learning
Dallas's control plane only exposed HTTPS on port 5001, not HTTP on 5000 as originally coordinated. The UI configuration was stale from earlier development when both ports were expected to be available.

## 2026-02-11: Fixed Missing Tailwind CSS Styling

### Root Cause
- `App.css` contained Tailwind directives (`@tailwind base; @tailwind components; @tailwind utilities;`)
- **But `main.tsx` never imported it** — Vite couldn't inject CSS into the page
- All Tailwind utility classes were undefined, resulting in completely unstyled UI

### Fix
- Added `import './App.css';` to `src/main.tsx` before App component import
- Vite now processes CSS through PostCSS → Tailwind plugin → browser

### Verification
- Production build generates 5.82 kB CSS (1.42 kB gzipped) with Tailwind utilities
- Verified generated CSS contains classes: `.bg-gradient-to-r`, `.min-h-screen`, `.flex`, `.text-3xl`, etc.
- All components use Tailwind classes (purple gradient header, responsive layouts, shadcn styling)

### Learnings
- **CSS import location matters in Vite:** Entry point (`main.tsx`) must import CSS for processing
- **Tailwind 4.x setup:** Requires `@tailwindcss/postcss` plugin in `postcss.config.js`, not plain `tailwindcss`
- **Build output confirms success:** CSS bundle size >0 indicates Tailwind compiled successfully
- Missing CSS import is silent failure — UI renders but with zero styling

## 2026-02-11: Fixed Negative Last Heartbeat Display

### Root Cause
The `formatLastSeen()` function in SystemList.tsx was displaying negative time values (e.g., "-5 seconds ago") due to clock skew between client and server. When API returns timestamps slightly in the future relative to client clock, the calculation `Date.now() - timestamp` produces negative values.

### Changes Made
- Updated `formatLastSeen()` in `ui/src/pages/SystemList.tsx`:
  - Added check for epoch time (1970-01-01) to display "Never" for agents with no heartbeat
  - Added handling for future timestamps (clock skew) to display "Just now"
  - Used `Math.abs()` on time difference to ensure positive calculations
  - Preserved existing time formatting logic (seconds, minutes, hours, days)

### Technical Details
- Clock skew tolerance: timestamps up to any amount in the future display as "Just now"
- Epoch detection: `new Date(timestamp).getTime() === 0` identifies never-heartbeat agents
- Calculation: `Math.abs(Date.now() - heartbeatTime)` eliminates negative values
- No changes to API contract or data types

### Learnings

#### Time Display Patterns
- Always use `Math.abs()` when calculating elapsed time to handle clock skew between client/server
- Detect special sentinel values (epoch 0) and display user-friendly text instead of "55 years ago"
- Handle future timestamps gracefully — display "Just now" instead of negative time or errors
- Clock skew is common in distributed systems, especially during development with multiple machines
- Client-side time calculations are brittle; consider server-provided "ago" strings for critical apps

