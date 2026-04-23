# APPROVE — Publish Draft Pages

**When:** A page has quality annotations or the user says "approve", "publish", or "this looks good".

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

## What Approve Does

- Strips quality annotations (`<!-- olw-auto: ... -->` HTML comments)
- Strips provenance markers (`^[inferred]`, `^[ambiguous]`, `^[extracted]`)
- Sets `is_draft: false` in frontmatter
- Commits the change

## Rules

- Do not approve pages with unresolved `^[ambiguous]` claims unless the user explicitly confirms.
- Ambiguity is signal, not noise — resolve it before approving.
- New pages are drafts by default. Approve them once reviewed.
