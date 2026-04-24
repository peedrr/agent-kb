# internal/lint/

**Parent:** `./AGENTS.md`

## OVERVIEW

Lint engine with 15 checkers across 3 categories: structural, template-driven, semantic.

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
| `type_exists.go` | type_exists | template-driven |
| `frontmatter_schema.go` | frontmatter_schema | template-driven |
| `category_dirs.go` | category_dirs | template-driven |
| `citations.go` | citations | template-driven |
| `confidence.go` | confidence | template-driven |
| `summary_length.go` | summary_length | template-driven |
| `provenance.go` | provenance | semantic |
| `freshness.go` | freshness | semantic |

## KEY TYPES

| Type | Purpose |
|------|---------|
| `LintEngine` | Orchestrates checkers, produces `LintReport` |
| `LintChecker` | Interface: `Name()` + `Check(ctx, *KB) ([]LintIssue, error)` |
| `LintIssue` | `Type`, `Message`, `Path`, `Severity` |
| `LintReport` | `Issues`, `PagesChecked`, `ByCheck` |
| `KB` | Lint context: root path, linkgraph, templates, manifest, parsed pages |
| `PageData` | Parsed page with content, frontmatter, provenance markers, annotations |

## THRESHOLDS (HARDCODED)

```go
ProvenanceDriftThreshold = 0.20   // |frontmatter - inline| ratio
FreshnessHalfLifeDays    = 30     // days for freshness decay
FreshnessScoreThreshold  = 50.0   // flag pages below this
SummaryMinLength         = 10     // chars
SummaryMaxLength         = 200    // chars
```

## NOTES

- All checkers query SQLite `links` table (not filesystem scan) for broken_links/orphans
- Template-driven checks use dynamically loaded templates, NOT hardcoded enum
- Provenance drift: `|frontmatter_confidence - inline_marker_ratio|` > 0.20
- Freshness: `score = 100 * 2^(-days/30) * confidence_weight`; flag if < 50
- `akb lint --json` outputs `LintReport` JSON