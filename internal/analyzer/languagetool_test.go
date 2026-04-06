package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLanguageToolAnalyzeWithoutBaseURLWarns(t *testing.T) {
	report, err := (LanguageTool{}).Analyze(context.Background(), "text")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(report.Warnings) != 1 || !strings.Contains(report.Warnings[0], "not configured") {
		t.Fatalf("unexpected warnings: %#v", report.Warnings)
	}
}

func TestLanguageToolAnalyzeUnsupportedLanguageWarns(t *testing.T) {
	report, err := (LanguageTool{BaseURL: "http://example.com"}).AnalyzeWithContext(context.Background(), "texto", ContextOptions{
		WritingLanguage: "es",
	})
	if err != nil {
		t.Fatalf("analyze with context: %v", err)
	}
	if len(report.Warnings) != 1 || !strings.Contains(report.Warnings[0], "not configured yet") {
		t.Fatalf("unexpected warnings: %#v", report.Warnings)
	}
}

func TestLanguageToolAnalyzeRequestFailures(t *testing.T) {
	nonOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer nonOK.Close()

	_, err := (LanguageTool{BaseURL: nonOK.URL}).AnalyzeWithContext(context.Background(), "text", ContextOptions{WritingLanguage: "en"})
	if err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("expected request failed error, got %v", err)
	}

	badJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{invalid"))
	}))
	defer badJSON.Close()

	_, err = (LanguageTool{BaseURL: badJSON.URL}).AnalyzeWithContext(context.Background(), "text", ContextOptions{WritingLanguage: "en"})
	if err == nil {
		t.Fatal("expected JSON decode error")
	}
}

func TestLanguageToolAnalyzeParsesMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/check" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.Form.Get("language") != "en-US" {
			t.Fatalf("expected en-US language code, got %q", r.Form.Get("language"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"matches":[
				{
					"message":"Prefer active voice",
					"rule":{"issueType":"style","category":{"name":"Clarity"}}
				},
				{
					"message":"Repeated word",
					"rule":{"issueType":"duplication","category":{"name":""}}
				}
			]
		}`))
	}))
	defer server.Close()

	report, err := (LanguageTool{BaseURL: server.URL}).AnalyzeWithContext(context.Background(), "Some text", ContextOptions{
		WritingLanguage: "en",
	})
	if err != nil {
		t.Fatalf("analyze with context: %v", err)
	}
	if report.Metrics["languagetool_matches"] != 2 {
		t.Fatalf("expected 2 matches metric, got %d", report.Metrics["languagetool_matches"])
	}
	if len(report.Findings) != 2 {
		t.Fatalf("expected 2 findings, got %#v", report.Findings)
	}
	if report.Findings[0].Category != "Clarity" {
		t.Fatalf("expected category from payload, got %#v", report.Findings[0])
	}
	if report.Findings[1].Category != "duplication" {
		t.Fatalf("expected fallback issue type category, got %#v", report.Findings[1])
	}
}

func TestNormalizeSeverityAndCoalesce(t *testing.T) {
	if coalesce("primary", "fallback") != "primary" {
		t.Fatal("expected non-empty primary")
	}
	if coalesce("   ", "fallback") != "fallback" {
		t.Fatal("expected fallback when primary is blank")
	}

	cases := map[string]string{
		"error":   "error",
		"warning": "warning",
		"note":    "note",
		"INFO":    "note",
	}
	for input, expected := range cases {
		if got := normalizeSeverity(input); got != expected {
			t.Fatalf("normalizeSeverity(%q) = %q, want %q", input, got, expected)
		}
	}
}
