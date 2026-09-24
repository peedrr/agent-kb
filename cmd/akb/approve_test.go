package main

import (
	"errors"
	"fmt"
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

// TestApproveAllDraftsRejectsSymlinkedPage pins the bulk-approve walk against a
// page that is a symlink to a file outside the base: the path is rejected
// before the page is read or rewritten.
func TestApproveAllDraftsRejectsSymlinkedPage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const outsideContent = "---\ntype: note\ntitle: Secret\nis_draft: true\n---\ncontent outside the base"
	outsideFile := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outsideFile, []byte(outsideContent), 0600); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, filepath.Join(kbRoot, "kb", "evil.md"), outsideFile)

	origAllDrafts := approveAllDrafts
	approveAllDrafts = true
	t.Cleanup(func() { approveAllDrafts = origAllDrafts })

	assertSymlinkEscape(t, runApprove(nil, nil))

	data, err := os.ReadFile(outsideFile) //nolint:gosec // test reading a known temp file
	if err != nil {
		t.Fatalf("read the outside file: %v", err)
	}
	if string(data) != outsideContent {
		t.Errorf("outside file = %q, want it untouched by approve", string(data))
	}
}

// gitLogOneline returns the commit subjects of the repository at kbRoot,
// newest first.
func gitLogOneline(t *testing.T, kbRoot string) string {
	t.Helper()
	cmd := exec.Command( //nolint:gosec // test helper reading the repository log
		"git", "log", "--oneline")
	cmd.Dir = kbRoot
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	return string(out)
}

// approveDraftsFixture writes each name as a draft page through akb write and
// returns the base-relative path of every page. Each body carries an
// annotation, so an approval is observable on disk as a rewrite that strips the
// annotation and sets is_draft: false.
func approveDraftsFixture(t *testing.T, kbRoot string, names ...string) []string {
	t.Helper()

	paths := make([]string, 0, len(names))
	for _, name := range names {
		content := fmt.Sprintf("---\ntype: note\ntitle: %s\nsummary: test\ntags: test\n---\n<!-- olw-auto: action=review -->\nContent of %s.", name, name)
		if out, err := writeRun(kbRoot, name, content); err != nil {
			t.Fatalf("write draft %s: %s: %v", name, out, err)
		}

		relPath := "kb/notes/" + name
		data, err := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
		if err != nil {
			t.Fatalf("read draft %s: %v", relPath, err)
		}
		if !strings.Contains(string(data), "olw-auto") {
			t.Fatalf("draft %s carries no annotation, so approval is not observable: %s", relPath, string(data))
		}

		paths = append(paths, relPath)
	}
	return paths
}

// assertPagesUnchanged fails unless every base-relative path still holds the
// content it held before the batch ran.
func assertPagesUnchanged(t *testing.T, kbRoot string, before map[string]string) {
	t.Helper()
	for relPath, want := range before {
		data, err := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
		if err != nil {
			t.Fatalf("read %s: %v", relPath, err)
		}
		if string(data) != want {
			t.Errorf("%s was rewritten although the batch failed:\ngot:  %q\nwant: %q", relPath, string(data), want)
		}
	}
}

// snapshotPages reads the content of every base-relative path.
func snapshotPages(t *testing.T, kbRoot string, relPaths []string) map[string]string {
	t.Helper()

	pages := make(map[string]string, len(relPaths))
	for _, relPath := range relPaths {
		data, err := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
		if err != nil {
			t.Fatalf("read %s: %v", relPath, err)
		}
		pages[relPath] = string(data)
	}
	return pages
}

// TestApproveAllDraftsApprovesNothingWhenACandidateFails pins the batch approve
// as all-or-nothing: when one candidate in the draft set is rejected, no draft
// is approved and no approval commit lands. The rejected candidate is an
// escaping symlink that sorts after the drafts, so a walk that approves as it
// goes would already have rewritten and committed them.
func TestApproveAllDraftsApprovesNothingWhenACandidateFails(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	before := snapshotPages(t, kbRoot, approveDraftsFixture(t, kbRoot, "alpha.md", "beta.md"))

	const outsideContent = "---\ntype: note\ntitle: Secret\nis_draft: true\n---\ncontent outside the base"
	outsideFile := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(outsideFile, []byte(outsideContent), 0600); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, filepath.Join(kbRoot, "kb", "zz-evil.md"), outsideFile)

	out, err := approveAllDraftsRun(kbRoot)
	if err == nil {
		t.Fatalf("expected approve --all-drafts to reject the symlinked candidate, got: %s", out)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("failure %v is not an exit error", err)
	}
	if exitErr.ExitCode() != exitFault {
		t.Errorf("exit code = %d, want %d; output: %s", exitErr.ExitCode(), exitFault, out)
	}

	assertPagesUnchanged(t, kbRoot, before)

	if log := gitLogOneline(t, kbRoot); strings.Contains(log, "akb: approve") {
		t.Errorf("git log carries an approval commit although the run failed:\n%s", log)
	}

	data, err := os.ReadFile(outsideFile) //nolint:gosec // test reading a known temp file
	if err != nil {
		t.Fatalf("read the outside file: %v", err)
	}
	if string(data) != outsideContent {
		t.Errorf("outside file = %q, want it untouched by approve", string(data))
	}
}

// TestApproveAllDraftsApprovesNothingWhenFrontmatterDoesNotParse pins the other
// validation the batch runs over its candidates: a page whose frontmatter does
// not parse fails the run before any draft is approved.
func TestApproveAllDraftsApprovesNothingWhenFrontmatterDoesNotParse(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	before := snapshotPages(t, kbRoot, approveDraftsFixture(t, kbRoot, "alpha.md"))
	writeRawPage(t, kbRoot, "kb/zz-broken.md", "---\ntype: [unclosed\n---\nbroken frontmatter")

	out, err := approveAllDraftsRun(kbRoot)
	if err == nil {
		t.Fatalf("expected approve --all-drafts to reject the unparseable candidate, got: %s", out)
	}
	if !strings.Contains(out, "zz-broken.md") {
		t.Errorf("output = %q, want it to name the rejected page", out)
	}

	assertPagesUnchanged(t, kbRoot, before)

	if log := gitLogOneline(t, kbRoot); strings.Contains(log, "akb: approve") {
		t.Errorf("git log carries an approval commit although the run failed:\n%s", log)
	}
}

// TestApproveAllDraftsApprovesEveryValidDraft pins that a draft set which
// validates fully is still approved page by page: every draft is rewritten with
// is_draft: false and its annotation stripped, each approval lands its own
// commit, and the run reports the count.
func TestApproveAllDraftsApprovesEveryValidDraft(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	paths := approveDraftsFixture(t, kbRoot, "alpha.md", "beta.md")

	out, err := approveAllDraftsRun(kbRoot)
	if err != nil {
		t.Fatalf("approve --all-drafts failed: %s: %v", out, err)
	}
	if !strings.Contains(out, "Approved 2 drafts") {
		t.Errorf("output = %q, want the approval count", out)
	}

	for _, relPath := range paths {
		data, err := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
		if err != nil {
			t.Fatalf("read %s: %v", relPath, err)
		}
		if !strings.Contains(string(data), "is_draft: false") {
			t.Errorf("%s = %q, want is_draft: false after approval", relPath, string(data))
		}
		if strings.Contains(string(data), "olw-auto") {
			t.Errorf("%s = %q, want the annotation stripped", relPath, string(data))
		}
	}

	log := gitLogOneline(t, kbRoot)
	for _, page := range []string{"notes/alpha.md", "notes/beta.md"} {
		if !strings.Contains(log, "akb: approve "+page) {
			t.Errorf("git log is missing the approval commit of %s:\n%s", page, log)
		}
	}
}

// TestApproveRefusesDraftMissingRequiredField pins the approval gate: a draft
// that leaves a schema-required frontmatter field unset is not approved, the
// missing fields are reported the way the write path reports them, and the run
// exits 1 as a validation failure. The planted page never passed write
// validation, so it models the pre-existing pages the gate has to cover.
func TestApproveRefusesDraftMissingRequiredField(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const relPath = "kb/decisions/planted.md"
	writeRawPage(t, kbRoot, relPath, adrMissingStatus)

	out, err := approveRun(kbRoot, "decisions/planted.md")
	if err == nil {
		t.Fatalf("expected approve to refuse the incomplete draft, got: %s", out)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %T: %v", err, err)
	}
	if code := exitErr.ExitCode(); code != exitFailure {
		t.Errorf("exit code = %d, want %d; output: %s", code, exitFailure, out)
	}
	if !strings.Contains(out, "cannot approve 'decisions/planted.md'") {
		t.Errorf("output = %q, want it to name the page", out)
	}
	if want := `missing required frontmatter field(s): status (declared required by template "adr")`; !strings.Contains(out, want) {
		t.Errorf("output = %q, want %q", out, want)
	}

	data, readErr := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
	if readErr != nil {
		t.Fatalf("read the planted page: %v", readErr)
	}
	if string(data) != adrMissingStatus {
		t.Errorf("refused page changed:\nbefore: %q\nafter:  %q", adrMissingStatus, data)
	}

	if log := gitLogOneline(t, kbRoot); strings.Contains(log, "akb: approve") {
		t.Errorf("git log carries an approval commit although the draft was refused:\n%s", log)
	}
}

// TestApproveAllDraftsRefusesOnlyIncompleteDrafts pins the batch semantics of
// the gate: the drafts that pass are approved, a draft the gate refuses keeps
// its draft state and its content, and the run fails so the refusal is never
// reported as part of a successful batch.
func TestApproveAllDraftsRefusesOnlyIncompleteDrafts(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	approveDraftsFixture(t, kbRoot, "alpha.md")

	const refusedRel = "kb/decisions/planted.md"
	writeRawPage(t, kbRoot, refusedRel, adrMissingStatus)
	before := snapshotPages(t, kbRoot, []string{refusedRel})

	out, err := approveAllDraftsRun(kbRoot)
	if err == nil {
		t.Fatalf("expected approve --all-drafts to fail on the refused draft, got: %s", out)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %T: %v", err, err)
	}
	if code := exitErr.ExitCode(); code != exitFailure {
		t.Errorf("exit code = %d, want %d; output: %s", code, exitFailure, out)
	}
	if !strings.Contains(out, "Approved 1 drafts") {
		t.Errorf("output = %q, want the count of the approved drafts", out)
	}
	if !strings.Contains(out, "1 draft(s) not approved") {
		t.Errorf("output = %q, want the count of the refused drafts", out)
	}
	if !strings.Contains(out, "cannot approve 'decisions/planted.md'") {
		t.Errorf("output = %q, want it to name the refused page", out)
	}
	if want := `missing required frontmatter field(s): status (declared required by template "adr")`; !strings.Contains(out, want) {
		t.Errorf("output = %q, want %q", out, want)
	}

	approvedData, readErr := os.ReadFile(filepath.Join(kbRoot, "kb", "notes", "alpha.md")) //nolint:gosec // test reading a known temp file
	if readErr != nil {
		t.Fatalf("read the approved draft: %v", readErr)
	}
	if !strings.Contains(string(approvedData), "is_draft: false") {
		t.Errorf("approved draft = %q, want is_draft: false", string(approvedData))
	}
	assertPagesUnchanged(t, kbRoot, before)

	log := gitLogOneline(t, kbRoot)
	if !strings.Contains(log, "akb: approve notes/alpha.md") {
		t.Errorf("git log is missing the approval commit of the passing draft:\n%s", log)
	}
	if strings.Contains(log, "akb: approve decisions/planted.md") {
		t.Errorf("git log carries an approval commit of the refused draft:\n%s", log)
	}
}

// TestApproveAcceptsPageTypeWithoutTemplate pins the boundary of the gate: a
// page whose type has no template declares no required fields, so its approval
// is left to the type checks rather than refused here.
func TestApproveAcceptsPageTypeWithoutTemplate(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const relPath = "kb/notes/legacy.md"
	content := "---\ntype: retired-type\ntitle: Legacy Page\n---\nContent of a page whose type has no template."
	writeRawPage(t, kbRoot, relPath, content)

	out, err := approveRun(kbRoot, "notes/legacy.md")
	if err != nil {
		t.Fatalf("approve failed: %s: %v", out, err)
	}
	if !strings.Contains(out, "Approved 'notes/legacy.md'") {
		t.Errorf("output = %q, want the approval message", out)
	}

	data, readErr := os.ReadFile(filepath.Join(kbRoot, filepath.FromSlash(relPath))) //nolint:gosec // test reading a known temp file
	if readErr != nil {
		t.Fatalf("read the approved page: %v", readErr)
	}
	if !strings.Contains(string(data), "is_draft: false") {
		t.Errorf("approved page = %q, want is_draft: false", string(data))
	}
}
