package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

func setupTemplatesWriteTestKB(t *testing.T) string {
	t.Helper()
	kbRoot := t.TempDir()
	setupTestKBWithGit(t, kbRoot)

	if err := os.MkdirAll(filepath.Join(kbRoot, ".akb", "templates"), 0750); err != nil {
		t.Fatal(err)
	}

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origCwd) }) //nolint:errcheck,gosec // test cleanup

	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	return kbRoot
}

func TestTemplatesWrite_AcceptValid(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	tmplPath := filepath.Join(kbRoot, "new.yaml")
	tmplData := `name: new
description: A new template
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
`
	if err := os.WriteFile(tmplPath, []byte(tmplData), 0600); err != nil {
		t.Fatal(err)
	}

	passPath := filepath.Join(kbRoot, "new_pass.md")
	validData := `---
type: new
title: Hello
---
# Hello
`
	if err := os.WriteFile(passPath, []byte(validData), 0600); err != nil {
		t.Fatal(err)
	}

	failPath := filepath.Join(kbRoot, "new_fail.md")
	invalidData := `---
type: new
title: ""
---
# Empty
`
	if err := os.WriteFile(failPath, []byte(invalidData), 0600); err != nil {
		t.Fatal(err)
	}

	twTemplate = tmplPath
	twPass = passPath
	twFail = failPath
	defer func() {
		twTemplate = ""
		twPass = ""
		twFail = ""
	}()

	if err := runTemplatesWrite(nil, []string{"new"}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"new.yaml", "new_pass.md", "new_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestTemplatesWrite_RejectCELSyntaxError(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	tmplPath := filepath.Join(kbRoot, "bad.yaml")
	tmplData := `name: bad
description: Bad template
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: bad_rule
    rule: 'page.frontmatter.title =='
    expect: syntax error
`
	if err := os.WriteFile(tmplPath, []byte(tmplData), 0600); err != nil {
		t.Fatal(err)
	}

	passPath := filepath.Join(kbRoot, "bad_pass.md")
	validData := `---
type: bad
title: Hello
---
# Hello
`
	if err := os.WriteFile(passPath, []byte(validData), 0600); err != nil {
		t.Fatal(err)
	}

	failPath := filepath.Join(kbRoot, "bad_fail.md")
	invalidData := `---
type: bad
title: ""
---
# Empty
`
	if err := os.WriteFile(failPath, []byte(invalidData), 0600); err != nil {
		t.Fatal(err)
	}

	twTemplate = tmplPath
	twPass = passPath
	twFail = failPath
	defer func() {
		twTemplate = ""
		twPass = ""
		twFail = ""
	}()

	err := runTemplatesWrite(nil, []string{"bad"})
	if err == nil {
		t.Fatal("expected error for CEL syntax error, got nil")
	}
	if !strings.Contains(err.Error(), "compile") {
		t.Errorf("expected compile error, got: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"bad.yaml", "bad_pass.md", "bad_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); !os.IsNotExist(err) {
			t.Errorf("expected file %s to NOT exist", f)
		}
	}
}

func TestTemplatesWrite_RejectPassFailsRules(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	tmplPath := filepath.Join(kbRoot, "badpass.yaml")
	tmplData := `name: badpass
description: Bad pass
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
`
	if err := os.WriteFile(tmplPath, []byte(tmplData), 0600); err != nil {
		t.Fatal(err)
	}

	passPath := filepath.Join(kbRoot, "badpass_pass.md")
	validData := `---
type: badpass
title: ""
---
# Empty
`
	if err := os.WriteFile(passPath, []byte(validData), 0600); err != nil {
		t.Fatal(err)
	}

	failPath := filepath.Join(kbRoot, "badpass_fail.md")
	invalidData := `---
type: badpass
title: Hello
---
# Hello
`
	if err := os.WriteFile(failPath, []byte(invalidData), 0600); err != nil {
		t.Fatal(err)
	}

	twTemplate = tmplPath
	twPass = passPath
	twFail = failPath
	defer func() {
		twTemplate = ""
		twPass = ""
		twFail = ""
	}()

	err := runTemplatesWrite(nil, []string{"badpass"})
	if err == nil {
		t.Fatal("expected error when pass mockup fails rules, got nil")
	}
	if !strings.Contains(err.Error(), "pass mockup") {
		t.Errorf("expected pass mockup error, got: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"badpass.yaml", "badpass_pass.md", "badpass_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); !os.IsNotExist(err) {
			t.Errorf("expected file %s to NOT exist", f)
		}
	}
}

func TestTemplatesWrite_RejectFailPassesRules(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	tmplPath := filepath.Join(kbRoot, "badfail.yaml")
	tmplData := `name: badfail
description: Bad fail
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
`
	if err := os.WriteFile(tmplPath, []byte(tmplData), 0600); err != nil {
		t.Fatal(err)
	}

	passPath := filepath.Join(kbRoot, "badfail_pass.md")
	validData := `---
type: badfail
title: Hello
---
# Hello
`
	if err := os.WriteFile(passPath, []byte(validData), 0600); err != nil {
		t.Fatal(err)
	}

	failPath := filepath.Join(kbRoot, "badfail_fail.md")
	invalidData := `---
type: badfail
title: Hello
---
# Hello
`
	if err := os.WriteFile(failPath, []byte(invalidData), 0600); err != nil {
		t.Fatal(err)
	}

	twTemplate = tmplPath
	twPass = passPath
	twFail = failPath
	defer func() {
		twTemplate = ""
		twPass = ""
		twFail = ""
	}()

	err := runTemplatesWrite(nil, []string{"badfail"})
	if err == nil {
		t.Fatal("expected error when fail mockup passes all rules, got nil")
	}
	if !strings.Contains(err.Error(), "fail mockup") {
		t.Errorf("expected fail mockup error, got: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"badfail.yaml", "badfail_pass.md", "badfail_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); !os.IsNotExist(err) {
			t.Errorf("expected file %s to NOT exist", f)
		}
	}
}

func TestTemplatesWrite_AtomicWrite(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	tmplPath := filepath.Join(kbRoot, "atomic.yaml")
	tmplData := `name: atomic
description: Atomic test
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
`
	if err := os.WriteFile(tmplPath, []byte(tmplData), 0600); err != nil {
		t.Fatal(err)
	}

	passPath := filepath.Join(kbRoot, "atomic_pass.md")
	validData := `---
type: atomic
title: ""
---
# Empty
`
	if err := os.WriteFile(passPath, []byte(validData), 0600); err != nil {
		t.Fatal(err)
	}

	failPath := filepath.Join(kbRoot, "atomic_fail.md")
	invalidData := `---
type: atomic
title: Hello
---
# Hello
`
	if err := os.WriteFile(failPath, []byte(invalidData), 0600); err != nil {
		t.Fatal(err)
	}

	twTemplate = tmplPath
	twPass = passPath
	twFail = failPath
	defer func() {
		twTemplate = ""
		twPass = ""
		twFail = ""
	}()

	if err := runTemplatesWrite(nil, []string{"atomic"}); err == nil {
		t.Fatal("expected error, got nil")
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "atomic") {
			t.Errorf("expected no atomic files, found: %s", entry.Name())
		}
	}
}

// TestTemplatesWriteCommitIsScopedToItsFiles pins that the template write
// commit records only the template and its mockups: a change the surrounding
// repository staged stays staged.
func TestTemplatesWriteCommitIsScopedToItsFiles(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)
	mustGitInDir(t, kbRoot, "commit", "-m", "initial")

	foreign := filepath.Join(kbRoot, "src", "app.go")
	if err := os.MkdirAll(filepath.Dir(foreign), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(foreign, []byte("package main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	mustGitInDir(t, kbRoot, "add", "--", "src/app.go")

	templatePath := filepath.Join(kbRoot, "scoped.yaml")
	templateBody := `name: scoped
description: Scoped template
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
`
	if err := os.WriteFile(templatePath, []byte(templateBody), 0600); err != nil {
		t.Fatal(err)
	}

	okMockupPath := filepath.Join(kbRoot, "scoped_ok.md")
	okMockupBody := `---
type: scoped
title: Hello
---
# Hello
`
	if err := os.WriteFile(okMockupPath, []byte(okMockupBody), 0600); err != nil {
		t.Fatal(err)
	}

	badMockupPath := filepath.Join(kbRoot, "scoped_bad.md")
	badMockupBody := `---
type: scoped
title: ""
---
# Empty
`
	if err := os.WriteFile(badMockupPath, []byte(badMockupBody), 0600); err != nil {
		t.Fatal(err)
	}

	twTemplate = templatePath
	twPass = okMockupPath
	twFail = badMockupPath
	defer func() {
		twTemplate = ""
		twPass = ""
		twFail = ""
	}()

	if err := runTemplatesWrite(nil, []string{"scoped"}); err != nil {
		t.Fatalf("template write failed: %v", err)
	}

	want := []string{
		".akb/templates/scoped.yaml",
		".akb/templates/scoped_fail.md",
		".akb/templates/scoped_pass.md",
	}
	if files := commitFilesIn(t, kbRoot); !reflect.DeepEqual(files, want) {
		t.Errorf("commit recorded %v, want only the template files", files)
	}

	status := mustGitInDir(t, kbRoot, "status", "--porcelain")
	if !strings.Contains(status, "A  src/app.go") {
		t.Errorf("staged change lost from the index:\n%s", status)
	}
}

// TestTemplatesWrite_RejectsSymlinkedTemplatesDir pins that writing a template
// rejects a templates directory that is a symlink out of the base before it
// writes anything there.
func TestTemplatesWrite_RejectsSymlinkedTemplatesDir(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	outsideDir := t.TempDir()
	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := os.RemoveAll(templatesDir); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, templatesDir, outsideDir)

	templatePath := filepath.Join(kbRoot, "outside-template.yaml")
	templateBody := `name: outside
description: A template
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
`
	if err := os.WriteFile(templatePath, []byte(templateBody), 0600); err != nil {
		t.Fatal(err)
	}

	origTemplate, origPass, origFail := twTemplate, twPass, twFail
	twTemplate, twPass, twFail = templatePath, "", ""
	t.Cleanup(func() { twTemplate, twPass, twFail = origTemplate, origPass, origFail })

	assertSymlinkEscape(t, runTemplatesWrite(nil, []string{"outside"}))

	entries, err := os.ReadDir(outsideDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("templates outside the base were written: %v", entries)
	}
}

// TestTemplatesWrite_RejectsSymlinkedMockup pins that a mockup read back from
// the templates directory is rejected when it is a symlink out of the base:
// the outside content is neither validated nor echoed.
func TestTemplatesWrite_RejectsSymlinkedMockup(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	const outsideSecret = "OUTSIDE_PASS_SECRET"
	outsidePass := filepath.Join(t.TempDir(), "store_pass.md")
	outsideContent := "---\ntype: store\ntitle: \"\"\n---\n" + outsideSecret + "\n"
	if err := os.WriteFile(outsidePass, []byte(outsideContent), 0600); err != nil {
		t.Fatal(err)
	}

	templatePath := filepath.Join(kbRoot, "store.yaml")
	templateBody := `name: store
description: Stored template
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
`
	if err := os.WriteFile(templatePath, []byte(templateBody), 0600); err != nil {
		t.Fatal(err)
	}

	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := os.WriteFile(filepath.Join(templatesDir, "store.yaml"), []byte(templateBody), 0600); err != nil {
		t.Fatal(err)
	}
	symlinkFixture(t, filepath.Join(templatesDir, "store_pass.md"), outsidePass)

	origTemplate, origPass, origFail := twTemplate, twPass, twFail
	twTemplate, twPass, twFail = templatePath, "", ""
	t.Cleanup(func() { twTemplate, twPass, twFail = origTemplate, origPass, origFail })

	var runErr error
	out := captureStdout(t, func() { runErr = runTemplatesWrite(nil, []string{"store"}) })

	assertSymlinkEscape(t, runErr)
	if strings.Contains(out, outsideSecret) {
		t.Errorf("outside mockup content reached stdout: %q", out)
	}
	if strings.Contains(runErr.Error(), outsideSecret) {
		t.Errorf("outside mockup content reached the error report: %q", runErr)
	}

	data, err := os.ReadFile(outsidePass) //nolint:gosec // test reading a known temp file
	if err != nil {
		t.Fatalf("read the outside file: %v", err)
	}
	if string(data) != outsideContent {
		t.Errorf("outside file = %q, want it untouched by templates write", string(data))
	}
}

// writeTemplatesWriteFixture writes a template and its mockups beside the KB and
// points the template write flags at them.
func writeTemplatesWriteFixture(t *testing.T, kbRoot, name, templateBody, passBody, failBody string) {
	t.Helper()

	write := func(path, body string) {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}

	tmplPath := filepath.Join(kbRoot, name+".yaml")
	passPath := filepath.Join(kbRoot, name+"_pass.md")
	failPath := filepath.Join(kbRoot, name+"_fail.md")
	write(tmplPath, templateBody)
	write(passPath, passBody)
	write(failPath, failBody)

	origTemplate, origPass, origFail := twTemplate, twPass, twFail
	twTemplate, twPass, twFail = tmplPath, passPath, failPath
	t.Cleanup(func() { twTemplate, twPass, twFail = origTemplate, origPass, origFail })
}

// TestTemplatesWrite_RejectsUnguardedOptionalKeyRead pins that a rule reading a
// schema-optional key without has() is rejected: the pass mockup supplies the
// key, but a page without it must validate too.
func TestTemplatesWrite_RejectsUnguardedOptionalKeyRead(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	writeTemplatesWriteFixture(t, kbRoot, "summary",
		`name: summary
description: Template with an optional summary
schema:
  frontmatter:
    title:
      type: string
      required: true
    summary:
      type: string
      required: false
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
  - id: summary_non_empty
    rule: 'page.frontmatter.summary != ""'
    expect: summary must not be empty when present
`,
		`---
type: summary
title: Hello
summary: A summary
---
# Hello
`,
		`---
type: summary
title: ""
summary: A summary
---
# Empty
`)

	err := runTemplatesWrite(nil, []string{"summary"})
	if err == nil {
		t.Fatal("expected error for an unguarded optional-key read, got nil")
	}
	for _, want := range []string{
		"rule summary_non_empty errored when optional key summary was absent from the pass mockup",
		"no such key: summary",
		"guard the access with has() or mark summary required: true in the schema",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"summary.yaml", "summary_pass.md", "summary_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); !os.IsNotExist(err) {
			t.Errorf("expected file %s to NOT exist", f)
		}
	}
}

// TestTemplatesWrite_AcceptsGuardedOptionalKeyRead pins that the has()-guarded
// equivalent of the rejected rule validates for a page that omits the key.
func TestTemplatesWrite_AcceptsGuardedOptionalKeyRead(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	writeTemplatesWriteFixture(t, kbRoot, "guarded",
		`name: guarded
description: Template with a guarded optional summary
schema:
  frontmatter:
    title:
      type: string
      required: true
    summary:
      type: string
      required: false
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
  - id: summary_non_empty
    rule: '!has(page.frontmatter.summary) || page.frontmatter.summary != ""'
    expect: summary must not be empty when present
`,
		`---
type: guarded
title: Hello
summary: A summary
---
# Hello
`,
		`---
type: guarded
title: ""
summary: A summary
---
# Empty
`)

	if err := runTemplatesWrite(nil, []string{"guarded"}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"guarded.yaml", "guarded_pass.md", "guarded_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

// TestTemplatesWrite_RejectsCreateOnlyOldPageRule pins that the pass mockup is
// also evaluated as an update of itself: a rule that only holds while old_page
// is absent is rejected.
func TestTemplatesWrite_RejectsCreateOnlyOldPageRule(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	writeTemplatesWriteFixture(t, kbRoot, "createonly",
		`name: createonly
description: Template with a create-only rule
schema:
  frontmatter:
    title:
      type: string
      required: true
    supersedes:
      type: string
      required: false
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
  - id: supersedes_stable
    rule: '!has(old_page.frontmatter) || old_page.frontmatter.supersedes == page.frontmatter.supersedes'
    expect: supersedes must not change
`,
		`---
type: createonly
title: Hello
---
# Hello
`,
		`---
type: createonly
title: ""
---
# Empty
`)

	err := runTemplatesWrite(nil, []string{"createonly"})
	if err == nil {
		t.Fatal("expected error for a create-only old_page rule, got nil")
	}
	for _, want := range []string{
		"rule supersedes_stable errored when the pass mockup was evaluated against itself as old_page",
		"no such key: supersedes",
		"guard the access with has()",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"createonly.yaml", "createonly_pass.md", "createonly_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); !os.IsNotExist(err) {
			t.Errorf("expected file %s to NOT exist", f)
		}
	}
}

// TestTemplatesWrite_ForceDoesNotBypassVariantChecks pins that --force skips
// only the overwrite confirmation: the variant checks still reject the write.
func TestTemplatesWrite_ForceDoesNotBypassVariantChecks(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)
	templatesDir := filepath.Join(kbRoot, ".akb", "templates")

	templateBody := `name: forced
description: Template with an unguarded optional read
schema:
  frontmatter:
    title:
      type: string
      required: true
    summary:
      type: string
      required: false
validations:
  - id: summary_non_empty
    rule: 'page.frontmatter.summary != ""'
    expect: summary must not be empty when present
`
	passBody := `---
type: forced
title: Hello
summary: A summary
---
# Hello
`
	failBody := `---
type: forced
title: Hello
---
# Hello
`
	if err := os.WriteFile(filepath.Join(templatesDir, "forced.yaml"), []byte(templateBody), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templatesDir, "forced_pass.md"), []byte(passBody), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templatesDir, "forced_fail.md"), []byte(failBody), 0600); err != nil {
		t.Fatal(err)
	}

	writeTemplatesWriteFixture(t, kbRoot, "forced", templateBody, passBody, failBody)

	origForce := twForce
	twForce = true
	t.Cleanup(func() { twForce = origForce })

	err := runTemplatesWrite(nil, []string{"forced"})
	if err == nil {
		t.Fatal("expected the variant check to reject the write despite --force")
	}
	if !strings.Contains(err.Error(), "errored when optional key summary was absent from the pass mockup") {
		t.Errorf("expected the stripped-variant error, got: %v", err)
	}

	data, readErr := os.ReadFile(filepath.Join(templatesDir, "forced.yaml")) //nolint:gosec // test reading a known temp file
	if readErr != nil {
		t.Fatalf("read the on-disk template: %v", readErr)
	}
	if string(data) != templateBody {
		t.Errorf("template on disk = %q, want it untouched by the rejected write", string(data))
	}
}

// TestOptionalKeysSuppliedByMockup pins that only schema-optional keys the
// mockup actually sets get a stripped variant, in a stable order, and that
// type and title never do even when the schema declares them optional.
func TestOptionalKeysSuppliedByMockup(t *testing.T) {
	tmpl := template.Template{
		Name: "unit",
		Schema: template.Schema{Frontmatter: map[string]template.FieldSchema{
			"type":   {Type: "string"},
			"title":  {Type: "string"},
			"zeta":   {Type: "string"},
			"alpha":  {Type: "string"},
			"absent": {Type: "string"},
		}},
	}
	fm := &frontmatter.ParsedFrontmatter{
		Type:   "unit",
		Title:  "Hello",
		Fields: map[string]any{"zeta": "z", "alpha": "a"},
	}

	got := optionalKeysSuppliedByMockup(&tmpl, fm)
	want := []string{"alpha", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("optionalKeysSuppliedByMockup = %v, want %v", got, want)
	}
}

// TestTemplatesWrite_AcceptsSchemaOptionalTitleAndType pins that type and title
// are never stripped: cel.BuildPage always injects both keys, so a rule reading
// either holds even when the schema declares it optional.
func TestTemplatesWrite_AcceptsSchemaOptionalTitleAndType(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	writeTemplatesWriteFixture(t, kbRoot, "loose",
		`name: loose
description: Template declaring type and title schema-optional
schema:
  frontmatter:
    type:
      type: string
      required: false
    title:
      type: string
      required: false
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
  - id: has_type
    rule: 'page.frontmatter.type != ""'
    expect: type must not be empty
`,
		`---
type: loose
title: Hello
---
# Hello
`,
		`---
type: loose
title: ""
---
# Empty
`)

	if err := runTemplatesWrite(nil, []string{"loose"}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"loose.yaml", "loose_pass.md", "loose_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

// TestTemplatesWrite_RejectsRuleRequiringOptionalKey pins the false-result side
// of the stripped variants: a rule that demands a schema-optional key is
// rejected because pages without that key must validate.
func TestTemplatesWrite_RejectsRuleRequiringOptionalKey(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	writeTemplatesWriteFixture(t, kbRoot, "requires",
		`name: requires
description: Template whose rule demands an optional key
schema:
  frontmatter:
    title:
      type: string
      required: true
    summary:
      type: string
      required: false
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
  - id: summary_required
    rule: 'has(page.frontmatter.summary)'
    expect: summary must be set
`,
		`---
type: requires
title: Hello
summary: A summary
---
# Hello
`,
		`---
type: requires
title: ""
summary: A summary
---
# Empty
`)

	err := runTemplatesWrite(nil, []string{"requires"})
	if err == nil {
		t.Fatal("expected error for a rule that demands an optional key, got nil")
	}
	if !strings.Contains(err.Error(), "pass mockup no longer validates without optional key summary: [summary_required]") {
		t.Errorf("expected the stripped-variant rejection, got: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"requires.yaml", "requires_pass.md", "requires_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); !os.IsNotExist(err) {
			t.Errorf("expected file %s to NOT exist", f)
		}
	}
}

// TestTemplatesWrite_RejectsNoOpUpdateRule pins the false-result side of the
// self-succession variant: a rule that demands a change on every update cannot
// hold for the mockup evaluated against itself.
func TestTemplatesWrite_RejectsNoOpUpdateRule(t *testing.T) {
	kbRoot := setupTemplatesWriteTestKB(t)

	writeTemplatesWriteFixture(t, kbRoot, "noop",
		`name: noop
description: Template that demands a change on every update
schema:
  frontmatter:
    title:
      type: string
      required: true
    updated:
      type: string
      required: false
validations:
  - id: has_title
    rule: 'page.frontmatter.title != ""'
    expect: title must not be empty
  - id: updated_must_change
    rule: '!has(old_page.frontmatter) || !has(old_page.frontmatter.updated) || page.frontmatter.updated != old_page.frontmatter.updated'
    expect: updated must change on every update
`,
		`---
type: noop
title: Hello
updated: 2024-01-01
---
# Hello
`,
		`---
type: noop
title: ""
updated: 2024-01-01
---
# Empty
`)

	err := runTemplatesWrite(nil, []string{"noop"})
	if err == nil {
		t.Fatal("expected error for a rule that rejects a no-op update, got nil")
	}
	if !strings.Contains(err.Error(), "pass mockup no longer validates as an update of itself: [updated_must_change]") {
		t.Errorf("expected the self-succession rejection, got: %v", err)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	for _, f := range []string{"noop.yaml", "noop_pass.md", "noop_fail.md"} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); !os.IsNotExist(err) {
			t.Errorf("expected file %s to NOT exist", f)
		}
	}
}
