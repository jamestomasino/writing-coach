package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/llm"
	"github.com/tomasino/writing-coach/internal/openai"
)

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
		return errors.New("gemini client disabled")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.withKey(c.baseURL+"/models"), nil)
	if err != nil {
		return err
	}
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
	payload, err := c.generateJSON(ctx, c.promptModel, openai.ExerciseSystemPrompt(), openai.ExerciseSchema(),
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
	payload, err := c.generateJSON(ctx, c.promptModel, openai.RevisionSystemPrompt(), openai.ExerciseSchema(),
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
	payload, err := c.generateJSON(ctx, c.reviewModel, openai.ReviewSystemPrompt(), openai.ReviewSchema(),
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

type generateContentRequest struct {
	SystemInstruction *content         `json:"systemInstruction,omitempty"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMIMEType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
}

type generateContentResponse struct {
	Candidates []candidate     `json:"candidates"`
	Error      *geminiAPIError `json:"error"`
}

type candidate struct {
	Content content `json:"content"`
}

type geminiAPIError struct {
	Message string `json:"message"`
}

func (c *Client) generateJSON(ctx context.Context, model, systemPrompt string, schema map[string]any, userInput string) ([]byte, error) {
	if !c.Enabled() {
		return nil, errors.New("gemini client disabled")
	}
	body := generateContentRequest{
		SystemInstruction: &content{Parts: []part{{Text: systemPrompt}}},
		Contents:          []content{{Parts: []part{{Text: userInput}}}},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
		},
	}
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.withKey(c.baseURL+"/models/"+model+":generateContent"), bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
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
		return nil, parseHTTPError(resp.StatusCode, responseBody)
	}
	var envelope generateContentResponse
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return nil, err
	}
	for _, candidate := range envelope.Candidates {
		for _, p := range candidate.Content.Parts {
			if strings.TrimSpace(p.Text) != "" {
				return []byte(p.Text), nil
			}
		}
	}
	return nil, errors.New("gemini api: no structured text returned")
}

func (c *Client) withKey(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	query := u.Query()
	query.Set("key", c.apiKey)
	u.RawQuery = query.Encode()
	return u.String()
}

func parseHTTPError(statusCode int, responseBody []byte) error {
	var failed struct {
		Error *geminiAPIError `json:"error"`
	}
	if json.Unmarshal(responseBody, &failed) == nil && failed.Error != nil && strings.TrimSpace(failed.Error.Message) != "" {
		return &llm.HTTPError{StatusCode: statusCode, Message: failed.Error.Message}
	}
	return &llm.HTTPError{StatusCode: statusCode}
}
