package analyzer

import (
	"context"
	"regexp"
	"strings"
)

type Heuristic struct{}

var adverbPattern = regexp.MustCompile(`\b\w+ly\b`)

func (Heuristic) Name() string {
	return "heuristic"
}

func (Heuristic) Analyze(_ context.Context, text string) (Report, error) {
	wordCount := len(strings.Fields(text))
	sentenceCount := estimateSentenceCount(text)
	paragraphCount := countParagraphs(text)
	dialogueLines := strings.Count(text, "\"")
	asCount := strings.Count(strings.ToLower(text), " as ")
	adverbCount := len(adverbPattern.FindAllString(strings.ToLower(text), -1))

	report := Report{
		Metrics: map[string]int{
			"word_count":      wordCount,
			"sentence_count":  sentenceCount,
			"paragraph_count": paragraphCount,
			"dialogue_marks":  dialogueLines,
			"adverb_count":    adverbCount,
			"comparison_as":   asCount,
		},
	}

	avgSentenceLength := 0
	if sentenceCount > 0 {
		avgSentenceLength = wordCount / sentenceCount
		report.Metrics["avg_sentence_length"] = avgSentenceLength
	}

	if avgSentenceLength > 24 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "clarity",
			Severity: "warning",
			Message:  "Average sentence length is high; simplify or vary sentence rhythm to improve clarity.",
		})
	}
	if avgSentenceLength > 0 && avgSentenceLength < 8 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "rhythm",
			Severity: "warning",
			Message:  "Average sentence length is very short; the scene may need more modulation and connective tissue.",
		})
	}
	if adverbCount > max(3, wordCount/120) {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "prose precision",
			Severity: "warning",
			Message:  "Adverb density is elevated; check whether stronger verbs can carry more of the weight.",
		})
	}
	if asCount > 5 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "image freshness",
			Severity: "warning",
			Message:  "Frequent comparative phrasing with 'as' may be flattening image precision.",
		})
	}
	if paragraphCount == 1 && wordCount > 250 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "scene architecture",
			Severity: "warning",
			Message:  "A long single paragraph can obscure beat changes; consider clearer scene staging.",
		})
	}
	if dialogueLines == 0 && wordCount > 700 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "dialogue intelligence",
			Severity: "note",
			Message:  "There is no quoted dialogue in a fairly long scene; confirm that silence is a deliberate choice.",
		})
	}
	if wordCount < 500 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "scene architecture",
			Severity: "warning",
			Message:  "The draft is brief enough that escalation may be arriving too quickly.",
		})
	}

	return report, nil
}

func estimateSentenceCount(text string) int {
	count := strings.Count(text, ".") + strings.Count(text, "!") + strings.Count(text, "?")
	if count == 0 && strings.TrimSpace(text) != "" {
		return 1
	}
	return count
}

func countParagraphs(text string) int {
	parts := strings.Split(strings.TrimSpace(text), "\n\n")
	count := 0
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			count++
		}
	}
	if count == 0 && strings.TrimSpace(text) != "" {
		return 1
	}
	return count
}
