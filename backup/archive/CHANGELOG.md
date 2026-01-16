# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Phase 3: Source Management
  - Implemented persistent cache metadata for repositories and downloads (`.smidr_meta.json`) with last-access timestamps.
  - Added TTL-based eviction (EvictOldCache) for selective cache cleanup.
  - Implemented per-repo lockfiles to prevent concurrent fetch corruption.
  - Added Downloader improvements: mirror support (try PREMIRRORS/MIRRORS), retries with exponential backoff, and unit tests covering mirror/fallback behavior.
  - Added unit tests and small integration tests to ensure stability.

### Changed

- Centralized cache metadata helpers in `internal/source/cachemeta.go`.

### Fixed

- Fixed race conditions and file placement issues discovered during implementation and testing.

---

For full details, review the commit history.
