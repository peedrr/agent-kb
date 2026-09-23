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

`escapeFTS5Query()` strips FTS5 operators (OR, AND, NOT) and special chars (`"'()*`).

## REBUILDINDEX

- Walks `kb/` directory
- Parses frontmatter from each `.md` file
- Drops/recreates `pages_fts` table (handles schema migration)
- Runs in one transaction: a failure mid-walk rolls back to the prior index
- Rebuilds `kb/index.md` via `index.RebuildIndex()`
