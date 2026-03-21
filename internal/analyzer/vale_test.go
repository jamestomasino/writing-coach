package analyzer

import (
	"strings"
	"testing"
)

func TestValeStylesForContext_FantasyFiction(t *testing.T) {
	styles := valeStylesForContext(ContextOptions{
		TreeSlug:    "fantasy-fiction-track",
		TemplateKey: "fantasy-fiction",
	})
	if !containsStyle(styles, "WritingCoachFiction") {
		t.Fatalf("expected fiction pack, got %v", styles)
	}
	if !containsStyle(styles, "WritingCoachFantasy") {
		t.Fatalf("expected fantasy pack, got %v", styles)
	}
}

func TestValeStylesForContext_TechnicalWriting(t *testing.T) {
	styles := valeStylesForContext(ContextOptions{
		WritingType:      "technical writing",
		AssignmentFormat: "how-to guide",
		TemplateKey:      "technical-writing",
	})
	if !containsStyle(styles, "WritingCoachTechnical") {
		t.Fatalf("expected technical pack, got %v", styles)
	}
	if containsStyle(styles, "WritingCoachFiction") {
		t.Fatalf("did not expect fiction pack, got %v", styles)
	}
}

func TestValeStylesForContext_MarketingWriting(t *testing.T) {
	styles := valeStylesForContext(ContextOptions{
		WritingType:      "marketing writing",
		AssignmentFormat: "landing page",
		TemplateKey:      "marketing-writing",
	})
	if !containsStyle(styles, "WritingCoachMarketing") {
		t.Fatalf("expected marketing pack, got %v", styles)
	}
}

func TestValeDynamicConfigIncludesBaseAndSelectedStyles(t *testing.T) {
	vale := Vale{
		StylesPath:    "/tmp/styles",
		MinAlertLevel: "warning",
		BaseStyles:    []string{"WritingCoachCore"},
	}
	config, ok := vale.dynamicConfig(ContextOptions{
		TemplateKey: "academic-essay",
	})
	if !ok {
		t.Fatal("expected dynamic config")
	}
	if !strings.Contains(config, "StylesPath = /tmp/styles") {
		t.Fatalf("missing styles path in config: %q", config)
	}
	if !strings.Contains(config, "MinAlertLevel = warning") {
		t.Fatalf("missing min alert level in config: %q", config)
	}
	if !strings.Contains(config, "WritingCoachCore") || !strings.Contains(config, "WritingCoachAcademic") {
		t.Fatalf("missing selected styles in config: %q", config)
	}
}

func containsStyle(styles []string, target string) bool {
	for _, style := range styles {
		if style == target {
			return true
		}
	}
	return false
}
