package analyzer

import (
	"context"
	"slices"
	"testing"
)

func TestDeterministicCategoryOwnershipSpecs_Completeness(t *testing.T) {
	specs := DeterministicCategoryOwnershipSpecs()
	required := map[string]bool{
		"clarity":                    false,
		"structure":                  false,
		"readability":                false,
		"mechanics":                  false,
		"style policy":               false,
		"narrative progression":      false,
		"instructional completeness": false,
		"argument support":           false,
		"actionability":              false,
		"message hierarchy":          false,
		"memo execution":             false,
		"cta architecture":           false,
		"grant compliance framing":   false,
		"poetic craft proxies":       false,
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

func TestArbitrateDeterministicFindings_DomainAndSpecialtyOwnershipMatrix(t *testing.T) {
	tests := []struct {
		name            string
		options         ContextOptions
		findings        []Finding
		wantAnalyzers   []string
		wantFindingSize int
	}{
		{
			name: "fiction narrative progression prefers nlp owner",
			options: ContextOptions{
				WritingType: "fiction",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "narrative progression", Message: "heuristic"},
				{Analyzer: "nlp", Category: "narrative progression", Message: "nlp"},
			},
			wantAnalyzers:   []string{"nlp"},
			wantFindingSize: 1,
		},
		{
			name: "technical instructional completeness prefers nlp owner",
			options: ContextOptions{
				WritingType: "technical writing",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "instructional completeness", Message: "heuristic"},
				{Analyzer: "nlp", Category: "instructional completeness", Message: "nlp"},
			},
			wantAnalyzers:   []string{"nlp"},
			wantFindingSize: 1,
		},
		{
			name: "academic argument support prefers nlp owner",
			options: ContextOptions{
				WritingType: "academic essay",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "argument support", Message: "heuristic"},
				{Analyzer: "nlp", Category: "argument support", Message: "nlp"},
			},
			wantAnalyzers:   []string{"nlp"},
			wantFindingSize: 1,
		},
		{
			name: "marketing message hierarchy prefers nlp owner",
			options: ContextOptions{
				WritingType: "marketing writing",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "message hierarchy", Message: "heuristic"},
				{Analyzer: "nlp", Category: "message hierarchy", Message: "nlp"},
			},
			wantAnalyzers:   []string{"nlp"},
			wantFindingSize: 1,
		},
		{
			name: "professional actionability keeps heuristic owner",
			options: ContextOptions{
				WritingType: "professional writing",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "actionability", Message: "heuristic"},
				{Analyzer: "nlp", Category: "actionability", Message: "nlp"},
			},
			wantAnalyzers:   []string{"heuristic"},
			wantFindingSize: 1,
		},
		{
			name: "general context keeps both findings when no domain ownership applies",
			options: ContextOptions{
				WritingType: "journal response",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "argument support", Message: "heuristic"},
				{Analyzer: "nlp", Category: "argument support", Message: "nlp"},
			},
			wantAnalyzers:   []string{"heuristic", "nlp"},
			wantFindingSize: 2,
		},
		{
			name: "memo specialty ownership keeps heuristic",
			options: ContextOptions{
				WritingType:      "professional writing",
				AssignmentFormat: "memo",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "memo execution", Message: "heuristic"},
				{Analyzer: "nlp", Category: "memo execution", Message: "nlp"},
			},
			wantAnalyzers:   []string{"heuristic"},
			wantFindingSize: 1,
		},
		{
			name: "landing page specialty ownership keeps heuristic",
			options: ContextOptions{
				WritingType:      "marketing writing",
				AssignmentFormat: "landing page",
			},
			findings: []Finding{
				{Analyzer: "heuristic", Category: "cta architecture", Message: "heuristic"},
				{Analyzer: "nlp", Category: "cta architecture", Message: "nlp"},
			},
			wantAnalyzers:   []string{"heuristic"},
			wantFindingSize: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := arbitrateDeterministicFindings(tc.findings, tc.options)
			if len(got) != tc.wantFindingSize {
				t.Fatalf("len(got) = %d, want %d, findings=%#v", len(got), tc.wantFindingSize, got)
			}
			for _, finding := range got {
				if !slices.Contains(tc.wantAnalyzers, finding.Analyzer) {
					t.Fatalf("unexpected analyzer %q in findings %#v", finding.Analyzer, got)
				}
			}
		})
	}
}
