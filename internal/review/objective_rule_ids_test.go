package review

import "testing"

func TestObjectiveRuleIDsFor(t *testing.T) {
	ids := ObjectiveRuleIDsFor("story-scene-architecture", "scene architecture", "deterministic_skill_bridge")
	if len(ids) < 2 {
		t.Fatalf("expected rule ids, got %#v", ids)
	}
	hasObjectiveRule := false
	hasBasisRule := false
	for _, id := range ids {
		if id == "objective.story-scene-architecture.presence" {
			hasObjectiveRule = true
		}
		if id == "objective.story-scene-architecture.basis.deterministic_skill_bridge" {
			hasBasisRule = true
		}
	}
	if !hasObjectiveRule {
		t.Fatalf("missing objective presence rule id: %#v", ids)
	}
	if !hasBasisRule {
		t.Fatalf("missing basis rule id: %#v", ids)
	}
}

func TestObjectiveRuleIDsForRequiresCode(t *testing.T) {
	ids := ObjectiveRuleIDsFor("", "scene architecture", "status_fallback")
	if len(ids) != 0 {
		t.Fatalf("expected no ids for empty code, got %#v", ids)
	}
}
