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

	treeDef, err := store.TreeDefinitionBySlug(context.Background(), "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("tree definition: %v", err)
	}
	report, err := store.ProgressReport(context.Background(), userID, treeID, treeDef.PrioritySkills, 5)
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
	if err := store.SaveReviewArtifacts(context.Background(), domain.ReviewArtifacts{
		ReviewID:           loaded.ID,
		AnalyzerReportJSON: `{"metrics":{"word_count":6}}`,
		RecommendationJSON: `{"focus":"tragic inevitability"}`,
		ComparisonJSON:     `{"summary":"tighter than before"}`,
	}); err != nil {
		t.Fatalf("save review artifacts: %v", err)
	}
	artifacts, err := store.GetReviewArtifacts(context.Background(), loaded.ID)
	if err != nil {
		t.Fatalf("get review artifacts: %v", err)
	}
	if artifacts.RecommendationJSON == "" || artifacts.AnalyzerReportJSON == "" {
		t.Fatalf("artifacts = %#v", artifacts)
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
	if active[0].ProgressMode != domain.ProgressModePercent {
		t.Fatalf("expected percent progress mode, got %q", active[0].ProgressMode)
	}
}

func TestTGOMasterySignalUsesRollingEvidence(t *testing.T) {
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
	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "youth-writing-foundations")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	exID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Test",
		Brief:           "Brief",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"sentence variety"},
		SuccessCriteria: []string{"two"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}

	statuses := []string{"secure", "mastered", "mastered"}
	var target domain.TGO
	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos: %v", err)
	}
	target = active[1]

	for i, status := range statuses {
		subID, err := store.SaveSubmission(ctx, domain.Submission{
			UserID:     userID,
			TreeID:     treeID,
			ExerciseID: exID,
			Content:    "Sentence control improves across drafts.",
			WordCount:  5,
		})
		if err != nil {
			t.Fatalf("save submission %d: %v", i, err)
		}
		if _, err := store.SaveReview(ctx, domain.Review{
			UserID:           userID,
			TreeID:           treeID,
			SubmissionID:     subID,
			ReviewKind:       "deterministic",
			Summary:          "Summary",
			Strengths:        []string{"s"},
			Weaknesses:       []string{"w"},
			AnalyzerFindings: []string{"f"},
			TGOAssessments: []domain.TGOAssessment{
				{TGOCode: target.Code, Status: status, Evidence: "evidence"},
			},
			NextFocus:       target.Title,
			MetricWordCount: 5,
		}, nil); err != nil {
			t.Fatalf("save review %d: %v", i, err)
		}
	}

	signal, err := store.TGOMasterySignal(ctx, enrollmentID, target, "")
	if err != nil {
		t.Fatalf("mastery signal: %v", err)
	}
	if signal.Percent == 0 || signal.EvidenceCount != 3 {
		t.Fatalf("unexpected signal: %#v", signal)
	}
	if !signal.Ready {
		t.Fatalf("signal should be ready: %#v", signal)
	}
}
