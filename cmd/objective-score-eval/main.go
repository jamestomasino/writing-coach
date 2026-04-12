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
	Version     string     `json:"version"`
	MinPassRate float64    `json:"min_pass_rate"`
	Cases       []evalCase `json:"cases"`
}

type evalCase struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	TreeSlug        string         `json:"tree_slug"`
	WritingType     string         `json:"writing_type"`
	SkillScores     []skillFixture `json:"skill_scores"`
	LeftCode        string         `json:"left_code"`
	RightCode       string         `json:"right_code"`
	LeftWinsMetrics map[string]int `json:"left_wins_metrics"`
	RightWinsMetrics map[string]int `json:"right_wins_metrics"`
	Code            string         `json:"code"`
	LowMetrics      map[string]int `json:"low_metrics"`
	HighMetrics     map[string]int `json:"high_metrics"`
}

type skillFixture struct {
	Skill string `json:"skill"`
	Score int    `json:"score"`
}

type failure struct {
	Case   string
	Detail string
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

	totalChecks := 0
	passedChecks := 0
	var failures []failure
	trackChecks := map[string]int{}
	trackPasses := map[string]int{}

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
			trackChecks[options.TreeSlug]++
			if leftMap[leftCode].Score > leftMap[rightCode].Score {
				passedChecks++
				trackPasses[options.TreeSlug]++
			} else {
				failures = append(failures, failure{Case: caseName, Detail: fmt.Sprintf("leftWins expected %s>%s got %d<=%d", leftCode, rightCode, leftMap[leftCode].Score, leftMap[rightCode].Score)})
			}

			rightPass := review.BuildObjectiveScores(900, active, assessments, skillScores, analyzer.Report{Metrics: tc.RightWinsMetrics}, options)
			rightMap := objectiveScoreByCode(rightPass)
			totalChecks++
			trackChecks[options.TreeSlug]++
			if rightMap[rightCode].Score > rightMap[leftCode].Score {
				passedChecks++
				trackPasses[options.TreeSlug]++
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
			trackChecks[options.TreeSlug]++
			if highScore >= lowScore {
				passedChecks++
				trackPasses[options.TreeSlug]++
			} else {
				failures = append(failures, failure{Case: caseName, Detail: fmt.Sprintf("monotonic expected high>=low got %d<%d", highScore, lowScore)})
			}
		default:
			failures = append(failures, failure{Case: caseName, Detail: "unknown case type: " + caseType})
		}
	}

	passRate := float64(passedChecks) / float64(totalChecks)
	fmt.Printf("objective-score-eval: corpus=%s version=%s\n", abs, strings.TrimSpace(c.Version))
	fmt.Printf("objective-score-eval: checks=%d passed=%d failed=%d pass_rate=%.3f required=%.3f\n", totalChecks, passedChecks, totalChecks-passedChecks, passRate, c.MinPassRate)

	slugs := make([]string, 0, len(trackChecks))
	for slug := range trackChecks {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	for _, slug := range slugs {
		checks := trackChecks[slug]
		passes := trackPasses[slug]
		fmt.Printf("objective-score-eval: track=%s checks=%d passed=%d failed=%d\n", slug, checks, passes, checks-passes)
	}

	if len(failures) > 0 {
		for _, item := range failures {
			fmt.Fprintf(os.Stderr, "objective-score-eval: case=%s %s\n", item.Case, item.Detail)
		}
	}
	if passRate < c.MinPassRate {
		dieMsg(fmt.Sprintf("pass rate %.3f below required %.3f", passRate, c.MinPassRate))
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

func die(msg string, err error) {
	fmt.Fprintf(os.Stderr, "objective-score-eval: %s: %v\n", msg, err)
	os.Exit(1)
}

func dieMsg(msg string) {
	fmt.Fprintf(os.Stderr, "objective-score-eval: %s\n", msg)
	os.Exit(1)
}
