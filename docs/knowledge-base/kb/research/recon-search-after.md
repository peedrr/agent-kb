---
as_of: "2026-09-27"
created: "2026-09-27"
grammar: 1
informs:
- RFC-001
- RFC-002
is_draft: false
provenance: agent-drafted
run: json-cel
scope:
- cmd/akb/**
- internal/search/**
- internal/db/**
status: final
summary: "Probe-verified recon: --after filters on documents.created, which is always index time and reset by rebuild/append/approve — not the page's creation date the help text promises."
tags:
- search
- recon
- probe
- database
title: "Recon: how akb search --after <date> behaves"
type: research
updated: "2026-09-27"
---
# Recon: how `akb search --after <date>` behaves (read-only)

Repo state: HEAD `073f275` (v0.21.0). Probe binary built from HEAD to `/tmp/akb-probe-bin`; probe KB at `/tmp/afterprobe/probe`; DB inspected with a throwaway `database/sql`+`modernc.org/sqlite` reader at `/tmp/dbprobe/main.go`. No repo files modified.

## Files Retrieved

1. `cmd/akb/search.go` (lines 22-70) — flag declaration + wiring into `search.SearchOptions`.
2. `internal/search/searcher.go` (lines 25-31, 54-61) — `SearchOptions{Tag,Type,After,Limit}` and `IndexPageTx` signature (no date parameter).
3. `internal/search/sqlite.go` (lines 52-84) — `IndexPageTx`: the only writer of `documents.created` / `documents.updated`.
4. `internal/search/sqlite.go` (lines 153-205) — `Search`: WHERE-clause assembly, `ORDER BY rank LIMIT ?`.
5. `internal/search/sqlite.go` (lines 213-301) — `RebuildIndex`: `DELETE FROM pages` / `DELETE FROM documents` then re-insert every file via `IndexPageTx`.
6. `internal/search/sqlite.go` (lines 306-339) — `ExtractTags` / `ExtractSummary` (the field-lifting precedent).
7. `internal/db/db.go` (lines 78-107) — DDL: `documents.created TEXT NOT NULL DEFAULT ''` (no default `datetime('now')`, no index).
8. `cmd/akb/write.go` (lines 376-380, 551-556) — frontmatter `created` injection + the `IndexPageTx` call that drops it.
9. `cmd/akb/index.go` (lines 312-316) — `akb index rebuild` → `searcher.RebuildIndex`.
10. `internal/skill/embedded/kb-management/references/QUERY.md` (line 41) and `internal/skill/embedded/kb-management/SKILL.md` (line 120) — the only user-facing docs for the flag.
11. `cmd/akb/AGENTS.md` (line 51), `AGENTS.md` (line 67) — repo-level claims.
12. `test/testdata/search_tests.txt` (whole file, 146 lines) — the only testscript coverage of `akb search`; no filter flags.

## Mechanism

Trace, end to end:

- `searchCmd.Flags().StringVar(&searchAfter, "after", "", "filter by creation date (YYYY-MM-DD)")` — `cmd/akb/search.go:41`; no validation, free-form string.
- `runSearch` copies it straight into `search.SearchOptions{Tag, Type, After: searchAfter}` — `cmd/akb/search.go:64-68`. `SearchOptions.Limit` is left at zero (`searcher.go:25-31`).
- `SQLiteFTS5Searcher.Search` appends the predicate `AND d.created >= ?` with the raw string as the bind arg — `internal/search/sqlite.go:175-178`. The query is the FTS5 MATCH over `pages_fts` joined to `documents d` (`sqlite.go:186-196`), so **the filter is a SQLite `documents` table column, not a filesystem/mtime scan**.
- The column is written in exactly one place: `IndexPageTx` — `INSERT OR REPLACE INTO documents (path, title, content, tags, summary, created, updated, type) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'), ?)` — `internal/search/sqlite.go:59-61`. That is `datetime('now')` (UTC, `'YYYY-MM-DD HH:MM:SS'`).
- Callers of `IndexPageTx`: `cmd/akb/write.go:553`, `cmd/akb/append.go:244`, `cmd/akb/approve.go:195`, and `RebuildIndex` at `internal/search/sqlite.go:284`. **None of them passes any date** — the signature is `(ctx, tx, path, title, content, tags, summary, pageType)` (`searcher.go:60`).

**Confirmed: the column is index time, never the page's frontmatter `created`.** `akb write` does inject `created: <time.Now().UTC().Format(time.RFC3339)>` into frontmatter when absent (`cmd/akb/write.go:377-379`), but that value never leaves the file — the index call at `write.go:553` passes only title/body/tags/summary/type. Live proof:

```
page frontmatter: created: 2020-01-01   →  documents row: created="2026-09-27 19:13:19"
search "zzprobe" --after 2021-01-01     →  HIT   (page declared 2020)
search "zzprobe" --after 2030-01-01     →  (empty)
```

**Rebuild resets every row's `created` to rebuild time — confirmed.** `RebuildIndex` runs `DELETE FROM pages` (`sqlite.go:227`) then `DELETE FROM documents` (`sqlite.go:230`) and re-inserts every file through `IndexPageTx` (`sqlite.go:284`), so `created` is rewritten to `datetime('now')` for *all* rows. Probe:

```
UPDATE documents SET created='2019-05-05 08:00:00'   → search --after 2020-01-01 : empty
akb index rebuild                                    → created="2026-09-27 19:13:31", id 1→2
                                                       search --after 2020-01-01 : HIT
```

Two extra facts the same probe surfaced:
- `INSERT OR REPLACE` + `path TEXT UNIQUE` (`db.go:89`) means a plain re-index (not just rebuild) also resets the value: `akb append notes/old.md` on a row with `created='2019-05-05 08:00:00'` produced `id 2→3, created="2026-09-27 19:13:49"`. So `--after` is semantically "on or after the last time this page was (re)indexed", not "created".
- `documents.updated` (`sqlite.go:60`) is also `datetime('now')` at the same instant and is **read by nothing** — grep for `d.updated` / `.updated` finds no query consumer anywhere in `cmd/` or `internal/`.

## Date comparison

**It is a raw SQL TEXT comparison, not a SQLite datetime comparison.** `d.created >= ?` with the user's string as the parameter (`sqlite.go:177`) — no `datetime()`, `date()`, `julianday()`, or `strftime()` wraps either side, and there is no normalization on the Go side. The column has TEXT affinity with format `'YYYY-MM-DD HH:MM:SS'` (UTC), and the parameter is a Go `string`, so SQLite compares byte strings lexicographically.

Verified matrix against the probe row `created = "2026-09-27 19:13:31"`:

| `--after` value | result |
|---|---|
| `2026-09-27` | HIT |
| `2026-09-26` | HIT |
| `2026-09-27 19:13` | HIT |
| `2026-09-27 19:13:31` | HIT (equal) |
| `2026-09-27 19:13:32` | empty |
| `2026-09-27T00:00:00Z` (RFC3339) | **empty** |
| `garbage`, `9999-99-99` | empty, exit 0, no error |

Does `created >= ?` behave sanely?
- **Date-only input: yes, and slightly better than naive.** Because the stored value always carries a trailing ` HH:MM:SS`, the bare date sorts *below* every value from that day, so `--after 2026-09-27` includes the whole named day — the intuitive "created on or after this date" reading. Note the boundary is inclusive only by accident of the format: a hypothetically date-only stored value `2026-09-27` would be *excluded* by `>= '2026-09-27'`… no, it would be equal and included — the real hazard is the opposite direction: any stored value strictly *shorter* than the input, which cannot occur with `datetime('now')`.
- **Exact `'YYYY-MM-DD HH:MM:SS'` input: yes**, exact inclusive boundary.
- **RFC3339 input (`2024-01-15T06:00:00Z`): broken.** At byte 10, `' '` (0x20) < `'T'` (0x54), so every same-day row compares low and is silently dropped — the probe returned empty for `2026-09-27T00:00:00Z` against a row created that very day. This matters because RFC3339 is the documented frontmatter date convention and the format `akb write` itself emits (`write.go:378`), so the "obvious" copy-paste value from a page's frontmatter is exactly the one that misbehaves.
- **Malformed input: silently empty.** No validation, no error, exit 0 — `garbage` just sorts above every row.
- **Timezone: UTC-only.** `datetime('now')` is UTC; a user in, say, UTC+13 asking `--after 2026-09-27` at 00:30 local is asking about a day that has not started in the stored clock.

## Usage evidence

Grep of the whole repo (`--include=*.go/*.md/*.txt/*.yaml/*.sh`):

- **Implementation**: `cmd/akb/search.go:41,67`; `internal/search/searcher.go:30`; `internal/search/sqlite.go:176-178`. Added in a single commit, `7f66982 feat(delete,approve,search,links): orphans, batch approve, dimensional search, empty backlinks` (`git log -S'searchAfter' -- cmd/akb/search.go`). One commit, no follow-up.
- **Documentation (4 places, all "surface" claims, none describing index-time semantics)**:
  - `internal/skill/embedded/kb-management/references/QUERY.md:41` — `akb search "API" --after 2025-01-01   # by creation date`
  - `internal/skill/embedded/kb-management/SKILL.md:120` — flag listed in the command table
  - `cmd/akb/AGENTS.md:51` — "`search` supports `--tag`, `--type`, `--after` dimensional filters"
  - `AGENTS.md:67` — "Dimensional search | `cmd/akb/search.go` | `--tag`, `--type`, `--after` filters"
- **Tests: zero.** `grep -rn -- "--after" test/` → nothing; `grep -rn "After"` over `internal/search/*_test.go`, `internal/linkgraph/*_test.go`, `cmd/akb/search_test.go` → nothing; no `SearchOptions{...After:...}` anywhere. `test/testdata/search_tests.txt` (146 lines) covers plain queries, `--json`, FTS5 escaping, and the missing-DB hint, but **no filter flag**. `--tag` and `--type` are equally untested (`test/` uses `--type` only for `akb log show`).
- **No internal consumers**: the only `SearchOptions` construction in non-test code is `cmd/akb/search.go:64`; every other occurrence is a test literal passing `Limit` only.
- **`--limit` does not exist.** `akb search --help` lists `--after`, `--json`, `--tag`, `--type` only; `limit := opts.Limit; if limit == 0 { limit = 10 }` at `internal/search/sqlite.go:159-162`, and no CLI path ever sets `Limit`.
- **README/CHANGELOG/docs/**: no mention of `akb search` flags at all.

**Verdict on usage: orphaned-looking surface.** Shipped with its docs, never exercised by any test, never called by any internal code path, and the CLI's own help text (`search.go:41`) asserts a semantics the implementation does not have.

## Lifting a declared field

**No temporal column in `documents` is page-derived.** The DDL (`internal/db/db.go:78-107`) has exactly two temporal columns, `created` and `updated`, both `TEXT NOT NULL DEFAULT ''`, both populated solely by `datetime('now')` at `sqlite.go:60`. The frontmatter values that *do* get lifted are `tags`, `summary` (and `title`/`type`): extracted at `internal/search/sqlite.go:306-323` / `325-339` and passed positionally into `IndexPageTx` from the walk in `RebuildIndex` (`sqlite.go:281-284`). So the precedent exists and is exactly the shape a `created`-lifting change would follow:

1. In `RebuildIndex`, read the declared field from the already-parsed `fm.Fields` map next to `tags := ExtractTags(fm.Fields)` / `summary := ExtractSummary(fm.Fields)` (`sqlite.go:281-282`).
2. Extend `IndexPageTx`'s parameter list (`searcher.go:60`, `sqlite.go:56`) — this touches all four call sites: `internal/search/sqlite.go:284`, `cmd/akb/write.go:553`, `cmd/akb/append.go:244`, `cmd/akb/approve.go:195`, plus the `TxSearcher` interface and `noop.go`.
3. Bind the value in the INSERT (`sqlite.go:60`) instead of `datetime('now')`.
4. **Normalize the stored form**, because the comparison in `Search` is lexicographic (`sqlite.go:177`). The CEL layer already has the parser to copy: `internal/cel/pagebuilder.go:27-42` accepts RFC3339 via `time.Parse(time.RFC3339, str)` and date-only via `time.Parse("2006-01-02", str)`. Storing RFC3339 would break every date-only `--after` (a date-only string sorts below a same-day RFC3339 value? not uniform — `2024-01-15` vs `2024-01-15T00:00:00Z` is a prefix, so still a HIT; the real breakage is the reverse and mixed offsets). Storing the current `datetime('now')` shape (`2006-01-02 15:04:05`, UTC) keeps today's `--after` semantics intact.
5. Consider a nullable column + `COALESCE` fallback so pages with no declared field keep index-time behavior, and note there is **no index on `created`** (only `idx_documents_type`, `db.go:106`) — the predicate is evaluated post-FTS-join on a handful of rows, so an index is not required for the current query shape.

Cost is mechanical, not architectural: 2 packages, ~5 files, one extra string on a signature that already carries 7.

## Verdict

**"Broken and unused-looking" is accurate on both halves, for a specific sense of "broken".** Unused: the flag has exactly one definition site and one construction site (`cmd/akb/search.go:41,67`), zero test coverage in `test/testdata/*.txt` and zero unit-test coverage of `SearchOptions.After`, no internal caller, and one commit (`7f66982`) of history — it is documented in four places and exercised in none, making it pure orphaned surface. Broken: the help text and both skill docs promise "filter by creation date", but the predicate reads `documents.created`, which is `datetime('now')` at index time (`internal/search/sqlite.go:60,177`) — so the frontmatter `created` that `akb write` itself stamps (`cmd/akb/write.go:377-379`) is never consulted, `akb index rebuild` rewrites every row's value to rebuild time (`sqlite.go:227-231` + `284`), and even a plain append/approve re-index resets it. The user-visible consequence is a filter that answers "pages whose index row was last written on or after X", which after any rebuild means *every* page, regardless of the dates the KB declares. What is merely **surprising rather than broken**: the comparison itself is a raw lexicographic TEXT compare, but for the `YYYY-MM-DD` input its own help advertises it happens to behave like a sane inclusive day filter (the stored value's ` HH:MM:SS` suffix sorts above the bare date); the sharp edges are that a *shared* predicate is really "last indexed", that RFC3339 input (the format akb writes into frontmatter) silently returns nothing for same-day rows, that malformed input is silently empty with exit 0, and that the day boundary is UTC. Secondary finding on the same lines: `documents.updated` is likewise index-time and read by nobody, and `--limit` genuinely does not exist (`sqlite.go:159-162`).

## Start Here

Open `internal/search/sqlite.go:56-84` (`IndexPageTx`) first — it is the single fact that explains the whole finding: the only writer of `documents.created` binds `datetime('now')` and receives no date from any caller. Then `internal/search/sqlite.go:175-178` for the predicate and `internal/db/db.go:78-107` for the column's TEXT affinity, which together establish the lexicographic comparison.
