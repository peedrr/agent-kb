package markdown

import (
	"regexp"
	"strings"
)

type Annotation struct {
	Type     string
	Fields   map[string]string
	Position int
}

var annotationRe = regexp.MustCompile(`<!--\s*olw-auto:\s*(.+?)\s*-->`)

func ParseAnnotations(content string) []Annotation {
	if content == "" {
		return nil
	}

	matches := annotationRe.FindAllStringSubmatchIndex(content, -1)
	var annotations []Annotation
	for _, m := range matches {
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
