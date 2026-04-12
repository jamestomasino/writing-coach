package api

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/review"
)

func TestApplyObjectiveEvalCalibrationGatePassThroughWhenPolicyPasses(t *testing.T) {
	original := evaluateObjectiveCalibrationCorpus
	t.Cleanup(func() {
		evaluateObjectiveCalibrationCorpus = original
	})
	evaluateObjectiveCalibrationCorpus = func() (review.ObjectiveEvalResult, error) {
		return review.ObjectiveEvalResult{
			TotalChecks:              10,
			PassedChecks:             10,
			PassRate:                 1.0,
			RequiredMinPassRate:      1.0,
			PassedPolicyRequirements: true,
		}, nil
	}

	highlights := []string{"existing highlight"}
	recommendations := []string{"existing recommendation"}
	nextHighlights, nextRecommendations, nextAdequate := applyObjectiveEvalCalibrationGate(highlights, recommendations, true)

	if len(nextHighlights) != 1 || nextHighlights[0] != "existing highlight" {
		t.Fatalf("unexpected highlights: %#v", nextHighlights)
	}
	if len(nextRecommendations) != 1 || nextRecommendations[0] != "existing recommendation" {
		t.Fatalf("unexpected recommendations: %#v", nextRecommendations)
	}
	if !nextAdequate {
		t.Fatal("expected data to remain adequate")
	}
}

func TestApplyObjectiveEvalCalibrationGateFailsDataAdequateOnPolicyFailure(t *testing.T) {
	original := evaluateObjectiveCalibrationCorpus
	t.Cleanup(func() {
		evaluateObjectiveCalibrationCorpus = original
	})
	evaluateObjectiveCalibrationCorpus = func() (review.ObjectiveEvalResult, error) {
		maxTie := 0.0
		return review.ObjectiveEvalResult{
			TotalChecks:              12,
			PassedChecks:             10,
			PassRate:                 0.833,
			RequiredMinPassRate:      0.95,
			PairwiseTieRate:          0.125,
			MaxPairwiseTieRate:       &maxTie,
			PolicyFailures:           []string{"track academic-track pass rate 0.500 below required 1.000"},
			PassedPolicyRequirements: false,
		}, nil
	}

	nextHighlights, nextRecommendations, nextAdequate := applyObjectiveEvalCalibrationGate(nil, nil, true)
	if nextAdequate {
		t.Fatal("expected objective eval failure to force data_adequate=false")
	}
	if len(nextHighlights) == 0 {
		t.Fatal("expected objective eval failure highlights")
	}
	if !strings.Contains(nextHighlights[0], "Deterministic objective-score gate failed") {
		t.Fatalf("unexpected highlight: %q", nextHighlights[0])
	}
	foundPolicyFailure := false
	for _, item := range nextRecommendations {
		if strings.Contains(item, "Objective eval policy failure") {
			foundPolicyFailure = true
			break
		}
	}
	if !foundPolicyFailure {
		t.Fatalf("expected policy failure recommendation, got %#v", nextRecommendations)
	}
}

func TestApplyObjectiveEvalCalibrationGateFailsClosedOnEvalError(t *testing.T) {
	original := evaluateObjectiveCalibrationCorpus
	t.Cleanup(func() {
		evaluateObjectiveCalibrationCorpus = original
	})
	evaluateObjectiveCalibrationCorpus = func() (review.ObjectiveEvalResult, error) {
		return review.ObjectiveEvalResult{}, fmt.Errorf("corpus unavailable")
	}

	nextHighlights, nextRecommendations, nextAdequate := applyObjectiveEvalCalibrationGate(nil, nil, true)
	if nextAdequate {
		t.Fatal("expected eval error to force data_adequate=false")
	}
	if len(nextHighlights) == 0 || !strings.Contains(nextHighlights[0], "could not run") {
		t.Fatalf("unexpected highlights: %#v", nextHighlights)
	}
	if len(nextRecommendations) == 0 || !strings.Contains(nextRecommendations[0], "Fix objective-score evaluation corpus") {
		t.Fatalf("unexpected recommendations: %#v", nextRecommendations)
	}
}
