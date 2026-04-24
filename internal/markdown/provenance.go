package markdown

import (
	"regexp"
	"strings"
)

// ProvenanceMarker represents a ^[type] provenance marker in content.
type ProvenanceMarker struct {
	Type     string // "inferred", "ambiguous", "extracted", or custom
	Position int    // byte offset in content
}

var provenanceRe = regexp.MustCompile(`\^\[([^\]]+?)\]`)

// ParseProvenanceMarkers extracts all ^[type] provenance markers from content.
// Markers inside code blocks, inline code, and HTML comments are excluded.
func ParseProvenanceMarkers(content string) []ProvenanceMarker {
	if content == "" {
		return nil
	}

	excluded := computeExclusions(content)

	matches := provenanceRe.FindAllStringSubmatchIndex(content, -1)
	var markers []ProvenanceMarker
	for _, m := range matches {
		innerStart := m[2]
		innerEnd := m[3]
		if excluded.isExcluded(innerStart, innerEnd) {
			continue
		}

		markerType := content[innerStart:innerEnd]
		markers = append(markers, ProvenanceMarker{
			Type:     markerType,
			Position: m[0], // start of full match
		})
	}

	return markers
}

// CountMarkersByType returns a map of marker type to count.
func CountMarkersByType(markers []ProvenanceMarker) map[string]int {
	counts := make(map[string]int)
	for _, m := range markers {
		counts[m.Type]++
	}
	return counts
}

var stripProvenanceRe = regexp.MustCompile(`\^\[(inferred|ambiguous|extracted)\]`)

// StripProvenanceMarkers removes ^[inferred], ^[ambiguous], and ^[extracted]
// markers from content, preserving those inside code blocks, inline code,
// and HTML comments.
func StripProvenanceMarkers(content string) string {
	if content == "" {
		return ""
	}

	excluded := computeExclusions(content)
	matches := stripProvenanceRe.FindAllStringIndex(content, -1)

	var result strings.Builder
	lastEnd := 0
	for _, m := range matches {
		start := m[0]
		end := m[1]
		result.WriteString(content[lastEnd:start])
		if excluded.isExcluded(start, end) {
			result.WriteString(content[start:end])
		}
		lastEnd = end
	}
	result.WriteString(content[lastEnd:])

	return result.String()
}
