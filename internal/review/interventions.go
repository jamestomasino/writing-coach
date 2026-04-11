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

type InterventionOutcome struct {
	Target             string   `json:"target"`
	Status             string   `json:"status"`
	Source             string   `json:"source"`
	ReasonCodes        []string `json:"reason_codes"`
	IntroducedReviewID int64    `json:"introduced_review_id,omitempty"`
	ResolvedReviewID   int64    `json:"resolved_review_id,omitempty"`
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

func BuildInterventionOutcomes(current []Intervention, prior []Intervention, comparison *Comparison, previousReviewID, currentReviewID int64) []InterventionOutcome {
	currentByKey := map[string]Intervention{}
	for _, item := range current {
		currentByKey[normalizeInterventionKey(item.Target)] = item
	}
	priorByKey := map[string]Intervention{}
	for _, item := range prior {
		priorByKey[normalizeInterventionKey(item.Target)] = item
	}

	persisting := map[string]bool{}
	addressed := map[string]bool{}
	if comparison != nil {
		for _, value := range comparison.PersistingWeaknesses {
			key := normalizeInterventionKey(value)
			if key != "" {
				persisting[key] = true
			}
		}
		for _, value := range comparison.AddressedWeaknesses {
			key := normalizeInterventionKey(value)
			if key != "" {
				addressed[key] = true
			}
		}
	}

	outcomes := make([]InterventionOutcome, 0, len(current)+len(prior))
	for _, item := range current {
		key := normalizeInterventionKey(item.Target)
		status := "introduced"
		reasonCodes := append([]string{}, item.ReasonCodes...)
		introducedReviewID := currentReviewID
		if _, ok := priorByKey[key]; ok || persisting[key] {
			status = "persisting"
			if previousReviewID != 0 {
				introducedReviewID = previousReviewID
			}
			reasonCodes = append(reasonCodes, "carried_from_prior_review")
		}
		outcomes = append(outcomes, InterventionOutcome{
			Target:             item.Target,
			Status:             status,
			Source:             item.Source,
			ReasonCodes:        dedupeReasonCodes(reasonCodes),
			IntroducedReviewID: introducedReviewID,
		})
	}

	resolved := make([]InterventionOutcome, 0, len(prior))
	for key, item := range priorByKey {
		if _, exists := currentByKey[key]; exists {
			continue
		}
		if !addressed[key] {
			continue
		}
		resolved = append(resolved, InterventionOutcome{
			Target:             item.Target,
			Status:             "resolved",
			Source:             item.Source,
			ReasonCodes:        []string{"addressed_in_comparison"},
			IntroducedReviewID: previousReviewID,
			ResolvedReviewID:   currentReviewID,
		})
	}
	sort.Slice(resolved, func(i, j int) bool {
		return strings.ToLower(resolved[i].Target) < strings.ToLower(resolved[j].Target)
	})

	return append(outcomes, resolved...)
}

func dedupeReasonCodes(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
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
