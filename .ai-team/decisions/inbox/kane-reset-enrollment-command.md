### 2026-02-11: Agent enrollment reset command
**By:** Kane
**What:** Added `reset-enrollment` command to agent that clears all enrollment state (cert, csr, key files and agent_id) and enables re-enrollment

**Why:** When control plane database is reset, agents with old certificates can't heartbeat because their agent ID no longer exists in the database. The reset command provides a clean way to clear stale enrollment state without manual file deletion. After reset, the agent daemon automatically generates new credentials and re-enrolls with the control plane. This is safer than manual `rm` commands and preserves important config like control_plane_url and heartbeat_interval.

**Implementation Pattern — Idempotent Reset Commands:**
This reset command follows an idempotent design pattern that's reusable for any cleanup/reset operation:

1. **Safe file deletion:** Check if file exists before deleting (no error if already deleted)
   ```go
   func deleteFileIfExists(path string) error {
       if _, err := os.Stat(path); err == nil {
           return os.Remove(path)
       } else if !errors.Is(err, os.ErrNotExist) {
           return err  // Report permission errors, etc.
       }
       return nil  // File doesn't exist, nothing to do
   }
   ```

2. **Selective config clearing:** Only clear reset-related fields, preserve others
   ```go
   updates := map[string]string{
       "agent_id":  "",
       "key_path":  "",
       "csr_path":  "",
       "cert_path": "",
   }
   // Preserves: control_plane_url, hostname, token, heartbeat_interval
   ```

3. **Clear user feedback:** Tell user what was done and next steps
4. **Test both scenarios:** Fresh reset and double reset (idempotency)

This pattern makes reset commands operator-friendly and automation-ready (can run multiple times during troubleshooting without errors).
