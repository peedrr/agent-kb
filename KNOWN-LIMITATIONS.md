# Known Limitations

Deliberate deferrals: divergences and leftovers that were reported, understood, and
consciously left in place. Each entry states what happens, why it was deferred, and the date
it was reported. An entry leaves this file only when the behavior is changed and pinned;
this file is the standing filter for future triage, not a bug list.

The wikilink behavior these entries qualify is pinned in [`docs/wikilinks.md`](docs/wikilinks.md).

## Parser ↔ goldmark divergences

### Paren balance differs in both directions

**What happens.** The wikilink parser's destination scan and goldmark's inline-link parser
do not agree on unbalanced parentheses, in both directions:

- goldmark accepts a destination this parser rejects. `[[a]](( )` yields goldmark destination
  `(` (so CEL's merged entry reports target `(`) while the parser treats the token as a plain
  wikilink and the link graph records the bracket target `a`.
- this parser accepts a destination goldmark rejects when a quoted span hides the group's
  closer: `[[a]](>'.(\(/[()')` records the destination `>'.(\(/[()'` while goldmark parses no
  link at all.

**Why deferred:** closing either direction means making one reader's verdict authoritative on
the other's path — a design change to the parser/goldmark split, not a local fix.
**Reported, deliberately deferred: 2026-09-24.**

### `[[a]]([[b]])`: `page.ast.links` and the link graph disagree

**What happens.** A wikilink whose parenthesized destination is itself a bracket token is
malformed. `page.ast.links` emits exactly one entry — the inner token, target `b` — dropping
the outer token and its goldmark link. The link graph, which consumes the parser output
directly, still records the bracket-shaped target `[[b]]` for the outer token.

**Why deferred:** the graph-side fix means extending span containment from the CEL merge into
`ParseWikilinks`/`internal/linkgraph` — degenerate input with no real-world authoring pattern.
An owner ruling on whether to close it was still pending when this file was written; the item
is ledgered by default until that ruling lands.
**Reported, deliberately deferred: 2026-09-24.**

### Indented-code detection is the simple 4-column rule

**What happens.** Indented code is measured as four columns past the content column of the
innermost open list item (four columns at the top level). Strict CommonMark would need six
spaces in some list contexts, so a 4-space line after a blank line inside a list item can be
treated as code where strict CommonMark would call it paragraph text. Scope-literal,
degenerate.

**Why deferred:** the rule is a deliberate simplification — full container-stack modelling is
a larger change than the divergence justifies.
**Reported, deliberately deferred: 2026-09-24.**

### Nested and empty list items are modelled as one flat item

**What happens.** Marker-only items such as `+ + +` are modelled as a single flat item, while
goldmark nests empty items. A blank line followed by a 4-space line then records one link
where goldmark renders indented code (zero links).

**Why deferred:** same container-stack modelling limit as above; pre-existing and unchanged by
the list-bookkeeping work.
**Reported, deliberately deferred: 2026-09-24.**

### Goldmark-only path: a malformed destination only goldmark accepts

**What happens.** When the parser leaves a wikilink-form token plain but goldmark still parses
a real inline link, CEL takes goldmark's destination, normalized. For a destination only
goldmark accepts (`[[a]](( )` → goldmark destination `(`) that yields CEL target `(` while the
link graph records the bracket target `a` — the two readers diverge on a malformed input.

**Why deferred:** closing it requires making the parser's verdict authoritative on the
goldmark-only path (see the paren-balance entry above), a design change.
**Reported, deliberately deferred: 2026-09-24.**

### Scheme-qualified destinations are normalized by the link graph

**What happens.** `[[a]](https://x/y.md)` records `https://x/y` — the parser path normalizes
every destination, including a scheme-qualified one, so a URL's `.md` suffix is stripped.
CEL matches the graph, per the one-spelling convention; the `://` no-mangle guard governs only
the goldmark-reported path (`normalizeWikilinkTarget`).

**Why deferred:** pre-existing link-graph behavior; changing it would alter recorded targets
for URL-shaped destinations, a behavior call with no current consumer.
**Reported, deliberately deferred: 2026-09-24.**

### Lowercase HTML declarations interrupt paragraphs goldmark keeps open

**What happens.** `startsParagraphInterrupt` lowercases the line before the HTML type-4 and
type-5 start checks, so a lowercase declaration line (`<!doctype html>`, `<!a>`, or
`<![cdata[`) inside an open list-item paragraph is judged a paragraph interrupt: the list
item closes and a following blank-line-plus-4-indented wikilink is dropped as indented code.
goldmark v1.8.2 requires an uppercase ASCII letter for type 4 (`^[ ]{0,3}<![A-Z]+`) and the
exact `<![CDATA[` spelling for type 5, so it treats those lines as lazy paragraph
continuations and renders the link.

**Why deferred:** the parser is spec-frozen as of v0.19.0 — the divergence needs the exact
lowercase spelling inside a list item to fire, and spec-frozen divergences on degenerate
input are documented, not patched.
**Reported, deliberately deferred: 2026-09-24.**

## Exclusions not shared with the marker parsers

### Provenance and annotation parsers keep their own, narrower exclusion set

**What happens.** `internal/cel/pagebuilder.go` carries duplicated exclusion helpers for the
provenance-marker and annotation parsers. They deliberately did **not** receive the
escaped-bracket and indented-code exclusions the wikilink parser gained, so a provenance
marker or annotation inside an escaped bracket run or a 4-space indented block can still be
recorded where the wikilink parser would skip the region.

**Why deferred:** out of scope for the exclusion work; the fix is to share one exclusion
helper across all three parsers, which touches the marker parsers' pinned behavior.
**Reported, deliberately deferred: 2026-09-24.**

## Error-classification consistency

### `BuildOldPage` failures are classified inconsistently

**What happens.** Write's full-stdin overwrite branch returns a `cel.BuildOldPage` failure as a
plain page error (exit 1), while `runAppend` and write's `--append`/`--frontmatter` branches
wrap it as an internal error with a "CEL engine error:" prefix (exit 2).

**Why deferred:** practically unreachable — `frontmatter.Parse` already succeeded on the same
locked content before `BuildOldPage` re-reads it — but the classification is latent and would
mislead if it ever fires.
**Reported, deliberately deferred: 2026-09-24.**

### An unreadable page reports "page not found" on append

**What happens.** `akb append` to an existing page that cannot be read (for example mode 000)
reports `page not found: <path>. Use 'akb write' to create` instead of a read error. The exit
code is still 1.

**Why deferred:** deliberate convergence to write's documented "unreadable = passed over"
helper semantics; recorded in case the misclassification of read errors matters later.
**Reported, deliberately deferred: 2026-09-24.**

## Code health

### Template loading is duplicated between append and type discovery

**What happens.** The `AssertContained` + `assertTemplateFilesContained` + `LoadTemplates`
block is written inline in both `runAppend` and `typeDirsFromDisk`. Identical today.

**Why deferred:** cosmetic duplication with a future-divergence risk; the fix is one shared
helper.
**Reported, deliberately deferred: 2026-09-24.**

### Pre-existing `//nolint:errcheck` directives in write

**What happens.** `cmd/akb/write.go` carries `//nolint:errcheck` directives that predate the
current write path.

**Why deferred:** outside the blast radius of every change since; trivial lint hygiene.
**Reported, deliberately deferred: 2026-09-24.**

## Test hygiene

### Orphan status goes stale until the linking page is rewritten

**What happens.** Write-time link resolution means an ambiguous-basename inbound link keeps
suppressing orphan status for its stale target until the linking page is rewritten or
`akb index rebuild` runs. Documented in the `GetOrphans` doc comment, but no automated test
pins the staleness.

**Why deferred:** documented behavior with a test gap, not a defect.
**Reported, deliberately deferred: 2026-09-24.**

### Embedded mockup revalidation is warning-only

**What happens.** `akb template get <name> --example` warns on stderr when the pass mockup no
longer validates; nothing forces a fail mockup to keep failing after rule edits. The embedded
`adr` fail mockup now has a pin asserting it fails exactly `valid_status`, but `note` has no
equivalent and the `template get --example` warning path has no automated test.

**Why deferred:** the adr case is pinned; the remainder is a test gap on a warning-only path.
**Reported, deliberately deferred: 2026-09-24.**

### A test swaps the process-global `os.Stdin`

**What happens.** A `cmd/akb` test replaces `os.Stdin` and restores it only in cleanup. Safe
while same-package tests run sequentially; a future `t.Parallel()` in that package would race
it.

**Why deferred:** no current failure mode; the fix is a helper or an explicit constraint.
**Reported, deliberately deferred: 2026-09-24.**

### Cached integration-suite results

**What happens.** A `go test ./test/` result can be served from go test's cache, so a change
is not observed to execute. Verification hygiene: force `-count=1` when the run itself is the
evidence.

**Why deferred:** standing practice note rather than a defect.
**Reported, deliberately deferred: 2026-09-24.**
