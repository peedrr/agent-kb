---
name: kb-management
description: >
  Knowledge Base management skill for agents using the `akb` CLI tool.
  Use whenever the user wants to create, maintain, query, or organize a
  markdown-based knowledge base. Triggers on: wiki, knowledge base, KB,
  akb, ingest, compile sources, query what we know, lint wiki, update pages,
  delete pages, add raw files, approve drafts, create template, define page type,
  add schema, design template, update template, change template, modify template,
  add validation rule, delete template, remove template, template list, what types exist,
  or any request to read/write structured project knowledge.
  Do NOT use for generic chat, unrelated file operations, or off-wiki trivia.
metadata:
  author: peedrr
  version: "2.0"
  tool: akb
---

# Agent KB — Task Router

You are maintaining a **persistent, compounding knowledge base** using the `akb` CLI. The KB is not a chat transcript — it is a **living reference** where knowledge is distilled once, kept current, and enriched over time.

**You are the brain.** You decide what to create, connect, update, and delete.
**`akb` is the hands.** It reads, writes, searches, links, and lints — but never calls an LLM.

## CRITICAL DIRECTIVE

**This SKILL.md is a router only. It does NOT contain complete instructions for any task.**

Before performing ANY KB-related work, you MUST:

1. Classify the user's request using the decision table below.
2. Read the corresponding reference file from `references/`.
3. Follow that reference's instructions EXACTLY.

**DO NOT attempt KB work by improvising from this file. Read the correct reference first.**

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
| "Update template", "change template", "modify template", "add validation rule" | Template update | `references/TEMPLATE.md` |
| "Delete template", "remove template" | Template deletion | `references/TEMPLATE.md` |
| "Template list", "what types exist", "available types" | Template query | `references/TEMPLATE.md` |

If the request matches multiple triggers, prefer the one higher in the table.

## Constraints

These rules apply to ALL KB operations. The reference files assume you know these.

1. **Never edit the index or log directly.** Use `akb index` and `akb log` commands only.
2. **Never edit the raw manifest directly.** Use `akb raw sync` only.
3. **Raw sources are immutable.** Once stored via `akb raw write`, never modify them.
4. **Batch writes with `--no-commit`.** All write commands auto-commit by default. Use `--no-commit` on all but the last in a batch to avoid one commit per operation.
5. **Wikilinks:** Prefer `[[path-form]]` (e.g., `[[concepts/attention]]`). Short names (`[[attention]]`) are supported but may be ambiguous.
6. **Drafts:** New pages are drafts by default. Use `akb approve` to publish.
7. **Always check the template before writing a page.** Run `akb template get <type>` before `akb write`. Writing without understanding requirements causes validation failures.
8. **Always run `akb lint` after template changes.** Changing or deleting a template affects every page of that type.

---

## Command Quick Reference

| Operation | Command |
|-----------|---------|
| Initialize KB | `akb init <name>` |
| Read page | `akb read <path>` |
| Write page | `akb write <path> [--frontmatter key=val] [--append] <<'EOF' ... EOF` |
| Delete page | `akb delete <path> [--orphans] [--force]` |
| List pages | `akb list [--json]` |
| Search | `akb search <query> [--tag <tag>] [--type <type>] [--after <date>] [--json]` |
| Show links | `akb links <path>` |
| Show backlinks | `akb backlinks <path> [--json]` |
| Show orphans | `akb orphans` |
| Lint | `akb lint [--json]` |
| Status | `akb status` |
| Approve draft | `akb approve <path> [--all-drafts]` |
| Index | `akb index show\|add\|remove\|rebuild` |
| Log | `akb log show\|append` |
| Template list | `akb template list` |
| Template get (schema + requirements) | `akb template get <name>` |
| Template get (example page) | `akb template get <name> --example` |
| Template get (full YAML) | `akb template get <name> --full` |
| Template write (new) | `akb templates write <name> --template <yaml> --pass <md> --fail <md>` |
| Template write (update) | `akb templates write <name> --template <yaml> [--pass <md>] [--fail <md>] --force` |
| Template delete | `akb template delete <name> [--force]` |
| Raw write | `akb raw write <path>` |
| Raw sync | `akb raw sync` |
| Raw status | `akb raw status` |
| Raw list | `akb raw list` |
| Raw read | `akb raw read <path>` |
| Raw delete | `akb raw delete <path>` |
| Registry | `akb registry` |
| Switch KB | `akb use <name>` |
