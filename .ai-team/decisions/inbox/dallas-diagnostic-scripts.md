### 2026-02-11: Added Diagnostic Scripts for Registration Troubleshooting

**By:** Dallas

**What:** Created three diagnostic scripts to help troubleshoot agent registration and database issues:
1. `check-database.sh` - Inspects database schema, migration history, and agent records
2. `test-registration.sh` - Tests registration endpoint with synthetic agent
3. `docs/TROUBLESHOOTING-REGISTRATION.md` - Comprehensive guide for diagnosing registration failures

**Why:**
- "Agent not registered" errors are confusing without visibility into database state
- Manual SQL queries are tedious and error-prone
- Need quick way to test registration without running full agent
- Team members (especially Kane working on agent) need to understand registration flow

**Usage:**
```bash
cd control-plane
./check-database.sh           # Inspect database state
./test-registration.sh        # Test registration endpoint
cat docs/TROUBLESHOOTING-REGISTRATION.md  # Read troubleshooting guide
```

**Files created:**
- `control-plane/check-database.sh`
- `control-plane/test-registration.sh`
- `control-plane/docs/TROUBLESHOOTING-REGISTRATION.md`
