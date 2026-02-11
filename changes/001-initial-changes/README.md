# Numbered changes convention

This repository follows a "numbered -changes" convention: each change is recorded in its own directory under changes/
with a three-digit numeric prefix followed by a short slug and the "-changes" suffix, for example:

  001-initial-changes/
  002-db-migration-changes/

Each change directory contains:
- change.md: machine-friendly metadata and description
- README.md: human-readable summary and the files changed

Use the helper script scripts/new_change.sh to create new change directories.
