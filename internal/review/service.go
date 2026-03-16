package review

import (
	"context"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/openai"
)

type deterministicReviewer struct{}

type Service struct {
	client    *openai.Client
	analyzers analyzer.Service
	fallback  deterministicReviewer
}

type Result struct {
	Review         domain.Review
	Scores         []domain.SkillScore
	AnalyzerReport analyzer.Report
}

func NewService(client *openai.Client, analyzers analyzer.Service) Service {
	return Service{
		client:    client,
		analyzers: analyzers,
		fallback:  deterministicReviewer{},
	}
}

func (s Service) ReviewSubmission(ctx context.Context, sub domain.Submission, activeTGOs []domain.TGO, completedTGOs []domain.TGO) (domain.Review, []domain.SkillScore) {
	result := s.ReviewSubmissionDetailed(ctx, sub, activeTGOs, completedTGOs)
	return result.Review, result.Scores
}

func (s Service) ReviewSubmissionDetailed(ctx context.Context, sub domain.Submission, activeTGOs []domain.TGO, completedTGOs []domain.TGO) Result {
	report := s.analyzers.Analyze(ctx, sub.Content)

	if s.client != nil && s.client.Enabled() {
		reviewResult, scores, err := s.client.ReviewSubmission(ctx, openai.ReviewRequest{
			SubmissionID:     sub.ID,
			Content:          sub.Content,
			WordCount:        sub.WordCount,
			ActiveTGOs:       activeTGOs,
			CompletedTGOs:    completedTGOs,
			AnalysisSummary:  analyzer.Summary(report),
			AnalyzerFindings: analyzer.TopFindings(report, 6),
		})
		if err == nil {
			reviewResult.ReviewKind = "openai"
			reviewResult.AnalyzerFindings = analyzer.TopFindings(report, 6)
			return Result{Review: reviewResult, Scores: scores, AnalyzerReport: report}
		}

		reviewResult, scores = s.fallback.ReviewSubmission(ctx, sub, report, activeTGOs, completedTGOs)
		reviewResult.ReviewKind = "deterministic-fallback"
		reviewResult.ProviderNote = err.Error()
		return Result{Review: reviewResult, Scores: scores, AnalyzerReport: report}
	}

	reviewResult, scores := s.fallback.ReviewSubmission(ctx, sub, report, activeTGOs, completedTGOs)
	reviewResult.ReviewKind = "deterministic"
	return Result{Review: reviewResult, Scores: scores, AnalyzerReport: report}
}

func (deterministicReviewer) ReviewSubmission(_ context.Context, sub domain.Submission, report analyzer.Report, activeTGOs []domain.TGO, completedTGOs []domain.TGO) (domain.Review, []domain.SkillScore) {
	wordCount := sub.WordCount
	sentences := report.Metrics["sentence_count"]
	avgSentenceLength := 0
	if sentences > 0 {
		avgSentenceLength = report.Metrics["avg_sentence_length"]
	}

	summary := "The draft shows a workable dramatic frame, but the next exercise should push harder on control and inevitability."
	strengths := []string{
		"The submission sustains a recognizably serious tonal register.",
		"The scene length is appropriate for a focused exercise.",
	}
	weaknesses := []string{
		"Symbolic and emotional turns can be made more concrete.",
		"Sentence rhythm likely needs stronger variation to avoid flattening the scene.",
	}
	nextFocus := "emotional compression"

	if avgSentenceLength > 24 {
		nextFocus = "narrative clarity"
	}
	if wordCount < 500 {
		nextFocus = "scene architecture"
	}
	weaknesses = append(weaknesses, analyzer.TopFindings(report, 3)...)

	scores := []domain.SkillScore{
		{SubmissionID: sub.ID, Skill: "scene architecture", Score: scoreFromWordCount(wordCount)},
		{SubmissionID: sub.ID, Skill: "narrative clarity", Score: scoreFromSentenceLength(avgSentenceLength)},
		{SubmissionID: sub.ID, Skill: "mythic tone", Score: 3},
		{SubmissionID: sub.ID, Skill: "tragic inevitability", Score: scoreFromFindingCount(len(report.Findings))},
	}

	return domain.Review{
		SubmissionID:       sub.ID,
		Summary:            summary,
		Strengths:          strengths,
		Weaknesses:         weaknesses,
		AnalyzerFindings:   analyzer.TopFindings(report, 6),
		TGOAssessments:     deterministicAssessments(activeTGOs, report),
		CompletedTGOChecks: deterministicCompletedChecks(completedTGOs, report),
		NextFocus:          nextFocus,
		MetricWordCount:    wordCount,
	}, scores
}

func scoreFromWordCount(wordCount int) int {
	switch {
	case wordCount >= 700 && wordCount <= 1000:
		return 4
	case wordCount >= 500:
		return 3
	default:
		return 2
	}
}

func scoreFromSentenceLength(avg int) int {
	switch {
	case avg >= 10 && avg <= 22:
		return 4
	case avg > 0:
		return 3
	default:
		return 1
	}
}

func scoreFromFindingCount(count int) int {
	switch {
	case count <= 1:
		return 4
	case count <= 3:
		return 3
	default:
		return 2
	}
}

func deterministicAssessments(activeTGOs []domain.TGO, report analyzer.Report) []domain.TGOAssessment {
	activeTGOs = ensureReviewTGOs(activeTGOs)
	status := "secure"
	if len(report.Findings) >= 4 {
		status = "developing"
	}
	if len(report.Findings) <= 1 {
		status = "mastered"
	}
	evidence := "Deterministic analyzer found no dominant issue."
	if findings := analyzer.TopFindings(report, 1); len(findings) > 0 {
		evidence = findings[0]
	}
	var out []domain.TGOAssessment
	for _, tgo := range activeTGOs {
		out = append(out, domain.TGOAssessment{
			TGOCode:  tgo.Code,
			Status:   status,
			Evidence: evidence,
		})
	}
	return out
}

func ensureReviewTGOs(active []domain.TGO) []domain.TGO {
	if len(active) == 3 {
		return active
	}
	var out []domain.TGO
	for _, code := range []string{"causal-clarity", "scene-architecture", "prose-precision"} {
		if tgo, ok := domain.TGOByCode(code); ok {
			out = append(out, tgo)
		}
	}
	return out
}

func deterministicCompletedChecks(completedTGOs []domain.TGO, report analyzer.Report) []domain.TGOAssessment {
	if len(completedTGOs) == 0 {
		return nil
	}
	status := "holding"
	if len(report.Findings) >= 6 {
		status = "slipping"
	}
	evidence := "Completed skills appear stable in this submission."
	if findings := analyzer.TopFindings(report, 1); len(findings) > 0 {
		evidence = findings[0]
	}
	limit := 2
	if len(completedTGOs) < limit {
		limit = len(completedTGOs)
	}
	out := make([]domain.TGOAssessment, 0, limit)
	for _, tgo := range completedTGOs[:limit] {
		out = append(out, domain.TGOAssessment{
			TGOCode:  tgo.Code,
			Status:   status,
			Evidence: evidence,
		})
	}
	return out
}
