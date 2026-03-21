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
	"github.com/tomasino/writing-coach/internal/session"
)

type CLI struct {
	Config     config.Config
	AppContext session.Context
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
	case "compare":
		return c.runCompare(ctx, args[1:])
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
	fmt.Println("  writing-coach serve")
	fmt.Println("  writing-coach prompt next")
	fmt.Println("  writing-coach prompt revise --submission <id>")
	fmt.Println("  writing-coach submit --exercise <id> --file <path> [--revise-from <submission-id>]")
	fmt.Println("  writing-coach review --submission <id>")
	fmt.Println("  writing-coach compare --submission <id> [--against <submission-id>]")
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
		fmt.Println("system openai fallback: disabled (set OPENAI_API_KEY to enable a shared provider fallback)")
	} else {
		fmt.Printf("system openai fallback: enabled (prompt model: %s, review model: %s)\n", c.Config.PromptModel, c.Config.ReviewModel)
	}
	if c.Config.AIKeySecret == "" {
		fmt.Println("ai key encryption: disabled (set WRITING_COACH_AI_KEY_SECRET before using personal provider keys)")
	} else {
		fmt.Println("ai key encryption: enabled")
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
	if c.Config.NLPAnalyzerURL == "" {
		fmt.Println("nlp analyzer: disabled (set WRITING_COACH_NLP_ANALYZER_URL to enable spaCy/TextDescriptives checks)")
	} else {
		fmt.Printf("nlp analyzer: enabled (%s)\n", c.Config.NLPAnalyzerURL)
	}
	return nil
}

func (c CLI) runPrompt(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: writing-coach prompt next | prompt revise --submission <id>")
	}
	switch args[0] {
	case "next":
		return c.runPromptNext(ctx)
	case "revise":
		return c.runPromptRevise(ctx, args[1:])
	default:
		return errors.New("usage: writing-coach prompt next | prompt revise --submission <id>")
	}
}

func (c CLI) runPromptNext(ctx context.Context) error {
	state, err := c.Store.GetCurriculumState(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}

	recentTitles, err := c.Store.RecentExerciseTitles(ctx, c.AppContext.UserID, c.AppContext.TreeID, 3)
	if err != nil {
		return err
	}
	recentWeaknesses, err := c.Store.RecurringWeaknesses(ctx, c.AppContext.UserID, c.AppContext.TreeID, 5)
	if err != nil {
		return err
	}
	recurringFindings, err := c.Store.RecurringAnalyzerFindings(ctx, c.AppContext.UserID, c.AppContext.TreeID, 5)
	if err != nil {
		return err
	}
	activeTGOs, err := c.Store.ActiveTGOs(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}

	ex := c.Prompts.NextExercise(ctx, prompt.Context{
		CurriculumState:   state,
		ActiveTGOs:        activeTGOs,
		RecentTitles:      recentTitles,
		RecentWeaknesses:  recentWeaknesses,
		RecurringFindings: recurringFindings,
	})
	ex.UserID = c.AppContext.UserID
	ex.TreeID = c.AppContext.TreeID
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
	if len(activeTGOs) > 0 {
		var tgoLines []string
		for _, tgo := range activeTGOs {
			tgoLines = append(tgoLines, fmt.Sprintf("%s: %s", tgo.Code, tgo.Title))
		}
		fmt.Printf("tgos: %s\n", strings.Join(tgoLines, "; "))
	}
	fmt.Printf("constraints: %s\n", strings.Join(ex.Constraints, "; "))
	fmt.Printf("focus skills: %s\n", strings.Join(ex.FocusSkills, ", "))
	fmt.Printf("success criteria: %s\n", strings.Join(ex.SuccessCriteria, "; "))
	return nil
}

func (c CLI) runPromptRevise(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("prompt revise", flag.ContinueOnError)
	submissionID := fs.Int64("submission", 0, "submission id to revise")
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *submissionID == 0 {
		return errors.New("usage: writing-coach prompt revise --submission <id>")
	}

	sub, err := c.Store.GetSubmission(ctx, *submissionID)
	if err != nil {
		return err
	}
	reviewResult, err := c.Store.LatestReviewForSubmission(ctx, sub.ID)
	if err != nil {
		return fmt.Errorf("submission %d has no review yet", sub.ID)
	}
	state, err := c.Store.GetCurriculumState(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}
	recentTitles, err := c.Store.RecentExerciseTitles(ctx, c.AppContext.UserID, c.AppContext.TreeID, 3)
	if err != nil {
		return err
	}
	recentWeaknesses, err := c.Store.RecurringWeaknesses(ctx, c.AppContext.UserID, c.AppContext.TreeID, 5)
	if err != nil {
		return err
	}
	recurringFindings, err := c.Store.RecurringAnalyzerFindings(ctx, c.AppContext.UserID, c.AppContext.TreeID, 5)
	if err != nil {
		return err
	}
	activeTGOs, err := c.Store.ActiveTGOs(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}

	var cmp *review.Comparison
	if previous, err := c.Store.PreviousSubmission(ctx, sub); err == nil {
		if previousReview, err := c.Store.LatestReviewForSubmission(ctx, previous.ID); err == nil {
			comparison := review.CompareSubmissions(sub, previous, reviewResult, previousReview)
			cmp = &comparison
		}
	}

	ex := c.Prompts.RevisionExercise(ctx, prompt.Context{
		CurriculumState:    state,
		ActiveTGOs:         activeTGOs,
		RecentTitles:       recentTitles,
		RecentWeaknesses:   recentWeaknesses,
		RecurringFindings:  recurringFindings,
		RevisionOf:         &sub,
		RevisionReview:     &reviewResult,
		RevisionComparison: cmp,
	})
	ex.UserID = c.AppContext.UserID
	ex.TreeID = c.AppContext.TreeID
	exerciseID, err := c.Store.SaveExercise(ctx, ex)
	if err != nil {
		return err
	}

	fmt.Printf("exercise %d\n", exerciseID)
	fmt.Printf("provider: %s\n", ex.GenerationKind)
	if ex.ProviderNote != "" {
		fmt.Printf("provider note: %s\n", ex.ProviderNote)
	}
	fmt.Printf("revision of submission: %d\n", sub.ID)
	fmt.Printf("title: %s\n", ex.Title)
	fmt.Printf("brief: %s\n", ex.Brief)
	if len(activeTGOs) > 0 {
		var tgoLines []string
		for _, tgo := range activeTGOs {
			tgoLines = append(tgoLines, fmt.Sprintf("%s: %s", tgo.Code, tgo.Title))
		}
		fmt.Printf("tgos: %s\n", strings.Join(tgoLines, "; "))
	}
	fmt.Printf("constraints: %s\n", strings.Join(ex.Constraints, "; "))
	fmt.Printf("focus skills: %s\n", strings.Join(ex.FocusSkills, ", "))
	fmt.Printf("success criteria: %s\n", strings.Join(ex.SuccessCriteria, "; "))
	return nil
}

func (c CLI) runSubmit(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("submit", flag.ContinueOnError)
	exerciseID := fs.Int64("exercise", 0, "exercise id")
	filePath := fs.String("file", "", "path to submission text")
	reviseFrom := fs.Int64("revise-from", 0, "prior submission id to revise")
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
		UserID:             c.AppContext.UserID,
		TreeID:             c.AppContext.TreeID,
		ExerciseID:         *exerciseID,
		ParentSubmissionID: *reviseFrom,
		Content:            string(content),
		WordCount:          db.CountWords(string(content)),
	}
	submissionID, err := c.Store.SaveSubmission(ctx, sub)
	if err != nil {
		return err
	}
	saved, err := c.Store.GetSubmission(ctx, submissionID)
	if err != nil {
		return err
	}

	fmt.Printf("submission %d saved with %d words\n", submissionID, saved.WordCount)
	fmt.Printf("draft number: %d\n", saved.DraftNumber)
	if saved.ParentSubmissionID != 0 {
		fmt.Printf("revises submission: %d\n", saved.ParentSubmissionID)
	}
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

	activeTGOs, err := c.Store.ActiveTGOs(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}
	completedTGOs, err := c.Store.CompletedTGOs(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}
	reviewResult, scores := c.Reviews.ReviewSubmission(ctx, sub, activeTGOs, completedTGOs)
	reviewResult.UserID = c.AppContext.UserID
	reviewResult.TreeID = c.AppContext.TreeID
	recommendation, err := c.Curriculum.SyncTGOs(ctx, c.Store, c.AppContext.TreeSlug, c.AppContext.EnrollmentID, reviewResult)
	if err != nil {
		return err
	}
	reviewResult.NextFocus = recommendation.Focus
	reviewID, err := c.Store.SaveReview(ctx, reviewResult, scores)
	if err != nil {
		return err
	}
	if err := c.Store.UpdateCurriculumState(ctx, c.AppContext.EnrollmentID, recommendation.Focus, recommendation.Difficulty, reviewID); err != nil {
		return err
	}

	state, err := c.Store.GetCurriculumState(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}

	fmt.Printf("review %d\n", reviewID)
	fmt.Printf("provider: %s\n", reviewResult.ReviewKind)
	if reviewResult.ProviderNote != "" {
		fmt.Printf("provider note: %s\n", reviewResult.ProviderNote)
	}
	fmt.Printf("summary: %s\n", reviewResult.Summary)
	if previous, err := c.Store.PreviousSubmission(ctx, sub); err == nil {
		fmt.Printf("revision delta: draft %d -> %d | word count %+d\n", previous.DraftNumber, sub.DraftNumber, sub.WordCount-previous.WordCount)
		if previousReview, err := c.Store.LatestReviewForSubmission(ctx, previous.ID); err == nil {
			comparison := review.CompareSubmissions(sub, previous, reviewResult, previousReview)
			fmt.Printf("revision summary: %s\n", comparison.Summary)
			if len(comparison.AddressedWeaknesses) > 0 {
				fmt.Printf("addressed weaknesses: %s\n", strings.Join(comparison.AddressedWeaknesses, "; "))
			}
			if len(comparison.PersistingWeaknesses) > 0 {
				fmt.Printf("persisting weaknesses: %s\n", strings.Join(comparison.PersistingWeaknesses, "; "))
			}
		}
	}
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
	if len(reviewResult.TGOAssessments) > 0 {
		var parts []string
		for _, assessment := range reviewResult.TGOAssessments {
			parts = append(parts, fmt.Sprintf("%s=%s", assessment.TGOCode, assessment.Status))
		}
		fmt.Printf("tgo assessments: %s\n", strings.Join(parts, ", "))
	}
	if len(reviewResult.CompletedTGOChecks) > 0 {
		var parts []string
		for _, check := range reviewResult.CompletedTGOChecks {
			parts = append(parts, fmt.Sprintf("%s=%s", check.TGOCode, check.Status))
		}
		fmt.Printf("completed tgo checks: %s\n", strings.Join(parts, ", "))
	}
	fmt.Printf("next focus: %s\n", reviewResult.NextFocus)
	fmt.Printf("focus rationale: %s\n", recommendation.Rationale)
	fmt.Printf("curriculum: %s\n", c.Curriculum.DescribeNextStep(state))
	return nil
}

func (c CLI) runHistory(ctx context.Context) error {
	state, err := c.Store.GetCurriculumState(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}
	items, err := c.Store.History(ctx, c.AppContext.UserID, c.AppContext.TreeID)
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
	state, err := c.Store.GetCurriculumState(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}
	treeDef, err := c.Store.TreeDefinitionBySlug(ctx, c.AppContext.TreeSlug)
	if err != nil {
		return err
	}
	items, err := c.Store.ProgressReport(ctx, c.AppContext.UserID, c.AppContext.TreeID, treeDef.PrioritySkills, 5)
	if err != nil {
		return err
	}
	strongest, weakest, err := c.Store.StrongestWeakestSkills(ctx, c.AppContext.UserID, c.AppContext.TreeID, treeDef.PrioritySkills, 5)
	if err != nil {
		return err
	}
	recurringWeaknesses, err := c.Store.RecurringWeaknesses(ctx, c.AppContext.UserID, c.AppContext.TreeID, 5)
	if err != nil {
		return err
	}
	recurringFindings, err := c.Store.RecurringAnalyzerFindings(ctx, c.AppContext.UserID, c.AppContext.TreeID, 5)
	if err != nil {
		return err
	}
	recurringSlips, err := c.Store.RecurringCompletedTGOSlips(ctx, c.AppContext.UserID, c.AppContext.TreeID, 5)
	if err != nil {
		return err
	}
	activeTGOs, err := c.Store.ActiveTGOs(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}
	completedTGOs, err := c.Store.CompletedTGOs(ctx, c.AppContext.EnrollmentID)
	if err != nil {
		return err
	}
	completedSet := map[string]bool{}
	for _, tgo := range completedTGOs {
		completedSet[tgo.Code] = true
	}
	activeSet := map[string]bool{}
	for _, tgo := range activeTGOs {
		activeSet[tgo.Code] = true
	}
	upcomingTGOs := domain.NextUnlockedFromDefinition(treeDef, completedSet, activeSet, 3)
	fmt.Printf("user: %s\n", c.AppContext.UserSlug)
	fmt.Printf("tree: %s\n", c.AppContext.TreeSlug)
	fmt.Printf("current focus: %s\n", state.CurrentFocus)
	if len(activeTGOs) > 0 {
		var parts []string
		for _, tgo := range activeTGOs {
			parts = append(parts, fmt.Sprintf("[%d] %s", tgo.ActiveSlot, tgo.Title))
		}
		fmt.Printf("active tgos: %s\n", strings.Join(parts, "; "))
		for _, tgo := range activeTGOs {
			fmt.Printf("  active %s: %s\n", tgo.Code, tgo.MasteryHint)
		}
	}
	if len(completedTGOs) > 0 {
		var parts []string
		for _, tgo := range completedTGOs {
			parts = append(parts, tgo.Title)
		}
		fmt.Printf("completed tgos: %s\n", strings.Join(parts, "; "))
	}
	if len(upcomingTGOs) > 0 {
		var parts []string
		for _, tgo := range upcomingTGOs {
			parts = append(parts, fmt.Sprintf("%s (%s)", tgo.Title, tgo.Stage))
		}
		fmt.Printf("upcoming tgos: %s\n", strings.Join(parts, "; "))
	}
	if len(items) == 0 {
		fmt.Println("no progress data yet")
		return nil
	}
	if len(strongest) > 0 {
		fmt.Printf("strongest skills: %s\n", strings.Join(strongest, "; "))
	}
	if len(weakest) > 0 {
		fmt.Printf("weakest skills: %s\n", strings.Join(weakest, "; "))
	}
	if len(recurringWeaknesses) > 0 {
		fmt.Printf("recurring weaknesses: %s\n", strings.Join(recurringWeaknesses, "; "))
	}
	if len(recurringFindings) > 0 {
		fmt.Printf("recurring analyzer findings: %s\n", strings.Join(recurringFindings, "; "))
	}
	if len(recurringSlips) > 0 {
		fmt.Printf("completed tgos slipping: %s\n", strings.Join(recurringSlips, "; "))
	}
	for _, item := range items {
		fmt.Println(item)
	}
	return nil
}

func (c CLI) runCompare(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	submissionID := fs.Int64("submission", 0, "current submission id")
	againstID := fs.Int64("against", 0, "baseline submission id")
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *submissionID == 0 {
		return errors.New("usage: writing-coach compare --submission <id> [--against <submission-id>]")
	}

	current, err := c.Store.GetSubmission(ctx, *submissionID)
	if err != nil {
		return err
	}
	var baseline domain.Submission
	if *againstID != 0 {
		baseline, err = c.Store.GetSubmission(ctx, *againstID)
	} else {
		baseline, err = c.Store.PreviousSubmission(ctx, current)
	}
	if err != nil {
		return fmt.Errorf("no baseline submission available for comparison")
	}

	currentReview, err := c.Store.LatestReviewForSubmission(ctx, current.ID)
	if err != nil {
		return fmt.Errorf("current submission has no review yet")
	}
	baselineReview, err := c.Store.LatestReviewForSubmission(ctx, baseline.ID)
	if err != nil {
		return fmt.Errorf("baseline submission has no review yet")
	}

	comparison := review.CompareSubmissions(current, baseline, currentReview, baselineReview)
	fmt.Printf("compare submission %d against %d\n", current.ID, baseline.ID)
	fmt.Printf("summary: %s\n", comparison.Summary)
	fmt.Printf("word delta: %+d\n", comparison.WordDelta)
	if len(comparison.AddedWords) > 0 {
		fmt.Printf("notable added words: %s\n", strings.Join(comparison.AddedWords, ", "))
	}
	if len(comparison.RemovedWords) > 0 {
		fmt.Printf("notable removed words: %s\n", strings.Join(comparison.RemovedWords, ", "))
	}
	if len(comparison.AddressedWeaknesses) > 0 {
		fmt.Printf("addressed weaknesses: %s\n", strings.Join(comparison.AddressedWeaknesses, "; "))
	}
	if len(comparison.PersistingWeaknesses) > 0 {
		fmt.Printf("persisting weaknesses: %s\n", strings.Join(comparison.PersistingWeaknesses, "; "))
	}
	return nil
}
