package review

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestBuildObjectiveScoresHasPublicTGOCoverage(t *testing.T) {
	var active []domain.TGO
	var assessments []domain.TGOAssessment
	skillSeen := map[string]bool{}
	var skillScores []domain.SkillScore

	for _, tree := range domain.PublicBuiltInTrees {
		for _, tgo := range tree.TGOs {
			active = append(active, domain.TGO{Code: tgo.Code})
			assessments = append(assessments, domain.TGOAssessment{TGOCode: tgo.Code, Status: "developing"})
			skill := strings.TrimSpace(domain.TGOCodeToSkill[tgo.Code])
			if skill == "" {
				t.Fatalf("missing TGOCodeToSkill mapping for %s in tree %s", tgo.Code, tree.Slug)
			}
			if skillSeen[skill] {
				continue
			}
			skillSeen[skill] = true
			skillScores = append(skillScores, domain.SkillScore{
				Skill:             skill,
				Score:             4,
				ScoreSource:       "deterministic",
				ScoreVersion:      "det-v1",
				ScoreEvidenceJSON: `{"rubric_id":"coverage-fixture"}`,
			})
		}
	}

	out := BuildObjectiveScores(999, active, assessments, skillScores)
	if len(out) != len(active) {
		t.Fatalf("expected objective score per active TGO; got %d for %d active TGOs", len(out), len(active))
	}

	seen := map[string]bool{}
	for _, score := range out {
		code := strings.TrimSpace(score.TGOCode)
		if code == "" {
			t.Fatalf("empty tgo code in %+v", score)
		}
		seen[code] = true
		if strings.TrimSpace(score.ScoreSource) != "deterministic" {
			t.Fatalf("expected deterministic source for %s, got %q", code, score.ScoreSource)
		}
		if score.Score < 1 || score.Score > 5 {
			t.Fatalf("score out of range for %s: %d", code, score.Score)
		}
		if strings.TrimSpace(score.ScoreEvidenceJSON) == "" || strings.TrimSpace(score.ScoreEvidenceJSON) == "{}" {
			t.Fatalf("missing evidence for %s", code)
		}
		var evidence map[string]any
		if err := json.Unmarshal([]byte(score.ScoreEvidenceJSON), &evidence); err != nil {
			t.Fatalf("decode evidence for %s: %v", code, err)
		}
		if strings.TrimSpace(anyToString(evidence["tgo_code"])) != code {
			t.Fatalf("evidence tgo_code mismatch for %s: %+v", code, evidence)
		}
	}
	for _, item := range active {
		if !seen[item.Code] {
			t.Fatalf("missing objective score for %s", item.Code)
		}
	}
}

func anyToString(value any) string {
	item, _ := value.(string)
	return item
}
