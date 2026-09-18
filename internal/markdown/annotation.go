// Package markdown parses wikilinks, annotations, and provenance markers.
package markdown

import (
	"regexp"
	"strings"
)

// Annotation represents an olw-auto HTML comment in page content.
type Annotation struct {
	Type     string
	Fields   map[string]string
	Position int
}

var annotationRe = regexp.MustCompile(`<!--\s*olw-auto:\s*(.+?)\s*-->`)

// ParseAnnotations extracts olw-auto annotations from content.
// Annotations inside fenced code blocks or inline code are ignored.
func ParseAnnotations(content string) []Annotation {
	if content == "" {
		return nil
	}

	// Only code ranges are excluded: an annotation is itself an HTML comment,
	// so including comment ranges here would exclude every real annotation.
	var excluded exclusionSet
	excluded.ranges = append(excluded.ranges, fencedCodeBlockRanges(content)...)
	excluded.ranges = append(excluded.ranges, inlineCodeRanges(content)...)

	matches := annotationRe.FindAllStringSubmatchIndex(content, -1)
	var annotations []Annotation
	for _, m := range matches {
		if excluded.isExcluded(m[0], m[1]) {
			continue
		}

		fieldsStr := content[m[2]:m[3]]
		fields := parseAnnotationFields(fieldsStr)
		annotations = append(annotations, Annotation{
			Type:     "olw-auto",
			Fields:   fields,
			Position: m[0],
		})
	}

	return annotations
}

func parseAnnotationFields(s string) map[string]string {
	fields := make(map[string]string)
	s = strings.TrimSpace(s)
	if s == "" {
		return fields
	}

	tokens := strings.Fields(s)
	for _, token := range tokens {
		key, val, ok := strings.Cut(token, "=")
		if !ok {
			continue
		}
		fields[key] = val
	}

	return fields
}

// StripAnnotations removes olw-auto annotation comments from content, leaving
// the annotations inside fenced code blocks and inline code untouched — the
// same exclusion ParseAnnotations applies when it reads them.
func StripAnnotations(content string) string {
	if content == "" {
		return content
	}

	// Only code ranges are excluded: an annotation is itself an HTML comment,
	// so including comment ranges here would exclude every real annotation.
	var excluded exclusionSet
	excluded.ranges = append(excluded.ranges, fencedCodeBlockRanges(content)...)
	excluded.ranges = append(excluded.ranges, inlineCodeRanges(content)...)

	matches := annotationRe.FindAllStringIndex(content, -1)

	var stripped strings.Builder
	lastEnd := 0
	for _, m := range matches {
		start := m[0]
		end := m[1]
		stripped.WriteString(content[lastEnd:start])
		if excluded.isExcluded(start, end) {
			stripped.WriteString(content[start:end])
		}
		lastEnd = end
	}
	stripped.WriteString(content[lastEnd:])

	return stripped.String()
}
