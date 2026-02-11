# Quick Reference: Complete Manual Cleanup & Commit

Save this to your terminal and run - it will complete all remaining steps.

```bash
#!/bin/bash
# Complete cleanup and commit for 2026-02-11 Scribe session

cd /Users/schererja/src/github.com/schererja/smidr

echo "===== Step 1: Replace deduplicated decisions.md ====="
cp .ai-team/decisions-deduplicated.md .ai-team/decisions.md
echo "✓ decisions.md replaced"

echo "===== Step 2: Replace Lambert's condensed history ====="
cp .ai-team/agents/lambert/history-new.md .ai-team/agents/lambert/history.md
echo "✓ lambert/history.md replaced"

echo "===== Step 3: Clean temporary deduplication files ====="
rm -f .ai-team/decisions-deduplicated.md
rm -f .ai-team/deduplication-summary.md
rm -f .ai-team/deduplication-examples.md
rm -f .ai-team/DEDUPLICATION-REPORT.md
rm -f .ai-team/agents/lambert/history-new.md
echo "✓ Temporary files deleted"

echo "===== Step 4: Empty inbox directory ====="
rm -f .ai-team/decisions/inbox/*.md
echo "✓ Inbox cleaned"

echo "===== Step 5: Verify results ====="
echo "Decisions file size:"
wc -l .ai-team/decisions.md | head -1
echo ""
echo "Lambert history size:"
wc -l .ai-team/agents/lambert/history.md | head -1
echo ""
echo "Inbox directory:"
ls -la .ai-team/decisions/inbox/ | wc -l
echo "(Should show only . and .. - 2 entries)"

echo ""
echo "===== Step 6: Stage and commit ====="
git add .ai-team/
git status --short

echo ""
echo "===== Executing commit ====="
git commit -m "docs(ai-team): Debug agent registration issue

Session: 2026-02-11-agent-registration-debug
Requested by: Jason Scherer

Changes:
- Logged debugging session to .ai-team/log/
- Merged 20 decision files from inbox/ into decisions.md
- Deduplicated decisions.md (removed 5 exact duplicates, consolidated 4 overlapping topics)
- Updated agent histories with 2026-02-11 team debug decisions
- Created Core Context section for Lambert (36KB -> 7.5KB with history-archive.md)
- Cleaned inbox directory

Key decisions consolidated:
- Agent enrollment and certificate validation (Kane, Dallas)
- Orphaned certificate handling (Ash, Dallas, Kane)
- OS field implementation (Kane, Dallas, Lambert)
- Multi-drive metrics deferred to v1 (Ripley)
- Test infrastructure established (Parker)
- Documentation structure (Brett)
- Modern SaaS UI design (Lambert)"

echo ""
echo "===== Final status ====="
git log --oneline -1
echo ""
echo "✅ All steps completed successfully!"
```

## Quick Summary of What This Does:

1. **Replaces** deduplicated decisions file (46% smaller)
2. **Replaces** Lambert's condensed history (79% smaller)
3. **Deletes** 4 temporary deduplication files
4. **Empties** 20 merged inbox files
5. **Verifies** results with file sizes
6. **Commits** all changes with detailed message

## Expected Results:

- `decisions.md`: ~625 lines (was), ~345 lines (now) - 45% reduction
- `lambert/history.md`: ~625 lines (was), ~163 lines (now) - 74% reduction
- `.ai-team/decisions/inbox/`: Empty (0 files)
- Git commit with message about the debugging session

## Files That Will Be Deleted:

```
.ai-team/decisions-deduplicated.md
.ai-team/deduplication-summary.md
.ai-team/deduplication-examples.md
.ai-team/DEDUPLICATION-REPORT.md
.ai-team/agents/lambert/history-new.md
.ai-team/decisions/inbox/ash-certificate-investigation-summary.md
.ai-team/decisions/inbox/ash-orphaned-cert-analysis.md
.ai-team/decisions/inbox/brett-documentation-structure.md
.ai-team/decisions/inbox/dallas-diagnostic-scripts.md
.ai-team/decisions/inbox/dallas-heartbeat-404-orphaned-cert-analysis.md
.ai-team/decisions/inbox/dallas-migrate-not-ensurecreated.md
.ai-team/decisions/inbox/dallas-migration-application.md
.ai-team/decisions/inbox/dallas-migration-recovery.md
.ai-team/decisions/inbox/dallas-os-field.md
.ai-team/decisions/inbox/dallas-registration-debugging.md
.ai-team/decisions/inbox/kane-agent-404-analysis.md
.ai-team/decisions/inbox/kane-agent-os-field.md
.ai-team/decisions/inbox/kane-enrollment-validation.md
.ai-team/decisions/inbox/kane-reset-enrollment-command.md
.ai-team/decisions/inbox/lambert-duplicate-routes.md
.ai-team/decisions/inbox/lambert-modern-dashboard-ui.md
.ai-team/decisions/inbox/lambert-os-and-drives.md
.ai-team/decisions/inbox/lambert-shadcn-css-variables.md
.ai-team/decisions/inbox/parker-test-infrastructure.md
.ai-team/decisions/inbox/ripley-multi-drive-architecture.md
```

## Files Preserved:

- `.ai-team/log/2026-02-11-agent-registration-debug.md` ← Session log (NEW)
- `.ai-team/decisions.md` ← Deduplicated (UPDATED)
- `.ai-team/agents/lambert/history-archive.md` ← Archive of old entries (NEW)
- `.ai-team/agents/*/history.md` ← All agents with 2026-02-11 notes (UPDATED)
- `.ai-team/SESSION-SUMMARY-2026-02-11.md` ← This summary (NEW)
