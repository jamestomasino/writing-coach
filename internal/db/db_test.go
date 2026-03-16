package db

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestMigrateSeedAndProgress(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(context.Background(), filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(context.Background(), "Tomasino"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, treeID, _, err := store.EnsureDefaultUserTree(context.Background(), "tomasino", "Tomasino", "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	exID, err := store.SaveExercise(context.Background(), domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Test",
		Brief:           "Brief",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"tragic inevitability"},
		SuccessCriteria: []string{"two"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	subID, err := store.SaveSubmission(context.Background(), domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exID,
		Content:    "A short scene with doomed choices.",
		WordCount:  6,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	_, err = store.SaveReview(context.Background(), domain.Review{
		UserID:           userID,
		TreeID:           treeID,
		SubmissionID:     subID,
		ReviewKind:       "deterministic",
		Summary:          "Summary",
		Strengths:        []string{"s"},
		Weaknesses:       []string{"w"},
		AnalyzerFindings: []string{"f"},
		CompletedTGOChecks: []domain.TGOAssessment{
			{TGOCode: "causal-clarity", Status: "holding", Evidence: "still stable"},
		},
		NextFocus:       "tragic inevitability",
		MetricWordCount: 6,
	}, []domain.SkillScore{
		{SubmissionID: subID, Skill: "tragic inevitability", Score: 2},
		{SubmissionID: subID, Skill: "symbolic control", Score: 3},
	})
	if err != nil {
		t.Fatalf("save review: %v", err)
	}

	report, err := store.ProgressReport(context.Background(), userID, treeID, 5)
	if err != nil {
		t.Fatalf("progress report: %v", err)
	}
	if len(report) == 0 {
		t.Fatal("expected progress lines")
	}
	loaded, err := store.LatestReviewForSubmission(context.Background(), subID)
	if err != nil {
		t.Fatalf("latest review: %v", err)
	}
	if len(loaded.CompletedTGOChecks) != 1 || loaded.CompletedTGOChecks[0].TGOCode != "causal-clarity" {
		t.Fatalf("completed tgo checks = %#v", loaded.CompletedTGOChecks)
	}
}

func TestEnsureDefaultUserTreeSeedsTreeSpecificTGOs(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
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
	_, _, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "kid", "Kid", "youth-writing-foundations")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos: %v", err)
	}
	if len(active) != 3 {
		t.Fatalf("active len = %d", len(active))
	}
	if active[0].Code != "word-choice" {
		t.Fatalf("first active tgo = %q", active[0].Code)
	}
}
