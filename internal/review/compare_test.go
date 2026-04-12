package review

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestCompareSubmissions(t *testing.T) {
	baseline := domain.Submission{ID: 1, WordCount: 100, Content: "The river seal broke and the city mourned in silence."}
	current := domain.Submission{ID: 2, WordCount: 120, Content: "The river token broke and the city mourned beneath the hall rafters."}
	baselineReview := domain.Review{
		SkillScores: []domain.SkillScore{
			{Skill: "narrative clarity", Score: 2, ScoreSource: "deterministic", ScoreEvidenceJSON: `{"applied_rules":["top score gate: nlp_unique_token_ratio >= 54 required"]}`},
			{Skill: "scene architecture", Score: 3, ScoreSource: "deterministic", ScoreEvidenceJSON: `{"applied_rules":["finding pressure <= 2 required"]}`},
		},
		Weaknesses: []string{
			"Adverb density is elevated; check whether stronger verbs can carry more of the weight.",
			"Relationship clarity is weak.",
		},
	}
	currentReview := domain.Review{
		SkillScores: []domain.SkillScore{
			{Skill: "narrative clarity", Score: 4, ScoreSource: "deterministic", ScoreEvidenceJSON: `{"applied_rules":["top score gate: nlp_unique_token_ratio >= 54 required"]}`},
			{Skill: "scene architecture", Score: 2, ScoreSource: "deterministic", ScoreEvidenceJSON: `{"applied_rules":["finding pressure <= 2 required"]}`},
		},
		Annotations: []domain.ReviewAnnotation{
			{Quote: "The river token broke and the city mourned beneath the hall rafters.", Category: "clarity", Comment: "Narrative clarity improved around causal flow."},
			{Quote: "The city mourned beneath the hall rafters.", Category: "structure", Comment: "Scene architecture remains compressed."},
		},
		Weaknesses: []string{
			"Relationship clarity is weak.",
		},
	}

	got := CompareSubmissions(current, baseline, currentReview, baselineReview)
	if got.WordDelta != 20 {
		t.Fatalf("word delta = %d", got.WordDelta)
	}
	if len(got.AddressedWeaknesses) == 0 {
		t.Fatal("expected addressed weaknesses")
	}
	if len(got.PersistingWeaknesses) == 0 {
		t.Fatal("expected persisting weaknesses")
	}
	if len(got.SkillDeltas) == 0 {
		t.Fatal("expected skill deltas")
	}
	if got.SkillDeltas[0].Skill != "narrative clarity" && got.SkillDeltas[1].Skill != "narrative clarity" {
		t.Fatalf("expected narrative clarity delta in %+v", got.SkillDeltas)
	}
	foundNarrativeDelta := false
	for _, delta := range got.SkillDeltas {
		if delta.Skill != "narrative clarity" {
			continue
		}
		foundNarrativeDelta = true
		if !delta.DeterministicDelta {
			t.Fatal("expected deterministic delta flag for narrative clarity")
		}
		if len(delta.EvidenceQuotes) == 0 {
			t.Fatal("expected evidence quotes for narrative clarity")
		}
		if delta.DeltaExplanation == "" {
			t.Fatal("expected delta explanation")
		}
	}
	if !foundNarrativeDelta {
		t.Fatal("expected to locate narrative clarity delta details")
	}
	if got.SkillSetMismatch {
		t.Fatal("expected stable skill set across revision chain")
	}
}

func TestCompareSubmissionsFlagsSkillSetMismatch(t *testing.T) {
	baseline := domain.Submission{ID: 10, WordCount: 200, Content: "baseline content"}
	current := domain.Submission{ID: 11, WordCount: 250, Content: "current content"}
	baselineReview := domain.Review{
		SkillScores: []domain.SkillScore{
			{Skill: "narrative clarity", Score: 3, ScoreSource: "deterministic"},
			{Skill: "scene architecture", Score: 3, ScoreSource: "deterministic"},
		},
	}
	currentReview := domain.Review{
		SkillScores: []domain.SkillScore{
			{Skill: "narrative clarity", Score: 4, ScoreSource: "deterministic"},
		},
	}

	got := CompareSubmissions(current, baseline, currentReview, baselineReview)
	if !got.SkillSetMismatch {
		t.Fatal("expected skill set mismatch")
	}
	if len(got.SkillDeltas) != 1 || got.SkillDeltas[0].Skill != "narrative clarity" {
		t.Fatalf("expected overlap delta only, got %+v", got.SkillDeltas)
	}
}

func TestCompareSubmissionsFiltersSkillDeltasToActiveAssessmentSkills(t *testing.T) {
	baseline := domain.Submission{ID: 20, WordCount: 200, Content: "baseline content"}
	current := domain.Submission{ID: 21, WordCount: 240, Content: "current content"}
	baselineReview := domain.Review{
		SkillScores: []domain.SkillScore{
			{Skill: "narrative clarity", Score: 3, ScoreSource: "deterministic"},
			{Skill: "scene architecture", Score: 3, ScoreSource: "deterministic"},
			{Skill: "image freshness", Score: 4, ScoreSource: "deterministic"},
		},
	}
	currentReview := domain.Review{
		SkillScores: []domain.SkillScore{
			{Skill: "narrative clarity", Score: 4, ScoreSource: "deterministic"},
			{Skill: "scene architecture", Score: 2, ScoreSource: "deterministic"},
			{Skill: "image freshness", Score: 2, ScoreSource: "deterministic"},
		},
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "story-causal-clarity", Status: "developing"},
			{TGOCode: "story-scene-architecture", Status: "developing"},
		},
	}

	got := CompareSubmissions(current, baseline, currentReview, baselineReview)
	if len(got.SkillDeltas) != 2 {
		t.Fatalf("expected only active assessment skill deltas, got %+v", got.SkillDeltas)
	}
	for _, delta := range got.SkillDeltas {
		if delta.Skill == "image freshness" {
			t.Fatalf("unexpected non-active skill delta: %+v", delta)
		}
	}
}
