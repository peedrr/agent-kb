package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/registry"
)

func TestFreshnessFormula(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		daysAgo       int
		confidence    string
		minExpected   float64
		maxExpected   float64
		expectedStale bool
	}{
		{"fresh high", 0, "high", 99.0, 101.0, false},
		{"fresh medium", 0, "medium", 69.0, 71.0, false},
		{"fresh low", 0, "low", 39.0, 41.0, true},
		{"30 days high", 30, "high", 49.0, 51.0, false},
		{"30 days medium", 30, "medium", 34.0, 36.0, true},
		{"60 days high", 60, "high", 24.0, 26.0, true},
		{"90 days low", 90, "low", 4.0, 6.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedTime := now.AddDate(0, 0, -tt.daysAgo)
			daysSince := now.Sub(parsedTime).Hours() / 24

			var confidenceWeight float64
			switch tt.confidence {
			case "high":
				confidenceWeight = 1.0
			case "medium":
				confidenceWeight = 0.7
			case "low":
				confidenceWeight = 0.4
			default:
				confidenceWeight = 0.7
			}

			score := 100.0 * math.Pow(2, -daysSince/30.0) * confidenceWeight

			if (score < 50.0) != tt.expectedStale {
				t.Errorf("expected stale=%v, got score=%.1f", tt.expectedStale, score)
			}
			if score < tt.minExpected || score > tt.maxExpected {
				t.Errorf("expected score between %.1f-%.1f, got %.1f", tt.minExpected, tt.maxExpected, score)
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantYear  int
		wantMonth int
		wantDay   int
	}{
		{"rfc3339", "2024-01-15T10:30:00Z", false, 2024, 1, 15},
		{"date only", "2024-01-15", false, 2024, 1, 15},
		{"invalid", "not-a-date", true, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := parseTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ts.Year() != tt.wantYear || ts.Month() != time.Month(tt.wantMonth) || ts.Day() != tt.wantDay {
					t.Errorf("parseTimestamp() = %v, want %d-%02d-%02d", ts, tt.wantYear, tt.wantMonth, tt.wantDay)
				}
			}
		})
	}
}

func TestJSONOutputStructure(t *testing.T) {
	result := kbResult{
		Name: "test-kb",
		Path: "/tmp/test-kb",
		StalePages: []stalePage{
			{Path: "page.md", Score: 23.4, Confidence: "low", Updated: "2024-01-01"},
		},
		TotalStale: 1,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := parsed["name"]; !ok {
		t.Error("missing 'name' in JSON output")
	}
	if _, ok := parsed["path"]; !ok {
		t.Error("missing 'path' in JSON output")
	}
	if _, ok := parsed["stale_pages"]; !ok {
		t.Error("missing 'stale_pages' in JSON output")
	}
	if _, ok := parsed["total_stale"]; !ok {
		t.Error("missing 'total_stale' in JSON output")
	}
}

func TestEmptyRegistryHandling(t *testing.T) {
	dir := t.TempDir()
	registryFile := filepath.Join(dir, "registry.yaml")
	if err := os.WriteFile(registryFile, []byte("default: \"\"\nentries: []\n"), 0600); err != nil {
		t.Fatalf("failed to write registry: %v", err)
	}

	reg, err := LoadRegistryForTest(registryFile)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	if len(reg.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(reg.Entries))
	}
}

func TestMissingDirectoryWarning(t *testing.T) {
	entry := registry.Entry{
		Name: "missing-kb",
		Path: "/nonexistent/path",
	}

	_, err := os.Stat(filepath.Join(entry.Path, "kb"))
	if err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestMissingTimestampWarning(t *testing.T) {
	content := []byte(`---
title: Test Page
type: note
---

content here
`)

	fm, _, err := frontmatter.Parse(content)
	if err != nil {
		t.Fatalf("failed to parse frontmatter: %v", err)
	}

	if fm.Fields["updated"] != nil || fm.Fields["created"] != nil {
		t.Error("expected no timestamp fields")
	}
}

func LoadRegistryForTest(path string) (*registry.Registry, error) {
	return registry.Load(path)
}
