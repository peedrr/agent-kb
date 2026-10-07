---
grammar: 1
type: research
title: "RFC-002-S1 research log — structured validation report (SPEC-002)"
status: final
provenance: agent-drafted
run: rfc-002-s1
as_of: "2026-10-07"
created: "2026-10-07"
updated: "2026-10-07"
scope:
  - "cmd/akb/**"
  - "internal/cel/**"
  - "internal/lint/**"
tags:
  - validation-report
  - spec-002
  - decision-log
  - research-log
summary: "Run log for RFC-002-S1: SPEC-002 scope, format authority, decision-fork strategy, and the child-run index for the structured validation report spec."
informs:
  - SPEC-002
---
# LOG — RFC-002-S1

## Question
Produce the spec required by RFC-002 (open validation engine) slice S1, per the
docs-writer skill, from the RFC + ADRs 001-006 in docs/knowledge-base/kb/decisions,
and commit it to the KB via `akb` (AKB_KB set).

## Run parameters
- MODE: start
- SUBJECT: RFC-002-S1
- Date: 2026-07-13 (session date; verify)
- INSTRUCTIONS: consume RFC-002 in full; ADRs written; prior research in
  .pi/research/rfc-002-a{1..4}* still in place; create required spec per
  docs-writer; commit via akb; approval questions via decision-forms.

## Decisions made
- RFC fully consumed by coordinator (903 lines) — owner instruction explicitly granted this.
- RFC-002-S1 = SPEC candidate "Structured validation report (widened runTemplateValidations)";
  referenced in RFC §7 (error merge carries from RFC-001) and §18 (status: proposed).
- Drafts as SPEC-002 (next free global number; SPEC-001-init-versioning exists in kb/specs/).
  Slug: structured-validation-report. Header records Origin: RFC-002 candidate RFC-002-S1.
- Status will be `proposed`, provenance `agent-drafted` (docs-writer: draft, don't ratify).
- Format authority: docs-writer references/spec.md (read in full) + KB `spec` template
  (.agent-kb/templates/spec.yaml — required frontmatter: grammar,type,id,title,kind,status,
  provenance,created,updated,scope,verified; CEL gates incl. challenge affordance, REQ/AC IDs,
  no [NEEDS CLARIFICATION], code blocks need language tags, >=50 words).
- First launch attempt malformed (tool-call corruption); no children ran; relaunched clean.
- COMMIT STRATEGY (owner-raised): draft SPEC-002 in run dir; single `akb write --no-commit`
  (+ index/log mutations also --no-commit if conventions require them); owner reviews in place;
  owner runs one plain-git commit bundling all KB changes. Rationale: ADR-era churn came from
  repeated committed writes; validation failures never commit (fail closed). No `akb commit`
  command exists — final commit is plain git, surfaced as a command for the owner.
- ADR-001/002 predate RFC-002 (akb-doctor, init-author-flags) — not S1 inputs; governing ADRs
  for S1 are 003/004/005/006.

## Key findings (code-recon, anchors in code-recon.md)
- `runTemplateValidations` cmd/akb/write.go:603 — collects ALL false/unevaluable rules as
  cel.ValidationError{RuleID,Message,Line:0,Severity:error}; prints each to stderr, returns
  bare `validationFailure{}` sentinel. FIRST compile error aborts as internalError (exit 2).
- Call sites: write.go:529, append.go:225 ONLY. template write uses separate
  `evaluateValidations` (templates_write.go:363) returning bare rule-ID strings — divergence
  the spec must address explicitly.
- NO machine-readable write-failure output today (confirmed; no JSON tags anywhere on path).
- lint --json envelope (engine.go:17,28 + lint.go:220 jsonOutput{issues,summary{total,
  pages_checked,by_check}}) is the ONLY structured-report precedent.
- Load-bearing: validationFailure sentinel matched by classifyExit (exit 1, silent) AND
  approve.go:265 errors.As; print-then-return contract (classifyExit reports nothing more);
  fault/result split (CEL engine faults = exit 2, content failures = exit 1); lint/write
  divergence on unevaluable rules is deliberate (write.go:598-602 comment).
- `<!-- FAILS: rule_id -->` convention is DOCUMENTARY ONLY — zero Go parses it (grep=0 hits);
  extending it to `FAILS: schema:` is net-new machinery.
- cel.ValidationError{RuleID,Message,Line,Severity} errors.go:11; Line/Severity unused at
  write time (0/"error" always).

## Coordinator decisions (one-right-way; logged, not escalated)
- Envelope is akb-OWNED (walk santhosh Causes; do NOT reuse library OutputUnit — quirky
  capitalized tag, and akb needs N5-style freeze control). Shape modeled on lint conventions:
  stdout JSON, empty arrays as [], indented encoder.
- All-errors aggregation preserved (current behavior + RFC-001 "merge into one report").
- Schema rows: keyword + JSON Pointer (instanceLocation) + message. CEL rows: rule_id +
  message; `line` optional (omitted when 0/absent; schema rows never carry it).
- severity field carried on rows, always "error" at write time (fail-closed); documents the
  mapping (schema violations + CEL validation failures are error-severity by definition).
- approve: no new surface; widened internals MUST preserve validationFailure sentinel
  (errors.As at approve.go:265 + classifyExit silent-exit-1) and print-then-return contract.
- Exit codes unchanged: content failure=1, engine fault=2, usage=2.
- Messages single-line; cel-go message text never pinned (ADR-004:108).
- verified.state: unverified + method manual-review for the proposed draft (SPEC-000 mockup
  precedent); anchor states unverified pending a verification sweep.
- index.md/log.md entries via `akb index add` / `akb log` (never direct writes), --no-commit.

## Key findings (research-extract + adr-conventions)
- RFC-001 §7:231-234 (verbatim): schema violations (keyword + JSON Pointer) + CEL failures
  (rule ID) merge into one report, one exit-1; human stderr stays; structured form requires
  widening runTemplateValidations. §11:329: report on STDOUT under `write --json` failure.
  :373-374 S1 charter: widening + merged model + stable IDs/pointers + stdout/stderr split.
- N5: lint --json envelope FROZEN — S1 must not touch it. akb JSON convention: results stdout,
  failures human stderr, empty collections [] not null.
- santhosh v6.0.3 error tree (freshly verified from module cache, NOT in prior research):
  ValidationError{SchemaURL, InstanceLocation []string, ErrorKind, Causes}; Validate returns
  synthetic root whose Causes are the real errors; kinds public (kind/kind.go, ~60 types);
  Basic/Detailed/Flag output exist (quirky) — own envelope via Causes walk.
- ADR-004:108 — cel-go error strings churn on upgrade; never pin as contract.
- ADR-005 I10/I11 — message-carries-rule-ID+path+remediation precedent. ADR-006 — rejection
  messages self-sufficient, single shared message text, exit 1.
- Origin convention: intro paragraph after the For-agents blockquote —
  "Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-S1 ..." (+ rfc-002 tag
  tolerated by non-strict loader; spec.yaml declares no tags key; SPEC-001 carries none).
- SPEC-001 shape: 253 lines, no body metadata table; sections: Why, Goals and Non-Goals,
  Current Behaviour, Requirements, Acceptance Criteria, Contract and Invariants, Preserved
  Behaviour, Decisions and Rejected Alternatives, Assumptions and Open Questions, Affected
  Surface and Ordering (+ bare depends_on: []), Verification Plan, Drift Ledger, Revisit
  Triggers, Evidence Appendix. Wikilinks [[ADR-003-...|ADR-003]].
- index.md `## Specs` entry = title link + em-dash description; log.md `## <date> <verb>`.

## Key findings
- Prior research dirs: .pi/research/rfc-002-a1-validator, -a2, -a3-celgo, -a4
- ADRs: ADR-001 akb-doctor, ADR-002 init-author-flags, ADR-003 santhosh-v6-format-assertion,
  ADR-004 cel-go-upgrade-pinned-extensions, ADR-005 date-bridge-determinism,
  ADR-006 schema-frontmatter-retirement
- RFC: docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md (903 lines)

## Open questions
- None blocking. Non-blocking: whether approve ever gains a --json surface; report type's Go
  package home (implementation choice, recorded in the SPEC's Open Questions).

## Parked/deferred
- Mockup attribution (`<!-- FAILS: schema: ... -->` parser + template-write mockup report):
  owner-excluded from SPEC-002 2026-10-07; rides Phase-1 "template write schema obligations"
  (ADR-006 N7 defers the parser). Recorded in SPEC-002 Decisions with do-not-re-propose clause.
- RFC-002 §18 S1 row amendment (proposed → ratified, Resolves-as SPEC-002 + body reference
  rewrite): HUMAN step at ratification, per §18 lifecycle. Also fixes the page's orphan
  warning (the RFC becomes its inbound link).
- SPEC-002 length 348 lines vs docs-writer's ≤250 design default (SPEC-001: 253): flagged to
  owner at review; trimming deferred to ratification feedback rather than cut unilaterally.
- Owner's interview note (RFC-001 references in fork framing): answered in chat — RFC-002 §3/
  §7/Appendix B carry the merged error model by name; RFC-001 cited only as carried text;
  phases cited are RFC-002 §17's; ADR-003..006 (accepted) prove the chain. No residual doubt.

## Owner decisions (decision form, 2026-10-07, both as recommended)
- SURFACE: RFC-literal — structured report on stdout ONLY under `akb write --json` failure
  (exit 1); markdown-mode write/append/approve keep human stderr only; no new flags.
- SCOPE: write-path report only; mockup attribution excluded (see Parked).

## Outcome (run complete pending owner actions)
- Draft: .pi/research/rfc-002-s1/SPEC-002-draft.md → written to KB as
  kb/specs/SPEC-002-structured-validation-report.md via `akb write --no-commit` (passed all
  spec-template CEL gates first attempt); index + log entries added (`akb index add` /
  `akb log append plan`, both --no-commit). All three files staged, UNCOMMITTED.
- `akb lint` exit 0; new page carries an orphan warning (no inbound links) until RFC-002
  links it at ratification.
- Frontmatter: status proposed, provenance agent-drafted, verified.state unverified (HEAD
  fc08a201, main), 3 anchors (runTemplateValidations / cel.ValidationError / classifyExit).
- No is_draft injected by the write path; page is not KB-draft.
- AWAITING OWNER: (1) review the page; (2) commit:
  `git -C docs/knowledge-base commit -m "akb: write kb/specs/SPEC-002-structured-validation-report.md"`
  (3) ratify: proposed → active + ratified_by + verification sweep (verified.state → live);
  (4) amend RFC-002 §18 S1 row.

## Child-run index
- Workflow 2279a8a6 (async, 2026-10-07): three scouts, outputs bound into run dir:
  - code-recon → code-recon.md: current runTemplateValidations/ValidationError/write-path/
    classifyExit/template-write-mockup/lint --json machinery, file:line anchors.
  - research-extract → research-extract.md: RFC-001 merged-error-model verbatim reqs;
    santhosh v6 error types (a1 research); a2/a3/a4 error-bearing findings; json-cel research.
  - adr-conventions → adr-conventions.md: ADR-003..006 decisions + Origin-recording convention;
    SPEC-001 structural conventions; index.md/log.md entry style.
