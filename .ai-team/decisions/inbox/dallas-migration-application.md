### 2026-02-11: Migration Application Required for OS Column

**By:** Dallas

**What:** The AddOSToAgent migration (20260211000000) exists but needs to be manually applied to the database before the control plane can use the OS field.

**Why:** EF Core migrations create the schema change code but don't automatically apply it to the database. The migration file was created in a previous session but wasn't applied, causing the "no such column: a.OS" SQLite error at runtime.

**Action Required:**
```bash
cd control-plane
dotnet ef database update
```

**Technical Details:**
- Migration file: `Migrations/20260211000000_AddOSToAgent.cs`
- Adds nullable TEXT column "OS" to Agents table
- Database location: `control-plane/data/controlplane.db` (SQLite)
- Column is nullable to support existing agents without OS data
