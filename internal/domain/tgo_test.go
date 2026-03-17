package domain

import "testing"

func TestInferProgressMode(t *testing.T) {
	if got := InferProgressMode("sentence variety"); got != ProgressModePercent {
		t.Fatalf("sentence variety progress mode = %q", got)
	}
	if got := InferProgressMode("symbolic control"); got != ProgressModeStage {
		t.Fatalf("symbolic control progress mode = %q", got)
	}
}

func TestComputeMasterySignalPercent(t *testing.T) {
	signal := ComputeMasterySignal(ProgressModePercent, []string{"mastered", "secure", "mastered"})
	if !signal.Ready {
		t.Fatalf("signal should be ready: %#v", signal)
	}
	if signal.Percent != 100 {
		t.Fatalf("percent = %d", signal.Percent)
	}
}

func TestComputeMasterySignalStage(t *testing.T) {
	signal := ComputeMasterySignal(ProgressModeStage, []string{"secure", "developing"})
	if signal.Stage != "developing" && signal.Stage != "strong control" {
		t.Fatalf("unexpected stage: %#v", signal)
	}
	if signal.Ready {
		t.Fatalf("signal should not be ready: %#v", signal)
	}
}
