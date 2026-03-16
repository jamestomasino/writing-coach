package review

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestCompareSubmissions(t *testing.T) {
	baseline := domain.Submission{ID: 1, WordCount: 100, Content: "The river seal broke and the city mourned in silence."}
	current := domain.Submission{ID: 2, WordCount: 120, Content: "The river token broke and the city mourned beneath the hall rafters."}
	baselineReview := domain.Review{
		Weaknesses: []string{
			"Adverb density is elevated; check whether stronger verbs can carry more of the weight.",
			"Relationship clarity is weak.",
		},
	}
	currentReview := domain.Review{
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
}
