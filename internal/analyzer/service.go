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
	var reports []Report
	for _, analyzer := range s.analyzers {
		report, err := analyzer.Analyze(ctx, text)
		if err != nil {
			reports = append(reports, Report{
				Warnings: []string{analyzer.Name() + ": " + err.Error()},
			})
			continue
		}
		reports = append(reports, report)
	}
	return Merge(reports...)
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
