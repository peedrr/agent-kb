# INGEST — Add Knowledge to the KB

**When:** User provides documents, URLs, pasted text, or says "add this to the wiki".

## Critical Rule

Raw sources and KB pages are **separate**. Never store undistilled source text as a KB page.

## Procedure

1. **Save the raw source** (immutable once stored):
   ```bash
   akb raw write <filename> <<'EOF'
   <full content>
   EOF
   ```

2. **Sync the manifest** so the KB can track the raw source:
   ```bash
   akb raw sync
   ```

3. **Check the template requirements** before writing any KB page:
   ```bash
   akb template get <type-name>
   ```
   This shows required frontmatter fields and content requirements. `type` and `title` are mandatory for all pages.

4. **Distill into a KB page.** Use `akb template list` to see available types. Include ALL required frontmatter fields for the chosen type:
   ```bash
   akb write <page-name>.md <<'EOF'
   ---
   title: <Page Title>
   type: <type-name>
   tags: [<tag1>, <tag2>]
   summary: <One or two sentences describing this page>
   sources: [<filename>]
   created: <YYYY-MM-DD>
   ---
   
   # <Page Title>
   
   <distilled content>
   
   See also [[<existing-page-1>]], [[<existing-page-2>]].
   EOF
   ```

5. **Update bookkeeping:**
   ```bash
   akb index add "<page-name>.md" "<one-line summary>"
   akb log append ingest "Ingested <source>; created <page-name>.md"
   ```

## Rules

- Always check `akb template get <type>` before writing. Writing without understanding requirements causes validation failures.
- Add 2–3 `[[wikilinks]]` to existing pages on every new page.
- After creating pages, search for mentions of the new topic and add backlinks where missing: `akb search "<topic>"`.
- If validation fails, `akb write` prints the missing fields. Fix and retry.
- New pages are drafts by default. Use `akb approve` to publish.
