# Session Log: 2026-02-11-stale-cert-fix

**Date:** 2026-02-11  
**Requested by:** Jason Scherer  
**Agent:** Kane

## Problem
Agent failing with "certificate is stale" error — required manual deletion.

## Solution
Implemented automatic re-enrollment logic in daemon.go:
- Created new reset.go file with DeleteEnrollmentFiles() function
- When agent detects "not registered" error, it now auto-cleans stale certs and re-enrolls (up to 3 attempts)

## Outcome
User confirmed: "It is working again"
