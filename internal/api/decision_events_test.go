package api

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestEmitProgressionHoldTransitionEventWritesActivation(t *testing.T) {
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("ensure default user tree: %v", err)
	}

	exerciseID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Transition test",
		Brief:           "brief",
		Constraints:     []string{"one paragraph"},
		FocusSkills:     []string{"narrative clarity"},
		SuccessCriteria: []string{"clear causality"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := store.SaveSubmission(ctx, domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exerciseID,
		Content:    "test",
		WordCount:  1,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	reviewID, err := store.SaveReview(ctx, domain.Review{
		UserID:           userID,
		TreeID:           treeID,
		SubmissionID:     submissionID,
		ReviewKind:       "deterministic",
		Summary:          "ok",
		Strengths:        []string{"s"},
		Weaknesses:       []string{"w"},
		AnalyzerFindings: []string{"f"},
		NextFocus:        "Causal Clarity",
		MetricWordCount:  1,
	}, nil)
	if err != nil {
		t.Fatalf("save review: %v", err)
	}

	server := Server{Store: store}
	review := domain.Review{ID: reviewID, SubmissionID: submissionID}
	if err := server.emitProgressionHoldTransitionEvent(ctx, userID, treeID, enrollmentID, review, false, true, "completed_tgo_slipping"); err != nil {
		t.Fatalf("emit transition event: %v", err)
	}

	events, err := store.DecisionEventsByReview(ctx, reviewID)
	if err != nil {
		t.Fatalf("load decision events: %v", err)
	}
	found := false
	for _, event := range events {
		if event.EventType == "progression_hold_activated" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected progression_hold_activated event, got %#v", events)
	}
}
