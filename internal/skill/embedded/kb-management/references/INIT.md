# INIT — Bootstrap a New KB

**When:** No KB exists; user wants a new knowledge base.

## Procedure

1. Initialize the KB:
   ```bash
   akb init <kb-name>
   ```
   Replace `<kb-name>` with a short identifier (no slashes, no `..`).

2. Verify:
   ```bash
   akb status
   akb list
   ```

3. (Optional) Create an overview page:
   ```bash
   akb write overview.md <<'EOF'
   ---
   title: Overview
   type: note
   summary: Purpose and scope of this knowledge base.
   tags: [overview]
   ---
   
   # Overview
   
   <purpose statement>
   
   ## Open Questions
   EOF
   ```

4. Update the index:
   ```bash
   akb index add "overview.md" "KB purpose and open questions"
   ```

5. Log the initialization:
   ```bash
   akb log append ingest "Initialized KB <kb-name>"
   ```

## Notes

- `akb init` creates `kb/`, `raw/`, `.akb/`, default templates, search index, and a git repo.
- The KB root is the current working directory. Run `akb init` from where you want the KB.
