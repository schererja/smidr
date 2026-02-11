---
name: "efcore-migration-sync"
description: "Synchronizing EF Core entity models with existing databases that have schema drift"
domain: "database"
confidence: "high"
source: "earned"
---

## Context
When working with EF Core, entity models and database schemas can drift apart if the database was manually created or modified outside of migrations. This causes runtime errors like "no such column" when EF Core tries to query or insert data. This skill covers how to synchronize the schema, establish migration tracking, and ensure automatic migration application.

## Patterns

### Automatic Migration Application on Startup
**RECOMMENDED**: Use `Database.Migrate()` in application startup to automatically apply pending migrations. This ensures the database schema always matches the entity models.

```csharp
// Program.cs or Startup.cs
using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<YourDbContext>();
    db.Database.Migrate();  // CORRECT: Applies pending migrations
}
```

**AVOID**: Using `Database.EnsureCreated()` which creates schema but ignores migrations:

```csharp
// DON'T DO THIS
db.Database.EnsureCreated();  // WRONG: Ignores migrations, causes drift
```

**Why Migrate() is better:**
- Applies all pending migrations automatically
- Keeps schema in sync with code changes
- Idempotent (safe to run multiple times)
- No manual intervention needed when schema changes
- Works for both new databases (applies all) and existing databases (applies only pending)

**When EnsureCreated() causes problems:**
1. Database created with `EnsureCreated()` before migration exists
2. New migration added for schema change (e.g., add column)
3. Code expects new column, but database doesn't have it
4. Runtime error: "no such column"
5. **Solution**: Switch to `Migrate()` and restart application

### Detecting Schema Drift
Check if columns expected by the entity model exist in the database using SQLite pragma or SQL information schema queries.

```bash
# SQLite
sqlite3 database.db "PRAGMA table_info(TableName);"

# PostgreSQL
SELECT column_name FROM information_schema.columns WHERE table_name = 'tablename';
```

### Manual Schema Synchronization
When the database exists but is out of sync, add missing columns manually using ALTER TABLE statements.

```bash
# Add nullable column
sqlite3 database.db "ALTER TABLE TableName ADD COLUMN ColumnName TEXT NULL;"

# Add non-nullable column with default
sqlite3 database.db "ALTER TABLE TableName ADD COLUMN ColumnName TEXT NOT NULL DEFAULT 'value';"
```

**Important**: SQLite does NOT support:
- Adding non-nullable columns without defaults
- Dropping columns (requires table recreation)
- Modifying column types (requires table recreation)

### Establishing Migration Tracking
After manually synchronizing the schema, create an EF Core migration to track the current state and mark it as applied.

```bash
# Create migration matching current schema
dotnet ef migrations add InitialCreate

# Manually mark as applied (when migration fails due to existing tables)
sqlite3 database.db "INSERT INTO __EFMigrationsHistory (MigrationId, ProductVersion) VALUES ('20260210123456_InitialCreate', '8.0.0');"

# Verify tracking
dotnet ef migrations list
```

### Migration Naming Convention
Use descriptive names that indicate what changed, or use `InitialCreate` for the first migration that establishes baseline schema tracking.

```bash
dotnet ef migrations add AddCertificateColumn
dotnet ef migrations add UpdateAgentSchema
dotnet ef migrations add InitialCreate  # First migration for existing DB
```

### Applying Migrations to Empty Database
For new deployments or clean databases, simply run the migration without manual SQL.

```bash
dotnet ef database update
```

## Examples

### Full Schema Synchronization Process
```bash
# 1. Check current schema
sqlite3 Data/database.db "PRAGMA table_info(Agents);"

# 2. Add missing columns
sqlite3 Data/database.db "ALTER TABLE Agents ADD COLUMN CertificatePem TEXT NULL;"
sqlite3 Data/database.db "ALTER TABLE Agents ADD COLUMN CurrentHealth TEXT NOT NULL DEFAULT 'Learning';"

# 3. Create migration to track schema
cd control-plane
dotnet ef migrations add InitialCreate

# 4. Mark migration as applied
sqlite3 Data/database.db "INSERT INTO __EFMigrationsHistory (MigrationId, ProductVersion) VALUES ('20260210202319_InitialCreate', '8.0.0');"

# 5. Verify
dotnet ef migrations list
```

### Handling Database Lock During Migration
If the application is running, the database may be locked. Options:
1. Stop the application first
2. Add columns one at a time (some may succeed before lock)
3. Use a connection with shorter timeout

```bash
# Check for running processes
ps aux | grep "control-plane\|dotnet"

# Kill if necessary (use specific PID)
kill 12345
```

## Anti-Patterns
- **Using EnsureCreated() in production code** — Ignores migrations, causes schema drift when new migrations are added
- **Ignoring migrations** — Creates drift between environments, makes deployments unpredictable
- **Deleting and recreating database** — Loses data, not viable for production (but acceptable for dev when migration history is corrupted)
- **Running migrations without backup** — Risk of data loss if migration fails
- **Not verifying schema after changes** — May miss additional drift
- **Using generic migration names** — Makes history unclear (e.g., "Migration1", "UpdateDb")

## Recovery from Corrupted Migration State
When `__EFMigrationsHistory` is out of sync with actual schema (tables exist but aren't marked as migrated), you may see errors like "table X already exists".

### Development Environment
For SQLite databases in development, the fastest solution is a clean slate:

```bash
# Delete database files
rm -f data/database.db data/database.db-shm data/database.db-wal

# Re-run all migrations
dotnet ef database update

# Verify all migrations applied
dotnet ef migrations list
```

### Production Environment
For production, requires careful manual reconciliation:
1. Backup database first
2. Compare `__EFMigrationsHistory` table with actual schema
3. Either manually add missing migration records, or manually apply missing schema changes
4. Verify schema matches model before marking migration as applied
5. Test thoroughly in staging environment first

### Automation Script
Create a `fix-migrations.sh` script for repeatable recovery in development:

```bash
#!/bin/bash
cd "$(dirname "$0")"
rm -f data/database.db data/database.db-shm data/database.db-wal
dotnet ef database update
dotnet ef migrations list
echo "Done! Database recreated with all migrations applied."
```

## Benefits
- **Environment consistency**: Same schema across dev, staging, prod
- **Version control**: Schema changes tracked in source control
- **Deployment automation**: `dotnet ef database update` handles all schema changes
- **Rollback capability**: Can revert migrations if needed
- **Team collaboration**: Everyone knows what schema changes happened and when
