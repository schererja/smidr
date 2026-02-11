### 2026-02-11: UI prepared for OS icons and multiple drives
**By:** Lambert
**What:** UI now has OSIcon component and types ready for OS detection. Added placeholder for multiple drive support.

**Current State:**
- UI has OS field in Agent type (optional)
- OSIcon component renders OS-appropriate icons (Linux/Windows/macOS)
- Agent cards and detail pages display OS icons
- Single disk usage metric displayed with TODO comment

**What Needs Backend Changes:**
1. **OS Detection** (requires agent + control plane changes):
   - Agent: Send runtime.GOOS in RegisterAgentRequest
   - Control Plane: Add OS field to Agent model
   - Control Plane: Accept OS in RegisterAgentRequest
   - Control Plane: Return OS in API responses

2. **Multiple Drives** (requires agent + control plane changes):
   - Agent: Detect all mounted filesystems, send array of disk metrics
   - Agent: Each disk should include mount point, total space, used space, used %
   - Control Plane: Update Heartbeat model to support disk array
   - Control Plane: Calculate baselines per drive
   - UI: Display each drive separately in metrics card

**Why:** Jason requested OS icons for each system and support for multiple drives. UI is now prepared to receive this data, but backend changes are required before it can be implemented.
