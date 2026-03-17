package prompt

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/openai"
	"github.com/tomasino/writing-coach/internal/review"
)

type deterministicGenerator struct{}

type Service struct {
	client   *openai.Client
	fallback deterministicGenerator
}

type Context struct {
	CurriculumState    domain.CurriculumState
	ActiveTGOs         []domain.TGO
	RecentTitles       []string
	RecentWeaknesses   []string
	RecurringFindings  []string
	CoachingBrief      string
	RevisionOf         *domain.Submission
	RevisionReview     *domain.Review
	RevisionComparison *review.Comparison
}

func NewService(client *openai.Client) Service {
	return Service{
		client:   client,
		fallback: deterministicGenerator{},
	}
}

func (s Service) NextExercise(ctx context.Context, input Context) domain.Exercise {
	if s.client != nil && s.client.Enabled() {
		exercise, err := s.client.GenerateExercise(ctx, openai.ExerciseRequest{
			CurrentFocus:      input.CurriculumState.CurrentFocus,
			DifficultyLevel:   input.CurriculumState.DifficultyLevel,
			ActiveTGOs:        input.ActiveTGOs,
			RecentTitles:      input.RecentTitles,
			RecentWeaknesses:  input.RecentWeaknesses,
			RecurringFindings: input.RecurringFindings,
			CoachingBrief:     input.CoachingBrief,
		})
		if err == nil {
			exercise.GenerationKind = "openai"
			return exercise
		}

		exercise = s.fallback.NextExercise(ctx, input)
		exercise.GenerationKind = "deterministic-fallback"
		exercise.ProviderNote = err.Error()
		return exercise
	}

	exercise := s.fallback.NextExercise(ctx, input)
	exercise.GenerationKind = "deterministic"
	return exercise
}

func (s Service) RevisionExercise(ctx context.Context, input Context) domain.Exercise {
	if input.RevisionOf == nil || input.RevisionReview == nil {
		return s.NextExercise(ctx, input)
	}
	if s.client != nil && s.client.Enabled() {
		exercise, err := s.client.GenerateRevisionExercise(ctx, openai.RevisionExerciseRequest{
			CurrentFocus:      input.CurriculumState.CurrentFocus,
			DifficultyLevel:   input.CurriculumState.DifficultyLevel,
			ActiveTGOs:        input.ActiveTGOs,
			SubmissionID:      input.RevisionOf.ID,
			SubmissionContent: input.RevisionOf.Content,
			Weaknesses:        input.RevisionReview.Weaknesses,
			AnalyzerFindings:  input.RevisionReview.AnalyzerFindings,
			ComparisonSummary: revisionSummary(input.RevisionComparison),
			RecentWeaknesses:  input.RecentWeaknesses,
			RecurringFindings: input.RecurringFindings,
			CoachingBrief:     input.CoachingBrief,
		})
		if err == nil {
			exercise.GenerationKind = "openai"
			exercise.SourceSubmissionID = input.RevisionOf.ID
			return exercise
		}

		exercise = s.fallback.RevisionExercise(ctx, input)
		exercise.GenerationKind = "deterministic-fallback"
		exercise.ProviderNote = err.Error()
		exercise.SourceSubmissionID = input.RevisionOf.ID
		return exercise
	}

	exercise := s.fallback.RevisionExercise(ctx, input)
	exercise.GenerationKind = "deterministic"
	exercise.SourceSubmissionID = input.RevisionOf.ID
	return exercise
}

func (deterministicGenerator) NextExercise(_ context.Context, input Context) domain.Exercise {
	focus := input.CurriculumState.CurrentFocus
	if focus == "" {
		focus = "scene architecture"
	}
	tgos := ensureDefaultTGOs(input.ActiveTGOs)

	title := fmt.Sprintf("Exercise in %s", titleCase(focus))
	if len(input.RecentTitles) > 0 {
		title = fmt.Sprintf("%s After %d Prior Trials", title, len(input.RecentTitles))
	}

	brief := fmt.Sprintf(
		"Write a new piece about %s. Show clear action, clear results, and a clear turn by the end.",
		focus,
	)
	if input.CoachingBrief != "" {
		brief += " Use this coaching goal: " + input.CoachingBrief + "."
	}
	if len(input.RecentWeaknesses) > 0 || len(input.RecurringFindings) > 0 {
		brief += " Work on the problem that showed up in recent feedback."
	}

	constraints := []string{"keep the piece small and focused", "make the main turn easy to follow", "use clear details instead of vague filler"}
	if len(input.RecurringFindings) > 0 {
		constraints = append(constraints, "avoid this repeated problem: "+input.RecurringFindings[0])
	}

	return domain.Exercise{
		Title:       title,
		Brief:       brief + " TGOs: " + strings.Join(tgoTitles(tgos), "; "),
		Constraints: constraints,
		FocusSkills: tgoSkills(tgos),
		TGOCodes:    tgoCodes(tgos),
		SuccessCriteria: []string{
			"the piece shows a clear change from start to finish",
			"the setting comes through action and image",
			"the ending feels clear and finished",
		},
	}
}

func (deterministicGenerator) RevisionExercise(_ context.Context, input Context) domain.Exercise {
	focus := input.CurriculumState.CurrentFocus
	if focus == "" {
		focus = "prose precision"
	}
	tgos := ensureDefaultTGOs(input.ActiveTGOs)
	weaknesses := []string{}
	findings := []string{}
	if input.RevisionReview != nil {
		weaknesses = input.RevisionReview.Weaknesses
		findings = input.RevisionReview.AnalyzerFindings
	}
	title := fmt.Sprintf("Revision of Draft %d in %s", input.RevisionOf.DraftNumber, titleCase(focus))
	brief := fmt.Sprintf(
		"Revise your existing draft rather than replacing it. Preserve the core dramatic event, but rewrite for %s with sharper causality, cleaner prose pressure, and more concrete consequence.",
		focus,
	)
	if input.CoachingBrief != "" {
		brief += " Coaching context: " + input.CoachingBrief + "."
	}
	if input.RevisionComparison != nil && input.RevisionComparison.Summary != "" {
		brief += " Comparison note: " + input.RevisionComparison.Summary
	}

	constraints := []string{
		"keep the same central scene and ending decision",
		"revise at the sentence and beat level rather than expanding lore",
		"make at least one previously vague consequence concrete on the page",
	}
	if len(weaknesses) > 0 {
		constraints = append(constraints, "directly address this weakness: "+weaknesses[0])
	}
	if len(findings) > 0 {
		constraints = append(constraints, "eliminate or reduce this analyzer issue: "+findings[0])
	}

	success := []string{
		"the revised draft makes the causal chain easier to follow",
		"the prose is tighter and less hedged",
		"the emotional cost is more concrete without overexplaining symbolism",
	}
	if input.RevisionComparison != nil && len(input.RevisionComparison.PersistingWeaknesses) > 0 {
		success = append(success, "the prior persistent weakness is materially reduced")
	}

	return domain.Exercise{
		Title:           title,
		Brief:           brief,
		Constraints:     constraints,
		FocusSkills:     tgoSkills(tgos),
		TGOCodes:        tgoCodes(tgos),
		SuccessCriteria: success,
	}
}

func revisionSummary(c *review.Comparison) string {
	if c == nil {
		return ""
	}
	parts := []string{c.Summary}
	if len(c.PersistingWeaknesses) > 0 {
		parts = append(parts, "Persisting: "+c.PersistingWeaknesses[0])
	}
	if len(c.AddressedWeaknesses) > 0 {
		parts = append(parts, "Addressed: "+c.AddressedWeaknesses[0])
	}
	return strings.Join(parts, " ")
}

func ensureDefaultTGOs(active []domain.TGO) []domain.TGO {
	if len(active) == 3 {
		return active
	}
	var out []domain.TGO
	for _, code := range []string{"causal-clarity", "scene-architecture", "prose-precision"} {
		if tgo, ok := domain.TGOByCode(code); ok {
			out = append(out, tgo)
		}
	}
	return out
}

func tgoCodes(tgos []domain.TGO) []string {
	var out []string
	for _, tgo := range tgos {
		out = append(out, tgo.Code)
	}
	return out
}

func tgoTitles(tgos []domain.TGO) []string {
	var out []string
	for _, tgo := range tgos {
		out = append(out, tgo.Title)
	}
	return out
}

func tgoSkills(tgos []domain.TGO) []string {
	seen := map[string]bool{}
	var out []string
	for _, tgo := range tgos {
		skill := domain.TGOCodeToSkill[tgo.Code]
		if skill == "" || seen[skill] {
			continue
		}
		seen[skill] = true
		out = append(out, skill)
	}
	return out
}

func titleCase(value string) string {
	words := strings.Fields(value)
	for i, word := range words {
		runes := []rune(word)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
