package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/voltic-software/aiops/internal/config"
	"github.com/voltic-software/aiops/internal/evolve"
	"github.com/voltic-software/aiops/internal/renderer"
	"github.com/voltic-software/aiops/internal/scanner"
	"github.com/voltic-software/aiops/internal/skills"
	"github.com/voltic-software/aiops/internal/target"
	"github.com/voltic-software/aiops/internal/updater"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		cmdInit()
	case "scan":
		cmdScan()
	case "status":
		cmdStatus()
	case "update":
		cmdUpdate()
	case "evolve":
		cmdEvolve()
	case "skills":
		cmdSkills()
	case "sync":
		cmdSync()
	case "doctor":
		cmdDoctor()
	case "uninstall":
		cmdUninstall()
	case "version":
		fmt.Printf("aiops %s\n", config.Version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`aiops — AI-powered operations infrastructure for your codebase

Usage:
  aiops init      Scan repo, generate config, install Cascade artifacts
  aiops scan      Re-scan repo and show detected stack (does not write files)
  aiops sync      Re-scan MCPs and targets, re-render rules (no questions)
  aiops status    Show what's installed and check for staleness
  aiops update    Regenerate artifacts from latest templates, show diff
  aiops evolve    Read directive logs and propose rule changes
  aiops skills    Generate skill scaffolds from detected frameworks
  aiops doctor    Check integrity of aiops installation
  aiops uninstall Remove all aiops artifacts from this repository
  aiops version   Show version

Options:
  --dir <path>    Project directory (default: current directory)
  --yes           Skip confirmation prompts (for uninstall)
  --help          Show this help`)
}

// getDir returns the project directory from --dir flag or current directory.
func getDir() string {
	for i, arg := range os.Args {
		if arg == "--dir" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine working directory: %v\n", err)
		os.Exit(1)
	}
	return dir
}

// --- init command ---

func cmdInit() {
	dir := getDir()
	fmt.Printf("aiops init — scanning %s\n\n", dir)

	// Check if already initialized
	if config.Exists(dir) {
		fmt.Println("⚠  .aiops.yaml already exists in this directory.")
		if !confirm("Reinitialize? This will overwrite generated files.") {
			fmt.Println("Aborted.")
			return
		}
		fmt.Println()
	}

	// 1. Scan the repo
	fmt.Println("Scanning repository...")
	stack, err := scanner.Scan(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	// 2. Show what we found
	fmt.Println()
	printDetected(stack)

	// 3. Confirm
	if !confirm("Is this correct?") {
		fmt.Println("\nYou can edit .aiops.yaml manually after init to correct any issues.")
		fmt.Println("Continuing with detected values...")
	}
	fmt.Println()

	// 4. Ask for project name
	projectName := ask("Project name", filepath.Base(dir))

	// 5. Detect IDE targets
	targets := target.Detect(dir)
	var targetNames []string
	for _, t := range targets {
		targetNames = append(targetNames, t.Name)
	}

	fmt.Println()
	fmt.Print("IDE targets: ")
	for i, t := range targets {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(t.DisplayName)
	}
	fmt.Println()

	// 5b. Detect project maturity
	maturity := scanner.DetectMaturity(dir)
	fmt.Printf("Project maturity: %s\n", maturity)

	// 6. Detect existing skills and specs (from previous init or manual creation)
	skills := scanner.DetectSkills(dir)
	specs := scanner.DetectSpecs(dir)
	stack.Skills = skills
	stack.Specs = specs

	// 7. Build config
	paths := config.DefaultPaths()
	paths.Targets = targetNames
	cfg := &config.ProjectConfig{
		Version: config.Version,
		Project: config.Project{
			Name:     projectName,
			Maturity: maturity,
		},
		Paths:    paths,
		Detected: *stack,
	}

	// 8. Save config
	if err := config.Save(dir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Created .aiops.yaml\n")

	// 9. Render templates (first pass — generates base specs)
	fmt.Println("\nGenerating artifacts...")
	files, err := renderer.RenderAll(dir, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering templates: %v\n", err)
		os.Exit(1)
	}

	// 10. Re-detect specs after first render (picks up newly generated specs)
	newSpecs := scanner.DetectSpecs(dir)
	if len(newSpecs) > len(specs) {
		cfg.Detected.Specs = newSpecs
		// Re-render workflows that reference specs (they now have the full list)
		files2, err := renderer.RenderAll(dir, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error re-rendering with specs: %v\n", err)
			os.Exit(1)
		}
		files = files2
		// Update saved config with detected specs
		_ = config.Save(dir, cfg)
	}

	for _, f := range files {
		fmt.Printf("  ✓ %s\n", f)
	}

	fmt.Printf("\n✅ aiops initialized! %d files generated.\n", len(files))

	if maturity == config.MaturityBootstrap {
		fmt.Println("\n🚀 Bootstrap mode detected — recommended first actions:")
		fmt.Println("  1. Open an AI session and run: /multiagency design.yaml")
		fmt.Println("  2. Produce architecture.md, risks.md, assumptions.md")
		fmt.Println("  3. After architecture is framed, start building (single-agent)")
		fmt.Println("  4. Run `aiops sync` after the project matures")
	} else {
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Review the generated files")
		fmt.Println("  2. Commit them to version control")
		fmt.Println("  3. Open a new AI session — the rules are now active")
		fmt.Println("  4. Run `aiops status` to check for updates later")
	}
}

// --- scan command ---

func cmdScan() {
	dir := getDir()
	fmt.Printf("aiops scan — analyzing %s\n\n", dir)

	stack, err := scanner.Scan(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	printDetected(stack)

	fmt.Println("\nThis is a read-only scan. Run `aiops init` to generate files.")
}

// --- status command ---

func cmdStatus() {
	dir := getDir()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("AIops is not installed in this repository.")
		fmt.Println("Run: aiops init")
		os.Exit(1)
	}

	fmt.Println("AIops installed ✔")
	fmt.Printf("Version:    %s\n", config.Version)
	fmt.Printf("Project:    %s\n", cfg.Project.Name)
	fmt.Printf("Maturity:   %s\n", cfg.Project.Maturity)

	// Targets
	if len(cfg.Paths.Targets) > 0 {
		fmt.Printf("Targets:    %s\n", strings.Join(cfg.Paths.Targets, ", "))
	}

	// MCP servers
	if len(cfg.Detected.MCPServers) > 0 {
		var mcpNames []string
		for _, m := range cfg.Detected.MCPServers {
			mcpNames = append(mcpNames, m.Name)
		}
		fmt.Printf("MCP:        %s\n", strings.Join(mcpNames, ", "))
	} else {
		fmt.Println("MCP:        none")
	}

	// Skills
	detectedSkills := scanner.DetectSkills(dir)
	fmt.Printf("Skills:     %d\n", len(detectedSkills))

	// Specs
	detectedSpecs := scanner.DetectSpecs(dir)
	fmt.Printf("Workflows:  %d\n", len(detectedSpecs))

	// Check which artifacts exist
	type artifact struct {
		path  string
		label string
	}

	artifacts := []artifact{
		{filepath.Join(dir, ".aiops", "soul.md"), "Soul (constitution)"},
		{filepath.Join(dir, cfg.Paths.Windsurf, "workflows", "default-mode.md"), "Default mode workflow"},
		{filepath.Join(dir, cfg.Paths.Windsurf, "workflows", "multiagency.md"), "Multiagency workflow"},
		{filepath.Join(dir, cfg.Paths.Windsurf, "workflows", "orchestrator.md"), "Orchestrator workflow"},
		{filepath.Join(dir, cfg.Paths.Windsurf, "orchestrator", "session_state.yaml"), "Session state"},
	}

	// Check for repo rules
	for _, t := range target.Detect(dir) {
		if t.RepoRulesPath != "" {
			artifacts = append(artifacts, artifact{filepath.Join(dir, t.RepoRulesPath), fmt.Sprintf("Repo rules (%s)", t.DisplayName)})
		}
	}

	fmt.Println("\nArtifacts:")
	installed := 0
	missing := 0
	for _, a := range artifacts {
		if _, err := os.Stat(a.path); err == nil {
			fmt.Printf("  ✓ %s\n", a.label)
			installed++
		} else {
			fmt.Printf("  ✗ %s (missing)\n", a.label)
			missing++
		}
	}

	fmt.Printf("\n%d installed, %d missing\n", installed, missing)

	// Re-scan and compare
	fmt.Println("\nDrift check...")
	stack, err := scanner.Scan(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: scan failed: %v\n", err)
		return
	}

	diffs := compareStack(&cfg.Detected, stack)
	if len(diffs) == 0 {
		fmt.Println("✓ No drift detected")
	} else {
		fmt.Println("⚠  Stack drift detected:")
		for _, d := range diffs {
			fmt.Printf("  - %s\n", d)
		}
		fmt.Println("\nRun `aiops sync` to update.")
	}

	// Check orchestrator state
	statePath := filepath.Join(dir, cfg.Paths.Windsurf, "orchestrator", "session_state.yaml")
	if _, err := os.Stat(statePath); err == nil {
		fmt.Println("\nOrchestrator: active")
	}
}

// --- update command ---

func cmdUpdate() {
	dir := getDir()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("✗ Not initialized. Run `aiops init` first.")
		os.Exit(1)
	}

	fmt.Printf("aiops update — %s\n\n", cfg.Project.Name)
	fmt.Println("Computing diff against latest templates...")

	plan, err := updater.ComputePlan(dir, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing plan: %v\n", err)
		os.Exit(1)
	}

	if plan.NewFiles == 0 && plan.Modified == 0 {
		fmt.Println("✓ All artifacts are up to date. No changes needed.")
		return
	}

	fmt.Printf("\nUpdate plan: %d new, %d modified, %d unchanged\n\n", plan.NewFiles, plan.Modified, plan.Unchanged)

	for _, diff := range plan.Diffs {
		switch diff.Status {
		case "new":
			fmt.Printf("  + %s (new)\n", diff.Path)
		case "modified":
			fmt.Printf("  ~ %s (modified)\n", diff.Path)
		}
	}

	fmt.Println()
	if !confirm("Apply these changes?") {
		fmt.Println("Aborted.")
		return
	}

	files, err := renderer.RenderAll(dir, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error applying update: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Updated %d files.\n", len(files))
}

// --- evolve command ---

func cmdEvolve() {
	dir := getDir()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("✗ Not initialized. Run `aiops init` first.")
		os.Exit(1)
	}

	windsurfDir := cfg.Paths.Windsurf
	if windsurfDir == "" {
		windsurfDir = ".windsurf"
	}

	fmt.Printf("aiops evolve — analyzing directive logs for %s\n\n", cfg.Project.Name)

	patterns, err := evolve.AnalyzeDirectives(dir, windsurfDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("\nNo directive log found. This is normal if no @directive overrides have been used yet.")
		return
	}

	// Count total directives
	total := 0
	for _, p := range patterns {
		total += p.Count
	}

	report := evolve.GenerateReport(patterns, total)
	fmt.Println(report)

	if len(patterns) > 0 {
		// Write report to file
		reportPath := filepath.Join(dir, windsurfDir, "orchestrator", "evolution_report.md")
		if err := os.WriteFile(reportPath, []byte(report), 0644); err == nil {
			fmt.Printf("Report saved to: %s\n", reportPath)
		}
	}
}

// --- skills command ---

func cmdSkills() {
	dir := getDir()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("✗ Not initialized. Run `aiops init` first.")
		os.Exit(1)
	}

	fmt.Printf("aiops skills — generating skill scaffolds for %s\n\n", cfg.Project.Name)

	// Show what would be generated
	detected := skills.DetectSkills(cfg)
	if len(detected) == 0 {
		fmt.Println("No skills detected for this project's stack.")
		return
	}

	fmt.Println("Detected skills to generate:")
	for _, s := range detected {
		fmt.Printf("  - @%s — %s\n", s.Name, s.Description)
	}

	fmt.Println()
	if !confirm("Generate these skill scaffolds?") {
		fmt.Println("Aborted.")
		return
	}

	created, err := skills.GenerateSkills(dir, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating skills: %v\n", err)
		os.Exit(1)
	}

	if len(created) == 0 {
		fmt.Println("\n✓ All skills already exist. Nothing to generate.")
		return
	}

	fmt.Println()
	for _, f := range created {
		fmt.Printf("  ✓ %s\n", f)
	}
	fmt.Printf("\n✅ Generated %d skill scaffolds.\n", len(created))
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review and customize each SKILL.md")
	fmt.Println("  2. Add reference files (e.g., design-reference.md, patterns.md)")
	fmt.Println("  3. Skills are auto-invoked by Cascade based on task description")
}

// --- sync command ---

func cmdSync() {
	dir := getDir()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("✗ Not initialized. Run `aiops init` first.")
		os.Exit(1)
	}

	fmt.Printf("aiops sync — %s\n\n", cfg.Project.Name)

	// Re-scan MCPs
	stack, err := scanner.Scan(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	// Re-detect targets
	targets := target.Detect(dir)
	var targetNames []string
	for _, t := range targets {
		targetNames = append(targetNames, t.Name)
	}

	// Compare MCPs
	oldMCPs := map[string]bool{}
	for _, m := range cfg.Detected.MCPServers {
		oldMCPs[m.Name] = true
	}
	newMCPs := map[string]bool{}
	for _, m := range stack.MCPServers {
		newMCPs[m.Name] = true
	}

	var added, removed []string
	for _, m := range stack.MCPServers {
		if !oldMCPs[m.Name] {
			added = append(added, m.Name+" ("+m.Source+")")
		}
	}
	for _, m := range cfg.Detected.MCPServers {
		if !newMCPs[m.Name] {
			removed = append(removed, m.Name+" ("+m.Source+")")
		}
	}

	// Compare targets
	oldTargets := map[string]bool{}
	for _, t := range cfg.Paths.Targets {
		oldTargets[t] = true
	}
	var addedTargets, removedTargets []string
	for _, t := range targetNames {
		if !oldTargets[t] {
			addedTargets = append(addedTargets, t)
		}
	}
	newTargets := map[string]bool{}
	for _, t := range targetNames {
		newTargets[t] = true
	}
	for _, t := range cfg.Paths.Targets {
		if !newTargets[t] {
			removedTargets = append(removedTargets, t)
		}
	}

	// Re-detect maturity
	newMaturity := scanner.DetectMaturity(dir)
	maturityChanged := newMaturity != cfg.Project.Maturity

	// Re-detect skills and specs
	newSkills := scanner.DetectSkills(dir)
	newSpecs := scanner.DetectSpecs(dir)
	skillsChanged := len(newSkills) != len(cfg.Detected.Skills)
	specsChanged := len(newSpecs) != len(cfg.Detected.Specs)

	hasChanges := len(added) > 0 || len(removed) > 0 || len(addedTargets) > 0 || len(removedTargets) > 0 || maturityChanged || skillsChanged || specsChanged

	if !hasChanges {
		fmt.Println("✓ No changes detected. MCPs, targets, maturity, skills, and specs are up to date.")
		fmt.Printf("  MCP servers: %d\n", len(cfg.Detected.MCPServers))
		fmt.Printf("  Targets: %s\n", strings.Join(cfg.Paths.Targets, ", "))
		fmt.Printf("  Maturity: %s\n", cfg.Project.Maturity)
		fmt.Printf("  Skills: %d\n", len(cfg.Detected.Skills))
		fmt.Printf("  Specs: %d\n", len(cfg.Detected.Specs))
		return
	}

	// Report changes
	for _, name := range added {
		fmt.Printf("  + MCP added: %s\n", name)
	}
	for _, name := range removed {
		fmt.Printf("  - MCP removed: %s\n", name)
	}
	for _, name := range addedTargets {
		fmt.Printf("  + Target added: %s\n", name)
	}
	for _, name := range removedTargets {
		fmt.Printf("  - Target removed: %s\n", name)
	}
	if maturityChanged {
		fmt.Printf("  ↑ Maturity changed: %s → %s\n", cfg.Project.Maturity, newMaturity)
	}
	if skillsChanged {
		fmt.Printf("  ↑ Skills changed: %d → %d\n", len(cfg.Detected.Skills), len(newSkills))
	}
	if specsChanged {
		fmt.Printf("  ↑ Specs changed: %d → %d\n", len(cfg.Detected.Specs), len(newSpecs))
	}

	// Update config
	cfg.Detected.MCPServers = stack.MCPServers
	cfg.Detected.Skills = newSkills
	cfg.Detected.Specs = newSpecs
	cfg.Paths.Targets = targetNames
	cfg.Project.Maturity = newMaturity

	if err := config.Save(dir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	// Re-render all artifacts
	fmt.Println("\nRe-rendering artifacts...")
	files, err := renderer.RenderAll(dir, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Synced. %d files updated.\n", len(files))
}

// --- uninstall command ---

func cmdUninstall() {
	dir := getDir()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("AIops is not installed in this repository.")
		os.Exit(1)
	}

	// Check for --yes flag
	autoConfirm := false
	for _, arg := range os.Args {
		if arg == "--yes" || arg == "-y" {
			autoConfirm = true
		}
	}

	// Collect all paths to remove
	type removal struct {
		path  string
		label string
		isDir bool
	}
	var removals []removal

	// .aiops.yaml config
	configPath := filepath.Join(dir, ".aiops.yaml")
	if _, err := os.Stat(configPath); err == nil {
		removals = append(removals, removal{configPath, ".aiops.yaml", false})
	}

	// .aiops/ directory (soul.md, soul.local.md, kill switch)
	aiopsDir := filepath.Join(dir, ".aiops")
	if _, err := os.Stat(aiopsDir); err == nil {
		removals = append(removals, removal{aiopsDir, ".aiops/", true})
	}

	// decisions/ directory (only if it contains the seed file and nothing else)
	decisionsDir := filepath.Join(dir, "decisions")
	if entries, err := os.ReadDir(decisionsDir); err == nil {
		if len(entries) == 1 && entries[0].Name() == "0001-aiops-initialized.md" {
			removals = append(removals, removal{decisionsDir, "decisions/ (seed only)", true})
		}
	}

	// multiagency/ directory
	multiagencyDir := cfg.Paths.Multiagency
	if multiagencyDir == "" {
		multiagencyDir = "multiagency"
	}
	multiagencyPath := filepath.Join(dir, multiagencyDir)
	if _, err := os.Stat(multiagencyPath); err == nil {
		removals = append(removals, removal{multiagencyPath, multiagencyDir + "/", true})
	}

	// Per-target artifacts
	targets := target.Detect(dir)
	for _, t := range targets {
		// Repo rules file
		if t.RepoRulesPath != "" {
			p := filepath.Join(dir, t.RepoRulesPath)
			if _, err := os.Stat(p); err == nil {
				removals = append(removals, removal{p, t.RepoRulesPath, false})
			}
		}
		// Workflows dir (only aiops-generated files)
		if t.WorkflowsDir != "" {
			wfDir := filepath.Join(dir, t.WorkflowsDir)
			for _, name := range []string{"default-mode.md", "multiagency.md", "orchestrator.md"} {
				p := filepath.Join(wfDir, name)
				if _, err := os.Stat(p); err == nil {
					removals = append(removals, removal{p, filepath.Join(t.WorkflowsDir, name), false})
				}
			}
		}
		// Orchestrator dir
		if t.OrchestrDir != "" {
			p := filepath.Join(dir, t.OrchestrDir)
			if _, err := os.Stat(p); err == nil {
				removals = append(removals, removal{p, t.OrchestrDir + "/", true})
			}
		}
	}

	if len(removals) == 0 {
		fmt.Println("Nothing to remove.")
		return
	}

	fmt.Println("This will remove AIops from this repository.")
	fmt.Println()
	fmt.Println("The following will be deleted:")
	for _, r := range removals {
		fmt.Printf("  - %s\n", r.label)
	}
	fmt.Println("\nGlobal tools and binaries will NOT be removed.")

	if !autoConfirm {
		fmt.Println()
		if !confirm("Proceed?") {
			fmt.Println("Aborted.")
			return
		}
	}

	// Remove everything
	removed := 0
	for _, r := range removals {
		var err error
		if r.isDir {
			err = os.RemoveAll(r.path)
		} else {
			err = os.Remove(r.path)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed to remove %s: %v\n", r.label, err)
		} else {
			fmt.Printf("  ✓ Removed %s\n", r.label)
			removed++
		}
	}

	fmt.Printf("\n✅ AIops uninstalled. %d items removed.\n", removed)
}

// --- doctor command ---

func cmdDoctor() {
	dir := getDir()

	fmt.Println("aiops doctor — checking installation integrity")
	fmt.Println()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Println("  ✗ .aiops.yaml — missing (not initialized)")
		fmt.Println("\nRun `aiops init` to initialize.")
		os.Exit(1)
	}

	passed := 0
	warned := 0
	failed := 0

	pass := func(label string) {
		fmt.Printf("  ✓ %s\n", label)
		passed++
	}
	warn := func(label, msg string) {
		fmt.Printf("  ⚠ %s — %s\n", label, msg)
		warned++
	}
	fail := func(label, msg string) {
		fmt.Printf("  ✗ %s — %s\n", label, msg)
		failed++
	}

	// 1. Config
	pass(".aiops.yaml")

	// 2. Soul
	soulPath := filepath.Join(dir, ".aiops", "soul.md")
	if _, err := os.Stat(soulPath); err == nil {
		// Verify it matches the canonical soul (not tampered)
		canonical, readErr := renderer.GetTemplateFS().ReadFile("templates/soul/soul.md")
		if readErr == nil {
			actual, _ := os.ReadFile(soulPath)
			if string(actual) == string(canonical) {
				pass("soul.md (canonical)")
			} else {
				warn("soul.md", "modified from canonical — run `aiops sync` to restore")
			}
		} else {
			pass("soul.md")
		}
	} else {
		fail("soul.md", "missing — run `aiops sync` to create")
	}

	// 3. Soul local (optional)
	soulLocalPath := filepath.Join(dir, ".aiops", "soul.local.md")
	if _, err := os.Stat(soulLocalPath); err == nil {
		pass("soul.local.md (optional)")
	} else {
		warn("soul.local.md", "not created — optional, run `aiops sync` to generate template")
	}

	// 4. Kill switch check
	disabledPath := filepath.Join(dir, ".aiops", "disabled")
	if _, err := os.Stat(disabledPath); err == nil {
		warn("kill switch", "ACTIVE — .aiops/disabled exists, all orchestration is disabled")
	} else {
		pass("kill switch (inactive)")
	}

	// 5. Repo rules per target
	targets := target.Detect(dir)
	for _, t := range targets {
		if t.RepoRulesPath != "" {
			p := filepath.Join(dir, t.RepoRulesPath)
			if _, err := os.Stat(p); err == nil {
				pass(fmt.Sprintf("repo rules (%s)", t.DisplayName))
			} else {
				fail(fmt.Sprintf("repo rules (%s)", t.DisplayName), "missing")
			}
		}
	}

	// 6. Workflows
	windsurfDir := cfg.Paths.Windsurf
	if windsurfDir == "" {
		windsurfDir = ".windsurf"
	}
	for _, wf := range []string{"default-mode.md", "multiagency.md", "orchestrator.md"} {
		p := filepath.Join(dir, windsurfDir, "workflows", wf)
		if _, err := os.Stat(p); err == nil {
			pass(fmt.Sprintf("workflow/%s", wf))
		} else {
			fail(fmt.Sprintf("workflow/%s", wf), "missing")
		}
	}

	// 7. Orchestrator
	statePath := filepath.Join(dir, windsurfDir, "orchestrator", "session_state.yaml")
	if _, err := os.Stat(statePath); err == nil {
		pass("session_state.yaml")
	} else {
		fail("session_state.yaml", "missing")
	}

	// 8. Decisions directory
	decisionsDir := filepath.Join(dir, "decisions")
	if _, err := os.Stat(decisionsDir); err == nil {
		pass("decisions/")
	} else {
		warn("decisions/", "not found — run `aiops sync` to create")
	}

	// 9. Multiagency module
	multiagencyDir := cfg.Paths.Multiagency
	if multiagencyDir == "" {
		multiagencyDir = "multiagency"
	}
	multiagencyMod := filepath.Join(dir, multiagencyDir, "go.mod")
	if _, err := os.Stat(multiagencyMod); err == nil {
		pass("multiagency/go.mod")
	} else {
		warn("multiagency/go.mod", "not found")
	}

	// 10. Version check
	if cfg.Version != config.Version {
		warn("version", fmt.Sprintf("config says %s, binary is %s — run `aiops update`", cfg.Version, config.Version))
	} else {
		pass(fmt.Sprintf("version (%s)", config.Version))
	}

	// Summary
	fmt.Printf("\n%d passed, %d warnings, %d failed\n", passed, warned, failed)
	if failed > 0 {
		fmt.Println("\nRun `aiops sync` to fix missing artifacts.")
		os.Exit(1)
	} else if warned > 0 {
		fmt.Println("\n⚠ Some warnings detected. Review above.")
	} else {
		fmt.Println("\n✅ Installation is healthy.")
	}
}

// --- helpers ---

func printDetected(stack *config.DetectedStack) {
	fmt.Println("Detected:")

	if len(stack.Languages) > 0 {
		fmt.Print("  Languages: ")
		names := make([]string, len(stack.Languages))
		for i, l := range stack.Languages {
			names[i] = l.Name
			if l.EntryDir != "" {
				names[i] += " (" + l.EntryDir + ")"
			}
		}
		fmt.Println(strings.Join(names, ", "))
	}

	if len(stack.Frameworks) > 0 {
		fmt.Print("  Frameworks: ")
		names := make([]string, len(stack.Frameworks))
		for i, f := range stack.Frameworks {
			names[i] = f.Name
			if f.Dir != "" && f.Dir != "." {
				names[i] += " (" + f.Dir + ")"
			}
		}
		fmt.Println(strings.Join(names, ", "))
	}

	if len(stack.Build.Commands) > 0 {
		fmt.Print("  Build: ")
		fmt.Println(strings.Join(stack.Build.Commands, ", "))
	}

	if len(stack.Build.GenerateCommands) > 0 {
		fmt.Print("  Generate: ")
		fmt.Println(strings.Join(stack.Build.GenerateCommands, ", "))
	}

	if len(stack.Patterns) > 0 {
		fmt.Print("  Patterns: ")
		fmt.Println(strings.Join(stack.Patterns, ", "))
	}

	if len(stack.MCPServers) > 0 {
		fmt.Print("  MCP servers: ")
		names := make([]string, len(stack.MCPServers))
		for i, s := range stack.MCPServers {
			names[i] = s.Name
			if s.Source != "" {
				names[i] += " (" + s.Source + ")"
			}
		}
		fmt.Println(strings.Join(names, ", "))
	}
}

func compareStack(old, new *config.DetectedStack) []string {
	var diffs []string

	oldLangs := map[string]bool{}
	for _, l := range old.Languages {
		oldLangs[l.Name] = true
	}
	for _, l := range new.Languages {
		if !oldLangs[l.Name] {
			diffs = append(diffs, fmt.Sprintf("New language detected: %s", l.Name))
		}
	}

	oldFws := map[string]bool{}
	for _, f := range old.Frameworks {
		oldFws[f.Name] = true
	}
	for _, f := range new.Frameworks {
		if !oldFws[f.Name] {
			diffs = append(diffs, fmt.Sprintf("New framework detected: %s", f.Name))
		}
	}

	oldPatterns := map[string]bool{}
	for _, p := range old.Patterns {
		oldPatterns[p] = true
	}
	for _, p := range new.Patterns {
		if !oldPatterns[p] {
			diffs = append(diffs, fmt.Sprintf("New pattern detected: %s", p))
		}
	}

	oldMCPs := map[string]bool{}
	for _, m := range old.MCPServers {
		oldMCPs[m.Name] = true
	}
	for _, m := range new.MCPServers {
		if !oldMCPs[m.Name] {
			diffs = append(diffs, fmt.Sprintf("New MCP server detected: %s (%s)", m.Name, m.Source))
		}
	}

	return diffs
}

func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [Y/n] ", prompt)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}

func ask(prompt, defaultVal string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}
