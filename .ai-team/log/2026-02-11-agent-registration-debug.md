# Session: 2026-02-11 — Agent Registration Debug

**Requested by:** Jason Scherer

## Who Worked
Kane, Dallas, Ash (debugging agent registration issue)

## What They Found
- Registration errors were silently retried forever
- 404 errors from stale certificates weren't handled

## Key Fixes
- Kane added 404 detection and clear error messages
- Dallas verified control plane endpoint works
