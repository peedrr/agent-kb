---
grammar: 1
type: research
title: "RFC-002-A4 research log — date-bridge determinism and the duration seam"
status: final
provenance: agent-drafted
run: rfc-002-a4
as_of: "2026-10-06"
created: "2026-10-06"
updated: "2026-10-06"
scope:
  - "internal/cel/**"
  - "cmd/akb/**"
  - "internal/template/**"
tags:
  - date-bridge
  - determinism
  - adr-005
  - research-log
summary: "Run log for RFC-002-A4: date-bridge determinism spec, template-write static temporal check, duration seam, recon anchors, and the decision forks behind ADR-005."
informs:
  - ADR-005
---
# RESEARCH LOG — rfc-002-a4

## Run parameters
- MODE: start
- SUBJECT: rfc-002-a4
- Date: 2025-… (session start)
- Owner instructions: Consume RFC-002-open-validation-engine.md in full; begin
  RFC-002-A4. Pending decisions live in ADR-003 (A1) and ADR-004 (A3). Use
  subagents per protocol. If auto-resolvable → create ADR via akb + docs-writer
  skill. If owner input needed → owner-escalation skill.
- AKB_KB=/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb/docs/knowledge-base

## Question
Draft ADR resolving RFC-002 candidate **RFC-002-A4**: date-bridge determinism +
`template write` static temporal check + duration seam. Deliverable: one ADR
(next free global number, Origin: RFC-002 candidate RFC-002-A4) written via akb
into the KB, IF all decisions auto-resolvable; else owner-escalation interview.

## Agreed requirements (from RFC-002 §5.2/§18 — the scope statement itself)
1. Bridge determinism spec: traversal mechanics for schema-declared temporal
   fields (format: date-time / date) at ARBITRARY depth (nested objects, arrays);
   determinism guarantees; applied identically to page and old_page (built from
   fresh on-disk read under CURRENT template); in-memory only, never on disk;
   date→midnight UTC (already pinned by ADR-003 I5); format: time never coerced
   (ADR-003 I6); bridge = Go-stdlib pre-conversion only (ADR-004 N6).
2. `template write` static check: CEL rule calling timestamp(x) where x lacks a
   schema format assertion → WARNING (not error).
3. Duration seam: CEL duration() is Go-style; JSON Schema format: duration is
   ISO 8601 → document the seam; iso_duration() registry function = OPTIONAL
   bridge (registry contents are RFC-002-A5 scope — likely defer).
Constraints carried: ADR-003 I4 strict date-time profile is sole acceptance
definition; ADR-004 N4/N6 (no cache-key/env changes here).

## Decision forks anticipated (arbitrate after recon lands)
- F1: traversal semantics under applicators (anyOf/oneOf/$ref) — permissive
  reachable-declaration coercion vs restricted-unconditional vs annotation-driven
- F2: static check mechanics/limits (unresolvable args, old_page paths,
  non-literal expressions) — severity already fixed = warning
- F3: duration seam disposition — document-only vs schema-side assertion
  behavior vs iso_duration() (A5 boundary)
- F4: coercion-failure semantics at sweep (write is fail-closed unreachable)

## Decisions made
- **D1 (F1 — traversal semantics):** permissive static reachability. Coercion plan
  = every `format: date-time|date` declaration statically reachable in the schema's
  frontmatter subtree (properties/patternProperties/additionalProperties/items/
  prefixItems/contains/allOf/anyOf/oneOf/if/then/else; local $ref with visited-set
  cycle guard; `not` subtrees excluded). Coercion gated by the ADR-003 strict
  checkers (bridge reuses the same checker functions → assertion acceptance ==
  bridge acceptance == CEL timestamp() acceptance). Parse failure leaves the raw
  string (only reachable on non-matching branches or sweep-bypassed pages).
  Rationale: declaration-faithful reading of RFC §5.2; annotation-driven couples
  page-build to validator internals; unconditional-only is a D9 trap.
- **D2 (F2 — static check):** syntactic AST walk at `template write` (new additive
  helper; CompileRule discards AST today — engine.go:41-60). Warnings (never
  errors) for: timestamp() on undeclared path; duration() on format:duration path.
  Covers validations + lint_rules, page./old_page. roots; non-path args silent
  (documented limitation; frontmatter types as dyn so type-based checking unsound).
- **D3 (F3 — duration seam):** document-only + defer iso_duration() to RFC-002-A5.
  Schema-side assertion already free (santhosh built-in duration checker = RFC 3339
  App A profile, under ADR-003 assert-all). ISO months/years have no fixed length —
  calendar-semantics question belongs to A5. SKILL documents the seam.
- **D4 (F4):** folded into D1 — raw-string fallback, P4-consistent.
- **D5:** all forks auto-resolvable under ratified principles (P1–P7, D7,
  ADR-003/004) → no owner interview needed pre-drafting; ADR created as `proposed`,
  owner ratification is the gate. Scope-adjacent additions flagged for review:
  the duration()-on-format:duration warning (I11) beyond the RFC's literal scope.

## Outcome
- **ADR-005 created**: `kb/decisions/ADR-005-date-bridge-determinism.md`
  (status: proposed, provenance: agent-drafted, Origin: RFC-002 candidate
  RFC-002-A4). Written via akb (passed all template CEL gates first try),
  indexed via `akb index add`. Commits ae1c54c (write), b1d2b16 (index).
- akb lint: only orphans warning (no inbound links) — same as ADR-001;
  resolves at ratification when RFC-002's A4 row is rewritten per §18.

## Awaiting owner
- Nothing. RUN COMPLETE 2026-10-06.

## Ratification (owner-directed, 2026-10-06 ~21:38Z)
- ADR-005 flipped proposed → accepted; deciders: [Pete Hope];
  updated: 2026-10-06T21:38:41Z; Context gained the ratification sentence
  (ADR-003/004 style). Passed accepted_requires_deciders gate.
- is_draft left True: ALL KB pages incl. accepted ADR-003/004 show is_draft True
  (implicit-draft precedent) — akb approve deliberately NOT run.
- RFC-002 amended per §18 lifecycle: A4 row → ratified | ADR-005 (link);
  §5.2 traversal-mechanics ref → [[ADR-005]] link (first prose mention);
  §5.2 scope line → "Resolved by ADR-005:"; §19 risk row → plain ADR-005;
  updated → 2026-10-06T21:38:41Z. Remaining RFC-002-A4 occurrences = 2,
  both correct: spawns list + §18 ID column (A1/A3 precedent).
- Post-ratification lint: ADR-005 orphan warning CLEARED (RFC-002 links it);
  remaining orphans = ADR-001, RFC-003 (pre-existing, not this run's scope).
- 2026-10-06 owner UX review: owner asked whether I11 is actionable per the
  akb doctrine "errors carry mitigation". Answer: remediations exist (pattern-
  align schema to Go syntax, or drop the duration() call; iso_duration via A5),
  but I11 as drafted only "stated" the mismatch while I10 named remediation.
  AMENDED I11 + its verification row to require rule ID, path, mismatch, and
  both remediations. Rewrite passed gates.
- 2026-10-06 owner Q2: does A5 make I11 moot? NO — duration() is a CEL builtin
  the closed registry cannot redefine; format: duration keeps asserting ISO.
  Mismatch permanent in every A5 outcome; A5 ratifying iso_duration() adds a
  third remediation ("use iso_duration()") instead of retiring the warning.
  AMENDED Consequences with a Neutral bullet stating this lifecycle. Rewrite
  passed gates.

## Key findings
Prior research base (delegate extraction, anchors: `.pi/subagents/proposals/json-cel/research/`):
- Current coercion `convertDateField` (pagebuilder.go:26-45): top-level strings only,
  name-agnostic, RFC3339 + date-only; nested/list values NOT coerced — the asymmetry
  A4's traversal kills (akb-recon.md:111, openness-document-model.md:104).
- Coercion's ONLY consumer is the CEL page map (openness-document-model.md:132);
  `page` ≠ `document` — JSON view must carry raw values (LOG.md:96-97).
- All BuildPage call sites already hold the template — schema-aware traversal is
  "mechanical, not architectural" (openness-document-model.md:173, SYNTHESIS-openness.md:84).
- old_page: identical bridge under CURRENT template (LOG.md:620-621); lint sweep has
  NO old_page, only now (seam-periphery.md:68-70).
- Verified v0.28: duration() = Go time.ParseDuration, NOT ISO 8601; no ISO-8601
  duration anywhere in cel-go core (LOG.md:335-336, akb-recon.md:94).
- Prior recommendation: document duration seam + optional closed-registry
  iso_duration() bridge (LOG.md:356-357) — matches RFC-002 §5.2 text.
- Static check: only prior mention is the ADR-candidate line itself (LOG.md:355-356);
  NO prior design. Relevant: rules compiled with AST DISCARDED (engine.go:43-61,
  akb-recon.md:196) — extraction needs a new AST-exposing entry point. Prior art for
  static analysis at template-write: cost gate via cel-go static estimate
  (cel-bridge-research-b.md:151,196).
- Guard doctrine narrows under schema: required-key presence structural; has()
  remains for optional keys + old_page (seam-periphery.md:151, akb-recon.md:78,114).

Upstream facts (researcher, VERIFIED against pinned sources; full report:
subagent-artifacts/outputs/af533d6f-…/research.md):
- cel-go v0.32 `duration()`: Go time.ParseDuration ONLY (common/types/string.go);
  zero ISO-8601 duration anywhere in core or ext/ (enumerated); cel-spec langdef
  explicitly excludes days/weeks. `string(duration)` outputs `<seconds>s` — NOT a
  valid format: duration value, so ISO↔CEL round-trips are lossy by construction.
- cel-go v0.32 `timestamp()`: strictRFC3339Pattern gate (PR #1338, landed v0.30)
  + time.Parse + range check. ADR-003 I4 remains a subset of both v0.28/v0.32 —
  no A4 action. NOTE: pattern allows [Tt]/[Zz]/sec=60; time.Parse backstops.
- JSON Schema 2020-12 §7.3.1: `format: duration` = RFC 3339 Appendix A `duration`
  ABNF (verbatim captured) — P + dur-date/dur-time/dur-week, no sign, no
  fractions. Zero string-syntax overlap with CEL duration() — any bridge is a
  TRANSLATION the host performs, not a parse.
- santhosh v6 (v6.0.2/v6.0.3): built-in `duration` checker implementing exactly
  that ABNF profile (uppercase P, ordered units, no fractions/sign); overridable
  via Compiler.RegisterFormat (only `regex` reserved); assertion opt-in via
  AssertFormat() for draft ≥2019-09 (akb already opts in per ADR-003 I2).
  19 built-in formats total. CONSEQUENCE: under ADR-003's assert-all policy,
  `format: duration` ALREADY asserts ISO-8601 ABNF with zero new code.
- cel-go AST introspection: common/ast NavigateAST + Pre/PostOrderVisit +
  CallExpr(FunctionName/Args) + SelectExpr(Operand/FieldName) chains;
  cel.Ast.NativeRep() is the entry point and EXISTS AT v0.28 (pkg.go.dev
  "added v0.32" is the module-rename artifact). Attribute paths must be rebuilt
  manually from SelectExpr chains (no dotted-path helper — minor gap,
  ReferenceMap fields unread). Macro calls: SourceInfo.MacroCalls() recovers
  pre-expansion call signatures (relevant for has()/exists rules).
- Repo moved: github.com/google/cel-go → github.com/cel-expr/cel-go (raw master
  paths 404; pinned tags fine). ADR-004 citations remain valid (tag-anchored).

## Open questions
- None blocking. Swept-side schema checker ownership (A4 vs later candidate)
  noted by scout as undetermined — left to implementation planning, not an ADR
  question.

## Parked/deferred
- iso_duration() registry function → RFC-002-A5 (registry contents/selection/cost).
- Sweep-side schema checker implementation ownership → implementation planning.
- Probe scratch dirs left by scout at /tmp/akbcelprobe, /tmp/gotimeprobe (outside
  repo, harmless).

## Parked/deferred
- (none yet)

## Child-run index
- scout 0064dd03 (async): code recon — current coercion, BuildPage/old_page,
  write pipeline, template-write warning hooks, sweep page build, cel-go AST
  API in module cache, duration usage. Return: inline summary.
- delegate 2d51e718 (async): extraction from prior research base
  `.pi/subagents/proposals/json-cel/research/` re: date bridge, duration seam,
  static checks. Return: inline bullet extraction.
- researcher af533d6f (async): upstream facts — cel-go v0.32 duration()/timestamp()
  acceptance, JSON Schema 2020-12 format: duration spec, santhosh v6 built-in
  duration checker + RegisterFormat override, cel-go AST introspection API.
  Return: inline sourced findings. DONE — full report at
  subagent-artifacts/outputs/af533d6f-e58f-416e-b475-2fb67aa3a637/research.md.

Scout key anchors (full report: subagent-artifacts/outputs/0064dd03-…/context.md):
- convertDateField pagebuilder.go:29-44 name-agnostic top-level only; parse
  failure silently keeps string; BuildOldPage (91-116) gets identical coercion.
- Probe (go1.26.1): today's acceptance WIDER than I4 (comma fractions, 1-digit
  hour, +24:00) — kills any 'keep raw time.Parse' option.
- BuildPage lacks template param (pagebuilder.go:51); 5 call sites all hold it:
  write.go:518, append.go:210, templates_write.go:355, lint/cel.go:62,
  pagebuilder.go:115 (BuildOldPage). Signature change is mechanical.
- template write compile gate templates_write.go:113-121 = static-check hook;
  ZERO warning precedents in templates_write.go today; warning style precedent
  = template.go:200-259 (display path).
- CompileRule discards AST (engine.go:41-60); cache keyed on expr only
  (ADR-004 N4 pins). cel-go common/ast API probe-VERIFIED at v0.28:
  MatchDescendants+FunctionMatcher(timestamp) + SelectExpr-chain path walk
  extracts page.frontmatter.* paths incl. arbitrary depth + old_page root.
- duration() used only in CEL rules with Go-style strings (adr.yaml:183,190,197,
  note.yaml:34); no format:duration anywhere; evidence of the '180d' trap
  (.sisyphus/evidence/task-25).
- Sweep: lint/cel.go reuses BuildPage; old_page NOT injected at sweep; eval
  failure degrades per-page (deliberate, write.go:594-603).
