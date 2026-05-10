# internal/lint/

**Parent:** `./AGENTS.md`

## OVERVIEW

Lint engine with 9 checkers across 3 categories: structural, template-driven, semantic.

## FILES

| File | Checker | Category |
|------|---------|----------|
| `engine.go` | LintEngine orchestrator | - |
| `thresholds.go` | Hardcoded constants | - |
| `broken_links.go` | broken_links | structural |
| `orphans.go` | orphans | structural |
| `empty_pages.go` | empty_pages | structural |
| `missing_frontmatter.go` | missing_frontmatter | structural |
| `index_consistency.go` | index_consistency | structural |
| `citations.go` | citations | semantic |
| `provenance.go` | provenance | semantic |
| `cel.go` | cel_lint | template-driven |
| `type_orphan.go` | type_orphan | structural |

## KEY TYPES

| Type | Purpose |
|------|---------|
| `LintEngine` | Orchestrates checkers, produces `LintReport` |
| `LintChecker` | Interface: `Name()` + `Check(ctx, *KB) ([]LintIssue, error)` |
| `LintIssue` | `Type`, `RuleID`, `Message`, `Path`, `Severity` |
| `LintReport` | `Issues`, `PagesChecked`, `ByCheck` |
| `KB` | Lint context: root path, linkgraph, templates, manifest, parsed pages |
| `PageData` | Parsed page with content, frontmatter, provenance markers, annotations |

## THRESHOLDS (HARDCODED)

```go
ProvenanceDriftThreshold = 0.20   // |frontmatter - inline| ratio
```

## NOTES

- All structural checkers query SQLite `links` table (not filesystem scan) for broken_links/orphans
- `cel_lint` evaluates CEL `lint_rules` from templates with `now` variable injection
- `citations` validates frontmatter `sources` against raw manifest entries
- Provenance drift: `|frontmatter_confidence - inline_marker_ratio|` > 0.20
- `akb lint --json` outputs `LintReport` JSON with `rule_id` field
- Retired checkers (replaced by CEL): `type_exists`, `frontmatter_schema`, `category_dirs`
- Removed checkers: `freshness`, `confidence`, `summary_length`
- `type_orphan` detects pages whose `type` frontmatter has no matching template in `.akb/templates/`; excludes pages without frontmatter or empty type; severity: error
