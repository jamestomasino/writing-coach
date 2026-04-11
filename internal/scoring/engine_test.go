package scoring

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

type evidenceEnvelope struct {
	RubricID     string   `json:"rubric_id"`
	Domain       string   `json:"domain"`
	AppliedRules []string `json:"applied_rules"`
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

func TestAcademicClaimEvidenceCoverageInfluencesClaimClarity(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	options := analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"}
	sub := domain.Submission{ID: 501, Content: "fixture", WordCount: 780}
	active := []domain.TGO{{Code: "claim-clarity"}}

	strong := analyzer.Report{
		Metrics: map[string]int{
			"word_count":                    780,
			"avg_sentence_length":           25,
			"paragraph_count":               8,
			"nlp_readability_grade":         9,
			"nlp_passive_sentences":         2,
			"nlp_unique_token_ratio":        58,
			"nlp_claim_evidence_coverage":   72,
			"nlp_evidence_marker_count":     4,
			"nlp_coref_ambiguity_count":     1,
			"nlp_semantic_repetition_ratio": 24,
		},
	}
	weak := analyzer.Report{
		Metrics: map[string]int{
			"word_count":                    780,
			"avg_sentence_length":           25,
			"paragraph_count":               8,
			"nlp_readability_grade":         9,
			"nlp_passive_sentences":         2,
			"nlp_unique_token_ratio":        58,
			"nlp_claim_evidence_coverage":   22,
			"nlp_evidence_marker_count":     0,
			"nlp_coref_ambiguity_count":     1,
			"nlp_semantic_repetition_ratio": 24,
		},
	}

	strongScores, err := engine.ScoreSubmission(sub, strong, options, active)
	if err != nil {
		t.Fatalf("score strong: %v", err)
	}
	weakScores, err := engine.ScoreSubmission(sub, weak, options, active)
	if err != nil {
		t.Fatalf("score weak: %v", err)
	}
	strongClaim, strongEvidence := scoreAndEvidenceForSkill(t, strongScores, "claim clarity")
	weakClaim, _ := scoreAndEvidenceForSkill(t, weakScores, "claim clarity")
	if strongClaim <= weakClaim {
		t.Fatalf("expected stronger claim-evidence coverage to lift claim clarity: strong=%d weak=%d", strongClaim, weakClaim)
	}
	if !containsRule(strongEvidence.AppliedRules, "claim-evidence") {
		t.Fatalf("expected claim-evidence rule trace, got %+v", strongEvidence.AppliedRules)
	}
}

func TestSentenceEconomyPenalizesHighSemanticRepetition(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	options := analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"}
	sub := domain.Submission{ID: 502, Content: "fixture", WordCount: 820}
	baseMetrics := map[string]int{
		"word_count":                  820,
		"avg_sentence_length":         21,
		"paragraph_count":             8,
		"nlp_readability_grade":       9,
		"nlp_passive_sentences":       0,
		"nlp_unique_token_ratio":      58,
		"nlp_claim_evidence_coverage": 70,
		"adverb_count":                7,
	}
	lowRep := analyzer.Report{Metrics: cloneMetrics(baseMetrics)}
	lowRep.Metrics["nlp_semantic_repetition_ratio"] = 20
	highRep := analyzer.Report{Metrics: cloneMetrics(baseMetrics)}
	highRep.Metrics["nlp_semantic_repetition_ratio"] = 72

	lowScores, err := engine.ScoreSubmission(sub, lowRep, options, nil)
	if err != nil {
		t.Fatalf("score low repetition: %v", err)
	}
	highScores, err := engine.ScoreSubmission(sub, highRep, options, nil)
	if err != nil {
		t.Fatalf("score high repetition: %v", err)
	}
	lowScore, _ := scoreAndEvidenceForSkill(t, lowScores, "sentence economy")
	highScore, highEvidence := scoreAndEvidenceForSkill(t, highScores, "sentence economy")
	if highScore >= lowScore {
		t.Fatalf("expected high repetition to reduce sentence economy: low=%d high=%d", lowScore, highScore)
	}
	if !containsRule(highEvidence.AppliedRules, "semantic repetition") {
		t.Fatalf("expected semantic repetition rule trace, got %+v", highEvidence.AppliedRules)
	}
}

func TestThoughtLeadershipSignpostingPenalizesTopicDrift(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	options := analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership"}
	sub := domain.Submission{ID: 503, Content: "fixture", WordCount: 820}
	baseMetrics := map[string]int{
		"word_count":                    820,
		"avg_sentence_length":           14,
		"paragraph_count":               9,
		"nlp_readability_grade":         8,
		"nlp_passive_sentences":         0,
		"nlp_unique_token_ratio":        59,
		"nlp_semantic_repetition_ratio": 24,
	}
	cohesive := analyzer.Report{Metrics: cloneMetrics(baseMetrics)}
	cohesive.Metrics["nlp_topic_drift_score"] = 35
	drifty := analyzer.Report{Metrics: cloneMetrics(baseMetrics)}
	drifty.Metrics["nlp_topic_drift_score"] = 78

	cohesiveScores, err := engine.ScoreSubmission(sub, cohesive, options, nil)
	if err != nil {
		t.Fatalf("score cohesive: %v", err)
	}
	driftyScores, err := engine.ScoreSubmission(sub, drifty, options, nil)
	if err != nil {
		t.Fatalf("score drifty: %v", err)
	}
	cohesiveScore, _ := scoreAndEvidenceForSkill(t, cohesiveScores, "structural signposting")
	driftyScore, driftyEvidence := scoreAndEvidenceForSkill(t, driftyScores, "structural signposting")
	if driftyScore >= cohesiveScore {
		t.Fatalf("expected topic drift to reduce structural signposting: cohesive=%d drifty=%d", cohesiveScore, driftyScore)
	}
	if !containsRule(driftyEvidence.AppliedRules, "topic drift") {
		t.Fatalf("expected topic drift rule trace, got %+v", driftyEvidence.AppliedRules)
	}
}

func TestTopScoreGateFailureIgnoresMissingOptionalMetrics(t *testing.T) {
	gate := TopScoreGate{
		MinMetrics: map[string]int{
			"nlp_claim_evidence_coverage": 60,
		},
		MaxMetrics: map[string]int{
			"nlp_topic_drift_score": 45,
		},
	}

	evidence := &ScoreEvidence{MetricSnapshot: map[string]int{}}
	reportMissing := analyzer.Report{
		Metrics: map[string]int{
			"word_count": 700,
		},
	}
	if reason := topScoreGateFailure(gate, reportMissing, map[string]int{}, evidence); reason != "" {
		t.Fatalf("expected missing optional gate metrics to pass, got reason %q", reason)
	}

	reportMinViolation := analyzer.Report{
		Metrics: map[string]int{
			"nlp_claim_evidence_coverage": 20,
		},
	}
	if reason := topScoreGateFailure(gate, reportMinViolation, map[string]int{}, evidence); reason == "" {
		t.Fatal("expected min-metric violation to fail top score gate")
	}

	reportMaxViolation := analyzer.Report{
		Metrics: map[string]int{
			"nlp_topic_drift_score": 90,
		},
	}
	if reason := topScoreGateFailure(gate, reportMaxViolation, map[string]int{}, evidence); reason == "" {
		t.Fatal("expected max-metric violation to fail top score gate")
	}
}

func cloneMetrics(input map[string]int) map[string]int {
	out := make(map[string]int, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func scoreAndEvidenceForSkill(t *testing.T, scores []domain.SkillScore, skill string) (int, evidenceEnvelope) {
	t.Helper()
	for _, score := range scores {
		if score.Skill != skill {
			continue
		}
		var evidence evidenceEnvelope
		if err := json.Unmarshal([]byte(score.ScoreEvidenceJSON), &evidence); err != nil {
			t.Fatalf("decode evidence for %s: %v", skill, err)
		}
		return score.Score, evidence
	}
	t.Fatalf("missing score for skill %q", skill)
	return 0, evidenceEnvelope{}
}

func containsRule(rules []string, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	for _, rule := range rules {
		if strings.Contains(strings.ToLower(rule), needle) {
			return true
		}
	}
	return false
}
