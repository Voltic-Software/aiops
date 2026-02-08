package target

import (
	"os"
	"path/filepath"
)

// Target represents an IDE or AI extension that aiops generates artifacts for.
type Target struct {
	Name          string // e.g., "windsurf", "cursor", "continue", "copilot"
	DisplayName   string // e.g., "Windsurf", "Cursor", "Continue (VS Code)", "GitHub Copilot"
	GlobalRules   string // Where global policy rules go (relative to $HOME). Empty if target has no global location.
	RepoRulesPath string // Where repo implementation rules go (relative to project dir)
	WorkflowsDir  string // Where workflow/prompt files go (relative to project dir)
	OrchestrDir   string // Where orchestrator state goes (relative to project dir)
	SkillsDir     string // Where skill scaffolds go (relative to project dir)
	RulesFormat   string // "markdown", "yaml", "mdc"
}

// All known targets.
var All = []Target{
	Windsurf,
	Cursor,
	Continue,
	Copilot,
}

var Windsurf = Target{
	Name:          "windsurf",
	DisplayName:   "Windsurf (Cascade)",
	GlobalRules:   filepath.Join(".codeium", "windsurf", "memories", "global_rules.md"),
	RepoRulesPath: ".windsurf/rules/aiops.md",
	WorkflowsDir:  ".windsurf/workflows",
	OrchestrDir:   ".windsurf/orchestrator",
	SkillsDir:     ".windsurf/skills",
	RulesFormat:   "markdown",
}

var Cursor = Target{
	Name:          "cursor",
	DisplayName:   "Cursor",
	GlobalRules:   "",
	RepoRulesPath: ".cursor/rules/aiops.mdc",
	WorkflowsDir:  ".cursor/prompts",
	OrchestrDir:   ".cursor/orchestrator",
	SkillsDir:     ".cursor/skills",
	RulesFormat:   "mdc",
}

var Continue = Target{
	Name:          "continue",
	DisplayName:   "Continue (VS Code)",
	GlobalRules:   "",
	RepoRulesPath: ".continue/rules/aiops.md",
	WorkflowsDir:  ".continue/prompts",
	OrchestrDir:   ".continue/orchestrator",
	SkillsDir:     ".continue/skills",
	RulesFormat:   "markdown",
}

var Copilot = Target{
	Name:          "copilot",
	DisplayName:   "GitHub Copilot",
	GlobalRules:   "",
	RepoRulesPath: ".github/copilot-instructions.md",
	WorkflowsDir:  "", // Copilot doesn't support custom workflows
	OrchestrDir:   "", // No orchestrator support
	SkillsDir:     "", // No skills support
	RulesFormat:   "markdown",
}

// ResolveGlobalRulesPath returns the absolute path for the global policy rules file.
// Returns empty string if this target has no global rules location.
func (t *Target) ResolveGlobalRulesPath() string {
	if t.GlobalRules == "" {
		return ""
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, t.GlobalRules)
}

// ResolveRepoRulesPath returns the absolute path for the repo implementation rules file.
func (t *Target) ResolveRepoRulesPath(projectDir string) string {
	if t.RepoRulesPath == "" {
		return ""
	}
	return filepath.Join(projectDir, t.RepoRulesPath)
}

// ResolveWorkflowsDir returns the absolute path for the workflows directory.
func (t *Target) ResolveWorkflowsDir(projectDir string) string {
	if t.WorkflowsDir == "" {
		return ""
	}
	return filepath.Join(projectDir, t.WorkflowsDir)
}

// ResolveOrchestrDir returns the absolute path for the orchestrator directory.
func (t *Target) ResolveOrchestrDir(projectDir string) string {
	if t.OrchestrDir == "" {
		return ""
	}
	return filepath.Join(projectDir, t.OrchestrDir)
}

// ResolveSkillsDir returns the absolute path for the skills directory.
func (t *Target) ResolveSkillsDir(projectDir string) string {
	if t.SkillsDir == "" {
		return ""
	}
	return filepath.Join(projectDir, t.SkillsDir)
}

// Detect scans the project directory and user home to determine which IDE targets are present.
func Detect(projectDir string) []Target {
	var detected []Target

	for _, t := range All {
		if isTargetPresent(projectDir, t) {
			detected = append(detected, t)
		}
	}

	// If nothing detected, default to all targets that use project-local files
	if len(detected) == 0 {
		detected = []Target{Windsurf, Cursor, Copilot}
	}

	return detected
}

func isTargetPresent(projectDir string, t Target) bool {
	switch t.Name {
	case "windsurf":
		// Check for Windsurf config dir or if Windsurf is installed
		home, _ := os.UserHomeDir()
		if dirExists(filepath.Join(home, ".codeium", "windsurf")) {
			return true
		}
		if dirExists(filepath.Join(projectDir, ".windsurf")) {
			return true
		}
	case "cursor":
		// Check for .cursor directory in project or Cursor app
		if dirExists(filepath.Join(projectDir, ".cursor")) {
			return true
		}
		home, _ := os.UserHomeDir()
		if dirExists(filepath.Join(home, ".cursor")) {
			return true
		}
	case "continue":
		// Check for .continue directory in project or home
		if dirExists(filepath.Join(projectDir, ".continue")) {
			return true
		}
		home, _ := os.UserHomeDir()
		if dirExists(filepath.Join(home, ".continue")) {
			return true
		}
	case "copilot":
		// Check for .github directory or VS Code with Copilot
		if dirExists(filepath.Join(projectDir, ".github")) {
			return true
		}
		home, _ := os.UserHomeDir()
		if dirExists(filepath.Join(home, ".vscode")) {
			return true
		}
	}
	return false
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
