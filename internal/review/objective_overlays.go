package review

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
)

type objectiveMetricRule struct {
	id     string
	metric string
	min    *int
	max    *int
	delta  int
	reason string
}

type objectiveCategoryRule struct {
	id       string
	category string
	maxHits  *int
	minHits  *int
	delta    int
	reason   string
}

type objectiveOverlayProfile struct {
	family     string
	ruleID     string
	metric     []objectiveMetricRule
	categories []objectiveCategoryRule
}

func applyObjectiveOverlay(code string, score int, report analyzer.Report, options analyzer.ContextOptions) (int, string, []string, []string) {
	profile := objectiveOverlayProfileForCode(code, options)
	if profile.family == "" {
		return score, "", nil, nil
	}
	adjusted := score
	firedRuleIDs := []string{}
	reasons := []string{}
	categoryHits := objectiveCategoryHistogram(report)

	for _, rule := range profile.metric {
		value, ok := report.Metrics[rule.metric]
		if !ok {
			continue
		}
		if rule.min != nil && value < *rule.min {
			continue
		}
		if rule.max != nil && value > *rule.max {
			continue
		}
		next := clampObjectiveScore(adjusted + rule.delta)
		if next != adjusted {
			adjusted = next
			firedRuleIDs = append(firedRuleIDs, rule.id)
			reasons = append(reasons, rule.reason)
		}
	}

	for _, rule := range profile.categories {
		hits := categoryHits[rule.category]
		if rule.maxHits != nil && hits > *rule.maxHits {
			next := clampObjectiveScore(adjusted + rule.delta)
			if next != adjusted {
				adjusted = next
				firedRuleIDs = append(firedRuleIDs, rule.id)
				reasons = append(reasons, fmt.Sprintf("%s (hits=%d)", rule.reason, hits))
			}
			continue
		}
		if rule.minHits != nil && hits < *rule.minHits {
			next := clampObjectiveScore(adjusted + rule.delta)
			if next != adjusted {
				adjusted = next
				firedRuleIDs = append(firedRuleIDs, rule.id)
				reasons = append(reasons, fmt.Sprintf("%s (hits=%d)", rule.reason, hits))
			}
		}
	}

	sort.Strings(firedRuleIDs)
	return adjusted, profile.family, firedRuleIDs, reasons
}

func objectiveOverlayProfileForCode(code string, options analyzer.ContextOptions) objectiveOverlayProfile {
	key := strings.ToLower(strings.TrimSpace(code))
	if key == "" {
		return objectiveOverlayProfile{}
	}
	domainName := analyzer.DomainForContext(options)

	switch {
	case strings.Contains(key, "scene-architecture"), strings.Contains(key, "spatial-legibility"), strings.Contains(key, "entrance-exit"), strings.Contains(key, "status-dynamics"), strings.Contains(key, "conflict-design"), strings.Contains(key, "reversal"):
		return objectiveOverlayProfile{
			family: "scene-architecture",
			ruleID: "objective.family.scene-architecture",
			metric: []objectiveMetricRule{
				metricRule("objective.family.scene-architecture.topic-drift-good", "nlp_topic_drift_score", nil, intPtr(45), 1, "scene architecture reward: low topic drift"),
				metricRule("objective.family.scene-architecture.topic-drift-high", "nlp_topic_drift_score", intPtr(70), nil, -1, "scene architecture penalty: high topic drift"),
				metricRule("objective.family.scene-architecture.coref-stable", "nlp_coref_ambiguity_count", nil, intPtr(1), 1, "scene architecture reward: stable referents"),
				metricRule("objective.family.scene-architecture.coref-unstable", "nlp_coref_ambiguity_count", intPtr(3), nil, -1, "scene architecture penalty: ambiguous referents"),
			},
			categories: []objectiveCategoryRule{
				categoryRuleMax("objective.family.scene-architecture.structure-noise", "structure", 1, -1, "scene architecture penalty: structure findings"),
				categoryRuleMax("objective.family.scene-architecture.clarity-noise", "clarity", 1, -1, "scene architecture penalty: clarity findings"),
			},
		}
	case strings.Contains(key, "causal-clarity"), strings.Contains(key, "cause-and-effect"), strings.Contains(key, "chain"), strings.Contains(key, "domino"):
		return objectiveOverlayProfile{
			family: "causal-clarity",
			ruleID: "objective.family.causal-clarity",
			metric: []objectiveMetricRule{
				metricRule("objective.family.causal-clarity.claim-support-good", "nlp_claim_support_alignment", intPtr(65), nil, 1, "causal clarity reward: strong claim-support alignment"),
				metricRule("objective.family.causal-clarity.claim-support-low", "nlp_claim_support_alignment", nil, intPtr(40), -1, "causal clarity penalty: weak claim-support alignment"),
				metricRule("objective.family.causal-clarity.transition-good", "nlp_transition_marker_density", intPtr(5), nil, 1, "causal clarity reward: transition support"),
				metricRule("objective.family.causal-clarity.transition-low", "nlp_transition_marker_density", nil, intPtr(2), -1, "causal clarity penalty: transition gaps"),
			},
			categories: []objectiveCategoryRule{
				categoryRuleMax("objective.family.causal-clarity.causal-noise", "causal clarity", 1, -1, "causal clarity penalty: causal findings"),
			},
		}
	case strings.Contains(key, "prompt-reading"), strings.Contains(key, "assignment-alignment"):
		return objectiveOverlayProfile{
			family: "prompt-reading",
			ruleID: "objective.family.prompt-reading",
			metric: []objectiveMetricRule{
				metricRule("objective.family.prompt-reading.structure-good", "nlp_structural_signpost_count", intPtr(4), nil, 1, "prompt reading reward: structural signposts present"),
				metricRule("objective.family.prompt-reading.structure-low", "nlp_structural_signpost_count", nil, intPtr(1), -1, "prompt reading penalty: weak structural signposts"),
				metricRule("objective.family.prompt-reading.topic-drift-low", "nlp_topic_drift_score", nil, intPtr(45), 1, "prompt reading reward: low topic drift"),
			},
			categories: []objectiveCategoryRule{
				categoryRuleMax("objective.family.prompt-reading.structure-noise", "structure", 1, -1, "prompt reading penalty: structure findings"),
			},
		}
	case strings.Contains(key, "evidence"), strings.Contains(key, "source"), strings.Contains(key, "quote"), strings.Contains(key, "citation"), strings.Contains(key, "synthesis"):
		return objectiveOverlayProfile{
			family: "evidence-source",
			ruleID: "objective.family.evidence-source",
			metric: []objectiveMetricRule{
				metricRule("objective.family.evidence-source.coverage-good", "nlp_claim_evidence_coverage", intPtr(60), nil, 1, "evidence/source reward: strong claim-evidence coverage"),
				metricRule("objective.family.evidence-source.coverage-low", "nlp_claim_evidence_coverage", nil, intPtr(35), -1, "evidence/source penalty: weak claim-evidence coverage"),
				metricRule("objective.family.evidence-source.markers-good", "nlp_evidence_marker_count", intPtr(2), nil, 1, "evidence/source reward: explicit evidence markers"),
				metricRule("objective.family.evidence-source.markers-low", "nlp_evidence_marker_count", nil, intPtr(0), -1, "evidence/source penalty: no evidence markers"),
			},
			categories: []objectiveCategoryRule{
				categoryRuleMax("objective.family.evidence-source.evidence-noise", "evidence", 1, -1, "evidence/source penalty: evidence findings"),
			},
		}
	case strings.Contains(key, "clarity"), strings.Contains(key, "sentence"), strings.Contains(key, "readability"), strings.Contains(key, "pronoun"), strings.Contains(key, "reference"):
		return objectiveOverlayProfile{
			family: "clarity-control",
			ruleID: "objective.family.clarity-control",
			metric: []objectiveMetricRule{
				metricRule("objective.family.clarity-control.readability-good", "nlp_readability_grade", nil, intPtr(10), 1, "clarity reward: readability target"),
				metricRule("objective.family.clarity-control.readability-high", "nlp_readability_grade", intPtr(14), nil, -1, "clarity penalty: readability too high"),
				metricRule("objective.family.clarity-control.coref-good", "nlp_coref_ambiguity_count", nil, intPtr(1), 1, "clarity reward: stable referents"),
				metricRule("objective.family.clarity-control.coref-high", "nlp_coref_ambiguity_count", intPtr(3), nil, -1, "clarity penalty: referent ambiguity"),
			},
			categories: []objectiveCategoryRule{
				categoryRuleMax("objective.family.clarity-control.clarity-noise", "clarity", 1, -1, "clarity penalty: clarity findings"),
				categoryRuleMax("objective.family.clarity-control.readability-noise", "readability", 1, -1, "clarity penalty: readability findings"),
			},
		}
	default:
		if domainName == analyzer.DomainTechnical {
			return objectiveOverlayProfile{
				family: "technical-default",
				ruleID: "objective.family.technical-default",
				metric: []objectiveMetricRule{
					metricRule("objective.family.technical-default.actionability-good", "nlp_action_verb_density", intPtr(8), nil, 1, "technical reward: action-verb density"),
					metricRule("objective.family.technical-default.actionability-low", "nlp_action_verb_density", nil, intPtr(3), -1, "technical penalty: low action-verb density"),
				},
				categories: []objectiveCategoryRule{
					categoryRuleMax("objective.family.technical-default.coverage-noise", "coverage", 1, -1, "technical penalty: coverage findings"),
				},
			}
		}
		return objectiveOverlayProfile{
			family: "general-default",
			ruleID: "objective.family.general-default",
			metric: []objectiveMetricRule{
				metricRule("objective.family.general-default.topic-drift-good", "nlp_topic_drift_score", nil, intPtr(45), 1, "general reward: low topic drift"),
				metricRule("objective.family.general-default.topic-drift-high", "nlp_topic_drift_score", intPtr(70), nil, -1, "general penalty: high topic drift"),
			},
			categories: []objectiveCategoryRule{
				categoryRuleMax("objective.family.general-default.clarity-noise", "clarity", 1, -1, "general penalty: clarity findings"),
			},
		}
	}
}

func metricRule(id, metric string, min, max *int, delta int, reason string) objectiveMetricRule {
	return objectiveMetricRule{
		id:     id,
		metric: metric,
		min:    min,
		max:    max,
		delta:  delta,
		reason: reason,
	}
}

func categoryRuleMax(id, category string, maxHits int, delta int, reason string) objectiveCategoryRule {
	return objectiveCategoryRule{
		id:       id,
		category: strings.ToLower(strings.TrimSpace(category)),
		maxHits:  intPtr(maxHits),
		delta:    delta,
		reason:   reason,
	}
}

func objectiveCategoryHistogram(report analyzer.Report) map[string]int {
	out := map[string]int{}
	for _, finding := range report.Findings {
		category := strings.ToLower(strings.TrimSpace(finding.Category))
		if category == "" {
			continue
		}
		out[category]++
	}
	return out
}

func intPtr(value int) *int {
	return &value
}
