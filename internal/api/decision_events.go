package api

import (
	"context"
	"encoding/json"

	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/domain"
)

func (s Server) emitReviewDecisionEvents(ctx context.Context, userID, treeID, enrollmentID int64, reviewResult domain.Review, recommendation curriculum.Recommendation, scoreCount int) error {
	scoredPayload, _ := json.Marshal(map[string]any{
		"review_kind":       reviewResult.ReviewKind,
		"provider_note":     reviewResult.ProviderNote,
		"score_row_count":   scoreCount,
		"analyzer_findings": len(reviewResult.AnalyzerFindings),
	})
	if err := s.Store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            reviewResult.ID,
		SubmissionID:        reviewResult.SubmissionID,
		EventType:           "review_scored",
		DecisionPayloadJSON: string(scoredPayload),
		RuleVersion:         "deterministic-scoring-v1",
		EvidenceRefsJSON:    `["analyzer_report","submission_skill_scores"]`,
	}); err != nil {
		return err
	}

	recommendationPayload, _ := json.Marshal(map[string]any{
		"focus":            recommendation.Focus,
		"difficulty":       recommendation.Difficulty,
		"hold_active":      recommendation.HoldActive,
		"hold_reason_code": recommendation.HoldReasonCode,
		"rationale":        recommendation.Rationale,
	})
	return s.Store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            reviewResult.ID,
		SubmissionID:        reviewResult.SubmissionID,
		EventType:           "recommendation_issued",
		DecisionPayloadJSON: string(recommendationPayload),
		RuleVersion:         "curriculum-sync-v1",
		EvidenceRefsJSON:    `["recommendation","completed_tgo_checks","tgo_assessments"]`,
	})
}
