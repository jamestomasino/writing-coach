package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestCalibrationRunAndNotificationLifecycle(t *testing.T) {
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
	userID, treeID, _, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}

	exerciseID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Calibration sample",
		Brief:           "Draft for calibration",
		Constraints:     []string{"single paragraph"},
		FocusSkills:     []string{"narrative clarity"},
		TGOCodes:        []string{"story-causal-clarity"},
		SuccessCriteria: []string{"clear causal chain"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := store.SaveSubmission(ctx, domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exerciseID,
		Content:    "A clear scene with a focused turn.",
		WordCount:  8,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO submission_skill_scores (submission_id, skill_name, score, score_source, score_version, score_evidence_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`, submissionID, "narrative clarity", 4, "deterministic", "det-v2", `{}`); err != nil {
		t.Fatalf("insert deterministic score: %v", err)
	}

	snapshots, err := store.ListCalibrationTrackSnapshots(ctx, domain.GlobalSkillGraphSlug, 50)
	if err != nil {
		t.Fatalf("list calibration snapshots: %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("expected at least one calibration snapshot")
	}

	run, err := store.CreateCalibrationRun(ctx, "manual", userID, 50, 200)
	if err != nil {
		t.Fatalf("create calibration run: %v", err)
	}
	run.Status = "succeeded"
	run.SubmissionCount = 1
	run.DeterministicScoreCount = 1
	run.TrackLearnings = []domain.CalibrationTrackLearning{{
		TreeSlug:                "story-craft-track",
		Domain:                  "fiction",
		SubmissionCount:         1,
		DeterministicScoreCount: 1,
		TopScoreRate:            0,
		AverageScore:            4,
		Issues:                  []string{"insufficient_samples"},
	}}
	run.DomainLearnings = []domain.CalibrationDomainLearning{{
		Domain:                  "fiction",
		TrackCount:              1,
		SubmissionCount:         1,
		DeterministicScoreCount: 1,
		TopScoreRate:            0,
		AverageScore:            4,
		Issues:                  []string{"insufficient_samples"},
	}}
	run.Highlights = []string{"track under min samples"}
	run.Recommendations = []string{"collect more data"}
	run.CompletedAt = time.Now().UTC()
	if err := store.FinalizeCalibrationRun(ctx, run); err != nil {
		t.Fatalf("finalize calibration run: %v", err)
	}

	runs, err := store.ListRecentCalibrationRuns(ctx, 10)
	if err != nil {
		t.Fatalf("list calibration runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("calibration runs len = %d", len(runs))
	}
	if runs[0].Status != "succeeded" {
		t.Fatalf("run status = %q", runs[0].Status)
	}
	if runs[0].ApprovalStatus != "pending" {
		t.Fatalf("approval status = %q", runs[0].ApprovalStatus)
	}
	if len(runs[0].TrackLearnings) != 1 {
		t.Fatalf("track learnings len = %d", len(runs[0].TrackLearnings))
	}
	if err := store.UpdateCalibrationRunApproval(ctx, run.ID, "approved", userID, "looks good"); err != nil {
		t.Fatalf("update approval: %v", err)
	}
	runs, err = store.ListRecentCalibrationRuns(ctx, 10)
	if err != nil {
		t.Fatalf("list calibration runs after approval: %v", err)
	}
	if runs[0].ApprovalStatus != "approved" {
		t.Fatalf("approval status after update = %q", runs[0].ApprovalStatus)
	}
	if runs[0].ApprovedByUserID != userID {
		t.Fatalf("approved_by_user_id = %d", runs[0].ApprovedByUserID)
	}
	if runs[0].ApprovalNotes != "looks good" {
		t.Fatalf("approval_notes = %q", runs[0].ApprovalNotes)
	}

	if err := store.SaveAdminNotification(ctx, domain.AdminNotification{
		Kind:         "calibration_run_completed",
		Title:        "Calibration run complete",
		Body:         "track under min samples",
		PayloadJSON:  `{"run_id":1}`,
		RelatedRunID: run.ID,
		CreatedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save notification: %v", err)
	}

	unread, err := store.CountUnreadAdminNotifications(ctx)
	if err != nil {
		t.Fatalf("count unread notifications: %v", err)
	}
	if unread != 1 {
		t.Fatalf("unread notifications = %d", unread)
	}
	if err := store.MarkCalibrationRunNotificationsRead(ctx, run.ID); err != nil {
		t.Fatalf("mark run notifications read: %v", err)
	}
	unread, err = store.CountUnreadAdminNotifications(ctx)
	if err != nil {
		t.Fatalf("count unread notifications after read: %v", err)
	}
	if unread != 0 {
		t.Fatalf("unread notifications after read = %d", unread)
	}
}

func TestListCalibrationTrackSnapshotsPrefersObjectiveScoresWhenPresent(t *testing.T) {
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
	userID, treeID, _, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}
	exerciseID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Calibration objective sample",
		Brief:           "Draft for objective calibration",
		Constraints:     []string{"single paragraph"},
		FocusSkills:     []string{"narrative clarity"},
		TGOCodes:        []string{"story-causal-clarity", "story-scene-architecture"},
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
		Content:    "A clear scene with a focused turn.",
		WordCount:  8,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	// Insert a legacy deterministic family score; objective rows should take precedence.
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO submission_skill_scores (submission_id, skill_name, score, score_source, score_version, score_evidence_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`, submissionID, "narrative clarity", 1, "deterministic", "det-v2", `{}`); err != nil {
		t.Fatalf("insert deterministic legacy score: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO submission_objective_scores (submission_id, tgo_code, score, score_source, score_version, score_evidence_json)
		VALUES (?, ?, ?, 'deterministic', 'obj-det-v1', '{}'),
		       (?, ?, ?, 'deterministic', 'obj-det-v1', '{}')
	`, submissionID, "story-causal-clarity", 4, submissionID, "story-scene-architecture", 5); err != nil {
		t.Fatalf("insert deterministic objective scores: %v", err)
	}

	snapshots, err := store.ListCalibrationTrackSnapshots(ctx, domain.GlobalSkillGraphSlug, 50)
	if err != nil {
		t.Fatalf("list calibration snapshots: %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("expected at least one calibration snapshot")
	}
	var found *domain.CalibrationTrackSnapshot
	for i := range snapshots {
		if snapshots[i].TreeSlug == "story-craft-track" {
			found = &snapshots[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected snapshot for story-craft-track, got %+v", snapshots)
	}
	if found.SubmissionCount != 1 {
		t.Fatalf("submission count = %d", found.SubmissionCount)
	}
	if found.DeterministicScoreCount != 2 {
		t.Fatalf("expected objective-preferred score count 2, got %d", found.DeterministicScoreCount)
	}
	if found.TopScoreCount != 1 {
		t.Fatalf("top score count = %d", found.TopScoreCount)
	}
	if found.AverageScore < 4.49 || found.AverageScore > 4.51 {
		t.Fatalf("average score = %.2f", found.AverageScore)
	}
}

func TestListCalibrationHybridSignalSnapshotsPrefersObjectiveHybridWhenPresent(t *testing.T) {
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
	userID, treeID, _, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "story-craft-track")
	if err != nil {
		t.Fatalf("default user tree: %v", err)
	}
	exerciseID, err := store.SaveExercise(ctx, domain.Exercise{
		UserID:          userID,
		TreeID:          treeID,
		Title:           "Calibration hybrid sample",
		Brief:           "Draft for hybrid calibration",
		Constraints:     []string{"single paragraph"},
		FocusSkills:     []string{"narrative clarity"},
		TGOCodes:        []string{"story-causal-clarity", "story-scene-architecture"},
		SuccessCriteria: []string{"clear causality"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}

	// Submission 1: mixed-era data. Objective hybrid rows should be preferred.
	submission1, err := store.SaveSubmission(ctx, domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exerciseID,
		Content:    "Submission one",
		WordCount:  2,
	})
	if err != nil {
		t.Fatalf("save submission1: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO submission_skill_scores (submission_id, skill_name, score, score_source, score_version, score_evidence_json)
		VALUES
			(?, 'narrative clarity', 3, 'hybrid', 'hyb-v1', '{"conflict":1,"applied_delta":1}'),
			(?, 'scene architecture', 4, 'hybrid', 'hyb-v1', '{"conflict":0,"applied_delta":0}')
	`, submission1, submission1); err != nil {
		t.Fatalf("insert submission1 skill hybrid rows: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO submission_objective_scores (submission_id, tgo_code, score, score_source, score_version, score_evidence_json)
		VALUES
			(?, 'story-causal-clarity', 4, 'hybrid', 'obj-hyb-v1', '{"conflict":1,"applied_delta":1}'),
			(?, 'story-scene-architecture', 4, 'hybrid', 'obj-hyb-v1', '{"conflict":0,"applied_delta":0}')
	`, submission1, submission1); err != nil {
		t.Fatalf("insert submission1 objective hybrid rows: %v", err)
	}

	// Submission 2: legacy-only hybrid data should still count via fallback.
	submission2, err := store.SaveSubmission(ctx, domain.Submission{
		UserID:     userID,
		TreeID:     treeID,
		ExerciseID: exerciseID,
		Content:    "Submission two",
		WordCount:  2,
	})
	if err != nil {
		t.Fatalf("save submission2: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `
		INSERT INTO submission_skill_scores (submission_id, skill_name, score, score_source, score_version, score_evidence_json)
		VALUES
			(?, 'prose precision', 4, 'hybrid', 'hyb-v1', '{"conflict":1,"applied_delta":2}'),
			(?, 'prose precision', 4, 'hybrid', 'hyb-v1', '{"conflict":0,"applied_delta":0}'),
			(?, 'image freshness', 3, 'hybrid', 'hyb-v1', '{"conflict":0,"applied_delta":0}')
	`, submission2, submission2, submission2); err != nil {
		t.Fatalf("insert submission2 skill hybrid rows: %v", err)
	}

	snapshots, err := store.ListCalibrationHybridSignalSnapshots(ctx, domain.GlobalSkillGraphSlug, 50)
	if err != nil {
		t.Fatalf("list calibration hybrid snapshots: %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("expected at least one hybrid snapshot")
	}
	var found *domain.CalibrationHybridSignalSnapshot
	for i := range snapshots {
		if snapshots[i].TreeSlug == "story-craft-track" {
			found = &snapshots[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected snapshot for story-craft-track, got %+v", snapshots)
	}
	// Submission 1 contributes 2 objective rows; submission 2 contributes latest 2 skill rows.
	if found.HybridScoreCount != 4 {
		t.Fatalf("hybrid score count = %d", found.HybridScoreCount)
	}
	// Conflicts should include only objective-preferred rows for submission 1 plus latest legacy rows for submission 2.
	if found.ConflictCount != 1 {
		t.Fatalf("conflict count = %d", found.ConflictCount)
	}
	// Adjusted includes applied_delta != 0 rows only.
	if found.AdjustedCount != 1 {
		t.Fatalf("adjusted count = %d", found.AdjustedCount)
	}
}
