// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/goccy/go-yaml"
	gocel "github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/cel"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/template"
)

var templateExample bool
var templateFull bool
var templateExamples bool

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Get template info",
}

var templateGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get template information",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateGet,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE:  runTemplateList,
}

func init() {
	templateGetCmd.Flags().BoolVar(&templateExample, "example", false, "Show pass mockup")
	templateGetCmd.Flags().BoolVar(&templateFull, "full", false, "Show complete template YAML")
	templateGetCmd.Flags().BoolVar(&templateExamples, "examples", false, "Read the embedded showcase templates instead of the knowledge base")
	templateListCmd.Flags().BoolVar(&templateExamples, "examples", false, "List the embedded showcase templates instead of the knowledge base")
	templateCmd.AddCommand(templateGetCmd)
	templateCmd.AddCommand(templateListCmd)
	RootCmd.AddCommand(templateCmd)
}

// resolveTemplateSource loads the templates a get or list command reads and
// returns a reader for the companion files beside them (the template YAML and
// the pass mockup). With --examples both come from the embedded showcase set
// and no KB is resolved, so the commands work without a base selected.
func resolveTemplateSource() (map[string]template.Template, func(string) ([]byte, error), error) {
	if templateExamples {
		templates, err := loadExampleTemplates()
		if err != nil {
			return nil, nil, err
		}
		readFile := func(name string) ([]byte, error) {
			return fs.ReadFile(template.DefaultTemplates, name)
		}
		return templates, readFile, nil
	}

	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve knowledge base: %w", err)
	}

	templatesDir := path.TemplatesDir(kbRoot)
	if err := path.AssertContained(kbRoot, templatesDir); err != nil {
		return nil, nil, fmt.Errorf("resolve templates directory: %w", err)
	}
	if err := assertTemplateFilesContained(kbRoot, templatesDir); err != nil {
		return nil, nil, fmt.Errorf("resolve templates directory: %w", err)
	}
	templates, err := template.LoadTemplates(templatesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("load templates: %w", err)
	}

	readFile := func(name string) ([]byte, error) {
		filePath := filepath.Join(templatesDir, name)
		if err := path.AssertContained(kbRoot, filePath); err != nil {
			return nil, fmt.Errorf("resolve template path: %w", err)
		}
		return os.ReadFile(filePath) //nolint:gosec // name validated by templateNameRe
	}
	return templates, readFile, nil
}

// loadExampleTemplates reads the embedded showcase templates, the copy source
// for a KB that has none of its own.
func loadExampleTemplates() (map[string]template.Template, error) {
	entries, err := fs.ReadDir(template.DefaultTemplates, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded templates: %w", err)
	}

	templates := make(map[string]template.Template)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		data, err := fs.ReadFile(template.DefaultTemplates, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read embedded %s: %w", entry.Name(), err)
		}

		var tmpl template.Template
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			return nil, fmt.Errorf("parse embedded %s: %w", entry.Name(), err)
		}
		templates[tmpl.Name] = tmpl
	}

	return templates, nil
}

func runTemplateGet(_ *cobra.Command, args []string) error {
	name := args[0]
	if !templateNameRe.MatchString(name) {
		return fmt.Errorf("invalid template name %q: must contain only letters, numbers, hyphens, and underscores", name)
	}

	templates, readFile, err := resolveTemplateSource()
	if err != nil {
		return err
	}

	tmpl, ok := templates[name]
	if !ok {
		if templateExamples {
			return fmt.Errorf("template %q not found in the embedded examples. Run `akb template list --examples` to see them", name)
		}
		return fmt.Errorf("template %q not found. Run `akb template list` to see available templates", name)
	}

	if templateExample {
		data, err := readFile(name + "_pass.md")
		if err != nil {
			return fmt.Errorf("pass mockup not found for template %q: %w", name, err)
		}

		warnMockupValidation(tmpl, name, data)

		fmt.Print(string(data))
		return nil
	}

	if templateFull {
		data, err := readFile(name + ".yaml")
		if err != nil {
			return fmt.Errorf("template file not found: %w", err)
		}
		fmt.Print(string(data))
		return nil
	}

	writerView := map[string]any{
		"name":        tmpl.Name,
		"description": tmpl.Description,
		"schema":      tmpl.Schema,
	}

	var requirements []string
	for _, v := range tmpl.Validations {
		if v.Requirement != "" {
			requirements = append(requirements, v.Requirement)
		}
	}
	if len(requirements) > 0 {
		writerView["requirements"] = requirements
	}

	out, err := yaml.Marshal(writerView)
	if err != nil {
		return fmt.Errorf("marshal writer view: %w", err)
	}
	fmt.Print(string(out))
	return nil
}

// warnMockupValidation reports on stderr every way the displayed pass mockup
// fails the checks `akb template write` applies before it records a template:
// the mockup as given, once per schema-optional key it supplies with that key
// removed, and once with the mockup as its own old_page. The display still
// succeeds — a stale mockup is a warning, not a command failure.
func warnMockupValidation(tmpl template.Template, name string, data []byte) {
	passFM, passBody, err := frontmatter.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: %v\n", err)
		return
	}

	celEnv, err := cel.NewEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: %v\n", err)
		return
	}

	passPage := buildTestPage(fmt.Sprintf("kb/%s_pass.md", name), passFM, passBody)
	warnMockupVariant(celEnv, tmpl, passPage, nil,
		"This mockup no longer passes validation against the current template rules.")

	for _, key := range optionalKeysSuppliedByMockup(&tmpl, passFM) {
		strippedPage := buildTestPage(fmt.Sprintf("kb/%s_pass.md", name), withoutFrontmatterKey(passFM, key), passBody)
		warnMockupVariant(celEnv, tmpl, strippedPage, nil,
			fmt.Sprintf("This mockup no longer validates without optional key %s.", key))
	}

	warnMockupVariant(celEnv, tmpl, passPage, passPage,
		"This mockup no longer validates as an update of itself.")
}

// warnMockupVariant evaluates the template's validation rules against one
// variant of the pass mockup and warns on stderr about every rule it breaks:
// a rule that cannot be compiled or evaluated gets its own warning, later rules
// are still checked, and the IDs of the rules that did not evaluate to true are
// reported together in one headline.
func warnMockupVariant(env *gocel.Env, tmpl template.Template, page, oldPage map[string]any, headline string) {
	var failed []string
	for _, rule := range tmpl.Validations {
		prg, err := cel.CompileRule(env, rule.Rule)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: CEL error in rule '%s': %v\n", rule.ID, err)
			continue
		}
		// A nil map is not the same as an absent page: an empty map makes
		// old_page.frontmatter a missing key instead of a null value.
		var oldPageValue any
		if oldPage != nil {
			oldPageValue = oldPage
		}
		result, err := cel.Evaluate(context.Background(), prg, map[string]any{
			"page":     page,
			"old_page": oldPageValue,
			"now":      time.Now(),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: CEL error in rule '%s': %v\n", rule.ID, err)
			continue
		}
		if result != types.True {
			failed = append(failed, rule.ID)
		}
	}
	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n         Failed rule(s): %v\n         The template may have been changed without updating the mockup.\n", headline, failed)
	}
}

func runTemplateList(_ *cobra.Command, _ []string) error {
	templates, _, err := resolveTemplateSource()
	if err != nil {
		return err
	}

	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}

	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("%s: %s\n", name, templates[name].Description)
	}
	return nil
}

// assertTemplateFilesContained checks every entry below templatesDir against
// kbRoot's symlink boundary. The directory check alone accepts a symlinked
// template file inside a real directory, which LoadTemplates then reads. A
// missing templates directory is not an error, matching LoadTemplates' empty
// result for it.
func assertTemplateFilesContained(kbRoot, templatesDir string) error {
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := path.AssertContained(kbRoot, filepath.Join(templatesDir, entry.Name())); err != nil {
			return fmt.Errorf("check template file %s: %w", entry.Name(), err)
		}
	}
	return nil
}
