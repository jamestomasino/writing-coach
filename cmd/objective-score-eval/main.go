package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomasino/writing-coach/internal/review"
)

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

	result, err := review.EvaluateObjectiveScoreCorpusPath(abs)
	if err != nil {
		die("evaluate corpus", err)
	}

	fmt.Printf("objective-score-eval: corpus=%s version=%s\n", abs, strings.TrimSpace(result.CorpusVersion))
	fmt.Printf("objective-score-eval: checks=%d passed=%d failed=%d pass_rate=%.3f required=%.3f\n", result.TotalChecks, result.PassedChecks, result.FailedChecks(), result.PassRate, result.RequiredMinPassRate)
	if result.MaxPairwiseTieRate != nil {
		fmt.Printf("objective-score-eval: pairwise_checks=%d ties=%d tie_rate=%.3f allowed<=%.3f\n", result.PairwiseChecks, result.PairwiseTies, result.PairwiseTieRate, *result.MaxPairwiseTieRate)
	}

	for _, slug := range review.SortedObjectiveEvalTrackSlugs(result.TrackAggregates) {
		agg := result.TrackAggregates[slug]
		trackPassRate := 0.0
		if agg.Checks > 0 {
			trackPassRate = float64(agg.Passes) / float64(agg.Checks)
		}
		fmt.Printf("objective-score-eval: track=%s checks=%d passed=%d failed=%d pass_rate=%.3f\n", slug, agg.Checks, agg.Passes, agg.Checks-agg.Passes, trackPassRate)
	}

	if len(result.Failures) > 0 {
		for _, item := range result.Failures {
			fmt.Fprintf(os.Stderr, "objective-score-eval: case=%s %s\n", item.Case, item.Detail)
		}
	}
	if len(result.PolicyFailures) > 0 {
		for _, msg := range result.PolicyFailures {
			fmt.Fprintf(os.Stderr, "objective-score-eval: %s\n", msg)
		}
	}
	if !result.PassedPolicyRequirements {
		dieMsg("objective eval corpus failed policy requirements")
	}
	fmt.Println("objective-score-eval: ok")
}

func die(msg string, err error) {
	fmt.Fprintf(os.Stderr, "objective-score-eval: %s: %v\n", msg, err)
	os.Exit(1)
}

func dieMsg(msg string) {
	fmt.Fprintf(os.Stderr, "objective-score-eval: %s\n", msg)
	os.Exit(1)
}
