---
as_of: "2026-10-06"
created: "2026-10-06"
grammar: 1
informs:
- ADR-006
is_draft: false
provenance: agent-drafted
run: rfc-002-a2
scope:
- internal/template/**
- cmd/akb/**
status: final
summary: Audit of the openness/coexistence research base and RFC-001-era material on schema.frontmatter retirement, migration mechanics, and deferrals for RFC-002-A2.
tags:
- prior-research
- templates
- migration
- rfc-002
title: Prior-Research Extraction on schema.frontmatter Retirement and Migration (RFC-002-A2)
type: research
updated: "2026-10-06"
---
# Prior-research extraction — RFC-002-A2 (`schema.frontmatter` retirement + migration)

Read-only. Sources mined: json-cel proposal research base, its LOG, RFC-001 + its research
ancestors, and the A1/A3/A4 run residue. RFC-002 itself was **not** read (coordinator-owned);
all RFC-002 quotations below are secondhand via A1's working copy or the A2 run LOG.

Confidence legend: **[H]** = direct anchor quoted; **[M]** = paraphrase of an anchored line;
**[L]** = inference/aggregate.

---

## 1. `.pi/subagents/proposals/json-cel/research/` — the openness/coexistence audit base

### 1.1 Retirement rationale + the "presence job moves to schema required" argument

- **The core retirement statement.** `seam-core-pipeline.md:259` (in the §6 TEMPLATE
  IMPLICATIONS table):
  > `` `schema.frontmatter` (`template.go:19-29`) | **Retire.** `type`/`enum` are dead today
  > (only `.Required` is read: required_fields.go:21, lint/required_fields.go:83,
  > templates_write.go:386). Its remaining job — presence — is `required` inside the JSON
  > Schema block | — `` **[H]**
  - i.e. `FieldSchema.Type`/`Enum` are **declared but never enforced**; the only live
    consumer of the block is `.Required` at three sites (write gate, lint mirror, mockup gate).
- **Why the block was judged decorative.** `.pi/subagents/plans/FIX-SHAPES-2026-09-07.md:73`:
  > "**The schema is decorative.** `schema.frontmatter` (`type`/`required`/`enum`) is parsed
  > but **never enforced** anywhere — the only consumer is `AllowedFields()`, which is dead
  > code… All real enforcement happens in CEL `validations`. Extra undeclared frontmatter
  > fields are silently allowed." **[H]** (file cited `internal/template/template.go:14-44,176-193`)
- **Session-1 confirmation.** `LOG.md:30` (json-cel LOG): "akb `schema.frontmatter`
  type/enum fields are **dead** (only `required` enforced); owner confirms: known, fix in the
  works." **[H]**
- **RFC-001 §6.1 shape carries it.** `docs/.../RFC-001-json-schema-cel-coexistence.md:198-200`:
  "**`schema.frontmatter` (the homegrown block) is retired.** Its one enforced job — presence —
  moves to JSON Schema `required` under `properties.frontmatter`. Templates using it get a hard
  rejection with a migration message (precedent: `detectOldFormat`, template.go:73-82).
  Migration path details → ADR." **[H]**
- **Same key, new meaning (not additive).** `rfc-002-a1-validator/recon-code.md:59-65`:
  current `schema:` key *is* the homegrown block (`Template.Schema` `yaml:"schema"`,
  `template.go:51`); RFC-002 §6 **reuses the same key** for a JSON Schema 2020-12 document, so
  it is "a type/meaning change of an existing key, with a hard-reject migration path. Precedent
  for that rejection already exists: `detectOldFormat`… RFC §6 explicitly cites this precedent
  for the `schema.frontmatter` retirement **[H]**." Also `:172-174` (same point as
  integration-constraint #1). **[H]**
- **A1 LOG restates it.** `rfc-002-a1-validator/LOG.md:105-106`: "`schema:` key collision:
  current `schema.frontmatter` homegrown block — V3 repurposes with hard-reject migration
  (precedent detectOldFormat, template.go:73-82)." **[H]**
- **Synthesis verdict.** `SYNTHESIS-openness.md:48`: RFC-002 "retires `schema.frontmatter` but
  keeps hardcoded `type`/`title` engine gates"; M1/M2/M4 are duplicate gates; `type` survives
  only as the *template selector*. `SYNTHESIS-openness.md:81`/D10 keep `type` as selector. **[H]**
- **Enum-duplication seam** (relevant to what the new `schema:` must absorb):
  `seam-periphery.md:191`: `schema.frontmatter.<f>.enum` is hand-mirrored into a CEL `in [...]`
  rule (`adr.yaml:29-32` vs `:68-72`); moving enum to JSON Schema creates a second source of
  truth "unless the CEL rule is removed or generated — decide explicitly." **[H]**
  Also `seam-core-pipeline.md:78` (rule #4 `valid_status` duplicates dead `.Enum`).
- **`template get` is the current human bridge for the block** —
  `command-inventory.md:145`: writer view emits `schema.frontmatter` + textual `requirements`;
  a JSON Schema "would want a machine-readable counterpart here." **[H]**
- **`command-inventory.md:149`** poses the fork A2 owns: "Does native JSON Schema replace
  `schema.frontmatter` (`FieldSchema{type,required,enum}`, template.go:15-24), or sit beside it
  in the template YAML? The required-fields gate (required_fields.go) is the only current
  structural enforcer." **[H]**

### 1.2 The guard invariant (presence job) — exact statement + consumers

- `seam-core-pipeline.md:263-270`:
  > "**Presence is load-bearing.** The write path refuses a page missing a `schema`-required
  > field *before* any rule runs (write.go:513-518) so CEL may read required keys unguarded
  > (the guard doctrine in `adr.yaml:1-7`). The JSON Schema block must therefore run at the same
  > point and supply the same information. Cleanest: keep one accessor (`checkRequiredFields`,
  > required_fields.go:19-33) but source the set from the JSON Schema's
  > `properties.frontmatter.required`; the sweep mirror (lint/required_fields.go:81-95) must
  > follow, and `templates_write.go:386-395` (`optionalKeysSuppliedByMockup`) too. Write the
  > invariant down: **the required set in the JSON Schema is exactly the set of keys CEL may
  > read unguarded.**" **[H]**
- RFC-001 §6 states the same invariant (`RFC-001…:201-205`) and that the three consumers of
  `checkRequiredFields` (write, lint, template-write mockups) follow. **[H]**
- **`openness-lint-engine.md:34`**: `required_fields` checker is "the *only* consumer of
  `.Required` besides the write/approve twins"; retirement path cites `seam-core-pipeline.md:258-259,268`. **[H]**
- **`openness-lint-engine.md:105`** verdict on `required_fields`: "STRUCTURALLY REQUIRED today,
  HISTORICAL tomorrow… three implementations of one rule, and the lint copy is a hand-built
  mirror of the write helper. JSON Schema `required` subsumes it." **[H]**
- **Special-casing caveat** (blocks a pure rename): `openness-write-path.md:123` — `type`/`title`
  are routed out of `Fields` by `frontmatter.Parse` (`frontmatter.go:70-82`), so CEL cannot
  `has("type")` today; the schema gate must keep special-casing them unless paired with making
  them ordinary fields. **[H]**
- **CEL field-selector caveat** (`seam-core-pipeline.md:271-273`): a `-`/unicode key is legal in
  JSON Schema `properties` but not a bare CEL selector; templates must keep keys
  CEL-identifier-safe. Coexistence "just moves" the problem. **[H]**

### 1.3 Migration mechanics raised by the audits

- **`seam-core-pipeline.md:297-300`** ("Template file parsing is strict-ish"): `detectOldFormat`
  (`template.go:73-82`) already rejects `required`/`optional`/`body`; "a new required block
  should get the same treatment (reject an unknown top-level block, or at least fail when both
  `schema.frontmatter` and `schema.json` declare `required` for the same key)." **[H]**
- **Loader is non-strict** — `template.go:106-114` unmarshal ignores unknown keys
  (`seam-core-pipeline.md:299`; A1 `recon-code.md:66-70` **[H]**). A1's landing: "unknown
  template keys rejected at authoring time (`template write`), warned at load" — load path stays
  lenient, `template write` becomes strict. **[H]**
- **Ordering/error aggregation** (`seam-core-pipeline.md:274-278`): schema + CEL errors must
  merge into one report / one exit-1; `--json` needs stable IDs (JSON Schema keyword + instance
  pointer; CEL rule `ID`). **[H]**
- **Pipeline placement** — RFC-001 §7 (`:214-220`): `merge-conflict → required gate (from schema)
  → JSON Schema → CEL validations → write`; structural first so CEL reads required keys
  unguarded. A1 `recon-code.md:52-54,163-165`: schema pass slots between `write.go:523` gate and
  `write.go:529` CEL, validating the RAW document view (no date coercion). **[H]**
- **Phasing** — RFC-001 §16 (`:383-388`): `schema.frontmatter` retirement + migration path sits
  in **Phase 3 — lifecycle**, not Phase 1. RFC-002's working phase plan (A1
  `rfc002-working.md:738-741`) instead puts "`schema.frontmatter` hard rejection" in **Phase 1 —
  validation core**. **Conflict to note for A2.** **[H]**
- **No migration code exists** — A2 `recon-code.md:171`: `grep -rni "migrate"` over
  `cmd/akb/` + `internal/template/` = **zero hits**; plan-level notes record the v1→v2 decision
  as "hard break, no migration tool, no dual format" (see §3.3). Hook point for a new subcommand:
  `cmd/akb/template.go:26-28` (`templateCmd`, `init()` self-registration). **[H]** (this file is
  the parallel code-recon child's artifact, not prior research — cross-referenced only.)

### 1.4 Mockup obligations (RFC-001 §6.2 and its research ancestors)

- RFC-001 §6.2 (`RFC-001…:214-220`): pass mockup must satisfy schema block **and** all CEL rules;
  fail mockup must fail the **intended** validator; `<!-- FAILS: rule_id — … -->` convention
  extends to `<!-- FAILS: schema: /frontmatter/status -->`; optional-key-stripping + self-`old_page`
  proofs stay CEL-side, "declared optional keys" re-derived from the schema block. **[H]**
- `seam-core-pipeline.md:279-290` gives the same three obligations with anchors
  (`templates_write.go:74-227`, pass at `:143-150`, CEL variants `:155-188`, fail at `:219-227`,
  self-`old_page` `:179-188`); adds a **satisfiability** check to catch has-no-solution schemas. **[H]**
- `seam-periphery.md:153` (#6) and `:154` (#7): `template write` must additionally prove the
  mockup against the JSON Schema; error text gains a structural branch; "a template without
  mockups is unproven". `references/TEMPLATE.md:12,153,117`. **[H]**
- `seam-periphery.md:148-160` (#1-9): the exact skill-doc surfaces that must change —
  `TEMPLATE.md:34-58` anatomy (two-layer), `:60` "Presence is all the schema enforces there…"
  (the sentence the design contradicts — **highest-priority rewrite**), `:102-119` `has()` guard
  rules (required reads become structural), `:269-292` DO/DON'T incl. old-format rejection. **[H]**
- `seam-periphery.md:196` (#13): `internal/template/embedded/adr.yaml:1-8` (guard-doctrine header
  + its `schema:` block `:12-48`) is "the **first artifact to restructure**, and `note.yaml` the
  minimal counterpoint." **[H]**
- `akb-recon.md:160-166`: current mockup machinery to reuse in a migration/rewrite —
  `buildTestPage` (`templates_write.go:341-345`), `evaluateValidations` (`:352-378`),
  `optionalKeysSuppliedByMockup` (`:384-397`), `withoutFrontmatterKey` (`:415-427`),
  `templateFilesMatch` (`:432-447`), atomic temp-dir + rename + single commit (`:261-337`). **[H]**

### 1.5 Existing-page / untyped-page disposition (the "not templates" migration surface)

- `seam-core-pipeline.md:339-341` (risk #8): "A JSON Schema block makes validation possible on
  pages that never passed `akb write` (git pull, index rebuild ingestion) — same as lint today;
  the sweep must run schema checks per page and **degrade like CEL does** (lint/cel.go:84-95),
  not abort the sweep." **[H]**
- `openness-document-model.md:111` (date-regime option 3 cost): "the schema must exist for a page
  to be coerced, so **lint-time handling of untyped pages needs a fallback**." **[H]**
- `SYNTHESIS-openness.md:108` ("Permissive vs safe"): demoting gates makes untyped/invalid pages
  possible; "the sweep (lint) becomes the only net" (ties to D11 index-rebuild ingestion). **[H]**
- `LOG.md:603` (D19): "untyped pages in derived artifacts = store raw, group explicitly"
  (`(untyped)` index.md section; never silently dropped; `type_orphan` flags). **[H]**
- `command-inventory.md:106,126`: index-rebuild ingestion bypass is the validation seam risk. **[H]**
- `openness-lint-engine.md:143` (OQ5): whether `index.md` is still a registry at all given
  `index rebuild` is authoritative. **[M]**
- **No prior source proposes a page-content migration** — only sweep-degrade policy. **[L]**

### 1.6 Open questions the audits left (A2-relevant subset)

- `openness-document-model.md:179-183` — OQ list:
  - **OQ2** (`:180`): is `type` an engine key or a schema property; who selects the schema?
  - **OQ4** (`:182`): **"Reorder write-time validation:"** `BuildPage` runs *before*
    required-field checks and CEL; if projections become schema-declared, declared-vs-parsed
    order matters ("a schema-driven view needs the schema loaded earlier than today's
    `checkRequiredFields`").
  - **OQ5** (`:183`): **"One parse or five?"** — does `frontmatter.Parse` keep its own
    `extractBody` once a single parse-with-extensions is the front door?
  - (`:181` OQ3 dialect declaration — adjacent, not A2.)
- `seam-core-pipeline.md:325-341` risk list, esp. #2 (duplicated exclusion logic becomes a
  published contract), #7 (no JSON Schema dep in `go.mod`), #8 (sweep degrade, §1.5 above). **[H]**
- `seam-periphery.md:191` enum duplication (§1.1) and `:192` divergent failure policy
  (write fails closed vs lint degrades) — a schema checker needs a stated policy for malformed
  page vs malformed schema. **[H]**
- `seam-periphery.md:193` sweep ordering: structural pass should run before `cel_lint`. **[H]**
- `openness-config-surface.md:117` (E5): old-format keys hard-rejected — "**STRUCTURAL**
  (one-way migration)". **[H]**

---

## 2. `.pi/subagents/proposals/json-cel/research/LOG.md` — session decisions that constrain A2

- `LOG.md:103` (session-3 template-format decision): "`lint_rules[]` stays CEL (temporal),
  **`schema.frontmatter` retired** (its presence job moves to JSON Schema `required`; invariant:
  the required set = exactly the keys CEL may read unguarded). Dialect must be pinned
  (recommend 2020-12)." **[H]**
- `LOG.md:176` (RFC-001 outcome): "`schema:` becomes a JSON Schema 2020-12 document over the
  document view; `schema.frontmatter` retired with loud rejection; `validations[]`/`lint_rules[]`
  unchanged." **[H]**
- `LOG.md:386`: "CEL churn verified mechanical (adr.yaml page.ast.*, page.content.word_count)" —
  renames ride "the already-breaking TemplateV2 revision." **[H]**
- Owner content decisions relevant to embedded-template restructure: `LOG.md:603` (D18 `title_field:`,
  default `title`; D19 untyped pages, §1.5). **[H]**

---

## 3. RFC-001-era material (migration precedent)

### 3.1 RFC-001 §6/§15/§16 (the source of A2's charter)

- `docs/knowledge-base/kb/rfcs/RFC-001-json-schema-cel-coexistence.md:157-159`: "## 6. Template
  format / ### 6.1 Shape (TemplateV2 revision — **breaking, loudly rejected like the v1
  rejection**)". **[H]**
- `:198-205` retirement + guard invariant (quoted §1.1/§1.2). **[H]**
- `:214-220` §6.2 mockup obligations (quoted §1.4). **[H]**
- **§15 candidate #2** (`:367-368`): "**ADR: `schema.frontmatter` retirement mechanics** — clean
  break vs one-release alias; migration error text; in-place migration command
  (`akb template migrate`?) vs manual." **[H]**
- §16 (`:386-387`): retirement + migration path = **Phase 3**. **[H]**
- §17 risk table (`:391`): "`schema.frontmatter` retirement + migration path." **[H]**

### 3.2 The v1→v2 rejection precedent — what `detectOldFormat` does

- `internal/template/template.go:73-82` (**two live copies**):
  ```go
  func detectOldFormat(raw map[string]any, filename string) error {
      for _, key := range []string{"required", "optional", "body"} {
          if _, ok := raw[key]; ok {
              return fmt.Errorf("parse %s: Template format has changed. Please update to the new schema.", filename)
          }
      }
      return nil
  }
  ```
  called from `LoadTemplates` at `:109`, after `yaml.Unmarshal(data, &raw)` into `map[string]any`
  (`:105-108`) and before the typed decode (`:113-116`). A duplicate with a **different message**
  (lowercase; `;` instead of `.`) exists at `cmd/akb/templates_write.go:461-467` for
  `akb template write` (`:93`). **[H]** (A2 `recon-code.md:42-49`; A1 `recon-code.md:63-70`.)
- **How users were messaged:** a single hard error naming the file and the three rejected keys —
  "Template format has changed. Please update to the new schema." No alias, no dual-format read,
  no auto-conversion. **[H]**
- `akb-recon.md:46`: loader's first unmarshal into `map[string]any` exists **only** to run this
  check. **[H]**
- `openness-config-surface.md:117` (E5): verdict **STRUCTURAL (one-way migration)**. **[H]**
- `seam-periphery.md:156` (#9): "The old-format rejection is the precedent for a hard, loudly-errored
  schema migration." (Skill doc `references/TEMPLATE.md:269-292` already teaches "DO NOT use the
  old template format".) **[H]**
- A1 `recon-code.md:65`: "RFC §6 explicitly cites this precedent for the `schema.frontmatter`
  retirement **[H]**." **[H]**

### 3.3 v1→v2 migration experience as recorded (plan-level)

- `.sisyphus/drafts/cel-validation-plan.md:232-240` (Decision 1, **RESOLVED — Hard Break**):
  > "No backwards compatibility required. AKB is not out in the wild nor used in production.
  > Breaking changes are absolutely acceptable… All existing `.akb/templates/*.yaml` files must
  > be rewritten to the new schema. **No migration tool, no dual-format support.** The embedded
  > default templates (`adr.yaml`, `note.yaml`) will be rewritten as part of this work. Users
  > creating new KBs via `akb init` will receive the new-format defaults." **[H]**
  Decision 2 (`:242-246`): `dir` kept; `body` removed, replaced by the `_pass.md` mockup as the
  canonical valid example. **[H]**
- `.sisyphus/plans/cel-validation.md:113` (Guardrails): "**NO** backwards compatibility /
  migration tooling (hard break)." **[H]**
- (A2 `recon-code.md:171` records both refs; the above are the primary anchors.) **[H]**

---

## 4. Items explicitly deferred to A2 / the migration ADR by A1/A3/A4

- **A1 (`rfc-002-a1-validator`)** — no line names "A2" by ID, but the migration mechanics were
  handed forward:
  - `LOG.md:105-106` — `schema:` collision → "V3 repurposes with hard-reject migration
    (precedent detectOldFormat, template.go:73-82)" **[H]**
  - `recon-code.md:172-174` — integration constraint #1: "**`schema:` is not a new key — it is
    the existing `Schema yaml:"schema"`** … repurposing/renaming that key with a hard-reject
    migration modeled on `detectOldFormat` … and making `template write` strict while
    `LoadTemplates` stays lenient." **[H]**
  - `LOG.md:122-133` "Parked/deferred" — all A1 residue is *other* (ADR-001/RFC-003 orphans,
    kb-management doctrine → RFC-002 Phase 4); **nothing migration-mechanical was parked here**. **[H]**
  - A1's ADR-003 adds a **migration constraint A2 must honour** (per A2 LOG:21-24):
    "unknown format names warn at `template write` (I7); no per-template opt-outs (N4). →
    migration that emits `format: date-time` declarations opts old fields into assertion (strict
    profile) — old pages with lax forms would fail sweep; must surface." **[H]**
    (ADR-003 N4 anchor: `docs/.../ADR-003-santhosh-v6-format-assertion.md:55`.)
- **A3 (`rfc-002-a3-celgo`)** — **nothing deferred to A2.** Its parked list
  (`LOG.md:164-179`) defers to A5/A6 and implementation-phase verification only; item 3 sends
  `internal/cel/types.go` cleanup to **A6**, not A2. **[H]**
- **A4 (`rfc-002-a4`)** — **nothing deferred to A2.** Parked (`LOG.md:163-168`): `iso_duration()`
  → **A5**; sweep-side schema checker ownership → **implementation planning**; open questions
  "None blocking." **[H]**
  - A4's ADR-005 constraint on A2 (per A2 LOG:26-31): migration "must detect CEL rules calling
    `timestamp()`/`duration()` on frontmatter paths and decide declarations (or warn) — same
    static analysis ADR-005 specifies"; bridge coerces only schema-declared temporal fields (I1). **[H]**

---

## 5. `akb template migrate` — every mention found

- `docs/knowledge-base/kb/rfcs/RFC-001-json-schema-cel-coexistence.md:368` — §15 candidate #2,
  posed as a question: "in-place migration command (`akb template migrate`?) vs manual." **[H]**
- `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md:792` (candidate table, A2 row):
  "`schema.frontmatter` retirement + migration (`akb template migrate`?)" — status `proposed`. **[H]**
  (secondhand copies: A1 `rfc002-working.md:781`; A3 `RFC-002-edited.md:792`.)
- A2 run LOG `:5`, `:47`, `:74` — the candidate title + the open fork. **[H]**
- A2 `recon-code.md:171-173` — **no such subcommand exists**; hook point `cmd/akb/template.go:26-28`. **[H]**
- **No prior research source names the command as a decision.** It appears only as a bracketed
  question. **[L]**

---

## 6. Gaps — what the prior research did NOT settle

1. **Automated vs manual migration is unanswered.** RFC-001 §15 poses the fork (`:368`); no
   source argues either side, and no cost/benefit is recorded. The v1→v2 precedent
   (`.sisyphus/drafts/cel-validation-plan.md:232-240`) resolved the *same* fork as **hard break,
   no tool** — but under "not in the wild / not production", a premise that may no longer hold.
2. **No concrete migration mechanics.** Nothing specifies in-place rewrite vs report/diff,
   per-template vs all, `_pass`/`_fail` mockup migration, or how `FieldSchema{type,enum}` maps to
   JSON Schema keywords (the mapping is implied only by `adr.yaml`-style rules).
3. **Existing-page disposition is policy-only, not mechanical.** Sources say the sweep must
   degrade (`seam-core-pipeline.md:339-341`) and untyped pages need a fallback
   (`openness-document-model.md:111`), but no source decides whether pages are migrated,
   re-validated, or left to fail the sweep; RFC-002 §19's "existing pages' `created`/`updated`
   stay on disk as data" is quoted only in the A2 LOG, not analysed.
4. **Embedded-template restructure ownership is unresolved.** `seam-periphery.md:196` names
   `adr.yaml`/`note.yaml` as "the first artifact to restructure", but the A2 LOG (`:55`) lists
   "Embedded templates restructure scope under A2 vs. Phase-4 doctrine work" as still open.
5. **Error text / alias posture is undecided.** RFC-001 §15 names "clean break vs one-release
   alias" and "migration error text"; no source recommends one. `detectOldFormat`'s message is
   the only precedent text (and it is duplicated with divergent wording at
   `templates_write.go:461-467`).
6. **Date-field auto-declaration is undecided** (A2 LOG `:53`; `openness-document-model.md:108-111`
   gives three regimes and their consumer impact but no choice).
7. **The Phase-1 vs Phase-3 placement conflict is unresolved** (RFC-001 §16 Phase 3 vs RFC-002
   working plan Phase 1; §1.3).
8. **`type`/`title` gate special-casing under the new schema is not resolved** —
   `openness-write-path.md:123` flags that `frontmatter.Parse` routes them out of `Fields`, so
   CEL `has("type")` is impossible; whether A2 makes them ordinary fields is open.
9. **No JSON Schema library is in `go.mod`** (`seam-core-pipeline.md:341`; A1 `recon-code.md:113`)
   — A2 cannot specify emitted dialect/keyword mapping until A1's ADR-003 (santhosh v6) is assumed.
10. **No source audits the in-tree TemplateV2 inventory** (this repo's own KB templates, other
    KBs' `.agent-kb/templates/`) for migration sizing — the A2 LOG (`:62-63`) assigns that to
    code-recon (the sibling child), so A2's ADR must fold in that child's inventory.

---

## 7. Start-here for the A2 author

1. `seam-core-pipeline.md:256-300` — the retirement table + migration notes (invariant,
   field-selector caveat, mockup growth, strict-ish parsing). The single densest A2 source.
2. `RFC-001…:198-220, 367-368, 383-388` — the charter: retirement, invariant, §6.2 mockups,
   the migration ADR question, phasing.
3. `internal/template/template.go:73-82` + `.sisyphus/drafts/cel-validation-plan.md:232-240` —
   the precedent A2 must follow or deliberately extend.
