package updater

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/voltic-software/aiops/internal/config"
	"github.com/voltic-software/aiops/internal/renderer"
)

// Diff represents a single file difference between current and updated.
type Diff struct {
	Path        string
	Status      string // "new", "modified", "unchanged"
	CurrentHash string
	NewHash     string
}

// Plan represents the update plan showing what would change.
type Plan struct {
	Diffs     []Diff
	NewFiles  int
	Modified  int
	Unchanged int
}

// ComputePlan compares current installed files against what templates would generate.
func ComputePlan(projectDir string, cfg *config.ProjectConfig) (*Plan, error) {
	// Render to temp dir using same relative paths as real project
	tmpDir, err := os.MkdirTemp("", "aiops-update-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use same relative paths so rendered list matches real structure
	tmpCfg := *cfg
	tmpCfg.Paths.Memories = filepath.Join(tmpDir, "memories")

	// Render templates to temp dir
	rendered, err := renderer.RenderAll(tmpDir, &tmpCfg)
	if err != nil {
		return nil, fmt.Errorf("rendering templates: %w", err)
	}

	plan := &Plan{}

	for _, relPath := range rendered {
		tmpPath := filepath.Join(tmpDir, relPath)

		// Determine real output path
		realPath := resolveRealPath(projectDir, cfg, relPath)

		// Use a display path that's project-relative
		displayPath := relPath
		if strings.HasPrefix(relPath, "../") || strings.HasPrefix(relPath, "/") {
			// Memory paths may be outside project dir
			displayPath = realPath
		}

		newHash := hashFile(tmpPath)
		currentHash := hashFile(realPath)

		diff := Diff{
			Path:        displayPath,
			CurrentHash: currentHash,
			NewHash:     newHash,
		}

		if currentHash == "" {
			diff.Status = "new"
			plan.NewFiles++
		} else if currentHash != newHash {
			diff.Status = "modified"
			plan.Modified++
		} else {
			diff.Status = "unchanged"
			plan.Unchanged++
		}

		plan.Diffs = append(plan.Diffs, diff)
	}

	return plan, nil
}

// Apply writes updated files for the given diffs.
func Apply(projectDir string, cfg *config.ProjectConfig, plan *Plan, includeUnchanged bool) ([]string, error) {
	// Render fresh templates
	files, err := renderer.RenderAll(projectDir, cfg)
	if err != nil {
		return nil, fmt.Errorf("rendering templates: %w", err)
	}

	var applied []string
	for _, diff := range plan.Diffs {
		if diff.Status == "unchanged" && !includeUnchanged {
			continue
		}
		applied = append(applied, diff.Path)
	}

	// RenderAll already wrote the files, so just return what changed
	_ = files
	return applied, nil
}

// resolveRealPath maps a temp-rendered relative path back to the real project path.
func resolveRealPath(projectDir string, cfg *config.ProjectConfig, relPath string) string {
	windsurfDir := cfg.Paths.Windsurf
	if windsurfDir == "" {
		windsurfDir = ".windsurf"
	}
	multiagencyDir := cfg.Paths.Multiagency
	if multiagencyDir == "" {
		multiagencyDir = "multiagency"
	}

	if strings.HasPrefix(relPath, ".windsurf/") || strings.HasPrefix(relPath, "windsurf/") {
		return filepath.Join(projectDir, windsurfDir, strings.TrimPrefix(strings.TrimPrefix(relPath, ".windsurf/"), "windsurf/"))
	}
	if strings.HasPrefix(relPath, "multiagency/") {
		return filepath.Join(projectDir, multiagencyDir, strings.TrimPrefix(relPath, "multiagency/"))
	}
	if strings.HasPrefix(relPath, "memories/") {
		memDir := cfg.Paths.Memories
		if memDir == "" {
			home, _ := os.UserHomeDir()
			memDir = filepath.Join(home, ".codeium", "windsurf", "memories")
		}
		return filepath.Join(memDir, strings.TrimPrefix(relPath, "memories/"))
	}
	return filepath.Join(projectDir, relPath)
}

func hashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}

// ListTemplateFiles returns the list of template files that would be rendered.
func ListTemplateFiles() ([]string, error) {
	var files []string
	templateFS := renderer.GetTemplateFS()
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		files = append(files, strings.TrimPrefix(path, "templates/"))
		return nil
	})
	return files, err
}
