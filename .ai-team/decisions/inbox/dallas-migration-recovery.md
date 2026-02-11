### 2026-02-11: Database Migration Recovery Strategy
**By:** Dallas
**What:** Established approach for recovering from corrupted EF Core migration state: delete database and re-run all migrations from scratch.
**Why:** When `__EFMigrationsHistory` is out of sync with actual schema (tables exist but aren't marked as migrated), attempting to fix manually is error-prone. For development environments, the fastest and safest solution is to delete the SQLite database and let EF Core recreate it cleanly. Created `control-plane/fix-migrations.sh` script to automate this. For production, would require backup and careful manual reconciliation.
