// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/template"
)

// nonGitTestKB creates a base that is not versioned in git: the layout
// writeSetupTestKB creates, with the config naming the none mode and no
// repository anywhere above it.
func nonGitTestKB(t *testing.T) string {
	t.Helper()

	kbRoot := t.TempDir()
	for _, dir := range []string{
		filepath.Join(kbRoot, "kb"),
		filepath.Join(kbRoot, "raw"),
		filepath.Join(kbRoot, ".agent-kb", "templates"),
	} {
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatalf("create dir %s: %v", dir, err)
		}
	}

	configContent := "name: nongit-test\ncreated: \"2024-01-01T00:00:00Z\"\nversioning: none\n"
	if err := os.WriteFile(filepath.Join(kbRoot, ".agent-kb", "akb.yaml"), []byte(configContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "index.md"), []byte("# Index\n\n"), 0600); err != nil {
		t.Fatalf("write index.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "log.md"), []byte("# Log\n\n"), 0600); err != nil {
		t.Fatalf("write log.md: %v", err)
	}
	if err := template.CopyDefaults(filepath.Join(kbRoot, ".agent-kb", "templates")); err != nil {
		t.Fatalf("copy templates: %v", err)
	}
	initTestSearchDB(t, kbRoot)

	useTestKBSelection(t, kbRoot)
	return kbRoot
}

// withoutGitOnPath makes the git binary unreachable for the duration of the
// test, so a command that invokes git in a base that is not versioned in git
// fails instead of silently shelling out.
func withoutGitOnPath(t *testing.T) {
	t.Helper()

	origPath, hadPath := os.LookupEnv("PATH")
	os.Setenv("PATH", t.TempDir()) //nolint:errcheck,gosec // test setup — failure is non-fatal
	t.Cleanup(func() {
		if hadPath {
			os.Setenv("PATH", origPath) //nolint:errcheck,gosec // test cleanup — failure is non-fatal
			return
		}
		os.Unsetenv("PATH") //nolint:errcheck,gosec // test cleanup — failure is non-fatal
	})
}

// withNoCommitFlag sets the base-wide --no-commit flag for one test.
func withNoCommitFlag(t *testing.T, value bool) {
	t.Helper()

	orig := noCommit
	noCommit = value
	t.Cleanup(func() { noCommit = orig })
}

// TestNonGitKBPageCommandsWriteWithoutGit drives the page commands of a base
// that is not versioned in git end-to-end: every one of them writes or deletes
// its files, no git invocation happens, and status reports the mode instead of
// a git status line.
func TestNonGitKBPageCommandsWriteWithoutGit(t *testing.T) {
	kbRoot := nonGitTestKB(t)
	withoutGitOnPath(t)
	withNoCommitFlag(t, false)
	withWriteFlags(t, nil, false)

	// write
	pointStdinAtTempFile(t, "---\ntype: note\ntitle: Non-git note\n---\nBody.\n")
	out, err := captureOutput(func() error { return runWrite(nil, []string{"notes/note.md"}) })
	if err != nil {
		t.Fatalf("write in a non-git base: %v", err)
	}
	if !strings.Contains(out, "Written to kb/notes/note.md") {
		t.Errorf("write output = %q, want the written page named", out)
	}
	notePath := filepath.Join(kbRoot, "kb", "notes", "note.md")

	// append
	pointStdinAtTempFile(t, "Appended body.\n")
	if _, err := captureOutput(func() error { return runAppend(nil, []string{"notes/note.md"}) }); err != nil {
		t.Fatalf("append in a non-git base: %v", err)
	}
	if data := readTestFile(t, notePath); !strings.Contains(data, "Appended body.") {
		t.Errorf("appended page does not carry the appended body:\n%s", data)
	}

	// index add
	if _, err := captureOutput(func() error { return runIndexAdd(nil, []string{"notes/note.md", "A non-git note"}) }); err != nil {
		t.Fatalf("index add in a non-git base: %v", err)
	}
	if data := readTestFile(t, filepath.Join(kbRoot, "kb", "index.md")); !strings.Contains(data, "kb/notes/note.md") {
		t.Errorf("index does not carry the new entry:\n%s", data)
	}

	// approve
	if _, err := captureOutput(func() error { return runApprove(nil, []string{"notes/note.md"}) }); err != nil {
		t.Fatalf("approve in a non-git base: %v", err)
	}
	if data := readTestFile(t, notePath); !strings.Contains(data, "is_draft: false") {
		t.Errorf("approved page does not carry is_draft: false:\n%s", data)
	}

	// delete
	if _, err := captureOutput(func() error { return runDeleteCmd(nil, []string{"notes/note.md"}) }); err != nil {
		t.Fatalf("delete in a non-git base: %v", err)
	}
	if _, err := os.Stat(notePath); !os.IsNotExist(err) {
		t.Errorf("page still on disk after delete: %v", err)
	}

	// status
	out, err = captureOutput(func() error { return runStatus(nil, nil) })
	if err != nil {
		t.Fatalf("status in a non-git base: %v", err)
	}
	if !strings.Contains(out, "Versioning: none") {
		t.Errorf("status output = %q, want the Versioning: none line", out)
	}
	if strings.Contains(out, "Git:") {
		t.Errorf("status output = %q, want no Git line in a non-git base", out)
	}
}

// TestNonGitKBRawAndTemplateCommandsWriteWithoutGit drives the raw and template
// commands of a base that is not versioned in git end-to-end: they write and
// remove their files, keep the manifest in step, and never invoke git.
func TestNonGitKBRawAndTemplateCommandsWriteWithoutGit(t *testing.T) {
	kbRoot := nonGitTestKB(t)
	withoutGitOnPath(t)
	withNoCommitFlag(t, false)

	// raw write
	pointStdinAtTempFile(t, "raw content")
	if _, err := captureOutput(func() error { return runRawWrite(nil, []string{"data/config.json"}) }); err != nil {
		t.Fatalf("raw write in a non-git base: %v", err)
	}
	rawPath := filepath.Join(kbRoot, "raw", "data", "config.json")
	if data := readTestFile(t, rawPath); data != "raw content" {
		t.Errorf("raw file content = %q, want %q", data, "raw content")
	}

	// raw delete
	if _, err := captureOutput(func() error { return runRawDelete(nil, []string{"data/config.json"}) }); err != nil {
		t.Fatalf("raw delete in a non-git base: %v", err)
	}
	if _, err := os.Stat(rawPath); !os.IsNotExist(err) {
		t.Errorf("raw file still on disk after delete: %v", err)
	}

	// raw sync: reconcile a file that was never written through akb.
	extraPath := filepath.Join(kbRoot, "raw", "extra.txt")
	if err := os.WriteFile(extraPath, []byte("extra"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := captureOutput(func() error { return runRawSync(nil, nil) }); err != nil {
		t.Fatalf("raw sync in a non-git base: %v", err)
	}
	entries, err := manifest.NewManager(kbRoot).ReadManifest()
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	synced := false
	for _, entry := range entries {
		if entry.Filename == "extra.txt" {
			synced = true
		}
	}
	if !synced {
		t.Errorf("manifest entries %+v do not carry the synced file", entries)
	}

	// template write: the template with a changed comment is a real swap.
	templatePath := filepath.Join(kbRoot, ".agent-kb", "templates", "note.yaml")
	yamlData := readTestFile(t, templatePath)
	editedPath := filepath.Join(t.TempDir(), "note.yaml")
	if err := os.WriteFile(editedPath, []byte(yamlData+"\n# edited by the non-git test\n"), 0600); err != nil {
		t.Fatal(err)
	}

	origTemplate, origPass, origFail, origForce := twTemplate, twPass, twFail, twForce
	twTemplate, twPass, twFail, twForce = editedPath, "", "", true
	t.Cleanup(func() { twTemplate, twPass, twFail, twForce = origTemplate, origPass, origFail, origForce })

	if _, err := captureOutput(func() error { return runTemplatesWrite(nil, []string{"note"}) }); err != nil {
		t.Fatalf("template write in a non-git base: %v", err)
	}
	if data := readTestFile(t, templatePath); !strings.Contains(data, "# edited by the non-git test") {
		t.Errorf("template file does not carry the rewrite:\n%s", data)
	}

	// template delete: an unused template is removed without a commit.
	origDeleteForce := tdForce
	tdForce = true
	t.Cleanup(func() { tdForce = origDeleteForce })

	if _, err := captureOutput(func() error { return runTemplateDelete(nil, []string{"adr"}) }); err != nil {
		t.Fatalf("template delete in a non-git base: %v", err)
	}
	for _, relPath := range []string{"adr.yaml", "adr_pass.md", "adr_fail.md"} {
		if _, err := os.Stat(filepath.Join(kbRoot, ".agent-kb", "templates", relPath)); !os.IsNotExist(err) {
			t.Errorf("template file %s still on disk after delete: %v", relPath, err)
		}
	}
}

// readTestFile returns the contents of a file of the test's own base.
func readTestFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // test reading its own temp file
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
