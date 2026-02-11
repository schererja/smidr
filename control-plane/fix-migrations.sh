#!/bin/bash

# Fix database migration state
cd "$(dirname "$0")"

echo "==> Deleting existing database..."
rm -f data/controlplane.db data/controlplane.db-shm data/controlplane.db-wal

echo "==> Running EF Core migrations..."
dotnet ef database update

echo "==> Verifying migrations applied..."
dotnet ef migrations list

echo "==> Done! Database recreated with all migrations applied."
