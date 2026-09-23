# internal/template/

**Parent:** `./AGENTS.md`

## OVERVIEW

TemplateV2 loader for typed page templates with CEL validations and lint rules. Replaces old `required[]`/`optional[]`/`body` format.

## FILES

| File | Purpose |
|------|---------|
| `template.go` | TemplateV2 struct, loader, old-format detection, embedded defaults |
| `embedded/adr.yaml` | Default ADR template with validations + lint_rules |
| `embedded/adr_pass.md` | Valid ADR mockup |
| `embedded/adr_fail.md` | Invalid ADR mockup (fails valid_status) |
| `embedded/note.yaml` | Default note template with validations + lint_rules |
| `embedded/note_pass.md` | Valid note mockup |
| `embedded/note_fail.md` | Invalid note mockup |

## KEY TYPES

| Type | Purpose |
|------|---------|
| `Template` | `Name`, `Description`, `Dir`, `Schema`, `Validations`, `LintRules` |
| `Schema` | `Frontmatter map[string]FieldSchema` |
| `FieldSchema` | `Type` (string/list), `Required` (bool), `Enum` ([]string) |
| `ValidationRule` | `ID`, `Rule` (CEL expr), `Requirement` (human-readable), `Expect` |
| `LintRule` | `ID`, `Rule` (CEL expr), `Severity` (warning/error), `Expect` |

## TEMPLATEV2 SCHEMA

```yaml
name: <template-name>
description: <description>
dir: <directory>
schema:
  frontmatter:
    <field>:
      type: string|list
      required: true|false
      enum: [<values>]
validations:
  - id: <rule-id>
    rule: <CEL-expression>
    requirement: <human-readable>  # shown in Writer View
    expect: <description>
lint_rules:
  - id: <rule-id>
    rule: <CEL-expression>
    severity: warning|error
    expect: <description>
```

## KEY FUNCTIONS

| Function | Purpose |
|----------|---------|
| `LoadTemplates(dir)` | Reads `.yaml` files from directory; rejects old format |
| `CopyDefaults(targetDir)` | Extracts embedded defaults (`.yaml` + `_pass.md` + `_fail.md`) |
| `DefaultFS` / `DefaultTemplates` | Embedded defaults (`//go:embed embedded/*`) that `CopyDefaults` reads |

## NOTES

- Old format detection: rejects YAML with `required`, `optional`, or `body` keys
- `//go:embed embedded/*` includes `.yaml` and `.md` mockup files
- Templates loaded from `.akb/templates/` per KB
- `get --example` validates mockup against current CEL rules at read-time; warns on stderr if stale rules found (still displays mockup, exit 0)
- `template write --force` bypasses existence warning only; stale mockups rejected with full content in error
- On overwrite, `template write` reuses existing `_pass.md`/`_fail.md` if `--pass`/`--fail` flags omitted
