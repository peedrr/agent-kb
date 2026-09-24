# MAINTAIN — Audit and Repair KB Health

**When:** "lint", "health check", "find gaps", "what needs fixing", or after batch ingests and template changes.

**Selection:** every command here picks its KB per invocation with `--kb <path>` or the `AKB_KB` environment variable; there is no stored default (`akb discover` lists nearby bases).

## Procedure

1. Run all lint checks:
   ```bash
   akb lint
   ```

   What lint detects:

   | Check | What it means |
   |-------|---------------|
   | `broken_links` | A wikilink points to a page that doesn't exist |
   | `orphans` | A page has zero inbound links — nothing connects to it |
   | `empty_pages` | A page has no body content |
   | `missing_frontmatter` | A page has no YAML frontmatter |
   | `required_fields` | A page leaves a schema-required frontmatter field unset |
   | `index_consistency` | A page is listed in the index but doesn't exist, or vice versa |
   | `type_orphan` | A page's `type` has no matching template — no enforced structure |
   | `cel_lint` | A page violates its template's `lint_rules` (skipped for type-orphans) |
   | `citations` | A page's `sources` field references a raw file that isn't tracked |
   | `provenance` | A page's confidence markers have drifted too far from its frontmatter |

2. Inspect specific issues:
   ```bash
   akb orphans                 # pages with zero inbound links
   akb links <path>            # outbound, broken, and ambiguous links for a page
   akb status                  # overview: name, path, page count
   ```

   **Freshness:** there is no standalone freshness command. Freshness policy is template-owned: express staleness as CEL `lint_rules` on the page type's template, and it is reported by the `cel_lint` check above (see `references/TEMPLATE.md`).

3. Fix issues:
   - **Broken links:** Create stub pages for missing targets, or fix the source page's wikilink.
   - **Orphans:** Add inbound links from related pages.
   - **Ambiguous links:** Use `[[path-form]]` instead of short names.
   - **Stale pages:** Update outdated content, or mark as deprecated. Staleness is template-owned CEL lint policy, not a separate command.
   - **Type-orphans:** Reassign the page's `type` to a valid template, create the missing template, or delete the page. See `references/TEMPLATE.md`.

4. Batch orphan cleanup:
   ```bash
   akb delete --orphans          # preview what would be deleted
   akb delete --orphans --force  # actually delete
   ```

   **Warning:** This deletes pages permanently. Preview first.

5. Log maintenance:
   ```bash
   akb log append lint "<description of fixes>"
   ```

## Deleting Pages

If lint identifies pages that should be removed:

```bash
akb delete <path>
akb lint  # verify no broken links remain
```

## Type-Orphan Recovery

When lint reports `[type_orphan] error`:

1. `akb template list` — see available templates
2. Reassign: `akb write <path> --frontmatter type=<existing-type>`
3. Or create the missing template (see `references/TEMPLATE.md`)
4. Or delete the page if no longer needed
5. `akb lint` — verify zero `type_orphan` errors remain

## Rules

- `akb lint` exits 0 for warnings only, 1 if errors found.
- Use `akb lint --json` for structured output.
- Run `akb lint` after batch ingests and template changes.
- `type_orphan` errors indicate structural gaps — pages without templates will drift. Fix them promptly.
