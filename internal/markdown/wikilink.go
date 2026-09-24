package markdown

import (
	"strings"
)

// Wikilink represents a [[target]], [[target|display]], [[target#heading]], or
// [[display]](dest) link.
type Wikilink struct {
	Target     string
	Display    string
	HasHeading bool
	Heading    string
	// Destination is the normalized explicit destination from an adjacent
	// trailing (dest). It is empty for the bracket-only forms.
	Destination string
	// Start and End are byte offsets into the source: the half-open span
	// [Start, End) covers the whole token, the [[...]] and any non-empty
	// adjacent (dest).
	Start int
	End   int
}

// wikilinkToken is the byte span of one wikilink-shaped token: the opening
// bracket run, the inner text, and the closing bracket run.
type wikilinkToken struct {
	start      int // first '[' of the opening run
	innerStart int
	innerEnd   int
	end        int // just past the closing run
}

// scanWikilinkTokens finds every wikilink-shaped token in content, left to
// right. A token opens with a run of two or more '[' and closes at the first
// following ']' run that is at least as long as the opening run; it consumes
// exactly as many ']' as the opening run has '['. An opening run that cannot
// close that way is retried one '[' further right, so [[[a]]] is the single
// token [[[a]]] whose inner text is "a" — never a bracket fragment such as
// "[a".
func scanWikilinkTokens(content string) []wikilinkToken {
	var tokens []wikilinkToken
	for i := 0; i+1 < len(content); {
		if content[i] != '[' {
			i++
			continue
		}
		open := countRun(content, i, '[')
		if open < 2 {
			i += open
			continue
		}

		closeStart := strings.IndexByte(content[i+open:], ']')
		if closeStart < 0 {
			break
		}
		closeStart += i + open

		if closeStart == i+open || countRun(content, closeStart, ']') < open {
			i++
			continue
		}

		end := closeStart + open
		tokens = append(tokens, wikilinkToken{
			start:      i,
			innerStart: i + open,
			innerEnd:   closeStart,
			end:        end,
		})
		i = end
	}
	return tokens
}

// countRun returns the length of the run of c that starts at s[i].
func countRun(s string, i int, c byte) int {
	n := 0
	for i+n < len(s) && s[i+n] == c {
		n++
	}
	return n
}

// ParseWikilinks extracts wikilinks from markdown content.
func ParseWikilinks(content string) []Wikilink {
	if content == "" {
		return nil
	}

	excluded := computeExclusions(content)

	// suppressedUntil is the end offset of an accepted explicit destination
	// whose internals cannot hold a link of their own: a bracket run inside a
	// quoted title — [[a]](x.md "see [[b]]") — or inside a path destination —
	// [[a]](notes/[[weird]].md) — is part of that destination, not a token of
	// its own. A token that opens before the offset is skipped. A destination
	// that is itself a bracket token ([[a]]([[b]])) sets no suppression: there
	// the inner token wins and the outer one is the malformed side.
	suppressedUntil := 0

	var links []Wikilink
	for _, token := range scanWikilinkTokens(content) {
		if token.start < suppressedUntil {
			continue
		}
		if excluded.isExcluded(token.innerStart, token.innerEnd) {
			continue
		}

		inner := content[token.innerStart:token.innerEnd]
		wl, hasPipeDisplay := parseWikilinkInner(inner)
		wl.Start = token.start
		wl.End = token.end

		// An explicit destination requires the opening paren to be adjacent to
		// the closing brackets (CommonMark's inline-link rule); whitespace
		// between them keeps the link a plain wikilink. A non-empty destination
		// that is a valid link destination overrides the target; an empty ()
		// or an invalid destination is a plain wikilink, parens ignored.
		if wl.End < len(content) && content[wl.End] == '(' {
			if raw, end, ok := scanParenDestination(content[wl.End:]); ok && isValidLinkDestination(raw) {
				if dest := normalizeDest(raw); dest != "" {
					// An explicit destination takes precedence over any #heading in
					// the bracket part: the destination is the target and the
					// heading is discarded, never merged into the destination. A
					// display derived from that heading falls back to the
					// bracket-part label; a pipe display still wins.
					if wl.HasHeading && !hasPipeDisplay {
						wl.Display = wl.Target
					}
					wl.Target = dest
					wl.Destination = dest
					wl.HasHeading = false
					wl.Heading = ""
					wl.End += end
					if !strings.HasPrefix(dest, "[[") {
						suppressedUntil = wl.End
					}
				}
			}
		}

		links = append(links, wl)
	}

	return links
}

// scanParenDestination reads a parenthesized explicit destination starting at
// s[0] == '('. It returns the raw text between the outer parens, the byte
// offset just past the matching ')', and whether a matching ')' exists.
//
// The scan follows the CommonMark inline-link tail: an angle-bracketed
// destination ends at its closing '>', a bare destination at the first
// unescaped whitespace or unbalanced ')', and a title that follows the
// destination can contain a ')'. Those spans are skipped so a ')' inside one
// does not end the token early.
func scanParenDestination(s string) (string, int, bool) {
	depth := 1
	atDestStart := true
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
			atDestStart = false
		case ')':
			depth--
			if depth == 0 {
				return s[1:i], i + 1, true
			}
			atDestStart = false
		case '<':
			if atDestStart {
				atDestStart = false
				if end, ok := scanAngleDestination(s, i); ok {
					i = end - 1
				}
			}
		case '"', '\'':
			// A title follows the destination, so it is always preceded by
			// whitespace or by the '>' that closed an angle-bracketed
			// destination.
			if i > 1 && (isSpace(s[i-1]) || s[i-1] == '>') {
				if end, ok := scanTitle(s, i); ok {
					i = end - 1
				}
			}
			atDestStart = false
		case '\\':
			if i+1 < len(s) && isPunct(s[i+1]) {
				i++
			}
			atDestStart = false
		case ' ', '\t', '\n', '\v', '\f', '\r':
			// Leading whitespace does not end the destination start.
		default:
			atDestStart = false
		}
	}
	return "", 0, false
}

// isValidLinkDestination reports whether raw, the text scanned between an
// adjacent pair of parens, is a valid CommonMark link destination with an
// optional trailing link title. Leading and trailing whitespace is ignored;
// the destination is valid when it is either angle-bracketed with no newline
// inside the brackets, or free of whitespace entirely. Balanced nested
// parentheses are accepted in both forms. A title, when present, is a quoted
// or parenthesized string that follows the destination.
func isValidLinkDestination(raw string) bool {
	_, end, ok := parseLinkDestination(raw)
	return ok && end == len(raw)
}

// parseLinkDestination parses a CommonMark link destination and an optional
// trailing link title from the start of s, skipping leading whitespace. It
// returns the raw destination text (angle brackets kept, title excluded), the
// byte offset just past the parsed destination-with-title, and whether a
// destination was found.
func parseLinkDestination(s string) (string, int, bool) {
	i := skipSpaces(s, 0)
	if i == len(s) {
		return "", 0, false
	}
	dest, i, ok := scanDestination(s, i)
	if !ok {
		return "", 0, false
	}
	i = skipSpaces(s, i)
	if i == len(s) {
		return dest, i, true
	}
	if s[i] != '"' && s[i] != '\'' && s[i] != '(' {
		return "", 0, false
	}
	end, ok := scanTitle(s, i)
	if !ok {
		return "", 0, false
	}
	return dest, skipSpaces(s, end), true
}

// scanDestination parses a CommonMark link destination starting at s[i]:
// either an angle-bracketed destination, or a bare one that ends at the first
// unescaped whitespace or unbalanced ')'. A backslash-escaped punctuation
// character is part of the destination.
func scanDestination(s string, i int) (string, int, bool) {
	if s[i] == '<' {
		end, ok := scanAngleDestination(s, i)
		if !ok {
			return "", 0, false
		}
		return s[i:end], end, true
	}
	depth := 0
	for j := i; j < len(s); j++ {
		c := s[j]
		if c == '\\' && j+1 < len(s) && isPunct(s[j+1]) {
			j++
			continue
		}
		switch c {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return s[i:j], j, j > i
			}
			depth--
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return s[i:j], j, j > i
		}
	}
	return s[i:], len(s), len(s) > i
}

// scanAngleDestination returns the byte offset just past the '>' that closes
// the angle-bracketed destination starting at s[i] == '<'. It reports false
// when no '>' appears before the end of the line; a backslash-escaped
// punctuation character does not close the destination.
func scanAngleDestination(s string, i int) (int, bool) {
	for j := i + 1; j < len(s); j++ {
		switch {
		case s[j] == '\n' || s[j] == '\r':
			return 0, false
		case s[j] == '\\' && j+1 < len(s) && isPunct(s[j+1]):
			j++
		case s[j] == '>':
			return j + 1, true
		}
	}
	return 0, false
}

// scanTitle returns the byte offset just past the closing delimiter of the
// link title starting at s[i], which must be a double quote, single quote, or
// '('. A title ends at the first unescaped matching delimiter; a
// parenthesized title may not contain a nested '(', and a backslash-escaped
// punctuation character does not close the title.
func scanTitle(s string, i int) (int, bool) {
	opener := s[i]
	closer := opener
	if opener == '(' {
		closer = ')'
	}
	for j := i + 1; j < len(s); j++ {
		switch {
		case s[j] == '\\' && j+1 < len(s) && isPunct(s[j+1]):
			j++
		case s[j] == closer:
			return j + 1, true
		case opener == '(' && s[j] == '(':
			return 0, false
		}
	}
	return 0, false
}

// skipSpaces returns the offset of the first byte at or after i that is not
// whitespace.
func skipSpaces(s string, i int) int {
	for i < len(s) && isSpace(s[i]) {
		i++
	}
	return i
}

// isSpace reports whether c is whitespace in CommonMark's sense.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r'
}

// isPunct reports whether c is an ASCII punctuation character, the set
// CommonMark allows a backslash to escape.
func isPunct(c byte) bool {
	return c >= '!' && c <= '/' || c >= ':' && c <= '@' || c >= '[' && c <= '`' || c >= '{' && c <= '~'
}

// NormalizeBareDestination cleans a link destination that carries no link
// title — the form goldmark reports, because it keeps a title in a field of
// its own rather than inside the destination. Surrounding whitespace and angle
// brackets are dropped, as are a leading "./" and a trailing ".md" (the link
// graph's resolution step re-adds the extension). Repeated affixes collapse,
// so "././a.md.md" becomes "a".
func NormalizeBareDestination(dest string) string {
	d := strings.TrimSpace(dest)
	if len(d) >= 2 && strings.HasPrefix(d, "<") && strings.HasSuffix(d, ">") {
		d = d[1 : len(d)-1]
	}
	d = strings.TrimSpace(d)
	for strings.HasPrefix(d, "./") {
		d = strings.TrimPrefix(d, "./")
	}
	for strings.HasSuffix(d, ".md") {
		d = strings.TrimSuffix(d, ".md")
	}
	return d
}

// normalizeDest cleans an explicit destination: an optional trailing link
// title is dropped, then NormalizeBareDestination strips the surrounding
// whitespace and angle brackets, the leading "./", and the trailing ".md".
func normalizeDest(dest string) string {
	d := strings.TrimSpace(dest)
	if parsed, _, ok := parseLinkDestination(d); ok {
		d = parsed
	}
	return NormalizeBareDestination(d)
}

// parseWikilinkInner parses the text between the brackets. It also reports
// whether a non-empty pipe display was present, so callers can tell a display
// the author wrote from one derived from the heading.
func parseWikilinkInner(inner string) (Wikilink, bool) {
	wl := Wikilink{}

	hasPipeDisplay := false
	var customDisplay string
	if pipeIdx := strings.Index(inner, "|"); pipeIdx >= 0 {
		customDisplay = inner[pipeIdx+1:]
		hasPipeDisplay = customDisplay != ""
		inner = inner[:pipeIdx]
	}

	if hashIdx := strings.Index(inner, "#"); hashIdx >= 0 {
		wl.HasHeading = true
		wl.Heading = inner[hashIdx+1:]
		wl.Target = inner[:hashIdx]
	} else {
		wl.Target = inner
	}

	switch {
	case customDisplay != "":
		wl.Display = customDisplay
	case wl.HasHeading:
		wl.Display = wl.Heading
	default:
		wl.Display = wl.Target
	}

	return wl, hasPipeDisplay
}

type exclusion struct {
	start int
	end   int
}

type exclusionSet struct {
	ranges []exclusion
}

func (es *exclusionSet) isExcluded(start, end int) bool {
	for _, r := range es.ranges {
		if start >= r.start && end <= r.end {
			return true
		}
	}
	return false
}

func computeExclusions(content string) exclusionSet {
	var es exclusionSet

	es.ranges = append(es.ranges, fencedCodeBlockRanges(content)...)
	es.ranges = append(es.ranges, indentedCodeBlockRanges(content)...)
	es.ranges = append(es.ranges, inlineCodeRanges(content)...)
	es.ranges = append(es.ranges, htmlCommentRanges(content)...)
	es.ranges = append(es.ranges, escapedBracketRanges(content)...)

	return es
}

// indentedCodeBlockRanges returns the byte ranges of indented code blocks: runs
// of lines indented four or more columns past the content column of the
// innermost open list item (four columns at the top level). A list item's
// content column is where its text starts, so item content such as
// "- item"'s four-space continuation is prose while four further columns are
// code. An indented line that follows a paragraph without an intervening blank
// line continues that paragraph — a lazy continuation — so only a block
// reaching the start of the content or following a blank line is code, and a
// lazy continuation line closes no open item even when it sits below the item's
// content column. Lines strictly inside a fenced code block are literal text
// and leave the list bookkeeping untouched; the fence delimiter lines are
// ordinary markdown and run it like any other line.
func indentedCodeBlockRanges(content string) []exclusion {
	var ranges []exclusion
	fencedInterior := fencedCodeBlockInteriors(content, fencedCodeBlockRanges(content))
	blockStart := -1
	prevBlank := true

	// contentColumns holds the content columns of the open list items,
	// outermost first; the last entry is the innermost open item.
	var contentColumns []int

	// paragraphOpen reports whether the previous line left a paragraph open.
	// Its text continues on a later line even when that line is indented below
	// the open item's content column — a lazy continuation, which closes no
	// item. A line that opens a block which interrupts a paragraph closes items
	// the way any other block-start line does.
	paragraphOpen := false

	for lineStart := 0; lineStart < len(content); {
		lineEnd := lineStart
		for lineEnd < len(content) && content[lineEnd] != '\n' {
			lineEnd++
		}
		line := content[lineStart:lineEnd]

		if lineOverlapsRanges(fencedInterior, lineStart, lineEnd) {
			// The lines strictly inside a fenced code block are opaque: they
			// are literal text, so they neither open nor close list items and
			// leave the blank-line state and any pending indented block
			// untouched. The delimiter lines fall through to the bookkeeping
			// below.
			lineStart = lineEnd + 1
			continue
		}

		trimmed := strings.TrimSpace(line)
		lineIndent, _ := leadingWhitespaceColumn(line)
		itemContentColumn, isMarker := scanListMarker(line)

		threshold := 4
		if n := len(contentColumns); n > 0 {
			threshold = contentColumns[n-1] + 4
		}
		isCodeLine := lineIndent >= threshold
		isLazyContinuation := false
		if isCodeLine && blockStart < 0 && !prevBlank {
			// A lazy continuation: the indented line continues the preceding
			// paragraph rather than starting an indented code block.
			isCodeLine = false
			isLazyContinuation = true
		}

		// After a blank line, a line indented below the innermost open item's
		// content column falls outside that item, so the items it falls
		// outside of close. The line is indented code only when it is still
		// indented four or more columns past whatever item remains open, or
		// four or more columns at the top level when none does; otherwise it
		// is that item's prose. This outranks marker detection: a marker
		// indented four or more columns is code, not a new item.
		if blockStart < 0 && prevBlank && trimmed != "" && lineIndent >= 4 &&
			len(contentColumns) > 0 && contentColumns[len(contentColumns)-1] > lineIndent {
			for len(contentColumns) > 0 && contentColumns[len(contentColumns)-1] > lineIndent {
				contentColumns = contentColumns[:len(contentColumns)-1]
			}
			remainingThreshold := 4
			if n := len(contentColumns); n > 0 {
				remainingThreshold = contentColumns[n-1] + 4
			}
			isCodeLine = lineIndent >= remainingThreshold
		}

		switch {
		case isCodeLine:
			paragraphOpen = false
			if blockStart < 0 {
				blockStart = lineStart
			}
		case isLazyContinuation:
			// Paragraph continuation text: it opens no item and closes none.
			paragraphOpen = true
		case isMarker:
			// A marker below an open item's content column cannot be nested
			// inside that item, so it closes the item and any deeper ones.
			for len(contentColumns) > 0 && contentColumns[len(contentColumns)-1] > lineIndent {
				contentColumns = contentColumns[:len(contentColumns)-1]
			}
			contentColumns = append(contentColumns, itemContentColumn)
			if blockStart >= 0 {
				ranges = append(ranges, exclusion{start: blockStart, end: lineStart})
				blockStart = -1
			}
			paragraphOpen = true
		case trimmed != "":
			// A non-blank line below an open item's content column lies outside
			// the item, so it closes the item and any deeper ones. Blank lines
			// never close a list. A line that continues an open paragraph is
			// lazy continuation text instead: it lies outside the item's indent
			// but stays inside its paragraph, so it closes nothing.
			interrupt := startsParagraphInterrupt(line)
			if paragraphOpen && !interrupt {
				break
			}
			for len(contentColumns) > 0 && contentColumns[len(contentColumns)-1] > lineIndent {
				contentColumns = contentColumns[:len(contentColumns)-1]
			}
			if blockStart >= 0 {
				ranges = append(ranges, exclusion{start: blockStart, end: lineStart})
				blockStart = -1
			}
			paragraphOpen = !interrupt
		default:
			paragraphOpen = false
			if blockStart >= 0 {
				ranges = append(ranges, exclusion{start: blockStart, end: lineStart})
				blockStart = -1
			}
		}

		prevBlank = trimmed == ""
		lineStart = lineEnd + 1
	}

	if blockStart >= 0 {
		ranges = append(ranges, exclusion{start: blockStart, end: len(content)})
	}
	return ranges
}

// lineOverlapsRanges reports whether the line span [lineStart, lineEnd)
// overlaps any of the ranges.
func lineOverlapsRanges(ranges []exclusion, lineStart, lineEnd int) bool {
	for _, r := range ranges {
		if lineStart < r.end && lineEnd > r.start {
			return true
		}
	}
	return false
}

// leadingWhitespaceColumn returns the column of the first non-whitespace byte
// of line and that byte's index. A tab advances to the next multiple of four,
// so a tab-indented line starts at column four.
func leadingWhitespaceColumn(line string) (column, index int) {
	for index < len(line) {
		switch line[index] {
		case ' ':
			column++
		case '\t':
			column += 4 - column%4
		default:
			return column, index
		}
		index++
	}
	return column, index
}

// scanListMarker reports whether line opens a list item and returns the column
// at which that item's content begins. A marker is '-', '*', or '+', or a run
// of one to nine digits followed by '.' or ')'; it must be followed by spaces,
// tabs, or end the line, so a word such as "-item" is not a marker. Three or
// more '-' or '*' characters separated only by spaces are a thematic break
// rather than a list item, and ten or more digits keep an ordered-marker line a
// paragraph. Up to four whitespace columns after the marker are part of the
// item's prefix, a tab advancing to the next multiple of four; more than four
// columns, or a marker ending the line, leave the content one column past the
// marker.
func scanListMarker(line string) (contentColumn int, ok bool) {
	markerColumn, i := leadingWhitespaceColumn(line)
	if i >= len(line) {
		return 0, false
	}

	width := 0
	switch line[i] {
	case '-', '*':
		if isThematicBreak(line[i:], line[i]) {
			return 0, false
		}
		width = 1
	case '+':
		width = 1
	default:
		for j := i; j < len(line) && line[j] >= '0' && line[j] <= '9'; j++ {
			width++
		}
		if width == 0 || width > 9 || i+width >= len(line) || (line[i+width] != '.' && line[i+width] != ')') {
			return 0, false
		}
		width++
	}

	afterMarker := i + width
	column := markerColumn + width
	whitespace := 0
	for afterMarker+whitespace < len(line) && (line[afterMarker+whitespace] == ' ' || line[afterMarker+whitespace] == '\t') {
		if line[afterMarker+whitespace] == '\t' {
			column += 4 - column%4
		} else {
			column++
		}
		whitespace++
	}
	atEndOfLine := afterMarker+whitespace == len(line)
	if !atEndOfLine && whitespace == 0 {
		return 0, false
	}

	if atEndOfLine || column-(markerColumn+width) > 4 {
		return markerColumn + width + 1, true
	}
	return column, true
}

// isThematicBreak reports whether s, which begins at a '-' or '*' bullet
// marker, consists of three or more repetitions of that marker separated only
// by spaces.
func isThematicBreak(s string, marker byte) bool {
	count := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case marker:
			count++
		case ' ':
		default:
			return false
		}
	}
	return count >= 3
}

// startsParagraphInterrupt reports whether line opens a block that interrupts
// an open paragraph: a fenced code block delimiter or a thematic break. Such a
// line closes the list items it falls outside of; an ordinary text line in the
// same position is a lazy continuation of the paragraph instead.
func startsParagraphInterrupt(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "```") {
		return true
	}
	if trimmed == "" || (trimmed[0] != '-' && trimmed[0] != '*') {
		return false
	}
	return isThematicBreak(trimmed, trimmed[0])
}

// escapedBracketRanges returns the inner range of every wikilink-shaped token
// whose opening '[' or closing ']' is backslash-escaped. CommonMark renders
// such a token as literal text rather than a link, so neither the link graph
// nor CEL should record it.
func escapedBracketRanges(content string) []exclusion {
	var ranges []exclusion
	for _, token := range scanWikilinkTokens(content) {
		if isEscapedByte(content, token.start) || isEscapedByte(content, token.innerEnd) {
			ranges = append(ranges, exclusion{start: token.start, end: token.innerEnd})
		}
	}
	return ranges
}

// isEscapedByte reports whether the byte at index i is escaped: preceded by an
// odd number of backslashes, the count CommonMark's escape rule uses.
func isEscapedByte(content string, i int) bool {
	count := 0
	for j := i - 1; j >= 0 && content[j] == '\\'; j-- {
		count++
	}
	return count%2 == 1
}

func fencedCodeBlockRanges(content string) []exclusion {
	var ranges []exclusion
	inBlock := false
	blockStart := 0

	for i := 0; i < len(content); {
		if strings.HasPrefix(content[i:], "```") {
			if !inBlock {
				inBlock = true
				blockStart = i
				i += 3
			} else {
				ranges = append(ranges, exclusion{start: blockStart, end: i + 3})
				inBlock = false
				i += 3
			}
		} else {
			i++
		}
	}

	return ranges
}

// fencedCodeBlockInteriors narrows each fenced code block range to the lines
// strictly between its delimiter lines: the interior begins just past the
// newline that ends the opening delimiter's line and ends at the start of the
// line holding the closing delimiter. A range whose delimiters share a line —
// an inline backtick span that the fence detector paired — has an empty
// interior and is dropped.
func fencedCodeBlockInteriors(content string, fences []exclusion) []exclusion {
	var interiors []exclusion
	for _, r := range fences {
		openLineEnd := strings.IndexByte(content[r.start:], '\n')
		if openLineEnd < 0 {
			continue
		}
		interiorStart := r.start + openLineEnd + 1

		// The closing delimiter is the range's final three backticks, so its
		// line starts at the last newline before them.
		closeLineStart := strings.LastIndexByte(content[:r.end-3], '\n') + 1

		if interiorStart >= closeLineStart {
			continue
		}
		interiors = append(interiors, exclusion{start: interiorStart, end: closeLineStart})
	}
	return interiors
}

func inlineCodeRanges(content string) []exclusion {
	var ranges []exclusion
	i := 0
	for i < len(content) {
		backtickRun := countBackticks(content, i)
		if backtickRun == 0 {
			i++
			continue
		}

		opener := i
		i += backtickRun

		closeIdx := findClosingBackticks(content, i, backtickRun)
		if closeIdx < 0 {
			continue
		}

		ranges = append(ranges, exclusion{start: opener, end: closeIdx + backtickRun})
		i = closeIdx + backtickRun
	}

	return ranges
}

func countBackticks(s string, pos int) int {
	count := 0
	for pos+count < len(s) && s[pos+count] == '`' {
		count++
	}
	return count
}

func findClosingBackticks(s string, start int, runLen int) int {
	for i := start; i <= len(s)-runLen; i++ {
		if s[i] == '`' {
			found := countBackticks(s, i)
			if found == runLen {
				return i
			}
		}
	}
	return -1
}

func htmlCommentRanges(content string) []exclusion {
	var ranges []exclusion
	commentOpen := "<!--"
	commentClose := "-->"

	i := 0
	for i <= len(content)-len(commentOpen) {
		if strings.HasPrefix(content[i:], commentOpen) {
			start := i
			closeIdx := strings.Index(content[i:], commentClose)
			if closeIdx < 0 {
				break
			}
			end := i + closeIdx + len(commentClose)
			ranges = append(ranges, exclusion{start: start, end: end})
			i = end
		} else {
			i++
		}
	}

	return ranges
}
