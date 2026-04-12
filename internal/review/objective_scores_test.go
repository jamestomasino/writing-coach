package review

import (
	"encoding/json"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

func TestBuildObjectiveScoresCoversAllActiveTGOs(t *testing.T) {
	active := []domain.TGO{
		{Code: "story-causal-clarity"},
		{Code: "story-scene-architecture"},
		{Code: "story-prose-precision"},
	}
	assessments := []domain.TGOAssessment{
		{TGOCode: "story-causal-clarity", Status: "secure"},
		{TGOCode: "story-scene-architecture", Status: "developing"},
		{TGOCode: "story-prose-precision", Status: "mastered"},
	}
	scores := []domain.SkillScore{
		{SubmissionID: 10, Skill: "narrative clarity", Score: 4, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
		{SubmissionID: 10, Skill: "scene architecture", Score: 3, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
		{SubmissionID: 10, Skill: "prose precision", Score: 5, ScoreSource: "deterministic", ScoreVersion: "det-v1"},
	}
	out := BuildObjectiveScores(10, active, assessments, scores, analyzer.Report{}, analyzer.ContextOptions{})
	if len(out) != 3 {
		t.Fatalf("expected 3 objective scores, got %d", len(out))
	}
	for _, score := range out {
		if score.ScoreSource != "deterministic" {
			t.Fatalf("unexpected source: %+v", score)
		}
		if score.Score < 1 || score.Score > 5 {
			t.Fatalf("score out of range: %+v", score)
		}
		if score.ScoreEvidenceJSON == "" || score.ScoreEvidenceJSON == "{}" {
			t.Fatalf("missing evidence for %+v", score)
		}
		var evidence map[string]any
		if err := json.Unmarshal([]byte(score.ScoreEvidenceJSON), &evidence); err != nil {
			t.Fatalf("decode evidence: %v", err)
		}
		ruleIDs, ok := evidence["objective_rule_ids"].([]any)
		if !ok || len(ruleIDs) == 0 {
			t.Fatalf("missing objective_rule_ids in %+v", evidence)
		}
		summary, _ := evidence["trigger_summary"].(string)
		if summary == "" {
			t.Fatalf("missing trigger_summary in %+v", evidence)
		}
	}
}

func TestBuildObjectiveScoresFallsBackToAssessmentStatus(t *testing.T) {
	active := []domain.TGO{{Code: "memoir-causal-thread"}}
	assessments := []domain.TGOAssessment{{TGOCode: "memoir-causal-thread", Status: "mastered"}}
	out := BuildObjectiveScores(21, active, assessments, nil, analyzer.Report{}, analyzer.ContextOptions{})
	if len(out) != 1 {
		t.Fatalf("expected single score, got %d", len(out))
	}
	if out[0].Score != 5 {
		t.Fatalf("expected mastered fallback score 5, got %d", out[0].Score)
	}
}

func TestBuildObjectiveScoresAcademicPairwiseDiscrimination(t *testing.T) {
	active := []domain.TGO{
		{Code: "academic-active-voice"},
		{Code: "academic-hedging-control"},
	}
	assessments := []domain.TGOAssessment{
		{TGOCode: "academic-active-voice", Status: "developing"},
		{TGOCode: "academic-hedging-control", Status: "developing"},
	}
	scores := []domain.SkillScore{
		{
			SubmissionID:      41,
			Skill:             "clarity and coherence",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-academic"}`,
		},
	}
	options := analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"}

	activeVoiceFavored := analyzer.Report{
		Metrics: map[string]int{
			"nlp_readability_grade":         10,
			"nlp_semantic_repetition_ratio": 40,
			"nlp_passive_sentences":         0,
			"nlp_modifier_overload_ratio":   20,
		},
	}
	outA := BuildObjectiveScores(41, active, assessments, scores, activeVoiceFavored, options)
	byCodeA := objectiveScoreByCode(outA)
	if byCodeA["academic-active-voice"].Score <= byCodeA["academic-hedging-control"].Score {
		t.Fatalf(
			"expected active-voice > hedging under passive=0 modifier=20, got active=%d hedging=%d",
			byCodeA["academic-active-voice"].Score,
			byCodeA["academic-hedging-control"].Score,
		)
	}

	hedgingFavored := analyzer.Report{
		Metrics: map[string]int{
			"nlp_readability_grade":         10,
			"nlp_semantic_repetition_ratio": 40,
			"nlp_passive_sentences":         5,
			"nlp_modifier_overload_ratio":   8,
		},
	}
	outB := BuildObjectiveScores(41, active, assessments, scores, hedgingFavored, options)
	byCodeB := objectiveScoreByCode(outB)
	if byCodeB["academic-hedging-control"].Score <= byCodeB["academic-active-voice"].Score {
		t.Fatalf(
			"expected hedging > active-voice under passive=5 modifier=8, got hedging=%d active=%d",
			byCodeB["academic-hedging-control"].Score,
			byCodeB["academic-active-voice"].Score,
		)
	}
}

func TestBuildObjectiveScoresAcademicPairwiseDiscriminationMultiplePairs(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"}
	sharedSkill := []domain.SkillScore{
		{
			SubmissionID:      52,
			Skill:             "source handling",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-academic-source"}`,
		},
		{
			SubmissionID:      52,
			Skill:             "structural signposting",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-academic-structure"}`,
		},
		{
			SubmissionID:      52,
			Skill:             "evidence integration",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-academic-evidence"}`,
		},
	}

	cases := []struct {
		name      string
		leftCode  string
		rightCode string
		leftWins  analyzer.Report
		rightWins analyzer.Report
	}{
		{
			name:      "quote integration vs source selection",
			leftCode:  "academic-quote-integration",
			rightCode: "academic-source-selection",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_evidence_marker_count":   4,
				"nlp_unique_token_ratio":      38,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_evidence_marker_count":   0,
				"nlp_unique_token_ratio":      60,
			}},
		},
		{
			name:      "evidence basics vs evidence proportion",
			leftCode:  "academic-evidence-basics",
			rightCode: "academic-evidence-proportion",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_evidence_marker_count":   4,
				"nlp_claim_support_alignment": 40,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_evidence_marker_count":   0,
				"nlp_claim_support_alignment": 78,
			}},
		},
		{
			name:      "introduction function vs abstract writing",
			leftCode:  "academic-introduction-function",
			rightCode: "academic-abstract-writing",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_structural_signpost_count": 6,
				"nlp_topic_drift_score":         35,
				"paragraph_count":               8,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_structural_signpost_count": 1,
				"nlp_topic_drift_score":         35,
				"paragraph_count":               2,
			}},
		},
		{
			name:      "section order vs transitions",
			leftCode:  "academic-section-order",
			rightCode: "academic-transitions",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_topic_drift_score":         35,
				"nlp_transition_marker_density": 1,
				"nlp_structural_signpost_count": 4,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_topic_drift_score":         80,
				"nlp_transition_marker_density": 7,
				"nlp_structural_signpost_count": 4,
			}},
		},
		{
			name:      "source evaluation vs citation discipline",
			leftCode:  "academic-source-evaluation",
			rightCode: "academic-citation-discipline",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_evidence_coverage":     60,
				"nlp_claim_support_alignment":     74,
				"nlp_reference_specificity_score": 40,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_evidence_coverage":     60,
				"nlp_claim_support_alignment":     40,
				"nlp_reference_specificity_score": 74,
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.leftCode}, {Code: tc.rightCode}}
			assessments := []domain.TGOAssessment{
				{TGOCode: tc.leftCode, Status: "developing"},
				{TGOCode: tc.rightCode, Status: "developing"},
			}

			leftPass := BuildObjectiveScores(52, active, assessments, sharedSkill, tc.leftWins, options)
			leftMap := objectiveScoreByCode(leftPass)
			if leftMap[tc.leftCode].Score <= leftMap[tc.rightCode].Score {
				t.Fatalf("expected %s score > %s score in leftWins scenario, got %d <= %d", tc.leftCode, tc.rightCode, leftMap[tc.leftCode].Score, leftMap[tc.rightCode].Score)
			}

			rightPass := BuildObjectiveScores(52, active, assessments, sharedSkill, tc.rightWins, options)
			rightMap := objectiveScoreByCode(rightPass)
			if rightMap[tc.rightCode].Score <= rightMap[tc.leftCode].Score {
				t.Fatalf("expected %s score > %s score in rightWins scenario, got %d <= %d", tc.rightCode, tc.leftCode, rightMap[tc.rightCode].Score, rightMap[tc.leftCode].Score)
			}
		})
	}
}

func TestBuildObjectiveScoresAcademicMetamorphicMonotonicity(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "academic-essay-track", WritingType: "academic essay"}

	type metricCase struct {
		name        string
		code        string
		skill       string
		lowMetrics  map[string]int
		highMetrics map[string]int
	}

	cases := []metricCase{
		{
			name:  "quote integration improves with evidence markers",
			code:  "academic-quote-integration",
			skill: "source handling",
			lowMetrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_evidence_marker_count":   0,
			},
			highMetrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_evidence_marker_count":   4,
			},
		},
		{
			name:  "source selection improves with lexical variety",
			code:  "academic-source-selection",
			skill: "source handling",
			lowMetrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_unique_token_ratio":      38,
			},
			highMetrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_unique_token_ratio":      60,
			},
		},
		{
			name:  "evidence proportion improves with claim-support alignment",
			code:  "academic-evidence-proportion",
			skill: "evidence integration",
			lowMetrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_claim_support_alignment": 40,
			},
			highMetrics: map[string]int{
				"nlp_claim_evidence_coverage": 60,
				"nlp_claim_support_alignment": 78,
			},
		},
		{
			name:  "transitions improve with transition density",
			code:  "academic-transitions",
			skill: "structural signposting",
			lowMetrics: map[string]int{
				"nlp_topic_drift_score":         35,
				"nlp_transition_marker_density": 1,
			},
			highMetrics: map[string]int{
				"nlp_topic_drift_score":         35,
				"nlp_transition_marker_density": 7,
			},
		},
		{
			name:  "active voice improves as passive usage drops",
			code:  "academic-active-voice",
			skill: "clarity and coherence",
			lowMetrics: map[string]int{
				"nlp_readability_grade":       10,
				"nlp_passive_sentences":       5,
				"nlp_modifier_overload_ratio": 12,
			},
			highMetrics: map[string]int{
				"nlp_readability_grade":       10,
				"nlp_passive_sentences":       0,
				"nlp_modifier_overload_ratio": 12,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.code}}
			assessments := []domain.TGOAssessment{{TGOCode: tc.code, Status: "developing"}}
			scores := []domain.SkillScore{
				{
					SubmissionID:      67,
					Skill:             tc.skill,
					Score:             3,
					ScoreSource:       "deterministic",
					ScoreVersion:      "det-v1",
					ScoreEvidenceJSON: `{"rubric_id":"metamorphic-fixture"}`,
				},
			}

			low := BuildObjectiveScores(
				67,
				active,
				assessments,
				scores,
				analyzer.Report{Metrics: tc.lowMetrics},
				options,
			)
			high := BuildObjectiveScores(
				67,
				active,
				assessments,
				scores,
				analyzer.Report{Metrics: tc.highMetrics},
				options,
			)
			lowScore := objectiveScoreByCode(low)[tc.code].Score
			highScore := objectiveScoreByCode(high)[tc.code].Score

			if highScore < lowScore {
				t.Fatalf(
					"metamorphic monotonicity violated for %s: improved evidence lowered score (%d -> %d)",
					tc.code,
					lowScore,
					highScore,
				)
			}
		})
	}
}

func TestBuildObjectiveScoresTechnicalPairwiseDiscrimination(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "technical-writing-track", WritingType: "technical writing"}
	shared := []domain.SkillScore{
		{
			SubmissionID:      88,
			Skill:             "actionability",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-technical-actionability"}`,
		},
		{
			SubmissionID:      88,
			Skill:             "technical precision",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-technical-precision"}`,
		},
		{
			SubmissionID:      88,
			Skill:             "scannability",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-technical-scan"}`,
		},
		{
			SubmissionID:      88,
			Skill:             "accuracy",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-technical-accuracy"}`,
		},
	}

	cases := []struct {
		name      string
		leftCode  string
		rightCode string
		leftWins  analyzer.Report
		rightWins analyzer.Report
	}{
		{
			name:      "step clarity vs prereq clarity",
			leftCode:  "technical-step-clarity",
			rightCode: "technical-prereq-clarity",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_action_verb_density":         12,
				"nlp_reference_specificity_score": 40,
				"nlp_topic_drift_score":           35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_action_verb_density":         2,
				"nlp_reference_specificity_score": 76,
				"nlp_topic_drift_score":           35,
			}},
		},
		{
			name:      "active voice vs sentence economy",
			leftCode:  "technical-active-voice",
			rightCode: "technical-sentence-economy",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_passive_sentences":         0,
				"nlp_semantic_repetition_ratio": 70,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_passive_sentences":         5,
				"nlp_semantic_repetition_ratio": 22,
			}},
		},
		{
			name:      "opening summary vs reference layout",
			leftCode:  "technical-opening-summary",
			rightCode: "technical-reference-layout",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_structural_signpost_count": 6,
				"nlp_scannability_marker_count": 2,
				"nlp_topic_drift_score":         35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_structural_signpost_count": 1,
				"nlp_scannability_marker_count": 9,
				"nlp_topic_drift_score":         35,
			}},
		},
		{
			name:      "troubleshooting vs versioning",
			leftCode:  "technical-troubleshooting-basics",
			rightCode: "technical-versioning-signals",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_support_alignment":     76,
				"nlp_temporal_clarity_score":      40,
				"nlp_reference_specificity_score": 60,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_support_alignment":     40,
				"nlp_temporal_clarity_score":      75,
				"nlp_reference_specificity_score": 60,
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.leftCode}, {Code: tc.rightCode}}
			assessments := []domain.TGOAssessment{
				{TGOCode: tc.leftCode, Status: "developing"},
				{TGOCode: tc.rightCode, Status: "developing"},
			}

			leftPass := BuildObjectiveScores(88, active, assessments, shared, tc.leftWins, options)
			leftMap := objectiveScoreByCode(leftPass)
			if leftMap[tc.leftCode].Score <= leftMap[tc.rightCode].Score {
				t.Fatalf("expected %s score > %s score in leftWins scenario, got %d <= %d", tc.leftCode, tc.rightCode, leftMap[tc.leftCode].Score, leftMap[tc.rightCode].Score)
			}

			rightPass := BuildObjectiveScores(88, active, assessments, shared, tc.rightWins, options)
			rightMap := objectiveScoreByCode(rightPass)
			if rightMap[tc.rightCode].Score <= rightMap[tc.leftCode].Score {
				t.Fatalf("expected %s score > %s score in rightWins scenario, got %d <= %d", tc.rightCode, tc.leftCode, rightMap[tc.rightCode].Score, rightMap[tc.leftCode].Score)
			}
		})
	}
}

func TestBuildObjectiveScoresTechnicalMetamorphicMonotonicity(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "technical-writing-track", WritingType: "technical writing"}
	cases := []struct {
		name        string
		code        string
		skill       string
		lowMetrics  map[string]int
		highMetrics map[string]int
	}{
		{
			name:  "step clarity improves with action density",
			code:  "technical-step-clarity",
			skill: "actionability",
			lowMetrics: map[string]int{
				"nlp_action_verb_density":         2,
				"nlp_reference_specificity_score": 60,
			},
			highMetrics: map[string]int{
				"nlp_action_verb_density":         12,
				"nlp_reference_specificity_score": 60,
			},
		},
		{
			name:  "sentence economy improves with lower repetition",
			code:  "technical-sentence-economy",
			skill: "technical precision",
			lowMetrics: map[string]int{
				"nlp_readability_grade":         10,
				"nlp_semantic_repetition_ratio": 72,
			},
			highMetrics: map[string]int{
				"nlp_readability_grade":         10,
				"nlp_semantic_repetition_ratio": 22,
			},
		},
		{
			name:  "reference layout improves with scan markers",
			code:  "technical-reference-layout",
			skill: "scannability",
			lowMetrics: map[string]int{
				"nlp_scannability_marker_count": 2,
				"nlp_structural_signpost_count": 4,
			},
			highMetrics: map[string]int{
				"nlp_scannability_marker_count": 9,
				"nlp_structural_signpost_count": 4,
			},
		},
		{
			name:  "versioning improves with temporal clarity",
			code:  "technical-versioning-signals",
			skill: "accuracy",
			lowMetrics: map[string]int{
				"nlp_temporal_clarity_score":  40,
				"nlp_claim_support_alignment": 60,
			},
			highMetrics: map[string]int{
				"nlp_temporal_clarity_score":  76,
				"nlp_claim_support_alignment": 60,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.code}}
			assessments := []domain.TGOAssessment{{TGOCode: tc.code, Status: "developing"}}
			scores := []domain.SkillScore{
				{
					SubmissionID:      89,
					Skill:             tc.skill,
					Score:             3,
					ScoreSource:       "deterministic",
					ScoreVersion:      "det-v1",
					ScoreEvidenceJSON: `{"rubric_id":"metamorphic-technical-fixture"}`,
				},
			}

			low := BuildObjectiveScores(89, active, assessments, scores, analyzer.Report{Metrics: tc.lowMetrics}, options)
			high := BuildObjectiveScores(89, active, assessments, scores, analyzer.Report{Metrics: tc.highMetrics}, options)
			lowScore := objectiveScoreByCode(low)[tc.code].Score
			highScore := objectiveScoreByCode(high)[tc.code].Score
			if highScore < lowScore {
				t.Fatalf("metamorphic monotonicity violated for %s: %d -> %d", tc.code, lowScore, highScore)
			}
		})
	}
}

func TestBuildObjectiveScoresProfessionalPairwiseDiscrimination(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "professional-writing-track", WritingType: "professional writing"}
	shared := []domain.SkillScore{
		{
			SubmissionID:      90,
			Skill:             "actionability",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-professional-actionability"}`,
		},
		{
			SubmissionID:      90,
			Skill:             "sentence economy",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-professional-sentence-economy"}`,
		},
		{
			SubmissionID:      90,
			Skill:             "scannability",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-professional-scannability"}`,
		},
		{
			SubmissionID:      90,
			Skill:             "evidence integration",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-professional-evidence"}`,
		},
	}

	cases := []struct {
		name      string
		leftCode  string
		rightCode string
		leftWins  analyzer.Report
		rightWins analyzer.Report
	}{
		{
			name:      "ask visibility vs ownership clarity",
			leftCode:  "ask-visibility",
			rightCode: "ownership-clarity",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_action_verb_density":         12,
				"nlp_reference_specificity_score": 50,
				"nlp_topic_drift_score":           35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_action_verb_density":         2,
				"nlp_reference_specificity_score": 78,
				"nlp_topic_drift_score":           35,
			}},
		},
		{
			name:      "active voice vs sentence economy",
			leftCode:  "active-voice-control",
			rightCode: "professional-sentence-economy",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_passive_sentences":         0,
				"nlp_semantic_repetition_ratio": 70,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_passive_sentences":         5,
				"nlp_semantic_repetition_ratio": 22,
			}},
		},
		{
			name:      "front loaded summary vs heading discipline",
			leftCode:  "front-loaded-summary",
			rightCode: "heading-discipline",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_structural_signpost_count": 6,
				"nlp_scannability_marker_count": 2,
				"nlp_topic_drift_score":         35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_structural_signpost_count": 1,
				"nlp_scannability_marker_count": 9,
				"nlp_topic_drift_score":         35,
			}},
		},
		{
			name:      "supporting rationale vs quantification basics",
			leftCode:  "supporting-rationale",
			rightCode: "quantification-basics",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_support_alignment": 78,
				"nlp_evidence_marker_count":   2,
				"nlp_topic_drift_score":       35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_support_alignment": 60,
				"nlp_evidence_marker_count":   5,
				"nlp_topic_drift_score":       35,
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.leftCode}, {Code: tc.rightCode}}
			assessments := []domain.TGOAssessment{
				{TGOCode: tc.leftCode, Status: "developing"},
				{TGOCode: tc.rightCode, Status: "developing"},
			}

			leftPass := BuildObjectiveScores(90, active, assessments, shared, tc.leftWins, options)
			leftMap := objectiveScoreByCode(leftPass)
			if leftMap[tc.leftCode].Score <= leftMap[tc.rightCode].Score {
				t.Fatalf("expected %s score > %s score in leftWins scenario, got %d <= %d", tc.leftCode, tc.rightCode, leftMap[tc.leftCode].Score, leftMap[tc.rightCode].Score)
			}

			rightPass := BuildObjectiveScores(90, active, assessments, shared, tc.rightWins, options)
			rightMap := objectiveScoreByCode(rightPass)
			if rightMap[tc.rightCode].Score <= rightMap[tc.leftCode].Score {
				t.Fatalf("expected %s score > %s score in rightWins scenario, got %d <= %d", tc.rightCode, tc.leftCode, rightMap[tc.rightCode].Score, rightMap[tc.leftCode].Score)
			}
		})
	}
}

func TestBuildObjectiveScoresProfessionalMetamorphicMonotonicity(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "professional-writing-track", WritingType: "professional writing"}
	cases := []struct {
		name        string
		code        string
		skill       string
		lowMetrics  map[string]int
		highMetrics map[string]int
	}{
		{
			name:  "ask visibility improves with action density",
			code:  "ask-visibility",
			skill: "actionability",
			lowMetrics: map[string]int{
				"nlp_action_verb_density":         2,
				"nlp_reference_specificity_score": 60,
			},
			highMetrics: map[string]int{
				"nlp_action_verb_density":         12,
				"nlp_reference_specificity_score": 60,
			},
		},
		{
			name:  "sentence economy improves with lower repetition",
			code:  "professional-sentence-economy",
			skill: "sentence economy",
			lowMetrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_semantic_repetition_ratio": 72,
			},
			highMetrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_semantic_repetition_ratio": 22,
			},
		},
		{
			name:  "heading discipline improves with markers",
			code:  "heading-discipline",
			skill: "scannability",
			lowMetrics: map[string]int{
				"nlp_scannability_marker_count": 2,
				"nlp_structural_signpost_count": 4,
			},
			highMetrics: map[string]int{
				"nlp_scannability_marker_count": 9,
				"nlp_structural_signpost_count": 4,
			},
		},
		{
			name:  "quantification basics improves with evidence markers",
			code:  "quantification-basics",
			skill: "evidence integration",
			lowMetrics: map[string]int{
				"nlp_evidence_marker_count":   1,
				"nlp_claim_support_alignment": 60,
			},
			highMetrics: map[string]int{
				"nlp_evidence_marker_count":   5,
				"nlp_claim_support_alignment": 60,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.code}}
			assessments := []domain.TGOAssessment{{TGOCode: tc.code, Status: "developing"}}
			scores := []domain.SkillScore{
				{
					SubmissionID:      91,
					Skill:             tc.skill,
					Score:             3,
					ScoreSource:       "deterministic",
					ScoreVersion:      "det-v1",
					ScoreEvidenceJSON: `{"rubric_id":"metamorphic-professional-fixture"}`,
				},
			}

			low := BuildObjectiveScores(91, active, assessments, scores, analyzer.Report{Metrics: tc.lowMetrics}, options)
			high := BuildObjectiveScores(91, active, assessments, scores, analyzer.Report{Metrics: tc.highMetrics}, options)
			lowScore := objectiveScoreByCode(low)[tc.code].Score
			highScore := objectiveScoreByCode(high)[tc.code].Score
			if highScore < lowScore {
				t.Fatalf("metamorphic monotonicity violated for %s: %d -> %d", tc.code, lowScore, highScore)
			}
		})
	}
}

func TestBuildObjectiveScoresThoughtLeadershipPairwiseDiscrimination(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership"}
	shared := []domain.SkillScore{
		{
			SubmissionID:      92,
			Skill:             "claim clarity",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-thought-claim-clarity"}`,
		},
		{
			SubmissionID:      92,
			Skill:             "sentence economy",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-thought-sentence-economy"}`,
		},
		{
			SubmissionID:      92,
			Skill:             "structural signposting",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-thought-structure"}`,
		},
		{
			SubmissionID:      92,
			Skill:             "evidence integration",
			Score:             3,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"fixture-thought-evidence"}`,
		},
	}

	cases := []struct {
		name      string
		leftCode  string
		rightCode string
		leftWins  analyzer.Report
		rightWins analyzer.Report
	}{
		{
			name:      "claim clarity vs stakes articulation",
			leftCode:  "claim-clarity",
			rightCode: "stakes-articulation",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_count":             4,
				"nlp_claim_support_alignment": 45,
				"nlp_topic_drift_score":       35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_claim_count":             1,
				"nlp_claim_support_alignment": 78,
				"nlp_topic_drift_score":       35,
			}},
		},
		{
			name:      "sentence pressure vs verb forward style",
			leftCode:  "sentence-pressure",
			rightCode: "verb-forward-style",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_semantic_repetition_ratio": 22,
				"nlp_action_verb_density":       3,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_readability_grade":         9,
				"nlp_semantic_repetition_ratio": 62,
				"nlp_action_verb_density":       12,
			}},
		},
		{
			name:      "bridge sentences vs section ordering",
			leftCode:  "bridge-sentences",
			rightCode: "section-ordering",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_transition_marker_density": 8,
				"nlp_structural_signpost_count": 3,
				"nlp_topic_drift_score":         35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_transition_marker_density": 2,
				"nlp_structural_signpost_count": 6,
				"nlp_topic_drift_score":         35,
			}},
		},
		{
			name:      "example selection vs quote restraint",
			leftCode:  "example-selection",
			rightCode: "quote-restraint",
			leftWins: analyzer.Report{Metrics: map[string]int{
				"nlp_reference_specificity_score": 78,
				"nlp_evidence_marker_count":       2,
				"nlp_claim_support_alignment":     55,
				"nlp_topic_drift_score":           35,
			}},
			rightWins: analyzer.Report{Metrics: map[string]int{
				"nlp_reference_specificity_score": 50,
				"nlp_evidence_marker_count":       1,
				"nlp_claim_support_alignment":     65,
				"nlp_topic_drift_score":           35,
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.leftCode}, {Code: tc.rightCode}}
			assessments := []domain.TGOAssessment{
				{TGOCode: tc.leftCode, Status: "developing"},
				{TGOCode: tc.rightCode, Status: "developing"},
			}

			leftPass := BuildObjectiveScores(92, active, assessments, shared, tc.leftWins, options)
			leftMap := objectiveScoreByCode(leftPass)
			if leftMap[tc.leftCode].Score <= leftMap[tc.rightCode].Score {
				t.Fatalf("expected %s score > %s score in leftWins scenario, got %d <= %d", tc.leftCode, tc.rightCode, leftMap[tc.leftCode].Score, leftMap[tc.rightCode].Score)
			}

			rightPass := BuildObjectiveScores(92, active, assessments, shared, tc.rightWins, options)
			rightMap := objectiveScoreByCode(rightPass)
			if rightMap[tc.rightCode].Score <= rightMap[tc.leftCode].Score {
				t.Fatalf("expected %s score > %s score in rightWins scenario, got %d <= %d", tc.rightCode, tc.leftCode, rightMap[tc.rightCode].Score, rightMap[tc.leftCode].Score)
			}
		})
	}
}

func TestBuildObjectiveScoresThoughtLeadershipMetamorphicMonotonicity(t *testing.T) {
	options := analyzer.ContextOptions{TreeSlug: "thought-leadership-track", WritingType: "thought leadership"}
	cases := []struct {
		name        string
		code        string
		skill       string
		lowMetrics  map[string]int
		highMetrics map[string]int
	}{
		{
			name:  "claim clarity improves with claim count",
			code:  "claim-clarity",
			skill: "claim clarity",
			lowMetrics: map[string]int{
				"nlp_claim_count":             1,
				"nlp_claim_support_alignment": 62,
			},
			highMetrics: map[string]int{
				"nlp_claim_count":             4,
				"nlp_claim_support_alignment": 62,
			},
		},
		{
			name:  "verb-forward style improves with action density",
			code:  "verb-forward-style",
			skill: "sentence economy",
			lowMetrics: map[string]int{
				"nlp_readability_grade":   9,
				"nlp_action_verb_density": 3,
			},
			highMetrics: map[string]int{
				"nlp_readability_grade":   9,
				"nlp_action_verb_density": 12,
			},
		},
		{
			name:  "bridge sentences improves with transition density",
			code:  "bridge-sentences",
			skill: "structural signposting",
			lowMetrics: map[string]int{
				"nlp_transition_marker_density": 2,
				"nlp_structural_signpost_count": 4,
			},
			highMetrics: map[string]int{
				"nlp_transition_marker_density": 8,
				"nlp_structural_signpost_count": 4,
			},
		},
		{
			name:  "example selection improves with reference specificity",
			code:  "example-selection",
			skill: "evidence integration",
			lowMetrics: map[string]int{
				"nlp_reference_specificity_score": 45,
				"nlp_claim_support_alignment":     65,
			},
			highMetrics: map[string]int{
				"nlp_reference_specificity_score": 78,
				"nlp_claim_support_alignment":     65,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := []domain.TGO{{Code: tc.code}}
			assessments := []domain.TGOAssessment{{TGOCode: tc.code, Status: "developing"}}
			scores := []domain.SkillScore{
				{
					SubmissionID:      93,
					Skill:             tc.skill,
					Score:             3,
					ScoreSource:       "deterministic",
					ScoreVersion:      "det-v1",
					ScoreEvidenceJSON: `{"rubric_id":"metamorphic-thought-fixture"}`,
				},
			}

			low := BuildObjectiveScores(93, active, assessments, scores, analyzer.Report{Metrics: tc.lowMetrics}, options)
			high := BuildObjectiveScores(93, active, assessments, scores, analyzer.Report{Metrics: tc.highMetrics}, options)
			lowScore := objectiveScoreByCode(low)[tc.code].Score
			highScore := objectiveScoreByCode(high)[tc.code].Score
			if highScore < lowScore {
				t.Fatalf("metamorphic monotonicity violated for %s: %d -> %d", tc.code, lowScore, highScore)
			}
		})
	}
}

func objectiveScoreByCode(scores []domain.ObjectiveScore) map[string]domain.ObjectiveScore {
	out := map[string]domain.ObjectiveScore{}
	for _, score := range scores {
		out[score.TGOCode] = score
	}
	return out
}
