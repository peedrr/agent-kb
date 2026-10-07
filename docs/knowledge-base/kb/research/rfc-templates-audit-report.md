---
as_of: "2026-10-02"
created: "2026-10-02"
grammar: 1
informs:
- RFC-003
provenance: agent-drafted
run: rfc-templates
scope:
- "docs/knowledge-base/**"
status: final
summary: "Independent lighter audit of the rfc-templates run: all 18 claims verified against fresh primaries — 18 supported, 0 corrected, 0 killed, with caveats noted."
tags:
- audit
- evidence
- rfc-format
title: "Evidence audit — RFC format research, 18 claims (rfc-templates)"
type: research
updated: "2026-10-02"
---
All 18 claims verified against their primaries. Composing the audit report.

# Evidence Audit — RFC format research (lighter pass, 18 claims)

Method: every primary fetched fresh this run (lane quotes not trusted). Content located via stored-slice `findText` on fetched primaries. No claim audited on a citation alone.

## Per-claim verdicts

**1. KEP approvers decide `implementable`; approvers ≠ authors — SUPPORTED.**
Primary: `raw.githubusercontent.com/kubernetes/enhancements/master/keps/sig-architecture/0000-kep-process/README.md`. Verbatim: "The approvers are the individuals who decide when to move this KEP to the `implementable` state. * Approvers should be a distinct set from authors." (Same "distinct set" rule also stated for *reviewers* — claim's wording matches.)

**2. PEP 1 historical-document clause + Resolution header link — SUPPORTED.**
Primary: `peps.python.org/pep-0001/`. Verbatim: "Once resolution is reached, a PEP is considered a historical document rather than a living specification." and "at the very least the Resolution header should be added with a direct link to the relevant post making a decision on the PEP."

**3. Oxide RFD 1 six states, editable `published`, `committed` post-implementation — SUPPORTED.**
Primary: `rfd.shared.oxide.computer/rfd/0001`. Enum verbatim (six states listed). "Note that just because something is in the `published` state does not mean that it cannot be updated and corrected." "…by the time an idea represents the consensus or direction, it should be in the `published` state." "Once an RFD has become implemented — that is, once it is not an idea of some future state but rather an explanation of how a system works — its state should be moved to be `committed`."

**4. CSSWG: Closed/Verified separate axes + `Commenter Timed Out (Assumed Satisfied)` — SUPPORTED.**
Primaries: `bin/issuegen.pl` (raw, main) + `github.com/w3c/csswg-drafts/labels`. Script legend: "An issue can be closed as `Accepted`, `OutOfScope`, `Invalid`, `Rejected`, or `Retracted`. `Verified` indicates commentor's acceptance of the response." Script fields: `Closed:` and `Verified:` are separate lines ("Verified: URL to a message where the commenter indicates satisfaction… Verification indicates full closure of the issue"). Labels page contains the literal label `Commenter Timed Out (Assumed Satisfied)`. Minor nuance (not material): the script's `%statusStyle` map has 7 keys (adds `objection`), the legend enumerates 5 — internal drift inside the primary, already flagged by the lane.

**5. IESG DISCUSS blocking/holder-cleared; COMMENT non-blocking — SUPPORTED.**
Primary: `datatracker.ietf.org/doc/statement-iesg-handling-ballot-positions-20220121/`. Verbatim: "a: *require* an explanation of the position and b: are blocking"; "These are blocking, and you **do** need to address them"; "Comments do not have to be addressed; they are not blocking"; "Once the AD (or ADs!) holding the DISCUSS position(s) are satisfied, they will clear their DISCUSS position"; "the AD will generally change the DISCUSS to 'No Objection' after those changes are in a new version of the draft."

**6. Squarespace: yes / not yet, deliberately no "no"; "yes, if" carries conditions — SUPPORTED.**
Primary: `engineering.squarespace.com/blog/2019/the-power-of-yes-if`. Verbatim: "Approvers say 'yes' or 'not yet.' We want to encourage constructive comments, so the template doesn't suggest 'no.'" "Yes, if you can show that each iteration was a useable milestone"… "means that the design author isn't blocked and doesn't need to wait for the approver to read again later." *Nuance (same post):* "Reviewers can and do say no" in Architecture Review — the no-"no" rule is scoped to the template's approver section, which is exactly what the claim asserts; not a correction.

**7. AIDR: append-only at `arbitrated` (except `superseded_by`); human-only arbitration; no rejected status — SUPPORTED.**
Primary: `raw.githubusercontent.com/snapsynapse/aidr/main/SPEC.md`. Verbatim: "`arbitrated`: the arbiter has decided. The record is now append-only except for `superseded_by`." "Arbitration is human. An agent MUST NOT author the Arbitration section." "There is no rejected status. A decision not to act is still an arbitrated decision."

**8. IETF: 185 days / "six months" boilerplate / anachronisms "illusory…wasted effort" / 2026bis softens to marking — SUPPORTED.**
Primaries: `ietf.org/ietf-ftp/ietf/1id-guidelines.txt` §8: "An Internet-Draft will expire exactly 185 days from the date that it is posted on the IETF Web site…" (and §2: "Use of the phrase 'expires in six months' or 'expires in 185 days' is not acceptable."). Boilerplate verbatim in draft-ietf-procon-2026bis: "Internet-Drafts are draft documents valid for a maximum of six months…" — confirms the mandated boilerplate retains six months. Anachronisms-07: "the expiry after six months of Internet-Drafts, as described in [RFC2026], is illusory and often leads to wasted effort. It is illusory because drafts, once posted on line, never disappear…" 2026bis: "A Internet-Draft that has been not been changed for more than six months will be marked as Expired and may be removed from some views of the collection." *Caveat:* 2026bis is still a draft (-11), not a published RFC — the claim's "softens" phrasing correctly frames it as draft text.

**9. KEP 617 kep.yaml rot — SUPPORTED (descriptive part verbatim; one clause is interpretation).**
Primary: `raw.githubusercontent.com/…/617-improve-kep-implementation/kep.yaml`. Verbatim: `title: Enhance KEP implementation … status: provisional … creation-date: 2018-09-08` (no stage, no last-updated). *Interpretation note:* "its substance is the shipped standard" is not asserted by the file itself; it is corroborated by the validator requiring `kep.yaml` metadata for every KEP (`pkg/repo/validate.go`) — reasonable but not a document-stated fact.

**10. Peer-reviewed expiry-pressure numbers — SUPPORTED (NudgeBot verified via venue primary; preprint unfetchable).**
- Nudge: `arxiv.org/abs/2011.12468` abstract verbatim: "In a randomized trial on 147 repositories in use at Microsoft, Nudge was able to reduce pull request resolution time by 60% for 8,500 pull requests…"
- Stale bot: `arxiv.org/html/2305.18150` full text verbatim: "15% more PRs were closed in the first month of adoption but overall 10% fewer PRs were closed by the end of the first year"; "overall 24% fewer PRs were merged by the end of the first year of adoption"; "overall 14% fewer contributors were active each month during the first year of adoption." (TOSEM/20 projects; matches claim.)
- NudgeBot: the preprint PDF (`users.encs.concordia.ca/...NudgeBot2022FSE-preprint.pdf`) timed out. Verified instead against the **official ESEC/FSE 2022 program page** (venue primary): "A/B cluster-randomized experiment on over 30k engineers. We observed substantial statistically significant decrease in both time in review (-6.8%, p=0.049) and time to first reviewer action (-9.9%…). We also used guard metrics… and saw no statistically significant change in these metrics." −6.8% + unchanged guardrails confirmed.

**11. Marzouk: 60% stale RFCs; hallucination-free extraction; 4-week backlog clear — SUPPORTED.**
Primary: `marzouk.io/posts/ai-rfc`. Verbatim: "nearly 60% of the ~20 RFCs created in the last nine months remained open with no updates for more than six weeks"; "This allows the AI to easily extract the data it needs to answer our questions without risk of hallucination or mix-ups"; "We were able to resolve the backlog of stalled RFCs within four weeks…" Note: these are operator self-reports (~20 RFCs, single org), not measured studies — the claim's framing should keep that strength level.

**12. philcalcado 2026-07-10: budget, rubric categories, Do Nothing, implementation boundary — SUPPORTED.**
Primary: `philcalcado.com/2026/07/10/writing_ai_assisted_rfcs.html`. Verbatim: "Mine is 2,500 words, with 1,500–2,000 as the target. I'm not religious about those numbers; I'm religious about having a budget." Rubric (five categories verbatim in one sentence): "find claims with no evidence, find alternatives that are obviously straw men, find bullets pretending to be reasoning, tell me where I'm explaining implementation instead of the proposal, check the word count." Do Nothing: "the only part of NABC I'm really married to is always including Do Nothing as an alternative." Boundary: "once we're debating filenames, method signatures, migration sequencing, or the exact DDL, we've probably crossed into a different document."

**13. lemieux/rfc-skills style guide + reviewer prompt — SUPPORTED.**
Primaries: raw `rfc-style-guide.md` + raw `rfc-reviewer-prompt.md` (main). Verbatim: "**Use tables** for dense reference data: API fields, status codes, configuration options - **Use prose with subheaders** for items needing explanation: risks, tradeoffs, design decisions"; "Target 500-1000 lines for most RFCs"; "Proposal | 40-60% of total RFC length"; "Diagrams and API contracts are dense information and don't count against length"; reviewer prompt: categories `CHOPPY_WRITING` and `AI_PATTERNS` present, "Only report issues with confidence ≥ 80"; "Warning Signs: RFC Became a Spec" list present (ToC needed, multiple appendices, config examples, full class implementations, etc.).

**14. tech-leads-club create-rfc mandatory fields + RFC≠TDD — SUPPORTED.**
Primary: raw `SKILL.md` (main). Verbatim: "At least 1 explicit assumption (with confidence level) - At least 2 decision criteria (with weights), stated before options - At least 2 options considered (including "do nothing" when relevant)"; "RFC is for decisions, not implementation — once the RFC is decided, create a TDD for the implementation plan."

**15. FEP: 2-year auto-withdrawal; PR #543 "doesn't work" + 1y→2y retune by counts — SUPPORTED.**
Primaries: `codeberg.org/fediverse/fep/raw/branch/main/fep/a4ed/fep-a4ed.md`: "If authors have not requested the proposal to be finalized, and there were no updates for 2 years or longer, a facilitator will set the status of the submission to `WITHDRAWN`." PR #543 (raw page + discussion): PR body verbatim: "FEP-a4ed requires facilitators to contact authors of a stale proposal before withdrawing it. As our experience shows, that doesn't work. In this PR I propose withdrawing stale proposals after 1 year of inactivity without contacting authors." Retune comment (silverpill): "I changed it to allow 2 years of inactivity. Today we have 3 drafts that have been inactive for 2+ years… 29 drafts have been inactive for 1 year, but some of them are not actually stale because discussion is still ongoing. So 2 years seems more reasonable to me." PR merged 2025-06-24. (PR started at 1y; the *original* pre-PR rule required facilitator contact — retune-by-counts confirmed.)

**16. kepval: typed control file only, strict YAML, 6 required keys, conditional PRR gating — SUPPORTED.**
Primaries: `api/proposal.go` + `pkg/repo/validate.go` + `pkg/kepval/approval.go` (master). Verbatim: `yaml.UnmarshalStrict(metadata, &kep)` ("no unknown keys"); struct tags `validate:"required"` on exactly `Title`, `Number(kep-number)`, `Authors`, `OwningSIG`, `Approvers`, `Status` — six keys; "metadata file missing for KEP: %s" (kep.yaml presence enforced); `isPRRRequired`: "required = kep.Status == api.ImplementableStatus || kep.Status == api.ImplementedStatus" with milestone ≥ v1.21 and stage/milestone-missing errors; PRR approver check: "this contributor (%s) is not a PRR approver (%v)". Negative half (README sections and questionnaire answers NOT machine-checked) is consistent with the traced call path (`validateFile` → `KEPHandler.Parse` + `kepval.ValidatePRR` only) and the lane's noted advisory-only `verify-toc-vs-template.sh`; no positive evidence of README-content validation exists in the fetched code. Note `KEPHandler.Validate()` (enum/SIG checks) exists in `api/proposal.go` but the CI path shown calls `Parse()` only — the lane's "may not run in presubmit" inference stands unrefuted.

**17. nesbitt.io AI-detection heuristics — SUPPORTED.**
Primary: `nesbitt.io/2026/05/21/rfc-artificial-contributors-to-open-source.md`, Appendix A verbatim: "Heuristics in informal use include perfectly formatted markdown, a commit message in the imperative mood that runs to four paragraphs, the substitution of a bulleted list where a sentence would do, and a level of politeness not previously observed in the project." Date on document: 2026-05-21 (checked metadata; matches).

**18. arXiv:2608.21747 format×model interaction — SUPPORTED.**
Primary: `arxiv.org/abs/2608.21747` abstract verbatim: "On the strongest models (Sonnet 4.6, GPT-5), format barely matters (quality spread 0.17-0.92). On weaker models, format produces spreads of 0.83-2.42 points, with code-proximate formats (OpenAPI, TypeScript contracts) recovering most of the capability gap." 5 formats × 6 models × 90 trials confirmed in abstract.

## Summary
**18 supported, 0 supported-with-correction, 0 killed.** No material contradictions found against primaries. Primary unfetchable in one case: NudgeBot preprint PDF (timed out) — verified against the official ESEC/FSE 2022 program-page abstract instead (venue-authoritative; not marked on a secondary). No claim rests on a secondary source.

## Residual caveats (interpretation/strength-level, not errors)
- Claim 9's "substance is the shipped standard" is researcher inference (corroborated by validator behavior, not stated by the artifact).
- Claims 11, 14, 13, 12 are single-operator/vendor or practitioner sources; strength appropriate only if the brief uses them as practice evidence, not measurement.
- Claim 8's 2026bis softening is draft text (-11), not published RFC text.