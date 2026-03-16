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
	_, _, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}
	if err := store.ReplaceActiveTGO(ctx, enrollmentID, 1, "causal-clarity", "emotional-compression"); err != nil {
		t.Fatalf("replace active tgo: %v", err)
	}

	service := NewService()
	recommendation, err := service.SyncTGOs(ctx, store, "mythic-tragedy-apprenticeship", enrollmentID, domain.Review{
		CompletedTGOChecks: []domain.TGOAssessment{
			{TGOCode: "causal-clarity", Status: "slipping", Evidence: "Causality became hard to follow again."},
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
	if active[0].Code != "emotional-compression" {
		t.Fatalf("unexpected promotion state: %q", active[0].Code)
	}
}
