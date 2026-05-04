---
type: adr
title: Use SQLite FTS5 for Full-Text Search
summary: Adopt FTS5 to enable fast, ranked full-text search across all knowledge base pages.
tags: [database, search, performance]
status: proposed
deciders: alice
created: 2024-01-01
updated: 2024-01-01
---

# Use SQLite FTS5 for Full-Text Search

## Context

The current knowledge base contains 500+ pages with growing content. Users need to find relevant pages by keyword, not just by title or exact wikilink. Full-text search would improve discoverability and reduce time spent hunting for existing documentation.

Current search is limited to title matching via SQLite LIKE queries. This misses content in page bodies and provides no relevance ranking.

<!-- PASSES: valid_status rule — status is "proposed", which is in the allowed enum ["proposed", "accepted", "deprecated", "superseded"] -->
<!-- PASSES: temporal_created rule — created date (2024-01-01) is in the past, so timestamp(created) <= now evaluates to true -->
<!-- PASSES: require_title rule — frontmatter has non-empty title field -->
<!-- PASSES: require_context rule — AST contains H2 heading with text exactly "Context" -->

## Decision

We will implement full-text search using SQLite FTS5 (Virtual Table module). FTS5 provides:

- Tokenized full-text indexing of page content
- BM25 ranking for relevance sorting
- Phrase matching and boolean operators
- Low overhead (WAL mode, incremental updates)

The implementation will replace the existing search index with an FTS5 table storing page path, title, and tokenized body content.

<!-- PASSES: require_decision rule — AST contains H2 heading with text exactly "Decision" -->
<!-- PASSES: require_consequences rule — AST contains H2 heading with text exactly "Consequences" -->
<!-- PASSES: disallow_options rule — no H2 "Options" heading exists, so negation evaluates to true -->
<!-- PASSES: disallow_pros_cons rule — no H2 "Pros and Cons" heading exists, so negation evaluates to true -->
<!-- PASSES: valid_state_transition rule — new page has no old_page.frontmatter, short-circuits to true -->

## Consequences

**Positive:**
- Users can search page content, not just titles
- Results ranked by relevance (BM25)
- Fast queries even with thousands of pages

**Negative:**
- Index rebuild required on schema changes
- Slightly more storage (typically 2-3x body size)
- FTS5 query syntax differs from simple LIKE

**Risks:**
- Migration cost for existing KBs (rebuild index)
- Query escaping complexity (FTS5 operators)
- No built-in stemming (English only)