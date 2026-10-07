---
as_of: "2026-09-26"
created: "2026-09-26"
grammar: 1
informs:
- RFC-001
- RFC-002
is_draft: false
provenance: agent-drafted
run: json-cel
scope:
- internal/**
- cmd/akb/**
status: final
summary: "Read-only survey of subsystems outside the core pipeline (search, link graph, lint, raw, index, log, git): where a JSON-Schema/CEL seam could attach and where it cannot."
tags:
- json-schema
- cel
- subsystems
- recon
title: Seam periphery — JSON Schema / CEL coexistence across akb subsystems
type: research
updated: "2026-09-26"
---
# Seam Periphery — JSON Schema / CEL coexistence across akb subsystems

Scope: everything outside the core write/read/validate/template/CEL pipeline. Read-only recon.
Repo: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb`. All paths relative to repo root.

Baseline facts that frame every verdict:

- No JSON Schema exists anywhere in the tree today (`grep -rn "jsonschema|JSON Schema"` → zero hits outside `.pi/`).
- The only validation surfaces are: (a) required-field **presence** (`cmd/akb/required_fields.go:19-29`, mirrored in lint `internal/lint/required_fields.go:81-91` and approve `cmd/akb/approve.go:218-231`), and (b) CEL (`internal/cel/`, template `validations` + `lint_rules`).
- Template schema is a *flat* map of `{type: string|list, required: bool, enum: [..]}` — `internal/template/template.go:19-28`. Type/enum are **declared but never enforced structurally**; enforcement lives in CEL rules (see `internal/template/embedded/adr.yaml:68-72` re-stating the enum as a CEL `in [...]` rule). That duplication is the central seam signal.
- Existing JSON in akb is **output-only** and ad hoc: 6 `--json` flags, each with a hand-rolled struct. No JSON input path anywhere except `raw write` (bytes verbatim) and nothing reads JSON.

---

## 1. SEARCH — `internal/search/`, `cmd/akb/search.go`

**Purpose.** FTS5/BM25 full-text index over page bodies, queryable with tag/type/date filters.

**Data in/out.**
- Write side: `SQLiteFTS5Searcher.IndexPageTx` (`internal/search/sqlite.go:56-78`) inserts `documents(path,title,content,tags,summary,type)`, `pages_fts(title,content,tags,summary)`, `pages(path,title,summary)`. `content` is the **raw page body verbatim**, provenance markers and annotations included. Only `tags` and `summary` are lifted from frontmatter (`ExtractTags`/`ExtractSummary`, `sqlite.go:300-332`).
- Read side: `SearchResult{Path,Title,Summary,Snippet,Rank}` (`internal/search/searcher.go:11-21`); `SearchOptions{Limit,Tag,Type,After}` (`searcher.go:23-30`, default limit 10 at `sqlite.go:176-179`; filters are SQL `LIKE`/`=`/`>=` at `sqlite.go:189-201`).
- Query escaping: `escapeFTS5Query` (`sqlite.go:155-185`) strips `OR|AND|NOT`, blanks FTS5 punctuation, quotes each token.
- Rebuild: `RebuildIndex` (`sqlite.go:220-300`) walks `kb/`, skips `index.md`/`log.md`, **silently skips pages whose frontmatter fails to parse** (`sqlite.go:275-278`) — so search tolerates exactly the pages JSON Schema would flag.

**Current JSON.** `--json` at `cmd/akb/search.go:38`; shape `searchJSONResult{path,title,summary,snippet,rank}`, always an array (`search.go:101-124`).

**Does search index content validation cares about?** It indexes the body, which CEL reads as `page.content.word_count` / `page.ast.*`. Search never validates and never will; it is a projection.

**Verdict.** JSON Schema role: **none** (a query/result view, not an input). CEL role: **none** today; a `--filter <CEL>` over results would be additive, not a seam. Search *consumes* the same validated page the schema describes, but the dependency is one-way and non-blocking.
Risk to note: the "silently skip unparsable page" policy means a stricter schema gate at write time does not propagate to search — rebuild can still index/omit pages that never passed validation.

---

## 2. LINK GRAPH — `internal/linkgraph/`, `cmd/akb/links.go`

**Purpose.** Cached wikilink resolution over the page set.

**Data in/out.** `Link{SourcePage,RawTarget,Display,ResolvedTo}` (`internal/linkgraph/sqlite.go:16-21`). Resolution is 3-step (`resolveTarget`, `sqlite.go:300-345`): exact `raw+".md"` → namespace `"kb/"+raw+".md"` → basename match (0 → nil = **broken**; 1 → path; 2+ → `"AMBIGUOUS:p1,p2"` sentinel, sorted). Broken = `resolved_to IS NULL` (`:258-264`); ambiguous = `resolved_to LIKE 'AMBIGUOUS%'` (`:266-272`). Rows are (re)derived from `markdown.ParseWikilinks` (`sqlite.go:71`), i.e. a **projection of page bodies**, rebuilt by `RebuildLinks` (`sqlite.go:164-227`) and by the write path.

**Current JSON.** `backlinks --json` only (`cmd/akb/links.go:56`); `backlinkJSON`/`backlinksResponse` at `links.go:150-162` (always an object wrapping a `backlinks` array). **`akb links` and `akb orphans` have no `--json`** (`links.go:21-47`).

**Do lint/CEL consume it?** Lint yes: `broken_links.go:38,52` and `orphans.go:27` query the graph. CEL **no**: `page.ast.links` is built independently by goldmark inside `internal/cel/pagebuilder.go` (used at `internal/lint/cel.go:61-62`), so CEL rules see the page's own link tokens, not resolved graph state. Staleness exists: `GetOrphans` doc comment (`sqlite.go:190-198`) notes `resolved_to` goes stale until re-resolution.

**Verdict.** This is **graph state**, not page structure — it is derived, cached, and re-computable, so it lives outside any page JSON view. JSON Schema role: **none**. CEL role: **indirect** — link-shaped *authoring* constraints (e.g. "must contain a wikilink", "superseded ADR must link its predecessor", `adr.yaml:167-178`) are CEL over `page.ast.links`; resolution correctness belongs to the graph, not to either validator. A JSON page view may carry resolved links as read-only derived data (JSON I/O output seam), never as schema input.

---

## 3. LINT ENGINE — `internal/lint/`

**Contract already JSON-shaped.** `LintIssue{Type:check,RuleID,Message,Path,Severity}` and `LintReport{Issues,PagesChecked,ByCheck}` carry JSON tags (`engine.go:17-42`); `akb lint --json` envelope at `cmd/akb/lint.go:233-268`. Engine is a checker registry (`engine.go:37-42`, `Run` `:87-110`); 10 checkers registered at `cmd/akb/lint.go:157-167`.

**Per-checker data class and seam.**

| Checker | File:line | What it inspects | Class | Seam |
|---|---|---|---|---|
| `required_fields` | `required_fields.go:44-91` | presence of template-`required` frontmatter keys | **page structure** | **JSON Schema** — `required` is exactly this; the write-path twin (`cmd/akb/required_fields.go:19-29`) and this checker collapse into one schema artifact |
| `type_orphan` | `type_orphan.go:34-51` | `frontmatter.type` → template registry | page structure (registry) | **JSON Schema** — `type` as `enum`/`const` over known template names |
| `missing_frontmatter` | `missing_frontmatter.go:33-45` | page lacks `---` delimiters | page structure (envelope) | JSON Schema describes the page *view* envelope; it does not operate on the on-disk markdown, so **partial** — the "is there a JSON object at all" question is a JSON I/O concern |
| `empty_pages` | `empty_pages.go:34-49` | body whitespace-empty | content-derived | **CEL/neither** — `minLength` on a body field is technically schema-able but CEL already owns `page.content.word_count` (`adr.yaml:117-120`) |
| `broken_links` | `broken_links.go:35-66` | linkgraph NULL/AMBIGUOUS | **graph/state** | neither (see §2) |
| `orphans` | `orphans.go:26-43` | linkgraph inbound-degree | **graph/state** | neither |
| `index_consistency` | `index_consistency.go:36-78` | `kb/index.md` entries vs filesystem | **state (managed file)** | neither |
| `citations` | `citations.go:26-68` | frontmatter `sources` vs raw manifest | cross-store state | **neither today** — the manifest is not in the page map; CEL *could* own it if the raw manifest were injected as a variable |
| `provenance` | `provenance.go:30-74` | declared confidence vs inline marker ratio, drift > 0.20 (`thresholds.go:11`) | content-derived arithmetic | **CEL** (arithmetic over content); JSON Schema cannot compare a derived ratio |
| `cel_lint` | `cel.go:41-100` | template `lint_rules` | **temporal/contextual** | **CEL only** |

**`cel_lint` vs write-time validations — the coexistence boundary.**
- Same expression language, different surfaces: write-time evaluates `validations` with `page` + `old_page` + `now`, failing **closed** (exit 1, write blocked); the sweep evaluates `lint_rules` with `page` + `now` **only** — no `old_page` (`cel.go:70-73`), and per-page eval errors degrade to an issue and the sweep continues (`cel.go:74-86`).
- `now` is injected once per sweep at `cel.go:48`; rules like `adr.yaml:182-199` (`now - timestamp(updated) < duration("4320h")`) are pure temporal policy.
- Can JSON Schema participate in sweeps? **Only as a structural prefilter.** It can answer presence/type/enum/shape per page, with no `now`, no `old_page`, and no content metrics. Temporal drift, state transitions, and body/AST predicates stay CEL. That means a sweep under coexistence is two passes over the same page list: `schema` (new checker, structural, pure function of the page) then `cel_lint` (policy, time-dependent) — both emitted as `LintIssue` rows so the existing `--json` envelope and `by_check` map absorb them without shape change.
- Note `required_fields` becomes redundant with a schema checker; the current comment (`required_fields.go:17-20`) explicitly frames it as the sweep-time rescue for pages that never passed `akb write` — the same argument applies verbatim to the schema checker.

---

## 4. INDEX + LOG — `internal/index/`, `internal/log/`, `cmd/akb/index*.go`, `log*.go`

**Index.** `IndexEntry{Path,Title,Summary,Type}` (`internal/index/index.go:20-25`); `ReadIndex` `:46-53`, `AddEntry` `:113-146`, `RebuildIndex` `:181-247`, `RenderIndex` `:250-288` (grouped under `Type+"s"` headings, entry = `- [Title](Path) — Summary`). Parsing is tolerant: malformed entry lines are skipped, not errors (`:68-73`).

**Log.** `LogEntry{Date,Operation,Title,Description}` (`internal/log/log.go:20-25`); heading grammar `^## (\d{4}-\d{2}-\d{2}) (\w+)(?: \| (.+))?$` (`log.go:27`); append stamps `time.Now()` (`log.go:93`). Operation is a **closed enum** hardcoded in the CLI (`cmd/akb/log.go:21`: ingest/delete/update/lint/query/distill/approve/plan).

**JSON / validation today.** No `--json` on any index or log command (`cmd/akb/index.go`, `cmd/akb/log.go`). Both files are **exempt from validation and from being walked as pages**: skipped by lint (`cmd/akb/lint.go:104-107`), search rebuild (`internal/search/sqlite.go:238`), index rebuild (`internal/index/index.go:199-202`), approve (`cmd/akb/approve.go:307-309`). The docs call them managed files; the CLI is the only writer.

**Verdict.** JSON Schema role: **neither for the files** — they are machine-owned markdown with their own tolerant parser, and a schema would only describe a format nothing external authors. Secondary seam: both are **derived from validated page frontmatter** (`Path/Title/Summary/Type`), so they are downstream of the page schema, not peers. CEL role: **none**. The log's operation enum is a trivial schema candidate with negligible value. Output-only `--json` additions for `index show` / `log show` are plausible JSON I/O work but carry no input schema.

---

## 5. RAW + MANIFEST — `internal/manifest/`, `cmd/akb/raw*.go`

**Purpose.** Immutable raw-file store + SHA-256 drift ledger.

**Data.** `Entry{Filename,SHA256,LastUpdated}` (`internal/manifest/manifest.go:20-25`), pipe-delimited line format in `raw/files.log` (`:65-79`), atomic temp+rename write (`:86-123`). `ValidateFilename` (`:226-236`) rejects `|`, newlines, and the literal `files.log`. `ComputeSHA256` streams (`:207-222`).

**Drift.** `cmd/akb/raw_status.go`: `driftFile{path,status}` / `driftSummary{modified,untracked,missing}` (`:21-35`), `--json` at `:54`, exit-code contract via `driftDetected` (`:190-192`). `raw_sync.go` reconciles and commits.

**JSON/validation intersection.** akb never parses raw content — `akb raw write data/config.json` stores bytes verbatim (`cmd/akb/raw.go:19` is documentation, not parsing). The only bridge to validation is the frontmatter `sources:` list, consumed by `citations` lint (`internal/lint/citations.go:37-66`). `raw list` has no `--json` (`cmd/akb/raw_list.go:28-75`).

**Verdict.** JSON Schema role: **none** for storage; the sole openings are (a) `sources` items if a template schema ever encodes them (pattern/format for filenames), and (b) describing the manifest format itself — low value, since it is auto-generated with an explicit corruption guard (`manifest.go:66-68`). CEL role: **none**. Validation here is hashing, not structure.

---

## 6. STORAGE / GIT — `internal/storage/`

**Path.** `Provider` is content-agnostic bytes-in/bytes-out (`provider.go:10-16`). `WriteWithCommitMsg` (`git.go:56-85`) = mkdir → `os.WriteFile` → `git add` → `git commit`; `Delete` (`:108-134`). Commit messages are **path-derived**: `akb: write <path>` (`:92`), `akb: delete <path>` (`:131`), `akb: approve <path>` (`cmd/akb/approve.go:172`).

**Guards.** `checkMergeConflicts` (`git.go:178-193`) runs **before** any write and thus before validation; `withRepoLock`/`LockRepo` (`:195-...`, `:244-...`) serializes whole mutation sequences; identity fallback `akbCommitIdentity` (`:32`) is used per-invocation, and only `akb init` writes repo config (`cmd/akb/init.go:232-252`). `managedCommitPaths = {kb/index.md, kb/log.md}` (`git.go:27`) join a commit's pathspec. `CommitFiles`/`StageFiles`/`NothingToCommit` (`:327-...`) scope commits by pathspec.

**Does JSON I/O change the commit path?** **No.** The provider commits bytes and derives messages from the path; the validation order (conflict check → write → stage → commit) is untouched. The one real risk is a **page file format** question, not an I/O-flag question: every walker filters `.md` (`internal/search/sqlite.go:238`, `internal/index/index.go:195`, `cmd/akb/lint.go:101`, `cmd/akb/approve.go:299`). If "JSON I/O" ever means storing pages *as* `.json`, those four walkers plus four rebuild paths change — that is a much larger change than adding `--json` stdin/stdout.

**Verdict.** JSON Schema role: **none**. CEL role: **none**.

---

## 7. APPROVE / DRAFTS — `cmd/akb/approve.go`

**Purpose.** Draft → published lifecycle: strip `olw-auto` annotations and provenance markers, set `is_draft: false`, rewrite frontmatter wholesale, commit, update search + linkgraph in one transaction.

**Where approval intersects validation.** Exactly one gate: `requiredFieldsRefusal` (`approve.go:218-231`) → `checkRequiredFields` (`cmd/akb/required_fields.go:19-29`) — **presence only**, and an unknown `type` passes by design (`:216-217`, since that belongs to `type_orphan`). CEL `validations` are **not** re-run at approve. Batch mode is two-phase: validate all candidates first, report the first non-per-page failure before approving anything (`approveAllDraftPages` `:252-280`); a required-field refusal keeps that draft in draft state without aborting the rest, and the run still exits 1 (`:258-279`). Draft detection: `frontmatter.IsDraft` at `:144` and `:324`. Frontmatter is re-serialized from `allFields` via `yaml.Marshal` (`:157-170`) — a JSON-Schema description of frontmatter would also document exactly what this re-serialization must preserve.

**Should approval re-validate, and against what?**
- Coexistence recommendation: approve = **full schema pass (structure: presence + type + enum) + the write-time CEL `validations` re-evaluated with `old_page` = the on-disk page and `now` = approval time**. Approval is the last gate before a page is "published"; today it lets a CEL-invalid page through (write-time rules are never re-run). `old_page` is available and meaningful here, unlike a lint sweep.
- Sweep-only `lint_rules` should stay out of approve (they are temporal policy, not a publish gate).
- Cheap win: with native JSON Schema the schema half of this gate is a single reusable call and `requiredFieldsRefusal` becomes a wrapper; unknown-type handling stays as-is.

**Verdict.** JSON Schema role: **yes** (structural gate + documented frontmatter shape). CEL role: **yes** (write-time `validations` re-run with `old_page`/`now`).

---

## 8. DISCOVER / INIT / CONFIG — `cmd/akb/discover.go`, `init.go`, `internal/config/`

**Config.** `Config{Name,Created,Description}` (`internal/config/config.go:18-22`); `Load` (`:25-46`) does a merge-conflict string guard (`:32-34`), YAML unmarshal, and a single required check — `name != ""` (`:41-43`); `Save` (`:49-57`). `akb.yaml` is the marker that defines a KB.

**Discover.** `akb discover --json` marshals `[]path.DiscoveredKB` (`cmd/akb/discover.go:54-61`; struct at `internal/path/path.go:420-424`) — note this is the one `--json` that uses `json.Marshal` with no indent, unlike the encoder-based commands.

**Init.** `cmd/akb/init.go:45-92`; seeds `kb/index.md`, `kb/log.md`, `raw/files.log` (`:135-147`); writes config (`:149-159`); **seeds no templates** (`:118-133`).

**Would config ever be JSON-Schema-described?** Plausibly, and the value grows over time: `Load` today enforces only `name`, while frontmatter lint thresholds are **hardcoded constants** (`internal/lint/thresholds.go:10-16`, and AGENTS.md states thresholds are not configurable in v1). If any policy knob (drift threshold, freshness half-life, default types) moves into `akb.yaml`, a config schema stops being ornamental. Today it is low value and **not a blocker**.

**Verdict.** JSON Schema role: **optional/marginal** for `akb.yaml` (small, already one-field-enforced); **output-contract only** for `discover --json` / `list --json` shapes. CEL role: **none**.

---

## 9. SKILL — `internal/skill/` (embedded kb-management doctrine)

Embedded via `//go:embed embedded/*` (`internal/skill/skill.go`). Files: `SKILL.md` (144 lines, router only), `references/{TEMPLATE,MAINTAIN,APPROVE,INGEST,UPDATE,QUERY,INIT}.md`. This is the doctrine a coexistence design must update — it is the agent-facing spec of the template/validation contract.

**Specific guidance that would change:**

1. `references/TEMPLATE.md:34-58` — "Anatomy of a Template" presents `schema.frontmatter` (type/required/enum) + `validations` + `lint_rules` as the whole contract. Must become a two-layer anatomy (structure vs context/arithmetic).
2. `references/TEMPLATE.md:60` — **"Presence is all the schema enforces there — a value's type and enum constraints stay with the validation rules."** This is the sentence the design directly contradicts; native JSON Schema would enforce type and enum structurally. Highest-priority rewrite.
3. `references/TEMPLATE.md:67-83` — "CEL Concepts" (`page`/`old_page`/`now`; `page.frontmatter.*`, `page.ast.headings/links/code_blocks`, `page.content.word_count/char_count`, date auto-coercion). Needs an explicit boundary statement of which facts the schema owns vs CEL.
4. `references/TEMPLATE.md:102-119` — **"Guarding frontmatter reads with has()"**: required reads unguarded, optional always guarded, `old_page` always guarded, lint rules guard everything but `type`/`title`; plus the deliberate three-surface divergence (write fails closed / lint degrades / template write rejects unguarded optional reads). Under a schema layer, required-key presence is structural, so the concise `has()` rule narrows to *optional keys + `old_page`*; the divergence table stays CEL-specific but its justification for required keys changes.
5. `references/TEMPLATE.md:119` — known limit: the optional-stripping proof enumerates **schema-declared optional keys only**, so an unguarded read of an undeclared key is caught only at first use. A schema layer changes what "declared" means; this paragraph must be re-derived.
6. `references/TEMPLATE.md:117` and `cmd/akb/templates_write.go:167-187` — `template write` proves the pass mockup three ways (as given, per-optional-key-stripped, self-as-`old_page`), with the error text telling authors to `has()`-guard or mark required. If the schema splits out, `template write` must additionally prove the mockup **against the JSON Schema**, and the error text gains a structural branch.
7. `references/TEMPLATE.md:12` — "a template without mockups is unproven"; `:153` "both `--pass` and `--fail` are required". Mockup contract expands if the schema is a second artifact.
8. `references/TEMPLATE.md:123-133, 289` — showcase workflow (`akb template get <n> --full --examples`) and "start from a showcase"; `--full` output shape changes if it emits JSON Schema fragments, and INIT.md:24,49 tells users to copy that output verbatim.
9. `references/TEMPLATE.md:269-292` — DO/DON'T list, incl. "DO NOT use the old template format" (`required:`/`optional:`/`body:` rejected; `internal/template/template.go:73-81`) and "DO use `enum` for fields with a closed set". The old-format rejection is the precedent for a hard, loudly-errored schema migration.
10. `references/MAINTAIN.md:14-27` — the 10-checker table the agent reads to interpret lint output. A new structural checker row must be added and `cel_lint`'s row re-scoped (currently the only "schema-ish" row). `:36,42` — freshness is explicitly "template-owned CEL `lint_rules`" → stays CEL.
11. `references/APPROVE.md:32` — "Approves every draft that passes schema required-field validation" → reword to full schema (+ whatever CEL policy §7 adopts).
12. `SKILL.md:103,117-124` and `references/{INGEST,UPDATE,QUERY}.md` — command tables and "Always check the template before writing" flow; JSON I/O flags (`--json` on more commands, stdin/stdout JSON on `read`/`write`/`template get`) must be added to `SKILL.md:113-144` and QUERY.md:45.
13. `internal/template/embedded/adr.yaml:1-8` — the **guard-doctrine header comment** shipped as the exemplar, plus its `schema:` block (`:12-48`), `validations:` (`:49-178`), `lint_rules:` (`:179-199`). This is the first artifact to restructure, and `note.yaml` the minimal counterpoint.

---

## BOUNDARY TABLE

| Subsystem | JSON Schema role | CEL role | Justification |
|---|---|---|---|
| SEARCH | **neither** | **neither** (optional `--filter`) | Derived BM25 projection; indexes raw body, only a read view; no validated input |
| LINK GRAPH | **neither** | **indirect** (via `page.ast.links`) | Re-computable graph state; CEL sees the page's own link tokens, not resolution |
| LINT → `required_fields` | **yes** | neither | Exactly JSON Schema `required`; duplicate of the write-path presence check |
| LINT → `type_orphan` | **yes** | neither | `type` as enum of known template names |
| LINT → `missing_frontmatter` | **partial** | neither | Envelope existence is a JSON I/O concern, not an on-disk markdown rule |
| LINT → `empty_pages` | neither | **yes** | Body emptiness is content; CEL already owns `word_count` |
| LINT → `broken_links` / `orphans` | neither | neither | Graph/state from linkgraph; not in the page JSON view |
| LINT → `index_consistency` | neither | neither | Managed-file state (`index.md` vs filesystem) |
| LINT → `citations` | neither (optional `sources.items`) | **marginal** | Cross-store (frontmatter vs raw manifest); manifest not in the page map |
| LINT → `provenance` | neither | **yes** | Threshold arithmetic over content-derived marker ratio |
| LINT → `cel_lint` | **no sweep participation** | **yes** | Temporal/contextual; needs `now`, no schema equivalent |
| INDEX + LOG | **neither** (output-only `--json`) | neither | Machine-owned, parser-driven, exempt from page validation; derived from page frontmatter |
| RAW + MANIFEST | **neither** | neither | Byte store + SHA-256 ledger; content never parsed |
| STORAGE / GIT | **neither** | neither | Provider is content-agnostic; commits bytes and path-derived messages |
| APPROVE / DRAFTS | **yes** | **yes** | Structural gate today is presence-only; publish gate should re-run schema + write-time CEL with `old_page`/`now` |
| DISCOVER / LIST / CONFIG | **marginal** (config schema; output contract) | neither | Config enforces only `name` today; value rises if lint thresholds become configurable |
| SKILL (kb-management) | **must be updated** | **must be updated** | Doctrine at TEMPLATE.md:34-119 is the agent-facing spec of exactly this boundary |

**One-line synthesis.** JSON Schema owns *page structure* (presence, type, enum, envelope) and plausibly the *config/manifest envelope*; CEL keeps *content, AST, arithmetic, and temporal* policy; every graph/state/file-management subsystem (search, linkgraph, index, log, raw, git) is **neither**, and its correct relationship to a schema layer is *downstream consumer or exempt walker*, not validator.

## Residual risks / open questions

1. **`.md`-filtered walkers.** `search` (`sqlite.go:238`), `index` (`index.go:195`), `lint` (`lint.go:101`), `approve` (`approve.go:299`) all key on `.md`. Any JSON *page file* format silently breaks all four; JSON I/O flags must not be conflated with a page-file format change.
2. **Enum duplication.** `schema.frontmatter.<f>.enum` is currently mirrored by hand into a CEL `in [...]` rule (`adr.yaml:29-32` vs `:68-72`). Moving enum enforcement to JSON Schema creates a second source of truth unless the CEL rule is removed or generated — decide explicitly.
3. **Divergent failure policy.** Write fails closed on CEL eval errors; lint degrades per page (`internal/lint/cel.go:74-86`). A new schema checker needs a stated policy for the equivalent case (malformed page vs malformed schema).
4. **Sweep ordering.** Lint rebuilds the link graph before checkers run (`cmd/akb/lint.go:86-89`); a structural pass should run before CEL so `cel_lint` never evaluates a page whose required keys are already known missing (currently `required_fields` and `cel_lint` are independent checkers with no ordering guarantee).
5. **Approve's blast radius.** Approve re-serializes all frontmatter (`approve.go:157-170`); if it gains schema/type coercion, it could silently normalize or drop fields a schema does not describe. Needs an explicit unknown-field policy.
6. **`now` provenance.** `cel_lint` takes `time.Now()` once per sweep (`internal/lint/cel.go:48`); any schema pass that also needs time must share that stamp or the two passes will disagree at day boundaries.
7. **No JSON Schema library in `go.mod`** — adding one is a dependency decision for the parent design, not a peripheral one.
