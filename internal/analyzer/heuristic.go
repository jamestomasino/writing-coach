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

func (h Heuristic) Analyze(ctx context.Context, text string) (Report, error) {
	return h.AnalyzeWithContext(ctx, text, ContextOptions{})
}

func (Heuristic) AnalyzeWithContext(_ context.Context, text string, options ContextOptions) (Report, error) {
	wordCount := len(strings.Fields(text))
	sentenceCount := estimateSentenceCount(text)
	paragraphCount := countParagraphs(text)
	dialogueLines := strings.Count(text, "\"")
	asCount := strings.Count(strings.ToLower(text), " as ")
	adverbCount := len(adverbPattern.FindAllString(strings.ToLower(text), -1))
	domain := DomainForContext(options)

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
			Message:  averageSentenceLengthHighMessage(domain),
		})
	}
	if avgSentenceLength > 0 && avgSentenceLength < 8 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "rhythm",
			Severity: "warning",
			Message:  shortSentenceRhythmMessage(domain),
		})
	}
	if adverbCount > max(3, wordCount/120) {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "prose precision",
			Severity: "warning",
			Message:  adverbDensityMessage(domain),
		})
	}
	if asCount > 5 && (domain == DomainFiction || domain == DomainFantasy || domain == DomainThoughtLeadership || domain == DomainGeneral) {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "image freshness",
			Severity: "warning",
			Message:  comparisonDensityMessage(domain),
		})
	}
	if paragraphCount == 1 && wordCount > 250 {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: paragraphCategory(domain),
			Severity: "warning",
			Message:  longSingleParagraphMessage(domain),
		})
	}
	if dialogueLines == 0 && wordCount > 700 && (domain == DomainFiction || domain == DomainFantasy) {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: "dialogue intelligence",
			Severity: "note",
			Message:  "There is no quoted dialogue in a fairly long scene; confirm that silence is a deliberate choice.",
		})
	}
	if wordCount < minimumExpectedWords(domain) {
		report.Findings = append(report.Findings, Finding{
			Analyzer: "heuristic",
			Category: brevityCategory(domain),
			Severity: "warning",
			Message:  briefDraftMessage(domain),
		})
	}

	return report, nil
}

func averageSentenceLengthHighMessage(domain string) string {
	switch domain {
	case DomainTechnical:
		return "Average sentence length is high; shorten a few instruction or explanation lines so steps stay easy to scan."
	case DomainAcademic:
		return "Average sentence length is high; tighten a few sentence frames so the argument remains easier to track."
	case DomainProfessional:
		return "Average sentence length is high; shorten key sentences so the request and ownership stay obvious."
	case DomainMarketing:
		return "Average sentence length is high; trim a few lines so the message lands faster and the main promise stays clear."
	case DomainThoughtLeadership:
		return "Average sentence length is high; tighten a few sentences so the claim progression stays easy to follow."
	default:
		return "Average sentence length is high; simplify or vary sentence rhythm to improve clarity."
	}
}

func shortSentenceRhythmMessage(domain string) string {
	switch domain {
	case DomainTechnical:
		return "Average sentence length is very short; add a bit more connective explanation so the reader can see how steps and results fit together."
	case DomainAcademic:
		return "Average sentence length is very short; add a little more connective reasoning so the argument does not feel list-like."
	case DomainProfessional:
		return "Average sentence length is very short; add a little more context so the action and rationale connect cleanly."
	case DomainMarketing:
		return "Average sentence length is very short; vary sentence rhythm so the copy does not start to feel choppy."
	case DomainThoughtLeadership:
		return "Average sentence length is very short; add some connective tissue so the idea progression feels deliberate."
	default:
		return "Average sentence length is very short; the draft may need more modulation and connective tissue."
	}
}

func adverbDensityMessage(domain string) string {
	switch domain {
	case DomainTechnical:
		return "Modifier density is elevated; trim extra qualifiers so the instructions read more directly."
	case DomainAcademic:
		return "Modifier density is elevated; remove softening words that are not helping precision."
	case DomainProfessional:
		return "Modifier density is elevated; tighten qualifiers so the message sounds direct and controlled."
	case DomainMarketing:
		return "Modifier density is elevated; trim extra intensifiers so the copy sounds more confident and specific."
	default:
		return "Adverb density is elevated; check whether stronger verbs can carry more of the weight."
	}
}

func comparisonDensityMessage(domain string) string {
	if domain == DomainThoughtLeadership {
		return "Frequent comparative phrasing with 'as' may be making the ideas feel more ornamental than precise."
	}
	return "Frequent comparative phrasing with 'as' may be flattening image precision."
}

func paragraphCategory(domain string) string {
	switch domain {
	case DomainTechnical:
		return "scanability"
	case DomainAcademic, DomainProfessional, DomainThoughtLeadership:
		return "structure"
	default:
		return "scene architecture"
	}
}

func longSingleParagraphMessage(domain string) string {
	switch domain {
	case DomainTechnical:
		return "A long single paragraph makes the piece harder to scan; break it into clearer steps or chunks."
	case DomainAcademic:
		return "A long single paragraph can hide the progression of the argument; break it where the reasoning shifts."
	case DomainProfessional:
		return "A long single paragraph can bury the key action or decision; split it so the main points stand out."
	case DomainMarketing:
		return "A long single paragraph can bury the core message; split it so the value and call to action are easier to spot."
	case DomainThoughtLeadership:
		return "A long single paragraph can obscure the turn of the idea; break it where the argument or example changes."
	default:
		return "A long single paragraph can obscure beat changes; consider clearer staging."
	}
}

func minimumExpectedWords(domain string) int {
	switch domain {
	case DomainMarketing:
		return 160
	case DomainProfessional:
		return 220
	case DomainTechnical:
		return 260
	case DomainAcademic, DomainThoughtLeadership:
		return 320
	default:
		return 500
	}
}

func brevityCategory(domain string) string {
	switch domain {
	case DomainTechnical:
		return "coverage"
	case DomainAcademic, DomainThoughtLeadership:
		return "development"
	case DomainProfessional:
		return "completeness"
	case DomainMarketing:
		return "message development"
	default:
		return "scene architecture"
	}
}

func briefDraftMessage(domain string) string {
	switch domain {
	case DomainTechnical:
		return "The draft is brief for instructional writing; check whether the reader has enough steps, examples, or expected outcomes."
	case DomainAcademic:
		return "The draft is brief for an argument-driven piece; check whether the claim, support, and implications are fully developed."
	case DomainProfessional:
		return "The draft is brief for a professional piece; confirm that the ask, context, and next steps are all present."
	case DomainMarketing:
		return "The draft is brief enough that the core value may not be fully supported yet; make sure the offer and next step are unmistakable."
	case DomainThoughtLeadership:
		return "The draft is brief for an idea-driven piece; make sure the central claim has enough development and support."
	default:
		return "The draft is brief enough that escalation may be arriving too quickly."
	}
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
