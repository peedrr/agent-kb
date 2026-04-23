# UPDATE — Revise Existing Pages

**When:** User wants edits, corrections, merges, or cascade updates.

## Procedure

1. Read the current page:
   ```bash
   akb read <path>
   ```

2. If adding content to the end, use append:
   ```bash
   akb append <path> <<'EOF'
   ## <New Section>
   
   <new content>
   EOF
   ```

3. If changing frontmatter or rewriting the page, use write with the FULL content:
   ```bash
   akb write <path> <<'EOF'
   ---
   title: <Title>
   type: <type>
   tags: [<tag1>, <tag2>]
   summary: <summary>
   sources: [<source1>, <source2>]
   created: <YYYY-MM-DD>
   updated: <YYYY-MM-DD>
   ---
   
   # <Title>
   
   <full merged content>
   EOF
   ```

4. Update the index:
   ```bash
   akb index add "<path>" "<updated summary>"
   ```

5. Log the update:
   ```bash
   akb log append update "Updated <path>; <reason>"
   ```

## Deleting Pages (Restructuring)

If restructuring requires removing a page:

1. Delete the page:
   ```bash
   akb delete <path>
   ```

2. Check for broken links:
   ```bash
   akb lint
   ```

3. Fix any broken links by updating pages that referenced the deleted page.

## Rules

- `akb append` preserves frontmatter and adds content to the body.
- `akb write` replaces the entire file. Include the full frontmatter and full content.
- After updating, assess related pages for impacts:
  - Does a linked page need to reference the new development?
  - Are there contradictions with existing claims?
- You cannot delete `kb/index.md` or `kb/log.md`. Use `akb index rebuild` to reset the index.
- You cannot delete raw files with `akb delete`. Use `akb raw delete` for raw files.
