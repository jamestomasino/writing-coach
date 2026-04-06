package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValeAnalyzeWarnings(t *testing.T) {
	report, err := (Vale{}).Analyze(context.Background(), "text")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(report.Warnings) != 1 {
		t.Fatalf("expected not-configured warning, got %#v", report)
	}

	unsupported, err := (Vale{Binary: "/bin/true"}).AnalyzeWithContext(context.Background(), "text", ContextOptions{
		WritingLanguage: "es",
	})
	if err != nil {
		t.Fatalf("analyze with context: %v", err)
	}
	if len(unsupported.Warnings) != 1 {
		t.Fatalf("expected unsupported-language warning, got %#v", unsupported)
	}
}

func TestValeAnalyzeParsesOutput(t *testing.T) {
	tempDir := t.TempDir()
	fakeVale := filepath.Join(tempDir, "fake-vale.sh")
	script := "#!/bin/sh\ncat <<'JSON'\n{\"doc.md\":[{\"Check\":\"WritingCoachTest.Rule\",\"Message\":\"Tighten this sentence.\",\"Severity\":\"warning\"}]}\nJSON\n"
	if err := os.WriteFile(fakeVale, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake vale script: %v", err)
	}

	report, err := (Vale{Binary: fakeVale}).Analyze(context.Background(), "draft text")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("expected one finding, got %#v", report.Findings)
	}
	finding := report.Findings[0]
	if finding.Analyzer != "vale" || finding.Category != "WritingCoachTest.Rule" || finding.Severity != "warning" {
		t.Fatalf("unexpected parsed finding: %#v", finding)
	}
}
