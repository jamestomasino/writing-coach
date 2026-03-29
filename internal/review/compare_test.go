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
			{Skill: "narrative clarity", Score: 2, ScoreSource: "deterministic"},
			{Skill: "scene architecture", Score: 3, ScoreSource: "deterministic"},
		},
		Weaknesses: []string{
			"Adverb density is elevated; check whether stronger verbs can carry more of the weight.",
			"Relationship clarity is weak.",
		},
	}
	currentReview := domain.Review{
		SkillScores: []domain.SkillScore{
			{Skill: "narrative clarity", Score: 4, ScoreSource: "deterministic"},
			{Skill: "scene architecture", Score: 2, ScoreSource: "deterministic"},
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
}
