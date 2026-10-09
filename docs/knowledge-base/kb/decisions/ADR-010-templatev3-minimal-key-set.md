---
created: "2026-10-09"
deciders:
- Pete Hope
grammar: 1
id: ADR-010
is_draft: false
provenance: agent-drafted
revisit:
- a later phase ships its template keys early or late — extend I1's set and the I2 message at that phase, not ad hoc
- the load-time warning proves invisible in practice (authors miss it) — re-evaluate load-time severity
- dir gains consumers that conflict with the Phase-3 rename to path — re-sequence the rename
scope:
- internal/template/**
- cmd/akb/**
- docs/knowledge-base/.agent-kb/templates/**
- internal/skill/embedded/**
status: accepted
summary: Phase-1 TemplateV3 recognizes exactly name, description, dir, schema, validations, lint_rules; template write hard-rejects any other key with a self-sufficient message, and load warns only.
tags:
- rfc-002
- phase-1
- template-format
- validation
title: Phase-1 TemplateV3 Recognizes Six Keys and Hard-Rejects the Rest
type: adr
updated: "2026-10-09"
---

# ADR-010: Phase-1 TemplateV3 Recognizes Six Keys and Hard-Rejects the Rest

> In the context of [[RFC-002-open-validation-engine|RFC-002]]'s TemplateV3 replacing the homegrown `schema.frontmatter` block, facing a full V3 vocabulary (`path`, `state_field`, `functions`, `projections`, `title_field`, `summary_field`, `strip_on_approve`) whose semantics belong to later phases and a loader that today silently ignores unknown keys, we decided that Phase-1 TemplateV3 recognizes exactly `name`, `description`, `dir`, `schema`, `validations`, and `lint_rules`, hard-rejecting any other top-level key at `template write` with a self-sufficient message while the load path only warns, and neglected recognize-and-ignore for deferred keys, renaming `dir` to `path` in Phase 1, and load-time strictness, to achieve a template format with no inert declarations, accepting that authors reading RFC §6 must wait for later phases to use its full vocabulary, because an unrecognized key that silently does nothing is the D9 trap species this RFC exists to eliminate.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

Phase-1 TemplateV3 is a strict, minimal format: the loader and `akb template write` recognize exactly six top-level keys — `name`, `description`, `dir`, `schema`, `validations`, `lint_rules` — and every other key is refused at authoring time with a message that names the offending key and the supported set, because a template key akb cannot honor must fail loudly at the door the author controls, while the load path warns only so a newer template never bricks an older binary.

## Invariants

- **I1**: The Phase-1 template loader MUST recognize exactly the top-level keys `name`, `description`, `dir`, `schema`, `validations`, and `lint_rules`.
- **I2**: WHEN `akb template write` encounters any other top-level key, the system MUST refuse the template with exit 1 and a message naming the unrecognized key and the supported set of I1.
- **I3**: WHEN the load path encounters an unrecognized top-level key in an otherwise-V3 template, the system MUST warn on stderr naming the key and MUST continue loading the recognized keys.
- **I4**: The `dir` key MUST keep its current semantics throughout Phase 1; the rename to `path` and its resolution-hint semantics (RFC-002 §8.3, D11) ride Phase 3.
- **I5**: WHEN a template file carries the retired `schema.frontmatter` shape, the system MUST apply [[ADR-006-schema-frontmatter-retirement|ADR-006]]'s refusal before any unknown-key judgment (ADR-006 I5's ordering).
- **I6**: The `schema` key's value MUST be a JSON Schema 2020-12 document over the current document shape, per [[ADR-008-phase1-current-document-shape|ADR-008]] I1/I2.

## Negative Constraints

- **N1** (MUST NOT · scope: `internal/template/**`, `cmd/akb/**`): The system MUST NOT recognize-then-ignore the deferred keys (`path`, `state_field`, `functions`, `projections`, `title_field`, `summary_field`, `strip_on_approve`) on any Phase-1 code path; the system MUST reject them at `template write` per I2 and warn at load per I3.
- **N2** (MUST NOT · scope: `internal/template/**`): The load path MUST NOT hard-fail on unknown top-level keys in an otherwise-V3 template; the system MUST warn only, so a template authored by a newer akb never bricks an older binary (RFC-002 §6's forward-compat rule, which ADR-006 N2 already binds for V3 keys).
- **N3** (MUST NOT · scope: `internal/template/**`, `cmd/akb/**`, `docs/knowledge-base/.agent-kb/templates/**`, `internal/skill/embedded/**`): Phase-1 work MUST NOT rename `dir` to `path` in the loader, the embedded templates, or this repo's KB templates; the system MUST perform the rename with Phase 3's D11 work, where the resolution-hint semantics it carries are defined.

## Exceptions

No exceptions are permitted. A template that needs a deferred key is a template waiting on a later phase; the I2 refusal says exactly that, and a need urgent enough to pull a key forward is a proposal to pull its whole phase's machinery forward, not to smuggle the key.

## Verification

- **I1, I2, N1**: testscript matrix — `template write` refuses each deferred key (`state_field`, `functions`, `projections`, `path`) and a typo'd key, each refusal naming the key and the supported set, exit 1 · gate: `go test ./test/` · mode: **block** once implemented.
- **I3, N2**: testscript — a V3 template carrying an unknown key still loads (a page write against it succeeds) with a stderr warning naming the key · gate: `go test ./test/` · mode: **block** once implemented.
- **I4, N3**: `rg -n '^path:' internal/template/embedded docs/knowledge-base/.agent-kb/templates` returns zero hits through Phase 1 · mode: **block**.
- **I5**: covered by ADR-006's own I5 verification (mixed-shape refusal precedes V3 interpretation) · mode: **advisory**.
- **I6**: covered by ADR-008's verification (current-shape schema blocks) · mode: **advisory**.
- **Human-only residue**: whether the supported-set message stays accurate as later phases extend the key set — each phase's tech-spec updates the message with its keys, checked at review.

## Context

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A21, researched and owner-ratified 2026-10-09 via decision form in the research run ([[rfc-002-phase1-scope-log|run LOG]]). RFC §6 defines the full V3 vocabulary but §17 distributes its semantics across phases: `schema` + `validations` + `lint_rules` are Phase 1; `state_field` is Phase 2 (approve unbundling, RFC-002-A15); `path`, `projections` (A17), `title_field`/`summary_field` (A13) are Phase 3; `functions` (A5) is Phase 4. Recon confirmed the loader is non-strict today (`internal/template/template.go:106,114` — a plain unmarshal, no `DisallowUnknownField`), so without this rule a Phase-1 author who writes `state_field:` produces a template that loads cleanly and does nothing — an inert declaration indistinguishable from a working one ([[rfc-002-phase1-scope-code-surface|code-surface recon]]). V2 already contains the cautionary example: `FieldSchema.Type` is declared in every shipped template and enforced by no code path, a declared-but-dead key exactly of the species this record prohibits. `dir` survives the phase not because its semantics are right but because its rename is load-bearing Phase-3 work (D11's resolution hints), and a bare cosmetic rename would churn every template twice. On ratification, RFC-002 is amended per its §18 lifecycle: row A21 moves to ratified as ADR-010 and §6 names the Phase-1 key set.

## Decision Drivers

- The D9 principle: an unenforceable declaration is a trap; traps fail loudly at authoring time or they fail silently in production.
- Phase coherence: a key is recognized exactly when the machinery that honors it ships — never before.
- Forward compatibility: warn-at-load keeps newer templates loadable by older binaries (RFC §6; ADR-006 N2 binds the same rule for V3 keys).
- Vocabulary stability: each later phase extends the set additively; no key is ever redefined or removed.

## Alternatives Considered

- **Recognize the deferred keys but ignore them with a warning** — rejected: an author who writes `state_field:` and sees it accepted believes the lifecycle gate is live; the declaration is inert and the belief is false — the precise trap this RFC was written to eliminate. Do not re-propose unless a key's machinery has shipped and only its default remains undecided.
- **Rename `dir` to `path` in Phase 1** — rejected: `path`'s resolution-hint semantics are Phase-3 D11 work; a bare rename without them churns every template now and again when the semantics land. Do not re-propose unless Phase 3 is pulled forward wholesale.
- **Strict unknown-key rejection at load** — rejected: a template authored by a newer akb would brick older binaries outright; RFC §6 fixes warn-at-load for exactly this forward-compat story. Do not re-propose unless akb's deployment model ever guarantees version lockstep.

## Consequences

- Good, because no template can carry an inert declaration: every key is either honored or refused, and the refusal teaches the supported set.
- Good, because the forward-compat story is symmetric with ADR-006's — old binaries warn and continue, new binaries enforce at authoring time.
- Good, because later phases extend the key set additively with their own specs, so this record never needs amendment to stay true — only its message text grows.
- Bad, because an author reading RFC §6's full example cannot use most of it until later phases; the I2 refusal message is the mitigation, and it must name the supported set plainly.
- Bad, because `dir` ships under its old name for three more phases; accepted to avoid a double rename.
- Neutral, because `FieldSchema.Type` — V2's declared-but-dead key — dies with the `schema.frontmatter` block under ADR-006; JSON Schema `type` keywords replace it with real enforcement.

## References

- Research run: [[rfc-002-phase1-scope-log|LOG]] (decision record), [[rfc-002-phase1-scope-code-surface|code-surface recon]] (loader leniency, TemplateV2 struct keys, FieldSchema.Type finding), [[rfc-002-phase1-scope-test-pinning|test-pinning recon]] (old-format guard pins).
- [[RFC-002-open-validation-engine|RFC-002]] §6 (full V3 vocabulary, strictness rule), §8.3 (D11 `path` semantics), §17 (phase distribution), §18 (candidate lifecycle).
- [[ADR-006-schema-frontmatter-retirement|ADR-006]] I5/N2 — refusal precedence for the retired block and the warn-at-load rule this record shares.
- [[ADR-008-phase1-current-document-shape|ADR-008]] I1/I2 — the document shape the Phase-1 `schema:` block validates.
- Code anchors at ratification (`a316f0a`): `internal/template/template.go:43-51` (current struct keys), `:106,114` (non-strict unmarshal), `:73-78` and `cmd/akb/templates_write.go:461-469` (the two `detectOldFormat` copies).
