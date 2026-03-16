package prompt

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/openai"
)

type deterministicGenerator struct{}

type Service struct {
	client   *openai.Client
	fallback deterministicGenerator
}

type Context struct {
	CurriculumState domain.CurriculumState
	RecentTitles    []string
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
			CurrentFocus:    input.CurriculumState.CurrentFocus,
			DifficultyLevel: input.CurriculumState.DifficultyLevel,
			RecentTitles:    input.RecentTitles,
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

func (deterministicGenerator) NextExercise(_ context.Context, input Context) domain.Exercise {
	focus := input.CurriculumState.CurrentFocus
	if focus == "" {
		focus = "scene architecture"
	}

	title := fmt.Sprintf("Exercise in %s", titleCase(focus))
	if len(input.RecentTitles) > 0 {
		title = fmt.Sprintf("%s After %d Prior Trials", title, len(input.RecentTitles))
	}

	brief := fmt.Sprintf(
		"Write 700-1000 words of mythopoeic tragic fantasy centered on %s. Build pressure through implication rather than exposition, and end on an irreversible moral or emotional turn.",
		focus,
	)

	return domain.Exercise{
		Title:       title,
		Brief:       brief,
		Constraints: []string{"third-person limited", "single scene", "one concrete symbol that changes meaning by the end"},
		FocusSkills: []string{focus, "narrative clarity"},
		SuccessCriteria: []string{
			"the scene carries a clear emotional progression",
			"worldbuilding is implied through action and image",
			"the ending closes a door rather than opening one",
		},
	}
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
