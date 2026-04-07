package analyzer

import (
	"context"
	"strings"
	"testing"
)

func TestHeuristicAnalyzerFindsSignals(t *testing.T) {
	text := strings.Repeat("He really seemed to move as if the ruin itself suddenly spoke. ", 20)
	report, err := (Heuristic{}).Analyze(context.Background(), text)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if report.Metrics["word_count"] == 0 {
		t.Fatal("expected word count metric")
	}
	if len(report.Findings) == 0 {
		t.Fatal("expected findings")
	}
}

func TestHeuristicAnalyzerUsesContextualMessages(t *testing.T) {
	text := strings.Repeat("Install it carefully and quickly. ", 50)
	report, err := (Heuristic{}).AnalyzeWithContext(context.Background(), text, ContextOptions{
		WritingType:      "technical writing",
		AssignmentFormat: "how-to guide",
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	joined := make([]string, 0, len(report.Findings))
	for _, finding := range report.Findings {
		joined = append(joined, finding.Message)
	}
	if !strings.Contains(strings.Join(joined, " "), "instructions") {
		t.Fatalf("expected technical-writing message, got %#v", report.Findings)
	}
}

func TestHeuristicAnalyzerSkipsDialogueCheckForPoetrySpecialty(t *testing.T) {
	text := strings.Repeat("word ", 800)
	report, err := (Heuristic{}).AnalyzeWithContext(context.Background(), text, ContextOptions{
		WritingType:      "fiction poetry",
		AssignmentFormat: "poem",
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	for _, finding := range report.Findings {
		if finding.Category == "dialogue intelligence" {
			t.Fatalf("unexpected dialogue finding in poetry specialty: %#v", finding)
		}
	}
}
