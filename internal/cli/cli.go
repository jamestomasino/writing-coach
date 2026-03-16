package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/prompt"
	"github.com/tomasino/writing-coach/internal/review"
)

type CLI struct {
	Config     config.Config
	Store      *db.Store
	Prompts    prompt.Service
	Reviews    review.Service
	Curriculum curriculum.Service
}

func (c CLI) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.usage()
	}

	switch args[0] {
	case "init":
		return c.runInit(ctx)
	case "prompt":
		return c.runPrompt(ctx, args[1:])
	case "submit":
		return c.runSubmit(ctx, args[1:])
	case "review":
		return c.runReview(ctx, args[1:])
	case "history":
		return c.runHistory(ctx)
	case "progress":
		return c.runProgress(ctx)
	default:
		return c.usage()
	}
}

func (c CLI) usage() error {
	fmt.Println("usage:")
	fmt.Println("  writing-coach init")
	fmt.Println("  writing-coach prompt next")
	fmt.Println("  writing-coach submit --exercise <id> --file <path>")
	fmt.Println("  writing-coach review --submission <id>")
	fmt.Println("  writing-coach history")
	fmt.Println("  writing-coach progress")
	return nil
}

func (c CLI) runInit(ctx context.Context) error {
	if err := config.Save(c.Config); err != nil {
		return err
	}
	if err := c.Store.Migrate(ctx, filepath.Join(c.Config.ProjectRoot, "migrations")); err != nil {
		return err
	}
	if err := c.Store.EnsureSeedData(ctx, c.Config.WriterName); err != nil {
		return err
	}

	fmt.Printf("initialized project state in %s\n", c.Config.DataDir)
	if c.Config.OpenAIAPIKey == "" {
		fmt.Println("openai: disabled (set OPENAI_API_KEY to enable model-backed prompt and review generation)")
	} else {
		fmt.Printf("openai: enabled (prompt model: %s, review model: %s)\n", c.Config.PromptModel, c.Config.ReviewModel)
	}
	if c.Config.ValeBinary == "" {
		fmt.Println("vale: auto-detect (set VALE_BINARY to override; uses repo .vale.ini when available)")
	} else {
		fmt.Printf("vale: enabled (%s)\n", c.Config.ValeBinary)
	}
	if c.Config.LanguageToolURL == "" {
		fmt.Println("languagetool: disabled (set LANGUAGETOOL_URL to enable LanguageTool checks)")
	} else {
		fmt.Printf("languagetool: enabled (%s)\n", c.Config.LanguageToolURL)
	}
	return nil
}

func (c CLI) runPrompt(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "next" {
		return errors.New("usage: writing-coach prompt next")
	}

	state, err := c.Store.GetCurriculumState(ctx)
	if err != nil {
		return err
	}

	recentTitles, err := c.Store.RecentExerciseTitles(ctx, 3)
	if err != nil {
		return err
	}

	ex := c.Prompts.NextExercise(ctx, prompt.Context{
		CurriculumState: state,
		RecentTitles:    recentTitles,
	})
	exerciseID, err := c.Store.SaveExercise(ctx, ex)
	if err != nil {
		return err
	}

	fmt.Printf("exercise %d\n", exerciseID)
	fmt.Printf("provider: %s\n", ex.GenerationKind)
	if ex.ProviderNote != "" {
		fmt.Printf("provider note: %s\n", ex.ProviderNote)
	}
	fmt.Printf("title: %s\n", ex.Title)
	fmt.Printf("brief: %s\n", ex.Brief)
	fmt.Printf("constraints: %s\n", strings.Join(ex.Constraints, "; "))
	fmt.Printf("focus skills: %s\n", strings.Join(ex.FocusSkills, ", "))
	fmt.Printf("success criteria: %s\n", strings.Join(ex.SuccessCriteria, "; "))
	return nil
}

func (c CLI) runSubmit(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("submit", flag.ContinueOnError)
	exerciseID := fs.Int64("exercise", 0, "exercise id")
	filePath := fs.String("file", "", "path to submission text")
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *exerciseID == 0 || *filePath == "" {
		return errors.New("usage: writing-coach submit --exercise <id> --file <path>")
	}

	content, err := os.ReadFile(*filePath)
	if err != nil {
		return err
	}

	sub := domain.Submission{
		ExerciseID: *exerciseID,
		Content:    string(content),
		WordCount:  db.CountWords(string(content)),
	}
	submissionID, err := c.Store.SaveSubmission(ctx, sub)
	if err != nil {
		return err
	}

	fmt.Printf("submission %d saved with %d words\n", submissionID, sub.WordCount)
	return nil
}

func (c CLI) runReview(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)
	submissionID := fs.Int64("submission", 0, "submission id")
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *submissionID == 0 {
		return errors.New("usage: writing-coach review --submission <id>")
	}

	sub, err := c.Store.GetSubmission(ctx, *submissionID)
	if err != nil {
		if db.IsNotFound(err) {
			return fmt.Errorf("submission %d not found", *submissionID)
		}
		return err
	}

	reviewResult, scores := c.Reviews.ReviewSubmission(ctx, sub)
	currentState, err := c.Store.GetCurriculumState(ctx)
	if err != nil {
		return err
	}
	recommendation, err := c.Curriculum.RecommendNextFocus(ctx, c.Store, currentState, reviewResult.NextFocus, scores)
	if err != nil {
		return err
	}
	reviewResult.NextFocus = recommendation.Focus
	reviewID, err := c.Store.SaveReview(ctx, reviewResult, scores)
	if err != nil {
		return err
	}
	if err := c.Store.UpdateCurriculumState(ctx, recommendation.Focus, recommendation.Difficulty, reviewID); err != nil {
		return err
	}

	state, err := c.Store.GetCurriculumState(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("review %d\n", reviewID)
	fmt.Printf("provider: %s\n", reviewResult.ReviewKind)
	if reviewResult.ProviderNote != "" {
		fmt.Printf("provider note: %s\n", reviewResult.ProviderNote)
	}
	fmt.Printf("summary: %s\n", reviewResult.Summary)
	fmt.Printf("strengths: %s\n", strings.Join(reviewResult.Strengths, "; "))
	fmt.Printf("weaknesses: %s\n", strings.Join(reviewResult.Weaknesses, "; "))
	if len(reviewResult.AnalyzerFindings) > 0 {
		fmt.Printf("analyzer findings: %s\n", strings.Join(reviewResult.AnalyzerFindings, "; "))
	}
	if len(scores) > 0 {
		var parts []string
		for _, score := range scores {
			parts = append(parts, fmt.Sprintf("%s=%d", score.Skill, score.Score))
		}
		fmt.Printf("skill scores: %s\n", strings.Join(parts, ", "))
	}
	fmt.Printf("next focus: %s\n", reviewResult.NextFocus)
	fmt.Printf("focus rationale: %s\n", recommendation.Rationale)
	fmt.Printf("curriculum: %s\n", c.Curriculum.DescribeNextStep(state))
	return nil
}

func (c CLI) runHistory(ctx context.Context) error {
	state, err := c.Store.GetCurriculumState(ctx)
	if err != nil {
		return err
	}
	items, err := c.Store.History(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("current focus: %s\n", state.CurrentFocus)
	fmt.Printf("difficulty level: %d\n", state.DifficultyLevel)
	if len(items) == 0 {
		fmt.Println("no exercises yet")
		return nil
	}
	for _, item := range items {
		fmt.Println(item)
	}
	return nil
}

func (c CLI) runProgress(ctx context.Context) error {
	state, err := c.Store.GetCurriculumState(ctx)
	if err != nil {
		return err
	}
	items, err := c.Store.ProgressReport(ctx, 5)
	if err != nil {
		return err
	}
	fmt.Printf("track: %s\n", domain.WriterTrackName)
	fmt.Printf("current focus: %s\n", state.CurrentFocus)
	if len(items) == 0 {
		fmt.Println("no progress data yet")
		return nil
	}
	for _, item := range items {
		fmt.Println(item)
	}
	return nil
}
