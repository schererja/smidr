# Database Migration Recovery

This script fixes corrupted EF Core migration state by deleting the SQLite database and re-running all migrations from scratch.

## When to Use

Run this script when you encounter migration errors like:
- `SQLite Error 1: 'table "X" already exists'`
- `no such column: Y` (when model has the property but DB doesn't have the column)
- Migration history (`__EFMigrationsHistory`) out of sync with actual schema

## What It Does

1. Deletes existing SQLite database files (`controlplane.db`, `controlplane.db-shm`, `controlplane.db-wal`)
2. Runs `dotnet ef database update` to apply all migrations in order
3. Lists applied migrations to verify success

## Usage

```bash
cd control-plane
chmod +x fix-migrations.sh
./fix-migrations.sh
```

## ⚠️ Warning

This script **deletes the database** and all its data. Only use in development environments.

For production databases:
- Take a backup first
- Manually reconcile migration history
- Test in staging before applying to production

## Expected Result

After running, you should see:
- All migrations applied successfully (InitialCreate, AddOSToAgent, etc.)
- New `controlplane.db` file in `data/` directory
- Control plane starts without errors
- All tables exist with correct schema including OS column

## Troubleshooting

If the script fails:
1. Ensure you're in the `control-plane` directory
2. Check that `dotnet ef` tools are installed: `dotnet tool restore`
3. Verify migrations exist in `Migrations/` folder
4. Check `appsettings.json` has `UsePostgres: false` for SQLite
