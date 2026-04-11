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

	return snapshot, nil
}
