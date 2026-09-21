package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestApproveStripsOnlyRealAnnotations(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\n---\n<!-- olw-auto: action=review -->\n```\n<!-- olw-auto: action=keep -->\n```"
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
	if strings.Count(body, "olw-auto") != 1 {
		t.Errorf("expected only the annotation inside the fenced block, got: %s", body)
	}
	if !strings.Contains(body, "<!-- olw-auto: action=keep -->") {
		t.Errorf("annotation inside the fenced block should survive approval, got: %s", body)
	}
	if !strings.Contains(body, "is_draft: false") {
		t.Errorf("expected is_draft: false in frontmatter, got: %s", body)
	}
}

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

func approveAllDraftsRun(kbRoot string) (string, error) {
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, "approve", "--all-drafts")
	cmd.Dir = kbRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func approveSearchRun(t *testing.T, kbRoot, query string) (string, error) {
	t.Helper()
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, "search", query)
	cmd.Dir = kbRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func approveBacklinksRun(t *testing.T, kbRoot, inputPath string) (string, error) {
	t.Helper()
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, "backlinks", inputPath)
	cmd.Dir = kbRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// approveReindexFixture writes a draft whose body carries an annotation and
// provenance markers plus a wikilink, then creates the link target afterwards.
// The draft is therefore indexed verbatim (marker text searchable) and its link
// entry stays unresolved until the page is approved.
func approveReindexFixture(t *testing.T, kbRoot string) {
	t.Helper()

	draft := "---\ntype: note\ntitle: Draft Note\nsummary: test\ntags: test\n---\nCanonical sentence.\n<!-- olw-auto: action=review -->\n^[inferred] Secondary sentence.\nSee [[target]]."
	if out, err := writeRun(kbRoot, "draft.md", draft); err != nil {
		t.Fatalf("write draft: %s: %v", out, err)
	}

	target := "---\ntype: note\ntitle: Link Target\nsummary: test\ntags: test\n---\nTarget body."
	if out, err := writeRun(kbRoot, "target.md", target); err != nil {
		t.Fatalf("write target: %s: %v", out, err)
	}

	out, err := approveSearchRun(t, kbRoot, "review")
	if err != nil {
		t.Fatalf("search before approve: %s: %v", out, err)
	}
	if !strings.Contains(out, "kb/notes/draft.md") {
		t.Fatalf("expected draft indexed verbatim before approval, got: %s", out)
	}

	out, err = approveBacklinksRun(t, kbRoot, "notes/target.md")
	if err != nil {
		t.Fatalf("backlinks before approve: %s: %v", out, err)
	}
	if strings.Contains(out, "kb/notes/draft.md") {
		t.Fatalf("expected draft link unresolved before approval, got: %s", out)
	}
}

// assertApproveReindexed checks that the search index and link graph reflect
// the approved page: annotation and marker text gone, stripped body searchable,
// and links re-resolved.
func assertApproveReindexed(t *testing.T, kbRoot string) {
	t.Helper()

	for _, query := range []string{"review", "inferred"} {
		out, err := approveSearchRun(t, kbRoot, query)
		if err != nil {
			t.Fatalf("search %q after approve: %s: %v", query, out, err)
		}
		if strings.Contains(out, "kb/notes/draft.md") {
			t.Errorf("search %q should not match approved page, got: %s", query, out)
		}
	}

	out, err := approveSearchRun(t, kbRoot, "Canonical")
	if err != nil {
		t.Fatalf("search after approve: %s: %v", out, err)
	}
	if !strings.Contains(out, "kb/notes/draft.md") {
		t.Errorf("approved page missing from search results, got: %s", out)
	}

	out, err = approveBacklinksRun(t, kbRoot, "notes/target.md")
	if err != nil {
		t.Fatalf("backlinks after approve: %s: %v", out, err)
	}
	if !strings.Contains(out, "kb/notes/draft.md") {
		t.Errorf("approved page missing from backlinks, got: %s", out)
	}
}

func TestApproveReindexesSearchAndLinks(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	approveReindexFixture(t, kbRoot)

	out, err := approveRun(kbRoot, "notes/draft.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}

	assertApproveReindexed(t, kbRoot)
}

func TestApproveAllDraftsReindexesSearchAndLinks(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	approveReindexFixture(t, kbRoot)

	out, err := approveAllDraftsRun(kbRoot)
	if err != nil {
		t.Fatalf("approve --all-drafts failed: %s: %v", out, err)
	}

	assertApproveReindexed(t, kbRoot)
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

// TestApproveLinkFailureRollsBackSearchIndex pins that approve reindexes the
// approved body and its links in one transaction: when the second step, the
// link-graph update, fails, the first step, the search-index write, does not
// persist and the search index keeps the pre-approval body.
func TestApproveLinkFailureRollsBackSearchIndex(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	approveReindexFixture(t, kbRoot)

	const draftRel = "kb/notes/draft.md"
	preBody, ok := searchDBDocumentBody(t, kbRoot, draftRel)
	if !ok {
		t.Fatalf("expected the draft indexed at %s before approval", draftRel)
	}
	if !strings.Contains(preBody, "review") {
		t.Fatalf("pre-approval indexed body %q does not carry the annotation text", preBody)
	}

	installSearchDBStatement(t, kbRoot, abortLinkInsertSQL)

	out, err := approveRun(kbRoot, "notes/draft.md")
	if err == nil {
		t.Fatalf("expected akb approve to fail when the link insert is aborted, got: %s", out)
	}
	if !strings.Contains(out, "update links:") {
		t.Fatalf("expected the link-graph step to fail, got: %s", out)
	}

	postBody, ok := searchDBDocumentBody(t, kbRoot, draftRel)
	if !ok {
		t.Fatalf("expected %s to stay in the search index after the failed link step", draftRel)
	}
	if postBody != preBody {
		t.Errorf("indexed body after the failed link step = %q, want the pre-approval body %q — the search-index step did not roll back", postBody, preBody)
	}

	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM links WHERE source_page = ? AND raw_target = ? AND resolved_to IS NULL", draftRel, "target"); got != 1 {
		t.Errorf("pre-approval link rows after the failed link step = %d, want 1", got)
	}
}

// TestApproveAllDraftsLinkFailureRollsBackSearchIndex pins the same rollback on
// the batch path, which approves each draft through the same per-page
// transaction.
func TestApproveAllDraftsLinkFailureRollsBackSearchIndex(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	approveReindexFixture(t, kbRoot)

	const draftRel = "kb/notes/draft.md"
	preBody, ok := searchDBDocumentBody(t, kbRoot, draftRel)
	if !ok {
		t.Fatalf("expected the draft indexed at %s before approval", draftRel)
	}

	installSearchDBStatement(t, kbRoot, abortLinkInsertSQL)

	out, err := approveAllDraftsRun(kbRoot)
	if err == nil {
		t.Fatalf("expected approve --all-drafts to fail when the link insert is aborted, got: %s", out)
	}

	postBody, ok := searchDBDocumentBody(t, kbRoot, draftRel)
	if !ok {
		t.Fatalf("expected %s to stay in the search index after the failed link step", draftRel)
	}
	if postBody != preBody {
		t.Errorf("indexed body after the failed link step = %q, want the pre-approval body %q — the search-index step did not roll back", postBody, preBody)
	}
}
