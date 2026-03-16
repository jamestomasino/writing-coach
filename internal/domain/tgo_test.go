package domain

import "testing"

func TestNextUnlockedTGOsRespectsPrerequisites(t *testing.T) {
	completed := map[string]bool{
		"causal-clarity":     true,
		"scene-architecture": true,
		"prose-precision":    true,
	}
	active := map[string]bool{}

	got := NextUnlockedTGOs(completed, active, 3)
	if len(got) == 0 {
		t.Fatal("expected unlocked TGOs")
	}
	if got[0].Code != "emotional-compression" {
		t.Fatalf("first unlocked = %q", got[0].Code)
	}
}
