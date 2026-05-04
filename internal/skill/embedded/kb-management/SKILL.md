---
name: kb-management
description: >
  Knowledge Base management skill for agents using the `akb` CLI tool.
  Use whenever the user wants to create, maintain, query, or organize a
  markdown-based knowledge base. Triggers on: wiki, knowledge base, KB,
  akb, ingest, compile sources, query what we know, lint wiki, update pages,
  delete pages, add raw files, approve drafts, create template, define page type,
  add schema, design template, or any request to read/write structured project knowledge.
  Do NOT use for generic chat, unrelated file operations, or off-wiki trivia.
metadata:
  author: peedrr
  version: "1.2"
  tool: akb
---

# Agent KB — Task Router

You are maintaining a **persistent, compounding knowledge base** using the `akb` CLI. The KB is not a chat transcript — it is a **compiled artifact** where knowledge is distilled once, kept current, and enriched over time.

**You are the brain.** You decide what to create, connect, update, and delete.
**`akb` is the hands.** It reads, writes, searches, links, and lints — but never calls an LLM.

## CRITICAL DIRECTIVE

**This SKILL.md is a router only. It does NOT contain complete instructions for any task.**

Before performing ANY KB-related work, you MUST:

1. Classify the user's request using the decision table below.
2. Read the corresponding reference file from `references/`.
3. Follow that reference's instructions EXACTLY.

**DO NOT attempt to perform KB work by improvising from this file. Read the correct reference first.**

## Task Classification

| Trigger | Intent | Read This Reference |
|---------|--------|---------------------|
| No KB exists; "create KB", "init wiki" | Initialize | `references/INIT.md` |
| "Add this", "ingest", "compile source", new documents | Add knowledge | `references/INGEST.md` |
| "What do we know", "search", "find", "compare" | Query | `references/QUERY.md` |
| "Update", "fix", "correct", "merge", "revise" | Revise pages | `references/UPDATE.md` |
| "Lint", "health check", "orphans", "broken links" | Audit / repair | `references/MAINTAIN.md` |
| "Approve", "publish", "strip annotations" | Publish drafts | `references/APPROVE.md` |
| "Create template", "define page type", "add schema", "design template", "template type" | Template authoring | `references/TEMPLATE.md` |

If the request matches multiple triggers, prefer the one higher in the table.

## Shared Conventions

These rules apply to ALL KB operations. The reference files assume you know these.

1. **Never write to `kb/index.md` or `kb/log.md` directly.** Use `akb index` and `akb log` commands only.
2. **Never write to `raw/files.log` directly.** Use `akb raw sync` only.
3. **Raw sources are immutable.** Once in `raw/`, never modify them.
4. **All mutations auto-commit.** `akb write`, `akb append`, `akb delete`, `akb approve` all create git commits. Use `--no-commit` for batching.
5. **Log and index only on mutations.** Read-only queries do not write to `kb/log.md` or `kb/index.md`.
6. **Path guard rails:** `akb write kb/test.md` strips the redundant `kb/` prefix. `akb write raw/test.md` errors — use `akb raw write`. No `..` or absolute paths.
7. **Type-driven directories:** Every page declares `type` in frontmatter. The template's `dir` field determines placement under `kb/`.
8. **Wikilinks:** Prefer `[[path-form]]` (e.g., `[[concepts/attention]]`). Short names (`[[attention]]`) are supported but may be ambiguous.
9. **Drafts:** New pages are drafts by default (`is_draft: true` is assumed). Use `akb approve` to publish.
10. **Multi-KB:** `akb registry` lists KBs; `akb use <name>` switches the active one.

---

## Command Quick Reference

| Operation | Command |
|-----------|---------|
| Initialize KB | `akb init <name>` |
| Read page | `akb read <path>` |
| Write page (stdin) | `akb write <path> [--frontmatter key=val] [--append] <<'EOF' ... EOF` |
| Append to page | `akb append <path> <<'EOF' ... EOF` |
| Delete page | `akb delete <path> [--orphans] [--force]` |
| List pages | `akb list [--json]` (shows [DRAFT] indicators) |
| Search | `akb search <query> [--tag <tag>] [--type <type>] [--after <date>] [--json]` |
| Show links | `akb links <path>` |
| Show backlinks | `akb backlinks <path> [--json]` |
| Show orphans | `akb orphans` |
| Lint | `akb lint [--json]` |
| Stale report | `akb stale [--all] [--json]` |
| Status | `akb status` |
| Approve draft | `akb approve <path> [--all-drafts]` |
| Index management | `akb index show|add|remove|rebuild` |
| Log management | `akb log show|append` |
| Raw write | `akb raw write <path>` |
| Raw sync | `akb raw sync` |
| Raw status | `akb raw status` |
| Raw list | `akb raw list` |
| Raw read | `akb raw read <path>` |
| Raw delete | `akb raw delete <path>` |
| Registry | `akb registry` |
| Switch KB | `akb use <name>` |
| Install skill | `akb skill install --location <path> <name>` |
