---
as_of: "2026-10-01"
created: "2026-10-01"
grammar: 1
informs:
- RFC-003
is_draft: false
provenance: agent-drafted
run: rfc-templates
scope:
- docs/knowledge-base/**
status: final
summary: "Deep lane on proposal lifecycle: IETF's 185-day expiry and its abolition, rotting hand-maintained status fields, and derived staleness postures over clock-driven kills."
tags:
- lane
- lifecycle
- staleness
- rfc-format
title: Lane lifecycle — proposal expiry, staleness, and abandonment mechanics (rfc-templates)
type: research
updated: "2026-10-09T16:35:05Z"
---
# Research: proposal lifecycle, expiry & staleness mechanics (lane-lifecycle)

Retrieved during this run. Latest primary-source dates observed: 2026-09-09 (IETF draft-carpenter-gendispatch-anachronisms-07),
2026-07-01 (draft-ietf-procon-2026bis-11), 2026-07-10 / 2026-08-18 (philcalcado), 2026-10-04 (open IESG comment
deadline on the PROCON recharter). Labels: **[direct]** = source states it; **[interpretation]** = my reading of
source text; **[inference]** = my design reasoning, in no source. Sibling lanes own the SPEC-side staleness
machinery, Rust RFC template sections, DoC/disposition, and kepval — not revisited here.

## Summary

Every mature proposal system has a stalled-state notion, but only one (IETF Internet-Drafts) uses a hard
date-driven kill, and the IETF spent 2023–2026 trying to abolish it because the measured effect was no-op
refreshes over a permanent archive, not removal. Meanwhile hand-maintained status fields rot: Kubernetes KEP 617
has said `status: provisional` since 2018-09-08 while what it proposes has been the shipped standard for years.
The only peer-reviewed evidence on deadlines says deadline pressure mostly *de-registers* items from the active
queue rather than completing them. Recommendation: a durable status enum with **no** clock-driven transition,
a derived `stale` posture computed by CEL, and a named steward as the only actor who may write `abandoned`.

## Findings

1. **IETF: 185 days, exact, per-version restart, with an enumerated suspension rule — now documented only in the
   legacy text.** **[direct]** "An Internet-Draft will expire exactly 185 days from the date that it is posted
   on the IETF Web site … unless it is replaced by an updated version (in which case the clock will start all
   over again for the new version)". Authors must print a concrete date on first and last page; "Use of the
   phrase 'expires in six months' or 'expires in 185 days' is not acceptable." Suspension is enumerated: when an
   I-D enters **"Publication Requested"** "it will not be expired until its status is resolved (e.g., it is
   published as an RFC). Data Tracker states not associated with a formal request to publish a document (e.g.,
   'AD is Watching') will not prevent an I-D from expiring." Also: no expiry while I-D posting is suspended
   around an IETF meeting (cutoff ≈ two weeks before the meeting). Source: Guidelines to Authors of
   Internet-Drafts, IESG/R. Housley, 2010-12-07, §2 + §8 —
   <https://www.ietf.org/ietf-ftp/ietf/1id-guidelines.txt>. **[direct]** The *current* guidance keeps the
   principle and drops the enumeration: "An I-D expires 185 days after it was placed in the Repository, unless
   it is in a state that prevents it from expiring. Examples of such states include being processed by the IESG
   for publication in the IETF stream, or being under review by the Independent Series Editor (ISE)" —
   <https://ietf.github.io/id-guidelines/>. **[interpretation]** The IETF prints a *computed literal date into the
   artifact* rather than storing a duration; for a frozen doc + CEL validator, store `posted_at` (or read the
   VCS date) and render `expires_at` — you get both robustness and auditability. And key any suspension on
   *"a formal disposition request has been filed"*, never on *"someone is looking at it"* — that distinction
   **is** the mechanism.

2. **Tombstones: frozen placeholder, version bump, never expire, resurrection by named actors only.**
   **[direct]** "When an I-D expires, a 'tombstone' file will be created that includes the filename and version
   number of the I-D that has expired … same as that of the expired I-D with the version number increased by
   one … Tombstone files will never expire and will always be available for reference unless they are replaced
   by updated versions of the subject I-D or the expired version is brought back by the explicit action of an
   Area Director." Unexpiration/extend requests must come "from an author, a working group chair, or an IESG
   member" with subject-line conventions (`Resurrect I-D <filename>`), and an author's request may be
   *overridden* by the WG chair. Source: legacy 1id-guidelines.txt §8. **[direct]** Operationally, tombstones
   "remain in the active repository for 185 days before being moved to the archive" — IAOC Standards Process
   Support Guidelines, June 2007,
   <https://iaoc.ietf.org/documents/Standards_Process_Support_Guidelines_June_2007.pdf> (second-hand: I did not
   fetch the PDF; treat this specific 185-day figure as unverified). **[interpretation]** The tombstone is
   append-only *evidence of a decision*, not a mutation of the document, and the resurrection path names
   actors rather than rules. Design implication: model `expired` as a derived posture over an immutable doc plus
   a sibling record `{id, version, expired_at, reason, resurrectable_by}` — never as an edit of the frozen doc.

3. **Expiry does not maintain a corpus; it relabels one. In 2026 IETF needed a human sweep of drafts "expired
   for more than 5 years".** **[direct]** IETF IDR WG thread, June 2026: "Spring Cleaning of IDR drafts expired
   for more than 5 years (6/2/2026)", closing with "Give these the tombstone they deserve." One 2011–2015-expired
   draft was found to have apparently passed last call and been lost — "I believe at one point John Scudder had
   indicated the document had completed last call and it was an error that it wasn't moved forward."
   <https://mailarchive.ietf.org/arch/msg/idr/xEwHc-Y-OuAmbxoADM4uqVes6j8/>. **[interpretation]** The actual
   remedy applied was a periodic human reconcile pass, not deletion-as-maintenance. Budget for one.

4. **The field is converging on "no hard clock": IETF's abolition draft expired, and its 2026 process revision
   softens the kill to a marking while *retaining* the six-month boilerplate.** **[direct]**
   draft-thomson-gendispatch-no-expiry-03 (Thomson & Hoffman, pub. 2024-01-17, intended BCP) would delete the
   "Expires:" header, amend the boilerplate, and replace expiration with an `active`/`inactive` marking in
   tooling, listing open options: "automatically marking drafts as 'inactive' after a certain period of time,
   for allowing working group chairs to control the marking … and for authors being able to change the status of
   their draft, either to mark a draft that has been overcome by events as 'inactive' or mark a draft as
   'active' when there is renewed interest."
   <https://datatracker.ietf.org/doc/html/draft-thomson-gendispatch-no-expiry-03> (running footer: "Drafts
   Aren't Milk"). **[direct]** Its own Datatracker state: IESG state **Expired**, last revision 2024-01-16,
   "Document has expired", intended BCP, no WG adoption state —
   <https://datatracker.ietf.org/doc/draft-thomson-gendispatch-no-expiry/>. **[direct]** The normative successor
   draft-ietf-procon-2026bis-11 (Salz & Bradner, 2026-07-01, obsoletes RFC 2026) keeps the mandated boilerplate
   verbatim ("valid for a maximum of six months") and replaces RFC 2026's hard sentence with: "A Internet-Draft
   that has been not been changed for more than six months will be marked as Expired and may be removed from
   some views of the collection. At any time, an Internet-Draft may be replaced by a more recent version of the
   same specification, restarting the six-month timeout period." (grammar verbatim) —
   <https://datatracker.ietf.org/doc/draft-ietf-procon-2026bis/>. **[direct]** PROCON is being rechartered with
   an IESG comment deadline of **2026-10-04**, and the proposed charter scopes non-editorial change to WG-milestone
   treatment and WG-adoption guidance only — i.e. removing expiry is out of scope —
   <https://mailarchive.ietf.org/arch/msg/ietf-announce/davRgBtVvfvZzNi_1wwEAOMlwIw/>. **[interpretation]** Two
   things to steal from -03: hard expiry can kill the very document that would have fixed expiry; and the
   convergence point for a mature process is *keep the vocabulary, degrade enforcement to a visible marking,
   push the marking rule into tooling*.

5. **The 185-vs-six-months discrepancy: acknowledged in 2005, fixed in the *guidelines*, never fixed in the
   *boilerplate*.** **[direct]** March 2005 review: "Sect. 2 mentions 'six months' for expiration, whereas the
   actual rule appears to be 185 days" → "Indeed, the secretariat's review also caught this error; thanks!"
   <https://www.mhonarc.org/archive/html/ietf/2005-03/msg00047.html>. **[direct]** The fix landed in the
   guidelines (§2 "maximum life of 185 days"; §8 "exactly 185 days"; loose phrasing explicitly banned) and did
   **not** land in the boilerplate: the mandated "Status of This Memo" text says "a maximum of six months" in
   2010, in the 2024 drafts, and in procon-2026bis-11 (2026-07-01). Sources as in findings 1 and 4.
   **[interpretation]** Structural, not clerical: the boilerplate is normative text inherited from RFC 2026 and
   needs a process RFC to change, which is why 2026bis keeps it. Design implication: **never express the same
   rule twice.** If a duration appears in prose and again as machine metadata they will diverge; the
   machine-readable field is the single source and prose renderings must be generated.

6. **IETF's own words: expiry "leads to wasted effort", with a liveness escape.** **[direct]**
   draft-carpenter-gendispatch-anachronisms-07 (2026-09-09, Informational), §2: "Experience has shown that the
   expiry after six months of Internet-Drafts … is illusory and often leads to wasted effort. It is illusory
   because drafts, once posted on line, never disappear; indeed the IETF maintains a public archive of them. It
   leads to wasted effort since some authors feel obliged to refresh a draft every six months with no
   significant change. This wastes effort and resources for the authors themselves, the IETF's own computing
   resources, and potentially the resources and time of innumerable others." Its replacement text adds a
   liveness qualifier — unchanged for six months "**and is not under active discussion in a working group**" →
   "marked as 'inactive' in tooling"; and concedes "marking" is not deletion. Source:
   <https://www.ietf.org/archive/id/draft-carpenter-gendispatch-anachronisms-07.txt> (pub. 2026-09-09, expires
   2027-03-13). **[interpretation]** The active-discussion qualifier is the right *idea* but needs an external
   activity signal, so it is not directly CEL-expressible from a frozen document alone — see the recommendation's
   `last_activity_at` substitute.

7. **Oxide RFD: six states, no clock, and the state is partly *derived by a bot* from repository events.**
   **[direct]** States: `prediscussion`, `ideation`, `discussion`, `published`, `committed`, `abandoned`;
   metadata is four AsciiDoc attributes (`authors`, `state`, `discussion` = PR link, `labels`). **[direct]** Bot
   repair: "If you move your RFD into `discussion` but fail to open a pull request, a friendly bot will do it for
   you. If you open a pull request but fail to update the state of the RFD to `discussion`, the bot will
   automatically correct the state … The bot will also cleanup the title of the pull request … The bot will
   automatically add the link to the pull request to the `discussion:` metadata." **[direct]** `ideation` is not
   "in progress": "Unlike the `prediscussion` state, there is no expectation that it is undergoing active
   revision… can be viewed as a scratchpad… Any member of the team is encouraged to start active development of
   such an RFD (moving it to the `prediscussion` state) with or without the participation of the original
   author." **[direct]** Merging is the author's call — "As a guideline, 3–5 business days to comment on your RFD
   before merging seems reasonable" and "RFDs shouldn't be merged if no one else has read or commented on it."
   **[direct]** `committed` = "once it is not an idea of some future state but rather an explanation of how a
   system works… This state is essentially no different from `published`, but represents ideas that have been
   more fully developed. While discussion on `committed` RFDs is permitted (and changes allowed), they would be
   expected to be infrequent." **[direct]** Post-publication discussion really does continue, recorded by link:
   "Discussion can continue on `published` RFDs! The `discussion:` link in the metadata should be retained,
   allowing discussion to continue on the original pull request," with a new issue as the escape hatch for
   larger threads. **[direct]** Tooling: an auto-published `.helpers/rfd.csv` plus Rust accessors "to automate or
   program tooling with RFD data", a rendered site, and `{num}.rfd.oxide.computer[/discussion]` short links.
   Source: <https://rfd.shared.oxide.computer/rfd/0001>. **[interpretation]** `abandoned` is defined as much for
   *"should be ignored"* as for *"non-viable"*; state lives in-band as a plain attribute with a bot repairing
   drift rather than trusting the human; and post-merge discussion is kept as a link, never re-absorbed into the
   document.

8. **KEP: seven statuses, four nouns for "not going anywhere", one forward transition owned by an approver, and
   an `editor` role instead of a clock.** **[direct]** "`status` … Must be one of `provisional`, `implementable`,
   `implemented`, `deferred`, `rejected`, `withdrawn`, or `replaced`." Definitions verbatim: `provisional` =
   "proposed and is actively being defined… The owning SIG has accepted that this work must be done";
   `implementable` = "The approvers have approved this KEP for implementation"; `implemented` = "has been
   implemented and is no longer actively changed"; `deferred` = "proposed but not actively being worked on";
   `rejected` = "The approvers and authors have decided that this KEP is not moving forward. The KEP is kept
   around as a historical document"; `withdrawn` = "The authors have withdrawn the KEP"; `replaced` = "has been
   replaced by a new KEP. The `superseded-by` metadata value should point to the new KEP." **[direct]** Who flips:
   "The approvers are the individuals who decide when to move this KEP to the `implementable` state. Approvers
   should be a distinct set from authors." Authors flip `withdrawn`. `rejected` is jointly attributed. **The
   process never assigns `deferred` to an actor.** A separate required `editor` field is "Someone to keep things
   moving forward." **[direct]** Structured fields: `stage: alpha|beta|stable`, `latest-milestone`,
   `milestone: {alpha,beta,stable}`, `replaces`, `creation-date`, optional `last-updated`, `see-also`,
   `feature-gates`, `metrics`. Sources:
   <https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/0000-kep-process/README.md> ·
   <https://github.com/kubernetes/enhancements/blob/master/keps/NNNN-kep-template/kep.yaml>.
   **[interpretation]** KEP has **no expiry and no revisit date**; `editor` ("keep things moving forward") is its
   substitute for the clock and is the design answer to "who re-affirms a stalled proposal".

9. **KEP's hand-maintained status demonstrably rots, and the maintainers say why.** **[direct]** KEP 617
   ("Enhance KEP implementation") `kep.yaml` on master reads `status: provisional`, `creation-date: 2018-09-08`,
   `replaces: []`, with no `stage`, no `milestone`, no `last-updated`
   (<https://raw.githubusercontent.com/kubernetes/enhancements/master/keps/sig-architecture/617-improve-kep-implementation/kep.yaml>);
   the official index renders 617 as `provisional`, stage `—` (<https://www.kubernetes.dev/resources/keps/>),
   while its substance (metadata moved to `kep.yaml`) is long since the shipped standard. Same index: KEP 1101
   "Immutable Fields" is `provisional`, stage `—`, created 2019-06-03. **[direct]** Maintainer account,
   kubernetes/enhancements #2960: "Between `stage`, `milestone`, and `status` there's overlap and ambiguity";
   "if, for some reason, the code missed the boat, someone has to go back and edit the KEP … That is prone to
   failure, and if it is forgotten, then people might take the wrong result"; "`latest-milestone` … is yet
   another field that's not common and easily falls out of date"; "`implemented` is generally used for KEPs that
   are completely finished post-GA"; "the logic itself is pretty easy to change, the tangly bits are in the data,
   since it's been mostly hand edited YAML that has historically not been validated against much."
   <https://github.com/kubernetes/enhancements/issues/2960>. **[interpretation]** A status field with no forcing
   function sits wrong for ~7–8 years and the generated index faithfully republishes the wrongness. You can
   always recompute `stale` from dates; you cannot recompute the truth of `implemented`. Three orthogonal
   lifecycle axes (stage / milestone / status) is one too many, and the release train is KEP's only reason for
   `milestone` — our format has no train, so: **one** status enum plus dates, no "phase" field.

10. **PEP: eight statuses, freeze-on-resolution, and *availability-driven* deferral; no date machinery at all.**
    **[direct]** `Status: <Draft | Active | Accepted | Provisional | Deferred | Rejected | Withdrawn | Final |
    Superseded>`. "In general, PEPs are no longer substantially modified after they have reached the Accepted,
    Final, Rejected or Superseded state. Once resolution is reached, a PEP is considered a historical document
    rather than a living specification. Formal documentation of the expected behavior should be maintained
    elsewhere." **[direct]** "A PEP can also be assigned the status 'Deferred'. The PEP author or an editor can
    assign the PEP this status when no progress is being made on the PEP. Once a PEP is deferred, a PEP editor
    can reassign it to draft status." **[direct]** And deferral is the *no-steward* state: if no PEP-Delegate can
    be found, "then the PEP will be marked as Deferred until one is available". **[direct]** Provisional is
    feedback-boxed, not date-boxed: "additional user feedback is needed before the full design can be considered
    'Final'", and "provisionally accepted PEPs may still be Rejected or Withdrawn even after the related changes
    have been included in a Python release." **[direct]** On decision, "at the very least the Resolution header
    should be added with a direct link to the relevant post making a decision on the PEP"; supersession is
    bidirectional (`Replaces` / `Superseded-By`). **[direct]** PEP 1 contains **no occurrence of the string
    "expir"** (verified by full-text search of the fetched document). Source: <https://peps.python.org/pep-0001/>.
    **[interpretation]** PEP answers our open question explicitly — *yes*, decided-but-not-implemented deserves
    its own state (`Accepted` ≠ `Final`), and the transition between them is keyed on **implementation reality,
    not a date**. Its `Deferred` is not an expiry: it is "no accountable human available", which is a cleaner
    trigger than a clock.

11. **FEP: the one live proposal system with a hard, date-driven, machine-checkable auto-withdrawal — and its
    self-reported failure modes.** **[direct]** Statuses `DRAFT | WITHDRAWN | FINAL`; metadata carries
    `dateReceived`, `dateWithdrawn`, `dateFinalized`. The rule: "If authors have not requested the proposal to be
    finalized, and there were no updates for 2 years or longer, a facilitator will set the status of the
    submission to `WITHDRAWN`." Reversible: "A proposal with status `WITHDRAWN` remains in the repository and can
    be resubmitted", and the state diagram allows `WITHDRAWN --> DRAFT`. `FINAL` is near-terminal: "can not be
    changed or updated in a way that would lead to adjustments to implementations. Minor corrections are
    allowed. Any substantial change to finalized proposal must be submitted as a separate FEP" with bidirectional
    `replaces`/`replacedBy`. Governance escape hatch: "FEP-a4ed … is a living document and can be updated
    despite having the `FINAL` status", with a ≥1-month cooling-off on process changes; submission gets a 7-day
    facilitator response; finalization needs ≥60 days in DRAFT plus a 14-day objection window.
    <https://codeberg.org/fediverse/fep/raw/branch/main/fep/a4ed/fep-a4ed.md>. **[direct]** 2025 change history
    and reasoning: the prior rule "requires facilitators to contact authors of a stale proposal before
    withdrawing it. As our experience shows, that doesn't work. In this PR I propose withdrawing stale proposals
    after 1 year of inactivity without contacting authors." Review pushed back → "I changed it to allow 2 years
    of inactivity. Today we have 3 drafts that have been inactive for 2+ years … 29 drafts have been inactive for
    1 year, but some of them are not actually stale because discussion is still ongoing. So 2 years seems more
    reasonable to me." Also recorded: "the proposed text allows bumping inactive FEPs with minor changes
    indefinitely. I think that is fine." And a predictability demand: "I think I'd be fine with this if there
    were an easy way of seeing which FEPs are close to expiring, especially if we move to stop notifying
    authors." Thread 2025-03-26 → 2025-06-24, merged.
    <https://codeberg.org/fediverse/fep/pulls/543>. **[interpretation]** Four ready-made lessons: notice-then-withdraw
    does not work, reported by the people operating it; an inactivity clock is *gameable* by trivial commits and
    the operators accepted that; the threshold needed empirical retuning (1y → 2y) driven by *counts*, not
    principle; and the moment you stop notifying authors you owe users a "what is about to expire" view — which
    is exactly a CEL-computable derived posture.

12. **Counter-datapoint: a system that *deleted* its decided-but-not-implemented state because every instance was
    a stale marker.** **[direct]** A system-design KB removed `proposed` from its ADR type: "in practice all five
    `proposed` ADRs (004, 005, 007, 019, 020) were long implemented — stale markers, not pending decisions";
    unadopted designs moved to a separate `proposals/` location, with the recorded rationale "Proposals become
    first-class and findable instead of squatting as hedged speculative notes" and "`Status: proposed` can no
    longer go stale, because it no longer exists"; proposals carry "a dated current-state anchor, with staleness
    against later ADRs an expected lifecycle event".
    <https://github.com/zby/commonplace/blob/main/kb/reference/adr/028-design-proposals-live-in-reference-proposals.md>.
    Label: **interpretation of a low-authority single-repo ADR, not a standard.** **[interpretation]** Direct
    counter-evidence to finding 10: the strongest argument that decided-but-not-implemented should be a *different
    document kind* or a *derived posture*, not a long-lived in-enum state. See Contradictions #2.

13. **The most transferable mechanism found: two-clock ownership + scope-scaled windows + pause levers.**
    **[direct]** An OSS project ADR (tamp-build/tamp, ADR 0017) makes auto-close depend on *whose turn it is*:
    "if it's from the author, the clock is on the maintainers (per 0016); if it's from a maintainer with a
    request-changes review, the clock is on the author (per 0017)". Windows scale with change scope — trivial
    14d / standard 30d / substantial 45d / major 60d, "LONGEST applicable tier wins (be generous)". Reminders
    fire "at the 50% and 90% marks". Three explicit pauses: `needs-time` label with an explanation, draft state,
    maintainer-applied `pinned`. Any author response — including "I'm still working on this" — resets the clock.
    Reformulating work into a new PR "is NEVER a valid reason to reopen a closed PR". Auto-close is explicitly
    reversible and non-punitive: "Authors who return with the requested changes pick up exactly where they left
    off. The close is a hygiene action." Cited motivation: "Maintainer time spent re-reading old PRs is a tax";
    "The graveyard is the worst aesthetic in OSS."
    <https://github.com/tamp-build/tamp/blob/main/docs/adr/0017-pr-staleness-autoclose.md>. Label: **low authority
    (repo design doc, unmeasured), high transferability.** **[interpretation]** The key import is ownership: a
    clock can only be "on" someone, so a document with no steward has an ill-defined staleness clock. Every rule
    in this ADR is a pure function of `{last_actor_role, last_activity_at, scope_tier, pause_labels, now}` —
    exactly the shape a CEL validator can evaluate.

14. **philcalcado: "my RFCs expire" means *revisit*, with the author as the re-affirmer — not a countdown.**
    **[direct]** The 2018 template header carries `Authors`, `To be reviewed by:`, `Revisit Date:`, `State:`
    (published example: reviewed by 10/5/2018, revisit 04/17/2019, state "Feedback Requested"). Lifecycle phases:
    `Draft` → `Feedback Requested` → `Active` (proceed) / `Abandoned` (decline) / `Retired` ("The changes proposed
    on this RFC aren't in effect anymore, the document is kept for historical purposes"). "Each RFC has a
    revisit date, by when the authors will update the mailing list on what they have learned since the feedback
    phase. This is a natural point for an RFC to be retired and a new approach proposed." Who re-affirms: the
    authors — and the revisit exists to unlock commitment, not to expire: "people are much more welcoming to
    change if they know that the decisions and assumptions will be revisited at some point in the future",
    citing Linda Rising's *Fearless Change* "Try it for a limited period" pattern; "Having an expiry date and a
    commitment from the authors to revisit the decision is one way to implement this Pattern." Disposition on
    staleness: "once an RFC moves away from *Feedback requested*, it is considered a historical artifact, if not
    discarded completely. RFCs aren't great as documentation, once the feedback period is over I usually ask the
    authors to document any relevant parts somewhere else like a wiki or even a different Google Doc."
    <https://philcalcado.com/2018/11/19/a_structured_rfc_process.html>. **[direct]** The 2026-07-10 follow-up
    restates it as the load-bearing property: "the most important part of a Structured RFC was never the
    document. It's ultimately a *change management process*… That's why my RFCs expire: so we can disagree,
    commit, and revisit instead of stalling on getting it right the first time." His 2026 template is "in an
    eternal *Feedback Requested* state and it has a revisit date, and I mean both seriously."
    <https://philcalcado.com/2026/07/10/writing_ai_assisted_rfcs.html>. **[interpretation]** His "expiry" is a
    **revisit reminder owned by the author**; `Retired` is the removal and it is a human judgement *at the
    revisit*, not the date. He and the IETF mean incompatible things by "expire" — see Contradictions #4.

15. **Deadline pressure changes the inventory; targeted reminders change behaviour. Peer-reviewed, with
    numbers.** **[direct, positive]** *Nudge* (Microsoft; randomized trial, 147 repositories, 8,500 overdue PRs):
    "Nudge was able to reduce pull request resolution time by 60% … compared to overdue pull requests for which
    Nudge did not send a notification. Furthermore, developers receiving Nudge notifications resolved 73% of
    these notifications as positive." Scaled to 8,000 repos / 210,000 notifications over a year.
    <https://arxiv.org/abs/2011.12468>. **[direct, positive]** *NudgeBot* (Meta; cluster-randomized A/B, >30k
    engineers, ~330k diffs): time in review −6.8% (p=0.049), time to first reviewer action −9.9% (p=0.010), share
    of diffs taking >3 days to close −11.89% (p=0.004); guardrails unchanged, so no rushing.
    <https://users.encs.concordia.ca/~pcr/paper/NudgeBot2022FSE-preprint.pdf> (ESEC/FSE 2022). **[direct,
    negative]** *Understanding the Helpfulness of Stale Bot* (TOSEM 2023; interrupted time series over 24
    months): "our predictions indicate that 15% more PRs were closed in the first month of adoption but overall
    10% fewer PRs were closed by the end of the first year of adoption… overall 24% fewer PRs were merged by the
    end of the first year of adoption." Positives found: merged PRs' first review latency −21%, closed PRs'
    resolution time −22%, discussion comment volume unchanged. Negatives: "overall 14% fewer contributors were
    active each month during the first year of adoption", and "Stale bot also tends to intervene more in PRs
    submitted by novice contributors." <https://export.arxiv.org/pdf/2305.18150v1.pdf>. **[direct, mechanism]**
    Bot first responses carry no pressure: across 111,094 closed PRs in ten mature projects, "bot-first response
    time explains 0.33% of the variance in the PR lifetime, while first human response time explains 64.7% of
    it"; and humans respond *more slowly* in bot-first threads ("using a bot might not generate pressure for
    project maintainers to follow up rapidly").
    <https://research.tudelft.nl/en/publications/nudge-accelerating-overdue-pull-requests-toward-completion/>
    (second half from the same result set's companion study; **search-summary level, primary PDF not fetched**).
    **[interpretation]** Where a deadline helped, it re-pointed an existing human's attention (Nudge/NudgeBot).
    Where it hurt, it closed items no human had been pointed at (Stale bot), and the closes scored as progress.
    For a format whose primary consumer is an agent, the honest assertable output of a validator is a *posture*
    plus a *pointer to the steward*; the closing act must remain an accountable write.

16. **Escape hatches are universal — no system examined enforces a hard content freeze.** **[direct]** PEP 1
    itself is `Status: Active` because it is "never meant to be completed"; FEP-a4ed is `FINAL` yet "a living
    document [that] can be updated despite having the `FINAL` status"; IETF states "I-Ds are not an archival
    document series" while hosting a permanent archive; Oxide permits edits to both `published` and `committed`
    RFDs; KEP's `implemented` is "no longer actively changed" yet the Release Signoff Checklist requires
    "Implementation History" to be up to date. Sources as in findings 1, 4, 7, 8, 10, 11. **[interpretation]** See
    Contradictions #1: the ratified rule here ("the format freezes at ratification; change = supersession") is
    stricter than every precedent found.

## Lifecycle state-machine candidates

Assumed frozen-payload fields: `status`, `steward`, `posted_at` (this revision), `opened_at`, optionally
`review_by`. All three candidates below are evaluable as pure functions of document + now.

### A — IETF-shaped: hard clock, per-revision restart, tombstone

| transition | who | CEL |
|---|---|---|
| new → `active` | author posts | — |
| `active` → `active` | new revision | `expires_at = posted_at + P185D` |
| `active` → `review-pending` | steward/approver files formal disposition request | suspends clock |
| `active` → `expired` | **clock alone** | `now >= expires_at && status not in ['review-pending','published']` |
| `expired` → `active` | named role (author / chair / AD) | manual, overridable |

**Pros:** zero discretion; the only precedent with an explicit tombstone artifact and a named-actor resurrection
path; arithmetic is trivially testable. **Cons:** it produced the failure its own designers documented (finding 6:
no-op refreshes with "no significant change"), it relabels rather than removes (finding 3), and it killed its own
abolition draft (finding 4). Worse for us: with a document that **freezes at ratification**, per-revision restart
is incoherent — there is no revision to reset the clock with, so a frozen doc would `expire` and be unresurrectable
except by supersession.

### B — KEP-shaped: durable status, no clock, an `editor` role instead

`provisional` → `implementable` → `implemented`, with `deferred` / `rejected` / `withdrawn` / `replaced` as exits.
Approver flips `implementable`; author flips `withdrawn`; a named `editor` ("someone to keep things moving
forward") performs hygiene; `replaced` pairs with `superseded-by`.

**Pros:** zero false expiry; matches the most-used process in this space; no arithmetic to get wrong; the `editor`
role answers "who re-affirms" with a person rather than a date. **Cons:** demonstrably rots — KEP 617 has been
`provisional` since 2018-09-08 (finding 9) — and gives a stalled proposal no visibility, so neither a reader nor
an agent can distinguish "actively being defined" from "abandoned in 2019" without archaeology.

### C — Oxide/FEP-shaped: minimal enum, reversible withdraw, clock only on the draft state

`draft` → `review` → `published` → `committed`; exit to `withdrawn`/`abandoned`, reversible back to `draft`;
supersession via `replaces`/`replaced_by`. Optionally the FEP rule: if `status == draft && now - last_update >
P730D` → `withdrawn` by the facilitator, **with no prior notice step**.

**Pros:** field evidence from a real proposal system with the same "state in-band, steal the PR link, never edit
after merge" architecture; `withdrawn`→`draft` gives a free resurrection path with no new state; `committed` =
"explanation of how a system works" is a genuinely useful definition of *not a proposal anymore*. **Cons:** the
FEP clock is gameable by trivial commits, conceded by its own operators; and the one available datapoint on
Oxide's terminal state is that the dashboard shows **0 abandoned RFDs** — a search-summary-level claim I could not
verify (see Missing evidence), i.e. the state may be aspirational.

### Recommendation: A's derived posture on a C-merged enum, with FEP hysteresis and the two-clock trigger

Store six statuses; derive one posture; **the clock never writes a status.**

```yaml
status: draft | review | accepted | implemented | superseded | abandoned
steward: <name|agent>            # required, non-empty, exactly one accountable party
superseded_by: <id>              # required iff status == superseded
posted_at: <ts>                  # this revision
opened_at: <ts>                  # first version, never changes
review_by: <ts>                  # REQUIRED iff status in [review, accepted]
last_activity_at: <ts>           # bumped by any recorded review/decision entry
disposition_requested_at: <ts>   # finding 1's "Publication Requested" analogue
```

Derived, never stored (CEL — pure function of document + now):

```cel
is_overdue    = has(review_by) && now > review_by
is_suspended  = has(disposition_requested_at) && status in ["review","accepted"]
is_stale      = status in ["review","accepted"] && is_overdue && !is_suspended
is_at_risk    = status in ["review","accepted"] && now > review_by - duration('336h')   # 14d warning
is_dormant    = status == "draft" && now > last_activity_at + duration('4380h')         # ~6mo silence
is_terminal   = status in ["superseded","abandoned"]
needs_steward = !has(steward) || steward == ""
```

| transition | who | mechanism |
|---|---|---|
| `draft → review` | steward | sets `review_by` |
| `review → accepted` | **named approver ≠ steward** | records decision + resolution link; sets a *new* `review_by` (the revisit) |
| `accepted → implemented` | steward | keyed on implementation reality, not a date (finding 10) |
| `review`/`accepted` → `abandoned` | **steward only, with a written reason** | the clock produces `is_stale`, never this transition |
| `draft → draft` (re-affirm) | steward | bumps `last_activity_at`; explicitly permits "no substantive change" |
| `* → superseded` | author of the successor | bidirectional `supersedes`/`superseded_by` (PEP + FEP both) |
| `abandoned → draft` | steward | resurrection; append-only, original doc untouched |

Why this shape, per owner decision it serves:

- **Pure function of document + now.** Every field is a timestamp or a token; every predicate a comparison. No
  predicate consults history, GitHub, or an external API — which matters because the validator must work on the
  frozen artifact alone. The one smart rule deliberately *rejected* is Carpenter's "don't expire while under
  active discussion" (finding 6): it needs an external activity signal. Its in-document substitute is
  `last_activity_at`, bumped by the steward — which is honest about who is asserting liveness.
- **`is_stale` is derived, so it cannot rot.** Finding 9 is the reason: KEP 617's wrong `provisional` is
  unfixable by computation, but `is_stale` is recomputed every run and can never itself be a stale field. This is
  the compromise between A and B.
- **`accepted` is kept** (answering the open question *yes*), because it is the highest-value place to hang
  `review_by`: accepted-and-unbuilt is exactly where a proposal silently becomes a lie. Paired with the
  counter-lesson from finding 12: `accepted` **requires** both `steward` and `review_by`, so it cannot become the
  five-dead-`proposed`-ADRs graveyard. An `accepted` without a revisit date is *invalid*, not merely stale.
- **`abandoned` requires a steward, never the clock** — findings 4, 6, 11, 15.
- **`draft` silence yields `is_dormant`, not `abandoned`.** FEP's lesson is that the *notice* step is the part
  that failed and its operators retuned the threshold from counts (finding 11). `is_dormant` supplies the
  "which proposals are close to expiring" view FEP reviewers asked for without committing to a kill.
- **`review_by`, never `expires_at`.** Finding 14: the field's word choice is the source of the very ambiguity
  the owners flagged. One word, one concept.

## What expiry pressure actually does

Ranked by strength; popular-practice sources labelled as such.

1. **[direct, randomized, peer-reviewed]** Actor-targeted reminders on overdue items compress latency
   substantially with no quality regression: Nudge −60% resolution time (147 repos, 8,500 PRs, 73% of
   notifications judged positive); NudgeBot −6.8% time in review, −9.9% time-to-first-action, −11.89% share of
   diffs >3 days to close, guardrails (24h-review share, reviewer eyeball time) unchanged.
2. **[direct, quasi-experimental, peer-reviewed]** Hard auto-close is a *removal* mechanism with a measurable
   cost: Stale bot gave +15% closes in month 1, then **−24% merges** and **−14% monthly active contributors**
   across the following year, skewed toward novice contributors.
3. **[direct, operator self-report]** Notice-then-withdraw does not work; FEP's operators replaced it with a
   silent date-driven withdrawal, then *loosened* the date 1y → 2y after counting 29 one-year-inactive proposals
   of which some were still under active discussion.
4. **[direct, operator self-report]** On Internet-Drafts, expiry "is illusory… [and] leads to wasted effort since
   some authors feel obliged to refresh a draft every six months with no significant change"; drafts "never
   disappear"; the practical consequence was a >5-year residue requiring a human "spring cleaning" pass.
5. **[direct, mechanism]** Machine-generated pressure does not work: bot first responses explain 0.33% of
   PR-lifetime variance vs 64.7% for human first responses, and humans respond *more slowly* in bot-first threads.
6. **[interpretation of low-authority practitioner designs]** Two independent repos converged on scope-scaled
   windows + 50%/90% reminders + "any response resets the clock", and on the insight that a staleness clock is
   ill-defined without an assigned steward.

**Net reading:** the *reminder* changes behaviour; the *deadline* changes the inventory. Deadline mechanisms
helped only where they re-pointed an existing accountable human; where they substituted for one, they removed
items and the removal was booked as progress. For a proposal format consumed primarily by an AI agent, the only
honest thing a validator can assert is a posture (`is_stale` / `is_dormant` / `is_at_risk`) plus a pointer to the
steward — the closing act must remain an accountable write.

## Contradictions

1. **Freeze-vs-live is unresolved field-wide, and our ratified rule is stricter than all of it.** PEP says
   post-resolution PEPs are "a historical document rather than a living specification", yet PEP 1 is `Active`,
   PEP states change (Provisional was added later), and `Accepted`→`Rejected` post-acceptance is explicitly
   permitted. FEP-a4ed is `FINAL` and simultaneously a living document. IETF says I-Ds are "not an archival
   document series" while hosting a permanent archive. Oxide allows changes to `published` and `committed`.
   Recorded honestly: **every precedent keeps an escape hatch; "change = supersession" has no precedent I could
   find.** The nearest precedent is FEP's `replaces`/`replacedBy` plus its one-month cooling-off — i.e. live
   editing is reserved for the *process* document, never for the proposals. If we want an escape hatch without
   violating freeze, the precedent-supported form is **one designated live artifact** (the process doc / a drift
   ledger), not a mutable field on frozen proposals. Sources: findings 1, 4, 7, 8, 10, 11, 16.
2. **Decided-but-not-implemented: three systems say yes, one deleted the state.** PEP keeps `Accepted` distinct
   from `Final` with a dedicated index section "Accepted PEPs (accepted; may not be implemented yet)"; KEP keeps
   `implementable` ≠ `implemented`; Oxide keeps `published` ≠ `committed`. Against that, a real KB deleted
   `proposed` because 5/5 instances were "stale markers, not pending decisions" — and KEP's own tracker shows the
   cost of the pattern (617 `provisional` since 2018-09-08). Both readings rest on primary evidence; I do not
   resolve them by preference — hence the recommendation's guard: `accepted` **requires** `review_by` + `steward`.
3. **185 days vs six months: acknowledged, half-fixed, deliberately retained.** The 2005 secretariat thread
   confirms the *guidelines* error was known and fixed; the *mandated boilerplate* still says "a maximum of six
   months" in 2010, in the 2024 drafts, and in the 2026 process revision. Fixed where editable, retained where
   normative — a two-source-of-truth defect, not a clerical oversight. Sources: findings 1, 4, 5.
4. **"Expire" means two incompatible things.** IETF = removed from the active set. philcalcado = time to
   re-affirm, with `Retired` as the removal and that removal a *human judgement at the revisit*, not the date.
   Three of the owner's open questions inherit this ambiguity.
5. **Expiry's effect: positive in two experiments, negative in one longitudinal study.** Nudge/NudgeBot show
   significant improvement with no quality regression; the Stale-bot ITS shows a same-direction short-term closure
   spike followed by materially worse merges and contributor retention. **[inference]** These are reconcilable as
   *targeted nudge to the correct actor* vs *blanket auto-close of anything idle* — but no source I found states
   that reconciliation, and I will not present it as sourced. Sources: finding 15.
6. **KEP metadata drift inside one project:** the process README documents `superseded-by` as a list of `KEP-123`
   identifiers, while the authoritative template's `kep.yaml` uses `replaces:` with *path* values and never shows
   `superseded-by` — despite the README's own claim that "The KEP template is the authoritative definition of
   things like the metadata schema." Minor, but it is live prose-vs-schema drift. Source: finding 8.

## Missing evidence

- **Current IETF tombstone behaviour is unverified.** Semantics come from the 2010 legacy guidelines and the 2007
  IAOC guidelines; the *current* id-guidelines omits the word, and I did not fetch the datatracker wiki that
  defines today's repository layout. The 2026 IDR thread is consistent with tombstones but is not a spec.
- **No evidence that philcalcado's revisit dates are honoured.** His published 2018 RFC lists revisit
  2019-04-17; I found no follow-up and no data. The 2026 post asserts the mechanism as a design constant, not as
  a measured outcome.
- **Oxide `abandoned` usage unverified.** `.helpers/rfd.csv` 404s on both `master` and `main`; the site root did
  not extract. The "0 abandoned RFDs" claim appears only in a search summary and is low confidence. Consequently
  I cannot state the actual state distribution of Oxide RFDs.
- **No quantitative study of *proposal-document* expiry.** Every number in "what expiry pressure actually does"
  is PR/diff-level, from code review. Transfer to design proposals is my inference, not evidence.
- **KEP: who flips `deferred` is never stated.** The README assigns flipping actors only for `implementable`
  (approvers) and `withdrawn` (authors); `rejected` is joint. This silence is itself informative for "who kills a
  stalled proposal".
- **No found acknowledgement that the six-months *boilerplate* is knowingly wrong.** Only the 2005 exchange about
  the guidelines text, and the subsequent fix, were found — no issue-tracker entry, no statement.
- **procon-2026bis's final state is open.** The original milestone was "Jun 2026 – Submit RFC2026bis to the IESG
  as BCP" and it is still -11 (2026-07-01); I did not verify IESG submission or approval. The "expiry is being
  softened to a marking" conclusion rests on draft text, not published RFC text.
- **Nothing on how an AI agent consumes a stale proposal.** The only machine-readable review-posture artifact
  found is `Hello-GregKulp/freshness-spec` (`.freshness.yml`: `lifecycle.mode/status`, `reviewed_at`,
  `review_after`, `stewardship.steward.type: agent`, a validator emitting warnings with `--fail-on-warning`;
  principles "Stale does NOT mean wrong", "Expired does NOT imply deletion", "Abandoned does NOT imply useless",
  "Freshness is a review posture") — a single-author experimental spec, not a standard, and **adjacent to the
  sibling SPEC-staleness lane's territory**, so not researched further.

## Sources

**Kept (primary, decision-relevant)**

- `https://www.ietf.org/ietf-ftp/ietf/1id-guidelines.txt` — IESG, 2010-12-07. The only source with the full §8:
  185 days, per-version restart, Publication Requested vs "AD is Watching", tombstone construction, never-expiring
  tombstones, named-actor unexpiration.
- `https://ietf.github.io/id-guidelines/` — what the 185-day rule says today; evidence the tombstone/suspension
  detail was dropped.
- `https://www.ietf.org/archive/id/draft-carpenter-gendispatch-anachronisms-07.txt` (2026-09-09) — operator-stated
  case against expiry ("illusory", "wasted effort") plus the active-discussion liveness qualifier.
- `https://datatracker.ietf.org/doc/html/draft-thomson-gendispatch-no-expiry-03` + `/doc/draft-thomson-gendispatch-no-expiry/`
  — the active/inactive replacement proposal, its open marking options, and proof it expired.
- `https://datatracker.ietf.org/doc/draft-ietf-procon-2026bis/` (2026-07-01) — the 2026 normative direction:
  six-month boilerplate retained, removal softened to "marked as Expired and may be removed from some views".
- `https://mailarchive.ietf.org/arch/msg/ietf-announce/davRgBtVvfvZzNi_1wwEAOMlwIw/` — PROCON recharter, comment
  deadline 2026-10-04, expiry removal out of scope.
- `https://mailarchive.ietf.org/arch/msg/idr/xEwHc-Y-OuAmbxoADM4uqVes6j8/` — operational proof that expired
  drafts persist 5+ years and that tombstoning is live vocabulary.
- `https://rfd.shared.oxide.computer/rfd/0001` — the six states, exact `committed` wording, post-publication
  discussion retention, bot state-repair.
- `https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/0000-kep-process/README.md` and
  `.../keps/NNNN-kep-template/kep.yaml` — seven statuses, who flips `implementable`, the `editor` role, the
  date/stage fields.
- `https://raw.githubusercontent.com/kubernetes/enhancements/master/keps/sig-architecture/617-improve-kep-implementation/kep.yaml`
  — the verifiable `status: provisional` / `creation-date: 2018-09-08` rot.
- `https://github.com/kubernetes/enhancements/issues/2960` — maintainers' own account of status/stage/milestone
  overlap and hand-edited-YAML staleness.
- `https://peps.python.org/pep-0001/` — status model, `Resolution` header, availability-driven `Deferred`, the
  "historical document rather than a living specification" clause, and (by textual absence) no expiry mechanism.
- `https://codeberg.org/fediverse/fep/raw/branch/main/fep/a4ed/fep-a4ed.md` — the only live date-driven
  auto-withdrawal rule found, with statuses, dates, and reversibility.
- `https://codeberg.org/fediverse/fep/pulls/543` — 2025 operator reasoning: notice-doesn't-work, 1y→2y retune,
  gaming vector, demand for a near-expiry view.
- `https://philcalcado.com/2018/11/19/a_structured_rfc_process.html` — revisit-date header, five lifecycle
  phases, `Retired` semantics, "once an RFC moves away from Feedback requested… historical artifact".
- `https://philcalcado.com/2026/07/10/writing_ai_assisted_rfcs.html` (2026-07-10) — "That's why my RFCs expire",
  the eternal-Feedback-Requested-with-revisit-date stance.
- `https://arxiv.org/abs/2011.12468` — Nudge: randomized, 8,500 PRs, −60% resolution time.
- `https://users.encs.concordia.ca/~pcr/paper/NudgeBot2022FSE-preprint.pdf` — NudgeBot: cluster-randomized,
  ~330k diffs, −6.8%/−9.9%/−11.89% with unchanged guardrails.
- `https://export.arxiv.org/pdf/2305.18150v1.pdf` — Stale bot ITS over 24 months: +15% closes month 1, −24%
  merges and −14% active contributors over the year, novice-skewed.
- `https://github.com/tamp-build/tamp/blob/main/docs/adr/0017-pr-staleness-autoclose.md` — two-clock ownership,
  scope-scaled windows, 50%/90% reminders, three pause mechanisms. Low authority, high transferability.
- `https://github.com/zby/commonplace/blob/main/kb/reference/adr/028-design-proposals-live-in-reference-proposals.md`
  — the counter-datapoint: `proposed` deleted because 5/5 instances were stale markers. Low authority.
- `https://iaoc.ietf.org/documents/Standards_Process_Support_Guidelines_June_2007.pdf` — tombstone
  active-repository/archive lifecycle. **Retained despite failing to fetch it directly; treat the "185 days then
  archive" figure as second-hand.**

**Rejected / deprioritized**

- Wikipedia *Internet Draft* — redundant; every claim traced back to the IETF primaries above.
- `draft-levine-iduse-00` and the 2024 internet-history "why six months?" thread — useful colour on the origin of
  the rule, non-normative, superseded by the gendispatch/procon drafts kept.
- kubernetes.dev KEP index (153 KB rendered table) — kept only as corroboration of KEP 617/1101 status; the
  authoritative `kep.yaml` was fetched instead.
- `Hello-GregKulp/freshness-spec` — single-author experimental spec; cited once as adjacent prior art and
  flagged as a sibling lane's territory.
- Assorted repo hygiene plans, stale-bot GitHub commits, and search-surfaced PR chatter about "outdated
  documentation" — off-topic (documentation freshness rather than proposal lifecycle).
- oxidecomputer `rfd-site`, `rfd-api` README, `cio/rfds.1s.sh` — tooling detail only; nothing on lifecycle
  semantics beyond RFD 1, and the CSV they describe was unreachable.

## Next steps

1. **Recover Oxide's `.helpers/rfd.csv`** (a tree listing of `oxidecomputer/rfd` will locate its current path).
   One lookup settles whether `abandoned` is in real use — the most decision-relevant unverified claim left, since
   it determines whether a terminal `abandoned` state is worth including at all.
2. **Validator-owner decision, not research:** should the CEL predicate set be normative *in* the format document
   (self-describing, verifiable offline against a frozen doc) or only in the CLI? Findings 1 and 4 push both ways
   — IETF prints a literal computed date; the modern direction pushes the rule into tooling. Recommended hybrid:
   literal dates stored, rule text normative in the format, CEL in the CLI.
3. **Check procon-2026bis's final disposition** before ratification if the format will cite IETF precedent as
   "settled" — as of -11 it is still a draft.
