---
created: "2026-10-07"
deciders:
- Pete Hope
grammar: 1
id: ADR-007
is_draft: false
provenance: agent-drafted
revisit:
- RFC-002's kb.* sweep-level rules land and the no-external-citation guard becomes a user-authored cross-page rule
- akb gains search or link-graph coverage of the raw layer, re-opening where externally-sourced research captures live
scope:
- docs/knowledge-base/**
status: accepted
summary: Agent-generated research reports and run logs are first-class research pages under kb/research/; raw/ holds external sources only; KB pages never cite files that do not travel with the KB.
tags:
- research
- kb-layout
- agent-memory
title: Research Reports Are First-Class Pages, Not Raw Sources or Chat Residue
type: adr
updated: "2026-10-07"
---

# ADR-007: Research Reports Are First-Class Pages, Not Raw Sources or Chat Residue

> In the context of akb's own embedded knowledge base citing research that lived in gitignored scratch directories, facing the choice of where agent-generated research reports live so that every KB citation travels with the KB, we decided that research reports and run logs are first-class `research` pages under `kb/research/`, written through `akb` under a minimal template, and neglected ingesting them into `raw/`, distilling the evidence into the citing records and dropping the corpus, and committing the `.pi/` scratch tree as-is, to achieve decision records whose every citation resolves inside the KB root — searchable, wikilinkable, lint-checked, and present in every worktree — accepting a one-time byte-faithful backfill of the existing corpus and a snapshot-frozen lifecycle discipline for research pages, because research is the evidence layer the ADR/SPEC/RFC records cite, and in the llm-wiki model agent-generated distillate is filed back into the wiki as pages; the raw layer is for external sources.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

Agent-generated research — findings reports (code recon, ecosystem surveys, audits, syntheses) and research-run logs — MUST be filed as pages of type `research` under `kb/research/` via `akb write`, governed by the `research` template's snapshot lifecycle: `draft` while a run is open, `final` when the run concludes, `superseded` when a later run replaces it, with the `as_of` snapshot date frozen once final. Decision records (ADR/SPEC/RFC pages) cite research with wikilinks, never with filesystem paths outside the KB root. `raw/` is reserved for externally-sourced, immutable documents. The pre-existing corpus — the cited reports and run logs under the gitignored scratch trees — is backfilled byte-faithfully (frontmatter added, bodies untouched), and every citing page's references are rewritten to wikilinks in the same wave.

## Invariants

- **I1**: WHEN a research run produces a durable finding or a run record, the operator (person or agent) MUST file it as a `research` page under `kb/research/` via `akb write`; scratch directories (`.pi/`, `.sisyphus/`) MAY hold ephemeral working artifacts only, never the durable record.
- **I2**: WHEN a KB page of type `adr`, `spec`, or `rfc` cites evidence, the citation MUST resolve inside the KB root — a wikilink or a KB-relative path; the page MUST NOT cite an untracked prose document outside `docs/knowledge-base/`.
- **I3**: The `research` template MUST enforce the lifecycle at write time: `status` is one of `draft`, `final`, `superseded`; the only transitions are `draft→final`, `draft→superseded`, and `final→superseded`.
- **I4**: WHEN a research page is `final` or `superseded`, its `as_of` snapshot date MUST NOT change; stale findings are corrected by a successor page that supersedes, never by rewriting the snapshot.
- **I5**: `raw/` MUST contain only externally-sourced, immutable documents; agent-generated analysis MUST NOT be filed there.
- **I6**: WHEN the pre-existing corpus is backfilled, report bodies MUST be preserved byte-faithfully; the backfill adds frontmatter and rewrites citations in the *citing* pages — it does not edit report content.

## Negative Constraints

- **N1** (MUST NOT · scope: `docs/knowledge-base/**`): Research reports MUST NOT be ingested into `raw/`; the raw layer is invisible to FTS5 search, the link graph, wikilink resolution, and the lint engine, so evidence filed there is undiscoverable and its citations unguardable.
- **N2** (MUST NOT · scope: `docs/knowledge-base/**`): A research page MUST NOT be summarized, restructured, or "cleaned up" on ingest; the report is the content, and silent edits destroy its value as evidence.
- **N3** (MUST NOT · scope: `docs/knowledge-base/**`): A `final` research page MUST NOT be deleted because the record it informs was ratified; ratified records' alternatives and revisit tripwires depend on the evidence baseline staying retrievable.
- **N4** (MUST NOT · scope: `docs/knowledge-base/**`): Research pages MUST NOT be organized into per-run subdirectories; the layout is flat under `kb/research/`, run identity lives in the `run` frontmatter key, and each run's log page is the wikilink hub for its reports.

## Exceptions

I2 governs prose evidence, not source-tree coordinates: a KB page MAY cite host-repository source files by path and line (e.g. `internal/cel/engine.go:29`) — those are code addresses in the repository that hosts the KB, not carried knowledge, and they travel with any checkout at the cited revision. External URLs are likewise exempt. Research pages MAY narrate historical scratch locations (a run log recording where its working artifacts lived is stating history, not citing evidence). No other exceptions are permitted; a case that appears to need one is a proposal to supersede this record.

## Verification

- **I3, I4**: `research` template CEL rules `valid_status`, `valid_state_transition`, `as_of_frozen_after_final`, `as_of_not_future` · gate: `akb write` · mode: **block** from template creation.
- **I2**: post-backfill gate — `grep -rn -E '\.pi/|\.sisyphus/' docs/knowledge-base/kb/decisions docs/knowledge-base/kb/specs docs/knowledge-base/kb/rfcs` returns matches only in ADR-007 (this rule's own definition), ADR-006 (the narrated provenance of its inline-quoted precedent), and RFC-003 (the loss note for its unrecoverable run record) · mode: **block** for declaring the backfill complete; a `no_external_citations` CEL lint rule (warning severity) on the adr/spec/rfc templates is the standing tripwire, with those same three pages as its documented accepted warnings.
- **I1, I5**: `akb template get research` exits 0 and `akb lint` reports no new `broken_links` or `orphans` errors after the backfill · gate: `akb lint` · mode: **block** for backfill completion.
- **I6**: spot-check — a sample of backfilled pages diffed body-only against their source files, byte-identical below the frontmatter block · mode: **block** for backfill completion.
- **Human-only residue**: whether a given artifact is a durable finding or ephemeral scratch is judgment; whether a research page's findings remain trustworthy as the code drifts is read off its `as_of` date, not mechanically detected.

## Context

This KB dogfoods akb's development, and its decision records cite the research that shaped them: RFC-001 and RFC-002 rest on thirteen reports in `.pi/subagents/proposals/json-cel/research/`; ADR-003 through ADR-006 and SPEC-002 rest on per-candidate runs under `.pi/research/rfc-002-*`; ADR-006 cites a precedent in `.sisyphus/drafts/`; RFC-003 cites a run record under `.pi/research/rfc-templates/`. Both scratch trees are gitignored, so none of that evidence travels with the KB — an agent in a fresh worktree implementing a SPEC cannot reach the research the SPEC stands on. The failure is not hypothetical: the run record RFC-003 cites was deleted before this record was written, and RFC-002 §16 cites the init-versioning spec at a path it no longer occupies. Two findings shaped the remedy. First, the corpus is two species — durable findings reports and run logs on one side, superseded working drafts on the other; only the first species needs preserving. Second, the llm-wiki model this KB follows already answers the placement question: raw sources are external and immutable, while agent-generated analysis is filed back into the wiki as pages — research reports are the distillate, not the source. RFC-002 §19 additionally names the research log as the living tracker for unaudited residue in later phases, a memory role that only a KB page can play (searchable, linkable, present in worktrees). The owner ratified the direction on 2026-10-07 with four forks: first-class pages over raw ingest and distill-then-drop; backfill scope of cited reports plus run logs, dropping superseded working drafts; the `.sisyphus` precedent quoted inline in ADR-006 rather than ingested whole; and execution split between the drafting session (this record and the template) and delegated backfill.

## Decision Drivers

- Worktree portability: an agent implementing a SPEC in a fresh worktree must be able to reach every document the SPEC and its ADRs cite.
- Retrievability: FTS5 search, wikilink resolution, the link graph, and the lint engine cover `kb/` pages only; evidence outside `kb/` is invisible to the machinery that makes the KB compound.
- Evidence permanence: "do not re-propose unless" clauses and `revisit` tripwires in ratified ADRs cite findings (e.g. compliance evidence verified on a date) whose baseline must stay retrievable to be checkable.
- Model fidelity: in the llm-wiki architecture, raw is for external immutable sources; agent-generated distillate belongs in the wiki layer.
- Dogfooding: the KB's own template machinery (CEL lifecycle rules, mockup-proved templates) should carry this contract, exercising the same path ADR-003 through ADR-006 use.

## Alternatives Considered

- **Ingest research into raw/** — rejected: raw is the external-source layer — invisible to search, link graph, wikilinks, and lint — and its manifest assumes immutability while run logs get appended during a run. Do not re-propose unless akb gains search and link coverage of the raw layer *and* a manifest semantics that tolerates appended files.
- **Distill-then-drop (self-sufficient records, no research layer)** — rejected: inlining all evidence into ADRs would either bloat them past readability or destroy the audit trail behind alternatives and revisit tripwires, and RFC-002 §19's living-tracker role has no home in a frozen decision record. Do not re-propose unless the KB adopts a policy that decision records must inline all evidence they depend on; research pages then become optional annexes.
- **Commit the `.pi/` research tree as-is** — rejected: `.pi/` mixes durable evidence with session-runtime state that must never travel, and committing it freezes no boundary between the two. Do not re-propose unless `.pi/` gains a durable/scratch separation enforced by the harness.
- **Per-run subdirectories under kb/research/** — rejected: nesting couples page identity to run organization, complicates wikilink basename resolution, and duplicates what the `run` frontmatter key and the run-log hub page already express. Do not re-propose unless the corpus grows past the point where a flat listing is navigable.

## Consequences

- Good, because every citation in every decision record resolves inside the KB root and travels to every worktree; research becomes searchable and wikilinkable agent memory; the `revisit` baselines in ADR-003 through ADR-006 become checkable again.
- Good, because the KB gains a worked example of filing agent-generated analysis as pages — the llm-wiki Query operation — dogfooded end-to-end through `akb template write`, `akb write`, and `akb lint`.
- Bad, because the backfill rewrites reference sections in ratified ADRs and RFCs (their `updated` fields bump, and the rewrites must be careful not to disturb ratified substance) and grows the KB by roughly 840K of historical markdown.
- Bad, because research pages demand discipline: a `final` page is a frozen snapshot, and readers must check `as_of` before trusting findings against drifted code; the template enforces the lifecycle but not the reading habit.
- Neutral, because RFC-003's cited run record is unrecoverable and receives a loss note rather than a reconstruction, and the stale init-versioning citation in RFC-002 §16 is rewritten to the SPEC-001 wikilink in the same wave.

## References

- Owner ratification of the four forks, 2026-10-07, in the session that drafted this record (first-class home; cited-reports-plus-logs scope; inline-quote for the `.sisyphus` precedent; split execution with delegated backfill).
- [[RFC-002-open-validation-engine|RFC-002]] §19 (research log as living tracker), Appendix C (research base inventory); [[RFC-001-json-schema-cel-coexistence|RFC-001]] (research base table).
- [[ADR-003-santhosh-v6-format-assertion|ADR-003]], [[ADR-004-cel-go-upgrade-pinned-extensions|ADR-004]], [[ADR-005-date-bridge-determinism|ADR-005]], [[ADR-006-schema-frontmatter-retirement|ADR-006]], [[SPEC-002-structured-validation-report|SPEC-002]] — the records whose References sections this decision rewrites.
- The llm-wiki pattern (A. Karpathy, "LLM Wiki" gist): raw sources external and immutable; the wiki layer is the persistent, compounding artifact; good answers are filed back into the wiki as pages.
