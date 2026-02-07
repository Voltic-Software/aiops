package renderer

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/voltic-software/aiops/internal/config"
)

//go:embed all:templates
var templateFS embed.FS

// TemplateData is the data passed to all templates.
type TemplateData struct {
	Project        config.Project
	Detected       config.DetectedStack
	Build          config.BuildInfo
	GoModule       string // Root Go module path (e.g., "github.com/org/project")
	MultiagencyMod string // Multiagency module path (e.g., "github.com/org/project/multiagency")
	HasGo          bool
	HasTS          bool
	HasPython      bool
	HasRust        bool
	HasNextJS      bool
	HasReact       bool
	HasEventsrc    bool
	HasMQTT        bool
	HasDDD         bool
	HasMonorepo    bool
	HasMCP         bool
}

// NewTemplateData builds template data from a project config.
func NewTemplateData(cfg *config.ProjectConfig) *TemplateData {
	// Derive multiagency module path from Go module
	goMod := cfg.Detected.GoModule
	multiagencyMod := cfg.Project.Name + "/multiagency"
	if goMod != "" {
		multiagencyMod = goMod + "/multiagency"
	}

	td := &TemplateData{
		Project:        cfg.Project,
		Detected:       cfg.Detected,
		Build:          cfg.Detected.Build,
		GoModule:       goMod,
		MultiagencyMod: multiagencyMod,
	}

	for _, lang := range cfg.Detected.Languages {
		switch lang.Name {
		case "go":
			td.HasGo = true
		case "typescript", "javascript":
			td.HasTS = true
		case "python":
			td.HasPython = true
		case "rust":
			td.HasRust = true
		}
	}

	for _, fw := range cfg.Detected.Frameworks {
		switch fw.Name {
		case "nextjs":
			td.HasNextJS = true
		case "react":
			td.HasReact = true
		case "eventsrc":
			td.HasEventsrc = true
		case "mqtt":
			td.HasMQTT = true
		}
	}

	for _, p := range cfg.Detected.Patterns {
		switch p {
		case "domain-driven-design":
			td.HasDDD = true
		case "event-sourcing":
			td.HasEventsrc = true
		case "monorepo":
			td.HasMonorepo = true
		case "mcp-server":
			td.HasMCP = true
		}
	}

	return td
}

// GetTemplateFS returns the embedded template filesystem for external inspection.
func GetTemplateFS() embed.FS {
	return templateFS
}

// RenderAll renders all templates to the project directory.
func RenderAll(projectDir string, cfg *config.ProjectConfig) ([]string, error) {
	data := NewTemplateData(cfg)
	var rendered []string

	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Compute output path
		relPath := strings.TrimPrefix(path, "templates/")
		outPath := resolveOutputPath(projectDir, cfg, relPath)

		// Strip .tmpl extension for template files
		if strings.HasSuffix(outPath, ".tmpl") {
			outPath = strings.TrimSuffix(outPath, ".tmpl")
		}

		// Read template content
		content, err := templateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		var output []byte
		if strings.HasSuffix(path, ".tmpl") {
			// Render as Go template
			tmpl, err := template.New(filepath.Base(path)).Funcs(templateFuncs()).Parse(string(content))
			if err != nil {
				return fmt.Errorf("parsing template %s: %w", path, err)
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return fmt.Errorf("rendering template %s: %w", path, err)
			}
			output = buf.Bytes()
		} else {
			// Copy as-is
			output = content
		}

		// Write output
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}
		if err := os.WriteFile(outPath, output, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}

		rel, _ := filepath.Rel(projectDir, outPath)
		rendered = append(rendered, rel)
		return nil
	})

	return rendered, err
}

// resolveOutputPath maps template paths to actual output paths.
func resolveOutputPath(projectDir string, cfg *config.ProjectConfig, relPath string) string {
	windsurfDir := cfg.Paths.Windsurf
	if windsurfDir == "" {
		windsurfDir = ".windsurf"
	}

	multiagencyDir := cfg.Paths.Multiagency
	if multiagencyDir == "" {
		multiagencyDir = "multiagency"
	}

	if strings.HasPrefix(relPath, "windsurf/") {
		return filepath.Join(projectDir, windsurfDir, strings.TrimPrefix(relPath, "windsurf/"))
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

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"join": strings.Join,
		"indent": func(n int, s string) string {
			pad := strings.Repeat(" ", n)
			lines := strings.Split(s, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = pad + line
				}
			}
			return strings.Join(lines, "\n")
		},
		"contains": func(slice []string, item string) bool {
			for _, s := range slice {
				if s == item {
					return true
				}
			}
			return false
		},
		"bullet": func(items []string) string {
			var lines []string
			for _, item := range items {
				lines = append(lines, "- "+item)
			}
			return strings.Join(lines, "\n")
		},
	}
}
