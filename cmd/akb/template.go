package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/cel-go/common/types"
	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/cel"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/template"
)

var templateExample bool
var templateFull bool

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
	templateCmd.AddCommand(templateGetCmd)
	templateCmd.AddCommand(templateListCmd)
	RootCmd.AddCommand(templateCmd)
}

func runTemplateGet(_ *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := path.AssertContained(kbRoot, templatesDir); err != nil {
		return fmt.Errorf("resolve templates directory: %w", err)
	}
	if err := assertTemplateFilesContained(kbRoot, templatesDir); err != nil {
		return fmt.Errorf("resolve templates directory: %w", err)
	}
	templates, err := template.LoadTemplates(templatesDir)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	name := args[0]
	if !templateNameRe.MatchString(name) {
		return fmt.Errorf("invalid template name %q: must contain only letters, numbers, hyphens, and underscores", name)
	}
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("template %q not found. Run `akb template list` to see available templates", name)
	}

	if templateExample {
		passPath := filepath.Join(templatesDir, name+"_pass.md")
		if err := path.AssertContained(kbRoot, passPath); err != nil {
			return fmt.Errorf("resolve template path: %w", err)
		}
		data, err := os.ReadFile(passPath) //nolint:gosec // name validated by templateNameRe
		if err != nil {
			return fmt.Errorf("pass mockup not found for template %q: %w", name, err)
		}

		passFM, passBody, fmErr := frontmatter.Parse(data)
		if fmErr == nil {
			passPage := buildTestPage(fmt.Sprintf("kb/%s_pass.md", name), passFM, passBody)
			celEnv, celErr := cel.NewEnv()
			if celErr == nil {
				var failedRules []string
				for _, rule := range tmpl.Validations {
					prg, compErr := cel.CompileRule(celEnv, rule.Rule)
					if compErr != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: CEL compile error in rule '%s': %v\n", rule.ID, compErr)
						continue
					}
					result, evalErr := cel.Evaluate(context.Background(), prg, map[string]any{
						"page":     passPage,
						"old_page": nil,
						"now":      time.Now(),
					})
					if evalErr != nil {
						fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: CEL evaluate error in rule '%s': %v\n", rule.ID, evalErr)
						continue
					}
					if result != types.True {
						failedRules = append(failedRules, rule.ID)
					}
				}
				if len(failedRules) > 0 {
					fmt.Fprintf(os.Stderr, "WARNING: This mockup no longer passes validation against the current template rules.\n         Failed rule(s): %v\n         The template may have been changed without updating the mockup.\n", failedRules)
				}
			} else {
				fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: %v\n", celErr)
			}
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Could not validate mockup: %v\n", fmErr)
		}

		fmt.Print(string(data))
		return nil
	}

	if templateFull {
		fullPath := filepath.Join(templatesDir, name+".yaml")
		if err := path.AssertContained(kbRoot, fullPath); err != nil {
			return fmt.Errorf("resolve template path: %w", err)
		}
		data, err := os.ReadFile(fullPath) //nolint:gosec // name validated by templateNameRe
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

func runTemplateList(_ *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := path.AssertContained(kbRoot, templatesDir); err != nil {
		return fmt.Errorf("resolve templates directory: %w", err)
	}
	if err := assertTemplateFilesContained(kbRoot, templatesDir); err != nil {
		return fmt.Errorf("resolve templates directory: %w", err)
	}
	templates, err := template.LoadTemplates(templatesDir)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}

	for name, tmpl := range templates {
		fmt.Printf("%s: %s\n", name, tmpl.Description)
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
