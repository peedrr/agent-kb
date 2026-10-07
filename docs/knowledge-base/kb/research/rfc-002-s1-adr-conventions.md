---
grammar: 1
type: research
title: "ADR conventions and governing decisions for SPEC-002 (RFC-002-S1)"
status: final
provenance: agent-drafted
run: rfc-002-s1
as_of: "2026-10-07"
created: "2026-10-07"
updated: "2026-10-07"
scope:
  - "docs/knowledge-base/**"
tags:
  - adr
  - conventions
  - spec-002
  - knowledge-base
summary: "Conventions and governing decisions (ADR-003..006, SPEC-001, the template contract) for authoring the SPEC-002 page recording RFC-002-S1."
informs:
  - SPEC-002
---
# Code Context

Conventions + governing decisions for the new SPEC page recording RFC-002 candidate RFC-002-S1 ("Structured validation report") at `docs/knowledge-base/kb/specs/`. All paths relative to repo root; anchors are `file:line`.

## Files Retrieved
1. `docs/knowledge-base/kb/decisions/ADR-003-santhosh-v6-format-assertion.md` (1-115) — frontmatter, origin line, format/error policy
2. `docs/knowledge-base/kb/decisions/ADR-004-cel-go-upgrade-pinned-extensions.md` (1-118) — pin, extensions, cel-go error-string churn note
3. `docs/knowledge-base/kb/decisions/ADR-005-date-bridge-determinism.md` (1-129) — write-time warning conventions (I9-I12, N6)
4. `docs/knowledge-base/kb/decisions/ADR-006-schema-frontmatter-retirement.md` (1-117) — rejection message/exit-1 conventions (I1-I4)
5. `docs/knowledge-base/kb/specs/SPEC-001-init-versioning.md` (1-253) — the only existing SPEC; shape to mirror
6. `docs/knowledge-base/kb/index.md` (1-21), `docs/knowledge-base/kb/log.md` (1-14) — indexing/logging entries
7. `docs/knowledge-base/.agent-kb/templates/spec.yaml` (1-281) — LINT/VALIDATION contract the new SPEC must satisfy
8. `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` (769-800, 334-335) — §18 origin rule + S1 row/scope

## 1. ADR-003 — santhosh v6 + format assertion
Frontmatter (lines 1-28) verbatim:
```yaml
---
created: "2026-10-06"
deciders:
- Pete Hope
grammar: 1
id: ADR-003
is_draft: false
provenance: agent-drafted
revisit:
- kaptinlin/jsonschema or another Go validator publishes draft 2020-12 test-suite or bowtie compliance plus a documented built-in-format override hook
- santhosh-tekuri/jsonschema maintenance lapses (no commits for a year) or the module is abandoned
- cel-go changes timestamp() string acceptance again (re-check at every cel-go upgrade)
- a built-in non-temporal format checker false-rejects values a template author reasonably declared
scope:
- internal/**
- cmd/akb/**
- go.mod
status: accepted
summary: JSON Schema validation delegates to santhosh-tekuri/jsonschema v6; every declared format asserts; akb-registered date-time/date checkers pin the strict RFC 3339 profile cel-go accepts.
tags:
- json-schema
- validation
- format-assertion
- rfc-002
title: JSON Schema Validation Uses santhosh v6 and Asserts Every Declared Format
type: adr
updated: "2026-10-06T17:16:30Z"
---
```
- Ratified decision + pin (ADR-003:38): "akb's JSON Schema layer is `github.com/santhosh-tekuri/jsonschema/v6`, compiled for draft 2020-12 with format assertion enabled globally, and akb registers its own `date-time` and `date` format checkers — pinned to the strict RFC 3339 profile of I4 — before any schema is compiled, because the write-time schema pass must guarantee CEL's `timestamp()` preconditions under both the currently pinned cel-go v0.28 and the RFC-002-A3 upgrade target v0.32".
- I2 (43): "The system MUST enable format assertion on the schema compiler (`Compiler.AssertFormat()`) for every compiled template schema, at write time and at sweep time." I3 (44) asserts every declared `format` keyword. Strict profiles: I4 (45) date-time pattern; I5 (46) RFC 3339 `full-date`; I6 (47) `format: time` stays a string.
- Error-output/type statements a spec must comply with: I7 (48): "WHEN `akb template write` encounters a `format` name outside the known registry, the system MUST warn on stderr that the format will not assert; the template-load path MUST stay lenient…"; verification (65): "a template declaring `format: email` rejects an invalid email at `akb write` with exit 1, and the sweep-time schema checker flags the same page"; (68): "`akb template write` with an unknown `format` name succeeds with a stderr warning naming the format". No machine-readable error type is defined here.

## 2. ADR-004 — cel-go pin + extension set (CEL error behavior)
Frontmatter (lines 1-31) verbatim:
```yaml
---
created: "2026-10-06"
deciders:
- Pete Hope
grammar: 1
id: ADR-004
is_draft: false
provenance: agent-drafted
revisit:
- cel-go publishes a release newer than v0.32.0 — re-check timestamp() acceptance, the cost model, and the pinned extension versions before any bump
- the issue
- akb begins ingesting templates from untrusted sources — set cel.RegexProgramSizeLimit
- a rule needs a capability behind OptionalTypes, ext.Regex, TwoVarComprehensions, Encoders, or Native — draft that capability's own ADR
- cel-go deprecates cel.CostLimit or changes interpreter.EvalCancelledError
scope:
- internal/**
- cmd/akb/**
- go.mod
- go.sum
status: accepted
summary: cel-go upgrades to cel.dev/cel-go v0.32.0 in one require+import rewrite; five ext libraries enabled in NewEnv pinned at their highest v0.32.0 versions; regex plan-size knob stays unbounded.
tags:
- cel
- dependency-upgrade
- extensions
- validation
- rfc-002
title: CEL Runs on cel.dev/cel-go v0.32 with a Pinned Extension Set
type: adr
updated: "2026-10-06T19:22:34Z"
---
```
- Pin (ADR-004:41): "Upgrade cel-go from `github.com/google/cel-go` v0.28.0 to `cel.dev/cel-go` v0.32.0 — one change rewriting the `go.mod` require and every import…". I1 (45) requires `cel.dev/cel-go` in go.mod and every import; N1 (54) forbids any `github.com/google/cel-go` reference.
- Extension set: I4 (48) "MUST enable exactly the extension libraries `ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`, and `ext.Bindings` in `NewEnv` (`internal/cel/engine.go`) for every CEL environment." I5 (49) pins numbered versions (`StringsVersion(5)`, `ListsVersion(2)`, `MathVersion(2)`). N4 (57) forbids changing the program-cache key or per-template env variation.
- Error-string churn (the one statement about CEL error text; ADR-004:108): "Bad, because tests pinning cel-go-owned error strings (`no such key: …`, the `overload` substring) may need message updates on the bump; the upgrade PR must distinguish message churn from behavior change." A report spec must therefore not pin cel-go message text as a contract.
- N6 (59): "The system MUST NOT define date-time acceptance by cel-go's string-to-timestamp gate or adopt v0.32's `types.ParseTimestamp` helper; the ADR-003 I4 schema checkers remain the sole acceptance definition".

## 3. ADR-005 — date bridge; write-time warning formats
Frontmatter (lines 1-31) verbatim:
```yaml
---
created: "2026-10-06"
deciders:
- Pete Hope
grammar: 1
id: ADR-005
is_draft: false
provenance: agent-drafted
revisit:
- cel-go changes duration() or timestamp() string acceptance (re-check at every cel-go upgrade alongside ADR-004's revisit)
- santhosh-tekuri/jsonschema changes its built-in duration checker or RegisterFormat override semantics
- a template needs ISO 8601 month/year duration arithmetic in CEL — route to RFC-002-A5's registry decision
- the permissive applicator traversal coerces a value a template author did not intend (restrict traversal or widen the static check)
- the CEL env gains typed frontmatter declarations — re-evaluate the syntactic-only static check
scope:
- internal/**
- cmd/akb/**
- internal/skill/embedded/**
status: accepted
summary: The date bridge coerces only schema-declared temporal fields, identically for page/old_page; template write warns on unbacked temporal calls; the duration seam is documented, bridging deferred to A5.
tags:
- cel
- json-schema
- date-bridge
- determinism
- template-authoring
- rfc-002
title: The Date Bridge Coerces Only Schema-Declared Fields and Is Statically Checked
type: adr
updated: "2026-10-06T21:39:49Z"
---
```
- Warning conventions (WARNING-ONLY, stderr, never a failure): I9 (53) statically inspects the compiled CEL AST for `timestamp`/`duration` on `page.frontmatter.*` / `old_page.frontmatter.*`; I10 (54): "MUST emit a stderr warning naming the rule ID, the attribute path, and the remediation"; I11 (55): "MUST emit a stderr warning naming the rule ID, the attribute path, the vocabulary mismatch … and the remediations"; I12 (56): non-resolvable argument → no warning.
- N6 (66): "The static temporal check MUST NOT fail `akb template write` or block a template from being written; findings MUST be emitted as stderr warnings only, per the RFC-002 §5.2 '→ warning' scope and the ADR-003 I7 unknown-format precedent."
- Exit contract per verification (82): undeclared `timestamp()` → "exit 0 plus a stderr warning naming rule and path".
- Coercion is in-memory only (N3, line 63): "MUST NOT persist a coerced value to disk or expose one in the document view (`read --json`, schema-validation input)".

## 4. ADR-006 — rejection message / exit-1 conventions
Frontmatter (lines 1-30) verbatim:
```yaml
---
created: "2026-10-06"
deciders:
- Pete Hope
grammar: 1
id: ADR-006
is_draft: false
provenance: agent-drafted
revisit:
- external KBs running TemplateV2 templates appear in the wild before Phase 1 ships (revisit the no-command decision)
- the rejection message or CHANGELOG recipe proves insufficient in practice (author confusion or support burden)
- a future template-format break re-evaluates alias posture under post-1.0 stability expectations
- the frontmatter-key detection predicate false-positives on a real V3 schema
scope:
- internal/template/**
- cmd/akb/**
- internal/skill/embedded/**
- CHANGELOG.md
status: accepted
summary: TemplateV3 hard-rejects the retired schema.frontmatter block with one self-sufficient message; migration is a CHANGELOG recipe driven by ADR-005 warnings — no command, no alias, no page rewriting.
tags:
- template-format
- migration
- json-schema
- validation
- rfc-002
title: TemplateV3 Retires schema.frontmatter — Migration Is a Recipe, Not a Command
type: adr
updated: "2026-10-06T22:49:48Z"
---
```
- I1 (44) / I2 (45): loader and `akb template write` "MUST refuse the template and exit with code 1". I3 (46): a single shared detection helper emitting a single message text (no per-call-site forking; the divergent `template.go:77` vs `templates_write.go:464` copies are named the anti-pattern at line 69).
- I4 (47) — the message must be self-sufficient: "The refusal message MUST name the template file, the retired construct (`schema.frontmatter`), the four-step migration essence … and the CHANGELOG as the recipe's home."
- N4 (57) and general posture: a rejection reports what was found + what to do; warnings do not block (cf. §3).

## 5. How each ADR records its RFC-002 origin (mechanism to copy)
- Governing rule — `RFC-002 §18` (RFC-002-open-validation-engine.md:775-776): "A candidate is drafted as a real ADR/SPEC, taking the next free global number; its header records `Origin: RFC-002 candidate <ID>`."
- Mechanism per ADR = (a) frontmatter `tags` includes `- rfc-002` (ADR-003:24, ADR-004:27, ADR-005:27, ADR-006:26); (b) the `## Context` section opens with a line beginning `Origin:`.
- ADR-003:75: "Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A1; ratified 2026-10-06 by Pete Hope in the research session (`.pi/research/rfc-002-a1-validator/`)…"
- ADR-004:80: "Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A3 (RFC §10), researched 2026-10-06 in `.pi/research/rfc-002-a3-celgo/`…"
- ADR-005:90: "Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A4 (RFC §5.2 names the scope…), researched 2026-10-06 in `.pi/research/rfc-002-a4/`…"
- ADR-006:81: "Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A2 (RFC §6 charters the hard rejection; §18 governs this record's lifecycle), researched 2026-10-06 in `.pi/research/rfc-002-a2/`…"
- There is **no** `origin:` frontmatter key: neither `templates/adr.yaml` (keys: grammar,type,id,title,status,created,updated,provenance,scope,tags,summary,deciders,supersedes,superseded_by,revisit) nor `templates/spec.yaml` defines one. For the SPEC, the spec template has no `Context` H2, so place `Origin: RFC-002 candidate RFC-002-S1 …` in SPEC-001's provenance slot — the intro paragraph immediately after the `> **For agents:**` blockquote (SPEC-001:93, which reads "Migrated from the pre-KB design record preserved at `raw/spec/init-versioning-spec.md`…") — and add `- rfc-002` to `tags` (spec.yaml does not declare `tags`, so check whether extra keys are tolerated; SPEC-001 carries no `tags`).

## 6. SPEC-001 structural conventions (`specs/SPEC-001-init-versioning.md`, 253 lines)
Full frontmatter (lines 1-84) verbatim:
```yaml
---
anchors:
- checked_at: "2026-10-01T16:22:52Z"
  claim: REQ-005
  evidence: resolves at git.go:430; commits exactly the named paths, unrelated staged changes stay staged; named test PASS
  kind: flow
  path: internal/storage/git.go
  state: live
  symbol: CommitFiles
  verify:
    check: go test ./internal/storage/ -run TestGitProviderCommitLeavesUnrelatedStagedChanges
    method: check
- checked_at: "2026-10-01T16:22:52Z"
  claim: REQ-006
  evidence: resolves at mode.go:82; ParseMode(cfg.Versioning) selects the provider — mode comes from the akb.yaml key
  kind: flow
  path: internal/storage/mode.go
  state: live
  symbol: OpenStore
  verify:
    check: go test ./internal/storage/...
    method: check
- checked_at: "2026-10-01T16:22:52Z"
  claim: REQ-011
  evidence: resolves at identity.go:121; precedence flags -> env -> git config -> default confirmed in the body, source decides recording
  kind: local
  path: internal/storage/identity.go
  state: live
  symbol: ResolveInitIdentity
  verify:
    check: go test ./internal/storage/...
    method: check
- checked_at: "2026-10-01T16:22:52Z"
  claim: REQ-012
  evidence: resolves at init.go:229; the IdentityFromFlag branch returns the pair for a complete flag set, which init records in akb.yaml, and nil for a partial pair, which stays invocation-scoped (ADR-002); flag tests PASS
  kind: local
  path: cmd/akb/init.go
  state: live
  symbol: initIdentity
  verify:
    check: go test ./cmd/akb/ -run 'TestInitAuthor|TestInitPartial|TestInitRecordsTheDefaultIdentity' -count=1
    method: check
- checked_at: "2026-10-01T16:22:52Z"
  claim: REQ-013
  evidence: resolves at identity.go:73; env -> akb.yaml git-author/git-email -> git-native confirmed; unresolvable yields ErrNoCommitIdentity
  kind: flow
  path: internal/storage/identity.go
  state: live
  symbol: ResolveIdentity
  verify:
    check: go test ./internal/storage/...
    method: check
- checked_at: "2026-10-01T16:22:52Z"
  claim: REQ-016
  evidence: resolves at git.go:197 as a GitProvider method; clean merges detected via MERGE_HEAD rev-parse in mergeInProgress
  kind: local
  path: internal/storage/git.go
  state: live
  symbol: checkMergeConflicts
  verify:
    check: go test ./internal/storage/...
    method: check
created: "2026-09-25"
grammar: 1
id: SPEC-001
kind: change
provenance: human
scope:
- cmd/akb/**
- internal/storage/**
- internal/config/**
- internal/skill/embedded/**
status: completed
title: Init Selects the KB Versioning Mode and Resolves Commit Identity
type: spec
updated: "2026-10-01T16:23:16Z"
verified:
  at: "2026-10-01T16:22:52Z"
  branch: main
  commit: eeebad88e0bf595af0490683fb4aafb80dd6c0b1
  method: symbol-resolve+blame-trace
  next_review_by: "2027-01-01"
  state: live
---
```
- No body metadata table. Instead a `> **For agents:**` blockquote (88-91): "this SPEC is authority only while `status: active` AND / `verified.state: live`. If your task conflicts with it, or any anchor fails / to resolve at HEAD, stop and name the conflict. Do not silently work around / it; propose an amendment or supersession instead." (Required verbatim by spec.yaml `challenge_affordance`: must contain "this SPEC is authority only while".)
- H1 (86) = `# SPEC-NNN: <title>` (identical to frontmatter title). Intro paragraph (93) = provenance/migration statement.
- Section list (all H2): Why (95), Goals and Non-Goals (99), Current Behaviour (109), Requirements (118, `REQ-NNN: WHEN … MUST …`), Acceptance Criteria (139, `AC-NNN (verifies REQ-NNN): …` + indented `verify: { method: check|ask, check|ask: … }`), Contract and Invariants (180), Preserved Behaviour (190), Decisions and Rejected Alternatives (198), Assumptions and Open Questions (207), Affected Surface and Ordering (213, closes with bare `depends_on: []` at 221), Verification Plan (223), Drift Ledger (229, append-only `- <date> · <event> · <from> → <to> · … · evidence:`), Revisit Triggers (238), Evidence Appendix (245).
- Wikilink style for ADR/RFC references: `[[ADR-002-init-author-flags]]` and `[[RFC-002-open-validation-engine|RFC-002]]` (target-only or `target|label`), used inline and in the Drift Ledger; ADRs instead link `[[ADR-003-santhosh-v6-format-assertion|ADR-003]]` in References.
- `verified` block values: `at` (RFC3339Z), `branch`, `commit` (full SHA), `method: symbol-resolve+blame-trace`, `next_review_by`, `state: live`. Anchor fields: `checked_at, claim (REQ-NNN), evidence, kind (local|flow|invariant), path (no line numbers), state (live|drifted|lost|unverified), symbol, verify.{check,method}`.
- Length: 253 lines total.
- spec.yaml also enforces: H2s Goals and Non-Goals, Requirements, Acceptance Criteria, Assumptions and Open Questions, Affected Surface and Ordering, Verification Plan, Drift Ledger, Revisit Triggers; `kind ∈ {change, capability}` (`capability` additionally requires `## Current Behaviour`); `id` `^SPEC-[0-9]{3}$`; `status` enum draft|proposed|active|completed|superseded|withdrawn (only `active` needs `ratified_by` + `verified.state == live`); body must contain `REQ-NNN` and `AC-NNN (verifies REQ-NNN)`; no `[NEEDS CLARIFICATION`; every code block needs a language; word_count ≥ 50; a `## Decisions and Rejected Alternatives` section must contain "Do not re-propose unless"; `kind: change` at `proposed` is valid without `ratified_by`.

## 7. index.md / log.md entries for SPEC-001
- `kb/index.md:14` under `## Specs`: "- [Init Selects the KB Versioning Mode and Resolves Commit Identity](kb/specs/SPEC-001-init-versioning.md) — akb init chooses the KB versioning mode (embed/no-git/standalone) and commit identity resolves env -> akb.yaml -> git; migrated from raw/spec/init-versioning-spec.md" (link text = title; path = `kb/<dir>/<file>`; description after an em dash).
- `kb/log.md:5` under `## 2026-10-01 ingest`: "Ingested pre-KB init-versioning design record as raw/spec/init-versioning-spec.md and recreated it as specs/SPEC-001-init-versioning.md (completed change SPEC, template-conformant)".
- log.md heading grammar: `## <YYYY-MM-DD> <verb>` (`ingest`, `delete | <title>`, `distill`) followed by a blank line and prose.

## 8. RFC-002-S1 scope + governing template
- `RFC-002-open-validation-engine.md:796`: `| RFC-002-S1 | SPEC | Structured validation report (widened `runTemplateValidations`) | proposed | — |`.
- Scope statement, `:334-335`: "Error merge (schema violations by keyword + JSON Pointer; CEL failures by rule ID; one exit-1 report) and the `runTemplateValidations` widening carry unchanged (RFC-002-S1)."
- Related exit contract, `:471-472`: "`akb lint` exits 1 iff any reported issue has severity `error`. Severities are KB-owned…".
- `templates/spec.yaml` is the loader-enforced contract; `dir: specs`, `name: spec`.

## Architecture
Every sub-document is a KB page validated by its type's CEL template (`.agent-kb/templates/*.yaml`): frontmatter keys are read unguarded only when `required: true`; body shape is enforced by `require_*` heading rules and raw-content rules. ADRs record their RFC origin as a `## Context` first line (`Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A<n>`) plus the `rfc-002` tag; SPEC-001 predates RFC-002 and instead puts a provenance sentence directly under the `For agents` blockquote — that is the slot the S1 SPEC should use. Acceptance criteria link back to `REQ-NNN`, `verified`/`anchors` provide the staleness signal, and `index.md`/`log.md` mirror each page's existence (index = title link + one-line description; log = dated verb entry).

## Start Here
Open `docs/knowledge-base/kb/specs/SPEC-001-init-versioning.md` (frontmatter 1-84, body 86-245) and clone its shape into `specs/SPEC-002-structured-validation-report.md`; then satisfy `docs/knowledge-base/.agent-kb/templates/spec.yaml` before writing, and add the `rfc-002` tag + `Origin: RFC-002 candidate RFC-002-S1` paragraph per §5.

## Supervisor coordination
No supervisor contact was needed; read-only extraction completed.
