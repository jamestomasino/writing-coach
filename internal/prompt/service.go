package prompt

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/llm"
	"github.com/tomasino/writing-coach/internal/review"
)

type deterministicGenerator struct{}

type Service struct {
	client            llm.Client
	clientKind        string
	fallback          deterministicGenerator
	generationTimeout time.Duration
}

const defaultGenerationTimeout = 20 * time.Second

type Context struct {
	CurriculumState    domain.CurriculumState
	ActiveTGOs         []domain.TGO
	OnboardingProfile  *domain.OnboardingProfile
	RecentTitles       []string
	RecentAssignments  []string
	RecentWeaknesses   []string
	RecurringFindings  []string
	CoachingBrief      string
	RevisionOf         *domain.Submission
	RevisionReview     *domain.Review
	RevisionComparison *review.Comparison
}

func NewService(client llm.Client) Service {
	return Service{
		client:            client,
		clientKind:        "openai",
		fallback:          deterministicGenerator{},
		generationTimeout: defaultGenerationTimeout,
	}
}

func (s Service) WithClient(client llm.Client, kind string) Service {
	s.client = client
	s.clientKind = strings.TrimSpace(kind)
	if s.clientKind == "" {
		s.clientKind = "openai"
	}
	return s
}

func (s Service) WithGenerationTimeout(timeout time.Duration) Service {
	s.generationTimeout = timeout
	return s
}

func (s Service) NextExercise(ctx context.Context, input Context) domain.Exercise {
	if s.client != nil && s.client.Enabled() {
		generationCtx, cancel := s.generationContext(ctx)
		defer cancel()
		request := llm.ExerciseRequest{
			WritingLanguage:   writingLanguageForProfile(input.OnboardingProfile),
			CurrentFocus:      input.CurriculumState.CurrentFocus,
			DifficultyLevel:   input.CurriculumState.DifficultyLevel,
			ActiveTGOs:        input.ActiveTGOs,
			OnboardingProfile: input.OnboardingProfile,
			RecentTitles:      input.RecentTitles,
			RecentAssignments: input.RecentAssignments,
			RecentWeaknesses:  input.RecentWeaknesses,
			RecurringFindings: input.RecurringFindings,
			CoachingBrief:     input.CoachingBrief,
		}
		exercise, err := s.client.GenerateExercise(generationCtx, request)
		if err == nil {
			if !isExercisePatternRepeat(exercise, input.RecentAssignments) {
				exercise.GenerationKind = s.clientKind
				exercise.ProviderNote = s.clientKind
				return exercise
			}
			retryRequest := request
			retryRequest.CoachingBrief = appendVarietyGuidance(request.CoachingBrief, input.RecentAssignments)
			retryExercise, retryErr := s.client.GenerateExercise(generationCtx, retryRequest)
			if retryErr == nil {
				retryExercise.GenerationKind = s.clientKind
				retryExercise.ProviderNote = s.clientKind
				return retryExercise
			}
			exercise.GenerationKind = "deterministic-fallback"
			exercise.ProviderNote = strings.TrimSpace(s.clientKind + ": " + retryErr.Error())
			return exercise
		}

		exercise = s.fallback.NextExercise(ctx, input)
		exercise.GenerationKind = "deterministic-fallback"
		exercise.ProviderNote = strings.TrimSpace(s.clientKind + ": " + err.Error())
		return exercise
	}

	exercise := s.fallback.NextExercise(ctx, input)
	exercise.GenerationKind = "deterministic"
	return exercise
}

func appendVarietyGuidance(existing string, recent []string) string {
	base := strings.TrimSpace(existing)
	history := joinOrDefault(recent, "none")
	note := "Variety requirement: choose a different core scenario pattern than recent assignments. Avoid repeating the same premise structure from this history: " + history
	if base == "" {
		return note
	}
	return base + " " + note
}

func joinOrDefault(values []string, fallback string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) == 0 {
		return fallback
	}
	return strings.Join(filtered, "; ")
}

func isExercisePatternRepeat(exercise domain.Exercise, recent []string) bool {
	if len(recent) == 0 {
		return false
	}
	candidateText := normalizePatternText(exercise.Title + " " + exercise.Brief)
	if candidateText == "" {
		return false
	}
	if hasChoiceUnderTimePressurePattern(candidateText) {
		for _, item := range recent {
			if hasChoiceUnderTimePressurePattern(normalizePatternText(item)) {
				return true
			}
		}
	}

	candidateTokens := tokenSet(candidateText)
	if len(candidateTokens) == 0 {
		return false
	}
	candidateBigrams := ngrams(candidateText, 2)
	for _, item := range recent {
		other := normalizePatternText(item)
		if other == "" {
			continue
		}
		similarity := jaccard(candidateTokens, tokenSet(other))
		if similarity >= 0.58 {
			return true
		}
		if overlapCount(candidateBigrams, ngrams(other, 2)) >= 3 && similarity >= 0.42 {
			return true
		}
	}
	return false
}

func hasChoiceUnderTimePressurePattern(text string) bool {
	if text == "" {
		return false
	}
	choiceWords := []string{"choose", "choice", "decide", "decision", "pick", "must choose", "must decide", "forced to choose"}
	timeWords := []string{"time pressure", "deadline", "before", "countdown", "minutes", "hours", "clock", "urgent", "by dawn", "by sunrise", "tonight"}
	hasChoice := false
	for _, word := range choiceWords {
		if strings.Contains(text, word) {
			hasChoice = true
			break
		}
	}
	if !hasChoice {
		return false
	}
	for _, word := range timeWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

func normalizePatternText(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(' ')
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func tokenSet(value string) map[string]struct{} {
	parts := strings.Fields(value)
	out := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		if len(part) <= 2 || isStopWord(part) {
			continue
		}
		out[part] = struct{}{}
	}
	return out
}

func ngrams(value string, size int) map[string]struct{} {
	if size <= 0 {
		return map[string]struct{}{}
	}
	parts := strings.Fields(value)
	if len(parts) < size {
		return map[string]struct{}{}
	}
	out := make(map[string]struct{}, len(parts)-size+1)
	for i := 0; i+size <= len(parts); i++ {
		window := parts[i : i+size]
		out[strings.Join(window, " ")] = struct{}{}
	}
	return out
}

func overlapCount(a, b map[string]struct{}) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	count := 0
	for key := range a {
		if _, ok := b[key]; ok {
			count++
		}
	}
	return count
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for key := range a {
		if _, ok := b[key]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return math.Abs(float64(intersection) / float64(union))
}

func isStopWord(token string) bool {
	switch token {
	case "the", "and", "for", "with", "from", "this", "that", "your", "into", "about", "after", "under", "when", "while", "then", "than":
		return true
	default:
		return false
	}
}

func (s Service) RevisionExercise(ctx context.Context, input Context) domain.Exercise {
	if input.RevisionOf == nil || input.RevisionReview == nil {
		return s.NextExercise(ctx, input)
	}
	if s.client != nil && s.client.Enabled() {
		generationCtx, cancel := s.generationContext(ctx)
		defer cancel()
		exercise, err := s.client.GenerateRevisionExercise(generationCtx, llm.RevisionExerciseRequest{
			WritingLanguage:   writingLanguageForProfile(input.OnboardingProfile),
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
			exercise.GenerationKind = s.clientKind
			exercise.ProviderNote = s.clientKind
			exercise.SourceSubmissionID = input.RevisionOf.ID
			return exercise
		}

		exercise = s.fallback.RevisionExercise(ctx, input)
		exercise.GenerationKind = "deterministic-fallback"
		exercise.ProviderNote = strings.TrimSpace(s.clientKind + ": " + err.Error())
		exercise.SourceSubmissionID = input.RevisionOf.ID
		return exercise
	}

	exercise := s.fallback.RevisionExercise(ctx, input)
	exercise.GenerationKind = "deterministic"
	exercise.SourceSubmissionID = input.RevisionOf.ID
	return exercise
}

func (s Service) generationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	base := context.WithoutCancel(ctx)
	timeout := s.generationTimeout
	if timeout <= 0 {
		return base, func() {}
	}
	return context.WithTimeout(base, timeout)
}

func (deterministicGenerator) NextExercise(_ context.Context, input Context) domain.Exercise {
	tgos := ensureDefaultTGOs(input.ActiveTGOs)
	profile := input.OnboardingProfile
	title := "New Writing Assignment"
	brief := "Write a new piece from scratch."
	constraints := []string{"keep the piece small and focused", "make the main turn easy to follow", "use clear details instead of vague filler"}
	success := []string{
		"the piece fits the assignment format",
		"the draft stays clear from start to finish",
		"the ending feels complete and intentional",
	}

	if profile != nil {
		title = deterministicTitle(*profile)
		brief = deterministicBrief(*profile)
		constraints = deterministicConstraints(*profile)
		success = deterministicSuccessCriteria(*profile)
	} else {
		focus := input.CurriculumState.CurrentFocus
		if focus == "" {
			focus = "scene architecture"
		}
		title = fmt.Sprintf("Exercise in %s", titleCase(focus))
		brief = fmt.Sprintf(
			"Write a new piece about %s. Show clear action, clear results, and a clear turn by the end.",
			focus,
		)
	}
	if len(input.RecentTitles) > 0 {
		title = fmt.Sprintf("%s After %d Prior Trials", title, len(input.RecentTitles))
	}
	if input.CoachingBrief != "" {
		brief += " Track context: " + input.CoachingBrief + "."
	}
	if len(input.RecentWeaknesses) > 0 || len(input.RecurringFindings) > 0 {
		brief += " Work on the problem that showed up in recent feedback."
	}
	if len(input.RecurringFindings) > 0 {
		constraints = append(constraints, "avoid this repeated problem: "+input.RecurringFindings[0])
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
	for _, code := range []string{"story-causal-clarity", "story-scene-architecture", "story-prose-precision"} {
		if tgo, ok := domain.TGOByCode(code); ok {
			out = append(out, tgo)
		}
	}
	return out
}

func writingLanguageForProfile(profile *domain.OnboardingProfile) string {
	if profile == nil {
		return domain.DefaultWritingLanguage
	}
	return domain.NormalizeWritingLanguage(profile.WritingLanguage)
}

func tgoCodes(tgos []domain.TGO) []string {
	var out []string
	for _, tgo := range tgos {
		out = append(out, tgo.Code)
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

func deterministicTitle(profile domain.OnboardingProfile) string {
	format := strings.TrimSpace(profile.AssignmentFormat)
	subject := strings.TrimSpace(profile.SubjectMatter)
	if format == "" {
		format = "writing piece"
	}
	if subject == "" {
		return "New " + titleCase(format)
	}
	return fmt.Sprintf("%s on %s", titleCase(format), titleCase(subject))
}

func deterministicBrief(profile domain.OnboardingProfile) string {
	format := fallbackText(profile.AssignmentFormat, "piece")
	audience := fallbackText(profile.TargetAudience, "your intended audience")
	subject := fallbackText(profile.SubjectMatter, "a topic that fits your track")
	tone := domain.PromptToneGuidance(profile)
	brief := fmt.Sprintf("Write a new %s for %s, using %s as the core situation.", format, audience, subject)
	if scenario := domain.PromptScenarioGuidance(profile); scenario != "" {
		parts := strings.Split(scenario, ";")
		if len(parts) > 1 {
			brief += " " + strings.TrimSpace(parts[len(parts)-1]) + "."
		}
	}
	if tone != "" {
		brief += " Tone guidance: " + tone + "."
	}
	return brief
}

func deterministicConstraints(profile domain.OnboardingProfile) []string {
	format := fallbackText(profile.AssignmentFormat, "piece")
	domainLabel := fallbackText(profile.WritingType, "track")
	constraints := []string{
		"stay inside the chosen format: " + format,
		"write for this audience: " + fallbackText(profile.TargetAudience, "your intended audience"),
		"use details that fit this writing domain: " + domainLabel,
	}
	if tone := domain.PromptToneGuidance(profile); tone != "" {
		constraints = append(constraints, tone)
	}
	return constraints
}

func deterministicSuccessCriteria(profile domain.OnboardingProfile) []string {
	success := []string{
		"the draft clearly fits the assignment format",
		"the piece feels written for the target audience",
		"the subject matter stays consistent from start to finish",
	}
	if goal := firstNonEmptyString(profile.WritingGoals, firstItem(profile.DesiredOutcomes)); goal != "" {
		success = append(success, "the draft supports this track goal: "+goal)
	}
	return success
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstItem(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
