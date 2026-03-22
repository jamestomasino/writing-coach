package review

import (
	"context"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/llm"
)

type deterministicReviewer struct{}

type Service struct {
	client     llm.Client
	clientKind string
	analyzers  analyzer.Service
	fallback   deterministicReviewer
}

type Options struct {
	AnalyzerContext analyzer.ContextOptions
}

type Result struct {
	Review         domain.Review
	Scores         []domain.SkillScore
	AnalyzerReport analyzer.Report
}

func NewService(client llm.Client, analyzers analyzer.Service) Service {
	return Service{
		client:     client,
		clientKind: "openai",
		analyzers:  analyzers,
		fallback:   deterministicReviewer{},
	}
}

func (s Service) WithClient(client llm.Client, kind string) Service {
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
	return s.ReviewSubmissionDetailedWithOptions(ctx, sub, activeTGOs, completedTGOs, Options{})
}

func (s Service) ReviewSubmissionWithOptions(ctx context.Context, sub domain.Submission, activeTGOs []domain.TGO, completedTGOs []domain.TGO, options Options) (domain.Review, []domain.SkillScore) {
	result := s.ReviewSubmissionDetailedWithOptions(ctx, sub, activeTGOs, completedTGOs, options)
	return result.Review, result.Scores
}

func (s Service) ReviewSubmissionDetailedWithOptions(ctx context.Context, sub domain.Submission, activeTGOs []domain.TGO, completedTGOs []domain.TGO, options Options) Result {
	report := s.analyzers.AnalyzeWithContext(ctx, sub.Content, options.AnalyzerContext)

	if s.client != nil && s.client.Enabled() {
		reviewResult, scores, err := s.client.ReviewSubmission(ctx, llm.ReviewRequest{
			SubmissionID:     sub.ID,
			Content:          sub.Content,
			WordCount:        sub.WordCount,
			WritingLanguage:  analyzerContextLanguage(options.AnalyzerContext),
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

		reviewResult, scores = s.fallback.ReviewSubmission(ctx, sub, report, activeTGOs, completedTGOs, options.AnalyzerContext)
		reviewResult.ReviewKind = "deterministic-fallback"
		reviewResult.ProviderNote = strings.TrimSpace(s.clientKind + ": " + err.Error())
		return Result{Review: reviewResult, Scores: scores, AnalyzerReport: report}
	}

	reviewResult, scores := s.fallback.ReviewSubmission(ctx, sub, report, activeTGOs, completedTGOs, options.AnalyzerContext)
	reviewResult.ReviewKind = "deterministic"
	return Result{Review: reviewResult, Scores: scores, AnalyzerReport: report}
}

func analyzerContextLanguage(options analyzer.ContextOptions) string {
	return domain.NormalizeWritingLanguage(options.WritingLanguage)
}

func (deterministicReviewer) ReviewSubmission(_ context.Context, sub domain.Submission, report analyzer.Report, activeTGOs []domain.TGO, completedTGOs []domain.TGO, options analyzer.ContextOptions) (domain.Review, []domain.SkillScore) {
	wordCount := sub.WordCount
	sentences := report.Metrics["sentence_count"]
	avgSentenceLength := 0
	if sentences > 0 {
		avgSentenceLength = report.Metrics["avg_sentence_length"]
	}
	domainName := analyzer.DomainForContext(options)
	writingLanguage := analyzerContextLanguage(options)

	if !domain.WritingLanguageSupported(writingLanguage) {
		scores := defaultScoresForActiveTGOs(sub.ID, activeTGOs, wordCount, avgSentenceLength, len(report.Findings))
		return domain.Review{
			SubmissionID:       sub.ID,
			Summary:            "This review is limited because deterministic coaching is not configured for the selected writing language yet.",
			Strengths:          []string{"The submission was saved and can still be reviewed by a model-backed coach if one is enabled."},
			Weaknesses:         []string{"Deterministic analyzers are currently configured only for English, so this fallback review cannot make confident craft judgments yet."},
			AnalyzerFindings:   analyzer.TopFindings(report, 6),
			TGOAssessments:     deterministicAssessments(activeTGOs, report),
			CompletedTGOChecks: deterministicCompletedChecks(completedTGOs, report),
			Annotations:        nil,
			NextFocus:          "clarity and coherence",
			MetricWordCount:    wordCount,
		}, scores
	}

	summary, strengths, weaknesses, nextFocus := deterministicReviewLanguage(domainName)

	if avgSentenceLength > 24 {
		nextFocus = clarityFocus(domainName)
	}
	if wordCount < recommendedMinimumWords(domainName) {
		nextFocus = developmentFocus(domainName)
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
		Annotations:        deterministicAnnotations(sub.Content, activeTGOs, completedTGOs, report, domainName),
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
	for _, code := range []string{"story-causal-clarity", "story-scene-architecture", "story-prose-precision"} {
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

func deterministicAnnotations(content string, activeTGOs []domain.TGO, completedTGOs []domain.TGO, report analyzer.Report, domainName string) []domain.ReviewAnnotation {
	sentences := splitSentences(content)
	if len(sentences) == 0 {
		return nil
	}
	activeTGOs = ensureReviewTGOs(activeTGOs)
	finding := defaultAnnotationComment(domainName)
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
			Comment:  maintenanceComment(domainName),
			Severity: "low",
		})
	}
	return annotations
}

func deterministicReviewLanguage(domainName string) (string, []string, []string, string) {
	switch domainName {
	case analyzer.DomainTechnical:
		return "The draft has a workable instructional frame, but the next pass should strengthen scanability, explicit sequencing, and user follow-through.",
			[]string{
				"The submission is clearly trying to teach one focused task.",
				"The draft has enough structure to support a more explicit next revision.",
			},
			[]string{
				"Some steps or outcomes can be made more explicit so the reader does not have to infer the process.",
				"Sentence control and chunking likely need work so the guidance is easier to scan under pressure.",
			},
			"step clarity"
	case analyzer.DomainAcademic:
		return "The draft has a workable argumentative frame, but the next pass should strengthen claim control, evidence flow, and reasoning clarity.",
			[]string{
				"The submission is making a focused attempt at a clear argument.",
				"The draft has enough material to support a more disciplined revision pass.",
			},
			[]string{
				"Some claims and transitions can be made more explicit so the reasoning is easier to track.",
				"Sentence control likely needs tightening so the argument carries more cleanly.",
			},
			"claim development"
	case analyzer.DomainProfessional:
		return "The draft has a workable professional frame, but the next pass should sharpen ownership, clarity, and next-step control.",
			[]string{
				"The submission is aiming at a clear practical purpose.",
				"The core message is established well enough to support a more focused revision pass.",
			},
			[]string{
				"The ask, responsibility, or rationale can be made easier to grasp on first read.",
				"The message likely needs firmer sentence control so key actions stand out.",
			},
			"request clarity"
	case analyzer.DomainMarketing:
		return "The draft has a workable persuasive frame, but the next pass should sharpen the value proposition, specificity, and conversion path.",
			[]string{
				"The submission is making a focused attempt at a clear persuasive message.",
				"The piece has enough shape to support a more targeted revision pass.",
			},
			[]string{
				"The value proposition can be made more concrete and faster to understand.",
				"Sentence rhythm and emphasis likely need tightening so the copy lands more decisively.",
			},
			"value clarity"
	case analyzer.DomainThoughtLeadership:
		return "The draft has a workable idea-driven frame, but the next pass should strengthen claim progression, structure, and payoff.",
			[]string{
				"The submission is trying to advance a focused central idea.",
				"The draft has enough material to support a more deliberate revision pass.",
			},
			[]string{
				"The key claims and turns can be made more concrete and easier to follow.",
				"The piece likely needs stronger rhythm and connective structure to carry the idea progression.",
			},
			"argument flow"
	default:
		return "The draft shows a workable frame, but the next exercise should push harder on control, clarity, and follow-through.",
			[]string{
				"The submission sustains a clear attempt at a focused mode.",
				"The draft length is appropriate for a focused exercise.",
			},
			[]string{
				"Key turns can be made more concrete and easier to follow.",
				"Sentence rhythm likely needs stronger variation to avoid flattening the draft.",
			},
			"emotional compression"
	}
}

func clarityFocus(domainName string) string {
	switch domainName {
	case analyzer.DomainTechnical:
		return "instruction clarity"
	case analyzer.DomainAcademic:
		return "argument clarity"
	case analyzer.DomainProfessional:
		return "message clarity"
	case analyzer.DomainMarketing:
		return "message clarity"
	case analyzer.DomainThoughtLeadership:
		return "idea clarity"
	default:
		return "narrative clarity"
	}
}

func developmentFocus(domainName string) string {
	switch domainName {
	case analyzer.DomainTechnical:
		return "coverage depth"
	case analyzer.DomainAcademic:
		return "claim development"
	case analyzer.DomainProfessional:
		return "completeness"
	case analyzer.DomainMarketing:
		return "offer support"
	case analyzer.DomainThoughtLeadership:
		return "idea development"
	default:
		return "scene architecture"
	}
}

func recommendedMinimumWords(domainName string) int {
	switch domainName {
	case analyzer.DomainMarketing:
		return 160
	case analyzer.DomainProfessional:
		return 220
	case analyzer.DomainTechnical:
		return 260
	case analyzer.DomainAcademic, analyzer.DomainThoughtLeadership:
		return 320
	default:
		return 500
	}
}

func defaultAnnotationComment(domainName string) string {
	switch domainName {
	case analyzer.DomainTechnical:
		return "Tighten the sentence so the instruction and expected outcome are easier to follow."
	case analyzer.DomainAcademic:
		return "Tighten the sentence so the claim and reasoning are easier to follow."
	case analyzer.DomainProfessional:
		return "Tighten the sentence so the request, ownership, or next step is easier to follow."
	case analyzer.DomainMarketing:
		return "Tighten the sentence so the value and next action are easier to follow."
	case analyzer.DomainThoughtLeadership:
		return "Tighten the sentence so the central idea moves forward more clearly."
	default:
		return "Tighten the sentence so the movement is easier to follow."
	}
}

func maintenanceComment(domainName string) string {
	switch domainName {
	case analyzer.DomainTechnical:
		return "This line is part of the lighter maintenance pass. Keep the previously established clarity and step control visible here."
	case analyzer.DomainAcademic:
		return "This line is part of the lighter maintenance pass. Keep the previously established argument control visible here."
	case analyzer.DomainProfessional:
		return "This line is part of the lighter maintenance pass. Keep the previously established clarity and action control visible here."
	case analyzer.DomainMarketing:
		return "This line is part of the lighter maintenance pass. Keep the previously established message clarity visible here."
	case analyzer.DomainThoughtLeadership:
		return "This line is part of the lighter maintenance pass. Keep the previously established idea control visible here."
	default:
		return "This line is part of the lighter completed-skill maintenance pass. Keep the previously established control visible here."
	}
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
	case "story-causal-clarity", "claim-clarity", "objective-clarity", "sentence-clarity":
		return "clarity"
	case "story-scene-architecture", "paragraph-control", "structural-signposting", "narrative-sequencing":
		return "structure"
	case "tone-calibration", "authority-and-voice":
		return "tone"
	case "image-freshness", "descriptive-specificity", "word-choice":
		return "imagery"
	case "dialogue-under-strain", "dialogue-basics":
		return "dialogue"
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
