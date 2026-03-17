package openai

import (
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestNormalizeReview(t *testing.T) {
	value := reviewResponse{
		Summary:    "  a  summary \n",
		Strengths:  []string{" one  strength ", ""},
		Weaknesses: []string{" weak \t point "},
		NextFocus:  " prose precision ",
		SkillScores: []skillScore{
			{Skill: " prose precision ", Score: 3},
		},
		TGOAssessments: []tgoAssessment{
			{Code: " prose-precision ", Status: " secure ", Evidence: " line control "},
		},
		CompletedTGOChecks: []tgoAssessment{
			{Code: " sentence-clarity ", Status: " holding ", Evidence: " still stable "},
		},
	}

	got := normalizeReview(value)
	if got.Summary != "a summary" {
		t.Fatalf("summary = %q", got.Summary)
	}
	if len(got.Strengths) != 1 || got.Strengths[0] != "one strength" {
		t.Fatalf("strengths = %#v", got.Strengths)
	}
	if got.NextFocus != "prose precision" {
		t.Fatalf("next focus = %q", got.NextFocus)
	}
	if got.SkillScores[0].Skill != "prose precision" {
		t.Fatalf("skill = %q", got.SkillScores[0].Skill)
	}
	if got.CompletedTGOChecks[0].Status != "holding" {
		t.Fatalf("completed status = %q", got.CompletedTGOChecks[0].Status)
	}
}

func TestMeasurabilityGuidanceUsesHintsNotRawTGOs(t *testing.T) {
	got := measurabilityGuidance([]domain.TGO{
		{Code: "dialogue-intelligence", Description: "Make speech reveal rank, motive, and fracture under restraint."},
		{Code: "scene-architecture", Description: "Stage turns clearly so ritual and conflict remain legible."},
		{Code: "symbolic-control", Description: "Let objects carry fate without explaining their meaning."},
	})

	lower := strings.ToLower(got)
	if strings.Contains(lower, "dialogue intelligence") || strings.Contains(lower, "stage turns clearly") {
		t.Fatalf("expected abstract measurability guidance, got %q", got)
	}
	if !strings.Contains(lower, "dialogue") {
		t.Fatalf("expected dialogue measurability hint, got %q", got)
	}
}
