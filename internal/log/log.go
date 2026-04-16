package log

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type LogEntry struct {
	Date        string
	Operation   string
	Title       string
	Description string
}

var headingRe = regexp.MustCompile(`^## (\d{4}-\d{2}-\d{2}) (\w+)(?: \| (.+))?$`)

func logPath(kbRoot string) string {
	return filepath.Join(kbRoot, "kb", "log.md")
}

func ReadLog(kbRoot string) ([]LogEntry, error) {
	data, err := os.ReadFile(logPath(kbRoot))
	if err != nil {
		return nil, fmt.Errorf("read log.md: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var entries []LogEntry
	var current *LogEntry

	for _, line := range lines {
		m := headingRe.FindStringSubmatch(line)
		if m != nil {
			if current != nil {
				current.Description = strings.TrimSpace(current.Description)
				entries = append(entries, *current)
			}
			current = &LogEntry{
				Date:      m[1],
				Operation: m[2],
				Title:     m[3],
			}
			continue
		}
		if current != nil && !strings.HasPrefix(line, "# ") {
			current.Description += line + "\n"
		}
	}

	if current != nil {
		current.Description = strings.TrimSpace(current.Description)
		entries = append(entries, *current)
	}

	return entries, nil
}

func AppendLog(kbRoot string, operation string, description string, title string) error {
	lp := logPath(kbRoot)

	var entries []LogEntry
	data, err := os.ReadFile(lp)
	if err != nil {
		if os.IsNotExist(err) {
			entries = []LogEntry{}
		} else {
			return fmt.Errorf("read log.md: %w", err)
		}
	} else {
		entries, err = ReadLog(kbRoot)
		if err != nil {
			return fmt.Errorf("parse log.md: %w", err)
		}
		_ = data
	}

	entries = append(entries, LogEntry{
		Date:        time.Now().Format("2006-01-02"),
		Operation:   operation,
		Title:       title,
		Description: description,
	})

	rendered := RenderLog(entries)
	if err := os.WriteFile(lp, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("write log.md: %w", err)
	}

	return nil
}

func FilterByType(entries []LogEntry, operation string) []LogEntry {
	result := []LogEntry{}
	for _, e := range entries {
		if e.Operation == operation {
			result = append(result, e)
		}
	}
	return result
}

func FilterByLast(entries []LogEntry, n int) []LogEntry {
	if n <= 0 {
		return []LogEntry{}
	}
	if n >= len(entries) {
		return entries
	}
	return entries[len(entries)-n:]
}

func RenderLog(entries []LogEntry) string {
	var b strings.Builder
	b.WriteString("# Log\n\n")

	for _, e := range entries {
		if e.Title != "" {
			fmt.Fprintf(&b, "## %s %s | %s\n\n%s\n\n", e.Date, e.Operation, e.Title, e.Description)
		} else {
			fmt.Fprintf(&b, "## %s %s\n\n%s\n\n", e.Date, e.Operation, e.Description)
		}
	}

	return b.String()
}
