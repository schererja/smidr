Phase 3 - Cache & Mirrors (summary)

This PR accompanies the Phase 3 work that implements:

- Persistent cache metadata for repositories and downloads (`.smidr_meta.json`) and TTL-based eviction.
- Per-repo lockfiles to avoid concurrent fetch corruption.
- Download retries with exponential backoff and mirror support (PREMIRRORS/MIRRORS).
- Unit tests for downloader mirror/retry behavior and cache eviction.
- Documentation (`docs/cache.md`) and changelog entry.

Note: the substantive code changes for Phase 3 were already pushed to `main` in previous commits; this PR provides a short human-facing note and a place for review comments if desired.
