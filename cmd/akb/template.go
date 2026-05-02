package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
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
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	templates, err := template.LoadTemplates(filepath.Join(kbRoot, ".akb", "templates"))
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	name := args[0]
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("template %q not found. Run `akb template list` to see available templates", name)
	}

	if templateExample {
		passPath := filepath.Join(kbRoot, ".akb", "templates", name+"_pass.md")
		data, err := os.ReadFile(passPath)
		if err != nil {
			return fmt.Errorf("pass mockup not found for template %q: %w", name, err)
		}
		fmt.Print(string(data))
		return nil
	}

	if templateFull {
		data, err := os.ReadFile(filepath.Join(kbRoot, ".akb", "templates", name+".yaml"))
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
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	templates, err := template.LoadTemplates(filepath.Join(kbRoot, ".akb", "templates"))
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
