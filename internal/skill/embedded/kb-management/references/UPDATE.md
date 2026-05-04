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

3. Append via write command (preferred)

   `akb write --append` is the preferred way to append content:
   ```bash
   akb write <path> --append <<'EOF'
   ## <New Section>

   <new content>
   EOF
   ```

   The standalone `akb append` command still works but `akb write --append` is preferred.

4. If changing frontmatter or rewriting the page, use write with the FULL content:
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

5. Partial frontmatter updates

   To update only specific frontmatter fields without rewriting the body:
   ```bash
   akb write <path> --frontmatter summary="Updated summary" --frontmatter tags="[go, cli]"
   ```

   This preserves the body and all other frontmatter fields. The page must already exist.

6. Update the index:
   ```bash
   akb index add "<path>" "<updated summary>"
   ```

7. Log the update:
   ```bash
   akb log append update "Updated <path>; <reason>"
   ```

## Re-checking Template Requirements

Before updating an existing page, always re-check the template requirements:

```bash
akb template get <type>
```

**Why this matters:** Updates can inadvertently violate validations that were satisfied in the original page. For example, if the template requires a "## References" section, removing that section during editing will cause validation to fail—even though the original page was valid.

**Workflow pattern:**

1. `akb read <path>` — read current content and note the `type` from frontmatter
2. `akb template get <type>` — review schema fields and requirements
3. Understand constraints — ensure your update won't remove required elements
4. `akb write <path> --frontmatter key=val` or `akb write <path> --append` — make the update

**Example:** A page has `type: reference` with a requirement that body must contain at least one wikilink. If you rewrite the page and forget to include the wikilink, validation will fail. Running `akb template get reference` beforehand reminds you to preserve that requirement.

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
