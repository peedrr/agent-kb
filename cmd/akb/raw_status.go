package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/path"
)

var statusJSON bool

type driftFile struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type driftSummary struct {
	Modified  int `json:"modified"`
	Untracked int `json:"untracked"`
	Missing   int `json:"missing"`
}

type driftOutput struct {
	Files   []driftFile  `json:"files"`
	Summary driftSummary `json:"summary"`
}

var rawStatusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Check raw file drift status",
	Long:  `Compare the manifest against the filesystem to detect modified, untracked, or missing files.`,
	Example: `  # Check drift status for all files
  akb raw status

  # Check drift for specific file
  akb raw status data/config.json

  # JSON output
  akb raw status --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRawStatus,
}

func init() {
	rawStatusCmd.Flags().BoolVar(&statusJSON, "json", false, "output in JSON format")
}

func runRawStatus(_ *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	rawDir := filepath.Join(kbRoot, "raw")
	if info, err := os.Stat(rawDir); err != nil {
		if os.IsNotExist(err) {
			return &usageError{msg: "raw/ directory not found. Run 'akb init' to create a knowledge base."}
		}
		return fmt.Errorf("stat raw directory: %w", err)
	} else if !info.IsDir() {
		return &usageError{msg: "raw/ is not a directory. Run 'akb init' to create a knowledge base."}
	}

	mgr := manifest.NewManager(kbRoot)
	entries, err := mgr.ReadManifest()
	if err != nil {
		return &internalError{err: err}
	}

	// Build a map of manifest entries for lookup
	manifestMap := make(map[string]string) // filename -> sha256
	for _, e := range entries {
		manifestMap[e.Filename] = e.SHA256
	}

	// Walk raw/ directory to get actual files
	diskFiles := make(map[string]bool) // relative path -> exists
	err = filepath.WalkDir(rawDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(rawDir, p)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		relPath = filepath.ToSlash(relPath)
		if relPath == "files.log" {
			return nil
		}
		diskFiles[relPath] = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk raw directory: %w", err)
	}

	// The drifted list stays non-nil so the JSON envelope always renders files
	// as an array, never as null.
	drifted := make([]driftFile, 0)

	// Check for modified and untracked files (on disk)
	for relPath := range diskFiles {
		sha256OnDisk, hashErr := manifest.ComputeSHA256(filepath.Join(rawDir, relPath))
		if hashErr != nil {
			return fmt.Errorf("compute SHA-256 for %s: %w", relPath, hashErr)
		}
		if manifestSHA, inManifest := manifestMap[relPath]; inManifest {
			if manifestSHA != sha256OnDisk {
				drifted = append(drifted, driftFile{Path: relPath, Status: "MODIFIED"})
			}
		} else {
			drifted = append(drifted, driftFile{Path: relPath, Status: "UNTRACKED"})
		}
	}

	// Check for missing files (in manifest but not on disk)
	for _, e := range entries {
		if !diskFiles[e.Filename] {
			drifted = append(drifted, driftFile{Path: e.Filename, Status: "MISSING"})
		}
	}

	// If path argument provided, filter to only that file
	if len(args) > 0 {
		inputPath := args[0]
		resolvedPath, resolveErr := path.ResolveRawPath(kbRoot, inputPath)
		if resolveErr != nil {
			return fmt.Errorf("resolve raw path: %w", resolveErr)
		}
		targetRel, relErr := filepath.Rel(rawDir, resolvedPath)
		if relErr != nil {
			return fmt.Errorf("compute relative path: %w", relErr)
		}
		targetRel = filepath.ToSlash(targetRel)

		// A filter target that names neither a file under raw/ nor a manifest
		// entry covers no raw file, so an empty drift list would misreport it as
		// an undrifted one.
		if _, inManifest := manifestMap[targetRel]; !inManifest && !diskFiles[targetRel] {
			return &usageError{msg: fmt.Sprintf("no raw file or manifest entry matches %s", inputPath)}
		}

		filtered := make([]driftFile, 0, len(drifted))
		for _, f := range drifted {
			if f.Path == targetRel {
				filtered = append(filtered, f)
			}
		}
		drifted = filtered
	}

	summary := driftSummary{
		Modified:  countByStatus(drifted, "MODIFIED"),
		Untracked: countByStatus(drifted, "UNTRACKED"),
		Missing:   countByStatus(drifted, "MISSING"),
	}

	if statusJSON {
		output := driftOutput{
			Files:   drifted,
			Summary: summary,
		}
		data, jsonErr := json.MarshalIndent(output, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("marshal JSON: %w", jsonErr)
		}
		fmt.Println(string(data))
	} else {
		if len(drifted) == 0 {
			fmt.Println("no drift")
		} else {
			for _, f := range drifted {
				fmt.Printf("%s %s\n", f.Status, f.Path)
			}
		}
	}

	if len(drifted) > 0 {
		return driftDetected{}
	}

	return nil
}

func countByStatus(files []driftFile, status string) int {
	count := 0
	for _, f := range files {
		if f.Status == status {
			count++
		}
	}
	return count
}
