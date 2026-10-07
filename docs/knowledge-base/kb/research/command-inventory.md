---
grammar: 1
type: research
title: "akb CLI user-facing surface inventory (JSON-vs-CEL recon)"
status: final
provenance: agent-drafted
run: json-cel
as_of: 2026-09-26
created: 2026-09-26
updated: 2026-09-26
scope:
  - cmd/akb/**
  - internal/**
tags: [cli, inventory, recon, json]
summary: "Inventory of 31 akb leaf commands: purpose, inputs, outputs/exit codes, subsystems touched, and which already expose --json — the surface a JSON-I/O design must fit."
informs:
  - RFC-001
  - RFC-002
---
# akb CLI — user-facing surface inventory (JSON-vs-CEL recon)

Read-only recon. Sources: `cmd/akb/*.go`, `internal/{cel,template,frontmatter,lint,search,linkgraph,markdown,path,manifest,storage}`, `.../agent-memory/DESIGN.md`.
Exit-code contract: `cmd/akb/main.go:20-24` — `0` success; `1` result-actionable (page validation failed, lint issues, raw drift); `2` invocation mistake (`usage:`) or akb fault (`internal:`).
KB resolution: every command except `init`/`discover`/`skill install` reaches a base via `path.ResolveKB(--kb | $AKB_KB)`; only commands in `mutatingCommands` (`cmd/akb/root.go:27-41`) print the `kb: <name> (<path>)` stderr identity line.

---

## 1. COMMAND INVENTORY

31 leaf commands. "Subsystems" uses the abbreviations: fm=frontmatter, tpl=template, db=db/index/search/linkgraph SQLite, idx=kb/index.md, log=kb/log.md, raw=manifest, git=storage, cfg=config, path. flag = reads/writes files.

| Command (file) | Purpose | Inputs | Outputs / exit | Subsystems | `--json` |
|---|---|---|---|---|---|
| `init <name>` (init.go:29) | Create a KB tree + git repo + config + DB | args: dir name; flag `--description` | stdout "Initialized KB..."; 1/2 | fm–, tpl–, db(create schema), raw(seed files.log), git(init+commit+identity), cfg(write akb.yaml), path | no |
| `discover [dir]` (discover.go:21) | List KBs near a dir (report-only) | optional dir; `--json` flag | list or JSON; 1/2 | path only (no base resolved) | **yes** → `[]DiscoveredKB{name,path,description?}` (path.go:420) |
| `status` (status.go:19) | Name, path, page count, git clean/dirty | none | 4-line text; 1/2 | cfg, git(`git status --porcelain`), path, fm(count only) | no |
| `write <path>` (write.go:58) | Create/overwrite a page from stdin, validate | **stdin** (fm+body, or bare body); `--append`,`--dated`,`--frontmatter k=v` | "Written to/Appended to/Updated frontmatter for <rel>"; 0/1(validation)/2 | fm, cel, tpl, db(search FTS+linkgraph), git(commit `akb: write <path>`), cfg, path | no |
| `read <path>` (read.go:24) | Dump raw page bytes | path arg | body verbatim to stdout; 1/2 | storage(filesystem), path, cfg | no |
| `append <path>` (append.go:36) | Append stdin to an existing page body | **stdin**; `--dated` | "Appended to <rel>"; 0/1/2 | fm, cel, tpl, db, git, cfg, path | no |
| `delete <path>` (delete.go:34) | Delete page (or `--orphans` batch) | path arg; `--orphans`,`--force` | "Deleted <rel>"/preview; 1/2 | fm, db(remove FTS+links), idx(RemoveEntry), log(append), git, path | no |
| `list` (list.go:27) | List all kb/**.md pages | none; `--json` | text or JSON; 0/2 | fm(is_draft), path | **yes** → `[]{path,is_draft}` (list.go:20) |
| `search <query>` (search.go:22) | FTS5 BM25 search | query arg; `--json`,`--tag`,`--type`,`--after` | lines or JSON; 1/2 | db(FTS5), path | **yes** → `[]{path,title,summary,snippet,rank}` (search.go:95) |
| `links <path>` (links.go:21) | Outbound / broken / ambiguous links | path arg | 3 text sections; 1/2 | db(linkgraph), path | no (missing) |
| `backlinks <path>` (links.go:30) | Inbound links | path arg; `--json` | list or JSON; 1/2 | db(linkgraph), path | **yes** → `{backlinks:[{source_page,raw_target,display,resolved_to}]}` (links.go:154-167) |
| `orphans` (links.go:38) | Pages with zero inbound links | none | paths; 1/2 | db(linkgraph), path | no (missing) |
| `index show` (index.go:47) | Print kb/index.md verbatim | none | file bytes; 1/2 | idx, path | no |
| `index add <path> <summary>` (index.go:57) | Add/update an index entry | args; reads page fm for title/type | "Added <rel> to index"; 1/2 | idx, fm, tpl(type-dir fallback), git(commit `akb: index add`) | no |
| `index remove <path>` (index.go:68) | Remove an index entry | arg | "Removed..."; 1/2 | idx, git | no |
| `index rebuild` (index.go:77) | Regenerate index.md + FTS + linkgraph from disk | none | "Index rebuilt"; 1/2 | idx, db(RebuildIndex+RebuildLinks), fm, git | no |
| `log show` (log.go:41) | Print/filter log entries | `--last`,`--type` | rendered md entries; 1/2 | log, path | no |
| `log append <op> <desc>` (log.go:55) | Append a log entry | args; `--title`; op ∈ {ingest,delete,update,lint,query,distill,approve,plan} (log.go:19) | 0/2 | log, git | no |
| `raw write <path>` (raw_write.go:22) | Write a raw file + SHA-256 manifest | **stdin** | "Written to raw/<rel>"; 1/2 | raw(manifest), git(commit raw pathspec), path | no |
| `raw read <path>` (raw_read.go:22) | Dump a raw file | path arg | bytes; 1/2 | storage, path | no |
| `raw list` (raw_list.go:20) | List raw/ files (excl. files.log) | none | paths; 1/2 | path | no |
| `raw status [path]` (raw_status.go:46) | Manifest-vs-disk drift | optional path; `--json` | report or JSON; 1(drift)/2 | raw(manifest), path | **yes** → `{files:[{path,status}],summary:{modified,untracked,missing}}` (raw_status.go:14-31) |
| `raw sync` (raw_sync.go:20) | Reconcile manifest with disk | none | summary; 1/2 | raw, git | no |
| `raw delete <path>` (raw_delete.go:23) | Delete raw file + manifest entry; scans `sources:` | path arg | "Deleted raw/<rel>" + referencing pages; 1/2 | raw, fm(sources scan), git | no |
| `lint` (lint.go:26) | Run 10 checkers incl. CEL lint rules | none; `--json` | report or JSON; **1 if any error-severity issue**/2 | fm, cel, tpl, db(linkgraph rebuild), manifest(citations), markdown(provenance), idx(index_consistency), path | **yes** → `{issues:[{check,rule_id?,message,path,severity}],summary:{total,pages_checked,by_check}}` (lint.go:232-246) |
| `approve <path>` / `--all-drafts` (approve.go:28) | Strip annotations/provenance, set `is_draft:false`, reindex | path arg or `--all-drafts` | "Approved '<path>'"/"Approved N drafts"; 0/1(required-field refusal)/2 | fm, tpl(required fields), cel–no, markdown(strip), db(reindex FTS+links), git | no |
| `template get <name>` (template.go:38) | Writer view (schema+requirements), or `--example` pass mockup, `--full` raw YAML, `--examples` embedded | name; 4 bool flags | YAML/markdown; 1/2 | tpl, fm(parse mockup), cel(validate mockup warnings), path | no |
| `template list` (template.go:43) | List template names+descriptions | `--examples` | lines; 1/2 | tpl, path | no |
| `template write <name>` (templates_write.go:35) | Author a template; compile CEL + test pass/fail mockups | `--template`(req),`--pass`,`--fail`,`--force`; files | "Template written"/diff; 1/2 | tpl, cel(compile+evaluate), fm(parse mockups), db(page count), git | no |
| `template delete <name>` (template_delete.go:22) | Delete template + mockups; impact count | name; `--force` | count warning/delete; 1/2 | tpl, db(count), lint(auto-run), git, path | no |
| `skill install <name>` (skill.go:33) | Copy an embedded skill to `--location` | name; `--location`(req) | 0/2 | skill(embed), path | no |

**JSON surface today:** 6 of 31 commands emit JSON (`discover`, `list`, `search`, `backlinks`, `raw status`, `lint`); all are newline-`Marshal`/indented `Encoder` of an ad-hoc struct. No input command consumes JSON; no `--format`, no `--json-schema`, no stdin-JSON anywhere. `akb read`/`links`/`orphans`/`status`/`template get` have no machine-readable mode.

---

## 2. DATA FLOWS

### 2a. Write path (`cmd/akb/write.go`, mirrored by `append.go` and `approve.go`)

**In (from stdin + args + flags):**
- body bytes; `--frontmatter k=v` pairs; `--append`/`--dated`/`--no-commit`/`--kb`.
- Frontmatter parsed to `ParsedFrontmatter{Type,Title,Fields map[string]any}` (frontmatter.go:18-40). `type`/`title` are pulled out; everything else lands raw in `Fields`.
- Auto-managed fields written back: `created` (new page only, RFC3339 UTC, write.go:376-378), `updated` (bumped on any content change / overwrite / explicit `--frontmatter updated=`, write.go:208-213, 455-460, 500-503), `is_draft` (deleted if `true` on new write, write.go:380-385; set `false` by approve).

**Validation inputs (the CEL `page` map, built by `cel.BuildPage`, pagebuilder.go:45-82):**
- `page.file {path,name,dir}` (relPath is `kb/<typeDir>/<cleanPath>`).
- `page.frontmatter`: all fields + `type`/`title`; RFC3339 and date-only strings converted to `time.Time` (`convertDateField`, pagebuilder.go:27-43).
- `page.content {raw, word_count, char_count}` — derived: `len(strings.Fields(bodyStr))`, `len(body)`.
- `page.ast {headings[{level,text,line}], links[{target,text,is_wikilink,line}], code_blocks[{language,line}]}` — goldmark parse + wikilink merge (pagebuilder.go:163-310, 322-343).
- `page.akb {provenance_markers[{type,position}], annotations}` (pagebuilder.go:71-74, 347-...).
- `old_page`: same map for the on-disk pre-modification page; `nil` on create (BuildOldPage, pagebuilder.go:85-111).
- `now`: `time.Now()` injected per rule (write.go:594-598).
- Pre-CEL gate: `checkRequiredFields` (required_fields.go:20-31) blocks on schema `required:true` presence before rules run.
- Rules: `tmpl.Validations[]` (template.go:27-32); **fail closed** (write.go:583-616) — unevaluable rule ≠ pass.

**Out (side effects, in order):** one repo lock (`storage.LockRepo`) wraps read→write→commit→index. `store.Write` (git commit `akb: write <path>`); one SQLite tx covering `searcher.IndexPageTx(relPath,title,body,tags,summary,type)` + `linkgraph.UpdatePageLinksTx(writeContent)`. Tags/summary derived by `search.ExtractTags/ExtractSummary` from `Fields` (write.go:518-520). Stdout line is the only structured-ish output.

### 2b. Read path

- `akb read` — raw bytes only; no parse, no JSON (read.go:60-64).
- `akb list` — walks `kb/`, parses each file only to read `is_draft` → `{path,is_draft}` (list.go:20-23, 62-92).
- `akb search` — `db.OpenKB` FTS5 → `SearchResult{Path,Title,Summary,Snippet,Rank}` (searcher.go:13-20); snippet is FTS5-generated, not body text.
- `akb links/backlinks/orphans` — linkgraph rows `Link{SourcePage,RawTarget,Display,ResolvedTo}` (sqlite.go:16).
- `akb lint` — the heaviest read: loads templates + manifest, **rebuilds the link graph**, walks pages into `PageData{RelPath,Content,Body,Frontmatter,ProvenanceMarkers,Annotations,HasFrontmatter}` (engine.go:56-64), then 10 checkers. CEL lint path builds the same `page` map but with **no `old_page`** and one shared `now` (cel.go:38-52, 55-70).
- `akb status` / `index show` / `log show` — text/concat, no validation.
- `akb raw *` — manifest SHA-256 + `sources:` frontmatter scan on delete (raw_delete.go:106-...).

---

## 3. ROADMAP (DESIGN.md — agent-memory design, amended 2026-09-23/24)

**Relationship agent-memory ↔ akb (explicit):** DESIGN.md states akb **is** the memory substrate — D6 "Invest in agent-kb" (line 19); §2 "Build the memory layer as **agent-kb knowledge bases in separate git-tracked KB repos** … schema-validated, provenance-tagged markdown records" (lines 23-38). The differentiators are named as **"schema-enforced writes (akb's CEL templates + lint)"** (lines 37-38) and structured failure memory. Wave 0-2 + backlog-correctness + parser-convergence already merged to master (lines 98-99, 256-260, 271-289). **DESIGN.md contains no mention of JSON Schema or JSON I/O** — the JSON-vs-CEL proposal is a new direction relative to the ratified design.

Features touching akb/validation/JSON I/O:

| # | Feature | DESIGN anchor | Interaction |
|---|---|---|---|
| 1 | **Six new content templates** `decision`, `known-error`, `lesson`, `convention`, `plan-distillate`, `harness-gotcha` with CEL-enforced H2 structure | §4.2 lines 112-125 | The core authorship target for the JSON/CEL boundary; each has structural (headings) + field constraints |
| 2 | **Bi-temporal frontmatter** `valid_from`/`valid_to` (+`created`/`updated`) on every template | §4.2 lines 126-128 | Needs generalized date conversion (P2, line 324) so `timestamp()` comparisons work |
| 3 | **Provenance classes** `source: user-explicit\|agent-inferred\|tool-output`, `confidence`, `sources:` list (cited against raw manifest) | §4.2 lines 129-132 | `source` is an enum → JSON-Schema-native; `sources` validation already exists in citations lint |
| 4 | **CEL `matches()`** for `lesson` date-regex H2s | §4.2 line 121; §8 lines 393-396 | RESOLVED as working; an example of CEL-only content validation |
| 5 | **`--dated` append flag + log-op enum** | §5 lines 256-257, 306-308 | ✅ landed (P3); log op enum now `ingest,delete,update,lint,query,distill,approve,plan` (log.go:19) |
| 6 | **auto-`updated` bump** | §5 line 306 (P1, ✅ landed) | Handled in write path (write.go:455-460) |
| 7 | **Write-time guards**: CEL validation, provenance, draft state, lint sweep — **to add: secret scanning on write** (Hindsight 45-regex set) | §4.3 lines 173-174; §5 line 309 | Secret scan is a new write-gate, neither schema nor CEL |
| 8 | **Typed gates**: `gate: {command: "akb lint --kb <path> --json", output: "json", schema}` | §4.3 lines 183-186 | **Depends directly on `akb lint --json` shape**; open Q on lint exit-nonzero vs gate PASS (line 405-408) |
| 9 | **T1 — enforce decorative schema: type/enum enforcement** | §5 line 304 (wave 3) | Native JSON Schema is a natural fit for type/enum; currently schema only checks `required` presence (required_fields.go:23) |
| 10 | **FTS5 phrase/prefix subset; `akb page info` / `akb mentions`** (new S/M commands) | §5 lines 307-308 | New commands that would likely want `--json` |
| 11 | **`akb stale`-equivalent as CEL lint rules (not a command)** | §5 lines 311-313 | Validation stays template-CEL-owned, deliberately |
| 12 | **T3 retrieval**: read-only `kb_search` tool registered in child sessions; main session uses `akb` CLI | §4.4 lines 218-229 | Consumes search output; caps + `--kb` addressing |
| 13 | **T1/T2 injection**: standing rules (`tier: standing`) + bounded index snapshot rendered at `session_start` | §4.4 lines 201-217 | Reads frontmatter/index; any new frontmatter key must survive |
| 14 | **index rebuild ingestion bypass (D11)** | §8 lines 411-414; §7 risk 4 lines 357-368 | Pages entering via rebuild never pass CEL/secret scan — a validation seam risk |
| 15 | **Trust-boundary wave** read-side containment file list incl. `cmd/akb/lint.go`, `list.go`, `status.go`, `raw_delete.go` + index-rebuild paths | §5 lines 266-268 | Any new JSON reader must respect the same containment |
| 16 | Phases 0-3 roadmap: templates→pilot, companion extension+contracts, lifecycle/gardening, factory patterns | §6 lines 318-347 | Phase 1 needs stable machine-readable lint/search output |

---

## 4. PRELIMINARY SEAM SKETCH

Keys: **[meta]** = validation input is pure frontmatter → JSON-Schema natural; **[content]** = needs body/AST/derived data → CEL natural; **[ctx]** = needs `now`/`old_page`/linkgraph/git → CEL-only; **[n/a]** = no validation relevance.

| Command | Classification | Notes |
|---|---|---|
| `write` | **[meta]+[content]+[ctx]** | required-fields = meta; word_count/AST/headings/links = content; `old_page`+`now` = ctx. The mixed case the design must split. |
| `write --frontmatter` | **[meta]** | metadata-only mutation; no body change (write.go:163-246) |
| `append` | **[content]+[meta]** | body grows; `now` stamps `updated` |
| `approve` | **[meta]+[n/a]** | required-field gate (meta) + strips annotations/provenance (content) but runs **no CEL rules** (approve.go:150-176) |
| `template write` | **[meta]** (authoring) + **[ctx]** (mockup self-`old_page` test) | It *defines* the schema + CEL rules; templates_write.go:159-186 evaluates pass/fail/optional-stripped/self-update |
| `template get` / `list` | **[n/a]** | read-only; exposes `Schema` (writer view) + raw CEL via `--full` |
| `template delete` | **[n/a]** | |
| `lint` | **[meta]+[content]+[ctx]** (cross-page) | `cel_lint` has `now` but **no `old_page`** (cel.go:55-70); structural checkers need linkgraph/manifest |
| `index rebuild` | **[n/a]** ⚠ | write-side none, but **ingests unvalidated pages** — validation-relevant on the ingestion edge (§8 line 411) |
| `index add/remove/show` | **[n/a]** | |
| `delete` | **[n/a]** | removes; no rules |
| `read` | **[n/a]** | |
| `list` | **[meta-read]** | reads only `is_draft`; the one field it surfaces |
| `search` | **[n/a]** | derived index; `--tag`/`--type` filters map to frontmatter fields |
| `links` / `backlinks` / `orphans` | **[n/a]** | linkgraph projections |
| `status` | **[n/a]** | |
| `discover` | **[n/a]** | |
| `log show/append` | **[n/a]** | enum op validated in code, not schema |
| `raw write/read/list/status/sync` | **[n/a]** | separate storage; manifest hashes not CEL |
| `raw delete` | **[meta-read]** ⚠ | scans `sources:` frontmatter to warn (raw_delete.go:106+) — a cross-page read amendable to a schema-aware check |
| `init` | **[n/a]** | |
| `skill install` | **[n/a]** | |

**Commands whose data resists a clean split:**
- **`write` / `append`** — one command, three validation flavours. The seam must be expressible per-template without bifurcating the command.
- **`approve`** — reads metadata (required fields) but also rewrites content (strips `olw-auto`/provenance) while running zero CEL rules; a JSON-Schema-only world could not express the strip, and a CEL-only world duplicates the required-field gate.
- **`index rebuild`** and **`raw delete`** — validation-relevant only as *readers* of unvalidated content; they sit outside both write-time gates today.
- **`template get` (writer view)** — emits `schema.frontmatter` + textual `requirements`, i.e. it is already the human-facing bridge between a structural schema and CEL prose; a JSON Schema would want a machine-readable counterpart here.
- **`lint`** — must aggregate both schema-type findings and CEL findings into one `LintIssue{check,rule_id?,message,path,severity}` list; the boundary is inside one command's output, not between commands.

## Open questions for the design
1. Does native JSON Schema replace `schema.frontmatter` (`FieldSchema{type,required,enum}`, template.go:15-24), or sit beside it in the template YAML? The required-fields gate (required_fields.go) is the only current structural enforcer.
2. `akb lint --json` is load-bearing for DESIGN §4.3 typed gates (line 184); any output-shape change ripples into the companion extension gate binding.
3. `page.content` (word_count/char_count) and `page.ast.*` are derived at validation time, not stored — a JSON input format (JSON I/O flags) would have to either accept them as given or recompute them from a body field.
4. `old_page`/`now` cannot be expressed by JSON Schema; if JSON I/O is added for writes, the CEL half still needs the pre-write page read (git/filesystem), which JSON payloads would not carry.
