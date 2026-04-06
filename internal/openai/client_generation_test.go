package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestGenerateExerciseAndRevisionAndReview(t *testing.T) {
	exerciseJSON := `{"title":"Concrete Assignment","brief":"Write a focused paragraph.","constraints":["Use one scene."],"focus_skills":["narrative clarity"],"success_criteria":["Clear causal chain."]}`
	revisionJSON := `{"title":"Revision Pass","brief":"Rewrite for clarity.","constraints":["Keep same premise."],"focus_skills":["narrative clarity"],"success_criteria":["Cleaner structure."]}`
	reviewJSON := `{"summary":"Strong effort","strengths":["Clear opening"],"weaknesses":["Needs tighter ending"],"next_focus":"ending precision","skill_scores":[{"skill":"narrative clarity","score":3}],"tgo_assessments":[{"code":"story-causal-clarity","status":"developing","evidence":"middle section"},{"code":"story-scene-architecture","status":"secure","evidence":"scene arc"},{"code":"story-description-focus","status":"developing","evidence":"imagery drift"}],"completed_tgo_checks":[{"code":"story-causal-clarity","status":"holding","evidence":"still stable"}],"annotations":[{"quote":"line","tgo_code":"story-causal-clarity","category":"clarity","comment":"good","severity":"low"}]}`

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q", r.Method)
		}
		calls++
		var payload string
		switch calls {
		case 1:
			payload = exerciseJSON
		case 2:
			payload = revisionJSON
		default:
			payload = reviewJSON
		}
		body := map[string]any{
			"output": []map[string]any{
				{
					"type": "message",
					"content": []map[string]any{
						{"type": "output_text", "text": payload},
					},
				},
			},
		}
		encoded, _ := json.Marshal(body)
		_, _ = w.Write(encoded)
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{APIKey: "sk-test", BaseURL: server.URL, PromptModel: "gpt-5-mini", ReviewModel: "gpt-5-mini"})

	exercise, err := client.GenerateExercise(context.Background(), ExerciseRequest{WritingLanguage: "en", ActiveTGOs: []domain.TGO{{Code: "story-causal-clarity"}}, OnboardingProfile: &domain.OnboardingProfile{WritingType: "fiction", AssignmentFormat: "scene"}})
	if err != nil {
		t.Fatalf("generate exercise: %v", err)
	}
	if exercise.Title == "" || len(exercise.SuccessCriteria) == 0 {
		t.Fatalf("exercise = %#v", exercise)
	}

	revision, err := client.GenerateRevisionExercise(context.Background(), RevisionExerciseRequest{WritingLanguage: "en", SubmissionID: 10, SubmissionContent: "draft", ActiveTGOs: []domain.TGO{{Code: "story-causal-clarity"}}, Weaknesses: []string{"w"}, AnalyzerFindings: []string{"a"}})
	if err != nil {
		t.Fatalf("generate revision: %v", err)
	}
	if revision.Title == "" || len(revision.Constraints) == 0 {
		t.Fatalf("revision = %#v", revision)
	}

	review, scores, err := client.ReviewSubmission(context.Background(), ReviewRequest{WritingLanguage: "en", SubmissionID: 42, WordCount: 100, ActiveTGOs: []domain.TGO{{Code: "story-causal-clarity"}}, Content: "draft"})
	if err != nil {
		t.Fatalf("review submission: %v", err)
	}
	if review.Summary == "" || len(scores) != 1 {
		t.Fatalf("review=%#v scores=%#v", review, scores)
	}
}

func TestRunStructuredResponseHandlesNoOutputText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"output":[{"type":"message","content":[{"type":"other","text":"ignored"}]}]}`)
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{APIKey: "sk-test", BaseURL: server.URL, PromptModel: "gpt-5-mini", ReviewModel: "gpt-5-mini"})
	_, err := client.GenerateExercise(context.Background(), ExerciseRequest{WritingLanguage: "en", ActiveTGOs: []domain.TGO{{Code: "story-causal-clarity"}}, OnboardingProfile: &domain.OnboardingProfile{WritingType: "fiction", AssignmentFormat: "scene"}})
	if err == nil {
		t.Fatal("expected error when no structured text is returned")
	}
}
