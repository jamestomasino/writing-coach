package db

import "context"

func (s *Store) TransferTrackData(ctx context.Context, userID, sourceEnrollmentID, sourceTreeID, targetEnrollmentID, targetTreeID int64) (err error) {
	if sourceEnrollmentID == targetEnrollmentID || sourceTreeID == targetTreeID {
		return nil
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

	for _, table := range []string{
		"exercises",
		"submissions",
		"reviews",
		"playground_sessions",
		"playground_reviews",
		"playground_drafts",
	} {
		if _, err = tx.ExecContext(ctx, `
			UPDATE `+table+`
			SET tree_id = ?
			WHERE user_id = ? AND tree_id = ?
		`, targetTreeID, userID, sourceTreeID); err != nil {
			return err
		}
	}

	for _, table := range []string{"ai_jobs", "review_jobs"} {
		if _, err = tx.ExecContext(ctx, `
			UPDATE `+table+`
			SET tree_id = ?, enrollment_id = ?
			WHERE user_id = ? AND tree_id = ? AND enrollment_id = ?
		`, targetTreeID, targetEnrollmentID, userID, sourceTreeID, sourceEnrollmentID); err != nil {
			return err
		}
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code, completed_at)
		SELECT ?, tgo_code, completed_at
		FROM enrollment_completed_tgos
		WHERE enrollment_id = ?
		ON CONFLICT(enrollment_id, tgo_code) DO NOTHING
	`, targetEnrollmentID, sourceEnrollmentID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO user_curriculum_state (
			enrollment_id,
			current_focus,
			difficulty_level,
			last_review_id,
			progression_hold_active,
			progression_hold_reason_code,
			hold_trigger_review_id,
			hold_cleared_review_id,
			hold_updated_at,
			updated_at
		)
		SELECT
			?,
			current_focus,
			difficulty_level,
			last_review_id,
			COALESCE(progression_hold_active, 0),
			COALESCE(progression_hold_reason_code, ''),
			hold_trigger_review_id,
			hold_cleared_review_id,
			hold_updated_at,
			updated_at
		FROM user_curriculum_state
		WHERE enrollment_id = ?
		ON CONFLICT(enrollment_id) DO UPDATE SET
			current_focus = excluded.current_focus,
			difficulty_level = excluded.difficulty_level,
			last_review_id = excluded.last_review_id,
			progression_hold_active = excluded.progression_hold_active,
			progression_hold_reason_code = excluded.progression_hold_reason_code,
			hold_trigger_review_id = excluded.hold_trigger_review_id,
			hold_cleared_review_id = excluded.hold_cleared_review_id,
			hold_updated_at = excluded.hold_updated_at,
			updated_at = excluded.updated_at
	`, targetEnrollmentID, sourceEnrollmentID); err != nil {
		return err
	}

	return tx.Commit()
}
