package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/review"
)

type corpus struct {
	Version            string        `json:"version"`
	MinPassRate        float64       `json:"min_pass_rate"`
	MaxPairwiseTieRate *float64      `json:"max_pairwise_tie_rate,omitempty"`
	TrackPolicies      []trackPolicy `json:"track_policies,omitempty"`
	Cases              []evalCase    `json:"cases"`
}

type trackPolicy struct {
	TreeSlug    string   `json:"tree_slug"`
	MinPassRate *float64 `json:"min_pass_rate,omitempty"`
	MinChecks   *int     `json:"min_checks,omitempty"`
}

type evalCase struct {
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

type failure struct {
	Case   string
	Detail string
}

type trackAggregate struct {
	Checks int
	Passes int
}

func main() {
	path := flag.String("corpus", "internal/review/testdata/objective_eval_corpus.json", "path to objective eval corpus JSON")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		die("resolve cwd", err)
	}
	abs := *path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(root, abs)
	}

	raw, err := os.ReadFile(abs)
	if err != nil {
		die("read corpus", err)
	}
	var c corpus
	if err := json.Unmarshal(raw, &c); err != nil {
		die("decode corpus", err)
	}
	if len(c.Cases) == 0 {
		dieMsg("corpus has zero cases")
	}
	if c.MinPassRate <= 0 || c.MinPassRate > 1 {
		dieMsg("min_pass_rate must be within (0,1]")
	}
	if c.MaxPairwiseTieRate != nil && (*c.MaxPairwiseTieRate < 0 || *c.MaxPairwiseTieRate > 1) {
		dieMsg("max_pairwise_tie_rate must be within [0,1]")
	}

	totalChecks := 0
	passedChecks := 0
	pairwiseChecks := 0
	pairwiseTies := 0
	var failures []failure
	trackAgg := map[string]*trackAggregate{}
	trackPolicyBySlug := map[string]trackPolicy{}
	for _, policy := range c.TrackPolicies {
		slug := strings.TrimSpace(policy.TreeSlug)
		if slug == "" {
			continue
		}
		trackPolicyBySlug[slug] = policy
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
				failures = append(failures, failure{Case: caseName, Detail: "pairwise case missing left_code or right_code"})
				continue
			}
			active := []domain.TGO{{Code: leftCode}, {Code: rightCode}}
			assessments := []domain.TGOAssessment{{TGOCode: leftCode, Status: "developing"}, {TGOCode: rightCode, Status: "developing"}}

			leftPass := review.BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.LeftWinsMetrics}, options)
			leftMap := objectiveScoreByCode(leftPass)
			totalChecks++
			pairwiseChecks++
			agg := ensureTrackAggregate(trackAgg, options.TreeSlug)
			agg.Checks++
			if leftMap[leftCode].Score == leftMap[rightCode].Score {
				pairwiseTies++
			}
			if leftMap[leftCode].Score > leftMap[rightCode].Score {
				passedChecks++
				agg.Passes++
			} else {
				failures = append(failures, failure{Case: caseName, Detail: fmt.Sprintf("leftWins expected %s>%s got %d<=%d", leftCode, rightCode, leftMap[leftCode].Score, leftMap[rightCode].Score)})
			}

			rightPass := review.BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.RightWinsMetrics}, options)
			rightMap := objectiveScoreByCode(rightPass)
			totalChecks++
			pairwiseChecks++
			agg = ensureTrackAggregate(trackAgg, options.TreeSlug)
			agg.Checks++
			if rightMap[rightCode].Score == rightMap[leftCode].Score {
				pairwiseTies++
			}
			if rightMap[rightCode].Score > rightMap[leftCode].Score {
				passedChecks++
				agg.Passes++
			} else {
				failures = append(failures, failure{Case: caseName, Detail: fmt.Sprintf("rightWins expected %s>%s got %d<=%d", rightCode, leftCode, rightMap[rightCode].Score, rightMap[leftCode].Score)})
			}
		case "monotonic":
			code := strings.TrimSpace(tc.Code)
			if code == "" {
				failures = append(failures, failure{Case: caseName, Detail: "monotonic case missing code"})
				continue
			}
			active := []domain.TGO{{Code: code}}
			assessments := []domain.TGOAssessment{{TGOCode: code, Status: "developing"}}
			low := review.BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.LowMetrics}, options)
			high := review.BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.HighMetrics}, options)
			lowScore := objectiveScoreByCode(low)[code].Score
			highScore := objectiveScoreByCode(high)[code].Score
			totalChecks++
			agg := ensureTrackAggregate(trackAgg, options.TreeSlug)
			agg.Checks++
			if highScore >= lowScore {
				passedChecks++
				agg.Passes++
			} else {
				failures = append(failures, failure{Case: caseName, Detail: fmt.Sprintf("monotonic expected high>=low got %d<%d", highScore, lowScore)})
			}
		default:
			failures = append(failures, failure{Case: caseName, Detail: "unknown case type: " + caseType})
		}
	}

	passRate := float64(passedChecks) / float64(totalChecks)
	pairwiseTieRate := 0.0
	if pairwiseChecks > 0 {
		pairwiseTieRate = float64(pairwiseTies) / float64(pairwiseChecks)
	}
	fmt.Printf("objective-score-eval: corpus=%s version=%s\n", abs, strings.TrimSpace(c.Version))
	fmt.Printf("objective-score-eval: checks=%d passed=%d failed=%d pass_rate=%.3f required=%.3f\n", totalChecks, passedChecks, totalChecks-passedChecks, passRate, c.MinPassRate)
	if c.MaxPairwiseTieRate != nil {
		fmt.Printf("objective-score-eval: pairwise_checks=%d ties=%d tie_rate=%.3f allowed<=%.3f\n", pairwiseChecks, pairwiseTies, pairwiseTieRate, *c.MaxPairwiseTieRate)
	}

	slugs := make([]string, 0, len(trackAgg))
	for slug := range trackAgg {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	for _, slug := range slugs {
		agg := trackAgg[slug]
		trackPassRate := 0.0
		if agg.Checks > 0 {
			trackPassRate = float64(agg.Passes) / float64(agg.Checks)
		}
		fmt.Printf("objective-score-eval: track=%s checks=%d passed=%d failed=%d pass_rate=%.3f\n", slug, agg.Checks, agg.Passes, agg.Checks-agg.Passes, trackPassRate)
	}

	if len(failures) > 0 {
		for _, item := range failures {
			fmt.Fprintf(os.Stderr, "objective-score-eval: case=%s %s\n", item.Case, item.Detail)
		}
	}
	if passRate < c.MinPassRate {
		dieMsg(fmt.Sprintf("pass rate %.3f below required %.3f", passRate, c.MinPassRate))
	}
	if c.MaxPairwiseTieRate != nil && pairwiseTieRate > *c.MaxPairwiseTieRate {
		dieMsg(fmt.Sprintf("pairwise tie rate %.3f above allowed %.3f", pairwiseTieRate, *c.MaxPairwiseTieRate))
	}
	for _, policy := range c.TrackPolicies {
		slug := strings.TrimSpace(policy.TreeSlug)
		if slug == "" {
			continue
		}
		agg := ensureTrackAggregate(trackAgg, slug)
		if policy.MinChecks != nil && agg.Checks < *policy.MinChecks {
			dieMsg(fmt.Sprintf("track %s has %d checks below required min_checks %d", slug, agg.Checks, *policy.MinChecks))
		}
		if policy.MinPassRate != nil {
			trackPassRate := 0.0
			if agg.Checks > 0 {
				trackPassRate = float64(agg.Passes) / float64(agg.Checks)
			}
			if trackPassRate < *policy.MinPassRate {
				dieMsg(fmt.Sprintf("track %s pass rate %.3f below required %.3f", slug, trackPassRate, *policy.MinPassRate))
			}
		}
	}
	fmt.Println("objective-score-eval: ok")
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
		out = append(out, domain.SkillScore{
			Skill:             skill,
			Score:             score,
			ScoreSource:       "deterministic",
			ScoreVersion:      "det-v1",
			ScoreEvidenceJSON: `{"rubric_id":"objective-score-eval-fixture"}`,
		})
	}
	if len(out) == 0 {
		out = append(out, domain.SkillScore{Skill: "clarity and coherence", Score: 3, ScoreSource: "deterministic", ScoreVersion: "det-v1", ScoreEvidenceJSON: `{"rubric_id":"objective-score-eval-fallback"}`})
	}
	return out
}

func objectiveScoreByCode(scores []domain.ObjectiveScore) map[string]domain.ObjectiveScore {
	out := map[string]domain.ObjectiveScore{}
	for _, score := range scores {
		out[score.TGOCode] = score
	}
	return out
}

func ensureTrackAggregate(store map[string]*trackAggregate, slug string) *trackAggregate {
	trim := strings.TrimSpace(slug)
	if trim == "" {
		trim = "unknown"
	}
	agg := store[trim]
	if agg == nil {
		agg = &trackAggregate{}
		store[trim] = agg
	}
	return agg
}

func die(msg string, err error) {
	fmt.Fprintf(os.Stderr, "objective-score-eval: %s: %v\n", msg, err)
	os.Exit(1)
}

func dieMsg(msg string) {
	fmt.Fprintf(os.Stderr, "objective-score-eval: %s\n", msg)
	os.Exit(1)
}
