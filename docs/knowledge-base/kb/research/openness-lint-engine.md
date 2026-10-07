---
as_of: "2026-09-27"
created: "2026-09-27"
grammar: 1
informs:
- RFC-001
- RFC-002
is_draft: false
provenance: agent-drafted
run: json-cel
scope:
- internal/lint/**
- cmd/akb/**
status: final
summary: "Audit of the lint engine: the checker registry is open in code but closed in the CLI, and each of the 10 built-in checkers is classified by what it forces and whether that is structural."
tags:
- openness
- lint
- audit
- checkers
title: Openness audit — lint engine and its 10 shipped checkers
type: research
updated: "2026-09-27"
---
# OPENNESS AUDIT — Lint Engine and its 10 Shipped Checkers

Seam: `internal/lint/*.go`, `cmd/akb/lint*.go`. Read-only recon; no repo files modified.
Prior research is cited, not re-derived: `seam-periphery.md:48-70` already established the
contract shape and the per-checker data class; `command-inventory.md:39,150` already
established the JSON envelope and its typed-gate dependency; `seam-core-pipeline.md:258-259`
already established `required_fields` ⇄ `schema.frontmatter.required`. This report adds the
**mandate/prescription** lens: what akb forces, who consumes it, and whether it is structural.

---

## Findings (mandate tables)

### A. Engine wiring — the registry is open in code, closed in the CLI

| # | Mandate | Anchor | Detail |
|---|---|---|---|
| A1 | The engine has exactly one checker list, built by the CLI | `internal/lint/engine.go:69-71` (`checkers []LintChecker`), `:80-82` (`NewLintEngine` returns an **empty** engine) | The package ships **no** defaults. `NewLintEngine()` is deliberately empty; nothing in `internal/lint` registers a checker. |
| A2 | The 10 built-ins exist only because `cmd/akb/lint.go:157-167` calls `AddChecker` ten times, in fixed order | `cmd/akb/lint.go:157-167` | Verified by exhaustive grep: `AddChecker`/`NewLintEngine` have **no non-test caller** anywhere else. `RunLint` is the sole producer of a `LintReport`. |
| A3 | No surface to add, remove, rename, or re-severity a checker | `internal/config/config.go:16-21` (`Config` = `name`, `created`, `description` only); `cmd/akb/lint.go:44-46` (`--json` is the only flag); `internal/template/template.go:39-45` (`LintRule` has no checker field) | No `akb.yaml` key, no CLI flag, no template key, no env var. The `LintChecker` interface is exported (`engine.go:37-40`) but unreachable from data. |
| A4 | Checker execution is all-or-nothing: one checker returning an error aborts the whole sweep | `internal/lint/engine.go:98-102` | No per-checker isolation, no partial report. Contrast `cel_lint`'s deliberate per-page degrade (`internal/lint/cel.go:76-85`, documented divergence `AGENTS.md` "CEL rule evaluation"). |
| A5 | Registration set is the public contract via `by_check` | `internal/lint/engine.go:106` sets `ByCheck[name] = len(issues)` for **every** registered checker, including 0-issue ones; `cmd/akb/lint.go:243-251` copies it into the envelope | Consumers can enumerate the checker set from output alone (`.sisyphus/evidence/final-qa/qa-results.md:25` did exactly that; it lists a stale count of 13). |
| A6 | Exit policy: fail (exit 1) iff any issue has severity exactly `"error"` | `cmd/akb/lint.go:221-229` (text) and `:259-267` (JSON), both returning `errLintIssues` (`:29`) | The severity vocabulary is a closed 2-value set in practice; a third value (e.g. `info`) would be reported but never fail. There is no `--fail-on`, `--severity`, or `--check` selector. |
| A7 | A third, duplicated copy of the same exit policy lives in `template delete` | `cmd/akb/template_delete.go:141-152` | `template delete` auto-runs the full sweep and re-implements the "count errors, return `errLintIssues`" logic. Any change to the lint gate has 3 sites. |

### B. Per-checker prescriptions (what each forces, and on what akb-owned fact)

| Checker | Must akb force this? | Hard-depends on (akb-prescribed fact) | Evidence |
|---|---|---|---|
| `broken_links` | A wikilink that does not resolve is an **error**; one that resolves to >1 page is a **warning** | The 3-step resolution algorithm (exact → `kb/` prefix → basename) and the notion that basename collisions are "ambiguous"; the SQLite `links` table as truth | `internal/lint/broken_links.go:38-49` (error), `:51-64` (warning); `internal/linkgraph/sqlite.go:306-312` (`resolveTarget`), `:283-304` (`GetBrokenLinks`/`GetAmbiguousLinks`) |
| `orphans` | Every page must have ≥1 inbound link; a self-link does not count | Inbound-degree semantics over the same linkgraph; stale `resolved_to` (documented) | `internal/lint/orphans.go:26-43` (warning); `internal/linkgraph/sqlite.go:246-280`, incl. the staleness caveat `:246-254` |
| `empty_pages` | Body (post-frontmatter bytes) must be non-empty after `TrimSpace`; pages **without** frontmatter are skipped | The body/envelope split produced by `frontmatter.Parse` | `internal/lint/empty_pages.go:33-48`; body built at `cmd/akb/lint.go:120-127,137` |
| `missing_frontmatter` | Every `.md` under `kb/` (except `index.md`/`log.md`) must carry YAML frontmatter — **error** | The walk filter and the `index.md`/`log.md` exemption list | `internal/lint/missing_frontmatter.go:33-45`; filter `cmd/akb/lint.go:101,105`; note a page whose frontmatter fails YAML parse is also reported with the wording "lacks YAML frontmatter delimiters" (`:120-121` sets `HasFrontmatter=false` on parse error) |
| `required_fields` | Presence of every `schema.frontmatter[key].required: true` key — **error**; presence only, values never inspected | `template.Schema.Frontmatter[].Required` (the *only* consumer of `.Required` besides the write/approve twins) | `internal/lint/required_fields.go:44-91`, `:83`; twins `cmd/akb/required_fields.go:19-29` and `cmd/akb/approve.go:218-231` (cited `seam-periphery.md:9`); retirement path `seam-core-pipeline.md:258-259,268` |
| `index_consistency` | `kb/index.md` must exist (**whole sweep aborts if not**), and its entry set must equal the page set: index-without-file = error, file-without-index = warning | `index.md` as an exhaustive second registry, plus its heading/entry grammar | `internal/lint/index_consistency.go:39-41` (missing index ⇒ error from `index.ReadIndex` ⇒ `engine.Run` abort), `:61` (error), `:73` (warning); grammar `internal/index/index.go:50-110`; `kb/index.md` is init-created (`cmd/akb/init.go:137`) and CLI-only-writable (`cmd/akb/write.go:163-164,287-288,412-413`) |
| `citations` | Frontmatter key **literally `sources`**, holding a string or a list of strings, must name a file present in `raw/files.log` — **error** | The field name `sources`; the manifest as the raw registry; silently ignores any other shape | `internal/lint/citations.go:37` (key literal), `:44-57` (accepted shapes; `default: continue` at `:56-57` means a map/number is invisible), `:60-65` (error); manifest read `cmd/akb/lint.go:81-84`, and an absent/empty `raw/files.log` yields `[]Entry{}` (`internal/manifest/manifest.go:50-53`) ⇒ every `sources` value becomes an error |
| `provenance` | Frontmatter key **literally `provenance`** as a numeric map; each key's declared value must be within **0.20** of `count(^[<same key>]) / non-empty-body-lines`; violation = warning | The `provenance` key name; numeric coercion of the value; the marker regex; the "lines" metric; the 0.20 constant | `internal/lint/provenance.go:39` (key), `:56-69` (ratio + `ProvenanceDriftThreshold`), `:71-79` (`countNonEmptyLines`), `:81-90` (`toFloat64` accepts only `float64|int|int64`); `internal/lint/thresholds.go:11`; marker regex `internal/markdown/provenance.go:17` is **open** (`\^\[([^\]]+?)\]`) so the *vocabulary* is user-defined on both sides — but `StripProvenanceMarkers` still only knows `inferred|ambiguous|extracted` (`:57`) |
| `cel_lint` | Nothing prescribed: rule body, `expect` text and severity all come from the template | Only `frontmatter.type` + a loaded template; `page` + `now`, **no `old_page`** | `internal/lint/cel.go:41-100`; severity passthrough `:88`; `now` once per sweep `:47`; no-`old_page` and degrade policy already established `seam-periphery.md:67-71`, `command-inventory.md:79` |
| `type_orphan` | Every page's `type` must resolve to a template in `.agent-kb/templates/` — **error** | The template registry as the type namespace | `internal/lint/type_orphan.go:34-51`; registry loaded `cmd/akb/lint.go:76-79` |

### C. Every hardcoded constant / severity / identity string in the package

| Anchor | Value | Live? |
|---|---|---|
| `internal/lint/thresholds.go:11` | `ProvenanceDriftThreshold = 0.20` | **live** — sole consumer `provenance.go:63` |
| `internal/lint/thresholds.go:12` | `FreshnessHalfLifeDays = 30` | **dead** — only `engine_test.go:206-208` |
| `internal/lint/thresholds.go:13` | `FreshnessScoreThreshold = 50.0` | **dead** — only `engine_test.go:209-211` |
| `internal/lint/thresholds.go:14` | `SummaryMinLength = 10` | **dead** — only `engine_test.go:212-214` |
| `internal/lint/thresholds.go:15` | `SummaryMaxLength = 200` | **dead** — only `engine_test.go:215-217` |
| `broken_links.go:48` / `:62` | `"error"` / `"warning"` | live |
| `missing_frontmatter.go:42`, `required_fields.go:70`, `type_orphan.go:48`, `citations.go:63`, `index_consistency.go:61`, `cel.go:83` | `"error"` | live |
| `empty_pages.go:46`, `orphans.go:39`, `provenance.go:69`, `index_consistency.go:73` | `"warning"` | live |
| 10 × `Name()` string + 10 × duplicated `Type:` literal | e.g. `orphans.go:22-24` and `:32`; `cel_lint` `cel.go:32-34` + `:66,79` | live — **two literal sites per checker**, no constant, no exported name |
| `internal/lint/engine.go:18-24` | `Severity` is a bare `string`; nothing validates it | live |
| `internal/template/template.go:42` | severity comment `"warning" | "error"` — **comment only**; `templates_write.go:118-122` compiles the CEL rule and never checks severity | ⇒ a template may ship `severity: bogus` or omit it; the issue is still reported but **cannot fail the run** (A6). This is an accidental, undocumented escape valve. |
| `internal/lint/engine.go:66` + `cmd/akb/lint.go:130,139` | `PageData.Annotations` is parsed and carried but **no checker reads it** (exhaustive grep) | dead payload on the lint path |

### D. JSON envelope (frozen contract — noted, not proposed for change)

`cmd/akb/lint.go:233-268`:

```json
{ "issues": [ {"check","rule_id?"|omitted,"message","path","severity"} ],
  "summary": { "total", "pages_checked", "by_check": {"<checker>": <count>, ...} } }
```

- `issues` is forced non-nil so it renders `[]` (`lint.go:246-250`); `by_check` is a Go map, so it renders `{}` when the engine registers nothing.
- `rule_id` is `omitempty` and is only ever set on `cel_lint` rows (`cel.go:66,79`); all built-ins set `rule_id` == their own name (`broken_links.go:45,59`, etc.).
- Load-bearing for external typed gates: `command-inventory.md:100,150` (DESIGN §4.3 `gate: {command: "akb lint --kb <path> --json", output: "json", schema}`), `seam-periphery.md:50`. Consequence for this audit: **any change to the registered checker set changes the `by_check` key set**, which a strict external schema can pin. Checker enable/disable is therefore a *contract* change unless new keys are additive.
- Pinned by tests: `test/testdata/lint_tests.txt:107` (`"by_check"`), `test/testdata/lint_required_fields.txt` (`"required_fields": 1` **and** `"required_fields": 0`), `test/testdata/cel_lint.txt:47-53` (`rule_id`, `severity`, `warning`).

---

## Consumer trace

`RunLint` (`cmd/akb/lint.go:52-174`) is the only assembler; its inputs are all akb-owned stores:

| Input | Anchor | Used by |
|---|---|---|
| `.agent-kb/search.db` open + `RebuildLinks` (mutation!) | `lint.go:58-66`, `:87-89` | `broken_links`, `orphans`; hard failure without DB pinned by `test/testdata/lint_tests_nodb.txt` |
| `.agent-kb/templates/*.yaml` | `lint.go:76-79` | `required_fields`, `cel_lint`, `type_orphan` |
| `raw/files.log` manifest | `lint.go:81-84` | `citations` |
| `kb/**/*.md` walk (`.md` only, minus `index.md`/`log.md`) | `lint.go:101,105,109-143` | all page-scoped checkers |
| `kb/index.md` | `internal/lint/index_consistency.go:39` | `index_consistency` |

Downstream consumers of the report:

1. `runLint` → human text/JSON + exit 1 on any error (`lint.go:178-194,196-231,233-268`) → `classifyExit` (`cmd/akb/main.go`), exit-code contract `command-inventory.md:4`.
2. `template delete` → `RunLint` + the same exit policy (`cmd/akb/template_delete.go:132,141-152`, plus its own pre-count at `:95-100`).
3. External typed gates on `--json` (`command-inventory.md:100,150`).
4. Human/agent doctrine: `internal/skill/embedded/kb-management/references/MAINTAIN.md:14-27` reproduces the 10-checker table verbatim as the agent's interpretation key; `internal/lint/AGENTS.md:29-45` repeats thresholds and checker semantics.
5. Twist: `kb.Pages[].Annotations` (`engine.go:66`) has **no** consumer, and `is_draft` never filters the sweep (no `IsDraft` reference anywhere in `internal/lint`) — a hand-planted or git-pulled **draft** page is swept and can fail with `error` severity, even though `approve` is the only draft-aware gate (`cmd/akb/approve.go:218-231`).

---

## Blast-radius / verdicts

| Mandate | Verdict | Why / blast radius |
|---|---|---|
| A1/A2 fixed 10-checker list in the CLI | **HISTORICAL ACCIDENT** | The engine is already a registry with zero baked defaults. Config-driven registration is a CLI-layer change only; package API unchanged. Blast radius of *removing* a checker: `by_check` key disappears (D), MAINTAIN.md table (item 4), `test/testdata/lint_tests*.txt` stdout assertions. |
| A3 no disable/severity/threshold surface | **HISTORICAL ACCIDENT** | No structural obstacle in the package (`KB`/`LintIssue` are exported, `Severity` is a free string). Blocked only by the absence of a config field: `Config` has 3 keys (`config.go:16-21`) and `AGENTS.md` states thresholds are "not configurable in akb.yaml in v1". |
| A4 all-or-nothing engine | **GENUINELY OPINIONATED-BUT-LOAD-BEARING (weakly)** | A checker error currently aborts the sweep; `cel_lint` compile errors do too (`cel.go:63-65`), so a single malformed template rule kills the entire run. Making this per-checker degrade is a deliberate policy change (it would change which runs exit 1). |
| A6 error-severity ⇒ exit 1, no selector | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** | The CI/gate meaning of `akb lint` is pinned by integration tests and by the external gate contract. Changing default exit semantics is a breaking change; additive per-checker enable/severity override is not. |
| `broken_links` | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** | It is the only surface that reports graph truth, and the resolution semantics live in `linkgraph` (not lint), so disabling the checker removes reporting but not the mandate. Severity pairing (broken=error / ambiguous=warning) is pure opinion. |
| `orphans` | **HISTORICAL ACCIDENT** | Zero structural need: no part of write/approve/index/read depends on inbound degree. It is policy ("a KB must be a connected graph") plus a known staleness wart (`sqlite.go:246-254`). Highest false-positive cost (entry points, index-orphaned pages). |
| `empty_pages` | **HISTORICAL ACCIDENT** | Expressible in CEL today via `page.content.word_count`/`char_count` (`pagebuilder.go:71-72`); the shipped `note.yaml`/`adr.yaml` do not, so the checker is currently the sole owner. Retiring it as a built-in means each template must opt in. |
| `missing_frontmatter` | **STRUCTURALLY REQUIRED given the page model**, opinionated as a universal | Frontmatter is the only metadata carrier; `type`, required-presence, draft state and `old_page` all key on it. But the *universal* scope (every `.md` in `kb/`) and the `error` severity are opinions, and the message is wrong for "frontmatter present but unparseable" (`lint.go:120-121`). |
| `required_fields` | **STRUCTURALLY REQUIRED today, HISTORICAL tomorrow** | It is the only sweep-time presence gate, but it duplicates the write path (`cmd/akb/required_fields.go`) and approve (`approve.go:218-231`) — three implementations of one rule, and the lint copy is a hand-built mirror of the write helper (`required_fields.go:78-108`). JSON Schema `required` subsumes it (`seam-core-pipeline.md:258-259`); CEL `has()` could too. |
| `index_consistency` | **HISTORICAL ACCIDENT** | `index.md` is a second registry whose only source of truth is the filesystem, and it is mechanically regenerable (`internal/index/index.go:181`, `akb index rebuild`). The checker adds a hard dependency of the sweep on index.md's existence and grammar — a missing/malformed index line surfaces as "page missing from index" (`index.go:70-74` silently drops malformed entry lines). Strongest removal candidate. |
| `citations` | Field name: **HISTORICAL ACCIDENT**; semantics: **GENUINELY OPINIONATED-BUT-OPTIONAL** | The literal key `sources` is a prescribed vocabulary shared with one other site (`cmd/akb/raw_delete.go:124`) and with the shipped templates (`adr.yaml` declares `sources: type: list, required: false`). The check itself is cross-store and cannot be re-expressed in CEL because the manifest is not in the page map (`seam-periphery.md:65`). Blast radius of renaming/configuring the key: citations, raw_delete, shipped templates, docs. |
| `provenance` | **GENUINELY OPINIONATED-BUT-LOAD-BEARING (temporarily)** | Not re-expressible in CEL today: CEL exposes `page.ast.provenance_markers[{type,position}]` (`pagebuilder.go:82,362-363`) and `word_count`/`char_count` (`:71-72`), but **no line count** and **no CEL string extension** (`internal/cel/engine.go:36-47` registers only `page`,`old_page`,`now`; no `split`), so the "ratio over non-empty body lines" metric is unreachable. Worse, the CEL-exposed marker parser (`pagebuilder.go:345-363`) does **not** apply the code-block/inline-code/HTML-comment exclusions the lint parser does (`markdown/provenance.go:24-38`) — a CEL re-expression would silently disagree with the checker. Severity (warning) and threshold 0.20 are opinions. |
| `cel_lint` | **GENUINELY OPEN** | Rule, message and severity are template data. Caveats: severity is unvalidated (C), and the sweep has no `old_page` (`seam-core-pipeline.md:39,258-259`). |
| `type_orphan` | **STRUCTURALLY REQUIRED for akb's type system** | Retiring it is equivalent to retiring "every page has a type", i.e. the core mandate. Blast radius: write-time type enforcement, `required_fields` skipping, `cel_lint` skipping, `template delete` impact counting. |

**Cross-cutting structural finding.** The 10 checkers split cleanly into two classes, and only one is open:

- **Page-local** (5): `missing_frontmatter`, `empty_pages` (definition-free enough), `required_fields`, `type_orphan`, `cel_lint` — the last two already read template data; the first three are expressible as template `lint_rules` today (or as JSON Schema), so they *could* be demoted to shipped-but-unprivileged defaults.
- **Cross-store** (4): `broken_links`, `orphans` (linkgraph), `citations` (manifest), `index_consistency` (index.md) — the user-facing rule language sees **none** of these stores. `KB` (`engine.go:45-52`) is the only carrier, and it is not reachable from `akb.yaml` or a template.

So "users define validity" is bounded today not by the single-binary model but by **what facts a user rule can name**. Making the engine registry config-driven without also exposing graph/manifest/index facts to a rule language just moves the mandate around.

---

## Feasibility sketches

1. **akb.yaml enable/disable + severity override** (5 lines of schema, ~15 lines of wiring):
   `akb.yaml: lint: {checkers: {orphans: off}, severity: {cel_lint: warning}}` → `Config` gains one struct; `RunLint` filters `engine.AddChecker` calls by name and post-processes `report.Issues[].Severity`. Blast radius: `by_check` key set (keep keys, report 0 for disabled) and the docs table. Note severity override post-hoc cannot change `cel_lint`'s `rule.Severity` provenance — it can, since `LintIssue.Severity` is a free string.

2. **Threshold config**: only `ProvenanceDriftThreshold` (`thresholds.go:11`) is live; move it to `Config.lint.provenance_drift` and pass it into `NewProvenanceChecker(threshold)`. The other four constants are dead (`engine_test.go` only) and can be deleted outright. Not viable for `cel_lint` rules — those thresholds already live in template text (`adr.yaml:182-199`).

3. **Demote built-ins to "shipped but unprivileged"**: ship the 5 page-local checks as default `lint_rules` in the embedded templates (`adr.yaml`, `note.yaml`) or as a `schema.required` block, and register the checker only for types that opt in. Requires a per-template, not per-KB, enable surface — a template key the `LintChecker` interface cannot express today (`Check(ctx, kb)` is global).

4. **User-defined checkers under a single binary**: `go.mod` has no plugin dependency and the only `exec.Command` sites are `git` (`cmd/akb/init.go:199,234`, `cmd/akb/status.go:103`, `internal/storage/git.go:306,443,460,533`) — there is no dynamic-loading or subprocess-checker path. Plausible cheap variants: (a) a data-driven checker whose predicate is a CEL string over a **KB-level** map (the missing piece is the KB-level variable — today `cel.NewEnv()` exposes only a page map, `engine.go:36-47`); (b) an `exec`-based checker run per lint invocation. Both are additive to the envelope (`check` = the user's name) and keep `by_check` stable for built-ins.

---

## Open questions for synthesis

1. Is checker enable/disable allowed to change `by_check`'s key set, or must disabled checkers emit `0` to keep external gate schemas valid (`command-inventory.md:100,150`)?
2. Is exit-1-on-any-error a fixed CLI contract, or may a KB downgrade a built-in to `warning` and thereby make `akb lint` exit 0? (`AGENTS.md` exit-code contract; `test/testdata/lint_required_fields.txt` asserts `! exec akb lint`.)
3. Should the sweep become draft-aware (`is_draft: true` pages currently can fail with `error`; no exemption exists in `internal/lint`), or is that `approve`'s job alone?
4. Does the JSON-Schema direction absorb `required_fields` **and** keep `missing_frontmatter`/`type_orphan` as separate checkers, or collapse all three into one structural pass (`seam-periphery.md:70` already sketches a two-pass sweep)?
5. For cross-store mandates (`broken_links`, `orphans`, `citations`, `index_consistency`): is the intended openness move (a) opt-in per KB, (b) expose the linkgraph/manifest/index as CEL variables, or (c) delete the prescriptions — in particular, is `index.md` still meant to be a registry at all given `akb index rebuild` is authoritative?
6. `provenance` re-expression needs two page-map additions (line count, exclusion-aware marker list) — is that in scope for the JSON-Schema/CEL seam (`seam-core-pipeline.md:318-320` lists the 4 call sites that build the page map), or does `provenance` stay a privileged built-in?

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Report written to the authoritative output path with file:line-anchored mandate tables, consumer trace, verdicts, feasibility sketches and open questions."
    }
  ],
  "changedFiles": [],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "grep -rn 'AddChecker|NewLintEngine' --include=*.go .",
      "result": "passed",
      "summary": "Confirmed the only non-test registration site is cmd/akb/lint.go:157-167"
    },
    {
      "command": "grep -rn 'ProvenanceDriftThreshold|FreshnessHalfLifeDays|FreshnessScoreThreshold|SummaryMinLength|SummaryMaxLength' --include=*.go .",
      "result": "passed",
      "summary": "4 of 5 lint thresholds are dead (referenced only by engine_test.go)"
    },
    {
      "command": "grep -rni 'severity' --include=*.go . | grep -v internal/lint",
      "result": "passed",
      "summary": "No validation/defaulting of LintRule.Severity anywhere; consumers only compare == \"error\""
    },
    {
      "command": "grep -rn 'exec.Command' --include=*.go cmd/ internal/",
      "result": "passed",
      "summary": "Only git subprocesses; no plugin/subprocess checker path exists"
    }
  ],
  "validationOutput": [
    "AddChecker non-test callers: cmd/akb/lint.go:158-167 only (engine.go defines it at :87-89)",
    "RunLint consumers: cmd/akb/lint.go:180 and cmd/akb/template_delete.go:132 (plus its own duplicate error counting at :141-152)",
    "No lint key in internal/config/config.go (Config = name/created/description); lint has exactly one flag, --json (lint.go:45)"
  ],
  "residualRisks": [
    "Line numbers for a few interior lines of cmp/akb/lint.go's WalkDir callback were re-derived from grep anchors after the read tool's offset disagreed with grep; the cited anchors (101,105,109,120-121,129-130,140) are grep-verified, but a spot check is advised before editing.",
    "Prior research files (.pi/subagents/proposals/json-cel/research/*) were skimmed via grep, not read end-to-end; a claim established only in prose there may be cited imprecisely.",
    "The claim that page-local checks are 'expressible in CEL today' was verified against the page-map keys exposed by pagebuilder.go, but no candidate rule was actually compiled/evaluated.",
    "DESIGN.md (agent-memory) was not on disk at the path prior research cites, so the typed-gate dependency on akb lint --json is cited secondhand from command-inventory.md:100,150."
  ],
  "noStagedFiles": true,
  "diffSummary": "No repo files modified; read-only recon. One artifact written to the configured output path.",
  "reviewFindings": [
    "no blockers",
    "note: duplicate exit-policy logic exists in three sites (cmd/akb/lint.go:221-229, :259-267, cmd/akb/template_delete.go:141-152)",
    "note: LintRule.Severity is unvalidated (internal/template/template.go:42 comment only; templates_write.go:118-122 compiles CEL only), so an unknown/empty severity reports but never fails the run",
    "note: PageData.Annotations (internal/lint/engine.go:66, cmd/akb/lint.go:130) has no consumer, and the lint sweep has no is_draft exemption"
  ],
  "manualNotes": "Headline verdicts: the engine is already an open registry with zero baked-in defaults — the 10-checker mandate exists only at cmd/akb/lint.go:157-167, so config-driven enable/disable is a CLI+config change, not a package redesign. The real openness bound is not the single-binary model but the fact that a user rule (CEL lint_rules) can name only page-local facts, while 4 of 10 checkers depend on cross-store state (linkgraph, raw manifest, index.md) that is exposed nowhere. index_consistency and orphans are the weakest mandates (HISTORICAL ACCIDENT); broken_links/type_orphan/missing_frontmatter are load-bearing; provenance is opinionated and not faithfully re-expressible in CEL today because the page map lacks line counts and uses an exclusion-unaware marker parser (pagebuilder.go:345-363 vs markdown/provenance.go:24-38). The --json by_check key set is the contract to watch for any enable/disable change."
}
```
