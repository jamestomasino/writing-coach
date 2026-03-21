package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LanguageTool struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (lt LanguageTool) Name() string {
	return "languagetool"
}

func (lt LanguageTool) Analyze(ctx context.Context, text string) (Report, error) {
	return lt.AnalyzeWithContext(ctx, text, ContextOptions{})
}

func (lt LanguageTool) AnalyzeWithContext(ctx context.Context, text string, options ContextOptions) (Report, error) {
	if lt.BaseURL == "" {
		return Report{Warnings: []string{"languagetool not configured"}}, nil
	}
	code := languageToolCode(options.WritingLanguage)
	if code == "" {
		return Report{Warnings: []string{unsupportedLanguageWarning(lt.Name(), options.WritingLanguage)}}, nil
	}

	client := lt.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	form := url.Values{}
	form.Set("text", text)
	form.Set("language", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(lt.BaseURL, "/")+"/v2/check", strings.NewReader(form.Encode()))
	if err != nil {
		return Report{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return Report{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Report{}, errors.New("languagetool request failed")
	}

	var payload struct {
		Matches []struct {
			Message string `json:"message"`
			Rule    struct {
				IssueType string `json:"issueType"`
				Category  struct {
					Name string `json:"name"`
				} `json:"category"`
			} `json:"rule"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Report{}, err
	}

	report := Report{
		Metrics: map[string]int{
			"languagetool_matches": len(payload.Matches),
		},
	}
	for _, match := range payload.Matches {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "languagetool",
			Category: coalesce(match.Rule.Category.Name, match.Rule.IssueType),
			Severity: "warning",
			Message:  match.Message,
		})
	}
	return report, nil
}

func coalesce(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(value) {
	case "error":
		return "error"
	case "warning":
		return "warning"
	default:
		return "note"
	}
}
