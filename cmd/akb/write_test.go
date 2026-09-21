package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
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
