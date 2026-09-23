---
type: adr
title: Draft Status Is Not an ADR Status
summary: Fail mockup for the adr template — every rule holds except the status enum
tags: [validation]
status: draft
deciders: akb maintainers
created: 2026-01-05
updated: 2026-01-05
---

# Draft Status Is Not an ADR Status

This page is the fail mockup of the `adr` template. It satisfies every rule except one, so the mockup proves that the status enum fires without a second failure blurring the report.

<!-- FAILS: valid_status — "draft" is not a valid ADR status (must be one of: proposed, accepted, deprecated, superseded) -->

## Context

A record type that gates its own lifecycle needs a closed set of states. The page-level draft flag is a review state owned by `akb approve`, not a decision state, so an ADR that carries `status: draft` has picked a value the lifecycle never defined. The mistake is common in pages that predate the enum, and the rule has to catch it on its own so the author sees one clear reason to change one field.

## Decision

Keep the status enum as the single source of truth for the lifecycle. A page that needs to signal "not yet reviewed" uses the `is_draft` frontmatter flag instead, and `akb approve` clears it. See [[templates/guard-doctrine]] for the surrounding rule set.

## Consequences

The failure is precise: `valid_status` is the only rule that fires on this page, so the report names the enum and nothing else. Every other rule stays green, which is what makes the mockup usable as proof that the enum rule is the one under test.
