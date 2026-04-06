package analyzer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

type stubAnalyzer struct {
	name   string
	report Report
	err    error
}

func (s stubAnalyzer) Name() string {
	return s.name
}

func (s stubAnalyzer) Analyze(_ context.Context, _ string) (Report, error) {
	return s.report, s.err
}

type stubContextualAnalyzer struct {
	stubAnalyzer
	called bool
}

func (s *stubContextualAnalyzer) AnalyzeWithContext(_ context.Context, _ string, _ ContextOptions) (Report, error) {
	s.called = true
	return s.report, s.err
}

func TestContextFromProfileAndMerge(t *testing.T) {
	profile := domain.OnboardingProfile{
		WritingLanguage:  " english ",
		WritingType:      "technical writing",
		AssignmentFormat: "how-to guide",
		TemplateKey:      "technical-writing",
	}
	options := ContextFromProfile("tech-track", profile)
	if options.WritingLanguage != "en" {
		t.Fatalf("expected normalized language, got %q", options.WritingLanguage)
	}
	if options.TreeSlug != "tech-track" {
		t.Fatalf("expected tree slug, got %q", options.TreeSlug)
	}

	merged := Merge(
		Report{
			Findings: []Finding{{Message: "f1"}},
			Metrics:  map[string]int{"word_count": 10, "adverb_count": 3},
			Warnings: []string{"w1"},
		},
		Report{
			Findings: []Finding{{Message: "f2"}},
			Metrics:  map[string]int{"word_count": 20},
			Warnings: []string{"w2"},
		},
	)
	if len(merged.Findings) != 2 {
		t.Fatalf("expected merged findings, got %#v", merged.Findings)
	}
	if len(merged.Warnings) != 2 {
		t.Fatalf("expected merged warnings, got %#v", merged.Warnings)
	}
	if merged.Metrics["word_count"] != 20 {
		t.Fatalf("expected later metric value to win, got %d", merged.Metrics["word_count"])
	}
	if merged.Metrics["adverb_count"] != 3 {
		t.Fatalf("expected adverb_count metric preserved, got %d", merged.Metrics["adverb_count"])
	}
}

func TestServiceAnalyzeWithContext_UsesContextualAndCollectsWarnings(t *testing.T) {
	contextual := &stubContextualAnalyzer{
		stubAnalyzer: stubAnalyzer{
			name: "ctx",
			report: Report{
				Findings: []Finding{{Message: "contextual finding"}},
				Metrics:  map[string]int{"word_count": 11},
			},
		},
	}
	failing := stubAnalyzer{name: "broken", err: errors.New("boom")}

	svc := NewService(contextual, failing)
	report := svc.AnalyzeWithContext(context.Background(), "text", ContextOptions{WritingLanguage: "en"})
	if !contextual.called {
		t.Fatal("expected contextual analyzer path to be used")
	}
	if len(report.Findings) != 1 {
		t.Fatalf("expected one finding, got %#v", report.Findings)
	}
	if len(report.Warnings) != 1 || !strings.Contains(report.Warnings[0], "broken: boom") {
		t.Fatalf("expected warning from failing analyzer, got %#v", report.Warnings)
	}
}

func TestServiceAnalyzeAndHelpers(t *testing.T) {
	svc := NewService(stubAnalyzer{
		name: "simple",
		report: Report{
			Findings: []Finding{
				{Message: "first"},
				{Message: "second"},
			},
			Metrics: map[string]int{
				"word_count":           100,
				"avg_sentence_length":  20,
				"adverb_count":         2,
				"languagetool_matches": 1,
				"nlp_long_sentences":   3,
			},
		},
	})

	report := svc.Analyze(context.Background(), "text")
	top := TopFindings(report, 1)
	if len(top) != 1 || top[0] != "first" {
		t.Fatalf("unexpected top findings: %#v", top)
	}
	if TopFindings(report, 0) != nil {
		t.Fatal("expected nil top findings for non-positive limit")
	}

	summary := Summary(report)
	for _, expected := range []string{
		"words=100",
		"avg_sentence_length=20",
		"adverbs=2",
		"languagetool_matches=1",
		"nlp_long_sentences=3",
	} {
		if !strings.Contains(summary, expected) {
			t.Fatalf("summary missing %q: %q", expected, summary)
		}
	}
}

func TestLanguageHelpers(t *testing.T) {
	if writingLanguageLabel("en") != "English" {
		t.Fatalf("expected english label, got %q", writingLanguageLabel("en"))
	}
	if !deterministicLanguageSupported("en") {
		t.Fatal("expected english to be supported")
	}
	if languageToolCode("en") == "" {
		t.Fatal("expected language tool code for english")
	}
	warning := unsupportedLanguageWarning("heuristic", "es")
	if !strings.Contains(warning, "heuristic skipped") || !strings.Contains(warning, "es") {
		t.Fatalf("unexpected warning: %q", warning)
	}
}
