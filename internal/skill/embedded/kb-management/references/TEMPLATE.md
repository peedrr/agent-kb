# TEMPLATE — Author, Read, Update, and Delete Templates

**When:** Create, inspect, modify, or remove a page type. Triggers on: create template, define page type, add schema, design template, update template, change template, modify template, add validation rule, delete template, remove template, template list, what types exist.

**Selection:** every command here picks its KB per invocation with `--kb <path>` or the `AKB_KB` environment variable; there is no stored default (`akb discover` lists nearby bases).

## The Template Mindset

Templates are **contracts**, not documentation. They define what "valid" means for every page of a given type.

- A page **without** a template is an orphan — no enforced structure, it will drift.
- A template **without** mockups is unproven — you cannot know if the rules work.
- A mockup that **doesn't validate** is a lie — the rules have changed but the mockup hasn't.

**Templates are the backbone of the KB.** Every page declares a `type` in its frontmatter. That type MUST map to an existing template. Without that mapping, `akb lint` reports `[type_orphan] error` and the page has no guardrails. Type-orphans are a real failure mode — frontmatter fields go missing, required sections vanish, the KB decays.

## Template Lifecycle

```
Create → Read → Update → Delete
  │        │        │        │
  ▼        ▼        ▼        ▼
write     get     write    delete
+mockups  --example +force  --force
          --full    +reuse   +auto-lint
```

Every transition has safeguards. Respect them.

## Anatomy of a Template

A template YAML has three sections:

```yaml
name: <type-name>
description: <one-line summary>

schema:
  frontmatter:
    <field>:
      type: string|list
      required: true|false
      enum: [<values>]

validations:          # Checked at write-time — failing ANY blocks the write
  - id: <rule-id>
    rule: <CEL-expression>
    requirement: <human-readable>
    expect: <description>

lint_rules:           # Checked during akb lint sweeps — monitor health over time
  - id: <rule-id>
    rule: <CEL-expression>
    severity: warning|error
    expect: <description>
```

A field marked `required: true` must be present in a page's frontmatter: `akb write` and `akb append` refuse a page that leaves one out (exit 1, every missing field named) before any validation rule runs. Presence is all the schema enforces there — a value's type and enum constraints stay with the validation rules.

Every template also requires two mockup files:

- **Pass mockup** (`<name>_pass.md`) — a page that carries every required field and passes all validations (proves rules are satisfiable)
- **Fail mockup** (`<name>_fail.md`) — a page that fails at least one validation (proves rules catch errors)

## CEL Concepts

CEL expressions evaluate against a `page` map built from frontmatter, content, and AST.

Key variables:
- `page` — current page state
- `old_page` — pre-modification state (nil for new pages)
- `now` — current timestamp (injected during lint)

Inspectable fields:
- `page.frontmatter.<field>` — frontmatter values
- `page.ast.headings` — list of `{level, text, line}`
- `page.ast.links` — list of `{target, text, is_wikilink, line}`; `is_wikilink` is true for every wikilink form
- `page.ast.code_blocks` — list of `{language, line}`
- `page.content.word_count`, `page.content.char_count`

Any frontmatter string that parses as RFC3339 or a date-only `2006-01-02` is auto-converted to a timestamp for CEL `timestamp()` and duration math.

### Wikilinks in `page.ast.links`

`is_wikilink` is true for all four wikilink forms, so a rule can inspect KB links without catching plain markdown links:

| Form | `target` | `text` |
|------|----------|--------|
| `[[target]]` | `target` | `target` |
| `[[target\|display]]` | `target` | `display` |
| `[[target#heading]]` | `target` | `heading` |
| `[[display]](dest)` | `dest` as written | `[display]` |

- **Adjacency:** the explicit-destination form needs `(` immediately after `]]`, with no whitespace. `[[Paris]] (the city)` stays a plain wikilink.
- **Precedence:** the destination beats the bracket label — `[[a|b]](concepts/a.md)` has target `concepts/a.md` and display `b`. A `#heading` in the bracket part is discarded when a destination is present.
- **Normalization:** the destination a page is *resolved* through is normalized — surrounding whitespace and angle brackets, a leading `./`, and a trailing `.md` are stripped — so `[[Paris]](concepts/paris.md)` resolves to `concepts/paris`. The CEL `target` is not that normalized value: it is the destination as written (`concepts/paris.md` keeps its `.md`), and for this form `text` is the rendered link label `[Paris]`.
- **Empty destinations:** `[[g]]()`, `[[g]](   )`, and `[[g]](<>)` are plain wikilinks whose target is the bracket target.
- Every wikilink yields exactly one entry, so `.all()` and `.exists()` over `page.ast.links` see each token once.

### Guarding frontmatter reads with has()

`has()` is what keeps a rule working when a key it reads may be absent: an absent key makes the rule vacuously true, while reading that key unguarded fails the rule instead of skipping it. Which reads need the guard depends on who guarantees the key:

- **Required keys** — schema fields marked `required: true` are read **unguarded** in validations: `akb write` refuses a page that leaves one out before any rule runs, so the key is always present. `type` and `title` are guaranteed by every write path, so they need no guard anywhere — lint rules included.
- **Optional keys** — **always guarded** with `has()`, e.g. `!has(page.frontmatter.updated) || now - timestamp(page.frontmatter.updated) < duration("2160h")`. A valid page may omit the key.
- **`old_page`** — **always guarded**: it is nil on create, and the on-disk page may predate presence enforcement or never have passed `akb write` at all (git pull, index rebuild ingestion).
- **Lint rules** — guard **everything except `type` and `title`**: the sweep evaluates pages that may never have passed `akb write`.

An unguarded read is an evaluation error, and the three surfaces react to it differently — the divergence is deliberate:

- **Write** — the rule fails, the write is blocked (exit 1), and the message names the offending rule.
- **Lint sweep** — the error degrades to a per-page `cel_lint` issue and the sweep continues; one broken rule never aborts the run.
- **Template write** — the optional-key-stripping variant check rejects the template, telling the author to guard the key with `has()` or mark it `required: true`.

`akb template write` proves the pass mockup three ways: as given, once with each schema-optional key the mockup supplies removed (`type` and `title` excepted — every write guarantees both), and once with `old_page` set to the mockup itself (a no-op update). A rule that only works at create time fails the write instead of the first real page.

**Known limit:** the stripped variants enumerate schema-declared optional keys only. A rule reading an **undeclared** key is not caught by them — deliberately, because an undeclared-but-guarded read is a legitimate pattern (open frontmatter, cross-cutting convention fields). An unguarded read of an undeclared key fails loudly and precisely at first use: exit 1 at write time naming the rule. The lint sweep evaluates only template `lint_rules`, so the same read surfaces as a `cel_lint` issue at sweep time only when it appears in a `lint_rules` entry.

## Workflow: Start from a Showcase

`akb init` seeds no templates: a KB refuses every write of a type it has no template for. The embedded showcase set is the copy source — it is read on demand with `--examples` and is never consulted at write or lint time.

```bash
akb template list --examples                            # the embedded showcase set: adr, note
akb template get adr --full --examples > /tmp/adr.yaml   # copy the template out
akb template get adr --example --examples > /tmp/adr_pass.md  # and its pass mockup
# edit /tmp/adr.yaml, write a fail mockup, then install:
akb template write adr --template /tmp/adr.yaml --pass /tmp/adr_pass.md --fail /tmp/adr_fail.md
```

`--examples` reads the embedded set instead of the KB, so it works with no base selected and in a KB with no templates of its own. `adr` is the showcase (full CEL validation surface); `note` is the minimal counterpoint.

## Workflow: Create a New Template

Creating a template is a **design act**. You are defining what "correct" means for every future page of this type.

1. **Define schema** — decide required vs optional fields, types, and enums
2. **Write validations** — express what makes a page valid at write-time
3. **Write lint rules** — express what makes a page healthy over time
4. **Create pass mockup** — a page that carries every required field and passes all validations
5. **Create fail mockup** — a page that fails at least one validation
6. **Write the template:**

```bash
akb template write <name> \
  --template <name>.yaml \
  --pass <name>_pass.md \
  --fail <name>_fail.md
```

**Both `--pass` and `--fail` are required for new templates.** This is non-negotiable — a template without proven mockups may contain rules that are impossible to satisfy or that fail to catch real errors.

7. **Lint existing pages:**

```bash
akb lint
```

## Workflow: Read Template Requirements

**ALWAYS read template requirements before writing ANY page.** This is not optional — `akb write` enforces validations, and writing without understanding requirements causes validation failures.

```bash
akb template get <type-name>              # Schema fields + requirements (use before writing)
akb template get <type-name> --example    # Pass mockup showing a valid page (use as model)
akb template get <type-name> --full       # Complete YAML with CEL rules (use when updating)
```

**Every page write follows this pattern:**

```
akb template get <type>  →  read requirements  →  craft content  →  akb write <path>
```

**If `--example` shows a staleness warning:** The template rules have changed since the mockup was updated. The content is still displayed but may no longer be a valid example. Check the current schema with `akb template get <type>`.

## Workflow: Update an Existing Template

Updating a template is **destructive by implication** — it may change what counts as "valid" for every existing page of that type.

### 1. Export and edit

```bash
akb template get <name> --full > /tmp/<name>.yaml
# Edit /tmp/<name>.yaml as needed
```

### 2. Write without --force (see impact)

```bash
akb template write <name> --template /tmp/<name>.yaml \
  --pass <name>_pass.md --fail <name>_fail.md
```

If the template exists, this will:
- Show a YAML diff between old and new
- Show how many pages use the type
- Print the exact `--force` command to proceed
- **Refuse to write** (exit 1)

### 3. Review the diff and page count

The diff shows exactly what will change. The page count shows how many existing pages are affected. If you are tightening validations, some existing pages may become invalid.

### 4. Overwrite with --force

```bash
akb template write <name> --template /tmp/<name>.yaml \
  --pass <name>_pass.md --fail <name>_fail.md --force
```

**Mockup reuse:** Omit `--pass`/`--fail` to reuse existing mockups. They will be validated against the new rules — if they still pass, the overwrite proceeds. If they fail, you MUST provide updated mockups — stale mockups are ALWAYS rejected, even with `--force`.

```bash
# Reuse existing mockups (when rules are unchanged)
akb template write <name> --template /tmp/<name>.yaml --force
```

### 5. Lint affected pages

```bash
akb lint
```

## Workflow: Delete a Template

Deleting a template is **high-impact** — every page of that type becomes a type-orphan with no enforced structure.

### 1. Check impact

```bash
akb template delete <name>
```

Without `--force`, this shows how many pages use the type and the exact command to proceed, then **refuses to delete** (exit 1).

### 2. Force delete (if intentional)

```bash
akb template delete <name> --force
```

This deletes the template and its mockups, then **automatically runs `akb lint`** and reports how many pages are now type-orphans.

Use `--no-commit` to skip the auto-commit (for batching):

```bash
akb template delete <name> --force --no-commit
```

### 3. Address type-orphans

After deleting, lint will show `[type_orphan] error` for every page that had the deleted type. You must either:

- **Reassign** — `akb write <path> --frontmatter type=<existing-type>`
- **Replace** — create a new template for the deleted type (see Create workflow)
- **Delete** — `akb delete <path>` if the page is no longer needed

### 4. Re-lint

```bash
akb lint
```

Verify zero `type_orphan` errors remain.

## DO NOT

- **DO NOT write a page without checking its template first.** `akb template get <type>` before every write. Failing validations waste your time and the user's.
- **DO NOT create a template without mockups.** `--pass` and `--fail` are required for new templates. Without them, you cannot prove your rules work.
- **DO NOT use `--force` to bypass mockup validation.** `--force` bypasses ONLY the overwrite confirmation. Stale mockups are ALWAYS rejected — there is no flag to skip mockup validation.
- **DO NOT delete a template without addressing the type-orphans.** Every page of that type becomes unstructured. Reassign, replace, or delete those pages.
- **DO NOT ignore stale mockup warnings.** When `akb template get --example` warns that the mockup no longer validates, fix the mockup.
- **DO NOT use the old template format.** Templates with `required:`, `optional:`, or `body:` keys are rejected. Use `schema.frontmatter`, `validations`, and `lint_rules`.
- **DO NOT add `--type` flag on `akb write` or `akb append`.** Type comes from frontmatter, not from a flag.
- **DO NOT skip `akb lint` after template changes.** Changing or deleting a template affects every page of that type. Lint reveals the impact.

## DO

- **DO run `akb template get <type>` before every `akb write`.** This is your pre-flight check.
- **DO use `akb template get <name> --example` as a model for new pages.** The pass mockup shows a valid page.
- **DO use `akb template get <name> --full` when updating a template.** Export, edit, re-write.
- **DO provide both mockups when creating a new template.** Prove your rules are satisfiable and catch real errors.
- **DO reuse existing mockups on overwrite when rules are unchanged.** Omit `--pass`/`--fail` — existing mockups are loaded and validated automatically.
- **DO review the diff and page count before overwriting.** The overwrite protection shows the blast radius for a reason.
- **DO run `akb lint` after any template change.** This catches type-orphans and pages that no longer validate.
- **DO start from a showcase when the KB has no template for the type.** `akb template get <name> --full --examples` copies one out; see *Start from a Showcase*.
- **DO start minimal (like `note`), then add rigor as needed.** Over-constraining creates friction. Under-constraining causes drift.
- **DO use `enum` for fields with a closed set of values** (e.g., `status: [proposed, accepted, deprecated]`).
- **DO keep `requirement` strings human-readable** — they appear in `akb template get` output and help the next agent understand intent.
