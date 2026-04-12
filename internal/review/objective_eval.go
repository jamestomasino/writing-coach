package review

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

type ObjectiveEvalCorpus struct {
	Version            string                     `json:"version"`
	MinPassRate        float64                    `json:"min_pass_rate"`
	MaxPairwiseTieRate *float64                   `json:"max_pairwise_tie_rate,omitempty"`
	TrackPolicies      []ObjectiveEvalTrackPolicy `json:"track_policies,omitempty"`
	Cases              []ObjectiveEvalCase        `json:"cases"`
}

type ObjectiveEvalTrackPolicy struct {
	TreeSlug    string   `json:"tree_slug"`
	MinPassRate *float64 `json:"min_pass_rate,omitempty"`
	MinChecks   *int     `json:"min_checks,omitempty"`
}

type ObjectiveEvalCase struct {
	Name             string         `json:"name"`
	Type             string         `json:"type"`
	TreeSlug         string         `json:"tree_slug"`
	WritingType      string         `json:"writing_type"`
	SkillScores      []skillFixture `json:"skill_scores"`
	LeftCode         string         `json:"left_code"`
	RightCode        string         `json:"right_code"`
	LeftWinsMetrics  map[string]int `json:"left_wins_metrics"`
	RightWinsMetrics map[string]int `json:"right_wins_metrics"`
	Code             string         `json:"code"`
	LowMetrics       map[string]int `json:"low_metrics"`
	HighMetrics      map[string]int `json:"high_metrics"`
}

type skillFixture struct {
	Skill string `json:"skill"`
	Score int    `json:"score"`
}

type ObjectiveEvalFailure struct {
	Case   string
	Detail string
}

type ObjectiveEvalTrackAggregate struct {
	Checks int
	Passes int
}

type ObjectiveEvalResult struct {
	CorpusVersion            string
	TotalChecks              int
	PassedChecks             int
	PassRate                 float64
	RequiredMinPassRate      float64
	PairwiseChecks           int
	PairwiseTies             int
	PairwiseTieRate          float64
	MaxPairwiseTieRate       *float64
	TrackAggregates          map[string]ObjectiveEvalTrackAggregate
	Failures                 []ObjectiveEvalFailure
	PolicyFailures           []string
	PassedPolicyRequirements bool
}

func (r ObjectiveEvalResult) FailedChecks() int {
	return r.TotalChecks - r.PassedChecks
}

func (r ObjectiveEvalResult) TrackPassRate(slug string) float64 {
	agg, ok := r.TrackAggregates[strings.TrimSpace(slug)]
	if !ok || agg.Checks == 0 {
		return 0
	}
	return float64(agg.Passes) / float64(agg.Checks)
}

func EvaluateObjectiveScoreCorpusPath(path string) (ObjectiveEvalResult, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ObjectiveEvalResult{}, err
	}
	return EvaluateObjectiveScoreCorpus(raw)
}

func EvaluateObjectiveScoreCorpus(raw []byte) (ObjectiveEvalResult, error) {
	var c ObjectiveEvalCorpus
	if err := json.Unmarshal(raw, &c); err != nil {
		return ObjectiveEvalResult{}, fmt.Errorf("decode corpus: %w", err)
	}
	if len(c.Cases) == 0 {
		return ObjectiveEvalResult{}, fmt.Errorf("corpus has zero cases")
	}
	if c.MinPassRate <= 0 || c.MinPassRate > 1 {
		return ObjectiveEvalResult{}, fmt.Errorf("min_pass_rate must be within (0,1]")
	}
	if c.MaxPairwiseTieRate != nil && (*c.MaxPairwiseTieRate < 0 || *c.MaxPairwiseTieRate > 1) {
		return ObjectiveEvalResult{}, fmt.Errorf("max_pairwise_tie_rate must be within [0,1]")
	}

	result := ObjectiveEvalResult{
		CorpusVersion:       strings.TrimSpace(c.Version),
		RequiredMinPassRate: c.MinPassRate,
		MaxPairwiseTieRate:  c.MaxPairwiseTieRate,
		TrackAggregates:     map[string]ObjectiveEvalTrackAggregate{},
	}

	for idx, tc := range c.Cases {
		caseName := strings.TrimSpace(tc.Name)
		if caseName == "" {
			caseName = fmt.Sprintf("case-%d", idx+1)
		}
		options := analyzer.ContextOptions{TreeSlug: strings.TrimSpace(tc.TreeSlug), WritingType: strings.TrimSpace(tc.WritingType)}
		skillScores := buildSkillScores(tc.SkillScores)
		caseType := strings.ToLower(strings.TrimSpace(tc.Type))
		if caseType == "" {
			caseType = "pairwise"
		}

		switch caseType {
		case "pairwise":
			leftCode := strings.TrimSpace(tc.LeftCode)
			rightCode := strings.TrimSpace(tc.RightCode)
			if leftCode == "" || rightCode == "" {
				result.Failures = append(result.Failures, ObjectiveEvalFailure{Case: caseName, Detail: "pairwise case missing left_code or right_code"})
				continue
			}
			active := []domain.TGO{{Code: leftCode}, {Code: rightCode}}
			assessments := []domain.TGOAssessment{{TGOCode: leftCode, Status: "developing"}, {TGOCode: rightCode, Status: "developing"}}

			leftPass := BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.LeftWinsMetrics}, options)
			leftMap := objectiveEvalScoreByCode(leftPass)
			result.TotalChecks++
			result.PairwiseChecks++
			agg := readTrackAggregate(result.TrackAggregates, options.TreeSlug)
			agg.Checks++
			if leftMap[leftCode].Score == leftMap[rightCode].Score {
				result.PairwiseTies++
			}
			if leftMap[leftCode].Score > leftMap[rightCode].Score {
				result.PassedChecks++
				agg.Passes++
			} else {
				result.Failures = append(result.Failures, ObjectiveEvalFailure{Case: caseName, Detail: fmt.Sprintf("leftWins expected %s>%s got %d<=%d", leftCode, rightCode, leftMap[leftCode].Score, leftMap[rightCode].Score)})
			}
			writeTrackAggregate(result.TrackAggregates, options.TreeSlug, agg)

			rightPass := BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.RightWinsMetrics}, options)
			rightMap := objectiveEvalScoreByCode(rightPass)
			result.TotalChecks++
			result.PairwiseChecks++
			agg = readTrackAggregate(result.TrackAggregates, options.TreeSlug)
			agg.Checks++
			if rightMap[rightCode].Score == rightMap[leftCode].Score {
				result.PairwiseTies++
			}
			if rightMap[rightCode].Score > rightMap[leftCode].Score {
				result.PassedChecks++
				agg.Passes++
			} else {
				result.Failures = append(result.Failures, ObjectiveEvalFailure{Case: caseName, Detail: fmt.Sprintf("rightWins expected %s>%s got %d<=%d", rightCode, leftCode, rightMap[rightCode].Score, rightMap[leftCode].Score)})
			}
			writeTrackAggregate(result.TrackAggregates, options.TreeSlug, agg)
		case "monotonic":
			code := strings.TrimSpace(tc.Code)
			if code == "" {
				result.Failures = append(result.Failures, ObjectiveEvalFailure{Case: caseName, Detail: "monotonic case missing code"})
				continue
			}
			active := []domain.TGO{{Code: code}}
			assessments := []domain.TGOAssessment{{TGOCode: code, Status: "developing"}}
			low := BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.LowMetrics}, options)
			high := BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.HighMetrics}, options)
			lowScore := objectiveEvalScoreByCode(low)[code].Score
			highScore := objectiveEvalScoreByCode(high)[code].Score
			result.TotalChecks++
			agg := readTrackAggregate(result.TrackAggregates, options.TreeSlug)
			agg.Checks++
			if highScore >= lowScore {
				result.PassedChecks++
				agg.Passes++
			} else {
				result.Failures = append(result.Failures, ObjectiveEvalFailure{Case: caseName, Detail: fmt.Sprintf("monotonic expected high>=low got %d<%d", highScore, lowScore)})
			}
			writeTrackAggregate(result.TrackAggregates, options.TreeSlug, agg)
		default:
			result.Failures = append(result.Failures, ObjectiveEvalFailure{Case: caseName, Detail: "unknown case type: " + caseType})
		}
	}

	if result.TotalChecks == 0 {
		return ObjectiveEvalResult{}, fmt.Errorf("corpus produced zero executable checks")
	}
	result.PassRate = float64(result.PassedChecks) / float64(result.TotalChecks)
	if result.PairwiseChecks > 0 {
		result.PairwiseTieRate = float64(result.PairwiseTies) / float64(result.PairwiseChecks)
	}

	if result.PassRate < c.MinPassRate {
		result.PolicyFailures = append(result.PolicyFailures, fmt.Sprintf("pass rate %.3f below required %.3f", result.PassRate, c.MinPassRate))
	}
	if c.MaxPairwiseTieRate != nil && result.PairwiseTieRate > *c.MaxPairwiseTieRate {
		result.PolicyFailures = append(result.PolicyFailures, fmt.Sprintf("pairwise tie rate %.3f above allowed %.3f", result.PairwiseTieRate, *c.MaxPairwiseTieRate))
	}
	for _, policy := range c.TrackPolicies {
		slug := strings.TrimSpace(policy.TreeSlug)
		if slug == "" {
			continue
		}
		agg := readTrackAggregate(result.TrackAggregates, slug)
		if policy.MinChecks != nil && agg.Checks < *policy.MinChecks {
			result.PolicyFailures = append(result.PolicyFailures, fmt.Sprintf("track %s has %d checks below required min_checks %d", slug, agg.Checks, *policy.MinChecks))
		}
		if policy.MinPassRate != nil {
			trackPassRate := 0.0
			if agg.Checks > 0 {
				trackPassRate = float64(agg.Passes) / float64(agg.Checks)
			}
			if trackPassRate < *policy.MinPassRate {
				result.PolicyFailures = append(result.PolicyFailures, fmt.Sprintf("track %s pass rate %.3f below required %.3f", slug, trackPassRate, *policy.MinPassRate))
			}
		}
		writeTrackAggregate(result.TrackAggregates, slug, agg)
	}

	result.PassedPolicyRequirements = len(result.Failures) == 0 && len(result.PolicyFailures) == 0
	return result, nil
}

func SortedObjectiveEvalTrackSlugs(aggs map[string]ObjectiveEvalTrackAggregate) []string {
	slugs := make([]string, 0, len(aggs))
	for slug := range aggs {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs
}

func buildSkillScores(items []skillFixture) []domain.SkillScore {
	out := make([]domain.SkillScore, 0, len(items))
	for _, item := range items {
		skill := strings.TrimSpace(item.Skill)
		if skill == "" {
			continue
		}
		score := item.Score
		if score <= 0 {
			score = 3
		}
		out = append(out, domain.SkillScore{Skill: skill, Score: score, ScoreSource: "deterministic", ScoreVersion: "det-v1", ScoreEvidenceJSON: `{"rubric_id":"objective-score-eval-fixture"}`})
	}
	if len(out) == 0 {
		out = append(out, domain.SkillScore{Skill: "clarity and coherence", Score: 3, ScoreSource: "deterministic", ScoreVersion: "det-v1", ScoreEvidenceJSON: `{"rubric_id":"objective-score-eval-fallback"}`})
	}
	return out
}

func objectiveEvalScoreByCode(scores []domain.ObjectiveScore) map[string]domain.ObjectiveScore {
	out := map[string]domain.ObjectiveScore{}
	for _, score := range scores {
		out[score.TGOCode] = score
	}
	return out
}

func readTrackAggregate(store map[string]ObjectiveEvalTrackAggregate, slug string) ObjectiveEvalTrackAggregate {
	norm := normalizeTrackSlug(slug)
	return store[norm]
}

func writeTrackAggregate(store map[string]ObjectiveEvalTrackAggregate, slug string, agg ObjectiveEvalTrackAggregate) {
	store[normalizeTrackSlug(slug)] = agg
}

func normalizeTrackSlug(slug string) string {
	trim := strings.TrimSpace(slug)
	if trim == "" {
		return "unknown"
	}
	return trim
}
