---
as_of: "2026-10-09"
created: "2026-10-09"
grammar: 1
informs:
- ADR-008
- ADR-009
- ADR-010
is_draft: false
provenance: agent-drafted
run: rfc-002-phase1-scope
scope:
- internal/**
- cmd/**
- test/**
- docs/knowledge-base/**
status: final
summary: Run log for the RFC-002 Phase-1 scope research — four owner-ratified scope decisions filed as ADR-008/009/010, with the evidence index, decision record, and parked items.
tags:
- rfc-002
- phase-1
- research-log
title: Phase-1 scope research run — LOG
type: research
updated: "2026-10-09"
---

# LOG — rfc-002-phase1-scope

## Question

Resolve the RFC-002 Phase-1 scope decisions that gate the Phase-1 tech-specs
(SPEC-003 V3 format, SPEC-004 schema-pass integration, SPEC-005 template-write
obligations — none yet drafted, no §18 candidate rows):

1. **Document shape for the Phase-1 schema pass + the `word_count` gap.**
   RFC-002 §5's flattened document (`content` string, `page.derived.ast.*`) and
   §5.3's CEL renames "ride the V3 revision" (Phase 1), but the document-builder
   rewrite is candidate A6 (Phase-3 parser work) and `word_count()` the registry
   function is Phase 4 (A5 unratified) — while §5.1 scraps materialized
   `word_count`/`char_count`. Flatten in Phase 1 ⇒ embedded templates'
   `page.content.word_count` rules break. Decide: flatten timing + word_count /
   char_count disposition during Phase 1.
2. **Sweep-time schema checker placement.** §17 Phase 1 omits it; §11 assigns
   the `schema` checker to lint (Phase-3 registry era); but ADR-003's blocking
   verification gate ("sweep-time schema checker flags the same page") and the
   `required_fields` subsumption row assume it exists when ADR-003 ships.
   Decide: Phase 1 as built-in vs Phase 3.
3. **Minimal V3 key set + forward-compat.** Which keys Phase-1 V3 carries
   (name/description/schema/validations/lint_rules?) vs deferred
   (path→P3, state_field→P2/A15, functions→P4/A5, projections→A17,
   title_field/summary_field→A13); what strict unknown-key rejection does with
   deferred keys in the interim.
4. (Candidate, pending owner confirmation) **SPEC-002 emission point.** REQ-005/
   AC-005..009 need `write --json` (Phase-2 surface). Decide: minimal
   failure-only `write --json` lands in Phase 1 vs those ACs defer to Phase 2.

Target artifact: Phase-1 scope ADR (new RFC-002 candidate → next free global
ADR number), per prior session agreement — confirm with owner at requirements
step / decision fork.

## Run parameters

- MODE: start
- Date: 2026-10-09
- KB: /home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb/docs/knowledge-base
  (AKB_KB set; `research` template verified present)
- Instructions: "Go ahead with researching the three scope questions."
- Run dir: .pi/research/rfc-002-phase1-scope/

## Agreed requirements

(Owner answers 2026-10-09, ask_user_question)

- Scope: FOUR questions — the three original + Q4 SPEC-002 emission point
  (minimal failure-only `write --json` in Phase 1 vs defer REQ-005/AC-005–009
  to Phase 2).
- Sources: INTERNAL ONLY — code + KB docs; no external re-verification
  (ADR-003/004/005 runs already verified upstream 2026-10-06).
- Artifact: DECIDE AFTER RESEARCH LANDS, on coordinator recommendation.
  Owner note verbatim: "ADRs *should* record just a single decision for
  future reference, however if any decisions are tightly coupled then they
  can be bundled. We'll revisit once all research is in and you make a
  recommendation."

## Decisions made

(Owner-ratified 2026-10-09 via interview form "RFC-002 Phase-1 scope
 decisions" — all four picked option A, the coordinator recommendation)

- **D-scope-1 (Q1):** Phase 1 validates the CURRENT document shape
  (`frontmatter.*`, `content.raw`, `ast.*`, `file.*`); the §5 flatten,
  `derived.*` renames, and `word_count()` registry all ship later with the
  flatten's natural cohort (Phase-2 JSON I/O or the A6/A17 parser phase).
  Requires amending RFC-002 §5.3's "renames ride the V3 revision" line
  (RFC still draft — legal). Why: ~zero churn vs ~27 test funcs + 4
  testscripts + ~30 rules + A5 ratification prerequisite; no Phase-1
  requirement needs the flatten.
- **D-scope-2 (Q2):** the sweep-time schema checker ships in PHASE 1 as a
  built-in (before cel_lint, shared `now`); lint `required_fields` checker
  stays until Phase 3's registry subsumes it with the keep-keys mechanism.
  Interim duplicate reporting accepted as documented, visible P3 tension.
  Why: ADR-003's blocking sweep gate + P7 net; by_check never loses a key.
- **D-scope-3 (Q3):** Phase-1 TemplateV3 recognizes exactly {name,
  description, dir, schema, validations, lint_rules}; any other key is
  HARD-REJECTED at `template write` with a self-sufficient message, warn-only
  at load (forward-compat). `dir` keeps current meaning; `path` rename rides
  Phase 3. Why: recognize-and-ignore would mint inert declarations — the D9
  trap species.
- **D-scope-4 (Q4):** `write --json` defers to Phase 2 entirely; Phase 1
  lands SPEC-002's report internals only (AC-001/004/006 close;
  AC-005–009 land with Phase 2's --json, exactly as SPEC-002's depends_on
  declares). Why: no envelope frozen before the flatten settles shape;
  JSON→YAML landmine stays Phase-2 scope. Coupled to D-scope-1 (the
  envelope/document divergence was the interaction term).
- **D-artifact (2026-10-09, interview):** THREE ADRs — D-scope-1+4 bundled
  (coupled via shape/envelope interaction) as candidate RFC-002-A19 →
  ADR-008; D-scope-2 standalone as A20 → ADR-009; D-scope-3 standalone as
  A21 → ADR-010. Owner ratified the recommendation.

## Key findings

Recon (recon-code-surface.md, all grep-verified, high confidence):

- Page map TODAY: `page.content` = OBJECT `{raw, word_count, char_count}`
  (pagebuilder.go:69-73), `page.ast` = {headings,links,code_blocks} (:75-79),
  `page.akb` = {provenance_markers, annotations} (:81-83); `page.file` and
  `page.frontmatter` exist TODAY (:54-66). BuildOldPage delegates to BuildPage
  (:116) — old_page moves with any shape change.
- `char_count`: ZERO rule consumers anywhere → safe scrap (matches D13).
- `word_count`: exactly 5 template consumers, all `page.content.word_count >= 50`
  (embedded adr.yaml:118; dogfood adr:210, rfc:309, spec:247, research:189).
  No Go consumer.
- `page.content.raw`: 8 dogfood rules; no embedded rules. `page.ast.*`:
  4 embedded + ~18 dogfood rules. `page.akb.*`: ZERO rule consumers.
  Total rename blast radius under flatten: ~30 rules, all mechanical.
- Required-field pre-gate reads ONLY `tmpl.Schema.Frontmatter` map
  (required_fields.go:19-29); if that block vanishes it SILENTLY NO-OPS and
  unguarded `page.frontmatter.<required>` CEL reads become eval errors
  (write fails closed confusingly, lint degrades). Replacement required-list
  must ship with any flatten — the guard invariant is load-bearing.
- Lint: new built-in checker needs only Name()+Check(ctx,*KB)
  (engine.go:37-40) + one AddChecker line (lint.go:157-167). PageData carries
  NO derived content/ast — a schema checker must rebuild via cel.BuildPage
  (pattern: internal/lint/cel.go:60-62, the 2nd production BuildPage caller).
- Loader is LENIENT (template.go:106,114 — no DisallowUnknownField): a new
  `schema:` block would load and be silently dropped today. detectOldFormat
  exists in TWO near-duplicate copies (template.go:73-78,
  templates_write.go:461-469) — ADR-006 I3's single-helper mandate targets this.
- FieldSchema.Type is declared but NEVER enforced; only Required (Go gate)
  and Enum (via CEL rules) are live. `grammar` key is page-frontmatter, not
  template-level.
- No --json on write or read today (write.go:69-73 registers only
  --append/--dated/--frontmatter; read.go has none). --json pattern exists on
  discover/links/lint/list/raw_status/search.

Recon wave 2 (recon-test-pinning.md, grep-verified):

- Flatten-now test churn: ~27 Go test funcs (21 in pagebuilder_test.go) + 4
  testscript assertions (cel_validation.txt:167, wikilink_forms.txt:146,
  exit_codes.txt:40, superseded_adr.txt:65) + 5 template min_word_count rule
  re-paths + 8 production BuildPage/BuildOldPage call sites.
  Validate-current-shape churn: ≈0 (new tests only). Required-fields wording
  (10 funcs / 2 testscripts) and V2 old-format guard are ORTHOGONAL to shape.
- page.akb pinned by only 2 structural asserts; zero rule consumers.
- Frontmatter parse types (goccy v1.19.2, empirically verified): date-like
  scalars stay STRING (only explicit !!timestamp → time.Time); ints
  uint64/int64; maps map[string]any (never map[interface{}]interface{}); lists
  []any. ⇒ schema document view needs VALUE normalization awareness (date
  formats, int signedness) but NO structural normalization pass.
- BuildPage/BuildOldPage are the sole chokepoints (pagebuilder.go:51,91);
  wrappers: buildTestPage, evaluateValidations, runTemplateValidations.
- COVERAGE GAP: approve refusal-wording parity (approve.go:221) has NO
  testscript guard — only approve_test.go:651,694.

## Open questions

- The four scope questions — AT DECISION FORK with owner 2026-10-09
  (interview, per decision-forms skill). Coordinator recommendations:
  Q1 validate-current-shape (strong); Q2 Phase-1 sweep checker +
  required_fields retained till Phase 3; Q3 strict minimal V3 key set;
  Q4 defer write --json to Phase 2.

## Parked / deferred

- Drafting of SPEC-003/004/005 themselves — later runs; this run only resolves
  the scope decisions that gate them.
- RFC-002 §18 table amendment (new candidate rows for the three tech-specs) —
  mechanics deferred to artifact filing.
- RFC-002 body amendments implied by D-scope-1 (§5.3 rename sentence) and
  D-scope-2 (§17 Phase-1 content gains the sweep schema checker) — ride the
  artifact ratification wave per §18 lifecycle.
- Recon-noted coverage gap: approve refusal-wording parity (approve.go:221)
  has no testscript guard, only approve_test.go:651,694 — candidate
  hardening note for SPEC-004.

## Child-run index

- recon-code-surface (scout, COMPLETE) — code surface map;
  filed as [[rfc-002-phase1-scope-code-surface]]
  (scratch: .pi/research/rfc-002-phase1-scope/recon-code-surface.md)
- recon-test-pinning (scout, COMPLETE) — test/testscript pinning inventory +
  frontmatter parse types + BuildPage call sites;
  filed as [[rfc-002-phase1-scope-test-pinning]]
  (scratch: .pi/research/rfc-002-phase1-scope/recon-test-pinning.md)
