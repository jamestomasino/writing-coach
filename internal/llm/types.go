package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

type Client interface {
	Enabled() bool
	ValidateCredentials(ctx context.Context) error
	GenerateExercise(ctx context.Context, input ExerciseRequest) (domain.Exercise, error)
	GenerateRevisionExercise(ctx context.Context, input RevisionExerciseRequest) (domain.Exercise, error)
	ReviewSubmission(ctx context.Context, input ReviewRequest) (domain.Review, []domain.SkillScore, error)
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return fmt.Sprintf("llm api: %s", e.Message)
	}
	return fmt.Sprintf("llm api: status %d", e.StatusCode)
}

type ExerciseRequest struct {
	CurrentFocus      string
	DifficultyLevel   int
	ActiveTGOs        []domain.TGO
	OnboardingProfile *domain.OnboardingProfile
	RecentTitles      []string
	RecentWeaknesses  []string
	RecurringFindings []string
	CoachingBrief     string
}

type RevisionExerciseRequest struct {
	CurrentFocus      string
	DifficultyLevel   int
	ActiveTGOs        []domain.TGO
	SubmissionID      int64
	SubmissionContent string
	Weaknesses        []string
	AnalyzerFindings  []string
	ComparisonSummary string
	RecentWeaknesses  []string
	RecurringFindings []string
	CoachingBrief     string
}

type ReviewRequest struct {
	SubmissionID     int64
	Content          string
	WordCount        int
	ActiveTGOs       []domain.TGO
	CompletedTGOs    []domain.TGO
	AnalysisSummary  string
	AnalyzerFindings []string
	CoachingBrief    string
}
