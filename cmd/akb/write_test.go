package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/template"
)

func writeSetupTestKB(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "akb-write-test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	kbRoot := tmpDir

	dirs := []string{
		filepath.Join(kbRoot, "kb"),
		filepath.Join(kbRoot, "raw"),
		filepath.Join(kbRoot, ".akb", "templates"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on setup failure
			t.Fatalf("create dir %s: %v", dir, err)
		}
	}

	configContent := "name: write-test\ncreated: \"2024-01-01T00:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", ".akb.yaml"), []byte(configContent), 0600); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on setup failure
		t.Fatalf("write config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "index.md"), []byte("# Index\n\n"), 0600); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on setup failure
		t.Fatalf("write index.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "log.md"), []byte("# Log\n\n"), 0600); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on setup failure
		t.Fatalf("write log.md: %v", err)
	}

	if err := template.CopyDefaults(filepath.Join(kbRoot, ".akb", "templates")); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on setup failure
		t.Fatalf("copy templates: %v", err)
	}

	initTestSearchDB(t, kbRoot)

	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "init")
	cmd.Dir = kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on setup failure
		t.Fatalf("git init: %s: %v", strings.TrimSpace(string(out)), err)
	}

	cmd = exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "config", "user.name", "akb-test")
	cmd.Dir = kbRoot
	_ = cmd.Run() //nolint:errcheck,gosec // best-effort git setup in test helper

	cmd = exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "config", "user.email", "akb-test@local")
	cmd.Dir = kbRoot
	_ = cmd.Run() //nolint:errcheck,gosec // best-effort git setup in test helper

	cmd = exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "add", "-A")
	cmd.Dir = kbRoot
	_ = cmd.Run() //nolint:errcheck,gosec // best-effort git setup in test helper

	cmd = exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "commit", "-m", "init test kb")
	cmd.Dir = kbRoot
	_ = cmd.Run() //nolint:errcheck,gosec // best-effort git setup in test helper

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", kbRoot)                         //nolint:errcheck,gosec // test setup — failure is non-fatal
	t.Cleanup(func() { os.Setenv("HOME", origHome) }) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	useTestKBSelection(t, kbRoot)

	return kbRoot
}

func writeCleanup(kbRoot string) {
	_ = os.RemoveAll(kbRoot) //nolint:errcheck,gosec // test cleanup — failure is non-fatal
}

func initTestSearchDB(t *testing.T, kbRoot string) {
	t.Helper()
	dbPath := filepath.Join(kbRoot, ".akb", "search.db")
	conn, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init test search DB: %v", err)
	}
	defer conn.Close() //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatalf("enable WAL mode: %v", err)
	}
	conn.SetMaxOpenConns(1)

	if err := db.CreateSchema(conn); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
}

func writeRun(kbRoot, inputPath, stdinContent string, extraArgs ...string) (string, error) {
	args := append([]string{"write"}, extraArgs...)
	args = append(args, inputPath)
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, args...)
	cmd.Dir = kbRoot
	cmd.Stdin = strings.NewReader(stdinContent)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// writeRawPage places a page on disk without going through akb write, so a test
// controls the frontmatter the command reads and rewrites.
func writeRawPage(t *testing.T, kbRoot, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(kbRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
		t.Fatalf("create page directory: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
		t.Fatalf("write raw page: %v", err)
	}
}

// pageFrontmatter parses the frontmatter of a written page.
func pageFrontmatter(t *testing.T, path string) *frontmatter.ParsedFrontmatter {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read page %s: %v", path, err)
	}
	fm, _, err := frontmatter.Parse(data)
	if err != nil {
		t.Fatalf("parse page %s: %v", path, err)
	}
	return fm
}

// frontmatterString returns one frontmatter field of a parsed page as a string.
func frontmatterString(t *testing.T, fm *frontmatter.ParsedFrontmatter, key string) string {
	t.Helper()
	raw, ok := fm.Fields[key]
	if !ok {
		t.Fatalf("frontmatter field %q is missing", key)
	}
	value, ok := raw.(string)
	if !ok {
		t.Fatalf("frontmatter field %q = %T (%v), want string", key, raw, raw)
	}
	return value
}

// assertRecentTimestamp fails unless value is an RFC3339 UTC stamp from now.
func assertRecentTimestamp(t *testing.T, value string) {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("timestamp %q is not RFC3339: %v", value, err)
	}
	if !strings.HasSuffix(value, "Z") {
		t.Errorf("timestamp %q does not carry the UTC Z suffix", value)
	}
	if delta := time.Since(parsed); delta < 0 || delta > 2*time.Minute {
		t.Errorf("timestamp %q is not current (delta %v)", value, delta)
	}
}

func TestWriteNoteToTypeDir(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test Note\nsummary: A test\ntags: test\n---\nContent here."
	out, err := writeRun(kbRoot, "my-note.md", content)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "my-note.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}

	data, err := os.ReadFile(writtenPath) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if !strings.Contains(string(data), "type: note") {
		t.Errorf("expected file to contain 'type: note', got: %s", string(data))
	}
	if !strings.Contains(string(data), "Content here.") {
		t.Errorf("expected file to contain body content, got: %s", string(data))
	}

	if !strings.Contains(out, "Written to kb/notes/my-note.md") {
		t.Errorf("expected output to contain 'Written to kb/notes/my-note.md', got: %s", out)
	}
	if !strings.Contains(out, "akb index add") {
		t.Errorf("expected output to contain index hint, got: %s", out)
	}
}

func TestWriteADRToDecisionsDir(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: adr\ntitle: Use Go\nsummary: We chose Go\ntags: decision\nstatus: accepted\ndeciders: team\ncreated: 2025-01-01\nupdated: 2025-01-01\n---\n## Context\n\nWe needed a language.\n\n## Decision\n\nWe decided to use Go.\n\n## Consequences\n\nEverything works better."
	out, err := writeRun(kbRoot, "my-adr.md", content)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "decisions", "my-adr.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}

	if !strings.Contains(out, "Written to kb/decisions/my-adr.md") {
		t.Errorf("expected output to contain 'Written to kb/decisions/my-adr.md', got: %s", out)
	}
}

func TestWriteUnknownType(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: recipe\ntitle: Cake\n---\nContent."
	out, err := writeRun(kbRoot, "cake.md", content)
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
	if !strings.Contains(out, "unknown type 'recipe'") {
		t.Errorf("expected error to contain \"unknown type 'recipe'\", got: %s", out)
	}
	if !strings.Contains(out, "Create .akb/templates/recipe.yaml") {
		t.Errorf("expected error to contain template suggestion, got: %s", out)
	}
}

func TestWriteMissingType(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntitle: No Type\n---\nContent."
	out, err := writeRun(kbRoot, "nope.md", content)
	if err == nil {
		t.Fatal("expected error for missing type, got nil")
	}
	if !strings.Contains(out, "missing required field 'type'") {
		t.Errorf("expected error to contain \"missing required field 'type'\", got: %s", out)
	}
}

func TestWriteMissingTitle(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\nsummary: No title\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "nope2.md", content)
	if err == nil {
		t.Fatal("expected error for missing title, got nil")
	}
	if !strings.Contains(out, "missing required field 'title'") {
		t.Errorf("expected error to contain \"missing required field 'title'\", got: %s", out)
	}
}

func TestWriteUnknownFieldAllowed(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Bad Field\nsummary: test\ntags: test\nconfidence: high\n---\nContent."
	out, err := writeRun(kbRoot, "bad-field.md", content)
	if err != nil {
		t.Fatalf("expected success for unknown field with CEL validation, got: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "bad-field.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}
}

func TestWriteIsDraftAllowed(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\nis_draft: true\n---\nContent."
	out, err := writeRun(kbRoot, "draft-note.md", content)
	if err != nil {
		t.Fatalf("akb write with is_draft failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "draft-note.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}
}

func TestWriteRawPrefixRejected(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "raw/test.md", content)
	if err == nil {
		t.Fatal("expected error for raw/ prefix, got nil")
	}
	if !strings.Contains(out, "use `akb raw write`") {
		t.Errorf("expected error to contain \"use `akb raw write`\", got: %s", out)
	}
}

func TestWriteNoMDExtension(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "my-note", content)
	if err == nil {
		t.Fatal("expected error for missing .md extension, got nil")
	}
	if !strings.Contains(out, "must end with .md") {
		t.Errorf("expected error to contain 'must end with .md', got: %s", out)
	}
}

func TestWriteStripsKBPrefix(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "kb/my-note.md", content)
	if err != nil {
		t.Fatalf("akb write with kb/ prefix failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "my-note.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}
}

func TestWriteStripsTypeDirPrefix(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "notes/my-note.md", content)
	if err != nil {
		t.Fatalf("akb write with notes/ prefix failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "my-note.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}
}

func TestWriteNoCommit(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: No Commit Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "nocommit.md", content, "--no-commit")
	if err != nil {
		t.Fatalf("akb write --no-commit failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "nocommit.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}

	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "log", "--oneline", "-1", "--", "kb/notes/nocommit.md")
	cmd.Dir = kbRoot
	gitOut, _ := cmd.Output()
	if strings.Contains(string(gitOut), "write") {
		t.Errorf("expected no git commit for nocommit.md, but found: %s", string(gitOut))
	}

	cmd = exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "status", "--porcelain")
	cmd.Dir = kbRoot
	statusOut, _ := cmd.Output()
	if !strings.Contains(string(statusOut), "nocommit.md") {
		t.Errorf("expected uncommitted file in git status, got: %s", string(statusOut))
	}
}

func TestWriteGitCommit(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Git Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "git-test.md", content)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}

	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "log", "--oneline", "-1")
	cmd.Dir = kbRoot
	gitOut, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if !strings.Contains(string(gitOut), "akb: write kb/notes/git-test.md") {
		t.Errorf("expected git commit message 'akb: write kb/notes/git-test.md', got: %s", string(gitOut))
	}
}

func TestWriteOverwrite(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content1 := "---\ntype: note\ntitle: Original\nsummary: first\ntags: test\n---\nOriginal content."
	out, err := writeRun(kbRoot, "overwrite.md", content1)
	if err != nil {
		t.Fatalf("first akb write failed: %s: %v", out, err)
	}

	content2 := "---\ntype: note\ntitle: Updated\nsummary: second\ntags: test\n---\nUpdated content."
	out, err = writeRun(kbRoot, "overwrite.md", content2)
	if err != nil {
		t.Fatalf("second akb write (overwrite) failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "overwrite.md")
	data, err := os.ReadFile(writtenPath) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if !strings.Contains(string(data), "Updated content.") {
		t.Errorf("expected overwritten content, got: %s", string(data))
	}
}

func TestWriteParentDirRejected(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "../escape.md", content)
	if err == nil {
		t.Fatal("expected error for .. path, got nil")
	}
	if !strings.Contains(out, "..") {
		t.Errorf("expected error to contain '..', got: %s", out)
	}
}

func TestWriteAbsolutePathRejected(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "/tmp/evil.md", content)
	if err == nil {
		t.Fatal("expected error for absolute path, got nil")
	}
	if !strings.Contains(out, "absolute") && !strings.Contains(out, "relative") {
		t.Errorf("expected error about absolute/relative path, got: %s", out)
	}
}

func TestWriteNoFrontmatter(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	out, err := writeRun(kbRoot, "nofm.md", "Just plain text without frontmatter")
	if err == nil {
		t.Fatal("expected error for missing frontmatter, got nil")
	}
	if !strings.Contains(out, "frontmatter") {
		t.Errorf("expected error about frontmatter, got: %s", out)
	}
}

func TestWriteEmptyFrontmatter(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	out, err := writeRun(kbRoot, "emptyfm.md", "---\n---\nContent.")
	if err == nil {
		t.Fatal("expected error for empty frontmatter, got nil")
	}
	if !strings.Contains(out, "empty frontmatter") {
		t.Errorf("expected error about empty frontmatter, got: %s", out)
	}
}

func TestWriteKBPrefixWithTypeDirPrefix(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "kb/notes/my-note.md", content)
	if err != nil {
		t.Fatalf("akb write with kb/notes/ prefix failed: %s: %v", out, err)
	}

	writtenPath := filepath.Join(kbRoot, "kb", "notes", "my-note.md")
	if _, err := os.Stat(writtenPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", writtenPath)
	}

	if !strings.Contains(out, "Written to kb/notes/my-note.md") {
		t.Errorf("expected output 'Written to kb/notes/my-note.md', got: %s", out)
	}
}

func TestWriteIndexHint(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Hint Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "hint-test.md", content)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}

	expected := "akb index add kb/notes/hint-test.md <summary>"
	if !strings.Contains(out, expected) {
		t.Errorf("expected index hint %q, got: %s", expected, out)
	}
}

func TestWriteGitCommitMessageADR(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: adr\ntitle: Use Go\nsummary: We chose Go\ntags: decision\nstatus: accepted\ndeciders: team\ncreated: 2025-01-01\nupdated: 2025-01-01\n---\n## Context\n\nWe needed a language.\n\n## Decision\n\nWe decided to use Go.\n\n## Consequences\n\nEverything works better."
	out, err := writeRun(kbRoot, "my-adr.md", content)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}

	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "log", "--oneline", "-1")
	cmd.Dir = kbRoot
	gitOut, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if !strings.Contains(string(gitOut), "akb: write kb/decisions/my-adr.md") {
		t.Errorf("expected commit message 'akb: write kb/decisions/my-adr.md', got: %s", string(gitOut))
	}
}

// TestConcurrentWriteAppendsSerializeWithoutLostUpdates runs several
// `akb write --append` processes against the same page: each process reads the
// page and assembles its new content under the repository lock, so every append
// survives in the page and in its own commit.
func TestConcurrentWriteAppendsSerializeWithoutLostUpdates(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const relPath = "kb/notes/shared-append.md"
	const commitSubject = "akb: write " + relPath

	seed := "---\ntype: note\ntitle: Shared Append Note\nsummary: test\ntags: test\n---\nOriginal body."
	if out, err := writeRun(kbRoot, "shared-append.md", seed); err != nil {
		t.Fatalf("seed write failed: %s: %v", out, err)
	}

	const writers = 4
	cmds := make([]*exec.Cmd, 0, writers)
	outputs := make([]*strings.Builder, 0, writers)
	for i := range writers {
		cmd := exec.Command(akbBinPath, "write", "--append", "notes/shared-append.md") //nolint:gosec // test helper launching akb binary
		cmd.Dir = kbRoot
		cmd.Stdin = strings.NewReader(fmt.Sprintf("Appended by writer %d.", i))
		output := &strings.Builder{}
		cmd.Stdout = output
		cmd.Stderr = output
		if err := cmd.Start(); err != nil {
			t.Fatalf("start write --append %d: %v", i, err)
		}
		cmds = append(cmds, cmd)
		outputs = append(outputs, output)
	}

	waitErr := make(chan error, writers)
	for i, cmd := range cmds {
		go func(index int, cmd *exec.Cmd) {
			if err := cmd.Wait(); err != nil {
				waitErr <- fmt.Errorf("write --append process %d failed: %w\n%s", index, err, outputs[index].String())
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
			t.Fatal("concurrent write --append processes did not finish in time")
		}
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "shared-append.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read shared page: %v", err)
	}
	body := string(data)
	for i := range writers {
		if !strings.Contains(body, fmt.Sprintf("Appended by writer %d.", i)) {
			t.Errorf("append of writer %d is missing from the page:\n%s", i, body)
		}
	}

	commits := 0
	for _, subject := range strings.Split(mustGitInDir(t, kbRoot, "log", "--format=%s"), "\n") {
		if strings.TrimSpace(subject) == commitSubject {
			commits++
		}
	}
	if commits != writers+1 {
		t.Errorf("commits for %s = %d, want %d (the seed write plus one per append) — an append was lost", relPath, commits, writers+1)
	}

	for _, line := range strings.Split(mustGitInDir(t, kbRoot, "status", "--porcelain"), "\n") {
		if len(line) > 3 && strings.HasPrefix(line[3:], "kb/") {
			t.Errorf("page left dirty after the concurrent write --append runs: %q", line)
		}
	}
}

func TestWriteNewPageStampsCreatedAndUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Stamped Note\nsummary: test\ntags: test\n---\nBody."
	if out, err := writeRun(kbRoot, "stamped.md", content); err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "stamped.md"))
	created := frontmatterString(t, fm, "created")
	updated := frontmatterString(t, fm, "updated")
	assertRecentTimestamp(t, created)
	assertRecentTimestamp(t, updated)
	if created > updated {
		t.Errorf("created = %q is after updated = %q", created, updated)
	}
}

func TestWriteNewPageKeepsExplicitUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const explicit = "2020-01-01T00:00:00Z"
	content := "---\ntype: note\ntitle: Explicit Updated\nsummary: test\ntags: test\nupdated: " + explicit + "\n---\nBody."
	if out, err := writeRun(kbRoot, "explicit-updated.md", content); err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "explicit-updated.md"))
	if got := frontmatterString(t, fm, "updated"); got != explicit {
		t.Errorf("updated = %q, want the explicit %q preserved on a new page", got, explicit)
	}
}

func TestWriteOverwriteBumpsUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const stale = "2020-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/notes/overwrite-updated.md",
		"---\ntype: note\ntitle: Overwrite Updated\nsummary: test\ntags: test\ncreated: "+stale+"\nupdated: "+stale+"\n---\nOriginal body.")

	replacement := "---\ntype: note\ntitle: Overwrite Updated\nsummary: test\ntags: test\ncreated: " + stale + "\nupdated: " + stale + "\n---\nReplacement body."
	if out, err := writeRun(kbRoot, "overwrite-updated.md", replacement); err != nil {
		t.Fatalf("akb write overwrite failed: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "overwrite-updated.md"))
	updated := frontmatterString(t, fm, "updated")
	if updated == stale {
		t.Errorf("updated = %q, want the overwrite to move the update time on", updated)
	}
	assertRecentTimestamp(t, updated)
	if got := frontmatterString(t, fm, "created"); got != stale {
		t.Errorf("created = %q, want %q preserved", got, stale)
	}
}

func TestWriteAppendBumpsUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const stale = "2020-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/notes/append-updated.md",
		"---\ntype: note\ntitle: Append Updated\nsummary: test\ntags: test\ncreated: "+stale+"\nupdated: "+stale+"\n---\nOriginal body.")

	out, err := writeRun(kbRoot, "notes/append-updated.md", "Appended body.", "--append")
	if err != nil {
		t.Fatalf("akb write --append failed: %s: %v", out, err)
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

func TestWriteFrontmatterBumpsUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const stale = "2020-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/notes/frontmatter-updated.md",
		"---\ntype: note\ntitle: Frontmatter Updated\nsummary: test\ntags: test\ncreated: "+stale+"\nupdated: "+stale+"\n---\nBody.")

	out, err := writeRun(kbRoot, "notes/frontmatter-updated.md", "", "--frontmatter", "summary=changed")
	if err != nil {
		t.Fatalf("akb write --frontmatter failed: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "frontmatter-updated.md"))
	updated := frontmatterString(t, fm, "updated")
	if updated == stale {
		t.Errorf("updated = %q, want the frontmatter update to move the update time on", updated)
	}
	assertRecentTimestamp(t, updated)
	if got := frontmatterString(t, fm, "summary"); got != "changed" {
		t.Errorf("summary = %q, want the requested update", got)
	}
}

func TestWriteFrontmatterExplicitUpdatedHonored(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const stale = "2020-01-01T00:00:00Z"
	const explicit = "2030-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/notes/frontmatter-explicit.md",
		"---\ntype: note\ntitle: Frontmatter Explicit\nsummary: test\ntags: test\ncreated: "+stale+"\nupdated: "+stale+"\n---\nBody.")

	out, err := writeRun(kbRoot, "notes/frontmatter-explicit.md", "", "--frontmatter", "updated="+explicit, "--frontmatter", "summary=changed")
	if err != nil {
		t.Fatalf("akb write --frontmatter failed: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "frontmatter-explicit.md"))
	if got := frontmatterString(t, fm, "updated"); got != explicit {
		t.Errorf("updated = %q, want the explicit %q honored", got, explicit)
	}
}

// TestWriteAppendCELOldPageSeesPreBumpUpdated pins that a CEL rule comparing
// old_page against page still sees the on-disk `updated` value from before the
// append bumped it.
func TestWriteAppendCELOldPageSeesPreBumpUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	tmplData := `name: oldpage
description: old_page freshness probe
dir: probes
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: old_updated_visible
    rule: 'old_page != null && old_page.frontmatter.updated == timestamp("2020-01-01T00:00:00Z") && page.frontmatter.updated > old_page.frontmatter.updated'
    expect: old_page must carry the pre-write updated timestamp
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", "templates", "oldpage.yaml"), []byte(tmplData), 0600); err != nil {
		t.Fatalf("write probe template: %v", err)
	}

	const stale = "2020-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/probes/probe.md",
		"---\ntype: oldpage\ntitle: Probe\nupdated: "+stale+"\n---\nOriginal body.")

	out, err := writeRun(kbRoot, "probes/probe.md", "Appended body.", "--append")
	if err != nil {
		t.Fatalf("akb write --append failed under the old_page probe rule: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "probes", "probe.md"))
	assertRecentTimestamp(t, frontmatterString(t, fm, "updated"))
}

// TestWriteFrontmatterCELOldPageSeesPreBumpUpdated pins that on the
// --frontmatter branch a CEL rule comparing old_page against page still sees
// the on-disk `updated` value from before the frontmatter update bumped it.
func TestWriteFrontmatterCELOldPageSeesPreBumpUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	tmplData := `name: oldpage
description: old_page freshness probe
dir: probes
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: old_updated_visible
    rule: 'old_page != null && old_page.frontmatter.updated == timestamp("2020-01-01T00:00:00Z") && page.frontmatter.updated > old_page.frontmatter.updated'
    expect: old_page must carry the pre-write updated timestamp
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", "templates", "oldpage.yaml"), []byte(tmplData), 0600); err != nil {
		t.Fatalf("write probe template: %v", err)
	}

	const stale = "2020-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/probes/probe.md",
		"---\ntype: oldpage\ntitle: Probe\nsummary: before\nupdated: "+stale+"\n---\nOriginal body.")

	out, err := writeRun(kbRoot, "probes/probe.md", "", "--frontmatter", "summary=changed")
	if err != nil {
		t.Fatalf("akb write --frontmatter failed under the old_page probe rule: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "probes", "probe.md"))
	assertRecentTimestamp(t, frontmatterString(t, fm, "updated"))
}

// TestWriteOverwriteCELOldPageSeesOnDiskState pins that a full stdin overwrite
// hands a write-time CEL rule the page as it is on disk: a rule comparing
// old_page against page sees the pre-overwrite value instead of null.
func TestWriteOverwriteCELOldPageSeesOnDiskState(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	tmplData := `name: approval
description: old_page state probe
dir: approvals
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: approval_advances
    rule: 'old_page != null && old_page.frontmatter.status == "draft" && page.frontmatter.status == "accepted"'
    expect: status must move from the on-disk draft to accepted
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", "templates", "approval.yaml"), []byte(tmplData), 0600); err != nil {
		t.Fatalf("write probe template: %v", err)
	}

	writeRawPage(t, kbRoot, "kb/approvals/probe.md",
		"---\ntype: approval\ntitle: Probe\nstatus: draft\n---\nDraft body.")

	accepted := "---\ntype: approval\ntitle: Probe\nstatus: accepted\n---\nAccepted body."
	if out, err := writeRun(kbRoot, "approvals/probe.md", accepted); err != nil {
		t.Fatalf("overwriting a draft page failed under the old_page state rule: %s: %v", out, err)
	}

	// The page on disk is accepted now, so the same overwrite no longer finds a
	// draft in old_page and the rule fails.
	out, err := writeRun(kbRoot, "approvals/probe.md", accepted)
	if err == nil {
		t.Fatalf("expected the second overwrite to fail the old_page state rule, got: %s", out)
	}
	if !strings.Contains(out, "status must move from the on-disk draft to accepted") {
		t.Errorf("output = %q, want the failed rule report", out)
	}
}

// TestWriteOverwriteCorruptFrontmatterIsPageError pins that overwriting a page
// whose on-disk frontmatter does not parse fails as a page error: the command
// reports the parse failure and classifies it as a result to act on rather than
// an akb fault, matching the sibling --frontmatter and --append branches.
func TestWriteOverwriteCorruptFrontmatterIsPageError(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)
	withWriteFlags(t, nil, false)

	writeRawPage(t, kbRoot, "kb/notes/corrupt.md",
		"---\ntype: note\ntitle: Corrupt\nsummary: [broken yaml {{{\n---\nOriginal body.")

	// runWrite reads the replacement page from stdin; a regular file keeps the
	// read from blocking and is not a character device.
	stdinPath := filepath.Join(t.TempDir(), "stdin.md")
	overwrite := "---\ntype: note\ntitle: Corrupt\nsummary: rewritten\ntags: test\n---\nRewritten body."
	if err := os.WriteFile(stdinPath, []byte(overwrite), 0600); err != nil {
		t.Fatalf("write stdin content: %v", err)
	}
	stdinFile, err := os.Open(stdinPath) //nolint:gosec // test opening known temp file
	if err != nil {
		t.Fatalf("open stdin content: %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = stdinFile
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = stdinFile.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	})

	err = runWrite(nil, []string{"notes/corrupt.md"})
	if err == nil {
		t.Fatal("expected overwriting a page with corrupt frontmatter to fail")
	}

	var internalErr *internalError
	if errors.As(err, &internalErr) {
		t.Fatalf("failure %v is an akb fault; corrupt page data must fail as a page error", err)
	}
	if !strings.Contains(err.Error(), "parse existing frontmatter") {
		t.Errorf("error = %v, want it to report the frontmatter parse failure", err)
	}
	if code, _ := classifyExit(err); code != exitFailure {
		t.Errorf("exit code = %d, want %d; error: %v", code, exitFailure, err)
	}
}

// TestWriteNewPageCELOldPageIsNull pins that a page akb write creates sees a
// null old_page, while overwriting that page does not.
func TestWriteNewPageCELOldPageIsNull(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	tmplData := `name: newest
description: old_page absence probe
dir: newest
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: page_is_new
    rule: 'old_page == null'
    expect: old_page must be null for a page that does not exist yet
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", "templates", "newest.yaml"), []byte(tmplData), 0600); err != nil {
		t.Fatalf("write probe template: %v", err)
	}

	content := "---\ntype: newest\ntitle: Newest\nsummary: test\ntags: test\n---\nBody."
	if out, err := writeRun(kbRoot, "newest/probe.md", content); err != nil {
		t.Fatalf("akb write of a new page failed under the old_page absence rule: %s: %v", out, err)
	}

	overwrite := "---\ntype: newest\ntitle: Newest\nsummary: test\ntags: test\n---\nOverwritten body."
	out, err := writeRun(kbRoot, "newest/probe.md", overwrite)
	if err == nil {
		t.Fatalf("expected the overwrite to fail the old_page absence rule, got: %s", out)
	}
	if !strings.Contains(out, "old_page must be null for a page that does not exist yet") {
		t.Errorf("output = %q, want the failed rule report", out)
	}
}

// TestWriteUnevaluableRuleIsValidationFailure pins that a write-time rule that
// compiles but cannot be evaluated is a template-authoring problem the caller
// acts on, not an akb fault: the write fails as a validation failure (exit 1),
// the page is not written, and the report names the rule, the evaluation error
// and the has() guard. The template also carries a rule that evaluates to
// false, so the report must keep carrying every failed rule next to the
// unevaluable one.
func TestWriteUnevaluableRuleIsValidationFailure(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	tmplData := `name: sourced
description: probe template whose rule reads an absent optional key unguarded
dir: sourced
schema:
  frontmatter:
    title:
      type: string
      required: true
    sources:
      type: list
      required: false
validations:
  - id: sources_named
    rule: 'page.frontmatter.sources[0] != ""'
    requirement: sources must name at least one source
    expect: sources must name at least one source
  - id: title_not_probe
    rule: 'page.frontmatter.title != "Unguarded"'
    requirement: the title must differ from the probe title
    expect: title must differ from the probe title
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", "templates", "sourced.yaml"), []byte(tmplData), 0600); err != nil {
		t.Fatalf("write probe template: %v", err)
	}

	withWriteFlags(t, nil, false)
	pointStdinAtTempFile(t, "---\ntype: sourced\ntitle: Unguarded\n---\nBody.")

	var runErr error
	stderr := captureStderr(t, func() {
		runErr = runWrite(nil, []string{"sourced/probe.md"})
	})
	if runErr == nil {
		t.Fatalf("expected the unevaluable rule to block the write, stderr: %s", stderr)
	}

	var internalErr *internalError
	if errors.As(runErr, &internalErr) {
		t.Errorf("failure %v is an akb fault; an unevaluable rule is a template-authoring problem", runErr)
	}
	var validationErr validationFailure
	if !errors.As(runErr, &validationErr) {
		t.Fatalf("failure %v is not a validation failure", runErr)
	}
	if code, report := classifyExit(runErr); code != exitFailure || report != "" {
		t.Errorf("classifyExit = (%d, %q), want (%d, %q)", code, report, exitFailure, "")
	}

	if _, err := os.Stat(filepath.Join(kbRoot, "kb", "sourced", "probe.md")); !os.IsNotExist(err) {
		t.Errorf("the page was written despite the unevaluable rule: %v", err)
	}

	for _, want := range []string{
		"rule sources_named could not be evaluated: no such key: sources",
		"template authoring problem",
		"guard optional frontmatter keys with has() (see the kb-management skill's TEMPLATE.md)",
		"title must differ from the probe title",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr = %q, want it to contain %q", stderr, want)
		}
	}
}

// TestWriteUncompilableRuleIsInternalFault pins that a template whose rule does
// not compile — a template file that bypassed `akb template write`, which
// rejects one — stays an akb fault: the write exits 2 and the page is not
// written.
func TestWriteUncompilableRuleIsInternalFault(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	tmplData := `name: broken
description: probe template whose rule does not compile
dir: broken
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: broken_rule
    rule: 'this is not valid CEL (('
    requirement: never satisfiable
    expect: never reported
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", "templates", "broken.yaml"), []byte(tmplData), 0600); err != nil {
		t.Fatalf("write probe template: %v", err)
	}

	withWriteFlags(t, nil, false)
	pointStdinAtTempFile(t, "---\ntype: broken\ntitle: Broken\n---\nBody.")

	var runErr error
	stderr := captureStderr(t, func() {
		runErr = runWrite(nil, []string{"broken/page.md"})
	})
	if runErr == nil {
		t.Fatalf("expected the uncompilable rule to block the write, stderr: %s", stderr)
	}

	var internalErr *internalError
	if !errors.As(runErr, &internalErr) {
		t.Fatalf("failure %v is not an akb fault; an uncompilable rule means the template bypassed `akb template write`", runErr)
	}
	code, report := classifyExit(runErr)
	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	if !strings.Contains(report, "internal: ") || !strings.Contains(report, "compile rule broken_rule") {
		t.Errorf("report = %q, want an internal report naming the rule that does not compile", report)
	}

	if _, err := os.Stat(filepath.Join(kbRoot, "kb", "broken", "page.md")); !os.IsNotExist(err) {
		t.Errorf("the page was written despite the uncompilable rule: %v", err)
	}
}

func TestApproveDoesNotBumpUpdated(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const stale = "2020-01-01T00:00:00Z"
	writeRawPage(t, kbRoot, "kb/notes/approve-no-bump.md",
		"---\ntype: note\ntitle: Approve No Bump\nsummary: test\ntags: test\nis_draft: true\nupdated: "+stale+"\n---\nDraft body.")

	out, err := approveRun(kbRoot, "notes/approve-no-bump.md")
	if err != nil {
		t.Fatalf("akb approve failed: %s: %v", out, err)
	}

	fm := pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "approve-no-bump.md"))
	if got := frontmatterString(t, fm, "updated"); got != stale {
		t.Errorf("updated = %q, want %q unchanged by approve", got, stale)
	}
	if frontmatter.IsDraft(fm.Fields) {
		t.Error("page is still a draft after approve")
	}
}

// abortDocumentInsertSQL, abortLinkInsertSQL, and abortLinkDeleteSQL arm one
// database step of a command path to fail: the first aborts the document insert
// of the search index, the second aborts the link insert of the link graph, and
// the third aborts the link removal of the link graph.
const (
	abortDocumentInsertSQL = `CREATE TRIGGER injected_document_insert_failure
		BEFORE INSERT ON documents
		BEGIN SELECT RAISE(ABORT, 'injected document insert failure'); END`
	abortLinkInsertSQL = `CREATE TRIGGER injected_link_insert_failure
		BEFORE INSERT ON links
		BEGIN SELECT RAISE(ABORT, 'injected link insert failure'); END`
	abortLinkDeleteSQL = `CREATE TRIGGER injected_link_delete_failure
		BEFORE DELETE ON links
		BEGIN SELECT RAISE(ABORT, 'injected link delete failure'); END`
)

// installSearchDBStatement runs one statement against the search database of a
// test KB, so a test can arm a failure before running the command under test.
func installSearchDBStatement(t *testing.T, kbRoot, statement string) {
	t.Helper()

	conn, err := db.OpenKB(kbRoot)
	if err != nil {
		t.Fatalf("open search database: %v", err)
	}
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal

	if _, err := conn.Exec(statement); err != nil {
		t.Fatalf("run search database statement: %v", err)
	}
}

// searchDBDocumentBody returns the body the search index holds for relPath and
// whether the index has a row for the page at all.
func searchDBDocumentBody(t *testing.T, kbRoot, relPath string) (string, bool) {
	t.Helper()

	conn, err := db.OpenKB(kbRoot)
	if err != nil {
		t.Fatalf("open search database: %v", err)
	}
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal

	var body string
	err = conn.QueryRow("SELECT content FROM documents WHERE path = ?", relPath).Scan(&body)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		t.Fatalf("query indexed body of %s: %v", relPath, err)
	}
	return body, true
}

// searchDBRowCount returns how many rows one query counts against the search
// database of a test KB.
func searchDBRowCount(t *testing.T, kbRoot, query string, args ...any) int {
	t.Helper()

	conn, err := db.OpenKB(kbRoot)
	if err != nil {
		t.Fatalf("open search database: %v", err)
	}
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal

	var count int
	if err := conn.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return count
}

// TestWriteIndexFailureAfterCommitReportsRemediation pins that a search-index
// failure after the page commit reports the committed page and the rebuild
// remediation, while the same write succeeds before the failure is armed.
func TestWriteIndexFailureAfterCommitReportsRemediation(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	controlContent := "---\ntype: note\ntitle: Index Control\nsummary: test\ntags: test\n---\nBody."
	controlOut, err := writeRun(kbRoot, "index-control.md", controlContent)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", controlOut, err)
	}
	if strings.Contains(controlOut, "akb index rebuild") {
		t.Errorf("a successful write must not carry rebuild remediation, got: %s", controlOut)
	}

	installSearchDBStatement(t, kbRoot, abortDocumentInsertSQL)

	failedContent := "---\ntype: note\ntitle: Index Failure\nsummary: test\ntags: test\n---\nBody."
	out, err := writeRun(kbRoot, "index-failure.md", failedContent)
	if err == nil {
		t.Fatalf("expected akb write to fail when the document insert is aborted, got: %s", out)
	}
	if !strings.Contains(out, "committed to git") {
		t.Errorf("expected the error to state the page is committed to git, got: %s", out)
	}
	if !strings.Contains(out, "akb index rebuild") {
		t.Errorf("expected the error to carry the rebuild remediation, got: %s", out)
	}
	if _, statErr := os.Stat(filepath.Join(kbRoot, "kb", "notes", "index-failure.md")); statErr != nil {
		t.Errorf("the page should stay on disk after the failed index step, stat error: %v", statErr)
	}
	committed := mustGitInDir(t, kbRoot, "log", "--oneline", "-1", "--", "kb/notes/index-failure.md")
	if !strings.Contains(committed, "akb: write kb/notes/index-failure.md") {
		t.Errorf("commit after the failed index step = %q, want the write commit", committed)
	}
}

// TestWriteLinkFailureAfterCommitReportsRemediation pins the same contract for
// a link-graph failure after the page commit.
func TestWriteLinkFailureAfterCommitReportsRemediation(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	controlContent := "---\ntype: note\ntitle: Link Control\nsummary: test\ntags: test\n---\nSee [[other-note]] for details."
	controlOut, err := writeRun(kbRoot, "link-control.md", controlContent)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", controlOut, err)
	}
	if strings.Contains(controlOut, "akb index rebuild") {
		t.Errorf("a successful write must not carry rebuild remediation, got: %s", controlOut)
	}

	installSearchDBStatement(t, kbRoot, abortLinkInsertSQL)

	failedContent := "---\ntype: note\ntitle: Link Failure\nsummary: test\ntags: test\n---\nSee [[other-note]] for details."
	out, err := writeRun(kbRoot, "link-failure.md", failedContent)
	if err == nil {
		t.Fatalf("expected akb write to fail when the link insert is aborted, got: %s", out)
	}
	if !strings.Contains(out, "committed to git") {
		t.Errorf("expected the error to state the page is committed to git, got: %s", out)
	}
	if !strings.Contains(out, "akb index rebuild") {
		t.Errorf("expected the error to carry the rebuild remediation, got: %s", out)
	}
	committed := mustGitInDir(t, kbRoot, "log", "--oneline", "-1", "--", "kb/notes/link-failure.md")
	if !strings.Contains(committed, "akb: write kb/notes/link-failure.md") {
		t.Errorf("commit after the failed link step = %q, want the write commit", committed)
	}
}

// TestWriteNoCommitIndexFailureReportsStagedState pins that a database failure
// after --no-commit reports the page as staged instead of claiming a commit.
func TestWriteNoCommitIndexFailureReportsStagedState(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	installSearchDBStatement(t, kbRoot, abortDocumentInsertSQL)

	content := "---\ntype: note\ntitle: Staged Failure\nsummary: test\ntags: test\n---\nBody."
	out, err := writeRun(kbRoot, "staged-failure.md", content, "--no-commit")
	if err == nil {
		t.Fatalf("expected akb write --no-commit to fail when the document insert is aborted, got: %s", out)
	}
	if strings.Contains(out, "committed to git") {
		t.Errorf("--no-commit must not claim the page is committed to git, got: %s", out)
	}
	if !strings.Contains(out, "staged but not committed") {
		t.Errorf("expected the error to report the staged page, got: %s", out)
	}
	if !strings.Contains(out, "akb index rebuild") {
		t.Errorf("expected the error to carry the rebuild remediation, got: %s", out)
	}
}

// TestWriteLinkFailureRollsBackSearchIndex pins that the write path indexes the
// page and its links in one transaction: when the second step, the link-graph
// update, fails, the first step, the search-index write, does not persist.
func TestWriteLinkFailureRollsBackSearchIndex(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	controlContent := "---\ntype: note\ntitle: Rollback Control\nsummary: test\ntags: test\n---\nSee [[other-note]] for details."
	if out, err := writeRun(kbRoot, "rollback-control.md", controlContent); err != nil {
		t.Fatalf("control write failed: %s: %v", out, err)
	}
	const controlRel = "kb/notes/rollback-control.md"
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM documents WHERE path = ?", controlRel); got != 1 {
		t.Fatalf("control page rows in documents = %d, want 1 — the write path did not index the control page", got)
	}

	installSearchDBStatement(t, kbRoot, abortLinkInsertSQL)

	failedContent := "---\ntype: note\ntitle: Rollback Failure\nsummary: test\ntags: test\n---\nSee [[other-note]] for details."
	out, err := writeRun(kbRoot, "rollback-failure.md", failedContent)
	if err == nil {
		t.Fatalf("expected akb write to fail when the link insert is aborted, got: %s", out)
	}
	if !strings.Contains(out, "update links:") {
		t.Fatalf("expected the link-graph step to fail, got: %s", out)
	}

	const failedRel = "kb/notes/rollback-failure.md"
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM documents WHERE path = ?", failedRel); got != 0 {
		t.Errorf("documents rows for %s = %d after the failed link step, want 0 — the search-index step did not roll back", failedRel, got)
	}
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM pages WHERE path = ?", failedRel); got != 0 {
		t.Errorf("pages rows for %s = %d after the failed link step, want 0 — the search-index step did not roll back", failedRel, got)
	}
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM links WHERE source_page = ?", failedRel); got != 0 {
		t.Errorf("links rows for %s = %d after the failed link step, want 0", failedRel, got)
	}
}

// symlinkFixture links target at linkPath, skipping the test when the
// filesystem does not support symlinks.
func symlinkFixture(t *testing.T, linkPath, target string) {
	t.Helper()
	if err := os.Symlink(target, linkPath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
}

// assertSymlinkEscape fails unless err is the guard error the resolvers report
// for a path that escapes the base through a symlink, and the CLI classifies it
// as a usage fault.
func assertSymlinkEscape(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected the symlink escape to be rejected, got nil error")
	}
	var guardErr *path.GuardError
	if !errors.As(err, &guardErr) {
		t.Fatalf("failure %v is not a path guard error", err)
	}
	if !errors.Is(err, path.ErrSymlinkEscape) {
		t.Fatalf("failure %v does not wrap ErrSymlinkEscape", err)
	}
	code, report := classifyExit(commandFailure{err: err})
	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	if !strings.Contains(report, "usage: ") || !strings.Contains(report, "escapes the knowledge base through a symlink") {
		t.Errorf("report = %q, want a usage report naming the symlink escape", report)
	}
}

// assertSymlinkEscapeExit fails unless the CLI invocation failed with the usage
// fault exit code and reported the symlink escape.
func assertSymlinkEscapeExit(t *testing.T, out string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected the symlink escape to be rejected, got output: %s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("failure %v is not an exit error", err)
	}
	if exitErr.ExitCode() != exitFault {
		t.Errorf("exit code = %d, want %d; output: %s", exitErr.ExitCode(), exitFault, out)
	}
	if !strings.Contains(out, "usage: ") || !strings.Contains(out, "escapes the knowledge base through a symlink") {
		t.Errorf("output = %q, want a usage report naming the symlink escape", out)
	}
}

// TestReadRejectsPageSymlinkedOutOfTheKB pins the read path against a page that
// is a symlink to a file outside the base: the path is rejected before the
// content is read.
func TestReadRejectsPageSymlinkedOutOfTheKB(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const outsideContent = "---\ntype: note\ntitle: Secret\n---\ncontent outside the base"
	outsideFile := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outsideFile, []byte(outsideContent), 0600); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, filepath.Join(kbRoot, "kb", "evil.md"), outsideFile)

	assertSymlinkEscape(t, runRead(nil, []string{"evil.md"}))

	data, err := os.ReadFile(outsideFile) //nolint:gosec // test reading a known temp file
	if err != nil {
		t.Fatalf("read the outside file: %v", err)
	}
	if string(data) != outsideContent {
		t.Errorf("outside file = %q, want it untouched", string(data))
	}
}

// TestApproveRejectsPageSymlinkedOutOfTheKB pins that approve resolves the page
// through the path guard instead of joining the base root on its own.
func TestApproveRejectsPageSymlinkedOutOfTheKB(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const outsideContent = "---\ntype: note\ntitle: Secret\n---\n<!-- olw-auto: action=review -->\ncontent"
	outsideFile := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outsideFile, []byte(outsideContent), 0600); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, filepath.Join(kbRoot, "kb", "evil.md"), outsideFile)

	assertSymlinkEscape(t, runApprove(nil, []string{"evil.md"}))

	data, err := os.ReadFile(outsideFile) //nolint:gosec // test reading a known temp file
	if err != nil {
		t.Fatalf("read the outside file: %v", err)
	}
	if string(data) != outsideContent {
		t.Errorf("outside file = %q, want it untouched by approve", string(data))
	}
}

// TestWriteRejectsSymlinkedTypeDirectoryCandidates pins that the type-derived
// directory the write command joins onto the validated path is checked for
// symlink escapes before it is used for I/O.
func TestWriteRejectsSymlinkedTypeDirectoryCandidates(t *testing.T) {
	const noteContent = "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."

	t.Run("type dir symlinked out of the base", func(t *testing.T) {
		kbRoot := writeSetupTestKB(t)
		defer writeCleanup(kbRoot)

		outsideDir := t.TempDir()
		symlinkFixture(t, filepath.Join(kbRoot, "kb", "notes"), outsideDir)

		out, err := writeRun(kbRoot, "note.md", noteContent)
		assertSymlinkEscapeExit(t, out, err)
		if _, err := os.Stat(filepath.Join(outsideDir, "note.md")); !os.IsNotExist(err) {
			t.Errorf("the write created a page outside the base: %v", err)
		}
	})

	t.Run("subdirectory below the type dir symlinked out of the base", func(t *testing.T) {
		kbRoot := writeSetupTestKB(t)
		defer writeCleanup(kbRoot)

		outsideDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(kbRoot, "kb", "notes"), 0750); err != nil {
			t.Fatal(err)
		}
		symlinkFixture(t, filepath.Join(kbRoot, "kb", "notes", "sub"), outsideDir)

		out, err := writeRun(kbRoot, "sub/note.md", noteContent)
		assertSymlinkEscapeExit(t, out, err)
		if _, err := os.Stat(filepath.Join(outsideDir, "note.md")); !os.IsNotExist(err) {
			t.Errorf("the write created a page outside the base: %v", err)
		}
	})

	t.Run("frontmatter candidate under a symlinked type dir", func(t *testing.T) {
		kbRoot := writeSetupTestKB(t)
		defer writeCleanup(kbRoot)

		outsideDir := t.TempDir()
		const outsideContent = "---\ntype: note\ntitle: Outside\nsummary: test\ntags: test\n---\noutside body"
		outsidePage := filepath.Join(outsideDir, "note.md")
		if err := os.WriteFile(outsidePage, []byte(outsideContent), 0600); err != nil {
			t.Fatal(err)
		}
		symlinkFixture(t, filepath.Join(kbRoot, "kb", "notes"), outsideDir)

		out, err := writeRun(kbRoot, "note.md", "", "--frontmatter", "title=Updated")
		assertSymlinkEscapeExit(t, out, err)

		data, err := os.ReadFile(outsidePage) //nolint:gosec // test reading a known temp file
		if err != nil {
			t.Fatalf("read the outside page: %v", err)
		}
		if string(data) != outsideContent {
			t.Errorf("the frontmatter update wrote through the symlink: %s", string(data))
		}
	})
}

// TestWriteRejectsSymlinkedTemplatesDir pins that a templates directory that is
// a symlink out of the base is rejected before the write reads templates.
func TestWriteRejectsSymlinkedTemplatesDir(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	outsideDir := t.TempDir()
	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := os.RemoveAll(templatesDir); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, templatesDir, outsideDir)

	content := "---\ntype: note\ntitle: Test\nsummary: test\ntags: test\n---\nContent."
	out, err := writeRun(kbRoot, "note.md", content)
	assertSymlinkEscapeExit(t, out, err)

	entries, err := os.ReadDir(outsideDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("templates outside the base were touched: %v", entries)
	}
}

// TestInTreeSymlinksStillWork pins that a symlink whose target stays inside the
// base keeps resolving: only links out of the base are rejected.
func TestInTreeSymlinksStillWork(t *testing.T) {
	t.Run("read through a page alias", func(t *testing.T) {
		kbRoot := writeSetupTestKB(t)
		defer writeCleanup(kbRoot)

		content := "---\ntype: note\ntitle: Real Note\nsummary: test\ntags: test\n---\nreal body"
		if out, err := writeRun(kbRoot, "real.md", content); err != nil {
			t.Fatalf("setup write failed: %s: %v", out, err)
		}
		symlinkFixture(t, filepath.Join(kbRoot, "kb", "alias.md"), filepath.Join(kbRoot, "kb", "notes", "real.md"))

		out, err := captureOutput(func() error { return runRead(nil, []string{"alias.md"}) })
		if err != nil {
			t.Fatalf("read through the in-tree alias failed: %v", err)
		}
		if !strings.Contains(out, "real body") {
			t.Errorf("read output = %q, want the aliased page content", out)
		}
	})

	t.Run("write through a type dir alias", func(t *testing.T) {
		kbRoot := writeSetupTestKB(t)
		defer writeCleanup(kbRoot)

		if err := os.MkdirAll(filepath.Join(kbRoot, "kb", "real-notes"), 0750); err != nil {
			t.Fatal(err)
		}
		symlinkFixture(t, filepath.Join(kbRoot, "kb", "notes"), filepath.Join(kbRoot, "kb", "real-notes"))

		content := "---\ntype: note\ntitle: Aliased\nsummary: test\ntags: test\n---\nBody."
		out, err := writeRun(kbRoot, "aliased.md", content)
		if err != nil && strings.Contains(out, "escapes the knowledge base") {
			t.Fatalf("the in-tree type dir alias was rejected as an escape: %s", out)
		}
		if _, err := os.Stat(filepath.Join(kbRoot, "kb", "real-notes", "aliased.md")); err != nil {
			t.Errorf("expected the page behind the alias, got: %v", err)
		}
	})
}

// resolvePageFixture lays out a page fixture: a KB root holding the pages given
// as KB-relative paths, so a test can drive page resolution without the CLI.
func resolvePageFixture(t *testing.T, pages map[string]string) string {
	t.Helper()

	kbRoot := t.TempDir()
	for relPath, content := range pages {
		fullPath := filepath.Join(kbRoot, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
			t.Fatalf("create directory for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			t.Fatalf("write page %s: %v", relPath, err)
		}
	}
	return kbRoot
}

// typeDirProviderStub supplies fixed type directories, for tests that drive the
// candidate order of page resolution directly.
func typeDirProviderStub(dirs ...string) typeDirsProvider {
	return func() ([]string, error) { return dirs, nil }
}

// TestResolveExistingPageCandidateOrder pins where the shared page resolution
// looks for a page: the path the input names first, then the type directories
// in ascending order, and the first readable candidate wins.
func TestResolveExistingPageCandidateOrder(t *testing.T) {
	const (
		baseBody      = "---\ntype: note\ntitle: Base\n---\nbase body"
		notesBody     = "---\ntype: note\ntitle: Notes\n---\nnotes body"
		decisionsBody = "---\ntype: adr\ntitle: Decisions\n---\ndecisions body"
	)

	cases := []struct {
		name        string
		pages       map[string]string
		inputPath   string
		dirs        []string
		wantPath    string
		wantRelPath string
		wantContent string
	}{
		{
			name:        "the named path wins over a type directory",
			pages:       map[string]string{"kb/shared.md": baseBody, "kb/notes/shared.md": notesBody},
			inputPath:   "shared.md",
			dirs:        []string{"notes"},
			wantPath:    "kb/shared.md",
			wantRelPath: "kb/shared.md",
			wantContent: baseBody,
		},
		{
			name:        "the type directory holds a page addressed by its filename",
			pages:       map[string]string{"kb/notes/shared.md": notesBody},
			inputPath:   "shared.md",
			dirs:        []string{"notes"},
			wantPath:    "kb/notes/shared.md",
			wantRelPath: "kb/notes/shared.md",
			wantContent: notesBody,
		},
		{
			name:        "type directories are tried in ascending order",
			pages:       map[string]string{"kb/notes/shared.md": notesBody, "kb/decisions/shared.md": decisionsBody},
			inputPath:   "shared.md",
			dirs:        []string{"notes", "decisions"},
			wantPath:    "kb/decisions/shared.md",
			wantRelPath: "kb/decisions/shared.md",
			wantContent: decisionsBody,
		},
		{
			name:        "a type directory that holds no page is passed over",
			pages:       map[string]string{"kb/notes/shared.md": notesBody},
			inputPath:   "shared.md",
			dirs:        []string{"decisions", "notes"},
			wantPath:    "kb/notes/shared.md",
			wantRelPath: "kb/notes/shared.md",
			wantContent: notesBody,
		},
		{
			name:        "a named path with a directory keeps that directory",
			pages:       map[string]string{"kb/notes/shared.md": notesBody},
			inputPath:   "notes/shared.md",
			dirs:        []string{"notes"},
			wantPath:    "kb/notes/shared.md",
			wantRelPath: "kb/notes/shared.md",
			wantContent: notesBody,
		},
		{
			name:        "the kb prefix of the input does not change the candidates",
			pages:       map[string]string{"kb/notes/shared.md": notesBody},
			inputPath:   "kb/shared.md",
			dirs:        []string{"notes"},
			wantPath:    "kb/notes/shared.md",
			wantRelPath: "kb/notes/shared.md",
			wantContent: notesBody,
		},
		{
			name:        "without a type directory the named path is the only candidate",
			pages:       map[string]string{"kb/shared.md": baseBody},
			inputPath:   "shared.md",
			dirs:        nil,
			wantPath:    "kb/shared.md",
			wantRelPath: "kb/shared.md",
			wantContent: baseBody,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kbRoot := resolvePageFixture(t, tc.pages)

			fullPath, relPath, content, err := resolveExistingPage(kbRoot, tc.inputPath, typeDirProviderStub(tc.dirs...))
			if err != nil {
				t.Fatalf("resolve %q: %v", tc.inputPath, err)
			}

			if want := filepath.Join(kbRoot, filepath.FromSlash(tc.wantPath)); fullPath != want {
				t.Errorf("path = %q, want %q", fullPath, want)
			}
			if relPath != tc.wantRelPath {
				t.Errorf("relative path = %q, want %q", relPath, tc.wantRelPath)
			}
			if string(content) != tc.wantContent {
				t.Errorf("content = %q, want %q", string(content), tc.wantContent)
			}
		})
	}
}

// TestResolveExistingPageReportsAbsentPage pins that a page no candidate holds
// reports the not-found sentinel rather than an error.
func TestResolveExistingPageReportsAbsentPage(t *testing.T) {
	kbRoot := resolvePageFixture(t, map[string]string{"kb/notes/other.md": "---\ntype: note\ntitle: Other\n---\nbody"})

	_, _, _, err := resolveExistingPage(kbRoot, "missing.md", typeDirProviderStub("notes", "decisions"))
	if !errors.Is(err, errPageNotFound) {
		t.Errorf("resolve of an absent page = %v, want errPageNotFound", err)
	}
}

// TestResolveExistingPageRejectsEscapingTypeDir pins that a type directory
// candidate is checked against the KB root before the page below it is read.
func TestResolveExistingPageRejectsEscapingTypeDir(t *testing.T) {
	kbRoot := resolvePageFixture(t, map[string]string{"kb/keep.md": "---\ntype: note\ntitle: Keep\n---\nbody"})

	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "escape.md"), []byte("---\ntype: note\ntitle: Outside\n---\noutside body"), 0600); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, filepath.Join(kbRoot, "kb", "notes"), outsideDir)

	_, _, _, err := resolveExistingPage(kbRoot, "escape.md", typeDirProviderStub("notes"))
	assertSymlinkEscape(t, err)
}

// TestTypeDirsFromTemplates pins the directories the provider reads from loaded
// templates: one per template that declares one.
func TestTypeDirsFromTemplates(t *testing.T) {
	templates := map[string]template.Template{
		"note": {Name: "note", Dir: "notes"},
		"adr":  {Name: "adr", Dir: "decisions"},
		"sink": {Name: "sink"},
	}

	dirs, err := typeDirsFromTemplates(templates)()
	if err != nil {
		t.Fatalf("type dirs of the loaded templates: %v", err)
	}

	sort.Strings(dirs)
	if got := strings.Join(dirs, ","); got != "decisions,notes" {
		t.Errorf("type dirs = %q, want %q", got, "decisions,notes")
	}
}

// TestWriteNotFoundMessages pins the not-found report each branch of akb write
// gives for a page that is under no type directory, byte for byte.
func TestWriteNotFoundMessages(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "frontmatter update",
			args: []string{"--frontmatter", "title=Updated"},
			want: "page does not exist; use 'akb write' without --frontmatter to create",
		},
		{
			name: "append",
			args: []string{"--append"},
			want: "page does not exist; use 'akb write' without --append to create",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kbRoot := writeSetupTestKB(t)
			defer writeCleanup(kbRoot)

			out, err := writeRun(kbRoot, "absent-note.md", "Appended body.", tc.args...)
			if err == nil {
				t.Fatalf("expected the update of an absent page to fail, got: %s", out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("output = %q, want %q", out, tc.want)
			}
			if _, statErr := os.Stat(filepath.Join(kbRoot, "kb", "notes", "absent-note.md")); !os.IsNotExist(statErr) {
				t.Errorf("the failed command created a page: %v", statErr)
			}
		})
	}
}

// TestWriteBareFilenameReachesTypeDir pins that both update branches of akb
// write address a page stored under the directory of its type by its bare
// filename, and report it by its stored path.
func TestWriteBareFilenameReachesTypeDir(t *testing.T) {
	t.Run("frontmatter update", func(t *testing.T) {
		kbRoot := writeSetupTestKB(t)
		defer writeCleanup(kbRoot)

		writeRawPage(t, kbRoot, "kb/notes/bare-frontmatter.md",
			"---\ntype: note\ntitle: Bare Frontmatter\nsummary: before\ntags: test\n---\nBody.")

		out, err := writeRun(kbRoot, "bare-frontmatter.md", "", "--frontmatter", "summary=changed")
		if err != nil {
			t.Fatalf("akb write --frontmatter by bare filename failed: %s: %v", out, err)
		}
		if !strings.Contains(out, "Updated frontmatter for kb/notes/bare-frontmatter.md") {
			t.Errorf("output = %q, want the page reported by its stored path", out)
		}
		if got := frontmatterString(t, pageFrontmatter(t, filepath.Join(kbRoot, "kb", "notes", "bare-frontmatter.md")), "summary"); got != "changed" {
			t.Errorf("summary = %q, want the requested update", got)
		}
	})

	t.Run("append", func(t *testing.T) {
		kbRoot := writeSetupTestKB(t)
		defer writeCleanup(kbRoot)

		const relPath = "kb/notes/bare-append.md"
		seed := "---\ntype: note\ntitle: Bare Append\nsummary: test\ntags: test\n---\nOriginal body."
		if out, err := writeRun(kbRoot, "bare-append.md", seed); err != nil {
			t.Fatalf("seed write failed: %s: %v", out, err)
		}

		out, err := writeRun(kbRoot, "bare-append.md", "Appended body.", "--append")
		if err != nil {
			t.Fatalf("akb write --append by bare filename failed: %s: %v", out, err)
		}
		if !strings.Contains(out, "Appended to "+relPath) {
			t.Errorf("output = %q, want the page reported by its stored path", out)
		}
		if _, statErr := os.Stat(filepath.Join(kbRoot, "kb", "bare-append.md")); !os.IsNotExist(statErr) {
			t.Errorf("the append reached a second page at the KB root: %v", statErr)
		}
		indexed, ok := searchDBDocumentBody(t, kbRoot, relPath)
		if !ok {
			t.Fatalf("expected the appended page indexed at %s", relPath)
		}
		if !strings.Contains(indexed, "Appended body.") {
			t.Errorf("indexed body = %q, want the appended content", indexed)
		}
	})
}

// requiredFieldsTestTemplate is the adr-shaped schema the required-field
// checker tests drive: eight required fields plus one optional.
func requiredFieldsTestTemplate() template.Template {
	return template.Template{
		Name: "unit",
		Schema: template.Schema{Frontmatter: map[string]template.FieldSchema{
			"title":    {Type: "string", Required: true},
			"type":     {Type: "string", Required: true},
			"summary":  {Type: "string", Required: true},
			"tags":     {Type: "list", Required: true},
			"status":   {Type: "string", Required: true},
			"deciders": {Type: "string", Required: true},
			"created":  {Type: "string", Required: true},
			"updated":  {Type: "string", Required: true},
			"sources":  {Type: "list"},
		}},
	}
}

// TestCheckRequiredFields pins the presence-only contract of the checker: every
// unset required field is reported once in a stable order, a required field
// with an empty value counts as set, and optional fields are never reported.
func TestCheckRequiredFields(t *testing.T) {
	tmpl := requiredFieldsTestTemplate()

	complete := map[string]any{
		"summary":  "A page",
		"tags":     []any{"test"},
		"status":   "accepted",
		"deciders": "team",
		"created":  "2026-01-01",
		"updated":  "2026-09-01",
	}
	emptyStatus := make(map[string]any, len(complete))
	for key, value := range complete {
		emptyStatus[key] = value
	}
	emptyStatus["status"] = ""

	cases := []struct {
		name string
		fm   *frontmatter.ParsedFrontmatter
		want []string
	}{
		{
			name: "missing one field",
			fm: &frontmatter.ParsedFrontmatter{
				Type:  "unit",
				Title: "Hello",
				Fields: map[string]any{
					"summary":  "A page",
					"tags":     []any{"test"},
					"deciders": "team",
					"created":  "2026-01-01",
					"updated":  "2026-09-01",
				},
			},
			want: []string{"status"},
		},
		{
			name: "missing many fields lists all of them",
			fm: &frontmatter.ParsedFrontmatter{
				Type:   "unit",
				Title:  "Hello",
				Fields: map[string]any{},
			},
			want: []string{"created", "deciders", "status", "summary", "tags", "updated"},
		},
		{
			name: "all required fields present",
			fm: &frontmatter.ParsedFrontmatter{
				Type:   "unit",
				Title:  "Hello",
				Fields: complete,
			},
			want: nil,
		},
		{
			name: "a required field with an empty value is present",
			fm: &frontmatter.ParsedFrontmatter{
				Type:   "unit",
				Title:  "Hello",
				Fields: emptyStatus,
			},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkRequiredFields(tmpl, tc.fm); !slices.Equal(got, tc.want) {
				t.Errorf("checkRequiredFields = %v, want %v", got, tc.want)
			}
		})
	}
}

// adrMissingStatus is an adr page that leaves out the schema-required status
// field, planted on disk so the update branches read it as an existing page.
const adrMissingStatus = `---
type: adr
title: Planted ADR
summary: Planted ADR without the required status field
tags: test
deciders: team
created: '2026-01-01'
updated: '2026-09-01'
---
## Context

Context body.

## Decision

Decision body.

## Consequences

Consequences body.`

// TestWriteMissingRequiredFieldsFailsValidation pins that a write of a page
// whose frontmatter leaves out schema-required fields is a validation failure:
// exit 1, every missing field named, and no page written.
func TestWriteMissingRequiredFieldsFailsValidation(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: adr\ntitle: Missing Fields\nsummary: ADR without status and deciders\ntags: test\ncreated: '2026-01-01'\nupdated: '2026-09-01'\n---\n## Context\n\nContext body.\n\n## Decision\n\nDecision body.\n\n## Consequences\n\nConsequences body."
	out, err := writeRun(kbRoot, "missing-fields.md", content)
	if err == nil {
		t.Fatalf("expected the write to fail, got: %s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %T: %v", err, err)
	}
	if code := exitErr.ExitCode(); code != exitFailure {
		t.Errorf("exit code = %d, want %d; output: %s", code, exitFailure, out)
	}
	if want := `missing required frontmatter field(s): deciders, status (declared required by template "adr")`; !strings.Contains(out, want) {
		t.Errorf("output = %q, want %q", out, want)
	}
	if _, statErr := os.Stat(filepath.Join(kbRoot, "kb", "decisions", "missing-fields.md")); !os.IsNotExist(statErr) {
		t.Errorf("the failed write created a page: %v", statErr)
	}
}

// TestWriteFrontmatterUpdateMissingRequiredFieldFailsValidation pins that the
// --frontmatter branch refuses a page that leaves out a schema-required field,
// and leaves the page byte-identical.
func TestWriteFrontmatterUpdateMissingRequiredFieldFailsValidation(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const relPath = "kb/decisions/planted.md"
	writeRawPage(t, kbRoot, relPath, adrMissingStatus)

	out, err := writeRun(kbRoot, "decisions/planted.md", "", "--frontmatter", "summary=changed")
	if err == nil {
		t.Fatalf("expected the frontmatter update to fail, got: %s", out)
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

	data, readErr := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
	if readErr != nil {
		t.Fatalf("read the planted page: %v", readErr)
	}
	if string(data) != adrMissingStatus {
		t.Errorf("page changed by the failed update:\nbefore: %q\nafter:  %q", adrMissingStatus, data)
	}
}

// TestWriteAppendMissingRequiredFieldFailsValidation pins that the --append
// branch refuses a page that leaves out a schema-required field, and leaves the
// page byte-identical.
func TestWriteAppendMissingRequiredFieldFailsValidation(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const relPath = "kb/decisions/planted-append.md"
	writeRawPage(t, kbRoot, relPath, adrMissingStatus)

	out, err := writeRun(kbRoot, "decisions/planted-append.md", "Appended body.", "--append")
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

	data, readErr := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
	if readErr != nil {
		t.Fatalf("read the planted page: %v", readErr)
	}
	if string(data) != adrMissingStatus {
		t.Errorf("page changed by the failed append:\nbefore: %q\nafter:  %q", adrMissingStatus, data)
	}
}
