package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/review"
	"github.com/tomasino/writing-coach/internal/scoring/objective_rules"
)

func main() {
	var failures []string
	seen := map[string]bool{}

	for _, tree := range domain.PublicBuiltInTrees {
		for _, tgo := range tree.TGOs {
			code := strings.TrimSpace(tgo.Code)
			if code == "" {
				failures = append(failures, fmt.Sprintf("tree=%s has objective with empty code", tree.Slug))
				continue
			}
			if seen[code] {
				continue
			}
			seen[code] = true

			skill := strings.TrimSpace(domain.TGOCodeToSkill[code])
			if skill == "" {
				failures = append(failures, fmt.Sprintf("objective=%s missing TGOCodeToSkill mapping", code))
				continue
			}

			ids := review.ObjectiveRuleIDsFor(code, skill, "deterministic_skill_bridge")
			if len(ids) == 0 {
				failures = append(failures, fmt.Sprintf("objective=%s resolved zero rule ids", code))
				continue
			}
			expected := fmt.Sprintf("objective.%s.presence", code)
			foundExpected := false
			for _, id := range ids {
				if strings.TrimSpace(id) == "" {
					failures = append(failures, fmt.Sprintf("objective=%s includes empty rule id", code))
					continue
				}
				if id == expected {
					foundExpected = true
				}
			}
			if !foundExpected {
				failures = append(failures, fmt.Sprintf("objective=%s missing canonical rule id %q", code, expected))
			}
			if strings.HasPrefix(code, "academic-") && !objective_rules.HasAnyForCodeDomain(code, analyzer.DomainAcademic) {
				failures = append(failures, fmt.Sprintf("objective=%s missing academic objective-rules manifest coverage", code))
			}
			if strings.HasPrefix(code, "technical-") && !objective_rules.HasAnyForCodeDomain(code, analyzer.DomainTechnical) {
				failures = append(failures, fmt.Sprintf("objective=%s missing technical objective-rules manifest coverage", code))
			}
			if tree.Slug == "professional-writing-track" && !objective_rules.HasAnyForCodeDomain(code, analyzer.DomainProfessional) {
				failures = append(failures, fmt.Sprintf("objective=%s missing professional objective-rules manifest coverage", code))
			}
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		for _, msg := range failures {
			fmt.Fprintf(os.Stderr, "objective-rule-lint: %s\n", msg)
		}
		os.Exit(1)
	}

	fmt.Printf("objective-rule-lint: ok (%d public objectives)\n", len(seen))
}
