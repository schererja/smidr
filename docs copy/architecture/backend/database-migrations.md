# Database Migrations

## Overview

Yggdrasil uses **golang-migrate** for database schema migrations. This document covers migration strategy, best practices, and common patterns.

---

## golang-migrate Setup

### Installation

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Go install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Configuration

Migrations are stored in the `/migrations` directory:

```
migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000002_add_tickets_table.up.sql
├── 000002_add_tickets_table.down.sql
├── 000003_add_indexes.up.sql
└── 000003_add_indexes.down.sql
```

### Running Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://user:pass@localhost:5432/yggdrasil?sslmode=disable"

# Migrate up (apply all pending)
migrate -path migrations -database "${DATABASE_URL}" up

# Migrate up by N versions
migrate -path migrations -database "${DATABASE_URL}" up 2

# Migrate down (revert all)
migrate -path migrations -database "${DATABASE_URL}" down

# Migrate down by N versions
migrate -path migrations -database "${DATABASE_URL}" down 1

# Force to specific version (use carefully)
migrate -path migrations -database "${DATABASE_URL}" force 3

# Check current version
migrate -path migrations -database "${DATABASE_URL}" version
```

    // golang-migrate Go integration example

package database

import (
"context"
"database/sql"
"log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/lib/pq"

)

// MigrateUp runs all pending migrations
func MigrateUp(databaseURL string, migrationsPath string) error {
db, err := sql.Open("postgres", databaseURL)
if err != nil {
return err
}
defer db.Close()

    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://"+migrationsPath,
        "postgres", driver)
    if err != nil {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil

}

// MigrateDown reverts all migrations
func MigrateDown(databaseURL string, migrationsPath string) error {
db, err := sql.Open("postgres", databaseURL)
if err != nil {
return err
}
defer db.Close()

    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://"+migrationsPath,
        "postgres", driver)
    if err != nil {
        return err
    }

    if err := m.Down(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil

}

````

---

## Migration Workflow

### Creating Migrations

**Create a new migration**:

```bash
# Create new migration files
migrate create -ext sql -dir migrations -seq add_ticket_table
````

This creates two files:

- `000002_add_ticket_table.up.sql` (apply changes)
- `000002_add_ticket_table.down.sql` (revert changes)

### Writing Migrations

**Example Up Migration** (`000002_add_tickets_table.up.sql`):

```sql
-- Create tickets table
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subject VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    priority VARCHAR(50) NOT NULL DEFAULT 'medium',
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_tickets_tenant_id ON tickets(tenant_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_assigned_to ON tickets(assigned_to);
CREATE INDEX idx_tickets_created_at ON tickets(created_at DESC);

-- Add row-level security
ALTER TABLE tickets ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON tickets
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
```

**Example Down Migration** (`000002_add_tickets_table.down.sql`):

```sql
-- Drop policies first
DROP POLICY IF EXISTS tenant_isolation ON tickets;

-- Drop indexes
DROP INDEX IF EXISTS idx_tickets_created_at;
DROP INDEX IF EXISTS idx_tickets_assigned_to;
DROP INDEX IF EXISTS idx_tickets_status;
DROP INDEX IF EXISTS idx_tickets_tenant_id;

-- Drop table
DROP TABLE IF EXISTS tickets;
```

-- Create tickets table
CREATE TABLE IF NOT EXISTS tickets (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
title VARCHAR(500) NOT NULL,
description TEXT,
status VARCHAR(50) NOT NULL DEFAULT 'new',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_tickets_tenant ON tickets(msp_tenant_id);
CREATE INDEX idx_tickets_client ON tickets(client_id);
CREATE INDEX idx_tickets_status ON tickets(status);

### Applying Migrations with golang-migrate

**Upgrade to latest**:

```bash
migrate -path migrations -database "${DATABASE_URL}" up
```

**Upgrade by N versions**:

```bash
migrate -path migrations -database "${DATABASE_URL}" up 2
```

**Downgrade by N versions**:

```bash
migrate -path migrations -database "${DATABASE_URL}" down 1
```

**Force to specific version**:

```bash
migrate -path migrations -database "${DATABASE_URL}" force 3
```

### Checking Migration Status

```bash
# Check current version
migrate -path migrations -database "${DATABASE_URL}" version

# Show migration history
migrate -path migrations -database "${DATABASE_URL}" force 0 && migrate -path migrations -database "${DATABASE_URL}" up
```

---

## Migration Patterns

### 1. Adding a Column (Non-Breaking)

**Safe Pattern**: Add nullable column first, backfill data, then make NOT NULL if needed.

```sql
-- Step 1: Add nullable column
ALTER TABLE tickets ADD COLUMN priority VARCHAR(50);

-- Step 2: Set default value for existing rows
UPDATE tickets SET priority = 'medium' WHERE priority IS NULL;

-- Step 3: Make NOT NULL (if required)
ALTER TABLE tickets ALTER COLUMN priority SET NOT NULL;

-- Step 4: Add index if needed
CREATE INDEX idx_tickets_priority ON tickets(priority);

-- Downgrade
DROP INDEX idx_tickets_priority;
ALTER TABLE tickets DROP COLUMN priority;
```

### 2. Renaming a Column (Breaking Change)

**Safe Pattern**: Add new column, copy data, deprecate old column, remove in next release.

**Migration 1** (Add new column):

```sql
-- Migration 1 (Add new column)

-- Add new column
ALTER TABLE tickets ADD COLUMN assigned_to UUID;

-- Copy data from old column
UPDATE tickets SET assigned_to = assignee_id WHERE assignee_id IS NOT NULL;

-- Add foreign key
ALTER TABLE tickets ADD CONSTRAINT fk_tickets_assigned_to
    FOREIGN KEY (assigned_to) REFERENCES users(id);

-- Note: Keep assignee_id for now (backwards compatibility)

-- Downgrade
ALTER TABLE tickets DROP CONSTRAINT fk_tickets_assigned_to;
ALTER TABLE tickets DROP COLUMN assigned_to;
```

**Migration 2** (Remove old column, next release):

```sql
-- Migration 2 (Remove old column, next release)

ALTER TABLE tickets DROP CONSTRAINT fk_tickets_assignee_id;
ALTER TABLE tickets DROP COLUMN assignee_id;

-- Downgrade
ALTER TABLE tickets ADD COLUMN assignee_id UUID;
UPDATE tickets SET assignee_id = assigned_to WHERE assigned_to IS NOT NULL;
ALTER TABLE tickets ADD CONSTRAINT fk_tickets_assignee_id
    FOREIGN KEY (assignee_id) REFERENCES users(id);
```

### 3. Adding an Index (Safe)

```sql
-- Add index
CREATE INDEX idx_tickets_created_at ON tickets (created_at);

-- Or for descending order:
CREATE INDEX idx_tickets_created_at_desc ON tickets (created_at DESC);

-- Downgrade
DROP INDEX idx_tickets_created_at;
```

### 4. Adding a Foreign Key

```sql
-- Add foreign key and index
ALTER TABLE tickets ADD COLUMN system_id UUID;

ALTER TABLE tickets ADD CONSTRAINT fk_tickets_system_id
    FOREIGN KEY (system_id) REFERENCES systems(id) ON DELETE SET NULL;

CREATE INDEX idx_tickets_system ON tickets (system_id);

-- Downgrade
DROP INDEX idx_tickets_system;
ALTER TABLE tickets DROP CONSTRAINT fk_tickets_system_id;
ALTER TABLE tickets DROP COLUMN system_id;
```

### 5. Data Migration

**Example**: Populate ticket numbers for existing tickets.

```sql
-- Add ticket_number column
ALTER TABLE tickets ADD COLUMN ticket_number INTEGER;

-- For each tenant, assign sequential ticket numbers
DO $$
DECLARE
    tenant_record RECORD;
BEGIN
    FOR tenant_record IN SELECT DISTINCT msp_tenant_id FROM tickets LOOP
        EXECUTE format('
            WITH numbered_tickets AS (
                SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) as ticket_num
                FROM tickets
                WHERE msp_tenant_id = %L
            )
            UPDATE tickets
            SET ticket_number = numbered_tickets.ticket_num
            FROM numbered_tickets
            WHERE tickets.id = numbered_tickets.id
        ', tenant_record.msp_tenant_id);
    END LOOP;
END $$;

-- Make NOT NULL after backfill
ALTER TABLE tickets ALTER COLUMN ticket_number SET NOT NULL;

-- Add unique constraint per tenant
ALTER TABLE tickets ADD CONSTRAINT uq_tickets_number_tenant
    UNIQUE (msp_tenant_id, ticket_number);

-- Downgrade
ALTER TABLE tickets DROP CONSTRAINT uq_tickets_number_tenant;
ALTER TABLE tickets DROP COLUMN ticket_number;
```

### 6. Creating TimescaleDB Hypertable

```sql
-- Create regular table
CREATE TABLE metric_data (
    time TIMESTAMPTZ NOT NULL,
    msp_tenant_id UUID NOT NULL REFERENCES msp_tenants(id) ON DELETE CASCADE,
    system_id UUID NOT NULL REFERENCES systems(id) ON DELETE CASCADE,
    metric_name VARCHAR(100) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    labels JSONB DEFAULT '{}',
    PRIMARY KEY (time, system_id, metric_name)
);

-- Convert to hypertable
SELECT create_hypertable('metric_data', 'time');

-- Create indexes
CREATE INDEX idx_metric_data_tenant ON metric_data (msp_tenant_id, time DESC);
CREATE INDEX idx_metric_data_system ON metric_data (system_id, time DESC);
CREATE INDEX idx_metric_data_name ON metric_data (metric_name, time DESC);

-- Downgrade
DROP TABLE metric_data;
```

### 7. Enabling Row-Level Security (RLS)

```sql
-- Enable RLS
ALTER TABLE tickets ENABLE ROW LEVEL SECURITY;

-- Create policy
CREATE POLICY tenant_isolation_policy ON tickets
    USING (msp_tenant_id::text = current_setting('app.current_tenant_id', TRUE));

-- Downgrade
DROP POLICY IF EXISTS tenant_isolation_policy ON tickets;
ALTER TABLE tickets DISABLE ROW LEVEL SECURITY;
```

---

## Best Practices

### 1. **Never Edit Existing Migrations**

Once a migration is committed and shared, never edit it. Create a new migration to fix issues.

### 2. **One Migration Per Logical Change**

Don't bundle unrelated changes in one migration. Easier to review and rollback.

### 3. **Always Include Downgrade**

Every `upgrade()` must have a corresponding `downgrade()`. Ensure rollback works.

### 4. **Test Migrations on Real Data**

Before deploying, test migrations against a copy of production data:

```bash
# Create snapshot of production DB
pg_dump production_db > snapshot.sql

# Restore to test DB
psql test_db < snapshot.sql

# Run migration
migrate -path migrations -database "${DATABASE_URL}" up

# Verify data integrity
```

### 5. **Use Concurrent Indexes**

For large tables, avoid locking by creating indexes concurrently:

```sql
-- For large tables, create index without locking
CREATE INDEX CONCURRENTLY idx_tickets_title ON tickets(title);

-- In migration file, handle the non-transactional nature:
-- Note: CREATE INDEX CONCURRENTLY cannot run in a transaction
-- Use separate migration files if needed
```

### 6. **Add Comments**

```sql
-- Add priority column for ticket triage
-- Default to 'medium' for existing tickets
ALTER TABLE tickets ADD COLUMN priority VARCHAR(50);
UPDATE tickets SET priority = 'medium' WHERE priority IS NULL;
ALTER TABLE tickets ALTER COLUMN priority SET NOT NULL;

-- Add table comment
COMMENT ON TABLE tickets IS 'Customer support tickets with tenant isolation';

-- Add column comments
COMMENT ON COLUMN tickets.priority IS 'Ticket priority: low, medium, high, critical';
```

### 7. **Check for Long-Running Migrations**

Some operations lock tables and can cause downtime:

- Adding NOT NULL constraint (use pattern from #1)
- Adding foreign key on large table
- Creating non-concurrent indexes

**Use**:

```sql
-- For indexes on large tables
CREATE INDEX CONCURRENTLY idx_large_table_column ON large_table(column);

-- For foreign keys, consider batching or doing during maintenance window
```

---

## Common Issues and Solutions

### Issue: "Can't locate revision identified by..."

**Cause**: Migration history mismatch between code and database.

**Solution**:

```bash
# Check current database version
migrate -path migrations -database "${DATABASE_URL}" version

# Check migration history
ls -la migrations/

# If database is ahead, force to specific version:
migrate -path migrations -database "${DATABASE_URL}" force <version>

# If database is behind, upgrade:
migrate -path migrations -database "${DATABASE_URL}" up
```

### Issue: Auto-generate creates unwanted changes

**Cause**: Model definitions don't match database exactly.

**Solution**:

- Review migration files
- Exclude tables from auto-generation if using ORM:
- For golang-migrate, write SQL migrations manually for full control

```go
// Example: Manual migration control
// Always review generated SQL before applying
```

### Issue: Migrations fail in production

**Cause**: Production database state differs from dev.

**Solution**:

- Always test migrations against production-like data
- Use feature flags to deploy code before migrations
- Have rollback plan ready

---

## Migration Checklist

Before merging migrations to main:

- [ ] Migration reviewed (not just auto-generated)
- [ ] Downgrade tested and works
- [ ] No breaking changes or backwards-compatible pattern used
- [ ] Indexes added where needed
- [ ] Foreign keys have proper ON DELETE behavior
- [ ] Data migrations tested on realistic data volume
- [ ] No long-running locks on large tables
- [ ] Comments explain non-obvious changes
- [ ] RLS policies updated if new tables added

---

## Emergency Rollback

If a migration causes issues in production:

**1. Rollback Code First**:

```bash
git revert <commit>
git push
# Deploy old code
```

**2. Then Rollback Migration**:

```bash
# Connect to production database
migrate -path migrations -database "${DATABASE_URL}" down 1

# Verify
migrate -path migrations -database "${DATABASE_URL}" version
```

**3. Investigate and Fix**:

- Review migration logs
- Test fix locally
- Create new migration with fix

---

## Next Steps

1. Review [Testing Strategy](../development/testing-strategy.md) for migration testing
2. Review [Development Workflow](../guides/development-workflow.md) for migration integration
3. Start creating initial migrations for v0.5 schema
