# TEMPLATE — Design and Validate Templates

**When:** You need a new page type with enforced structure, validation, or health checks.

## Why Templates Matter

Templates turn conventions into enforceable rules. Without them, pages drift. With them, every page of a given type has consistent frontmatter, required sections, and observable health.

The KB ships with two contrasting defaults:
- `note` — minimal schema, mostly optional fields, one stale-date lint rule
- `adr` — strict schema, required fields, enum constraints, heading validations, stale-date lint rule

Use `note` as a model for lightweight types. Use `adr` as a model for rigorous types.

## Template Anatomy

A template is a YAML file in `.akb/templates/` with three sections:

```yaml
schema:
  frontmatter:
    <field>:
      type: string|list
      required: true|false
      enum: [<values>]

validations:
  - id: <rule-id>
    rule: <CEL-expression>
    requirement: <human-readable>
    expect: <description>

lint_rules:
  - id: <rule-id>
    rule: <CEL-expression>
    severity: warning|error
    expect: <description>
```

- `schema` defines frontmatter fields, types, and whether they are required
- `validations` are checked at write-time; failing any blocks the write
- `lint_rules` are checked during `akb lint` sweeps; they monitor health over time

## CEL Concepts

CEL expressions evaluate against a `page` map built from frontmatter, content, and AST.

Key variables:
- `page` — map of the current page state
- `old_page` — map of the pre-modification state (nil for new pages)
- `now` — current timestamp (injected during lint sweeps)

What you can inspect:
- `page.frontmatter.<field>` — frontmatter values
- `page.ast.headings` — list of `{level, text, line}`
- `page.ast.links` — list of `{target, text, is_wikilink, line}`
- `page.ast.code_blocks` — list of `{language, line}`
- `page.content.word_count`, `page.content.char_count`

Date fields (`created`, `updated`) are auto-converted from ISO-8601 strings to timestamps for CEL `timestamp()` and duration math.

## Design Workflow

1. **Define schema** — decide required vs optional fields, types, and enums
2. **Write validations** — express what makes a page valid at write-time
3. **Write lint rules** — express what makes a page healthy over time
4. **Create `_pass.md`** — a mockup that passes all validations
5. **Create `_fail.md`** — a mockup that fails at least one validation
6. **Run `akb templates write`** — validate CEL syntax and both mockups

## Mandate

Agents **MUST** create both `_pass.md` and `_fail.md` and run `akb templates write` before considering a template complete. This proves the rules are syntactically valid, the pass case works, and the fail case catches real errors.

```bash
akb templates write <name> \
  --template <name>.yaml \
  --pass <name>_pass.md \
  --fail <name>_fail.md
```

## Best Practices

- Start minimal (like `note`), add rigor only when needed
- Use `enum` for fields with a closed set of values (e.g., `adr.status`)
- Validation rules should block writes; lint rules should warn about drift
- Keep `requirement` strings human-readable — they appear in `akb template get` output
- Re-run `akb lint` after deploying a new template to assess existing pages
