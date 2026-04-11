package db

import (
	"context"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) SaveDecisionEvent(ctx context.Context, event domain.DecisionEvent) error {
	eventType := strings.TrimSpace(event.EventType)
	if eventType == "" {
		eventType = "unspecified"
	}
	payload := strings.TrimSpace(event.DecisionPayloadJSON)
	if payload == "" {
		payload = "{}"
	}
	ruleVersion := strings.TrimSpace(event.RuleVersion)
	evidenceRefs := strings.TrimSpace(event.EvidenceRefsJSON)
	if evidenceRefs == "" {
		evidenceRefs = "[]"
	}
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO decision_events (
			user_id,
			tree_id,
			enrollment_id,
			review_id,
			submission_id,
			event_type,
			decision_payload_json,
			rule_version,
			evidence_refs_json
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.UserID, event.TreeID, event.EnrollmentID, nullableID(event.ReviewID), nullableID(event.SubmissionID), eventType, payload, ruleVersion, evidenceRefs)
	return err
}

func (s *Store) DecisionEventsByReview(ctx context.Context, reviewID int64) ([]domain.DecisionEvent, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, enrollment_id, COALESCE(review_id, 0), COALESCE(submission_id, 0), event_type, decision_payload_json, rule_version, evidence_refs_json, created_at
		FROM decision_events
		WHERE review_id = ?
		ORDER BY id ASC
	`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.DecisionEvent, 0, 4)
	for rows.Next() {
		var item domain.DecisionEvent
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.TreeID,
			&item.EnrollmentID,
			&item.ReviewID,
			&item.SubmissionID,
			&item.EventType,
			&item.DecisionPayloadJSON,
			&item.RuleVersion,
			&item.EvidenceRefsJSON,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, item)
	}
	return events, rows.Err()
}
