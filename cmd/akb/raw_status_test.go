package main

import (
	"errors"
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

// rawStatusExitCode returns the exit code of one akb raw status invocation.
func rawStatusExitCode(t *testing.T, out string, err error) int {
	t.Helper()

	if err == nil {
		return exitSuccess
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("akb raw status did not exit normally: %s: %v", out, err)
	}
	return exitErr.ExitCode()
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

// writeRawFile writes a file under raw/ directly, bypassing akb so the manifest
// stays untouched.
func writeRawFile(t *testing.T, kbRoot, relPath, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(kbRoot, "raw", relPath), []byte(content), 0600); err != nil {
		t.Fatal(err)
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

// TestRawStatusNonexistentPathArgumentExitsTwo pins the fix for a path argument
// that names no raw file: it is an invocation mistake reported as a usage
// error, not a healthy result.
func TestRawStatusNonexistentPathArgumentExitsTwo(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	writeTrackedRaw(t, kbRoot, "data.csv", "content\n")

	out, err := rawStatusRun(t, kbRoot, "missing.csv")
	if code := rawStatusExitCode(t, out, err); code != exitFault {
		t.Errorf("exit code = %d, want %d: %s", code, exitFault, out)
	}

	const wantReport = "usage: no raw file or manifest entry matches missing.csv"
	if !strings.Contains(out, wantReport) {
		t.Errorf("output = %q, want it to contain %q", out, wantReport)
	}
	if strings.Contains(out, "no drift") {
		t.Errorf("output = %q, want no healthy report for a path that covers no raw file", out)
	}

	// The same invocation in process returns the error main classifies.
	assertUsageFailure(t, runRawStatus(nil, []string{"missing.csv"}), "no raw file or manifest entry matches missing.csv")
}

// TestRawStatusUndriftedPathArgumentExitsZero keeps the healthy contract: a
// tracked file whose content still matches the manifest reports exit 0.
func TestRawStatusUndriftedPathArgumentExitsZero(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	writeTrackedRaw(t, kbRoot, "data.csv", "content\n")

	out, err := rawStatusRun(t, kbRoot, "data.csv")
	if code := rawStatusExitCode(t, out, err); code != exitSuccess {
		t.Errorf("exit code = %d, want %d: %s", code, exitSuccess, out)
	}
	if !strings.Contains(out, "no drift") {
		t.Errorf("output = %q, want a no drift report", out)
	}
}

// TestRawStatusDriftedPathArgumentExitsOne keeps the drift contract for every
// way a path argument can name a drifted raw file.
func TestRawStatusDriftedPathArgumentExitsOne(t *testing.T) {
	cases := []struct {
		name     string
		prepare  func(t *testing.T, kbRoot string)
		wantText string
	}{
		{
			name: "modified tracked file",
			prepare: func(t *testing.T, kbRoot string) {
				t.Helper()
				writeTrackedRaw(t, kbRoot, "data.csv", "content\n")
				writeRawFile(t, kbRoot, "data.csv", "changed\n")
			},
			wantText: "MODIFIED data.csv",
		},
		{
			name: "untracked file on disk",
			prepare: func(t *testing.T, kbRoot string) {
				t.Helper()
				writeRawFile(t, kbRoot, "data.csv", "loose\n")
			},
			wantText: "UNTRACKED data.csv",
		},
		{
			name: "tracked file removed from disk",
			prepare: func(t *testing.T, kbRoot string) {
				t.Helper()
				writeTrackedRaw(t, kbRoot, "data.csv", "content\n")
				if err := os.Remove(filepath.Join(kbRoot, "raw", "data.csv")); err != nil {
					t.Fatal(err)
				}
			},
			wantText: "MISSING data.csv",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kbRoot := writeSetupTestKB(t)
			defer writeCleanup(kbRoot)

			tc.prepare(t, kbRoot)

			out, err := rawStatusRun(t, kbRoot, "data.csv")
			if code := rawStatusExitCode(t, out, err); code != exitFailure {
				t.Errorf("exit code = %d, want %d: %s", code, exitFailure, out)
			}
			if !strings.Contains(out, tc.wantText) {
				t.Errorf("output = %q, want it to contain %q", out, tc.wantText)
			}
		})
	}
}
