package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestNormalizeReview(t *testing.T) {
	value := ReviewResponse{
		Summary:    "  a  summary \n",
		Strengths:  []string{" one  strength ", ""},
		Weaknesses: []string{" weak \t point "},
		NextFocus:  " prose precision ",
		SkillScores: []SkillScore{
			{Skill: " prose precision ", Score: 3},
		},
		TGOAssessments: []TGOAssessment{
			{Code: " prose-precision ", Status: " secure ", Evidence: " line control "},
		},
		CompletedTGOChecks: []TGOAssessment{
			{Code: " sentence-clarity ", Status: " holding ", Evidence: " still stable "},
		},
	}

	got := NormalizeReview(value)
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
	got := MeasurabilityGuidance([]domain.TGO{
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

func TestExerciseSystemPromptRequiresConcretePremise(t *testing.T) {
	got := strings.ToLower(ExerciseSystemPrompt())
	if !strings.Contains(got, "concrete premise") {
		t.Fatalf("expected concrete-premise instruction, got %q", got)
	}
	if !strings.Contains(got, "real starting point") {
		t.Fatalf("expected real-starting-point instruction, got %q", got)
	}
	if !strings.Contains(got, "brief itself must contain the core situation") {
		t.Fatalf("expected brief/core situation instruction, got %q", got)
	}
}

func TestValidateCredentialsUsesModelsEndpoint(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/models" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q", r.Method)
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		APIKey:      "sk-test-1234",
		BaseURL:     server.URL,
		PromptModel: "gpt-5-mini",
		ReviewModel: "gpt-5-mini",
	})
	if err := client.ValidateCredentials(context.Background()); err != nil {
		t.Fatalf("validate credentials: %v", err)
	}
	if authHeader != "Bearer sk-test-1234" {
		t.Fatalf("authorization = %q", authHeader)
	}
}

func TestValidateCredentialsReturnsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"invalid api key"}}`)
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		APIKey:      "sk-bad",
		BaseURL:     server.URL,
		PromptModel: "gpt-5-mini",
		ReviewModel: "gpt-5-mini",
	})
	err := client.ValidateCredentials(context.Background())
	if err == nil {
		t.Fatal("expected validation error")
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", httpErr.StatusCode)
	}
	if httpErr.Message != "invalid api key" {
		t.Fatalf("message = %q", httpErr.Message)
	}
}
