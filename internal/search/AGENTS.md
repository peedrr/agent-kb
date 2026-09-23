# internal/search/

**Parent:** `./AGENTS.md`

## OVERVIEW

SQLite FTS5 full-text search with BM25 ranking. Replaces former NoOp stub.

## FILES

| File | Purpose |
|------|---------|
| `searcher.go` | Interfaces: `Searcher` (IndexPage/RemovePage/Search/RebuildIndex, line 33) and `TxSearcher` (IndexPageTx/RemovePageTx) |
| `sqlite.go` | `SQLiteFTS5Searcher` implementation (316 lines) |
| `noop.go` | No-op stub for testing |
| `sqlite_test.go` | Unit tests |

## SCHEMA

```
documents(id, title, content, tags, path, summary, created, updated)
pages_fts(title, content, tags, summary) -- FTS5 virtual table
pages(path, title, summary)
```

## QUERY ESCAPING

`escapeFTS5Query()` strips FTS5 operators (OR, AND, NOT) and special chars (`"`, `'`, `(`, `)`, `*`, `:`), then wraps each remaining whitespace-separated token in double quotes. Inside the quotes punctuation stays literal — the hyphen in `event-driven` matches as written instead of being read as query syntax — and the stripped `:` plus the per-token quoting leave no operator or column filter for a caller to inject. A query that is empty after stripping is rejected with `search query must not be empty`.

## REBUILDINDEX

- Walks `kb/` directory
- Parses frontmatter from each `.md` file
- Drops/recreates `pages_fts` table (handles schema migration)
- Runs in one transaction: a failure mid-walk rolls back to the prior index
- Rebuilds `kb/index.md` via `index.RebuildIndex()`
