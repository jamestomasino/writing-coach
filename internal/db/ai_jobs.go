package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) EnqueueAIJob(ctx context.Context, job domain.AIJob) (domain.AIJob, error) {
	if job.MaxAttempts <= 0 {
		job.MaxAttempts = 3
	}
	if job.ResourceKey != "" {
		existing, err := s.AIJobByResourceKey(ctx, job.ResourceKey)
		if err == nil {
			_, err = s.SQL.ExecContext(ctx, `
				UPDATE ai_jobs
				SET user_id = ?,
					tree_id = ?,
					enrollment_id = ?,
					submission_id = ?,
					payload_json = ?,
					status = CASE
						WHEN exercise_id IS NOT NULL OR review_id IS NOT NULL OR result_json <> '' THEN 'completed'
						ELSE 'queued'
					END,
					last_error = '',
					updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, job.UserID, job.TreeID, job.EnrollmentID, nullableID(job.SubmissionID), job.PayloadJSON, existing.ID)
			if err != nil {
				return domain.AIJob{}, err
			}
			return s.AIJobByID(ctx, existing.ID)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return domain.AIJob{}, err
		}
	}

	res, err := s.SQL.ExecContext(ctx, `
		INSERT INTO ai_jobs (
			user_id, tree_id, enrollment_id, kind, resource_key, submission_id, status, max_attempts, payload_json
		)
		VALUES (?, ?, ?, ?, ?, ?, 'queued', ?, ?)
	`, job.UserID, job.TreeID, job.EnrollmentID, job.Kind, job.ResourceKey, nullableID(job.SubmissionID), job.MaxAttempts, job.PayloadJSON)
	if err != nil {
		return domain.AIJob{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.AIJob{}, err
	}
	return s.AIJobByID(ctx, id)
}

func (s *Store) AIJobByID(ctx context.Context, id int64) (domain.AIJob, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, enrollment_id, kind, resource_key, COALESCE(exercise_id, 0), COALESCE(submission_id, 0), COALESCE(review_id, 0), status, attempt_count, max_attempts, last_error, payload_json, result_json, created_at, updated_at
		FROM ai_jobs
		WHERE id = ?
	`, id)
	if err != nil {
		return domain.AIJob{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.AIJob{}, sql.ErrNoRows
	}
	return scanAIJob(rows)
}

func (s *Store) AIJobByResourceKey(ctx context.Context, resourceKey string) (domain.AIJob, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, enrollment_id, kind, resource_key, COALESCE(exercise_id, 0), COALESCE(submission_id, 0), COALESCE(review_id, 0), status, attempt_count, max_attempts, last_error, payload_json, result_json, created_at, updated_at
		FROM ai_jobs
		WHERE resource_key = ?
	`, resourceKey)
	if err != nil {
		return domain.AIJob{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.AIJob{}, sql.ErrNoRows
	}
	return scanAIJob(rows)
}

func (s *Store) LatestAIJobBySubmission(ctx context.Context, userID, treeID, submissionID int64, kind string) (domain.AIJob, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, enrollment_id, kind, resource_key, COALESCE(exercise_id, 0), COALESCE(submission_id, 0), COALESCE(review_id, 0), status, attempt_count, max_attempts, last_error, payload_json, result_json, created_at, updated_at
		FROM ai_jobs
		WHERE user_id = ? AND tree_id = ? AND submission_id = ? AND kind = ?
		ORDER BY id DESC
		LIMIT 1
	`, userID, treeID, submissionID, kind)
	if err != nil {
		return domain.AIJob{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.AIJob{}, sql.ErrNoRows
	}
	return scanAIJob(rows)
}

func (s *Store) RequeueStaleAIJobs(ctx context.Context, staleAfter time.Duration) error {
	if staleAfter <= 0 {
		staleAfter = 3 * time.Minute
	}
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE ai_jobs
		SET status = 'queued',
			last_error = CASE
				WHEN last_error = '' THEN 'job worker was interrupted; retrying'
				ELSE last_error
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'running' AND updated_at <= ?
	`, time.Now().UTC().Add(-staleAfter))
	return err
}

func (s *Store) ClaimNextAIJob(ctx context.Context) (domain.AIJob, error) {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return domain.AIJob{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var job domain.AIJob
	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, enrollment_id, kind, resource_key, COALESCE(exercise_id, 0), COALESCE(submission_id, 0), COALESCE(review_id, 0), status, attempt_count, max_attempts, last_error, payload_json, result_json, created_at, updated_at
		FROM ai_jobs
		WHERE status = 'queued'
		ORDER BY id ASC
		LIMIT 1
	`).Scan(
		&job.ID,
		&job.UserID,
		&job.TreeID,
		&job.EnrollmentID,
		&job.Kind,
		&job.ResourceKey,
		&job.ExerciseID,
		&job.SubmissionID,
		&job.ReviewID,
		&job.Status,
		&job.AttemptCount,
		&job.MaxAttempts,
		&job.LastError,
		&job.PayloadJSON,
		&job.ResultJSON,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return domain.AIJob{}, err
		}
		return domain.AIJob{}, err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE ai_jobs
		SET status = 'running',
			attempt_count = attempt_count + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, job.ID); err != nil {
		return domain.AIJob{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.AIJob{}, err
	}
	job.AttemptCount++
	job.Status = "running"
	job.UpdatedAt = time.Now().UTC()
	return job, nil
}

func (s *Store) CompleteAIJob(ctx context.Context, jobID, exerciseID, reviewID int64, resultJSON string) error {
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE ai_jobs
		SET exercise_id = ?,
			review_id = ?,
			result_json = ?,
			status = 'completed',
			last_error = '',
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, nullableID(exerciseID), nullableID(reviewID), resultJSON, jobID)
	return err
}

func (s *Store) FailAIJob(ctx context.Context, job domain.AIJob, lastError string) error {
	status := "queued"
	if job.AttemptCount >= job.MaxAttempts {
		status = "failed"
	}
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE ai_jobs
		SET status = ?, last_error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, lastError, job.ID)
	return err
}
