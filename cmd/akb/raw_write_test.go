// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
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

// rawWriteRun drives the akb binary for one raw write and returns its combined
// output.
func rawWriteRun(t *testing.T, kbRoot, inputPath, content string, extraEnv ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(akbBinPath, "raw", "write", inputPath) //nolint:gosec // test helper launching akb binary
	cmd.Dir = kbRoot
	cmd.Stdin = strings.NewReader(content)
	cmd.Env = append(os.Environ(), extraEnv...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// gitInDir runs one git command in kbRoot and returns its trimmed output.
func gitInDir(t *testing.T, kbRoot string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper launching the git binary
	cmd.Dir = kbRoot
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// mustGitInDir runs one git command in kbRoot and fails the test when it fails.
func mustGitInDir(t *testing.T, kbRoot string, args ...string) string {
	t.Helper()

	out, err := gitInDir(t, kbRoot, args...)
	if err != nil {
		t.Fatalf("git %s in %s: %s: %v", strings.Join(args, " "), kbRoot, out, err)
	}
	return out
}

// commitFilesIn returns the paths recorded by HEAD.
func commitFilesIn(t *testing.T, kbRoot string) []string {
	t.Helper()

	var files []string
	for _, line := range strings.Split(mustGitInDir(t, kbRoot, "show", "--name-only", "--format=", "HEAD"), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files
}

// fallbackIdentityEnv keeps an ambient global git identity out of the akb
// process under test, so the raw write path resolves the akb fallback identity.
// Some git builds (nix) ignore GIT_CONFIG_GLOBAL and read a compiled-in global
// config; an empty user.name reaches the same branch, because the identity
// resolution treats an empty value as unset.
func fallbackIdentityEnv(t *testing.T) []string {
	t.Helper()

	return []string{
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=" + filepath.Join(t.TempDir(), "gitconfig"),
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_KEY_0=user.name",
		"GIT_CONFIG_VALUE_0=",
	}
}

func TestRawWriteCommitIsScopedToItsFiles(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	// A change the surrounding repository staged stays staged.
	foreign := filepath.Join(kbRoot, "src", "app.go")
	if err := os.MkdirAll(filepath.Dir(foreign), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(foreign, []byte("package main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	mustGitInDir(t, kbRoot, "add", "--", "src/app.go")

	out, err := rawWriteRun(t, kbRoot, "data.csv", "hello\n")
	if err != nil {
		t.Fatalf("akb raw write failed: %s: %v", out, err)
	}

	if files := commitFilesIn(t, kbRoot); !reflect.DeepEqual(files, []string{"raw/data.csv", "raw/files.log"}) {
		t.Errorf("commit recorded %v, want only the raw files", files)
	}
	status := mustGitInDir(t, kbRoot, "status", "--porcelain")
	if !strings.Contains(status, "A  src/app.go") {
		t.Errorf("staged change lost from the index:\n%s", status)
	}
	for _, line := range strings.Split(status, "\n") {
		if strings.Contains(line, "raw/") {
			t.Errorf("raw file left dirty after the commit: %q", line)
		}
	}
}

func TestRawWriteCommitUsesConfiguredIdentity(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	out, err := rawWriteRun(t, kbRoot, "data.csv", "hello\n")
	if err != nil {
		t.Fatalf("akb raw write failed: %s: %v", out, err)
	}

	if author := mustGitInDir(t, kbRoot, "log", "-1", "--format=%an <%ae>"); author != "akb-test <akb-test@local>" {
		t.Errorf("author = %q, want %q", author, "akb-test <akb-test@local>")
	}
}

func TestRawWriteCommitFallsBackToAKBIdentityWithoutConfiguredUser(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	mustGitInDir(t, kbRoot, "config", "--unset", "user.name")
	mustGitInDir(t, kbRoot, "config", "--unset", "user.email")

	out, err := rawWriteRun(t, kbRoot, "data.csv", "hello\n", fallbackIdentityEnv(t)...)
	if err != nil {
		t.Fatalf("akb raw write failed: %s: %v", out, err)
	}

	if author := mustGitInDir(t, kbRoot, "log", "-1", "--format=%an <%ae>"); author != "akb <akb@local>" {
		t.Errorf("author = %q, want %q", author, "akb <akb@local>")
	}

	for _, key := range []string{"user.name", "user.email"} {
		if value, err := gitInDir(t, kbRoot, "config", "--local", key); err == nil {
			t.Errorf("repository config %s = %q, want it unset", key, value)
		}
	}
}
