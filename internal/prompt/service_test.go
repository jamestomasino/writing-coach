package prompt

import (
	"context"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/review"
)

func TestDeterministicRevisionExercise(t *testing.T) {
	service := NewService(nil)
	sub := domain.Submission{ID: 7, DraftNumber: 2, Content: "A court scene."}
	rev := domain.Review{
		Weaknesses:       []string{"Relationship clarity is weak."},
		AnalyzerFindings: []string{"Adverb density is elevated."},
	}
	cmp := review.Comparison{
		Summary:              "Revision still carries forward several earlier weaknesses.",
		PersistingWeaknesses: []string{"Relationship clarity is weak."},
	}

	ex := service.RevisionExercise(context.Background(), Context{
		CurriculumState:    domain.CurriculumState{CurrentFocus: "tragic inevitability"},
		RevisionOf:         &sub,
		RevisionReview:     &rev,
		RevisionComparison: &cmp,
	})

	if !strings.Contains(ex.Title, "Revision") {
		t.Fatalf("unexpected title: %q", ex.Title)
	}
	if ex.SourceSubmissionID != 7 {
		t.Fatalf("source submission id = %d", ex.SourceSubmissionID)
	}
	if len(ex.Constraints) == 0 || len(ex.SuccessCriteria) == 0 {
		t.Fatal("expected populated revision brief")
	}
}

func TestDeterministicNextExerciseUsesFreshDraftLanguage(t *testing.T) {
	service := NewService(nil)

	ex := service.NextExercise(context.Background(), Context{
		CurriculumState: domain.CurriculumState{CurrentFocus: "causal clarity"},
	})

	lower := strings.ToLower(ex.Brief)
	if strings.Contains(lower, "rewrite") || strings.Contains(lower, "revise") {
		t.Fatalf("expected fresh-draft language, got %q", ex.Brief)
	}
	if !strings.Contains(lower, "write a new piece") {
		t.Fatalf("expected explicit new-piece language, got %q", ex.Brief)
	}
}
