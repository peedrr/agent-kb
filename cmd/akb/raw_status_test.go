package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestRawStatusJSONReportsEmptyDriftAsArray(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	out, err := rawStatusRun(t, kbRoot, "--json")
	if err != nil {
		t.Fatalf("akb raw status --json on a clean KB failed: %s: %v", out, err)
	}
	assertEmptyFilesArray(t, out)
}

func TestRawStatusJSONReportsFilteredDriftAsEmptyArray(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	// A raw file missing from the manifest drifts, but the path argument
	// filters the drifted list down to nothing.
	rawFile := filepath.Join(kbRoot, "raw", "data.csv")
	if err := os.WriteFile(rawFile, []byte("content\n"), 0600); err != nil { //nolint:gosec // test writing into its temp KB
		t.Fatal(err)
	}

	out, err := rawStatusRun(t, kbRoot, "other.csv", "--json")
	if err != nil {
		t.Fatalf("akb raw status other.csv --json failed: %s: %v", out, err)
	}
	assertEmptyFilesArray(t, out)
}
