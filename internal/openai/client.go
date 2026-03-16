package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/domain"
)

type Client struct {
	apiKey      string
	baseURL     string
	promptModel string
	reviewModel string
	httpClient  *http.Client
}

type ExerciseRequest struct {
	CurrentFocus    string
	DifficultyLevel int
	RecentTitles    []string
}

type ReviewRequest struct {
	SubmissionID     int64
	Content          string
	WordCount        int
	AnalysisSummary  string
	AnalyzerFindings []string
}

type exerciseResponse struct {
	Title           string   `json:"title"`
	Brief           string   `json:"brief"`
	Constraints     []string `json:"constraints"`
	FocusSkills     []string `json:"focus_skills"`
	SuccessCriteria []string `json:"success_criteria"`
}

type reviewResponse struct {
	Summary     string       `json:"summary"`
	Strengths   []string     `json:"strengths"`
	Weaknesses  []string     `json:"weaknesses"`
	NextFocus   string       `json:"next_focus"`
	SkillScores []skillScore `json:"skill_scores"`
}

type skillScore struct {
	Skill string `json:"skill"`
	Score int    `json:"score"`
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		apiKey:      cfg.OpenAIAPIKey,
		baseURL:     strings.TrimRight(cfg.OpenAIBaseURL, "/"),
		promptModel: cfg.PromptModel,
		reviewModel: cfg.ReviewModel,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != ""
}

func (c *Client) GenerateExercise(ctx context.Context, input ExerciseRequest) (domain.Exercise, error) {
	payload, err := c.runStructuredResponse(ctx, requestSpec{
		Model:       c.promptModel,
		SchemaName:  "exercise_prompt",
		Schema:      exerciseSchema(),
		SystemInput: exerciseSystemPrompt(),
		UserInput: fmt.Sprintf(
			"Current focus: %s\nDifficulty level: %d\nRecent exercise titles: %s",
			emptyDefault(input.CurrentFocus, "scene architecture"),
			input.DifficultyLevel,
			joinOrDefault(input.RecentTitles, "none"),
		),
	})
	if err != nil {
		return domain.Exercise{}, err
	}

	var parsed exerciseResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Exercise{}, err
	}
	parsed = normalizeExercise(parsed)
	if err := validateExercise(parsed); err != nil {
		return domain.Exercise{}, err
	}

	return domain.Exercise{
		Title:           parsed.Title,
		Brief:           parsed.Brief,
		Constraints:     parsed.Constraints,
		FocusSkills:     parsed.FocusSkills,
		SuccessCriteria: parsed.SuccessCriteria,
	}, nil
}

func (c *Client) ReviewSubmission(ctx context.Context, input ReviewRequest) (domain.Review, []domain.SkillScore, error) {
	payload, err := c.runStructuredResponse(ctx, requestSpec{
		Model:       c.reviewModel,
		SchemaName:  "submission_review",
		Schema:      reviewSchema(),
		SystemInput: reviewSystemPrompt(),
		UserInput: fmt.Sprintf(
			"Submission ID: %d\nWord count: %d\nDeterministic analysis summary: %s\nDeterministic findings: %s\nSubmission:\n%s",
			input.SubmissionID,
			input.WordCount,
			emptyDefault(input.AnalysisSummary, "none"),
			joinOrDefault(input.AnalyzerFindings, "none"),
			input.Content,
		),
	})
	if err != nil {
		return domain.Review{}, nil, err
	}

	var parsed reviewResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Review{}, nil, err
	}
	parsed = normalizeReview(parsed)
	if err := validateReview(parsed); err != nil {
		return domain.Review{}, nil, err
	}

	scores := make([]domain.SkillScore, 0, len(parsed.SkillScores))
	for _, score := range parsed.SkillScores {
		scores = append(scores, domain.SkillScore{
			SubmissionID: input.SubmissionID,
			Skill:        score.Skill,
			Score:        score.Score,
		})
	}

	return domain.Review{
		SubmissionID:    input.SubmissionID,
		Summary:         parsed.Summary,
		Strengths:       parsed.Strengths,
		Weaknesses:      parsed.Weaknesses,
		NextFocus:       parsed.NextFocus,
		MetricWordCount: input.WordCount,
	}, scores, nil
}

type requestSpec struct {
	Model       string
	SchemaName  string
	Schema      map[string]any
	SystemInput string
	UserInput   string
}

type responsesRequest struct {
	Model string      `json:"model"`
	Input []inputItem `json:"input"`
	Text  textConfig  `json:"text"`
}

type inputItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type textConfig struct {
	Format responseFormat `json:"format"`
}

type responseFormat struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
	Strict bool           `json:"strict"`
}

type responsesEnvelope struct {
	Output []outputItem `json:"output"`
	Error  *apiError    `json:"error"`
}

type outputItem struct {
	Type    string          `json:"type"`
	Content []outputContent `json:"content"`
}

type outputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type apiError struct {
	Message string `json:"message"`
}

func (c *Client) runStructuredResponse(ctx context.Context, spec requestSpec) ([]byte, error) {
	if !c.Enabled() {
		return nil, errors.New("openai client disabled")
	}

	body := responsesRequest{
		Model: spec.Model,
		Input: []inputItem{
			{Role: "system", Content: spec.SystemInput},
			{Role: "user", Content: spec.UserInput},
		},
		Text: textConfig{
			Format: responseFormat{
				Type:   "json_schema",
				Name:   spec.SchemaName,
				Schema: spec.Schema,
				Strict: true,
			},
		},
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		var failed responsesEnvelope
		if json.Unmarshal(responseBody, &failed) == nil && failed.Error != nil && failed.Error.Message != "" {
			return nil, fmt.Errorf("openai api: %s", failed.Error.Message)
		}
		return nil, fmt.Errorf("openai api: status %d", resp.StatusCode)
	}

	var envelope responsesEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return nil, err
	}

	var text strings.Builder
	for _, item := range envelope.Output {
		for _, content := range item.Content {
			if content.Type == "output_text" || content.Type == "text" {
				text.WriteString(content.Text)
			}
		}
	}
	if text.Len() == 0 {
		return nil, errors.New("openai api: no structured text returned")
	}

	return []byte(text.String()), nil
}

func exerciseSystemPrompt() string {
	return strings.TrimSpace(`
You are a professional fiction coach generating one brief exercise.
Return only schema-compliant JSON.
Target mode: mythopoeic epic tragedy with fantasy influences.
Favor discipline over ornament.
Avoid derivative references to named authors.
The exercise should train one main weakness and one supporting skill.
Choose focus skills only from the supplied taxonomy.
`)
}

func reviewSystemPrompt() string {
	return strings.TrimSpace(`
You are a professional fiction coach reviewing a short fiction exercise.
Return only schema-compliant JSON.
Evaluate for narrative clarity, tragic pressure, symbolic control, tonal discipline, and scene construction.
Do not flatter. Be concrete and developmental.
Choose the next focus that would most improve the following exercise.
Choose next_focus only from the supplied taxonomy.
`)
}

func exerciseSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "brief", "constraints", "focus_skills", "success_criteria"},
		"properties": map[string]any{
			"title":            map[string]any{"type": "string"},
			"brief":            map[string]any{"type": "string"},
			"constraints":      stringArraySchema(2, 5),
			"focus_skills":     enumStringArraySchema(1, 4, domain.SupportedSkills),
			"success_criteria": stringArraySchema(2, 5),
		},
	}
}

func reviewSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"summary", "strengths", "weaknesses", "next_focus", "skill_scores"},
		"properties": map[string]any{
			"summary":    map[string]any{"type": "string"},
			"strengths":  stringArraySchema(2, 4),
			"weaknesses": stringArraySchema(2, 4),
			"next_focus": map[string]any{"type": "string", "enum": domain.SupportedSkills},
			"skill_scores": map[string]any{
				"type":     "array",
				"minItems": 3,
				"maxItems": 6,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"skill", "score"},
					"properties": map[string]any{
						"skill": map[string]any{"type": "string", "enum": domain.SupportedSkills},
						"score": map[string]any{"type": "integer", "minimum": 1, "maximum": 5},
					},
				},
			},
		},
	}
}

func stringArraySchema(minItems, maxItems int) map[string]any {
	return map[string]any{
		"type":     "array",
		"minItems": minItems,
		"maxItems": maxItems,
		"items": map[string]any{
			"type": "string",
		},
	}
}

func enumStringArraySchema(minItems, maxItems int, values []string) map[string]any {
	return map[string]any{
		"type":     "array",
		"minItems": minItems,
		"maxItems": maxItems,
		"items": map[string]any{
			"type": "string",
			"enum": values,
		},
	}
}

func validateExercise(value exerciseResponse) error {
	if strings.TrimSpace(value.Title) == "" || strings.TrimSpace(value.Brief) == "" {
		return errors.New("openai exercise response missing title or brief")
	}
	if len(value.Constraints) == 0 || len(value.FocusSkills) == 0 || len(value.SuccessCriteria) == 0 {
		return errors.New("openai exercise response missing required arrays")
	}
	return nil
}

func validateReview(value reviewResponse) error {
	if strings.TrimSpace(value.Summary) == "" || strings.TrimSpace(value.NextFocus) == "" {
		return errors.New("openai review response missing summary or next focus")
	}
	if len(value.Strengths) == 0 || len(value.Weaknesses) == 0 || len(value.SkillScores) == 0 {
		return errors.New("openai review response missing required arrays")
	}
	return nil
}

func normalizeExercise(value exerciseResponse) exerciseResponse {
	value.Title = normalizeString(value.Title)
	value.Brief = normalizeString(value.Brief)
	value.Constraints = normalizeStrings(value.Constraints)
	value.FocusSkills = normalizeStrings(value.FocusSkills)
	value.SuccessCriteria = normalizeStrings(value.SuccessCriteria)
	return value
}

func normalizeReview(value reviewResponse) reviewResponse {
	value.Summary = normalizeString(value.Summary)
	value.Strengths = normalizeStrings(value.Strengths)
	value.Weaknesses = normalizeStrings(value.Weaknesses)
	value.NextFocus = normalizeString(value.NextFocus)
	for i := range value.SkillScores {
		value.SkillScores[i].Skill = normalizeString(value.SkillScores[i].Skill)
	}
	return value
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if cleaned := normalizeString(value); cleaned != "" {
			out = append(out, cleaned)
		}
	}
	return out
}

func normalizeString(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func emptyDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func joinOrDefault(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, ", ")
}
