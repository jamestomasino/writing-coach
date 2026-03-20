package review

import (
	"context"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/openai"
)

type deterministicReviewer struct{}

type Service struct {
	client     *openai.Client
	clientKind string
	analyzers  analyzer.Service
	fallback   deterministicReviewer
}

type Result struct {
	Review         domain.Review
	Scores         []domain.SkillScore
	AnalyzerReport analyzer.Report
}

func NewService(client *openai.Client, analyzers analyzer.Service) Service {
	return Service{
		client:     client,
		clientKind: "openai",
		analyzers:  analyzers,
		fallback:   deterministicReviewer{},
	}
}

func (s Service) WithClient(client *openai.Client, kind string) Service {
	s.client = client
	s.clientKind = strings.TrimSpace(kind)
	if s.clientKind == "" {
		s.clientKind = "openai"
	}
	return s
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
			reviewResult.ReviewKind = s.clientKind
			reviewResult.ProviderNote = s.clientKind
			reviewResult.AnalyzerFindings = analyzer.TopFindings(report, 6)
			return Result{Review: reviewResult, Scores: scores, AnalyzerReport: report}
		}

		reviewResult, scores = s.fallback.ReviewSubmission(ctx, sub, report, activeTGOs, completedTGOs)
		reviewResult.ReviewKind = "deterministic-fallback"
		reviewResult.ProviderNote = strings.TrimSpace(s.clientKind + ": " + err.Error())
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

	summary := "The draft shows a workable frame, but the next exercise should push harder on control, clarity, and follow-through."
	strengths := []string{
		"The submission sustains a clear attempt at a focused mode.",
		"The scene length is appropriate for a focused exercise.",
	}
	weaknesses := []string{
		"Key turns can be made more concrete and easier to follow.",
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

	scores := defaultScoresForActiveTGOs(sub.ID, activeTGOs, wordCount, avgSentenceLength, len(report.Findings))

	return domain.Review{
		SubmissionID:       sub.ID,
		Summary:            summary,
		Strengths:          strengths,
		Weaknesses:         weaknesses,
		AnalyzerFindings:   analyzer.TopFindings(report, 6),
		TGOAssessments:     deterministicAssessments(activeTGOs, report),
		CompletedTGOChecks: deterministicCompletedChecks(completedTGOs, report),
		Annotations:        deterministicAnnotations(sub.Content, activeTGOs, completedTGOs, report),
		NextFocus:          nextFocus,
		MetricWordCount:    wordCount,
	}, scores
}

func defaultScoresForActiveTGOs(submissionID int64, activeTGOs []domain.TGO, wordCount, avgSentenceLength, findingCount int) []domain.SkillScore {
	activeTGOs = ensureReviewTGOs(activeTGOs)
	seen := map[string]bool{}
	scores := []domain.SkillScore{
		{SubmissionID: submissionID, Skill: "scene architecture", Score: scoreFromWordCount(wordCount)},
		{SubmissionID: submissionID, Skill: "narrative clarity", Score: scoreFromSentenceLength(avgSentenceLength)},
	}
	for _, score := range scores {
		seen[score.Skill] = true
	}
	for _, tgo := range activeTGOs {
		skill := domain.TGOCodeToSkill[tgo.Code]
		if skill == "" || seen[skill] {
			continue
		}
		seen[skill] = true
		scores = append(scores, domain.SkillScore{
			SubmissionID: submissionID,
			Skill:        skill,
			Score:        scoreFromFindingCount(findingCount),
		})
		if len(scores) >= 4 {
			break
		}
	}
	return scores
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

func deterministicAnnotations(content string, activeTGOs []domain.TGO, completedTGOs []domain.TGO, report analyzer.Report) []domain.ReviewAnnotation {
	sentences := splitSentences(content)
	if len(sentences) == 0 {
		return nil
	}
	activeTGOs = ensureReviewTGOs(activeTGOs)
	finding := "Tighten the sentence so the dramatic movement is easier to follow."
	if findings := analyzer.TopFindings(report, 1); len(findings) > 0 {
		finding = findings[0]
	}

	annotations := make([]domain.ReviewAnnotation, 0, 4)
	for i, tgo := range activeTGOs {
		if i >= len(sentences) {
			break
		}
		annotations = append(annotations, domain.ReviewAnnotation{
			Quote:    shortQuote(sentences[i]),
			TGOCode:  tgo.Code,
			Category: annotationCategoryForTGO(tgo.Code),
			Comment:  finding,
			Severity: annotationSeverity(report.Findings),
		})
	}
	if len(completedTGOs) > 0 && len(sentences) > len(annotations) {
		annotations = append(annotations, domain.ReviewAnnotation{
			Quote:    shortQuote(sentences[len(annotations)]),
			TGOCode:  completedTGOs[0].Code,
			Category: "revision",
			Comment:  "This line is part of the lighter completed-skill maintenance pass. Keep the previously established control visible here.",
			Severity: "low",
		})
	}
	return annotations
}

func splitSentences(content string) []string {
	parts := strings.FieldsFunc(content, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		cleaned := strings.Join(strings.Fields(strings.TrimSpace(part)), " ")
		if cleaned != "" {
			out = append(out, cleaned)
		}
	}
	return out
}

func shortQuote(sentence string) string {
	words := strings.Fields(sentence)
	if len(words) <= 14 {
		return sentence
	}
	return strings.Join(words[:14], " ")
}

func annotationCategoryForTGO(code string) string {
	switch code {
	case "causal-clarity", "claim-clarity", "objective-clarity", "sentence-clarity":
		return "clarity"
	case "scene-architecture", "paragraph-control", "structural-signposting", "narrative-sequencing":
		return "structure"
	case "mythic-register", "tone-calibration", "authority-and-voice":
		return "tone"
	case "image-freshness", "descriptive-specificity", "word-choice":
		return "imagery"
	case "dialogue-under-strain", "dialogue-basics":
		return "dialogue"
	case "symbolic-discipline":
		return "symbolism"
	default:
		return "revision"
	}
}

func annotationSeverity(findings []analyzer.Finding) string {
	switch {
	case len(findings) >= 6:
		return "high"
	case len(findings) >= 3:
		return "medium"
	default:
		return "low"
	}
}
