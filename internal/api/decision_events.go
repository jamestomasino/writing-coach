package api

import (
	"context"
	"encoding/json"
	"strings"

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
	if err := s.emitReviewScoredDecisionEvent(
		ctx,
		userID,
		treeID,
		enrollmentID,
		reviewResult,
		skillScoreCount,
		objectiveScoreCount,
	); err != nil {
		return err
	}
	return s.emitRecommendationIssuedDecisionEvent(ctx, userID, treeID, enrollmentID, reviewResult, recommendation)
}

func (s Server) emitReviewScoredDecisionEvent(
	ctx context.Context,
	userID, treeID, enrollmentID int64,
	reviewResult domain.Review,
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
	return nil
}

func (s Server) emitRecommendationIssuedDecisionEvent(
	ctx context.Context,
	userID, treeID, enrollmentID int64,
	reviewResult domain.Review,
	recommendation curriculum.Recommendation,
) error {
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

func (s Server) ensureReviewDecisionEvents(
	ctx context.Context,
	userID, treeID, enrollmentID int64,
	reviewResult domain.Review,
) error {
	events, err := s.Store.DecisionEventsByReview(ctx, reviewResult.ID)
	if err != nil {
		return err
	}
	hasScored := false
	hasRecommendation := false
	for _, event := range events {
		switch strings.TrimSpace(event.EventType) {
		case "review_scored":
			hasScored = true
		case "recommendation_issued":
			hasRecommendation = true
		}
	}
	if hasScored && hasRecommendation {
		return nil
	}

	skillScores, err := s.Store.SubmissionSkillScores(ctx, reviewResult.SubmissionID)
	if err != nil {
		return err
	}
	objectiveScores, err := s.Store.SubmissionObjectiveScores(ctx, reviewResult.SubmissionID)
	if err != nil {
		return err
	}

	recommendation := curriculum.Recommendation{}
	if artifacts, err := s.Store.GetReviewArtifacts(ctx, reviewResult.ID); err == nil {
		var payload struct {
			Focus          string `json:"focus"`
			Difficulty     int    `json:"difficulty"`
			HoldActive     bool   `json:"hold_active"`
			HoldReasonCode string `json:"hold_reason_code"`
			Rationale      string `json:"rationale"`
		}
		if json.Unmarshal([]byte(artifacts.RecommendationJSON), &payload) == nil {
			recommendation = curriculum.Recommendation{
				Focus:          strings.TrimSpace(payload.Focus),
				Difficulty:     payload.Difficulty,
				HoldActive:     payload.HoldActive,
				HoldReasonCode: strings.TrimSpace(payload.HoldReasonCode),
				Rationale:      strings.TrimSpace(payload.Rationale),
			}
		}
	}

	if !hasScored {
		if err := s.emitReviewScoredDecisionEvent(ctx, userID, treeID, enrollmentID, reviewResult, len(skillScores), len(objectiveScores)); err != nil {
			return err
		}
	}
	if !hasRecommendation {
		if err := s.emitRecommendationIssuedDecisionEvent(ctx, userID, treeID, enrollmentID, reviewResult, recommendation); err != nil {
			return err
		}
	}
	return nil
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
