package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/cel-go/common/types"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
	"github.com/peedrr/agent-kb/internal/cel"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/template"
)

var twTemplate string
var twPass string
var twFail string

var templateWriteCmd = &cobra.Command{
	Use:   "write <name>",
	Short: "Write a template with validation",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplatesWrite,
}

func init() {
	templateWriteCmd.Flags().StringVar(&twTemplate, "template", "", "path to template YAML")
	templateWriteCmd.Flags().StringVar(&twPass, "pass", "", "path to pass mockup")
	templateWriteCmd.Flags().StringVar(&twFail, "fail", "", "path to fail mockup")
	_ = templateWriteCmd.MarkFlagRequired("template")
	_ = templateWriteCmd.MarkFlagRequired("pass")
	_ = templateWriteCmd.MarkFlagRequired("fail")
	templateCmd.AddCommand(templateWriteCmd)
}

func runTemplatesWrite(_ *cobra.Command, args []string) error {
	name := args[0]

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	tmplData, err := os.ReadFile(twTemplate) //nolint:gosec // path provided by user flag
	if err != nil {
		return fmt.Errorf("read template file: %w", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(tmplData, &raw); err != nil {
		return fmt.Errorf("parse template YAML: %w", err)
	}
	if err := detectOldFormat(raw, filepath.Base(twTemplate)); err != nil {
		return err
	}

	var tmpl template.Template
	if err := yaml.Unmarshal(tmplData, &tmpl); err != nil {
		return fmt.Errorf("parse template YAML: %w", err)
	}
	if tmpl.Name == "" {
		return fmt.Errorf("missing name field in template YAML")
	}
	if tmpl.Name != name {
		return fmt.Errorf("template name %q does not match argument %q", tmpl.Name, name)
	}

	celEnv, err := cel.NewEnv()
	if err != nil {
		return fmt.Errorf("create CEL environment: %w", err)
	}

	for _, rule := range tmpl.Validations {
		if _, err := cel.CompileRule(celEnv, rule.Rule); err != nil {
			return fmt.Errorf("compile validation rule %q: %w", rule.ID, err)
		}
	}
	for _, rule := range tmpl.LintRules {
		if _, err := cel.CompileRule(celEnv, rule.Rule); err != nil {
			return fmt.Errorf("compile lint rule %q: %w", rule.ID, err)
		}
	}

	passData, err := os.ReadFile(twPass) //nolint:gosec // path provided by user flag
	if err != nil {
		return fmt.Errorf("read pass mockup: %w", err)
	}
	passFM, passBody, err := frontmatter.Parse(passData)
	if err != nil {
		return fmt.Errorf("parse pass mockup frontmatter: %w", err)
	}
	passPage := buildTestPage(fmt.Sprintf("kb/%s_pass.md", name), passFM, passBody)
	var passFailed []string
	for _, rule := range tmpl.Validations {
		prg, err := cel.CompileRule(celEnv, rule.Rule)
		if err != nil {
			return fmt.Errorf("pass mockup: compile rule %q: %w", rule.ID, err)
		}
		result, err := cel.Evaluate(context.Background(), prg, map[string]any{
			"page":     passPage,
			"old_page": nil,
			"now":      time.Now(),
		}, 100000)
		if err != nil {
			return fmt.Errorf("pass mockup: evaluate rule %q: %w", rule.ID, err)
		}
		if result != types.True {
			passFailed = append(passFailed, rule.ID)
		}
	}
	if len(passFailed) > 0 {
		return fmt.Errorf("pass mockup: validation(s) failed: %v", passFailed)
	}

	failData, err := os.ReadFile(twFail) //nolint:gosec // path provided by user flag
	if err != nil {
		return fmt.Errorf("read fail mockup: %w", err)
	}
	failFM, failBody, err := frontmatter.Parse(failData)
	if err != nil {
		return fmt.Errorf("parse fail mockup frontmatter: %w", err)
	}
	failPage := buildTestPage(fmt.Sprintf("kb/%s_fail.md", name), failFM, failBody)
	var failFailed []string
	for _, rule := range tmpl.Validations {
		prg, err := cel.CompileRule(celEnv, rule.Rule)
		if err != nil {
			return fmt.Errorf("fail mockup: compile rule %q: %w", rule.ID, err)
		}
		result, err := cel.Evaluate(context.Background(), prg, map[string]any{
			"page":     failPage,
			"old_page": nil,
			"now":      time.Now(),
		}, 100000)
		if err != nil {
			return fmt.Errorf("fail mockup: evaluate rule %q: %w", rule.ID, err)
		}
		if result != types.True {
			failFailed = append(failFailed, rule.ID)
		}
	}
	if len(failFailed) == 0 {
		return fmt.Errorf("fail mockup: expected at least one validation to fail, but all passed")
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("create templates directory: %w", err)
	}

	tmpDir, err := os.MkdirTemp(targetDir, ".tmp-write-")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck,gosec // cleanup

	tmpYAML := filepath.Join(tmpDir, name+".yaml")
	tmpPass := filepath.Join(tmpDir, name+"_pass.md")
	tmpFail := filepath.Join(tmpDir, name+"_fail.md")

	if err := os.WriteFile(tmpYAML, tmplData, 0600); err != nil {
		return fmt.Errorf("write temp template: %w", err)
	}
	if err := os.WriteFile(tmpPass, passData, 0600); err != nil {
		return fmt.Errorf("write temp pass mockup: %w", err)
	}
	if err := os.WriteFile(tmpFail, failData, 0600); err != nil {
		return fmt.Errorf("write temp fail mockup: %w", err)
	}

	finalYAML := filepath.Join(targetDir, name+".yaml")
	finalPass := filepath.Join(targetDir, name+"_pass.md")
	finalFail := filepath.Join(targetDir, name+"_fail.md")

	if err := os.Rename(tmpYAML, finalYAML); err != nil {
		return fmt.Errorf("write template file: %w", err)
	}
	if err := os.Rename(tmpPass, finalPass); err != nil {
		_ = os.Remove(finalYAML)
		return fmt.Errorf("write pass mockup: %w", err)
	}
	if err := os.Rename(tmpFail, finalFail); err != nil {
		_ = os.Remove(finalYAML)
		_ = os.Remove(finalPass)
		return fmt.Errorf("write fail mockup: %w", err)
	}

	fmt.Printf("Template %q written. Run `akb lint` to evaluate existing pages against new rules.\n", name)
	return nil
}

func buildTestPage(relPath string, fm *frontmatter.ParsedFrontmatter, body []byte) map[string]any {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(body))
	return cel.BuildPage(relPath, fm, body, doc, body)
}

func detectOldFormat(raw map[string]any, filename string) error {
	for _, key := range []string{"required", "optional", "body"} {
		if _, ok := raw[key]; ok {
			return fmt.Errorf("parse %s: Template format has changed. Please update to the new schema.", filename)
		}
	}
	return nil
}
