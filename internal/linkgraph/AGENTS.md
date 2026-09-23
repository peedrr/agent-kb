# internal/linkgraph/

**Parent:** `./AGENTS.md`

## OVERVIEW

SQLite-backed wikilink tracking. Tracks `[[wikilink]]` references between pages.

## FILES

| File | Purpose |
|------|---------|
| `updater.go` | Interfaces: `Updater` (`UpdatePageLinks`/`RemovePage`) and `TxUpdater` (`UpdatePageLinksTx`/`RemovePageTx`) |
| `sqlite.go` | `SQLiteLinkGraph` implementation (379 lines) |
| `noop.go` | No-op stub |
| `sqlite_test.go` | Unit tests |

## LINK RESOLUTION (3-STEP)

1. **Exact match**: `raw_target + ".md"` in pages table
2. **Namespace prefix**: `"kb/" + raw_target + ".md"`
3. **Basename match**: filename without `.md` equals `raw_target`
   - 0 matches → broken link (nil)
   - 1 match → resolved
   - 2+ matches → `"AMBIGUOUS:path1,path2"` (sorted)

## SCHEMA

```
pages(path, title, summary)
links(source_page, raw_target, display, resolved_to)
idx_links_source ON links(source_page)
idx_links_resolved ON links(resolved_to)
```

## KEY FUNCTIONS

| Function | Purpose |
|----------|---------|
| `UpdatePageLinks()` | Parse wikilinks (deduped by raw target), resolve targets, store links; `UpdatePageLinksTx()` writes on a caller transaction |
| `RemovePage()` | Delete page + all links (outbound + inbound); `RemovePageTx()` writes on a caller transaction |
| `RebuildLinks()` | Walk kb/, update all pages in one transaction |
| `GetOutboundLinks()` | Links FROM a page |
| `GetInboundLinks()` | Links TO a page |
| `GetOrphans()` | Pages with no inbound links |
| `GetBrokenLinks()` | Links with NULL resolved_to |
| `GetAmbiguousLinks()` | Links with multiple targets |
