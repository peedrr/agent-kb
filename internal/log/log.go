// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package log manages kb/log.md parsing and append operations.
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// LogEntry represents a single entry in kb/log.md.
//
//nolint:revive // intentionally exported for use by tests
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

// ReadLog parses kb/log.md and returns its entries.
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

// AppendLog adds a new entry to kb/log.md.
func AppendLog(kbRoot string, operation string, description string, title string) error {
	lp := logPath(kbRoot)

	var entries []LogEntry
	_, err := os.ReadFile(lp) //nolint:gosec // path constructed by logPath within KB root
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
	}

	entries = append(entries, LogEntry{
		Date:        time.Now().Format("2006-01-02"),
		Operation:   operation,
		Title:       title,
		Description: description,
	})

	rendered := RenderLog(entries)
	if err := os.WriteFile(lp, []byte(rendered), 0600); err != nil {
		return fmt.Errorf("write log.md: %w", err)
	}

	return nil
}

// FilterByType returns log entries matching the given operation.
func FilterByType(entries []LogEntry, operation string) []LogEntry {
	result := []LogEntry{}
	for _, e := range entries {
		if e.Operation == operation {
			result = append(result, e)
		}
	}
	return result
}

// FilterByLast returns the last N log entries.
func FilterByLast(entries []LogEntry, n int) []LogEntry {
	if n <= 0 {
		return []LogEntry{}
	}
	if n >= len(entries) {
		return entries
	}
	return entries[len(entries)-n:]
}

// RenderLog formats log entries as markdown.
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
