### 2026-02-11: Multi-drive disk metrics architecture review

**By:** Ripley  
**What:** Reviewed agent disk collection architecture in response to Jason's observation that systems can have multiple drives. Confirmed current implementation only tracks root filesystem (`/`).  
**Recommendation:** Keep single aggregate disk metric for v0, defer per-drive tracking to v1.

---

## Current State

**Agent (Go):**
- `agent/internal/signals/signals.go` line 49: `collectDiskUsage("/")` — hardcoded to root filesystem
- Uses `syscall.Statfs()` on a single mount point
- Returns single `DiskUsedPct` float in `Snapshot` struct
- No enumeration of `/proc/mounts` or multiple filesystems

**Control Plane (C#):**
- `Heartbeat` model has single `DiskUsedPct` field (double)
- Health evaluation treats "DiskUsedPct" as one metric name (line 103, HealthEvaluationService.cs)
- Baseline computed per metric name — one baseline for aggregate disk usage
- Database schema: `heartbeats` table has single `disk_used_pct` column

**API Contract:**
- `/v0/agents/heartbeat` expects single `diskUsedPct` field (docs/API.md line 143)
- Documented as "Root disk usage percentage (0-100)"

**UI:**
- Displays single disk percentage value in metrics cards

---

## Architecture Decision: Root-Only for v0

**Recommendation:** Keep single-filesystem tracking for Smidr v0. Do NOT implement per-drive metrics now.

### Rationale

**1. Scope Creep Risk**
- v0 spec defined five signals: uptime, load, memory, disk, process count
- Adding per-drive tracking changes data model, API contract, UI, baselines, health evaluation
- Would require 4-component change (agent, control plane, database migration, UI)

**2. Most Servers Have Simple Disk Topologies**
- Target: Linux x86_64 systems (cloud VMs, containers, basic servers)
- Common pattern: single root filesystem, or root + separate /data mount
- For multi-drive systems, root filesystem health is still the critical signal (OS binaries, logs, temp space)

**3. Baseline Complexity**
- Per-drive baselines complicate evaluation: Do we alert if ANY drive is anomalous? Or majority?
- Different drives have different usage patterns (e.g., /var/log grows linearly, /home is bursty)
- Single root metric is simple, explainable, and sufficient for "is this system healthy?" question

**4. API Versioning Would Be Required**
- Changing heartbeat payload from single `diskUsedPct` to array of drives breaks existing agent/control-plane contract
- Would need `/v1/agents/heartbeat` endpoint or feature flag
- Not worth the versioning burden for v0

**5. UI Display Challenges**
- How do we show 5+ drives in a compact metrics card?
- Which drive do we show in the system list view health badge?
- Do we aggregate across drives (back to single metric)? If so, what did we gain?

### What Jason's Right About

Jason correctly identified that **production systems often have multiple drives:**
- Separate data volumes (`/data`, `/var/lib/docker`)
- Database volumes (`/var/lib/postgresql`, `/mnt/db`)
- Log aggregation mounts (`/var/log`)
- Ephemeral storage for temp files

If `/` is only 10% full but `/data` is at 98%, we'd miss a real problem.

### The Right Solution for v1

When we tackle this properly (v1+), here's the architecture:

**1. Agent Change:**
```go
type DiskSnapshot struct {
    MountPoint string  `json:"mountPoint"`
    UsedPct    float64 `json:"usedPct"`
    TotalGB    float64 `json:"totalGB"`
    UsedGB     float64 `json:"usedGB"`
}

type Snapshot struct {
    // ... existing fields
    Disks []DiskSnapshot `json:"disks"`  // NEW: array of disks
}
```

Parse `/proc/mounts`, filter to real filesystems (ext4, xfs, btrfs, not tmpfs/devtmpfs), run Statfs on each.

**2. Control Plane Changes:**
- New `DiskMetric` table: `(heartbeat_id, mount_point, used_pct, total_gb, used_gb)`
- Baselines: Store per-mount baseline (e.g., "DiskUsedPct:/data", "DiskUsedPct:/")
- Health evaluation: Treat each mount as independent metric, OR aggregate anomaly count across drives

**3. API Contract:**
- Increment to `/v1/agents/heartbeat`
- Change `diskUsedPct` from float to array of objects
- Maintain `/v0/` for backward compatibility (or deprecate)

**4. UI Changes:**
- System detail view: Expandable "Disks" section with table of mount points
- System list view: Show worst disk percentage or aggregate health across all disks
- Baselines chart: Per-mount sparklines

**5. Migration Path:**
- Ship v0 with root-only disk tracking, gather user feedback
- If users report "missed full disk on /data mount" issues, prioritize v1 multi-disk feature
- If root-only proves sufficient, defer indefinitely

---

## Recommendation Summary

**Keep it simple for v0:**
1. No changes to agent disk collection
2. No changes to control plane models or API
3. Document limitation in README: "v0 tracks root filesystem only"
4. Add to backlog: "v1: per-mount disk metrics" as future enhancement

**If Jason insists on multi-disk for v0:**
1. Spawn Kane to add `/proc/mounts` parsing and multi-drive collection
2. Spawn Dallas to add `DiskMetric` model and update heartbeat schema
3. Spawn Lambert to update UI for disk arrays
4. Requires database migration, API version bump, UI redesign
5. Estimated effort: 1-2 days (vs. 0 hours for deferral)

**My call as Lead:** Ship v0 with root-only disk. Gather feedback. Implement per-drive in v1 if users need it.

---

## If We Must Implement Now (Fallback Plan)

If Jason decides this is critical for v0, here's the minimal-viable approach:

**Option A: Agent-side aggregation**
- Agent reads `/proc/mounts`, collects all non-tmpfs filesystems
- Computes weighted average: `total_used_blocks / total_available_blocks` across all drives
- Still sends single `diskUsedPct` float
- **Pros:** No API/schema changes, captures multi-drive reality
- **Cons:** Loses per-drive granularity, can mask one full drive if others are empty

**Option B: Array field with backward compat**
- Add `disks: []` array to heartbeat request, keep `diskUsedPct` as deprecated field
- Agent sends both: single root metric + optional array
- Control plane stores array in JSON column or separate table
- v0 evaluation uses root only, v1 evaluation uses array
- **Pros:** Forward-compatible, no API version bump
- **Cons:** Doubles storage for disk metrics, UI doesn't show array yet

I recommend **Option A** if we must ship multi-disk in v0.

---

**Follow-up:** Jason, let me know your decision. If you want to proceed with multi-drive now, I'll coordinate Kane, Dallas, and Lambert for the implementation.
