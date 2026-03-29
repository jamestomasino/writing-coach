package scoring

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestEngineScoresTechnicalDomainFixture(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	report := analyzer.Report{
		Metrics: map[string]int{
			"word_count":            420,
			"avg_sentence_length":   14,
			"paragraph_count":       6,
			"nlp_readability_grade": 9,
			"nlp_passive_sentences": 1,
		},
		Findings: []analyzer.Finding{{Category: "clarity"}},
	}
	sub := domain.Submission{ID: 42, Content: "Technical draft", WordCount: 420}
	scores, err := engine.ScoreSubmission(sub, report, analyzer.ContextOptions{TreeSlug: "technical-writing-track", WritingType: "technical writing"}, nil)
	if err != nil {
		t.Fatalf("score: %v", err)
	}
	if len(scores) == 0 {
		t.Fatal("expected scores")
	}
	for _, score := range scores {
		if score.ScoreSource != "deterministic" {
			t.Fatalf("score source = %q", score.ScoreSource)
		}
		if score.ScoreVersion == "" {
			t.Fatalf("score version empty for %q", score.Skill)
		}
		if score.ScoreEvidenceJSON == "" {
			t.Fatalf("score evidence empty for %q", score.Skill)
		}
	}
}

func TestEngineAppliesTrackOverrideFixture(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	report := analyzer.Report{
		Metrics: map[string]int{
			"word_count":            700,
			"avg_sentence_length":   16,
			"nlp_readability_grade": 10,
		},
	}
	sub := domain.Submission{ID: 7, Content: "Fantasy draft", WordCount: 700}
	scores, err := engine.ScoreSubmission(sub, report, analyzer.ContextOptions{TreeSlug: "fantasy-fiction-track", WritingType: "fantasy fiction"}, nil)
	if err != nil {
		t.Fatalf("score: %v", err)
	}
	if len(scores) == 0 {
		t.Fatal("expected scores")
	}
	found := false
	for _, score := range scores {
		if score.Skill == "worldbuilding economy" {
			found = true
			if score.Score < 1 || score.Score > 5 {
				t.Fatalf("score out of range: %d", score.Score)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected worldbuilding economy score")
	}
}

func TestEngineFallsBackToGeneralDomain(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	report := analyzer.Report{Metrics: map[string]int{"word_count": 90, "avg_sentence_length": 30}}
	sub := domain.Submission{ID: 9, Content: "Generic draft", WordCount: 90}
	scores, err := engine.ScoreSubmission(sub, report, analyzer.ContextOptions{TreeSlug: "unknown-track", WritingType: "other"}, nil)
	if err != nil {
		t.Fatalf("score: %v", err)
	}
	if len(scores) == 0 {
		t.Fatal("expected fallback scores")
	}
}

func TestEngineCandidateSkillsUsesActiveTGOs(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	report := analyzer.Report{Metrics: map[string]int{"word_count": 500, "avg_sentence_length": 14}}
	sub := domain.Submission{ID: 11, Content: "Draft", WordCount: 500}
	active := []domain.TGO{{Code: "claim-clarity"}}
	scores, err := engine.ScoreSubmission(sub, report, analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership"}, active)
	if err != nil {
		t.Fatalf("score: %v", err)
	}
	if len(scores) == 0 {
		t.Fatal("expected scores")
	}
}
