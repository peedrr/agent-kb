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
- cmd/akb/**
- internal/frontmatter/**
- internal/storage/**
- internal/index/**
- internal/log/**
status: final
summary: "Audit of akb's write path: created/updated injection, is_draft deletion, whole-frontmatter re-marshal, type-dir prefix forcing — which prescriptions are structural, load-bearing, or accidental."
tags:
- openness
- write-path
- audit
- frontmatter
title: Openness audit — write path & frontline policy
type: research
updated: "2026-09-27"
---
# Openness Audit — WRITE PATH & FRONTLINE POLICY (scout 4/4, read-only)

Repo root: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb`
Scope: `cmd/akb/{write,append,approve,read,delete,resolve_page,required_fields,index,log}.go`, `internal/{frontmatter,storage,index,log}/`, `internal/path/path.go`.
Prior research cited, not re-derived: `.pi/subagents/proposals/json-cel/research/{command-inventory,seam-core-pipeline,seam-periphery,akb-recon}.md`.

Question being answered: **what does akb still prescribe on the write path, who consumes the prescription downstream, and is the prescription structurally required, a historical accident, or opinionated-but-load-bearing?**

---

## 0. Owner's headline claims — verified

| Claim | Verified at | Verdict |
|---|---|---|
| `created`/`updated` injection | `write.go:377-379` (created, new page only), `write.go:231` (explicit `--frontmatter updated`), `write.go:334-335` (`--append` inside write), `write.go:478-481` (overwrite), `append.go:176` | TRUE |
| `is_draft: true` silently deleted on write | `write.go:382-385` (`delete(fm.Fields,"is_draft")` when `true`/`"true"`) | TRUE; the inverse (`is_draft: false` from stdin) is preserved — asymmetric |
| Whole-frontmatter re-marshal destroys comments/key-order/quoting/indent | `write.go:242-253`, `write.go:345-352`, `write.go:484-495`; `append.go:186-194`; `approve.go:157-170`. Key order: `seam-core-pipeline.md:205-213` (measured: `yaml.Marshal(map[string]any)` sorts lexicographically) | TRUE. Comments/indent/quoting are lost because the write side is a map round-trip, never a document round-trip |
| Verbatim passthrough only when `created`+`updated` present **and** page is new | `write.go:374` `writeContent = stdinContent`; reserialize is triggered by `write.go:377-379` (no `created`), `write.go:382-385` (`is_draft:true`), `write.go:478-481` (`updated` absent **or** page exists). So verbatim ⇔ page new ∧ `created` present ∧ `updated` present ∧ `is_draft` not true | TRUE |
| Template dir prefix path rewrite | `write.go:418-424` (strip `dir/` prefix if it matches the type's dir), `write.go:433-441` (re-prefix with the type's dir) | TRUE — and stronger than stated: the prefix is *forced*, see M8 |
| `--append` always re-marshals | `write.go:345-352`, `append.go:186-194` (no verbatim branch on either append path) | TRUE |
| `--dated` blank-line behavior | `datedSection` (`append.go:62-65`) returns `"## YYYY-MM-DD\n\n" + content`; both append paths join with `string(body) + "\n" + appended` (`write.go:342`, `append.go:183`) | TRUE: the tests use fixture bodies with **no trailing newline** (`append_test.go:432-484`, `appendSetupTestKB` writes `"...---\nOriginal body."`), so they assert `body\n## date\n\ncontent`. Extract on a real file returns the body **including** its trailing `\n` (`frontmatter.go:92-120`), so the same append yields a blank line before `##`. The `--dated` contract is only correct for trailing-newline-less files |
| `--frontmatter` stores all values as strings | `write.go:210-227` — `fm.Fields[key] = value`, `value := parts[1]` (string) | TRUE; no type coercion, no YAML parse of the value |

---

## Findings — mandate tables

Legend for verdict: **SR** = structurally required · **HA** = historical accident · **OL** = opinionated-but-load-bearing.

### A. Type / title / template identity

| # | What akb forces | Where | Consumer(s) | Configurable / derived / removable? | Verdict |
|---|---|---|---|---|---|
| M1 | Frontmatter **must** contain a non-empty `type` | `frontmatter.go:130-132` (`ValidateType`), called `write.go:194`, `write.go:235`, `write.go:318/365`, `append.go:162`, `write.go:399…` | `templates[fm.Type]` lookup (validation), search `documents.type` + `--type` filter (`sqlite.go:60,180`), index.md grouping/headings (`index.go:261-274`), lint `required_fields`/`cel`/`type_orphan` (`lint/required_fields.go:51`, `lint/cel.go:52`, `lint/type_orphan.go:38`) | Redundant with schema `required: true`: both embedded templates already declare `type: required` (`embedded/adr.yaml:17-18`, `embedded/note.yaml:8-9`) and `checkRequiredFields` enforces it (`required_fields.go:19-29` + `frontmatterKeyPresent` special-casing `type`). Hardcoded check could be deleted once the schema-driven path is the only gate. Removing it entirely is unsafe: `parsed.Type` is routed out of `Fields` (`frontmatter.go:70-82`), so `type` becomes unreachable by CEL `has()` if the field moves | **HA** (duplicate gate), the *field* itself **OL** |
| M2 | Frontmatter **must** contain a non-empty `title` | `frontmatter.go:139-142` (`ValidateTitle`), same call sites | FTS title column with BM25 weight 10.0 (`sqlite.go:186`), `index.md` entry text (`index.go:229,274`), log entry title (`delete.go:127,179,235`), `akb index add` entry title (`cmd/akb/index.go:198`) | Same shape as M1: `title: required: true` in both embedded templates, so the hardcoded check is a duplicate. Full removal is a behaviour change: pages without `title` would pass write but render `- [](kb/…)` in index.md and score 0 against the title weight | **HA** (duplicate gate), field **OL** |
| M3 | `type` must name an existing template → **no page may exist without a template** | `frontmatter.go:134-136`; re-checked at `write.go:239-242, 390-394, 502-505`, `append.go:200-203` | Every write-path consumer via `templates[fm.Type]`. Lint has the *opposite* posture: `type_orphan` (`lint/type_orphan.go:46`) reports it as an issue and `required_fields` **skips** unknown types (`lint/required_fields.go:52-55`), and approve skips too (`approve.go:219-221`) | This is the single most load-bearing mandate: the template *is* the validity definition. With JSON Schema + CEL per-template, "unknown type" could become "no schema → no validation" (lint's posture) instead of a write refusal. Blast radius of relaxing: `requiredFieldsRefusal`/`checkRequiredFields` no-op, CEL rules never run, `dirFromType` empty (page takes the caller's literal path) | **OL** |
| M4 | `type`/`title` YAML values must decode as `string` | `frontmatter.go:72-83` (`field 'type' must be a string, got %T`) | — | If types become schema-owned, this is the only place that can reject e.g. `type: 3`. Constraining the *shape* of a user field is exactly the kind of thing the user's schema should decide | **HA** |
| M5 | The page file **must** declare `type`/`title` in frontmatter *itself*, not in a sidecar/metadata file | `frontmatter.go:56-58` (no frontmatter → error), `frontmatter.go:64-66` (empty map → error) | Same as M1/M2 | Rejected outright: a KB whose pages are plain markdown (no frontmatter) cannot be written at all (`write.go:361-365`); `list` treats parse failure as draft (`list.go:107-110`), `search.RebuildIndex` silently *skips* unparsable pages (`sqlite.go:270-273`), `index.RebuildIndex` likewise (`index.go:204-207`) | **OL** for agent-KBs, **HA** if "akb provides the hands" — a body-only page is a legal markdown file |

### B. Auto-managed fields

| # | What akb forces | Where | Consumer(s) | Verdict |
|---|---|---|---|---|
| M6 | Injects `created: <RFC3339 UTC>` when absent | `write.go:377-379` | **None inside akb.** `documents.created` is `datetime('now')` at index time (`sqlite.go:60`), so `akb search --after` (`search.go:41`, `sqlite.go:177`) reads *index* time, not this field. The only reader is user CEL: `embedded/adr.yaml:136-151` (`created_immutable`, `updated_not_before_created`, `temporal_created`) | **HA** as a tool mandate; **OL** once a template declares it |
| M7 | Bumps `updated` on every content change; honours an explicit value only on create (`write.go:478-481`) and via `--frontmatter updated=` (`write.go:221-231`) | `write.go:231, 334-335, 478-481`; `append.go:176`; deliberately **not** in approve (`approve.go`, pinned by `write_test.go:1273-1296`) | Same as M6 — only CEL (`adr.yaml:142,183,190`; `note.yaml:37` "stale note" lint). Never indexed, never rendered | **HA**; the approve asymmetry is an accident |
| M8 | Deletes `is_draft` when it is `true` in stdin; keeps `false`; treats absence as draft (`frontmatter.go:148-163`) | `write.go:382-385` | `akb approve` gate (`approve.go:144,324`), `akb list` `[DRAFT]` marker + `--json.is_draft` (`list.go:24,64,110`), transitive: approve rewrites `is_draft: false` (`approve.go:155`) | **OL** (draft lifecycle is genuinely load-bearing: approve, list, batch approve all read it) but the *write-time deletion* is **HA** — it makes a write non-byte-faithful for zero semantic gain (absence and `true` are the same state) |

### C. Body and frontmatter fidelity (the input=output gap)

| # | What akb forces | Where | Consumer(s) | Verdict |
|---|---|---|---|---|
| M9 | Any write that is not "new page with `created`+`updated`" re-serializes the **whole** frontmatter from a `map[string]any` | `write.go:484-495`, `write.go:242-253`, `write.go:345-352`, `append.go:186-194`, `approve.go:157-170` | Nothing downstream needs the rewrite; it exists only to carry the injected fields | **HA** — the single largest openness/faithfulness defect. A document-level round-trip (or appending the injected keys textually) keeps comments, order, quoting, indentation |
| M10 | Page path is **not** the caller's path: it is forced under the type's `dir` | `write.go:418-424` (strip), `write.go:433-441` (re-prefix); only escaped when `tmpl.Dir == ""` | `relPath` → search index, link graph, index.md, commit message. `resolveExistingPage` (`resolve_page.go:89-127`) can *find* a page outside the type dir but `write` never *puts* one there | **OL** (dir-per-type is a real organizing convention) but the asymmetry is an accident: `append` accepts either location, `write` accepts one. Consequence: `akb write archive/foo.md` with `type: note` and `dir: notes` writes `kb/notes/archive/foo.md` — the trailing sub-path **survives**, so the caller cannot choose the directory and can still bury the page arbitrarily deep |
| M11 | `--frontmatter type=X` changes the type **without relocating the file** | `write.go:210-227` (assignment), `write.go:176-189` (path already resolved from the old location) | Retype leaves the page in the old type dir → type→dir derivation is immediately violated for every later consumer; `dirFromType` is not re-derived | **HA** — a direct contradiction inside the mandate set |
| M12 | A plain write to a page that already exists **elsewhere** silently creates a second copy (no cross-location existence check) | `write.go:455-460` (`os.Stat(fullPath)` only); `resolveExistingPage` is never consulted on the plain write path | Two pages, one identity → duplicate `documents.path` rows, duplicate index.md entries, orphaned links | **HA** consequence of M10 |
| M13 | Append always joins with a bare `"\n"` | `write.go:342`, `append.go:183` | Markdown rendering only | **HA** (and source of the `--dated` blank-line mismatch, §0) |
| M14 | Body mutation at approve: annotations and provenance markers are stripped | `approve.go:152-153` (`markdown.StripAnnotations`, `StripProvenanceMarkers`) | `provenance` and `citations` lint checkers read the markers (`internals/lint/*`); once stripped they are gone permanently. `approve_test.go:119,148,242-262` pins both the stripping and the "draft is indexed verbatim before approve" behaviour | **OL** (publish = clean is a coherent policy) but it is a *body* rewrite: `${input} ≠ output` on a command whose stated job is a one-field flip |

### D. Validation gates and their holes

| # | What akb forces | Where | Consumer(s) | Verdict |
|---|---|---|---|---|
| M15 | Schema `required: true` presence gate runs **before** CEL, per write path | `write.go:515-518`, `append.go:217-220`, `approve.go:224-229` via `required_fields.go:19-36`; mirrored in lint (`lint/required_fields.go:70-86`) | The refusal message is duplicated in two packages (`required_fields.go:33-36` and `lint/required_fields.go:70-73`) — a wording freeze across the seam | **OL** (it is the workaround for CEL's vacuous truth on absent keys — `write.go:511-514`). With typed schema validation it becomes the schema validator's job, and the *reason* for a separate pre-CEL pass disappears |
| M16 | **approve never re-runs CEL.** Only presence is checked | `approve.go:144-156` (gate at 148, then straight to `is_draft: false`) | A page that cannot be written cannot be approved into compliance, and vice versa: a page hand-edited to violate a validation still approves | **HA** — known hole; the write-time and lint-time rules (and `old_page`, available here) are simply absent |
| M17 | approve does not bump `updated` even though it rewrites the body and frontmatter | `approve.go:155-170` | `updated`-based CEL lint rules (`note.yaml:37`, `adr.yaml:183,190`) keep firing on freshly-approved pages | **HA** |
| M18 | Every mutating command must open the search DB first | `write.go:111-118`, `append.go:88-95`, `approve.go:61-68`, `delete.go:61-68` | A KB without `.agent-kb/search.db` cannot be written to at all ("run `akb index rebuild`") | **OL** — content write is coupled to an index artifact |

### E. Path and encoding prescriptions

| # | What akb forces | Where | Verdict |
|---|---|---|---|
| M19 | Filename must end `.md` | `write.go:150,274,399` (three copies), `append.go` inherits via resolve | **OL** — every walker filters `.md` (`sqlite.go:257`, `index.go:195`, `lint.go:101`, `approve.go:299`), so widening it is a cross-cutting change |
| M20 | `kb/` prefix is stripped, `raw/` prefix is rejected, both with bespoke error text | `write.go:153-159,277-283,402-408`; `path.go:104-118`; `delete.go:135-137`; `append.go:102-107` | **OL** for the guard; **HA** for the five-times-duplicated string prefix handling |
| M21 | `..` is rejected **anywhere** in the input path, including inside a filename | `path.go:99-101` (`strings.Contains(inputPath, "..")`) | **HA** — `my..draft.md` and `v1.2..3.md` are legitimate names refused for the wrong reason; the correct rule (already present) is segment-wise (`filepath.Clean` + containment at `path.go:118-124`) |
| M22 | Pages live under `kb/`, raw under `raw/`; no alternative root | `path.go:118`, `ResolveRawPath` | **OL** |
| M23 | No encoding check anywhere: no BOM tolerance, no UTF-8 validation, no line-ending normalisation, no size limit | `frontmatter.go:92-120` (`extractBody` requires the *first* line to start with `---`, so a UTF-8 BOM makes the file unparsable); `\r` is trimmed per line (`frontmatter.go:97,111`) | **HA** — the failure mode is a confusing "no frontmatter found" on a BOM'd file |
| M24 | Symlink containment for page paths, type-dir candidates, and every template file | `path.AssertContained` (`resolve_page.go:129-137`), `assertTemplateFilesContained` (`template.go:292-312`) | **SR** — security, not policy |

### F. Managed files (`kb/index.md`, `kb/log.md`)

| # | What akb forces | Where | Consumer(s) | Verdict |
|---|---|---|---|---|
| M25 | `index.md`/`log.md` are write-protected per command | `write.go:161-167, 285-291, 410-416`; `append.go:109-115`; `delete.go:155-160`; `index.go:114-117, 228-231`; orphans path filters them (`delete.go:79-83`); walkers skip them (`list.go:96`, `sqlite.go:261-263`, `index.go:199-202`, `lint.go:104-106`, `approve.go:307`) | The exclusion set is replicated in **eight** places. `status.go:85-89` excludes them only when at the `kb/` top level, unlike the other walkers | **OL** — the protection is load-bearing (index.md is regenerated wholesale by `index.RebuildIndex` `index.go:243`); log.md is parsed by a strict heading regex `log.go:26`). But the guards are **per-command copies**, not one storage-layer rule, and `git.go` has one canonical list (`managedCommitPaths`, `git.go:27`) that is *restaging*, not protection. A single `path.ManagedPage()` predicate is the obvious consolidation |
| M26 | `delete` auto-maintains index.md + log.md and restages them into the commit; `write`/`append`/`approve` do **not** | delete: `delete.go:224-241`; write: `write.go:572-574` prints "Don't forget to update the index!" | `index_consistency` lint (`lint/index_consistency.go:73-84`) flags every page written since the last `akb index rebuild`/`akb index add` — the warning is the *expected* state after a normal write | **HA** — the asymmetry is the clearest sign the current index concept is a leftover, not a design |
| M27 | `akb index add` takes the summary from **argv**, not from the page's frontmatter `summary` | `cmd/akb/index.go:108-109, 192-201` vs `index.RebuildIndex` reading `fm.Fields["summary"]` (`index.go:221`) | Two sources of truth for the same rendered field; a rebuild silently rewrites what `index add` produced | **HA** |

### G. Storage-layer prescriptions

| # | What akb forces | Where | Verdict |
|---|---|---|---|
| M28 | Every write is auto-committed with a `akb: <verb> <path>` message | `git.go:92,131`; `write.go:57-63` (the post-write state message); `append.go:226`; `approve.go:172`; `delete.go:244`; 9 more verbs (`raw_*`, `log`, `index`, `template*`, `init`) | **OL/Q**. Nothing in akb parses commit messages (grep: only producers), so the *format* is unconstrained; the *auto-commit* is the opinion (`--no-commit` is the global escape). Verdict: opinionated-but-load-bearing for the "git-tracked KB" contract |
| M29 | Commit identity `-c user.name=akb -c user.email=akb@local` when the repo configures none; repo config never written except by `init` | `git.go:32-33`, `git.go:503-511` (already documented in `seam-periphery.md:101-111`) | **OL** |
| M30 | A merge conflict blocks every write before validation | `git.go:178-193` | **SR** for a git-tracked KB |
| M31 | `delete --orphans` deletes anything in the link graph's orphan set, silently skipping failures | `delete.go:72-133` | **OL** — the orphan predicate is the link graph's, not the user's |
| M32 | Read path is byte-verbatim but path-constrained: `read` strips nothing from the file (no frontmatter, no JSON) and reads through `FilesystemProvider` (no git) | `read.go:52-65`, `filesystem.go:37-44` | **HA wrt openness**: `akb read` is the only place where input==output already holds, and it deliberately does *no* normalisation — the asymmetry with `write` is the audit's headline |

---

## Consumer trace (the core deliverable)

Every downstream reader of the four auto-managed fields and of `type`/`title`. "No consumer" is stated only after an exhaustive grep of non-test code.

| Producer | Written by | Downstream readers **inside akb** | Extracted consumers |
|---|---|---|---|
| `type` | stdin, `--frontmatter type=`, `approve`/`append` re-marshal | (1) template lookup on all four write paths; (2) `documents.type` column (`sqlite.go:60`) → `akb search --type` (`sqlite.go:173`); (3) index.md grouping + `## Types` heading (`index.go:261-274`, `typeToHeading` `index.go:31-37`); (4) lint `required_fields`/`cel`/`type_orphan`; (5) `cmd/akb/index.go:176` (`fm.Type == tmpl.Name` guard while resolving a bare filename); (6) `approve.go:219` template lookup; (7) `--frontmatter` unknown-type re-check (`write.go:239`) | (8) the write path's own dir derivation (M10) |
| `title` | stdin, `--frontmatter title=` | (1) FTS `title` column, BM25 weight 10.0 (`sqlite.go:186`); (2) `pages.title` (`sqlite.go:75`); (3) index.md entry text (`index.go:229,274`); (4) log entry title on delete (`delete.go:127,179,235`); (5) `akb index add` entry title (`index.go:198`); (6) `akb template write`'s mockup gate (`templates_write.go:417`) | (7) commit messages do **not** carry it |
| `created` | `write.go:377-379`; `--frontmatter created=` | **NONE.** `search --after` filters `documents.created` = `datetime('now')` at index time (`sqlite.go:60,177`), reset by every `akb index rebuild` (`sqlite.go:227-231` deletes all rows). The only readers are user CEL rules (`embedded/adr.yaml:136-151`) | `internal/config.Config.Created` (`config.go:20`) is a **different** `created` (the KB's own creation date) — same name, unrelated |
| `updated` | `write.go:231,334-335,478-481`; `append.go:176`; **not** approve | **NONE inside akb.** Only user CEL / lint rules (`embedded/note.yaml:37`, `adr.yaml:183,190`) | — |
| `is_draft` | absence = draft; `true` deleted on write (`write.go:382-385`); `false` set by approve (`approve.go:155`) | (1) `approve` gate + batch candidate collection (`approve.go:144, 324`); (2) `list` `[DRAFT]` and `--json.is_draft` (`list.go:24,64,110`). **Not** in the search index, **not** in index.md, **not** in any lint checker | — |
| type→`dir` derivation | `tmpl.Dir` (`template.go:43-47`); `declaredTypeDirs` (`resolve_page.go:62-70`) | (1) write-path path construction (M10); (2) `resolveExistingPage` bare-filename addressing (`resolve_page.go:104-118`) used by `append`, `--append`, `--frontmatter`; (3) `akb index add` bare-filename fallback (`index.go:161-181`) | — |

**Read-out:** `created` and `updated` have *zero* akb consumers. `type`/`title` are heavily consumed (search, index, lint, dir derivation) — they cannot simply be dropped, but nothing forces them to be *tool*-managed. `is_draft` has exactly two consumers, both lifecycle commands.

---

## Blast-radius / verdicts

**Structurally required (keep, no debate):** symlink containment (M24), merge-conflict block (M30), the `.md`-filtered walkers (M19, until a page format change), the managed-file *protection* (M25) as a concept.

**Historical accidents — candidates for removal or de-mandating, with blast radius:**

1. **M6/M7 (`created`/`updated` injection) — no consumers.** Removing the injection changes: nothing in search, index, lint, links or list. It only changes what CEL sees, and only for templates that declare those fields — i.e. *exactly the templates that can now declare them as required/validated themselves*. Blast radius: `embedded/adr.yaml` + `embedded/note.yaml` rules would become vacuous (`!has(...)` guards already present at `adr.yaml:137,183`), and 6 tests (`write_test.go:672-800`, `1273`) would fail. Verdict: **HA, highest-value removal candidate**, gated on moving "dates are RFC3339" from an akb convention into a schema-declared type.
2. **M1/M2/M4 (hardcoded `type`/`title` gates) — duplicated by the schema.** Blast radius of deleting the hardcoded checks: `type`/`title` stay enforced *only if* the template says `required: true`; both embedded templates and `templates_write.go:390-397` (mockup gate) already assume that. The real cost is `frontmatter.go`'s structural routing of `type`/`title` out of `Fields` (`frontmatter.go:70-82`) — CEL cannot `has("type")` today, so the schema gate must keep special-casing them (`required_fields.go:22-26`). Verdict: **HA** — but removal must be paired with making `type`/`title` ordinary fields in `page.frontmatter`, otherwise the special-casing just moves.
3. **M8 (delete `is_draft: true` on write).** Blast radius of no longer deleting: pages keep an explicit `is_draft: true` line that `IsDraft` already treats identically (`frontmatter.go:151-163`); `list`, `approve`, batch approve unaffected. Zero risk. Verdict: **HA, free win** for input==output.
4. **M9 (whole-frontmatter re-marshal).** Blast radius of a document-level round-trip: the injected-key writes (`created`, `updated`, `is_draft`) become surgical text edits; the sorted-key reordering and the quoted-date churn (`seam-core-pipeline.md:205-213`) disappear; the `--frontmatter` path must merge key=value pairs into the parsed document rather than a map. Tests that assert only field *presence* survive; none assert key order. Verdict: **HA, highest lexical value** — it is the whole "input=output" gap on the write path, and it is *independent* of the JSON/CEL seam.
5. **M10/M11/M12 (forced type dir + retype-without-move + silent duplicate).** Blast radius of relaxing the force: `resolveExistingPage` already resolves both the named path and the type dirs, so append/`--frontmatter` need nothing; search/index/link keys are `relPath`, so a page anywhere under `kb/` is indexed identically; `index.RebuildIndex` re-derives groups from frontmatter, not from directories. Consequence of *keeping* it: the M11 retype hole and M12 duplicate remain. Verdict: **OL for the convention, HA for the enforcement**.
6. **M26/M27 (index maintenance asymmetry).** Blast radius of making writes maintain index.md: the lock already covers write+index+links in one transaction (`write.go:531-575`), and `index.AddEntry`/`RemoveEntry` are already proven by `delete`. Verdict: **HA** — but note the deeper option: `index.md` may be *derivable* (`index.RebuildIndex` proves it), which would make the whole protected-file class disappear.
7. **M21 (`..` anywhere).** Blast radius: nil — `filepath.Clean` + containment already rejects real traversal. Verdict: **HA**.

**Opinionated-but-load-bearing (keep, make configurable at most):**

- M3 (type must name a template) — the template *is* the validity source; relaxing it means "unvalidated pages are legal", which is a legitimate design choice but a *product* decision, not a cleanup.
- M8 draft lifecycle, M14 approve stripping, M15 required-field pre-gate, M28/M29 auto-commit + identity, M31 orphans, M18 DB-required-to-write, M25 managed-file protection.

---

## Feasibility sketches (a few lines each; no implementation designed)

**(a) Verbatim frontmatter when nothing must change.** In the plain-write branch, when `statErr != nil` (new page) *or* the caller supplied `created`+`updated`, write `stdinContent`; otherwise perform a *line-level* insertion of only the injected keys instead of `yaml.Marshal(allFields)`. ~30 lines, local to `write.go:374-495`, no consumer changes.

**(b) Drop the `created`/`updated` injection entirely.** Delete `write.go:231, 334-335, 377-379, 478-481` and `append.go:176`. A template that wants a timestamp declares `created: {type: string, format: date-time, required: true}` and a CEL rule `!has(old_page.frontmatter.created) || old_page.frontmatter.created == page.frontmatter.created`. Nothing else changes because nothing else reads them.

**(c) One managed-file predicate.** Move the eight copies (M25) behind `path.ManagedPage(rel string) bool` (and reuse `storage.managedCommitPaths` as the single list); `status.go:85-89` is the only site whose current behaviour differs.

**(d) Dir force as a template flag.** Interpret `dir` as a *default* rather than a *destination*: if the caller's `cleanPath` already names a directory outside `dir`, honour it (the containment check at `path.go:118-124` is unaffected). Fixes M10/M11/M12 without touching `resolveExistingPage`.

**(e) `--frontmatter` value typing.** `fm.Fields[key] = value` → parse the value as YAML when it parses to a scalar/list/map, else keep the string; or drop `--frontmatter` in favour of the raw-markdown/JSON stdin path already sketched in `seam-core-pipeline.md:150-165`. Today a `tags` list becomes the string `"a b"` and any list-shaped CEL rule goes unevaluable → the write is refused (`write.go:604-616`).

---

## Open questions for synthesis

1. **Which package owns "a valid entry"?** M3 plus M1/M2/M4 want the template to be the only gate, but `frontmatter.Parse` structurally owns `type`/`title` (routing them out of `Fields`) and `required_fields.go` exists in *two* packages with identical wording. Should `type`/`title` become ordinary frontmatter keys so CEL/JSON-Schema can see them, and `ParsedFrontmatter.Type/Title` be derived accessors only?
2. **Is `created`/`updated` a KB concern or a template concern?** The evidence (zero consumers, all readers are user rules) says template. If confirmed, the "input=output" gap collapses to M9 alone.
3. **Is the type→dir mapping a mandate or a default?** M11/M12 show the current form cannot be honoured consistently; deciding "default" also decides whether `resolveExistingPage` keeps its two-candidate search.
4. **Does `approve` remain a validator at all?** Today it is a body-rewriting, CEL-skipping, `updated`-preserving command with a presence-only gate. Its three distinct jobs (flip `is_draft`, strip markers, gate on required fields) are independently optional.
5. **Is index.md a page or a derived artifact?** `index.RebuildIndex` can regenerate it from frontmatter, and `index add` takes a summary from argv — the protected-file class (M25) and M26/M27 both dissolve if it is derived.
6. **Does the write path need the search DB and a git repo to accept content?** M18/M28 make "akb provides the hands" conditional on two side-effect systems. `read.go` needs neither — the asymmetry is worth a product decision.

---

## Residual risks / caveats

- Line numbers are from the working tree at the time of this read; `write.go` is 653 lines with three near-duplicate branches (frontmatter / append / plain), so several prescriptions appear 2–3× and a single-line citation may under-count call sites.
- The claim "no consumer of `created`/`updated`" is based on exhaustive `grep` of `cmd/` and `internal/` for `"created"`, `"updated"`, `.Created`, `.Updated` (excluding tests). I did not audit the embedded skill docs (`internal/skill/embedded/*`) for *prose* consumers that tell agents to read those fields.
- Blast radii are stated for in-repo code only; `test/testdata/*.txt` testscript integration tests and the embedded skill instructions were not enumerated.
- M9's "document-level round-trip" sketch is my own inference of the fix shape; it is explicitly *not* a design (hard rule: no implementation design).
