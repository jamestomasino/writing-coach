package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/llm"
	"github.com/tomasino/writing-coach/internal/openai"
)

const apiVersion = "2023-06-01"

type Client struct {
	apiKey      string
	baseURL     string
	promptModel string
	reviewModel string
	httpClient  *http.Client
}

type ClientOptions struct {
	APIKey      string
	BaseURL     string
	PromptModel string
	ReviewModel string
}

func NewClientWithOptions(opts ClientOptions) *Client {
	return &Client{
		apiKey:      opts.APIKey,
		baseURL:     strings.TrimRight(opts.BaseURL, "/"),
		promptModel: opts.PromptModel,
		reviewModel: opts.ReviewModel,
		httpClient:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && strings.TrimSpace(c.apiKey) != ""
}

func (c *Client) ValidateCredentials(ctx context.Context) error {
	if !c.Enabled() {
		return errors.New("anthropic client disabled")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return parseHTTPError(resp.StatusCode, body)
	}
	return nil
}

func (c *Client) GenerateExercise(ctx context.Context, input llm.ExerciseRequest) (domain.Exercise, error) {
	payload, err := c.runToolCall(ctx, c.promptModel, "exercise_prompt", openai.ExerciseSystemPrompt(), openai.ExerciseSchema(),
		openai.ExerciseUserInput(input),
	)
	if err != nil {
		return domain.Exercise{}, err
	}
	var parsed openai.ExerciseResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Exercise{}, err
	}
	parsed = openai.NormalizeExercise(parsed)
	if err := openai.ValidateExercise(parsed); err != nil {
		return domain.Exercise{}, err
	}
	return domain.Exercise{
		Title:           parsed.Title,
		Brief:           parsed.Brief,
		Constraints:     parsed.Constraints,
		FocusSkills:     openai.CoalesceFocusSkills(input.ActiveTGOs, parsed.FocusSkills),
		TGOCodes:        openai.ActiveTGOCodes(input.ActiveTGOs),
		SuccessCriteria: parsed.SuccessCriteria,
	}, nil
}

func (c *Client) GenerateRevisionExercise(ctx context.Context, input llm.RevisionExerciseRequest) (domain.Exercise, error) {
	payload, err := c.runToolCall(ctx, c.promptModel, "revision_prompt", openai.RevisionSystemPrompt(), openai.ExerciseSchema(),
		openai.RevisionUserInput(input),
	)
	if err != nil {
		return domain.Exercise{}, err
	}
	var parsed openai.ExerciseResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Exercise{}, err
	}
	parsed = openai.NormalizeExercise(parsed)
	if err := openai.ValidateExercise(parsed); err != nil {
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

func (c *Client) ReviewSubmission(ctx context.Context, input llm.ReviewRequest) (domain.Review, []domain.SkillScore, error) {
	payload, err := c.runToolCall(ctx, c.reviewModel, "submission_review", openai.ReviewSystemPrompt(), openai.ReviewSchema(),
		openai.ReviewUserInput(input),
	)
	if err != nil {
		return domain.Review{}, nil, err
	}
	var parsed openai.ReviewResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Review{}, nil, err
	}
	parsed = openai.NormalizeReview(parsed)
	if err := openai.ValidateReview(parsed); err != nil {
		return domain.Review{}, nil, err
	}
	scores := make([]domain.SkillScore, 0, len(parsed.SkillScores))
	for _, score := range parsed.SkillScores {
		scores = append(scores, domain.SkillScore{SubmissionID: input.SubmissionID, Skill: score.Skill, Score: score.Score})
	}
	return domain.Review{
		SubmissionID:       input.SubmissionID,
		Summary:            parsed.Summary,
		Strengths:          parsed.Strengths,
		Weaknesses:         parsed.Weaknesses,
		TGOAssessments:     openai.ToDomainAssessments(parsed.TGOAssessments),
		CompletedTGOChecks: openai.ToDomainAssessments(parsed.CompletedTGOChecks),
		Annotations:        openai.ToDomainAnnotations(parsed.Annotations),
		NextFocus:          parsed.NextFocus,
		MetricWordCount:    input.WordCount,
	}, scores, nil
}

type messagesRequest struct {
	Model      string        `json:"model"`
	MaxTokens  int           `json:"max_tokens"`
	System     string        `json:"system,omitempty"`
	Messages   []message     `json:"messages"`
	Tools      []tool        `json:"tools"`
	ToolChoice messageChoice `json:"tool_choice"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema"`
}

type messageChoice struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type messagesResponse struct {
	Type    string            `json:"type"`
	Content []responseContent `json:"content"`
	Error   *apiError         `json:"error"`
}

type responseContent struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *Client) runToolCall(ctx context.Context, model, toolName, systemPrompt string, schema map[string]any, userInput string) ([]byte, error) {
	if !c.Enabled() {
		return nil, errors.New("anthropic client disabled")
	}
	body := messagesRequest{
		Model:      model,
		MaxTokens:  4096,
		System:     systemPrompt,
		Messages:   []message{{Role: "user", Content: userInput}},
		Tools:      []tool{{Name: toolName, InputSchema: schema}},
		ToolChoice: messageChoice{Type: "tool", Name: toolName},
	}
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/messages", bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("content-type", "application/json")
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
		return nil, parseHTTPError(resp.StatusCode, responseBody)
	}
	var envelope messagesResponse
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return nil, err
	}
	for _, item := range envelope.Content {
		if item.Type == "tool_use" && len(item.Input) > 0 {
			return item.Input, nil
		}
	}
	return nil, errors.New("anthropic api: no tool response returned")
}

func parseHTTPError(statusCode int, responseBody []byte) error {
	var failed struct {
		Error *apiError `json:"error"`
	}
	if json.Unmarshal(responseBody, &failed) == nil && failed.Error != nil && strings.TrimSpace(failed.Error.Message) != "" {
		return &llm.HTTPError{StatusCode: statusCode, Message: failed.Error.Message}
	}
	return &llm.HTTPError{StatusCode: statusCode}
}
