package main

import (
	"os/exec"
	"strings"
	"testing"
)

func rawStatusRun(t *testing.T, kbRoot string, args ...string) (string, error) {
	t.Helper()

	cmdArgs := append([]string{"raw", "status"}, args...)
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, cmdArgs...)
	cmd.Dir = kbRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// assertEmptyFilesArray checks the JSON envelope renders an empty file list as
// an array rather than null.
func assertEmptyFilesArray(t *testing.T, out string) {
	t.Helper()

	if !strings.Contains(out, `"files": []`) {
		t.Errorf("expected an empty files array, got: %s", out)
	}
	if strings.Contains(out, "null") {
		t.Errorf("expected no null in the JSON envelope, got: %s", out)
	}
}

// writeTrackedRaw records one raw file in the manifest through akb itself, so
// the file starts out undrifted.
func writeTrackedRaw(t *testing.T, kbRoot, inputPath, content string) {
	t.Helper()

	out, err := rawWriteRun(t, kbRoot, inputPath, content)
	if err != nil {
		t.Fatalf("akb raw write %s failed: %s: %v", inputPath, out, err)
	}
}

func TestRawStatusJSONReportsEmptyDriftAsArray(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	out, err := rawStatusRun(t, kbRoot, "--json")
	if err != nil {
		t.Fatalf("akb raw status --json on a clean KB failed: %s: %v", out, err)
	}
	assertEmptyFilesArray(t, out)
}

// TestRawStatusJSONReportsUndriftedFilterAsEmptyArray pins the filtered JSON
// envelope for a path that names a tracked raw file: the file exists, so the
// filter is valid, and it has not drifted, so the file list stays an array.
func TestRawStatusJSONReportsUndriftedFilterAsEmptyArray(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	writeTrackedRaw(t, kbRoot, "data.csv", "content\n")

	out, err := rawStatusRun(t, kbRoot, "data.csv", "--json")
	if err != nil {
		t.Fatalf("akb raw status data.csv --json failed: %s: %v", out, err)
	}
	assertEmptyFilesArray(t, out)
}
