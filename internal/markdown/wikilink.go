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

// scanParenDestination reads a parenthesized destination starting at s[0] == '('.
// It returns the raw destination, the byte offset just past the matching ')',
// and whether a matching ')' exists.
func scanParenDestination(s string) (string, int, bool) {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return s[1:i], i + 1, true
			}
		}
	}
	return "", 0, false
}

// isValidLinkDestination reports whether raw, the text scanned between an
// adjacent pair of parens, is a valid CommonMark link destination. Surrounding
// whitespace is ignored; the remainder is valid when it is either
// angle-bracketed with no newline inside the brackets, or free of whitespace
// entirely. Balanced nested parentheses are accepted in both forms.
func isValidLinkDestination(raw string) bool {
	d := strings.TrimSpace(raw)
	if d == "" {
		return false
	}
	if len(d) >= 2 && strings.HasPrefix(d, "<") && strings.HasSuffix(d, ">") {
		return !strings.ContainsAny(d[1:len(d)-1], "\n\r")
	}
	return !strings.ContainsAny(d, " \t\n\r")
}

// normalizeDest cleans an explicit destination: surrounding whitespace and
// angle brackets are dropped, as are a leading "./" and a trailing ".md" (the
// link graph's resolution step re-adds the extension).
func normalizeDest(dest string) string {
	d := strings.TrimSpace(dest)
	if len(d) >= 2 && strings.HasPrefix(d, "<") && strings.HasSuffix(d, ">") {
		d = d[1 : len(d)-1]
	}
	d = strings.TrimPrefix(d, "./")
	d = strings.TrimSuffix(d, ".md")
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
