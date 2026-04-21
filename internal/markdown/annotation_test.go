package markdown

import (
	"reflect"
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
