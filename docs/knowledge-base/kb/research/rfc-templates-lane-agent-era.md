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
summary: "Deep lane on the 2026 agent-era RFC literature: a hybrid contract of machine-extractable header, mandatory alternatives slot, pre-flight rubric, and prose argument."
tags:
- lane
- agent-era
- rfc-format
- prose
title: Lane agent-era — RFC format decisions for agent consumers (rfc-templates)
type: research
updated: "2026-10-01"
---
# Research: Agent-era RFC (proposal) format — lane-agent-era

**Scope:** format decisions for an RFC whose primary consumer is an AI coding agent (drafted by agents,
reviewed by agents + humans, later read by adjudication agents). Read-only lane; output file only.
**Observation date:** all fetches this session. No reliable system clock was available to me; the newest
source timestamp observed was **2026-09-28** (explainx.ai mirror of the create-rfc skill). Every source
below carries its own publication date, cited where material.
**Validation limitation:** no `source_check` tool was exposed to this lane, so claims were validated by
fetching the original documents and quoting them directly. Cross-checking against third-party fact
databases was not possible; all `[direct]` labels mean "quoted from the primary document I fetched".

---

## Summary

The agent-era RFC literature converges on a **hybrid contract**: a small machine-extractable header
(table/frontmatter) + a mandatory alternatives/do-nothing slot + a machine-runnable pre-flight rubric +
a hard length budget + exclusion of implementation artifacts — with **narrative prose reserved for the
argument itself**. The single sharpest split with our sibling ADR/SPEC contracts is prose-vs-bullets:
Calçado, Lubow, Raviv and the lemieux Claude-Code RFC skill all explicitly argue for prose in the
argument sections, and lemieux ships a reviewer prompt that *flags bullet-lists-read-as-sentences as a
defect*. That advocacy is justified by **human reviewer comprehension and AI-slop detection**, not by
any measured agent-comprehension result — no controlled study of proposal-format comprehension by agents
exists (confirmed negative). Adopt the hybrid: machine contract in the header, prose for claims and
alternatives, lists only for enumerations and comparisons.

---

## Findings

1. **[direct] Calçado's 2026 format deltas are six, and they are all format-level, not process-level.**
   He states the process itself "works almost exactly as before" and that "the part I had to change is the
   part I always said mattered least: the format." The deltas: (a) accountability over disclosure, (b) keep
   NABC with mandatory *Do Nothing*, (c) argument is narrative prose not bullets, (d) implementation detail
   stays out, (e) diagrams are first-class, (f) claims traceable + word budget + LLM rubric.
   **Source:** https://philcalcado.com/2026/07/10/writing_ai_assisted_rfcs.html (2026-07-10; accessed this session).
   **Confidence:** high. **Implication:** an RFC format can borrow these six without importing a lifecycle.

2. **[direct] Word budget: 2,500 words hard, 1,500–2,000 target — and he insists the number is not the point.**
   Verbatim: "Mine is 2,500 words, with 1,500–2,000 as the target. I'm not religious about those numbers;
   I'm religious about having a budget." Rationale is attention pricing, not intelligence: "It's just as easy
   to hide a poorly researched argument behind six bullet points as it is under six thousand words, so the
   goal isn't 'shorter is smarter.'" Purpose of the budget is to force the appendix split.
   **Source:** same post. **Confidence:** high. **Implication:** adopt a budget *field* (declared max + target)
   rather than a fixed constant; the sibling contracts' "minimal ceremony" goal is compatible, but the
   budget's job is to force "shared argument vs appendix", which is a *placement* rule, not a length rule.

3. **[direct] NABC is retained, but the only part he is "really married to" is mandatory Do Nothing.**
   Verbatim: "the only part of NABC I'm really married to is always including Do Nothing as an alternative.
   I can't tell you how many times over the last ten years I've started an RFC convinced something was a
   great idea, only to discover that I couldn't make a convincing case for doing it." He also explicitly
   blesses format variance: "There are many formats and even variations of NABC that work just fine, so if
   your company has adapted to something you all like you should probably keep using it."
   **Source:** same post. **Confidence:** high. **Implication:** make do-nothing a *validated required field*
   (not a section title), and do not hard-code NABC as the section list.

4. **[direct] Prose over bullets — the argument is that AI turns bullet lists into plausible prose without
   adding reasoning.** Verbatim: "AI makes this worse because it can turn those bullets into plausible prose
   without filling in the missing reasoning. The output looks like the writing problem has been solved, while
   the argument underneath can still be exactly the same list." He also dates the bullet epidemic *before*
   AI: "Just before AI-generated content became common enough to warrant changing the RFC process, I was
   already seeing an epidemic of RFCs that were little more than lists of bullets."
   **Source:** same post. **Confidence:** high (direct quote). **Implication:** the anti-bullets rule is an
   **anti-slop** rule, not an agent-comprehension rule. See Contradictions §1.

5. **[direct] Implementation-detail exclusion is stated with a precise boundary: shape yes, artifacts no.**
   Verbatim: "Give an agent access to the repository and suddenly your architectural proposal comes back with
   class names, schemas, method signatures, migration sequences, and a list of files to modify… I still want
   enough technical detail to understand the shape of the proposal. A small interface, some pseudocode, or a
   toy example can be the clearest way to communicate an idea. But once we're debating filenames, method
   signatures, migration sequencing, or the exact DDL, we've probably crossed into a different document."
   **Source:** same post. **Confidence:** high. **Implication:** the RFC needs a *positive* allowance list
   (interface sketch, pseudocode, toy example) alongside the prohibition list — a bare "no implementation
   detail" prohibition would delete the shape information reviewers need.

6. **[direct] The rubric is an author-side private pre-flight, explicitly not a grading system.** Verbatim:
   "before I ask anyone to read an RFC, I hand the document and the rubric to an LLM and tell it to be
   annoying: find claims with no evidence, find alternatives that are obviously straw men, find bullets
   pretending to be reasoning, tell me where I'm explaining implementation instead of the proposal, check the
   word count." And: "the rubric is not a grading system. Passing it doesn't mean the proposal is good. It
   means the document is ready for humans to have the argument." He names an unanticipated second effect:
   it makes it safer to put forward an incomplete idea, because the author finds the hole privately first.
   **Source:** same post. **Confidence:** high. **Implication:** the five checks above are the closest thing to
   a published, citable RFC rubric — adopt them verbatim as the pre-flight check IDs.

7. **[interpretation] The exact rubric questions are NOT publicly available in the post.** The post gives the
   five *check categories* quoted in F6, not a numbered question list. The full rubric and the 2026 RFC
   template live in a Google Doc that could not be fetched (JavaScript-rendered / auth-gated): the post links
   `https://docs.google.com/document/d/1FzwfdJ7k_zSeCAICjGYhh6rUYR-cUDCitl0yUTVDXU4/edit`. I searched
   specifically for a mirror of the rubric text and found none.
   **Source:** fetch attempt, this session (error: "Page appears to be JavaScript-rendered"). **Confidence:**
   high that it is unfetchable here; **medium** that no mirror exists anywhere.
   **Implication:** cite the five categories, not an invented rubric.

8. **[direct] Traceable claims are justified by *machine verification*, not by honesty alone.** Verbatim:
   "references make AI useful on the review side too. A reviewer can give the RFC and its sources to an LLM
   and ask it to check whether the sources actually support the claims being made. The author gets to use
   cheap machine research, but also has to leave enough of an evidence trail for cheap machine verification."
   **Source:** same post. **Confidence:** high. **Implication:** traceability is a *two-sided* contract —
   the source pointer must be machine-resolvable and must be attached to the specific claim, so a reviewer
   agent can check claim↔source support rather than just source existence.

9. **[direct] Diagrams are first-class content, partly as an anti-agent hack.** Verbatim: "I would have added
   this even without AI… There's also a temporary hack here. At least as I write this, LLMs are still
   noticeably worse at producing a good architectural diagram than they are at producing convincing
   architectural prose. Asking for a diagram therefore often forces the author back into the loop: you have
   to decide what the important boxes actually are and how they relate. I don't expect that advantage to last."
   **Source:** same post. **Confidence:** high. **Implication:** if the RFC mandates a diagram, note it as a
   *contingent* anti-slop device (self-declared as temporary by its author) — do not present it as a durable
   comprehension win.

10. **[direct] AI-disclosure stance: "disclosure becomes less useful to me"; accountability is the rule.**
    Verbatim: "This is the most important rule, and also the simplest: I don't care who typed the words, you
    own them. If the model hallucinated something and you didn't catch it, that's your mistake, not the
    model's… As AI becomes a normal part of producing these documents, disclosure becomes less useful to me.
    What matters is whether there is a human being willing and able to defend everything they put their name
    on." His 2026 companion post supplies the replacement mechanism: a two-value label distinguishing
    "I slopped it and here's what I got" from "I did some research and here's what I think", whose purpose is
    that "the reader knows how to engage with it."
    **Sources:** same post; https://philcalcado.com/2026/05/31/nutrition_labels_for_ai.html (2026-05-31).
    **Confidence:** high. **Implication:** adopt accountable-ownership; treat disclosure as *contested* (see
    Contradictions §2) — the majority empirical position in adjacent ecosystems is the opposite.

11. **[direct] Calçado 2018 already prohibited implementation detail, with a documented failure story.** The
    2018 RFC says: "A good RFC will describe the scope and the approach. It should not contain a list of
    specific tasks or project plan," and "existing RFCs… often get into too much detail about the 'how' and
    not enough on the 'what'." The annotated post supplies the cautionary tale (a SoundCloud RFC where one
    casual sentence about writing tooling in Python triggered a weekend-long language debate, while the
    latency/availability/durability content went undiscussed).
    **Source:** https://philcalcado.com/2018/11/19/a_structured_rfc_process.html (2018-11-19).
    **Confidence:** high. **Implication:** the rule is pre-AI and independently motivated; agents make it
    worse but did not create it — so it should survive into the agent-era format unchanged.

12. **[direct] Raviv's four questions an RFC must answer, and the RFC-vs-plan distinction.** The four:
    "1. What exactly is the problem? 2. How bad or important is it? 3. Will this solution fix the problem?
    4. How costly will the solution be, vs. the alternatives?" On plans: "an RFC is not an actual
    implementation plan (like a plan in Cursor or plan mode in Claude Code): the focus is on the problem, the
    general shape of the solution, and the alternatives. For a plan, the details are everything."
    **Source:** https://ohadravid.github.io/posts/2026-08-let-them-write-rfcs/ (2026-08; accessed this session).
    **Confidence:** high. **Implication:** these four questions are a compact, testable acceptance predicate
    for an agent-drafted RFC — closer to a rubric than to a section list.

13. **[direct] Raviv's iteration claim: version the file, don't perfect the first draft.** "An RFC is unlikely
    to be good enough on the first try, so I'll iterate and revise (V1.md, V2.md, etc.)… I wanted the LLM to
    produce an artifact that makes it easier for me to both understand the problem and identify what I don't
    (yet) understand about it." He spent "hundreds of millions of tokens" on four versions of one markdown file.
    **Source:** same post. **Confidence:** high. **Implication:** the format should tolerate and *name* iteration
    (version field / `Vn` filenames) rather than punishing it; a draft-state field is load-bearing.

14. **[direct] Raviv's flexible-structure stance, including "short prose".** He prompts: "Write a markdown
    file called $PROBLEM_RFC.md about this problem and how to solve it. Keep the prose short, have good code
    examples." Footnote 4: "I sometimes ask for the *style* of Rust's RFC format… but the actual structure
    should be flexible enough to fit the problem." Footnote 3 concedes the tradeoff: "I know LLM prose can be
    tiring… But LLM code can be much more tiring, so there's a tradeoff."
    **Source:** same post. **Confidence:** high. **Implication:** this is the *weakest* prose endorsement among
    the four human authors — Raviv wants short prose + explanatory code, which sits closer to our sibling
    bullets+fragments formats than to Calçado's narrative-prose rule.

15. **[direct] Marzouk's extractor argument: the header table exists to prevent hallucination and mix-ups.**
    Verbatim: "Our RFC template starts with a simple table where the author must fill in key info, such as
    reviewers and dates for main milestones. This allows the AI to easily extract the data it needs to answer
    our questions without risk of hallucination or mix-ups." This is a *production* system (daily job over
    Confluence RFCs → OpenAI API → structured answers → Slack reminders + dashboard).
    **Source:** https://www.marzouk.io/posts/ai-rfc (accessed this session; page title "AI as Process
    Infrastructure: Practical Learnings from an AI-augmented RFC Process").
    **Confidence:** high (direct quote). **Implication:** this is the strongest available evidence that a
    leading structured table is a *reliability* device for agent readers, not just a convention.

16. **[direct] Emoji-in-titles state encoding, with a caveat the author himself hedges.** Verbatim: "One thing
    that worked really well for us was using emojis (📝💬✅❌🚀) in the RFC titles to indicate their state. In
    addition to being easier for human readers to parse on the RFCs list page in Confluence, it also made
    filtering the RFCs (slightly) easier for the tool." Note the parenthetical "(slightly)" — this is the
    author's own downgrade of the machine benefit.
    **Source:** same post. **Confidence:** high. **Implication:** a redundant *human*-legible state channel,
    not a machine contract. Keep the real state in a typed field; emoji is at most an index decoration.

17. **[direct] Marzouk's five explicit quality criteria for the AI's structural check.** Verbatim: "Don't ask
    the AI to just 'evaluate the quality of the RFC.' Instead, give it clear criteria and examples… we found
    it highly reliable for structural checks: Were enough alternatives considered? Is the problem definition
    clear? Are the pros and cons listed properly? Are there contradictions in the proposed solution? Are there
    unanswered questions in the comments?" He separates this from engineering correctness: "we don't trust it
    blindly to assess engineering correctness."
    **Source:** same post. **Confidence:** high. **Implication:** the adjudication/quality lane should score
    *structure and discussion health*, and must not be asked to score technical correctness — the author
    states the opposite of trust there.

18. **[direct] Marzouk's keep-AI-out-of-content decision and its rationale.** Verbatim: "We debated whether
    the tool should leverage AI to enrich the document… or answer comments, but we opted to focus strictly on
    process governance and facilitation. We trust our engineers to leverage AI for writing and responding to
    comments when it makes sense, and to verify its input." Paired with: "our goal was not to have the AI
    judge technical decisions, but to facilitate visibility, timeliness, and closure."
    **Source:** same post. **Confidence:** high. **Implication:** a clean role split to copy — the agent
    *facilitates and checks structure*; the human authors and decides. If our RFC has an adjudication agent,
    this is direct evidence for scoping it to structure/convergence, not technical verdicts.

19. **[direct] Marzouk's "AI needs guardrails" cites the multi-turn degradation literature — but slightly
    misstates it.** His text: "AI loses 39% accuracy when given 6 tasks." The cited paper (arXiv:2505.06120,
    "LLMs Get Lost In Multi-Turn Conversation") actually reports "an average drop of 39% across six generation
    tasks" for **multi-turn vs single-turn** performance, decomposed into "a minor loss in aptitude and a
    significant increase in unreliability", with the mechanism "LLMs often make assumptions in early turns and
    prematurely attempt to generate final solutions, on which they overly rely."
    **Sources:** https://www.marzouk.io/posts/ai-rfc ; https://arxiv.org/abs/2505.06120 (v1 2025-05-09).
    **Confidence:** high (both quotes verified directly). **Implication:** treat Marzouk's "one question per
    call" architecture as sound, but do **not** repeat his statistic in our own docs — the paper supports
    "multi-turn degrades reliability and early assumptions ossify", which is *itself* an argument for an
    up-front written proposal (the RFC) over iterative agent Q&A.

20. **[direct] Lubow's five essential elements and the over-indexing rule.** Verbatim: "A functional template
    only needs to prompt for five things: the problem, the proposed solution, the 'why' behind the chosen path,
    the discarded alternatives, and the risks. If it takes longer to document the change than it would to code
    it, you're over-indexing on detail. We aren't writing a Wikipedia entry; we're documenting a decision-making
    process."
    **Source:** https://eric.lubow.org/2026/implementing-an-rfc-process-that-engineers-dont-hate/ (accessed
    this session). **Confidence:** high. **Implication:** five fields is the floor, and the over-indexing test
    is a *cost* test (documentation time vs coding time), which is the practical version of Calçado's word budget.

21. **[direct] Lubow's Documented-No outcome and silent-RFC failure mode.** On rejection: "Don't delete these.
    A rejected RFC is still valuable as it documents why we decided not to do something, which prevents the
    same proposal from coming back six months later without new information." On silence: "A 'Silent RFC' is
    effectively a dead one. It kills the engineer's motivation when nothing happens… The author is responsible
    for driving the feedback loop."
    **Source:** same post. **Confidence:** high. **Implication:** an RFC needs a terminal *rejected* state whose
    content is preserved and indexed — a directly machine-consumable artefact for an adjudication agent
    ("was this already decided against, and on what grounds?").

22. **[direct] Lubow's AI stance: useful for drafting, disqualifying if you don't understand it.** Verbatim:
    "We've found that Claude and similar tools are useful for drafting RFCs. They are not to generate the
    ideas, but to organize the engineer's thinking… If someone asks an AI to generate a proposal they don't
    actually understand, the meeting will expose that immediately." Also: "the value of an RFC isn't in the
    prose. It's in the thinking the prose represents."
    **Source:** same post. **Confidence:** high. **Implication:** independent (non-derivative) support for
    Calçado's accountability-over-disclosure position — but note Lubow's enforcement mechanism is a *human
    meeting*, which an agent-primary format does not have.

23. **[direct] The Tech Leads Club `create-rfc` skill is the highest-volume field contract in the ecosystem,
    and its mandatory-field list is explicit.** Mandatory: title (clear, action-oriented); Background/context;
    Driver; Approver(s); Impact level (HIGH/MEDIUM/LOW); ≥1 explicit assumption *with confidence level*; ≥2
    decision criteria *with weights, stated before options*; ≥2 options (including "do nothing" when relevant);
    recommended option with rationale tied back to the criteria. Seven mandatory sections (Header & Metadata,
    Background, Assumptions, Decision Criteria, Options Considered, Action Items, Outcome) + four recommended
    (Relevant Data, Pros/Cons, Estimated Cost, Resources). Ships a 13-item quality checklist and a
    "anti-patterns" section (predetermined conclusion, vague background, missing do-nothing, criteria after
    options, hidden assumptions).
    **Sources:** https://github.com/tech-leads-club/agent-skills/blob/main/packages/skills-catalog/skills/(creation)/create-rfc/SKILL.md
    (fetched raw); mirror at https://explainx.ai/skills/tech-leads-club/agent-skills/create-rfc (page stamp
    "Updated Sep 28, 2026"). **Confidence:** high (SKILL.md read in full).
    **Implication:** this is what agents have actually seen; the two design moves worth stealing are
    **criteria-before-options** and **assumption + confidence + invalidation trigger**.

24. **[direct] The create-rfc skill hard-separates RFC from TDD and treats the RFC/TDD boundary as a rule.**
    Its own comparison table: RFC = "Propose + decide", audience "broad stakeholders, leadership", output
    "Decision + rationale"; TDD = "Design + plan implementation". And: "**RFC is for decisions, not
    implementation** — once the RFC is decided, create a TDD for the implementation plan." The `Outcome`
    section is deliberately a placeholder during drafting, filled after the approvers decide.
    **Source:** same SKILL.md. **Confidence:** high. **Implication:** an RFC format should *name its handoff
    target* (here: the sibling SPEC/plan format) so the prohibition has a destination instead of just a ban.

25. **[direct] The lemieux/rfc-skills Claude Code plugin is the only ecosystem artefact that implements a
    *two-agent* RFC pipeline with a machine-checkable defect taxonomy.** It ships `rfc-writer` and
    `rfc-reviewer` subagents plus four skills (brainstorming → writing → finalizing → incorporating feedback),
    built on the Anthropic Agent Skills spec. The reviewer prompt defines 11 issue categories, three severity
    levels, and a confidence threshold: "Only report issues with confidence ≥ 80." Categories include
    `CHOPPY_WRITING`, `UNDEFINED_REFERENCE`, `TABLE_OVERUSE` ("Table used for items needing prose explanation
    (risks, tradeoffs)"), `THIN_SECTION`, `AI_PATTERNS` ("Em dashes, 'Let's dive in', 'It's worth noting',
    rhetorical questions"), and `HORIZONTAL_LINES`.
    **Sources:** https://github.com/lemieux/rfc-skills ; raw files
    `skills/writing-technical-docs/references/rfc-reviewer-prompt.md`,
    `agents/rfc-reviewer.md` (repo tree read via GitHub API this session). **Confidence:** high.
    **Implication:** this is a directly transplantable pre-flight design: **category + severity + confidence
    threshold + quote-the-location + never-suggest-fixes**. The confidence gate is what keeps a rubric from
    generating noise.

26. **[direct] The lemieux style guide's length budget is in *lines with a section-share rule*, and it bans
    time estimates and Open Questions.** Verbatim: "Target 500-1000 lines for most RFCs… Diagrams and API
    contracts are dense information and don't count against length… Full code implementations DO count against
    length (and probably shouldn't be there)." Section budgets: Abstract 2–3 sentences; Proposal "40-60% of
    total RFC length"; Abandoned Ideas "150-300 words per alternative"; Risks "50-100 words per risk";
    Rollout "200-400 words"; "Never include time estimates (weeks, sprints, dates). An RFC proposes what to
    build, not when." And: "Do NOT include an 'Open Questions' section… An 'Open Questions' section signals
    the author didn't finish their job."
    **Source:** https://raw.githubusercontent.com/lemieux/rfc-skills/main/skills/writing-technical-docs/references/rfc-style-guide.md
    (accessed this session). **Confidence:** high. **Implication:** two adoptable mechanics our sibling formats
    may lack: (a) **exempt machine-dense blocks (diagrams, contracts) from the budget** — this is the token-cost
    lever; (b) **Proposal-share rule (40–60%)** — a shape constraint that resists a bloated Background.

27. **[direct] The lemieux style guide's prose/tables rule is explicit and two-sided.** Verbatim: "**Use
    tables** for dense reference data: API fields, status codes, configuration options. **Use prose with
    subheaders** for items needing explanation: risks, tradeoffs, design decisions. Tables compress information.
    When readers need to understand *why*, prose is better." Its `CHOPPY WRITING CHECK`: "Read each section
    aloud. If it sounds like a bulleted list read as sentences, it's choppy… Three or more consecutive short
    sentences (under 10 words each)."
    **Source:** same style guide. **Confidence:** high. **Implication:** the resolution to the prose/bullets
    fight is a *per-section-type* rule, not a global one. Machine contract fields → table/frontmatter;
    enumerations → lists; argument/rationale → prose.

28. **[direct] The lemieux guide's "RFC became a spec" warning list is the most concrete
    implementation-detail-exclusion test found.** Verbatim warning signs: "Table of Contents needed (too long);
    Multiple appendices; Config file examples (YAML, JSON configs, Kubernetes manifests); Full class
    implementations with methods; Retry/backoff code; Monitoring metric code; Operational runbooks; Heavy
    diagrams (lifelines, numbered arrows) for simple linear flows; 'Option A / Option B' in Proposal section
    (pick one, move alternatives to Abandoned Ideas); Week-by-week rollout schedules."
    **Source:** same style guide. **Confidence:** high. **Implication:** adoptable as a negative checklist an
    agent can execute mechanically (each item is a greppable artefact class).

29. **[direct] The lemieux Abandoned-Ideas format is prose-narrative, explicitly not a template — and it
    supplies a word range.** Verbatim: "Write each alternative as a short narrative (150-300 words), not a
    templated format… Avoid templated formats like 'What it was: … Why attractive: … Why rejected: …'" and
    "The reader should feel like they're hearing your thought process, not reading a form."
    **Source:** `references/rfc-template.md` and `references/rfc-style-guide.md`. **Confidence:** high.
    **Implication:** *direct contradiction* with the Tech Leads Club options-table contract (F23). Both are in
    the ecosystem; a format must pick a lane or define when each applies.

30. **[direct] A second independent agent-skill family exists with an "answer-first" contract and a
    Decision-weight field.** `new-rfc` (eugenelim/agent-ready-repo): opens with a "Reviewer brief" grid
    (Decision · Recommended outcome · Change if accepted · Affected surface · Stakes · Review focus · Not in
    scope) then "The ask"; renders decisions as a table `| ID | Question | Recommendation | Why | Decide by |
    Reviewer action |`; carries a `Decision weight` header field (light | standard | heavy) that right-sizes
    research depth and pre-handoff ceremony but "never licenses dropping a gate check"; mandates a gated
    research/de-risk phase before any body prose and a mandatory pre-handoff gate with an `adversarial-reviewer`
    dispatch "re-run until clean"; caps genuinely open questions at ~3; and requires citations be
    "fetched and confirmed" before use.
    **Source:** https://github.com/eugenelim/agent-ready-repo/blob/main/.claude/skills/new-rfc/SKILL.md
    (accessed this session via search excerpt; SKILL.md content quoted). **Confidence:** medium-high (I read a
    substantial verbatim excerpt but not the whole file). **Implication:** the **Reviewer brief / BLUF grid is
    a strong candidate for the RFC header** — it is the agent-era version of Calçado's "at a glance" header,
    optimised for a reviewer agent's first screen.

31. **[direct] Other published RFC-authoring skills, tabulated in §Field-contract comparison.** Distinct
    field contracts found beyond the above: NVIDIA OpenShell `.agents/skills/create-rfc` (repo-template-driven;
    "Preserve uncertainty in Open questions instead of silently deciding unknowns"; "Keep rejected or left-out
    designs in Alternatives, not Proposal"); Bruno125 `rfc-authoring` gist (engineer-focused: problem statement,
    scoped goals/non-goals, component + sequence diagrams, explicit API inventory, security, rollout/testing,
    risks; "Prefer concrete interfaces, flows, data, and failure modes over narrative"; numbered `## 1. …`
    sections "so feedback can reference sections precisely"); `jrepp/docuchango` template (frontmatter with
    REQUIRED FIELDS validated by tooling, `title: "RFC-XXX: …"` pattern, state enum); `mcpmarket` RFC Architect
    (11-section guided questionnaire, codebase exploration, exports to Notion); guybary's Octocode
    `octocode-rfc-generator` (RFC as an alternative to plan mode).
    **Sources:** listed in §Sources. **Confidence:** medium (search-result excerpts, not full files, for the
    last three). **Implication:** the ecosystem is *not* convergent on a single section list — it converges on
    a small mandatory core (problem, alternatives, recommendation, ownership/state) with divergent section
    names. Our format should therefore make the **field semantics** the contract and section headings
    optional.

32. **[direct] The only measured experiment on architecture-specification *format* for coding agents found a
    strong format × model interaction, with format mattering little on frontier models.** arXiv:2608.21747
    (v1 2026-08-22, Canedo): 5 informationally equivalent formats (informal prose, Mermaid+constraints+ADRs,
    OpenAPI, C4/Structurizr DSL, TypeScript interface contracts) × 6 models × 90 multi-turn agent trials.
    Verbatim: "On the strongest models (Sonnet 4.6, GPT-5), format barely matters (quality spread 0.17-0.92).
    On weaker models, format produces spreads of 0.83-2.42 points, with code-proximate formats (OpenAPI,
    TypeScript contracts) recovering most of the capability gap." And: "TypeScript contracts triple API route
    coverage for the weakest model (33% to 100%)." Also: "Mid-tier models can consume more tokens than frontier
    models for worse output when they enter compilation debugging loops."
    **Sources:** https://arxiv.org/abs/2608.21747 ; https://arxiv.org/html/2608.21747. **Confidence:** high
    (abstract + paper body read). **Implication [inference]:** for *adjudication* agents (frontier) format
    choice is a weak lever; for *cheap/weak* drafting or screening agents, structured, code-proximate formats
    are the strong lever. This supports spending the RFC's structure budget on the machine contract rather than
    on prose formatting. Caveat: this study measures spec→code generation, **not** proposal comprehension —
    it is adjacent evidence, and belongs to the sibling SPEC lane's territory if it is claimed there.

33. **[direct] Corroboration that agents over-concretize proposals — the pattern is named and guarded against
    in three independent agent-tooling codebases.** (a) GitHub `spec-kit` (spec-driven development): "The
    feature specification template explicitly instructs: ✅ Focus on WHAT users need and WHY ❌ Avoid HOW to
    implement (no tech stack, APIs, code structure)" and lists as a template-driven quality property "Prevent
    premature implementation details"; also "Any code samples, detailed algorithms, or extensive technical
    specifications must be placed in the appropriate `implementation-details/` file." (b) SuperSpec "Principle 3:
    Intent Over Implementation": "Premature implementation details constrain solutions", with a bad example
    naming a file path and an exported function vs a good example stating requirements only. (c) `prd-to-plan`
    skill guidance: "Prevent premature implementation detail… If the output starts naming exact files, classes,
    or low-level functions too early, redirect it."
    **Sources:** https://raw.githubusercontent.com/github/spec-kit/main/spec-driven.md ;
    https://asasugar.github.io/SuperSpec/concepts/first-principles.html ;
    https://agentskillsfinder.com/skills/prd-to-plan. **Confidence:** medium-high (quoted from fetched/echoed
    primary text; the spec-kit file was retrieved via search-provided raw content rather than a direct fetch).
    **Implication:** the over-concretization failure is corroborated by *tooling authors' countermeasures*,
    not by measurement. Treat it as a well-attested engineering observation, not an empirically quantified one.

34. **[direct] The AI-provenance debate has a quantitative side, and it points the other way from Calçado.**
    An empirical study of 1,000 popular GitHub repos found 118 AI policies: "78% of the AI policies allow
    AI-assisted contributions, while 22% explicitly discourage AI use… 51% of the AI policies require the
    disclosure of AI-assisted contributions; and 74% of the AI policies require a human in the loop during
    contribution." It also documents the failure mode: "we encountered vague and fuzzy terms regarding when
    disclosure is required, such as 'significant', 'substantial', and 'meaningful'."
    **Source:** https://arxiv.org/html/2605.16706v2 (arXiv:2605.16706v2). **Confidence:** high (direct quotes
    from the paper's HTML). **Implication:** if our RFC format carries any provenance field, it must be a
    **typed enum with a definition** (e.g. fully-generated / drafted-then-edited / human-led), because the
    literature shows free-text "significant AI use" is unusable.

35. **[direct] Named project policies supply concrete provenance enums and a hard limit on AI adjudication.**
    (a) sipeed/picoclaw requires three levels: "Fully AI-generated", "Mostly AI-generated", "Mostly
    Human-written". (b) hashicorp/terraform asks agent-submitted PRs to self-identify in the title "for
    expedited processing". (c) Blender's policy: "You must disclose the use of AI tools when a significant part
    of your contribution, descriptions or comments are taken from a tool… Disclosure may be a note in the
    design document, issue description or comment," with an `Assisted-by:` trailer, plus: "You must not use AI
    as the sole or final arbiter in making a substantive or subjective judgment on a contribution." (d) Fedora
    v0.1.4b-2 mirrors both, using RFC 2119 keywords ("contributors SHOULD disclose…"; "an AI MUST NOT be the
    sole or final arbiter… nor may it be used to evaluate a person's standing within the community").
    **Sources:** https://arxiv.org/html/2605.16706v2 (quotes from policies) ;
    https://devtalk.blender.org/t/ai-contributions-policy-proposal/44202 ; the Fedora policy text quoted in the
    same arXiv paper's appendix/search excerpts. **Confidence:** high for Blender and the paper's policy quotes;
    medium for the Fedora draft version number (a draft, versioned "v0.1.4b-2"). **Implication:** direct
    support for a **"no agent as sole arbiter"** clause in any RFC review/adjudication design — and note the
    Blender/Fedora wording explicitly contemplates design documents as a disclosure venue.

36. **[direct] A 2026 Internet-Draft-style RFC on AI contributors specifies a *detection heuristic list* that
    is directly about bullet-vs-prose.** Verbatim: "No reliable mechanism exists for determining whether a
    contribution was produced by an AC, and this document does not propose one. Heuristics in informal use
    include perfectly formatted markdown, a commit message in the imperative mood that runs to four paragraphs,
    the substitution of a bulleted list where a sentence would do, and a level of politeness not previously
    observed in the project. Maintainers report that human contributors have begun to exhibit all four."
    It also requires degree-accurate disclosure: "A fully generated patch MUST NOT be described as
    'AI-assisted', and unread output MUST NOT be described as 'reviewed'."
    **Source:** https://nesbitt.io/2026/05/21/rfc-artificial-contributors-to-open-source.md (2026-05-21).
    **Confidence:** high (read in full). **Implication:** **"a bulleted list where a sentence would do" is now
    a *stigmatised* pattern in review culture.** If our RFC format mandates bullets for reasoning, it will be
    read as machine-slop by human reviewers. This is a reputational cost, not a comprehension one — but it is
    the strongest single argument for prose in the argument sections.

37. **[direct] The silent-RFC / stale-document failure mode was quantified in the marzouk production system.**
    Verbatim: "nearly 60% of the ~20 RFCs created in the last nine months remained open with no updates for
    more than six weeks", with three named causes: visibility bias, hesitation to commit, and stale
    documentation ("Several RFCs had reached a conclusion, and the author acted on them but did not update the
    document to reflect the final state"). After the AI coordinator was deployed: "We were able to resolve the
    backlog of stalled RFCs within four weeks."
    **Source:** https://www.marzouk.io/posts/ai-rfc. **Confidence:** high. **Implication:** an RFC format that
    an agent reads must have a **state field the agent is expected to advance**, and the most common defect is
    a *stale* state, not a badly written one. (Lifecycle/expiry mechanics are a sibling lane's remit — flagged
    here only as a format-field requirement.)

38. **[inference] No controlled study of proposal-format (RFC) comprehension by agents was found.** Across
    four search rounds with varied angles I found no measured evaluation of RFC/proposal/design-doc *format*
    for agent comprehension. Nearest neighbours: arXiv:2608.21747 (architecture spec → code generation; F32),
    DesignQA (arXiv:2404.07917, engineering *rulebook* QA — retrieval/comprehension/compliance on a 140-page
    Formula SAE rule document), and ENGDESIGN (NeurIPS 2025, simulation-verified engineering design tasks).
    None compares proposal formats. **Sources:** as cited; search queries listed in §Missing evidence.
    **Confidence:** medium-high (absence of evidence over the queries run, not proof of absence).
    **Implication:** every prose-vs-bullets and word-budget recommendation in this brief is **not** backed by
    measured agent-comprehension data. Design decisions here must be justified by reliability, reviewability
    and slop-resistance — not by an appeal to measured comprehension.

39. **[direct] The "five convincing pages" problem and the budget do not conflict as much as they appear.**
    Calçado says LLMs let "the document arrive first: five convincing pages before anyone has had a single
    serious thought about the problem" — i.e. ~2,500 words of plausible prose. His own budget is 2,500 words.
    **[inference]** The budget therefore cannot be the defence against plausible-prose inflation; the rubric
    (F6) and the traceability rule (F8) are. The budget's distinct job is to force appendix placement.
    **Source:** https://philcalcado.com/2026/07/10/writing_ai_assisted_rfcs.html. **Confidence:** high for the
    quotes; the inference is mine. **Implication:** do not treat "word budget" and "anti-plausible-prose" as
    one control; they are two independent controls with different failure modes.

---

## Convergent moves — verdict

Legend: **adopt** / **adapt** (adopt with a change) / **reject** / **hold** (evidence insufficient).
"Independence" = whether the advocates plausibly converged independently. Derivative clusters are noted.

| # | Move | Who advocates it | Independence | Verdict | Why |
|---|------|------------------|--------------|---------|-----|
| 1 | **Structured header / frontmatter as machine contract** | Marzouk (F15, production extractor); Tech Leads Club `create-rfc` (F23); Lubow template (metadata table); lemieux (Draft Status + Abstract); Calçado 2018 header (Authors/To-be-reviewed-by/Revisit Date/State — pre-AI); eugenelim `new-rfc` Reviewer-brief grid + Decision-weight field (F30) | **Independent** — Marzouk's is an operational need; TLC's and lemieux's are separate skill families; Calçado's predates AI by 8 years | **Adopt** | Only move with *production* evidence of a reliability benefit (extraction without hallucination). It is also the cheapest token-wise: N typed fields vs a prose paragraph the agent must parse. |
| 2 | **Rubric pre-flight (agent-runnable quality gate)** | Calçado (F6, 5 check categories); Marzouk (F17, 5 structural criteria); lemieux reviewer prompt (F25, 11 categories × severity × conf ≥ 80); eugenelim `new-rfc` pre-handoff gate + adversarial reviewer (F30) | **Independent** — four separate implementations, three different mechanisms (rubric doc, dashboard job, subagent prompt, gated workflow) | **Adopt** | Highest-confidence convergent move in the corpus. Copy lemieux's *mechanics* (category, severity, confidence threshold, quote-the-location, never-suggest-fixes) and Calçado's *scope* (private author-side pre-flight, explicitly not a grade). |
| 3 | **Evidence traceability (claim↔source, machine-checkable)** | Calçado (F8, explicit reviewer-side machine check); eugenelim `new-rfc` ("fetch it and confirm it resolves *and* contains the borrowed claim… Never pass an unverified citation through"); Marzouk (structure to avoid hallucination); lemieux (API contracts must show real shapes, not prose) | **Independent** | **Adopt** | The only anti-hallucination control that scales to an agent reviewer. Requirement is stricter than "cite a URL": the pointer must be resolvable *and* the cited text must be checkable against the claim. |
| 4 | **Implementation-detail exclusion** | Calçado 2018 + 2026 (F5, F11); Raviv (F12, RFC vs plan); Lubow (F20); lemieux "RFC became a spec" list (F28); Tech Leads Club (F24, RFC→TDD handoff); OpenShell ("rejected/left-out designs in Alternatives, not Proposal"); spec-kit / SuperSpec / prd-to-plan (F33) | **Independent, and pre-dates AI** (Calçado 2018) | **Adopt, with a positive allowance** | Most corroborated move, and the only one with a named failure story (the SoundCloud Python paragraph). But every source allows *shape* — interface, pseudocode, toy example, real API request/response shapes. A prohibition-only rule would strip information reviewers need, which is exactly the mistake our sibling prohibition-first formats must avoid here. |
| 5 | **Anti-plausible-prose (slop detection)** | Calçado (F4, F6); lemieux `AI_PATTERNS` + `CHOPPY_WRITING` (F25, F27); eugenelim ("AI-obvious patterns"); nesbitt heuristics incl. "bulleted list where a sentence would do" (F36); Blender ("quality bar for AI generated contributions is higher") | **Independent** | **Adopt** | This is a *detection* control, distinct from #4 and #6. Note the asymmetric risk: a false positive costs an author an edit; a false negative costs a reviewer a wasted read. Bias toward flagging. |
| 6 | **Word / length budget** | Calçado (2,500 max, 1,500–2,000 target, F2); lemieux (500–1000 lines, Proposal 40–60%, per-section word ranges, exempting diagrams/contracts, F26); Lubow (over-indexing cost test, F20); Raviv ("keep the prose short", F14) | **Independent** on the *principle*; **divergent on units** (words vs lines vs section share vs time-cost) | **Adapt** | Adopt the principle, but as a **declared budget field + per-section shares**, not a fixed constant — Calçado himself says "I'm not religious about those numbers; I'm religious about having a budget." Also steal lemieux's exemption: machine-dense blocks (diagrams, contracts) don't count, which is the token-cost lever. |
| 7 | **Mandatory do-nothing alternative** | Calçado (F3, "the only part of NABC I'm really married to"); Tech Leads Club (F23, "Do nothing is always an option", first-class); Sourcegraph handbook formats ("What happens if we do nothing" in the tension format); Raviv's Q4 cost-vs-alternatives (F12); Lubow's Documented-No (F21) | **Independent** | **Adopt + pair with a straw-man check** | Uniquely cheap to enforce (one validated field). But Calçado's *own* rubric flags "alternatives that are obviously straw men" — so the do-nothing slot must be paired with the rubric check, or an agent will fill it with a token paragraph. |
| 8 | **AI-provenance stance** | Calçado: **disclosure is less useful**, accountability is the rule (F10); Lubow: useful for drafting, disqualifying if you don't understand it (F22). Counter: 51% of 118 OSS AI policies require disclosure / 74% require human-in-the-loop (F34); Blender + Fedora + nesbitt require degree-accurate disclosure with `Assisted-by:` and ban AI as sole arbiter (F35, F36); Ghostty/stdlib close undisclosed-but-suspected PRs | **Contested, not convergent** — Calçado's position is a minority of one among the human authors, and is contradicted by the only quantitative evidence found | **Adapt: adopt accountability + a typed provenance enum; do not adopt "disclosure is unnecessary"; reject any AI-as-sole-arbiter design** | Accountability is uncontested (Calçado, Lubow, Blender, Fedora, the 74% human-in-the-loop figure). Disclosure is contested, so make it a cheap typed field rather than a prose obligation: the empirical study shows free-text "significant AI use" thresholds are unusable, and picoclaw's three-level enum is the working alternative. The "no AI as sole arbiter" clause is directly relevant to an adjudication agent. |

**Overall verdict:** the four moves with production or measured backing are the **structured header** (#1,
Marzouk), the **rubric pre-flight** (#2, four independent implementations), **implementation-detail exclusion**
(#4, ten years of pre-AI practice plus three agent-tooling countermeasures), and the **state field** implied by
F37. The prose-over-bullets rule (#5's *rationale*) is the weakest-justified move for our purposes — it is a
human-reviewer and slop-signal argument, and our sibling ADR/SPEC contracts chose bullets+fragments. The
reconciliation offered by lemieux (F27) is the one to copy: **table/frontmatter for the machine contract,
prose for claims/rationale/alternatives, lists for enumerations** — a per-section-type rule rather than a
global one.

---

## Field-contract comparison table — agent-skill ecosystem

| Source (type) | Mandatory / validated fields | Options contract | Header / state | Quality gate | Length | AI stance |
|---|---|---|---|---|---|---|
| **Tech Leads Club `create-rfc`** (Claude/Cursor/Copilot skill; highest-volume) | title; Background/context; Driver; Approver(s); Impact HIGH/MED/LOW; ≥1 assumption + confidence; ≥2 decision criteria + weights *before* options; ≥2 options (incl. do-nothing); recommended option tied to criteria | ≥2 options + comparison matrix; do-nothing first-class | Header & Metadata table (Impact, Status, Driver, Approver, Contributors, Informed, Due Date) | 13-item checklist + 5 named anti-patterns; Outcome left as placeholder | "Be concise in options"; no numeric budget | RFC≠TDD; "RFC is for decisions, not implementation"; hand off to TDD for implementation |
| **eric.lubow `rfc-generator`** (Claude skill + Confluence template) | five prompt fields: problem, proposed solution, why, discarded alternatives, risks | Alternatives considered (why not picked) | Metadata table: Name, Summary (2–3 sentences), Status (Draft→In Progress→Feedback Requested→Accepted/Rejected), Author(s), RFC Sponsor, Proposed Date, Feedback Due By, Feedback Requested From | "Use what applies—skip sections that don't"; no placeholder/N/A sections; template "exists to prompt thinking, not create busywork" | "If it takes longer to document the change than it would to code it, you're over-indexing on detail" | "Useful for drafting… not to generate the ideas"; AI organizes the engineer's thinking |
| **lemieux/rfc-skills** (Claude Code plugin: writer + reviewer subagents) | Abstract (2–3 sentences); Background (checklist: current state, why insufficient, constraints, stakeholders); Proposal; Abandoned Ideas | ONE solution in Proposal; alternatives → Abandoned Ideas, 150–300 words each, narrative not template | `## Draft Status` block with State + `<!-- REVIEW: … -->` items; inline markers duplicated there | Reviewer prompt: 11 categories, MAJOR/MODERATE/MINOR, confidence ≥80, never suggest fixes; 30+ item review checklist | 500–1000 lines; Proposal 40–60%; diagrams/contracts exempt; full code counts | Reviewer explicitly flags `AI_PATTERNS` (em dashes, "Let's dive in", "It's worth noting", rhetorical questions) |
| **eugenelim `new-rfc`** (repo skill) | Reviewer brief grid (Decision · Recommended outcome · Change if accepted · Affected surface · Stakes · Review focus · Not in scope); "The ask"; decisions table `ID/Question/Recommendation/Why/Decide by/Reviewer action`; `Decision weight` (light/standard/heavy) | per-decision options with trade-offs + recommended answer + owner + decide-by; "A bare list of option *names* is not decidable" | Header fields incl. `Decision weight`; open questions capped at ~3 | Mandatory gated research/de-risk phase before body prose; mandatory pre-handoff gate; `adversarial-reviewer` re-run until clean; each item "executed and its result recorded, never self-certified" | "aim for ≤…" (truncated in retrieved excerpt) | "Cite as you go… fetch it and confirm it resolves *and* contains the borrowed claim" |
| **NVIDIA OpenShell `.agents/skills/create-rfc`** (repo skill) | repo `rfc/0000-template/README.md` is the source of truth; front matter (author, `state: draft`, related links) | "Keep rejected or left-out designs in Alternatives, not Proposal" | YAML front matter: author, `state: draft`, related links | Verify every template section present or intentionally marked not-applicable; RFC number/folder consistency | section "suggested section length" from template | "Preserve uncertainty in Open questions instead of silently deciding unknowns"; ask the user if a missing decision blocks coherence |
| **Bruno125 `rfc-authoring`** (gist skill) | Executive Summary (5–10 lines incl. highest-risk assumptions); Background/Context; Proposed Architecture Overview + component map (existing vs proposed); Problem Statement; Goals/Non-Goals; Key Decisions | Key decisions written **as questions**, each with options + pros/cons + decision | Short header + metadata bullets (status/authors/date); numbered `## 1. …` sections "so feedback can reference sections precisely" | Review checklist (7 items) | Executive summary 5–10 lines; no global budget | "Prefer concrete interfaces, flows, data, and failure modes over narrative"; "Avoid meta commentary about the writing process" |
| **`jrepp/docuchango` `rfc-template.md`** (tooling-validated template) | frontmatter "REQUIRED FIELDS - All must be present for validation to pass"; `title` must match `"RFC-XXX: …"`; status enum from `Draft` | "Alternatives Considered"; "Trade-offs and Implications" | YAML frontmatter, machine-validated | Validation pass/fail on required fields | none | none stated |
| **mcpmarket "RFC Architect"** (Claude Code skill) | 11-section guided questionnaire (goals, engineering design, drawbacks, cost analysis, business motivation…) | guided section-by-section questioning | exports to Notion | automated codebase exploration for dependencies/constraints/prior art | none | integrates with codebase to gather context |
| **guybary Octocode `octocode-rfc-generator`** (skill) | Summary; Motivation/problem; Proposed design; Alternatives considered; Drawbacks and risks; Open questions; Recommendation; Implementation plan | alternatives + drawbacks + recommendation | none stated | "forces the agent to answer the harder questions first" | none | positioned explicitly against plan mode: RFC is "a structured way to think before execution" |
| **Sourcegraph handbook RFC formats** (not a skill; de-facto org standard) | per-format required attributes; default template = "just the required attributes" | format-dependent; tension format has "What happens if we do nothing" | **Tags** for people + roles (decider, input providers, approvers, approvals) and affected teams | separate formats documented outside the template "so we keep RFCs lightweight and flexible" | lightweight default | none stated |
| **Adjacent: `spec-kit` / SuperSpec / `prd-to-plan`** (spec/PRD tooling, transferred) | WHAT and WHY only; "❌ Avoid HOW to implement (no tech stack, APIs, code structure)"; explicit uncertainty markers | — | template + constitution gates | "Prevent premature implementation details"; "Force explicit uncertainty markers"; gates require documented justification for complexity | code samples/algorithms must go to `implementation-details/` | templates constrain AI behaviour; "Template-Driven Quality" |

**Reading of the table:** the ecosystem converges on a *semantic* core — problem, alternatives (incl.
do-nothing), recommendation + rationale, ownership/state, a length-or-shape limit — and diverges on section
names, on whether the document presents one decided design (lemieux, Bruno125) or an options matrix for a
decider (Tech Leads Club), and on whether "Open Questions" is a legitimate section (banned by lemieux;
required-by-design in OpenShell; capped at 3 in `new-rfc`).

---

## Contradictions

1. **Prose vs bullets for the argument — the sharpest conflict, and it cuts against our sibling formats.**
   *Prose side:* Calçado ("The argument is narrative prose, not bullet points", F4); lemieux (bans
   "three or more consecutive short sentences under 10 words", ships `CHOPPY_WRITING` as a MODERATE defect,
   and `TABLE_OVERUSE` for risks/tradeoffs, F27); nesbitt (lists "the substitution of a bulleted list where a
   sentence would do" as an AI tell, F36); Blender (higher quality bar for AI-generated contributions, F35).
   *Structured side:* Tech Leads Club mandates tables, matrices and weighted-criteria lists (F23); Bruno125
   wants numbered sections + API inventories + sequence diagrams and "concrete interfaces… over narrative";
   Marzouk's extractor wants a table and uses emoji-in-title (F15, F16); Sourcegraph uses tags; and our sibling
   ADR/SPEC contracts chose bullets+fragments.
   **Resolution offered by the evidence [interpretation]:** the two camps are not arguing about the same
   region of the document. Every "prose" advocate reserves tables for dense reference data and enumerations;
   every "structured" advocate still writes rationale. lemieux states the rule explicitly (F27). What the
   evidence does **not** contain is any measurement of agent comprehension for prose vs bullets (F38) — so
   the prose advocacy rests on *human* reviewer behaviour and on *slop signalling*, both of which matter here
   because humans review these RFCs and adjudication agents must not be fooled by fluent emptiness.
   **Unresolved:** no source addresses whether a *prose* rule survives an all-agent drafting pipeline, where
   the slop signal is worthless (there is no human "thinking behind the prose" to detect).

2. **AI disclosure: "less useful" vs the majority empirical position.**
   Calçado: "As AI becomes a normal part of producing these documents, disclosure becomes less useful to me"
   (F10). Against: 51% of 118 OSS AI policies require disclosure (F34); Blender, Fedora, nesbitt,
   Ghostty/stdlib require degree-accurate disclosure and reserve the right to close undisclosed contributions
   (F35, F36). Note these are *code-contribution* policies, not design-document policies — but Blender and
   Fedora explicitly name design documents as a valid disclosure venue, so the conflict is live for RFCs.
   **Unresolved:** no source tests whether a provenance label changes reviewer behaviour on a *proposal*
   document.

3. **"Open Questions" is either a required honesty device or proof of unfinished work.**
   lemieux: "Do NOT include an 'Open Questions' section… An 'Open Questions' section signals the author didn't
   finish their job" (F26). OpenShell: "Preserve uncertainty in Open questions instead of silently deciding
   unknowns" (F31). `new-rfc`: cap genuinely-open questions at ~3, each with owner + decide-by (F30). Raviv
   implicitly tolerates unresolved problems, since the whole point is to discover what you don't understand
   (F13). **Unresolved:** which is right depends on whether the RFC is a *proposal seeking a decision*
   (lemieux's model: decide first, then propose) or a *thinking artefact* (Raviv's model). Our format must
   pick, and the choice is downstream of whether the RFC is drafted by a human-with-context or an agent.

4. **One decided design vs an options matrix.** lemieux: "Present ONE solution in the Proposal section… 'Option
   A / Option B' in Proposal section" is a *warning sign*, and the tone should be "I thought this through,
   here's the design" (F26, F28). Tech Leads Club: ≥2 options + a comparison matrix + weighted criteria, with a
   named anti-pattern of "predetermined conclusion disguised as RFC" (F23). Calçado's NABC wants alternatives
   compared (F3) but his prose rule implies a single argued case. **Partly reconcilable:** lemieux's
   Abandoned-Ideas narrative *is* a alternatives-with-rejection-reason record; the difference is where the
   alternatives live and whether the recommendation is decided before review.

5. **Time estimates: banned vs required.** lemieux: "Never include time estimates… An RFC proposes what to
   build, not when" (F26). Tech Leads Club: Due Date in the header, "Estimated Cost (effort/complexity/
   monetary)" as a recommended section, action items with owners *and due dates* (F23). Lubow: timebox a
   questionable idea via "a very short PoC in the milestones of the RFC document" (F21). **Unresolved:**
   no source adjudicates; likely a genuine difference between an approval-gated RFC (dates matter) and an
   advisory RFC (dates don't).

6. **Marzouk's cited statistic is a misstatement of its source.** He writes "AI loses 39% accuracy when given
   6 tasks" citing arXiv:2505.06120; the paper reports a 39% average drop for **multi-turn vs single-turn**
   across six generation tasks, not a task-count penalty (F19). Both texts verified directly.
   **Impact:** do not propagate the claim; the paper's actual mechanism (early wrong turns ossify into
   over-relied-on assumptions) is arguably *more* supportive of a written up-front proposal.

7. **Calçado's "five convincing pages" vs his own 2,500-word budget** — see F39. Not a contradiction in his
   text (he never claims the budget prevents slop), but it is a trap for a reader who adopts the budget as the
   anti-slop control.

---

## Missing evidence

- **Calçado's actual rubric text and 2026 RFC template** — the Google Doc is unfetchable (JS-rendered/auth);
  no mirror found. Only the five check categories in the post are citable. *Targeted search for a mirror
  returned nothing.*
- **No measured evaluation of proposal/design-doc format for agent comprehension** (F38). Queries run:
  `benchmark study LLM comprehend engineering design document format bullets vs prose 2025 2026`;
  `"design doc" LLM agent readability evaluation measured experiment prose vs structured markdown`;
  `no benchmark measures LLM comprehension of RFC vs design doc format controlled study`;
  `measured evaluation "design doc" format LLM comprehension tokens 2026`.
  Nearest: arXiv:2608.21747 (spec→code), DesignQA (rulebook QA), ENGDESIGN (design tasks). **Confirmed negative
  for the queries run, not proof of absence.**
- **No measurement of the do-nothing alternative's effect** on proposal quality, human or agent.
- **No evidence on whether a typed provenance enum changes reviewer/adjudicator behaviour** on a proposal doc.
- **Marzouk's extractor claim is asserted, not measured** — no accuracy numbers for the header table or the
  emoji convention (he himself hedges the emoji benefit with "(slightly)").
- **The "39% accuracy" mechanism** is often mis-cited across the blogosphere (F19); no source I found corrects
  it publicly.
- **No source addresses the all-agent drafting case** — every human author's quality argument (F4, F22, F36)
  depends on a human author whose understanding can be tested in a meeting. Our pipeline may not have that
  meeting.
- **`new-rfc`'s full SKILL.md length budget** ("aim for ≤…") was truncated in the retrieved excerpt.

---

## Sources

**Kept (primary, decisive):**
- Writing AI-Assisted RFCs — Phil Calçado (https://philcalcado.com/2026/07/10/writing_ai_assisted_rfcs.html) — the six format deltas, the word budget, the rubric categories, the implementation-detail boundary, prose-not-bullets. Read in full.
- A Structured RFC Process — Phil Calçado (https://philcalcado.com/2018/11/19/a_structured_rfc_process.html) — proves implementation-detail exclusion and the header table pre-date AI; the SoundCloud failure story. Read in full (first ~30k chars).
- Nutrition Labels for AI — Phil Calçado (https://philcalcado.com/2026/05/31/nutrition_labels_for_ai.html) — the two-value engagement label that replaces disclosure in his model.
- Let Them Write RFCs — Ohad Raviv (https://ohadravid.github.io/posts/2026-08-let-them-write-rfcs/) — four questions, RFC-vs-plan, V1/V2 iteration, flexible structure, "keep the prose short". Read in full.
- AI as Process Infrastructure — Marwan Marzouk (https://www.marzouk.io/posts/ai-rfc) — the only production extractor evidence; header table, emoji state, five quality criteria, keep-AI-out-of-content. Read in full.
- Implementing an RFC Process That Engineers Don't Hate — Eric Lubow (https://eric.lubow.org/2026/implementing-an-rfc-process-that-engineers-dont-hate/) — five elements, over-indexing rule, Documented-No, silent RFC, AI-drafting stance. Read in full.
- RFC Template + Claude Skill for RFCs — Eric Lubow (https://eric.lubow.org/references/rfc-documents/rfc-template/, .../claude-skill-for-rfcs/) — the metadata table and the skill's "include only sections with actual content, no N/A" rule.
- tech-leads-club `create-rfc` SKILL.md (https://github.com/tech-leads-club/agent-skills/blob/main/packages/skills-catalog/skills/(creation)/create-rfc/SKILL.md) — the ecosystem's mandatory-field contract. Read in full via raw fetch.
- lemieux/rfc-skills — repo + `references/rfc-style-guide.md`, `references/rfc-template.md`, `references/rfc-reviewer-prompt.md`, `agents/rfc-reviewer.md` (https://github.com/lemieux/rfc-skills) — the only two-agent RFC pipeline with a machine-checkable defect taxonomy; the prose/table rule and the "became a spec" list.
- eugenelim `new-rfc` SKILL.md (https://github.com/eugenelim/agent-ready-repo/blob/main/.claude/skills/new-rfc/SKILL.md) — Reviewer-brief grid, Decision weight, gated pre-handoff checks, cite-and-confirm rule.
- NVIDIA OpenShell `create-rfc` (https://github.com/NVIDIA/OpenShell/blob/main/.agents/skills/create-rfc/SKILL.md) — Open-questions preservation; alternatives placement.
- Bruno125 `rfc-authoring` gist (https://gist.github.com/Bruno125/97fd9dcad9fca0f226816618232b28a1) — decisions-as-questions contract; "concrete over narrative".
- Architecture as Capability Equalizer for Coding Agents — arXiv:2608.21747 (https://arxiv.org/abs/2608.21747) — the only controlled format experiment relevant to agent consumption (spec→code; not proposals). Transferred/adjacent.
- LLMs Get Lost In Multi-Turn Conversation — arXiv:2505.06120 (https://arxiv.org/abs/2505.06120) — the real content behind Marzouk's mis-cited 39%.
- AI Policy, Disclosure, and Human in the Loop — arXiv:2605.16706v2 (https://arxiv.org/html/2605.16706v2) — 51% disclosure / 74% human-in-the-loop across 118 policies; the three-level provenance enum; the "fuzzy terms" failure.
- RFC: Artificial Contributors to Open Source — nesbitt.io (https://nesbitt.io/2026/05/21/rfc-artificial-contributors-to-open-source.md) — the AI-detection heuristic list including bulleted-list-where-a-sentence-would-do. Read in full.
- Blender AI Contributions Policy Proposal (https://devtalk.blender.org/t/ai-contributions-policy-proposal/44202) — disclosure + "no AI as sole arbiter", design documents named as a disclosure venue.
- sourcegraph handbook RFC formats (https://github.com/sourcegraph/handbook/blob/main/content/company-info-and-process/communication/rfcs/index.md) — multiple formats outside the template; tags-for-roles; do-nothing variant.

**Rejected / deprioritised:**
- explainx.ai mirror of `create-rfc` — byte-identical to the GitHub original, wrapped in generic SEO boilerplate; kept only as evidence of *distribution reach* (agents reach it via explainx, Claude Code, Cursor, Copilot, Windsurf, Codex, Goose, Zed).
- LinkedIn reshares of Calçado (activity posts) — secondary restatements; used only to confirm reach.
- mcpmarket "RFC Architect", guybary Medium (Octocode), agentskillsfinder (`prd-to-plan`) — marketplace/blog aggregators; treated as ecosystem-signal only (medium confidence), not as field-contract evidence.
- DesignQA (arXiv:2404.07917) and ENGDESIGN (NeurIPS 2025) — engineering *rulebook QA* and *simulation-verified design tasks*; not proposal-format comprehension. Cited only to bound the negative result.
- Unicode UTS #51 emoji spec — search noise from the emoji query; irrelevant.
- Rust/PEP/Gerrit/Google RFC templates — prior art explicitly assigned to sibling lanes; cited only where a source referenced them (Raviv's footnote 4).

---

## Next steps

1. **Obtain the Calçado 2026 Google Doc** by an authenticated/manual route (browser fetch with a Google session,
   or asking the author). It is the only cited artefact in this lane whose full content is unread, and it
   contains the exact rubric — the single most transplantable mechanism found.
2. **Run one decisive local experiment** rather than searching further: take 3–4 candidate RFC shapes
   (prose-argument vs bullets-argument, table header vs YAML frontmatter, with/without rubric section) and
   measure agent comprehension on a fixed question set (the four Raviv questions + the five Calçado rubric
   checks) at equal token budgets. Nothing in the literature answers this, and it is cheap to test in-house.
3. **Adjudicate the Open-Questions and time-estimate conflicts against our own pipeline**, not the literature:
   they resolve on whether our RFC is drafted by a context-rich agent (then open questions are a smell) or by
   a context-poor one (then they are a required uncertainty channel).
4. Only if a *design-document* policy question becomes blocking: check Ghostty's `AI_POLICY.md` and the
   stdlib-js issue #9347 thread as design-doc-adjacent provenance precedents (both surfaced but were not read
   in full).

---

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Concise, well-sourced research brief written to the authoritative output path: 39 numbered findings each labelled [direct]/[interpretation]/[inference] with source URL, observation-date convention, confidence and a one-line design implication; an 8-move convergent-moves verdict table with per-move independence and adopt/adapt/reject calls; a 10-source field-contract comparison table across the agent-skill ecosystem; 7 recorded contradictions; an explicit missing-evidence section including the confirmed-negative measured-evaluation search; and kept/rejected source lists. Residual risks disclosed inline (no source_check tool available; Calçado Google Doc unfetchable; no measured proposal-format comprehension study exists)."
    }
  ],
  "changedFiles": [
    ".pi/research/rfc-templates/lane-agent-era.md"
  ],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "fetch_content (primary sources: philcalcado 2026 + 2018 + nutrition-labels, ohadravid, marzouk, lubow + template + Claude skill, lemieux rfc-skills raw files, Tech Leads Club create-rfc SKILL.md, arXiv 2608.21747, arXiv 2505.06120, nesbitt.io, Google Doc template attempt)",
      "result": "passed",
      "summary": "All named primary sources read in full except the Calçado 2026 Google Doc (JS-rendered, error returned) and the 2018 Google Doc."
    },
    {
      "command": "web_search (11 queries across 4 rounds: direct answer, authoritative source, practical/production experience, measured-evaluation negative, over-concretization corroboration, ecosystem field contracts, AI-disclosure controversy)",
      "result": "passed",
      "summary": "Surfaced the agent-skill ecosystem (lemieux, new-rfc, OpenShell, Bruno125, docuchango, RFC Architect, Octocode), the arXiv format experiment, and the AI-policy empirical study."
    },
    {
      "command": "get_search_content (stored-content retrieval + findText probes on the lemieux style guide and marzouk post)",
      "result": "passed",
      "summary": "Extracted the style guide's length/section budgets, prose-vs-table rule and 'became a spec' list; confirmed the emoji-in-title and header-table quotes verbatim."
    }
  ],
  "validationOutput": [
    "Verified verbatim (direct quotes retrieved from the primary documents): Calçado word budget 2,500/1,500-2,000; the five rubric check categories; the 'once we're debating filenames, method signatures, migration sequencing, or the exact DDL' boundary; the disclosure stance; Raviv's four questions and RFC-vs-plan; Marzouk's header-table and emoji quotes and his five quality criteria; Lubow's five elements and over-indexing rule; Tech Leads Club's nine mandatory fields; lemieux's reviewer categories, confidence>=80 gate, 500-1000 line budget, Proposal 40-60%, and the 'became a spec' list; arXiv:2608.21747 abstract (format x model interaction, 0.17-0.92 spread on frontier vs 0.83-2.42 on weaker models, 33%->100% route coverage); arXiv:2505.06120 abstract (39% multi-turn drop across six generation tasks).",
    "Contradiction verified directly: Marzouk's 'AI loses 39% accuracy when given 6 tasks' vs arXiv:2505.06120's actual finding (multi-turn vs single-turn drop) - both texts fetched independently.",
    "Negative result recorded: no controlled study of proposal/RFC format comprehension by agents found across four query angles; nearest neighbours identified (arXiv:2608.21747 spec->code, DesignQA rulebook QA, ENGDESIGN design tasks)."
  ],
  "residualRisks": [
    "No source_check tool was exposed to this lane, so no third-party claim validation was possible; all validation was direct primary-document reading. Decision-critical claims are quoted verbatim to compensate.",
    "The Calçado 2026 RFC template and rubric Google Doc could not be fetched (JavaScript-rendered/auth-gated) and no mirror was found; only the five rubric check categories quoted in the blog post are citable. The exact rubric questions remain unknown.",
    "The measured-evaluation negative (finding 38) is an absence-of-evidence result bounded by the queries run, not proof of absence; a study behind a paywall or using different vocabulary (e.g. 'context engineering', 'spec quality') could exist.",
    "Several ecosystem field contracts (RFC Architect, Octocode, prd-to-plan, and the eugenelim new-rfc file in full) were read only via search-provided excerpts rather than full direct fetches; those rows are marked medium confidence.",
    "Prose-vs-bullets evidence is entirely human-reviewer and slop-signal based; it does not transfer to an all-agent drafting pipeline, and the brief flags this as unresolved rather than resolving it.",
    "Freshness: the corpus is 2018-2026 with the agent-era material concentrated in 2026; findings tied to specific model capabilities (e.g. Calçado's diagram hack, arXiv:2608.21747's model tiers) are self-declared as perishable."
  ],
  "noStagedFiles": true,
  "diffSummary": "Added one new research brief at .pi/research/rfc-templates/lane-agent-era.md (read-only lane; no other file touched, nothing staged).",
  "reviewFindings": [
    "no blockers"
  ],
  "manualNotes": "Key headline for the parent: the strongest-justified moves are the structured header (Marzouk's production extractor is the only operational evidence), the rubric pre-flight (four independent implementations, with lemieux's category+severity+confidence>=80 mechanics being the most transplantable), implementation-detail exclusion (pre-AI, ten years of practice, plus three agent-tooling countermeasures), and a state field the agent advances (60% of marzouk's RFCs went stale, not badly written). The prose-over-bullets rule is the weakest-justified move and directly contradicts our sibling ADR/SPEC bullets+fragments choice; the reconciliation the evidence actually supports is per-section-type (table/frontmatter = machine contract, prose = claims/rationale/alternatives, lists = enumerations) - lemieux states this rule explicitly. AI-disclosure is genuinely contested (Calçado says it is becoming useless; the only quantitative study found says 51% of OSS AI policies require it and 74% require human-in-the-loop), so recommend accountability + a typed provenance enum rather than adopting either extreme, plus a 'no agent as sole arbiter' clause now that Blender and Fedora both state that rule explicitly. One transferable caveat for the sibling lanes: arXiv:2608.21747 finds format barely matters on frontier models (spread 0.17-0.92) but hugely on weak ones (0.83-2.42) - worth checking it is not double-claimed by the SPEC lane."
}
```
