package scoring

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestPublicTGOsHaveDeterministicScoreCoverage(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	report := analyzer.Report{
		Metrics: map[string]int{
			"word_count":                         640,
			"avg_sentence_length":                14,
			"paragraph_count":                    8,
			"nlp_readability_grade":              8,
			"nlp_passive_sentences":              1,
			"nlp_unique_token_ratio":             58,
			"nlp_claim_count":                    4,
			"nlp_evidence_marker_count":          3,
			"nlp_claim_evidence_coverage":        75,
			"nlp_coref_ambiguity_count":          0,
			"nlp_semantic_repetition_ratio":      16,
			"nlp_topic_drift_score":              18,
			"nlp_action_verb_density":            12,
			"nlp_transition_marker_density":      7,
			"nlp_audience_reference_count":       3,
			"nlp_paragraph_length_variation":     14,
			"nlp_sentence_variety_index":         42,
			"nlp_concrete_noun_ratio":            26,
			"nlp_dialogue_ratio":                 12,
			"nlp_world_specific_term_count":      6,
			"nlp_goal_conflict_marker_count":     4,
			"nlp_structural_signpost_count":      5,
			"nlp_insight_density_ratio":          22,
			"nlp_imperative_verb_count":          4,
			"nlp_heading_count":                  3,
			"nlp_scannability_marker_count":      9,
			"nlp_claim_support_alignment":        70,
			"nlp_redundancy_ratio":               8,
			"nlp_paragraph_focus_score":          68,
			"nlp_sentence_boundary_error_count":  0,
			"nlp_lexical_density":                49,
			"nlp_modifier_overload_ratio":        9,
			"nlp_discourse_coherence_score":      74,
			"nlp_reference_specificity_score":    71,
			"nlp_temporal_clarity_score":         76,
			"nlp_character_distinctiveness_score": 64,
		},
		Findings: []analyzer.Finding{
			{Category: "clarity"},
		},
	}

	submissionID := int64(10_000)
	for _, tree := range domain.PublicBuiltInTrees {
		for _, tgo := range tree.TGOs {
			submissionID++
			mappedSkill := strings.TrimSpace(domain.TGOCodeToSkill[tgo.Code])
			if mappedSkill == "" {
				t.Fatalf("missing skill mapping for tgo=%s tree=%s", tgo.Code, tree.Slug)
			}

			scores, scoreErr := engine.ScoreSubmission(
				domain.Submission{ID: submissionID, Content: "deterministic fixture", WordCount: 640},
				report,
				analyzer.ContextOptions{TreeSlug: tree.Slug, WritingType: tree.Title},
				[]domain.TGO{{Code: tgo.Code}},
			)
			if scoreErr != nil {
				t.Fatalf("score tgo=%s tree=%s: %v", tgo.Code, tree.Slug, scoreErr)
			}

			score, ok := deterministicScoreForSkill(scores, mappedSkill)
			if !ok {
				t.Fatalf("no deterministic score for mapped skill=%q tgo=%s tree=%s", mappedSkill, tgo.Code, tree.Slug)
			}
			if strings.TrimSpace(score.ScoreEvidenceJSON) == "" || strings.TrimSpace(score.ScoreEvidenceJSON) == "{}" {
				t.Fatalf("missing deterministic evidence for tgo=%s skill=%q tree=%s", tgo.Code, mappedSkill, tree.Slug)
			}
			var evidence map[string]any
			if err := json.Unmarshal([]byte(score.ScoreEvidenceJSON), &evidence); err != nil {
				t.Fatalf("decode evidence for tgo=%s skill=%q tree=%s: %v", tgo.Code, mappedSkill, tree.Slug, err)
			}
			rubricID, _ := evidence["rubric_id"].(string)
			if strings.TrimSpace(rubricID) == "" {
				t.Fatalf("missing rubric_id evidence for tgo=%s skill=%q tree=%s", tgo.Code, mappedSkill, tree.Slug)
			}
		}
	}
}

func deterministicScoreForSkill(scores []domain.SkillScore, skill string) (domain.SkillScore, bool) {
	target := strings.ToLower(strings.TrimSpace(skill))
	for _, score := range scores {
		if strings.TrimSpace(score.ScoreSource) != "deterministic" {
			continue
		}
		if strings.ToLower(strings.TrimSpace(score.Skill)) == target {
			return score, true
		}
	}
	return domain.SkillScore{}, false
}
