package review

import (
	"reflect"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestPrioritizeInterventionsRanksPersistingWeaknessesFirst(t *testing.T) {
	current := domain.Review{
		NextFocus:        "narrative clarity",
		Weaknesses:       []string{"narrative clarity slips in the middle", "ending stakes stay abstract"},
		AnalyzerFindings: []string{"long sentence cluster in paragraph 2"},
	}
	comparison := &Comparison{
		PersistingWeaknesses: []string{"narrative clarity slips in the middle"},
	}

	got := PrioritizeInterventions(current, comparison)
	if len(got) == 0 {
		t.Fatal("expected interventions")
	}
	if got[0].Source != "comparison" {
		t.Fatalf("first source = %q", got[0].Source)
	}
	if got[0].Impact < got[1].Impact {
		t.Fatalf("expected descending impact order: %#v", got)
	}
	foundAlignment := false
	for _, code := range got[0].ReasonCodes {
		if code == "next_focus_alignment" {
			foundAlignment = true
			break
		}
	}
	if !foundAlignment {
		t.Fatalf("expected next focus alignment reason code: %#v", got[0])
	}
}

func TestPrioritizeInterventionsIsDeterministic(t *testing.T) {
	current := domain.Review{
		NextFocus: "prose precision",
		Weaknesses: []string{
			"weakened opening line control",
			"weak transitions between beats",
		},
		AnalyzerFindings: []string{
			"repeated filler phrase usage",
			"comma splice pattern detected",
		},
	}
	comparison := &Comparison{
		PersistingWeaknesses: []string{
			"weak transitions between beats",
		},
	}

	first := PrioritizeInterventions(current, comparison)
	second := PrioritizeInterventions(current, comparison)
	if len(first) != len(second) {
		t.Fatalf("intervention length mismatch: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if !reflect.DeepEqual(first[i], second[i]) {
			t.Fatalf("interventions differ at %d: %#v vs %#v", i, first[i], second[i])
		}
	}
	if len(first) > 5 {
		t.Fatalf("expected max 5 interventions, got %d", len(first))
	}
	for i, item := range first {
		if item.Rank != i+1 {
			t.Fatalf("rank mismatch at %d: %d", i, item.Rank)
		}
	}
}

func TestBuildInterventionOutcomesTracksLifecycle(t *testing.T) {
	current := []Intervention{
		{Rank: 1, Target: "narrative clarity slips in the middle", Source: "comparison", ReasonCodes: []string{"persisting_weakness"}},
		{Rank: 2, Target: "ending stakes stay abstract", Source: "review", ReasonCodes: []string{"current_weakness"}},
	}
	prior := []Intervention{
		{Rank: 1, Target: "narrative clarity slips in the middle", Source: "review", ReasonCodes: []string{"current_weakness"}},
		{Rank: 2, Target: "weak transitions between beats", Source: "review", ReasonCodes: []string{"current_weakness"}},
	}
	comparison := &Comparison{
		PersistingWeaknesses: []string{"narrative clarity slips in the middle"},
		AddressedWeaknesses:  []string{"weak transitions between beats"},
	}

	got := BuildInterventionOutcomes(current, prior, comparison, 77, 99)
	if len(got) != 3 {
		t.Fatalf("outcome count = %d", len(got))
	}
	if got[0].Status != "persisting" {
		t.Fatalf("first status = %q", got[0].Status)
	}
	if got[0].IntroducedReviewID != 77 {
		t.Fatalf("first introduced review id = %d", got[0].IntroducedReviewID)
	}
	if got[1].Status != "introduced" {
		t.Fatalf("second status = %q", got[1].Status)
	}
	if got[1].IntroducedReviewID != 99 {
		t.Fatalf("second introduced review id = %d", got[1].IntroducedReviewID)
	}
	if got[2].Status != "resolved" {
		t.Fatalf("third status = %q", got[2].Status)
	}
	if got[2].ResolvedReviewID != 99 {
		t.Fatalf("resolved review id = %d", got[2].ResolvedReviewID)
	}
}
