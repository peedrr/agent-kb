# APPROVE — Publish Draft Pages

**When:** A page has been reviewed and the user says "approve", "publish", or "this looks good".

**Selection:** every command here picks its KB per invocation with `--kb <path>` or the `AKB_KB` environment variable; there is no stored default (`akb discover` lists nearby bases).

## Procedure

1. Read the page to verify its current state:
   ```bash
   akb read <path>
   ```

2. Approve the page:
   ```bash
   akb approve <path>
   ```

3. Verify the approval:
   ```bash
   akb read <path>
   ```

   After approval, quality annotations and provenance markers will be stripped from the page. The page will show `is_draft: false`.

## Batch Approval

```bash
akb approve --all-drafts
```

Approves every draft that passes schema required-field validation; already-approved pages are skipped silently and a summary count is printed. A draft that leaves a schema-required frontmatter field unset keeps its draft state, is reported, and the run exits 1.

**When to use batch:** After a batch ingest where all pages have been reviewed.
**When to use single:** When reviewing individual pages for quality.

## Rules

- **DO NOT approve pages with unresolved `^[ambiguous]` claims** unless the user explicitly confirms. Ambiguity is signal, not noise — resolve it before approving.
- New pages are drafts by default. Approve them once reviewed.
