#!/bin/sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 short-slug (e.g. db-migration)" >&2
  exit 2
fi

slug="$1"
# sanitize slug: lowercase, replace non-alphanum with -, trim leading/trailing -
sanitized=$(printf '%s' "$slug" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g' | sed -E 's/^-+|-+$//g')

script_dir=$(cd "$(dirname "$0")" && pwd)
repo_root=$(cd "$script_dir/.." && pwd)
changes_dir="$repo_root/changes"

mkdir -p "$changes_dir"

# find max existing numeric prefix
max=0
for d in "$changes_dir"/*-*-changes; do
  [ -d "$d" ] || continue
  base=$(basename "$d")
  num=$(printf '%s' "$base" | sed -E 's/^([0-9]+).*/\1/')
  case "$num" in
    ''|*[!0-9]*) continue ;;
    *) n=$num ;;
  esac
  if [ "$n" -gt "$max" ]; then max=$n; fi
done

next=$(expr $max + 1)
next_padded=$(printf "%03d" "$next")
dir_name="$next_padded-$sanitized-changes"
new_dir="$changes_dir/$dir_name"

if [ -e "$new_dir" ]; then
  echo "Error: directory $new_dir already exists" >&2
  exit 1
fi

mkdir -p "$new_dir"

author=$(git config user.name 2>/dev/null || true)
if [ -z "$author" ]; then
  author=$(whoami 2>/dev/null || echo "unknown")
fi

date_iso=$(date -u +%Y-%m-%dT%H:%M:%SZ)

cat > "$new_dir/README.md" <<EOT
# $dir_name

Author: $author
Date: $date_iso

Description:
- Briefly describe the change in this directory.

Files changed:
- list files and paths modified by this change
EOT

cat > "$new_dir/change.md" <<EOT
author: $author
date: $date_iso
description: |
  Write a short description of the change here.

files:
  - path/to/file1
  - path/to/file2
EOT

printf '%s\n' "Created $new_dir"
