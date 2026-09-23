package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
