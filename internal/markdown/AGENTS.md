# internal/markdown/

**Parent:** `./AGENTS.md`

## OVERVIEW

Markdown parsing utilities: wikilinks, quality annotations, provenance markers. All parsers exclude fenced code blocks, inline code, and HTML comments.

## FILES

| File | Purpose |
|------|---------|
| `wikilink.go` | Parse `[[target]]`, `[[target|display]]`, `[[target#heading]]` |
| `annotation.go` | Parse `<!-- olw-auto: key=val -->` HTML comments |
| `provenance.go` | Parse `^[inferred]`, `^[ambiguous]`, `^[extracted]` markers |

## KEY TYPES

| Type | Fields / Purpose |
|------|------------------|
| `Wikilink` | `Target`, `Display`, `HasHeading`, `Heading` |
| `Annotation` | `Type` (always "olw-auto"), `Fields` map, `Position` |
| `ProvenanceMarker` | `Type` (inferred/ambiguous/extracted), `Position` |

## EXCLUSION LOGIC

All three parsers skip matches inside:
- Fenced code blocks (```...```)
- Inline code (backtick-delimited)
- HTML comments (<!-- ... -->)

`computeExclusions()` builds ranges; `exclusionSet.isExcluded()` checks before accepting matches.

## KEY FUNCTIONS

| Function | Purpose |
|----------|---------|
| `ParseWikilinks(content)` | `[]Wikilink` with heading anchor support |
| `ParseAnnotations(content)` | `[]Annotation` from olw-auto HTML comments |
| `ParseProvenanceMarkers(content)` | `[]ProvenanceMarker` excluding protected regions |
| `StripProvenanceMarkers(content)` | Removes `^[inferred/ambiguous/extracted]` markers |
| `CountMarkersByType(markers)` | Returns `map[type]count` |

## NOTES

- Wikilink heading anchors (`#heading`) are stripped before link resolution
- `StripProvenanceMarkers` preserves markers inside code blocks/comments
- Used by `cmd/akb/approve.go` to clean annotations before setting `is_draft: false`