### 2026-02-11: Use Migrate() Instead of EnsureCreated() for Database Initialization

**By:** Dallas

**What:** Changed Program.cs to use `db.Database.Migrate()` instead of `db.Database.EnsureCreated()` for database initialization on startup.

**Why:** 
- `EnsureCreated()` creates the database with the current schema but completely ignores EF Core migrations
- When new migrations are added (like AddOSToAgent), existing databases don't get updated
- This causes "no such column" errors when code expects fields added by migrations
- `Migrate()` applies all pending migrations automatically on startup, keeping schema in sync
- Eliminates need for manual `dotnet ef database update` commands during development

**Impact:**
- Control plane now auto-applies migrations on every start
- Developers don't need to remember to run migration commands
- Reduces "my code works but yours doesn't" issues from schema drift

**Files changed:**
- `control-plane/Program.cs`: Line 66 changed from `EnsureCreated()` to `Migrate()`

**Note:** This is safe for development and production. Migrations are idempotent and only apply once.
