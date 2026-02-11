### 2026-02-11: Removed duplicate navigation routes

**By:** Lambert

**What:** Removed the `/agents` route and "Agents" sidebar navigation item. Dashboard at `/systems` is now the single view for monitoring all agents.

**Why:** Both "Dashboard" and "Agents" navigation items pointed to the same `<SystemList />` component, creating user confusion. The Dashboard label better describes the page's purpose (overview with stats, search, and agent list). If we need a distinct "Agents" page in the future for operational tasks (add/remove/configure), we can add it back with different functionality.

**Files Changed:**
- `ui/src/components/Sidebar.tsx` - Removed `/agents` nav item from navItems array
- `ui/src/App.tsx` - Removed `<Route path="/agents" ...>` definition
