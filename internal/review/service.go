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

func NewService(client *openai.Client, analyzers analyzer.Service) Service {
	return Service{
		client:    client,
		analyzers: analyzers,
		fallback:  deterministicReviewer{},
	}
}

func (s Service) ReviewSubmission(ctx context.Context, sub domain.Submission) (domain.Review, []domain.SkillScore) {
	report := s.analyzers.Analyze(ctx, sub.Content)

	if s.client != nil && s.client.Enabled() {
		reviewResult, scores, err := s.client.ReviewSubmission(ctx, openai.ReviewRequest{
			SubmissionID:     sub.ID,
			Content:          sub.Content,
			WordCount:        sub.WordCount,
			AnalysisSummary:  analyzer.Summary(report),
			AnalyzerFindings: analyzer.TopFindings(report, 6),
		})
		if err == nil {
			reviewResult.ReviewKind = "openai"
			reviewResult.AnalyzerFindings = analyzer.TopFindings(report, 6)
			return reviewResult, scores
		}

		reviewResult, scores = s.fallback.ReviewSubmission(ctx, sub, report)
		reviewResult.ReviewKind = "deterministic-fallback"
		reviewResult.ProviderNote = err.Error()
		return reviewResult, scores
	}

	reviewResult, scores := s.fallback.ReviewSubmission(ctx, sub, report)
	reviewResult.ReviewKind = "deterministic"
	return reviewResult, scores
}

func (deterministicReviewer) ReviewSubmission(_ context.Context, sub domain.Submission, report analyzer.Report) (domain.Review, []domain.SkillScore) {
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
		SubmissionID:     sub.ID,
		Summary:          summary,
		Strengths:        strengths,
		Weaknesses:       weaknesses,
		AnalyzerFindings: analyzer.TopFindings(report, 6),
		NextFocus:        nextFocus,
		MetricWordCount:  wordCount,
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
