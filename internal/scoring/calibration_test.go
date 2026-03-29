package scoring

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestDomainRubricCalibrationProfiles(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	domains := []struct {
		name    string
		options analyzer.ContextOptions
	}{
		{name: "general", options: analyzer.ContextOptions{TreeSlug: "open-practice-track", WritingType: "reflection"}},
		{name: "fiction", options: analyzer.ContextOptions{TreeSlug: "story-craft-track", WritingType: "fiction"}},
		{name: "fantasy", options: analyzer.ContextOptions{TreeSlug: "fantasy-fiction-track", WritingType: "fantasy fiction"}},
		{name: "technical", options: analyzer.ContextOptions{TreeSlug: "technical-writing-track", WritingType: "technical writing"}},
		{name: "academic", options: analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"}},
		{name: "professional", options: analyzer.ContextOptions{TreeSlug: "professional-writing-track", WritingType: "professional writing"}},
		{name: "thought leadership", options: analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership"}},
		{name: "marketing", options: analyzer.ContextOptions{TreeSlug: "marketing-writing-track", WritingType: "marketing writing"}},
	}

	weak := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             110,
			"avg_sentence_length":    31,
			"paragraph_count":        1,
			"nlp_readability_grade":  16,
			"nlp_passive_sentences":  8,
			"nlp_unique_token_ratio": 30,
			"nlp_long_sentences":     8,
			"adverb_count":           20,
		},
		Findings: []analyzer.Finding{
			{Category: "clarity"},
			{Category: "clarity"},
			{Category: "readability"},
			{Category: "structure"},
			{Category: "sentence control"},
		},
	}
	baseline := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             360,
			"avg_sentence_length":    18,
			"paragraph_count":        4,
			"nlp_readability_grade":  11,
			"nlp_passive_sentences":  2,
			"nlp_unique_token_ratio": 42,
			"nlp_long_sentences":     2,
			"adverb_count":           10,
		},
		Findings: []analyzer.Finding{
			{Category: "clarity"},
			{Category: "structure"},
		},
	}
	strong := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             760,
			"avg_sentence_length":    14,
			"paragraph_count":        8,
			"nlp_readability_grade":  9,
			"nlp_passive_sentences":  1,
			"nlp_unique_token_ratio": 52,
			"nlp_long_sentences":     1,
			"adverb_count":           4,
		},
		Findings: []analyzer.Finding{},
	}

	for i, tc := range domains {
		t.Run(tc.name, func(t *testing.T) {
			sub := domain.Submission{ID: int64(200 + i), Content: "calibration fixture", WordCount: 760}
			weakScores, err := engine.ScoreSubmission(sub, weak, tc.options, nil)
			if err != nil {
				t.Fatalf("weak score: %v", err)
			}
			baseScores, err := engine.ScoreSubmission(sub, baseline, tc.options, nil)
			if err != nil {
				t.Fatalf("baseline score: %v", err)
			}
			strongScores, err := engine.ScoreSubmission(sub, strong, tc.options, nil)
			if err != nil {
				t.Fatalf("strong score: %v", err)
			}

			weakAvg := averageScore(weakScores)
			baseAvg := averageScore(baseScores)
			strongAvg := averageScore(strongScores)

			if !(weakAvg < baseAvg && baseAvg <= strongAvg) {
				t.Fatalf("expected weak < baseline <= strong, got weak=%.2f baseline=%.2f strong=%.2f", weakAvg, baseAvg, strongAvg)
			}
			if strongAvg-weakAvg < 0.8 {
				t.Fatalf("expected calibration spread >= 0.8, got %.2f", strongAvg-weakAvg)
			}
		})
	}
}

func TestTopEndSeparationFantasyAndTechnical(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	tracks := []struct {
		name    string
		options analyzer.ContextOptions
	}{
		{name: "fantasy", options: analyzer.ContextOptions{TreeSlug: "fantasy-fiction-track", WritingType: "fantasy fiction"}},
		{name: "technical", options: analyzer.ContextOptions{TreeSlug: "technical-writing-track", WritingType: "technical writing"}},
	}

	strongNotElite := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             780,
			"avg_sentence_length":    14,
			"paragraph_count":        8,
			"nlp_readability_grade":  9,
			"nlp_passive_sentences":  1,
			"nlp_unique_token_ratio": 53,
			"nlp_long_sentences":     1,
			"adverb_count":           4,
		},
		Findings: []analyzer.Finding{},
	}
	elite := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             960,
			"avg_sentence_length":    13,
			"paragraph_count":        9,
			"nlp_readability_grade":  8,
			"nlp_passive_sentences":  0,
			"nlp_unique_token_ratio": 58,
			"nlp_long_sentences":     0,
			"adverb_count":           3,
		},
		Findings: []analyzer.Finding{},
	}

	for i, track := range tracks {
		t.Run(track.name, func(t *testing.T) {
			sub := domain.Submission{ID: int64(800 + i), Content: "top-end fixture", WordCount: 960}
			strongScores, err := engine.ScoreSubmission(sub, strongNotElite, track.options, nil)
			if err != nil {
				t.Fatalf("strong score: %v", err)
			}
			eliteScores, err := engine.ScoreSubmission(sub, elite, track.options, nil)
			if err != nil {
				t.Fatalf("elite score: %v", err)
			}

			if maxScore(strongScores) >= 5 {
				t.Fatalf("expected top-end separation (no 5s) for strong-not-elite in %s, got max=%d", track.name, maxScore(strongScores))
			}
			if maxScore(eliteScores) < 5 {
				t.Fatalf("expected elite profile to reach 5 in %s, got max=%d", track.name, maxScore(eliteScores))
			}
		})
	}
}

func TestTopEndSeparationAllBuiltInTracks(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	strongNotElite := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             780,
			"avg_sentence_length":    14,
			"paragraph_count":        8,
			"nlp_readability_grade":  9,
			"nlp_passive_sentences":  1,
			"nlp_unique_token_ratio": 53,
			"nlp_long_sentences":     1,
			"adverb_count":           4,
		},
		Findings: []analyzer.Finding{},
	}
	elite := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             960,
			"avg_sentence_length":    13,
			"paragraph_count":        9,
			"nlp_readability_grade":  8,
			"nlp_passive_sentences":  0,
			"nlp_unique_token_ratio": 58,
			"nlp_long_sentences":     0,
			"adverb_count":           3,
		},
		Findings: []analyzer.Finding{},
	}

	for i, tree := range domain.BuiltInTrees {
		tree := tree
		t.Run(tree.Slug, func(t *testing.T) {
			sub := domain.Submission{ID: int64(1400 + i), Content: tree.Title + " calibration fixture", WordCount: 960}
			options := analyzer.ContextOptions{TreeSlug: tree.Slug, WritingType: tree.Title}
			strongScores, err := engine.ScoreSubmission(sub, strongNotElite, options, nil)
			if err != nil {
				t.Fatalf("strong score: %v", err)
			}
			eliteScores, err := engine.ScoreSubmission(sub, elite, options, nil)
			if err != nil {
				t.Fatalf("elite score: %v", err)
			}
			if maxScore(strongScores) >= 5 {
				t.Fatalf("expected no 5s in strong-not-elite profile, got max=%d", maxScore(strongScores))
			}
			if maxScore(eliteScores) < 5 {
				t.Fatalf("expected at least one 5 in elite profile, got max=%d", maxScore(eliteScores))
			}
		})
	}
}

func TestTopEndDistributionAllBuiltInTracks(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	strongNotElite := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             780,
			"avg_sentence_length":    14,
			"paragraph_count":        8,
			"nlp_readability_grade":  9,
			"nlp_passive_sentences":  1,
			"nlp_unique_token_ratio": 53,
			"nlp_long_sentences":     1,
			"adverb_count":           4,
		},
		Findings: []analyzer.Finding{},
	}
	elite := analyzer.Report{
		Metrics: map[string]int{
			"word_count":             960,
			"avg_sentence_length":    13,
			"paragraph_count":        9,
			"nlp_readability_grade":  8,
			"nlp_passive_sentences":  0,
			"nlp_unique_token_ratio": 58,
			"nlp_long_sentences":     0,
			"adverb_count":           3,
		},
		Findings: []analyzer.Finding{},
	}

	for i, tree := range domain.BuiltInTrees {
		tree := tree
		t.Run(tree.Slug, func(t *testing.T) {
			sub := domain.Submission{ID: int64(1800 + i), Content: tree.Title + " distribution fixture", WordCount: 960}
			options := analyzer.ContextOptions{TreeSlug: tree.Slug, WritingType: tree.Title}
			strongScores, err := engine.ScoreSubmission(sub, strongNotElite, options, nil)
			if err != nil {
				t.Fatalf("strong score: %v", err)
			}
			eliteScores, err := engine.ScoreSubmission(sub, elite, options, nil)
			if err != nil {
				t.Fatalf("elite score: %v", err)
			}

			strongAvg := averageScore(strongScores)
			eliteAvg := averageScore(eliteScores)
			if eliteAvg-strongAvg < 0.6 {
				t.Fatalf("expected elite-average lift >= 0.6, got strong=%.2f elite=%.2f", strongAvg, eliteAvg)
			}
			if eliteAvg < 4.6 {
				t.Fatalf("expected elite average >= 4.6, got %.2f", eliteAvg)
			}
			if countScore(eliteScores, 5) < 1 {
				t.Fatalf("expected elite profile to include at least one score of 5")
			}
		})
	}
}

func averageScore(scores []domain.SkillScore) float64 {
	if len(scores) == 0 {
		return 0
	}
	total := 0
	for _, score := range scores {
		total += score.Score
	}
	return float64(total) / float64(len(scores))
}

func countScore(scores []domain.SkillScore, target int) int {
	count := 0
	for _, score := range scores {
		if score.Score == target {
			count++
		}
	}
	return count
}

func maxScore(scores []domain.SkillScore) int {
	value := 0
	for _, score := range scores {
		if score.Score > value {
			value = score.Score
		}
	}
	return value
}
