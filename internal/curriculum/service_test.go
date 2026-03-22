package curriculum

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestSyncTGOsBlocksAdvancementWhenCompletedTGOSlips(t *testing.T) {
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
	_, _, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}
	if err := store.ReplaceActiveTGO(ctx, enrollmentID, 1, "story-causal-clarity", "emotional-compression"); err != nil {
		t.Fatalf("replace active tgo: %v", err)
	}

	service := NewService()
	recommendation, err := service.SyncTGOs(ctx, store, "story-craft-track", enrollmentID, domain.Review{
		CompletedTGOChecks: []domain.TGOAssessment{
			{TGOCode: "story-causal-clarity", Status: "slipping", Evidence: "Causality became hard to follow again."},
		},
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "emotional-compression", Status: "mastered", Evidence: "stable"},
		},
	})
	if err != nil {
		t.Fatalf("sync tgos: %v", err)
	}
	if recommendation.Focus != "Causal Clarity" {
		t.Fatalf("focus = %q", recommendation.Focus)
	}
	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos: %v", err)
	}
	if active[0].Code != "story-scene-architecture" {
		t.Fatalf("unexpected promotion state: %q", active[0].Code)
	}
}

func TestSyncTGOsDoesNotRotateActiveSkillsMidChainAfterMastery(t *testing.T) {
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
		t.Fatalf("default user tree: %v", err)
	}

	activeBefore, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos before: %v", err)
	}
	target := activeBefore[0]

	exerciseID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Mastery evidence",
		Brief:           "Gather enough evidence for mastery.",
		Constraints:     []string{"one scene"},
		FocusSkills:     []string{target.Title},
		TGOCodes:        []string{target.Code},
		SuccessCriteria: []string{"stable control"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	for i, status := range []string{"secure", "mastered"} {
		submissionID, err := store.SaveSubmission(ctx, domain.Submission{
			UserID:     userID,
			TreeID:     treeID,
			ExerciseID: exerciseID,
			Content:    "Evidence draft",
			WordCount:  2,
		})
		if err != nil {
			t.Fatalf("save submission %d: %v", i, err)
		}
		if _, err := store.SaveReview(ctx, domain.Review{
			UserID:           userID,
			TreeID:           treeID,
			SubmissionID:     submissionID,
			ReviewKind:       "deterministic",
			Summary:          "Evidence",
			Strengths:        []string{"stable"},
			Weaknesses:       []string{},
			AnalyzerFindings: []string{},
			TGOAssessments: []domain.TGOAssessment{
				{TGOCode: target.Code, Status: status, Evidence: "evidence"},
			},
			NextFocus:       target.Title,
			MetricWordCount: 2,
		}, nil); err != nil {
			t.Fatalf("save review %d: %v", i, err)
		}
	}

	service := NewService()
	recommendation, err := service.SyncTGOs(ctx, store, "story-craft-track", enrollmentID, domain.Review{
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: target.Code, Status: "mastered", Evidence: "stable mastery"},
		},
	})
	if err != nil {
		t.Fatalf("sync tgos: %v", err)
	}
	if recommendation.Focus == "" {
		t.Fatal("expected recommendation focus")
	}

	activeAfter, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos after: %v", err)
	}
	if len(activeAfter) != len(activeBefore) || activeAfter[0].Code != activeBefore[0].Code {
		t.Fatalf("active skills changed mid-chain: before=%q after=%q", activeBefore[0].Code, activeAfter[0].Code)
	}

	completed, err := store.CompletedTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("completed tgos: %v", err)
	}
	found := false
	for _, tgo := range completed {
		if tgo.Code == target.Code {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected mastered tgo %q to be marked completed", target.Code)
	}
}
