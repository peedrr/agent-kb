package markdown

import (
	"regexp"
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

var wikilinkRe = regexp.MustCompile(`\[\[([^\]]+?)\]\]`)

// ParseWikilinks extracts wikilinks from markdown content.
func ParseWikilinks(content string) []Wikilink {
	if content == "" {
		return nil
	}

	excluded := computeExclusions(content)

	matches := wikilinkRe.FindAllStringSubmatchIndex(content, -1)
	var links []Wikilink
	for _, m := range matches {
		innerStart := m[2]
		innerEnd := m[3]
		if excluded.isExcluded(innerStart, innerEnd) {
			continue
		}

		inner := content[innerStart:innerEnd]
		wl, hasPipeDisplay := parseWikilinkInner(inner)
		wl.Start = m[0]
		wl.End = m[1]

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

// normalizeDest cleans an explicit destination: an optional trailing link
// title is dropped, then surrounding whitespace and angle brackets, as are a
// leading "./" and a trailing ".md" (the link graph's resolution step re-adds
// the extension). Repeated affixes collapse, so "././a.md.md" becomes "a".
func normalizeDest(dest string) string {
	d := strings.TrimSpace(dest)
	if parsed, _, ok := parseLinkDestination(d); ok {
		d = parsed
	}
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
	es.ranges = append(es.ranges, inlineCodeRanges(content)...)
	es.ranges = append(es.ranges, htmlCommentRanges(content)...)

	return es
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
