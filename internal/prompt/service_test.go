package prompt

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/llm"
	"github.com/tomasino/writing-coach/internal/review"
)

type fakeLLMClient struct {
	generateExercise func(context.Context, llm.ExerciseRequest) (domain.Exercise, error)
}

func (f fakeLLMClient) Enabled() bool { return true }

func (f fakeLLMClient) ValidateCredentials(context.Context) error { return nil }

func (f fakeLLMClient) GenerateExercise(ctx context.Context, input llm.ExerciseRequest) (domain.Exercise, error) {
	return f.generateExercise(ctx, input)
}

func (f fakeLLMClient) GenerateRevisionExercise(context.Context, llm.RevisionExerciseRequest) (domain.Exercise, error) {
	return domain.Exercise{}, errors.New("not implemented")
}

func (f fakeLLMClient) ReviewSubmission(context.Context, llm.ReviewRequest) (domain.Review, []domain.SkillScore, error) {
	return domain.Review{}, nil, errors.New("not implemented")
}

func TestDeterministicRevisionExercise(t *testing.T) {
	service := NewService(nil)
	sub := domain.Submission{ID: 7, DraftNumber: 2, Content: "A court scene."}
	rev := domain.Review{
		Weaknesses:       []string{"Relationship clarity is weak."},
		AnalyzerFindings: []string{"Adverb density is elevated."},
	}
	cmp := review.Comparison{
		Summary:              "Revision still carries forward several earlier weaknesses.",
		PersistingWeaknesses: []string{"Relationship clarity is weak."},
	}

	ex := service.RevisionExercise(context.Background(), Context{
		CurriculumState:    domain.CurriculumState{CurrentFocus: "tragic inevitability"},
		RevisionOf:         &sub,
		RevisionReview:     &rev,
		RevisionComparison: &cmp,
	})

	if !strings.Contains(ex.Title, "Revision") {
		t.Fatalf("unexpected title: %q", ex.Title)
	}
	if ex.SourceSubmissionID != 7 {
		t.Fatalf("source submission id = %d", ex.SourceSubmissionID)
	}
	if len(ex.Constraints) == 0 || len(ex.SuccessCriteria) == 0 {
		t.Fatal("expected populated revision brief")
	}
}

func TestDeterministicNextExerciseUsesFreshDraftLanguage(t *testing.T) {
	service := NewService(nil)

	ex := service.NextExercise(context.Background(), Context{
		CurriculumState: domain.CurriculumState{CurrentFocus: "causal clarity"},
	})

	lower := strings.ToLower(ex.Brief)
	if strings.Contains(lower, "rewrite") || strings.Contains(lower, "revise") {
		t.Fatalf("expected fresh-draft language, got %q", ex.Brief)
	}
	if !strings.Contains(lower, "write a new piece") {
		t.Fatalf("expected explicit new-piece language, got %q", ex.Brief)
	}
}

func TestDeterministicNextExerciseUsesOnboardingProfileAsPromptSeed(t *testing.T) {
	service := NewService(nil)

	profile := &domain.OnboardingProfile{
		WritingType:      "marketing",
		AssignmentFormat: "landing page",
		TargetAudience:   "product-led growth teams",
		SubjectMatter:    "B2B SaaS launches",
		DesiredTone:      "clear and persuasive",
		WritingGoals:     "Drive sharper conversion-focused drafts.",
		DesiredOutcomes:  []string{"improve professional communication"},
	}
	active := []domain.TGO{
		{Code: "dialogue-intelligence", Title: "Dialogue Intelligence"},
		{Code: "scene-architecture", Title: "Scene Architecture"},
		{Code: "prose-precision", Title: "Prose Precision"},
	}

	ex := service.NextExercise(context.Background(), Context{
		OnboardingProfile: profile,
		ActiveTGOs:        active,
	})

	lower := strings.ToLower(ex.Brief)
	if !strings.Contains(lower, "landing page") || !strings.Contains(lower, "b2b saas launches") {
		t.Fatalf("expected onboarding profile to shape brief, got %q", ex.Brief)
	}
	if strings.Contains(lower, "dialogue intelligence") || strings.Contains(lower, "scene architecture") {
		t.Fatalf("expected brief to avoid naming review rubric skills, got %q", ex.Brief)
	}
	if len(ex.TGOCodes) != 3 {
		t.Fatalf("expected selected review tg os to persist, got %v", ex.TGOCodes)
	}
}

func TestDeterministicNextExerciseAddsConcreteScenarioGuidance(t *testing.T) {
	service := NewService(nil)

	profile := &domain.OnboardingProfile{
		WritingType:      "fantasy fiction",
		AssignmentFormat: "scene",
		TargetAudience:   "fantasy readers",
		SubjectMatter:    "inheritance fights over sacred relics",
		DesiredTone:      "serious and emotional",
	}

	ex := service.NextExercise(context.Background(), Context{
		OnboardingProfile: profile,
	})

	lower := strings.ToLower(ex.Brief)
	if !strings.Contains(lower, "specific pressure point") {
		t.Fatalf("expected concrete scenario pressure in brief, got %q", ex.Brief)
	}
	if !strings.Contains(lower, "inheritance fights over sacred relics") {
		t.Fatalf("expected subject matter in brief, got %q", ex.Brief)
	}
}

func TestNextExerciseIgnoresCanceledParentContextForProviderCall(t *testing.T) {
	parentCtx, cancel := context.WithCancel(context.Background())
	cancel()

	service := NewService(fakeLLMClient{
		generateExercise: func(ctx context.Context, input llm.ExerciseRequest) (domain.Exercise, error) {
			if err := ctx.Err(); err != nil {
				return domain.Exercise{}, err
			}
			return domain.Exercise{
				Title:           "Provider Draft",
				Brief:           "Generated by provider.",
				Constraints:     []string{"keep it focused"},
				FocusSkills:     []string{"narrative clarity"},
				SuccessCriteria: []string{"clear causal chain"},
			}, nil
		},
	}).WithGenerationTimeout(100 * time.Millisecond)

	exercise := service.NextExercise(parentCtx, Context{
		CurriculumState: domain.CurriculumState{CurrentFocus: "narrative clarity", DifficultyLevel: 2},
		ActiveTGOs:      []domain.TGO{{Code: "story-causal-clarity", Title: "Causal Clarity", Description: "Keep cause and effect readable."}},
	})

	if exercise.Title != "Provider Draft" {
		t.Fatalf("exercise title = %q", exercise.Title)
	}
	if exercise.GenerationKind != "openai" {
		t.Fatalf("generation kind = %q", exercise.GenerationKind)
	}
}

func TestNextExerciseFallsBackWhenProviderGenerationTimesOut(t *testing.T) {
	service := NewService(fakeLLMClient{
		generateExercise: func(ctx context.Context, input llm.ExerciseRequest) (domain.Exercise, error) {
			<-ctx.Done()
			return domain.Exercise{}, ctx.Err()
		},
	}).WithGenerationTimeout(10 * time.Millisecond)

	start := time.Now()
	exercise := service.NextExercise(context.Background(), Context{
		CurriculumState: domain.CurriculumState{CurrentFocus: "narrative clarity", DifficultyLevel: 2},
		ActiveTGOs: []domain.TGO{
			{Code: "story-causal-clarity", Title: "Causal Clarity", Description: "Keep cause and effect readable."},
			{Code: "story-scene-architecture", Title: "Scene Architecture", Description: "Make scene turns legible."},
			{Code: "story-prose-precision", Title: "Prose Precision", Description: "Tighten sentence-level decisions."},
		},
	})
	elapsed := time.Since(start)

	if exercise.GenerationKind != "deterministic-fallback" {
		t.Fatalf("generation kind = %q", exercise.GenerationKind)
	}
	if !strings.Contains(exercise.ProviderNote, "context deadline exceeded") {
		t.Fatalf("provider note = %q", exercise.ProviderNote)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("fallback took too long: %s", elapsed)
	}
}

func TestNextExerciseRetriesWhenScenarioPatternRepeats(t *testing.T) {
	callCount := 0
	service := NewService(fakeLLMClient{
		generateExercise: func(ctx context.Context, input llm.ExerciseRequest) (domain.Exercise, error) {
			callCount++
			if callCount == 1 {
				return domain.Exercise{
					Title:           "Choose Before Dawn",
					Brief:           "Your hero must choose between saving a sibling or guarding the city before dawn.",
					Constraints:     []string{"keep it concrete"},
					FocusSkills:     []string{"narrative clarity"},
					SuccessCriteria: []string{"decision is clear"},
				}, nil
			}
			if !strings.Contains(strings.ToLower(input.CoachingBrief), "variety requirement") {
				t.Fatalf("expected retry coaching brief to include variety guidance, got %q", input.CoachingBrief)
			}
			return domain.Exercise{
				Title:           "Broken Treaty Meeting",
				Brief:           "Two rivals negotiate a fragile ceasefire while both hide the same critical fact.",
				Constraints:     []string{"keep it concrete"},
				FocusSkills:     []string{"narrative clarity"},
				SuccessCriteria: []string{"stakes are clear"},
			}, nil
		},
	})

	ex := service.NextExercise(context.Background(), Context{
		CurriculumState: domain.CurriculumState{CurrentFocus: "narrative clarity"},
		RecentAssignments: []string{
			"Night Gate Ultimatum: The hero must choose between duty and family before sunrise.",
		},
	})

	if callCount != 2 {
		t.Fatalf("provider call count = %d", callCount)
	}
	if ex.Title != "Broken Treaty Meeting" {
		t.Fatalf("exercise title = %q", ex.Title)
	}
}

func TestNextExerciseDoesNotRetryWhenPatternIsDistinct(t *testing.T) {
	callCount := 0
	service := NewService(fakeLLMClient{
		generateExercise: func(ctx context.Context, input llm.ExerciseRequest) (domain.Exercise, error) {
			callCount++
			return domain.Exercise{
				Title:           "Shipyard Confession",
				Brief:           "A mechanic admits a hidden sabotage to a captain moments before inspection.",
				Constraints:     []string{"keep it concrete"},
				FocusSkills:     []string{"narrative clarity"},
				SuccessCriteria: []string{"conflict is concrete"},
			}, nil
		},
	})

	_ = service.NextExercise(context.Background(), Context{
		CurriculumState: domain.CurriculumState{CurrentFocus: "narrative clarity"},
		RecentAssignments: []string{
			"Council Vote: The city clerk has one hour to persuade a hostile council to fund flood barriers.",
		},
	})

	if callCount != 1 {
		t.Fatalf("provider call count = %d", callCount)
	}
}

func TestPatternDetectionChoiceUnderTimePressure(t *testing.T) {
	text := normalizePatternText("The hero must choose between the crown and exile before dawn.")
	if !hasChoiceUnderTimePressurePattern(text) {
		t.Fatalf("expected choice-under-time-pressure pattern for %q", text)
	}
	if !isExercisePatternRepeat(domain.Exercise{
		Title: "Choice at First Light",
		Brief: "A courier must decide between two oaths by sunrise.",
	}, []string{"City Oath: choose which gate to save before dawn."}) {
		t.Fatal("expected repeated pattern to be detected")
	}
}
