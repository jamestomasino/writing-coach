package openai

import (
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestPromptUserInputBuildersIncludeCoreFields(t *testing.T) {
	exerciseInput := ExerciseRequest{
		WritingLanguage: "en",
		OnboardingProfile: &domain.OnboardingProfile{
			WritingType:      "technical writing",
			AssignmentFormat: "how-to guide",
		},
		ActiveTGOs:        []domain.TGO{{Code: "story-causal-clarity", Description: "clarify causal movement"}},
		CurrentFocus:      "clarity",
		DifficultyLevel:   3,
		RecentTitles:      []string{"Prior One"},
		RecentWeaknesses:  []string{"passive voice"},
		RecurringFindings: []string{"long sentences"},
		CoachingBrief:     "short brief",
	}
	exercisePrompt := ExerciseUserInput(exerciseInput)
	for _, needle := range []string{"Writing language:", "Writing track profile:", "Hidden review guidance:", "Current coaching emphasis:", "Difficulty level:"} {
		if !strings.Contains(exercisePrompt, needle) {
			t.Fatalf("exercise prompt missing %q: %q", needle, exercisePrompt)
		}
	}

	revisionPrompt := RevisionUserInput(RevisionExerciseRequest{
		WritingLanguage:   "en",
		CurrentFocus:      "prose precision",
		DifficultyLevel:   2,
		ActiveTGOs:        []domain.TGO{{Code: "story-scene-architecture"}},
		SubmissionID:      42,
		SubmissionContent: "draft text",
		Weaknesses:        []string{"weak opening"},
		AnalyzerFindings:  []string{"hedging"},
		ComparisonSummary: "better transitions",
		RecentWeaknesses:  []string{"wordiness"},
		RecurringFindings: []string{"comma splices"},
		CoachingBrief:     "coach",
	})
	for _, needle := range []string{"Submission ID: 42", "Submission:\ndraft text", "Comparison summary:", "Recurring analyzer findings:"} {
		if !strings.Contains(revisionPrompt, needle) {
			t.Fatalf("revision prompt missing %q: %q", needle, revisionPrompt)
		}
	}

	reviewPrompt := ReviewUserInput(ReviewRequest{
		WritingLanguage: "en",
		SubmissionID:    7,
		WordCount:       120,
		ActiveTGOs:      []domain.TGO{{Code: "story-causal-clarity"}},
		CompletedTGOs:   []domain.TGO{{Code: "story-scene-architecture"}},
		AnalysisSummary: "analysis",
		AnalyzerFindings: []string{
			"finding a",
		},
		CoachingBrief: "brief",
		Content:       "submission body",
	})
	for _, needle := range []string{"Word count: 120", "Completed TGOs to monitor for regression:", "Deterministic analysis summary:", "Submission:\nsubmission body"} {
		if !strings.Contains(reviewPrompt, needle) {
			t.Fatalf("review prompt missing %q: %q", needle, reviewPrompt)
		}
	}
}

func TestOpenAIHelpersNormalizeAndValidate(t *testing.T) {
	if got := EmptyDefault("  ", "fallback"); got != "fallback" {
		t.Fatalf("EmptyDefault = %q", got)
	}
	if got := JoinOrDefault([]string{" a ", "", "b"}, "none"); got != " a , , b" {
		t.Fatalf("JoinOrDefault = %q", got)
	}
	if got := JoinTGOs([]domain.TGO{{Code: "a", Description: "d1"}, {Code: "b", Description: "d2"}}); got != "a: d1 | b: d2" {
		t.Fatalf("JoinTGOs = %q", got)
	}
	if got := JoinTGOs(nil); got != "none" {
		t.Fatalf("JoinTGOs nil = %q", got)
	}

	focus := CoalesceFocusSkills([]domain.TGO{{Code: "story-causal-clarity", Title: "Clarity"}}, nil)
	if len(focus) != 1 || focus[0] != "narrative clarity" {
		t.Fatalf("CoalesceFocusSkills = %#v", focus)
	}
	codes := ActiveTGOCodes([]domain.TGO{{Code: "x"}, {Code: "y"}})
	if len(codes) != 2 || codes[0] != "x" || codes[1] != "y" {
		t.Fatalf("ActiveTGOCodes = %#v", codes)
	}

	if err := ValidateExercise(ExerciseResponse{Title: "t", Brief: "b", Constraints: []string{"c"}, FocusSkills: []string{"f"}, SuccessCriteria: []string{"s"}}); err != nil {
		t.Fatalf("ValidateExercise unexpected err: %v", err)
	}
	if err := ValidateExercise(ExerciseResponse{Title: "", Brief: "b", Constraints: []string{"c"}, FocusSkills: []string{"f"}, SuccessCriteria: []string{"s"}}); err == nil {
		t.Fatal("expected ValidateExercise error for empty title")
	}

	if err := ValidateReview(ReviewResponse{Summary: "ok", Strengths: []string{"s"}, Weaknesses: []string{"w"}, NextFocus: "n", SkillScores: []SkillScore{{Skill: "k", Score: 3}}, TGOAssessments: []TGOAssessment{{Code: "c1", Status: "secure", Evidence: "e"}, {Code: "c2", Status: "secure", Evidence: "e"}, {Code: "c3", Status: "secure", Evidence: "e"}}}); err != nil {
		t.Fatalf("ValidateReview unexpected err: %v", err)
	}
	if err := ValidateReview(ReviewResponse{Summary: "", Strengths: []string{"s"}, Weaknesses: []string{"w"}, NextFocus: "n", SkillScores: []SkillScore{{Skill: "k", Score: 3}}, TGOAssessments: []TGOAssessment{{Code: "c1", Status: "secure", Evidence: "e"}, {Code: "c2", Status: "secure", Evidence: "e"}, {Code: "c3", Status: "secure", Evidence: "e"}}}); err == nil {
		t.Fatal("expected ValidateReview error")
	}
}

func TestSchemaHelpersAndDomainConversions(t *testing.T) {
	arr := StringArraySchema(2, 4)
	if arr["type"] != "array" {
		t.Fatalf("StringArraySchema type = %#v", arr["type"])
	}
	enumArr := EnumStringArraySchema(1, 2, []string{"a", "b"})
	items, ok := enumArr["items"].(map[string]any)
	if !ok || items["type"] != "string" {
		t.Fatalf("EnumStringArraySchema items = %#v", enumArr["items"])
	}

	assessments := ToDomainAssessments([]TGOAssessment{{Code: "code", Status: "secure", Evidence: "ev"}})
	if len(assessments) != 1 || assessments[0].TGOCode != "code" {
		t.Fatalf("ToDomainAssessments = %#v", assessments)
	}
	annotations := ToDomainAnnotations([]Annotation{{Quote: "q", TGOCode: "t", Category: "c", Comment: "m", Severity: "high"}})
	if len(annotations) != 1 || annotations[0].Quote != "q" {
		t.Fatalf("ToDomainAnnotations = %#v", annotations)
	}
}
