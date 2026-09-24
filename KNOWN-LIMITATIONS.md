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

### A normalized path destination that lands on a bracket token repeats that disagreement

**What happens.** `[[a]](./[[b]].md)` normalizes its destination to `[[b]]`, a bracket token,
so the parser suppresses no bracket run below it and emits both the outer token and the inner
one. The shape reproduces the disagreement above by a second route: the link graph records
two rows — raw targets `[[b]]` and `b` — while `page.ast.links` keeps one entry, the inner
token `b`. The angle-bracketed spelling `[[a]](<./[[b]].md>)` measures identically. Both
readers classify the destination as a bracket token, so the outcome is pinned by behavior,
not a second defect.

**Why deferred:** same graph-side fix as the entry above — span containment extended from the
CEL merge into `ParseWikilinks`/`internal/linkgraph`.
**Reported, deliberately deferred: 2026-09-24.**

### Indented-code detection is the simple 4-column rule

**What happens.** Indented code is measured as four columns past the content column of the
innermost open list item (four columns at the top level), and the parser's item-closing
bookkeeping — not a container stack — decides which item a line belongs to. The bookkeeping
tracks the open items' content columns: a non-blank line below an open item's content column
closes that item and every deeper one, except a lazy continuation of an open paragraph, which
closes nothing whatever its indent. A marker line opens an item and opens a paragraph only
when it carries item content; a paragraph interrupt — a fence delimiter, a thematic break, an
ATX heading, a block quote, or an HTML block of CommonMark type 1-6 — closes the items it
falls outside of. The lazy-continuation test keys on the previous line being non-blank rather
than on a paragraph actually being open, so a line indented four or more columns directly
after a paragraph-interrupting block is prose here where goldmark renders an indented code
block (`- item`, `# H`, `    [[a]]` records `a`).

**Why deferred:** the 4-column rule and the paragraph-state approximation are deliberate
simplifications; full container-stack modelling with per-item paragraph state is a larger
change than the divergence justifies, and the parser is spec-frozen.
**Reported, deliberately deferred: 2026-09-24.**

### Nested and empty list items are modelled as one flat item

**What happens.** Marker-only items such as `+ + +` are modelled as a single flat item, while
goldmark nests empty items. A blank line followed by a 4-space line then records one link
where goldmark renders indented code (zero links). The trailing markers count as item
content on the flat item, so its paragraph state opens and a low-indent line before the blank
line is a lazy continuation that closes nothing: `+ + +`, `x`, blank, `    [[a]]` records the
link too.

**Why deferred:** same container-stack modelling limit as above; the low-indent shape is a
consequence of the paragraph-open tracking, not a separate mechanism.
**Reported, deliberately deferred: 2026-09-24.**

### An indented line after a fence's closing delimiter stays a lazy continuation

**What happens.** The list bookkeeping's lazy-continuation test keys on the previous line
being non-blank, and a closing fence delimiter is a non-blank line, so a line indented four
or more columns directly after the closing delimiter of a fenced block is read as prose: the
parser records the link where goldmark renders an indented code block (zero links). A fenced
block is not a paragraph, so there is nothing to continue.

**Why deferred:** the spec-freeze classifies the residual lazy-continuation approximation as
documented, not patched.
**Reported, deliberately deferred: 2026-09-24.**

### `fencedCodeBlockRanges` accepts a fence anywhere and ignores an unclosed one

**What happens.** The fence detector opens a block on a fence delimiter anywhere in a line —
not only at the line start — and emits no range for a fence that never closes. Both readers
therefore drop a token between a mid-line delimiter and the next one (a mid-line opener with a
later closing delimiter, or two delimiters on one line) while goldmark renders the text as a
paragraph, and both record a token inside an unclosed fence while goldmark's fence runs to
the end of the file. The list bookkeeping narrows a fence range to the lines strictly between
its delimiters, so a same-line pair contributes no indented-code judgment; the range function
itself is unchanged.

**Why deferred:** line-start-only fencing and an end-of-file range change token exclusion for
every consumer and need a goldmark-parity pass, not a local fix.
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

### An indented line after a paragraph-interrupting block is a lazy continuation

**What happens.** The lazy-continuation check keys on the previous line being non-blank,
not on whether a paragraph is still open, so an indented line immediately after a
paragraph-interrupting block — an ATX heading, a block quote, a fence delimiter, a thematic
break, or an HTML block of CommonMark type 1-6 — with no blank line between them is judged
a lazy paragraph continuation and its wikilink is recorded. A `- item` or `para` line,
followed immediately by `# H` and then by an indented `    [[a]]` line, records link `a`;
goldmark renders that indented line as an indented code block (zero links).

**Why deferred:** pre-existing — the parent of the paragraph-interrupt fix behaved
identically. The parser is spec-frozen as of v0.19.0, so divergences on such input are
documented, not patched.
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

## Path guards

### `assertContained` allows the walk when the base root cannot be resolved

**What happens.** `filepath.EvalSymlinks(kbRoot)` failing (a looping or unreadable ancestor)
makes `assertContained` return no error: with no resolved root there is no boundary to compare
a link against, so the component walk is skipped. CLI calls are unreachable this way — the
base's config is stat'ed through the same path first — but a library caller handling an exotic
`kbRoot` gets no escape check.

**Why deferred:** fail-open on a path the CLI cannot reach; closing it means deciding what an
unresolvable base means, which changes the library calling contract.
**Reported, deliberately deferred: 2026-09-24.**

### `..` is rejected inside a filename, not only as a path component

**What happens.** `ResolveKBPath`/`ResolveRawPath` reject any input containing `..`, so a
legitimate mid-filename occurrence such as `notes/v2..bak.md` is refused as a
parent-directory escape (`path must not contain '..'`).

**Why deferred:** the lexical guard is deliberately blunt — allowing `..` only as a whole
component means a component-aware parse of every input, and no page name needs the sequence.
**Reported, deliberately deferred: 2026-09-24.**

### A symlink planted at `kb/` changes what the bulk sweeps see

**What happens.** The bulk page sweeps open the pages directory without the path resolver's
containment check, and `kb/` is itself a component that check would inspect. With `kb/` a
symlink to a directory outside the base, the sweeps that walk it (`akb list`, `akb status`,
the page walk behind `akb lint`, the search and link-graph rebuilds, the raw-delete link
scan) open the symlinked root as a non-directory entry and descend no further, so the base
reports an empty pages tree; direct reads that join `kb/<name>` follow the link instead —
`akb lint`'s index read resolves `kb/index.md` through it and reports on the outside file.
Commands that address a page by name are unaffected: the resolver rejects the escaping `kb/`
component as a usage fault. The constraint stays: the triage parked this case, and this
entry is its durable record.

**Why deferred:** making every bulk read containment-check the pages directory and each
entry below it is deferred to a future read-side containment pass over the bulk-read
commands.
**Reported, deliberately deferred: 2026-09-24.**

## Template authoring

### `loadExampleTemplates` skips the old-format check `LoadTemplates` applies

**What happens.** `template.LoadTemplates` rejects a template that still carries the old
`required`/`optional`/`body` keys; `loadExampleTemplates` (`cmd/akb/template.go`), the reader
behind `akb template list --examples` and `akb template get --examples`, unmarshals the
embedded showcase files without that check. The embedded set is build-controlled and covered
by template tests, so an old-format file cannot ship through it today; an in-package caller
of the function would get no rejection.

**Why deferred:** hardening an unreachable path — the duplicate checks would have to be
re-derived on the embedded reader for no current failure mode.
**Reported, deliberately deferred: 2026-09-24.**

### Templates whose rules read optional keys unguarded cannot be overwritten

**What happens.** `akb template write` re-evaluates the pass mockup with each schema-optional
key it supplies removed, and once more as a no-op update. A template authored before those
checks whose CEL rules read an optional key without `has()` fails the stripped evaluation,
so an overwrite of that template is rejected even with `--force` until the rule is guarded
or the key is marked `required: true`. Nothing migrates existing templates.

**Why deferred:** the checks are the deliberate proof that `optional: true` fields are safe;
the remedy is a per-template authoring edit, and the error names the key to guard.
**Reported, deliberately deferred: 2026-09-24.**

### Dead `unknown type` guards keep the terse message

**What happens.** `cmd/akb/write.go` (three sites) and `cmd/akb/append.go` (one) guard the
template lookup with `if !ok { return fmt.Errorf("unknown type %q", fm.Type) }` after
`frontmatter.ValidateType` already rejected an unknown type. The guards are unreachable, and
their message lacks the authoring guidance `ValidateType` gives; harmonizing the two was
scheduled and never run.

**Why deferred:** unreachable code — the message would matter only if the earlier check were
removed.
**Reported, deliberately deferred: 2026-09-24.**

### Required-field presence keys `type` and `title` on non-emptiness

**What happens.** `frontmatterKeyPresent` (`cmd/akb/templates_write.go`) reports `type` and
`title` present only when non-empty — `Parse` routes them out of `Fields`, and an explicitly
empty value counts as absent — while every other key is reported present by key-set
membership. The lint-side mirror `frontmatterFieldPresent` has the same semantics. Every
write path rejects an empty `type` or `title` first, so the two readings agree on real pages.

**Why deferred:** an empty `type` or `title` is not a usable page, so key-set membership
would not change an outcome today.
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

### The retired freshness and summary thresholds stay exported

**What happens.** `internal/lint/thresholds.go` still exports `FreshnessHalfLifeDays`,
`FreshnessScoreThreshold`, `SummaryMinLength`, and `SummaryMaxLength`, whose checkers were
removed, alongside the consumed `ProvenanceDriftThreshold`. Only `engine_test.go` reads the
four.

**Why deferred:** decision recorded — keep: they are stable threshold spellings a consumer
may read, and deleting exported identifiers is a breaking change with no current need; a
removal belongs to a lint-package cleanup of its own.
**Reported, deliberately deferred: 2026-09-24.**

### `internal/lint/AGENTS.md` lists one of the five threshold constants

**What happens.** The THRESHOLDS block in `internal/lint/AGENTS.md` shows only
`ProvenanceDriftThreshold`, while `thresholds.go` defines five exported constants, four of
them unconsumed by the engine.

**Why deferred:** documentation drift on an internal map; the block is corrected together
with the constants' fate (the entry above).
**Reported, deliberately deferred: 2026-09-24.**

## Test hygiene

### Orphan status goes stale until the linking page is rewritten

**What happens.** Write-time link resolution means an ambiguous-basename inbound link keeps
suppressing orphan status for its stale target until the linking page is rewritten or
`akb index rebuild` runs. Documented in the `GetOrphans` doc comment, but no automated test
pins the staleness.

**Why deferred:** documented behavior with a test gap, not a defect.
**Reported, deliberately deferred: 2026-09-24.**

### The embedded TEMPLATE.md carries claims no test pins

**What happens.** The kb-management skill's `references/TEMPLATE.md` documents the
optional-key stripping and self-succession checks, the `has()` discipline, and the
three-surface failure contract. No test reads the document or asserts those claims against
the code — the only reference to it is the literal path in the write-time error message. The
behavior paths are pinned; a drift between the shipped document and the code would be silent.

**Why deferred:** a test gap on a documentation surface; a doc-claim test would have to
track prose.
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

## Fail-closed template loading

### One unparseable template file blocks every approve

**What happens.** `akb approve` loads the base's templates fail-closed (`approveTemplates` →
`template.LoadTemplates`) before it approves anything, so a single corrupt or unparseable YAML
file in `.akb/templates/` makes every approve fail — including approvals of a page whose type
is unrelated to the broken template — with the load error naming the offending file. The write
path loads templates with the same fail-closed behavior.

**Why deferred:** a broken template file is a base-level fault the author must repair;
silently skipping unloadable templates would hide real corruption.
**Reported, deliberately deferred: 2026-09-24.**

## Lint coverage gaps

### A page with frontmatter but an empty type escapes every checker

**What happens.** A page whose frontmatter carries an empty `type` value is skipped by every
checker that could flag it: the required-fields check skips it (there is no template to check
it against), the type-orphan check skips it through its pre-existing empty-type exclusion, and
the missing-frontmatter check skips it (frontmatter is present). No checker reports the page.

**Why deferred:** the page is degenerate — a type-less page cannot be validated against any
schema — and closing the gap means deciding which checker owns empty-type pages, a
checker-responsibility design call rather than a local fix.
**Reported, deliberately deferred: 2026-09-24.**

## Git and storage edges

### A staged-but-reverted index still fails a no-op template write

**What happens.** `akb template write --force` can still fail at the commit step when the
template files are in a staged-but-reverted git state: the index differs from HEAD while the
worktree matches HEAD and matches the content being written. The byte-identity check
(`templateFilesMatch`) passes, but `storage.NothingToCommit` returns false because
`git status --porcelain` reports the staged difference, so the write is not treated as a
no-op. The swap-plus-commit path then stages the worktree content — identical to HEAD,
flattening the index — and git refuses the commit. The failure surfaces with the
`commit template write: git commit:` prefix; git's own wording after it varies
(`nothing to commit, working tree clean`, or `nothing added to commit but untracked files
present` when untracked files exist).

**Why deferred:** detecting the index-versus-worktree-versus-HEAD divergence needs more
git-state machinery than this pre-existing edge justifies, and the failure message already
names the cause.
**Reported, deliberately deferred: 2026-09-24.**

### `NothingToCommit` hands git unnormalized OS-specific pathspecs

**What happens.** KB-relative git pathspecs are built with `filepath.Join`
(`templateCommitPaths`), so on Windows they carry backslash separators.
`storage.NothingToCommit` passes them to `git status --porcelain --` unnormalized, so the
status pathspecs of the template-write no-op detection would use backslashes and mismatch
git's slash-separated pathspec format; the no-op check would misbehave on Windows.
`storage.CommitFiles` is not affected: it normalizes paths via `recordablePaths`
(`filepath.ToSlash`) before staging and committing.

**Why deferred:** a pre-existing convention across the storage layer; the project's toolchain
(Makefile, Nix flake) targets unix and Windows is not a supported platform.
**Reported, deliberately deferred: 2026-09-24.**
