package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func approveRun(kbRoot, inputPath string, extraArgs ...string) (string, error) {
	args := append([]string{"approve"}, extraArgs...)
	args = append(args, inputPath)
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, args...) //nolint:gosec // test helper launching akb binary
	cmd.Dir = kbRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestApproveStripsAnnotations(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\n---\nSome content.\n<!-- olw-auto: action=review -->\nMore content."
	_, err := writeRun(kbRoot, "draft.md", content)
	if err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	out, err := approveRun(kbRoot, "notes/draft.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}

	if !strings.Contains(out, "Approved 'notes/draft.md'") {
		t.Errorf("expected approval message, got: %s", out)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "draft.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read approved file: %v", err)
	}

	if strings.Contains(string(data), "olw-auto") {
		t.Errorf("annotation should be stripped, got: %s", string(data))
	}
}

func TestApproveStripsProvenanceMarkers(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\n---\n^[inferred] Some claim.\n^[ambiguous] Another claim.\n^[extracted] Data point."
	_, err := writeRun(kbRoot, "draft.md", content)
	if err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	out, err := approveRun(kbRoot, "notes/draft.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "draft.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read approved file: %v", err)
	}

	body := string(data)
	if strings.Contains(body, "^[inferred]") {
		t.Errorf("inferred marker should be stripped")
	}
	if strings.Contains(body, "^[ambiguous]") {
		t.Errorf("ambiguous marker should be stripped")
	}
	if strings.Contains(body, "^[extracted]") {
		t.Errorf("extracted marker should be stripped")
	}
	if !strings.Contains(body, "Some claim.") {
		t.Errorf("regular content should be preserved")
	}
}

func TestApprovePreservesMarkersInCodeBlocks(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\n---\n^[inferred] Outside code.\n```\n^[inferred] Inside code.\n```"
	_, err := writeRun(kbRoot, "draft.md", content)
	if err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	out, err := approveRun(kbRoot, "notes/draft.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "draft.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read approved file: %v", err)
	}

	body := string(data)
	if strings.Contains(body, "^[inferred] Outside code.") {
		t.Errorf("marker outside code should be stripped")
	}
	if !strings.Contains(body, "^[inferred] Inside code.") {
		t.Errorf("marker inside code block should be preserved")
	}
}

func TestApproveSetsIsDraftFalse(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\n---\nContent."
	_, err := writeRun(kbRoot, "draft.md", content)
	if err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	out, err := approveRun(kbRoot, "notes/draft.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}

	data, err := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "draft.md")) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read approved file: %v", err)
	}

	if !strings.Contains(string(data), "is_draft: false") {
		t.Errorf("expected is_draft: false in frontmatter, got: %s", string(data))
	}
}

func TestApproveNoOpWhenAlreadyApproved(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Approved Note\nsummary: test\ntags: test\nis_draft: false\n---\nContent."
	_, err := writeRun(kbRoot, "approved.md", content)
	if err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	out, err := approveRun(kbRoot, "notes/approved.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}

	if !strings.Contains(out, "Page 'notes/approved.md' is already approved") {
		t.Errorf("expected no-op message, got: %s", out)
	}
}

func TestApproveErrorWhenPageNotFound(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	out, err := approveRun(kbRoot, "nonexistent.md")
	if err == nil {
		t.Fatal("expected error for missing page, got nil")
	}

	if !strings.Contains(out, "nonexistent.md' not found") {
		t.Errorf("expected not-found error, got: %s", out)
	}
	if !strings.Contains(out, "akb list") {
		t.Errorf("expected suggestion to use 'akb list', got: %s", out)
	}
}

func TestApproveGitCommitMessage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\n---\nContent."
	_, err := writeRun(kbRoot, "draft.md", content)
	if err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	out, err := approveRun(kbRoot, "notes/draft.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}

	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "log", "--oneline", "-1")
	cmd.Dir = kbRoot
	gitOut, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if !strings.Contains(string(gitOut), "akb: approve notes/draft.md") {
		t.Errorf("expected commit message 'akb: approve notes/draft.md', got: %s", string(gitOut))
	}
}
