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
		wl := parseWikilinkInner(inner)
		wl.Start = m[0]
		wl.End = m[1]

		// An explicit destination requires the opening paren to be adjacent to
		// the closing brackets (CommonMark's inline-link rule); whitespace
		// between them keeps the link a plain wikilink. A non-empty destination
		// overrides the target; an empty () is a plain wikilink, parens ignored.
		if wl.End < len(content) && content[wl.End] == '(' {
			if raw, end, ok := scanParenDestination(content[wl.End:]); ok {
				if dest := normalizeDest(raw); dest != "" {
					// An explicit destination takes precedence over any #heading in
					// the bracket part: the destination is the target and the two
					// are never combined.
					wl.Target = dest
					wl.Destination = dest
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

func parseWikilinkInner(inner string) Wikilink {
	wl := Wikilink{}

	var customDisplay string
	if pipeIdx := strings.Index(inner, "|"); pipeIdx >= 0 {
		customDisplay = inner[pipeIdx+1:]
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

	return wl
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
