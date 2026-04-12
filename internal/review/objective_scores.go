package review

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

func BuildObjectiveScores(
	submissionID int64,
	activeTGOs []domain.TGO,
	assessments []domain.TGOAssessment,
	skillScores []domain.SkillScore,
	report analyzer.Report,
	options analyzer.ContextOptions,
) []domain.ObjectiveScore {
	if len(activeTGOs) == 0 {
		return nil
	}
	deterministicSkill := map[string]domain.SkillScore{}
	for _, score := range skillScores {
		if strings.TrimSpace(score.ScoreSource) != "deterministic" {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(score.Skill))
		if key == "" {
			continue
		}
		if _, exists := deterministicSkill[key]; exists {
			continue
		}
		deterministicSkill[key] = score
	}
	assessmentByCode := map[string]domain.TGOAssessment{}
	for _, assessment := range assessments {
		code := strings.TrimSpace(assessment.TGOCode)
		if code == "" {
			continue
		}
		assessmentByCode[code] = assessment
	}

	out := make([]domain.ObjectiveScore, 0, len(activeTGOs))
	for _, tgo := range activeTGOs {
		code := strings.TrimSpace(tgo.Code)
		if code == "" {
			continue
		}
		skill := strings.TrimSpace(domain.TGOCodeToSkill[code])
		score := 3
		sourceBasis := "status_fallback"
		version := "obj-det-v1"
		status := "developing"
		var bridgedSkillScore *domain.SkillScore
		if item, ok := assessmentByCode[code]; ok {
			status = strings.TrimSpace(item.Status)
			if derived := scoreFromAssessmentStatus(status); derived > 0 {
				score = derived
			}
		}
		if skill != "" {
			if skillScore, ok := deterministicSkill[strings.ToLower(skill)]; ok {
				score = clampObjectiveScore(skillScore.Score)
				sourceBasis = "deterministic_skill_bridge"
				copy := skillScore
				bridgedSkillScore = &copy
				if strings.TrimSpace(skillScore.ScoreVersion) != "" {
					version = strings.TrimSpace(skillScore.ScoreVersion)
				}
			}
		}
		overlayFamily := ""
		overlayRuleIDs := []string{}
		overlayReasons := []string{}
		overlayOrigin := ""

		if manifestScore, manifestRuleID, manifestRuleIDs, manifestReasons, ok := applyManifestObjectiveOverlay(code, score, report, options); ok {
			score = manifestScore
			overlayFamily = manifestRuleID
			overlayRuleIDs = append(overlayRuleIDs, manifestRuleIDs...)
			overlayReasons = append(overlayReasons, manifestReasons...)
			overlayOrigin = "manifest"
		} else {
			overlayScore, family, familyRuleIDs, familyReasons := applyObjectiveOverlay(code, score, report, options)
			score = overlayScore
			overlayFamily = family
			overlayRuleIDs = append(overlayRuleIDs, familyRuleIDs...)
			overlayReasons = append(overlayReasons, familyReasons...)
			overlayOrigin = "family_fallback"
		}
		ruleIDs := ObjectiveRuleIDsFor(code, skill, sourceBasis)
		if strings.TrimSpace(overlayFamily) != "" {
			ruleIDs = append(ruleIDs, fmt.Sprintf("objective.overlay.%s", normalizeRuleToken(overlayFamily)))
		}
		ruleIDs = append(ruleIDs, overlayRuleIDs...)
		sort.Strings(ruleIDs)
		ruleIDs = dedupeStringSlice(ruleIDs)
		evidence := map[string]any{
			"kind":               "objective_deterministic_bridge",
			"tgo_code":           code,
			"mapped_skill":       skill,
			"assessment_status":  status,
			"basis":              sourceBasis,
			"objective_family":   overlayFamily,
			"overlay_origin":     overlayOrigin,
			"overlay_reasons":    overlayReasons,
			"objective_rule_ids": ruleIDs,
			"trigger_summary":    objectiveTriggerSummary(code, sourceBasis, status, len(ruleIDs), len(overlayRuleIDs), overlayFamily, overlayOrigin),
		}
		if bridgedSkillScore != nil {
			attachBridgedSkillEvidence(evidence, *bridgedSkillScore)
		}
		rawEvidence := "{}"
		if encoded, err := json.Marshal(evidence); err == nil {
			rawEvidence = string(encoded)
		}
		out = append(out, domain.ObjectiveScore{
			SubmissionID:      submissionID,
			TGOCode:           code,
			Score:             score,
			ScoreSource:       "deterministic",
			ScoreVersion:      version,
			ScoreEvidenceJSON: rawEvidence,
		})
	}
	return out
}

func scoreFromAssessmentStatus(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "mastered":
		return 5
	case "secure":
		return 4
	case "developing":
		return 3
	default:
		return 2
	}
}

func clampObjectiveScore(value int) int {
	if value < 1 {
		return 1
	}
	if value > 5 {
		return 5
	}
	return value
}

func objectiveTriggerSummary(code, basis, status string, ruleCount, overlayTriggeredCount int, overlayFamily, overlayOrigin string) string {
	overlayTail := ""
	if overlayTriggeredCount > 0 {
		overlayTail = fmt.Sprintf(" Objective overlay %q (%s) triggered %d rule(s).", strings.TrimSpace(overlayFamily), strings.TrimSpace(overlayOrigin), overlayTriggeredCount)
	}
	switch strings.TrimSpace(basis) {
	case "deterministic_skill_bridge":
		return fmt.Sprintf("Objective %s scored from mapped deterministic skill evidence (%d rule ids).%s", code, ruleCount, overlayTail)
	default:
		return fmt.Sprintf("Objective %s scored from assessment status %q fallback (%d rule ids).%s", code, strings.TrimSpace(status), ruleCount, overlayTail)
	}
}

func attachBridgedSkillEvidence(evidence map[string]any, score domain.SkillScore) {
	raw := strings.TrimSpace(score.ScoreEvidenceJSON)
	if raw == "" || raw == "{}" {
		return
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return
	}
	bridged := map[string]any{}
	if item, ok := decoded["rubric_id"]; ok {
		bridged["rubric_id"] = item
	}
	if item, ok := decoded["score_version"]; ok {
		bridged["score_version"] = item
	}
	if item, ok := decoded["applied_rules"]; ok {
		bridged["applied_rules"] = item
	}
	if item, ok := decoded["metric_snapshot"]; ok {
		bridged["metric_snapshot"] = item
	}
	if len(bridged) > 0 {
		evidence["bridged_skill_evidence"] = bridged
	}
}

func dedupeStringSlice(values []string) []string {
	if len(values) <= 1 {
		return values
	}
	out := make([]string, 0, len(values))
	var prev string
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if i > 0 && value == prev {
			continue
		}
		out = append(out, value)
		prev = value
	}
	return out
}
