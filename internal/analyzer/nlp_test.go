package analyzer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNLPAnalyzerReadsFindingsAndMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/analyze" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"metrics":{
				"nlp_long_sentences":2,
				"nlp_passive_sentences":1,
				"nlp_claim_evidence_coverage":67,
				"nlp_coref_ambiguity_count":1,
				"nlp_semantic_repetition_ratio":12,
				"nlp_topic_drift_score":44
			},
			"findings":[
				{"category":"clarity","severity":"warning","message":"Several sentences are carrying too many clauses."}
			]
		}`))
	}))
	defer server.Close()

	report, err := (NLP{BaseURL: server.URL}).Analyze(context.Background(), "text")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if report.Metrics["nlp_long_sentences"] != 2 {
		t.Fatalf("long sentence metric = %d", report.Metrics["nlp_long_sentences"])
	}
	if report.Metrics["nlp_claim_evidence_coverage"] != 67 {
		t.Fatalf("claim/evidence coverage metric = %d", report.Metrics["nlp_claim_evidence_coverage"])
	}
	if len(report.Findings) != 1 {
		t.Fatalf("findings = %#v", report.Findings)
	}
	if report.Findings[0].Analyzer != "nlp" {
		t.Fatalf("unexpected analyzer: %#v", report.Findings[0])
	}
}

func TestNLPAnalyzerWithoutURLWarnsInsteadOfFailing(t *testing.T) {
	report, err := (NLP{}).Analyze(context.Background(), "text")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(report.Warnings) != 1 {
		t.Fatalf("warnings = %#v", report.Warnings)
	}
}

func TestNLPAnalyzerSendsContext(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"metrics":{},"findings":[]}`))
	}))
	defer server.Close()

	_, err := (NLP{BaseURL: server.URL}).AnalyzeWithContext(context.Background(), "text", ContextOptions{
		WritingLanguage:  "en",
		WritingType:      "technical writing",
		AssignmentFormat: "how-to guide",
		TemplateKey:      "technical-writing",
		TreeSlug:         "technical-writing-track",
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if payload["domain"] != DomainTechnical {
		t.Fatalf("expected technical domain, got %#v", payload["domain"])
	}
	if payload["writing_language"] != "en" {
		t.Fatalf("expected writing language, got %#v", payload["writing_language"])
	}
	if payload["writing_type"] != "technical writing" {
		t.Fatalf("expected writing type, got %#v", payload["writing_type"])
	}
}

func TestNLPAnalyzerSkipsUnsupportedLanguage(t *testing.T) {
	report, err := (NLP{BaseURL: "http://example.com"}).AnalyzeWithContext(context.Background(), "texto", ContextOptions{
		WritingLanguage: "es",
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected warning for unsupported language, got %#v", report)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected no findings, got %#v", report.Findings)
	}
}
