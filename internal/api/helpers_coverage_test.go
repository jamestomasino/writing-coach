package api

import (
	"testing"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestNextSeedCodesSelectionOrder(t *testing.T) {
	tree := domain.TGOTreeDefinition{
		SeedCodes: []string{"seed-a", "seed-b", "seed-c"},
		TGOs: []domain.TGO{
			{Code: "seed-a"},
			{Code: "seed-b"},
			{Code: "seed-c"},
			{Code: "unlocked-1", Prerequisites: []string{"seed-a"}},
			{Code: "locked-1", Prerequisites: []string{"missing"}},
			{Code: "fallback-1"},
		},
	}

	selected := nextSeedCodes(tree, map[string]bool{
		"seed-a": true,
		"seed-b": true,
		"seed-c": true,
	})
	if len(selected) != 3 {
		t.Fatalf("expected 3 selected codes, got %#v", selected)
	}
	if selected[0] != "unlocked-1" || selected[1] != "fallback-1" || selected[2] != "locked-1" {
		t.Fatalf("unexpected selection order: %#v", selected)
	}
}

func TestToReviewJobResponse(t *testing.T) {
	now := time.Now().UTC().Add(-2 * time.Minute)
	resp := toReviewJobResponse(domain.ReviewJob{
		ID:           5,
		SubmissionID: 7,
		ReviewID:     9,
		Status:       "queued",
		AttemptCount: 1,
		MaxAttempts:  3,
		LastError:    "none",
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	if resp.ID != 5 || resp.SubmissionID != 7 || resp.ReviewID != 9 {
		t.Fatalf("unexpected id fields: %#v", resp)
	}
	if resp.Status != "queued" || resp.AttemptCount != 1 || resp.MaxAttempts != 3 || resp.LastError != "none" {
		t.Fatalf("unexpected queue fields: %#v", resp)
	}
	if resp.CreatedAt == "" || resp.UpdatedAt == "" {
		t.Fatalf("expected relative timestamps, got %#v", resp)
	}
}

func TestAIValidationLimitErrorMessage(t *testing.T) {
	var nilErr *aiValidationLimitError
	if nilErr.Error() != "" {
		t.Fatalf("expected nil error string to be empty, got %q", nilErr.Error())
	}

	err := &aiValidationLimitError{RetryAfter: 500 * time.Millisecond}
	if msg := err.Error(); msg == "" || msg == "provider validation is rate-limiting requests right now. Wait about 0 seconds and try again" {
		t.Fatalf("unexpected message for sub-second duration: %q", msg)
	}
}
