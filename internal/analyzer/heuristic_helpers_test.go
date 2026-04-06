package analyzer

import (
	"strings"
	"testing"
)

func TestHeuristicHelperMessagesByDomain(t *testing.T) {
	if !strings.Contains(averageSentenceLengthHighMessage(DomainTechnical), "instruction") {
		t.Fatalf("technical average sentence message not domain-specific")
	}
	if !strings.Contains(averageSentenceLengthHighMessage(DomainGeneral), "clarity") {
		t.Fatalf("default average sentence message not used")
	}

	if !strings.Contains(shortSentenceRhythmMessage(DomainAcademic), "argument") {
		t.Fatalf("academic short rhythm message not domain-specific")
	}
	if !strings.Contains(shortSentenceRhythmMessage(DomainGeneral), "connective tissue") {
		t.Fatalf("default short rhythm message not used")
	}

	if !strings.Contains(adverbDensityMessage(DomainMarketing), "intensifiers") {
		t.Fatalf("marketing adverb message not domain-specific")
	}
	if !strings.Contains(adverbDensityMessage(DomainGeneral), "stronger verbs") {
		t.Fatalf("default adverb message not used")
	}

	if !strings.Contains(comparisonDensityMessage(DomainThoughtLeadership), "ornamental") {
		t.Fatalf("thought leadership comparison message not domain-specific")
	}
	if !strings.Contains(comparisonDensityMessage(DomainFiction), "image precision") {
		t.Fatalf("default comparison message not used")
	}
}

func TestHeuristicCategoriesAndThresholds(t *testing.T) {
	if paragraphCategory(DomainTechnical) != "scanability" {
		t.Fatalf("unexpected paragraph category for technical")
	}
	if paragraphCategory(DomainProfessional) != "structure" {
		t.Fatalf("unexpected paragraph category for professional")
	}
	if paragraphCategory(DomainGeneral) != "scene architecture" {
		t.Fatalf("unexpected paragraph category default")
	}

	if minimumExpectedWords(DomainMarketing) != 160 || minimumExpectedWords(DomainAcademic) != 320 || minimumExpectedWords(DomainGeneral) != 500 {
		t.Fatalf("unexpected minimum expected words thresholds")
	}

	if brevityCategory(DomainTechnical) != "coverage" || brevityCategory(DomainMarketing) != "message development" || brevityCategory(DomainGeneral) != "scene architecture" {
		t.Fatalf("unexpected brevity categories")
	}
}

func TestBriefDraftAndParagraphMessages(t *testing.T) {
	if !strings.Contains(longSingleParagraphMessage(DomainProfessional), "key action") {
		t.Fatalf("professional paragraph message not used")
	}
	if !strings.Contains(longSingleParagraphMessage(DomainGeneral), "beat changes") {
		t.Fatalf("default paragraph message not used")
	}

	if !strings.Contains(briefDraftMessage(DomainTechnical), "instructional writing") {
		t.Fatalf("technical brief draft message not used")
	}
	if !strings.Contains(briefDraftMessage(DomainGeneral), "escalation") {
		t.Fatalf("default brief draft message not used")
	}
}

func TestSentenceAndParagraphCounters(t *testing.T) {
	if estimateSentenceCount("No punctuation here") != 1 {
		t.Fatalf("expected implicit single sentence")
	}
	if estimateSentenceCount("One. Two! Three?") != 3 {
		t.Fatalf("expected explicit sentence count")
	}
	if countParagraphs("alpha\n\nbeta") != 2 {
		t.Fatalf("expected two paragraphs")
	}
	if countParagraphs("   lone paragraph   ") != 1 {
		t.Fatalf("expected one paragraph for non-empty text")
	}
}
