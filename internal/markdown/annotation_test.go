package markdown

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseAnnotations_Simple(t *testing.T) {
	content := "<!-- olw-auto: confidence=low -->"
	annotations := ParseAnnotations(content)

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Type != "olw-auto" {
		t.Errorf("expected Type 'olw-auto', got %q", annotations[0].Type)
	}

	expected := map[string]string{"confidence": "low"}
	if !reflect.DeepEqual(annotations[0].Fields, expected) {
		t.Errorf("expected Fields %v, got %v", expected, annotations[0].Fields)
	}
}

func TestParseAnnotations_MultipleFields(t *testing.T) {
	content := "<!-- olw-auto: confidence=low score=0.32 -->"
	annotations := ParseAnnotations(content)

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	expected := map[string]string{"confidence": "low", "score": "0.32"}
	if !reflect.DeepEqual(annotations[0].Fields, expected) {
		t.Errorf("expected Fields %v, got %v", expected, annotations[0].Fields)
	}
}

func TestParseAnnotations_RegularHTMLComment(t *testing.T) {
	content := "<!-- regular comment -->"
	annotations := ParseAnnotations(content)

	if len(annotations) != 0 {
		t.Fatalf("expected 0 annotations, got %d", len(annotations))
	}
}

func TestParseAnnotations_MalformedNoFields(t *testing.T) {
	content := "<!-- olw-auto: -->"
	annotations := ParseAnnotations(content)

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if len(annotations[0].Fields) != 0 {
		t.Errorf("expected empty Fields map, got %v", annotations[0].Fields)
	}
}

func TestParseAnnotations_MultipleAnnotations(t *testing.T) {
	content := `Some text <!-- olw-auto: confidence=high --> more text <!-- olw-auto: verified=true -->`
	annotations := ParseAnnotations(content)

	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}

	if annotations[0].Fields["confidence"] != "high" {
		t.Errorf("expected first annotation confidence=high, got %q", annotations[0].Fields["confidence"])
	}

	if annotations[1].Fields["verified"] != "true" {
		t.Errorf("expected second annotation verified=true, got %q", annotations[1].Fields["verified"])
	}
}

func TestParseAnnotations_NoHTMLComments(t *testing.T) {
	content := "This is just plain text with no HTML comments."
	annotations := ParseAnnotations(content)

	if len(annotations) != 0 {
		t.Fatalf("expected 0 annotations, got %d", len(annotations))
	}
}

func TestParseAnnotations_MixedComments(t *testing.T) {
	content := `<!-- regular comment --> some text <!-- olw-auto: confidence=low -->`
	annotations := ParseAnnotations(content)

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Fields["confidence"] != "low" {
		t.Errorf("expected confidence=low, got %q", annotations[0].Fields["confidence"])
	}
}

func TestParseAnnotations_FencedCodeBlock(t *testing.T) {
	content := "Some text\n```\n<!-- olw-auto: confidence=low -->\n```\nBut <!-- olw-auto: verified=true --> here."
	annotations := ParseAnnotations(content)

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Fields["verified"] != "true" {
		t.Errorf("expected verified=true, got %v", annotations[0].Fields)
	}
}

func TestParseAnnotations_InlineCode(t *testing.T) {
	content := "Use `<!-- olw-auto: confidence=low -->` in a doc, but <!-- olw-auto: verified=true --> is real."
	annotations := ParseAnnotations(content)

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Fields["verified"] != "true" {
		t.Errorf("expected verified=true, got %v", annotations[0].Fields)
	}
}

func TestStripAnnotations_ProseAnnotation(t *testing.T) {
	content := "Some text.\n<!-- olw-auto: confidence=low -->\nMore text."
	stripped := StripAnnotations(content)

	if strings.Contains(stripped, "olw-auto") {
		t.Errorf("prose annotation should be removed, got: %q", stripped)
	}
	if !strings.Contains(stripped, "Some text.") || !strings.Contains(stripped, "More text.") {
		t.Errorf("surrounding text should be preserved, got: %q", stripped)
	}
}

func TestStripAnnotations_FencedCodeBlock(t *testing.T) {
	content := "Some text\n```\n<!-- olw-auto: confidence=low -->\n```\nBut <!-- olw-auto: verified=true --> here."
	stripped := StripAnnotations(content)

	if !strings.Contains(stripped, "```\n<!-- olw-auto: confidence=low -->\n```") {
		t.Errorf("annotation inside a fenced block should be preserved, got: %q", stripped)
	}
	if strings.Contains(stripped, "verified=true") {
		t.Errorf("prose annotation should be removed, got: %q", stripped)
	}
}

func TestStripAnnotations_InlineCode(t *testing.T) {
	content := "Use `<!-- olw-auto: confidence=low -->` in a doc, but <!-- olw-auto: verified=true --> is real."
	stripped := StripAnnotations(content)

	if !strings.Contains(stripped, "`<!-- olw-auto: confidence=low -->`") {
		t.Errorf("annotation inside inline code should be preserved, got: %q", stripped)
	}
	if strings.Contains(stripped, "verified=true") {
		t.Errorf("prose annotation should be removed, got: %q", stripped)
	}
}

func TestStripAnnotations_WithoutAnnotations(t *testing.T) {
	if stripped := StripAnnotations(""); stripped != "" {
		t.Errorf("empty content = %q, want %q", stripped, "")
	}

	content := "Just prose, no annotations.\n\n```\nsome code\n```\n"
	if stripped := StripAnnotations(content); stripped != content {
		t.Errorf("content without annotations changed to %q, want %q", stripped, content)
	}
}
