package review

import (
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

type Intervention struct {
	Rank        int      `json:"rank"`
	Target      string   `json:"target"`
	Source      string   `json:"source"`
	Impact      int      `json:"impact"`
	Effort      int      `json:"effort"`
	Confidence  int      `json:"confidence"`
	ReasonCodes []string `json:"reason_codes"`
}

func PrioritizeInterventions(current domain.Review, comparison *Comparison) []Intervention {
	type candidate struct {
		target      string
		source      string
		impact      int
		effort      int
		confidence  int
		reasonCodes []string
	}

	candidates := make([]candidate, 0, 12)
	seen := map[string]bool{}
	nextFocusKey := normalizeInterventionKey(current.NextFocus)

	addCandidate := func(target, source string, impact, effort, confidence int, reasonCodes ...string) {
		target = strings.TrimSpace(target)
		if target == "" {
			return
		}
		key := normalizeInterventionKey(target)
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		candidates = append(candidates, candidate{
			target:      target,
			source:      source,
			impact:      impact,
			effort:      effort,
			confidence:  confidence,
			reasonCodes: append([]string(nil), reasonCodes...),
		})
	}

	if comparison != nil {
		for _, weakness := range comparison.PersistingWeaknesses {
			reasons := []string{"persisting_weakness", "closure_required"}
			if nextFocusKey != "" && strings.Contains(normalizeInterventionKey(weakness), nextFocusKey) {
				reasons = append(reasons, "next_focus_alignment")
			}
			addCandidate(weakness, "comparison", 5, 3, 4, reasons...)
		}
	}

	for _, weakness := range current.Weaknesses {
		reasons := []string{"current_weakness"}
		if nextFocusKey != "" && strings.Contains(normalizeInterventionKey(weakness), nextFocusKey) {
			reasons = append(reasons, "next_focus_alignment")
		}
		addCandidate(weakness, "review", 4, 2, 4, reasons...)
	}

	for _, finding := range current.AnalyzerFindings {
		reasons := []string{"analyzer_finding"}
		if nextFocusKey != "" && strings.Contains(normalizeInterventionKey(finding), nextFocusKey) {
			reasons = append(reasons, "next_focus_alignment")
		}
		addCandidate(finding, "analyzer", 3, 2, 3, reasons...)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].impact != candidates[j].impact {
			return candidates[i].impact > candidates[j].impact
		}
		if candidates[i].effort != candidates[j].effort {
			return candidates[i].effort < candidates[j].effort
		}
		if candidates[i].confidence != candidates[j].confidence {
			return candidates[i].confidence > candidates[j].confidence
		}
		return strings.ToLower(candidates[i].target) < strings.ToLower(candidates[j].target)
	})

	if len(candidates) > 5 {
		candidates = candidates[:5]
	}

	out := make([]Intervention, 0, len(candidates))
	for i, item := range candidates {
		out = append(out, Intervention{
			Rank:        i + 1,
			Target:      item.target,
			Source:      item.source,
			Impact:      item.impact,
			Effort:      item.effort,
			Confidence:  item.confidence,
			ReasonCodes: item.reasonCodes,
		})
	}
	return out
}

func normalizeInterventionKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(' ')
	}
	parts := strings.Fields(b.String())
	if len(parts) > 7 {
		parts = parts[:7]
	}
	return strings.Join(parts, " ")
}
