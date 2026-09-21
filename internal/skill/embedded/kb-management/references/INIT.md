# INIT — Bootstrap a New KB

**When:** No KB exists; user wants a new knowledge base.

## Procedure

1. Initialize the KB:
   ```bash
   akb init <kb-name>
   ```
   Replace `<kb-name>` with a short identifier (no slashes, no `..`). The KB
   lands at `<working directory>/<kb-name>`. `akb init` does not select it.

2. Verify (every command below addresses the new KB with `--kb`):
   ```bash
   akb --kb <kb-name> status
   akb --kb <kb-name> list
   ```

3. (Optional) Create an overview page:
   ```bash
   akb --kb <kb-name> write overview.md <<'EOF'
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
   akb --kb <kb-name> index add "overview.md" "KB purpose and open questions"
   akb --kb <kb-name> log append ingest "Initialized KB <kb-name>"
   ```

## Notes

- `akb init` sets up everything — structure, default templates, search index, and version control.
- Run `akb init` from the directory where you want the KB to live.
- After `akb init` (which does not select a base), address the KB on every
  invocation with `--kb <path>` or `AKB_KB=<path>`; a relative path resolves
  against the working directory, so `--kb <kb-name>` works from the directory
  you initialized in. See `SKILL.md` for the full selection rules.
