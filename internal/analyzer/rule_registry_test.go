package analyzer

import "testing"

func TestCurrentHeuristicRuleSpecsUniqueIDs(t *testing.T) {
	seen := make(map[string]struct{})
	for _, spec := range CurrentHeuristicRuleSpecs() {
		if spec.ID == "" {
			t.Fatal("expected non-empty rule id")
		}
		if _, ok := seen[spec.ID]; ok {
			t.Fatalf("duplicate rule id: %s", spec.ID)
		}
		seen[spec.ID] = struct{}{}
	}
}

func TestShouldEvaluateHeuristicRuleDialogueGuard(t *testing.T) {
	fictionScene := ContextOptions{
		WritingType:      "fiction",
		AssignmentFormat: "scene",
	}
	if !shouldEvaluateHeuristicRule("heuristic.dialogue_absence_long_scene", fictionScene, 800) {
		t.Fatal("expected dialogue rule to apply for fiction scene")
	}

	fictionPoetry := ContextOptions{
		WritingType:      "fiction poetry",
		AssignmentFormat: "poem",
	}
	if shouldEvaluateHeuristicRule("heuristic.dialogue_absence_long_scene", fictionPoetry, 800) {
		t.Fatal("expected dialogue rule to skip for poetry specialty")
	}
}

func TestShouldEvaluateHeuristicRuleUnknownDefaultsToTrue(t *testing.T) {
	if !shouldEvaluateHeuristicRule("heuristic.unknown", ContextOptions{}, 100) {
		t.Fatal("expected unknown rules to default true")
	}
}

func TestContextSpecialtiesDetection(t *testing.T) {
	options := ContextOptions{
		WritingType:      "fiction poetry memo",
		AssignmentFormat: "landing page",
	}
	specialties := contextSpecialties(options)
	if !containsString(specialties, "poetry") {
		t.Fatalf("expected poetry specialty, got %#v", specialties)
	}
	if !containsString(specialties, "memo") {
		t.Fatalf("expected memo specialty, got %#v", specialties)
	}
	if !containsString(specialties, "landing_page") {
		t.Fatalf("expected landing_page specialty, got %#v", specialties)
	}
}
