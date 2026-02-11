#!/bin/bash

# Check database state
cd "$(dirname "$0")"

DB_PATH="data/controlplane.db"

if [ ! -f "$DB_PATH" ]; then
  echo "❌ Database not found at $DB_PATH"
  echo "Run 'dotnet run' to create it, or 'dotnet ef database update' to apply migrations."
  exit 1
fi

echo "==> Database found: $DB_PATH"
echo ""

echo "==> Checking tables..."
sqlite3 "$DB_PATH" ".tables"
echo ""

echo "==> Checking Agents table schema..."
sqlite3 "$DB_PATH" ".schema Agents"
echo ""

echo "==> Checking migration history..."
sqlite3 "$DB_PATH" "SELECT MigrationId FROM __EFMigrationsHistory ORDER BY MigrationId;"
echo ""

echo "==> Checking for OS column..."
if sqlite3 "$DB_PATH" "PRAGMA table_info(Agents);" | grep -q "OS"; then
  echo "✅ OS column exists in Agents table"
else
  echo "❌ OS column NOT found in Agents table"
  echo "   Run './fix-migrations.sh' to recreate database with all migrations"
fi
echo ""

echo "==> Counting registered agents..."
AGENT_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM Agents;")
echo "Total agents: $AGENT_COUNT"

if [ "$AGENT_COUNT" -gt 0 ]; then
  echo ""
  echo "==> Registered agents:"
  sqlite3 "$DB_PATH" "SELECT Id, Hostname, RegisteredAt, LastHeartbeatAt, CurrentHealth, OS FROM Agents;" | column -t -s '|'
fi

echo ""
echo "==> Database check complete"
