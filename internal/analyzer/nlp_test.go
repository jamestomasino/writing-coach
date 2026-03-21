package analyzer

import (
	"context"
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
			"metrics":{"nlp_long_sentences":2,"nlp_passive_sentences":1},
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
