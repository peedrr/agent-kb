package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func appendSetupTestKB(t *testing.T) string {
	t.Helper()
	return writeSetupTestKB(t)
}

func appendCleanup(kbRoot string) {
	writeCleanup(kbRoot)
}

func appendRun(kbRoot, inputPath, stdinContent string, extraArgs ...string) (string, error) {
	args := append([]string{"append"}, extraArgs...)
	args = append(args, inputPath)
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, args...) //nolint:gosec // test helper launching akb binary
	cmd.Dir = kbRoot
	cmd.Stdin = strings.NewReader(stdinContent)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writePageForAppend(t *testing.T, kbRoot, filename, content string) {
	t.Helper()
	out, err := writeRun(kbRoot, filename, content)
	if err != nil {
		t.Fatalf("failed to write page for append test: %s: %v", out, err)
	}
}

func TestAppendSuccessful(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test Note\nsummary: A test\ntags: test\n---\nOriginal content."
	writePageForAppend(t, kbRoot, "my-note.md", content)

	appendContent := "Appended content."
	out, err := appendRun(kbRoot, "notes/my-note.md", appendContent)
	if err != nil {
		t.Fatalf("akb append failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "my-note.md")
	data, err := os.ReadFile(writtenPath) //nolint:gosec // test reading known temp file //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	fileStr := string(data)
	if !strings.Contains(fileStr, "type: note") {
		t.Errorf("expected frontmatter 'type: note' to be preserved, got: %s", fileStr)
	}
	if !strings.Contains(fileStr, "title: Test Note") {
		t.Errorf("expected frontmatter 'title: Test Note' to be preserved, got: %s", fileStr)
	}
	if !strings.Contains(fileStr, "Original content.") {
		t.Errorf("expected original body content, got: %s", fileStr)
	}
	if !strings.Contains(fileStr, "Appended content.") {
		t.Errorf("expected appended content, got: %s", fileStr)
	}

	if !strings.Contains(out, "Appended to kb/notes/my-note.md") {
		t.Errorf("expected output to contain 'Appended to kb/notes/my-note.md', got: %s", out)
	}
}

func TestAppendPreservesFrontmatter(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: FM Test\nsummary: A test\ntags: test\n---\nOriginal body."
	writePageForAppend(t, kbRoot, "fm-test.md", content)

	appendContent := "New content."
	out, err := appendRun(kbRoot, "notes/fm-test.md", appendContent)
	if err != nil {
		t.Fatalf("akb append failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "fm-test.md")
	data, err := os.ReadFile(writtenPath) //nolint:gosec // test reading known temp file //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	fileStr := string(data)
	// Verify frontmatter fields are exactly preserved
	if !strings.Contains(fileStr, "type: note") {
		t.Errorf("expected 'type: note' in frontmatter, got: %s", fileStr)
	}
	if !strings.Contains(fileStr, "title: FM Test") {
		t.Errorf("expected 'title: FM Test' in frontmatter, got: %s", fileStr)
	}
	if !strings.Contains(fileStr, "summary: A test") {
		t.Errorf("expected 'summary: A test' in frontmatter, got: %s", fileStr)
	}
	if !strings.Contains(fileStr, "tags: test") {
		t.Errorf("expected 'tags: test' in frontmatter, got: %s", fileStr)
	}
}

func TestAppendNonExistentPage(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	appendContent := "Some content."
	out, err := appendRun(kbRoot, "nonexistent.md", appendContent)
	if err == nil {
		t.Fatal("expected error for non-existent page, got nil")
	}
	if !strings.Contains(out, "page not found") {
		t.Errorf("expected error to contain 'page not found', got: %s", out)
	}
	if !strings.Contains(out, "akb write") {
		t.Errorf("expected error to suggest 'akb write', got: %s", out)
	}
}

func TestAppendRawPrefix(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	appendContent := "Some content."
	out, err := appendRun(kbRoot, "raw/test.txt", appendContent)
	if err == nil {
		t.Fatal("expected error for raw/ prefix, got nil")
	}
	if !strings.Contains(out, "use `akb raw write`") {
		t.Errorf("expected error to contain 'use `akb raw write`', got: %s", out)
	}
}

func TestAppendBodyContentNotRawAppend(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Body Test\nsummary: test\ntags: test\n---\nFirst line."
	writePageForAppend(t, kbRoot, "body-test.md", content)

	appendContent := "Second line."
	out, err := appendRun(kbRoot, "notes/body-test.md", appendContent)
	if err != nil {
		t.Fatalf("akb append failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "body-test.md")
	data, err := os.ReadFile(writtenPath) //nolint:gosec // test reading known temp file //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	fileStr := string(data)

	// The appended content should appear after the existing body content
	// with a newline separator, not as a raw file append
	bodyStart := strings.Index(fileStr, "First line.")
	if bodyStart == -1 {
		t.Fatalf("expected 'First line.' in body, got: %s", fileStr)
	}

	afterFirstLine := fileStr[bodyStart:]
	if !strings.Contains(afterFirstLine, "First line.\nSecond line.") {
		t.Errorf("expected body to have 'First line.\\nSecond line.', got: %q", afterFirstLine)
	}

	// Verify frontmatter is NOT duplicated or modified
	fmCount := strings.Count(fileStr, "type: note")
	if fmCount != 1 {
		t.Errorf("expected frontmatter 'type: note' to appear exactly once, got %d occurrences", fmCount)
	}
}

func TestAppendGitCommit(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Git Test\nsummary: test\ntags: test\n---\nOriginal."
	writePageForAppend(t, kbRoot, "git-test.md", content)

	appendContent := "Appended."
	out, err := appendRun(kbRoot, "notes/git-test.md", appendContent)
	if err != nil {
		t.Fatalf("akb append failed: %s: %v", out, err)
	}

	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "log", "--oneline", "-1")
	cmd.Dir = kbRoot
	gitOut, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if !strings.Contains(string(gitOut), "akb: append kb/notes/git-test.md") {
		t.Errorf("expected git commit message 'akb: append kb/notes/git-test.md', got: %s", string(gitOut))
	}
}

func TestAppendNoCommit(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: No Commit Test\nsummary: test\ntags: test\n---\nOriginal."
	writePageForAppend(t, kbRoot, "nocommit.md", content)

	appendContent := "Appended."
	out, err := appendRun(kbRoot, "notes/nocommit.md", appendContent, "--no-commit")
	if err != nil {
		t.Fatalf("akb append --no-commit failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "nocommit.md")
	data, err := os.ReadFile(writtenPath) //nolint:gosec // test reading known temp file //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if !strings.Contains(string(data), "Appended.") {
		t.Errorf("expected appended content in file, got: %s", string(data))
	}

	// Verify no git commit was made for the append
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "log", "--oneline", "-1", "--", "kb/notes/nocommit.md")
	cmd.Dir = kbRoot
	gitOut, _ := cmd.Output()
	// The last commit for this file should be the write, not the append
	if strings.Contains(string(gitOut), "append") {
		t.Errorf("expected no append commit for nocommit.md, but found: %s", string(gitOut))
	}
}

func TestAppendWithKBPrefix(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Prefix Test\nsummary: test\ntags: test\n---\nOriginal."
	writePageForAppend(t, kbRoot, "prefix-test.md", content)

	appendContent := "Appended."
	out, err := appendRun(kbRoot, "kb/notes/prefix-test.md", appendContent)
	if err != nil {
		t.Fatalf("akb append with kb/ prefix failed: %s: %v", out, err)
	}

	if !strings.Contains(out, "Appended to kb/notes/prefix-test.md") {
		t.Errorf("expected output 'Appended to kb/notes/prefix-test.md', got: %s", out)
	}
}

func TestAppendTTYStdin(t *testing.T) {
	// Test that isStdinTTY returns error when stdin is a TTY
	// We override the isStdinTTY function to simulate TTY
	origIsStdinTTY := isStdinTTY
	defer func() { isStdinTTY = origIsStdinTTY }()

	isStdinTTY = func() (bool, error) {
		return true, nil
	}

	err := runAppend(nil, []string{"test.md"})
	if err == nil {
		t.Fatal("expected error for TTY stdin, got nil")
	}
	if !strings.Contains(err.Error(), "input required: pipe content to stdin") {
		t.Errorf("expected error to contain 'input required: pipe content to stdin', got: %v", err)
	}
}

func TestAppendParentDirRejected(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	appendContent := "Some content."
	out, err := appendRun(kbRoot, "../escape.md", appendContent)
	if err == nil {
		t.Fatal("expected error for .. path, got nil")
	}
	if !strings.Contains(out, "..") {
		t.Errorf("expected error to contain '..', got: %s", out)
	}
}

func TestAppendAbsolutePathRejected(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	appendContent := "Some content."
	out, err := appendRun(kbRoot, "/tmp/evil.md", appendContent)
	if err == nil {
		t.Fatal("expected error for absolute path, got nil")
	}
	if !strings.Contains(out, "absolute") && !strings.Contains(out, "relative") {
		t.Errorf("expected error about absolute/relative path, got: %s", out)
	}
}

// TestConcurrentAppendsSerializeWithoutLostUpdates runs several akb append
// processes against the same page: their read-modify-write cycles serialize on
// the repository lock, so every append survives in the page and in its own
// commit.
func TestConcurrentAppendsSerializeWithoutLostUpdates(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	seed := "---\ntype: note\ntitle: Shared Note\nsummary: test\ntags: test\n---\nOriginal body."
	if out, err := writeRun(kbRoot, "shared.md", seed); err != nil {
		t.Fatalf("seed write failed: %s: %v", out, err)
	}

	const writers = 4
	cmds := make([]*exec.Cmd, 0, writers)
	outputs := make([]*strings.Builder, 0, writers)
	for i := range writers {
		cmd := exec.Command(akbBinPath, "append", "notes/shared.md") //nolint:gosec // test helper launching akb binary
		cmd.Dir = kbRoot
		cmd.Stdin = strings.NewReader(fmt.Sprintf("Appended by writer %d.", i))
		output := &strings.Builder{}
		cmd.Stdout = output
		cmd.Stderr = output
		if err := cmd.Start(); err != nil {
			t.Fatalf("start append %d: %v", i, err)
		}
		cmds = append(cmds, cmd)
		outputs = append(outputs, output)
	}

	waitErr := make(chan error, writers)
	for i, cmd := range cmds {
		go func(index int, cmd *exec.Cmd) {
			if err := cmd.Wait(); err != nil {
				waitErr <- fmt.Errorf("append process %d failed: %w\n%s", index, err, outputs[index].String())
				return
			}
			waitErr <- nil
		}(i, cmd)
	}
	deadline := time.After(90 * time.Second)
	for range writers {
		select {
		case err := <-waitErr:
			if err != nil {
				for _, cmd := range cmds {
					_ = cmd.Process.Kill() //nolint:errcheck // test cleanup
				}
				t.Fatal(err)
			}
		case <-deadline:
			for _, cmd := range cmds {
				_ = cmd.Process.Kill() //nolint:errcheck // test cleanup
			}
			t.Fatal("concurrent append processes did not finish in time")
		}
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "shared.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read shared page: %v", err)
	}
	body := string(data)
	for i := range writers {
		if !strings.Contains(body, fmt.Sprintf("Appended by writer %d.", i)) {
			t.Errorf("append of writer %d is missing from the page:\n%s", i, body)
		}
	}

	appends := 0
	for _, subject := range strings.Split(mustGitInDir(t, kbRoot, "log", "--format=%s"), "\n") {
		if strings.TrimSpace(subject) == "akb: append kb/notes/shared.md" {
			appends++
		}
	}
	if appends != writers {
		t.Errorf("append commits = %d, want %d — an append was lost", appends, writers)
	}
	for _, line := range strings.Split(mustGitInDir(t, kbRoot, "status", "--porcelain"), "\n") {
		if len(line) > 3 && strings.HasPrefix(line[3:], "kb/") {
			t.Errorf("page left dirty after the concurrent appends: %q", line)
		}
	}
}

func TestAppendBumpsUpdated(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	const stale = "2020-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/notes/append-updated.md",
		"---\ntype: note\ntitle: Append Updated\nsummary: test\ntags: test\ncreated: "+stale+"\nupdated: "+stale+"\n---\nOriginal body.")

	out, err := appendRun(kbRoot, "notes/append-updated.md", "Appended body.")
	if err != nil {
		t.Fatalf("akb append failed: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "append-updated.md"))
	updated := frontmatterString(t, fm, "updated")
	if updated == stale {
		t.Errorf("updated = %q, want the append to move the update time on", updated)
	}
	assertRecentTimestamp(t, updated)
	if got := frontmatterString(t, fm, "created"); got != stale {
		t.Errorf("created = %q, want %q preserved", got, stale)
	}
}

// datedHeading is the `## YYYY-MM-DD` heading a --dated append writes.
func datedHeading(t *testing.T) string {
	t.Helper()
	return "## " + time.Now().Format("2006-01-02")
}

func TestAppendDatedWrapsContentUnderHeading(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Dated Append\nsummary: test\ntags: test\n---\nOriginal body."
	writePageForAppend(t, kbRoot, "dated-append.md", content)

	out, err := appendRun(kbRoot, "notes/dated-append.md", "Dated addition.", "--dated")
	if err != nil {
		t.Fatalf("akb append --dated failed: %s: %v", out, err)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "dated-append.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	body := string(data)
	want := "Original body.\n" + datedHeading(t) + "\n\nDated addition."
	if !strings.Contains(body, want) {
		t.Errorf("expected the page to contain %q, got: %s", want, body)
	}
	if count := strings.Count(body, "type: note"); count != 1 {
		t.Errorf("frontmatter 'type: note' appears %d times, want 1", count)
	}
}

func TestWriteAppendDatedWrapsContentUnderHeading(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Dated Write Append\nsummary: test\ntags: test\n---\nOriginal body."
	writePageForAppend(t, kbRoot, "dated-write.md", content)

	out, err := writeRun(kbRoot, "notes/dated-write.md", "Dated addition.", "--append", "--dated")
	if err != nil {
		t.Fatalf("akb write --append --dated failed: %s: %v", out, err)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "dated-write.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	body := string(data)
	want := "Original body.\n" + datedHeading(t) + "\n\nDated addition."
	if !strings.Contains(body, want) {
		t.Errorf("expected the page to contain %q, got: %s", want, body)
	}
	if count := strings.Count(body, "type: note"); count != 1 {
		t.Errorf("frontmatter 'type: note' appears %d times, want 1", count)
	}
}

func TestWriteDatedRequiresAppend(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	out, err := writeRun(kbRoot, "notes/undated.md", "body without a dated append", "--dated")
	if err == nil {
		t.Fatalf("expected --dated without --append to fail, got: %s", out)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %T: %v", err, err)
	}
	if code := exitErr.ExitCode(); code != exitFault {
		t.Errorf("exit code = %d, want %d; output: %s", code, exitFault, out)
	}
	if !strings.Contains(out, "usage: --dated requires --append") {
		t.Errorf("expected the usage report for --dated without --append, got: %s", out)
	}
	if _, statErr := os.Stat(filepath.Join(kbRoot, "kb", "notes", "undated.md")); !os.IsNotExist(statErr) {
		t.Errorf("expected no page to be written, stat error = %v", statErr)
	}
}

// TestAppendLinkFailureRollsBackSearchIndex pins that the append path indexes
// the appended body and its links in one transaction: when the second step, the
// link-graph update, fails, the first step, the search-index write, does not
// persist and the search index keeps the pre-append body.
func TestAppendLinkFailureRollsBackSearchIndex(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	const relPath = "kb/notes/rollback-append.md"
	seed := "---\ntype: note\ntitle: Rollback Append\nsummary: test\ntags: test\n---\nOriginal body. See [[seed-target]]."
	if out, err := writeRun(kbRoot, "rollback-append.md", seed); err != nil {
		t.Fatalf("seed write failed: %s: %v", out, err)
	}

	preBody, ok := searchDBDocumentBody(t, kbRoot, relPath)
	if !ok {
		t.Fatalf("expected the seed page indexed at %s", relPath)
	}

	installSearchDBStatement(t, kbRoot, abortLinkInsertSQL)

	out, err := appendRun(kbRoot, "notes/rollback-append.md", "Appended body. See [[appended-target]].")
	if err == nil {
		t.Fatalf("expected akb append to fail when the link insert is aborted, got: %s", out)
	}
	if !strings.Contains(out, "update links:") {
		t.Fatalf("expected the link-graph step to fail, got: %s", out)
	}

	postBody, ok := searchDBDocumentBody(t, kbRoot, relPath)
	if !ok {
		t.Fatalf("expected %s to stay in the search index after the failed link step", relPath)
	}
	if postBody != preBody {
		t.Errorf("indexed body after the failed link step = %q, want the pre-append body %q — the search-index step did not roll back", postBody, preBody)
	}

	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM links WHERE source_page = ? AND raw_target = ?", relPath, "seed-target"); got != 1 {
		t.Errorf("seed link rows after the failed link step = %d, want 1", got)
	}
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM links WHERE source_page = ? AND raw_target = ?", relPath, "appended-target"); got != 0 {
		t.Errorf("appended link rows after the failed link step = %d, want 0", got)
	}
}

// TestAppendBareFilenameReachesTypeDir pins that akb append addresses a page
// stored under the directory of its type by its bare filename, and reports it
// by its stored path.
func TestAppendBareFilenameReachesTypeDir(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	const relPath = "kb/notes/bare-name.md"
	content := "---\ntype: note\ntitle: Bare Name\nsummary: test\ntags: test\n---\nOriginal body."
	writePageForAppend(t, kbRoot, "bare-name.md", content)

	out, err := appendRun(kbRoot, "bare-name.md", "Appended body.")
	if err != nil {
		t.Fatalf("akb append by bare filename failed: %s: %v", out, err)
	}
	if !strings.Contains(out, "Appended to "+relPath) {
		t.Errorf("output = %q, want the page reported by its stored path", out)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "bare-name.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if !strings.Contains(string(data), "Appended body.") {
		t.Errorf("page = %q, want the appended content", string(data))
	}
	if _, statErr := os.Stat(filepath.Join(kbRoot, "kb", "bare-name.md")); !os.IsNotExist(statErr) {
		t.Errorf("the append reached a second page at the KB root: %v", statErr)
	}

	indexed, ok := searchDBDocumentBody(t, kbRoot, relPath)
	if !ok {
		t.Fatalf("expected the appended page indexed at %s", relPath)
	}
	if !strings.Contains(indexed, "Appended body.") {
		t.Errorf("indexed body = %q, want the appended content", indexed)
	}
}

// TestAppendBareFilenameNotFound pins the not-found report for a page that is
// under no type directory.
func TestAppendBareFilenameNotFound(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	out, err := appendRun(kbRoot, "absent-note.md", "Some content.")
	if err == nil {
		t.Fatalf("expected the append of an absent page to fail, got: %s", out)
	}
	if want := "page not found: absent-note.md. Use `akb write` to create"; !strings.Contains(out, want) {
		t.Errorf("output = %q, want %q", out, want)
	}
}

// appendValidADR is a page the adr template accepts as written: it carries
// every required field and the Context, Decision and Consequences headings the
// template requires.
const appendValidADR = `---
type: adr
title: Append Validation ADR
summary: ADR for append validation
tags: test
status: proposed
deciders: team
created: '2020-01-01'
updated: '2020-01-01'
---
## Context

Context body.

## Decision

Decision body.

## Consequences

Consequences body.`

// TestAppendValidationFailureReportsAllRulesAndKeepsPage pins that an append
// runs the template's write-time rules against the page as the append leaves
// it: every failed rule is reported, the command exits with the validation
// exit code, and the page is left byte-identical.
func TestAppendValidationFailureReportsAllRulesAndKeepsPage(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	if out, err := writeRun(kbRoot, "append-validation.md", appendValidADR); err != nil {
		t.Fatalf("seed write failed: %s: %v", out, err)
	}

	pagePath := filepath.Join(kbRoot, "kb", "decisions", "append-validation.md")
	before, err := os.ReadFile(pagePath) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read page before the append: %v", err)
	}

	// Both headings are rejected by the adr template, so a report that carries
	// only the first failure would mean the rules are not all evaluated.
	out, err := appendRun(kbRoot, "decisions/append-validation.md",
		"## Options\n\nOption A.\n\n## Pros and Cons\n\nPros and cons.")
	if err == nil {
		t.Fatalf("expected the append to fail validation, got: %s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %T: %v", err, err)
	}
	if code := exitErr.ExitCode(); code != exitFailure {
		t.Errorf("exit code = %d, want %d; output: %s", code, exitFailure, out)
	}
	for _, want := range []string{
		"disallow_options",
		"page must not have a '## Options' heading",
		"disallow_pros_cons",
		"page must not have a '## Pros and Cons' heading",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want it to report %q", out, want)
		}
	}
	for _, unwanted := range []string{"internal:", "CEL engine error"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("output = %q, want a validation failure rather than %q", out, unwanted)
		}
	}

	after, err := os.ReadFile(pagePath) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read page after the failed append: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("page changed by a failed append:\nbefore: %q\nafter:  %q", before, after)
	}
}

// TestAppendTypelessPageFailsValidation pins that a page without a type is no
// longer appended to: the append reports the missing type and leaves the page
// byte-identical.
func TestAppendTypelessPageFailsValidation(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	const relPath = "kb/notes/typeless.md"
	const typeless = "---\ntitle: Typeless Page\n---\nOriginal body."
	writeRawPage(t, kbRoot, relPath, typeless)

	out, err := appendRun(kbRoot, "notes/typeless.md", "Appended body.")
	if err == nil {
		t.Fatalf("expected the append of a typeless page to fail, got: %s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %T: %v", err, err)
	}
	if code := exitErr.ExitCode(); code != exitFailure {
		t.Errorf("exit code = %d, want %d; output: %s", code, exitFailure, out)
	}
	if want := "validate type: missing required field 'type' in frontmatter"; !strings.Contains(out, want) {
		t.Errorf("output = %q, want %q", out, want)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read page after the failed append: %v", err)
	}
	if string(data) != typeless {
		t.Errorf("page = %q, want it left as %q", data, typeless)
	}
}

// TestAppendSuccessOutputLineStable pins the success line of a valid append to
// the exact text the command has always printed on stdout.
func TestAppendSuccessOutputLineStable(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	writePageForAppend(t, kbRoot, "output-line.md",
		"---\ntype: note\ntitle: Output Line\nsummary: test\ntags: test\n---\nOriginal body.")

	cmd := exec.Command(akbBinPath, "append", "notes/output-line.md") //nolint:gosec // test helper launching akb binary
	cmd.Dir = kbRoot
	cmd.Stdin = strings.NewReader("Appended body.")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("akb append failed: %v; stderr: %s", err, stderr.String())
	}
	if want := "Appended to kb/notes/output-line.md\n"; stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}

// TestAppendMissingRequiredFieldFailsValidation pins that an append to a page
// that leaves out a schema-required field is a validation failure: exit 1, the
// missing field named, and the page left byte-identical.
func TestAppendMissingRequiredFieldFailsValidation(t *testing.T) {
	kbRoot := appendSetupTestKB(t)
	defer appendCleanup(kbRoot)

	const relPath = "kb/decisions/planted.md"
	writeRawPage(t, kbRoot, relPath, adrMissingStatus)

	out, err := appendRun(kbRoot, "decisions/planted.md", "Appended body.")
	if err == nil {
		t.Fatalf("expected the append to fail, got: %s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %T: %v", err, err)
	}
	if code := exitErr.ExitCode(); code != exitFailure {
		t.Errorf("exit code = %d, want %d; output: %s", code, exitFailure, out)
	}
	if want := `missing required frontmatter field(s): status (declared required by template "adr")`; !strings.Contains(out, want) {
		t.Errorf("output = %q, want %q", out, want)
	}

	data, readErr := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading known temp file
	if readErr != nil {
		t.Fatalf("read page after the failed append: %v", readErr)
	}
	if string(data) != adrMissingStatus {
		t.Errorf("page changed by the failed append:\nbefore: %q\nafter:  %q", adrMissingStatus, data)
	}
}
