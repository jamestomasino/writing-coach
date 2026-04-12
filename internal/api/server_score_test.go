package api

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestToScoreResponsesPrefersDeterministic(t *testing.T) {
	in := []domain.SkillScore{
		{Skill: "claim clarity", Score: 2, ScoreSource: "llm:openai:non_authoritative", ScoreVersion: "llm-v1"},
		{Skill: "claim clarity", Score: 4, ScoreSource: "deterministic", ScoreVersion: "det-v1", ScoreEvidenceJSON: `{"domain":"technical","applied_rules":["word_count within [300,1200]: +1"]}`},
	}

	out := toScoreResponses(in, nil)
	if len(out) != 1 {
		t.Fatalf("expected 1 score, got %d", len(out))
	}
	if out[0].Score != 4 {
		t.Fatalf("expected deterministic score 4, got %d", out[0].Score)
	}
	if out[0].ScoreSource != "deterministic" {
		t.Fatalf("source = %q", out[0].ScoreSource)
	}
	if out[0].ScoreVersion != "det-v1" {
		t.Fatalf("version = %q", out[0].ScoreVersion)
	}
	if out[0].ScoreEvidence == nil {
		t.Fatal("expected parsed evidence")
	}
}

func TestToScoreResponsesFallsBackToLegacy(t *testing.T) {
	in := []domain.SkillScore{
		{Skill: "narrative clarity", Score: 3, ScoreSource: "llm_or_legacy", ScoreVersion: "legacy-unknown"},
		{Skill: "scene architecture", Score: 4, ScoreSource: "", ScoreVersion: ""},
	}

	out := toScoreResponses(in, nil)
	if len(out) != 2 {
		t.Fatalf("expected legacy fallback set, got %d", len(out))
	}
}

func TestToScoreResponsesFiltersToAssessmentSkills(t *testing.T) {
	in := []domain.SkillScore{
		{Skill: "narrative clarity", Score: 3, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
		{Skill: "scene architecture", Score: 4, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
		{Skill: "image freshness", Score: 4, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
	}
	assessments := []domain.TGOAssessment{
		{TGOCode: "story-causal-clarity", Status: "developing"},
		{TGOCode: "story-scene-architecture", Status: "secure"},
	}
	out := toScoreResponses(in, assessments)
	if len(out) != 2 {
		t.Fatalf("expected only focused assessment scores, got %d", len(out))
	}
	for _, score := range out {
		if score.Skill == "image freshness" {
			t.Fatalf("unexpected non-focused score included: %+v", score)
		}
	}
}
