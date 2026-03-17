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
	if !strings.Contains(joined, "assignment seed:") {
		t.Fatalf("expected assignment seed guidance, got %q", joined)
	}
}

func TestPromptScenarioGuidanceDemandsConcreteSituation(t *testing.T) {
	got := PromptScenarioGuidance(OnboardingProfile{
		WritingType:      "fantasy fiction",
		AssignmentFormat: "scene",
		TargetAudience:   "fantasy readers",
		SubjectMatter:    "oaths, sacred objects, and succession fights",
	})

	lower := strings.ToLower(got)
	if !strings.Contains(lower, "concrete") || !strings.Contains(lower, "specific pressure point") {
		t.Fatalf("expected concrete fiction guidance, got %q", got)
	}
	if !strings.Contains(lower, "oaths") {
		t.Fatalf("expected subject matter to shape guidance, got %q", got)
	}
}

func TestGeneratedTreeDisplayUsesProfileLanguageInsteadOfTemplateBranding(t *testing.T) {
	profile := OnboardingProfile{
		WritingType:      "fiction",
		AssignmentFormat: "scene",
		TargetAudience:   "fantasy readers",
		DesiredTone:      "serious and emotional, literary, philosophical",
		WritingGoals:     "I want to become a great author at mythopoeic literature, epic fantasy, and world building.",
	}

	title, description := GeneratedTreeDisplay("James Tomasino", profile, "Mythic Fiction Track", "Advanced mythopoeic tragic fiction track.")
	if strings.Contains(title, "Mythic Fiction Track") {
		t.Fatalf("expected title to use profile language, got %q", title)
	}
	if title != "James Tomasino's Fiction Track" {
		t.Fatalf("unexpected title %q", title)
	}
	if strings.Contains(description, "Advanced mythopoeic tragic fiction track.") {
		t.Fatalf("expected description to avoid template branding, got %q", description)
	}
	if !strings.Contains(description, "Skill track for fiction, with scene assignments for fantasy readers.") {
		t.Fatalf("expected description summary to use profile language, got %q", description)
	}
}

func TestTemplateKeyForProfileUsesBroaderFictionTemplates(t *testing.T) {
	cases := []struct {
		name    string
		profile OnboardingProfile
		want    string
	}{
		{
			name: "generic fiction defaults to story craft",
			profile: OnboardingProfile{
				WritingType:     "fiction",
				ExperienceLevel: "advanced",
				DesiredTone:     "serious and emotional",
				WritingGoals:    "write stronger scenes",
			},
			want: "story-craft",
		},
		{
			name: "fantasy gets fantasy template",
			profile: OnboardingProfile{
				WritingType:     "fantasy fiction",
				ExperienceLevel: "advanced",
				DesiredTone:     "serious and emotional",
				WritingGoals:    "write an epic fantasy novel",
			},
			want: "fantasy-fiction",
		},
		{
			name: "science fiction gets science fiction template",
			profile: OnboardingProfile{
				WritingType:     "science fiction",
				ExperienceLevel: "advanced",
				DesiredTone:     "clear and reflective",
				WritingGoals:    "write stronger science fiction",
			},
			want: "science-fiction",
		},
		{
			name: "romance gets romance template",
			profile: OnboardingProfile{
				WritingType:     "romance",
				ExperienceLevel: "intermediate",
				DesiredTone:     "warm and emotional",
				WritingGoals:    "write stronger romance scenes",
			},
			want: "romance-fiction",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := TemplateKeyForProfile(tc.profile); got != tc.want {
				t.Fatalf("template key = %q, want %q", got, tc.want)
			}
		})
	}
}
