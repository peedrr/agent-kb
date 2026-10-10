---
grammar: 1
type: research
title: "RFC-002-S6 prior-research extraction — ratified decisions and open gaps"
status: final
provenance: agent-drafted
run: rfc-002-s6
as_of: "2026-10-09"
created: "2026-10-09"
updated: "2026-10-09"
scope:
- docs/knowledge-base/kb/**
tags:
- rfc-002
- templatev3
- prior-research
summary: "Extraction of everything prior KB research already ratified for RFC-002-S6 (ADR-006/008/010 invariants, S1 SPEC conventions), with the open gaps that drove the run's decision forks."
informs:
- SPEC-003
---

# RFC-002-S6 (→ SPEC-003) — Prior-Research Extraction

Read-only extraction. Repo `agent-kb/agent-kb` @ HEAD `5a2ea67` (`main`); ratification-period
anchors cited by the KB were taken at `a316f0a` (v0.23.1), so line numbers drift — treat all
`file:line` as approximate. All KB anchors are paths under `docs/knowledge-base/kb/`.

S6 identity: `RFC-002-open-validation-engine.md:833` —
`| RFC-002-S6 | SPEC | 1 | TemplateV3 core format + loader: minimal key set (ADR-010), strict keys, type/title un-routing (D10), embedded template conversion — drafts as SPEC-003 | proposed |`.
So S6 = loader/format only; the schema **pass** (compilation, pipeline placement, sweep checker)
is S7/SPEC-004 (`:834`), and `template write` schema obligations + `FAILS:` parser are S8/SPEC-005 (`:835`).

---

## 1. DECIDED

### 1.1 The exact minimal key set (ADR-010)
- **I1** (`decisions/ADR-010-templatev3-minimal-key-set.md` §Invariants I1): the Phase-1 template
  loader MUST recognize **exactly** the top-level keys `name`, `description`, `dir`, `schema`,
  `validations`, `lint_rules`.
- The deferred (refuse/warn) keys are named in **N1** (`:51`): `path`, `state_field`, `functions`,
  `projections`, `title_field`, `summary_field`, `strip_on_approve`.
- `Template` struct already carries exactly these six YAML keys today:
  `internal/template/template.go:43-51` (`name; description; dir; schema; validations; lint_rules`),
  with `Schema.Frontmatter map[string]FieldSchema` (`:15-17`), `FieldSchema{Type,Required,Enum}`
  (`:20-24`), `ValidationRule` (`:27-32`), `LintRule` (`:35-39`) — per
  `research/rfc-002-phase1-scope-code-surface.md` §Q4 and `research/rfc-002-a2-recon-code.md` §1.

### 1.2 Strict-key posture
- **I2**: `akb template write` encountering any other top-level key MUST refuse with **exit 1**
  and a message naming the unrecognized key and the I1 supported set.
- **I3 / N2**: the **load path** encountering an unrecognized top-level key in an otherwise-V3
  template MUST **warn on stderr naming the key** and continue loading the recognized keys —
  **never hard-fail** (forward-compat: a newer-akb template must not brick an older binary).
- **N1**: recognize-then-ignore of deferred keys is forbidden on any Phase-1 path (that is the D9
  "inert declaration" trap).
- **I5**: when a file carries the retired `schema.frontmatter` shape, **ADR-006's refusal applies
  before any unknown-key judgment** (ADR-006 I5 ordering). ADR-006 I1/I2/I3 require that refusal to
  come from **one shared helper with one message text** at both load and `template write` — the two
  divergent `detectOldFormat` copies are named the anti-pattern (`ADR-006` §Verification I3/I4;
  `cmd/akb/templates_write.go:461-469`, `internal/template/template.go:73-78`).
- Verification gates (ADR-010 §Verification): testscript matrix refusing each deferred key + a typo,
  each naming the key and supported set, exit 1; a V3 template with an unknown key still loads (a
  page write against it succeeds) with a stderr warning; `rg -n '^path:' internal/template/embedded docs/knowledge-base/.agent-kb/templates` returns zero hits.

### 1.3 `schema` key semantics
- **I6**: the `schema` value MUST be a **JSON Schema 2020-12 document over the current document
  shape** (ADR-010 I6 ↔ ADR-008 I1/I2).
- Current shape ADR-008 I1 fixes: `frontmatter` (raw YAML values), `content` = `{raw, word_count,
  char_count}`, `ast` = `{headings, links, code_blocks}`, `akb` = `{provenance_markers,
  annotations}`, `file` = `{path, name, dir}`. Concrete property paths are `content.raw`,
  `ast.headings`, `/frontmatter/title` (ADR-008 I2; `code-surface.md` §Q1).
- The flatten (`content` string, `derived.*` renames, `word_count()` registry, `char_count` scrap)
  is **not** S6 — it rides a later cohort (ADR-008 I3/N2/N4; RFC-002 §5.3 as amended).

### 1.4 `dir` stays; `path` is Phase 3
- **I4**: `dir` keeps current semantics throughout Phase 1. **N3**: Phase-1 work MUST NOT rename
  `dir` → `path` in loader, embedded templates, or this repo's KB templates — the rename + D11
  resolution-hint semantics ride Phase 3 (RFC-002 §8.3, D11 register `:883`).

### 1.5 D10 `type`/`title` un-routing — concrete meaning
- RFC-002 §6 (`rfcs/RFC-002-open-validation-engine.md:289-295`): `type`/`title` **stop being engine
  gates** (D10 register `:882`). `type` survives only as the **template selector** (unknown template
  refused at write = the one surviving door-bouncer; flagged by `type_orphan` at sweep).
  `frontmatter.Parse` **stops routing `type`/`title` out of `Fields`**; they become ordinary keys
  visible to both validators.
- Code blast radius (decided work, anchored `research/rfc-002-a2-recon-code.md` §4):
  - `internal/frontmatter/frontmatter.go:61-72` — the routing switch (`type`→`fm.Type`,
    `title`→`fm.Title`, else→`Fields`) is removed; `type`/`title` land in `Fields`.
  - `internal/cel/pagebuilder.go:60-66` — the `fmMap["type"]=fm.Type; fmMap["title"]=fm.Title`
    injection becomes redundant (must avoid double-write).
  - Delete the `type`/`title` special-cases in `cmd/akb/required_fields.go:38-49` and
    `cmd/akb/templates_write.go:399-401,411-447`.
- `title` becomes the **declared display field** (D18) feeding search title column / `index.md` /
  log / CEL page map; declared via `title_field:` (default `title`) — **but `title_field` is a
  deferred key in Phase 1** (ADR-010 N1), so the declaration machinery rides Phase 3/A13 while the
  default-`title` behavior is the Phase-1 surface (RFC §6 `:293-298`). This split is under-specified
  (see §2).
- D10 disposition source: `research/json-cel-log.md:450`; `research/synthesis-openness.md:111`.

### 1.6 Embedded template conversion (S6 scope)
- S6 scope line explicitly includes "embedded template conversion" (`:833`). ADR-006 records the
  embedded adr/note rewrite as **Phase-1 implementation scope outside ADR-006**
  (`ADR-006` §Consequences, "Neutral…" line `:108`). `rfc-002-a2-log.md:179` parks it as
  "Embedded adr/note rewrite as V3 showcases — Phase-1 implementation".
- Inventory (`research/rfc-002-a2-recon-code.md` §5/§7; verified on disk): embedded
  `internal/template/embedded/{adr.yaml, adr_pass.md, adr_fail.md, note.yaml, note_pass.md,
  note_fail.md}` — the two YAMLs plus their 4 mockups convert; this repo's KB corpus
  `docs/knowledge-base/.agent-kb/templates/{adr,rfc,spec,research}.yaml` + 8 mockups is the **dogfood**
  corpus.
- Guard invariant both the conversion and the schema pass must honor (RFC §6 `:286-288`): *the
  schema `required` set under `properties.frontmatter` is exactly the set of keys CEL rules may
  read unguarded.* The presence job of the retired `schema.frontmatter` moves to JSON Schema
  `required`.
- Every shipped template calls `timestamp()` on `created`/`updated` currently declared plain
  `type: string`; under ADR-003's assert-all policy the corpus mixes shapes (`created` date-only,
  `updated` date-time), so conversion must declare `format: date` vs `format: date-time`
  deliberately — blind `date-time` breaks existing-page sweeps (`rfc-002-a2-log.md` §Key findings;
  ADR-006 §Context). This is an author judgment, not an auto-rule (ADR-006 N3).
- Embedded adr.yaml's derived-surface rules to re-express against a current-shape schema:
  `ast.headings` ×5, `ast.code_blocks`, `ast.links` ×2, `page.content.word_count >= 50`
  (`code-surface.md` §Q5a).

### 1.7 Loader facts S6 builds on (ratified context)
- Loader is **lenient today**: `yaml.Unmarshal(data, &raw)` then `&tmpl` with no
  `DisallowUnknownField`/`Strict` option (`internal/template/template.go:106,114`); `detectOldFormat`
  runs off the raw map in between — `code-surface.md` §Q4, `a2-recon-code.md` §1.
- `template get --full` serializes `tmpl.Schema` verbatim (`cmd/akb/template.go:174`) — the V3
  `schema:` block must flow through that path (`a2-recon-code.md` §7 note 5).
- Exit contract: plain errors → exit 1 with `Error:` prefix; `usageError`/`internalError` → exit 2
  (`cmd/akb/main.go` `classifyExit`; `AGENTS.md` §NOTES). ADR-010 I2 pins the unknown-key refusal at
  exit 1.

*Confidence: high for 1.1–1.4, 1.6 (ratified ADR text read verbatim). High for 1.5/1.7 mechanism
(anchored recon), medium for the exact Phase-1 `title` display behavior (deferred-key inference).*

---

## 2. OPEN / GAPS

These are **not** settled by prior research and drive S6 gap-fill.

1. **Embedded-conversion ownership vs RFC §17.** RFC §17 places "embedded example restructure" in
   **Phase 4** (`:781`), while S6/ADR-006/a2-log place the adr/note V3 rewrite in **Phase 1**. The
   exact Phase-1 deliverable (format-only conversion vs showcase restructure) is unresolved —
   `research/rfc-002-a2-extract-prior-research.md:336-338` lists this as an explicit open item.
2. **`title` display behavior in Phase 1 without `title_field`.** RFC §6 declares `title_field:`
   (default `title`) and fallbacks, but ADR-010 N1 defers `title_field`. What S6 ships for the search
   title column / `index.md` / log (literal `title`? empty-column fallback? filename?) is undecided.
3. **`schema` block's Go representation in the Template struct.** Is `Template.Schema` a
   `json.RawMessage` / `map[string]any` / compiled `*jsonschema.Schema`? Where does santhosh v6
   compilation live — loader (S6) or schema pass (S7)? ADR-010 I6 fixes the *format*, not the field
   type or the compile site; S6/S7 boundary text (`:833-834`) does not draw it.
4. **Unknown-key message text and single-helper design for V3.** ADR-010 requires a "self-sufficient
   message naming the offending key and the supported set" but fixes no wording; whether V3 unfies
   with ADR-006's shared `detectOldFormat` helper (one rejection surface) or adds a sibling is open.
5. **KB-template (dogfood) conversion timing.** ADR-008 I5 speaks of re-pathing "the two embedded
   templates and this repo's four KB templates" at flatten time; ADR-006 says the KB adr/rfc/spec
   migrate "via the recipe as its dogfood". Whether S6 converts the 4 KB templates to V3 now or
   leaves them to the recipe is not pinned.
6. **`type`/`title` un-routing downstream scope (S6 vs later).** Removal of `frontmatter.Parse`
   routing forces changes in `search` type/title columns, the write-path re-marshal that re-merges
   `type`/`title` (`cmd/akb/write.go:490-497`), `list --json`, and `index.md`/log title reads. Which
   of these S6 owns vs S7/other phases is unspecified (`a2-log.md` D-c: "V3 implementation").
7. **`template get`/`--example`/`--full` behavior under V3.** `--full` serialization, and
   `--example` stale-mockup validation, must be re-derived for the schema block
   (`a2-recon-code.md` §7.5); not decided.
8. **Test-fallout enumeration.** RFC §19 ("Known unknowns") requires *each phase's tech-spec* to
   enumerate its test fallout — S6 must produce this list (embedded conversion + un-routing +
   strictness touch the 5 YAMLs, 8 mockups, `cel_old_format.txt`, `template_write_*`, required-field
   wording tests). Not yet enumerated for S6.
9. **Repo/AGENTS doctrine updates.** RFC §17 `:782-786` says every phase corrects stale doctrine
   (AGENTS.md asserts "TemplateV2", engine `type`/`title` gates). Which doctrine lines S6 edits is
   undecided.
10. **`schema` key collision mechanics on the load path.** ADR-006 I5 handles the mixed retired/V3
    shape; the interplay between the `schema.frontmatter` predicate and the new strict-key set for an
    otherwise-V3 template carrying a stray `frontmatter` key (ADR-006 §Context accepts the
    false-positive) is acknowledged but the exact precedence/message for S6 is not spelled out.

*Confidence: high that these are genuinely open (each traced to a parked/deferred note or a
deferred-key ADR line); the resolution direction is not pre-determined.*

---

## 3. CONVENTIONS (from the SPEC-002 / RFC-002-S1 run — the template S6 must follow)

Mandatory constraints for SPEC-003 (drafted from RFC-002-S6), from
`research/rfc-002-s1-adr-conventions.md` (esp. §5–§8) and `research/rfc-002-s1-log.md`.

- **Format authority**: docs-writer `references/spec.md` + the KB `spec` template
  (`docs/knowledge-base/.agent-kb/templates/spec.yaml`), which is a live loader-enforced contract
  (`s1-adr-conventions.md` §1, §6, §8).
- **Frontmatter (spec.yaml-enforced)**: `grammar: 1`, `type: spec`, `id` `^SPEC-[0-9]{3}$`,
  `title`, `kind ∈ {change,capability}`, `status ∈ {draft,proposed,active,completed,superseded,
  withdrawn}`, `provenance`, `created`, `updated`, `scope` (globs), `verified` (map). Next free global
  number is **SPEC-003** (SPEC-001, SPEC-002 taken). Draft as `status: proposed`, `provenance:
  agent-drafted`, `verified.state: unverified`, `method: manual-review`; anchors `state: unverified`
  pending a verification sweep (`s1-log.md` §Coordinator decisions; SPEC-002 frontmatter).
- **Challenge affordance** must be verbatim: body must contain `this SPEC is authority only while`
  (`spec.yaml` validation `challenge_affordance`; SPEC-001/002 `> **For agents:**` blockquote).
- **Body skeleton (spec.yaml H2 gates)**: `## Why`, `## Goals and Non-Goals`, (`## Current
  Behaviour` iff `kind: capability`), `## Requirements`, `## Acceptance Criteria`, `## Contract and
  Invariants`, `## Preserved Behaviour`, `## Decisions and Rejected Alternatives`, `## Assumptions
  and Open Questions`, `## Affected Surface and Ordering`, `## Verification Plan`, `## Drift Ledger`,
  `## Revisit Triggers`, `## Evidence Appendix` (SPEC-001/002 shapes; `s1-adr-conventions.md` §6).
- **IDs**: `REQ-NNN` in EARS-2119 grammar; `AC-NNN (verifies REQ-NNN)`; body must contain both
  patterns; no `[NEEDS CLARIFICATION`; every fenced block needs a language tag; ≥50 words;
  a `## Decisions and Rejected Alternatives` section must contain "Do not re-propose unless"
  (`spec.yaml` validations).
- **`depends_on`** is declared as data at the end of `## Affected Surface and Ordering`
  (SPEC-002 `:294-297` shows the style).
- **Origin convention** (RFC §18 lifecycle, `s1-adr-conventions.md` §5): the intro paragraph
  immediately after the `For agents` blockquote reads
  `Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-S6 …`, and `tags` includes
  `- rfc-002`. (Note: `spec.yaml` declares no `tags` key and SPEC-002 carries none; the interim
  tolerance hinges on the loader, so if S6 tightens the loader this convention must be re-checked —
  but strictness is scoped to **template** keys, not page frontmatter, so extra page keys remain
  governed by the schema block, not the key set.)
- **Wikilinks**: `[[ADR-003-santhosh-v6-format-assertion|ADR-003]]`, `[[RFC-002-open-validation-engine|RFC-002]]`
  (target-only or `target|label`).
- **index.md / log.md**: index entry = title link + em-dash one-line description under `## Specs`;
  log entry `## <YYYY-MM-DD> <verb>`. Never write either file directly — use `akb index add` /
  `akb log` (`AGENTS.md` §ANTI-PATTERNS; `s1-log.md` §Key findings).
- **Commit strategy** (owner-raised in S1): draft in the run dir; land via `akb write --no-commit`
  (+ index/log `--no-commit`); owner reviews in place; owner runs one plain-git commit bundling all
  KB changes (validation failures never commit — fail closed). No `akb commit` command exists
  (`s1-log.md` §Decisions made).
- **Ratification lifecycle** (RFC §18): the S6 candidate row moves `proposed → ratified` with
  `Resolves-as SPEC-003` and body references rewritten — a **human** step at ratification; the page
  stays an orphan until RFC-002 links it (`s1-log.md` §Parked; RFC §18 `:786-800`).
- **Evidence standards** (ADR run conventions, `s1-adr-conventions.md` §5, and
  `research/rfc-002-s1-code-recon.md`): anchors are `{claim, path, symbol, kind}` + `verify.check`
  with `method: check|ask`, no line numbers in anchors; **never pin cel-go-owned error strings**
  (ADR-004 `:108`); refusal/warning messages must be single-sourced and self-sufficient (ADR-006
  I3/I4, ADR-005 I10/I11).
- **Exit-code contract** (AGENTS.md §NOTES): 0 success; 1 result-to-act-on (validation/refusal);
  2 could-not-do-work (`usage:` / `internal:` prefixes). Refusals must be self-sufficient.

*Confidence: high — these are explicit in the S1 artifacts and the spec template.*

---

## 4. CONFLICTS

1. **Embedded restructure phase.** RFC-002 §17 `:781` puts "embedded example restructure" in
   **Phase 4**; ADR-006 §Consequences `:108`, `rfc-002-a2-log.md:179`, and the S6 scope line
   (`:833`) put the **embedded adr/note V3 conversion in Phase 1**. Resolution is open (see §2.1);
   the S6 scope line is the most recent authority, but §17 is unamended on this point.
2. **RFC §6 schema example vs ADR-008.** The §6 YAML example (`:253-282`) authors the schema over
   the **flattened** shape (`content: {type: string}`, `derived.ast.headings`). ADR-008 I2/N1
   (ratified) bounds Phase 1 to the **current** shape and forbids `derived.*` paths. ADR-008 wins;
   the RFC §6 example must not be copied literally into SPEC-003. (RFC §6's own Phase-1 bullet
   `:327-331` already defers to ADR-010's key set.)
3. **`dir` rename.** RFC §6 `:326` says "`dir` is renamed/re-specified as `path`"; ADR-010 N3
   forbids the Phase-1 rename. ADR-010 (ratified, later) supersedes the RFC's unqualified statement.
4. **Anchor/line drift.** ADR-010 and the phase-1-scope recons cite `template.go:106,114` / `:43-51`
   / `:73-78` and `templates_write.go:461-469` at `a316f0a`; `a2-recon-code.md` cites
   `template.go:105-116` / `:14-55` / `:84-131`. HEAD is now `5a2ea67`. Re-verify anchors at
   drafting time; do not trust exact line numbers.
5. **KB-template count.** ADR-008 I5 says "this repo's **four** KB templates"; `a2-recon-code.md`
   (2026-10-06) enumerated **three** (`adr`, `rfc`, `spec`). On disk today there are **four**
   (`adr`, `rfc`, `spec`, `research`) — evolution, not a decision conflict, but SPEC-003's counts
   must use the current four.
6. **Validator choice (survey vs ADR).** `research/schema-survey.md` describes **pi-subagents'**
   TypeBox validator (draft 3→2020-12, `additionalProperties`, explicit-type transport rule) — that
   is the *bridge* side, not akb. akb's schema validator is **santhosh v6** per ADR-003. Not a
   conflict; do not import TypeBox findings into SPEC-003 except as contrast.
7. **`schema.frontmatter` refusal vs strict-key refusal ordering** is resolved (ADR-006 I5 /
   ADR-010 I5), not a conflict — recorded here so it is not re-litigated.

*Confidence: high — each conflict is a direct textual contradiction between ratified ADRs and the
still-draft RFC, with the ADR/A DR-consistent resolution identified.*

---

## Files Retrieved (highest-value, for the SPEC-003 drafter)
1. `docs/knowledge-base/kb/decisions/ADR-010-templatev3-minimal-key-set.md` (1-142) — THE core record: key set, I1-I6, N1-N3, verification, context.
2. `docs/knowledge-base/kb/decisions/ADR-008-phase1-current-document-shape.md` (1-120) — current-shape bound (I1/I2, N1/N2) the `schema` block validates against.
3. `docs/knowledge-base/kb/decisions/ADR-006-schema-frontmatter-retirement.md` (1-117) — shared-helper old-format refusal (I1-I5, N2), embedded rewrite ownership (`:108`).
4. `docs/knowledge-base/kb/decisions/ADR-009-phase1-sweep-schema-checker.md` (1-95) — S7 boundary; S6 must not absorb the checker.
5. `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` (250-331, 755-800, 860-890) — §6 V3 vocabulary, §17 phases, §18 candidate/lifecycle, D10/D11/D18 register.
6. `docs/knowledge-base/kb/research/rfc-002-phase1-scope-code-surface.md` (1-245) — loader leniency, Template struct, in-tree inventories, rule consumers.
7. `docs/knowledge-base/kb/research/rfc-002-a2-recon-code.md` (1-240) — `frontmatter.Parse` type/title routing + every consumer; embedded inventory.
8. `docs/knowledge-base/kb/research/rfc-002-phase1-scope-log.md` (1-205) — the four ratified scope decisions and evidence index.
9. `docs/knowledge-base/kb/research/rfc-002-a2-log.md` (1-220) — A2 decisions D-a..D-c, parked items (FAILS parser = S8; un-routing = V3 impl).
10. `docs/knowledge-base/kb/research/rfc-002-s1-adr-conventions.md` (1-314) + `rfc-002-s1-log.md` (1-179) — SPEC drafting conventions + S1 process.
11. `docs/knowledge-base/kb/specs/SPEC-002-structured-validation-report.md` (1-56, 290-297) — the closest live SPEC exemplar (frontmatter, `depends_on`, Origin slot).
12. `docs/knowledge-base/.agent-kb/templates/spec.yaml` (1-281) — the loader-enforced contract SPEC-003 must satisfy.
13. `internal/template/template.go:14-137`, `internal/frontmatter/frontmatter.go:18-166`, `internal/cel/pagebuilder.go:26-120` — the code S6 edits.
14. `internal/template/embedded/{adr.yaml,note.yaml}` + `docs/knowledge-base/.agent-kb/templates/{adr,rfc,spec,research}.yaml` — the conversion corpus.

### Note on `rfc-templates-*.md`
The `research/rfc-templates-*` run (report, synthesis, recon-local/-landscape, audit, lanes) is
about the **RFC process document format** (deliberation unit, EARS-2119, 7-status lifecycle) — it
does **not** bear on akb page-template format or the loader. Its only tangential hook is that the
future akb `rfc.yaml` adopter template is deferred until "RFC-002 V3 lands"
(`rfc-templates-report.md:189-197`), i.e. an adopter of S6, not an input to it. No S6 facts were
taken from it. Likewise `research/schema-survey.md` concerns the pi-subagents bridge, not akb.

*Confidence: high that rfc-templates-* is out of S6 scope for format/loader purposes; medium only
on whether the deferred `rfc.yaml` V3 adopter template should be named in SPEC-003's downstream
surface (likely S8/Phase-4 doctrine, not S6).*
