package review

import (
	"fmt"
	"sort"
	"strings"
)

// ObjectiveRuleIDsFor returns deterministic objective rule identifiers used by
// objective-level score evidence and linting.
func ObjectiveRuleIDsFor(tgoCode, mappedSkill, basis string) []string {
	code := normalizeRuleToken(tgoCode)
	if code == "" {
		return nil
	}
	out := []string{
		fmt.Sprintf("objective.%s.presence", code),
	}
	skill := normalizeRuleToken(mappedSkill)
	if skill != "" {
		out = append(out, fmt.Sprintf("objective.%s.skill.%s.bridge", code, skill))
	}
	basisToken := normalizeRuleToken(basis)
	if basisToken != "" {
		out = append(out, fmt.Sprintf("objective.%s.basis.%s", code, basisToken))
	}
	sort.Strings(out)
	return out
}

func normalizeRuleToken(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, " ", "-")
	var b strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('-')
	}
	out := b.String()
	out = strings.Trim(out, "-")
	out = strings.ReplaceAll(out, "--", "-")
	return out
}
