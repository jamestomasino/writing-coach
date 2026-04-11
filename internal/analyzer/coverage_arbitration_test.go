package analyzer

import (
	"context"
	"testing"
)

func TestDeterministicCategoryOwnershipSpecs_Completeness(t *testing.T) {
	specs := DeterministicCategoryOwnershipSpecs()
	required := map[string]bool{
		"clarity":               false,
		"structure":             false,
		"readability":           false,
		"mechanics":             false,
		"style policy":          false,
		"narrative progression": false,
		"instructional completeness": false,
		"argument support":           false,
		"actionability":             false,
		"message hierarchy":         false,
		"memo execution":            false,
		"cta architecture":          false,
		"grant compliance framing":  false,
		"poetic craft proxies":      false,
	}
	for _, spec := range specs {
		key := normalizedValue(spec.Category)
		if _, ok := required[key]; ok {
			required[key] = true
		}
	}
	for category, present := range required {
		if !present {
			t.Fatalf("missing required ownership category %q", category)
		}
	}
}

func TestArbitrateDeterministicFindings_PrefersOwnerWhenAvailable(t *testing.T) {
	findings := []Finding{
		{Analyzer: "heuristic", Category: "clarity", Severity: "warning", Message: "heuristic clarity"},
		{Analyzer: "nlp", Category: "clarity", Severity: "warning", Message: "nlp clarity"},
	}
	filtered := arbitrateDeterministicFindings(findings, ContextOptions{WritingType: "technical writing"})
	if len(filtered) != 1 {
		t.Fatalf("filtered findings = %#v", filtered)
	}
	if filtered[0].Analyzer != "nlp" {
		t.Fatalf("expected nlp owner finding, got %#v", filtered[0])
	}
}

func TestArbitrateDeterministicFindings_AllowsHeuristicFallbackWhenOwnerUnavailable(t *testing.T) {
	findings := []Finding{
		{Analyzer: "heuristic", Category: "clarity", Severity: "warning", Message: "heuristic clarity"},
	}
	filtered := arbitrateDeterministicFindings(findings, ContextOptions{WritingType: "technical writing"})
	if len(filtered) != 1 || filtered[0].Analyzer != "heuristic" {
		t.Fatalf("expected heuristic fallback finding, got %#v", filtered)
	}
}

func TestArbitrateDeterministicFindings_SpecialtyPrecedenceOverridesGlobal(t *testing.T) {
	findings := []Finding{
		{Analyzer: "heuristic", Category: "structure", Severity: "warning", Message: "poetry structure"},
		{Analyzer: "nlp", Category: "structure", Severity: "warning", Message: "general structure"},
	}
	filtered := arbitrateDeterministicFindings(findings, ContextOptions{
		WritingType:      "poetry",
		AssignmentFormat: "poem",
	})
	if len(filtered) != 1 {
		t.Fatalf("filtered findings = %#v", filtered)
	}
	if filtered[0].Analyzer != "heuristic" {
		t.Fatalf("expected specialty heuristic finding, got %#v", filtered[0])
	}
}

func TestCanonicalCategoryForFinding_ThirdPartyOwnershipBuckets(t *testing.T) {
	if got := canonicalCategoryForFinding(Finding{Analyzer: "languagetool", Category: "Clarity"}); got != "mechanics" {
		t.Fatalf("expected mechanics for languagetool, got %q", got)
	}
	if got := canonicalCategoryForFinding(Finding{Analyzer: "vale", Category: "WritingCoach.Rule"}); got != "style policy" {
		t.Fatalf("expected style policy for vale, got %q", got)
	}
}

func TestServiceAnalyzeWithContext_AppliesCoverageArbitration(t *testing.T) {
	svc := NewService(
		stubAnalyzer{
			name: "heuristic",
			report: Report{
				Findings: []Finding{{Analyzer: "heuristic", Category: "clarity", Message: "heuristic"}},
			},
		},
		stubAnalyzer{
			name: "nlp",
			report: Report{
				Findings: []Finding{{Analyzer: "nlp", Category: "clarity", Message: "nlp"}},
			},
		},
	)

	report := svc.AnalyzeWithContext(context.Background(), "text", ContextOptions{WritingType: "technical writing"})
	if len(report.Findings) != 1 {
		t.Fatalf("expected one arbitrated finding, got %#v", report.Findings)
	}
	if report.Findings[0].Analyzer != "nlp" {
		t.Fatalf("expected nlp finding to survive arbitration, got %#v", report.Findings[0])
	}
}
