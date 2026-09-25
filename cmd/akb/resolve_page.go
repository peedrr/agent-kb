// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/template"
)

// errPageNotFound reports that no candidate held a readable page. Each command
// turns it into the not-found message of its own path argument.
var errPageNotFound = errors.New("page not found")

// typeDirsProvider returns the directories a page can be stored under besides
// the path the caller named. It runs only when that path holds no readable
// page, so a command that resolves a named path never reads the templates that
// declare the directories.
type typeDirsProvider func() ([]string, error)

// typeDirsFromTemplates returns the provider for loaded templates: the
// directories the templates declare.
func typeDirsFromTemplates(templates map[string]template.Template) typeDirsProvider {
	return func() ([]string, error) {
		return declaredTypeDirs(templates), nil
	}
}

// typeDirsFromDisk returns the provider for the templates of a knowledge base.
// Nothing is read until the provider runs; the templates directory is checked
// against the KB root before it is read, and a template problem is reported by
// the provider, not before it.
func typeDirsFromDisk(kbRoot string) typeDirsProvider {
	return func() ([]string, error) {
		templatesDir := filepath.Join(kbRoot, ".akb", "templates")
		if err := path.AssertContained(kbRoot, templatesDir); err != nil {
			return nil, fmt.Errorf("resolve templates directory: %w", err)
		}
		// The guard's error names the offending template file, so it is
		// reported as it stands.
		if err := assertTemplateFilesContained(kbRoot, templatesDir); err != nil {
			return nil, err
		}
		templates, err := template.LoadTemplates(templatesDir)
		if err != nil {
			return nil, fmt.Errorf("load templates: %w", err)
		}
		return declaredTypeDirs(templates), nil
	}
}

// declaredTypeDirs lists the directories the templates declare, one entry per
// template that names one.
func declaredTypeDirs(templates map[string]template.Template) []string {
	dirs := make([]string, 0, len(templates))
	for _, tmpl := range templates {
		if tmpl.Dir != "" {
			dirs = append(dirs, tmpl.Dir)
		}
	}
	return dirs
}

// resolveExistingPage locates the page the input names and returns the path it
// was found at, its KB-relative path, and its current content.
//
// The page is read from the path the input resolves to and, when no readable
// file is there, from the directories the provider returns: a page is stored
// under the directory of its type, so its bare filename addresses it. A type
// directory is a component the resolver did not validate, so every candidate is
// checked against the KB root on its own before it is read — a candidate that
// escapes the root fails the resolution, while a candidate that merely holds no
// readable file is passed over. The named path is tried first, then the type
// directories in ascending order, and the first readable candidate wins.
//
// The KB-relative path is derived from the candidate the page was found at
// rather than from the input, because it names the page for the search index,
// the link graph and the commit message.
//
// A page that no candidate holds is reported as errPageNotFound.
func resolveExistingPage(kbRoot, inputPath string, typeDirs typeDirsProvider) (string, string, []byte, error) {
	namedPath, err := path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve path: %w", err)
	}

	fullPath := namedPath
	content, found, err := readPageCandidate(kbRoot, fullPath)
	if err != nil {
		return "", "", nil, err
	}

	if !found {
		dirs, err := typeDirs()
		if err != nil {
			return "", "", nil, err
		}
		cleanPath := strings.TrimPrefix(inputPath, "kb/")
		for _, dir := range sortedDistinct(dirs) {
			candidate := filepath.Join(kbRoot, "kb", dir, cleanPath)
			content, found, err = readPageCandidate(kbRoot, candidate)
			if err != nil {
				return "", "", nil, err
			}
			if found {
				fullPath = candidate
				break
			}
		}
	}

	if !found {
		return "", "", nil, errPageNotFound
	}
	return fullPath, kbRelativePath(kbRoot, fullPath), content, nil
}

// readPageCandidate reads one candidate page. It reports whether the candidate
// holds a readable file — without an error when it does not — and reports a
// path failure when the candidate escapes the KB root.
func readPageCandidate(kbRoot, candidate string) ([]byte, bool, error) {
	if err := path.AssertContained(kbRoot, candidate); err != nil {
		return nil, false, fmt.Errorf("resolve path: %w", err)
	}
	data, err := os.ReadFile(candidate) //nolint:gosec // candidate is checked against the KB root above
	if err != nil {
		return nil, false, nil
	}
	return data, true, nil
}

// sortedDistinct returns the directories without duplicates, in ascending
// order, so a page name that exists under more than one of them resolves the
// same way on every run.
func sortedDistinct(dirs []string) []string {
	seen := make(map[string]struct{}, len(dirs))
	distinct := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		distinct = append(distinct, dir)
	}
	sort.Strings(distinct)
	return distinct
}

// kbRelativePath names an on-disk page the way the knowledge base addresses it:
// below the KB root, with forward slashes.
func kbRelativePath(kbRoot, fullPath string) string {
	return filepath.ToSlash(strings.TrimPrefix(fullPath, kbRoot+string(filepath.Separator)))
}
