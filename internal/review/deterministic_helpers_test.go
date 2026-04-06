package review

import (
	"context"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestDeterministicScoringAndHelpers(t *testing.T) {
	if got := scoreFromWordCount(800); got != 4 {
		t.Fatalf("scoreFromWordCount = %d", got)
	}
	if got := scoreFromSentenceLength(15); got != 4 {
		t.Fatalf("scoreFromSentenceLength = %d", got)
	}
	if got := scoreFromFindingCount(4); got != 2 {
		t.Fatalf("scoreFromFindingCount = %d", got)
	}

	if got := clarityFocus(analyzer.DomainTechnical); got != "instruction clarity" {
		t.Fatalf("clarityFocus = %q", got)
	}
	if got := developmentFocus(analyzer.DomainAcademic); got != "claim development" {
		t.Fatalf("developmentFocus = %q", got)
	}
	if got := recommendedMinimumWords(analyzer.DomainMarketing); got != 160 {
		t.Fatalf("recommendedMinimumWords = %d", got)
	}
	if !strings.Contains(defaultAnnotationComment(analyzer.DomainTechnical), "instruction") {
		t.Fatalf("defaultAnnotationComment = %q", defaultAnnotationComment(analyzer.DomainTechnical))
	}
	if !strings.Contains(maintenanceComment(analyzer.DomainThoughtLeadership), "idea") {
		t.Fatalf("maintenanceComment = %q", maintenanceComment(analyzer.DomainThoughtLeadership))
	}

	sentences := splitSentences("One. Two! Three?")
	if len(sentences) != 3 {
		t.Fatalf("splitSentences = %#v", sentences)
	}
	if got := shortQuote("one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen"); len(strings.Fields(got)) != 14 {
		t.Fatalf("shortQuote = %q", got)
	}
	if got := annotationCategoryForTGO("story-scene-architecture"); got != "structure" {
		t.Fatalf("annotationCategoryForTGO = %q", got)
	}
	if got := annotationSeverity([]analyzer.Finding{{}, {}, {}, {}}); got != "medium" {
		t.Fatalf("annotationSeverity = %q", got)
	}
}

func TestDeterministicReviewLanguageProfiles(t *testing.T) {
	summary, strengths, weaknesses, next := deterministicReviewLanguage(analyzer.DomainProfessional)
	if summary == "" || len(strengths) == 0 || len(weaknesses) == 0 || next == "" {
		t.Fatalf("unexpected deterministicReviewLanguage response: %q %#v %#v %q", summary, strengths, weaknesses, next)
	}
}

func TestDeterministicAssessmentsAndAnnotations(t *testing.T) {
	active := []domain.TGO{{Code: "story-causal-clarity"}, {Code: "story-scene-architecture"}}
	completed := []domain.TGO{{Code: "story-prose-precision"}, {Code: "story-description-focus"}, {Code: "story-dialogue-intelligence"}}
	report := analyzer.Report{Findings: []analyzer.Finding{{Message: "issue a"}, {Message: "issue b"}, {Message: "issue c"}, {Message: "issue d"}}, Metrics: map[string]int{"sentence_count": 3, "avg_sentence_length": 20}}

	assessments := deterministicAssessments(active, report)
	if len(assessments) != 2 {
		t.Fatalf("deterministicAssessments = %#v", assessments)
	}
	checks := deterministicCompletedChecks(completed, report)
	if len(checks) != 2 {
		t.Fatalf("deterministicCompletedChecks = %#v", checks)
	}
	annotations := deterministicAnnotations("Sentence one. Sentence two. Sentence three.", active, completed, report, analyzer.DomainThoughtLeadership)
	if len(annotations) < 2 {
		t.Fatalf("deterministicAnnotations = %#v", annotations)
	}
}

func TestDeterministicReviewerReviewSubmissionLanguageFallbackAndDefaultScope(t *testing.T) {
	reviewer := deterministicReviewer{}
	sub := domain.Submission{ID: 10, Content: "One sentence. Two sentence.", WordCount: 120}
	report := analyzer.Report{Findings: []analyzer.Finding{{Message: "f1"}, {Message: "f2"}}, Metrics: map[string]int{"sentence_count": 2, "avg_sentence_length": 15}}

	reviewUnsupported, scoresUnsupported := reviewer.ReviewSubmission(context.Background(), sub, report, []domain.TGO{{Code: "story-causal-clarity"}}, nil, analyzer.ContextOptions{WritingLanguage: "es"})
	if !strings.Contains(strings.ToLower(reviewUnsupported.Summary), "deterministic coaching") || len(scoresUnsupported) == 0 {
		t.Fatalf("unsupported-language review=%#v scores=%#v", reviewUnsupported, scoresUnsupported)
	}

	reviewSupported, scoresSupported := reviewer.ReviewSubmission(context.Background(), sub, report, nil, nil, analyzer.ContextOptions{WritingLanguage: "en", WritingType: "technical writing"})
	if reviewSupported.NextFocus == "" || len(scoresSupported) == 0 {
		t.Fatalf("supported-language review=%#v scores=%#v", reviewSupported, scoresSupported)
	}

	scoped := reviewTGOs([]domain.TGO{{Code: "a"}}, false)
	if len(scoped) != 3 {
		t.Fatalf("reviewTGOs fallback = %#v", scoped)
	}
}
