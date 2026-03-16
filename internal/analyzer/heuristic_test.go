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
