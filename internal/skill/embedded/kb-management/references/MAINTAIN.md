# MAINTAIN — Audit and Repair KB Health

**When:** "lint", "health check", "find gaps", "what needs fixing", or after a batch of ingests.

## Procedure

1. Run all lint checks:
   ```bash
   akb lint
   ```

2. Inspect specific issues:
   ```bash
   akb orphans                 # pages with zero inbound links
   akb links <path>            # outbound, broken, and ambiguous links
   akb stale                   # freshness report across all registered KBs
   akb status                  # overview: name, path, page count, git status
   ```

3. Fix issues:
   - **Broken links:** Create stub pages for missing targets, or fix the source page's wikilink.
   - **Orphans:** Add inbound links from related pages.
   - **Ambiguous links:** Use `[[path-form]]` instead of short names.
   - **Stale pages:** Update outdated content, or mark as deprecated.

4. Log maintenance:
   ```bash
   akb log append lint "<description of fixes>"
   ```

## Deleting Pages (Cleanup)

If lint identifies pages that should be removed (e.g., empty pages, duplicates):

1. Delete the page:
   ```bash
   akb delete <path>
   ```

2. Run lint again to verify no broken links remain:
   ```bash
   akb lint
   ```

## Rules

- `akb lint` exits 0 if no issues, 1 if issues found.
- Use `akb lint --json` for structured output.
- Use `akb stale --json` for structured freshness data.
- Run `akb lint` regularly, especially after batch ingests.
- `akb stale` checks ALL registered KBs, not just the active one.
