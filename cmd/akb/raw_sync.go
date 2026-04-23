package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/spf13/cobra"
)

var rawSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize raw file manifest with filesystem",
	Long:  `Reconcile the manifest with the actual files on disk. Adds untracked files, updates modified files, and removes missing files from the manifest.`,
	Args:  cobra.NoArgs,
	RunE:  runRawSync,
}

func runRawSync(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return err
	}

	// Check that raw/ directory exists
	rawDir := filepath.Join(kbRoot, "raw")
	if _, err := os.Stat(rawDir); os.IsNotExist(err) {
		return fmt.Errorf("raw/ directory does not exist; run `akb init` first")
	}

	// Read existing manifest
	mgr := manifest.NewManager(kbRoot)
	existingEntries, err := mgr.ReadManifest()
	if err != nil {
		return err
	}

	// Build a map of existing entries by filename
	existingMap := make(map[string]manifest.Entry)
	for _, e := range existingEntries {
		existingMap[e.Filename] = e
	}

	// Walk raw/ directory to get actual files
	actualFiles := make(map[string]string) // filename -> sha256
	if err := filepath.WalkDir(rawDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		// Skip files.log itself
		if d.Name() == "files.log" {
			return nil
		}
		relPath, err := filepath.Rel(rawDir, path)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		relPath = filepath.ToSlash(relPath)

		sha256hash, err := manifest.ComputeSHA256(path)
		if err != nil {
			return fmt.Errorf("compute SHA-256 for %s: %w", relPath, err)
		}
		actualFiles[relPath] = sha256hash
		return nil
	}); err != nil {
		return fmt.Errorf("walk raw directory: %w", err)
	}

	// Compare and build updated entries
	var updatedEntries []manifest.Entry
	newCount := 0
	modifiedCount := 0
	deletedCount := 0
	unchangedCount := 0

	// Process files on disk
	for filename, sha256hash := range actualFiles {
		if existingEntry, exists := existingMap[filename]; exists {
			if existingEntry.SHA256 == sha256hash {
				// Unchanged - keep existing entry with its original timestamp
				updatedEntries = append(updatedEntries, existingEntry)
				unchangedCount++
			} else {
				// Modified - update entry with new hash and current timestamp
				updatedEntries = append(updatedEntries, manifest.Entry{
					Filename:    filename,
					SHA256:      sha256hash,
					LastUpdated: time.Now().UTC().Format(time.RFC3339),
				})
				modifiedCount++
			}
			// Remove from existingMap so we know what's processed
			delete(existingMap, filename)
		} else {
			// New file - add entry
			updatedEntries = append(updatedEntries, manifest.Entry{
				Filename:    filename,
				SHA256:      sha256hash,
				LastUpdated: time.Now().UTC().Format(time.RFC3339),
			})
			newCount++
		}
	}

	// Remaining entries in existingMap are missing from disk
	for range existingMap {
		deletedCount++
	}

	// Write the updated manifest
	if err := mgr.WriteManifest(updatedEntries); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	// Git add and commit if there are changes
	totalChanges := newCount + modifiedCount + deletedCount
	if totalChanges > 0 && !noCommit {
		// Ensure git config is set
		if err := ensureGitConfig(kbRoot); err != nil {
			return fmt.Errorf("git config: %w", err)
		}

		// Stage raw/ directory and manifest
		gitAdd1 := exec.Command("git", "add", "-A", "raw/")
		gitAdd1.Dir = kbRoot
		if out, err := gitAdd1.CombinedOutput(); err != nil {
			return fmt.Errorf("git add raw/: %s: %w", strings.TrimSpace(string(out)), err)
		}

		gitAdd2 := exec.Command("git", "add", "raw/files.log")
		gitAdd2.Dir = kbRoot
		if out, err := gitAdd2.CombinedOutput(); err != nil {
			return fmt.Errorf("git add raw/files.log: %s: %w", strings.TrimSpace(string(out)), err)
		}

		// Commit with summary
		commitMsg := fmt.Sprintf("akb: raw sync (%d new, %d modified, %d deleted, %d unchanged)",
			newCount, modifiedCount, deletedCount, unchangedCount)
		gitCommit := exec.Command("git", "commit", "-m", commitMsg)
		gitCommit.Dir = kbRoot
		if out, err := gitCommit.CombinedOutput(); err != nil {
			return fmt.Errorf("git commit: %s: %w", strings.TrimSpace(string(out)), err)
		}
	}

	// Print summary
	if totalChanges == 0 {
		fmt.Println("No changes to sync.")
	} else {
		fmt.Printf("Synced: %d new, %d modified, %d deleted, %d unchanged\n",
			newCount, modifiedCount, deletedCount, unchangedCount)
	}

	return nil
}
