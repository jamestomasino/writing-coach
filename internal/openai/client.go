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
	"github.com/tomasino/writing-coach/internal/llm"
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

type HTTPError = llm.HTTPError

type ExerciseRequest = llm.ExerciseRequest
type RevisionExerciseRequest = llm.RevisionExerciseRequest
type ReviewRequest = llm.ReviewRequest

type ExerciseResponse struct {
	Title           string   `json:"title"`
	Brief           string   `json:"brief"`
	Constraints     []string `json:"constraints"`
	FocusSkills     []string `json:"focus_skills"`
	SuccessCriteria []string `json:"success_criteria"`
}

type ReviewResponse struct {
	Summary            string          `json:"summary"`
	Strengths          []string        `json:"strengths"`
	Weaknesses         []string        `json:"weaknesses"`
	NextFocus          string          `json:"next_focus"`
	SkillScores        []SkillScore    `json:"skill_scores"`
	TGOAssessments     []TGOAssessment `json:"tgo_assessments"`
	CompletedTGOChecks []TGOAssessment `json:"completed_tgo_checks"`
	Annotations        []Annotation    `json:"annotations"`
}

type SkillScore struct {
	Skill string `json:"skill"`
	Score int    `json:"score"`
}

type TGOAssessment struct {
	Code     string `json:"code"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type Annotation struct {
	Quote    string `json:"quote"`
	TGOCode  string `json:"tgo_code"`
	Category string `json:"category"`
	Comment  string `json:"comment"`
	Severity string `json:"severity"`
}

func NewClient(cfg config.Config) *Client {
	return NewClientWithOptions(ClientOptions{
		APIKey:      cfg.OpenAIAPIKey,
		BaseURL:     cfg.OpenAIBaseURL,
		PromptModel: cfg.PromptModel,
		ReviewModel: cfg.ReviewModel,
	})
}

func NewClientWithOptions(opts ClientOptions) *Client {
	return &Client{
		apiKey:      opts.APIKey,
		baseURL:     strings.TrimRight(opts.BaseURL, "/"),
		promptModel: opts.PromptModel,
		reviewModel: opts.ReviewModel,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != ""
}

func (c *Client) ValidateCredentials(ctx context.Context) error {
	if !c.Enabled() {
		return errors.New("openai client disabled")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return parseHTTPError(resp.StatusCode, responseBody)
	}
	return nil
}

func (c *Client) GenerateExercise(ctx context.Context, input ExerciseRequest) (domain.Exercise, error) {
	payload, err := c.runStructuredResponse(ctx, requestSpec{
		Model:       c.promptModel,
		SchemaName:  "exercise_prompt",
		Schema:      ExerciseSchema(),
		SystemInput: ExerciseSystemPrompt(),
		UserInput: fmt.Sprintf(
			"Writing language: %s\nWriting track profile:\n%s\nHidden review guidance: %s\nUse this guidance only to make the finished draft reviewable. Do not name it, quote it, or turn it into visible checklist items in the assignment.\nCurrent coaching emphasis: %s\nDifficulty level: %d\nRecent exercise titles: %s\nRecent weaknesses: %s\nRecurring analyzer findings: %s\nCoaching context: %s",
			FormatWritingLanguage(input.WritingLanguage),
			FormatOnboardingProfile(input.OnboardingProfile),
			MeasurabilityGuidance(input.ActiveTGOs),
			EmptyDefault(input.CurrentFocus, "none"),
			input.DifficultyLevel,
			JoinOrDefault(input.RecentTitles, "none"),
			JoinOrDefault(input.RecentWeaknesses, "none"),
			JoinOrDefault(input.RecurringFindings, "none"),
			EmptyDefault(input.CoachingBrief, "none"),
		),
	})
	if err != nil {
		return domain.Exercise{}, err
	}

	var parsed ExerciseResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Exercise{}, err
	}
	parsed = NormalizeExercise(parsed)
	if err := ValidateExercise(parsed); err != nil {
		return domain.Exercise{}, err
	}

	return domain.Exercise{
		Title:           parsed.Title,
		Brief:           parsed.Brief,
		Constraints:     parsed.Constraints,
		FocusSkills:     CoalesceFocusSkills(input.ActiveTGOs, parsed.FocusSkills),
		TGOCodes:        ActiveTGOCodes(input.ActiveTGOs),
		SuccessCriteria: parsed.SuccessCriteria,
	}, nil
}

func (c *Client) GenerateRevisionExercise(ctx context.Context, input RevisionExerciseRequest) (domain.Exercise, error) {
	payload, err := c.runStructuredResponse(ctx, requestSpec{
		Model:       c.promptModel,
		SchemaName:  "revision_prompt",
		Schema:      ExerciseSchema(),
		SystemInput: RevisionSystemPrompt(),
		UserInput: fmt.Sprintf(
			"Writing language: %s\nCurrent focus: %s\nDifficulty level: %d\nActive TGOs: %s\nSubmission ID: %d\nSubmission:\n%s\nCurrent weaknesses: %s\nAnalyzer findings: %s\nComparison summary: %s\nRecent weaknesses: %s\nRecurring analyzer findings: %s\nCoaching context: %s",
			FormatWritingLanguage(input.WritingLanguage),
			EmptyDefault(input.CurrentFocus, "prose precision"),
			input.DifficultyLevel,
			JoinTGOs(input.ActiveTGOs),
			input.SubmissionID,
			input.SubmissionContent,
			JoinOrDefault(input.Weaknesses, "none"),
			JoinOrDefault(input.AnalyzerFindings, "none"),
			EmptyDefault(input.ComparisonSummary, "none"),
			JoinOrDefault(input.RecentWeaknesses, "none"),
			JoinOrDefault(input.RecurringFindings, "none"),
			EmptyDefault(input.CoachingBrief, "none"),
		),
	})
	if err != nil {
		return domain.Exercise{}, err
	}

	var parsed ExerciseResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Exercise{}, err
	}
	parsed = NormalizeExercise(parsed)
	if err := ValidateExercise(parsed); err != nil {
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
		Schema:      ReviewSchema(),
		SystemInput: ReviewSystemPrompt(),
		UserInput: fmt.Sprintf(
			"Writing language: %s\nSubmission ID: %d\nWord count: %d\nActive TGOs: %s\nCompleted TGOs to monitor for regression: %s\nDeterministic analysis summary: %s\nDeterministic findings: %s\nCoaching context: %s\nSubmission:\n%s",
			FormatWritingLanguage(input.WritingLanguage),
			input.SubmissionID,
			input.WordCount,
			JoinTGOs(input.ActiveTGOs),
			JoinTGOs(input.CompletedTGOs),
			EmptyDefault(input.AnalysisSummary, "none"),
			JoinOrDefault(input.AnalyzerFindings, "none"),
			EmptyDefault(input.CoachingBrief, "none"),
			input.Content,
		),
	})
	if err != nil {
		return domain.Review{}, nil, err
	}

	var parsed ReviewResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return domain.Review{}, nil, err
	}
	parsed = NormalizeReview(parsed)
	if err := ValidateReview(parsed); err != nil {
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
		SubmissionID:       input.SubmissionID,
		Summary:            parsed.Summary,
		Strengths:          parsed.Strengths,
		Weaknesses:         parsed.Weaknesses,
		TGOAssessments:     ToDomainAssessments(parsed.TGOAssessments),
		CompletedTGOChecks: ToDomainAssessments(parsed.CompletedTGOChecks),
		Annotations:        ToDomainAnnotations(parsed.Annotations),
		NextFocus:          parsed.NextFocus,
		MetricWordCount:    input.WordCount,
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
		return nil, parseHTTPError(resp.StatusCode, responseBody)
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

func parseHTTPError(statusCode int, responseBody []byte) error {
	var failed responsesEnvelope
	if json.Unmarshal(responseBody, &failed) == nil && failed.Error != nil && failed.Error.Message != "" {
		return &HTTPError{StatusCode: statusCode, Message: failed.Error.Message}
	}
	return &HTTPError{StatusCode: statusCode}
}

func ExerciseSystemPrompt() string {
	return strings.TrimSpace(`
You are a professional writing coach generating one brief exercise.
Return only schema-compliant JSON.
This is a new assignment, not a revision.
Ask for a fresh draft written from scratch unless the user input clearly says otherwise.
Write at about a 6th-grade reading level.
Use short, plain sentences.
Make the request easy to understand on the first read.
If a writing language is supplied, write the assignment in that language.
Build the assignment from the writing track profile first: writing domain, format, audience, subject matter, tone, and goals.
Use the track profile to create a concrete premise, situation, or communication problem, not a generic template.
The brief must give the writer a real starting point: a pressure point, choice, conflict, reader need, or message to handle.
If the assignment is fictional, ground it in a specific dramatic situation rather than only naming a setting or mood.
If the assignment is nonfiction, ground it in a specific reader need, decision, objection, or communication task.
Make the title specific enough to feel like a real assignment, not a category label.
The supplied review rubric skills are for later evaluation, not for choosing the assignment topic.
If a review skill needs a visible feature to be measurable, quietly make room for that feature in the assignment.
Example: if the review skill depends on dialogue quality, the assignment should create a natural reason for dialogue to appear.
Never copy the hidden review guidance into the visible brief, constraints, or success criteria.
Do not turn the assignment into a beat sheet, rubric, or checklist derived from the review criteria.
Do not let the constraints carry the whole assignment. The brief itself must contain the core situation.
Match the user's writing mode and tone only when the supplied coaching context supports it.
Favor discipline, clarity, and specificity over ornament.
Avoid derivative references to named authors or genres unless the coaching context clearly calls for them.
Do not introduce genre or tone framing words such as "mythic", "fantasy", "epic", or "tragic" unless they are clearly present in the supplied track profile or coaching context.
The exercise should give the writer a strong starting point inside the track, not a disguised TGO drill.
The brief should be 1-2 short sentences.
The constraints and success criteria should use simple, direct language.
Do not use words like "rewrite" or "revise" unless this is explicitly a revision task.
Choose focus skills only from the supplied taxonomy.
`)
}

func RevisionSystemPrompt() string {
	return strings.TrimSpace(`
You are a professional writing coach generating a rewrite brief for the author's next draft.
Return only schema-compliant JSON.
Do not generate a fresh unrelated exercise.
Write at about a 6th-grade reading level.
Use short, plain sentences.
Make each instruction easy to follow.
If a writing language is supplied, write the revision brief in that language.
Preserve the core scene, but focus the revision on the most important weaknesses.
Keep the brief aligned to the supplied coaching context without repeating every profile detail.
Do not introduce genre or tone framing words such as "mythic", "fantasy", "epic", or "tragic" unless they are clearly present in the supplied track profile or coaching context.
The brief should be 1-2 short sentences.
The constraints and success criteria should use simple, direct language.
Choose focus skills only from the supplied taxonomy.
`)
}

func ReviewSystemPrompt() string {
	return strings.TrimSpace(`
You are a professional writing coach reviewing a writing exercise.
Return only schema-compliant JSON.
Evaluate for narrative clarity, control, tonal discipline, and scene construction in the mode implied by the coaching context.
If a writing language is supplied, write the review in that language.
Keep quoted annotations exactly as they appear in the submission.
Do not flatter. Be concrete and developmental.
Choose the next focus that would most improve the following exercise.
Choose next_focus only from the supplied taxonomy.
Assess each active TGO with one of: developing, secure, mastered.
Optionally flag up to two completed TGOs as holding or slipping if the draft regresses on already-established skills.
Return up to six short annotations for the UI. Each annotation must cite a short exact quote from the submission, map to a TGO, classify the issue, and explain the coaching note.
`)
}

func ExerciseSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "brief", "constraints", "focus_skills", "success_criteria"},
		"properties": map[string]any{
			"title":            map[string]any{"type": "string"},
			"brief":            map[string]any{"type": "string"},
			"constraints":      StringArraySchema(2, 5),
			"focus_skills":     EnumStringArraySchema(1, 4, domain.SupportedSkills),
			"success_criteria": StringArraySchema(2, 5),
		},
	}
}

func ReviewSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"summary", "strengths", "weaknesses", "next_focus", "skill_scores", "tgo_assessments", "completed_tgo_checks", "annotations"},
		"properties": map[string]any{
			"summary":    map[string]any{"type": "string"},
			"strengths":  StringArraySchema(2, 4),
			"weaknesses": StringArraySchema(2, 4),
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
			"tgo_assessments": map[string]any{
				"type":     "array",
				"minItems": 3,
				"maxItems": 3,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"code", "status", "evidence"},
					"properties": map[string]any{
						"code":     map[string]any{"type": "string"},
						"status":   map[string]any{"type": "string", "enum": []string{"developing", "secure", "mastered"}},
						"evidence": map[string]any{"type": "string"},
					},
				},
			},
			"completed_tgo_checks": map[string]any{
				"type":     "array",
				"minItems": 0,
				"maxItems": 2,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"code", "status", "evidence"},
					"properties": map[string]any{
						"code":     map[string]any{"type": "string"},
						"status":   map[string]any{"type": "string", "enum": []string{"holding", "slipping"}},
						"evidence": map[string]any{"type": "string"},
					},
				},
			},
			"annotations": map[string]any{
				"type":     "array",
				"minItems": 0,
				"maxItems": 6,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"quote", "tgo_code", "category", "comment", "severity"},
					"properties": map[string]any{
						"quote":    map[string]any{"type": "string"},
						"tgo_code": map[string]any{"type": "string"},
						"category": map[string]any{"type": "string", "enum": []string{"clarity", "structure", "tone", "imagery", "dialogue", "symbolism", "grammar", "revision"}},
						"comment":  map[string]any{"type": "string"},
						"severity": map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
					},
				},
			},
		},
	}
}

func StringArraySchema(minItems, maxItems int) map[string]any {
	return map[string]any{
		"type":     "array",
		"minItems": minItems,
		"maxItems": maxItems,
		"items": map[string]any{
			"type": "string",
		},
	}
}

func EnumStringArraySchema(minItems, maxItems int, values []string) map[string]any {
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

func ValidateExercise(value ExerciseResponse) error {
	if strings.TrimSpace(value.Title) == "" || strings.TrimSpace(value.Brief) == "" {
		return errors.New("openai exercise response missing title or brief")
	}
	if len(value.Constraints) == 0 || len(value.FocusSkills) == 0 || len(value.SuccessCriteria) == 0 {
		return errors.New("openai exercise response missing required arrays")
	}
	return nil
}

func ValidateReview(value ReviewResponse) error {
	if strings.TrimSpace(value.Summary) == "" || strings.TrimSpace(value.NextFocus) == "" {
		return errors.New("openai review response missing summary or next focus")
	}
	if len(value.Strengths) == 0 || len(value.Weaknesses) == 0 || len(value.SkillScores) == 0 || len(value.TGOAssessments) != 3 {
		return errors.New("openai review response missing required arrays")
	}
	return nil
}

func NormalizeExercise(value ExerciseResponse) ExerciseResponse {
	value.Title = normalizeString(value.Title)
	value.Brief = normalizeString(value.Brief)
	value.Constraints = normalizeStrings(value.Constraints)
	value.FocusSkills = normalizeStrings(value.FocusSkills)
	value.SuccessCriteria = normalizeStrings(value.SuccessCriteria)
	return value
}

func FormatOnboardingProfile(profile *domain.OnboardingProfile) string {
	if profile == nil {
		return "none"
	}
	lines := domain.PromptProfileLines(*profile)
	if len(lines) == 0 {
		return "none"
	}
	return strings.Join(lines, "\n")
}

func FormatWritingLanguage(code string) string {
	normalized := domain.NormalizeWritingLanguage(code)
	if normalized == "" {
		normalized = domain.DefaultWritingLanguage
	}
	return fmt.Sprintf("%s (%s)", domain.WritingLanguageLabel(normalized), normalized)
}

func CoalesceFocusSkills(activeTGOs []domain.TGO, fallback []string) []string {
	if len(activeTGOs) == 0 {
		return fallback
	}
	seen := map[string]bool{}
	var out []string
	for _, tgo := range activeTGOs {
		skill := strings.TrimSpace(domain.TGOCodeToSkill[tgo.Code])
		if skill == "" || seen[skill] {
			continue
		}
		seen[skill] = true
		out = append(out, skill)
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func ActiveTGOCodes(activeTGOs []domain.TGO) []string {
	var out []string
	for _, tgo := range activeTGOs {
		code := strings.TrimSpace(tgo.Code)
		if code == "" {
			continue
		}
		out = append(out, code)
	}
	return out
}

func MeasurabilityGuidance(activeTGOs []domain.TGO) string {
	if len(activeTGOs) == 0 {
		return "Keep the assignment concrete enough that craft choices can be reviewed."
	}

	seen := map[string]bool{}
	var hints []string
	add := func(hint string) {
		if hint == "" || seen[hint] {
			return
		}
		seen[hint] = true
		hints = append(hints, hint)
	}

	for _, tgo := range activeTGOs {
		skill := strings.ToLower(strings.TrimSpace(domain.TGOCodeToSkill[tgo.Code]))
		text := strings.ToLower(strings.Join([]string{tgo.Code, tgo.Title, tgo.Description}, " "))
		switch {
		case skill == "dialogue intelligence" || strings.Contains(text, "dialogue"):
			add("Make room for at least a little dialogue if it fits the format.")
		case skill == "scene architecture" || skill == "narrative clarity" || skill == "clarity and coherence" || skill == "structural signposting" || strings.Contains(text, "scene") || strings.Contains(text, "causal") || strings.Contains(text, "sequence"):
			add("Give the piece a clear sequence so a reviewer can follow what happens and why.")
		case skill == "symbolic control" || skill == "image freshness" || skill == "descriptive precision" || strings.Contains(text, "symbol") || strings.Contains(text, "image") || strings.Contains(text, "object"):
			add("Include at least one concrete image, object, or detail that can carry meaning.")
		case skill == "evidence integration" || skill == "reasoning quality" || skill == "analysis depth" || strings.Contains(text, "evidence") || strings.Contains(text, "reason") || strings.Contains(text, "analysis"):
			add("Leave room for a clear example, support, or line of reasoning.")
		case skill == "audience alignment" || skill == "actionability" || skill == "scannability" || skill == "user goal alignment" || strings.Contains(text, "audience") || strings.Contains(text, "reader"):
			add("Make the purpose and audience legible in the piece.")
		case skill == "voice presence" || skill == "tone calibration" || skill == "authority and voice" || skill == "rhetorical force" || strings.Contains(text, "voice") || strings.Contains(text, "tone"):
			add("Give the writer room to show voice and tone through the draft itself.")
		}
	}

	if len(hints) == 0 {
		return "Keep the assignment concrete enough that the selected review skills can be judged from the draft."
	}
	if len(hints) > 2 {
		hints = hints[:2]
	}
	return strings.Join(hints, " ")
}

func NormalizeReview(value ReviewResponse) ReviewResponse {
	value.Summary = normalizeString(value.Summary)
	value.Strengths = normalizeStrings(value.Strengths)
	value.Weaknesses = normalizeStrings(value.Weaknesses)
	value.NextFocus = normalizeString(value.NextFocus)
	for i := range value.SkillScores {
		value.SkillScores[i].Skill = normalizeString(value.SkillScores[i].Skill)
	}
	for i := range value.TGOAssessments {
		value.TGOAssessments[i].Code = normalizeString(value.TGOAssessments[i].Code)
		value.TGOAssessments[i].Status = normalizeString(value.TGOAssessments[i].Status)
		value.TGOAssessments[i].Evidence = normalizeString(value.TGOAssessments[i].Evidence)
	}
	for i := range value.CompletedTGOChecks {
		value.CompletedTGOChecks[i].Code = normalizeString(value.CompletedTGOChecks[i].Code)
		value.CompletedTGOChecks[i].Status = normalizeString(value.CompletedTGOChecks[i].Status)
		value.CompletedTGOChecks[i].Evidence = normalizeString(value.CompletedTGOChecks[i].Evidence)
	}
	for i := range value.Annotations {
		value.Annotations[i].Quote = normalizeString(value.Annotations[i].Quote)
		value.Annotations[i].TGOCode = normalizeString(value.Annotations[i].TGOCode)
		value.Annotations[i].Category = normalizeString(value.Annotations[i].Category)
		value.Annotations[i].Comment = normalizeString(value.Annotations[i].Comment)
		value.Annotations[i].Severity = normalizeString(value.Annotations[i].Severity)
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

func EmptyDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func JoinOrDefault(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, ", ")
}

func JoinTGOs(tgos []domain.TGO) string {
	if len(tgos) == 0 {
		return "none"
	}
	var parts []string
	for _, tgo := range tgos {
		parts = append(parts, fmt.Sprintf("%s: %s", tgo.Code, tgo.Description))
	}
	return strings.Join(parts, " | ")
}

func ToDomainAssessments(values []TGOAssessment) []domain.TGOAssessment {
	out := make([]domain.TGOAssessment, 0, len(values))
	for _, value := range values {
		out = append(out, domain.TGOAssessment{
			TGOCode:  value.Code,
			Status:   value.Status,
			Evidence: value.Evidence,
		})
	}
	return out
}

func ToDomainAnnotations(values []Annotation) []domain.ReviewAnnotation {
	out := make([]domain.ReviewAnnotation, 0, len(values))
	for _, value := range values {
		out = append(out, domain.ReviewAnnotation{
			Quote:    value.Quote,
			TGOCode:  value.TGOCode,
			Category: value.Category,
			Comment:  value.Comment,
			Severity: value.Severity,
		})
	}
	return out
}
