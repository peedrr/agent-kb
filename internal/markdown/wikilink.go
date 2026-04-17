package markdown

import (
	"regexp"
	"strings"
)

type Wikilink struct {
	Target     string
	Display    string
	HasHeading bool
	Heading    string
}

var wikilinkRe = regexp.MustCompile(`\[\[([^\]]+?)\]\]`)

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
		links = append(links, parseWikilinkInner(inner))
	}

	return links
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
