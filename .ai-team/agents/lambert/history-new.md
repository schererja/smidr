# Lambert Work History

## Core Context

**Role:** Frontend developer focused on React UI, API integration, and user experience

**Key Accomplishments (Condensed):**
- Built modern SaaS dashboard UI with Tailwind CSS + shadcn/ui component library
- Integrated UI with control plane REST API, replacing mock data with real-time agent monitoring
- Implemented responsive design for mobile/tablet/desktop with hamburger menu navigation
- Created reusable component library: Sidebar, Header, StatCard, AgentCard, OSIcon, HealthBadge, MetricsCard
- Established dashboard architecture: summary stats, grid/table views, search/filtering, health visualization
- Fixed critical bugs: hamburger overlay opacity, healthy fleet % calculation, missing CSS imports, negative last heartbeat display
- Integrated OS detection with platform-appropriate icons (ready for backend data)
- Deferred multi-drive metrics to v1; current implementation tracks root filesystem only

**Tech Stack:**
- React 19 + TypeScript (strict mode), Vite build, Tailwind CSS 4.x + shadcn/ui components, React Router, Axios
- API integration via environment variables (VITE_ prefix), self-signed HTTPS on port 5001
- Mobile-responsive: sidebar overlay on mobile, inline on desktop; hamburger menu; responsive grids

**Key Learning:** shadcn/ui requires two-part setup (CSS variables in @layer base + Tailwind config mapping), CSS import location critical in Vite, modern Tailwind uses slash notation for opacity (bg-black/60)

**Files:**
- `ui/src/api/client.ts` - API integration layer
- `ui/src/pages/SystemList.tsx` - Dashboard with summary stats and agent grid
- `ui/src/pages/SystemDetail.tsx` - Agent detail with signals, baselines, heartbeats
- `ui/src/components/` - Sidebar, Header, StatCard, AgentCard, HealthBadge, MetricsCard, OSIcon
- `ui/tailwind.config.js` - Tailwind theme with semantic color variables

---

## 2026-02-11: Recent Work Sessions

### Mobile Responsiveness and Bug Fixes

**Changes Made:**
- Made entire UI mobile-responsive with hamburger menu, responsive layouts, fixed calculation bugs
- Files modified: App.tsx, Sidebar.tsx, Header.tsx, SystemList.tsx, SystemDetail.tsx, types/index.ts, components/OSIcon.tsx, components/AgentCard.tsx, components/MetricsCard.tsx
- Created `ui/MOBILE-TESTING.md` comprehensive mobile testing guide

**Bugs Fixed:**

1. **Healthy Fleet Percentage Calculation (SystemList.tsx)**
   - Problem: Expression `((healthyAgents / totalAgents || 0) * 100)` had incorrect operator precedence
   - Fix: Changed to `totalAgents > 0 ? ((healthyAgents / totalAgents) * 100).toFixed(0) : 0`
   - Lesson: Always check divisor explicitly, don't rely on `||` for division-by-zero when numerator can be 0

2. **Metrics Count "Bug"** (Investigation)
   - Discovered: Agent collects 5 signals, control plane calculates baselines for only 4 (excludes uptimeSeconds)
   - Reason: uptimeSeconds is monotonically increasing, doesn't make statistical sense for baselines
   - UI correctly displays "4 metrics tracked"

### Mobile Responsiveness Patterns Documented

**Sidebar Navigation on Mobile:**
- Fixed positioning with `fixed lg:static` - overlay on mobile, inline on desktop
- CSS transforms for slide: `translate-x-0` vs `-translate-x-full`
- Dark overlay backdrop: `fixed inset-0 bg-black/60 z-40`
- Close sidebar on navigation

**Hamburger Menu Pattern:**
- Visible only on mobile: `lg:hidden`
- Icon from lucide-react Menu component
- Positioned in header

**Responsive Grid Breakpoints:**
- Stat cards: `grid-cols-2 lg:grid-cols-4` (2→4 columns)
- Agent grid: `grid-cols-1 md:grid-cols-2 lg:grid-cols-3` (1→2→3)
- Table columns: `hidden md:table-cell` and `hidden lg:table-cell`
- Padding: `p-4 md:p-8` (progressive enhancement)

**Table Responsiveness:**
- Overflow scroll: `overflow-x-auto` container
- Hide columns progressively
- Mobile priority: Hostname, Health, Last Seen (always visible)
- Desktop: Agent ID, Load, Memory, Disk, Registered (hidden on mobile)

### OS Detection and Icons

**Current State:**
- UI prepared with OS type: `'linux' | 'windows' | 'darwin' | 'unknown'`
- OSIcon component created with placeholder icons (Server for Linux, Monitor for Windows/Mac)
- Agent cards and detail pages display OS icons
- Backend changes complete: Agent sends `runtime.GOOS`, Control Plane stores OS, returns in API

### Multiple Drives Support (Deferred to v1)

**Current Implementation:**
- Agent collects disk usage for root filesystem only (`collectDiskUsage("/")`)
- Single `diskUsedPct` field in signals
- UI displays aggregate disk usage

**Future Enhancement (v1):**
- Agent detects all mounted filesystems via `/proc/mounts`
- Control Plane: DiskMetric table with (heartbeat_id, mount_point, used_pct, total_gb, used_gb)
- Per-mount baselines
- UI: Expandable "Disks" section with table of mount points
- Design challenge: How to display 5+ drives compactly

### Fixed Hamburger Menu Overlay Opacity

**Change:** `bg-black bg-opacity-50` → `bg-black/60` (Tailwind 3.x+ slash notation)
- Modern syntax more reliable and concise
- 60% opacity: visible but allows content context awareness
- Users can now see content behind overlay

### Removed Duplicate Navigation Route

**Problem:** Both `/systems` and `/agents` pointed to same `<SystemList />` component
**Solution:** Removed `/agents` route and "Agents" nav item; `/systems` is single entry point
**Rationale:** Dashboard more descriptive for view with stats + list; "Agents" would be for operational tasks (registration, bulk ops) in future

---

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

---

For earlier work history and detailed learnings, see `history-archive.md`.
