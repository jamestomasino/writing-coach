package db

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestMigrateSeedAndProgress(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(context.Background(), filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(context.Background(), "Tomasino"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, treeID, _, err := store.EnsureDefaultUserTree(context.Background(), "tomasino", "Tomasino", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	exID, err := store.SaveExercise(context.Background(), domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Test",
		Brief:           "Brief",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"narrative clarity"},
		SuccessCriteria: []string{"two"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	subID, err := store.SaveSubmission(context.Background(), domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exID,
		Content:    "A short scene with doomed choices.",
		WordCount:  6,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	_, err = store.SaveReview(context.Background(), domain.Review{
		UserID:           userID,
		TreeID:           treeID,
		SubmissionID:     subID,
		ReviewKind:       "deterministic",
		Summary:          "Summary",
		Strengths:        []string{"s"},
		Weaknesses:       []string{"w"},
		AnalyzerFindings: []string{"f"},
		CompletedTGOChecks: []domain.TGOAssessment{
			{TGOCode: "story-causal-clarity", Status: "holding", Evidence: "still stable"},
		},
		NextFocus:       "narrative clarity",
		MetricWordCount: 6,
	}, []domain.SkillScore{
		{SubmissionID: subID, Skill: "narrative clarity", Score: 2},
		{SubmissionID: subID, Skill: "scene architecture", Score: 3},
	})
	if err != nil {
		t.Fatalf("save review: %v", err)
	}

	treeDef, err := store.TreeDefinitionBySlug(context.Background(), "story-craft-track")
	if err != nil {
		t.Fatalf("tree definition: %v", err)
	}
	report, err := store.ProgressReport(context.Background(), userID, treeID, treeDef.PrioritySkills, 5)
	if err != nil {
		t.Fatalf("progress report: %v", err)
	}
	if len(report) == 0 {
		t.Fatal("expected progress lines")
	}
	loaded, err := store.LatestReviewForSubmission(context.Background(), subID)
	if err != nil {
		t.Fatalf("latest review: %v", err)
	}
	if len(loaded.CompletedTGOChecks) != 1 || loaded.CompletedTGOChecks[0].TGOCode != "story-causal-clarity" {
		t.Fatalf("completed tgo checks = %#v", loaded.CompletedTGOChecks)
	}
	if err := store.SaveReviewArtifacts(context.Background(), domain.ReviewArtifacts{
		ReviewID:           loaded.ID,
		AnalyzerReportJSON: `{"metrics":{"word_count":6}}`,
		RecommendationJSON: `{"focus":"narrative clarity"}`,
		ComparisonJSON:     `{"summary":"tighter than before"}`,
	}); err != nil {
		t.Fatalf("save review artifacts: %v", err)
	}
	artifacts, err := store.GetReviewArtifacts(context.Background(), loaded.ID)
	if err != nil {
		t.Fatalf("get review artifacts: %v", err)
	}
	if artifacts.RecommendationJSON == "" || artifacts.AnalyzerReportJSON == "" {
		t.Fatalf("artifacts = %#v", artifacts)
	}
}

func TestEnsureDefaultUserTreeSeedsTreeSpecificTGOs(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, _, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "kid", "Kid", "youth-writing-foundations")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos: %v", err)
	}
	if len(active) != 3 {
		t.Fatalf("active len = %d", len(active))
	}
	if active[0].Code != "word-choice" {
		t.Fatalf("first active tgo = %q", active[0].Code)
	}
	if active[0].ProgressMode != domain.ProgressModePercent {
		t.Fatalf("expected percent progress mode, got %q", active[0].ProgressMode)
	}
}

func TestMigrateAddsPlaygroundStep2AndStep3SchemaToOlderDatabase(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if _, err := store.SQL.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		t.Fatalf("create schema migrations: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "migrations")
	for _, version := range []string{
		"0001_init.sql", "0002_add_exercise_generation_kind.sql", "0003_add_review_analyzer_findings.sql",
		"0004_add_submission_revision_fields.sql", "0005_add_tgo_tables.sql", "0006_add_users_and_trees.sql",
		"0007_add_review_completed_tgo_checks.sql", "0008_add_exercise_source_submission_id.sql",
		"0009_add_review_artifacts.sql", "0010_add_tree_definition_metadata.sql", "0011_add_tree_versions.sql",
		"0012_add_admin_identities.sql", "0013_add_user_onboarding.sql", "0014_add_review_annotations.sql",
		"0015_add_tgo_progress_mode.sql", "0016_add_onboarding_prompt_seed_fields.sql", "0017_add_review_jobs.sql",
		"0018_add_user_ai_provider_settings.sql", "0019_add_provider_notes.sql", "0020_add_ai_provider_events.sql",
		"0021_add_enrollment_onboarding_profiles.sql", "0022_add_archived_at_to_user_tree_enrollments.sql",
		"0023_add_closed_at_to_exercises.sql", "0024_add_writing_language_to_enrollment_onboarding_profiles.sql",
		"0025_add_ai_jobs.sql",
	} {
		sqlBytes, err := os.ReadFile(filepath.Join(migrationsDir, version))
		if err != nil {
			t.Fatalf("read migration %s: %v", version, err)
		}
		if _, err := store.SQL.ExecContext(ctx, string(sqlBytes)); err != nil {
			if !strings.Contains(err.Error(), "duplicate column name") {
				t.Fatalf("apply migration %s: %v", version, err)
			}
		}
		if _, err := store.SQL.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			t.Fatalf("record migration %s: %v", version, err)
		}
	}

	if err := store.Migrate(ctx, migrationsDir); err != nil {
		t.Fatalf("migrate latest: %v", err)
	}

	for _, table := range []string{"playground_sessions", "playground_reviews", "playground_drafts"} {
		var count int
		if err := store.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("expected table %s to exist", table)
		}
	}

	for _, column := range []struct {
		table  string
		column string
	}{
		{table: "playground_sessions", column: "latest_draft_id"},
		{table: "playground_sessions", column: "draft_count"},
		{table: "playground_reviews", column: "draft_id"},
		{table: "playground_reviews", column: "comparison_json"},
	} {
		rows, err := store.SQL.QueryContext(ctx, `PRAGMA table_info(`+column.table+`)`)
		if err != nil {
			t.Fatalf("pragma %s: %v", column.table, err)
		}
		found := false
		for rows.Next() {
			var cid int
			var name, ctype string
			var notNull int
			var dfltValue any
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notNull, &dfltValue, &pk); err != nil {
				t.Fatalf("scan pragma %s: %v", column.table, err)
			}
			if name == column.column {
				found = true
				break
			}
		}
		rows.Close()
		if !found {
			t.Fatalf("expected column %s.%s to exist", column.table, column.column)
		}
	}
}

func TestEnsureDefaultUserTreeReconcilesChangedSeedsUsingLeastProgressedSlot(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	def := domain.TGOTreeDefinition{
		Slug:           "seed-reconcile-track",
		Title:          "Seed Reconcile Track",
		Description:    "Test track for seed migration behavior.",
		SeedCodes:      []string{"seed-core-a", "seed-core-b", "seed-domain-c"},
		PrioritySkills: []string{"claim clarity", "audience alignment", "reasoning quality"},
		TGOs: []domain.TGO{
			{Code: "seed-core-a", Title: "Seed A", Description: "A", Stage: "core", StageOrder: 1, ProgressMode: domain.ProgressModePercent},
			{Code: "seed-core-b", Title: "Seed B", Description: "B", Stage: "core", StageOrder: 2, ProgressMode: domain.ProgressModePercent},
			{Code: "seed-domain-c", Title: "Seed C", Description: "C", Stage: "core", StageOrder: 3, ProgressMode: domain.ProgressModePercent},
			{Code: "seed-domain-d", Title: "Seed D", Description: "D", Stage: "core", StageOrder: 4, ProgressMode: domain.ProgressModePercent},
		},
	}
	if err := store.SaveTreeDefinition(ctx, def); err != nil {
		t.Fatalf("save initial tree: %v", err)
	}

	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", def.Slug)
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	exID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Seed Practice",
		Brief:           "Practice one seed.",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"claim clarity"},
		TGOCodes:        []string{"seed-core-b"},
		SuccessCriteria: []string{"progress"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	subID, err := store.SaveSubmission(ctx, domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exID,
		Content:    "progress",
		WordCount:  1,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	if _, err := store.SaveReview(ctx, domain.Review{
		UserID:           userID,
		TreeID:           treeID,
		SubmissionID:     subID,
		ReviewKind:       "deterministic",
		Summary:          "progress",
		Strengths:        []string{"one"},
		Weaknesses:       []string{},
		AnalyzerFindings: []string{},
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "seed-core-b", Status: "secure", Evidence: "kept"},
		},
		NextFocus:       "Seed B",
		MetricWordCount: 1,
	}, nil); err != nil {
		t.Fatalf("save review: %v", err)
	}

	def.SeedCodes = []string{"seed-core-a", "seed-core-b", "seed-domain-d"}
	if err := store.SaveTreeDefinition(ctx, def); err != nil {
		t.Fatalf("save updated tree: %v", err)
	}

	if _, _, _, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", def.Slug); err != nil {
		t.Fatalf("re-ensure default user tree: %v", err)
	}

	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos: %v", err)
	}
	if len(active) != 3 {
		t.Fatalf("active len = %d", len(active))
	}
	if active[0].Code != "seed-core-a" {
		t.Fatalf("slot 1 = %q", active[0].Code)
	}
	if active[1].Code != "seed-core-b" {
		t.Fatalf("slot 2 = %q", active[1].Code)
	}
	if active[2].Code != "seed-domain-d" {
		t.Fatalf("slot 3 = %q", active[2].Code)
	}
}

func TestTGOMasterySignalUsesRollingEvidence(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "youth-writing-foundations")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	exID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Test",
		Brief:           "Brief",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"sentence variety"},
		SuccessCriteria: []string{"two"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}

	statuses := []string{"secure", "mastered", "mastered"}
	var target domain.TGO
	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos: %v", err)
	}
	target = active[1]

	for i, status := range statuses {
		subID, err := store.SaveSubmission(ctx, domain.Submission{
			UserID:     userID,
			TreeID:     treeID,
			ExerciseID: exID,
			Content:    "Sentence control improves across drafts.",
			WordCount:  5,
		})
		if err != nil {
			t.Fatalf("save submission %d: %v", i, err)
		}
		if _, err := store.SaveReview(ctx, domain.Review{
			UserID:           userID,
			TreeID:           treeID,
			SubmissionID:     subID,
			ReviewKind:       "deterministic",
			Summary:          "Summary",
			Strengths:        []string{"s"},
			Weaknesses:       []string{"w"},
			AnalyzerFindings: []string{"f"},
			TGOAssessments: []domain.TGOAssessment{
				{TGOCode: target.Code, Status: status, Evidence: "evidence"},
			},
			NextFocus:       target.Title,
			MetricWordCount: 5,
		}, nil); err != nil {
			t.Fatalf("save review %d: %v", i, err)
		}
	}

	signal, err := store.TGOMasterySignal(ctx, enrollmentID, target, "")
	if err != nil {
		t.Fatalf("mastery signal: %v", err)
	}
	if signal.Percent == 0 || signal.EvidenceCount != 3 {
		t.Fatalf("unexpected signal: %#v", signal)
	}
	if !signal.Ready {
		t.Fatalf("signal should be ready: %#v", signal)
	}
}

func TestAIProviderSettingsCRUD(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, _, _, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	validatedAt := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)
	settings := domain.AIProviderSettings{
		UserID:              userID,
		Provider:            "openai",
		APIKeyEncrypted:     "enc:test",
		APIKeyLast4:         "abcd",
		BaseURLOverride:     "https://api.openai.com/v1",
		PromptModelOverride: "gpt-5.4",
		ReviewModelOverride: "gpt-5.4-mini",
		Enabled:             true,
		ValidatedAt:         validatedAt,
		LastValidationError: "",
	}
	if err := store.SaveAIProviderSettings(ctx, settings); err != nil {
		t.Fatalf("save provider settings: %v", err)
	}

	loaded, err := store.AIProviderSettingsByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("load provider settings: %v", err)
	}
	if loaded.Provider != "openai" || loaded.APIKeyEncrypted != "enc:test" || loaded.APIKeyLast4 != "abcd" {
		t.Fatalf("loaded provider settings = %#v", loaded)
	}
	if !loaded.Enabled {
		t.Fatalf("expected enabled settings, got %#v", loaded)
	}
	if loaded.ValidatedAt.IsZero() {
		t.Fatalf("expected validated_at to be set, got %#v", loaded)
	}

	settings.Provider = "groq"
	settings.APIKeyEncrypted = "enc:next"
	settings.APIKeyLast4 = "wxyz"
	settings.Enabled = false
	settings.ValidatedAt = time.Time{}
	settings.LastValidationError = "invalid credentials"
	if err := store.SaveAIProviderSettings(ctx, settings); err != nil {
		t.Fatalf("update provider settings: %v", err)
	}

	updated, err := store.AIProviderSettingsByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("reload provider settings: %v", err)
	}
	if updated.Provider != "groq" || updated.APIKeyEncrypted != "enc:next" || updated.APIKeyLast4 != "wxyz" {
		t.Fatalf("updated provider settings = %#v", updated)
	}
	if updated.Enabled {
		t.Fatalf("expected disabled settings, got %#v", updated)
	}
	if !updated.ValidatedAt.IsZero() {
		t.Fatalf("expected validated_at cleared, got %#v", updated)
	}
	if updated.LastValidationError != "invalid credentials" {
		t.Fatalf("last validation error = %q", updated.LastValidationError)
	}

	if err := store.DeleteAIProviderSettings(ctx, userID); err != nil {
		t.Fatalf("delete provider settings: %v", err)
	}
	if _, err := store.AIProviderSettingsByUserID(ctx, userID); !IsNotFound(err) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestAIProviderEventsFilteringAndRetention(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, _, _, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	now := time.Now().UTC()
	for _, event := range []domain.AIProviderEvent{
		{UserID: userID, Provider: "openai", Event: "settings_validate_failed", Category: "auth", StatusCode: 400, CreatedAt: now},
		{UserID: userID, Provider: "openai", Event: "settings_validate_rate_limited", Category: "local_rate_limit", StatusCode: 429, CreatedAt: now.Add(-time.Hour)},
		{UserID: userID, Provider: "anthropic", Event: "generation_fallback", Category: "quota", StatusCode: 502, CreatedAt: now.Add(-48 * time.Hour)},
	} {
		if err := store.SaveAIProviderEvent(ctx, event); err != nil {
			t.Fatalf("save provider event: %v", err)
		}
	}

	recent, err := store.ListRecentAIProviderEvents(ctx, 20, now.Add(-24*time.Hour), "openai", "")
	if err != nil {
		t.Fatalf("list recent provider events: %v", err)
	}
	if len(recent) != 2 {
		t.Fatalf("recent filtered events = %d", len(recent))
	}

	summary, err := store.SummarizeAIProviderEventsSince(ctx, now.Add(-24*time.Hour), "openai", "")
	if err != nil {
		t.Fatalf("summarize provider events: %v", err)
	}
	if summary.Total != 2 {
		t.Fatalf("summary total = %d", summary.Total)
	}

	if err := store.DeleteAIProviderEventsOlderThan(ctx, now.Add(-24*time.Hour)); err != nil {
		t.Fatalf("delete old provider events: %v", err)
	}
	allAfterRetention, err := store.ListRecentAIProviderEvents(ctx, 20, now.Add(-7*24*time.Hour), "", "")
	if err != nil {
		t.Fatalf("list provider events after retention: %v", err)
	}
	if len(allAfterRetention) != 2 {
		t.Fatalf("provider events after retention = %d", len(allAfterRetention))
	}
}

func TestUpdateProgressionHoldStateActivatesAndClears(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, _, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	if err := store.UpdateProgressionHoldState(ctx, enrollmentID, true, "completed_tgo_slipping", 21, 2); err != nil {
		t.Fatalf("activate hold: %v", err)
	}
	state, err := store.GetCurriculumState(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("get curriculum state after activation: %v", err)
	}
	if !state.ProgressionHoldActive {
		t.Fatal("expected progression hold to be active")
	}
	if state.ProgressionHoldReasonCode != "completed_tgo_slipping" {
		t.Fatalf("hold reason code = %q", state.ProgressionHoldReasonCode)
	}
	if state.HoldTriggerReviewID != 21 {
		t.Fatalf("hold trigger review id = %d", state.HoldTriggerReviewID)
	}
	if state.HoldUpdatedAt.IsZero() {
		t.Fatal("expected hold updated timestamp")
	}

	if err := store.UpdateProgressionHoldState(ctx, enrollmentID, false, "", 22, 2); err != nil {
		t.Fatalf("attempt hold clear (1): %v", err)
	}
	state, err = store.GetCurriculumState(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("get curriculum state after clear attempt 1: %v", err)
	}
	if !state.ProgressionHoldActive {
		t.Fatal("expected progression hold to remain active until streak threshold is reached")
	}
	if state.HoldClearStreak != 1 {
		t.Fatalf("hold clear streak = %d", state.HoldClearStreak)
	}
	if err := store.UpdateProgressionHoldState(ctx, enrollmentID, false, "", 23, 2); err != nil {
		t.Fatalf("attempt hold clear (2): %v", err)
	}
	state, err = store.GetCurriculumState(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("get curriculum state after clear attempt 2: %v", err)
	}
	if state.ProgressionHoldActive {
		t.Fatal("expected progression hold to be cleared after streak threshold")
	}
	if state.ProgressionHoldReasonCode != "" {
		t.Fatalf("expected empty hold reason code, got %q", state.ProgressionHoldReasonCode)
	}
	if state.HoldTriggerReviewID != 21 {
		t.Fatalf("expected trigger review id to remain 21, got %d", state.HoldTriggerReviewID)
	}
	if state.HoldClearedReviewID != 23 {
		t.Fatalf("hold cleared review id = %d", state.HoldClearedReviewID)
	}
}

func TestUpdateProgressionHoldStateIsEnrollmentScoped(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, _, storyEnrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("story enrollment: %v", err)
	}
	_, _, technicalEnrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "technical-writing-track")
	if err != nil {
		t.Fatalf("technical enrollment: %v", err)
	}

	if err := store.UpdateProgressionHoldState(ctx, storyEnrollmentID, true, "completed_tgo_slipping", 99, 2); err != nil {
		t.Fatalf("activate story hold: %v", err)
	}

	storyState, err := store.GetCurriculumState(ctx, storyEnrollmentID)
	if err != nil {
		t.Fatalf("get story curriculum state: %v", err)
	}
	if !storyState.ProgressionHoldActive {
		t.Fatal("expected story enrollment hold to be active")
	}

	technicalState, err := store.GetCurriculumState(ctx, technicalEnrollmentID)
	if err != nil {
		t.Fatalf("get technical curriculum state: %v", err)
	}
	if technicalState.ProgressionHoldActive {
		t.Fatal("expected technical enrollment hold to remain inactive")
	}
}

func TestSaveAndLoadDecisionEventsByReview(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}
	exerciseID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Decision events",
		Brief:           "Test payload.",
		Constraints:     []string{"one paragraph"},
		FocusSkills:     []string{"narrative clarity"},
		SuccessCriteria: []string{"show cause and consequence"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := store.SaveSubmission(ctx, domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exerciseID,
		Content:    "A short draft.",
		WordCount:  3,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	reviewID, err := store.SaveReview(ctx, domain.Review{
		UserID:           userID,
		TreeID:           treeID,
		SubmissionID:     submissionID,
		ReviewKind:       "deterministic",
		Summary:          "ok",
		Strengths:        []string{"clear"},
		Weaknesses:       []string{"thin detail"},
		AnalyzerFindings: []string{"finding"},
		NextFocus:        "Causal Clarity",
		MetricWordCount:  3,
	}, []domain.SkillScore{{SubmissionID: submissionID, Skill: "narrative clarity", Score: 3}})
	if err != nil {
		t.Fatalf("save review: %v", err)
	}

	if err := store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            reviewID,
		SubmissionID:        submissionID,
		EventType:           "review_scored",
		DecisionPayloadJSON: `{"review_kind":"deterministic"}`,
		RuleVersion:         "deterministic-scoring-v1",
		EvidenceRefsJSON:    `["submission_skill_scores"]`,
	}); err != nil {
		t.Fatalf("save decision event review_scored: %v", err)
	}
	if err := store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            reviewID,
		SubmissionID:        submissionID,
		EventType:           "recommendation_issued",
		DecisionPayloadJSON: `{"focus":"Causal Clarity"}`,
		RuleVersion:         "curriculum-sync-v1",
		EvidenceRefsJSON:    `["recommendation"]`,
	}); err != nil {
		t.Fatalf("save decision event recommendation_issued: %v", err)
	}

	events, err := store.DecisionEventsByReview(ctx, reviewID)
	if err != nil {
		t.Fatalf("decision events by review: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("decision events len = %d", len(events))
	}
	if events[0].EventType != "review_scored" || events[1].EventType != "recommendation_issued" {
		t.Fatalf("unexpected decision events = %#v", events)
	}
}

func TestPedagogyIntegritySnapshot(t *testing.T) {
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}
	exerciseID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Snapshot",
		Brief:           "brief",
		Constraints:     []string{"one paragraph"},
		FocusSkills:     []string{"narrative clarity"},
		SuccessCriteria: []string{"clear causality"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := store.SaveSubmission(ctx, domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exerciseID,
		Content:    "test",
		WordCount:  1,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	reviewID, err := store.SaveReview(ctx, domain.Review{
		UserID:           userID,
		TreeID:           treeID,
		SubmissionID:     submissionID,
		ReviewKind:       "deterministic",
		Summary:          "ok",
		Strengths:        []string{"s"},
		Weaknesses:       []string{"w"},
		AnalyzerFindings: []string{"f"},
		NextFocus:        "Causal Clarity",
		MetricWordCount:  1,
	}, nil)
	if err != nil {
		t.Fatalf("save review: %v", err)
	}
	if err := store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            reviewID,
		SubmissionID:        submissionID,
		EventType:           "review_scored",
		DecisionPayloadJSON: `{"review_kind":"deterministic"}`,
		RuleVersion:         "deterministic-scoring-v1",
		EvidenceRefsJSON:    `["submission_skill_scores"]`,
	}); err != nil {
		t.Fatalf("save review_scored event: %v", err)
	}
	if err := store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		EventType:           "progression_hold_blocked",
		DecisionPayloadJSON: `{"reason_code":"completed_tgo_slipping"}`,
		RuleVersion:         "progression-hold-block-v1",
		EvidenceRefsJSON:    `["user_curriculum_state"]`,
	}); err != nil {
		t.Fatalf("save blocked event: %v", err)
	}
	if err := store.UpdateProgressionHoldState(ctx, enrollmentID, true, "completed_tgo_slipping", reviewID, 2); err != nil {
		t.Fatalf("activate hold state: %v", err)
	}
	if err := store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            reviewID,
		SubmissionID:        submissionID,
		EventType:           "progression_hold_activated",
		DecisionPayloadJSON: `{"reason_code":"completed_tgo_slipping"}`,
		RuleVersion:         "progression-hold-activate-v1",
		EvidenceRefsJSON:    `["completed_tgo_checks","user_curriculum_state"]`,
	}); err != nil {
		t.Fatalf("save hold activated event: %v", err)
	}
	if err := store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            reviewID,
		SubmissionID:        submissionID,
		EventType:           "progression_hold_cleared",
		DecisionPayloadJSON: `{"reason_code":""}`,
		RuleVersion:         "progression-hold-clear-v1",
		EvidenceRefsJSON:    `["completed_tgo_checks","user_curriculum_state"]`,
	}); err != nil {
		t.Fatalf("save hold cleared event: %v", err)
	}
	if err := store.SaveReviewArtifacts(ctx, domain.ReviewArtifacts{
		ReviewID:           reviewID,
		AnalyzerReportJSON: `{"summary":"ok"}`,
		RecommendationJSON: `{"intervention_outcomes":[{"status":"resolved"},{"status":"persisting"},{"status":"resolved"}]}`,
		ComparisonJSON:     `{"summary":"ok"}`,
		AnnotationsJSON:    `[]`,
	}); err != nil {
		t.Fatalf("save review artifacts: %v", err)
	}
	if err := store.MarkTGOCompleted(ctx, enrollmentID, "story-causal-clarity"); err != nil {
		t.Fatalf("mark tgo completed: %v", err)
	}

	snapshot, err := store.PedagogyIntegritySnapshot(ctx, time.Now().UTC().Add(-24*time.Hour), 24)
	if err != nil {
		t.Fatalf("pedagogy integrity snapshot: %v", err)
	}
	if snapshot.TotalReviews < 1 {
		t.Fatalf("total reviews = %d", snapshot.TotalReviews)
	}
	if snapshot.ReviewScoredEvents < 1 {
		t.Fatalf("review scored events = %d", snapshot.ReviewScoredEvents)
	}
	if snapshot.RecommendationEvents != 0 {
		t.Fatalf("recommendation events = %d", snapshot.RecommendationEvents)
	}
	if snapshot.ReviewsMissingDecisionEvents < 1 {
		t.Fatalf("missing decision events = %d", snapshot.ReviewsMissingDecisionEvents)
	}
	if snapshot.HoldBlockedEvents < 1 {
		t.Fatalf("hold blocked events = %d", snapshot.HoldBlockedEvents)
	}
	if snapshot.ActiveHoldEnrollments < 1 {
		t.Fatalf("active hold enrollments = %d", snapshot.ActiveHoldEnrollments)
	}
	if snapshot.AvgHoldClearHours < 0 {
		t.Fatalf("avg hold clear hours = %f", snapshot.AvgHoldClearHours)
	}
	if snapshot.InterventionResolvedCount != 2 {
		t.Fatalf("intervention resolved count = %d", snapshot.InterventionResolvedCount)
	}
	if snapshot.InterventionPersistingCount != 1 {
		t.Fatalf("intervention persisting count = %d", snapshot.InterventionPersistingCount)
	}
	if snapshot.InterventionResolutionRate <= 0 {
		t.Fatalf("intervention resolution rate = %f", snapshot.InterventionResolutionRate)
	}
	if snapshot.MasteryCompletions < 1 {
		t.Fatalf("mastery completions = %d", snapshot.MasteryCompletions)
	}
	if snapshot.MasteryVelocityPer100Reviews <= 0 {
		t.Fatalf("mastery velocity per 100 reviews = %f", snapshot.MasteryVelocityPer100Reviews)
	}
}
