package review

import "testing"

func TestEvaluateObjectiveScoreCorpusDefaultPassesPolicy(t *testing.T) {
	result, err := EvaluateObjectiveScoreCorpus(DefaultObjectiveEvalCorpus())
	if err != nil {
		t.Fatalf("evaluate corpus: %v", err)
	}
	if result.TotalChecks == 0 {
		t.Fatal("expected objective eval checks")
	}
	if !result.PassedPolicyRequirements {
		t.Fatalf("expected policy pass, got failures=%v policy_failures=%v", result.Failures, result.PolicyFailures)
	}
}
