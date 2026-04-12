package review

import (
	"encoding/json"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

func BuildObjectiveScores(submissionID int64, activeTGOs []domain.TGO, assessments []domain.TGOAssessment, skillScores []domain.SkillScore) []domain.ObjectiveScore {
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
				if strings.TrimSpace(skillScore.ScoreVersion) != "" {
					version = strings.TrimSpace(skillScore.ScoreVersion)
				}
			}
		}
		evidence := map[string]any{
			"kind":              "objective_deterministic_bridge",
			"tgo_code":          code,
			"mapped_skill":      skill,
			"assessment_status": status,
			"basis":             sourceBasis,
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
