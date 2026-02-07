package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/voltic-software/aiops/internal/config"
)

// SkillDef defines a skill scaffold to generate.
type SkillDef struct {
	Name        string
	Trigger     string
	Description string
	SkillMD     string
}

// DetectSkills determines which skills should be generated based on the detected stack.
func DetectSkills(cfg *config.ProjectConfig) []SkillDef {
	var skills []SkillDef

	for _, p := range cfg.Detected.Patterns {
		switch p {
		case "domain-driven-design":
			skills = append(skills, SkillDef{
				Name:        "domain-changes",
				Trigger:     "domain modifications, aggregate changes, read model updates",
				Description: "Guide for modifying domain entities, commands, events, and read models",
				SkillMD:     domainChangesSkill(cfg),
			})
		case "event-sourcing":
			// Covered by domain-changes above if DDD is also detected
			if !hasPattern(cfg, "domain-driven-design") {
				skills = append(skills, SkillDef{
					Name:        "event-sourcing",
					Trigger:     "event sourcing changes, aggregate modifications",
					Description: "Guide for working with event-sourced aggregates",
					SkillMD:     eventSourcingSkill(cfg),
				})
			}
		}
	}

	for _, fw := range cfg.Detected.Frameworks {
		switch fw.Name {
		case "nextjs":
			skills = append(skills, SkillDef{
				Name:        "frontend-component",
				Trigger:     "React/Next.js UI development, component creation",
				Description: "Guide for developing frontend components following project patterns",
				SkillMD:     frontendSkill(cfg, fw),
			})
		case "mqtt":
			skills = append(skills, SkillDef{
				Name:        "mqtt-integration",
				Trigger:     "MQTT message flow changes, API contract modifications",
				Description: "Guide for modifying MQTT message flows and API contracts",
				SkillMD:     mqttSkill(cfg),
			})
		}
	}

	// Always include code-review skill
	skills = append(skills, SkillDef{
		Name:        "code-review",
		Trigger:     "code review, quality audit, PR review",
		Description: "Guide for performing code reviews",
		SkillMD:     codeReviewSkill(cfg),
	})

	return skills
}

// GenerateSkills writes skill scaffolds to the project.
func GenerateSkills(projectDir string, cfg *config.ProjectConfig) ([]string, error) {
	skills := DetectSkills(cfg)
	windsurfDir := cfg.Paths.Windsurf
	if windsurfDir == "" {
		windsurfDir = ".windsurf"
	}

	var created []string

	for _, skill := range skills {
		skillDir := filepath.Join(projectDir, windsurfDir, "skills", skill.Name)

		// Don't overwrite existing skills
		skillPath := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			continue // Already exists
		}

		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return nil, fmt.Errorf("creating skill dir %s: %w", skill.Name, err)
		}

		if err := os.WriteFile(skillPath, []byte(skill.SkillMD), 0644); err != nil {
			return nil, fmt.Errorf("writing SKILL.md for %s: %w", skill.Name, err)
		}

		rel, _ := filepath.Rel(projectDir, skillPath)
		created = append(created, rel)
	}

	return created, nil
}

func hasPattern(cfg *config.ProjectConfig, pattern string) bool {
	for _, p := range cfg.Detected.Patterns {
		if p == pattern {
			return true
		}
	}
	return false
}

func buildCommands(cfg *config.ProjectConfig) string {
	if len(cfg.Detected.Build.Commands) == 0 {
		return "- Run appropriate build commands after changes"
	}
	var lines []string
	for _, cmd := range cfg.Detected.Build.Commands {
		lines = append(lines, fmt.Sprintf("- `%s`", cmd))
	}
	return strings.Join(lines, "\n")
}

func generateCommands(cfg *config.ProjectConfig) string {
	if len(cfg.Detected.Build.GenerateCommands) == 0 {
		return ""
	}
	var lines []string
	lines = append(lines, "\n### Generation Commands\n")
	for _, cmd := range cfg.Detected.Build.GenerateCommands {
		lines = append(lines, fmt.Sprintf("- `%s`", cmd))
	}
	return strings.Join(lines, "\n")
}

// --- Skill templates ---

func domainChangesSkill(cfg *config.ProjectConfig) string {
	return fmt.Sprintf(`---
description: Guide for modifying domain entities, commands, events, and read models. Use when the task involves adding or changing domain entities.
---

# Domain Changes

## When to Use

This skill is auto-invoked when the task involves:
- Adding or modifying domain commands, events, or read models
- Changing aggregate behavior
- Adding new domain entities

## Procedure

1. **Load the domain schema** before making changes
2. **Understand the current structure** — commands, events, read models
3. **Make changes to the domain definition** (e.g., domain.go)
4. **Run code generation** if applicable
5. **Implement handlers** in the appropriate files
6. **Verify compilation**

### Build Commands

%s
%s

## Critical Rules

- Do not edit generated files (check for generation markers)
- One aggregate root per domain
- Command handlers validate, event applicators record
- Use generated enum types, avoid string comparisons
`, buildCommands(cfg), generateCommands(cfg))
}

func eventSourcingSkill(cfg *config.ProjectConfig) string {
	return fmt.Sprintf(`---
description: Guide for working with event-sourced aggregates
---

# Event Sourcing

## Procedure

1. Define commands and events in the domain definition
2. Run code generation
3. Implement command handlers (validation + emit events)
4. Implement event applicators (state changes)
5. Write tests using ApplyEvent to set up state

### Build Commands

%s

## Rules

- Command handlers validate, event applicators record state
- One aggregate root per domain
- Do not edit generated files
`, buildCommands(cfg))
}

func frontendSkill(cfg *config.ProjectConfig, fw config.Framework) string {
	dir := fw.Dir
	if dir == "" {
		dir = "frontend"
	}
	return fmt.Sprintf(`---
description: Guide for developing React/Next.js frontend components. Use when creating or modifying UI components, pages, or hooks.
---

# Frontend Component Development

## When to Use

This skill is auto-invoked when the task involves:
- Creating new React components
- Modifying existing UI components
- Building new pages
- Working with hooks or data fetching

## Project Structure

Frontend code is in ` + "`%s/`" + `

## Procedure

1. **Check existing components** before creating new ones
2. **Follow existing patterns** for pages and components
3. **Use existing shared components** from packages directory
4. **Maintain strict TypeScript** typing
5. **Run type check** after changes

### Build Commands

- ` + "`npx tsc --noEmit`" + ` — Type check
- ` + "`npm run build`" + ` — Full build

## Design Rules

- Follow existing component patterns
- Use existing design system components
- Maintain consistent styling with project config
- Use React Query for data fetching where applicable
`, dir)
}

func mqttSkill(cfg *config.ProjectConfig) string {
	return fmt.Sprintf(`---
description: Guide for modifying MQTT message flows and API contracts. Use when changing MQTT topics, message formats, or system integration points.
---

# MQTT Integration

## When to Use

This skill is auto-invoked when the task involves:
- Changing MQTT topics or message formats
- Modifying API endpoints
- Changing system integration points

## Procedure

1. **Map the full message flow** before making changes
2. **Identify all producers and consumers** of the affected messages
3. **Update message definitions** in the appropriate files
4. **Run code generation** if applicable
5. **Update both sides** (sender and receiver)
6. **Test end-to-end**

### Build Commands

%s

## Critical Rules

- Always update both producer and consumer when changing message format
- Consider backward compatibility for messages in flight
- Map the complete flow before making changes
`, buildCommands(cfg))
}

func codeReviewSkill(cfg *config.ProjectConfig) string {
	markers := strings.Join(cfg.Detected.Build.GeneratedFileMarks, ", ")
	if markers == "" {
		markers = "(none detected)"
	}
	return fmt.Sprintf(`---
description: Guide for performing code reviews. Use when reviewing PRs, auditing code quality, or assessing changes.
---

# Code Review

## When to Use

This skill is auto-invoked when the task involves:
- Reviewing code changes
- Auditing code quality
- Assessing PRs

## Review Checklist

1. **Correctness** — Does the code do what it claims?
2. **Style** — Does it follow existing patterns?
3. **Safety** — Are there security, performance, or reliability risks?
4. **Generated files** — Are generated files left untouched? Markers: %s
5. **Tests** — Are changes tested? Do existing tests still pass?
6. **Scope** — Does the change stay within its stated intent?

## Severity Classification

- **Critical** — Security vulnerability, data loss, broken functionality
- **High** — Incorrect behavior, performance regression, missing validation
- **Medium** — Code smell, unclear naming, missing error handling
- **Low** — Style, documentation, minor improvements
`, markers)
}
