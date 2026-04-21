package markdown

import (
	"reflect"
	"testing"
)

func TestParseProvenanceMarkers_Simple(t *testing.T) {
	content := "text ^[inferred] more"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}
	if markers[0].Type != "inferred" {
		t.Errorf("expected type 'inferred', got '%s'", markers[0].Type)
	}
}

func TestParseProvenanceMarkers_MultipleTypes(t *testing.T) {
	content := "^[inferred] and ^[ambiguous]"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
	if markers[0].Type != "inferred" {
		t.Errorf("first marker: expected 'inferred', got '%s'", markers[0].Type)
	}
	if markers[1].Type != "ambiguous" {
		t.Errorf("second marker: expected 'ambiguous', got '%s'", markers[1].Type)
	}
}

func TestParseProvenanceMarkers_FencedCodeBlock(t *testing.T) {
	content := "```\n^[inferred]\n```"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 0 {
		t.Fatalf("expected 0 markers in fenced code, got %d", len(markers))
	}
}

func TestParseProvenanceMarkers_InlineCode(t *testing.T) {
	content := "`^[inferred]`"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 0 {
		t.Fatalf("expected 0 markers in inline code, got %d", len(markers))
	}
}

func TestParseProvenanceMarkers_HTMLComment(t *testing.T) {
	content := "<!-- ^[inferred] -->"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 0 {
		t.Fatalf("expected 0 markers in HTML comment, got %d", len(markers))
	}
}

func TestParseProvenanceMarkers_EmptyContent(t *testing.T) {
	markers := ParseProvenanceMarkers("")

	if len(markers) != 0 {
		t.Fatalf("expected 0 markers for empty content, got %d", len(markers))
	}
}

func TestParseProvenanceMarkers_UnknownType(t *testing.T) {
	content := "^[unknown]"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}
	if markers[0].Type != "unknown" {
		t.Errorf("expected type 'unknown', got '%s'", markers[0].Type)
	}
}

func TestParseProvenanceMarkers_Adjacent(t *testing.T) {
	content := "^[inferred]^[ambiguous]"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
	if markers[0].Type != "inferred" || markers[1].Type != "ambiguous" {
		t.Errorf("marker types not as expected")
	}
}

func TestCountMarkersByType(t *testing.T) {
	markers := []ProvenanceMarker{
		{Type: "inferred", Position: 0},
		{Type: "inferred", Position: 10},
		{Type: "ambiguous", Position: 20},
		{Type: "extracted", Position: 30},
	}

	counts := CountMarkersByType(markers)

	expected := map[string]int{
		"inferred":   2,
		"ambiguous": 1,
		"extracted":  1,
	}

	if !reflect.DeepEqual(counts, expected) {
		t.Errorf("expected %v, got %v", expected, counts)
	}
}

func TestParseProvenanceMarkers_MixedExclusions(t *testing.T) {
	// Marker outside code block (should be found), marker inside (excluded)
	content := "text ^[extracted] text\n```\n^[inferred]\n```\ntext"
	markers := ParseProvenanceMarkers(content)

	if len(markers) != 1 {
		t.Fatalf("expected 1 marker (only outside code block), got %d", len(markers))
	}
	if markers[0].Type != "extracted" {
		t.Errorf("expected 'extracted', got '%s'", markers[0].Type)
	}
}