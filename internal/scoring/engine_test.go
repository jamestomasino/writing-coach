package scoring

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

type evidenceEnvelope struct {
	RubricID string `json:"rubric_id"`
	Domain   string `json:"domain"`
}

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

func TestEngineScoresAllDomainRubricFamilies(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	cases := []struct {
		name               string
		options            analyzer.ContextOptions
		expectedRubricHint string
	}{
		{
			name:               "general fallback",
			options:            analyzer.ContextOptions{TreeSlug: "open-practice-track", WritingType: "journal response"},
			expectedRubricHint: "general",
		},
		{
			name:               "fiction",
			options:            analyzer.ContextOptions{TreeSlug: "story-craft-track", WritingType: "fiction"},
			expectedRubricHint: "fiction",
		},
		{
			name:               "fantasy",
			options:            analyzer.ContextOptions{TreeSlug: "fantasy-fiction-track", WritingType: "fantasy fiction"},
			expectedRubricHint: "fantasy",
		},
		{
			name:               "technical",
			options:            analyzer.ContextOptions{TreeSlug: "technical-writing-track", WritingType: "technical writing"},
			expectedRubricHint: "technical",
		},
		{
			name:               "academic",
			options:            analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"},
			expectedRubricHint: "academic",
		},
		{
			name:               "professional",
			options:            analyzer.ContextOptions{TreeSlug: "professional-writing-track", WritingType: "professional writing"},
			expectedRubricHint: "professional",
		},
		{
			name:               "thought leadership",
			options:            analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership"},
			expectedRubricHint: "thought",
		},
		{
			name:               "marketing",
			options:            analyzer.ContextOptions{TreeSlug: "marketing-writing-track", WritingType: "marketing writing"},
			expectedRubricHint: "marketing",
		},
	}

	report := analyzer.Report{
		Metrics: map[string]int{
			"word_count":            540,
			"avg_sentence_length":   15,
			"paragraph_count":       8,
			"nlp_readability_grade": 9,
			"nlp_passive_sentences": 2,
		},
		Findings: []analyzer.Finding{
			{Category: "clarity"},
			{Category: "style"},
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sub := domain.Submission{ID: int64(100 + idx), Content: "fixture", WordCount: 540}
			scores, err := engine.ScoreSubmission(sub, report, tc.options, []domain.TGO{{Code: "claim-clarity"}})
			if err != nil {
				t.Fatalf("score: %v", err)
			}
			if len(scores) == 0 {
				t.Fatalf("expected scores for %s", tc.name)
			}

			for _, score := range scores {
				if score.Score < 1 || score.Score > 5 {
					t.Fatalf("score out of range for %s: %d", score.Skill, score.Score)
				}
				if score.ScoreSource != "deterministic" {
					t.Fatalf("score source = %q", score.ScoreSource)
				}
				if score.ScoreEvidenceJSON == "" || score.ScoreEvidenceJSON == "{}" {
					t.Fatalf("missing evidence for %s", score.Skill)
				}

				var evidence evidenceEnvelope
				if err := json.Unmarshal([]byte(score.ScoreEvidenceJSON), &evidence); err != nil {
					t.Fatalf("evidence decode: %v", err)
				}
				if evidence.RubricID == "" {
					t.Fatalf("missing rubric_id in evidence for %s", score.Skill)
				}
				if evidence.Domain == "" {
					t.Fatalf("missing domain in evidence for %s", score.Skill)
				}
				if tc.expectedRubricHint != "" && !strings.Contains(strings.ToLower(evidence.RubricID), strings.ToLower(tc.expectedRubricHint)) {
					t.Fatalf("rubric_id %q does not include %q", evidence.RubricID, tc.expectedRubricHint)
				}
			}
		})
	}
}
