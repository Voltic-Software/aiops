package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/voltic-software/aiops/internal/config"
)

// Scan analyzes a project directory and returns the detected technology stack.
func Scan(dir string) (*config.DetectedStack, error) {
	stack := &config.DetectedStack{}

	langs := detectLanguages(dir)
	stack.Languages = langs

	frameworks := detectFrameworks(dir, langs)
	stack.Frameworks = frameworks

	stack.Build = detectBuild(dir, langs, frameworks)
	stack.Patterns = detectPatterns(dir)
	stack.GoModule = detectGoModule(dir)

	return stack, nil
}

// detectGoModule finds the Go module path from go.mod files.
func detectGoModule(dir string) string {
	// Check root first, then subdirectories
	modPaths := findFiles(dir, "go.mod", 2)
	for _, modPath := range modPaths {
		data, err := os.ReadFile(modPath)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				mod := strings.TrimPrefix(line, "module ")
				mod = strings.TrimSpace(mod)
				// Prefer root-level or shortest module path
				return mod
			}
		}
	}
	return ""
}

func detectLanguages(dir string) []Language {
	var langs []Language

	checks := []struct {
		indicator  string
		name       string
		confidence string
		entryDir   string
	}{
		{"go.mod", "go", "high", ""},
		{"go.sum", "go", "high", ""},
		{"package.json", "typescript", "medium", ""},
		{"tsconfig.json", "typescript", "high", ""},
		{"requirements.txt", "python", "high", ""},
		{"pyproject.toml", "python", "high", ""},
		{"Cargo.toml", "rust", "high", ""},
		{"pom.xml", "java", "high", ""},
		{"build.gradle", "java", "high", ""},
		{"Gemfile", "ruby", "high", ""},
		{"mix.exs", "elixir", "high", ""},
	}

	seen := map[string]bool{}

	// Check root
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.indicator)); err == nil {
			if !seen[c.name] {
				seen[c.name] = true
				langs = append(langs, Language{Name: c.name, Confidence: c.confidence})
			}
		}
	}

	// Check one level of subdirectories
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" || entry.Name() == "vendor" {
			continue
		}
		subdir := filepath.Join(dir, entry.Name())
		for _, c := range checks {
			if _, err := os.Stat(filepath.Join(subdir, c.indicator)); err == nil {
				if !seen[c.name] {
					seen[c.name] = true
					langs = append(langs, Language{Name: c.name, Confidence: c.confidence, EntryDir: entry.Name()})
				}
			}
		}
	}

	return toLangConfig(langs)
}

func detectFrameworks(dir string, langs []Language) []Framework {
	var frameworks []Framework

	// Go frameworks — check root and subdirectory go.mod files
	if hasLang(langs, "go") {
		goModPaths := findFiles(dir, "go.mod", 2)
		for _, modPath := range goModPaths {
			modDir := filepath.Dir(modPath)
			relDir, _ := filepath.Rel(dir, modDir)

			if containsInFile(modDir, "go.mod", "eventsrc") {
				frameworks = appendFrameworkUnique(frameworks, Framework{Name: "eventsrc", Language: "go", Confidence: "high", Dir: relDir})
			}
			if containsInFile(modDir, "go.mod", "temporal") {
				frameworks = appendFrameworkUnique(frameworks, Framework{Name: "temporal", Language: "go", Confidence: "high", Dir: relDir})
			}
			if containsInFile(modDir, "go.mod", "gin-gonic") {
				frameworks = appendFrameworkUnique(frameworks, Framework{Name: "gin", Language: "go", Confidence: "high", Dir: relDir})
			}
			if containsInFile(modDir, "go.mod", "go-chi") || containsInFile(modDir, "go.mod", "chi/v5") {
				frameworks = appendFrameworkUnique(frameworks, Framework{Name: "chi", Language: "go", Confidence: "high", Dir: relDir})
			}
			if containsInFile(modDir, "go.mod", "fiber") {
				frameworks = appendFrameworkUnique(frameworks, Framework{Name: "fiber", Language: "go", Confidence: "high", Dir: relDir})
			}
			if containsInFile(modDir, "go.mod", "eclipse/paho") {
				frameworks = appendFrameworkUnique(frameworks, Framework{Name: "mqtt", Language: "go", Confidence: "high", Dir: relDir})
			}
		}
	}

	// TypeScript/JavaScript frameworks
	if hasLang(langs, "typescript") || hasLang(langs, "javascript") {
		// Check root and subdirectories for package.json
		pkgPaths := findFiles(dir, "package.json", 2)
		for _, pkgPath := range pkgPaths {
			relDir, _ := filepath.Rel(dir, filepath.Dir(pkgPath))

			if containsInFile(filepath.Dir(pkgPath), "package.json", "\"next\"") {
				frameworks = append(frameworks, Framework{Name: "nextjs", Language: "typescript", Confidence: "high", Dir: relDir})
			}
			if containsInFile(filepath.Dir(pkgPath), "package.json", "\"react\"") && !containsInFile(filepath.Dir(pkgPath), "package.json", "\"next\"") {
				frameworks = append(frameworks, Framework{Name: "react", Language: "typescript", Confidence: "high", Dir: relDir})
			}
			if containsInFile(filepath.Dir(pkgPath), "package.json", "\"vue\"") {
				frameworks = append(frameworks, Framework{Name: "vue", Language: "typescript", Confidence: "high", Dir: relDir})
			}
			if containsInFile(filepath.Dir(pkgPath), "package.json", "\"svelte\"") {
				frameworks = append(frameworks, Framework{Name: "svelte", Language: "typescript", Confidence: "high", Dir: relDir})
			}
			if containsInFile(filepath.Dir(pkgPath), "package.json", "\"angular\"") || containsInFile(filepath.Dir(pkgPath), "package.json", "\"@angular/core\"") {
				frameworks = append(frameworks, Framework{Name: "angular", Language: "typescript", Confidence: "high", Dir: relDir})
			}
			if containsInFile(filepath.Dir(pkgPath), "package.json", "\"tailwindcss\"") {
				frameworks = append(frameworks, Framework{Name: "tailwindcss", Language: "typescript", Confidence: "high", Dir: relDir})
			}
		}
	}

	// Python frameworks
	if hasLang(langs, "python") {
		if containsInAnyFile(dir, []string{"requirements.txt", "pyproject.toml"}, "django") {
			frameworks = append(frameworks, Framework{Name: "django", Language: "python", Confidence: "high"})
		}
		if containsInAnyFile(dir, []string{"requirements.txt", "pyproject.toml"}, "fastapi") {
			frameworks = append(frameworks, Framework{Name: "fastapi", Language: "python", Confidence: "high"})
		}
		if containsInAnyFile(dir, []string{"requirements.txt", "pyproject.toml"}, "flask") {
			frameworks = append(frameworks, Framework{Name: "flask", Language: "python", Confidence: "high"})
		}
	}

	return toFrameworkConfig(frameworks)
}

func detectBuild(dir string, langs []Language, frameworks []Framework) config.BuildInfo {
	info := config.BuildInfo{}

	for _, lang := range langs {
		switch lang.Name {
		case "go":
			info.Commands = append(info.Commands, "go build ./...")
			info.TestCommands = append(info.TestCommands, "go test ./...")
			info.GeneratedFileMarks = append(info.GeneratedFileMarks,
				"// Code generated",
				"// THIS FILE WAS GENERATED; DO NOT EDIT!",
			)
		case "typescript":
			info.Commands = append(info.Commands, "npx tsc --noEmit")
			info.TestCommands = append(info.TestCommands, "npm test")
		case "python":
			info.Commands = append(info.Commands, "python -m py_compile")
			info.TestCommands = append(info.TestCommands, "pytest")
		case "rust":
			info.Commands = append(info.Commands, "cargo build")
			info.TestCommands = append(info.TestCommands, "cargo test")
		case "java":
			if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
				info.Commands = append(info.Commands, "mvn compile")
				info.TestCommands = append(info.TestCommands, "mvn test")
			} else {
				info.Commands = append(info.Commands, "gradle build")
				info.TestCommands = append(info.TestCommands, "gradle test")
			}
		}
	}

	for _, fw := range frameworks {
		switch fw.Name {
		case "eventsrc":
			info.GenerateCommands = append(info.GenerateCommands, "go run main.go generate-dsl [domain]")
		case "nextjs":
			info.Commands = appendUnique(info.Commands, "npm run build")
		}
	}

	return info
}

func detectPatterns(dir string) []string {
	var patterns []string

	// Domain-driven design — check root and one level deep
	dddPaths := []string{"internal/domain", "src/domain", "domain", "pkg/domain"}
	for _, p := range dddPaths {
		if dirExists(dir, p) {
			patterns = append(patterns, "domain-driven-design")
			break
		}
	}
	if !contains(patterns, "domain-driven-design") {
		// Check subdirectories
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			for _, p := range dddPaths {
				if dirExists(filepath.Join(dir, entry.Name()), p) {
					patterns = append(patterns, "domain-driven-design")
					break
				}
			}
			if contains(patterns, "domain-driven-design") {
				break
			}
		}
	}

	// Event sourcing — check for eventsrc package or aggregate files
	eventsrcPaths := []string{"pkg/eventsrc", "internal/eventsrc", "eventsrc"}
	for _, p := range eventsrcPaths {
		if dirExists(dir, p) {
			patterns = append(patterns, "event-sourcing")
			break
		}
	}
	if !contains(patterns, "event-sourcing") {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			for _, p := range eventsrcPaths {
				if dirExists(filepath.Join(dir, entry.Name()), p) {
					patterns = append(patterns, "event-sourcing")
					break
				}
			}
			if contains(patterns, "event-sourcing") {
				break
			}
		}
	}

	// Code generation — check for domainspec or dslgen packages
	codegenPaths := []string{"pkg/domainspec", "pkg/dslgen", "internal/codegen"}
	for _, p := range codegenPaths {
		if dirExists(dir, p) {
			patterns = append(patterns, "code-generation")
			break
		}
	}
	if !contains(patterns, "code-generation") {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			for _, p := range codegenPaths {
				if dirExists(filepath.Join(dir, entry.Name()), p) {
					patterns = append(patterns, "code-generation")
					break
				}
			}
			if contains(patterns, "code-generation") {
				break
			}
		}
	}

	// Monorepo
	if _, err := os.Stat(filepath.Join(dir, "pnpm-workspace.yaml")); err == nil {
		patterns = append(patterns, "monorepo")
	}
	if _, err := os.Stat(filepath.Join(dir, "lerna.json")); err == nil {
		patterns = append(patterns, "monorepo")
	}

	// Docker
	dockerFiles := findFiles(dir, "Dockerfile", 2)
	if len(dockerFiles) > 0 {
		patterns = append(patterns, "containerized")
	}
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
		patterns = appendUnique(patterns, "containerized")
	}
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yaml")); err == nil {
		patterns = appendUnique(patterns, "containerized")
	}

	// CI/CD
	if dirExists(dir, ".github/workflows") {
		patterns = append(patterns, "github-actions")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitlab-ci.yml")); err == nil {
		patterns = append(patterns, "gitlab-ci")
	}

	// MCP server
	if dirExists(dir, "mcp-server") {
		patterns = append(patterns, "mcp-server")
	}

	return dedupe(patterns)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// --- helpers ---

type Language = config.Language
type Framework = config.Framework

func toLangConfig(langs []Language) []config.Language {
	return langs
}

func toFrameworkConfig(fws []Framework) []config.Framework {
	return fws
}

func hasLang(langs []Language, name string) bool {
	for _, l := range langs {
		if l.Name == name {
			return true
		}
	}
	return false
}

func containsInFile(dir, filename, substr string) bool {
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}

func containsInAnyFile(dir string, filenames []string, substr string) bool {
	for _, f := range filenames {
		if containsInFile(dir, f, substr) {
			return true
		}
	}
	return false
}

func findFiles(dir string, name string, maxDepth int) []string {
	var results []string
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		depth := strings.Count(rel, string(os.PathSeparator))
		if depth > maxDepth {
			return filepath.SkipDir
		}
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == "vendor") {
			return filepath.SkipDir
		}
		if d.Name() == name {
			results = append(results, path)
		}
		return nil
	})
	return results
}

func dirExists(base, sub string) bool {
	info, err := os.Stat(filepath.Join(base, sub))
	return err == nil && info.IsDir()
}

func appendFrameworkUnique(slice []Framework, fw Framework) []Framework {
	for _, f := range slice {
		if f.Name == fw.Name {
			return slice
		}
	}
	return append(slice, fw)
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func dedupe(items []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
