package review

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/llm"
)

type fakeReviewLLM struct{}

func (fakeReviewLLM) Enabled() bool                             { return true }
func (fakeReviewLLM) ValidateCredentials(context.Context) error { return nil }
func (fakeReviewLLM) GenerateExercise(context.Context, llm.ExerciseRequest) (domain.Exercise, error) {
	return domain.Exercise{}, nil
}
func (fakeReviewLLM) GenerateRevisionExercise(context.Context, llm.RevisionExerciseRequest) (domain.Exercise, error) {
	return domain.Exercise{}, nil
}
func (fakeReviewLLM) ReviewSubmission(_ context.Context, input llm.ReviewRequest) (domain.Review, []domain.SkillScore, error) {
	return domain.Review{
		SubmissionID:     input.SubmissionID,
		Summary:          "Model review",
		Strengths:        []string{"strong claim"},
		Weaknesses:       []string{"tighten cadence"},
		NextFocus:        "claim clarity",
		MetricWordCount:  input.WordCount,
		TGOAssessments:   []domain.TGOAssessment{},
		Annotations:      []domain.ReviewAnnotation{},
		AnalyzerFindings: []string{},
	}, []domain.SkillScore{{SubmissionID: input.SubmissionID, Skill: "claim clarity", Score: 4}}, nil
}

func TestReviewSubmissionIncludesDeterministicAndNonAuthoritativeProviderScores(t *testing.T) {
	service := NewService(fakeReviewLLM{}, analyzer.NewService(analyzer.Heuristic{})).WithClient(fakeReviewLLM{}, "openai")
	sub := domain.Submission{ID: 55, Content: "A compact draft with clear steps.", WordCount: 220}
	result := service.ReviewSubmissionDetailedWithOptions(context.Background(), sub, []domain.TGO{{Code: "claim-clarity"}}, nil, Options{
		AnalyzerContext: analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership", WritingLanguage: "en"},
	})

	if len(result.Scores) == 0 {
		t.Fatal("expected scores")
	}
	hasDeterministic := false
	hasProvider := false
	for _, score := range result.Scores {
		if score.ScoreSource == "deterministic" {
			hasDeterministic = true
		}
		if score.ScoreSource == "llm:openai:non_authoritative" {
			hasProvider = true
		}
	}
	if !hasDeterministic {
		t.Fatal("expected deterministic scores")
	}
	if !hasProvider {
		t.Fatal("expected non-authoritative provider scores")
	}
}

func TestConstrainedCalibrationScoresAppliesBoundedAdjustmentAndFlagsConflict(t *testing.T) {
	deterministic := []domain.SkillScore{
		{SubmissionID: 9, Skill: "claim clarity", Score: 2, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
	}
	provider := []domain.SkillScore{
		{SubmissionID: 9, Skill: "claim clarity", Score: 5},
		{SubmissionID: 9, Skill: "extra skill", Score: 4},
	}

	hybrid, summary := constrainedCalibrationScores(deterministic, provider, "openai")
	if len(hybrid) != 1 {
		t.Fatalf("expected 1 hybrid score, got %d", len(hybrid))
	}
	if hybrid[0].ScoreSource != "hybrid" {
		t.Fatalf("expected hybrid source, got %q", hybrid[0].ScoreSource)
	}
	if hybrid[0].Score != 3 {
		t.Fatalf("expected bounded score 3, got %d", hybrid[0].Score)
	}
	if summary.AppliedCount != 1 || summary.AdjustedCount != 1 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if summary.ConflictCount != 1 {
		t.Fatalf("expected 1 conflict, got %d", summary.ConflictCount)
	}
	if summary.UnsupportedCount != 1 {
		t.Fatalf("expected 1 unsupported provider skill, got %d", summary.UnsupportedCount)
	}

	var evidence map[string]any
	if err := json.Unmarshal([]byte(hybrid[0].ScoreEvidenceJSON), &evidence); err != nil {
		t.Fatalf("unmarshal evidence: %v", err)
	}
	if evidence["kind"] != "bounded_calibration" {
		t.Fatalf("unexpected evidence kind: %v", evidence["kind"])
	}
	if evidence["conflict"] != true {
		t.Fatalf("expected conflict=true evidence, got %v", evidence["conflict"])
	}
}

func TestReviewSubmissionAddsHybridCalibrationStream(t *testing.T) {
	service := NewService(fakeReviewLLM{}, analyzer.NewService(analyzer.Heuristic{})).WithClient(fakeReviewLLM{}, "openai")
	sub := domain.Submission{ID: 77, Content: "A compact draft with clear steps and explicit ownership.", WordCount: 260}
	result := service.ReviewSubmissionDetailedWithOptions(context.Background(), sub, []domain.TGO{{Code: "claim-clarity"}}, nil, Options{
		AnalyzerContext: analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership", WritingLanguage: "en"},
	})

	hasHybrid := false
	hasProvider := false
	for _, score := range result.Scores {
		if score.ScoreSource == "hybrid" {
			hasHybrid = true
		}
		if score.ScoreSource == "llm:openai:non_authoritative" {
			hasProvider = true
		}
	}
	if !hasHybrid {
		t.Fatal("expected bounded hybrid calibration scores")
	}
	if !hasProvider {
		t.Fatal("expected non-authoritative provider scores")
	}
	if result.Review.ProviderNote == "" {
		t.Fatal("expected provider note to include calibration summary")
	}
}
