package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) CreateCalibrationRun(ctx context.Context, runKind string, triggeredByUserID int64, minSamples, limitPerTrack int) (domain.CalibrationRun, error) {
	runKind = strings.TrimSpace(runKind)
	if runKind == "" {
		runKind = "scheduled"
	}
	if minSamples <= 0 {
		minSamples = 50
	}
	if limitPerTrack <= 0 {
		limitPerTrack = 200
	}

	now := time.Now().UTC()
	res, err := s.SQL.ExecContext(ctx, `
		INSERT INTO calibration_runs (
			run_kind,
			status,
			triggered_by_user_id,
			min_samples,
			limit_per_track,
			data_adequate,
			approval_status,
			started_at,
			updated_at
		)
		VALUES (?, 'running', ?, ?, ?, 0, 'pending', ?, ?)
	`, runKind, nullableID(triggeredByUserID), minSamples, limitPerTrack, now, now)
	if err != nil {
		return domain.CalibrationRun{}, err
	}
	runID, err := res.LastInsertId()
	if err != nil {
		return domain.CalibrationRun{}, err
	}
	return domain.CalibrationRun{
		ID:                runID,
		RunKind:           runKind,
		Status:            "running",
		TriggeredByUserID: triggeredByUserID,
		MinSamples:        minSamples,
		LimitPerTrack:     limitPerTrack,
		DataAdequate:      false,
		ApprovalStatus:    "pending",
		StartedAt:         now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (s *Store) FinalizeCalibrationRun(ctx context.Context, run domain.CalibrationRun) error {
	tracksJSON, err := json.Marshal(run.TrackLearnings)
	if err != nil {
		return err
	}
	domainsJSON, err := json.Marshal(run.DomainLearnings)
	if err != nil {
		return err
	}
	highlightsJSON, err := json.Marshal(run.Highlights)
	if err != nil {
		return err
	}
	recommendationsJSON, err := json.Marshal(run.Recommendations)
	if err != nil {
		return err
	}

	completedAt := run.CompletedAt.UTC()
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}
	_, err = s.SQL.ExecContext(ctx, `
		UPDATE calibration_runs
		SET status = ?,
			submission_count = ?,
			deterministic_score_count = ?,
			data_adequate = ?,
			track_learnings_json = ?,
			domain_learnings_json = ?,
			highlights_json = ?,
			recommendations_json = ?,
			error_text = ?,
			completed_at = ?,
			updated_at = ?
		WHERE id = ?
	`, strings.TrimSpace(run.Status), run.SubmissionCount, run.DeterministicScoreCount, boolToInt(run.DataAdequate), string(tracksJSON), string(domainsJSON), string(highlightsJSON), string(recommendationsJSON), strings.TrimSpace(run.ErrorText), completedAt, time.Now().UTC(), run.ID)
	return err
}

func (s *Store) UpdateCalibrationRunApproval(ctx context.Context, runID int64, status string, approvedByUserID int64, notes string) error {
	if runID <= 0 {
		return fmt.Errorf("invalid run id")
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "approved" && status != "rejected" && status != "pending" {
		return fmt.Errorf("invalid approval status")
	}
	approvedAt := nullableTime(time.Now().UTC())
	if status == "pending" {
		approvedAt = nil
	}
	approvedBy := nullableID(approvedByUserID)
	if status == "pending" {
		approvedBy = nil
	}
	res, err := s.SQL.ExecContext(ctx, `
		UPDATE calibration_runs
		SET approval_status = ?,
			approved_by_user_id = ?,
			approved_at = ?,
			approval_notes = ?,
			updated_at = ?
		WHERE id = ?
	`, status, approvedBy, approvedAt, strings.TrimSpace(notes), time.Now().UTC(), runID)
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

func (s *Store) ListRecentCalibrationRuns(ctx context.Context, limit int) ([]domain.CalibrationRun, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT
			id,
			run_kind,
			status,
			COALESCE(triggered_by_user_id, 0),
			min_samples,
			limit_per_track,
			submission_count,
			deterministic_score_count,
			data_adequate,
			approval_status,
			COALESCE(approved_by_user_id, 0),
			approved_at,
			approval_notes,
			track_learnings_json,
			domain_learnings_json,
			highlights_json,
			recommendations_json,
			error_text,
			started_at,
			completed_at,
			created_at,
			updated_at
		FROM calibration_runs
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []domain.CalibrationRun
	for rows.Next() {
		run, err := scanCalibrationRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *Store) SaveAdminNotification(ctx context.Context, notification domain.AdminNotification) error {
	if strings.TrimSpace(notification.Kind) == "" {
		notification.Kind = "general"
	}
	if strings.TrimSpace(notification.PayloadJSON) == "" {
		notification.PayloadJSON = "{}"
	}
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO admin_notifications (kind, title, body, payload_json, related_run_id, is_read, created_at, read_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, strings.TrimSpace(notification.Kind), strings.TrimSpace(notification.Title), strings.TrimSpace(notification.Body), strings.TrimSpace(notification.PayloadJSON), nullableID(notification.RelatedRunID), boolToInt(notification.IsRead), notification.CreatedAt.UTC(), nullableTime(notification.ReadAt))
	return err
}

func (s *Store) ListRecentAdminNotifications(ctx context.Context, limit int) ([]domain.AdminNotification, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, kind, title, body, payload_json, COALESCE(related_run_id, 0), is_read, created_at, read_at
		FROM admin_notifications
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.AdminNotification, 0, limit)
	for rows.Next() {
		item, err := scanAdminNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CountUnreadAdminNotifications(ctx context.Context) (int, error) {
	var count int
	err := s.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_notifications WHERE is_read = 0`).Scan(&count)
	return count, err
}

func (s *Store) MarkAdminNotificationRead(ctx context.Context, notificationID int64) error {
	if notificationID <= 0 {
		return fmt.Errorf("invalid notification id")
	}
	res, err := s.SQL.ExecContext(ctx, `
		UPDATE admin_notifications
		SET is_read = 1,
			read_at = COALESCE(read_at, CURRENT_TIMESTAMP)
		WHERE id = ?
	`, notificationID)
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

func (s *Store) MarkCalibrationRunNotificationsRead(ctx context.Context, runID int64) error {
	if runID <= 0 {
		return fmt.Errorf("invalid run id")
	}
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE admin_notifications
		SET is_read = 1,
			read_at = COALESCE(read_at, CURRENT_TIMESTAMP)
		WHERE related_run_id = ?
	`, runID)
	return err
}

func (s *Store) ListCalibrationTrackSnapshots(ctx context.Context, defaultTreeSlug string, limitPerTrack int) ([]domain.CalibrationTrackSnapshot, error) {
	if strings.TrimSpace(defaultTreeSlug) == "" {
		defaultTreeSlug = domain.GlobalSkillGraphSlug
	}
	if limitPerTrack <= 0 {
		limitPerTrack = 200
	}
	rows, err := s.SQL.QueryContext(ctx, `
		WITH ranked_scores AS (
			SELECT
				submission_id,
				skill_name,
				score,
				ROW_NUMBER() OVER (
					PARTITION BY submission_id, skill_name
					ORDER BY id DESC
				) AS rn
			FROM submission_skill_scores
			WHERE score_source = 'deterministic'
		),
		track_ranked AS (
			SELECT
				s.id AS submission_id,
				COALESCE(NULLIF(t.slug, ''), NULLIF(u.active_tree_slug, ''), ?) AS tree_slug,
				ROW_NUMBER() OVER (
					PARTITION BY COALESCE(NULLIF(t.slug, ''), NULLIF(u.active_tree_slug, ''), ?)
					ORDER BY s.id DESC
				) AS track_rank
			FROM submissions s
			LEFT JOIN users u ON u.id = s.user_id
			LEFT JOIN tgo_trees t ON t.id = s.tree_id
		),
		sampled AS (
			SELECT submission_id, tree_slug
			FROM track_ranked
			WHERE track_rank <= ?
		)
		SELECT
			sampled.tree_slug,
			COUNT(DISTINCT sampled.submission_id) AS submission_count,
			COUNT(rs.score) AS deterministic_score_count,
			COALESCE(SUM(CASE WHEN rs.score = 5 THEN 1 ELSE 0 END), 0) AS top_score_count,
			COALESCE(AVG(rs.score), 0) AS avg_score
		FROM sampled
		LEFT JOIN ranked_scores rs
			ON rs.submission_id = sampled.submission_id
			AND rs.rn = 1
		GROUP BY sampled.tree_slug
		ORDER BY submission_count DESC, sampled.tree_slug ASC
	`, defaultTreeSlug, defaultTreeSlug, limitPerTrack)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.CalibrationTrackSnapshot, 0)
	for rows.Next() {
		var item domain.CalibrationTrackSnapshot
		if err := rows.Scan(&item.TreeSlug, &item.SubmissionCount, &item.DeterministicScoreCount, &item.TopScoreCount, &item.AverageScore); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ListCalibrationHybridSignalSnapshots(ctx context.Context, defaultTreeSlug string, limitPerTrack int) ([]domain.CalibrationHybridSignalSnapshot, error) {
	if strings.TrimSpace(defaultTreeSlug) == "" {
		defaultTreeSlug = domain.GlobalSkillGraphSlug
	}
	if limitPerTrack <= 0 {
		limitPerTrack = 200
	}
	rows, err := s.SQL.QueryContext(ctx, `
		WITH track_ranked AS (
			SELECT
				s.id AS submission_id,
				COALESCE(NULLIF(t.slug, ''), NULLIF(u.active_tree_slug, ''), ?) AS tree_slug,
				ROW_NUMBER() OVER (
					PARTITION BY COALESCE(NULLIF(t.slug, ''), NULLIF(u.active_tree_slug, ''), ?)
					ORDER BY s.id DESC
				) AS track_rank
			FROM submissions s
			LEFT JOIN users u ON u.id = s.user_id
			LEFT JOIN tgo_trees t ON t.id = s.tree_id
		),
		sampled AS (
			SELECT submission_id, tree_slug
			FROM track_ranked
			WHERE track_rank <= ?
		),
		latest_hybrid AS (
			SELECT
				sss.submission_id,
				sss.skill_name,
				sss.score_evidence_json,
				ROW_NUMBER() OVER (
					PARTITION BY sss.submission_id, sss.skill_name
					ORDER BY sss.id DESC
				) AS rn
			FROM submission_skill_scores sss
			WHERE sss.score_source = 'hybrid'
		)
		SELECT
			sampled.tree_slug,
			COUNT(latest_hybrid.skill_name) AS hybrid_score_count,
			COALESCE(SUM(CASE WHEN json_extract(latest_hybrid.score_evidence_json, '$.conflict') = 1 THEN 1 ELSE 0 END), 0) AS conflict_count,
			COALESCE(SUM(CASE WHEN ABS(COALESCE(json_extract(latest_hybrid.score_evidence_json, '$.applied_delta'), 0)) > 0 THEN 1 ELSE 0 END), 0) AS adjusted_count
		FROM sampled
		LEFT JOIN latest_hybrid
			ON latest_hybrid.submission_id = sampled.submission_id
			AND latest_hybrid.rn = 1
		GROUP BY sampled.tree_slug
		ORDER BY hybrid_score_count DESC, sampled.tree_slug ASC
	`, defaultTreeSlug, defaultTreeSlug, limitPerTrack)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.CalibrationHybridSignalSnapshot, 0)
	for rows.Next() {
		var item domain.CalibrationHybridSignalSnapshot
		if err := rows.Scan(&item.TreeSlug, &item.HybridScoreCount, &item.ConflictCount, &item.AdjustedCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanCalibrationRun(scanner interface{ Scan(...any) error }) (domain.CalibrationRun, error) {
	var run domain.CalibrationRun
	var tracksJSON, domainsJSON, highlightsJSON, recommendationsJSON string
	var completedAt, approvedAt sql.NullTime
	var dataAdequate int
	if err := scanner.Scan(
		&run.ID,
		&run.RunKind,
		&run.Status,
		&run.TriggeredByUserID,
		&run.MinSamples,
		&run.LimitPerTrack,
		&run.SubmissionCount,
		&run.DeterministicScoreCount,
		&dataAdequate,
		&run.ApprovalStatus,
		&run.ApprovedByUserID,
		&approvedAt,
		&run.ApprovalNotes,
		&tracksJSON,
		&domainsJSON,
		&highlightsJSON,
		&recommendationsJSON,
		&run.ErrorText,
		&run.StartedAt,
		&completedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return domain.CalibrationRun{}, err
	}
	if completedAt.Valid {
		run.CompletedAt = completedAt.Time
	}
	run.DataAdequate = dataAdequate == 1
	if approvedAt.Valid {
		run.ApprovedAt = approvedAt.Time
	}
	if strings.TrimSpace(tracksJSON) != "" {
		_ = json.Unmarshal([]byte(tracksJSON), &run.TrackLearnings)
	}
	if strings.TrimSpace(domainsJSON) != "" {
		_ = json.Unmarshal([]byte(domainsJSON), &run.DomainLearnings)
	}
	if strings.TrimSpace(highlightsJSON) != "" {
		_ = json.Unmarshal([]byte(highlightsJSON), &run.Highlights)
	}
	if strings.TrimSpace(recommendationsJSON) != "" {
		_ = json.Unmarshal([]byte(recommendationsJSON), &run.Recommendations)
	}
	if run.TrackLearnings == nil {
		run.TrackLearnings = []domain.CalibrationTrackLearning{}
	}
	if run.DomainLearnings == nil {
		run.DomainLearnings = []domain.CalibrationDomainLearning{}
	}
	if run.Highlights == nil {
		run.Highlights = []string{}
	}
	if run.Recommendations == nil {
		run.Recommendations = []string{}
	}
	return run, nil
}

func scanAdminNotification(scanner interface{ Scan(...any) error }) (domain.AdminNotification, error) {
	var item domain.AdminNotification
	var isRead int
	var readAt sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&item.Kind,
		&item.Title,
		&item.Body,
		&item.PayloadJSON,
		&item.RelatedRunID,
		&isRead,
		&item.CreatedAt,
		&readAt,
	); err != nil {
		return domain.AdminNotification{}, err
	}
	item.IsRead = isRead == 1
	if readAt.Valid {
		item.ReadAt = readAt.Time
	}
	if strings.TrimSpace(item.PayloadJSON) == "" {
		item.PayloadJSON = "{}"
	}
	return item, nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}
