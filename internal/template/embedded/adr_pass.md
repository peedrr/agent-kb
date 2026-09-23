---
type: adr
title: Adopt CEL-Based Validation for Typed Page Templates
summary: Typed pages are gated by CEL rules at write time and at sweep time, with a guard doctrine that keeps every optional read honest.
tags: [validation, templates, cel]
status: accepted
deciders: akb maintainers
created: 2026-01-05
updated: 2026-02-10
sources: [internal/skill/embedded/kb-management/references/TEMPLATE.md, internal/cel/pagebuilder.go]
---

# Adopt CEL-Based Validation for Typed Page Templates

This page is the pass mockup of the `adr` template, so it is also the template's own rationale: the rules it satisfies are described below, grouped by the CEL concept each one demonstrates, and annotated with a `PASSES` HTML comment naming the rule.

## Context

A typed page is only as trustworthy as the rules that gate it, and a rule that reads a key the page does not carry is worse than no rule at all: it either fails the page for a reason the author cannot see, or it passes the page vacuously. Both defects followed from that one ambiguity. Authors wrote presence guards for keys the schema already guarantees, which let a typo slip through as a skipped rule, and they read optional keys directly, which turned an absent `updated` into an evaluation error instead of a rule that simply does not apply.

We settled the question at both ends of the pipeline. `akb write` refuses a page that leaves out a schema-required field before any CEL rule runs, so a rule may read a required key unguarded; a rule that still cannot be evaluated is a fail-closed validation failure, exit 1, naming the rule and the key that broke it. The lint sweep, which sees pages that never passed `akb write` — pulled from git, or ingested by an index rebuild — does the opposite: an unevaluable rule becomes one `cel_lint` issue against one page and the sweep continues, so a single broken rule cannot hide every other finding.

`akb template write` closes the gap at authoring time. It evaluates the pass mockup as given, once per schema-optional key the mockup supplies with that key removed, and once with `old_page` set to the mockup itself, so a rule that only holds on a create, or only holds while an optional key is present, fails the template write instead of the first real page.

<!-- PASSES: require_context rule — the AST contains an H2 heading with text exactly "Context" -->
<!-- PASSES: require_title rule — frontmatter carries a non-empty title -->
<!-- PASSES: temporal_created rule — created (2026-01-05) is in the past, so timestamp(created) <= now evaluates to true -->
<!-- PASSES: min_word_count rule — the body holds far more than the 50 words the threshold demands -->

## Decision

The schema marks `title`, `type`, `summary`, `tags`, `status`, `deciders`, `created`, and `updated` required, and `sources` and `supersedes` optional. Each rule below names the CEL concept it exists to demonstrate, so an author copying this template learns the shape of the language rather than a single house style.

### Presence and string shape

- `require_title` reads the schema-required `title` with no `has()` guard, because presence is enforced before the rules run.
- `title_no_period` is a suffix predicate: `!page.frontmatter.title.endsWith(".")`.
- `title_style` combines a regex match with a string size: `page.frontmatter.title.matches("^[A-Z]") && size(page.frontmatter.title) <= 80`.

<!-- PASSES: title_no_period rule — the title ends with "Templates", not a period -->
<!-- PASSES: title_style rule — the title opens with "A" and is shorter than 80 characters -->

### Enumerations and lists

- `valid_status` is a membership test over the status enum. It reads `status` unguarded because the schema requires it.
- `tags_nonempty` is a list size check: `size(page.frontmatter.tags) > 0`.
- `tags_format` is a universal quantifier with a per-item regex: `page.frontmatter.tags.all(t, t.matches("^[a-z0-9-]+$"))`. An empty tag list satisfies `all` vacuously, which is exactly why `tags_nonempty` sits beside it.

<!-- PASSES: valid_status rule — status is "accepted", a member of ["proposed", "accepted", "deprecated", "superseded"] -->
<!-- PASSES: tags_nonempty rule — tags holds three entries -->
<!-- PASSES: tags_format rule — every tag (validation, templates, cel) matches ^[a-z0-9-]+$ -->

### Document structure

- `require_context`, `require_decision`, and `require_consequences` are AST collection existence checks over H2 headings, written as `page.ast.headings.exists(h, h.level == 2 && h.text == "Context")`.
- `disallow_options` and `disallow_pros_cons` are the negation of the same check, and they keep the decision record from drifting into a survey of alternatives that were never chosen.

<!-- PASSES: require_decision rule — the AST contains an H2 heading with text exactly "Decision" -->
<!-- PASSES: require_consequences rule — the AST contains an H2 heading with text exactly "Consequences" -->
<!-- PASSES: disallow_options rule — no H2 "Options" heading exists, so the negation holds -->
<!-- PASSES: disallow_pros_cons rule — no H2 "Pros and Cons" heading exists, so the negation holds -->

### Content and links

- `min_word_count` is a content metric over the body alone: `page.content.word_count >= 50`. The threshold is deliberately low; it catches a stub, not a style.
- `code_blocks_have_language` quantifies over the AST code blocks: `page.ast.code_blocks.all(c, c.language != "")`. A block with no language is unhighlightable and usually a paste that lost its fence.

```cel
page.ast.code_blocks.all(c, c.language != "")
```

<!-- PASSES: code_blocks_have_language rule — the one fenced block in this body is tagged cel -->

### Cross-field and temporal rules

- `updated_not_before_created` orders two fields against each other with `timestamp()`: `timestamp(page.frontmatter.updated) >= timestamp(page.frontmatter.created)`. Any frontmatter string that parses as RFC3339 or a date-only value is converted to a timestamp for exactly this kind of comparison.
- `temporal_created` compares a field against `now`, so a page cannot be backdated into the future.
- `superseded_requires_field` is a conditional requirement, and it guards the optional key it reads: a superseded record must name what it replaces, and the check is `!has(page.frontmatter.supersedes) || page.frontmatter.supersedes != ""` on the branch that applies.
- `proposed_has_no_supersedes` is the mirror image, a conditional prohibition: while a record is still proposed it must not claim to have replaced anything.

<!-- PASSES: updated_not_before_created rule — updated (2026-02-10) is after created (2026-01-05) -->
<!-- PASSES: superseded_requires_field rule — status is "accepted", so the conditional never reaches the optional key -->
<!-- PASSES: proposed_has_no_supersedes rule — status is "accepted", so the prohibition does not apply -->

### The old_page state machine

- `valid_state_transition` reads the pre-modification state to allow only the lifecycle moves the record type recognises, and every read of `old_page` is guarded, because a create has no previous state at all.
- `created_immutable` compares the two states to keep the created date fixed: `!has(old_page.frontmatter) || !has(old_page.frontmatter.created) || old_page.frontmatter.created == page.frontmatter.created`.

<!-- PASSES: valid_state_transition rule — this mockup is evaluated against itself as old_page, so the old and new status are both "accepted" -->
<!-- PASSES: created_immutable rule — old_page is this same page, so both created values are 2026-01-05 -->

### Guard every optional read with has()

This is the part worth copying even if nothing else is. CEL treats a missing key as an evaluation error, not as a false value, so the guard is the rule's only protection against an absent field.

- A **schema-required** key is read unguarded: `akb write` refuses the page before the rules run.
- A **schema-optional** key is always read behind `has()`, on the same branch that needs it: `!has(page.frontmatter.updated) || now - timestamp(page.frontmatter.updated) < duration("2160h")`.
- Every `old_page` read is guarded, because `old_page` is nil on a create and the on-disk page may predate the presence check or never have passed `akb write` at all.
- Every `lint_rule` guards everything except `type` and `title`: the sweep evaluates pages the write path never saw.

## Consequences

Authors copying this template may delete any rule they do not need, but the mockups must be updated in the same `akb template write`. The command re-checks the pass mockup as given, without each optional key it supplies, and against itself as `old_page`, so a rule and its exemplar move together or the write is refused. Deleting a rule is therefore a design act with a proof obligation attached, not a cleanup.

Rules that survive the copy keep their guards. A rule that reads an optional key directly, or that assumes `old_page` is present, will pass authoring only until the first page that omits the key, and then fail closed at write time or surface as a `cel_lint` issue at sweep time. The guard doctrine in this decision is recorded for reuse in [[templates/guard-doctrine]], so a copied template can cite it instead of restating it.
