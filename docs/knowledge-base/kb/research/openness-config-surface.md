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
- internal/config/**
- internal/path/**
- internal/cel/**
- internal/search/**
- internal/storage/**
status: final
summary: "Audit finding: akb.yaml holds three fields, two read and zero behavioral; every knob worth exposing (compute budget, drift threshold, indexed fields, layout) is a Go constant."
tags:
- openness
- configuration
- akb-yaml
- audit
title: Openness audit — configurability
type: research
updated: "2026-09-27"
---
# Openness Audit — Seam: CONFIGURABILITY

Scope: `internal/config/`, `internal/path/`, `internal/cel/engine.go`, `internal/search/sqlite.go`,
`internal/linkgraph/`, `internal/storage/git.go`, `internal/index/`, `internal/log/`,
`internal/manifest/`, `internal/template/`, `cmd/akb/` (flags only).

Read-only recon at cwd `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb` (working-tree
HEAD, v0.21.0 line per AGENTS.md). No repo file modified.

**Headline:** `akb.yaml` has **three** fields; **two** are read by anything, **zero** are behavioral.
Every behavioral knob the owner may want to expose (compute budget, drift threshold, indexed
fields, commit format, managed-file set, layout) is a Go `const`/`var`, not a config key. The
container for an openness policy exists and is essentially empty — that is the finding, not a
missing feature. The one genuinely load-bearing opinion is the **write-path stamping + type gate**
(`created`/`updated`/`is_draft`/`type`/`title`), which is not a constant at all but a hardcoded
*document mutation*. Prior research already establishes the validation seam itself
(`json-cel/research/seam-core-pipeline.md` §1–2, `akb-recon.md` §2–3); this report does not re-derive it.

---

## Findings — mandate tables

### Table 1 — `akb.yaml` TODAY (the whole surface)

`internal/config/config.go:14-18`:

```go
type Config struct {
	Name        string `yaml:"name"`
	Created     string `yaml:"created"`
	Description string `yaml:"description,omitempty"`
}
```

| Field | Default | Required | Real consumers | Verdict |
|---|---|---|---|---|
| `name` | none | **yes** — load fails if empty (`config.go:40-42`) | `cmd/akb/root.go:105-109` (identity prefix line `kb: <name> (<path>)`); `cmd/akb/status.go:36,56`; `internal/path/path.go:594-606` → discover listing (`path.go:620-622`) and `discover --json` (`path.go:423`); fallback to dir basename at `path.go:604-606` | **STRUCTURALLY REQUIRED** as a *marker*, not as data. `hasKBConfig` (`path.go:340-347`) only requires a regular file at that path; `name` is informational. |
| `created` | none (written by `cmd/akb/init.go:150`) | no | **NO READER AT ALL** — `grep -rn "\.Created"` → only the write site `init.go:150` | **HISTORICAL ACCIDENT** — write-only field. |
| `description` | `""` | no | `internal/path/path.go:600` → discover (`path.go:620-622`, `discover.go`); `--description` flag `cmd/akb/init.go:42` | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** (discover needs a label); harmless, cannot break anything. |

Two structural facts about the file itself:

1. **The file is a marker first, a config second.** `path.ResolveKB` refuses any directory without
   a *regular file* at `<kbRoot>/.agent-kb/akb.yaml` (`path.go:333-347`). Five commands re-load it
   purely as a validity gate and discard the result: `cmd/akb/write.go:109`, `append.go:98`,
   `read.go:42`, `status.go:36`, `root.go:105`. So `config.Load`'s only *hard* function today is
   "the marker exists and parses YAML with a `name`".
2. **No unknown-key rejection.** `yaml.Unmarshal` into a 3-field struct (`config.go:34-37`) silently
   drops everything else. A user can put `compute_budget: 500000` in `akb.yaml` today and get
   **no error and no effect**. Any new config field is therefore additive with zero migration risk
   (there is literally no way to detect a stale config), but also means today's "config surface"
   cannot express intent and akb cannot warn about it.

### Table 2 — Hardcoded constants inventory (core deliverable)

Legend for verdict: **TRIVIAL** = one config field, mechanical; **NEEDS DESIGN** = configurable but
the consumer contract must be defined first; **STRUCTURAL** = must exist as a bound/identity for
the tool to function; **ACCIDENT** = vestigial or a value nobody chose deliberately.

**A. CEL engine (`internal/cel/engine.go`)**

| # | Constant / hardcode | Anchor | Consumer trace | Config verdict | Blast radius if changed |
|---|---|---|---|---|---|
| A1 | `MaxCostLimit = 100000` | `engine.go:22`, applied `engine.go:57` (`cel.CostLimit`) | every `CompileRule`: `write.go:598`, `append.go` (via write), `templates_write.go:114,119,355`, `template.go:235`, `lint/cel.go:65`. Cancel → `ErrComputeBudget` (`engine.go:25`, mapped `engine.go:88-90,97-99`) | **TRIVIAL** config field (`validation.cost_limit`), but the *bound itself* is **STRUCTURAL** (an unbounded user-supplied CEL expression is a DoS on the write path). The **value** is an unchosen round number → **ACCIDENT**. | Only knob re-tuning; no schema change. Note: on write, cost-cancel fails **closed** (blocked write) — raising the limit loosens the write gate; on lint it degrades per-page (`lint/cel.go:74-86`). Divergence is deliberate (AGENTS.md "CEL rule evaluation"). |
| A2 | Program cache keyed by **expression string only** | `engine.go:30,44,56` | all `CompileRule` callers | **NEEDS DESIGN** — a per-template/per-KB env extension (see §Feasibility) makes the cache key wrong: same expression, different env = stale program | Silent wrong-verdict. Must key on (env identity, expr). |
| A3 | Env is fixed to exactly 3 variables, no functions, no extensions | `engine.go:35-41` (`page`, `old_page`, `now`) | `NewEnv()` at `write.go:131`, `append.go:209`, `templates_write.go:108`, `template.go:207`, `lint/cel.go:42` | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** — this is the entire language surface; §3 below | — |
| A4 | `now` injected as `time.Now()` per call | `write.go:609`, `templates_write.go:368`, `template.go:249`, `lint/cel.go:48,72` | temporal rules (e.g. `adr.yaml` `adr_stale`) | **TRIVIAL** (`--now` for reproducibility/tests) | Test determinism only; no behavior change. |

**B. Search / FTS5 (`internal/search/sqlite.go`, `internal/db/db.go`)**

| # | Hardcode | Anchor | Consumer trace | Verdict |
|---|---|---|---|---|
| B1 | Indexed columns are fixed: `pages_fts(title, content, tags, summary)`; `documents(path,title,content,tags,summary,created,updated,type)` | `db.go:78-99`, recreated in `RebuildIndex` `sqlite.go:238-241` | `Search` `sqlite.go:184-194`; `akb search` | **NEEDS DESIGN** (owner candidate: configurable index fields). The FTS5 table is `content=documents` (external-content) so a new indexed column needs DDL + rebuild, not just a query change. |
| B2 | Frontmatter keys lifted into the index are hardcoded: `"tags"` (`sqlite.go:307`), `"summary"` (`sqlite.go:326`) | `sqlite.go:306-332` | `IndexPageTx` callers; `ExtractTags`/`ExtractSummary` | **NEEDS DESIGN / OPINIONATED-LOAD-BEARING** — the indexer must name *some* fields; opening it means the user declares the indexed field set (natural home: `akb.yaml` or the template). |
| B3 | BM25 weights `(10.0, 1.0, 5.0, 3.0)` | `sqlite.go:187` | ranking only | **TRIVIAL** |
| B4 | `snippet(..., 32)` window | `sqlite.go:185` | `--json`/text snippet | **TRIVIAL** |
| B5 | Default result `limit = 10` | `sqlite.go:176-179` | `SearchOptions.Limit`; `--limit` flag absent (`cmd/akb/search.go:38-41`) | **TRIVIAL** (not even a flag today) |
| B6 | `content` indexed = **raw body verbatim**, provenance markers + annotations included | `sqlite.go:62`, `sqlite.go:287` | FTS corpus | **OPINIONATED-LOAD-BEARING** (prior research: `seam-periphery.md` §1). Search indexes what the author wrote, not the approved text. |
| B7 | Rebuild skips `index.md`/`log.md` **by basename**; silently skips pages whose frontmatter fails to parse | `sqlite.go:262-264`, `sqlite.go:275-278` | `akb index rebuild`, `akb init` | **ACCIDENT-ish**: the *policy* (managed files aren't pages) is structural; its **hardcoded spelling** is not. |
| B8 | Query escaping: strip `OR|AND|NOT` (`sqlite.go:123,130`), blank `"'()*:` (`sqlite.go:132-139`), quote each token (`sqlite.go:146-149`) | `sqlite.go:120-151` | every search | **OPINIONATED-LOAD-BEARING** — this is a deliberate "no FTS5 syntax for agents" mandate. Owner's thesis may want it removed (input = output). |
| B9 | Filters compile to `tags LIKE '%x%'` / `type = ?` / `created >= ?` | `sqlite.go:189-201` | `--tag/--type/--after` (`search.go:39-41`) | **NEEDS DESIGN** — `LIKE` substring on the flat `tags` string is the coupling point for a user-defined tag field. |
| B10 | DSN `busy_timeout(5000)&_txlock=immediate`; `SetMaxOpenConns(1)`; `PRAGMA journal_mode=WAL` | `db.go:19-21`, `db.go:149,156`, `init.go:157` | every DB open (`db.OpenKB` `db.go:142-166`) | **STRUCTURAL** (the single-connection + immediate-tx design is what makes concurrent akb processes safe; AGENTS.md anti-pattern "Do NOT run concurrent DB operations"). Exposure is a footgun, not openness. |
| B11 | Fixed schema + `VerifySchema` expected-object set; failure text "run 'akb init' to fix" | `db.go:41-49`, `db.go:160-163` | `OpenKB` | **STRUCTURAL** |
| B12 | Before/after `updated`/`created` columns are `datetime('now')` at insert | `sqlite.go:57-61` | `--after` filter reads `d.created` while `documents.created` is **index time, not page `created` frontmatter** | **ACCIDENT** — worth flagging (a search filter named "creation date" reads index time). |

**C. Git / storage (`internal/storage/git.go`)**

| # | Hardcode | Anchor | Consumer trace | Verdict |
|---|---|---|---|---|
| C1 | Managed restage set `{kb/index.md, kb/log.md}` | `git.go:27` | `gitCommit` pathspec (`git.go:483-520`); mirrored at `cmd/akb/delete.go:239` (`StageFiles`) | **TRIVIAL** if managed files become configurable; today the set is duplicated in ≥6 files (`write.go:413-419`, `append.go:111-115`, `delete.go:85,157-161`, `list.go:96`, `approve.go:307`, `index.go:116`) |
| C2 | Commit message format `akb: <verb> <path>` | `git.go:92` (`write`), `git.go:131` (`delete`), `append.go:226`, `approve.go:172`, `index.go:216,277,335`, `log.go:145`, `raw_delete.go:92`, `raw_sync.go:141`, `raw_write.go:112`, `init.go:219`, `template_delete.go:121`, `templates_write.go:331` | git history only | **TRIVIAL** (owner candidate) — but note it is **13 literal format sites**, not one; making it configurable means one `commitMessage(verb, path)` helper plus a template string. |
| C3 | Fallback identity `-c user.name=akb -c user.email=akb@local` when repo has no `user.name` | `git.go:32`, `CommitIdentityArgs` `git.go:505-515` | every commit | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** — deliberate: never writes repo config (AGENTS.md commits convention). Configuring it would invite exactly the config mutation the design avoids. |
| C4 | Index-lock retry: 12 attempts, 10ms → 250ms exponential cap | `git.go:37,40,41`, `RunGit` `git.go:527-543`, `indexLockBackoff` `git.go:547-553` | all git invocations incl. `akb init` | **STRUCTURAL** (concurrency bound), values **ACCIDENT** |
| C5 | **No git timeout at all** — `exec.Command` + `CombinedOutput`, no `context`, no `--timeout` | `git.go:528-535`; `grep -rn "context.WithTimeout"` over `cmd/ internal/` → **0 hits** | every command path | **ACCIDENT / risk** — a hung `git` (network FS, credential prompt, hook) hangs akb forever. The lock retry loop *waits* rather than bounding. |
| C6 | Commit is `--only -- <paths>` (partial commit) + merge-conflict gate `U`/`AA`/`DD` on `status --porcelain` | `git.go:520`, `git.go:490-500` | all writes | **STRUCTURAL** — this is the "akb commits only what akb touched / refuses mid-conflict" mandate |
| C7 | File mode `0600`, dir `0750` | `git.go:79-82` | writes | **TRIVIAL/ACCIDENT** |

**D. Layout & path rules (`internal/path/path.go`)**

| # | Hardcode | Anchor | Verdict |
|---|---|---|---|
| D1 | `StateDirName = ".agent-kb"`, `ConfigFileName = "akb.yaml"` | `path.go:29,34` | **STRUCTURAL as an identity, TRIVIAL as a value.** These are documented in AGENTS.md as the *only* definition site (v0.20.0 rename was BREAKING with no shim). Making the state dir name configurable is self-referential (you must find the config before you know where it is) → effectively impossible; keep hardcoded, but this is the one case where the project already decided. |
| D2 | `templates/` and `search.db` names inside the state dir | `path.go:48-52` (`TemplatesDir`, `SearchDBPath`) | **TRIVIAL** (created/consumed centrally; `init.go:121-131`, `db.go:142`) |
| D3 | **Page root is `kb/` and raw root is `raw/`, hardcoded at ~25 joins** | `path.go:122,166`; `cmd/akb/{write,append,delete,index,list,log,links,status,lint,approve,resolve_page}.go`; `internal/search/sqlite.go:222`; `internal/linkgraph/sqlite.go:177`; `internal/index/index.go:27`; `internal/storage/git.go:27`; `internal/manifest/manifest.go:39` | **STRUCTURAL-but-leaky.** Unlike D1 this is *not* centralized: prior research flagged the type→dir path as partly unvalidated (`write.go:429-448` re-checks via `AssertContained`), and the literal `"kb"`/`"raw"` strings are scattered. Opening this is a large refactor, not a config field. |
| D4 | Marker semantics: `.agent-kb` **must be a directory** (`isKBRoot` `path.go:392-395`) **and** `akb.yaml` a **regular file** (`hasKBConfig` `path.go:340-347`) | `path.go:340-347,392-395` | **STRUCTURAL** (the KB boundary) |
| D5 | Discovery bounds: depth 2, budget 1000, skip `{.git,node_modules,vendor}` + hidden dirs | `path.go:404-417` | **TRIVIAL** (a `discover.*` config block); values **ACCIDENT** |
| D6 | Guard rails: reject any `..` substring, any absolute path, cross-family `raw/`↔`kb/` | `path.go:100-104,139-145` | **STRUCTURAL** (security). Note `strings.Contains(inputPath, "..")` (`path.go:104`) also rejects a *legitimate* file named e.g. `v1..2.md` — a conservative over-rejection, not a configurability gap. |

**E. Template / type (`internal/template/`, `internal/frontmatter/`, `cmd/akb/write.go`)**

| # | Hardcode | Anchor | Verdict |
|---|---|---|---|
| E1 | Template name regex `^[a-zA-Z0-9_-]+$`, enforced at 3 gates | `cmd/akb/template_delete.go:44`; used `template.go:133`, `templates_write.go:55`, `template_delete.go:59` | **STRUCTURAL** (path-traversal guard: the name becomes `<name>.yaml` under `.agent-kb/templates/`; AGENTS.md "Template name validation"). Do not relax. |
| E2 | Every page **must** declare `type`, must match a template, and must declare non-empty `title` | `frontmatter.go:130-140` (`ValidateType`), `frontmatter.go:143-146` (`ValidateTitle`), call sites `write.go:365,370` | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** — the single biggest openness question. This *is* akb's type system; without it there is no template to validate against and the whole CEL seam loses its driver. Owner thesis ("users decide what a valid entry is") does not remove the need for a *named* type, but it does mean "a page with no type" could be an RFC if the owner wants an untyped escape hatch. |
| E3 | `type`→directory derivation from `tmpl.Dir` | `write.go:390-397,429-448`; `index.go:164`; `resolve_page.go:108` | **ALREADY USER-CONTROLLED** — `dir` is a template field (`template.go:44`, `adr.yaml` `dir: decisions`). Owner candidate "type→directory derivation" is **already flexible**; what is *not* flexible is the fallback (no `dir` → page lands wherever the CLI arg says) and that the prefix is stripped/reattached at 3 sites. |
| E4 | Template file naming `<name>.yaml`; only `.yaml` loaded | `template.go:93-95` | **TRIVIAL** |
| E5 | Old-format keys `{required, optional, body}` are hard-rejected | `template.go:66-76` (`detectOldFormat`) | **STRUCTURAL** (one-way migration) |

**F. Managed prose files: `index.md`, `log.md`, `raw/files.log`**

| # | Hardcode | Anchor | Verdict |
|---|---|---|---|
| F1 | `index.md` is machine-owned: cannot be written/appended/deleted; `akb index add/remove/rebuild` is the only channel | `write.go:413-419`, `append.go:111-115`, `delete.go:157-161`, `index.go:116` | **STRUCTURAL** (the derived-view mandate). Owner input≠output here. |
| F2 | `index.md` format: `# Index`, sections `## <Type>s` via naive pluralization `Type[:1] upper + rest + "s"` | `index.go:31-37` (`typeToHeading`), `index.go:38-44` (`headingToType` strips a trailing `s`), render `index.go:250-288` | **ACCIDENT / NEEDS DESIGN** — "decision"→"Decisions" works, "analysis"→"Analysiss" does not, and `headingToType("Analysiss")` loses the `s`. Also `Type` is the *page* frontmatter `type`, not the directory. This is akb prescribing document *structure* far beyond validity. |
| F3 | `index.md` entry line = `- [Title](Path) — Summary` | `index.go:275-283` | **NEEDS DESIGN** — em-dash separator is the parse contract (`index.go:80-111`) |
| F4 | `log.md` heading grammar `^## (\d{4}-\d{2}-\d{2}) (\w+)(?: \| (.+))?$` and render format | `log.go:27`, `log.go:130-140` | **STRUCTURAL for parse, NEEDS DESIGN for openness** — the log is append-only prose with a regex contract |
| F5 | Log `operation` is a **closed enum** `{ingest,delete,update,lint,query,distill,approve,plan}` | `cmd/akb/log.go:21`, enforced `:125`, rendered `:154` | **GENUINELY OPINIONATED-BUT-LOAD-BEARING / ACCIDENT** — the vocabulary (`ingest`, `distill`, `plan`, `query`) is *agent-workflow* language (this is the "must not be tied to a specific harness" smell: `distill`/`plan` presume a pipeline). Nothing consumes the enum except the validator. Configurable enum or free-form operation is a **TRIVIAL** change with near-zero blast radius (`grep validOperations` → 3 sites, all in `log.go`). |
| F6 | Manifest: `raw/files.log`, line format `<filename> | <sha256> | <last_updated>` (`SplitN "|" ,3`), 2 `#` header lines, atomic temp+rename write | `manifest.go:38-40,66-72,89-92`, `AddEntry` RFC3339 `manifest.go:138` | **STRUCTURAL** (raw drift detection contract; `raw/files.log` rebuilt by `akb raw sync`). Format is an internal file, but adding a `|` to a filename silently corrupts it — `ValidateFilename` (`manifest.go:226`) is the only guard. |

**G. Lint policy (`internal/lint/`)**

| # | Hardcode | Anchor | Consumer | Verdict |
|---|---|---|---|---|
| G1 | `ProvenanceDriftThreshold = 0.20` | `thresholds.go:11`, consumed `provenance.go:63` | provenance checker | **TRIVIAL** (owner candidate — this one is real). Threshold value is **ACCIDENT**; the check itself is a hardcoded *semantic policy* (declared confidence vs inline marker ratio) that a CEL lint rule could express given a marker-count variable. |
| G2 | `FreshnessHalfLifeDays=30`, `FreshnessScoreThreshold=50.0`, `SummaryMinLength=10`, `SummaryMaxLength=200` | `thresholds.go:12-15` | **NOTHING** — production grep = 0 hits; only `engine_test.go:206-216` asserts the constants equal their own literals | **ACCIDENT (dead surface)** — 4 of 5 "config" thresholds are vestigial, pinned by a test that enshrines the literals. Free deletion (or the natural seed for a config schema). |
| G3 | Checker set is a fixed registry of 10, hardcoded at the call site; no enable/disable, no severity override | `cmd/akb/lint.go:157-167` | `akb lint` | **NEEDS DESIGN** (owner candidate: lint checks + severities). Engine supports `AddChecker` (`engine.go:80-84`) so the mechanism exists; only the *selection* is hardcoded. |
| G4 | Per-checker severity is hardcoded in each checker's body: `broken_links` error+warning (`broken_links.go:48,62`), `orphans` warning (`orphans.go:39`), `empty_pages` warning (`empty_pages.go:46`), `missing_frontmatter` error (`missing_frontmatter.go:42`), `required_fields` error (`required_fields.go:70`), `index_consistency` error+warning (`index_consistency.go:61,73`), `citations` error (`citations.go:63`), `type_orphan` error (`type_orphan.go:48`), `provenance` warning (`provenance.go:69`), `cel_lint` error-on-eval-failure + rule severity (`cel.go:83,94`) | see left | `LintIssue.Severity` → exit code | **NEEDS DESIGN** — only `lint_rules[].severity` (`template.go:38`) is template-controlled; the other 9 are Go literals. |
| G5 | Exit code = "any issue with Severity == error" | `lint.go:225-232`, `lint.go:262-270` | process exit | **OPINIONATED-LOAD-BEARING** (the severity→exit coupling is what makes lint a gate) |
| G6 | Hardcoded frontmatter key `"sources"` for citations, checked against the raw manifest | `citations.go:42` | citations checker | **NEEDS DESIGN** — akb prescribes a *field name* into user frontmatter |
| G7 | Hardcoded frontmatter key `"provenance"` (map of type→confidence) and marker vocabulary `^[inferred|ambiguous|extracted]` | `provenance.go:41,57`; `markdown/provenance.go:57` (`stripProvenanceRe`); `markdown/provenance.go:18` | provenance checker + `akb approve` strip | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** — `^[inferred]` provenance markers are a core akb concept (AGENTS.md), but the *vocabulary is closed*. |
| G8 | Annotation prefix `olw-auto:` hardcoded in the regex | `markdown/annotation.go:20`; duplicated in `cel/pagebuilder.go:369` | annotation parser, `akb approve` | **ACCIDENT / OPINIONATED** — `olw-` is a foreign-project prefix (prior research notes the duplication at `pagebuilder.go:117-120`). |

**H. Write-path stamps & document mutation (`cmd/akb/`) — not constants, mandates**

| # | Mandate | Anchor | Verdict |
|---|---|---|---|
| H1 | `created` auto-stamped (UTC RFC3339) when absent | `write.go:377-379` | **OPINIONATED-LOAD-BEARING** (breaks input=output; a caller cannot withhold a timestamp) |
| H2 | `updated` auto-bumped to now on `--append`/`--frontmatter`/re-serialize | `write.go:231,335,476-483`, `append.go:176` | same |
| H3 | `is_draft` **deleted** from frontmatter on write (normalized to implicit draft) | `write.go:381-386`; semantics `frontmatter.go:151-166` | **OPINIONATED-LOAD-BEARING** — `akb approve` sets `is_draft: false` (`approve.go:155`) |
| H4 | `akb approve` strips **both** `olw-auto` annotations **and** provenance markers | `approve.go:34,153` (`markdown.StripProvenanceMarkers`) | **OPINIONATED** |
| H5 | `--dated` prepends `## YYYY-MM-DD` using the **local** date while `created`/`updated` use **UTC** | `append.go:56-57` vs `write.go:378` | **ACCIDENT** (inconsistent clock policy, undocumented) |
| H6 | Global `--no-commit` | `root.go:57` | **already flexible** |
| H7 | Exit codes 0/1/2 with prefix conventions; `mutatingCommands` set gates whether a KB is resolved | `main.go:20-23,68-90`; `root.go:26-40` | **STRUCTURAL** |

---

## Consumer trace — three mandates end-to-end

**1. Compute budget.** `engine.go:22` → `cel.CostLimit(MaxCostLimit)` (`engine.go:57`) → program →
`Evaluate` (`engine.go:74-100`) recovers `interpreter.EvalCancelledError` and maps to
`ErrComputeBudget` (`engine.go:88-90,97-99`). Callers: `write.go:606-620` (a non-true or error
result → collected into a failure message → `validationFailure{}` → **exit 1, file not written**);
`lint/cel.go:65-86` (an eval error → one `LintIssue` with `Severity: "error"` → **exit 1 for the
whole sweep**); `templates_write.go:355-378` (mockup gate); `template.go:235-255` (example display).
So one knob reaches: write gate, lint gate, template-write gate. Raising it loosens all three.

**2. Provenance drift threshold.** `thresholds.go:11` → `provenance.go:63` → `LintIssue{Severity:
"warning"}` → `LintReport.Issues` (`engine.go:101-104`) → `lint.go:157-167` registry →
`printLintText`/`printLintJSON` → exit code. Because the severity is `warning`, this threshold
**cannot fail a build**; a config knob here is cosmetic unless G4 becomes configurable too. That is
the coupling to report: *threshold configurability is inert without severity configurability.*

**3. Commit message.** 13 literal sites (`git.go:92,131` + 11 command sites, Table 2 C2) → `gitCommit`
(`git.go:461-493`) → `commitArgs` (`git.go:517-524`) → `RunGit` (`git.go:527-543`) → git history.
Nothing parses these messages back (`grep "akb: "` shows no reader). **Zero consumers** → changing
the format is free; only user expectations (`git log`) and integration test data could break.

---

## Blast-radius / verdicts

**STRUCTURALLY REQUIRED (do not open / relax):**
- KB boundary: `.agent-kb/akb.yaml` as regular file, `.agent-kb` as directory (D1, D4) — self-referential to locate otherwise.
- Path guards: absolute/`..`/cross-family rejection (D6), template-name regex (E1), path containment incl. symlinks (`path.go:181-247`).
- SQLite single-connection + WAL + busy timeout + fixed schema (B10, B11).
- Git: `--only --` partial commit, merge-conflict gate, fallback identity (C3, C6), index-lock retry (C4).
- Managed prose files: `index.md`/`log.md` not directly writable (F1); manifest format as a drift contract (F6).
- Cost limit **as a bound** (A1), lint severity→exit coupling (G5).
- The type gate (E2) **conditionally**: load-bearing only because it names the template. If the owner wants an untyped escape hatch, E2 is the change, not a config field.

**TRIVIAL config fields (mechanical, near-zero blast radius):**
A1 value, A4 (`--now`), B3, B4, B5, C1/C2 (managed set + commit format), D5, E4, F5 (log operation
enum), G1 (drift threshold — but see the severity coupling).

**NEEDS DESIGN (must define the consumer contract first):**
A2 (cache key under extensible env), B1/B2/B9 (indexed fields, lifted keys, filter mapping), E3
fallback behaviour, F2/F3 (`index.md` structure), G3/G4 (checker set + severities), G6 (`sources`).

**HISTORICAL ACCIDENT (free to delete or to seed a config schema):**
`config.Created` (write-only), `FreshnessHalfLifeDays`/`FreshnessScoreThreshold`/`SummaryMinLength`/
`SummaryMaxLength` (dead; pinned by `engine_test.go:206-216`), `typeToHeading`'s naive pluralization
(F2), the `olw-auto` prefix (G8), the UTC-vs-local clock split (H5), the missing git timeout (C5),
`documents.created` = index time while `--after` is documented as "creation date" (B12).

**GENUINELY OPINIONATED-BUT-LOAD-BEARING (the owner should keep these deliberately, and say so):**
The CEL env surface (A3), body-verbatim indexing (B6), FTS5-syntax stripping (B8), the write-path
stamps H1–H4, the provenance vocabulary (G7), `is_draft` semantics (H3).

**Contract with the owner's candidate list (§4 of the brief):**

| Owner candidate | Today | Real gap? |
|---|---|---|
| CEL compute budget | `const MaxCostLimit = 100000` (`engine.go:22`) | **REAL**, TRIVIAL |
| Provenance drift threshold | `const 0.20` (`thresholds.go:11`) | **REAL but inert** without G4 severity config; and it is a *policy*, expressible as a CEL lint rule if marker counts were exposed |
| Lint checks + severities | 10 hardcoded at `lint.go:157-167`; 9 severity literals | **REAL**, NEEDS DESIGN (mechanism `AddChecker` already exists) |
| Search-index fields | columns + `tags`/`summary` key names hardcoded (`db.go:78-99`, `sqlite.go:307,326`) | **REAL**, NEEDS DESIGN (external-content FTS5 ⇒ DDL + rebuild) |
| Type→directory derivation | `tmpl.Dir` is **already template-controlled** (`template.go:44`, `write.go:390`) | **ALREADY FLEXIBLE** — only the no-`dir` fallback and the 3 duplicated strip/attach sites are rigid |
| `is_draft` lifecycle | implicit draft, stripped on write, `approve` flips it (`write.go:381-386`, `frontmatter.go:151-166`, `approve.go:155`) | **REAL but opinionated** — it is a workflow mandate (review gate), not a validity rule; "maximal openness" would move it into the template as ordinary frontmatter, at the cost of the `akb list`/`approve` draft pipeline |
| Commit message format | 13 literal sites, **zero consumers** | **REAL**, TRIVIAL |

---

## Feasibility sketches (≤ few lines each; not designs)

**Custom CEL functions declared by akb.yaml / template.** The repo is on `cel-go v0.28.0`
(`go.mod:12`), *not* v0.31 — so the copy-on-write `Env.Extend` behaviour prior research cited must
be checked against v0.28: `cel.Function(name, cel.Overload(id, args, result, cel.FunctionBinding(fn)))`
exists (`$GOMODCACHE/github.com/google/cel-go@v0.28.0/cel/decls.go:195,303,339`), and
`(*Env).Extend(opts ...EnvOption) (*Env, error)` exists (`cel/env.go:478`) with a documented
"should not share memory" caveat. So:

```go
func NewEnvWithFuncs(funcs []FuncDecl) (*cel.Env, error) {
    base, _ := NewEnv()
    return base.Extend(funcs...)   // each: cel.Function(...)+cel.Overload(...)+cel.FunctionBinding(...)
}
```

Cost-accounting: a Go `functions.FunctionOp` runs **outside** CEL's cost tracking — `CostLimit`
only charges interpreter steps, so a function that loops over `page.content.raw` is an unbounded
escape hatch around `MaxCostLimit`. Any template-declared function therefore needs its own
per-call bound (or the whole feature is a DoS vector on the write path where errors fail closed).
Trust: templates are **git-committed KB content**, i.e. the same trust level as the pages, and the
same level as CEL rule *strings* today — but a rule string is bounded by the interpreter, whereas a
native binding is arbitrary Go reachable from a template edit. Recommended posture: a template may
select from a **closed, akb-shipped function registry** (e.g. `regex_has_group`, `url_host`,
`path_base`) rather than inject code; that keeps env description declarative and preserves the
cost model. The cache key (A2) must become `(envID, expr)` for any of this to be correct.

**Per-template env extension.** Same `Extend` per template, memoized by template name; the cost is
N envs instead of 1, and every `CompileRule` call site (6 sites, Table A1) must pass the
template-scoped env. Feasible; the risk is cache/A2 plus error-message divergence (`template write`
compiles against the same env — `templates_write.go:108-121`).

**Input = output on the write path.** The stamps (H1–H3) are the only unconditional mutations on the
plain-stdin path; prior research already established that byte passthrough happens unless
`needsReserialize` (`seam-core-pipeline.md` §4; `write.go:374,482-497`). A `--no-stamp` (or a
template flag) is a TRIVIAL change with the blast radius: `created`/`updated` temporal CEL rules
(`adr.yaml` `temporal_created`, `adr_stale`) start failing on unstamped pages — which is *correct*
under the owner's thesis (the user's rules, not akb's defaults).

---

## Open questions for synthesis

1. **Which layer owns behavior config — `akb.yaml` or the template?** Indexed search fields and
   lint severities are per-*KB* today, but `lint_rules[].severity` is already per-*template*
   (`template.go:38`). Picking one layer determines whether the openness surface is one file or N.
   The template is git-committed KB content; `akb.yaml` is too. Same trust level, different lifecycle.
2. **Is the compute budget a user knob or a hard safety bound?** A native CEL function binding
   bypasses it entirely (above). If openness includes "user-supplied functions", the bound must be
   re-justified; if not, A1 is cosmetic.
3. **Does maximal openness include removing the type gate (E2)?** This is the one mandate whose
   removal changes the architecture (no template ⇒ nothing to validate against ⇒ the JSON-Schema/CEL
   seam loses its driver). The owner's phrase "users decide what a valid KB entry is" is satisfiable
   *with* the type gate (the *template* is the user's declaration); it is not satisfiable *without*
   redefining what akb validates.
4. **Should dead surface be deleted or repurposed?** `config.Created` and the four unused lint
   thresholds (`thresholds.go:12-15`, pinned by `engine_test.go:206-216`) are the only pre-existing
   "config" in the repo. Deleting them is honest; repurposing them as the first fields of a real
   config schema is cheaper — but they carry no semantics worth keeping.
5. **`--no-commit` is the sole existing behavioral escape hatch** (`root.go:57`). It suggests the
   project's own precedent for openness is a *flag per invocation*, not `akb.yaml` — worth deciding
   deliberately rather than by drift, since flags are agent-hostile (long command lines) and
   `akb.yaml` is agent-friendly (one edit, many invocations).
6. **C5 (no git timeout)** is the only hardcode here with a liveness risk rather than a policy
   cost — it deserves a decision independent of the openness framing.

---

### Prior research relied on (not re-derived)

- `json-cel/research/akb-recon.md` §2 (env construction, cost limit, cache key), §3 (write flow, stamps), §7 (constraint inventory).
- `json-cel/research/seam-core-pipeline.md` §1 (page map, derived views), §4 (`needsReserialize`, byte passthrough).
- `json-cel/research/seam-periphery.md` §1 (search/indexer: body verbatim, `tags`/`summary` lifting, limit 10), §3 (checker table + severities + `provenance` threshold), §4 (index/log formats, log operation enum).
