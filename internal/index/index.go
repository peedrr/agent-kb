// Package index manages kb/index.md parsing, rendering, and updates.
package index

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/frontmatter"
)

// IndexEntry represents a single entry in kb/index.md.
//
//nolint:revive // intentionally exported for use by tests and external commands
type IndexEntry struct {
	Path    string
	Title   string
	Summary string
	Type    string
}

func indexPath(kbRoot string) string {
	return filepath.Join(kbRoot, "kb", "index.md")
}

func typeToHeading(typ string) string {
	if len(typ) == 0 {
		return ""
	}
	return strings.ToUpper(typ[:1]) + typ[1:] + "s"
}

func headingToType(heading string) string {
	if len(heading) == 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSuffix(heading, "s"))
}

// ReadIndex parses kb/index.md and returns its entries.
func ReadIndex(kbRoot string) ([]IndexEntry, error) {
	path := indexPath(kbRoot)
	data, err := os.ReadFile(path) //nolint:gosec // path constructed by indexPath within KB root
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}
	return parseIndex(string(data))
}

func parseIndex(content string) ([]IndexEntry, error) {
	var entries []IndexEntry
	var currentType string

	for line := range strings.SplitSeq(content, "\n") {
		line = strings.TrimRight(line, "\r")

		if line, ok := strings.CutPrefix(line, "## "); ok {
			heading := strings.TrimSpace(line)
			currentType = headingToType(heading)
			continue
		}

		if strings.HasPrefix(line, "- [") {
			entry, err := parseEntryLine(line, currentType)
			if err != nil {
				continue
			}
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func parseEntryLine(line string, typ string) (IndexEntry, error) {
	_, after, ok := strings.Cut(line, "[")
	if !ok {
		return IndexEntry{}, fmt.Errorf("malformed entry line: missing title")
	}
	title, _, ok := strings.Cut(after, "]")
	if !ok {
		return IndexEntry{}, fmt.Errorf("malformed entry line: missing title")
	}

	_, after, ok = strings.Cut(line, "(")
	if !ok {
		return IndexEntry{}, fmt.Errorf("malformed entry line: missing path")
	}
	path, _, ok := strings.Cut(after, ")")
	if !ok {
		return IndexEntry{}, fmt.Errorf("malformed entry line: missing path")
	}

	summary := ""
	if _, after, ok := strings.Cut(line, " \u2014 "); ok {
		summary = after
	}

	return IndexEntry{
		Path:    path,
		Title:   title,
		Summary: summary,
		Type:    typ,
	}, nil
}

// AddEntry adds or updates a page entry in kb/index.md.
func AddEntry(kbRoot string, entry IndexEntry) error {
	path := indexPath(kbRoot)

	var entries []IndexEntry
	data, err := os.ReadFile(path) //nolint:gosec // path constructed by indexPath within KB root
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read index: %w", err)
	}
	if err == nil {
		entries, err = parseIndex(string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse index: %v — run 'akb index rebuild' to regenerate\n", err)
			entries = nil
		}
	}

	found := false
	for i, e := range entries {
		if e.Path == entry.Path {
			entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, entry)
	}

	rendered := RenderIndex(entries)
	if err := os.WriteFile(path, []byte(rendered), 0600); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// RemoveEntry removes a page entry from kb/index.md.
func RemoveEntry(kbRoot string, entryPath string) error {
	path := indexPath(kbRoot)

	data, err := os.ReadFile(path) //nolint:gosec // path constructed by indexPath within KB root
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read index: %w", err)
	}

	entries, err := parseIndex(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to parse index: %v — run 'akb index rebuild' to regenerate\n", err)
		entries = nil
	}

	filtered := make([]IndexEntry, 0, len(entries))
	for _, e := range entries {
		if e.Path != entryPath {
			filtered = append(filtered, e)
		}
	}

	rendered := RenderIndex(filtered)
	if err := os.WriteFile(path, []byte(rendered), 0600); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// RebuildIndex regenerates kb/index.md from the filesystem.
func RebuildIndex(kbRoot string) error {
	kbDir := filepath.Join(kbRoot, "kb")

	var entries []IndexEntry

	err := filepath.WalkDir(kbDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		base := filepath.Base(path)
		if base == "index.md" || base == "log.md" {
			return nil
		}

		content, err := os.ReadFile(path) //nolint:gosec // path from filepath.WalkDir within KB root
		if err != nil {
			return nil
		}

		fm, _, err := frontmatter.Parse(content)
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(kbRoot, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		summary := ""
		if s, ok := fm.Fields["summary"]; ok {
			if str, ok := s.(string); ok {
				summary = str
			}
		}

		entries = append(entries, IndexEntry{
			Path:    relPath,
			Title:   fm.Title,
			Summary: summary,
			Type:    fm.Type,
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk kb directory: %w", err)
	}

	rendered := RenderIndex(entries)
	idxPath := indexPath(kbRoot)
	if err := os.WriteFile(idxPath, []byte(rendered), 0600); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// RenderIndex renders index entries as markdown.
func RenderIndex(entries []IndexEntry) string {
	var sb strings.Builder
	sb.WriteString("# Index\n\n")

	if len(entries) == 0 {
		return sb.String()
	}

	typeGroups := make(map[string][]IndexEntry)
	var typeOrder []string
	for _, e := range entries {
		if _, exists := typeGroups[e.Type]; !exists {
			typeOrder = append(typeOrder, e.Type)
		}
		typeGroups[e.Type] = append(typeGroups[e.Type], e)
	}

	for _, typ := range typeOrder {
		sb.WriteString("## ")
		sb.WriteString(typeToHeading(typ))
		sb.WriteString("\n\n")

		for _, e := range typeGroups[typ] {
			sb.WriteString("- [")
			sb.WriteString(e.Title)
			sb.WriteString("](")
			sb.WriteString(e.Path)
			sb.WriteString(")")
			if e.Summary != "" {
				sb.WriteString(" \u2014 ")
				sb.WriteString(e.Summary)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
