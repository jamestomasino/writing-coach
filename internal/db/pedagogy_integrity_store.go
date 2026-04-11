package db

import (
	"context"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) PedagogyIntegritySnapshot(ctx context.Context, since time.Time, windowHours int) (domain.PedagogyIntegritySnapshot, error) {
	snapshot := domain.PedagogyIntegritySnapshot{
		Since:       since.UTC(),
		WindowHours: windowHours,
	}

	if err := s.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM reviews
		WHERE created_at >= ?
	`, since.UTC()).Scan(&snapshot.TotalReviews); err != nil {
		return domain.PedagogyIntegritySnapshot{}, err
	}

	if err := s.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT
				r.id,
				COALESCE(SUM(CASE WHEN e.event_type = 'review_scored' THEN 1 ELSE 0 END), 0) AS scored_count,
				COALESCE(SUM(CASE WHEN e.event_type = 'recommendation_issued' THEN 1 ELSE 0 END), 0) AS recommendation_count
			FROM reviews r
			LEFT JOIN decision_events e ON e.review_id = r.id
			WHERE r.created_at >= ?
			GROUP BY r.id
			HAVING scored_count = 0 OR recommendation_count = 0
		)
	`, since.UTC()).Scan(&snapshot.ReviewsMissingDecisionEvents); err != nil {
		return domain.PedagogyIntegritySnapshot{}, err
	}

	eventTypes := []struct {
		eventType string
		target    *int
	}{
		{eventType: "review_scored", target: &snapshot.ReviewScoredEvents},
		{eventType: "recommendation_issued", target: &snapshot.RecommendationEvents},
		{eventType: "progression_hold_activated", target: &snapshot.HoldActivationEvents},
		{eventType: "progression_hold_cleared", target: &snapshot.HoldClearEvents},
		{eventType: "progression_hold_blocked", target: &snapshot.HoldBlockedEvents},
	}
	for _, item := range eventTypes {
		if err := s.SQL.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM decision_events
			WHERE event_type = ? AND created_at >= ?
		`, item.eventType, since.UTC()).Scan(item.target); err != nil {
			return domain.PedagogyIntegritySnapshot{}, err
		}
	}

	if err := s.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_curriculum_state
		WHERE progression_hold_active = 1
	`).Scan(&snapshot.ActiveHoldEnrollments); err != nil {
		return domain.PedagogyIntegritySnapshot{}, err
	}

	if err := s.SQL.QueryRowContext(ctx, `
		SELECT COALESCE(AVG((julianday(clear_at) - julianday(activated_at)) * 24.0), 0)
		FROM (
			SELECT
				e.created_at AS clear_at,
				(
					SELECT MAX(ea.created_at)
					FROM decision_events ea
					WHERE ea.enrollment_id = e.enrollment_id
					  AND ea.event_type = 'progression_hold_activated'
					  AND ea.created_at <= e.created_at
				) AS activated_at
			FROM decision_events e
			WHERE e.event_type = 'progression_hold_cleared'
			  AND e.created_at >= ?
		)
		WHERE activated_at IS NOT NULL
	`, since.UTC()).Scan(&snapshot.AvgHoldClearHours); err != nil {
		return domain.PedagogyIntegritySnapshot{}, err
	}

	if err := s.SQL.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN json_extract(io.value, '$.status') = 'resolved' THEN 1 ELSE 0 END), 0) AS resolved_count,
			COALESCE(SUM(CASE WHEN json_extract(io.value, '$.status') = 'persisting' THEN 1 ELSE 0 END), 0) AS persisting_count
		FROM review_artifacts ra
		JOIN reviews r ON r.id = ra.review_id
		LEFT JOIN json_each(ra.recommendation_json, '$.intervention_outcomes') io
		WHERE r.created_at >= ?
	`, since.UTC()).Scan(&snapshot.InterventionResolvedCount, &snapshot.InterventionPersistingCount); err != nil {
		return domain.PedagogyIntegritySnapshot{}, err
	}

	outcomeTotal := snapshot.InterventionResolvedCount + snapshot.InterventionPersistingCount
	if outcomeTotal > 0 {
		snapshot.InterventionResolutionRate = (float64(snapshot.InterventionResolvedCount) / float64(outcomeTotal)) * 100.0
		snapshot.InterventionRecurrenceRate = (float64(snapshot.InterventionPersistingCount) / float64(outcomeTotal)) * 100.0
	}

	if err := s.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM enrollment_completed_tgos
		WHERE completed_at >= ?
	`, since.UTC()).Scan(&snapshot.MasteryCompletions); err != nil {
		return domain.PedagogyIntegritySnapshot{}, err
	}
	if snapshot.TotalReviews > 0 {
		snapshot.MasteryVelocityPer100Reviews = (float64(snapshot.MasteryCompletions) / float64(snapshot.TotalReviews)) * 100.0
	}

	return snapshot, nil
}
