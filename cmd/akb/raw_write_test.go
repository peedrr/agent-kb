package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/peedrr/agent-kb/internal/manifest"
)

// pointStdinAtTempFile makes os.Stdin readable with the given content for the
// duration of the test.
func pointStdinAtTempFile(t *testing.T, content string) {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	old := os.Stdin
	os.Stdin = f
	t.Cleanup(func() {
		os.Stdin = old
		_ = f.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	})
}

func TestRunRawWrite_RepeatedWriteKeepsSingleManifestEntry(t *testing.T) {
	kbRoot := setupLinksTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	const rawPath = "data/config.json"
	fullPath := filepath.Join(kbRoot, "raw", rawPath)
	mgr := manifest.NewManager(kbRoot)

	pointStdinAtTempFile(t, "first version")
	if _, err := captureOutput(func() error { return runRawWrite(nil, []string{rawPath}) }); err != nil {
		t.Fatalf("first raw write: %v", err)
	}

	entries, err := mgr.ReadManifest()
	if err != nil {
		t.Fatalf("ReadManifest after first write: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 manifest entry after first write, got %d: %+v", len(entries), entries)
	}
	firstHash, err := manifest.ComputeSHA256(fullPath)
	if err != nil {
		t.Fatalf("ComputeSHA256 after first write: %v", err)
	}
	if entries[0].SHA256 != firstHash {
		t.Errorf("entry SHA256 = %q, want %q", entries[0].SHA256, firstHash)
	}

	// Age the entry so the second write's timestamp refresh is observable even
	// when both writes land within the same RFC3339 second.
	aged := entries[0]
	aged.LastUpdated = "2000-01-01T00:00:00Z"
	if err := mgr.WriteManifest([]manifest.Entry{aged}); err != nil {
		t.Fatalf("age manifest entry: %v", err)
	}

	pointStdinAtTempFile(t, "second version")
	if _, err := captureOutput(func() error { return runRawWrite(nil, []string{rawPath}) }); err != nil {
		t.Fatalf("second raw write: %v", err)
	}

	entries, err = mgr.ReadManifest()
	if err != nil {
		t.Fatalf("ReadManifest after second write: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 manifest entry after repeated write, got %d: %+v", len(entries), entries)
	}

	secondHash, err := manifest.ComputeSHA256(fullPath)
	if err != nil {
		t.Fatalf("ComputeSHA256 after second write: %v", err)
	}
	if entries[0].SHA256 != secondHash {
		t.Errorf("entry SHA256 = %q, want %q (hash of rewritten file)", entries[0].SHA256, secondHash)
	}
	if entries[0].SHA256 == firstHash {
		t.Errorf("entry SHA256 still %q after rewrite with different content", firstHash)
	}

	timestamp, err := time.Parse(time.RFC3339, entries[0].LastUpdated)
	if err != nil {
		t.Fatalf("entry LastUpdated %q is not RFC3339: %v", entries[0].LastUpdated, err)
	}
	if !timestamp.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("entry LastUpdated = %q, want refreshed timestamp after second write", entries[0].LastUpdated)
	}
}
