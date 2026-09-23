# UPDATE — Revise Existing Pages

**When:** User wants edits, corrections, merges, or cascade updates.

**Selection:** every command here picks its KB per invocation with `--kb <path>` or the `AKB_KB` environment variable; there is no stored default (`akb discover` lists nearby bases).

## Procedure

1. **Read the current page:**
   ```bash
   akb read <path>
   ```

2. **Check the template requirements** before updating:
   ```bash
   akb template get <type>
   ```
   Updates can inadvertently violate validations that were satisfied in the original page. If the template requires a section you might remove, validation will fail.

3. **Make the update** using one of these methods:

   **Append content** (preserves frontmatter, adds to body):
   ```bash
   akb write <path> --append <<'EOF'
   ## <New Section>
   
   <new content>
   EOF
   ```

   **Rewrite the page** (replaces entire file — include full frontmatter and content):
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

   **Partial frontmatter update** (preserves body and all other frontmatter fields):
   ```bash
   akb write <path> --frontmatter summary="Updated summary" --frontmatter tags="[go, cli]"
   ```

4. **Update the index and log:**
   ```bash
   akb index add "<path>" "<updated summary>"
   akb log append update "Updated <path>; <reason>"
   ```

5. **Assess related pages** for impacts:
   - Does a linked page need to reference the new development?
   - Are there contradictions with existing claims?

## Deleting Pages

If restructuring requires removing a page:

```bash
akb delete <path>
akb lint  # check for broken links
```

Fix any broken links by updating pages that referenced the deleted page.

## Rules

- `akb write <path> --append` preserves frontmatter and adds content to the body.
- `akb write <path>` replaces the entire file. Include full frontmatter and full content.
- `akb write <path> --frontmatter key=val` updates only specified frontmatter fields. The page must already exist.
- Always check `akb template get <type>` before updating — validation may reject your change.
