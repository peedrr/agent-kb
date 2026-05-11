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

4. Update the index and log:
   ```bash
   akb index add "overview.md" "KB purpose and open questions"
   akb log append ingest "Initialized KB <kb-name>"
   ```

## Notes

- `akb init` sets up everything — structure, default templates, search index, and version control.
- Run `akb init` from the directory where you want the KB to live.
