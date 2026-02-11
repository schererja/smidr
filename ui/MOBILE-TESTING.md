# Mobile Responsiveness Testing Guide

## Testing Approach

Since automated mobile testing with Playwright requires installation, use browser DevTools for testing:

### Chrome/Edge DevTools
1. Open the UI at http://localhost:3000
2. Press F12 or Cmd+Option+I (Mac) / Ctrl+Shift+I (Windows)
3. Click the device toolbar icon or press Cmd+Shift+M (Mac) / Ctrl+Shift+M (Windows)
4. Select preset devices or custom dimensions

### Test Scenarios

#### 1. Sidebar Navigation (Mobile Menu)
- **Mobile (<1024px)**: Sidebar should be hidden by default
- Click hamburger menu icon in header → sidebar slides in from left
- Dark overlay appears behind sidebar
- Click overlay or navigate → sidebar closes
- **Desktop (≥1024px)**: Sidebar always visible, no hamburger menu

#### 2. Dashboard Layout (SystemList)
**Mobile (320-640px):**
- Header: Hamburger menu + notification bell only, search hidden
- Page title: smaller font (text-2xl)
- Stat cards: 2 columns (grid-cols-2)
- Search bar: full width
- Grid view: 1 column
- Table view: Only Hostname, Health, Last Seen columns visible
  - Agent ID, Load, Memory, Disk hidden (hidden on md/lg breakpoints)

**Tablet (641-1024px):**
- Header: Hamburger menu + search bar + notifications
- Stat cards: 2 columns (grid-cols-2)
- Grid view: 2 columns
- Table view: Hostname, Agent ID, Health, Last Seen visible

**Desktop (≥1024px):**
- Sidebar always visible
- Stat cards: 4 columns (grid-cols-4)
- Grid view: 3 columns
- Table view: All columns visible

#### 3. System Detail Page
**Mobile:**
- Breadcrumb truncates hostname if needed
- Hero card padding reduced (p-4)
- Agent icon smaller (w-12 h-12)
- Health badge stacks below agent info on very small screens
- Metrics cards: 1 column
- Baselines grid: 1-2 columns depending on screen size

**Tablet/Desktop:**
- Standard multi-column layouts
- All content visible

### Test Viewports
- **iPhone SE**: 375x667
- **iPhone 12 Pro**: 390x844
- **iPad**: 768x1024
- **iPad Pro**: 1024x1366
- **Desktop**: 1920x1080

### What Was Fixed

1. **Sidebar Component**:
   - Added `isOpen` and `onClose` props
   - Fixed positioning: `fixed lg:static`
   - Added mobile overlay with backdrop
   - Slides in/out with CSS transforms
   - Links close sidebar on mobile

2. **Header Component**:
   - Added `onMenuClick` prop for hamburger menu
   - Hamburger menu button: `lg:hidden`
   - Search bar: `hidden md:block`
   - Mobile search icon: `md:hidden`

3. **App Component**:
   - State management for sidebar open/close
   - Passes props to Header and Sidebar

4. **SystemList Page**:
   - Responsive padding: `p-4 md:p-8`
   - Stat cards: `grid-cols-2 lg:grid-cols-4`
   - Agent grid: `grid-cols-1 md:grid-cols-2 lg:grid-cols-3`
   - Table columns: Hide on mobile with `hidden md:table-cell`, `hidden lg:table-cell`
   - Overflow scroll for table on mobile

5. **SystemDetail Page**:
   - Hero section: Flex column on mobile, row on desktop
   - Responsive text sizes: `text-xl md:text-3xl`
   - Responsive icon sizes: `w-12 h-12 md:w-16 md:h-16`
   - Baselines grid: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3`

## Known Issues

### OS Icons
- UI is ready to display OS icons
- Backend needs to send OS information (runtime.GOOS from agent)
- Currently shows generic server icon for all agents

### Multiple Drives
- Agent currently sends single disk usage percentage
- UI displays aggregate disk usage
- Future: Agent should send array of drives with mount points

## Bugs Fixed

### 1. Healthy Fleet Percentage (Line 136 in SystemList.tsx)
**Before:** `((healthyAgents / totalAgents || 0) * 100)`
- Bug: When `healthyAgents = 0`, division gives `0` which is falsy, so `|| 0` returns 0
- Result: Always showed 0%

**After:** `totalAgents > 0 ? ((healthyAgents / totalAgents) * 100).toFixed(0) : 0`
- Correct: Check if totalAgents > 0 first, then calculate percentage

### 2. Metrics Count
**Not a bug!** The UI shows "4 metrics tracked" which is correct:
1. LoadAverage1m
2. MemoryUsedPct
3. DiskUsedPct
4. ProcessCount

Note: `uptimeSeconds` is collected but doesn't get a baseline (it's monotonically increasing).
