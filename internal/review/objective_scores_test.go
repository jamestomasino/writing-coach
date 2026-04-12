package review

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestBuildObjectiveScoresCoversAllActiveTGOs(t *testing.T) {
	active := []domain.TGO{
		{Code: "story-causal-clarity"},
		{Code: "story-scene-architecture"},
		{Code: "story-prose-precision"},
	}
	assessments := []domain.TGOAssessment{
		{TGOCode: "story-causal-clarity", Status: "secure"},
		{TGOCode: "story-scene-architecture", Status: "developing"},
		{TGOCode: "story-prose-precision", Status: "mastered"},
	}
	scores := []domain.SkillScore{
		{SubmissionID: 10, Skill: "narrative clarity", Score: 4, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
		{SubmissionID: 10, Skill: "scene architecture", Score: 3, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
		{SubmissionID: 10, Skill: "prose precision", Score: 5, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
	}
	out := BuildObjectiveScores(10, active, assessments, scores)
	if len(out) != 3 {
		t.Fatalf("expected 3 objective scores, got %d", len(out))
	}
	for _, score := range out {
		if score.ScoreSource != "deterministic" {
			t.Fatalf("unexpected source: %+v", score)
		}
		if score.Score < 1 || score.Score > 5 {
			t.Fatalf("score out of range: %+v", score)
		}
		if score.ScoreEvidenceJSON == "" || score.ScoreEvidenceJSON == "{}" {
			t.Fatalf("missing evidence for %+v", score)
		}
	}
}

func TestBuildObjectiveScoresFallsBackToAssessmentStatus(t *testing.T) {
	active := []domain.TGO{{Code: "memoir-causal-thread"}}
	assessments := []domain.TGOAssessment{{TGOCode: "memoir-causal-thread", Status: "mastered"}}
	out := BuildObjectiveScores(21, active, assessments, nil)
	if len(out) != 1 {
		t.Fatalf("expected single score, got %d", len(out))
	}
	if out[0].Score != 5 {
		t.Fatalf("expected mastered fallback score 5, got %d", out[0].Score)
	}
}
