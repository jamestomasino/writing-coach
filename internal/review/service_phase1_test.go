package review

import (
	"context"
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
