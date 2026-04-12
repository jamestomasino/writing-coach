package review

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/scoring/objective_rules"
)

func applyManifestObjectiveOverlay(code string, score int, report analyzer.Report, options analyzer.ContextOptions) (int, string, []string, []string, bool) {
	sets := objective_rules.ResolveAll(code, options)
	if len(sets) == 0 {
		return score, "", nil, nil, false
	}
	categoryHits := objectiveCategoryHistogram(report)
	adjusted := score
	fired := []string{}
	reasons := []string{}
	appliedSetIDs := []string{}

	for _, set := range sets {
		appliedSetIDs = append(appliedSetIDs, strings.TrimSpace(set.RuleID))
		for _, rule := range set.MetricRules {
			metric := strings.TrimSpace(rule.Metric)
			if metric == "" {
				continue
			}
			value, exists := report.Metrics[metric]
			if !exists {
				continue
			}
			if rule.Min != nil && value < *rule.Min {
				continue
			}
			if rule.Max != nil && value > *rule.Max {
				continue
			}
			next := clampObjectiveScore(adjusted + rule.Delta)
			if next != adjusted {
				adjusted = next
				fired = append(fired, strings.TrimSpace(rule.ID))
				reasons = append(reasons, strings.TrimSpace(rule.Reason))
			}
		}

		for _, rule := range set.CategoryRules {
			category := strings.ToLower(strings.TrimSpace(rule.Category))
			if category == "" {
				continue
			}
			hits := categoryHits[category]
			apply := false
			switch {
			case rule.MaxHits != nil && hits > *rule.MaxHits:
				apply = true
			case rule.MinHits != nil && hits < *rule.MinHits:
				apply = true
			}
			if !apply {
				continue
			}
			next := clampObjectiveScore(adjusted + rule.Delta)
			if next != adjusted {
				adjusted = next
				fired = append(fired, strings.TrimSpace(rule.ID))
				reasons = append(reasons, fmt.Sprintf("%s (hits=%d)", strings.TrimSpace(rule.Reason), hits))
			}
		}

		if set.TopScoreGate.RequireRuleTriggerCount > 0 && adjusted == 5 && len(fired) < set.TopScoreGate.RequireRuleTriggerCount {
			adjusted = 4
			fired = append(fired, fmt.Sprintf("%s.top-score-gate.rule-trigger-count", strings.TrimSpace(set.RuleID)))
			reasons = append(reasons, "manifest top-score gate: insufficient triggered rules")
		}
		if adjusted == 5 && len(set.TopScoreGate.MinMetrics) > 0 {
			for metric, min := range set.TopScoreGate.MinMetrics {
				value, ok := report.Metrics[metric]
				if !ok {
					continue
				}
				if value < min {
					adjusted = 4
					fired = append(fired, fmt.Sprintf("%s.top-score-gate.min-metric", strings.TrimSpace(set.RuleID)))
					reasons = append(reasons, fmt.Sprintf("manifest top-score gate: %s >= %d required", metric, min))
					break
				}
			}
		}
		if adjusted == 5 && len(set.TopScoreGate.MaxMetrics) > 0 {
			for metric, max := range set.TopScoreGate.MaxMetrics {
				value, ok := report.Metrics[metric]
				if !ok {
					continue
				}
				if value > max {
					adjusted = 4
					fired = append(fired, fmt.Sprintf("%s.top-score-gate.max-metric", strings.TrimSpace(set.RuleID)))
					reasons = append(reasons, fmt.Sprintf("manifest top-score gate: %s <= %d required", metric, max))
					break
				}
			}
		}
	}

	if len(fired) > 0 {
		sort.Strings(fired)
		fired = dedupeStringSlice(fired)
	}
	if len(appliedSetIDs) == 0 {
		return adjusted, "", fired, reasons, true
	}
	sort.Strings(appliedSetIDs)
	appliedSetIDs = dedupeStringSlice(appliedSetIDs)
	return adjusted, strings.Join(appliedSetIDs, "+"), fired, reasons, true
}
