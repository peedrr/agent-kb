package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	gocel "github.com/google/cel-go/cel"
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

	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := path.AssertContained(kbRoot, templatesDir); err != nil {
		return fmt.Errorf("resolve templates directory: %w", err)
	}

	templateYAMLPath := filepath.Join(templatesDir, name+".yaml")
	if err := path.AssertContained(kbRoot, templateYAMLPath); err != nil {
		return fmt.Errorf("resolve template path: %w", err)
	}
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
		existingPass := filepath.Join(templatesDir, name+"_pass.md")
		if err := path.AssertContained(kbRoot, existingPass); err != nil {
			return fmt.Errorf("resolve pass mockup path: %w", err)
		}
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

	// The pass mockup is the exemplar every page of the type is written from,
	// so it must carry every field the schema marks required. The fail mockup
	// is exempt: it exists to fail a rule, not to model a valid page.
	if missing := checkRequiredFields(tmpl, passFM); len(missing) > 0 {
		return fmt.Errorf("pass mockup %s\n\n--- pass mockup ---\n%s\n\nProvide updated mockup with --pass <path>",
			requiredFieldsMessage(tmpl, missing), string(passData))
	}

	passPage := buildTestPage(fmt.Sprintf("kb/%s_pass.md", name), passFM, passBody)

	passFailed, _, err := evaluateValidations(celEnv, tmpl.Validations, passPage, nil)
	if err != nil {
		return fmt.Errorf("pass mockup: %w", err)
	}
	if len(passFailed) > 0 {
		return fmt.Errorf("pass mockup no longer validates: %v\n\n--- pass mockup ---\n%s\n\nProvide updated mockup with --pass <path>", passFailed, string(passData))
	}

	// A page that omits a schema-optional field must validate too, so the mockup
	// is re-evaluated once per optional key it supplies, with that key removed.
	// Keys the mockup already omits are covered by the evaluation above.
	for _, key := range optionalKeysSuppliedByMockup(&tmpl, passFM) {
		strippedPage := buildTestPage(fmt.Sprintf("kb/%s_pass.md", name), withoutFrontmatterKey(passFM, key), passBody)
		failed, ruleID, err := evaluateValidations(celEnv, tmpl.Validations, strippedPage, nil)
		if err != nil {
			return fmt.Errorf("rule %s errored when optional key %s was absent from the pass mockup:\n%w — guard the access with has() or mark %s required: true in the schema", ruleID, key, err, key)
		}
		if len(failed) > 0 {
			return fmt.Errorf("pass mockup no longer validates without optional key %s: %v\n\n--- pass mockup ---\n%s\n\nProvide updated mockup with --pass <path>", key, failed, string(passData))
		}
	}

	// The mockup must also survive a no-op update, where the page is its own
	// pre-modification state.
	selfFailed, selfRuleID, err := evaluateValidations(celEnv, tmpl.Validations, passPage, passPage)
	if err != nil {
		return fmt.Errorf("rule %s errored when the pass mockup was evaluated against itself as old_page:\n%w — guard the access with has()", selfRuleID, err)
	}
	if len(selfFailed) > 0 {
		return fmt.Errorf("pass mockup no longer validates as an update of itself: %v\n\n--- pass mockup ---\n%s\n\nProvide updated mockup with --pass <path>", selfFailed, string(passData))
	}

	var failData []byte
	switch {
	case twFail != "":
		failData, err = os.ReadFile(twFail) //nolint:gosec // path provided by user flag
		if err != nil {
			return fmt.Errorf("read fail mockup: %w", err)
		}
	case templateExists:
		existingFail := filepath.Join(templatesDir, name+"_fail.md")
		if err := path.AssertContained(kbRoot, existingFail); err != nil {
			return fmt.Errorf("resolve fail mockup path: %w", err)
		}
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
	failFailed, _, err := evaluateValidations(celEnv, tmpl.Validations, failPage, nil)
	if err != nil {
		return fmt.Errorf("fail mockup: %w", err)
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

	targetDir := templatesDir
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

	// The atomic swap replaces the destination directory entry, so a final*
	// path that is a symlink is replaced rather than followed and the rename
	// cannot write through a link. Containment for this command comes from the
	// path.AssertContained checks above on the templates directory, the target
	// YAML path, and the mockups read back from .akb/templates. A future
	// refactor that writes finalYAML/finalPass/finalFail in place (os.WriteFile
	// instead of the temp-file plus rename swap) would follow a symlinked
	// destination rather than replace it and must re-check containment first.
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

// evaluateValidations compiles and evaluates every validation rule against
// page, with oldPage as the pre-modification state, and returns the IDs of the
// rules that did not evaluate to true. When a rule cannot be compiled or
// evaluated it returns that rule's ID with the error; the ID is empty
// otherwise.
func evaluateValidations(env *gocel.Env, rules []template.ValidationRule, page, oldPage map[string]any) ([]string, string, error) {
	var failed []string
	for _, rule := range rules {
		prg, err := cel.CompileRule(env, rule.Rule)
		if err != nil {
			return nil, rule.ID, fmt.Errorf("compile rule %q: %w", rule.ID, err)
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
			return nil, rule.ID, fmt.Errorf("evaluate rule %q: %w", rule.ID, err)
		}
		if result != types.True {
			failed = append(failed, rule.ID)
		}
	}
	return failed, "", nil
}

// optionalKeysSuppliedByMockup returns the schema-optional frontmatter keys the
// mockup sets, sorted for a stable evaluation order. type and title are never
// enumerated: cel.BuildPage always injects both keys, so removing them cannot
// produce an absent-key variant.
func optionalKeysSuppliedByMockup(tmpl *template.Template, fm *frontmatter.ParsedFrontmatter) []string {
	var keys []string
	for key, field := range tmpl.Schema.Frontmatter {
		if key == "type" || key == "title" {
			continue
		}
		if field.Required || !frontmatterKeyPresent(fm, key) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// frontmatterKeyPresent reports whether the parsed frontmatter sets key.
func frontmatterKeyPresent(fm *frontmatter.ParsedFrontmatter, key string) bool {
	_, ok := fm.Fields[key]
	return ok
}

// withoutFrontmatterKey returns a copy of fm with key removed, so the mockup
// can be evaluated as a page that omits a schema-optional field.
func withoutFrontmatterKey(fm *frontmatter.ParsedFrontmatter, key string) *frontmatter.ParsedFrontmatter {
	stripped := &frontmatter.ParsedFrontmatter{
		Type:   fm.Type,
		Title:  fm.Title,
		Fields: make(map[string]any, len(fm.Fields)),
	}
	for k, v := range fm.Fields {
		if k != key {
			stripped.Fields[k] = v
		}
	}
	return stripped
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
