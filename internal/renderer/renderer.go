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
	"github.com/voltic-software/aiops/internal/target"
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
	TargetName     string // Current target being rendered (e.g., "windsurf", "cursor")
	TargetDisplay  string // Human-readable target name
	OrchestrDir    string // Orchestrator dir for this target (e.g., ".windsurf/orchestrator")
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

// RenderAll renders all templates to the project directory for all detected targets.
func RenderAll(projectDir string, cfg *config.ProjectConfig) ([]string, error) {
	targets := resolveTargets(cfg)
	var allRendered []string

	// Render target-specific artifacts (rules, workflows, orchestrator) for each target
	for _, t := range targets {
		files, err := renderForTarget(projectDir, cfg, t)
		if err != nil {
			return nil, fmt.Errorf("rendering for %s: %w", t.DisplayName, err)
		}
		allRendered = append(allRendered, files...)
	}

	// Render target-independent artifacts (multiagency) once
	files, err := renderShared(projectDir, cfg)
	if err != nil {
		return nil, fmt.Errorf("rendering shared artifacts: %w", err)
	}
	allRendered = append(allRendered, files...)

	return allRendered, nil
}

// resolveTargets returns the targets to render for.
func resolveTargets(cfg *config.ProjectConfig) []target.Target {
	if len(cfg.Paths.Targets) > 0 {
		var targets []target.Target
		for _, name := range cfg.Paths.Targets {
			for _, t := range target.All {
				if t.Name == name {
					targets = append(targets, t)
				}
			}
		}
		if len(targets) > 0 {
			return targets
		}
	}
	// Fallback: windsurf only (backward compat)
	return []target.Target{target.Windsurf}
}

// renderForTarget renders rules, workflows, and orchestrator for a single target.
func renderForTarget(projectDir string, cfg *config.ProjectConfig, t target.Target) ([]string, error) {
	data := NewTemplateData(cfg)
	data.TargetName = t.Name
	data.TargetDisplay = t.DisplayName
	data.OrchestrDir = t.OrchestrDir

	var rendered []string

	// 1. Render rules
	rulesOut := t.ResolveRulesPath(projectDir)
	if rulesOut != "" {
		output, err := renderTemplate("templates/memories/global_rules.md.tmpl", data)
		if err != nil {
			return nil, fmt.Errorf("rendering rules for %s: %w", t.Name, err)
		}
		if err := writeFile(rulesOut, output); err != nil {
			return nil, err
		}
		rel, _ := filepath.Rel(projectDir, rulesOut)
		rendered = append(rendered, rel)
	}

	// 2. Render workflows
	workflowsDir := t.ResolveWorkflowsDir(projectDir)
	if workflowsDir != "" {
		wfFiles, err := renderDir("templates/windsurf/workflows", workflowsDir, data)
		if err != nil {
			return nil, err
		}
		for _, f := range wfFiles {
			rel, _ := filepath.Rel(projectDir, f)
			rendered = append(rendered, rel)
		}
	}

	// 3. Render orchestrator
	orchestrDir := t.ResolveOrchestrDir(projectDir)
	if orchestrDir != "" {
		orcFiles, err := renderDir("templates/windsurf/orchestrator", orchestrDir, data)
		if err != nil {
			return nil, err
		}
		for _, f := range orcFiles {
			rel, _ := filepath.Rel(projectDir, f)
			rendered = append(rendered, rel)
		}
	}

	return rendered, nil
}

// renderShared renders target-independent artifacts (multiagency).
func renderShared(projectDir string, cfg *config.ProjectConfig) ([]string, error) {
	data := NewTemplateData(cfg)
	data.TargetName = "shared"
	var rendered []string

	multiagencyDir := cfg.Paths.Multiagency
	if multiagencyDir == "" {
		multiagencyDir = "multiagency"
	}

	err := fs.WalkDir(templateFS, "templates/multiagency", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath := strings.TrimPrefix(path, "templates/multiagency/")
		outPath := filepath.Join(projectDir, multiagencyDir, relPath)
		outPath = strings.TrimSuffix(outPath, ".tmpl")

		output, err := renderFileContent(path, data)
		if err != nil {
			return err
		}

		if err := writeFile(outPath, output); err != nil {
			return err
		}

		rel, _ := filepath.Rel(projectDir, outPath)
		rendered = append(rendered, rel)
		return nil
	})

	return rendered, err
}

// renderTemplate renders a single template file and returns the output bytes.
func renderTemplate(templatePath string, data *TemplateData) ([]byte, error) {
	content, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("reading template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(templateFuncs()).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering template %s: %w", templatePath, err)
	}

	return buf.Bytes(), nil
}

// renderDir renders all files in a template directory to an output directory.
func renderDir(templateDir, outputDir string, data *TemplateData) ([]string, error) {
	var rendered []string

	err := fs.WalkDir(templateFS, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath := strings.TrimPrefix(path, templateDir+"/")
		outPath := filepath.Join(outputDir, relPath)
		outPath = strings.TrimSuffix(outPath, ".tmpl")

		output, err := renderFileContent(path, data)
		if err != nil {
			return err
		}

		if err := writeFile(outPath, output); err != nil {
			return err
		}

		rendered = append(rendered, outPath)
		return nil
	})

	return rendered, err
}

// renderFileContent reads and optionally templates a file.
func renderFileContent(path string, data *TemplateData) ([]byte, error) {
	content, err := templateFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	if strings.HasSuffix(path, ".tmpl") {
		tmpl, err := template.New(filepath.Base(path)).Funcs(templateFuncs()).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", path, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("rendering template %s: %w", path, err)
		}
		return buf.Bytes(), nil
	}

	return content, nil
}

// writeFile creates parent directories and writes content.
func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
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
