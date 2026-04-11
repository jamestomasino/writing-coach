package analyzer

import (
	"context"
	"strconv"
	"strings"
)

type Service struct {
	analyzers []Analyzer
}

func NewService(analyzers ...Analyzer) Service {
	return Service{analyzers: analyzers}
}

func (s Service) Analyze(ctx context.Context, text string) Report {
	return s.AnalyzeWithContext(ctx, text, ContextOptions{})
}

func (s Service) AnalyzeWithContext(ctx context.Context, text string, options ContextOptions) Report {
	var results []analyzerReportResult
	for _, analyzer := range s.analyzers {
		report, err := analyzeWithOptions(ctx, analyzer, text, options)
		if err != nil {
			results = append(results, analyzerReportResult{
				analyzer: analyzer.Name(),
				report: Report{
					Warnings: []string{analyzer.Name() + ": " + err.Error()},
				},
			})
			continue
		}
		results = append(results, analyzerReportResult{
			analyzer: analyzer.Name(),
			report:   report,
		})
	}
	return mergeWithCoverageArbitration(options, results...)
}

func analyzeWithOptions(ctx context.Context, analyzer Analyzer, text string, options ContextOptions) (Report, error) {
	if contextual, ok := analyzer.(ContextualAnalyzer); ok {
		return contextual.AnalyzeWithContext(ctx, text, options)
	}
	return analyzer.Analyze(ctx, text)
}

func TopFindings(report Report, limit int) []string {
	if limit <= 0 {
		return nil
	}

	items := make([]string, 0, limit)
	for _, finding := range report.Findings {
		items = append(items, finding.Message)
		if len(items) == limit {
			break
		}
	}
	return items
}

func Summary(report Report) string {
	var parts []string
	if value, ok := report.Metrics["word_count"]; ok {
		parts = append(parts, "words="+itoa(value))
	}
	if value, ok := report.Metrics["avg_sentence_length"]; ok {
		parts = append(parts, "avg_sentence_length="+itoa(value))
	}
	if value, ok := report.Metrics["adverb_count"]; ok {
		parts = append(parts, "adverbs="+itoa(value))
	}
	if value, ok := report.Metrics["languagetool_matches"]; ok {
		parts = append(parts, "languagetool_matches="+itoa(value))
	}
	if value, ok := report.Metrics["nlp_long_sentences"]; ok {
		parts = append(parts, "nlp_long_sentences="+itoa(value))
	}
	if value, ok := report.Metrics["nlp_passive_sentences"]; ok {
		parts = append(parts, "nlp_passive_sentences="+itoa(value))
	}
	return strings.Join(parts, ", ")
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
