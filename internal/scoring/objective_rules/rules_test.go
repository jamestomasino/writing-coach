package objective_rules

import (
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestResolveAcademicRules(t *testing.T) {
	set, ok := Resolve("academic-thesis-clarity", analyzer.ContextOptions{
		TreeSlug:    "academic-essay-track",
		WritingType: "academic essay",
	})
	if !ok {
		t.Fatal("expected academic objective rule resolution")
	}
	if strings.TrimSpace(set.RuleID) == "" {
		t.Fatal("expected non-empty rule id")
	}
	if len(set.MetricRules) == 0 && len(set.CategoryRules) == 0 {
		t.Fatal("expected at least one metric/category rule")
	}
}

func TestAllAcademicPublicObjectivesHaveManifestCoverage(t *testing.T) {
	tree, ok := domain.BuiltInTreeBySlug("academic-essay-track")
	if !ok {
		t.Fatal("missing academic track")
	}
	for _, tgo := range tree.TGOs {
		if !HasAnyForCodeDomain(tgo.Code, analyzer.DomainAcademic) {
			t.Fatalf("missing manifest coverage for %s", tgo.Code)
		}
	}
}

func TestResolveAllIncludesSpecificAndBaseRulesets(t *testing.T) {
	sets := ResolveAll("academic-active-voice", analyzer.ContextOptions{
		TreeSlug:    "academic-essay-track",
		WritingType: "academic essay",
	})
	if len(sets) < 2 {
		t.Fatalf("expected layered rulesets for active voice, got %#v", sets)
	}
	if !strings.Contains(sets[0].RuleID, "active-voice-specific") {
		t.Fatalf("expected specific ruleset first, got %q", sets[0].RuleID)
	}
}

func TestAllTechnicalPublicObjectivesHaveManifestCoverage(t *testing.T) {
	tree, ok := domain.BuiltInTreeBySlug("technical-writing-track")
	if !ok {
		t.Fatal("missing technical track")
	}
	for _, tgo := range tree.TGOs {
		if !HasAnyForCodeDomain(tgo.Code, analyzer.DomainTechnical) {
			t.Fatalf("missing manifest coverage for %s", tgo.Code)
		}
	}
}

func TestAllProfessionalPublicObjectivesHaveManifestCoverage(t *testing.T) {
	tree, ok := domain.BuiltInTreeBySlug("professional-writing-track")
	if !ok {
		t.Fatal("missing professional track")
	}
	for _, tgo := range tree.TGOs {
		if !HasAnyForCodeDomain(tgo.Code, analyzer.DomainProfessional) {
			t.Fatalf("missing manifest coverage for %s", tgo.Code)
		}
	}
}

func TestAllThoughtLeadershipPublicObjectivesHaveManifestCoverage(t *testing.T) {
	tree, ok := domain.BuiltInTreeBySlug("thought-leadership-track")
	if !ok {
		t.Fatal("missing thought leadership track")
	}
	for _, tgo := range tree.TGOs {
		if !HasAnyForCodeDomain(tgo.Code, analyzer.DomainThoughtLeadership) {
			t.Fatalf("missing manifest coverage for %s", tgo.Code)
		}
	}
}

func TestAllPersuasivePublicObjectivesHaveManifestCoverage(t *testing.T) {
	tree, ok := domain.BuiltInTreeBySlug("persuasive-writing-track")
	if !ok {
		t.Fatal("missing persuasive track")
	}
	for _, tgo := range tree.TGOs {
		if !HasAnyForCodeDomain(tgo.Code, analyzer.DomainThoughtLeadership) {
			t.Fatalf("missing manifest coverage for %s", tgo.Code)
		}
	}
}

func TestAllMemoirPublicObjectivesHaveManifestCoverage(t *testing.T) {
	tree, ok := domain.BuiltInTreeBySlug("memoir-personal-narrative-track")
	if !ok {
		t.Fatal("missing memoir track")
	}
	for _, tgo := range tree.TGOs {
		if !HasAnyForCodeDomain(tgo.Code, analyzer.DomainFiction) {
			t.Fatalf("missing manifest coverage for %s", tgo.Code)
		}
	}
}
