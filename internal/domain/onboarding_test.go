package domain

import (
	"strings"
	"testing"
)

func TestPromptToneGuidanceInterpretsSimpleToneLabels(t *testing.T) {
	got := PromptToneGuidance(OnboardingProfile{
		DesiredTone: "serious and emotional",
	})

	lower := strings.ToLower(got)
	if strings.Contains(lower, "serious and emotional") {
		t.Fatalf("expected interpreted guidance, got %q", got)
	}
	if !strings.Contains(lower, "weighty") || !strings.Contains(lower, "emotional stakes") {
		t.Fatalf("expected mapped guidance, got %q", got)
	}
}

func TestPromptProfileLinesUseInterpretedToneGuidance(t *testing.T) {
	lines := PromptProfileLines(OnboardingProfile{
		WritingType:      "marketing",
		AssignmentFormat: "landing page",
		TargetAudience:   "buyers",
		SubjectMatter:    "product launch",
		DesiredTone:      "clear and persuasive",
		WritingGoals:     "Write sharper copy.",
	})

	joined := strings.ToLower(strings.Join(lines, "\n"))
	if !strings.Contains(joined, "tone guidance:") {
		t.Fatalf("expected tone guidance line, got %q", joined)
	}
	if strings.Contains(joined, "clear and persuasive") {
		t.Fatalf("expected interpreted tone guidance, got %q", joined)
	}
}
