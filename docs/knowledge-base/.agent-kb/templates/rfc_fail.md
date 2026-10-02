---
grammar: 1
type: rfc
id: RFC-000
title: A Self-Ratified Proposal Is Refused by the Identity Gate
status: accepted
provenance: agent-drafted
author: pete
steward: pete
created: 2026-10-02
updated: 2026-10-02
review_by: 2027-01-15
decided_by: pete
decided_at: 2026-10-02
scope: ["docs/knowledge-base/**"]
tags: [rfc, governance]
summary: The fail mockup — a page claiming acceptance where the ratifier is also the author, which the identity rule refuses.
---

# RFC-000: A Self-Ratified Proposal Is Refused by the Identity Gate

This page is the fail mockup of the `rfc` template. It claims `status: accepted` while naming the same party as `author` and `decided_by`, so the identity gate has something to refuse.

The rule that must fail is `decided_by_not_author`: the ratifier may not be the drafter, because an agent must never ratify its own proposal and a drafter must not silently promote their own page. Every other validation is deliberately satisfied, including the future `review_by`, the required `decided_at`, the ISO date formats, the non-empty scope and tags, and the body word count. The failure should therefore name exactly one rule, which is what makes this mockup a precise regression test for the write gate rather than a general smoke test.
