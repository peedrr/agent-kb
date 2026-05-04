# INGEST — Add Knowledge to the KB

**When:** User provides documents, URLs, pasted text, or says "add this to the wiki".

## Critical Rule

Raw sources and KB pages are **separate**. Never write raw source text directly into `kb/` without distillation.

## Discover Template Requirements

Before writing ANY KB page, check what the template requires. Different `type` values have different mandatory fields, validations, and linking requirements.

```bash
akb template get <type-name>
```

This shows:
- `schema`: Required frontmatter fields
- `requirements`: What must be present in the page content

**Why this matters:** Writing without checking requirements causes validation failures. `akb write` will reject the page if required fields are missing or content doesn't meet rules.

Pattern:
```
akb template get <type>  →  read requirements  →  craft content  →  akb write <path>
```

## Procedure

1. Save the raw source (immutable):
   ```bash
   akb raw write <filename> <<'EOF'
   <full content>
   EOF
   ```
   Examples: `akb raw write paper.pdf`, `akb raw write notes.txt`.

2. Sync the manifest:
   ```bash
   akb raw sync
   akb raw status
   ```

3. Distill into a KB page. Choose the appropriate `type` (check `.akb/templates/*.yaml` for available types). Include ALL required frontmatter fields for that type. `type` and `title` are mandatory for all pages.
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

4. Update bookkeeping:
   ```bash
   akb index add "<page-name>.md" "<one-line summary>"
   akb log append ingest "Ingested <source>; created <page-name>.md"
   ```

## Rules

- Add 2–3 `[[wikilinks]]` to existing pages on every new page.
- After creating pages, search for mentions of the new topic and add backlinks where missing.
- Use `akb search "<topic>"` to find existing pages for linking.
- `akb write` resolves the output path based on the page's `type` template. Just provide the filename.
- If frontmatter validation fails, `akb write` prints the missing fields. Fix and retry.
- New pages are drafts by default. Use `akb approve` to publish.
