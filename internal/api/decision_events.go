package api

import (
	"context"
	"encoding/json"

	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/domain"
)

func reviewScoredEvidenceRefsJSON(skillScoreCount, objectiveScoreCount int) string {
	refs := []string{"analyzer_report"}
	if skillScoreCount > 0 {
		refs = append(refs, "submission_skill_scores")
	}
	if objectiveScoreCount > 0 {
		refs = append(refs, "submission_objective_scores")
	}
	return string(mustJSON(refs))
}

func (s Server) emitReviewDecisionEvents(
	ctx context.Context,
	userID, treeID, enrollmentID int64,
	reviewResult domain.Review,
	recommendation curriculum.Recommendation,
	skillScoreCount int,
	objectiveScoreCount int,
) error {
	scoreCount := skillScoreCount
	if objectiveScoreCount > 0 {
		scoreCount = objectiveScoreCount
	}
	scoredPayload, _ := json.Marshal(map[string]any{
		"review_kind":               reviewResult.ReviewKind,
		"provider_note":             reviewResult.ProviderNote,
		"score_row_count":           scoreCount,
		"skill_score_row_count":     skillScoreCount,
		"objective_score_row_count": objectiveScoreCount,
		"analyzer_findings":         len(reviewResult.AnalyzerFindings),
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
		EvidenceRefsJSON:    reviewScoredEvidenceRefsJSON(skillScoreCount, objectiveScoreCount),
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

func (s Server) emitProgressionHoldTransitionEvent(ctx context.Context, userID, treeID, enrollmentID int64, review domain.Review, wasHoldActive bool, isHoldActive bool, reasonCode string) error {
	if wasHoldActive == isHoldActive {
		return nil
	}
	eventType := "progression_hold_cleared"
	ruleVersion := "progression-hold-clear-v1"
	if isHoldActive {
		eventType = "progression_hold_activated"
		ruleVersion = "progression-hold-activate-v1"
	}
	payload, _ := json.Marshal(map[string]any{
		"was_hold_active": wasHoldActive,
		"is_hold_active":  isHoldActive,
		"reason_code":     reasonCode,
	})
	return s.Store.SaveDecisionEvent(ctx, domain.DecisionEvent{
		UserID:              userID,
		TreeID:              treeID,
		EnrollmentID:        enrollmentID,
		ReviewID:            review.ID,
		SubmissionID:        review.SubmissionID,
		EventType:           eventType,
		DecisionPayloadJSON: string(payload),
		RuleVersion:         ruleVersion,
		EvidenceRefsJSON:    `["completed_tgo_checks","user_curriculum_state"]`,
	})
}
