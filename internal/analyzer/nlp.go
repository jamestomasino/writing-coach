package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type NLP struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (n NLP) Name() string {
	return "nlp"
}

func (n NLP) Analyze(ctx context.Context, text string) (Report, error) {
	return n.AnalyzeWithContext(ctx, text, ContextOptions{})
}

func (n NLP) AnalyzeWithContext(ctx context.Context, text string, options ContextOptions) (Report, error) {
	if n.BaseURL == "" {
		return Report{Warnings: []string{"nlp analyzer not configured"}}, nil
	}

	client := n.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	body, err := json.Marshal(struct {
		Text             string `json:"text"`
		TreeSlug         string `json:"tree_slug,omitempty"`
		WritingType      string `json:"writing_type,omitempty"`
		AssignmentFormat string `json:"assignment_format,omitempty"`
		TemplateKey      string `json:"template_key,omitempty"`
		Domain           string `json:"domain,omitempty"`
	}{
		Text:             text,
		TreeSlug:         strings.TrimSpace(options.TreeSlug),
		WritingType:      strings.TrimSpace(options.WritingType),
		AssignmentFormat: strings.TrimSpace(options.AssignmentFormat),
		TemplateKey:      strings.TrimSpace(options.TemplateKey),
		Domain:           DomainForContext(options),
	})
	if err != nil {
		return Report{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(n.BaseURL, "/")+"/analyze", bytes.NewReader(body))
	if err != nil {
		return Report{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Report{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Report{}, errors.New("nlp analyzer request failed")
	}

	var payload struct {
		Metrics  map[string]int `json:"metrics"`
		Findings []struct {
			Category string `json:"category"`
			Severity string `json:"severity"`
			Message  string `json:"message"`
		} `json:"findings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Report{}, err
	}

	report := Report{
		Metrics: payload.Metrics,
	}
	for _, finding := range payload.Findings {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "nlp",
			Category: strings.TrimSpace(finding.Category),
			Severity: normalizeSeverity(finding.Severity),
			Message:  strings.TrimSpace(finding.Message),
		})
	}
	return report, nil
}
