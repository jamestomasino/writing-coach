package domain

import "testing"

func TestNextUnlockedTGOsRespectsPrerequisites(t *testing.T) {
	completed := map[string]bool{
		"causal-clarity":     true,
		"scene-architecture": true,
		"prose-precision":    true,
	}
	active := map[string]bool{}

	got := NextUnlockedTGOs("mythic-tragedy-apprenticeship", completed, active, 3)
	if len(got) == 0 {
		t.Fatal("expected unlocked TGOs")
	}
	if got[0].Code != "emotional-compression" {
		t.Fatalf("first unlocked = %q", got[0].Code)
	}
}

func TestSeedTGOsAreTreeSpecific(t *testing.T) {
	got := SeedTGOs("youth-writing-foundations")
	if len(got) != 3 {
		t.Fatalf("seed len = %d", len(got))
	}
	if got[0] != "word-choice" {
		t.Fatalf("first seed = %q", got[0])
	}
}
