# internal/markdown/

**Parent:** `./AGENTS.md`

## OVERVIEW

Markdown parsing utilities: wikilinks, quality annotations, provenance markers. Wikilinks and provenance markers are skipped inside fenced code blocks, inline code, and HTML comments; annotations are skipped inside code only, because an annotation is itself an HTML comment.

## FILES

| File | Purpose |
|------|---------|
| `wikilink.go` | Parse `[[target]]`, `[[target|display]]`, `[[target#heading]]` |
| `annotation.go` | Parse `<!-- olw-auto: key=val -->` HTML comments |
| `provenance.go` | Parse `^[inferred]`, `^[ambiguous]`, `^[extracted]` markers |
| `offset.go` | `offsetToLine(source, offset)` helper |

## KEY TYPES

| Type | Fields / Purpose |
|------|------------------|
| `Wikilink` | `Target`, `Display`, `HasHeading`, `Heading` |
| `Annotation` | `Type` (always "olw-auto"), `Fields` map, `Position` |
| `ProvenanceMarker` | `Type` (inferred/ambiguous/extracted), `Position` |

## EXCLUSION LOGIC

Wikilinks and provenance markers skip matches inside:
- Fenced code blocks (```...```)
- Inline code (backtick-delimited)
- HTML comments (<!-- ... -->)

Annotations skip matches inside fenced code blocks and inline code only: an annotation is itself an HTML comment, so comment ranges would exclude every annotation.

`computeExclusions()` builds the wikilink/provenance ranges; the annotation parser builds its own code-only set. `exclusionSet.isExcluded()` checks before accepting matches.

## KEY FUNCTIONS

| Function | Purpose |
|----------|---------|
| `ParseWikilinks(content)` | `[]Wikilink` with heading anchor support |
| `ParseAnnotations(content)` | `[]Annotation` from olw-auto HTML comments |
| `ParseProvenanceMarkers(content)` | `[]ProvenanceMarker` excluding protected regions |
| `StripAnnotations(content)` | Removes `olw-auto` annotations, leaving those inside code untouched |
| `StripProvenanceMarkers(content)` | Removes `^[inferred/ambiguous/extracted]` markers |
| `CountMarkersByType(markers)` | Returns `map[type]count` |
| `offsetToLine(source, offset)` | Returns 1-based line number from byte offset |

## NOTES

- Wikilink heading anchors (`#heading`) are stripped before link resolution
- `StripProvenanceMarkers` preserves markers inside code blocks/comments
- Used by `cmd/akb/approve.go` to clean annotations before setting `is_draft: false`
- `page.ast` is built by `internal/cel/pagebuilder.go`, which carries its own Goldmark walk; this package no longer parses the AST
