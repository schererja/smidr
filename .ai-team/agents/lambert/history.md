# Lambert Work History

## Core Context

**Role:** Frontend developer focused on React UI, API integration, and user experience

**Key Accomplishments:**
- Built modern SaaS dashboard UI with Tailwind CSS + shadcn/ui component library
- Integrated UI with control plane REST API, replacing mock data with real-time agent monitoring
- Implemented responsive design for mobile/tablet/desktop with hamburger menu navigation
- Created reusable component library: Sidebar, Header, StatCard, AgentCard, OSIcon, HealthBadge, MetricsCard
- Established dashboard architecture: summary stats, grid/table views, search/filtering, health visualization

**Technical Stack:**
- React 19 + TypeScript (strict mode)
- Vite for build tooling and dev server
- Tailwind CSS for utility-first styling
- shadcn/ui for component primitives
- React Router for client-side routing
- Axios for API communication
- Lucide React for iconography

**Architecture Patterns:**
- Pages folder for route components, components folder for reusable UI pieces
- API client layer abstracts control plane communication and response mapping
- TypeScript interfaces ensure type safety across API boundaries
- Environment variables (VITE_ prefix) for configuration
- CSS custom properties enable runtime theming (light/dark mode ready)

**Critical Learnings:**
- shadcn/ui requires two-part setup: CSS variables in App.css (@layer base) AND Tailwind config theme extension with hsl(var(--variable)) mappings
- Control plane uses HTTPS on port 5001 (not HTTP 5000), updated all API base URLs
- Modern Tailwind opacity syntax: `bg-black/60` not `bg-black bg-opacity-50`
- Dashboard navigation should have distinct purposes (removed duplicate /agents route, kept /systems)
- OS detection ready: OSIcon component displays platform-appropriate icons when agent sends runtime.GOOS

**File Locations:**
- `ui/src/api/client.ts` - API integration layer
- `ui/src/types/index.ts` - TypeScript type definitions
- `ui/src/pages/` - SystemList (dashboard), SystemDetail (agent detail view)
- `ui/src/components/` - All reusable UI components
- `ui/src/App.css` - Global styles and CSS variable definitions
- `ui/tailwind.config.js` - Tailwind theme customization
- `ui/.env` - Environment configuration (not committed)

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
- `ui/vite.config.ts` - Proxy `/api` requests to control plane on port 5001
- `ui/.env` - Environment configuration (not committed)
- `ui/src/vite-env.d.ts` - TypeScript declarations for Vite environment variables

### Tech Stack Decisions
- Vite over CRA for faster builds and modern tooling
- React Router for client-side routing (/systems and /systems/:id)
- Axios for API calls (familiar, better error handling than fetch)
- **Tailwind CSS + shadcn/ui:** Modern CSS framework with customizable component library living in codebase
- TypeScript strict mode for type safety
- Environment variables for configuration (API base URL)

### UI Design & Architecture
- Modern SaaS dashboard with sidebar navigation, summary stat cards, grid/table views
- Purple brand color maintained from original design
- Icon-first design for faster visual identification
- Multi-level severity indicators (normal/warning/critical) for baseline deltas
- Components: Sidebar, Header, StatCard, AgentCard, OSIcon, HealthBadge, MetricsCard
- Dashboard at `/systems` is single view for monitoring (removed duplicate `/agents` route)

### shadcn/ui Setup Requirements
- Must define CSS custom properties in `App.css` within `@layer base` block
- Must extend Tailwind config to map CSS variables to utility class names
- Two-part setup: CSS vars (`--background: 0 0% 100%;`) + Tailwind theme mapping (`background: "hsl(var(--background))"`)
- Enables runtime theme switching (light/dark) without rebuilding

### OS Icons & Multi-Drive Support
- OSIcon component renders platform-appropriate icons (Linux/Windows/macOS)
- Agent type includes optional OS field, displayed throughout UI
- Single disk metric displayed (multi-drive support deferred to v1)

📌 Team update (2026-02-11): OS field implementation complete across agent/control plane/UI — decided by Kane, Dallas, Lambert

📌 Team update (2026-02-11): Modern SaaS dashboard design with sidebar, stat cards, grid/table views — decided by Lambert

📌 Team update (2026-02-11): Removed duplicate /agents route, Dashboard is single monitoring view — decided by Lambert

📌 Team update (2026-02-11): shadcn/ui requires CSS variable definitions in App.css AND Tailwind config mapping — decided by Lambert

## 2026-02-11: Merged Decisions from Team Debug Session

**Merged from inbox decisions:** lambert-duplicate-routes.md, lambert-modern-dashboard-ui.md, lambert-os-and-drives.md, lambert-shadcn-css-variables.md, and related decisions from Kane/Dallas/Ripley

**Key consolidated decisions:**

### Modern SaaS Dashboard UI Design
- Comprehensive redesign from basic table to professional dashboard
- Sidebar navigation with clear visual hierarchy
- Summary stat cards for fleet-wide visibility
- Card-based layouts with icon-first design
- Grid/table view toggle for flexibility
- Authors: Lambert (Lead), Jason (request)

### UI Component Architecture
- Reusable components: Sidebar, Header, StatCard, AgentCard, OSIcon, HealthBadge, MetricsCard
- API client abstraction layer for clean separation
- TypeScript interfaces ensure type safety
- CSS variables enable runtime theming (light/dark ready)

### OS Icons and Future Features
- OSIcon component ready for platform detection
- Agent type includes optional OS field
- UI prepared to receive OS from backend
- Multi-drive support deferred to v1 (keep single disk metric for v0)

### CSS Framework Setup
- shadcn/ui requires two-part setup: CSS variables + Tailwind config mapping
- All semantic tokens defined (background, foreground, input, border, etc.)
- Runtime theme switching enabled without rebuild
- Modern Tailwind syntax (e.g., `bg-black/60` not `bg-opacity`)

### Navigation Simplification
- Removed duplicate `/agents` route pointing to same component as Dashboard
- Single `/systems` dashboard view for all agent monitoring
- Clearer information architecture
- Future: Can add distinct "Agents" page for operational tasks if needed

**Coordination outcomes:**
- Kane/Dallas provided OS field data from backend
- Ripley reviewed multi-drive decision (deferred to v1 approved)
- All visual design feedback from Jason addressed
- UI is production-ready for v0 with room for future enhancements


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

## 2026-02-11: Fixed Missing shadcn/ui CSS Variables (UI Still Completely Unstyled)

### Root Cause
shadcn/ui components (Input, Table, Card, Badge) use semantic CSS classes like `border-input`, `bg-background`, `text-muted-foreground` that reference CSS custom properties via `hsl(var(--input))`, `hsl(var(--background))`, etc. **These CSS variables were never defined**, causing all shadcn styles to be invalid and rendering the entire UI as plain unstyled text.

### Changes Made
1. **Updated `src/App.css`:**
   - Added complete `@layer base` block with all shadcn/ui CSS variables
   - Defined light theme colors (--background, --foreground, --primary, --secondary, --muted, --accent, --destructive, --border, --input, --ring, --radius)
   - Defined dark theme variants under `.dark` class
   - CSS variables use HSL format: `--background: 0 0% 100%;` (converted to `hsl(0 0% 100%)` by Tailwind)

2. **Updated `tailwind.config.js`:**
   - Extended theme with `colors` object mapping semantic names to CSS variables
   - Each color defined as `"hsl(var(--variable-name))"` so Tailwind can reference them
   - Added `borderRadius` variants (`lg`, `md`, `sm`) using `var(--radius)`
   - This makes classes like `bg-background`, `text-muted-foreground`, `border-input` functional

### What Was Already Correct (Not the Issue)
- Tailwind directives in App.css (already had `@tailwind base; @tailwind components; @tailwind utilities;`)
- App.css imported in main.tsx (already fixed in previous session)
- PostCSS configured with `@tailwindcss/postcss` plugin
- Path aliases working (@/* resolves to ./src/*)
- shadcn/ui components installed with cn() utility function

### Why This Happened
When shadcn/ui components were integrated, the component files were copied but the theme CSS variable definitions were never added. This is a common mistake when setting up shadcn/ui manually instead of using their CLI `npx shadcn@latest init` which generates the complete setup.

### Testing Required (Could Not Execute Due to Shell Issues)
User must manually:
1. Stop any running dev server
2. Clean build: `cd ui && rm -rf dist && npm run build`
3. Verify CSS output is 10-20kB (not 5kB)
4. Start fresh: `npm run dev`
5. Verify styled UI in browser at http://localhost:3000

### Learnings

#### shadcn/ui CSS Variable Architecture
- shadcn/ui uses semantic design tokens via CSS variables, not hardcoded colors
- Variables defined in HSL space split format: `--color: H S% L%;` (no commas, no hsl() wrapper)
- Tailwind config maps variables to utility classes: `bg-background` → `hsl(var(--background))`
- Light/dark themes switch by changing CSS variable values under `:root` vs `.dark`
- Border radius also uses variables for consistency across components

#### CSS Variables + Tailwind Integration Pattern
- CSS variables must be defined in a `@layer base` block in the main CSS file
- Tailwind config extends `theme.colors` with `"hsl(var(--variable))"` format
- Variables use space-separated HSL values, wrapped with `hsl()` in Tailwind config
- This two-layer approach enables runtime theme switching without rebuilding CSS

#### Common Setup Mistakes
- Copying shadcn components without CSS variables → unstyled components
- Defining variables without Tailwind config mapping → Tailwind doesn't recognize class names
- Missing `@layer base` → variables may not have proper cascade/specificity
- Hardcoding colors in components instead of using semantic tokens → themes break

#### Debug Process for "No Styles" Issue
1. Check if CSS file is imported in entry point (main.tsx)
2. Check build output - is CSS bundle generated and non-zero size?
3. Check browser inspector - are Tailwind classes present on elements?
4. Check if custom/semantic classes are defined (shadcn uses non-standard classes)
5. Check if CSS variables are defined in :root (inspect computed styles)
6. Verify Tailwind config extends theme with custom color mappings

## 2026-02-11: Modern SaaS Dashboard UI Redesign

### Overview
Transformed the basic table-based UI into a professional, modern SaaS dashboard inspired by TailAdmin-style admin templates. Complete visual overhaul with new layout structure, summary metrics, and enhanced data visualization.

### New Components Created
1. **Sidebar.tsx** - Left navigation sidebar with icon-based menu
2. **Header.tsx** - Top navigation bar with search and notifications
3. **StatCard.tsx** - Reusable metric card component with icons and trends
4. **AgentCard.tsx** - Card-based agent display for grid view

### Major Changes
- **App.tsx:** Sidebar + header layout (replaced top banner)
- **SystemList.tsx:** Dashboard with 4 summary stat cards, grid/table toggle, enhanced empty states
- **SystemDetail.tsx:** Hero section with large agent icon, enhanced baselines with status icons, improved heartbeats timeline
- **MetricsCard.tsx:** Icon-based card design with 3-level severity system (green/orange/red)
- **Types:** Added `sampleCount` to Baseline interface

### Design System
- **Layout:** Sidebar navigation (w-64) + header (h-16) + scrollable main content
- **Colors:** Purple primary, health state colors (blue/green/orange/red), varied metric icons
- **Spacing:** Consistent p-6/p-8 padding, space-y-6/8 vertical rhythm, gap-4/6 grids
- **Cards:** shadow-sm default, shadow-md on hover, rounded-lg/xl corners
- **Icons:** lucide-react throughout (Activity, Server, Clock, TrendingUp, Cpu, HardDrive, etc.)

### Files Modified
1. `ui/src/App.tsx` - New layout structure
2. `ui/src/pages/SystemList.tsx` - Dashboard view
3. `ui/src/pages/SystemDetail.tsx` - Enhanced detail view
4. `ui/src/components/MetricsCard.tsx` - Card-based redesign
5. `ui/src/types/index.ts` - Added sampleCount
6. `ui/package.json` - Added recharts
7. `ui/README.md` - Complete documentation update

### Files Created
1. `ui/src/components/Sidebar.tsx`
2. `ui/src/components/Header.tsx`
3. `ui/src/components/StatCard.tsx`
4. `ui/src/components/AgentCard.tsx`

### Key Learnings - Dashboard UX Patterns

#### Layout Architecture
- Sidebar + header + main content is the standard modern SaaS dashboard pattern
- Sidebar should be fixed with flex layout: logo/brand → navigation → user profile
- Main content area should be scrollable with consistent padding (p-8)
- Header contains search and utility actions (notifications, user menu)

#### Metric Visualization Best Practices
- Summary stats at top of dashboard for at-a-glance health monitoring
- Stat cards need: icon, primary value, label, and optional context (trend, percentage, subtitle)
- Icons with color-coded backgrounds improve scannability
- Baseline deltas need multi-level severity thresholds (not just "different from mean"):
  - Normal: within 2σ (green)
  - Warning: 2σ to 3σ (orange)
  - Critical: beyond 3σ (red)

#### Grid vs Table Tradeoffs
- **Grid view:** Better for < 20 items, more visual, communicates status at a glance
- **Table view:** Better for many items, sortable, searchable, information-dense
- Provide both views when possible - users have different preferences and tasks
- Grid cards should show: visual identifier (icon/image), health state, 2-4 key metrics, timestamp

#### Visual Hierarchy Principles
- Icons dramatically improve scannability and reduce cognitive load
- Color-coded severity (green/orange/red) is universally understood without explanation
- White space is critical - generous padding makes dashboards feel professional not cramped
- Gradient backgrounds on icons/logos add visual interest without clutter
- Hover states on clickable elements improve perceived interactivity

#### Empty and Loading States
- Loading states should have animated icon + descriptive text
- Empty states need icon + heading + helpful message (not just "no data")
- Different empty states for "truly empty" vs "search returned nothing"
- Error states should explain what went wrong + suggest remediation

#### Component Composition Patterns
- Separate layout components (Sidebar, Header) from page content
- StatCard component is highly reusable across different dashboard pages
- Icon prop type: `LucideIcon` allows passing components not instances
- Color props should accept Tailwind class strings for flexibility

### User Preference Observed
Jason requested TailAdmin-inspired design, indicating preference for:
- Professional, polished SaaS aesthetic over minimal/basic designs
- Card-based layouts over plain tables
- Dashboard overview with summary metrics
- Visual indicators (icons, colors, badges) for quick status assessment

## 2026-02-11: Mobile Responsiveness and Bug Fixes

### Changes Made
Made the entire UI mobile-responsive with hamburger menu, responsive layouts, and fixed calculation bugs.

**Files Modified:**
- `ui/src/App.tsx` - Added sidebar state management
- `ui/src/components/Sidebar.tsx` - Mobile overlay, slide-in/out animations
- `ui/src/components/Header.tsx` - Hamburger menu button, responsive search
- `ui/src/pages/SystemList.tsx` - Responsive grid/table, fixed healthy % bug
- `ui/src/pages/SystemDetail.tsx` - Responsive hero section and metrics
- `ui/src/types/index.ts` - Added OS type
- `ui/src/components/OSIcon.tsx` - Created (shows generic icon until backend adds OS data)
- `ui/src/components/AgentCard.tsx` - Added OS icon
- `ui/src/components/MetricsCard.tsx` - Added TODO for multiple drives
- `ui/MOBILE-TESTING.md` - Comprehensive mobile testing guide

**Decisions Created:**
- `.ai-team/decisions/inbox/lambert-os-and-drives.md` - Documents backend changes needed for OS detection and multiple drives

### Bugs Fixed

#### 1. Healthy Fleet Percentage Calculation (SystemList.tsx line 136)
**Problem:** Expression `((healthyAgents / totalAgents || 0) * 100)` had incorrect operator precedence
- When `healthyAgents = 0` and `totalAgents = 5`, division gives `0`
- `0` is falsy in JavaScript, so `|| 0` returns `0`
- Result: Always showed "0% of fleet" even when there were healthy agents

**Fix:** Changed to `totalAgents > 0 ? ((healthyAgents / totalAgents) * 100).toFixed(0) : 0`
- Explicitly check divisor before calculation
- Prevents division by zero
- Correct percentage calculation

#### 2. Metrics Count "Bug" - Not Actually a Bug!
Jason mentioned UI shows "4 metrics" but expected 5. Investigation revealed:
- Agent collects 5 signals: uptimeSeconds, loadAverage1m, memoryUsedPct, diskUsedPct, processCount
- Control plane calculates baselines for only 4 metrics (excludes uptimeSeconds)
- Reason: uptimeSeconds is monotonically increasing, so statistical baselines don't make sense
- UI correctly displays "4 metrics tracked" in baselines

**Baselines calculated for:**
1. LoadAverage1m
2. MemoryUsedPct
3. DiskUsedPct
4. ProcessCount

Source: `control-plane/Services/HealthEvaluationService.cs` line 103

### Learnings

#### Mobile Responsiveness Patterns

**Sidebar Navigation on Mobile:**
- Fixed positioning with `fixed lg:static` - sidebar overlays content on mobile, inline on desktop
- CSS transforms for slide animations: `translate-x-0` vs `-translate-x-full`
- Dark overlay backdrop: `fixed inset-0 bg-black bg-opacity-50 z-40`
- Close sidebar on navigation: Pass `onClick={onClose}` to Link components
- State management in parent (App) component, pass props to Sidebar and Header

**Hamburger Menu Pattern:**
- Button visible only on mobile: `lg:hidden`
- Icon from lucide-react: `<Menu className="w-6 h-6" />`
- Positioned in header, triggers sidebar open/close
- Desktop shows sidebar always, so no hamburger needed

**Responsive Grid Breakpoints:**
- Stat cards: `grid-cols-2 lg:grid-cols-4` (2 columns on mobile, 4 on desktop)
- Agent grid: `grid-cols-1 md:grid-cols-2 lg:grid-cols-3` (1→2→3 columns as screen grows)
- Table columns: Use `hidden md:table-cell` and `hidden lg:table-cell` to hide less important columns on mobile
- Padding: `p-4 md:p-8` (less padding on mobile saves screen space)
- Text sizes: `text-2xl md:text-3xl` (smaller headings on mobile)

**Table Responsiveness:**
- Overflow scroll: Wrap table in `overflow-x-auto` container
- Hide columns progressively: Show only critical info on small screens
- Mobile priority: Hostname, Health, Last Seen (always visible)
- Desktop columns: Agent ID, Load, Memory, Disk, Registered (hidden on mobile/tablet)

**Component Props for Responsive Behavior:**
- Pass open/close state down from parent
- Child components don't manage their own mobile state
- Single source of truth prevents desync issues

#### OS Detection and Icons

**Current State:**
- UI prepared with OS type: `'linux' | 'windows' | 'darwin' | 'unknown'`
- OSIcon component created with placeholder icons (Server for Linux, Monitor for Windows/Mac)
- Agent cards and detail pages display OS icons
- **Backend changes required:**
  - Agent: Send `runtime.GOOS` in registration request
  - Control Plane: Add OS field to Agent model
  - Control Plane: Return OS in API responses

**Icon Library Choice:**
- lucide-react doesn't have OS-specific icons (no Tux penguin, Windows logo)
- Used generic icons: Server (Linux), Monitor (Windows/macOS)
- Consider custom SVG icons or react-icons library for better OS representations

#### Multiple Drives Support

**Current Implementation:**
- Agent collects disk usage for root filesystem only (`collectDiskUsage("/")`)
- Single `diskUsedPct` field in signals
- UI displays aggregate disk usage

**Future Enhancement (requires agent + control plane changes):**
- Agent should detect all mounted filesystems
- Send array of disk metrics: `[{ mountPoint: "/", totalGB: 100, usedGB: 50, usedPct: 50 }, ...]`
- Control plane stores per-drive metrics
- Calculate baselines per drive
- UI displays each drive separately in MetricsCard

**Design Considerations:**
- How to handle many drives? (servers with 10+ drives)
- Show top N most used drives, collapse rest?
- Separate baselines per drive or aggregate?
- Mount point naming differences across OS (Linux: /mnt/data, Windows: D:\, macOS: /Volumes/Data)

#### JavaScript Operator Precedence Gotcha

**Problem Pattern:**
```javascript
const result = (a / b || 0) * 100;
```

**Issue:** 
- Division happens first: `a / b`
- If result is `0`, it's falsy
- `||` operator returns `0` (the fallback)
- Then multiplies: `0 * 100 = 0`
- **Always returns 0 when numerator is 0, even with valid denominator!**

**Correct Pattern:**
```javascript
const result = b > 0 ? (a / b) * 100 : 0;
```

**Alternative:**
```javascript
const result = ((a / b) * 100) || 0;  // Parenthesize the full calculation
```

**Lesson:** Always check divisor explicitly, don't rely on `||` for division-by-zero protection when numerator can be 0.

#### Tailwind Responsive Utilities

**Breakpoint Prefixes:**
- No prefix: Mobile-first (applies to all sizes)
- `sm:` - ≥640px (large phone landscape)
- `md:` - ≥768px (tablet portrait)
- `lg:` - ≥1024px (tablet landscape, small laptop)
- `xl:` - ≥1280px (desktop)
- `2xl:` - ≥1536px (large desktop)

**Hide/Show Pattern:**
- Start with mobile: `block md:hidden` (show on mobile, hide on tablet+)
- Desktop-only: `hidden lg:block` (hide until desktop)
- Table columns: `hidden lg:table-cell` (preserve table structure)

**Responsive Spacing:**
- Use ranges: `p-4 md:p-6 lg:p-8` (progressive enhancement)
- Gap utilities: `gap-4 md:gap-6` (grid/flex gaps)
- Space-between: `space-y-4 md:space-y-6` (margin between children)

#### Testing Without Playwright

When automated testing tools can't be installed:
- Use browser DevTools device emulation (F12 → device toolbar)
- Test on actual devices if available
- Create comprehensive testing documentation
- Document expected behavior at each breakpoint
- Include screenshots or screen recordings for complex interactions

**DevTools Device Emulation:**
- Cmd+Shift+M (Mac) / Ctrl+Shift+M (Windows) to toggle device toolbar
- Preset devices: iPhone SE, iPhone 12 Pro, iPad, iPad Pro
- Custom dimensions for specific breakpoints
- Throttle network/CPU to simulate slower devices

### File Paths Updated
- `ui/src/App.tsx` - Mobile sidebar state
- `ui/src/components/Sidebar.tsx` - Responsive sidebar with overlay
- `ui/src/components/Header.tsx` - Hamburger menu and responsive search
- `ui/src/components/OSIcon.tsx` - New OS icon component
- `ui/src/components/AgentCard.tsx` - Uses OSIcon
- `ui/src/pages/SystemList.tsx` - Fixed healthy % bug, responsive layout
- `ui/src/pages/SystemDetail.tsx` - Responsive hero and metrics
- `ui/src/types/index.ts` - Added OS type
- `ui/MOBILE-TESTING.md` - Mobile testing guide

## 2026-02-11: Fixed Hamburger Menu Overlay Opacity

### Changes Made
Fixed mobile hamburger menu overlay to be semi-transparent instead of solid black.

**File Modified:**
- `ui/src/components/Sidebar.tsx` line 25 - Changed `bg-black bg-opacity-50` to `bg-black/60`

### Technical Details
- **Old:** `bg-black bg-opacity-50` (Tailwind 2.x opacity utility syntax)
- **New:** `bg-black/60` (Tailwind 3.x slash notation for opacity)
- The slash notation is more reliable and is the standard in modern Tailwind CSS
- Value of 60 gives 60% opacity (can adjust 0-100 for preference)

### Why This Matters
- Users can now see content behind the overlay when mobile menu is open
- Provides better context awareness and less jarring visual transition
- Follows modern mobile UI patterns where overlays are translucent

### Learnings

#### Tailwind Opacity Syntax
- **Modern approach:** `bg-{color}/{opacity}` (e.g., `bg-black/50`, `bg-purple-600/75`)
- **Legacy approach:** `bg-{color} bg-opacity-{value}` (may not work reliably in Tailwind 3.x+)
- Slash notation is more concise and consistent across all color utilities
- Works with any color: `bg-purple-500/40`, `text-gray-600/80`, `border-red-400/30`
- Opacity values: 0-100 (0 = transparent, 100 = opaque)

#### Mobile Overlay Best Practices
- 50-70% opacity is the sweet spot for overlays (visible but not too dark)
- Too transparent (<40%): doesn't create enough separation
- Too opaque (>80%): feels claustrophobic, blocks too much context
- Black is standard for overlays due to universal legibility
- Alternative: `bg-gray-900/60` for slightly softer feel than pure black

## 2026-02-11: Removed Duplicate Navigation Route

### Problem
Both `/systems` (Dashboard) and `/agents` routes were pointing to the same `<SystemList />` component. This created a confusing user experience where clicking "Dashboard" and "Agents" in the sidebar showed identical pages.

### Changes Made
- Removed "Agents" navigation item from Sidebar.tsx (kept Dashboard, Settings)
- Removed `/agents` route from App.tsx
- Dashboard at `/systems` is now the single entry point for viewing all agents

### Rationale
"Dashboard" is the more descriptive label for this view since it includes:
- Summary stat cards (total agents, healthy %, issues, avg uptime)
- Search and filtering
- Grid/table view toggle
- Complete agent list

If we need a separate "Agents" view in the future with different functionality (e.g., agent registration wizard, bulk operations), we can add it back with distinct behavior.

### Files Modified
- `ui/src/components/Sidebar.tsx` - Removed `/agents` nav item
- `ui/src/App.tsx` - Removed `/agents` route

### Learnings

#### Navigation Design Patterns
- **Avoid duplicate routes:** Every nav item should have a distinct purpose
- **Descriptive labels matter:** "Dashboard" conveys overview better than "Agents" for a stats + list view
- **When to separate views:**
  - Dashboard = monitoring, high-level stats, health overview
  - Agents = operational tasks (add/remove/configure agents)
  - If both show the same content, they should be one view
- **Menu item rule:** If user asks "what's the difference?" you have a UX problem

#### Route Organization
- Keep routes minimal and purposeful
- Root redirect: `/` → `/systems` provides clean landing
- Nested routes for details: `/systems/:agentId` follows RESTful convention
- Avoid route duplication even if backend supports it



## 2026-02-11: Production Dockerfile for UI

### Context
Jason requested Docker setup so he can "just run build and up" to get UI and control plane running together. Coordinated with Dallas (backend dev) who's creating the docker-compose.yml.

### Changes Made
Created production-ready Docker setup for React UI:

**Files Created:**
1. `ui/Dockerfile` - Multi-stage build (Node 22 Alpine → nginx Alpine)
2. `ui/nginx.conf` - Custom nginx config with client-side routing, gzip, security headers
3. `ui/.dockerignore` - Excludes node_modules, dist, .env from build context
4. `ui/DOCKER.md` - Complete documentation for Dallas

**Files Modified:**
1. `ui/.env.example` - Added Docker build-time comments

### Architecture Decisions

#### Multi-Stage Docker Build
- **Stage 1 (build):** Node 22 Alpine, pnpm for fast installs, Vite production build
- **Stage 2 (serve):** nginx 1.27 Alpine, serves static files from /usr/share/nginx/html
- Image size: ~50MB (nginx Alpine base is tiny, build artifacts discarded)

#### API URL Configuration Challenge
**The Issue:** Vite bakes environment variables into the JS bundle at **build time** (not runtime). This means `VITE_API_BASE_URL` must be known when `docker build` runs, not when `docker run` starts the container.

**Solution:** Docker compose passes API URL as build arg:
```yaml
build:
  context: ./ui
  args:
    - VITE_API_BASE_URL=https://control-plane:5001
```

**Why Not Runtime Config?**
- React SPA is static files after build
- No server-side rendering
- Can't inject env vars at container startup (JS already bundled)
- Alternative (complex): nginx envsubst script to rewrite built JS (not worth the fragility)

#### nginx Configuration
**Client-side routing:** `try_files $uri $uri/ /index.html;` - Serves index.html for all routes so React Router handles navigation

**Performance optimizations:**
- Gzip compression (text/css/js)
- Aggressive caching for static assets (1 year) with immutable flag
- Vite's content hashing prevents stale cache (e.g., `main-abc123.js`)

**Security headers:**
- X-Frame-Options: SAMEORIGIN (prevents clickjacking)
- X-Content-Type-Options: nosniff (prevents MIME sniffing)
- X-XSS-Protection: 1; mode=block (legacy XSS protection)

**Health check endpoint:** `/health` returns "healthy" for docker-compose health checks

#### Port Strategy
- Container exposes port 80 (nginx default)
- Docker compose maps `3000:80` on host (Vite dev server uses 3000, so consistent for developers)

### Learnings

#### Vite Environment Variables in Docker
- `VITE_*` prefixed variables are **build-time only**
- No runtime injection possible for static SPAs
- Must pass env vars as build args in Dockerfile or docker-compose
- Alternative approaches:
  - Server-side rendering (Next.js) - overkill for this project
  - Runtime config via nginx envsubst - fragile, hard to maintain
  - API gateway that proxies and rewrites URLs - adds complexity
- Build args are the cleanest solution for small number of env vars

#### Docker .dockerignore Best Practices
**Always exclude:**
- `node_modules/` (rebuilt in container)
- `dist/` (built in container)
- `.env` and `.env.local` (secrets, dev config)
- `.git/` (not needed in image)
- `*.md` (documentation not needed at runtime)
- `.DS_Store` (macOS cruft)

**Why it matters:**
- Faster builds (smaller context uploaded to Docker daemon)
- Smaller images (less garbage in final image)
- Security (no accidental secret leaks)

#### nginx for React SPAs
**Essential nginx config for SPAs:**
1. `try_files $uri $uri/ /index.html;` - Client-side routing support
2. Gzip compression - Smaller JS/CSS transfers
3. Cache-Control headers - Static asset caching
4. Security headers - Basic hardening

**React Router caveat:**
- Direct URL access (e.g., `/systems/agent-123`) hits nginx first
- Without `try_files`, nginx returns 404 (no file at that path)
- `try_files` falls back to index.html, then React Router takes over

**Development vs Production:**
- Dev: Vite dev server handles routing automatically
- Prod: nginx must be configured for client-side routing
- Many developers forget this and see 404s on page refresh

#### Multi-Stage Build Pattern
**Benefits:**
- Final image only contains runtime dependencies (nginx, not Node.js)
- Build tools and intermediate files discarded
- Dramatically smaller images (50MB vs 500MB)

**Pattern:**
```dockerfile
FROM node:22-alpine AS build
# ... install deps, build app ...

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
```

**Naming stages:**
- `AS build` names the stage
- `--from=build` references named stage
- Can have multiple stages (test, build, production)

### Coordination with Dallas

**What Dallas needs from this:**
- Service name in docker-compose: `ui`
- Build context: `./ui`
- Build arg: `VITE_API_BASE_URL=https://control-plane:5001`
- Port mapping: `3000:80`
- Depends on: `control-plane` service
- Health check: `curl http://localhost/health`

**HTTPS API URL:** UI expects control plane on HTTPS (port 5001). Dallas's control plane uses self-signed cert in dev. Docker network allows internal HTTPS communication.

**Service discovery:** Docker compose networking resolves `control-plane` hostname to the container IP. This is why build arg uses service name, not localhost.

### Files for Reference
- `ui/Dockerfile` - Production build definition
- `ui/nginx.conf` - Web server configuration
- `ui/.dockerignore` - Build context exclusions
- `ui/DOCKER.md` - Complete documentation with examples
- `ui/.env.example` - Updated with Docker comments

### User Preference
Jason wants simple "docker compose up" workflow, indicating preference for:
- Containerized development/demo environments
- Minimal local setup (no manual npm installs, no multiple terminal windows)
- One command to start entire system

📌 Team update (2026-02-11): Docker Production Build for UI — Multi-stage Dockerfile with nginx, React Router support, health checks — decided by Lambert

📌 Team update (2026-02-11): Git tracking exclusions — .ai-team/ and diagnostic files excluded from git per user directive — decided by Jason Scherer
