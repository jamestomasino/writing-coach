package db

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/tomasino/writing-coach/internal/domain"
)

type Store struct {
	SQL *sql.DB
}

type Options struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Open(databasePath string) (*Store, error) {
	return OpenWithOptions(databasePath, defaultOptions(databasePath))
}

func OpenWithOptions(databasePath string, opts Options) (*Store, error) {
	sqlDB, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, err
	}

	maxOpenConns := opts.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 4
	}
	maxIdleConns := opts.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = maxOpenConns
	}
	if maxIdleConns > maxOpenConns {
		maxIdleConns = maxOpenConns
	}
	connMaxLifetime := opts.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = 30 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	setupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pragmas := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA synchronous = NORMAL`,
		`PRAGMA busy_timeout = 5000`,
	}
	for _, stmt := range pragmas {
		if _, err := sqlDB.ExecContext(setupCtx, stmt); err != nil {
			_ = sqlDB.Close()
			return nil, err
		}
	}

	return &Store{SQL: sqlDB}, nil
}

func defaultOptions(databasePath string) Options {
	if databasePath == ":memory:" || strings.Contains(databasePath, "mode=memory") {
		return Options{
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: 30 * time.Minute,
		}
	}
	return Options{
		MaxOpenConns:    4,
		MaxIdleConns:    4,
		ConnMaxLifetime: 30 * time.Minute,
	}
}

func (s *Store) Close() error {
	if s == nil || s.SQL == nil {
		return nil
	}
	return s.SQL.Close()
}

func (s *Store) EnsureSeedData(ctx context.Context, writerName string) error {
	if _, err := s.SQL.ExecContext(ctx, `
		INSERT INTO writer_profile (name, aesthetic_target)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM writer_profile)
	`, writerName, "Writing Coach: story craft across fiction, nonfiction, technical, academic, and professional writing."); err != nil {
		return err
	}

	if _, err := s.SQL.ExecContext(ctx, `
		INSERT INTO curriculum_state (id, current_focus, difficulty_level, updated_at)
		SELECT 1, 'narrative clarity', 2, CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM curriculum_state WHERE id = 1)
	`); err != nil {
		return err
	}

	for _, skill := range domain.SupportedSkills {
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO skill_dimensions (name)
			SELECT ?
			WHERE NOT EXISTS (SELECT 1 FROM skill_dimensions WHERE name = ?)
		`, skill, skill); err != nil {
			return err
		}
	}

	for _, tgo := range domain.AllTGOs() {
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order)
			SELECT ?, ?, ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM tgo_catalog WHERE code = ?)
		`, tgo.Code, tgo.Title, tgo.Description, tgo.Stage, tgo.StageOrder, tgo.Code); err != nil {
			return err
		}
		if _, err := s.SQL.ExecContext(ctx, `
			UPDATE tgo_catalog
			SET title = ?, description = ?, stage = ?, stage_order = ?
			WHERE code = ?
		`, tgo.Title, tgo.Description, tgo.Stage, tgo.StageOrder, tgo.Code); err != nil {
			return err
		}
	}

	for _, tree := range domain.BuiltInTrees {
		if err := s.SaveTreeDefinition(ctx, tree); err != nil {
			return err
		}
	}
	if err := s.SaveTreeDefinition(ctx, domain.GlobalSkillGraphDefinition()); err != nil {
		return err
	}

	return nil
}

func (s *Store) ListEnrollments(ctx context.Context) ([]domain.Enrollment, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT e.id, e.user_id, e.tree_id, u.slug, t.slug, e.created_at
		FROM user_tree_enrollments e
		JOIN users u ON u.id = e.user_id
		JOIN tgo_trees t ON t.id = e.tree_id
		WHERE e.archived_at IS NULL
		ORDER BY u.slug ASC, t.slug ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []domain.Enrollment
	for rows.Next() {
		var enrollment domain.Enrollment
		if err := rows.Scan(&enrollment.ID, &enrollment.UserID, &enrollment.TreeID, &enrollment.UserSlug, &enrollment.TreeSlug, &enrollment.CreatedAt); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, enrollment)
	}
	return enrollments, rows.Err()
}

func (s *Store) ListEnrollmentsByUserID(ctx context.Context, userID int64) ([]domain.Enrollment, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT e.id, e.user_id, e.tree_id, u.slug, t.slug, e.created_at
		FROM user_tree_enrollments e
		JOIN users u ON u.id = e.user_id
		JOIN tgo_trees t ON t.id = e.tree_id
		WHERE e.user_id = ? AND e.archived_at IS NULL
		ORDER BY e.created_at ASC, e.id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []domain.Enrollment
	for rows.Next() {
		var enrollment domain.Enrollment
		if err := rows.Scan(&enrollment.ID, &enrollment.UserID, &enrollment.TreeID, &enrollment.UserSlug, &enrollment.TreeSlug, &enrollment.CreatedAt); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, enrollment)
	}
	return enrollments, rows.Err()
}

func (s *Store) EnrollmentByID(ctx context.Context, enrollmentID int64) (domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := s.SQL.QueryRowContext(ctx, `
		SELECT e.id, e.user_id, e.tree_id, u.slug, t.slug, e.created_at
		FROM user_tree_enrollments e
		JOIN users u ON u.id = e.user_id
		JOIN tgo_trees t ON t.id = e.tree_id
		WHERE e.id = ? AND e.archived_at IS NULL
	`, enrollmentID).Scan(&enrollment.ID, &enrollment.UserID, &enrollment.TreeID, &enrollment.UserSlug, &enrollment.TreeSlug, &enrollment.CreatedAt)
	return enrollment, err
}

func (s *Store) EnrollmentID(ctx context.Context, userID, treeID int64) (int64, error) {
	var id int64
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id FROM user_tree_enrollments WHERE user_id = ? AND tree_id = ? AND archived_at IS NULL
	`, userID, treeID).Scan(&id)
	return id, err
}

func (s *Store) ActiveEnrollmentIDByUserID(ctx context.Context, userID int64) (int64, error) {
	var enrollmentID int64
	err := s.SQL.QueryRowContext(ctx, `
		SELECT e.id
		FROM user_tree_enrollments e
		JOIN users u ON u.id = e.user_id
		JOIN tgo_trees t ON t.id = e.tree_id
		WHERE e.user_id = ? AND t.slug = u.active_tree_slug AND e.archived_at IS NULL
	`, userID).Scan(&enrollmentID)
	return enrollmentID, err
}

func (s *Store) ListUserTracks(ctx context.Context, userID int64) ([]domain.UserTrack, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT e.id, t.id, t.slug, t.title, t.description, CASE WHEN u.active_tree_slug = t.slug THEN 1 ELSE 0 END, e.created_at
		FROM user_tree_enrollments e
		JOIN tgo_trees t ON t.id = e.tree_id
		JOIN users u ON u.id = e.user_id
		WHERE e.user_id = ? AND e.archived_at IS NULL
		ORDER BY CASE WHEN u.active_tree_slug = t.slug THEN 0 ELSE 1 END, e.created_at ASC, e.id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []domain.UserTrack
	for rows.Next() {
		var track domain.UserTrack
		var active int
		if err := rows.Scan(&track.EnrollmentID, &track.TreeID, &track.TreeSlug, &track.Title, &track.Description, &active, &track.CreatedAt); err != nil {
			return nil, err
		}
		track.IsActive = active == 1
		tracks = append(tracks, track)
	}
	return tracks, rows.Err()
}

func (s *Store) ArchiveUserTrack(ctx context.Context, userID int64, treeSlug string) error {
	res, err := s.SQL.ExecContext(ctx, `
		UPDATE user_tree_enrollments
		SET archived_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND tree_id = (SELECT id FROM tgo_trees WHERE slug = ?) AND archived_at IS NULL
	`, userID, strings.TrimSpace(treeSlug))
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetCurriculumState(ctx context.Context, enrollmentID int64) (domain.CurriculumState, error) {
	var state domain.CurriculumState
	var holdActive int
	var holdUpdatedAt sql.NullTime
	err := s.SQL.QueryRowContext(ctx, `
		SELECT
			enrollment_id,
			current_focus,
			difficulty_level,
			COALESCE(last_review_id, 0),
			COALESCE(progression_hold_active, 0),
			COALESCE(progression_hold_reason_code, ''),
			COALESCE(hold_trigger_review_id, 0),
			COALESCE(hold_cleared_review_id, 0),
			COALESCE(hold_clear_streak, 0),
			hold_updated_at,
			updated_at
		FROM user_curriculum_state
		WHERE enrollment_id = ?
	`, enrollmentID).Scan(
		&state.ID,
		&state.CurrentFocus,
		&state.DifficultyLevel,
		&state.LastReviewID,
		&holdActive,
		&state.ProgressionHoldReasonCode,
		&state.HoldTriggerReviewID,
		&state.HoldClearedReviewID,
		&state.HoldClearStreak,
		&holdUpdatedAt,
		&state.UpdatedAt,
	)
	if err != nil {
		return domain.CurriculumState{}, err
	}
	state.ProgressionHoldActive = holdActive == 1
	if holdUpdatedAt.Valid {
		state.HoldUpdatedAt = holdUpdatedAt.Time
	}
	return state, nil
}

func (s *Store) SaveExercise(ctx context.Context, ex domain.Exercise) (int64, error) {
	const query = `
		INSERT INTO exercises (
			user_id,
			tree_id,
			title,
			brief,
			constraints_json,
			focus_skills_json,
			tgo_codes_json,
			success_criteria_json,
			generation_kind,
			provider_note,
			source_submission_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := s.SQL.ExecContext(
		ctx,
		query,
		ex.UserID,
		ex.TreeID,
		ex.Title,
		ex.Brief,
		mustJSON(ex.Constraints),
		mustJSON(ex.FocusSkills),
		mustJSON(ex.TGOCodes),
		mustJSON(ex.SuccessCriteria),
		ex.GenerationKind,
		ex.ProviderNote,
		nullableID(ex.SourceSubmissionID),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Store) GetExercise(ctx context.Context, exerciseID int64) (domain.Exercise, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json, success_criteria_json, generation_kind, provider_note, COALESCE(source_submission_id, 0), closed_at, created_at
		FROM exercises
		WHERE id = ?
	`, exerciseID)
	if err != nil {
		return domain.Exercise{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Exercise{}, sql.ErrNoRows
	}
	exercise, err := scanExercise(rows)
	if err != nil {
		return domain.Exercise{}, err
	}
	return exercise, rows.Err()
}

func (s *Store) ListExercises(ctx context.Context, userID, treeID int64, limit int) ([]domain.Exercise, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json, success_criteria_json, generation_kind, provider_note, COALESCE(source_submission_id, 0), closed_at, created_at
		FROM exercises
		WHERE user_id = ? AND tree_id = ?
		ORDER BY id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []domain.Exercise
	for rows.Next() {
		exercise, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}
	return exercises, rows.Err()
}

func (s *Store) CloseExercise(ctx context.Context, userID, treeID, exerciseID int64) error {
	res, err := s.SQL.ExecContext(ctx, `
		UPDATE exercises
		SET closed_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND tree_id = ? AND closed_at IS NULL
	`, exerciseID, userID, treeID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) ReopenExercise(ctx context.Context, userID, treeID, exerciseID int64) error {
	res, err := s.SQL.ExecContext(ctx, `
		UPDATE exercises
		SET closed_at = NULL
		WHERE id = ? AND user_id = ? AND tree_id = ? AND closed_at IS NOT NULL
	`, exerciseID, userID, treeID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) SaveSubmission(ctx context.Context, sub domain.Submission) (int64, error) {
	if sub.DraftNumber == 0 {
		nextDraft, err := s.NextDraftNumber(ctx, sub.ExerciseID, sub.UserID, sub.TreeID)
		if err != nil {
			return 0, err
		}
		sub.DraftNumber = nextDraft
	}
	res, err := s.SQL.ExecContext(ctx, `
		INSERT INTO submissions (user_id, tree_id, exercise_id, parent_submission_id, draft_number, content, word_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sub.UserID, sub.TreeID, sub.ExerciseID, nullableID(sub.ParentSubmissionID), sub.DraftNumber, sub.Content, sub.WordCount)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Store) GetSubmission(ctx context.Context, submissionID int64) (domain.Submission, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE id = ?
	`, submissionID)
	if err != nil {
		return domain.Submission{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Submission{}, sql.ErrNoRows
	}
	submission, err := scanSubmission(rows)
	if err != nil {
		return domain.Submission{}, err
	}
	return submission, rows.Err()
}

func (s *Store) ListSubmissions(ctx context.Context, userID, treeID, exerciseID int64, limit int) ([]domain.Submission, error) {
	query := `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE user_id = ? AND tree_id = ?
	`
	args := []any{userID, treeID}
	if exerciseID != 0 {
		query += " AND exercise_id = ?"
		args = append(args, exerciseID)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []domain.Submission
	for rows.Next() {
		submission, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	return submissions, rows.Err()
}

func (s *Store) NextDraftNumber(ctx context.Context, exerciseID, userID, treeID int64) (int, error) {
	var next int
	err := s.SQL.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(draft_number), 0) + 1
		FROM submissions
		WHERE exercise_id = ? AND user_id = ? AND tree_id = ?
	`, exerciseID, userID, treeID).Scan(&next)
	if err != nil {
		return 0, err
	}
	if next == 0 {
		next = 1
	}
	return next, nil
}

func (s *Store) LatestSubmissionForExercise(ctx context.Context, exerciseID, userID, treeID int64) (domain.Submission, error) {
	var sub domain.Submission
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE exercise_id = ? AND user_id = ? AND tree_id = ?
		ORDER BY draft_number DESC, id DESC
		LIMIT 1
	`, exerciseID, userID, treeID).Scan(&sub.ID, &sub.UserID, &sub.TreeID, &sub.ExerciseID, &sub.ParentSubmissionID, &sub.DraftNumber, &sub.Content, &sub.WordCount, &sub.CreatedAt)
	if err != nil {
		return domain.Submission{}, err
	}
	return sub, nil
}

func (s *Store) PreviousSubmission(ctx context.Context, sub domain.Submission) (domain.Submission, error) {
	if sub.ParentSubmissionID != 0 {
		return s.GetSubmission(ctx, sub.ParentSubmissionID)
	}
	var prev domain.Submission
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE exercise_id = ? AND user_id = ? AND tree_id = ? AND draft_number < ?
		ORDER BY draft_number DESC, id DESC
		LIMIT 1
	`, sub.ExerciseID, sub.UserID, sub.TreeID, sub.DraftNumber).Scan(&prev.ID, &prev.UserID, &prev.TreeID, &prev.ExerciseID, &prev.ParentSubmissionID, &prev.DraftNumber, &prev.Content, &prev.WordCount, &prev.CreatedAt)
	if err != nil {
		return domain.Submission{}, err
	}
	return prev, nil
}

func (s *Store) SaveReview(ctx context.Context, review domain.Review, scores []domain.SkillScore) (int64, error) {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO reviews (
			user_id,
			tree_id,
			submission_id,
			review_kind,
			provider_note,
			summary,
			strengths_json,
			weaknesses_json,
			analyzer_findings_json,
			completed_tgo_checks_json,
			next_focus,
			metric_word_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		review.UserID,
		review.TreeID,
		review.SubmissionID,
		review.ReviewKind,
		review.ProviderNote,
		review.Summary,
		mustJSON(review.Strengths),
		mustJSON(review.Weaknesses),
		mustJSON(review.AnalyzerFindings),
		mustJSON(review.CompletedTGOChecks),
		review.NextFocus,
		review.MetricWordCount,
	)
	if err != nil {
		return 0, err
	}

	reviewID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, score := range scores {
		scoreSource := strings.TrimSpace(score.ScoreSource)
		if scoreSource == "" {
			scoreSource = "deterministic"
		}
		scoreVersion := strings.TrimSpace(score.ScoreVersion)
		if scoreVersion == "" {
			scoreVersion = "det-v1"
		}
		scoreEvidence := strings.TrimSpace(score.ScoreEvidenceJSON)
		if scoreEvidence == "" {
			scoreEvidence = "{}"
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO submission_skill_scores (submission_id, skill_name, score, score_source, score_version, score_evidence_json)
			VALUES (?, ?, ?, ?, ?, ?)
		`, review.SubmissionID, score.Skill, score.Score, scoreSource, scoreVersion, scoreEvidence); err != nil {
			return 0, err
		}
	}

	for _, assessment := range review.TGOAssessments {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO review_tgo_assessments (review_id, submission_id, tgo_code, status, evidence)
			VALUES (?, ?, ?, ?, ?)
		`, reviewID, review.SubmissionID, assessment.TGOCode, assessment.Status, assessment.Evidence); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return reviewID, nil
}

func (s *Store) SaveReviewArtifacts(ctx context.Context, artifacts domain.ReviewArtifacts) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO review_artifacts (review_id, analyzer_report_json, recommendation_json, comparison_json, annotations_json)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(review_id) DO UPDATE SET
			analyzer_report_json = excluded.analyzer_report_json,
			recommendation_json = excluded.recommendation_json,
			comparison_json = excluded.comparison_json,
			annotations_json = excluded.annotations_json
	`, artifacts.ReviewID, artifacts.AnalyzerReportJSON, artifacts.RecommendationJSON, artifacts.ComparisonJSON, artifacts.AnnotationsJSON)
	return err
}

func (s *Store) ActiveTGOs(ctx context.Context, enrollmentID int64) ([]domain.TGO, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, c.progress_mode, a.slot, tt.prerequisites_json, tt.mastery_hint
		FROM enrollment_active_tgos a
		JOIN tgo_catalog c ON c.code = a.tgo_code
		LEFT JOIN user_tree_enrollments e ON e.id = a.enrollment_id
		LEFT JOIN tree_tgos tt ON tt.tree_id = e.tree_id AND tt.tgo_code = c.code
		WHERE a.enrollment_id = ?
		ORDER BY a.slot ASC
	`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tgos []domain.TGO
	for rows.Next() {
		var tgo domain.TGO
		var prereqsJSON sql.NullString
		var masteryHint sql.NullString
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ProgressMode, &tgo.ActiveSlot, &prereqsJSON, &masteryHint); err != nil {
			return nil, err
		}
		if prereqsJSON.Valid {
			if tgo.Prerequisites, err = DecodeStringSlice(prereqsJSON.String); err != nil {
				return nil, err
			}
		} else {
			tgo.Prerequisites = []string{}
		}
		tgo.MasteryHint = masteryHint.String
		tgos = append(tgos, tgo)
	}
	return tgos, rows.Err()
}

func (s *Store) CompletedTGOs(ctx context.Context, enrollmentID int64) ([]domain.TGO, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, c.progress_mode, 0, tt.prerequisites_json, tt.mastery_hint
		FROM enrollment_completed_tgos x
		JOIN tgo_catalog c ON c.code = x.tgo_code
		LEFT JOIN user_tree_enrollments e ON e.id = x.enrollment_id
		LEFT JOIN tree_tgos tt ON tt.tree_id = e.tree_id AND tt.tgo_code = c.code
		WHERE x.enrollment_id = ?
		ORDER BY x.completed_at ASC
	`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tgos []domain.TGO
	for rows.Next() {
		var tgo domain.TGO
		var prereqsJSON sql.NullString
		var masteryHint sql.NullString
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ProgressMode, &tgo.ActiveSlot, &prereqsJSON, &masteryHint); err != nil {
			return nil, err
		}
		if prereqsJSON.Valid {
			if tgo.Prerequisites, err = DecodeStringSlice(prereqsJSON.String); err != nil {
				return nil, err
			}
		} else {
			tgo.Prerequisites = []string{}
		}
		tgo.MasteryHint = masteryHint.String
		tgos = append(tgos, tgo)
	}
	return tgos, rows.Err()
}

func (s *Store) RecentTGOStatuses(ctx context.Context, enrollmentID int64, code string, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT status
		FROM review_tgo_assessments
		WHERE tgo_code = ? AND EXISTS (
			SELECT 1 FROM reviews r WHERE r.id = review_tgo_assessments.review_id AND r.user_id = (
				SELECT user_id FROM user_tree_enrollments WHERE id = ?
			) AND r.tree_id = (
				SELECT tree_id FROM user_tree_enrollments WHERE id = ?
			)
		)
		ORDER BY id DESC
		LIMIT ?
	`, code, enrollmentID, enrollmentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return nil, err
		}
		out = append(out, status)
	}
	return out, rows.Err()
}

func (s *Store) TGOMasterySignal(ctx context.Context, enrollmentID int64, tgo domain.TGO, currentStatus string) (domain.TGOMasterySignal, error) {
	statuses, err := s.RecentTGOStatuses(ctx, enrollmentID, tgo.Code, 5)
	if err != nil {
		return domain.TGOMasterySignal{}, err
	}
	if currentStatus != "" {
		statuses = append([]string{currentStatus}, statuses...)
	}
	return domain.ComputeMasterySignal(tgo.ProgressMode, statuses), nil
}

func (s *Store) ReplaceActiveTGO(ctx context.Context, enrollmentID int64, slot int, completedCode string, nextCode string) error {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM enrollment_completed_tgos WHERE enrollment_id = ? AND tgo_code = ?)
	`, enrollmentID, completedCode, enrollmentID, completedCode); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE enrollment_active_tgos SET tgo_code = ?, activated_at = CURRENT_TIMESTAMP WHERE enrollment_id = ? AND slot = ?
	`, nextCode, enrollmentID, slot); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) MarkTGOCompleted(ctx context.Context, enrollmentID int64, code string) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM enrollment_completed_tgos WHERE enrollment_id = ? AND tgo_code = ?)
	`, enrollmentID, code, enrollmentID, code)
	return err
}

func (s *Store) SetActiveTGOs(ctx context.Context, enrollmentID int64, codes []string) error {
	if len(codes) != 3 {
		return fmt.Errorf("exactly 3 TGO codes are required")
	}
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM enrollment_active_tgos WHERE enrollment_id = ?`, enrollmentID); err != nil {
		return err
	}
	for i, code := range codes {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code)
			VALUES (?, ?, ?)
		`, enrollmentID, i+1, code); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) NextAvailableTGO(ctx context.Context, enrollmentID int64) (domain.TGO, error) {
	var tgo domain.TGO
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, code, title, description, stage, stage_order
		FROM tgo_catalog
		WHERE code NOT IN (SELECT tgo_code FROM enrollment_active_tgos WHERE enrollment_id = ?)
		  AND code NOT IN (SELECT tgo_code FROM enrollment_completed_tgos WHERE enrollment_id = ?)
		ORDER BY stage_order ASC
		LIMIT 1
	`, enrollmentID, enrollmentID).Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder)
	if err != nil {
		return domain.TGO{}, err
	}
	return tgo, nil
}

func (s *Store) LatestReviewForSubmission(ctx context.Context, submissionID int64) (domain.Review, error) {
	var review domain.Review
	var strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, submission_id, review_kind, provider_note, summary, strengths_json, weaknesses_json, analyzer_findings_json, completed_tgo_checks_json, next_focus, metric_word_count, created_at
		FROM reviews
		WHERE submission_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, submissionID).Scan(
		&review.ID,
		&review.UserID,
		&review.TreeID,
		&review.SubmissionID,
		&review.ReviewKind,
		&review.ProviderNote,
		&review.Summary,
		&strengthsJSON,
		&weaknessesJSON,
		&findingsJSON,
		&completedChecksJSON,
		&review.NextFocus,
		&review.MetricWordCount,
		&review.CreatedAt,
	)
	if err != nil {
		return domain.Review{}, err
	}
	review, err = hydrateReview(review, strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON)
	if err != nil {
		return domain.Review{}, err
	}
	review.TGOAssessments, err = s.ReviewTGOAssessments(ctx, review.ID)
	if err != nil {
		return domain.Review{}, err
	}
	review.SkillScores, err = s.SubmissionSkillScores(ctx, review.SubmissionID)
	if err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

func (s *Store) GetReview(ctx context.Context, reviewID int64) (domain.Review, error) {
	var review domain.Review
	var strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, submission_id, review_kind, provider_note, summary, strengths_json, weaknesses_json, analyzer_findings_json, completed_tgo_checks_json, next_focus, metric_word_count, created_at
		FROM reviews
		WHERE id = ?
	`, reviewID).Scan(
		&review.ID,
		&review.UserID,
		&review.TreeID,
		&review.SubmissionID,
		&review.ReviewKind,
		&review.ProviderNote,
		&review.Summary,
		&strengthsJSON,
		&weaknessesJSON,
		&findingsJSON,
		&completedChecksJSON,
		&review.NextFocus,
		&review.MetricWordCount,
		&review.CreatedAt,
	)
	if err != nil {
		return domain.Review{}, err
	}
	review, err = hydrateReview(review, strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON)
	if err != nil {
		return domain.Review{}, err
	}
	review.TGOAssessments, err = s.ReviewTGOAssessments(ctx, review.ID)
	if err != nil {
		return domain.Review{}, err
	}
	review.SkillScores, err = s.SubmissionSkillScores(ctx, review.SubmissionID)
	if err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

func (s *Store) ListReviews(ctx context.Context, userID, treeID, submissionID int64, limit int) ([]domain.Review, error) {
	query := `
		SELECT id, user_id, tree_id, submission_id, review_kind, provider_note, summary, strengths_json, weaknesses_json, analyzer_findings_json, completed_tgo_checks_json, next_focus, metric_word_count, created_at
		FROM reviews
		WHERE user_id = ? AND tree_id = ?
	`
	args := []any{userID, treeID}
	if submissionID != 0 {
		query += " AND submission_id = ?"
		args = append(args, submissionID)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range reviews {
		assessments, err := s.ReviewTGOAssessments(ctx, reviews[i].ID)
		if err != nil {
			return nil, err
		}
		reviews[i].TGOAssessments = assessments
		reviews[i].SkillScores, err = s.SubmissionSkillScores(ctx, reviews[i].SubmissionID)
		if err != nil {
			return nil, err
		}
	}
	return reviews, nil
}

func (s *Store) SubmissionSkillScores(ctx context.Context, submissionID int64) ([]domain.SkillScore, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT submission_id, skill_name, score, score_source, score_version, score_evidence_json
		FROM submission_skill_scores
		WHERE submission_id = ?
		ORDER BY score DESC, skill_name ASC
	`, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []domain.SkillScore
	for rows.Next() {
		var score domain.SkillScore
		if err := rows.Scan(
			&score.SubmissionID,
			&score.Skill,
			&score.Score,
			&score.ScoreSource,
			&score.ScoreVersion,
			&score.ScoreEvidenceJSON,
		); err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}
	return scores, rows.Err()
}

func (s *Store) EnqueueReviewJob(ctx context.Context, job domain.ReviewJob) (domain.ReviewJob, error) {
	if job.MaxAttempts <= 0 {
		job.MaxAttempts = 3
	}
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO review_jobs (user_id, tree_id, enrollment_id, submission_id, status, max_attempts)
		VALUES (?, ?, ?, ?, 'queued', ?)
		ON CONFLICT(submission_id) DO UPDATE SET
			user_id = excluded.user_id,
			tree_id = excluded.tree_id,
			enrollment_id = excluded.enrollment_id,
			status = CASE
				WHEN review_jobs.review_id IS NOT NULL THEN 'completed'
				ELSE 'queued'
			END,
			last_error = '',
			updated_at = CURRENT_TIMESTAMP
	`, job.UserID, job.TreeID, job.EnrollmentID, job.SubmissionID, job.MaxAttempts)
	if err != nil {
		return domain.ReviewJob{}, err
	}
	return s.ReviewJobBySubmission(ctx, job.UserID, job.TreeID, job.SubmissionID)
}

func (s *Store) ReviewJobBySubmission(ctx context.Context, userID, treeID, submissionID int64) (domain.ReviewJob, error) {
	var job domain.ReviewJob
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, enrollment_id, submission_id, COALESCE(review_id, 0), status, attempt_count, max_attempts, last_error, created_at, updated_at
		FROM review_jobs
		WHERE user_id = ? AND tree_id = ? AND submission_id = ?
	`, userID, treeID, submissionID).Scan(
		&job.ID,
		&job.UserID,
		&job.TreeID,
		&job.EnrollmentID,
		&job.SubmissionID,
		&job.ReviewID,
		&job.Status,
		&job.AttemptCount,
		&job.MaxAttempts,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return domain.ReviewJob{}, err
	}
	return job, nil
}

func (s *Store) RequeueStaleReviewJobs(ctx context.Context, staleAfter time.Duration) error {
	if staleAfter <= 0 {
		staleAfter = 3 * time.Minute
	}
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE review_jobs
		SET status = 'queued',
			last_error = CASE
				WHEN last_error = '' THEN 'review worker was interrupted; retrying'
				ELSE last_error
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'running' AND updated_at <= ?
	`, time.Now().UTC().Add(-staleAfter))
	return err
}

func (s *Store) ClaimNextReviewJob(ctx context.Context) (domain.ReviewJob, error) {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return domain.ReviewJob{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var job domain.ReviewJob
	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, enrollment_id, submission_id, COALESCE(review_id, 0), status, attempt_count, max_attempts, last_error, created_at, updated_at
		FROM review_jobs
		WHERE status = 'queued'
		ORDER BY id ASC
		LIMIT 1
	`).Scan(
		&job.ID,
		&job.UserID,
		&job.TreeID,
		&job.EnrollmentID,
		&job.SubmissionID,
		&job.ReviewID,
		&job.Status,
		&job.AttemptCount,
		&job.MaxAttempts,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return domain.ReviewJob{}, err
		}
		return domain.ReviewJob{}, err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE review_jobs
		SET status = 'running',
			attempt_count = attempt_count + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, job.ID); err != nil {
		return domain.ReviewJob{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.ReviewJob{}, err
	}
	job.AttemptCount++
	job.Status = "running"
	job.UpdatedAt = time.Now().UTC()
	return job, nil
}

func (s *Store) CompleteReviewJob(ctx context.Context, jobID, reviewID int64) error {
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE review_jobs
		SET review_id = ?, status = 'completed', last_error = '', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, reviewID, jobID)
	return err
}

func (s *Store) FailReviewJob(ctx context.Context, job domain.ReviewJob, lastError string) error {
	status := "queued"
	if job.AttemptCount >= job.MaxAttempts {
		status = "failed"
	}
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE review_jobs
		SET status = ?, last_error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, lastError, job.ID)
	return err
}

func (s *Store) ReviewTGOAssessments(ctx context.Context, reviewID int64) ([]domain.TGOAssessment, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT tgo_code, status, evidence
		FROM review_tgo_assessments
		WHERE review_id = ?
		ORDER BY id ASC
	`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assessments []domain.TGOAssessment
	for rows.Next() {
		var assessment domain.TGOAssessment
		if err := rows.Scan(&assessment.TGOCode, &assessment.Status, &assessment.Evidence); err != nil {
			return nil, err
		}
		assessment.ReviewID = reviewID
		assessments = append(assessments, assessment)
	}
	return assessments, rows.Err()
}

func (s *Store) GetReviewArtifacts(ctx context.Context, reviewID int64) (domain.ReviewArtifacts, error) {
	var artifacts domain.ReviewArtifacts
	err := s.SQL.QueryRowContext(ctx, `
		SELECT review_id, analyzer_report_json, recommendation_json, comparison_json, annotations_json, created_at
		FROM review_artifacts
		WHERE review_id = ?
	`, reviewID).Scan(
		&artifacts.ReviewID,
		&artifacts.AnalyzerReportJSON,
		&artifacts.RecommendationJSON,
		&artifacts.ComparisonJSON,
		&artifacts.AnnotationsJSON,
		&artifacts.CreatedAt,
	)
	if err != nil {
		return domain.ReviewArtifacts{}, err
	}
	return artifacts, nil
}

func (s *Store) UpdateCurriculumState(ctx context.Context, enrollmentID int64, focus string, difficulty int, reviewID int64) error {
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE user_curriculum_state
		SET current_focus = ?, difficulty_level = ?, last_review_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE enrollment_id = ?
	`, focus, difficulty, reviewID, enrollmentID)
	return err
}

func (s *Store) UpdateProgressionHoldState(ctx context.Context, enrollmentID int64, holdRequested bool, reasonCode string, reviewID int64, clearStreakRequired int) error {
	if clearStreakRequired <= 0 {
		clearStreakRequired = 1
	}
	if holdRequested {
		_, err := s.SQL.ExecContext(ctx, `
			UPDATE user_curriculum_state
			SET
				progression_hold_active = 1,
				progression_hold_reason_code = ?,
				hold_trigger_review_id = CASE
					WHEN progression_hold_active = 1 AND hold_trigger_review_id IS NOT NULL THEN hold_trigger_review_id
					ELSE ?
				END,
				hold_cleared_review_id = NULL,
				hold_clear_streak = 0,
				hold_updated_at = CURRENT_TIMESTAMP
			WHERE enrollment_id = ?
		`, strings.TrimSpace(reasonCode), reviewID, enrollmentID)
		return err
	}

	_, err := s.SQL.ExecContext(ctx, `
		UPDATE user_curriculum_state
		SET
			hold_clear_streak = CASE
				WHEN progression_hold_active = 1 THEN CASE
					WHEN (hold_clear_streak + 1) >= ? THEN 0
					ELSE (hold_clear_streak + 1)
				END
				ELSE 0
			END,
			progression_hold_active = CASE
				WHEN progression_hold_active = 1 AND (hold_clear_streak + 1) >= ? THEN 0
				ELSE progression_hold_active
			END,
			progression_hold_reason_code = CASE
				WHEN progression_hold_active = 1 AND (hold_clear_streak + 1) >= ? THEN ''
				ELSE progression_hold_reason_code
			END,
			hold_cleared_review_id = CASE
				WHEN progression_hold_active = 1 AND (hold_clear_streak + 1) >= ? THEN ?
				ELSE hold_cleared_review_id
			END,
			hold_updated_at = CASE
				WHEN progression_hold_active = 1 AND (hold_clear_streak + 1) >= ? THEN CURRENT_TIMESTAMP
				ELSE hold_updated_at
			END
		WHERE enrollment_id = ?
	`, clearStreakRequired, clearStreakRequired, clearStreakRequired, clearStreakRequired, reviewID, clearStreakRequired, enrollmentID)
	return err
}
