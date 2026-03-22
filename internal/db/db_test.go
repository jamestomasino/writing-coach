package db

import (
	"context"
	"encoding/json"
	"path/filepath"
	"slices"
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

func TestEnsureSeedDataMigratesMythicTrackDataIntoStoryCraft(t *testing.T) {
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

	for _, code := range []string{"story-causal-clarity", "story-scene-architecture", "story-prose-precision"} {
		if _, err := store.SQL.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order, progress_mode)
			VALUES (?, ?, ?, 'core', 1, ?)
		`, code, code, code, domain.ProgressModePercent); err != nil {
			t.Fatalf("insert story tgo %q: %v", code, err)
		}
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO tgo_trees (slug, title, description, seed_codes_json, priority_skills_json)
		VALUES (?, 'Story Craft', 'Story track', ?, ?)
	`, newStoryCraftTreeSlug, mustJSON([]string{"story-causal-clarity", "story-scene-architecture", "story-prose-precision"}), mustJSON([]string{"narrative clarity"})); err != nil {
		t.Fatalf("insert story tree: %v", err)
	}
	var storyTreeID int64
	if err := store.SQL.QueryRowContext(ctx, `SELECT id FROM tgo_trees WHERE slug = ?`, newStoryCraftTreeSlug).Scan(&storyTreeID); err != nil {
		t.Fatalf("story tree id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO users (slug, name, active_tree_slug)
		VALUES ('legacy-user', 'Legacy User', ?)
	`, oldMythicTreeSlug); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var userID int64
	if err := store.SQL.QueryRowContext(ctx, `SELECT id FROM users WHERE slug = 'legacy-user'`).Scan(&userID); err != nil {
		t.Fatalf("user id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO user_onboarding_profiles (
			user_id, writing_type, assignment_format, target_audience, subject_matter, experience_level, desired_tone,
			biggest_weaknesses_json, desired_outcomes_json, difficulty_intensity, writing_goals, generated_tree_slug, template_key
		) VALUES (?, 'fiction', '', '', '', 'beginner', 'serious', '[]', '[]', 'steady', 'learn fiction', ?, 'fiction')
	`, userID, oldMythicTreeSlug); err != nil {
		t.Fatalf("insert user onboarding profile: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO user_tree_enrollments (user_id, tree_id) VALUES (?, ?)
	`, userID, storyTreeID); err != nil {
		t.Fatalf("insert story enrollment: %v", err)
	}
	var storyEnrollmentID int64
	if err := store.SQL.QueryRowContext(ctx, `
		SELECT id FROM user_tree_enrollments WHERE user_id = ? AND tree_id = ?
	`, userID, storyTreeID).Scan(&storyEnrollmentID); err != nil {
		t.Fatalf("story enrollment id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code)
		VALUES (?, 1, 'story-causal-clarity'), (?, 2, 'story-scene-architecture'), (?, 3, 'story-prose-precision')
	`, storyEnrollmentID, storyEnrollmentID, storyEnrollmentID); err != nil {
		t.Fatalf("insert story actives: %v", err)
	}

	for _, code := range []string{"causal-clarity", "scene-architecture", "prose-precision"} {
		if _, err := store.SQL.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order, progress_mode)
			VALUES (?, ?, ?, 'core', 1, ?)
		`, code, code, code, domain.ProgressModePercent); err != nil {
			t.Fatalf("insert mythic tgo %q: %v", code, err)
		}
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO tgo_trees (slug, title, description, seed_codes_json, priority_skills_json)
		VALUES (?, 'Mythic', 'Old retired track', ?, ?)
	`, oldMythicTreeSlug, mustJSON([]string{"causal-clarity", "scene-architecture", "prose-precision"}), mustJSON([]string{"narrative clarity"})); err != nil {
		t.Fatalf("insert mythic tree: %v", err)
	}
	var mythicTreeID int64
	if err := store.SQL.QueryRowContext(ctx, `SELECT id FROM tgo_trees WHERE slug = ?`, oldMythicTreeSlug).Scan(&mythicTreeID); err != nil {
		t.Fatalf("mythic tree id: %v", err)
	}
	for _, code := range []string{"causal-clarity", "scene-architecture", "prose-precision"} {
		if _, err := store.SQL.ExecContext(ctx, `
			INSERT INTO tree_tgos (tree_id, tgo_code, prerequisites_json, mastery_hint)
			VALUES (?, ?, '[]', '')
		`, mythicTreeID, code); err != nil {
			t.Fatalf("insert mythic tree tgo %q: %v", code, err)
		}
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO user_tree_enrollments (user_id, tree_id) VALUES (?, ?)
	`, userID, mythicTreeID); err != nil {
		t.Fatalf("insert mythic enrollment: %v", err)
	}
	var mythicEnrollmentID int64
	if err := store.SQL.QueryRowContext(ctx, `
		SELECT id FROM user_tree_enrollments WHERE user_id = ? AND tree_id = ?
	`, userID, mythicTreeID).Scan(&mythicEnrollmentID); err != nil {
		t.Fatalf("mythic enrollment id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO user_curriculum_state (enrollment_id, current_focus, difficulty_level)
		VALUES (?, 'causal-clarity', 2)
	`, mythicEnrollmentID); err != nil {
		t.Fatalf("insert mythic curriculum state: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code)
		VALUES (?, 1, 'causal-clarity'), (?, 2, 'scene-architecture'), (?, 3, 'prose-precision')
	`, mythicEnrollmentID, mythicEnrollmentID, mythicEnrollmentID); err != nil {
		t.Fatalf("insert mythic actives: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code)
		VALUES (?, 'causal-clarity')
	`, mythicEnrollmentID); err != nil {
		t.Fatalf("insert mythic completed: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_onboarding_profiles (
			enrollment_id, writing_language, writing_type, assignment_format, target_audience, subject_matter,
			experience_level, desired_tone, biggest_weaknesses_json, desired_outcomes_json, difficulty_intensity,
			writing_goals, generated_tree_slug, template_key
		) VALUES (?, 'en', 'fiction', '', '', '', 'beginner', 'serious', '[]', '[]', 'steady', 'learn fiction', ?, 'fiction')
	`, mythicEnrollmentID, oldMythicTreeSlug); err != nil {
		t.Fatalf("insert mythic enrollment profile: %v", err)
	}
	res, err := store.SQL.ExecContext(ctx, `
		INSERT INTO exercises (
			user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json,
			success_criteria_json, generation_kind, provider_note
		) VALUES (?, ?, 'Legacy Exercise', 'Brief', '[]', ?, ?, '[]', 'deterministic', '')
	`, userID, mythicTreeID, mustJSON([]string{"tragic inevitability"}), mustJSON([]string{"causal-clarity"}))
	if err != nil {
		t.Fatalf("insert mythic exercise: %v", err)
	}
	exerciseID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("exercise last insert id: %v", err)
	}
	res, err = store.SQL.ExecContext(ctx, `
		INSERT INTO submissions (user_id, tree_id, exercise_id, content, word_count)
		VALUES (?, ?, ?, 'old draft', 2)
	`, userID, mythicTreeID, exerciseID)
	if err != nil {
		t.Fatalf("insert mythic submission: %v", err)
	}
	submissionID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("submission last insert id: %v", err)
	}
	completedChecks := mustJSON([]domain.TGOAssessment{{TGOCode: "causal-clarity", Status: "secure", Evidence: "evidence"}})
	res, err = store.SQL.ExecContext(ctx, `
		INSERT INTO reviews (
			user_id, tree_id, submission_id, review_kind, provider_note, summary, strengths_json, weaknesses_json,
			analyzer_findings_json, completed_tgo_checks_json, next_focus, metric_word_count
		) VALUES (?, ?, ?, 'deterministic', '', 'Summary', '[]', '[]', '[]', ?, 'focus', 2)
	`, userID, mythicTreeID, submissionID, completedChecks)
	if err != nil {
		t.Fatalf("insert mythic review: %v", err)
	}
	reviewID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("review last insert id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO review_tgo_assessments (review_id, submission_id, tgo_code, status, evidence)
		VALUES (?, ?, 'causal-clarity', 'secure', 'evidence')
	`, reviewID, submissionID); err != nil {
		t.Fatalf("insert mythic assessment: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO review_artifacts (review_id, analyzer_report_json, recommendation_json, comparison_json, annotations_json)
		VALUES (?, '{}', '{}', '{}', ?)
	`, reviewID, mustJSON([]domain.ReviewAnnotation{{Quote: "old", TGOCode: "causal-clarity", Category: "clarity", Comment: "ok", Severity: "info"}})); err != nil {
		t.Fatalf("insert mythic artifacts: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO review_jobs (user_id, tree_id, enrollment_id, submission_id, status, max_attempts)
		VALUES (?, ?, ?, ?, 'queued', 3)
	`, userID, mythicTreeID, mythicEnrollmentID, submissionID); err != nil {
		t.Fatalf("insert mythic review job: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO submission_skill_scores (submission_id, skill_name, score)
		VALUES (?, 'tragic inevitability', 4), (?, 'symbolic control', 3)
	`, submissionID, submissionID); err != nil {
		t.Fatalf("insert legacy submission skill scores: %v", err)
	}

	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("ensure seed data: %v", err)
	}

	var gotActiveTree string
	if err := store.SQL.QueryRowContext(ctx, `SELECT active_tree_slug FROM users WHERE id = ?`, userID).Scan(&gotActiveTree); err != nil {
		t.Fatalf("active tree slug: %v", err)
	}
	if gotActiveTree != newStoryCraftTreeSlug {
		t.Fatalf("active tree slug = %q", gotActiveTree)
	}

	var mythicTreeCount int
	if err := store.SQL.QueryRowContext(ctx, `SELECT COUNT(1) FROM tgo_trees WHERE slug = ?`, oldMythicTreeSlug).Scan(&mythicTreeCount); err != nil {
		t.Fatalf("mythic tree count: %v", err)
	}
	if mythicTreeCount != 0 {
		t.Fatalf("expected mythic tree deleted, count=%d", mythicTreeCount)
	}

	var enrollmentCount int
	if err := store.SQL.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM user_tree_enrollments WHERE user_id = ?
	`, userID).Scan(&enrollmentCount); err != nil {
		t.Fatalf("enrollment count: %v", err)
	}
	if enrollmentCount != 1 {
		t.Fatalf("enrollment count = %d", enrollmentCount)
	}

	completed, err := store.CompletedTGOs(ctx, storyEnrollmentID)
	if err != nil {
		t.Fatalf("completed tgos: %v", err)
	}
	if len(completed) != 1 || completed[0].Code != "story-causal-clarity" {
		t.Fatalf("completed = %#v", completed)
	}

	exercise, err := store.GetExercise(ctx, exerciseID)
	if err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	if exercise.TreeID != storyTreeID || len(exercise.TGOCodes) != 1 || exercise.TGOCodes[0] != "story-causal-clarity" {
		t.Fatalf("exercise after migration = %#v", exercise)
	}
	if len(exercise.FocusSkills) != 1 || exercise.FocusSkills[0] != "narrative clarity" {
		t.Fatalf("exercise focus skills after migration = %#v", exercise.FocusSkills)
	}

	submission, err := store.GetSubmission(ctx, submissionID)
	if err != nil {
		t.Fatalf("get submission: %v", err)
	}
	if submission.TreeID != storyTreeID {
		t.Fatalf("submission tree id = %d", submission.TreeID)
	}

	review, err := store.GetReview(ctx, reviewID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if review.TreeID != storyTreeID || len(review.CompletedTGOChecks) != 1 || review.CompletedTGOChecks[0].TGOCode != "story-causal-clarity" {
		t.Fatalf("review after migration = %#v", review)
	}
	if len(review.TGOAssessments) != 1 || review.TGOAssessments[0].TGOCode != "story-causal-clarity" {
		t.Fatalf("review assessments after migration = %#v", review.TGOAssessments)
	}

	job, err := store.ReviewJobBySubmission(ctx, userID, storyTreeID, submissionID)
	if err != nil {
		t.Fatalf("review job by submission: %v", err)
	}
	if job.EnrollmentID != storyEnrollmentID {
		t.Fatalf("review job enrollment = %d", job.EnrollmentID)
	}

	artifacts, err := store.GetReviewArtifacts(ctx, reviewID)
	if err != nil {
		t.Fatalf("get review artifacts: %v", err)
	}
	var annotations []domain.ReviewAnnotation
	if err := json.Unmarshal([]byte(artifacts.AnnotationsJSON), &annotations); err != nil {
		t.Fatalf("unmarshal annotations: %v", err)
	}
	if len(annotations) != 1 || annotations[0].TGOCode != "story-causal-clarity" {
		t.Fatalf("annotations after migration = %#v", annotations)
	}

	profile, err := store.OnboardingProfileByEnrollmentID(ctx, storyEnrollmentID)
	if err != nil {
		t.Fatalf("onboarding profile by enrollment: %v", err)
	}
	if profile.GeneratedTreeSlug != newStoryCraftTreeSlug {
		t.Fatalf("generated tree slug = %q", profile.GeneratedTreeSlug)
	}

	var userProfileSlug string
	if err := store.SQL.QueryRowContext(ctx, `
		SELECT generated_tree_slug FROM user_onboarding_profiles WHERE user_id = ?
	`, userID).Scan(&userProfileSlug); err != nil {
		t.Fatalf("user onboarding generated tree slug: %v", err)
	}
	if userProfileSlug != newStoryCraftTreeSlug {
		t.Fatalf("user onboarding generated tree slug = %q", userProfileSlug)
	}

	scores, err := store.SubmissionSkillScores(ctx, submissionID)
	if err != nil {
		t.Fatalf("submission skill scores: %v", err)
	}
	if len(scores) != 2 || scores[0].Skill != "narrative clarity" || scores[1].Skill != "scene architecture" {
		t.Fatalf("submission skill scores after migration = %#v", scores)
	}

	var migrationRecorded int
	if err := store.SQL.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM schema_migrations WHERE version = ?
	`, legacySkillPathUpgradeVersion).Scan(&migrationRecorded); err != nil {
		t.Fatalf("migration recorded: %v", err)
	}
	if migrationRecorded != 1 {
		t.Fatalf("migration record count = %d", migrationRecorded)
	}
}

func TestEnsureSeedDataMigratesLegacyCodesWithinCurrentTrackSlug(t *testing.T) {
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

	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO tgo_catalog (code, title, description, stage, stage_order, progress_mode)
		VALUES ('causal-clarity', 'Causal Clarity', 'legacy', 'core', 1, ?)
	`, domain.ProgressModePercent); err != nil {
		t.Fatalf("insert legacy tgo catalog row: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO tgo_trees (slug, title, description, seed_codes_json, priority_skills_json)
		VALUES ('story-craft-track', 'Legacy Story', 'legacy', '[]', '[]')
	`); err != nil {
		t.Fatalf("insert legacy story tree: %v", err)
	}
	var storyTreeID int64
	if err := store.SQL.QueryRowContext(ctx, `SELECT id FROM tgo_trees WHERE slug = 'story-craft-track'`).Scan(&storyTreeID); err != nil {
		t.Fatalf("story tree id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO tree_tgos (tree_id, tgo_code, prerequisites_json, mastery_hint)
		VALUES (?, 'causal-clarity', '[]', '')
	`, storyTreeID); err != nil {
		t.Fatalf("insert legacy story tree_tgo: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO users (slug, name, active_tree_slug)
		VALUES ('legacy-story-user', 'Legacy Story User', 'story-craft-track')
	`); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var userID int64
	if err := store.SQL.QueryRowContext(ctx, `SELECT id FROM users WHERE slug = 'legacy-story-user'`).Scan(&userID); err != nil {
		t.Fatalf("user id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO user_tree_enrollments (user_id, tree_id) VALUES (?, ?)
	`, userID, storyTreeID); err != nil {
		t.Fatalf("insert enrollment: %v", err)
	}
	var enrollmentID int64
	if err := store.SQL.QueryRowContext(ctx, `
		SELECT id FROM user_tree_enrollments WHERE user_id = ? AND tree_id = ?
	`, userID, storyTreeID).Scan(&enrollmentID); err != nil {
		t.Fatalf("enrollment id: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO user_curriculum_state (enrollment_id, current_focus, difficulty_level)
		VALUES (?, 'causal-clarity', 2)
	`, enrollmentID); err != nil {
		t.Fatalf("insert curriculum state: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code)
		VALUES (?, 1, 'causal-clarity')
	`, enrollmentID); err != nil {
		t.Fatalf("insert active tgo: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code)
		VALUES (?, 'causal-clarity')
	`, enrollmentID); err != nil {
		t.Fatalf("insert completed tgo: %v", err)
	}
	res, err := store.SQL.ExecContext(ctx, `
		INSERT INTO exercises (
			user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json,
			success_criteria_json, generation_kind, provider_note
		) VALUES (?, ?, 'Legacy Current Track Exercise', 'Brief', '[]', ?, ?, '[]', 'deterministic', '')
	`, userID, storyTreeID, mustJSON([]string{"tragic inevitability"}), mustJSON([]string{"causal-clarity"}))
	if err != nil {
		t.Fatalf("insert exercise: %v", err)
	}
	exerciseID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("exercise id: %v", err)
	}

	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("ensure seed data: %v", err)
	}

	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("active tgos: %v", err)
	}
	if len(active) == 0 || active[0].Code != "story-causal-clarity" {
		t.Fatalf("active after migration = %#v", active)
	}

	completed, err := store.CompletedTGOs(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("completed tgos: %v", err)
	}
	if len(completed) != 1 || completed[0].Code != "story-causal-clarity" {
		t.Fatalf("completed after migration = %#v", completed)
	}

	state, err := store.GetCurriculumState(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("get curriculum state: %v", err)
	}
	if state.CurrentFocus != "story-causal-clarity" {
		t.Fatalf("current focus after migration = %q", state.CurrentFocus)
	}

	exercise, err := store.GetExercise(ctx, exerciseID)
	if err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	if len(exercise.TGOCodes) != 1 || exercise.TGOCodes[0] != "story-causal-clarity" {
		t.Fatalf("exercise tgo codes after migration = %#v", exercise.TGOCodes)
	}
	if len(exercise.FocusSkills) != 1 || exercise.FocusSkills[0] != "narrative clarity" {
		t.Fatalf("exercise focus skills after migration = %#v", exercise.FocusSkills)
	}
}

func TestEnsureSeedDataMigratesLegacyCodesForGeneratedTreeMetadataAndExercises(t *testing.T) {
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

	for _, code := range []string{"literary-story-causal-clarity", "literary-story-scene-architecture", "literary-story-prose-precision"} {
		if _, err := store.SQL.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order, progress_mode)
			VALUES (?, ?, ?, 'core', 1, ?)
		`, code, code, code, domain.ProgressModePercent); err != nil {
			t.Fatalf("insert generated tree tgo %q: %v", code, err)
		}
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO tgo_trees (slug, title, description, seed_codes_json, priority_skills_json)
		VALUES (?, 'Generated Literary Track', 'legacy generated track', ?, ?)
	`, "kratos-legacy-track", mustJSON([]string{"causal-clarity", "scene-architecture", "prose-precision"}), mustJSON([]string{"narrative clarity"})); err != nil {
		t.Fatalf("insert generated tree: %v", err)
	}
	var treeID int64
	if err := store.SQL.QueryRowContext(ctx, `SELECT id FROM tgo_trees WHERE slug = ?`, "kratos-legacy-track").Scan(&treeID); err != nil {
		t.Fatalf("tree id: %v", err)
	}
	for _, code := range []string{"literary-story-causal-clarity", "literary-story-scene-architecture", "literary-story-prose-precision"} {
		if _, err := store.SQL.ExecContext(ctx, `
			INSERT INTO tree_tgos (tree_id, tgo_code, prerequisites_json, mastery_hint)
			VALUES (?, ?, '[]', '')
		`, treeID, code); err != nil {
			t.Fatalf("insert tree tgo %q: %v", code, err)
		}
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO tree_versions (tree_id, version, title, description, seed_codes_json, priority_skills_json, tgos_json)
		VALUES (?, 1, 'Generated Literary Track', 'legacy generated track', ?, ?, ?)
	`, treeID, mustJSON([]string{"causal-clarity", "scene-architecture", "prose-precision"}), mustJSON([]string{"narrative clarity"}), mustJSON([]domain.TGO{
		{Code: "causal-clarity", Title: "Causal Clarity", Description: "legacy", Stage: "core", StageOrder: 1, ProgressMode: domain.ProgressModePercent},
		{Code: "scene-architecture", Title: "Scene Architecture", Description: "legacy", Stage: "core", StageOrder: 2, ProgressMode: domain.ProgressModePercent, Prerequisites: []string{"causal-clarity"}},
		{Code: "prose-precision", Title: "Prose Precision", Description: "legacy", Stage: "core", StageOrder: 3, ProgressMode: domain.ProgressModePercent, Prerequisites: []string{"scene-architecture"}},
	})); err != nil {
		t.Fatalf("insert legacy tree version: %v", err)
	}

	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO users (slug, name, active_tree_slug)
		VALUES ('generated-legacy-user', 'Generated Legacy User', ?)
	`, "kratos-legacy-track"); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var userID int64
	if err := store.SQL.QueryRowContext(ctx, `SELECT id FROM users WHERE slug = 'generated-legacy-user'`).Scan(&userID); err != nil {
		t.Fatalf("user id: %v", err)
	}
	res, err := store.SQL.ExecContext(ctx, `
		INSERT INTO exercises (
			user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json,
			success_criteria_json, generation_kind, provider_note
		) VALUES (?, ?, 'Legacy Generated Exercise', 'Brief', '[]', ?, ?, '[]', 'deterministic', '')
	`, userID, treeID, mustJSON([]string{"narrative clarity", "scene architecture", "prose precision"}), mustJSON([]string{"causal-clarity", "scene-architecture", "prose-precision"}))
	if err != nil {
		t.Fatalf("insert exercise: %v", err)
	}
	exerciseID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("exercise id: %v", err)
	}

	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("ensure seed data: %v", err)
	}

	def, err := store.TreeDefinitionBySlug(ctx, "kratos-legacy-track")
	if err != nil {
		t.Fatalf("tree definition by slug: %v", err)
	}
	if !slices.Equal(def.SeedCodes, []string{"literary-story-causal-clarity", "literary-story-scene-architecture", "literary-story-prose-precision"}) {
		t.Fatalf("seed codes after migration = %#v", def.SeedCodes)
	}

	_, versionDef, err := store.TreeVersionByNumber(ctx, "kratos-legacy-track", 1)
	if err != nil {
		t.Fatalf("tree version by number: %v", err)
	}
	if !slices.Equal(versionDef.SeedCodes, []string{"literary-story-causal-clarity", "literary-story-scene-architecture", "literary-story-prose-precision"}) {
		t.Fatalf("version seed codes after migration = %#v", versionDef.SeedCodes)
	}
	if len(versionDef.TGOs) != 3 || versionDef.TGOs[0].Code != "literary-story-causal-clarity" {
		t.Fatalf("version tgos after migration = %#v", versionDef.TGOs)
	}
	if len(versionDef.TGOs) > 1 && (len(versionDef.TGOs[1].Prerequisites) != 1 || versionDef.TGOs[1].Prerequisites[0] != "literary-story-causal-clarity") {
		t.Fatalf("version prerequisites after migration = %#v", versionDef.TGOs[1].Prerequisites)
	}

	exercise, err := store.GetExercise(ctx, exerciseID)
	if err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	if !slices.Equal(exercise.TGOCodes, []string{"literary-story-causal-clarity", "literary-story-scene-architecture", "literary-story-prose-precision"}) {
		t.Fatalf("exercise tgo codes after migration = %#v", exercise.TGOCodes)
	}
	if !slices.Equal(exercise.FocusSkills, []string{"narrative clarity", "scene architecture", "prose precision"}) {
		t.Fatalf("exercise focus skills after migration = %#v", exercise.FocusSkills)
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
