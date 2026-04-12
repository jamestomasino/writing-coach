package api

import (
	"context"
	"encoding/json"
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

func TestReviewScoredEvidenceRefsJSON(t *testing.T) {
	tests := []struct {
		name      string
		skill     int
		objective int
		expect    []string
	}{
		{
			name:      "objective only",
			skill:     0,
			objective: 3,
			expect:    []string{"analyzer_report", "submission_objective_scores"},
		},
		{
			name:      "legacy only",
			skill:     3,
			objective: 0,
			expect:    []string{"analyzer_report", "submission_skill_scores"},
		},
		{
			name:      "mixed",
			skill:     3,
			objective: 3,
			expect:    []string{"analyzer_report", "submission_skill_scores", "submission_objective_scores"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := reviewScoredEvidenceRefsJSON(tc.skill, tc.objective)
			var got []string
			if err := json.Unmarshal([]byte(raw), &got); err != nil {
				t.Fatalf("unmarshal refs: %v", err)
			}
			if len(got) != len(tc.expect) {
				t.Fatalf("expected %d refs, got %d (%v)", len(tc.expect), len(got), got)
			}
			for idx := range tc.expect {
				if got[idx] != tc.expect[idx] {
					t.Fatalf("refs mismatch at %d: expected %q got %q", idx, tc.expect[idx], got[idx])
				}
			}
		})
	}
}

func TestEnsureReviewDecisionEventsBackfillsMissingAndIsIdempotent(t *testing.T) {
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
		Title:           "Backfill test",
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
	reviewID, err := store.SaveReviewWithObjectiveScores(ctx, domain.Review{
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
	}, []domain.SkillScore{
		{SubmissionID: submissionID, Skill: "clarity and coherence", Score: 3, ScoreSource: "deterministic", ScoreVersion: "det-v1", ScoreEvidenceJSON: `{"id":"fixture"}`},
	}, []domain.ObjectiveScore{
		{SubmissionID: submissionID, TGOCode: "story-causal-clarity", Score: 3, ScoreSource: "deterministic", ScoreVersion: "det-v1", ScoreEvidenceJSON: `{"id":"fixture"}`},
	})
	if err != nil {
		t.Fatalf("save review: %v", err)
	}
	if err := store.SaveReviewArtifacts(ctx, domain.ReviewArtifacts{
		ReviewID:           reviewID,
		AnalyzerReportJSON: `{}`,
		RecommendationJSON: `{"focus":"Causal Clarity","difficulty":2,"hold_active":false,"hold_reason_code":"","rationale":"keep iterating"}`,
		ComparisonJSON:     `{}`,
		AnnotationsJSON:    `[]`,
	}); err != nil {
		t.Fatalf("save review artifacts: %v", err)
	}

	reviewRow, err := store.GetReview(ctx, reviewID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}

	server := Server{Store: store}
	if err := server.ensureReviewDecisionEvents(ctx, userID, treeID, enrollmentID, reviewRow); err != nil {
		t.Fatalf("ensure decision events first run: %v", err)
	}
	if err := server.ensureReviewDecisionEvents(ctx, userID, treeID, enrollmentID, reviewRow); err != nil {
		t.Fatalf("ensure decision events second run: %v", err)
	}

	events, err := store.DecisionEventsByReview(ctx, reviewID)
	if err != nil {
		t.Fatalf("decision events by review: %v", err)
	}
	counts := map[string]int{}
	for _, event := range events {
		counts[event.EventType]++
	}
	if counts["review_scored"] != 1 {
		t.Fatalf("review_scored count = %d", counts["review_scored"])
	}
	if counts["recommendation_issued"] != 1 {
		t.Fatalf("recommendation_issued count = %d", counts["recommendation_issued"])
	}
}
