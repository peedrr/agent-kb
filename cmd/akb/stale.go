package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/registry"
)

const (
	FreshnessHalfLifeDays   = 30
	FreshnessScoreThreshold = 50.0
)

type stalePage struct {
	Path       string  `json:"path"`
	Score      float64 `json:"score"`
	Confidence string  `json:"confidence"`
	Updated    string  `json:"updated"`
}

type kbResult struct {
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	StalePages []stalePage `json:"stale_pages"`
	TotalStale int         `json:"total_stale"`
}

var staleJSON bool
var staleAll bool

var staleCmd = &cobra.Command{
	Use:   "stale",
	Short: "Report stale pages across all KBs",
	Long: `Check all registered KBs for stale pages based on last update time.

A page is considered stale if its freshness score drops below 50.0.
The freshness score decays exponentially based on days since last update,
weighted by confidence level (high=1.0, medium=0.7, low=0.4).

By default, only the active KB is checked. Use --all to check all registered KBs.`,
	Example: `  # Check for stale pages in the active KB
  akb stale

  # Check all registered KBs
  akb stale --all

  # Output as JSON
  akb stale --json`,
	RunE: runStale,
}

func init() {
	staleCmd.Flags().BoolVar(&staleJSON, "json", false, "output as JSON")
	staleCmd.Flags().BoolVar(&staleAll, "all", false, "check all registered KBs")
}

func runStale(_ *cobra.Command, _ []string) error {
	regPath, err := registry.Path()
	if err != nil {
		return fmt.Errorf("get registry path: %w", err)
	}

	reg, err := registry.Load(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if len(reg.Entries) == 0 {
		fmt.Println("No KBs registered. Use 'akb init <name>' to create one.")
		return nil
	}

	var entriesToCheck []registry.Entry

	if staleAll {
		entriesToCheck = reg.Entries
	} else {
		defaultEntry, err := registry.GetDefault()
		if err != nil {
			return fmt.Errorf("get active KB: %w", err)
		}
		entriesToCheck = []registry.Entry{defaultEntry}
	}

	now := time.Now()
	var results []kbResult
	totalStale := 0

	for _, entry := range entriesToCheck {
		kbPath := entry.Path
		kbDir := filepath.Join(kbPath, "kb")

		if _, err := os.Stat(kbDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: KB '%s' directory not found at %s. Skipping.\n", entry.Name, kbDir)
			continue
		}

		kbResult := kbResult{
			Name:       entry.Name,
			Path:       kbPath,
			StalePages: []stalePage{},
		}

		err := filepath.WalkDir(kbDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}

			content, err := os.ReadFile(path) //nolint:gosec // path validated by filepath.WalkDir within KB root
			if err != nil {
				return nil
			}

			fm, _, err := frontmatter.Parse(content)
			if err != nil {
				return nil
			}

			var tsStr string
			if fm.Fields != nil {
				if v, ok := fm.Fields["updated"]; ok {
					if s, ok := v.(string); ok {
						tsStr = s
					}
				}
				if tsStr == "" {
					if v, ok := fm.Fields["created"]; ok {
						if s, ok := v.(string); ok {
							tsStr = s
						}
					}
				}
			}

			if tsStr == "" {
				relPath, _ := filepath.Rel(kbDir, path)
				fmt.Fprintf(os.Stderr, "Warning: %s missing updated/created timestamp. Skipping.\n", relPath)
				return nil
			}

			parsedTime, err := parseTimestamp(tsStr)
			if err != nil {
				return nil
			}

			if parsedTime.After(now) {
				return nil
			}

			daysSince := now.Sub(parsedTime).Hours() / 24

			confidenceWeight := 0.7
			confidence := "medium"
			if fm.Fields != nil {
				if v, ok := fm.Fields["confidence"]; ok {
					if s, ok := v.(string); ok {
						switch s {
						case "high":
							confidenceWeight = 1.0
							confidence = "high"
						case "medium":
							confidenceWeight = 0.7
							confidence = "medium"
						case "low":
							confidenceWeight = 0.4
							confidence = "low"
						}
					}
				}
			}

			score := 100.0 * math.Pow(2, -daysSince/float64(FreshnessHalfLifeDays)) * confidenceWeight

			if score < FreshnessScoreThreshold {
				relPath, _ := filepath.Rel(kbDir, path)
				kbResult.StalePages = append(kbResult.StalePages, stalePage{
					Path:       relPath,
					Score:      score,
					Confidence: confidence,
					Updated:    tsStr,
				})
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("walk KB %s: %w", entry.Name, err)
		}

		kbResult.TotalStale = len(kbResult.StalePages)
		totalStale += kbResult.TotalStale
		results = append(results, kbResult)
	}

	if staleJSON {
		output := map[string]any{
			"kbs":         results,
			"total_stale": totalStale,
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		for _, r := range results {
			fmt.Printf("KB: %s (%s)\n", r.Name, r.Path)
			if len(r.StalePages) == 0 {
				fmt.Println("  No stale pages")
			} else {
				for _, p := range r.StalePages {
					fmt.Printf("  STALE: %s (score: %.1f, confidence: %s, last updated: %s)\n", p.Path, p.Score, p.Confidence, p.Updated)
				}
			}
		}
		fmt.Printf("Total: %d stale pages across %d KBs\n", totalStale, len(results))
	}

	if totalStale > 0 {
		os.Exit(1)
	}

	return nil
}

func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unparseable date: %s", s)
}
