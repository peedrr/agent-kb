---
as_of: "2026-10-10"
created: "2026-10-10"
grammar: 1
informs:
- SPEC-004
- ADR-011
is_draft: false
provenance: agent-drafted
run: rfc-002-s7
scope:
- cmd/akb/**
- internal/**
- docs/knowledge-base/**
status: final
summary: "Hub log of the rfc-002-s7 run: S7 scope, key findings, owner forks OQ-1/OQ-2, decisions, and the filing record for SPEC-004 and ADR-011."
tags:
- rfc-002
- spec-004
- adr-011
- run-log
title: RFC-002-S7 run log — schema-pass integration research, SPEC-004 and ADR-011 drafting
type: research
updated: "2026-10-10"
---

## Filed pages

- [[research/rfc-002-s7-rfc-extract|rfc-002-s7-rfc-extract]] — S7 verbatim scope, SPEC inventory, deliverables
- [[research/rfc-002-s7-prior-research|rfc-002-s7-prior-research]] — 46-page prior-research coverage map
- [[research/rfc-002-s7-code-recon|rfc-002-s7-code-recon]] — pipeline attachment points, pre-gate anatomy, test fallout
- [[research/rfc-002-s7-libs-verify|rfc-002-s7-libs-verify]] — santhosh v6.0.3 API verification (source + probes)
- [[research/rfc-002-s7-spec-conventions|rfc-002-s7-spec-conventions]] — spec contract, house anatomy, report contract
- [[research/rfc-002-s7-nocommit-recon|rfc-002-s7-nocommit-recon]] — --no-commit vs merge/identity preflight behavior

Artifacts drafted by this run: [[ADR-011-validation-before-git-preflight|ADR-011]] (proposed) and [[SPEC-004-schema-pass-integration|SPEC-004]] (proposed); RFC-002 amended (§7 pipeline ordering, S7 resolves-as, A22 row).

---
# LOG — research run rfc-002-s7

## Question
Research RFC-002-S7 (from `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` + supporting docs) to enable authoring **SPEC-004** per the docs-writer skill. Existing research must not be repeated.

## Run parameters
- MODE: start · SUBJECT: RFC-002-S7 · Date: 2026-10-10
- KB: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb/docs/knowledge-base` (`research` template present)
- Owner note: user AFK — work autonomously until user decisions are the blocker, then summarize progress in chat and STOP; call `interview` only after owner's "OK".

## Agreed requirements
Research the six S7 work items (RFC-002:834) to drafting readiness for SPEC-004:
1. Document construction (build step before schema pass; ADR-008 current shape: frontmatter raw, content{raw,word_count,char_count}, ast{headings,links,code_blocks}, akb{...}, file{...})
2. Pipeline placement (RFC §7:331-338: merge-conflict → selector gate → build document → JSON Schema → CEL → write; no mutation step)
3. Compiled-schema cache (ADR-003 I8: compile once, keyed by template identity, RegisterFormat before compile; I2 AssertFormat write+sweep; compiler not thread-safe)
4. Error-tree → SPEC-002 report mapping (schema short-circuits CEL; keyword + JSON Pointer rows, no `line`; schema rows first; santhosh v6 tree shape needs re-verification vs pinned version)
5. Pre-gate removal (`cmd/akb/required_fields.go:18`; write-path only — sweep keeps required_fields per ADR-009 I5)
6. Sweep checker (ADR-009 I1-I6: before cel_lint, shared now, degrade-to-issue, LintIssue shape, by_check stability)
Plus: SPEC-004 page contract (spec template, REQ/AC conventions mirroring SPEC-002/003).
Boundary guards: SPEC-003 REQ-007 — loader passes schema UNCOMPILED (compile-at-load rejected, do not re-propose unless pass-through proven unworkable). S8 owns FAILS: parser/mockup proofs. ADR-008 forbids flatten/word_count/write --json in Phase 1.
Scope is RFC-defined → AGREED without owner fork (owner AFK; no plural-valid scope options).

## Decisions made
- Run slug: `rfc-002-s7`; run dir `.pi/research/rfc-002-s7/`.
- Research scope AGREED without owner fork (RFC row + ratified ADRs fully determine it).
- D-F1: package home = new `internal/schema` (low stakes, one right way).
- D-F2: one mutex-guarded shared cache, write path + sweep (ADR-003 I8 reading).
- D-F3: sweep fault isolation = per-page degrade (ADR-009 I3/N3 already fixes it — not a fork).
- D-F4: approve wording parity = SPEC-004 REQ-015 (ADR-009 consequence, test-pinned).
- D-F6: schema short-circuit is ratified (SPEC-002 REQ-002); `exit_codes.txt:41`/`-count=6` churn is test work, not a fork.
- D-F8: row messages via library rendering (BasicOutput-style), no direct x/text dep.
- SPEC-004 drafted → `.pi/research/rfc-002-s7/SPEC-004-draft.md` (status proposed, 22 REQs, 8 ACs, ~280 lines).

## STATE (2026-10-10, owner engaged)
Research COMPLETE (5/5 child reports). SPEC-004 draft v2 complete. ADR-011 drafted.
- OQ-1 RESOLVED: recon (`nocommit-recon.md`) showed `--no-commit` skips ONLY identity; merge check ALWAYS runs in git mode (mode.go:115 before :118; rechecked at git.go:56). Owner branch (b) applied: parked out of SPEC-004. Owner decisions: (i) ALL mutating commands should order validate → git-state preflight (agents move fast; merges clear while validation is fixed); (ii) `--no-commit` should bypass ALL git-state checks (merge + identity); (iii) author ADR (rolled into RFC-002 + post-RFC-002 work unit). Sequencing owner-chose: ADR now → OQ-2 → one filing wave. Coordinator pushback on (ii) — working-tree entanglement — offered and rejected (git never sweeps unstaged paths into a merge commit); recorded in ADR-011 alternatives.
- ADR-011 drafted → `ADR-011-draft.md` (status proposed): validate-first ordering I1-I2, --no-commit full bypass I3/N2, safety-net preflight I4, init excepted.
- SPEC-004 draft v2: REQ-007 (early preflight) REMOVED; REQs renumbered (now REQ-001..021); OQ-1 marked RESOLVED w/ ADR-011 pointer; Preserved Behaviour + Non-Goals + Drift Ledger updated; Store.Write anchor dropped.
- OQ-2 RESOLVED (owner, 2026-10-10): Option A — malformed template schema = exit 2 `internal:` fault on write path + per-page degrade in sweep. Owner's condition #1 confirmed correct (inter-phase HEAD risk acceptable; S8 is Phase 1 so window is small); condition #2 corrected (CEL mockups exercise at template-write; schema block NOT exercised there until SPEC-005) — owner ratified Option A after the correction. REQ-021 finalized + sweep-degrade sentence added.
- FILING WAVE STAGING (all checks done): `akb template get rfc`/`adr` contracts retrieved; 3 frontmatter blocks prepared (libs-verify.fm, spec-conventions.fm, nocommit-recon.fm); rfc-extract/prior-research/code-recon carry valid frontmatter (prior-research gained created/updated; code-recon summary trimmed to 192); drafts verified free of `.pi/` citations; approve each new page with --no-commit during wave.
- `rfc-002-prep` child (66fbbfe2, async): preparing rfc-002-updated.md — 5 edits: (A) §7 pipeline text → git-state preflight immediately before storage write, note ADR-011; (B) S7 row resolves-as → SPEC-004; (C) new row RFC-002-A22 → ADR-011 work unit; (D) D-register entry if present; (E) updated bump.
- FILING ORDER (when child lands): research/rfc-002-s7-{rfc-extract,prior-research,code-recon,libs-verify,spec-conventions,nocommit-recon}.md → research/rfc-002-s7-log.md (hub w/ wikilinks) → decisions/ADR-011-validation-before-git-preflight.md → RFC-002 overwrite → SPEC-004 LAST (no --no-commit = the wave's one commit). Then akb lint, resolve, request ratification of SPEC-004 + ADR-011 (both proposed; owner-only move to active/accepted).

## Key findings
- **S7 definition** (RFC-002 §18 row, `rfcs/RFC-002-open-validation-engine.md:834`): "Schema-pass integration: document construction, pipeline placement, compiled-schema cache, error-tree → SPEC-002 report mapping, pre-gate removal, sweep checker (ADR-009) — drafts as SPEC-004." [via prior-research.md]
- **Governing decisions**: ADR-003 (santhosh v6, format assertion), ADR-008 (Phase-1 validates today's page-map shape; S7 bound), ADR-009 (sweep-time schema checker, before cel_lint, shared `now`; required_fields kept — S7 owns), ADR-010 (TemplateV3 six keys). SPEC-002 = report contract; SPEC-003 = TemplateV3 schema block S7 compiles/caches/passes. [prior-research.md]
- **46 prior research pages mapped**; deep coverage already exists for write-path internals, CEL engine, lint engine, TemplateV2→V3 loader, santhosh v6 survey, SPEC-002 report model, document model. Do NOT re-research these. [prior-research.md]
- **S7 research gaps** (the fanout targets) [prior-research.md §Gaps]:
  1. Compiled-schema cache design (key/lifetime/invalidation, concurrency vs single-DB rule, compile site S6-vs-S7 boundary — open per `rfc-002-s6-prior-research.md:151-155`, `rfc-002-s6-log.md:41` F2)
  2. Error-tree → SPEC-002 report mapping (santhosh v6 `*ValidationError` tree → report)
  3. Concrete pipeline placement of schema pass vs required-gate/CEL under Phase-1 shape
  4. Sweep-time schema checker internals (compile reuse, issue mapping, by_check stability)
  5. Pre-gate removal mechanics (required_fields kept per ADR-009 vs JSON-Schema pre-gate removal; ordering/double-reporting)
  6. Dogfood conversion + test fallout for schema pass
  7. Malformed-schema error surface ownership (S6 load vs S7 compile) — explicitly undrawn

## Open questions (candidate decision forks for owner, post-research)
- F1: Go package home for combined report type + compiled-schema cache (SPEC-002 OQ :274-280; `internal/cel` is CEL-only). DECIDED (low stakes, one right way): new `internal/schema` package — internal/template is forbidden the compiler dep (SPEC-003 AC-007), internal/cel is CEL-only.
- F7 (new, from libs-verify): malformed `schema:` block surfacing — loader passes through uncompiled (SPEC-003 REQ-007), so first compile happens in SPEC-004's cache. Where does a malformed schema error: template write proof, write/append (exit 2 internal vs exit 1), sweep (per-page issue)? Compile faults during validation are engine faults not page faults (cf. cel_lint compile-error aborts sweep; write path exit 2 via internalError). PLURAL-VALID → owner.
- F8 (new, from libs-verify): message rendering for schema rows — no library Message field; (a) BasicOutput().Error.String() — no new direct dep; (b) direct golang.org/x/text/message dep + LocalizedString(printer). DECIDED (low stakes): (a) BasicOutput, single-line-normalized; avoids new dep, library already renders.
- F2: Compiler/cache sharing between write path and sweep checker (one shared cache keyed by template identity vs per-path instances; ADR-003 I8 says compile once keyed by template identity → shared cache is the natural reading; concurrency constraint: santhosh Compile not thread-safe, Validate-on-compiled undocumented)
- F3: Sweep checker fault-isolation depth (ADR-009 defers to "section-11 Phase-A scope"; minimal = degrade per page per I3/N3)
- F4: approve refusal-wording parity hardening (approve.go:221 lacks testscript guard; ADR-009 requires schema pass to emit the same `missing required frontmatter field(s)` sentence → likely SPEC-004 requirement, low stakes)
- F5 → became OQ-1 in the draft: SPEC-004 owns the preflight reorder (explicit store.Preflight() before validation, matching RFC §7 + all other mutating commands) vs leave preflight inside store.Write (validation-first precedence). Recommendation: reorder. OWNER FORK.
- F7 → became OQ-2 in the draft: malformed template schema = internal fault exit 2 (mirrors CEL compile-fault; recommended) vs exit 1 validation-style; sweep-side: degrade per page (recommended) vs abort sweep (cel_lint precedent). OWNER FORK.
- F6 (from code-recon G5): schema short-circuit breaks `exit_codes.txt:41` (-count=6) and `superseded_adr.txt:65` — short-circuit ALREADY ratified (SPEC-002 REQ-002) → test churn, not a fork. CLOSED.

## Parked/deferred
- S8 owns: FAILS: parser, mockup-vs-schema proofs, overlap warning, ADR-005 static-check wiring (RFC:835; SPEC-003:314-316).
- `-race` verification of concurrent Validate (no gcc/CGO locally) — SPEC-004 verification plan should include it.
- YAML shape survival through goccy YAMLToJSON for the SCHEMA document (anchors, !!timestamp, non-string keys) — flagged by libs-verify gap 5; the schema block arrives PARSED from the loader (SPEC-003), so route is re-marshal-to-JSON, not YAMLToJSON of source text. Confirm at authoring.
- libs-verify probe harness at /tmp/jsv6smoke (throwaway, not committed).

## Child-run index
- `rfc-extract` (scout, DONE): S7 scope + SPEC-004 context → `.pi/research/rfc-002-s7/rfc-extract.md`
- `prior-research` (scout, DONE): coverage map → `.pi/research/rfc-002-s7/prior-research.md`
- `code-recon` (scout 74549ac1, DONE): → `code-recon.md` (child self-repaired the malformed binding; full 267-line report at the intended path). Key yield: insertion points write.go:518-528 / append.go:210-224; pre-gate anatomy + 4 call sites; cel_lint as checker model (registration lint.go:157-167 before :166); test fallout 11 Go funcs + 8 testscripts. **New deltas**: D5 merge/identity preflight runs INSIDE store.Write (after validation) — RFC §7 ordering needs explicit store.Preflight() added; D6 page["frontmatter"] is date-coerced — schema document must be built from fm.Fields WITHOUT convertDateField; D7 no shared-now mechanism exists (ADR-009 I2); D8 santhosh dep net-new in go.mod.
- `libs-verify` (scout 58c9504c, DONE): → `libs-verify.md`. Verified against v6.0.3 module source + runtime probes (/tmp/jsv6smoke). Key yield: SPEC-002 assumed error-tree shape VERIFIED verbatim (validator.go:960-972); corrections — 39 kind types (not ~60), NO Message field (LocalizedString(nil) PANICS), no exported JSON-Pointer helper (jsonPtr/escape private; escape order ~ then /); Validate accepts ONLY JSON-shaped Go values (all int/uint/float OK — goccy numerics need NO normalization; time.Time/RawMessage REJECTED → confirms raw-document requirement D6); Compile cached per Compiler (same pointer), Compiler goroutine-safety undocumented (shared-map writes evidence unsafe), concurrent Validate on one *Schema probably safe (no -race available); AssertFormat mandatory for 2020-12; unknown format names silently unasserted; malformed-schema classifiable via typed errors (*SchemaValidationError etc.); ADR-003 correction: thread-unsafety is NOT documented (zero concurrency text in module).
- `nocommit-recon` (scout 5a30ee6e, DONE): → `nocommit-recon.md`. Verdict: --no-commit skips ONLY identity (mode.go:118); merge check always runs in git mode (mode.go:115 + git.go:56 double check); ModeNone bases fully git-free (independent mechanism); NO test pins --no-commit+merge (gap → ADR-011 verification names it).
- `spec-conventions` (scout b3c71df2, DONE): → `spec-conventions.md`. Key yield: spec.yaml contract (required keys, CEL rules, required H2s, challenge affordance, `Do not re-propose unless` clause); SPEC-003 anatomy to clone; SPEC-002 report contract quoted (schema row = `{source:schema, keyword, pointer, message}` NO line/rule_id; short-circuit REQ-002; ordering REQ-006); SPEC-004 starts `status: proposed`; cite tracked `kb/research/` wikilinks never `.pi/` (no_external_citations lint).
