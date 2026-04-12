package review

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
)

func TestApplyObjectiveOverlaySceneArchitectureRewardsAndPenalties(t *testing.T) {
	base := 3
	strongReport := analyzer.Report{
		Metrics: map[string]int{
			"nlp_topic_drift_score":     30,
			"nlp_coref_ambiguity_count": 0,
		},
	}
	strong, family, fired, reasons := applyObjectiveOverlay(
		"story-scene-architecture",
		base,
		strongReport,
		analyzer.ContextOptions{TreeSlug: "story-craft-track", WritingType: "fiction"},
	)
	if family != "scene-architecture" {
		t.Fatalf("family = %q", family)
	}
	if strong <= base {
		t.Fatalf("expected positive overlay adjustment, base=%d strong=%d", base, strong)
	}
	if len(fired) == 0 || len(reasons) == 0 {
		t.Fatalf("expected fired rules and reasons, fired=%v reasons=%v", fired, reasons)
	}

	weakReport := analyzer.Report{
		Metrics: map[string]int{
			"nlp_topic_drift_score":     80,
			"nlp_coref_ambiguity_count": 4,
		},
		Findings: []analyzer.Finding{
			{Category: "structure"},
			{Category: "clarity"},
			{Category: "clarity"},
		},
	}
	weak, _, weakFired, _ := applyObjectiveOverlay(
		"story-scene-architecture",
		base,
		weakReport,
		analyzer.ContextOptions{TreeSlug: "story-craft-track", WritingType: "fiction"},
	)
	if weak >= strong {
		t.Fatalf("expected weak overlay score below strong, weak=%d strong=%d", weak, strong)
	}
	if len(weakFired) == 0 {
		t.Fatalf("expected weak fired rules")
	}
}

func TestApplyManifestObjectiveOverlayAcademic(t *testing.T) {
	base := 3
	report := analyzer.Report{
		Metrics: map[string]int{
			"nlp_claim_evidence_coverage": 72,
			"nlp_claim_count":             4,
		},
	}
	adjusted, ruleID, fired, reasons, ok := applyManifestObjectiveOverlay(
		"academic-thesis-clarity",
		base,
		report,
		analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"},
	)
	if !ok {
		t.Fatal("expected academic manifest overlay")
	}
	if ruleID == "" {
		t.Fatal("expected non-empty manifest rule id")
	}
	if adjusted <= base {
		t.Fatalf("expected positive adjustment from manifest overlay, base=%d adjusted=%d", base, adjusted)
	}
	if len(fired) == 0 || len(reasons) == 0 {
		t.Fatalf("expected fired rules and reasons, fired=%v reasons=%v", fired, reasons)
	}
}
