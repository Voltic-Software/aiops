package evolve

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DirectiveEntry represents a single @directive override logged in session_state.yaml.
type DirectiveEntry struct {
	Session        string `yaml:"session"`
	Directive      string `yaml:"directive"`
	Reason         string `yaml:"reason"`
	Timestamp      string `yaml:"timestamp"`
	RuleOverridden string `yaml:"rule_overridden"`
}

// SessionState is a minimal parse of session_state.yaml for reading directive_log.
type SessionState struct {
	DirectiveLog []DirectiveEntry `yaml:"directive_log"`
}

// Pattern represents a detected pattern from directive logs.
type Pattern struct {
	Rule       string
	Count      int
	Directives []DirectiveEntry
	Proposal   string
}

// AnalyzeDirectives reads the directive log and detects patterns.
func AnalyzeDirectives(projectDir, windsurfDir string) ([]Pattern, error) {
	statePath := filepath.Join(projectDir, windsurfDir, "orchestrator", "session_state.yaml")

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("no session_state.yaml found: %w", err)
	}

	var state SessionState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing session_state.yaml: %w", err)
	}

	if len(state.DirectiveLog) == 0 {
		return nil, nil
	}

	// Group by rule_overridden
	ruleCount := map[string][]DirectiveEntry{}
	for _, d := range state.DirectiveLog {
		rule := d.RuleOverridden
		if rule == "" {
			rule = "unknown"
		}
		ruleCount[rule] = append(ruleCount[rule], d)
	}

	// Detect patterns (threshold: 2+ occurrences)
	var patterns []Pattern
	for rule, directives := range ruleCount {
		if len(directives) < 2 {
			continue
		}
		p := Pattern{
			Rule:       rule,
			Count:      len(directives),
			Directives: directives,
			Proposal:   generateProposal(rule, directives),
		}
		patterns = append(patterns, p)
	}

	return patterns, nil
}

// GenerateReport produces a human-readable evolution report.
func GenerateReport(patterns []Pattern, totalDirectives int) string {
	var sb strings.Builder

	sb.WriteString("# Evolution Analysis Report\n\n")
	sb.WriteString(fmt.Sprintf("Total directives logged: %d\n", totalDirectives))
	sb.WriteString(fmt.Sprintf("Patterns detected: %d\n\n", len(patterns)))

	if len(patterns) == 0 {
		sb.WriteString("No recurring patterns found. Default rules appear well-calibrated.\n")
		return sb.String()
	}

	sb.WriteString("## Detected Patterns\n\n")
	for i, p := range patterns {
		sb.WriteString(fmt.Sprintf("### Pattern %d: `%s` overridden %d times\n\n", i+1, p.Rule, p.Count))

		sb.WriteString("**Occurrences:**\n")
		for _, d := range p.Directives {
			sb.WriteString(fmt.Sprintf("- [%s] Session `%s`: \"%s\"\n", d.Timestamp, d.Session, d.Directive))
			if d.Reason != "" {
				sb.WriteString(fmt.Sprintf("  Reason: %s\n", d.Reason))
			}
		}

		sb.WriteString(fmt.Sprintf("\n**Proposed Rule Change:**\n%s\n\n", p.Proposal))
		sb.WriteString("---\n\n")
	}

	sb.WriteString("## Next Steps\n\n")
	sb.WriteString("1. Review each proposed rule change\n")
	sb.WriteString("2. If approved, update `global_rules.md` accordingly\n")
	sb.WriteString("3. Clear the `directive_log` in `session_state.yaml`\n")

	return sb.String()
}

func generateProposal(rule string, directives []DirectiveEntry) string {
	switch rule {
	case "escalation":
		return "Consider raising the escalation threshold or increasing the escalation budget from 2 to 3 per session. Users are frequently overriding escalation decisions, suggesting the current threshold is too conservative."
	case "tier_classification":
		return "Consider adjusting tier classification criteria. Users are frequently overriding tier decisions, suggesting some task types are being classified at a higher tier than needed."
	case "intent_scope":
		return "Consider relaxing the intent guardrail for the patterns seen in these overrides. The current scope may be too narrow for common workflows."
	default:
		// Generic proposal based on directive content
		reasons := []string{}
		for _, d := range directives {
			if d.Reason != "" {
				reasons = append(reasons, d.Reason)
			}
		}
		if len(reasons) > 0 {
			return fmt.Sprintf("Rule `%s` was overridden %d times. Common reasons: %s. Consider adjusting the default to accommodate these cases.", rule, len(directives), strings.Join(reasons, "; "))
		}
		return fmt.Sprintf("Rule `%s` was overridden %d times. Review whether the default is too restrictive.", rule, len(directives))
	}
}
