package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/cel-go/common/types"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"

	"github.com/peedrr/agent-kb/internal/cel"
	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/peedrr/agent-kb/internal/template"
)

var twTemplate string
var twPass string
var twFail string
var twForce bool

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
	templateWriteCmd.Flags().BoolVar(&twForce, "force", false, "overwrite existing template without confirmation")
	_ = templateWriteCmd.MarkFlagRequired("template")
	templateCmd.AddCommand(templateWriteCmd)
}

func runTemplatesWrite(_ *cobra.Command, args []string) error {
	name := args[0]
	if !templateNameRe.MatchString(name) {
		return fmt.Errorf("invalid template name %q: must contain only letters, numbers, hyphens, and underscores", name)
	}

	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	templateYAMLPath := filepath.Join(kbRoot, ".akb", "templates", name+".yaml")
	templateExists := false
	if _, err := os.Stat(templateYAMLPath); err == nil {
		templateExists = true
	}

	if !templateExists {
		if twPass == "" || twFail == "" {
			return fmt.Errorf("--pass and --fail are required for new templates")
		}
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

	var passData []byte
	switch {
	case twPass != "":
		passData, err = os.ReadFile(twPass) //nolint:gosec // path provided by user flag
		if err != nil {
			return fmt.Errorf("read pass mockup: %w", err)
		}
	case templateExists:
		existingPass := filepath.Join(kbRoot, ".akb", "templates", name+"_pass.md")
		passData, err = os.ReadFile(existingPass) //nolint:gosec // known path
		if err != nil {
			return fmt.Errorf("read existing pass mockup: %w", err)
		}
	default:
		return fmt.Errorf("--pass and --fail are required for new templates")
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
		}, cel.MaxCostLimit)
		if err != nil {
			return fmt.Errorf("pass mockup: evaluate rule %q: %w", rule.ID, err)
		}
		if result != types.True {
			passFailed = append(passFailed, rule.ID)
		}
	}
	if len(passFailed) > 0 {
		return fmt.Errorf("pass mockup no longer validates: %v\n\n--- pass mockup ---\n%s\n\nProvide updated mockup with --pass <path>", passFailed, string(passData))
	}

	var failData []byte
	switch {
	case twFail != "":
		failData, err = os.ReadFile(twFail) //nolint:gosec // path provided by user flag
		if err != nil {
			return fmt.Errorf("read fail mockup: %w", err)
		}
	case templateExists:
		existingFail := filepath.Join(kbRoot, ".akb", "templates", name+"_fail.md")
		failData, err = os.ReadFile(existingFail) //nolint:gosec // known path
		if err != nil {
			return fmt.Errorf("read existing fail mockup: %w", err)
		}
	default:
		return fmt.Errorf("--pass and --fail are required for new templates")
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
		}, cel.MaxCostLimit)
		if err != nil {
			return fmt.Errorf("fail mockup: evaluate rule %q: %w", rule.ID, err)
		}
		if result != types.True {
			failFailed = append(failFailed, rule.ID)
		}
	}
	if len(failFailed) == 0 {
		return fmt.Errorf("fail mockup no longer validates: expected at least one validation to fail, but all passed\n\n--- fail mockup ---\n%s\n\nProvide updated mockup with --fail <path>", string(failData))
	}

	if templateExists && !twForce {
		oldYAMLData, err := os.ReadFile(templateYAMLPath) //nolint:gosec // known path
		if err != nil {
			return fmt.Errorf("read existing template: %w", err)
		}

		diff := diffStrings(string(oldYAMLData), string(tmplData))

		pageCount := 0
		dbConn, dbErr := db.OpenKB(kbRoot)
		if dbErr == nil {
			row := dbConn.QueryRow("SELECT COUNT(*) FROM documents WHERE type = ?", name)
			if scanErr := row.Scan(&pageCount); scanErr != nil {
				pageCount = 0
			}
			_ = dbConn.Close() //nolint:errcheck // best effort
		}

		cmdParts := []string{"akb", "template", "write", name, "--template", twTemplate}
		if twPass != "" {
			cmdParts = append(cmdParts, "--pass", twPass)
		}
		if twFail != "" {
			cmdParts = append(cmdParts, "--fail", twFail)
		}
		cmdParts = append(cmdParts, "--force")

		fmt.Printf("Template %q exists and is used by %d pages.\n\n--- diff ---\n%s\n\nRun `%s` to overwrite.\n",
			name, pageCount, diff, strings.Join(cmdParts, " "))
		return fmt.Errorf("template %q exists; use --force to overwrite", name)
	}

	targetDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("create templates directory: %w", err)
	}

	// Hold the repository lock across the swap of the template files and the
	// commit that records it.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	tmpDir, err := os.MkdirTemp(targetDir, ".tmp-write-")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck,gosec // cleanup

	tmpYAML := filepath.Join(tmpDir, name+".yaml")
	tmpPass := filepath.Join(tmpDir, name+"_pass.md")
	tmpFail := filepath.Join(tmpDir, name+"_fail.md")

	if err := os.WriteFile(tmpYAML, tmplData, 0600); err != nil { //nolint:gosec // name validated by templateNameRe, path inside temp dir
		return fmt.Errorf("write temp template: %w", err)
	}
	if err := os.WriteFile(tmpPass, passData, 0600); err != nil { //nolint:gosec // name validated by templateNameRe, path inside temp dir
		return fmt.Errorf("write temp pass mockup: %w", err)
	}
	if err := os.WriteFile(tmpFail, failData, 0600); err != nil { //nolint:gosec // name validated by templateNameRe, path inside temp dir
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

	if !noCommit {
		commitMsg := fmt.Sprintf("akb: template write %s", name)
		if err := storage.CommitFiles(kbRoot, commitMsg, templateCommitPaths(name)...); err != nil {
			return fmt.Errorf("commit template write: %w", err)
		}
	}

	fmt.Printf("Template %q written. Run `akb lint` to evaluate existing pages.\n", name)
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
			return fmt.Errorf("parse %s: template format has changed; please update to the new schema", filename)
		}
	}
	return nil
}

// diffStrings produces a simple aligned diff between two strings,
// showing unchanged lines with "  ", removed lines with "- ", and
// added lines with "+ ".
func diffStrings(a, b string) string {
	alines := strings.Split(a, "\n")
	blines := strings.Split(b, "\n")
	var out strings.Builder
	for i := 0; i < len(alines) || i < len(blines); i++ {
		if i < len(alines) && i < len(blines) && alines[i] == blines[i] {
			out.WriteString("  " + alines[i] + "\n")
		} else {
			if i < len(alines) {
				out.WriteString("- " + alines[i] + "\n")
			}
			if i < len(blines) {
				out.WriteString("+ " + blines[i] + "\n")
			}
		}
	}
	return out.String()
}
